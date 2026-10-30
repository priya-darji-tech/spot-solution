package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"spot-solution/internal/config"
	"spot-solution/internal/database"
	"spot-solution/internal/email"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
)

type ContactService struct {
	cfg         *config.Config
	db          *database.DB
	contactRepo *repository.ContactRepository
	crmService  CRMService
	outboxRepo  *repository.OutboxRepository
	mailer      *email.Mailer
}

func NewContactService(
	cfg *config.Config,
	db *database.DB,
	contactRepo *repository.ContactRepository,
	crmService CRMService,
	outboxRepo *repository.OutboxRepository,
	mailer *email.Mailer,
) *ContactService {
	return &ContactService{
		cfg:         cfg,
		db:          db,
		contactRepo: contactRepo,
		crmService:  crmService,
		outboxRepo:  outboxRepo,
		mailer:      mailer,
	}
}

func (s *ContactService) ProcessSubmission(ctx context.Context, req *models.ContactRequest, ip, userAgent string) (*models.ContactSubmission, error) {
	// 1. Sanitize Inputs
	cleanFirst := sanitizeString(req.FirstName)
	cleanLast := sanitizeString(req.LastName)
	cleanCompany := sanitizeString(req.Company)
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanPhone := sanitizeString(req.Phone)
	cleanService := sanitizeString(req.Service)
	cleanMessage := html.EscapeString(strings.TrimSpace(req.Message))

	// 2. Generate Deduplication Hash (email + message)
	hashInput := fmt.Sprintf("%s|%s", cleanEmail, cleanMessage)
	h := sha256.Sum256([]byte(hashInput))
	subHash := hex.EncodeToString(h[:])

	// 3. Deduplication Check (within configured time window)
	window := time.Duration(s.cfg.ContactRateLimitMinutes) * time.Minute
	if window < 1*time.Minute {
		window = 5 * time.Minute
	}

	isDuplicate, err := s.contactRepo.HasRecentDuplicate(ctx, subHash, window)
	if err != nil {
		return nil, fmt.Errorf("duplicate check error: %w", err)
	}
	if isDuplicate {
		return nil, fmt.Errorf("a duplicate inquiry was recently submitted. Please wait before submitting again")
	}

	// 4. Prepare Models
	sub := &models.ContactSubmission{
		ID:             uuid.New(),
		FirstName:      cleanFirst,
		LastName:       cleanLast,
		CompanyName:    cleanCompany,
		Email:          cleanEmail,
		Phone:          cleanPhone,
		Service:        cleanService,
		Message:        cleanMessage,
		SubmissionHash: subHash,
		IPAddress:      ip,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
	}

	adminEmailHTML := s.mailer.BuildAdminNotificationHTML(sub)
	customerEmailHTML := s.mailer.BuildCustomerAcknowledgmentHTML(sub)

	adminOutbox := &models.EmailOutboxItem{
		ID:             uuid.New(),
		SubmissionID:   &sub.ID,
		RecipientEmail: s.cfg.NotificationRecipient,
		RecipientName:  "Admin Team",
		Subject:        fmt.Sprintf("New Inquiry: %s (%s)", sub.Service, sub.CompanyName),
		BodyHTML:       adminEmailHTML,
		Status:         models.OutboxPending,
		RetryCount:     0,
		MaxRetries:     s.cfg.EmailMaxRetries,
		NextRetryAt:    time.Now(),
		CreatedAt:      time.Now(),
	}

	customerOutbox := &models.EmailOutboxItem{
		ID:             uuid.New(),
		SubmissionID:   &sub.ID,
		RecipientEmail: sub.Email,
		RecipientName:  fmt.Sprintf("%s %s", sub.FirstName, sub.LastName),
		Subject:        fmt.Sprintf("We received your inquiry - %s", s.cfg.AppName),
		BodyHTML:       customerEmailHTML,
		Status:         models.OutboxPending,
		RetryCount:     0,
		MaxRetries:     s.cfg.EmailMaxRetries,
		NextRetryAt:    time.Now(),
		CreatedAt:      time.Now(),
	}

	// 5. Execute Atomic Database Transaction (Submission + CRM Lead + Outbox Emails)
	err = s.db.ExecTx(ctx, func(tx *sql.Tx) error {
		// A. Save Contact Submission
		if err := s.contactRepo.CreateSubmissionTx(ctx, tx, sub); err != nil {
			return err
		}

		// B. Save CRM Lead
		if _, err := s.crmService.RecordLeadFromSubmissionTx(ctx, tx, sub); err != nil {
			return err
		}

		// C. Queue Admin Notification Email
		if err := s.outboxRepo.QueueEmailTx(ctx, tx, adminOutbox); err != nil {
			return err
		}

		// D. Queue Customer Acknowledgment Email
		if err := s.outboxRepo.QueueEmailTx(ctx, tx, customerOutbox); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed during inquiry processing: %w", err)
	}

	return sub, nil
}

func sanitizeString(input string) string {
	cleaned := strings.TrimSpace(input)
	cleaned = html.EscapeString(cleaned)
	return cleaned
}

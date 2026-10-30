package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
)

type LeadService struct {
	repo *repository.LeadRepository
}

func NewLeadService(repo *repository.LeadRepository) *LeadService {
	return &LeadService{repo: repo}
}

func (s *LeadService) RecordLeadFromSubmissionTx(ctx context.Context, tx *sql.Tx, sub *models.ContactSubmission) (*models.Lead, error) {
	lead := &models.Lead{
		ID:           uuid.New(),
		SubmissionID: &sub.ID,
		FirstName:    sub.FirstName,
		LastName:     sub.LastName,
		CompanyName:  sub.CompanyName,
		Email:        sub.Email,
		Phone:        sub.Phone,
		Service:      sub.Service,
		Message:      sub.Message,
		Source:       "Website",
		Status:       models.LeadStatusNew,
		Notes:        fmt.Sprintf("Auto-generated lead from website inquiry on service: %s", sub.Service),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateLeadTx(ctx, tx, lead); err != nil {
		return nil, fmt.Errorf("error creating lead in tx: %w", err)
	}

	return lead, nil
}

func (s *LeadService) CreateLead(ctx context.Context, lead *models.Lead) error {
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}
	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()
	if lead.Status == "" {
		lead.Status = models.LeadStatusNew
	}
	if lead.Source == "" {
		lead.Source = "Direct"
	}
	return s.repo.CreateLead(ctx, lead)
}

func (s *LeadService) GetLeads(ctx context.Context, page, pageSize int, statusFilter, searchQuery, sortBy, sortDir string) ([]models.Lead, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset, statusFilter, searchQuery, sortBy, sortDir)
}

func (s *LeadService) GetLeadByID(ctx context.Context, idStr string) (*models.Lead, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid lead UUID: %w", err)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *LeadService) UpdateLeadStatus(ctx context.Context, idStr string, status models.LeadStatus, notes string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid lead UUID: %w", err)
	}
	return s.repo.UpdateStatus(ctx, id, status, notes)
}

func (s *LeadService) DeleteLead(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid lead UUID: %w", err)
	}
	return s.repo.Delete(ctx, id)
}

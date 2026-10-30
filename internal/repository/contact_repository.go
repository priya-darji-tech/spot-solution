package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"spot-solution/internal/database"
	"spot-solution/internal/models"
)

type ContactRepository struct {
	db *database.DB
}

func NewContactRepository(db *database.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) CreateSubmissionTx(ctx context.Context, tx *sql.Tx, s *models.ContactSubmission) error {
	query := `
		INSERT INTO contact_submissions (
			id, first_name, last_name, company_name, email, phone, service, message, submission_hash, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := tx.ExecContext(ctx, query,
		s.ID, s.FirstName, s.LastName, s.CompanyName, s.Email, s.Phone,
		s.Service, s.Message, s.SubmissionHash, s.IPAddress, s.UserAgent, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert contact submission: %w", err)
	}
	return nil
}

func (r *ContactRepository) HasRecentDuplicate(ctx context.Context, hash string, window time.Duration) (bool, error) {
	cutoff := time.Now().Add(-window)
	query := `SELECT EXISTS(SELECT 1 FROM contact_submissions WHERE submission_hash = $1 AND created_at >= $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, hash, cutoff).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate submission: %w", err)
	}
	return exists, nil
}

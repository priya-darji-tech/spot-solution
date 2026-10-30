package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"spot-solution/internal/database"
	"spot-solution/internal/models"
)

type LeadRepository struct {
	db *database.DB
}

func NewLeadRepository(db *database.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

func (r *LeadRepository) CreateLeadTx(ctx context.Context, tx *sql.Tx, lead *models.Lead) error {
	query := `
		INSERT INTO leads (
			id, submission_id, first_name, last_name, company_name, email, phone, service, message, source, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := tx.ExecContext(ctx, query,
		lead.ID, lead.SubmissionID, lead.FirstName, lead.LastName, lead.CompanyName,
		lead.Email, lead.Phone, lead.Service, lead.Message, lead.Source,
		lead.Status, lead.Notes, lead.CreatedAt, lead.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) CreateLead(ctx context.Context, lead *models.Lead) error {
	query := `
		INSERT INTO leads (
			id, submission_id, first_name, last_name, company_name, email, phone, service, message, source, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(ctx, query,
		lead.ID, lead.SubmissionID, lead.FirstName, lead.LastName, lead.CompanyName,
		lead.Email, lead.Phone, lead.Service, lead.Message, lead.Source,
		lead.Status, lead.Notes, lead.CreatedAt, lead.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert lead: %w", err)
	}
	return nil
}

func sanitizeSort(sortBy, sortDir string) (string, string) {
	var col string
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "name", "first_name":
		col = "first_name"
	case "status":
		col = "status"
	case "company", "company_name":
		col = "company_name"
	case "service":
		col = "service"
	case "updated_at":
		col = "updated_at"
	default:
		col = "created_at"
	}

	var dir string
	switch strings.ToLower(strings.TrimSpace(sortDir)) {
	case "asc":
		dir = "ASC"
	default:
		dir = "DESC"
	}

	return col, dir
}

func (r *LeadRepository) GetAll(ctx context.Context, limit, offset int, statusFilter, searchQuery, sortBy, sortDir string) ([]models.Lead, int, error) {
	sortCol, sortDirection := sanitizeSort(sortBy, sortDir)

	searchPattern := ""
	if searchQuery != "" {
		searchPattern = "%" + searchQuery + "%"
	}

	countQuery := `
		SELECT COUNT(*)
		FROM leads
		WHERE ($1 = '' OR status::text = $1)
		  AND ($2 = '' OR first_name ILIKE $3 OR last_name ILIKE $3 OR company_name ILIKE $3 OR email ILIKE $3 OR message ILIKE $3)
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, statusFilter, searchQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leads: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, submission_id, first_name, last_name, company_name, email, phone, service, message, source, status, COALESCE(notes, ''), created_at, updated_at
		FROM leads
		WHERE ($1 = '' OR status::text = $1)
		  AND ($2 = '' OR first_name ILIKE $3 OR last_name ILIKE $3 OR company_name ILIKE $3 OR email ILIKE $3 OR message ILIKE $3)
		ORDER BY %s %s
		LIMIT $4 OFFSET $5
	`, sortCol, sortDirection)

	rows, err := r.db.QueryContext(ctx, query, statusFilter, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch leads: %w", err)
	}
	defer rows.Close()

	var leads []models.Lead
	for rows.Next() {
		var l models.Lead
		err := rows.Scan(
			&l.ID, &l.SubmissionID, &l.FirstName, &l.LastName, &l.CompanyName,
			&l.Email, &l.Phone, &l.Service, &l.Message, &l.Source,
			&l.Status, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan lead row: %w", err)
		}
		leads = append(leads, l)
	}

	return leads, total, nil
}

func (r *LeadRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lead, error) {
	query := `
		SELECT id, submission_id, first_name, last_name, company_name, email, phone, service, message, source, status, COALESCE(notes, ''), created_at, updated_at
		FROM leads
		WHERE id = $1
	`
	var l models.Lead
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.SubmissionID, &l.FirstName, &l.LastName, &l.CompanyName,
		&l.Email, &l.Phone, &l.Service, &l.Message, &l.Source,
		&l.Status, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lead by id: %w", err)
	}
	return &l, nil
}

func (r *LeadRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.LeadStatus, notes string) error {
	var res sql.Result
	var err error
	if status == "" {
		query := `
			UPDATE leads
			SET notes = CASE WHEN $1 != '' THEN $1 ELSE notes END, updated_at = $2
			WHERE id = $3
		`
		res, err = r.db.ExecContext(ctx, query, notes, time.Now(), id)
	} else {
		query := `
			UPDATE leads
			SET status = $1, notes = CASE WHEN $2 != '' THEN $2 ELSE notes END, updated_at = $3
			WHERE id = $4
		`
		res, err = r.db.ExecContext(ctx, query, status, notes, time.Now(), id)
	}
	if err != nil {
		return fmt.Errorf("failed to update lead status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("lead not found")
	}
	return nil
}

func (r *LeadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM leads WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("lead not found")
	}
	return nil
}

package services

import (
	"context"
	"database/sql"

	"spot-solution/internal/models"
)

// CRMService defines the interface for creating and managing leads,
// allowing external CRM integrations (HubSpot, Salesforce, etc.) in the future.
type CRMService interface {
	RecordLeadFromSubmissionTx(ctx context.Context, tx *sql.Tx, sub *models.ContactSubmission) (*models.Lead, error)
	CreateLead(ctx context.Context, lead *models.Lead) error
	GetLeads(ctx context.Context, page, pageSize int, statusFilter, searchQuery, sortBy, sortDir string) ([]models.Lead, int, error)
	GetLeadByID(ctx context.Context, id string) (*models.Lead, error)
	UpdateLeadStatus(ctx context.Context, id string, status models.LeadStatus, notes string) error
	DeleteLead(ctx context.Context, id string) error
}

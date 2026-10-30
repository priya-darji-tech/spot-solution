package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleViewer  UserRole = "viewer"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	TokenHash string    `json:"token_hash"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ContactRequest struct {
	FirstName string `json:"first_name" form:"first_name" binding:"required,min=2,max=100"`
	LastName  string `json:"last_name" form:"last_name" binding:"required,min=2,max=100"`
	Company   string `json:"company_name" form:"company_name" binding:"max=150"`
	Email     string `json:"email" form:"email" binding:"required,email,max=255"`
	Phone     string `json:"phone" form:"phone" binding:"max=50"`
	Service   string `json:"service" form:"service" binding:"required,max=100"`
	Message   string `json:"message" form:"message" binding:"required,min=10,max=5000"`
	CSRFToken string `json:"csrf_token" form:"csrf_token"`
}

type ContactSubmission struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CompanyName    string    `json:"company_name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Service        string    `json:"service"`
	Message        string    `json:"message"`
	SubmissionHash string    `json:"submission_hash"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	CreatedAt      time.Time `json:"created_at"`
}

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "New"
	LeadStatusContacted LeadStatus = "Contacted"
	LeadStatusQualified LeadStatus = "Qualified"
	LeadStatusConverted LeadStatus = "Converted"
	LeadStatusLost      LeadStatus = "Lost"
)

type Lead struct {
	ID           uuid.UUID  `json:"id"`
	SubmissionID *uuid.UUID `json:"submission_id,omitempty"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	CompanyName  string     `json:"company_name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Service      string     `json:"service"`
	Message      string     `json:"message"`
	Source       string     `json:"source"`
	Status       LeadStatus `json:"status"`
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UpdateLeadStatusRequest struct {
	Status LeadStatus `json:"status" binding:"omitempty,oneof=New Contacted Qualified Converted Lost"`
	Notes  string     `json:"notes"`
}

type CreateLeadRequest struct {
	FirstName   string `json:"first_name" binding:"required,min=2,max=100"`
	LastName    string `json:"last_name" binding:"required,min=2,max=100"`
	CompanyName string `json:"company_name" binding:"max=150"`
	Email       string `json:"email" binding:"required,email,max=255"`
	Phone       string `json:"phone" binding:"max=50"`
	Service     string `json:"service" binding:"required,max=100"`
	Message     string `json:"message"`
	Source      string `json:"source" binding:"max=50"`
	Notes       string `json:"notes"`
}

type OutboxStatus string

const (
	OutboxPending    OutboxStatus = "pending"
	OutboxProcessing OutboxStatus = "processing"
	OutboxSent       OutboxStatus = "sent"
	OutboxFailed     OutboxStatus = "failed"
)

type EmailOutboxItem struct {
	ID             uuid.UUID    `json:"id"`
	SubmissionID   *uuid.UUID   `json:"submission_id,omitempty"`
	RecipientEmail string       `json:"recipient_email"`
	RecipientName  string       `json:"recipient_name"`
	Subject        string       `json:"subject"`
	BodyHTML       string       `json:"body_html"`
	Status         OutboxStatus `json:"status"`
	RetryCount     int          `json:"retry_count"`
	MaxRetries     int          `json:"max_retries"`
	NextRetryAt    time.Time    `json:"next_retry_at"`
	LastError      *string      `json:"last_error,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	SentAt         *time.Time   `json:"sent_at,omitempty"`
}

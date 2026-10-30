package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"spot-solution/internal/database"
	"spot-solution/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = TRUE
	`
	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &roleStr, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	u.Role = models.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`
	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &roleStr, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	u.Role = models.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, s *models.Session) error {
	query := `
		INSERT INTO sessions (token_hash, user_id, ip_address, user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, s.TokenHash, s.UserID, s.IPAddress, s.UserAgent, s.ExpiresAt, s.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (r *UserRepository) GetSession(ctx context.Context, tokenHash string) (*models.Session, *models.User, error) {
	query := `
		SELECT s.token_hash, s.user_id, s.ip_address, s.user_agent, s.expires_at, s.created_at,
		       u.id, u.email, u.full_name, u.role, u.is_active, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token_hash = $1 AND s.expires_at > $2 AND u.is_active = TRUE
	`
	var s models.Session
	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, tokenHash, time.Now()).Scan(
		&s.TokenHash, &s.UserID, &s.IPAddress, &s.UserAgent, &s.ExpiresAt, &s.CreatedAt,
		&u.ID, &u.Email, &u.FullName, &roleStr, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve session: %w", err)
	}
	u.Role = models.UserRole(roleStr)
	return &s, &u, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}

func (r *UserRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

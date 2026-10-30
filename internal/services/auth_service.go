package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"spot-solution/internal/config"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, email, password, ip, userAgent string) (string, *models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("database query error: %w", err)
	}
	if user == nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	// Generate raw session token
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", nil, fmt.Errorf("error generating session token: %w", err)
	}
	rawToken := hex.EncodeToString(rawBytes)

	// Hash token for database storage
	tokenHash := HashSessionToken(rawToken)

	session := &models.Session{
		TokenHash: tokenHash,
		UserID:    user.ID,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.SessionMaxAgeSeconds) * time.Second),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return "", nil, fmt.Errorf("error saving session: %w", err)
	}

	return rawToken, user, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, rawToken string) (*models.Session, *models.User, error) {
	if rawToken == "" {
		return nil, nil, nil
	}
	tokenHash := HashSessionToken(rawToken)
	return s.userRepo.GetSession(ctx, tokenHash)
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	tokenHash := HashSessionToken(rawToken)
	return s.userRepo.DeleteSession(ctx, tokenHash)
}

func HashSessionToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

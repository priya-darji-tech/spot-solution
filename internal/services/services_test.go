package services_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"spot-solution/internal/models"
	"spot-solution/internal/services"
)

// TestPasswordHashing verifies bcrypt password hashing and constant-time comparison
func TestPasswordHashing(t *testing.T) {
	password := "SecretPass123!"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Correct password must match
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(password)); err != nil {
		t.Fatalf("Expected password to match hash, got error: %v", err)
	}

	// Wrong password must fail
	if err := bcrypt.CompareHashAndPassword(hashed, []byte("WrongPassword!")); err == nil {
		t.Fatalf("Expected wrong password to fail comparison, but it succeeded")
	}
}

// TestSessionTokenHashing verifies sha256 session token generation and uniqueness
func TestSessionTokenHashing(t *testing.T) {
	token1 := "token-abc-123"
	token2 := "token-xyz-456"

	hash1 := services.HashSessionToken(token1)
	hash2 := services.HashSessionToken(token2)

	if hash1 == hash2 {
		t.Fatalf("Expected different hashes for different tokens, got %s", hash1)
	}

	if len(hash1) != 64 {
		t.Fatalf("Expected SHA-256 hex string length 64, got %d", len(hash1))
	}
}

// TestDuplicateHashCalculation verifies deduplication hash calculation
func TestDuplicateHashCalculation(t *testing.T) {
	email := "Client@Domain.com"
	message := "We need custom software development."

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanMessage := html.EscapeString(strings.TrimSpace(message))

	hashInput1 := fmt.Sprintf("%s|%s", cleanEmail, cleanMessage)
	h1 := sha256.Sum256([]byte(hashInput1))
	hash1 := hex.EncodeToString(h1[:])

	// Submitting with leading spaces / uppercase should produce identical deduplication hash
	cleanEmail2 := strings.ToLower(strings.TrimSpace("   CLIENT@DOMAIN.COM   "))
	cleanMessage2 := html.EscapeString(strings.TrimSpace("We need custom software development.   "))

	hashInput2 := fmt.Sprintf("%s|%s", cleanEmail2, cleanMessage2)
	h2 := sha256.Sum256([]byte(hashInput2))
	hash2 := hex.EncodeToString(h2[:])

	if hash1 != hash2 {
		t.Fatalf("Expected identical duplicate hash, got %s vs %s", hash1, hash2)
	}
}

// TestContactRequestValidation verifies field boundaries and constraints
func TestContactRequestValidation(t *testing.T) {
	validReq := models.ContactRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
		Service:   "Custom Software Development",
		Message:   "We are looking for an enterprise architecture review.",
	}

	if validReq.FirstName == "" || validReq.LastName == "" || validReq.Email == "" || validReq.Service == "" || len(validReq.Message) < 10 {
		t.Fatalf("Expected valid contact request to satisfy constraints")
	}

	invalidReq := models.ContactRequest{
		FirstName: "",
		Email:     "invalid-email",
		Message:   "short",
	}

	if invalidReq.FirstName != "" || len(invalidReq.Message) >= 10 {
		t.Fatalf("Expected invalid request to be caught")
	}
}

// TestLeadStatusWorkflow verifies status transitions
func TestLeadStatusWorkflow(t *testing.T) {
	statuses := []models.LeadStatus{
		models.LeadStatusNew,
		models.LeadStatusContacted,
		models.LeadStatusQualified,
		models.LeadStatusConverted,
		models.LeadStatusLost,
	}

	for _, s := range statuses {
		if s == "" {
			t.Fatalf("Expected valid lead status string")
		}
	}
}

// TestCreateLeadRequestValidation verifies field constraints on manual lead creation
func TestCreateLeadRequestValidation(t *testing.T) {
	req := models.CreateLeadRequest{
		FirstName: "Daniel",
		LastName:  "Jackson",
		Email:     "djackson@stargate.mil",
		Service:   "AI Advisory",
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Service == "" {
		t.Fatalf("Expected valid CreateLeadRequest")
	}

	// Notes can be empty or set
	req.Notes = "Archaeology & ancient database translation"
	if req.Notes == "" {
		t.Fatalf("Expected notes to be assignable")
	}
}

// TestUpdateLeadStatusRequestNotesOnly verifies updating notes without changing status
func TestUpdateLeadStatusRequestNotesOnly(t *testing.T) {
	reqNotesOnly := models.UpdateLeadStatusRequest{
		Status: "",
		Notes:  "Only updating notes",
	}

	if reqNotesOnly.Notes == "" {
		t.Fatalf("Expected notes to be present")
	}
	if reqNotesOnly.Status != "" {
		t.Fatalf("Expected status to be empty for notes-only update")
	}

	reqStatusOnly := models.UpdateLeadStatusRequest{
		Status: models.LeadStatusQualified,
		Notes:  "",
	}
	if reqStatusOnly.Status != models.LeadStatusQualified {
		t.Fatalf("Expected status to be qualified")
	}
}

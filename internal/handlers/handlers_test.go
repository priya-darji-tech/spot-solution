package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"spot-solution/internal/config"
	"spot-solution/internal/database"
	"spot-solution/internal/email"
	"spot-solution/internal/handlers"
	"spot-solution/internal/middleware"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
	"spot-solution/internal/services"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *database.DB, *config.Config) {
	gin.SetMode(gin.TestMode)
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed loading config: %v", err)
	}

	db, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skipping integration test; PostgreSQL not accessible: %v", err)
	}

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	_ = db.SeedDefaultAdmin(cfg)

	contactRepo := repository.NewContactRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	userRepo := repository.NewUserRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	mailer := email.NewMailer(cfg)
	crmService := services.NewLeadService(leadRepo)
	contactService := services.NewContactService(cfg, db, contactRepo, crmService, outboxRepo, mailer)
	authService := services.NewAuthService(cfg, userRepo)

	contactHandler := handlers.NewContactHandler(contactService)
	authHandler := handlers.NewAuthHandler(cfg, authService)
	leadHandler := handlers.NewLeadHandler(crmService)

	r := gin.New()
	r.Use(middleware.CSRFMiddleware(cfg))
	r.Use(middleware.AuthMiddleware(authService))

	// Routes under test
	r.POST("/api/contact", contactHandler.SubmitContact)
	r.POST("/api/admin/login", authHandler.Login)
	r.GET("/admin/leads", middleware.RequireAuth(), leadHandler.AdminDashboard)

	apiLeads := r.Group("/api/leads")
	apiLeads.Use(middleware.RequireAuth())
	{
		apiLeads.GET("", leadHandler.GetLeads)
		apiLeads.GET("/:id", leadHandler.GetLeadByID)
		apiLeads.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleManager), leadHandler.CreateLead)
		apiLeads.PATCH("/:id/status", middleware.RequireRole(models.RoleAdmin, models.RoleManager), leadHandler.UpdateStatus)
		apiLeads.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), leadHandler.DeleteLead)
	}

	return r, db, cfg
}

func extractCookiesAndCSRF(w *httptest.ResponseRecorder) (cookieHeader, csrfToken string) {
	var parts []string
	for _, c := range w.Result().Cookies() {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
		if c.Name == "csrf_token" {
			csrfToken = c.Value
		}
	}
	return strings.Join(parts, "; "), csrfToken
}

func createTestUser(t *testing.T, db *database.DB, email, password string, role models.UserRole) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password for %s: %v", email, err)
	}
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO users (email, password_hash, full_name, role, is_active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (email) DO UPDATE SET password_hash = $2, role = $4
	`, email, string(hashed), "Test "+string(role), string(role))
	if err != nil {
		t.Fatalf("Failed to create test user %s: %v", email, err)
	}
}

func loginAs(t *testing.T, r *gin.Engine, email, password string) (cookieHeader, csrfToken string) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Login failed for %s: %d %s", email, w.Code, w.Body.String())
	}
	return extractCookiesAndCSRF(w)
}

func TestContactSubmissionAPI(t *testing.T) {
	r, db, _ := setupTestRouter(t)
	defer db.Close()

	// 1. Valid submission
	validBody := map[string]string{
		"first_name":   "Integration",
		"last_name":    "Tester",
		"company_name": "Test Co",
		"email":        fmt.Sprintf("test-inquiry-%d@example.com", time.Now().UnixNano()),
		"phone":        "+1-555-0199",
		"service":      "Custom Software Development",
		"message":      "This is an automated integration test message with sufficient length.",
	}
	jsonBytes, _ := json.Marshal(validBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/contact", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for valid contact submission, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Duplicate submission test (immediately resubmitting the same payload must be rejected)
	reqDup, _ := http.NewRequest(http.MethodPost, "/api/contact", bytes.NewBuffer(jsonBytes))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	r.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("Expected HTTP 400 for duplicate contact submission, got %d", wDup.Code)
	}

	// 3. Invalid submission test (missing required first name and short message)
	invalidBody := map[string]string{
		"first_name": "",
		"email":      "bad-email",
		"message":    "too short",
	}
	invBytes, _ := json.Marshal(invalidBody)
	reqInv, _ := http.NewRequest(http.MethodPost, "/api/contact", bytes.NewBuffer(invBytes))
	reqInv.Header.Set("Content-Type", "application/json")
	wInv := httptest.NewRecorder()
	r.ServeHTTP(wInv, reqInv)

	if wInv.Code != http.StatusBadRequest {
		t.Fatalf("Expected HTTP 400 for invalid contact submission, got %d", wInv.Code)
	}
}

func TestAdminAuthAndProtectedRoutes(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	// 1. Unauthenticated request to /api/leads must be 401 Unauthorized
	reqUnauth, _ := http.NewRequest(http.MethodGet, "/api/leads", nil)
	wUnauth := httptest.NewRecorder()
	r.ServeHTTP(wUnauth, reqUnauth)

	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized for unauthenticated request, got %d", wUnauth.Code)
	}

	// 2. Valid Login
	cookie, _ := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)

	// 3. Authenticated request to /api/leads using session cookie
	reqAuth, _ := http.NewRequest(http.MethodGet, "/api/leads", nil)
	reqAuth.Header.Set("Cookie", cookie)
	wAuth := httptest.NewRecorder()
	r.ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for authenticated /api/leads request, got %d: %s", wAuth.Code, wAuth.Body.String())
	}
}

func TestCRMLeadOperations(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	// 1. Login as Admin
	cookie, csrfToken := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)

	// 2. Create Lead via POST /api/leads
	newLead := map[string]string{
		"first_name":   "Samantha",
		"last_name":    "Carter",
		"company_name": "Stargate Ops",
		"email":        fmt.Sprintf("scarter-%d@stargate.mil", time.Now().UnixNano()),
		"phone":        "+1-555-0144",
		"service":      "Custom Software Development",
		"message":      "Deep space telemetry parsing software needed.",
		"notes":        "High priority lead",
	}
	nlBytes, _ := json.Marshal(newLead)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(nlBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", cookie)
	reqCreate.Header.Set("X-CSRF-Token", csrfToken)
	wCreate := httptest.NewRecorder()
	r.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for manual lead creation, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createdRes struct {
		Success bool        `json:"success"`
		Data    models.Lead `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createdRes)
	leadID := createdRes.Data.ID.String()

	// 3. Search and Pagination via GET /api/leads?search=Samantha&page=1&page_size=10
	reqSearch, _ := http.NewRequest(http.MethodGet, "/api/leads?search=Samantha&page=1&page_size=10&sort_by=first_name&sort_dir=asc", nil)
	reqSearch.Header.Set("Cookie", cookie)
	wSearch := httptest.NewRecorder()
	r.ServeHTTP(wSearch, reqSearch)

	if wSearch.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for search request, got %d", wSearch.Code)
	}

	var searchRes struct {
		Success    bool          `json:"success"`
		Data       []models.Lead `json:"data"`
		Total      int           `json:"total"`
		TotalPages int           `json:"total_pages"`
	}
	_ = json.Unmarshal(wSearch.Body.Bytes(), &searchRes)
	if searchRes.Total < 1 || len(searchRes.Data) < 1 {
		t.Fatalf("Expected to find created lead by search query")
	}

	// 4. Update Notes Only (Preserving status) via PATCH /api/leads/:id/status
	updateNoteBody := map[string]string{
		"notes": "Follow-up meeting scheduled with Colonel",
	}
	unBytes, _ := json.Marshal(updateNoteBody)
	reqUpdate, _ := http.NewRequest(http.MethodPatch, "/api/leads/"+leadID+"/status", bytes.NewBuffer(unBytes))
	reqUpdate.Header.Set("Content-Type", "application/json")
	reqUpdate.Header.Set("Cookie", cookie)
	reqUpdate.Header.Set("X-CSRF-Token", csrfToken)
	wUpdate := httptest.NewRecorder()
	r.ServeHTTP(wUpdate, reqUpdate)

	if wUpdate.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for updating notes, got %d: %s", wUpdate.Code, wUpdate.Body.String())
	}

	// 5. Verify status preserved and notes updated via GET /api/leads/:id
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/leads/"+leadID, nil)
	reqGet.Header.Set("Cookie", cookie)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)

	var getRes struct {
		Success bool        `json:"success"`
		Data    models.Lead `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &getRes)
	if getRes.Data.Status != models.LeadStatusNew {
		t.Fatalf("Expected status to be preserved as 'New', got '%s'", getRes.Data.Status)
	}
	if getRes.Data.Notes != "Follow-up meeting scheduled with Colonel" {
		t.Fatalf("Expected notes to be updated, got '%s'", getRes.Data.Notes)
	}

	// 6. Delete Lead via DELETE /api/leads/:id
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/leads/"+leadID, nil)
	reqDel.Header.Set("Cookie", cookie)
	reqDel.Header.Set("X-CSRF-Token", csrfToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for lead deletion, got %d", wDel.Code)
	}
}

// TestCRMAuthenticationAndRBAC verifies 401 unauthorized & 403 forbidden roles across all endpoints
func TestCRMAuthenticationAndRBAC(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	dummyID := uuid.New().String()

	// 1. Unauthenticated checks across all CRM endpoints
	unauthEndpoints := []struct {
		method string
		url    string
		expect int
	}{
		{http.MethodGet, "/api/leads", http.StatusUnauthorized},
		{http.MethodGet, "/api/leads/" + dummyID, http.StatusUnauthorized},
		{http.MethodPost, "/api/leads", http.StatusUnauthorized},
		{http.MethodPatch, "/api/leads/" + dummyID + "/status", http.StatusUnauthorized},
		{http.MethodDelete, "/api/leads/" + dummyID, http.StatusUnauthorized},
		{http.MethodGet, "/admin/leads", http.StatusSeeOther}, // HTML redirects to login
	}

	for _, ep := range unauthEndpoints {
		req, _ := http.NewRequest(ep.method, ep.url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != ep.expect {
			t.Errorf("[%s %s] Expected %d for unauthenticated request, got %d", ep.method, ep.url, ep.expect, w.Code)
		}
	}

	// 2. Setup Viewer Role User
	viewerEmail := "viewer-test@spotcrm.com"
	viewerPass := "ViewerPass123!"
	createTestUser(t, db, viewerEmail, viewerPass, models.RoleViewer)
	viewerCookie, viewerCSRF := loginAs(t, r, viewerEmail, viewerPass)

	// Viewer can read leads
	reqViewerGet, _ := http.NewRequest(http.MethodGet, "/api/leads", nil)
	reqViewerGet.Header.Set("Cookie", viewerCookie)
	wViewerGet := httptest.NewRecorder()
	r.ServeHTTP(wViewerGet, reqViewerGet)
	if wViewerGet.Code != http.StatusOK {
		t.Fatalf("Viewer expected 200 OK for GET /api/leads, got %d", wViewerGet.Code)
	}

	// Viewer cannot write (POST /api/leads -> 403)
	postBody, _ := json.Marshal(map[string]string{
		"first_name": "Viewer", "last_name": "Lead", "email": "viewerlead@test.com", "service": "Web Applications",
	})
	reqViewerPost, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(postBody))
	reqViewerPost.Header.Set("Content-Type", "application/json")
	reqViewerPost.Header.Set("Cookie", viewerCookie)
	reqViewerPost.Header.Set("X-CSRF-Token", viewerCSRF)
	wViewerPost := httptest.NewRecorder()
	r.ServeHTTP(wViewerPost, reqViewerPost)
	if wViewerPost.Code != http.StatusForbidden {
		t.Fatalf("Viewer expected 403 Forbidden for POST /api/leads, got %d", wViewerPost.Code)
	}

	// Viewer cannot update status (PATCH /api/leads/:id/status -> 403)
	patchBody, _ := json.Marshal(map[string]string{"status": "Contacted"})
	reqViewerPatch, _ := http.NewRequest(http.MethodPatch, "/api/leads/"+dummyID+"/status", bytes.NewBuffer(patchBody))
	reqViewerPatch.Header.Set("Content-Type", "application/json")
	reqViewerPatch.Header.Set("Cookie", viewerCookie)
	reqViewerPatch.Header.Set("X-CSRF-Token", viewerCSRF)
	wViewerPatch := httptest.NewRecorder()
	r.ServeHTTP(wViewerPatch, reqViewerPatch)
	if wViewerPatch.Code != http.StatusForbidden {
		t.Fatalf("Viewer expected 403 Forbidden for PATCH /api/leads/:id/status, got %d", wViewerPatch.Code)
	}

	// Viewer cannot delete (DELETE /api/leads/:id -> 403)
	reqViewerDel, _ := http.NewRequest(http.MethodDelete, "/api/leads/"+dummyID, nil)
	reqViewerDel.Header.Set("Cookie", viewerCookie)
	reqViewerDel.Header.Set("X-CSRF-Token", viewerCSRF)
	wViewerDel := httptest.NewRecorder()
	r.ServeHTTP(wViewerDel, reqViewerDel)
	if wViewerDel.Code != http.StatusForbidden {
		t.Fatalf("Viewer expected 403 Forbidden for DELETE /api/leads/:id, got %d", wViewerDel.Code)
	}

	// 3. Setup Manager Role User
	managerEmail := "manager-test@spotcrm.com"
	managerPass := "ManagerPass123!"
	createTestUser(t, db, managerEmail, managerPass, models.RoleManager)
	mgrCookie, mgrCSRF := loginAs(t, r, managerEmail, managerPass)

	// Manager CAN create lead (POST /api/leads -> 201)
	mgrLeadBody, _ := json.Marshal(map[string]string{
		"first_name": "ManagerCreated", "last_name": "Lead", "email": fmt.Sprintf("mgr-%d@test.com", time.Now().UnixNano()), "service": "Web Applications",
	})
	reqMgrPost, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(mgrLeadBody))
	reqMgrPost.Header.Set("Content-Type", "application/json")
	reqMgrPost.Header.Set("Cookie", mgrCookie)
	reqMgrPost.Header.Set("X-CSRF-Token", mgrCSRF)
	wMgrPost := httptest.NewRecorder()
	r.ServeHTTP(wMgrPost, reqMgrPost)
	if wMgrPost.Code != http.StatusCreated {
		t.Fatalf("Manager expected 201 Created for POST /api/leads, got %d: %s", wMgrPost.Code, wMgrPost.Body.String())
	}

	var mgrCreated struct {
		Data models.Lead `json:"data"`
	}
	_ = json.Unmarshal(wMgrPost.Body.Bytes(), &mgrCreated)
	createdLeadID := mgrCreated.Data.ID.String()

	// Manager CAN update status (PATCH /api/leads/:id/status -> 200)
	reqMgrPatch, _ := http.NewRequest(http.MethodPatch, "/api/leads/"+createdLeadID+"/status", bytes.NewBuffer(patchBody))
	reqMgrPatch.Header.Set("Content-Type", "application/json")
	reqMgrPatch.Header.Set("Cookie", mgrCookie)
	reqMgrPatch.Header.Set("X-CSRF-Token", mgrCSRF)
	wMgrPatch := httptest.NewRecorder()
	r.ServeHTTP(wMgrPatch, reqMgrPatch)
	if wMgrPatch.Code != http.StatusOK {
		t.Fatalf("Manager expected 200 OK for PATCH /api/leads/:id/status, got %d", wMgrPatch.Code)
	}

	// Manager CANNOT delete lead (DELETE /api/leads/:id -> 403 Forbidden, admin-only)
	reqMgrDel, _ := http.NewRequest(http.MethodDelete, "/api/leads/"+createdLeadID, nil)
	reqMgrDel.Header.Set("Cookie", mgrCookie)
	reqMgrDel.Header.Set("X-CSRF-Token", mgrCSRF)
	wMgrDel := httptest.NewRecorder()
	r.ServeHTTP(wMgrDel, reqMgrDel)
	if wMgrDel.Code != http.StatusForbidden {
		t.Fatalf("Manager expected 403 Forbidden for DELETE /api/leads/:id (admin-only), got %d", wMgrDel.Code)
	}

	// Admin CAN delete lead (DELETE /api/leads/:id -> 200 OK)
	adminCookie, adminCSRF := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)
	reqAdminDel, _ := http.NewRequest(http.MethodDelete, "/api/leads/"+createdLeadID, nil)
	reqAdminDel.Header.Set("Cookie", adminCookie)
	reqAdminDel.Header.Set("X-CSRF-Token", adminCSRF)
	wAdminDel := httptest.NewRecorder()
	r.ServeHTTP(wAdminDel, reqAdminDel)
	if wAdminDel.Code != http.StatusOK {
		t.Fatalf("Admin expected 200 OK for DELETE /api/leads/:id, got %d", wAdminDel.Code)
	}
}

// TestCRMCSRFProtection verifies CSRF protection on authenticated state-changing requests
func TestCRMCSRFProtection(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	cookie, csrfToken := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)

	payload, _ := json.Marshal(map[string]string{
		"first_name": "CSRF", "last_name": "Test", "email": fmt.Sprintf("csrf-%d@example.com", time.Now().UnixNano()), "service": "Custom Software Development",
	})

	// 1. Authenticated POST without CSRF token must be 403 Forbidden
	reqNoToken, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(payload))
	reqNoToken.Header.Set("Content-Type", "application/json")
	reqNoToken.Header.Set("Cookie", cookie)
	wNoToken := httptest.NewRecorder()
	r.ServeHTTP(wNoToken, reqNoToken)

	if wNoToken.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for authenticated POST without CSRF token, got %d", wNoToken.Code)
	}

	// 2. Authenticated POST with invalid CSRF token must be 403 Forbidden
	reqBadToken, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(payload))
	reqBadToken.Header.Set("Content-Type", "application/json")
	reqBadToken.Header.Set("Cookie", cookie)
	reqBadToken.Header.Set("X-CSRF-Token", "invalid-forged-csrf-token-1234567890")
	wBadToken := httptest.NewRecorder()
	r.ServeHTTP(wBadToken, reqBadToken)

	if wBadToken.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for authenticated POST with forged CSRF token, got %d", wBadToken.Code)
	}

	// 3. Authenticated POST with matching CSRF token must succeed (201 Created)
	reqValid, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(payload))
	reqValid.Header.Set("Content-Type", "application/json")
	reqValid.Header.Set("Cookie", cookie)
	reqValid.Header.Set("X-CSRF-Token", csrfToken)
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for authenticated POST with valid CSRF token, got %d: %s", wValid.Code, wValid.Body.String())
	}
}

// TestCRMSearchSortFilterAndPaginationEdgeCases verifies safe parameterized queries and boundary behavior
func TestCRMSearchSortFilterAndPaginationEdgeCases(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	cookie, csrfToken := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)

	// Seed lead with unique keyword
	uniqueKey := fmt.Sprintf("QuantumTelemetry-%d", time.Now().UnixNano())
	seedLead, _ := json.Marshal(map[string]string{
		"first_name":   "Gwen",
		"last_name":    "Stacy",
		"company_name": "Web Labs",
		"email":        fmt.Sprintf("gwen-%d@weblabs.org", time.Now().UnixNano()),
		"service":      "Database Architecture",
		"message":      uniqueKey,
		"notes":        "Initial evaluation",
	})
	reqSeed, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(seedLead))
	reqSeed.Header.Set("Content-Type", "application/json")
	reqSeed.Header.Set("Cookie", cookie)
	reqSeed.Header.Set("X-CSRF-Token", csrfToken)
	wSeed := httptest.NewRecorder()
	r.ServeHTTP(wSeed, reqSeed)
	if wSeed.Code != http.StatusCreated {
		t.Fatalf("Failed seeding lead for edge cases: %d", wSeed.Code)
	}

	// 1. Search with SQL injection payloads must not cause syntax error or leak
	sqliPayloads := []string{
		"'; DROP TABLE leads;--",
		"' OR '1'='1",
		"\" OR \"1\"=\"1",
		"%' UNION SELECT NULL--",
		"%%",
		"__",
	}

	for _, p := range sqliPayloads {
		req, _ := http.NewRequest(http.MethodGet, "/api/leads?search="+strings.ReplaceAll(p, " ", "%20"), nil)
		req.Header.Set("Cookie", cookie)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("SQLi search attempt '%s' caused unexpected status %d", p, w.Code)
		}
	}

	// 2. Search for unique keyword must return exactly matching record
	reqExact, _ := http.NewRequest(http.MethodGet, "/api/leads?search="+uniqueKey, nil)
	reqExact.Header.Set("Cookie", cookie)
	wExact := httptest.NewRecorder()
	r.ServeHTTP(wExact, reqExact)
	if wExact.Code != http.StatusOK {
		t.Fatalf("Exact keyword search failed with status %d", wExact.Code)
	}
	var exactRes struct {
		Data  []models.Lead `json:"data"`
		Total int           `json:"total"`
	}
	_ = json.Unmarshal(wExact.Body.Bytes(), &exactRes)
	if exactRes.Total < 1 || len(exactRes.Data) < 1 {
		t.Fatalf("Expected at least 1 lead matching uniqueKey, got %d", exactRes.Total)
	}

	// 3. Search non-existent string must return total=0
	reqNone, _ := http.NewRequest(http.MethodGet, "/api/leads?search=NonExistentGibberishString9999", nil)
	reqNone.Header.Set("Cookie", cookie)
	wNone := httptest.NewRecorder()
	r.ServeHTTP(wNone, reqNone)
	var noneRes struct {
		Total int `json:"total"`
	}
	_ = json.Unmarshal(wNone.Body.Bytes(), &noneRes)
	if noneRes.Total != 0 {
		t.Fatalf("Expected total 0 for non-existent search, got %d", noneRes.Total)
	}

	// 4. Status filter edge cases
	statuses := []string{"New", "Contacted", "Qualified", "Converted", "Lost", "NonExistentStatus"}
	for _, st := range statuses {
		req, _ := http.NewRequest(http.MethodGet, "/api/leads?status="+st, nil)
		req.Header.Set("Cookie", cookie)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Status query '%s' returned status %d", st, w.Code)
		}
	}

	// 5. Sorting allowlist tests (valid columns and direction fallback)
	sortCases := []string{
		"sort_by=first_name&sort_dir=asc",
		"sort_by=status&sort_dir=desc",
		"sort_by=company_name&sort_dir=asc",
		"sort_by=service&sort_dir=desc",
		"sort_by=created_at&sort_dir=asc",
		"sort_by=password&sort_dir=EVIL_DIR",      // safe fallback to created_at DESC
		"sort_by=';DROP TABLE leads;--&sort_dir=", // injection fallback
	}
	for _, sc := range sortCases {
		req, _ := http.NewRequest(http.MethodGet, "/api/leads?"+sc, nil)
		req.Header.Set("Cookie", cookie)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Sort case '%s' returned status %d", sc, w.Code)
		}
	}

	// 6. Pagination edge cases
	pageCases := []string{
		"page=0&page_size=0",       // defaults to 1, 20
		"page=-5&page_size=-10",    // defaults
		"page=1&page_size=500",     // capped at 100
		"page=999999&page_size=10", // valid page with empty data
	}
	for _, pc := range pageCases {
		req, _ := http.NewRequest(http.MethodGet, "/api/leads?"+pc, nil)
		req.Header.Set("Cookie", cookie)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Pagination case '%s' returned status %d", pc, w.Code)
		}
	}
}

// TestCRMInputValidationAndPersistence verifies field validation and actual DB persistence
func TestCRMInputValidationAndPersistence(t *testing.T) {
	r, db, cfg := setupTestRouter(t)
	defer db.Close()

	cookie, csrfToken := loginAs(t, r, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword)

	// 1. Validation: Missing required fields in POST /api/leads
	badInputs := []map[string]string{
		{"first_name": "", "last_name": "Doe", "email": "j@doe.com", "service": "Web Applications"},
		{"first_name": "John", "last_name": "", "email": "j@doe.com", "service": "Web Applications"},
		{"first_name": "John", "last_name": "Doe", "email": "invalid-email-format", "service": "Web Applications"},
		{"first_name": "John", "last_name": "Doe", "email": "j@doe.com", "service": ""},
	}

	for i, inp := range badInputs {
		b, _ := json.Marshal(inp)
		req, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		req.Header.Set("X-CSRF-Token", csrfToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Input case %d expected 400 Bad Request, got %d", i, w.Code)
		}
	}

	// 2. Validation: PATCH /api/leads/:id/status errors
	// A. Invalid UUID
	reqBadUUID, _ := http.NewRequest(http.MethodPatch, "/api/leads/not-a-uuid/status", bytes.NewBuffer([]byte(`{"status":"Contacted"}`)))
	reqBadUUID.Header.Set("Content-Type", "application/json")
	reqBadUUID.Header.Set("Cookie", cookie)
	reqBadUUID.Header.Set("X-CSRF-Token", csrfToken)
	wBadUUID := httptest.NewRecorder()
	r.ServeHTTP(wBadUUID, reqBadUUID)
	if wBadUUID.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for non-UUID lead ID, got %d", wBadUUID.Code)
	}

	// B. Non-existent UUID
	reqNonExistent, _ := http.NewRequest(http.MethodPatch, "/api/leads/00000000-0000-0000-0000-000000000000/status", bytes.NewBuffer([]byte(`{"status":"Contacted"}`)))
	reqNonExistent.Header.Set("Content-Type", "application/json")
	reqNonExistent.Header.Set("Cookie", cookie)
	reqNonExistent.Header.Set("X-CSRF-Token", csrfToken)
	wNonExistent := httptest.NewRecorder()
	r.ServeHTTP(wNonExistent, reqNonExistent)
	if wNonExistent.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for non-existent lead ID, got %d", wNonExistent.Code)
	}

	// C. Neither status nor notes provided
	reqEmptyPatch, _ := http.NewRequest(http.MethodPatch, "/api/leads/00000000-0000-0000-0000-000000000000/status", bytes.NewBuffer([]byte(`{}`)))
	reqEmptyPatch.Header.Set("Content-Type", "application/json")
	reqEmptyPatch.Header.Set("Cookie", cookie)
	reqEmptyPatch.Header.Set("X-CSRF-Token", csrfToken)
	wEmptyPatch := httptest.NewRecorder()
	r.ServeHTTP(wEmptyPatch, reqEmptyPatch)
	if wEmptyPatch.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 when neither status nor notes are provided, got %d", wEmptyPatch.Code)
	}

	// 3. Database Persistence Verification
	persistEmail := fmt.Sprintf("persist-%d@example.com", time.Now().UnixNano())
	validLead := map[string]string{
		"first_name":   "Persistence",
		"last_name":    "Check",
		"company_name": "ACME Persistent",
		"email":        persistEmail,
		"phone":        "+1-555-4321",
		"service":      "Database Architecture",
		"message":      "Verify raw SQL database persistence.",
		"notes":        "Initial note for persistence",
	}
	vlBytes, _ := json.Marshal(validLead)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/leads", bytes.NewBuffer(vlBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", cookie)
	reqCreate.Header.Set("X-CSRF-Token", csrfToken)
	wCreate := httptest.NewRecorder()
	r.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("Failed creating lead for persistence test: %d %s", wCreate.Code, wCreate.Body.String())
	}

	var created struct {
		Data models.Lead `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &created)
	targetID := created.Data.ID

	// Directly query PostgreSQL to verify insertion
	var dbFirstName, dbEmail, dbNotes, dbStatus string
	err := db.QueryRowContext(context.Background(), `
		SELECT first_name, email, COALESCE(notes, ''), status
		FROM leads
		WHERE id = $1
	`, targetID).Scan(&dbFirstName, &dbEmail, &dbNotes, &dbStatus)

	if err != nil {
		t.Fatalf("Failed directly querying database for created lead: %v", err)
	}
	if dbFirstName != "Persistence" || dbEmail != persistEmail || dbNotes != "Initial note for persistence" || dbStatus != "New" {
		t.Fatalf("Database row mismatch: got (%s, %s, %s, %s)", dbFirstName, dbEmail, dbNotes, dbStatus)
	}

	// Direct persistence check for updating notes only
	updatedNoteText := "Updated note after initial consultation"
	patchNotes, _ := json.Marshal(map[string]string{
		"notes": updatedNoteText,
	})
	reqPatch, _ := http.NewRequest(http.MethodPatch, "/api/leads/"+targetID.String()+"/status", bytes.NewBuffer(patchNotes))
	reqPatch.Header.Set("Content-Type", "application/json")
	reqPatch.Header.Set("Cookie", cookie)
	reqPatch.Header.Set("X-CSRF-Token", csrfToken)
	wPatch := httptest.NewRecorder()
	r.ServeHTTP(wPatch, reqPatch)
	if wPatch.Code != http.StatusOK {
		t.Fatalf("Failed updating notes: %d", wPatch.Code)
	}

	// Directly query PostgreSQL again to verify notes persisted and status preserved
	err = db.QueryRowContext(context.Background(), `
		SELECT COALESCE(notes, ''), status
		FROM leads
		WHERE id = $1
	`, targetID).Scan(&dbNotes, &dbStatus)
	if err != nil {
		t.Fatalf("Failed directly querying database after note update: %v", err)
	}
	if dbNotes != updatedNoteText {
		t.Fatalf("Expected notes in DB to be '%s', got '%s'", updatedNoteText, dbNotes)
	}
	if dbStatus != "New" {
		t.Fatalf("Expected status to be preserved as 'New' in DB, got '%s'", dbStatus)
	}

	// Delete and verify database deletion
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/leads/"+targetID.String(), nil)
	reqDel.Header.Set("Cookie", cookie)
	reqDel.Header.Set("X-CSRF-Token", csrfToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("Failed deleting lead: %d", wDel.Code)
	}

	var dummy string
	err = db.QueryRowContext(context.Background(), `SELECT id FROM leads WHERE id = $1`, targetID).Scan(&dummy)
	if err != sql.ErrNoRows {
		t.Fatalf("Expected sql.ErrNoRows after lead deletion, got %v", err)
	}
}

// TestContactDeduplicationAndOutboxPersistence verifies transactional outbox and lead creation from contact form
func TestContactDeduplicationAndOutboxPersistence(t *testing.T) {
	r, db, _ := setupTestRouter(t)
	defer db.Close()

	testEmail := fmt.Sprintf("inquiry-%d@example.com", time.Now().UnixNano())
	testMsg := "We require custom enterprise software development and database design."

	payload := map[string]string{
		"first_name":   "Arthur",
		"last_name":    "Dent",
		"company_name": "Earth Corp",
		"email":        testEmail,
		"phone":        "+1-555-4242",
		"service":      "Custom Software Development",
		"message":      testMsg,
	}
	b, _ := json.Marshal(payload)

	// 1. Submit contact inquiry
	req, _ := http.NewRequest(http.MethodPost, "/api/contact", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for contact submission, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Query contact_submissions directly
	var submissionID uuid.UUID
	err := db.QueryRowContext(context.Background(), `
		SELECT id FROM contact_submissions WHERE email = $1 ORDER BY created_at DESC LIMIT 1
	`, testEmail).Scan(&submissionID)
	if err != nil {
		t.Fatalf("Failed to find contact submission in DB: %v", err)
	}

	// 3. Query leads directly
	var leadID uuid.UUID
	var leadSource string
	err = db.QueryRowContext(context.Background(), `
		SELECT id, source FROM leads WHERE submission_id = $1
	`, submissionID).Scan(&leadID, &leadSource)
	if err != nil {
		t.Fatalf("Failed to find associated CRM lead in DB: %v", err)
	}
	if leadSource != "Website" {
		t.Fatalf("Expected lead source 'Website', got '%s'", leadSource)
	}

	// 4. Query email_outbox directly
	var outboxCount int
	err = db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM email_outbox WHERE submission_id = $1
	`, submissionID).Scan(&outboxCount)
	if err != nil {
		t.Fatalf("Failed querying email_outbox: %v", err)
	}
	if outboxCount != 2 {
		t.Fatalf("Expected 2 queued emails in outbox (admin + customer), got %d", outboxCount)
	}

	// 5. Immediate duplicate submission must be rejected
	reqDup, _ := http.NewRequest(http.MethodPost, "/api/contact", bytes.NewBuffer(b))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	r.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for duplicate submission, got %d", wDup.Code)
	}
}

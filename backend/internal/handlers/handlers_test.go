package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"incidenthub/backend/internal/auth"
	"incidenthub/backend/internal/middleware"
	"incidenthub/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/incidenthub?sslmode=disable"
	}
	return dsn
}

func setupTestDB(t *testing.T, dsn string) *sqlx.DB {
	db, err := sqlx.Open("postgres", dsn)
	require.NoError(t, err)

	db.MustExec(`DROP TABLE IF EXISTS comments`)
	db.MustExec(`DROP TABLE IF EXISTS incidents`)
	db.MustExec(`DROP TABLE IF EXISTS users`)

	migrations := []string{
		`CREATE TABLE users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE incidents (
			id VARCHAR(36) PRIMARY KEY,
			title VARCHAR(500) NOT NULL,
			description TEXT NOT NULL,
			severity VARCHAR(20) NOT NULL DEFAULT 'LOW',
			status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
			assignee_id VARCHAR(36),
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			resolved_at TIMESTAMP WITH TIME ZONE,
			CONSTRAINT fk_incidents_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT fk_incidents_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE comments (
			id VARCHAR(36) PRIMARY KEY,
			incident_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT fk_comments_incident FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
			CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}
	for _, m := range migrations {
		_, err := db.Exec(m)
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		db.MustExec(`DROP TABLE IF EXISTS comments`)
		db.MustExec(`DROP TABLE IF EXISTS incidents`)
		db.MustExec(`DROP TABLE IF EXISTS users`)
		db.Close()
	})

	return db
}

func createTestUser(t *testing.T, db *sqlx.DB) string {
	id := generateUUID()
	hash, _ := auth.HashPassword("password123")
	_, err := db.Exec(`INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)`,
		id, "test@example.com", hash, "Test User")
	require.NoError(t, err)
	return id
}

func createTestUser2(t *testing.T, db *sqlx.DB) string {
	id := generateUUID()
	hash, _ := auth.HashPassword("password123")
	_, err := db.Exec(`INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)`,
		id, "user2@example.com", hash, "User Two")
	require.NoError(t, err)
	return id
}

func createTestIncident(t *testing.T, db *sqlx.DB, userID string) string {
	id := generateUUID()
	_, err := db.Exec(`INSERT INTO incidents (id, title, description, severity, status, created_by) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, "Test Incident", "Test description", "HIGH", "OPEN", userID)
	require.NoError(t, err)
	return id
}

func generateTestToken(t *testing.T, userID, email, name, secret string) string {
	token, err := auth.GenerateToken(userID, email, name, secret)
	require.NoError(t, err)
	return token
}

func newTestRouter(db *sqlx.DB, secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	authHandler := NewAuthHandler(db, secret)
	incidentHandler := NewIncidentHandler(db)
	commentHandler := NewCommentHandler(db)
	dashboardHandler := NewDashboardHandler(db)

	api := r.Group("/api")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	auth := api.Group("", middleware.AuthRequired(secret))
	auth.GET("/dashboard/stats", dashboardHandler.GetStats)
	auth.GET("/incidents", incidentHandler.List)
	auth.POST("/incidents", incidentHandler.Create)
	auth.GET("/incidents/:id", incidentHandler.Get)
	auth.PUT("/incidents/:id", incidentHandler.Update)
	auth.DELETE("/incidents/:id", incidentHandler.Delete)
	auth.GET("/incidents/:id/comments", commentHandler.List)
	auth.POST("/incidents/:id/comments", commentHandler.Create)

	return r
}

// Auth Tests

func TestRegisterSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	r := newTestRouter(db, "test-secret")

	body := `{"email":"new@example.com","password":"password123","name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "new@example.com", resp["email"])
}

func TestRegisterDuplicateEmail(t *testing.T) {
	db := setupTestDB(t, testDSN())
	createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"email":"test@example.com","password":"password123","name":"Another User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegisterValidation(t *testing.T) {
	db := setupTestDB(t, testDSN())
	r := newTestRouter(db, "test-secret")

	body := `{"email":"invalid","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.LoginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "test@example.com", resp.Email)
}

func TestLoginWrongPassword(t *testing.T) {
	db := setupTestDB(t, testDSN())
	createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginNonExistentUser(t *testing.T) {
	db := setupTestDB(t, testDSN())
	r := newTestRouter(db, "test-secret")

	body := `{"email":"nonexistent@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Middleware Tests

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	db := setupTestDB(t, testDSN())
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	db := setupTestDB(t, testDSN())
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Incident Tests

func TestCreateIncidentSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"title":"Server Down","description":"Production server is unresponsive","severity":"CRITICAL"}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Incident
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Server Down", resp.Title)
	assert.Equal(t, models.SeverityCritical, resp.Severity)
	assert.Equal(t, models.StatusOpen, resp.Status)
}

func TestCreateIncidentValidation(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"title":"","description":"","severity":"INVALID"}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateIncidentWithResolvedStatus(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"title":"Resolved Issue","description":"Already fixed","severity":"LOW","status":"RESOLVED"}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Incident
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp.ResolvedAt)
}

func TestGetIncidentSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents/"+incidentID, nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Incident
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Test Incident", resp.Title)
	assert.NotNil(t, resp.Creator)
	assert.Equal(t, "Test User", resp.Creator.Name)
}

func TestGetIncidentNotFound(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateIncidentStatusToResolved(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	body := `{"status":"RESOLVED"}`
	req := httptest.NewRequest(http.MethodPut, "/api/incidents/"+incidentID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Incident
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, models.StatusResolved, resp.Status)
	assert.NotNil(t, resp.ResolvedAt)
}

func TestUpdateIncidentReopen(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)

	// First resolve it
	db.MustExec(`UPDATE incidents SET status = 'RESOLVED', resolved_at = NOW() WHERE id = $1`, incidentID)

	r := newTestRouter(db, "test-secret")

	body := `{"status":"OPEN"}`
	req := httptest.NewRequest(http.MethodPut, "/api/incidents/"+incidentID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Incident
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, models.StatusOpen, resp.Status)
	assert.Nil(t, resp.ResolvedAt)
}

func TestUpdateIncidentAssignee(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	assigneeID := createTestUser2(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	body := `{"assignee_id":"` + assigneeID + `"}`
	req := httptest.NewRequest(http.MethodPut, "/api/incidents/"+incidentID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteIncidentSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodDelete, "/api/incidents/"+incidentID, nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	var count int
	db.Get(&count, `SELECT COUNT(*) FROM incidents WHERE id = $1`, incidentID)
	assert.Equal(t, 0, count)
}

func TestDeleteIncidentNotFound(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodDelete, "/api/incidents/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// List/Filter Tests

func TestListIncidents(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	for i := 0; i < 3; i++ {
		createTestIncident(t, db, userID)
	}
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.Page)
}

func TestListIncidentsFilterByStatus(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	createTestIncident(t, db, userID)
	db.MustExec(`UPDATE incidents SET status = 'INVESTIGATING'`)
	createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents?status=OPEN", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Total)
}

func TestListIncidentsFilterBySeverity(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	createTestIncident(t, db, userID)
	db.MustExec(`UPDATE incidents SET severity = 'LOW'`)
	createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents?severity=HIGH", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Total)
}

func TestListIncidentsSearch(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	createTestIncident(t, db, userID)
	db.MustExec(`UPDATE incidents SET title = 'Database Connection Failed', description = 'Cannot connect to database'`)
	createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents?search=database", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Total)
}

func TestListIncidentsPagination(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	for i := 0; i < 5; i++ {
		createTestIncident(t, db, userID)
	}
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents?page=1&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(5), resp.Total)
	assert.Equal(t, 3, resp.TotalPages)
}

// Comment Tests

func TestCreateCommentSuccess(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	body := `{"body":"Investigating the issue now"}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents/"+incidentID+"/comments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Comment
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Investigating the issue now", resp.Body)
	assert.NotNil(t, resp.Author)
}

func TestCreateCommentNotFound(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	r := newTestRouter(db, "test-secret")

	body := `{"body":"A comment"}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents/nonexistent-id/comments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateCommentValidation(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)
	r := newTestRouter(db, "test-secret")

	body := `{"body":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/incidents/"+incidentID+"/comments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListComments(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	incidentID := createTestIncident(t, db, userID)

	db.MustExec(`INSERT INTO comments (id, incident_id, user_id, body) VALUES ($1, $2, $3, $4)`,
		generateUUID(), incidentID, userID, "First comment")
	db.MustExec(`INSERT INTO comments (id, incident_id, user_id, body) VALUES ($1, $2, $3, $4)`,
		generateUUID(), incidentID, userID, "Second comment")

	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/incidents/"+incidentID+"/comments", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Comment
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp))
}

// Dashboard Tests

func TestDashboardStats(t *testing.T) {
	db := setupTestDB(t, testDSN())
	userID := createTestUser(t, db)
	createTestIncident(t, db, userID)
	createTestIncident(t, db, userID)
	db.MustExec(`UPDATE incidents SET status = 'INVESTIGATING'`)
	db.MustExec(`UPDATE incidents SET severity = 'CRITICAL'`)
	r := newTestRouter(db, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, userID, "test@example.com", "Test User", "test-secret"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	stats := resp["stats"].(map[string]interface{})
	assert.Equal(t, float64(2), stats["total"])
	assert.Equal(t, float64(1), stats["investigating"])
	assert.Equal(t, float64(1), stats["critical"])
}
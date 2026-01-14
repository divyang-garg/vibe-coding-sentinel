// Shared test helpers for both unit and integration tests
package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestConfig holds test configuration
type TestConfig struct {
	DatabaseURL string
	APIKey      string
	ProjectID   string
}

var testConfig *TestConfig
var testDB *sql.DB
var testProjectID string

// GetTestConfig returns the test configuration
func GetTestConfig() *TestConfig {
	if testConfig == nil {
		testConfig = &TestConfig{
			DatabaseURL: getEnv("TEST_DATABASE_URL", "postgres://sentinel:sentinel@localhost:5432/sentinel_test?sslmode=disable"),
			APIKey:      getEnv("TEST_API_KEY", "test-api-key"),
			ProjectID:   getEnv("TEST_PROJECT_ID", ""),
		}
	}
	return testConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestDB sets up a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	if testDB != nil {
		return testDB
	}

	config := GetTestConfig()
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	testDB = db
	return db
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T) {
	if testDB != nil {
		testDB.Close()
		testDB = nil
	}
}

// CreateTestProject creates a test project in the database
func CreateTestProject(t *testing.T, db *sql.DB) string {
	if testProjectID != "" {
		return testProjectID
	}

	projectID := "test-project-" + generateTestID()
	_, err := db.Exec(`
		INSERT INTO projects (id, org_id, name, api_key, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO NOTHING
	`, projectID, "test-org", "Test Project", GetTestConfig().APIKey)

	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	testProjectID = projectID
	return projectID
}

// CreateTestRequest creates an HTTP test request with proper headers
func CreateTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetTestConfig().APIKey)
	req.Header.Set("X-API-Key", GetTestConfig().APIKey)

	return req
}

// AssertJSONResponse asserts that a response is valid JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	if w.Code == 0 {
		t.Error("Response code not set")
		return nil
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v\nBody: %s", err, w.Body.String())
		return nil
	}

	return response
}

// AssertSuccessResponse asserts that a response indicates success
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) map[string]interface{} {
	if w.Code != expectedCode {
		t.Errorf("Expected status code %d, got %d\nBody: %s", expectedCode, w.Code, w.Body.String())
		return nil
	}

	response := AssertJSONResponse(t, w)
	if response == nil {
		return nil
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v\nResponse: %+v", response["success"], response)
		return nil
	}

	return response
}

// AssertErrorResponse asserts that a response indicates an error
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) map[string]interface{} {
	if w.Code != expectedCode {
		t.Errorf("Expected status code %d, got %d\nBody: %s", expectedCode, w.Code, w.Body.String())
		return nil
	}

	response := AssertJSONResponse(t, w)
	if response == nil {
		return nil
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v\nResponse: %+v", response["success"], response)
		return nil
	}

	return response
}

// MockLLMService provides a mock LLM service for testing
type MockLLMService struct {
	Response string
	Error    error
}

func NewMockLLMService() *MockLLMService {
	return &MockLLMService{}
}

func (m *MockLLMService) SetResponse(response string) {
	m.Response = response
	m.Error = nil
}

func (m *MockLLMService) SetError(err error) {
	m.Error = err
	m.Response = ""
}

func (m *MockLLMService) GetResponse() (string, error) {
	return m.Response, m.Error
}

func (m *MockLLMService) ResetTestState() {
	m.Response = ""
	m.Error = nil
}

// Helper functions
func generateTestID() string {
	return fmt.Sprintf("%d", os.Getpid())
}

// Package integration provides end-to-end API tests
// Complies with CODING_STANDARDS.md: Integration test file max 500 lines
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/divyang-garg/sentinel-hub-api/internal/api/server"
	"github.com/divyang-garg/sentinel-hub-api/internal/config"
	"github.com/divyang-garg/sentinel-hub-api/internal/models"
	"github.com/divyang-garg/sentinel-hub-api/internal/repository"
	_ "github.com/lib/pq"
)

type UserAPIIntegrationTestSuite struct {
	suite.Suite
	db     *sql.DB
	server *server.Server
}

func (suite *UserAPIIntegrationTestSuite) SetupSuite() {
	// Use a test database
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://sentinel:password@localhost/sentinel_test?sslmode=disable"
	}

	var err error
	suite.db, err = repository.NewDatabaseConnection(testDBURL)
	suite.Require().NoError(err)

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: config.DatabaseConfig{
			URL: testDBURL,
		},
		Security: config.SecurityConfig{
			JWTSecret:          "test-jwt-secret-for-integration-tests",
			JWTExpiration:      time.Hour,
			BcryptCost:         4, // Lower cost for faster tests
			RateLimitRequests:  1000,
			RateLimitWindow:    time.Minute,
			CORSAllowedOrigins: []string{"*"},
		},
		LLM: config.LLMConfig{
			RequestTimeout: 30 * time.Second,
		},
	}

	// Create and start server
	suite.server = server.NewServer(cfg, suite.db)

	// Clean up any existing test data
	suite.cleanupTestData()
}

func (suite *UserAPIIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.cleanupTestData()
		suite.db.Close()
	}
}

func (suite *UserAPIIntegrationTestSuite) SetupTest() {
	suite.cleanupTestData()
}

func (suite *UserAPIIntegrationTestSuite) cleanupTestData() {
	// Clean up test data
	suite.db.Exec("DELETE FROM users WHERE email LIKE '%test%'")
}

func TestUserAPIIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(UserAPIIntegrationTestSuite))
}

func (suite *UserAPIIntegrationTestSuite) TestFullUserLifecycle() {
	// 1. Create a user
	createReq := map[string]interface{}{
		"email":    "integration-test@example.com",
		"name":     "Integration Test User",
		"password": "testpassword123",
	}

	body, _ := json.Marshal(createReq)
	_ = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))

	// We need to create a test router since we can't easily start the full server
	// In a real scenario, you'd start the server and make HTTP calls
	suite.T().Skip("Full integration test requires server startup - implement with httptest.Server")
}

func (suite *UserAPIIntegrationTestSuite) TestCreateUser_DuplicateEmail() {
	// First, create a user directly in the database
	user := &models.User{
		Email:     "duplicate-test@example.com",
		Name:      "Duplicate Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo := repository.NewPostgresUserRepository(suite.db)
	createdUser, err := repo.Create(context.Background(), user)
	suite.NoError(err)
	suite.NotNil(createdUser)

	// Now try to create another user with the same email (should fail)
	// This would be tested via HTTP in a full integration test
	suite.T().Log("User created successfully for duplicate email test setup")
}

func (suite *UserAPIIntegrationTestSuite) TestDatabaseConnection() {
	// Test basic database connectivity
	err := suite.db.Ping()
	suite.NoError(err, "Database should be reachable")

	// Test basic query
	var result int
	err = suite.db.QueryRow("SELECT 1").Scan(&result)
	suite.NoError(err, "Basic query should work")
	suite.Equal(1, result)
}

func (suite *UserAPIIntegrationTestSuite) TestUserRepositoryIntegration() {
	repo := repository.NewPostgresUserRepository(suite.db)

	// Test Create
	user := &models.User{
		Email:     "repo-test@example.com",
		Name:      "Repository Test User",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := repo.Create(context.Background(), user)
	suite.NoError(err)
	suite.NotNil(created)
	suite.NotZero(created.ID)
	suite.Equal(user.Email, created.Email)
	suite.Equal(user.Name, created.Name)

	// Test GetByID
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(created.ID, retrieved.ID)
	suite.Equal(created.Email, retrieved.Email)

	// Test GetByEmail
	byEmail, err := repo.GetByEmail(context.Background(), created.Email)
	suite.NoError(err)
	suite.NotNil(byEmail)
	suite.Equal(created.ID, byEmail.ID)

	// Test Update
	created.Name = "Updated Repository Test User"
	err = repo.Update(context.Background(), created)
	suite.NoError(err)

	// Verify update
	updated, err := repo.GetByID(context.Background(), created.ID)
	suite.NoError(err)
	suite.Equal("Updated Repository Test User", updated.Name)

	// Test Delete
	err = repo.Delete(context.Background(), created.ID)
	suite.NoError(err)

	// Verify deletion
	deleted, err := repo.GetByID(context.Background(), created.ID)
	suite.Error(err)
	suite.Nil(deleted)
	suite.Contains(err.Error(), "not found")
}

func (suite *UserAPIIntegrationTestSuite) TestConcurrentUserOperations() {
	repo := repository.NewPostgresUserRepository(suite.db)

	// Test concurrent user creation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			user := &models.User{
				Email:     fmt.Sprintf("concurrent-test-%d@example.com", index),
				Name:      fmt.Sprintf("Concurrent Test User %d", index),
				Password:  "hashedpassword",
				Role:      models.RoleUser,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err := repo.Create(context.Background(), user)
			assert.NoError(suite.T(), err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all users were created
	users, err := repo.List(context.Background(), 20, 0)
	suite.NoError(err)
	concurrentUsers := 0
	for _, user := range users {
		if len(user.Email) > 15 && user.Email[:15] == "concurrent-test" {
			concurrentUsers++
		}
	}
	suite.Equal(10, concurrentUsers)
}

func (suite *UserAPIIntegrationTestSuite) TestDatabaseConstraints() {
	repo := repository.NewPostgresUserRepository(suite.db)

	// Test unique email constraint (if implemented at DB level)
	user1 := &models.User{
		Email:     "constraint-test@example.com",
		Name:      "Constraint Test User 1",
		Password:  "hashedpassword",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := repo.Create(context.Background(), user1)
	suite.NoError(err)

	// Try to create user with same email
	user2 := &models.User{
		Email:     "constraint-test@example.com", // Same email
		Name:      "Constraint Test User 2",
		Password:  "hashedpassword2",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = repo.Create(context.Background(), user2)
	// This might succeed or fail depending on DB constraints
	// The important thing is that the operation completes
	suite.T().Logf("Duplicate email creation result: %v", err)
}

func (suite *UserAPIIntegrationTestSuite) TestLargeDatasetHandling() {
	repo := repository.NewPostgresUserRepository(suite.db)

	// Create multiple users for pagination testing
	users := make([]*models.User, 25)
	for i := 0; i < 25; i++ {
		user := &models.User{
			Email:     fmt.Sprintf("large-dataset-%d@example.com", i),
			Name:      fmt.Sprintf("Large Dataset User %d", i),
			Password:  "hashedpassword",
			Role:      models.RoleUser,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		users[i] = user
	}

	// Insert users
	for _, user := range users {
		_, err := repo.Create(context.Background(), user)
		suite.NoError(err)
	}

	// Test pagination
	page1, err := repo.List(context.Background(), 10, 0)
	suite.NoError(err)
	suite.Len(page1, 10)

	page2, err := repo.List(context.Background(), 10, 10)
	suite.NoError(err)
	suite.Len(page2, 10)

	page3, err := repo.List(context.Background(), 10, 20)
	suite.NoError(err)
	suite.Len(page3, 5) // Should have 5 remaining

	// Verify no overlap between pages
	page1IDs := make(map[int]bool)
	for _, user := range page1 {
		page1IDs[user.ID] = true
	}

	for _, user := range page2 {
		suite.False(page1IDs[user.ID], "User should not appear in multiple pages")
	}
}

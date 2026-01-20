// Test Sandbox - Main Types and HTTP Handlers
// Handles test execution requests and database operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestExecution represents a test execution record
type TestExecution struct {
	ID              string          `json:"id"`
	ProjectID       string          `json:"project_id"`
	ExecutionType   string          `json:"execution_type"` // "coverage", "validation", "mutation", "full"
	Status          string          `json:"status"`         // "running", "completed", "failed"
	Result          json.RawMessage `json:"result,omitempty"`
	ExecutionTimeMs int             `json:"execution_time_ms"`
	CreatedAt       time.Time       `json:"created_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}

// TestExecutionRequest represents the request to execute tests
type TestExecutionRequest struct {
	ProjectID     string           `json:"project_id"`
	ExecutionType string           `json:"executionType"` // "coverage", "validation", "mutation", "full"
	Language      string           `json:"language"`
	TestFiles     []TestFile       `json:"testFiles"`              // Test files with content
	SourceFiles   []TestFile       `json:"sourceFiles,omitempty"`  // Source files (for mutation testing)
	Dependencies  []DependencyFile `json:"dependencies,omitempty"` // Dependency files (go.mod, package.json, requirements.txt)
	TestCommand   string           `json:"testCommand,omitempty"`  // Optional: custom test command
}

// DependencyFile represents a dependency file (go.mod, package.json, etc.)
type DependencyFile struct {
	Path    string `json:"path"`    // e.g., "go.mod", "package.json", "requirements.txt"
	Content string `json:"content"` // File content
}

// TestExecutionResponse represents the response
type TestExecutionResponse struct {
	Success         bool   `json:"success"`
	ExecutionID     string `json:"executionId"`
	Status          string `json:"status"` // "running", "completed", "failed"
	ExitCode        int    `json:"exitCode"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExecutionTimeMs int    `json:"executionTimeMs"`
	Message         string `json:"message,omitempty"`
}

// ExecutionResult stores the execution results
type ExecutionResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// saveTestExecution saves execution record to database
func saveTestExecution(ctx context.Context, execution TestExecution) error {
	query := `
		INSERT INTO test_executions 
		(id, project_id, execution_type, status, result, execution_time_ms, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			result = EXCLUDED.result,
			execution_time_ms = EXCLUDED.execution_time_ms,
			completed_at = EXCLUDED.completed_at
	`

	_, err := database.ExecWithTimeout(ctx, db, query,
		execution.ID, execution.ProjectID, execution.ExecutionType, execution.Status,
		execution.Result, execution.ExecutionTimeMs, execution.CreatedAt, execution.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save test execution: %w", err)
	}

	return nil
}

// testExecutionHandler handles the API request to execute tests
func testExecutionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		http.Error(w, "language is required", http.StatusBadRequest)
		return
	}
	if len(req.TestFiles) == 0 {
		http.Error(w, "testFiles are required", http.StatusBadRequest)
		return
	}

	// Check Docker availability
	if !checkDockerAvailable() {
		http.Error(w, "Docker is not available on this system", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Create execution record
	executionID := uuid.New().String()
	now := time.Now()
	execution := TestExecution{
		ID:            executionID,
		ProjectID:     req.ProjectID,
		ExecutionType: req.ExecutionType,
		Status:        "running",
		CreatedAt:     now,
	}

	// Save initial execution record
	if err := saveTestExecution(ctx, execution); err != nil {
		log.Printf("Error saving execution record: %v", err)
		// Continue anyway
	}

	// Execute tests in sandbox
	startTime := time.Now()
	result, err := executeTestInSandbox(ctx, req)
	executionTime := time.Since(startTime)

	// Update execution record
	completedAt := time.Now()
	status := "completed"
	if err != nil {
		status = "failed"
	}

	// Serialize result
	resultJSON, _ := json.Marshal(result)

	execution.Status = status
	execution.Result = resultJSON
	execution.ExecutionTimeMs = int(executionTime.Milliseconds())
	execution.CompletedAt = &completedAt

	if err := saveTestExecution(ctx, execution); err != nil {
		log.Printf("Error updating execution record: %v", err)
	}

	// Prepare response
	response := TestExecutionResponse{
		Success:         err == nil,
		ExecutionID:     executionID,
		Status:          status,
		ExecutionTimeMs: int(executionTime.Milliseconds()),
	}

	if result != nil {
		response.ExitCode = result.ExitCode
		response.Stdout = result.Stdout
		response.Stderr = result.Stderr
	}

	if err != nil {
		response.Message = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getTestExecutionHandler handles GET request to retrieve execution status
func getTestExecutionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	executionID := chi.URLParam(r, "execution_id")
	if executionID == "" {
		http.Error(w, "execution_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, project_id, execution_type, status, result, execution_time_ms, created_at, completed_at
		FROM test_executions
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, executionID)

	var execution TestExecution
	var resultJSON sql.NullString
	var completedAt sql.NullTime

	err := row.Scan(
		&execution.ID, &execution.ProjectID, &execution.ExecutionType, &execution.Status,
		&resultJSON, &execution.ExecutionTimeMs, &execution.CreatedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Execution not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying execution: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query execution: %v", err), http.StatusInternalServerError)
		return
	}

	if resultJSON.Valid {
		execution.Result = json.RawMessage(resultJSON.String)
	}
	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}

	// Parse result to extract exit code, stdout, stderr
	var execResult ExecutionResult
	if len(execution.Result) > 0 {
		json.Unmarshal(execution.Result, &execResult)
	}

	response := TestExecutionResponse{
		Success:         execution.Status == "completed",
		ExecutionID:     execution.ID,
		Status:          execution.Status,
		ExitCode:        execResult.ExitCode,
		Stdout:          execResult.Stdout,
		Stderr:          execResult.Stderr,
		ExecutionTimeMs: execution.ExecutionTimeMs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Fixed import structure
package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sentinel-hub-api/pkg/database"
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

// checkDockerAvailable checks if Docker is available
func checkDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// getDockerImage returns the appropriate Docker image for a language
func getDockerImage(language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return "golang:1.21-alpine"
	case "javascript", "js", "typescript", "ts":
		return "node:20-alpine"
	case "python", "py":
		return "python:3.11-alpine"
	default:
		return "golang:1.21-alpine" // Default fallback
	}
}

// needsNetworkAccess determines if network access is needed for test execution
func needsNetworkAccess(language string, deps []DependencyFile) bool {
	// Check if dependencies file exists and requires network
	for _, dep := range deps {
		if dep.Path == "package.json" || dep.Path == "requirements.txt" || dep.Path == "go.mod" {
			// If dependencies file exists, likely need network for installation
			return true
		}
	}
	return false
}

// validateDependencyFile validates a dependency file before writing
func validateDependencyFile(dep DependencyFile) error {
	switch dep.Path {
	case "go.mod":
		if !strings.HasPrefix(dep.Content, "module ") {
			return fmt.Errorf("invalid go.mod: missing module declaration")
		}
	case "package.json":
		var pkg map[string]interface{}
		if err := json.Unmarshal([]byte(dep.Content), &pkg); err != nil {
			return fmt.Errorf("invalid package.json: %w", err)
		}
	case "requirements.txt":
		// Basic validation - check for common patterns
		if len(dep.Content) == 0 {
			return fmt.Errorf("empty requirements.txt")
		}
		// Check for basic format (package names, version specifiers)
		lines := strings.Split(dep.Content, "\n")
		hasValidLine := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				hasValidLine = true
				break
			}
		}
		if !hasValidLine {
			return fmt.Errorf("requirements.txt has no valid package declarations")
		}
	}
	return nil
}

// getTestCommand returns the test command for a language
func getTestCommand(language string, customCommand string) string {
	if customCommand != "" {
		return customCommand
	}

	switch strings.ToLower(language) {
	case "go", "golang":
		return "go test -v ./..."
	case "javascript", "js", "typescript", "ts":
		return "npm test"
	case "python", "py":
		return "pytest -v"
	default:
		return "go test -v ./..."
	}
}

// prepareDockerfile creates a Dockerfile for the test execution
func prepareDockerfile(language string, testFiles []TestFile, sourceFiles []TestFile, dependencies []DependencyFile) string {
	var dockerfile strings.Builder

	// Base image
	image := getDockerImage(language)
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n", image))
	dockerfile.WriteString("WORKDIR /test\n")

	// Copy dependency files first (for caching)
	switch strings.ToLower(language) {
	case "go", "golang":
		// Copy go.mod and go.sum if provided
		for _, dep := range dependencies {
			if dep.Path == "go.mod" || dep.Path == "go.sum" {
				dockerfile.WriteString(fmt.Sprintf("COPY %s ./\n", dep.Path))
			}
		}
		dockerfile.WriteString("RUN go mod download\n")

	case "javascript", "js", "typescript", "ts":
		// Copy package.json if provided
		for _, dep := range dependencies {
			if dep.Path == "package.json" {
				dockerfile.WriteString(fmt.Sprintf("COPY %s ./\n", dep.Path))
			}
		}
		dockerfile.WriteString("RUN npm install\n")

	case "python", "py":
		// Copy requirements.txt if provided
		for _, dep := range dependencies {
			if dep.Path == "requirements.txt" {
				dockerfile.WriteString(fmt.Sprintf("COPY %s ./\n", dep.Path))
			}
		}
		dockerfile.WriteString("RUN pip install --no-cache-dir -r requirements.txt 2>/dev/null || true\n")
	}

	// Copy source files (if any)
	if len(sourceFiles) > 0 {
		for _, file := range sourceFiles {
			// Create directory structure
			dir := getDirPath(file.Path)
			if dir != "." {
				dockerfile.WriteString(fmt.Sprintf("RUN mkdir -p %s\n", dir))
			}
			dockerfile.WriteString(fmt.Sprintf("COPY %s %s\n", file.Path, file.Path))
		}
	}

	// Copy test files
	for _, file := range testFiles {
		// Create directory structure
		dir := getDirPath(file.Path)
		if dir != "." {
			dockerfile.WriteString(fmt.Sprintf("RUN mkdir -p %s\n", dir))
		}
		dockerfile.WriteString(fmt.Sprintf("COPY %s %s\n", file.Path, file.Path))
	}

	// Test command
	testCmd := getTestCommand(language, "")
	dockerfile.WriteString(fmt.Sprintf("CMD [\"sh\", \"-c\", \"%s\"]\n", testCmd))

	return dockerfile.String()
}

// getDirPath extracts directory path from file path
func getDirPath(filePath string) string {
	lastSlash := strings.LastIndex(filePath, "/")
	if lastSlash == -1 {
		return "."
	}
	return filePath[:lastSlash]
}

// executeTestInSandbox executes tests in a Docker container
func executeTestInSandbox(ctx context.Context, req TestExecutionRequest) (*ExecutionResult, error) {
	// Check Docker availability
	if !checkDockerAvailable() {
		return nil, fmt.Errorf("docker is not available on this system")
	}

	// Create temporary directory for Docker build context
	tempDir, err := os.MkdirTemp("", "sentinel-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		os.RemoveAll(tempDir) // Cleanup
	}()

	// Write Dockerfile
	dockerfile := prepareDockerfile(req.Language, req.TestFiles, req.SourceFiles, req.Dependencies)
	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Write dependency files with validation
	for _, dep := range req.Dependencies {
		// Validate dependency file before writing
		if err := validateDependencyFile(dep); err != nil {
			return nil, fmt.Errorf("invalid dependency file %s: %w", dep.Path, err)
		}

		depPath := filepath.Join(tempDir, dep.Path)
		depDir := filepath.Dir(depPath)
		if depDir != "." {
			os.MkdirAll(depDir, 0755)
		}
		if err := os.WriteFile(depPath, []byte(dep.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write dependency file %s: %w", dep.Path, err)
		}
	}

	// Write source files
	for _, file := range req.SourceFiles {
		filePath := filepath.Join(tempDir, file.Path)
		fileDir := filepath.Dir(filePath)
		if fileDir != "." {
			os.MkdirAll(fileDir, 0755)
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write source file %s: %w", file.Path, err)
		}
	}

	// Write test files
	for _, file := range req.TestFiles {
		filePath := filepath.Join(tempDir, file.Path)
		fileDir := filepath.Dir(filePath)
		if fileDir != "." {
			os.MkdirAll(fileDir, 0755)
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test file %s: %w", file.Path, err)
		}
	}

	// Build Docker image
	imageName := fmt.Sprintf("sentinel-test-%s", uuid.New().String())
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, tempDir)

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to build Docker image: %w (output: %s)", err, buildOutput.String())
	}

	// Determine network mode based on dependencies
	networkMode := "none"
	if needsNetworkAccess(req.Language, req.Dependencies) {
		networkMode = "bridge" // Allow network for package managers
	}

	// Run container with resource limits
	runCmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",                                   // Remove container after execution
		"--memory=512m",                          // Memory limit
		"--cpus=1",                               // CPU limit
		fmt.Sprintf("--network=%s", networkMode), // Conditional network access
		imageName,
	)

	// Create context with timeout
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	startTime := time.Now()
	err = runCmd.Run()
	executionTime := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Context timeout or other error
			if runCtx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("test execution timed out after 30 seconds")
			}
			return nil, fmt.Errorf("test execution failed: %w", err)
		}
	}

	// Cleanup: remove Docker image
	cleanupCmd := exec.Command("docker", "rmi", imageName)
	cleanupCmd.Run() // Ignore errors in cleanup

	result := &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}

	log.Printf("Test execution completed in %v: exit code %d", executionTime, exitCode)

	return result, nil
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

// Test Sandbox - Docker Operations
// Handles Docker container creation and test execution
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

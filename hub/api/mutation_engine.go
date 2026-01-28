// Phase 10D: Mutation Testing Engine
// Generates code mutants and calculates mutation score

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// MutationResult represents mutation testing results
type MutationResult struct {
	ID                string    `json:"id"`
	TestRequirementID string    `json:"test_requirement_id"`
	MutationScore     float64   `json:"mutation_score"` // 0.0 to 1.0
	TotalMutants      int       `json:"total_mutants"`
	KilledMutants     int       `json:"killed_mutants"`
	SurvivedMutants   int       `json:"survived_mutants"`
	ExecutionTimeMs   int       `json:"execution_time_ms"`
	CreatedAt         time.Time `json:"created_at"`
}

// MutationTestRequest represents the request to run mutation testing
type MutationTestRequest struct {
	ProjectID         string `json:"project_id"`
	SourceCode        string `json:"sourceCode"` // Code to mutate
	SourcePath        string `json:"sourcePath"` // File path
	Language          string `json:"language"`
	TestCode          string `json:"testCode"`                    // Test code to run
	TestRequirementID string `json:"testRequirementId,omitempty"` // Optional: specific requirement
}

// MutationTestResponse represents the response
type MutationTestResponse struct {
	Success         bool    `json:"success"`
	MutationScore   float64 `json:"mutationScore"` // 0.0 to 1.0
	TotalMutants    int     `json:"totalMutants"`
	KilledMutants   int     `json:"killedMutants"`
	SurvivedMutants int     `json:"survivedMutants"`
	ExecutionTimeMs int     `json:"executionTimeMs"`
	Message         string  `json:"message,omitempty"`
}

// Mutant represents a code mutant
type Mutant struct {
	ID       string `json:"id"`
	Original string `json:"original"` // Original code
	Mutated  string `json:"mutated"`  // Mutated code
	Operator string `json:"operator"` // Mutation operator used
	Line     int    `json:"line"`     // Line number where mutation occurred
	Killed   bool   `json:"killed"`   // Whether tests killed this mutant
}

// Mutation cache for performance
type mutationCacheEntry struct {
	Result  MutationResult
	Expires time.Time
}

var mutationCache = make(map[string]*mutationCacheEntry)
var mutationCacheMutex sync.RWMutex
var mutationCacheTTL = 1 * time.Hour // Cache mutation results for 1 hour

// applyMutation applies a mutation to source code using line-number-based replacement
func applyMutation(sourceCode string, mutant Mutant) string {
	// Use line-number-based replacement for accuracy
	if mutant.Line > 0 {
		lines := strings.Split(sourceCode, "\n")
		if mutant.Line <= len(lines) {
			// Line numbers are 1-indexed, array is 0-indexed
			lines[mutant.Line-1] = mutant.Mutated
			return strings.Join(lines, "\n")
		}
	}
	// Fallback to string replacement if line number invalid
	return strings.Replace(sourceCode, mutant.Original, mutant.Mutated, 1)
}

// cleanupMutationCache periodically cleans up expired cache entries
func cleanupMutationCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mutationCacheMutex.Lock()
		now := time.Now()
		for key, entry := range mutationCache {
			if now.After(entry.Expires) {
				delete(mutationCache, key)
			}
		}
		mutationCacheMutex.Unlock()
	}
}

func init() {
	// Start cache cleanup goroutine
	go cleanupMutationCache()
}

// generateMutants generates mutants for source code (file-level, limited per function)
func generateMutants(sourceCode string, language string, maxMutantsPerFunction int) []Mutant {
	var mutants []Mutant
	mutantID := 0

	lines := strings.Split(sourceCode, "\n")

	// Limit total mutants to prevent excessive execution time
	maxTotalMutants := 50
	if maxMutantsPerFunction > 0 {
		maxTotalMutants = maxMutantsPerFunction * 10 // Assume ~10 functions per file
	}

	// Track mutations per function to limit them
	mutationsPerFunction := make(map[string]int)
	currentFunction := "global"

	for lineNum, line := range lines {
		// Detect function boundaries (simplified)
		if isFunctionStart(line, language) {
			// Extract function name
			currentFunction = extractFunctionNameForMutation(line, language)
			mutationsPerFunction[currentFunction] = 0
		}

		// Skip if we've hit the limit for this function
		if mutationsPerFunction[currentFunction] >= maxMutantsPerFunction {
			continue
		}

		// Skip if we've hit total limit
		if len(mutants) >= maxTotalMutants {
			break
		}

		// Generate mutants for this line
		lineMutants := generateLineMutants(line, lineNum+1, language)

		for _, mutant := range lineMutants {
			if mutationsPerFunction[currentFunction] >= maxMutantsPerFunction {
				break
			}
			mutant.ID = fmt.Sprintf("mutant_%d", mutantID)
			mutantID++
			mutants = append(mutants, mutant)
			mutationsPerFunction[currentFunction]++
		}
	}

	return mutants
}

// isFunctionStart checks if a line starts a function definition
func isFunctionStart(line string, language string) bool {
	lineTrimmed := strings.TrimSpace(line)
	switch strings.ToLower(language) {
	case "go", "golang":
		return strings.HasPrefix(lineTrimmed, "func ")
	case "javascript", "js", "typescript", "ts":
		return strings.Contains(lineTrimmed, "function ") ||
			strings.Contains(lineTrimmed, "=>") ||
			regexp.MustCompile(`^\s*(const|let|var)\s+\w+\s*=\s*\(`).MatchString(lineTrimmed)
	case "python", "py":
		return strings.HasPrefix(lineTrimmed, "def ")
	default:
		return false
	}
}

// extractFunctionNameForMutation extracts function name from function definition line (for mutation testing)
func extractFunctionNameForMutation(line string, language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		// func FunctionName(...)
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "func" && i+1 < len(parts) {
				funcName := parts[i+1]
				// Remove receiver if present
				if strings.Contains(funcName, "(") {
					continue
				}
				return funcName
			}
		}
	case "javascript", "js", "typescript", "ts":
		// function name(...) or const name = (...)
		if match := regexp.MustCompile(`function\s+(\w+)`).FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
		if match := regexp.MustCompile(`(const|let|var)\s+(\w+)\s*=`).FindStringSubmatch(line); len(match) > 2 {
			return match[2]
		}
	case "python", "py":
		// def name(...)
		if match := regexp.MustCompile(`def\s+(\w+)`).FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
	}
	return "unknown"
}

// generateLineMutants generates mutants for a single line of code
func generateLineMutants(line string, lineNum int, language string) []Mutant {
	var mutants []Mutant

	// Arithmetic operator mutations: + → -, * → /, etc.
	arithmeticOps := map[string][]string{
		"+": {"-", "*"},
		"-": {"+", "*"},
		"*": {"/", "+"},
		"/": {"*", "-"},
	}

	for op, replacements := range arithmeticOps {
		if strings.Contains(line, op) {
			for _, replacement := range replacements {
				mutated := strings.Replace(line, op, replacement, 1)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("arithmetic_%s_to_%s", op, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Comparison operator mutations: == → !=, < → <=, etc.
	comparisonOps := map[string][]string{
		"==": {"!=", "<", ">"},
		"!=": {"==", "<", ">"},
		"<":  {"<=", ">", "=="},
		">":  {">=", "<", "=="},
		"<=": {"<", ">", "=="},
		">=": {">", "<", "=="},
	}

	for op, replacements := range comparisonOps {
		if strings.Contains(line, op) {
			for _, replacement := range replacements {
				mutated := strings.Replace(line, op, replacement, 1)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("comparison_%s_to_%s", op, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Boolean operator mutations: && → ||, ! → remove
	if strings.Contains(line, "&&") {
		mutated := strings.Replace(line, "&&", "||", 1)
		mutants = append(mutants, Mutant{
			Original: line,
			Mutated:  mutated,
			Operator: "boolean_and_to_or",
			Line:     lineNum,
		})
	}
	if strings.Contains(line, "||") {
		mutated := strings.Replace(line, "||", "&&", 1)
		mutants = append(mutants, Mutant{
			Original: line,
			Mutated:  mutated,
			Operator: "boolean_or_to_and",
			Line:     lineNum,
		})
	}

	// Constant mutations: 1 → 0, true → false, etc.
	constantMutations := map[string]string{
		"1":     "0",
		"0":     "1",
		"true":  "false",
		"false": "true",
		"nil":   "not_nil", // Special case - would need proper replacement
		"null":  "not_null",
	}

	for constant, replacement := range constantMutations {
		if strings.Contains(line, constant) {
			// Only replace whole-word matches
			re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(constant)))
			if re.MatchString(line) {
				mutated := re.ReplaceAllString(line, replacement)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("constant_%s_to_%s", constant, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Return value mutations: return x → return nil (for languages that support it)
	if strings.Contains(line, "return") && !strings.Contains(line, "return nil") && !strings.Contains(line, "return null") {
		// Extract return value
		if match := regexp.MustCompile(`return\s+(\S+)`).FindStringSubmatch(line); len(match) > 1 {
			returnValue := match[1]
			var nilValue string
			switch strings.ToLower(language) {
			case "go", "golang":
				nilValue = "nil"
			case "javascript", "js", "typescript", "ts":
				nilValue = "null"
			case "python", "py":
				nilValue = "None"
			default:
				nilValue = "nil"
			}
			mutated := strings.Replace(line, returnValue, nilValue, 1)
			mutants = append(mutants, Mutant{
				Original: line,
				Mutated:  mutated,
				Operator: fmt.Sprintf("return_%s_to_%s", returnValue, nilValue),
				Line:     lineNum,
			})
		}
	}

	return mutants
}

// executeTestsAgainstMutants executes tests against each mutant using sandbox
func executeTestsAgainstMutants(ctx context.Context, mutants []Mutant, sourceCode string, sourcePath string, testCode string, language string, projectID string) (int, int, error) {
	killedCount := 0
	survivedCount := 0

	// Limit concurrent executions
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range mutants {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(mutant Mutant) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Create mutated source code using line-number-based replacement
			mutatedSource := applyMutation(sourceCode, mutant)

			// Prepare test execution request
			testReq := TestExecutionRequest{
				ProjectID:     projectID,
				ExecutionType: "mutation",
				Language:      language,
				TestFiles: []TestFile{
					{
						Path:    sourcePath + "_test",
						Content: testCode,
					},
				},
				SourceFiles: []TestFile{
					{
						Path:    sourcePath,
						Content: mutatedSource,
					},
				},
			}

			// Execute test with timeout
			mutantCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			result, err := executeTestInSandbox(mutantCtx, testReq)

			mu.Lock()
			if err != nil {
				// Distinguish between test failure and execution error
				errMsg := err.Error()
				if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "failed to build") || strings.Contains(errMsg, "Docker is not available") {
					// Execution error - don't count as killed, log and skip
					log.Printf("Mutant %s execution error (not counted): %v", mutant.ID, err)
				} else {
					// Test failure - mutant killed
					killedCount++
					mutant.Killed = true
				}
			} else if result.ExitCode != 0 {
				// Tests failed - mutant killed
				killedCount++
				mutant.Killed = true
			} else {
				// Tests passed - mutant survived
				survivedCount++
				mutant.Killed = false
			}
			mu.Unlock()
		}(mutants[i])
	}

	wg.Wait()

	return killedCount, survivedCount, nil
}

// calculateMutationScore calculates mutation score
func calculateMutationScore(totalMutants, killedMutants int) float64 {
	if totalMutants == 0 {
		return 0.0
	}
	return float64(killedMutants) / float64(totalMutants)
}

// getCachedMutationResult checks cache for mutation results
func getCachedMutationResult(sourceCodeHash string) (*MutationResult, bool) {
	mutationCacheMutex.RLock()
	defer mutationCacheMutex.RUnlock()

	if entry, ok := mutationCache[sourceCodeHash]; ok {
		if time.Now().Before(entry.Expires) {
			return &entry.Result, true
		}
		// Expired, remove from cache
		delete(mutationCache, sourceCodeHash)
	}

	return nil, false
}

// cacheMutationResult caches mutation result
func cacheMutationResult(sourceCodeHash string, result MutationResult) {
	mutationCacheMutex.Lock()
	defer mutationCacheMutex.Unlock()

	mutationCache[sourceCodeHash] = &mutationCacheEntry{
		Result:  result,
		Expires: time.Now().Add(mutationCacheTTL),
	}
}

// hashSourceCode creates a hash of source code for caching
func hashSourceCode(sourceCode, testCode string) string {
	combined := sourceCode + "|||" + testCode
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// saveMutationResult saves mutation result to database
func saveMutationResult(ctx context.Context, result MutationResult) error {
	query := `
		INSERT INTO mutation_results 
		(id, test_requirement_id, mutation_score, total_mutants, killed_mutants, survived_mutants, execution_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			mutation_score = EXCLUDED.mutation_score,
			total_mutants = EXCLUDED.total_mutants,
			killed_mutants = EXCLUDED.killed_mutants,
			survived_mutants = EXCLUDED.survived_mutants,
			execution_time_ms = EXCLUDED.execution_time_ms
	`

	_, err := execWithTimeout(ctx, query,
		result.ID, result.TestRequirementID, result.MutationScore,
		result.TotalMutants, result.KilledMutants, result.SurvivedMutants,
		result.ExecutionTimeMs, result.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save mutation result: %w", err)
	}

	return nil
}

// mutationTestHandler handles the API request to run mutation testing
func mutationTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MutationTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	if req.SourceCode == "" {
		http.Error(w, "sourceCode is required", http.StatusBadRequest)
		return
	}
	if req.TestCode == "" {
		http.Error(w, "testCode is required", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		http.Error(w, "language is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second) // 5 minutes for mutation testing
	defer cancel()

	// Check cache
	sourceHash := hashSourceCode(req.SourceCode, req.TestCode)
	if cached, ok := getCachedMutationResult(sourceHash); ok {
		response := MutationTestResponse{
			Success:         true,
			MutationScore:   cached.MutationScore,
			TotalMutants:    cached.TotalMutants,
			KilledMutants:   cached.KilledMutants,
			SurvivedMutants: cached.SurvivedMutants,
			ExecutionTimeMs: cached.ExecutionTimeMs,
			Message:         "Result from cache",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	startTime := time.Now()

	// Generate mutants (limit to 20 per function to control execution time)
	mutants := generateMutants(req.SourceCode, req.Language, 20)

	if len(mutants) == 0 {
		response := MutationTestResponse{
			Success:         false,
			MutationScore:   0.0,
			TotalMutants:    0,
			KilledMutants:   0,
			SurvivedMutants: 0,
			ExecutionTimeMs: 0,
			Message:         "No mutants generated - code may be too simple or unsupported",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Execute tests against mutants
	killedCount, survivedCount, err := executeTestsAgainstMutants(
		ctx, mutants, req.SourceCode, req.SourcePath, req.TestCode, req.Language, req.ProjectID,
	)

	executionTime := time.Since(startTime)

	if err != nil {
		log.Printf("Error executing tests against mutants: %v", err)
		http.Error(w, fmt.Sprintf("Mutation testing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate mutation score
	totalMutants := len(mutants)
	mutationScore := calculateMutationScore(totalMutants, killedCount)

	// Save result
	result := MutationResult{
		ID:                uuid.New().String(),
		TestRequirementID: req.TestRequirementID,
		MutationScore:     mutationScore,
		TotalMutants:      totalMutants,
		KilledMutants:     killedCount,
		SurvivedMutants:   survivedCount,
		ExecutionTimeMs:   int(executionTime.Milliseconds()),
		CreatedAt:         time.Now(),
	}

	if err := saveMutationResult(ctx, result); err != nil {
		log.Printf("Error saving mutation result: %v", err)
		// Continue anyway
	}

	// Cache result
	cacheMutationResult(sourceHash, result)

	response := MutationTestResponse{
		Success:         true,
		MutationScore:   mutationScore,
		TotalMutants:    totalMutants,
		KilledMutants:   killedCount,
		SurvivedMutants: survivedCount,
		ExecutionTimeMs: int(executionTime.Milliseconds()),
		Message:         fmt.Sprintf("Generated %d mutants, %d killed, %d survived", totalMutants, killedCount, survivedCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getMutationResultHandler handles GET request to retrieve mutation result
func getMutationResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testRequirementID := chi.URLParam(r, "test_requirement_id")
	if testRequirementID == "" {
		http.Error(w, "test_requirement_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, test_requirement_id, mutation_score, total_mutants, killed_mutants, 
		       survived_mutants, execution_time_ms, created_at
		FROM mutation_results
		WHERE test_requirement_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := queryRowWithTimeout(ctx, query, testRequirementID)

	var result MutationResult
	err := row.Scan(
		&result.ID, &result.TestRequirementID, &result.MutationScore,
		&result.TotalMutants, &result.KilledMutants, &result.SurvivedMutants,
		&result.ExecutionTimeMs, &result.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Mutation result not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying mutation result: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query mutation result: %v", err), http.StatusInternalServerError)
		return
	}

	response := MutationTestResponse{
		Success:         true,
		MutationScore:   result.MutationScore,
		TotalMutants:    result.TotalMutants,
		KilledMutants:   result.KilledMutants,
		SurvivedMutants: result.SurvivedMutants,
		ExecutionTimeMs: result.ExecutionTimeMs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

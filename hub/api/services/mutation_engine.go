// Mutation Testing Engine - Main Handler and Types
// Handles mutation testing requests and caching
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

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
	"strings"
	"sync"
	"time"

	"sentinel-hub-api/pkg/database"

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

	_, err := database.ExecWithTimeout(ctx, db, query,
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

	row := database.QueryRowWithTimeout(ctx, db, query, testRequirementID)

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

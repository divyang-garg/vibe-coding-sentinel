// Mutation Engine - Testing Functions
// Executes tests against mutants and calculates mutation scores
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

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

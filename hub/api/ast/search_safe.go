// Package ast provides safe search operations with timeout and concurrency
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"context"
	"errors"
	"sync"
	"time"
)

// DefaultSearchTimeout is the default timeout for codebase searches
const DefaultSearchTimeout = 30 * time.Second

// DefaultMaxConcurrency is the default maximum concurrent validations
const DefaultMaxConcurrency = 4

// SearchWithTimeout wraps SearchCodebase with timeout protection
func SearchWithTimeout(pattern, projectRoot string, timeout time.Duration) ([]SearchResult, error) {
	if timeout <= 0 {
		timeout = DefaultSearchTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Channel to receive results
	resultChan := make(chan searchResultOrError, 1)

	// Run search in goroutine
	go func() {
		results, err := SearchCodebase(pattern, projectRoot, nil)
		resultChan <- searchResultOrError{results: results, err: err}
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, errors.New("search timeout exceeded")
	case result := <-resultChan:
		return result.results, result.err
	}
}

// searchResultOrError wraps search results or error
type searchResultOrError struct {
	results []SearchResult
	err     error
}

// ValidateFindingsConcurrent validates multiple findings in parallel
func ValidateFindingsConcurrent(findings []ASTFinding, filePath, projectRoot, language string, maxConcurrency int) error {
	if maxConcurrency <= 0 {
		maxConcurrency = DefaultMaxConcurrency
	}

	// Create semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// Validate each finding
	for i := range findings {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Validate finding
			err := ValidateFinding(&findings[idx], filePath, projectRoot, language)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(i)
	}

	// Wait for all validations to complete
	wg.Wait()

	return firstErr
}

// ValidateFindingsConcurrentWithTimeout validates findings with both concurrency and timeout
func ValidateFindingsConcurrentWithTimeout(
	findings []ASTFinding,
	filePath, projectRoot, language string,
	maxConcurrency int,
	timeout time.Duration,
) error {
	if timeout <= 0 {
		timeout = DefaultSearchTimeout * time.Duration(len(findings))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Channel to receive completion signal
	done := make(chan error, 1)

	// Run concurrent validation
	go func() {
		done <- ValidateFindingsConcurrent(findings, filePath, projectRoot, language, maxConcurrency)
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		return errors.New("validation timeout exceeded")
	case err := <-done:
		return err
	}
}

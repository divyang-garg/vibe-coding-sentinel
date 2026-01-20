// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker prevents cascading failures by opening circuit after threshold failures
type CircuitBreaker struct {
	failures    int
	lastFailure time.Time
	threshold   int
	timeout     time.Duration
	mu          sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.RLock()
	if cb.failures >= cb.threshold {
		if time.Since(cb.lastFailure) < cb.timeout {
			cb.mu.RUnlock()
			return fmt.Errorf("circuit breaker is open")
		}
		// Reset after timeout
		cb.mu.RUnlock()
		cb.mu.Lock()
		cb.failures = 0
		cb.mu.Unlock()
	} else {
		cb.mu.RUnlock()
	}

	err := fn()
	if err != nil {
		cb.mu.Lock()
		cb.failures++
		cb.lastFailure = time.Now()
		cb.mu.Unlock()
	} else {
		// Reset on success
		cb.mu.Lock()
		if cb.failures > 0 {
			cb.failures = 0
		}
		cb.mu.Unlock()
	}

	return err
}

// IsOpen returns true if circuit breaker is currently open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures >= cb.threshold && time.Since(cb.lastFailure) < cb.timeout
}

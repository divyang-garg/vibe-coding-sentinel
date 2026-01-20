// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("allows calls when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(3, time.Second)
		
		callCount := 0
		err := cb.Call(func() error {
			callCount++
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.False(t, cb.IsOpen())
	})

	t.Run("opens after threshold failures", func(t *testing.T) {
		cb := NewCircuitBreaker(2, time.Second)
		
		// First failure
		_ = cb.Call(func() error { return errors.New("fail 1") })
		assert.False(t, cb.IsOpen())
		
		// Second failure - should open
		_ = cb.Call(func() error { return errors.New("fail 2") })
		assert.True(t, cb.IsOpen())
		
		// Should reject call when open
		err := cb.Call(func() error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})

	t.Run("resets on success", func(t *testing.T) {
		cb := NewCircuitBreaker(3, time.Second)
		
		// One failure
		_ = cb.Call(func() error { return errors.New("fail") })
		
		// Success resets
		_ = cb.Call(func() error { return nil })
		
		// More failures should require threshold again
		_ = cb.Call(func() error { return errors.New("fail 1") })
		_ = cb.Call(func() error { return errors.New("fail 2") })
		assert.False(t, cb.IsOpen()) // Not yet at threshold
	})

	t.Run("allows retry after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 100*time.Millisecond)
		
		// Open circuit
		_ = cb.Call(func() error { return errors.New("fail") })
		assert.True(t, cb.IsOpen())
		
		// Wait for timeout
		time.Sleep(150 * time.Millisecond)
		
		// Should allow call again
		callCount := 0
		err := cb.Call(func() error {
			callCount++
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.False(t, cb.IsOpen())
	})
}

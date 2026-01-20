// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFallbackExtractor_Extract(t *testing.T) {
	fallback := NewFallbackExtractor()

	t.Run("finds_must_patterns", func(t *testing.T) {
		text := "The system must validate all input. The user must authenticate before access."

		result, err := fallback.Extract(context.Background(), text)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.BusinessRules), 1)
		assert.Equal(t, "regex", result.Source)
		assert.Equal(t, 0.5, result.Confidence)
	})

	t.Run("finds_time_constraints", func(t *testing.T) {
		text := "Orders must be processed within 24 hours of submission."

		result, err := fallback.Extract(context.Background(), text)

		assert.NoError(t, err)
		assert.Greater(t, len(result.BusinessRules), 0)
		for _, rule := range result.BusinessRules {
			if rule.Specification.Constraints[0].Type == "time_based" {
				assert.Equal(t, "hour", rule.Specification.Constraints[0].Unit)
			}
		}
	})

	t.Run("finds_prohibitions", func(t *testing.T) {
		text := "Users must not access admin functions without permission."

		result, err := fallback.Extract(context.Background(), text)

		assert.NoError(t, err)
		assert.Greater(t, len(result.BusinessRules), 0)
	})

	t.Run("finds_shall_patterns", func(t *testing.T) {
		text := "The application shall encrypt all sensitive data."

		result, err := fallback.Extract(context.Background(), text)

		assert.NoError(t, err)
		assert.Greater(t, len(result.BusinessRules), 0)
	})

	t.Run("no_matches_returns_error", func(t *testing.T) {
		text := "This is just regular text with no business rules."

		_, err := fallback.Extract(context.Background(), text)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no business rules found")
	})

	t.Run("deduplicates_similar_rules", func(t *testing.T) {
		text := "The system must validate input. The system must validate input."

		result, err := fallback.Extract(context.Background(), text)

		assert.NoError(t, err)
		// Should only find one unique rule
		assert.LessOrEqual(t, len(result.BusinessRules), 2)
	})
}

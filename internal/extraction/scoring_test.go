// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfidenceScorer_ScoreRule(t *testing.T) {
	scorer := NewConfidenceScorer()

	t.Run("complete_rule_high_score", func(t *testing.T) {
		rule := BusinessRule{
			Title:       "Complete Rule",
			Description: "This is a complete description with enough detail",
			Specification: Specification{
				Constraints: []Constraint{{
					ID:         "C1",
					Pseudocode: "order.age < 24h",
				}},
			},
			Traceability: Traceability{
				SourceDocument: "requirements.md",
			},
		}

		score := scorer.ScoreRule(rule)

		assert.Greater(t, score, 0.8)
	})

	t.Run("minimal_rule_low_score", func(t *testing.T) {
		rule := BusinessRule{
			Title: "Minimal",
		}

		score := scorer.ScoreRule(rule)

		assert.Less(t, score, 0.3)
	})

	t.Run("rule_with_title_only", func(t *testing.T) {
		rule := BusinessRule{
			Title: "Test Rule",
		}

		score := scorer.ScoreRule(rule)

		assert.Greater(t, score, 0.0)
		assert.Less(t, score, 0.2)
	})

	t.Run("rule_with_constraints", func(t *testing.T) {
		rule := BusinessRule{
			Title: "Test",
			Specification: Specification{
				Constraints: []Constraint{{ID: "C1", Type: "state_based"}},
			},
		}

		score := scorer.ScoreRule(rule)

		assert.Greater(t, score, 0.3)
	})

	t.Run("score_overall", func(t *testing.T) {
		rules := []BusinessRule{
			{Title: "Rule 1", Description: "Description 1", Specification: Specification{Constraints: []Constraint{{ID: "C1"}}}},
			{Title: "Rule 2", Description: "Description 2", Specification: Specification{Constraints: []Constraint{{ID: "C2"}}}},
		}

		score := scorer.ScoreOverall(rules)

		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("score_overall_empty_list", func(t *testing.T) {
		rules := []BusinessRule{}

		score := scorer.ScoreOverall(rules)

		assert.Equal(t, 0.0, score)
	})

	t.Run("score_overall_with_existing_confidence", func(t *testing.T) {
		rules := []BusinessRule{
			{Title: "Rule 1", Confidence: 0.9},
			{Title: "Rule 2", Confidence: 0.8},
		}

		score := scorer.ScoreOverall(rules)

		assert.InDelta(t, 0.85, score, 0.01)
	})
}

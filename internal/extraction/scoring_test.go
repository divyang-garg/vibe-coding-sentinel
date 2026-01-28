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
			ID:          "BR1",
			Title:       "Complete Rule",
			Description: "Users must authenticate before accessing the system. The system shall validate credentials within 24 hours. Minimum password length is 8 characters.",
			Specification: Specification{
				Constraints: []Constraint{{
					ID:         "C1",
					Type:       "state_based",
					Pseudocode: "order.age < 24h",
				}},
			},
			Traceability: Traceability{
				SourceDocument: "requirements.md",
				SourceQuote:    "Section 3.2.1",
			},
			Priority: "high",
			Status:   "approved",
		}

		score := scorer.ScoreRule(rule)

		assert.Greater(t, score, 0.8, "Complete rule with all fields should score >0.8")
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
			ID:          "BR1",
			Title:       "Test Rule with Constraints",
			Description: "This rule has constraints",
			Specification: Specification{
				Constraints: []Constraint{{ID: "C1", Type: "state_based"}},
			},
		}

		score := scorer.ScoreRule(rule)

		// Rule with title, description, and constraints should score >0.3
		// Structure: 0.35 (title) + 0.40 (desc) = 0.75 -> 0.75*0.30 = 0.225
		// Semantics: 0.5 (desc length) -> 0.5*0.20 = 0.10
		// Traceability: 0 -> 0*0.25 = 0
		// Constraints: 0.9 (constraints + type) -> 0.9*0.25 = 0.225
		// Total: ~0.55, so >0.3 is reasonable
		assert.Greater(t, score, 0.3, "Rule with constraints should score >0.3")
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

func TestConfidenceScorer_EdgeCases(t *testing.T) {
	scorer := NewConfidenceScorer()

	t.Run("rule with ID only", func(t *testing.T) {
		rule := BusinessRule{
			ID: "BR1",
		}
		score := scorer.ScoreRule(rule)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("rule with traceability only", func(t *testing.T) {
		rule := BusinessRule{
			Title: "Test",
			Traceability: Traceability{
				SourceDocument: "doc.md",
			},
		}
		score := scorer.ScoreRule(rule)
		// Traceability alone gives 0.25 weight * 1.0 = 0.25
		// Plus structure: title 0.35 * 0.30 = 0.105
		// Total: ~0.355, should be >0.3
		assert.Greater(t, score, 0.3)
	})

	t.Run("rule with pseudocode constraint", func(t *testing.T) {
		rule := BusinessRule{
			Title:       "Test",
			Description: "Test description with enough characters to qualify",
			Specification: Specification{
				Constraints: []Constraint{{
					ID:         "C1",
					Pseudocode: "value > 0",
				}},
			},
		}
		score := scorer.ScoreRule(rule)
		assert.Greater(t, score, 0.4)
	})

	t.Run("semantics score with action words", func(t *testing.T) {
		rule := BusinessRule{
			Title:       "Test",
			Description: "Users must authenticate. Password should be at least 8 characters long.",
		}
		score := scorer.ScoreRule(rule)
		// Should get bonus for action words ("must", "should") and numbers
		assert.Greater(t, score, 0.2)
	})

	t.Run("semantics score without description", func(t *testing.T) {
		rule := BusinessRule{
			Title: "Test",
			// No description
		}
		score := scorer.ScoreRule(rule)
		// Semantics should be 0.0 when no description
		// Structure: title 0.35 * 0.30 = 0.105
		// Total should be low
		assert.Less(t, score, 0.2)
	})

	t.Run("score_traceability_with_source_document", func(t *testing.T) {
		scorer := NewConfidenceScorer().(*confidenceScorer)
		rule := BusinessRule{
			Title: "Test",
			Traceability: Traceability{
				SourceDocument: "requirements.md",
			},
		}
		score := scorer.scoreTraceability(rule)
		assert.Equal(t, 1.0, score, "SourceDocument should give full score")
	})

	t.Run("score_traceability_with_source_quote", func(t *testing.T) {
		scorer := NewConfidenceScorer().(*confidenceScorer)
		rule := BusinessRule{
			Title: "Test",
			Traceability: Traceability{
				SourceQuote: "Section 3.2.1",
			},
		}
		score := scorer.scoreTraceability(rule)
		assert.Equal(t, 1.0, score, "SourceQuote should give full score")
	})

	t.Run("score_traceability_without_traceability", func(t *testing.T) {
		scorer := NewConfidenceScorer().(*confidenceScorer)
		rule := BusinessRule{
			Title: "Test",
		}
		score := scorer.scoreTraceability(rule)
		assert.Equal(t, 0.0, score, "No traceability should give 0 score")
	})

	t.Run("score_traceability_with_both_source_document_and_quote", func(t *testing.T) {
		scorer := NewConfidenceScorer().(*confidenceScorer)
		rule := BusinessRule{
			Title: "Test",
			Traceability: Traceability{
				SourceDocument: "requirements.md",
				SourceQuote:    "Section 3.2.1",
			},
		}
		score := scorer.scoreTraceability(rule)
		assert.Equal(t, 1.0, score, "Both should give full score")
	})
}

func TestClassifyConfidence(t *testing.T) {
	t.Run("high confidence", func(t *testing.T) {
		assert.Equal(t, ConfidenceHigh, ClassifyConfidence(0.8))
		assert.Equal(t, ConfidenceHigh, ClassifyConfidence(0.9))
		assert.Equal(t, ConfidenceHigh, ClassifyConfidence(1.0))
	})

	t.Run("medium confidence", func(t *testing.T) {
		assert.Equal(t, ConfidenceMedium, ClassifyConfidence(0.5))
		assert.Equal(t, ConfidenceMedium, ClassifyConfidence(0.7))
		assert.Equal(t, ConfidenceMedium, ClassifyConfidence(0.79))
	})

	t.Run("low confidence", func(t *testing.T) {
		assert.Equal(t, ConfidenceLow, ClassifyConfidence(0.0))
		assert.Equal(t, ConfidenceLow, ClassifyConfidence(0.3))
		assert.Equal(t, ConfidenceLow, ClassifyConfidence(0.49))
	})
}

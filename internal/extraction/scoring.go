// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"math"
	"regexp"
	"strings"
)

// ConfidenceScorer calculates confidence scores for extracted rules
type ConfidenceScorer interface {
	ScoreRule(rule BusinessRule) float64
	ScoreOverall(rules []BusinessRule) float64
}

// ConfidenceLevel represents confidence classification
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"   // >= 0.8
	ConfidenceMedium ConfidenceLevel = "medium" // 0.5-0.8
	ConfidenceLow    ConfidenceLevel = "low"    // < 0.5
)

// ClassifyConfidence classifies a confidence score into a level
func ClassifyConfidence(score float64) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceHigh
	case score >= 0.5:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}

// confidenceScorer implements ConfidenceScorer
type confidenceScorer struct {
	weights ConfidenceWeights
}

// ConfidenceWeights defines scoring weights
type ConfidenceWeights struct {
	TitlePresent        float64
	DescriptionPresent  float64
	ConstraintsPresent  float64
	PseudocodePresent   float64
	TraceabilityPresent float64
}

// NewConfidenceScorer creates a scorer with default weights
func NewConfidenceScorer() ConfidenceScorer {
	return &confidenceScorer{
		weights: ConfidenceWeights{
			TitlePresent:        0.15,
			DescriptionPresent:  0.15,
			ConstraintsPresent:  0.30,
			PseudocodePresent:   0.25,
			TraceabilityPresent: 0.15,
		},
	}
}

// ScoreRule calculates confidence for a single rule using multi-factor analysis
func (s *confidenceScorer) ScoreRule(rule BusinessRule) float64 {
	// Structural completeness (30%)
	structScore := s.scoreStructure(rule)

	// Semantic quality (20%)
	semanticScore := s.scoreSemantics(rule)

	// Traceability (25%) - increased importance
	traceScore := s.scoreTraceability(rule)

	// Constraint quality (25%) - increased weight as constraints are critical
	constraintScore := s.scoreConstraints(rule)

	score := structScore*0.30 + semanticScore*0.20 + traceScore*0.25 + constraintScore*0.25

	return math.Min(score, 1.0)
}

// scoreStructure evaluates structural completeness (30% weight)
func (s *confidenceScorer) scoreStructure(rule BusinessRule) float64 {
	score := 0.0
	if rule.ID != "" {
		score += 0.15
	}
	if rule.Title != "" {
		score += 0.35 // Increased - title is essential
	}
	if len(rule.Description) > 20 {
		score += 0.40 // Increased - description adds significant value
	}
	if rule.Priority != "" {
		score += 0.05
	}
	if rule.Status != "" {
		score += 0.05
	}
	return math.Min(score, 1.0)
}

// scoreSemantics evaluates semantic quality (20% weight)
func (s *confidenceScorer) scoreSemantics(rule BusinessRule) float64 {
	score := 0.0
	desc := strings.ToLower(rule.Description)

	// If no description, return minimal score
	if len(rule.Description) == 0 {
		return 0.0
	}

	// Check for actionable language
	actionWords := []string{"must", "shall", "should", "will", "can", "may"}
	for _, word := range actionWords {
		if strings.Contains(desc, word) {
			score += 0.3
			break
		}
	}

	// Check for measurable criteria (numbers)
	if regexp.MustCompile(`\d+`).MatchString(desc) {
		score += 0.3
	}

	// Description length quality (50-500 chars is optimal)
	descLen := len(rule.Description)
	if descLen > 50 && descLen < 500 {
		score += 0.5 // Increased
	} else if descLen >= 20 {
		score += 0.5 // Increased - give substantial credit for reasonable descriptions
	}

	return math.Min(score, 1.0)
}

// scoreTraceability evaluates traceability (25% weight)
func (s *confidenceScorer) scoreTraceability(rule BusinessRule) float64 {
	score := 0.0
	if rule.Traceability.SourceDocument != "" {
		score = 1.0 // Full credit for SourceDocument - it's a strong indicator of quality
	}
	if rule.Traceability.SourceQuote != "" {
		// SourceQuote is bonus but not required for full score
		if score < 1.0 {
			score = 1.0
		}
	}
	return score
}

// scoreConstraints evaluates constraint quality (25% weight - increased importance)
func (s *confidenceScorer) scoreConstraints(rule BusinessRule) float64 {
	score := 0.0
	if len(rule.Specification.Constraints) > 0 {
		score += 0.7 // Increased - having constraints is very valuable

		// Bonus for pseudocode
		hasPseudocode := false
		hasType := false
		for _, c := range rule.Specification.Constraints {
			if c.Pseudocode != "" {
				hasPseudocode = true
			}
			if c.Type != "" {
				hasType = true
			}
		}
		if hasPseudocode {
			score += 0.3 // Bonus for pseudocode
		} else if hasType {
			score += 0.2 // Partial credit for having constraint type
		}
	}
	return math.Min(score, 1.0)
}

// ScoreOverall calculates average confidence across all rules
func (s *confidenceScorer) ScoreOverall(rules []BusinessRule) float64 {
	if len(rules) == 0 {
		return 0.0
	}

	total := 0.0
	for _, rule := range rules {
		if rule.Confidence > 0 {
			total += rule.Confidence
		} else {
			total += s.ScoreRule(rule)
		}
	}

	return total / float64(len(rules))
}

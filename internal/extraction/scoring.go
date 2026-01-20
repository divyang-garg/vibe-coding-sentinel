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
	// Structural completeness (40%)
	structScore := s.scoreStructure(rule)

	// Semantic quality (30%)
	semanticScore := s.scoreSemantics(rule)

	// Traceability (20%)
	traceScore := s.scoreTraceability(rule)

	// Constraint quality (10%)
	constraintScore := s.scoreConstraints(rule)

	score := structScore*0.4 + semanticScore*0.3 + traceScore*0.2 + constraintScore*0.1

	return math.Min(score, 1.0)
}

// scoreStructure evaluates structural completeness (40% weight)
func (s *confidenceScorer) scoreStructure(rule BusinessRule) float64 {
	score := 0.0
	if rule.ID != "" {
		score += 0.15
	}
	if rule.Title != "" {
		score += 0.25
	}
	if len(rule.Description) > 20 {
		score += 0.30
	}
	if rule.Priority != "" {
		score += 0.15
	}
	if rule.Status != "" {
		score += 0.15
	}
	return math.Min(score, 1.0)
}

// scoreSemantics evaluates semantic quality (30% weight)
func (s *confidenceScorer) scoreSemantics(rule BusinessRule) float64 {
	score := 0.0
	desc := strings.ToLower(rule.Description)

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
		score += 0.4
	} else if descLen >= 20 {
		score += 0.2 // Partial credit for reasonable length
	}

	return math.Min(score, 1.0)
}

// scoreTraceability evaluates traceability (20% weight)
func (s *confidenceScorer) scoreTraceability(rule BusinessRule) float64 {
	score := 0.0
	if rule.Traceability.SourceDocument != "" {
		score += 0.5
	}
	if rule.Traceability.SourceQuote != "" {
		score += 0.5
	}
	return score
}

// scoreConstraints evaluates constraint quality (10% weight)
func (s *confidenceScorer) scoreConstraints(rule BusinessRule) float64 {
	score := 0.0
	if len(rule.Specification.Constraints) > 0 {
		score += 0.5

		// Bonus for pseudocode
		hasPseudocode := false
		for _, c := range rule.Specification.Constraints {
			if c.Pseudocode != "" {
				hasPseudocode = true
				break
			}
		}
		if hasPseudocode {
			score += 0.5
		}
	}
	return score
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

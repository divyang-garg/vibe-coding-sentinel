// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// FallbackExtractor provides regex-based extraction when LLM fails
type FallbackExtractor interface {
	Extract(ctx context.Context, text string) (*ExtractResult, error)
}

// regexFallback implements FallbackExtractor
type regexFallback struct {
	patterns []rulePattern
}

type rulePattern struct {
	name    string
	regex   *regexp.Regexp
	extract func(matches []string) *BusinessRule
}

var (
	ruleCounter int
	ruleMutex   sync.Mutex
)

// NewFallbackExtractor creates a new regex-based fallback extractor
func NewFallbackExtractor() FallbackExtractor {
	return &regexFallback{
		patterns: []rulePattern{
			{
				name:  "must_pattern",
				regex: regexp.MustCompile(`(?i)(?:the\s+)?(?:system|user|application)\s+(?:must|shall|should)\s+(.+?)(?:\.|$)`),
				extract: func(matches []string) *BusinessRule {
					if len(matches) < 2 {
						return nil
					}
					return &BusinessRule{
						ID:          generateRuleID(),
						Version:     "1.0",
						Status:      "draft",
						Title:       truncate(matches[1], 100),
						Description: matches[0],
						Priority:    "medium",
						Specification: Specification{
							Constraints: []Constraint{{
								ID:         "C1",
								Type:       "state_based",
								Expression: matches[1],
							}},
						},
					}
				},
			},
			{
				name:  "time_constraint",
				regex: regexp.MustCompile(`(?i)within\s+(\d+)\s+(hours?|days?|minutes?|seconds?)`),
				extract: func(matches []string) *BusinessRule {
					if len(matches) < 3 {
						return nil
					}
					return &BusinessRule{
						ID:          generateRuleID(),
						Version:     "1.0",
						Status:      "draft",
						Title:       fmt.Sprintf("Time Constraint: %s %s", matches[1], matches[2]),
						Description: matches[0],
						Priority:    "medium",
						Specification: Specification{
							Constraints: []Constraint{{
								ID:         "C1",
								Type:       "time_based",
								Expression: fmt.Sprintf("Within %s %s", matches[1], matches[2]),
								Unit:       strings.TrimSuffix(matches[2], "s"),
							}},
						},
					}
				},
			},
			{
				name:  "not_allowed",
				regex: regexp.MustCompile(`(?i)(?:must\s+not|shall\s+not|cannot|should\s+not)\s+(.+?)(?:\.|$)`),
				extract: func(matches []string) *BusinessRule {
					if len(matches) < 2 {
						return nil
					}
					return &BusinessRule{
						ID:          generateRuleID(),
						Version:     "1.0",
						Status:      "draft",
						Title:       "Prohibition: " + truncate(matches[1], 80),
						Description: matches[0],
						Priority:    "high",
						Specification: Specification{
							Constraints: []Constraint{{
								ID:         "C1",
								Type:       "state_based",
								Expression: "NOT: " + matches[1],
							}},
						},
					}
				},
			},
		},
	}
}

// Extract performs regex-based extraction
func (f *regexFallback) Extract(ctx context.Context, text string) (*ExtractResult, error) {
	var rules []BusinessRule
	seen := make(map[string]bool)

	for _, pattern := range f.patterns {
		matches := pattern.regex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			// Avoid duplicates
			key := strings.ToLower(match[0])
			if seen[key] {
				continue
			}
			seen[key] = true

			if rule := pattern.extract(match); rule != nil {
				rule.Confidence = 0.5 // Lower confidence for regex
				rules = append(rules, *rule)
			}
		}
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no business rules found via regex patterns")
	}

	return &ExtractResult{
		BusinessRules: rules,
		Confidence:    0.5,
		Source:        "regex",
		Metadata: ExtractionMetadata{
			ProcessedAt: time.Now(),
		},
	}, nil
}

func generateRuleID() string {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	ruleCounter++
	return fmt.Sprintf("BR-%03d", ruleCounter)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

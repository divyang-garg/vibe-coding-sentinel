// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ResponseParser parses LLM responses into structured data
type ResponseParser interface {
	Parse(response string) (*ExtractResult, error)
}

// responseParser implements ResponseParser
type responseParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() ResponseParser {
	return &responseParser{}
}

// Parse parses LLM response JSON into ExtractResult
func (p *responseParser) Parse(response string) (*ExtractResult, error) {
	cleaned := p.cleanResponse(response)

	var wrapper struct {
		BusinessRules []BusinessRule  `json:"business_rules,omitempty"`
		Entities      []Entity        `json:"entities,omitempty"`
		APIContracts  []APIContract   `json:"api_contracts,omitempty"`
		UserJourneys  []UserJourney   `json:"user_journeys,omitempty"`
		Glossary      []GlossaryTerm  `json:"glossary,omitempty"`
	}

	if err := json.Unmarshal([]byte(cleaned), &wrapper); err != nil {
		repaired := p.repairJSON(cleaned)
		if err := json.Unmarshal([]byte(repaired), &wrapper); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	var errors []ExtractionError
	if len(wrapper.BusinessRules) > 0 {
		errors = append(errors, p.validateRules(wrapper.BusinessRules)...)
	}
	if len(wrapper.Entities) > 0 {
		errors = append(errors, p.validateEntities(wrapper.Entities)...)
	}
	if len(wrapper.APIContracts) > 0 {
		errors = append(errors, p.validateAPIContracts(wrapper.APIContracts)...)
	}

	return &ExtractResult{
		BusinessRules: wrapper.BusinessRules,
		Entities:      wrapper.Entities,
		APIContracts:  wrapper.APIContracts,
		UserJourneys:  wrapper.UserJourneys,
		Glossary:      wrapper.Glossary,
		Errors:        errors,
		Metadata:      ExtractionMetadata{},
	}, nil
}

func (p *responseParser) cleanResponse(response string) string {
	// Remove markdown code fences
	re := regexp.MustCompile("```(?:json)?\\s*")
	cleaned := re.ReplaceAllString(response, "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	return strings.TrimSpace(cleaned)
}

func (p *responseParser) repairJSON(response string) string {
	// Fix common LLM JSON errors
	repaired := response

	// Fix trailing commas before closing brackets
	re := regexp.MustCompile(",\\s*([\\]\\}])")
	repaired = re.ReplaceAllString(repaired, "$1")

	return repaired
}

func (p *responseParser) validateRules(rules []BusinessRule) []ExtractionError {
	var errors []ExtractionError

	for i, rule := range rules {
		if rule.Title == "" {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_TITLE",
				Message: fmt.Sprintf("Rule %d missing required field: title", i),
			})
		}
		if len(rule.Specification.Constraints) == 0 {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_CONSTRAINTS",
				Message: fmt.Sprintf("Rule %s has no constraints", rule.ID),
			})
		}
	}

	return errors
}

func (p *responseParser) validateEntities(entities []Entity) []ExtractionError {
	var errors []ExtractionError
	for i, entity := range entities {
		if entity.Name == "" {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_NAME",
				Message: fmt.Sprintf("Entity %d missing required field: name", i),
			})
		}
		if len(entity.Fields) == 0 {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_FIELDS",
				Message: fmt.Sprintf("Entity %s has no fields", entity.ID),
			})
		}
	}
	return errors
}

func (p *responseParser) validateAPIContracts(contracts []APIContract) []ExtractionError {
	var errors []ExtractionError
	for i, contract := range contracts {
		if contract.Endpoint == "" {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_ENDPOINT",
				Message: fmt.Sprintf("API contract %d missing required field: endpoint", i),
			})
		}
		if contract.Method == "" {
			errors = append(errors, ExtractionError{
				Code:    "MISSING_METHOD",
				Message: fmt.Sprintf("API contract %s missing required field: method", contract.ID),
			})
		}
	}
	return errors
}

// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseParser_Parse(t *testing.T) {
	parser := NewResponseParser()

	t.Run("valid_json", func(t *testing.T) {
		input := `{"business_rules":[{"id":"BR-001","title":"Test","description":"Desc","specification":{"constraints":[{"id":"C1","type":"time_based","expression":"Within 24 hours"}]}}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.BusinessRules, 1)
		assert.Equal(t, "BR-001", result.BusinessRules[0].ID)
	})

	t.Run("json_with_markdown_fences", func(t *testing.T) {
		input := "```json\n{\"business_rules\":[]}\n```"

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.BusinessRules, 0)
	})

	t.Run("trailing_comma_repair", func(t *testing.T) {
		input := `{"business_rules":[{"id":"BR-001","title":"Test",}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.BusinessRules, 1)
	})

	t.Run("invalid_json_returns_error", func(t *testing.T) {
		input := `{invalid json}`

		_, err := parser.Parse(input)

		assert.Error(t, err)
	})

	t.Run("validates_missing_title", func(t *testing.T) {
		input := `{"business_rules":[{"id":"BR-001","specification":{"constraints":[{"id":"C1"}]}}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "MISSING_TITLE", result.Errors[0].Code)
	})

	t.Run("validates_missing_constraints", func(t *testing.T) {
		input := `{"business_rules":[{"id":"BR-001","title":"Test","specification":{"constraints":[]}}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "MISSING_CONSTRAINTS", result.Errors[0].Code)
	})

	t.Run("validates_entities", func(t *testing.T) {
		input := `{"entities":[{"id":"E1"},{"id":"E2","name":"Entity2","fields":[]}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		// Should have validation errors for missing name and fields
		assert.Greater(t, len(result.Errors), 0)
	})

	t.Run("validates_entities_with_name_and_fields", func(t *testing.T) {
		input := `{"entities":[{"id":"E1","name":"User","fields":[{"name":"id","type":"string"}]}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		// Should not have validation errors
		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Entities, 1)
	})

	t.Run("validates_api_contracts", func(t *testing.T) {
		input := `{"api_contracts":[{"id":"API1"},{"id":"API2","endpoint":"/users"}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		// Should have validation errors for missing endpoint and method
		assert.Greater(t, len(result.Errors), 0)
	})

	t.Run("validates_api_contracts_complete", func(t *testing.T) {
		input := `{"api_contracts":[{"id":"API1","endpoint":"/users","method":"GET"}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		// Should not have validation errors
		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.APIContracts, 1)
	})

	t.Run("parses_user_journeys", func(t *testing.T) {
		input := `{"user_journeys":[{"id":"UJ1","title":"Login Flow","steps":[]}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.UserJourneys, 1)
	})

	t.Run("parses_glossary", func(t *testing.T) {
		input := `{"glossary":[{"term":"API","definition":"Application Programming Interface"}]}`

		result, err := parser.Parse(input)

		assert.NoError(t, err)
		assert.Len(t, result.Glossary, 1)
	})
}

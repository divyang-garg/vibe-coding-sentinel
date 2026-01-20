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
}

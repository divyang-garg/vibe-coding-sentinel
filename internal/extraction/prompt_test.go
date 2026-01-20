// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder_BuildBusinessRulesPrompt(t *testing.T) {
	builder := NewPromptBuilder()

	t.Run("prompt_contains_text", func(t *testing.T) {
		text := "The system must validate input."
		prompt := builder.BuildBusinessRulesPrompt(text)

		assert.Contains(t, prompt, text)
		assert.Contains(t, prompt, "business_rules")
		assert.Contains(t, prompt, "OUTPUT FORMAT")
	})

	t.Run("prompt_has_correct_structure", func(t *testing.T) {
		text := "Test document"
		prompt := builder.BuildBusinessRulesPrompt(text)

		assert.Contains(t, prompt, "You are extracting business rules")
		assert.Contains(t, prompt, "DOCUMENT TEXT:")
		assert.Contains(t, prompt, "Return ONLY valid JSON")
	})
}

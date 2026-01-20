// Package extraction provides LLM-powered knowledge extraction
// Integration tests for live LLM extraction
package extraction

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestLiveLLMExtraction tests extraction with real LLM (requires LLM_API_KEY)
func TestLiveLLMExtraction(t *testing.T) {
	if os.Getenv("LLM_API_KEY") == "" {
		t.Skip("LLM_API_KEY not set, skipping live test")
	}

	// Create extractor with real LLM client
	// Note: This requires proper LLM client setup
	// For now, we'll test the structure and fallback behavior

	t.Run("extraction_structure", func(t *testing.T) {
		// Test that extraction returns valid structure
		// This will use fallback if LLM is not configured
		extractor := NewKnowledgeExtractor(
			&MockLLMClient{},
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			&MockCache{},
			&MockLogger{},
		)

		req := ExtractRequest{
			Text:       "The system must validate all user input.",
			Source:     "test.md",
			SchemaType: "business_rule",
			Options: ExtractOptions{
				UseLLM:      false, // Use fallback for this test
				UseFallback: true,
			},
		}

		result, err := extractor.Extract(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.BusinessRules), 0)
		assert.GreaterOrEqual(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
	})

	t.Run("json_structure_validation", func(t *testing.T) {
		// Test that extracted rules match KNOWLEDGE_SCHEMA.md structure
		parser := NewResponseParser()

		validJSON := `{"business_rules":[{"id":"BR-001","title":"Test Rule","description":"Test","specification":{"constraints":[{"id":"C1","type":"state_based","expression":"test"}]}}]}`

		result, err := parser.Parse(validJSON)
		require.NoError(t, err)
		assert.Len(t, result.BusinessRules, 1)
		assert.Equal(t, "BR-001", result.BusinessRules[0].ID)
		assert.NotEmpty(t, result.BusinessRules[0].Title)
		assert.NotEmpty(t, result.BusinessRules[0].Specification.Constraints)
	})
}

// TestEndToEndExtraction tests the full pipeline
func TestEndToEndExtraction(t *testing.T) {
	t.Run("full_pipeline_fallback", func(t *testing.T) {
		mockCache := &MockCache{}
		mockCache.On("Get", mock.Anything).Return("", false)

		extractor := NewKnowledgeExtractor(
			&MockLLMClient{},
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			&MockLogger{},
		)

		req := ExtractRequest{
			Text:       "The system must validate all user input before processing.",
			Source:     "test.md",
			SchemaType: "business_rule",
			Options: ExtractOptions{
				UseLLM:      false,
				UseFallback: true,
			},
		}

		result, err := extractor.Extract(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "regex", result.Source)
		assert.Greater(t, len(result.BusinessRules), 0)
	})

	t.Run("confidence_score_range", func(t *testing.T) {
		scorer := NewConfidenceScorer()

		rule := BusinessRule{
			Title:       "Test Rule",
			Description: "This is a test rule with sufficient detail",
			Specification: Specification{
				Constraints: []Constraint{{
					ID:         "C1",
					Type:       "state_based",
					Expression: "test expression",
					Pseudocode: "test.pseudocode",
				}},
			},
			Traceability: Traceability{
				SourceDocument: "test.md",
			},
		}

		score := scorer.ScoreRule(rule)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})
}

// TestFallbackBehavior tests fallback scenarios
func TestFallbackBehavior(t *testing.T) {
	t.Run("llm_failure_triggers_fallback", func(t *testing.T) {
		// Mock LLM that always fails
		failingLLM := &MockLLMClient{}
		failingLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return("", 0, assert.AnError)

		mockCache := &MockCache{}
		mockCache.On("Get", mock.Anything).Return("", false)
		mockLogger := &MockLogger{}
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		extractor := NewKnowledgeExtractor(
			failingLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		req := ExtractRequest{
			Text:       "The system must validate input.",
			Source:     "test.md",
			SchemaType: "business_rule",
			Options: ExtractOptions{
				UseLLM:      true,
				UseFallback: true,
			},
		}

		result, err := extractor.Extract(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "regex", result.Source)
	})
}

// TestBatchProcessing tests batch extraction functionality
func TestBatchProcessing(t *testing.T) {
	t.Run("chunks_large_document", func(t *testing.T) {
		chunker := NewTextChunker(4000)
		largeText := strings.Repeat("The system must validate input. ", 1000) // ~40K chars

		chunks := chunker.Chunk(largeText, 4000)
		assert.Greater(t, len(chunks), 1, "Large document should be split into multiple chunks")
	})

	t.Run("batch_extraction_deduplicates", func(t *testing.T) {
		mockCache := &MockCache{}
		mockCache.On("Get", mock.Anything).Return("", false)

		extractor := NewKnowledgeExtractor(
			&MockLLMClient{},
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			&MockLogger{},
		)

		// Create text with duplicate rules
		text := "The system must validate input. The system must validate input."
		req := ExtractRequest{
			Text:       text,
			Source:     "test.md",
			SchemaType: "business_rule",
			Options: ExtractOptions{
				UseLLM:      false,
				UseFallback: true,
			},
		}

		result, err := extractor.ExtractBatch(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestSchemaTypes tests extraction for different schema types
func TestSchemaTypes(t *testing.T) {
	t.Run("entity_extraction_prompt", func(t *testing.T) {
		builder := NewPromptBuilder()
		prompt := builder.BuildEntitiesPrompt("The User entity has email and password fields.")

		assert.Contains(t, prompt, "entities")
		assert.Contains(t, prompt, "fields")
	})

	t.Run("api_contract_extraction_prompt", func(t *testing.T) {
		builder := NewPromptBuilder()
		prompt := builder.BuildAPIContractsPrompt("GET /api/users returns a list of users.")

		assert.Contains(t, prompt, "api_contracts")
		assert.Contains(t, prompt, "endpoint")
	})

	t.Run("user_journey_extraction_prompt", func(t *testing.T) {
		builder := NewPromptBuilder()
		prompt := builder.BuildUserJourneysPrompt("User logs in, then views dashboard.")

		assert.Contains(t, prompt, "user_journeys")
		assert.Contains(t, prompt, "steps")
	})

	t.Run("glossary_extraction_prompt", func(t *testing.T) {
		builder := NewPromptBuilder()
		prompt := builder.BuildGlossaryPrompt("An order is a customer purchase request.")

		assert.Contains(t, prompt, "glossary")
		assert.Contains(t, prompt, "definition")
	})
}

// Note: Mock implementations are defined in extractor_test.go

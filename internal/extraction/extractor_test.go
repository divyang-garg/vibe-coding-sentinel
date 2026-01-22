// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package extraction

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient for testing
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
	args := m.Called(ctx, prompt, taskType)
	return args.String(0), args.Int(1), args.Error(2)
}

// MockCache for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (string, bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

func (m *MockCache) Set(key string, value string, tokensUsed int) {
	m.Called(key, value, tokensUsed)
}

// MockLogger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestKnowledgeExtractor_Extract(t *testing.T) {
	t.Run("success_with_llm", func(t *testing.T) {
		// Given
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		llmResponse := `{"business_rules":[{"id":"BR-001","title":"Test Rule","description":"Test description","specification":{"constraints":[{"id":"C1","type":"state_based","expression":"test"}]}}]}`
		mockLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return(llmResponse, 100, nil)
		mockCache.On("Get", mock.Anything).Return("", false)
		mockCache.On("Set", mock.Anything, mock.Anything, 100).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		req := ExtractRequest{
			Text:       "The system must validate user input",
			SchemaType: "business_rule",
			Options:    ExtractOptions{UseLLM: true, UseFallback: true},
		}

		// When
		result, err := extractor.Extract(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "llm", result.Source)
		assert.Len(t, result.BusinessRules, 1)
		mockLLM.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("fallback_to_regex", func(t *testing.T) {
		// Given
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		mockLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return("", 0, fmt.Errorf("LLM unavailable"))
		mockCache.On("Get", mock.Anything).Return("", false)
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		req := ExtractRequest{
			Text:       "The system must validate all user inputs before processing.",
			SchemaType: "business_rule",
			Options:    ExtractOptions{UseLLM: true, UseFallback: true},
		}

		// When
		result, err := extractor.Extract(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "regex", result.Source)
		assert.Greater(t, len(result.BusinessRules), 0)
	})

	t.Run("validation_error_empty_text", func(t *testing.T) {
		// Given
		extractor := NewKnowledgeExtractor(
			new(MockLLMClient),
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			new(MockCache),
			new(MockLogger),
		)

		req := ExtractRequest{
			Text:       "",
			SchemaType: "business_rule",
			Options:    ExtractOptions{UseLLM: true},
		}

		// When
		_, err := extractor.Extract(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "text is required")
	})

	t.Run("cache_hit", func(t *testing.T) {
		// Given
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		cachedResponse := `{"business_rules":[{"id":"BR-001","title":"Cached Rule","description":"Cached","specification":{"constraints":[{"id":"C1","type":"state_based","expression":"test"}]}}]}`

		extractor := NewKnowledgeExtractor(
			new(MockLLMClient),
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		mockCache.On("Get", mock.Anything).Return(cachedResponse, true)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		req := ExtractRequest{
			Text:       "Test text",
			SchemaType: "business_rule",
			Options:    ExtractOptions{UseLLM: true},
		}

		// When
		result, err := extractor.Extract(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Metadata.CacheHit)
		mockCache.AssertExpectations(t)
	})

	t.Run("extract_with_entity_schema", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		llmResponse := `{"entities":[{"name":"User","attributes":["id","email"]}]}`
		mockLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return(llmResponse, 100, nil)
		mockCache.On("Get", mock.Anything).Return("", false)
		mockCache.On("Set", mock.Anything, mock.Anything, 100).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		req := ExtractRequest{
			Text:       "User entity with id and email",
			SchemaType: "entity",
			Options:    ExtractOptions{UseLLM: true},
		}

		result, err := extractor.Extract(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0.7, result.Confidence) // Default for non-business-rule
	})

	t.Run("extract_with_api_contract_schema", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		llmResponse := `{"api_contracts":[{"endpoint":"/api/users","method":"GET"}]}`
		mockLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return(llmResponse, 100, nil)
		mockCache.On("Get", mock.Anything).Return("", false)
		mockCache.On("Set", mock.Anything, mock.Anything, 100).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		req := ExtractRequest{
			Text:       "GET /api/users endpoint",
			SchemaType: "api_contract",
			Options:    ExtractOptions{UseLLM: true},
		}

		result, err := extractor.Extract(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("extract_with_unsupported_schema", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		mockCache.On("Get", mock.Anything).Return("", false)
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return().Maybe()

		req := ExtractRequest{
			Text:       "Test text",
			SchemaType: "unsupported_type",
			Options:    ExtractOptions{UseLLM: true, UseFallback: false},
		}

		_, err := extractor.Extract(context.Background(), req)
		// The error may be from extractWithLLM or from fallback, both are acceptable
		assert.Error(t, err)
		// Check if error contains expected messages
		errMsg := err.Error()
		if !strings.Contains(errMsg, "unsupported schema type") && 
		   !strings.Contains(errMsg, "all extraction methods failed") &&
		   !strings.Contains(errMsg, "extraction disabled") {
			t.Errorf("Expected error about unsupported schema or extraction failure, got: %v", err)
		}
	})

	t.Run("extract_with_parse_error", func(t *testing.T) {
		mockLLM := new(MockLLMClient)
		mockCache := new(MockCache)
		mockLogger := new(MockLogger)

		extractor := NewKnowledgeExtractor(
			mockLLM,
			NewPromptBuilder(),
			NewResponseParser(),
			NewConfidenceScorer(),
			NewFallbackExtractor(),
			mockCache,
			mockLogger,
		)

		mockLLM.On("Call", mock.Anything, mock.Anything, "knowledge_extraction").
			Return("invalid json", 100, nil)
		mockCache.On("Get", mock.Anything).Return("", false)
		mockCache.On("Set", mock.Anything, mock.Anything, 100).Return()
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()

		req := ExtractRequest{
			Text:       "The system must validate all user inputs before processing.",
			SchemaType: "business_rule",
			Options:    ExtractOptions{UseLLM: true, UseFallback: true},
		}

		// Should fallback to regex
		result, err := extractor.Extract(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "regex", result.Source)
	})
}

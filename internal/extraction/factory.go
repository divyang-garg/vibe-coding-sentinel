// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import "time"

// ExtractorFactory creates configured extractors
type ExtractorFactory struct{}

// NewExtractorFactory creates a new factory
func NewExtractorFactory() *ExtractorFactory {
	return &ExtractorFactory{}
}

// CreateDefault creates an extractor with default configuration
func (f *ExtractorFactory) CreateDefault() *KnowledgeExtractor {
	cfg := DefaultOllamaConfig()
	llmClient := NewOllamaClient(cfg)
	promptBuilder := NewPromptBuilder()
	parser := NewResponseParser()
	scorer := NewConfidenceScorer()
	fallback := NewFallbackExtractor()
	cache := NewMemoryCache(1000, 24*time.Hour)
	logger := NewStdLogger()

	return NewKnowledgeExtractor(
		llmClient,
		promptBuilder,
		parser,
		scorer,
		fallback,
		cache,
		logger,
	)
}

// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package extraction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// LLMClient interface for LLM calls (wraps hub/api/llm)
type LLMClient interface {
	Call(ctx context.Context, prompt string, taskType string) (string, int, error)
}

// Cache interface for extraction caching
type Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string, tokensUsed int)
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Extractor defines the knowledge extraction interface
type Extractor interface {
	Extract(ctx context.Context, req ExtractRequest) (*ExtractResult, error)
	ExtractBatch(ctx context.Context, req ExtractRequest) (*ExtractResult, error)
}

// KnowledgeExtractor implements LLM-based extraction with fallback
type KnowledgeExtractor struct {
	llmClient     LLMClient
	promptBuilder PromptBuilder
	parser        ResponseParser
	scorer        ConfidenceScorer
	fallback      FallbackExtractor
	cache         Cache
	logger        Logger
}

// NewKnowledgeExtractor creates a new extractor with dependencies
// Complies with CODING_STANDARDS.md Section 7.1 Constructor Injection
func NewKnowledgeExtractor(
	llmClient LLMClient,
	promptBuilder PromptBuilder,
	parser ResponseParser,
	scorer ConfidenceScorer,
	fallback FallbackExtractor,
	cache Cache,
	logger Logger,
) *KnowledgeExtractor {
	return &KnowledgeExtractor{
		llmClient:     llmClient,
		promptBuilder: promptBuilder,
		parser:        parser,
		scorer:        scorer,
		fallback:      fallback,
		cache:         cache,
		logger:        logger,
	}
}

// Extract performs knowledge extraction from text
func (e *KnowledgeExtractor) Extract(ctx context.Context, req ExtractRequest) (*ExtractResult, error) {
	startTime := time.Now()

	// Validate request (CODING_STANDARDS Section 11.1)
	if err := e.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate cache key
	cacheKey := e.generateCacheKey(req.Text)

	// Check cache first
	if cached, ok := e.cache.Get(cacheKey); ok {
		e.logger.Debug("cache hit for extraction")
		result, err := e.parser.Parse(cached)
		if err == nil {
			result.Metadata.CacheHit = true
			result.Metadata.ProcessedAt = time.Now()
			result.Metadata.ProcessingMs = time.Since(startTime).Milliseconds()
			return result, nil
		}
	}

	// Try LLM extraction
	if req.Options.UseLLM {
		result, err := e.extractWithLLM(ctx, req)
		if err == nil {
			result.Metadata.ProcessedAt = time.Now()
			result.Metadata.ProcessingMs = time.Since(startTime).Milliseconds()
			return result, nil
		}
		e.logger.Warn("LLM extraction failed, trying fallback", "error", err)
	}

	// Fallback to regex
	if req.Options.UseFallback {
		result, err := e.fallback.Extract(ctx, req.Text)
		if err == nil {
			result.Source = "regex"
			result.Metadata.ProcessedAt = time.Now()
			result.Metadata.ProcessingMs = time.Since(startTime).Milliseconds()
			return result, nil
		}
		return nil, fmt.Errorf("all extraction methods failed: %w", err)
	}

	return nil, fmt.Errorf("extraction disabled and fallback not allowed")
}

func (e *KnowledgeExtractor) validateRequest(req ExtractRequest) error {
	if req.Text == "" {
		return &ValidationError{Field: "text", Message: "text is required"}
	}
	if len(req.Text) > 100000 { // ~25K tokens max
		return &ValidationError{Field: "text", Message: "text exceeds maximum length"}
	}
	return nil
}

func (e *KnowledgeExtractor) extractWithLLM(ctx context.Context, req ExtractRequest) (*ExtractResult, error) {
	// Build prompt based on schema type
	var prompt string
	switch req.SchemaType {
	case "business_rule", "":
		prompt = e.promptBuilder.BuildBusinessRulesPrompt(req.Text)
	case "entity":
		prompt = e.promptBuilder.BuildEntitiesPrompt(req.Text)
	case "api_contract":
		prompt = e.promptBuilder.BuildAPIContractsPrompt(req.Text)
	case "user_journey":
		prompt = e.promptBuilder.BuildUserJourneysPrompt(req.Text)
	case "glossary":
		prompt = e.promptBuilder.BuildGlossaryPrompt(req.Text)
	default:
		return nil, fmt.Errorf("unsupported schema type: %s", req.SchemaType)
	}

	// Call LLM with retry logic
	response, tokens, err := e.callLLMWithRetry(ctx, prompt, "knowledge_extraction", 3)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse response
	result, err := e.parser.Parse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Calculate confidence scores (for business rules only, other types use default)
	if len(result.BusinessRules) > 0 {
		for i := range result.BusinessRules {
			result.BusinessRules[i].Confidence = e.scorer.ScoreRule(result.BusinessRules[i])
		}
		result.Confidence = e.scorer.ScoreOverall(result.BusinessRules)
	} else {
		// Default confidence for non-business-rule schemas
		result.Confidence = 0.7
	}
	result.Source = "llm"
	result.Metadata.TokensUsed = tokens

	// Cache successful result
	cacheKey := e.generateCacheKey(req.Text)
	e.cache.Set(cacheKey, response, tokens)

	return result, nil
}

func (e *KnowledgeExtractor) generateCacheKey(text string) string {
	input := fmt.Sprintf("extract:business_rules:v1:%s", text)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// ValidationError for input validation failures
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for '%s': %s", e.Field, e.Message)
}

// ExtractBatch processes large documents by splitting into chunks
func (e *KnowledgeExtractor) ExtractBatch(ctx context.Context, req ExtractRequest) (*ExtractResult, error) {
	chunker := NewTextChunker(4000) // ~1000 tokens per chunk
	chunks := chunker.Chunk(req.Text, 4000)

	var allRules []BusinessRule
	var allErrors []ExtractionError

	for i, chunk := range chunks {
		chunkReq := ExtractRequest{
			Text:       chunk,
			Source:     fmt.Sprintf("%s:chunk-%d", req.Source, i),
			SchemaType: req.SchemaType,
			Options:    req.Options,
		}

		result, err := e.Extract(ctx, chunkReq)
		if err != nil {
			allErrors = append(allErrors, ExtractionError{
				Code:    "CHUNK_EXTRACTION_FAILED",
				Message: fmt.Sprintf("chunk %d failed: %v", i, err),
			})
			continue
		}

		allRules = append(allRules, result.BusinessRules...)
	}

	// Deduplicate rules by ID
	deduplicated := deduplicateRules(allRules)

	// Calculate overall confidence
	confidence := e.scorer.ScoreOverall(deduplicated)
	if len(deduplicated) == 0 {
		confidence = 0.0
	}

	return &ExtractResult{
		BusinessRules: deduplicated,
		Confidence:    confidence,
		Source:        "llm",
		Errors:        allErrors,
		Metadata: ExtractionMetadata{
			ProcessedAt: time.Now(),
		},
	}, nil
}

// deduplicateRules removes duplicate rules based on ID
func deduplicateRules(rules []BusinessRule) []BusinessRule {
	seen := make(map[string]bool)
	var unique []BusinessRule

	for _, rule := range rules {
		if rule.ID == "" {
			// Generate ID if missing
			rule.ID = fmt.Sprintf("BR-%d", len(unique)+1)
		}
		if !seen[rule.ID] {
			seen[rule.ID] = true
			unique = append(unique, rule)
		}
	}

	return unique
}

// callLLMWithRetry calls LLM with exponential backoff retry
func (e *KnowledgeExtractor) callLLMWithRetry(ctx context.Context, prompt string, taskType string, maxRetries int) (string, int, error) {
	backoff := time.Second
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		response, tokens, err := e.llmClient.Call(ctx, prompt, taskType)
		if err == nil {
			return response, tokens, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return "", 0, err
		}

		// Wait before retry (except on last attempt)
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return "", 0, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			}
		}
	}

	return "", 0, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Retry on network errors, rate limits, timeouts
	retryablePatterns := []string{
		"timeout", "rate limit", "temporary", "network", "connection",
		"503", "502", "429", // HTTP status codes
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

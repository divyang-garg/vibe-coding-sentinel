// Package extraction provides tests for extractor factory
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package extraction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewExtractorFactory(t *testing.T) {
	t.Run("creates factory instance", func(t *testing.T) {
		// When
		factory := NewExtractorFactory()

		// Then
		assert.NotNil(t, factory)
	})
}

func TestExtractorFactory_CreateDefault(t *testing.T) {
	t.Run("creates extractor with default configuration", func(t *testing.T) {
		// Given
		factory := NewExtractorFactory()

		// When
		extractor := factory.CreateDefault()

		// Then
		assert.NotNil(t, extractor)
		// Verify extractor has all required components
		assert.NotNil(t, extractor.llmClient)
		assert.NotNil(t, extractor.promptBuilder)
		assert.NotNil(t, extractor.parser)
		assert.NotNil(t, extractor.scorer)
		assert.NotNil(t, extractor.fallback)
		assert.NotNil(t, extractor.cache)
		assert.NotNil(t, extractor.logger)
	})

	t.Run("creates functional extractor", func(t *testing.T) {
		// Given
		factory := NewExtractorFactory()
		extractor := factory.CreateDefault()

		// When - extractor should be ready to use
		// Verify it's not nil and has all dependencies
		assert.NotNil(t, extractor)

		// Verify cache is properly initialized
		cache := extractor.cache
		if memCache, ok := cache.(*MemoryCache); ok {
			assert.Equal(t, 1000, memCache.maxSize)
			assert.Equal(t, 24*time.Hour, memCache.ttl)
		}
	})
}

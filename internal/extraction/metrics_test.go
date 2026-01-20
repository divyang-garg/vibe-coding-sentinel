// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector(t *testing.T) {
	t.Run("records extraction metrics", func(t *testing.T) {
		mc := NewMetricsCollector()

		mc.RecordExtraction(100*time.Millisecond, true, "llm")
		mc.RecordExtraction(50*time.Millisecond, true, "regex")
		mc.RecordExtraction(200*time.Millisecond, false, "llm")

		metrics := mc.GetMetrics()
		assert.Equal(t, int64(3), metrics.TotalExtractions)
		assert.Equal(t, int64(2), metrics.SuccessfulExtractions)
		assert.Equal(t, int64(1), metrics.FailedExtractions)
	})

	t.Run("records token usage", func(t *testing.T) {
		mc := NewMetricsCollector()

		mc.RecordTokenUsage(100, "openai")
		mc.RecordTokenUsage(200, "openai")
		mc.RecordTokenUsage(50, "anthropic")

		metrics := mc.GetMetrics()
		assert.Equal(t, int64(350), metrics.TotalTokens)
	})

	t.Run("records cache hits and misses", func(t *testing.T) {
		mc := NewMetricsCollector()

		mc.RecordCacheHit(true)
		mc.RecordCacheHit(true)
		mc.RecordCacheHit(false)

		metrics := mc.GetMetrics()
		assert.Equal(t, int64(2), metrics.CacheHits)
		assert.Equal(t, int64(1), metrics.CacheMisses)
	})

	t.Run("calculates average duration", func(t *testing.T) {
		mc := NewMetricsCollector()

		mc.RecordExtraction(100*time.Millisecond, true, "llm")
		mc.RecordExtraction(200*time.Millisecond, true, "llm")

		metrics := mc.GetMetrics()
		assert.Equal(t, 150*time.Millisecond, metrics.AverageDuration)
	})

	t.Run("handles zero extractions", func(t *testing.T) {
		mc := NewMetricsCollector()

		metrics := mc.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalExtractions)
		assert.Equal(t, time.Duration(0), metrics.AverageDuration)
	})
}

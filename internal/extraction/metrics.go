// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"sync"
	"time"
)

// MetricsCollector tracks extraction metrics
type MetricsCollector interface {
	RecordExtraction(duration time.Duration, success bool, source string)
	RecordTokenUsage(tokens int, provider string)
	RecordCacheHit(hit bool)
	GetMetrics() ExtractionMetrics
}

// ExtractionMetrics contains aggregated metrics
type ExtractionMetrics struct {
	TotalExtractions      int64
	SuccessfulExtractions int64
	FailedExtractions     int64
	TotalTokens           int64
	CacheHits             int64
	CacheMisses           int64
	AverageDuration       time.Duration
}

// simpleMetricsCollector implements MetricsCollector
type simpleMetricsCollector struct {
	mu                    sync.RWMutex
	totalExtractions      int64
	successfulExtractions int64
	failedExtractions     int64
	totalTokens           int64
	cacheHits             int64
	cacheMisses           int64
	totalDuration         time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() MetricsCollector {
	return &simpleMetricsCollector{}
}

func (m *simpleMetricsCollector) RecordExtraction(duration time.Duration, success bool, source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalExtractions++
	if success {
		m.successfulExtractions++
	} else {
		m.failedExtractions++
	}
	m.totalDuration += duration
}

func (m *simpleMetricsCollector) RecordTokenUsage(tokens int, provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalTokens += int64(tokens)
}

func (m *simpleMetricsCollector) RecordCacheHit(hit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if hit {
		m.cacheHits++
	} else {
		m.cacheMisses++
	}
}

func (m *simpleMetricsCollector) GetMetrics() ExtractionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := time.Duration(0)
	if m.totalExtractions > 0 {
		avgDuration = m.totalDuration / time.Duration(m.totalExtractions)
	}

	return ExtractionMetrics{
		TotalExtractions:      m.totalExtractions,
		SuccessfulExtractions: m.successfulExtractions,
		FailedExtractions:     m.failedExtractions,
		TotalTokens:           m.totalTokens,
		CacheHits:             m.cacheHits,
		CacheMisses:           m.cacheMisses,
		AverageDuration:       avgDuration,
	}
}

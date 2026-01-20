// Package metrics provides Prometheus metrics
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec
	
	// Business metrics
	TasksCreated        prometheus.Counter
	TasksCompleted      prometheus.Counter
	DocumentsProcessed  prometheus.Counter
	ExtractionDuration  *prometheus.HistogramVec
	ExtractionConfidence *prometheus.HistogramVec
	
	// System metrics
	ActiveConnections   prometheus.Gauge
	GoroutineCount      prometheus.Gauge
	MemoryUsage         prometheus.Gauge
}

// NewMetrics creates and registers all metrics
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "sentinel_hub_api"
	}
	
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
			},
			[]string{"method", "path", "status"},
		),
		TasksCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_created_total",
				Help:      "Total number of tasks created",
			},
		),
		TasksCompleted: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_completed_total",
				Help:      "Total number of tasks completed",
			},
		),
		DocumentsProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "documents_processed_total",
				Help:      "Total number of documents processed",
			},
		),
		ExtractionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "extraction_duration_seconds",
				Help:      "Knowledge extraction duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"type", "source"},
		),
		ExtractionConfidence: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "extraction_confidence",
				Help:      "Knowledge extraction confidence scores",
				Buckets:   []float64{.1, .2, .3, .4, .5, .6, .7, .8, .9, 1.0},
			},
			[]string{"type"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_connections",
				Help:      "Number of active HTTP connections",
			},
		),
		GoroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutine_count",
				Help:      "Number of goroutines",
			},
		),
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),
	}
}

// Package services provides unit tests for monitoring service.
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"context"
	"testing"
	"time"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringServiceImpl_ReportError(t *testing.T) {
	service := NewMonitoringService()

	report := models.ErrorReport{
		Message:    "Test error occurred",
		Severity:   models.ErrorSeverityHigh,
		Category:   "test",
		UserID:     "user-123",
		RequestID:  "req-456",
		StackTrace: "stack trace here",
		Context: map[string]interface{}{
			"key": "value",
		},
		Timestamp: time.Now(),
	}

	err := service.ReportError(context.Background(), report)
	assert.NoError(t, err)

	// Verify error was stored (accessing private field through GetErrorDashboard)
	dashboard, err := service.GetErrorDashboard(context.Background())
	assert.NoError(t, err)

	dashboardMap, ok := dashboard.(map[string]interface{})
	assert.True(t, ok)

	totalErrors, ok := dashboardMap["total_errors"].(int)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, totalErrors, 1)
}

func TestMonitoringServiceImpl_GetErrorDashboard(t *testing.T) {
	service := NewMonitoringService()

	// Add some test errors
	errors := []models.ErrorReport{
		{
			Message:   "Database connection failed",
			Severity:  models.ErrorSeverityHigh,
			Category:  "database",
			Timestamp: time.Now(),
		},
		{
			Message:   "API rate limit exceeded",
			Severity:  models.ErrorSeverityMedium,
			Category:  "api",
			Timestamp: time.Now(),
		},
		{
			Message:   "Invalid input validation",
			Severity:  models.ErrorSeverityLow,
			Category:  "validation",
			Timestamp: time.Now(),
		},
	}

	for _, err := range errors {
		assert.NoError(t, service.ReportError(context.Background(), err))
	}

	result, err := service.GetErrorDashboard(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate dashboard structure
	assert.Contains(t, resultMap, "total_errors")
	assert.Contains(t, resultMap, "error_rate")
	assert.Contains(t, resultMap, "top_errors")
	assert.Contains(t, resultMap, "error_trends")
	assert.Contains(t, resultMap, "system_health")
	assert.Contains(t, resultMap, "last_updated")
	assert.Contains(t, resultMap, "active_alerts")

	totalErrors, ok := resultMap["total_errors"].(int)
	assert.True(t, ok)
	assert.Equal(t, 3, totalErrors)

	systemHealth, ok := resultMap["system_health"].(string)
	assert.True(t, ok)
	assert.Contains(t, []string{"healthy", "warning", "critical"}, systemHealth)

	topErrors, ok := resultMap["top_errors"].([]map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, topErrors)

	// Each top error should have required fields
	for _, error := range topErrors {
		assert.Contains(t, error, "message")
		assert.Contains(t, error, "count")
		assert.Contains(t, error, "percentage")
	}
}

func TestMonitoringServiceImpl_GetErrorAnalysis(t *testing.T) {
	service := NewMonitoringService()

	// Add test errors
	assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
		Message:   "Connection timeout",
		Severity:  models.ErrorSeverityHigh,
		Category:  "database",
		Timestamp: time.Now(),
	}))
	assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
		Message:   "Invalid request",
		Severity:  models.ErrorSeverityMedium,
		Category:  "api",
		Timestamp: time.Now(),
	}))

	result, err := service.GetErrorAnalysis(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate analysis structure
	assert.Contains(t, resultMap, "patterns")
	assert.Contains(t, resultMap, "recommendations")
	assert.Contains(t, resultMap, "severity_trends")
	assert.Contains(t, resultMap, "error_clusters")
	assert.Contains(t, resultMap, "time_window")
	assert.Contains(t, resultMap, "analysis_timestamp")

	patterns, ok := resultMap["patterns"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, patterns)

	recommendations, ok := resultMap["recommendations"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, recommendations)

	severityTrends, ok := resultMap["severity_trends"].(map[string]int)
	assert.True(t, ok)
	assert.Contains(t, severityTrends, "low")
	assert.Contains(t, severityTrends, "medium")
	assert.Contains(t, severityTrends, "high")
	assert.Contains(t, severityTrends, "critical")
}

func TestMonitoringServiceImpl_GetErrorStats(t *testing.T) {
	service := NewMonitoringService()

	// Add test errors
	assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
		Message:  "DB error 1",
		Category: "database",
		Severity: models.ErrorSeverityHigh,
	}))
	assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
		Message:  "DB error 2",
		Category: "database",
		Severity: models.ErrorSeverityMedium,
	}))
	assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
		Message:  "API error",
		Category: "api",
		Severity: models.ErrorSeverityLow,
	}))

	result, err := service.GetErrorStats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate stats structure
	assert.Contains(t, resultMap, "total_count")
	assert.Contains(t, resultMap, "by_category")
	assert.Contains(t, resultMap, "by_severity")
	assert.Contains(t, resultMap, "resolution_rate")
	assert.Contains(t, resultMap, "avg_resolution_time")
	assert.Contains(t, resultMap, "trending_categories")
	assert.Contains(t, resultMap, "generated_at")

	totalCount, ok := resultMap["total_count"].(int)
	assert.True(t, ok)
	assert.Equal(t, 3, totalCount)

	byCategory, ok := resultMap["by_category"].(map[string]int)
	assert.True(t, ok)
	assert.Contains(t, byCategory, "database")
	assert.Contains(t, byCategory, "api")
	assert.Greater(t, byCategory["database"], 0)

	bySeverity, ok := resultMap["by_severity"].(map[string]int)
	assert.True(t, ok)
	assert.Contains(t, bySeverity, "low")
	assert.Contains(t, bySeverity, "medium")
	assert.Contains(t, bySeverity, "high")
}

func TestMonitoringServiceImpl_ClassifyError(t *testing.T) {
	tests := []struct {
		name           string
		classification models.ErrorClassification
		wantErr        bool
	}{
		{
			name: "valid error classification",
			classification: models.ErrorClassification{
				Category:    "database",
				Severity:    models.ErrorSeverityHigh,
				Recovery:    "retry",
				Retryable:   true,
				UserVisible: true,
				Context:     map[string]interface{}{"service": "db"},
				Suggestions: []string{"Check database connection", "Increase timeout"},
				ErrorCode:   500,
				Timestamp:   time.Now(),
			},
			wantErr: false,
		},
		{
			name:           "empty classification",
			classification: models.ErrorClassification{},
			wantErr:        false, // Should handle empty gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewMonitoringService()

			result, err := service.ClassifyError(context.Background(), tt.classification)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "error_category")
				assert.Contains(t, resultMap, "severity_assessment")
				assert.Contains(t, resultMap, "impact_analysis")
				assert.Contains(t, resultMap, "recommended_actions")
				assert.Contains(t, resultMap, "similar_errors")
				assert.Contains(t, resultMap, "classification_confidence")

				confidence, ok := resultMap["classification_confidence"].(float64)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, confidence, 0.0)
				assert.LessOrEqual(t, confidence, 100.0)
			}
		})
	}
}

func TestMonitoringServiceImpl_GetHealthMetrics(t *testing.T) {
	service := NewMonitoringService()

	result, err := service.GetHealthMetrics(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate health metrics structure
	assert.Contains(t, resultMap, "system_status")
	assert.Contains(t, resultMap, "uptime_seconds")
	assert.Contains(t, resultMap, "cpu_usage_percent")
	assert.Contains(t, resultMap, "memory_usage_percent")
	assert.Contains(t, resultMap, "disk_usage_percent")
	assert.Contains(t, resultMap, "active_connections")
	assert.Contains(t, resultMap, "request_rate_per_second")
	assert.Contains(t, resultMap, "error_rate_percent")
	assert.Contains(t, resultMap, "response_time_avg_ms")
	assert.Contains(t, resultMap, "database_connections")
	assert.Contains(t, resultMap, "cache_hit_rate_percent")
	assert.Contains(t, resultMap, "last_updated")

	systemStatus, ok := resultMap["system_status"].(string)
	assert.True(t, ok)
	assert.Equal(t, "healthy", systemStatus)

	uptime, ok := resultMap["uptime_seconds"].(int)
	assert.True(t, ok)
	assert.Greater(t, uptime, 0)

	cpuUsage, ok := resultMap["cpu_usage_percent"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, cpuUsage, 0.0)
	assert.LessOrEqual(t, cpuUsage, 100.0)

	memoryUsage, ok := resultMap["memory_usage_percent"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, memoryUsage, 0.0)
	assert.LessOrEqual(t, memoryUsage, 100.0)
}

func TestMonitoringServiceImpl_GetPerformanceMetrics(t *testing.T) {
	service := NewMonitoringService()

	result, err := service.GetPerformanceMetrics(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate performance metrics structure
	assert.Contains(t, resultMap, "response_times")
	assert.Contains(t, resultMap, "throughput")
	assert.Contains(t, resultMap, "resource_usage")
	assert.Contains(t, resultMap, "bottlenecks")
	assert.Contains(t, resultMap, "optimization_opportunities")
	assert.Contains(t, resultMap, "performance_score")
	assert.Contains(t, resultMap, "measured_at")

	responseTimes, ok := resultMap["response_times"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, responseTimes, "p50_ms")
	assert.Contains(t, responseTimes, "p95_ms")
	assert.Contains(t, responseTimes, "p99_ms")

	throughput, ok := resultMap["throughput"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, throughput, "requests_per_second")
	assert.Contains(t, throughput, "bytes_per_second")

	resourceUsage, ok := resultMap["resource_usage"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, resourceUsage, "cpu_cores_used")
	assert.Contains(t, resourceUsage, "memory_mb_used")

	bottlenecks, ok := resultMap["bottlenecks"].([]map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, bottlenecks)

	performanceScore, ok := resultMap["performance_score"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, performanceScore, 0.0)
	assert.LessOrEqual(t, performanceScore, 100.0)
}

func TestMonitoringServiceImpl_ErrorRateCalculation(t *testing.T) {
	service := NewMonitoringService()

	// Initially should have low/zero error rate
	dashboard, err := service.GetErrorDashboard(context.Background())
	assert.NoError(t, err)

	dashboardMap, ok := dashboard.(map[string]interface{})
	assert.True(t, ok)

	initialRate, ok := dashboardMap["error_rate"].(float64)
	assert.True(t, ok)

	// Add some errors
	for i := 0; i < 5; i++ {
		assert.NoError(t, service.ReportError(context.Background(), models.ErrorReport{
			Message:   "Test error",
			Severity:  models.ErrorSeverityMedium,
			Category:  "test",
			Timestamp: time.Now(),
		}))
	}

	// Check error rate increased
	dashboard, err = service.GetErrorDashboard(context.Background())
	assert.NoError(t, err)

	dashboardMap, ok = dashboard.(map[string]interface{})
	assert.True(t, ok)

	newRate, ok := dashboardMap["error_rate"].(float64)
	assert.True(t, ok)

	// Error rate should be higher after adding errors
	assert.GreaterOrEqual(t, newRate, initialRate)
	assert.LessOrEqual(t, newRate, 100.0)
}

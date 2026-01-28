// Package services provides tests for quality metrics and enhanced vibe analysis
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sentinel-hub-api/models"
)

func TestCodeAnalysisServiceImpl_calculateQualityMetrics(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := `
package main

func example() {
	// Example function
	return
}
`
		language := "go"
		issues := []interface{}{
			map[string]interface{}{
				"type":     "style",
				"severity": "low",
				"line":     5,
				"message":  "Minor style issue",
			},
		}
		duplicates := []interface{}{}
		orphaned := []interface{}{}

		// When
		metrics := service.calculateQualityMetrics(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, metrics)
		assert.Greater(t, metrics.OverallScore, 0.0)
		assert.Greater(t, metrics.Maintainability, 0.0)
		assert.Greater(t, metrics.Readability, 0.0)
		assert.Equal(t, 1, metrics.IssueCount)
		assert.NotNil(t, metrics.SeverityBreakdown)
		assert.NotNil(t, metrics.CategoryScores)
	})

	t.Run("with_duplicates_and_orphaned", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := "package main\nfunc test() {}"
		language := "go"
		issues := []interface{}{}
		duplicates := []interface{}{
			map[string]interface{}{"functions": []string{"test"}},
		}
		orphaned := []interface{}{
			map[string]interface{}{"type": "unused_function", "name": "unused"},
		}

		// When
		metrics := service.calculateQualityMetrics(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, metrics)
		assert.Equal(t, 2, metrics.IssueCount) // duplicates + orphaned
	})

	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When
		metrics := service.calculateQualityMetrics(ctx, "code", "go", nil, nil, nil)

		// Then
		assert.Equal(t, QualityMetrics{}, metrics)
	})
}

func TestCodeAnalysisServiceImpl_calculateMaintainabilityIndex(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := `
package main

func simple() {
	return
}
`
		language := "go"

		// When
		index := service.calculateMaintainabilityIndex(ctx, code, language)

		// Then
		assert.Greater(t, index, 0.0)
		assert.LessOrEqual(t, index, 100.0)
	})

	t.Run("empty_code", func(t *testing.T) {
		// Given
		ctx := context.Background()

		// When
		index := service.calculateMaintainabilityIndex(ctx, "", "go")

		// Then
		assert.Equal(t, 0.0, index)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When
		index := service.calculateMaintainabilityIndex(ctx, "code", "go")

		// Then
		assert.Equal(t, 0.0, index)
	})
}

func TestCodeAnalysisServiceImpl_estimateTechnicalDebt(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := "package main\nfunc test() {}"
		language := "go"
		issues := []interface{}{
			map[string]interface{}{
				"type":     "bug",
				"severity": "high",
				"line":     2,
				"message":  "High priority bug",
			},
		}
		duplicates := []interface{}{
			map[string]interface{}{"functions": []string{"test"}},
		}
		orphaned := []interface{}{}

		// When
		debt := service.estimateTechnicalDebt(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, debt)
		assert.Greater(t, debt.TotalDebtHours, 0.0)
		assert.NotNil(t, debt.DebtByCategory)
		assert.NotEmpty(t, debt.PriorityIssues)
		assert.NotEmpty(t, debt.EstimatedCost)
		assert.NotEmpty(t, debt.PayoffTime)
		assert.GreaterOrEqual(t, debt.DebtRatio, 0.0)
		assert.LessOrEqual(t, debt.DebtRatio, 1.0)
	})

	t.Run("with_multiple_issues", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := strings.Repeat("package main\nfunc test() {}\n", 10)
		language := "go"
		issues := []interface{}{
			map[string]interface{}{
				"type":     "critical",
				"severity": "critical",
				"line":     1,
				"message":  "Critical issue",
			},
			map[string]interface{}{
				"type":     "style",
				"severity": "low",
				"line":     2,
				"message":  "Style issue",
			},
		}
		duplicates := []interface{}{}
		orphaned := []interface{}{}

		// When
		debt := service.estimateTechnicalDebt(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, debt)
		assert.Greater(t, debt.TotalDebtHours, 0.0)
		assert.Equal(t, 2, len(debt.PriorityIssues))
	})

	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When
		debt := service.estimateTechnicalDebt(ctx, "code", "go", nil, nil, nil)

		// Then
		assert.Equal(t, TechnicalDebtEstimate{}, debt)
	})
}

func TestCodeAnalysisServiceImpl_calculateRefactoringPriority(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := "package main\nfunc test() {}"
		language := "go"
		issues := []interface{}{
			map[string]interface{}{
				"type":     "bug",
				"severity": "critical",
				"line":     2,
				"message":  "Critical bug",
			},
		}
		duplicates := []interface{}{
			map[string]interface{}{"functions": []string{"test"}},
		}
		orphaned := []interface{}{}

		// When
		priorities := service.calculateRefactoringPriority(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, priorities)
		assert.Greater(t, len(priorities), 0)
		// Should be sorted by score (descending)
		if len(priorities) > 1 {
			assert.GreaterOrEqual(t, priorities[0].Score, priorities[1].Score)
		}
	})

	t.Run("with_orphaned_code", func(t *testing.T) {
		// Given
		ctx := context.Background()
		code := "package main"
		language := "go"
		issues := []interface{}{}
		duplicates := []interface{}{}
		orphaned := []interface{}{
			map[string]interface{}{"type": "unused_function", "name": "unused"},
		}

		// When
		priorities := service.calculateRefactoringPriority(ctx, code, language, issues, duplicates, orphaned)

		// Then
		assert.NotNil(t, priorities)
		// Should include orphaned code priority
		foundOrphaned := false
		for _, p := range priorities {
			if p.Type == "orphaned_code" {
				foundOrphaned = true
				break
			}
		}
		assert.True(t, foundOrphaned)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When
		priorities := service.calculateRefactoringPriority(ctx, "code", "go", nil, nil, nil)

		// Then
		assert.Empty(t, priorities)
	})
}

func TestCodeAnalysisServiceImpl_AnalyzeVibe(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("success", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := models.CodeAnalysisRequest{
			Code: `
package main

func example() {
	return
}
`,
			Language: "go",
		}

		// When
		result, err := service.AnalyzeVibe(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "go", resultMap["language"])
		
		// Check that all expected keys are present
		_, hasVibeIssues := resultMap["vibe_issues"]
		assert.True(t, hasVibeIssues, "vibe_issues should be present")
		
		_, hasDuplicates := resultMap["duplicate_functions"]
		assert.True(t, hasDuplicates, "duplicate_functions should be present")
		
		_, hasOrphaned := resultMap["orphaned_code"]
		assert.True(t, hasOrphaned, "orphaned_code should be present")
		
		// Quality metrics should be present
		_, hasQualityMetrics := resultMap["quality_metrics"]
		assert.True(t, hasQualityMetrics, "quality_metrics should be present")
		
		// Maintainability index should be present and a number
		mi, hasMI := resultMap["maintainability_index"]
		assert.True(t, hasMI, "maintainability_index should be present")
		if mi != nil {
			_, ok := mi.(float64)
			assert.True(t, ok, "maintainability_index should be a float64")
		}
		
		// Technical debt should be present
		_, hasTD := resultMap["technical_debt"]
		assert.True(t, hasTD, "technical_debt should be present")
		
		// Refactoring priority should be present (may be empty slice)
		_, hasRP := resultMap["refactoring_priority"]
		assert.True(t, hasRP, "refactoring_priority should be present")
		
		assert.NotEmpty(t, resultMap["analyzed_at"])
	})

	t.Run("missing_code", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := models.CodeAnalysisRequest{
			Language: "go",
		}

		// When
		result, err := service.AnalyzeVibe(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "code is required")
	})

	t.Run("missing_language", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := models.CodeAnalysisRequest{
			Code: "package main",
		}

		// When
		result, err := service.AnalyzeVibe(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "language is required")
	})
}

func TestCodeAnalysisServiceImpl_calculateReadabilityScore(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	t.Run("good_readability", func(t *testing.T) {
		// Given
		code := `
// This is a well-documented function
func example() {
	// With comments
	return
}
`
		language := "go"

		// When
		score := service.calculateReadabilityScore(code, language)

		// Then
		assert.Greater(t, score, 80.0)
		assert.LessOrEqual(t, score, 100.0)
	})

	t.Run("poor_readability", func(t *testing.T) {
		// Given
		code := strings.Repeat("x", 150) + "\n" + strings.Repeat("y", 150)
		language := "go"

		// When
		score := service.calculateReadabilityScore(code, language)

		// Then
		assert.Less(t, score, 100.0)
		assert.GreaterOrEqual(t, score, 0.0)
	})
}

func TestCodeAnalysisServiceImpl_estimateIssueEffort(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	tests := []struct {
		name     string
		severity string
		issueType string
		expected float64
	}{
		{"critical_severity", "critical", "bug", 4.0 * 1.2},
		{"high_severity", "high", "security", 2.0 * 1.5},
		{"medium_severity", "medium", "style", 1.0 * 0.8},
		{"low_severity", "low", "general", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.estimateIssueEffort(tt.severity, tt.issueType)
			assert.Greater(t, result, 0.0)
		})
	}
}

func TestCodeAnalysisServiceImpl_determinePriority(t *testing.T) {
	service := NewCodeAnalysisService().(*CodeAnalysisServiceImpl)

	tests := []struct {
		name         string
		severity     string
		effortHours  float64
		expected     string
	}{
		{"critical_high_effort", "critical", 5.0, "high"},
		{"high_severity", "high", 2.5, "medium"},
		{"medium_severity", "medium", 1.0, "low"},
		{"low_severity", "low", 0.5, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determinePriority(tt.severity, tt.effortHours)
			assert.Equal(t, tt.expected, result)
		})
	}
}

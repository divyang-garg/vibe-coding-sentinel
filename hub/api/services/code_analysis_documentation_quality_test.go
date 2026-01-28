// Package services provides unit tests for documentation quality assessment functions
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreLanguageSpecificQuality_Go(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Go doc with parameter format
	docStr := "CalculateSum adds two numbers. a: first number, b: second number"
	score := impl.scoreLanguageSpecificQuality(docStr, "go")
	assert.Greater(t, score, 0.0)
}

func TestScoreLanguageSpecificQuality_GoWithComment(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Go doc with comment style
	docStr := "// CalculateSum adds two numbers"
	score := impl.scoreLanguageSpecificQuality(docStr, "go")
	assert.Greater(t, score, 0.0)
}

func TestScoreLanguageSpecificQuality_JavaScriptJSDoc(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// JSDoc format
	docStr := `/**
 * Calculates sum
 * @param {number} a - First number
 * @returns {number} The sum
 */`
	score := impl.scoreLanguageSpecificQuality(docStr, "javascript")
	assert.GreaterOrEqual(t, score, 10.0)
}

func TestScoreLanguageSpecificQuality_TypeScriptJSDoc(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// JSDoc format
	docStr := `/**
 * Calculates sum
 * @param {number} a - First number
 * @return {number} The sum
 */`
	score := impl.scoreLanguageSpecificQuality(docStr, "typescript")
	assert.GreaterOrEqual(t, score, 10.0)
}

func TestScoreLanguageSpecificQuality_PythonDocstring(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Python docstring
	docStr := `"""
Calculate sum of two numbers.

Args:
	a: First number
	b: Second number

Returns:
	Sum of a and b
"""`
	score := impl.scoreLanguageSpecificQuality(docStr, "python")
	assert.Greater(t, score, 0.0)
}

func TestScoreLanguageSpecificQuality_PythonGoogleStyle(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Google style docstring
	docStr := `Calculate sum.

Args:
	a: First number
	b: Second number

Returns:
	Sum of a and b`
	score := impl.scoreLanguageSpecificQuality(docStr, "python")
	assert.GreaterOrEqual(t, score, 5.0)
}

func TestScoreLanguageSpecificQuality_PythonNumPyStyle(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// NumPy style docstring
	docStr := `Calculate sum.

Parameters
----------
a : int
	First number
b : int
	Second number

Returns
-------
int
	Sum of a and b`
	score := impl.scoreLanguageSpecificQuality(docStr, "python")
	assert.GreaterOrEqual(t, score, 5.0)
}

func TestScoreLanguageSpecificQuality_UnknownLanguage(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	docStr := "Some documentation"
	score := impl.scoreLanguageSpecificQuality(docStr, "unknown")
	assert.Equal(t, 0.0, score)
}

func TestScoreFunctionDocumentation_WithParameters(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": "CalculateSum adds two numbers. Parameter a is the first number, parameter b is the second.",
		"parameters":   []string{"a", "b"},
		"returnType":   "int",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + parameter mention
}

func TestScoreFunctionDocumentation_WithReturnType(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": "CalculateSum adds two numbers and returns an int.",
		"returnType":    "int",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + return type mention
}

func TestScoreFunctionDocumentation_WithExamples(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": `CalculateSum adds two numbers.

Example:
	result := CalculateSum(1, 2)
	fmt.Println(result)`,
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + example
}

func TestScoreFunctionDocumentation_WithUsage(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": `CalculateSum adds two numbers.

Usage:
	result := CalculateSum(1, 2)`,
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + usage
}

func TestScoreFunctionDocumentation_WithCodeBlock(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": "CalculateSum adds two numbers.\n\n```\nresult := CalculateSum(1, 2)\n```",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + code block
}

func TestScoreFunctionDocumentation_OptimalLength(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation between 20-500 characters
	docStr := "This is a good documentation string that is between 20 and 500 characters long and should get full points for length."
	doc := map[string]interface{}{
		"documentation": docStr,
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + length bonus
}

func TestScoreFunctionDocumentation_TooShort(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation less than 20 characters
	doc := map[string]interface{}{
		"documentation": "Short",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Less(t, score, 60.0) // Base score only, no length bonus
}

func TestScoreFunctionDocumentation_TooLong(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation over 500 characters
	longDoc := ""
	for i := 0; i < 600; i++ {
		longDoc += "a"
	}
	doc := map[string]interface{}{
		"documentation": longDoc,
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + partial length bonus
	assert.Less(t, score, 60.0)     // But not full length bonus
}

func TestScoreFunctionDocumentation_WithFormatting(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation with newlines
	doc := map[string]interface{}{
		"documentation": `CalculateSum adds two numbers.

This function takes two parameters and returns their sum.`,
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + formatting bonus
}

func TestScoreFunctionDocumentation_WithDoubleSpaces(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation with double spaces (structured)
	doc := map[string]interface{}{
		"documentation": "CalculateSum  adds  two  numbers  with  proper  spacing.",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Greater(t, score, 50.0) // Base score + formatting bonus
}

func TestScoreFunctionDocumentation_NoDocumentation(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": "",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Equal(t, 0.0, score)
}

func TestScoreFunctionDocumentation_WhitespaceOnly(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	doc := map[string]interface{}{
		"documentation": "   ",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.Equal(t, 0.0, score)
}

func TestScoreFunctionDocumentation_ScoreBounds(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	// Documentation with all features
	doc := map[string]interface{}{
		"documentation": `CalculateSum adds two numbers. Parameter a is the first number, parameter b is the second. Returns an int.

Example:
	result := CalculateSum(1, 2)
	fmt.Println(result)`,
		"parameters": []string{"a", "b"},
		"returnType": "int",
		"language":   "go",
	}

	score := impl.scoreFunctionDocumentation(doc)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 100.0)
}

func TestAssessDocumentationQuality_InterfaceFormat(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	docs := map[string]interface{}{
		"functions": []interface{}{
			map[string]interface{}{
				"name":         "Function1",
				"documentation": "Function1 does something good",
			},
			map[string]interface{}{
				"name":         "Function2",
				"documentation": "Function2 does something else",
			},
		},
	}

	quality := impl.assessDocumentationQuality(docs)
	assert.Greater(t, quality, 0.0)
}

func TestAssessDocumentationQuality_InvalidFormat(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	docs := map[string]interface{}{
		"functions": "invalid", // Not a slice
	}

	quality := impl.assessDocumentationQuality(docs)
	assert.Equal(t, 0.0, quality)
}

func TestAssessDocumentationQuality_NotMap(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	quality := impl.assessDocumentationQuality("not a map")
	assert.Equal(t, 0.0, quality)
}

func TestAssessDocumentationQuality_MixedQuality(t *testing.T) {
	service := NewCodeAnalysisService()
	impl := service.(*CodeAnalysisServiceImpl)

	docs := map[string]interface{}{
		"functions": []map[string]interface{}{
			{
				"name":         "GoodFunction",
				"documentation": "This is a good documentation string that is between 20 and 500 characters long and should get full points.",
				"parameters":   []string{"a", "b"},
				"returnType":   "int",
			},
			{
				"name":         "BadFunction",
				"documentation": "x", // Poor quality
			},
		},
	}

	quality := impl.assessDocumentationQuality(docs)
	// Should be average of good and bad
	assert.Greater(t, quality, 0.0)
	assert.Less(t, quality, 100.0)
}

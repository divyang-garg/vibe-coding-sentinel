// Package services provides unit tests for code analysis service.
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestCodeAnalysisServiceImpl_AnalyzeCode(t *testing.T) {
	tests := []struct {
		name    string
		req     models.ASTAnalysisRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid Go code analysis",
			req: models.ASTAnalysisRequest{
				Code: `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func calculateSum(a, b int) int {
	return a + b
}
`,
				Language: "go",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			req: models.ASTAnalysisRequest{
				Code:     "",
				Language: "go",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "missing language",
			req: models.ASTAnalysisRequest{
				Code:     "some code",
				Language: "",
			},
			wantErr: true,
			errMsg:  "language is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCodeAnalysisService()

			result, err := service.AnalyzeCode(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "language")
				assert.Contains(t, resultMap, "code_length")
				assert.Contains(t, resultMap, "lines_count")
				assert.Contains(t, resultMap, "complexity")
				assert.Contains(t, resultMap, "quality_score")
				assert.Contains(t, resultMap, "issues")
				assert.Contains(t, resultMap, "suggestions")

				assert.Equal(t, tt.req.Language, resultMap["language"])
				assert.Equal(t, len(tt.req.Code), resultMap["code_length"])

				// Quality score should be reasonable
				qualityScore, ok := resultMap["quality_score"].(float64)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, qualityScore, 0.0)
				assert.LessOrEqual(t, qualityScore, 100.0)
			}
		})
	}
}

func TestCodeAnalysisServiceImpl_LintCode(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CodeLintRequest
		rules   []string
		wantErr bool
	}{
		{
			name: "valid linting",
			req: models.CodeLintRequest{
				Code: `
func badFunction() {
    x := 1
    y := 2
    z := x + y + 3 + 4 + 5  // Long line
    fmt.Println(z)
}
`,
				Language: "go",
			},
			rules:   []string{"long_line"},
			wantErr: false,
		},
		{
			name: "empty code",
			req: models.CodeLintRequest{
				Code:     "",
				Language: "go",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCodeAnalysisService()

			result, err := service.LintCode(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "language")
				assert.Contains(t, resultMap, "issues")
				assert.Contains(t, resultMap, "issue_count")
				assert.Contains(t, resultMap, "severity_breakdown")

				assert.Equal(t, tt.req.Language, resultMap["language"])

				issues, ok := resultMap["issues"].([]map[string]interface{})
				assert.True(t, ok)

				issueCount, ok := resultMap["issue_count"].(int)
				assert.True(t, ok)
				assert.Equal(t, len(issues), issueCount)
			}
		})
	}
}

func TestCodeAnalysisServiceImpl_RefactorCode(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CodeRefactorRequest
		wantErr bool
	}{
		{
			name: "valid refactoring request",
			req: models.CodeRefactorRequest{
				Code: `
func processData(data []int) {
    for i, v := range data {
        if v > 10 {
            data[i] = v * 2
        }
    }
}
`,
				Language: "go",
				Action:   "extract_method",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			req: models.CodeRefactorRequest{
				Code:     "",
				Language: "go",
				Action:   "extract_method",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCodeAnalysisService()

			result, err := service.RefactorCode(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "language")
				assert.Contains(t, resultMap, "action")
				assert.Contains(t, resultMap, "suggestions")
				assert.Contains(t, resultMap, "confidence_score")
				assert.Contains(t, resultMap, "estimated_savings")

				assert.Equal(t, tt.req.Language, resultMap["language"])
				assert.Equal(t, tt.req.Action, resultMap["action"])

				suggestions, ok := resultMap["suggestions"].([]map[string]interface{})
				assert.True(t, ok)
				assert.NotEmpty(t, suggestions)
			}
		})
	}
}

func TestCodeAnalysisServiceImpl_GenerateDocumentation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.DocumentationRequest
		wantErr bool
	}{
		{
			name: "valid documentation generation",
			req: models.DocumentationRequest{
				Code: `
package main

// User represents a system user
type User struct {
    ID   string
    Name string
}

// GetUser retrieves a user by ID
func GetUser(id string) (*User, error) {
    return &User{ID: id, Name: "Test"}, nil
}
`,
				Language: "go",
				Format:   "markdown",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			req: models.DocumentationRequest{
				Code:     "",
				Language: "go",
				Format:   "markdown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCodeAnalysisService()

			result, err := service.GenerateDocumentation(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "language")
				assert.Contains(t, resultMap, "format")
				assert.Contains(t, resultMap, "documentation")
				assert.Contains(t, resultMap, "coverage")
				assert.Contains(t, resultMap, "quality_score")

				assert.Equal(t, tt.req.Language, resultMap["language"])
				assert.Equal(t, tt.req.Format, resultMap["format"])

				coverage, ok := resultMap["coverage"].(float64)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, coverage, 0.0)
				assert.LessOrEqual(t, coverage, 100.0)
			}
		})
	}
}

func TestCodeAnalysisServiceImpl_ValidateCode(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CodeValidationRequest
		wantErr bool
	}{
		{
			name: "valid Go code validation",
			req: models.CodeValidationRequest{
				Code: `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`,
				Language: "go",
			},
			wantErr: false,
		},
		{
			name: "code with syntax issues",
			req: models.CodeValidationRequest{
				Code: `
package main

func main() {
    fmt.Println("Missing import")
}
`,
				Language: "go",
			},
			wantErr: false, // Our validation is basic
		},
		{
			name: "empty code",
			req: models.CodeValidationRequest{
				Code:     "",
				Language: "go",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCodeAnalysisService()

			result, err := service.ValidateCode(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "language")
				assert.Contains(t, resultMap, "is_valid")
				assert.Contains(t, resultMap, "errors")
				assert.Contains(t, resultMap, "warnings")
				assert.Contains(t, resultMap, "compliance")

				assert.Equal(t, tt.req.Language, resultMap["language"])
				assert.Contains(t, resultMap, "is_valid")
			}
		})
	}
}

func TestCodeAnalysisServiceImpl_ComplexityAnalysis(t *testing.T) {
	service := NewCodeAnalysisService()

	code := `
package main

import "fmt"

func main() {
    if true {
        for i := 0; i < 10; i++ {
            switch i % 2 {
            case 0:
                fmt.Println("even")
            case 1:
                fmt.Println("odd")
            }
        }
    }
}

func simpleFunction() int {
    return 42
}
`

	req := models.ASTAnalysisRequest{
		Code:     code,
		Language: "go",
	}

	result, err := service.AnalyzeCode(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	complexity, ok := resultMap["complexity"].(map[string]interface{})
	assert.True(t, ok)

	assert.Contains(t, complexity, "cyclomatic")
	assert.Contains(t, complexity, "functions")
	assert.Contains(t, complexity, "lines")

	cyclomatic, ok := complexity["cyclomatic"].(int)
	assert.True(t, ok)
	assert.Greater(t, cyclomatic, 0) // Should detect control structures
}

func TestCodeAnalysisServiceImpl_QualityScore(t *testing.T) {
	service := NewCodeAnalysisService()

	// High quality code
	goodCode := `
package main

import "fmt"

// User represents a user
type User struct {
    ID   string
    Name string
}

// NewUser creates a new user
func NewUser(id, name string) *User {
    return &User{
        ID:   id,
        Name: name,
    }
}

func main() {
    user := NewUser("1", "John")
    fmt.Printf("User: %+v\n", user)
}
`

	req := models.ASTAnalysisRequest{
		Code:     goodCode,
		Language: "go",
	}

	result, err := service.AnalyzeCode(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	score, ok := resultMap["quality_score"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, score, 70.0) // Should be reasonably high
	assert.LessOrEqual(t, score, 100.0)
}

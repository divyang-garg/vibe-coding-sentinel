// Package services - Unit tests for fix service
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestFixService_ApplyFix_Security(t *testing.T) {
	service := NewFixService()

	req := models.ApplyFixRequest{
		Code:     "const apiKey = 'secret123';",
		Language: "javascript",
		FixType:  "security",
	}

	result, err := service.ApplyFix(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Summary, "security")
	assert.NotEmpty(t, result.FixedCode)
}

func TestFixService_ApplyFix_Style(t *testing.T) {
	service := NewFixService()

	req := models.ApplyFixRequest{
		Code:     "const x = 1;   \n",
		Language: "javascript",
		FixType:  "style",
	}

	result, err := service.ApplyFix(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Summary, "style")
}

func TestFixService_ApplyFix_Performance(t *testing.T) {
	service := NewFixService()

	req := models.ApplyFixRequest{
		Code:     "for (let i = 0; i < 10; i++) { expensiveCall(); }",
		Language: "javascript",
		FixType:  "performance",
	}

	result, err := service.ApplyFix(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Summary, "performance")
}

func TestFixService_ApplyFix_InvalidFixType(t *testing.T) {
	service := NewFixService()

	req := models.ApplyFixRequest{
		Code:     "const x = 1;",
		Language: "javascript",
		FixType:  "invalid",
	}

	result, err := service.ApplyFix(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported fix type")
}

func TestFixService_ApplyFix_ContextCancellation(t *testing.T) {
	service := NewFixService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := models.ApplyFixRequest{
		Code:     "const x = 1;",
		Language: "javascript",
		FixType:  "security",
	}

	result, err := service.ApplyFix(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

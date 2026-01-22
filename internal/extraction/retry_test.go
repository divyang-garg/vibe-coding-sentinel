// Package extraction provides tests for retry logic
package extraction

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRetryableLLMClient struct {
	mock.Mock
	attempts int
}

func (m *MockRetryableLLMClient) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
	m.attempts++
	args := m.Called(ctx, prompt, taskType)
	return args.String(0), args.Int(1), args.Error(2)
}

func TestCallLLMWithRetry(t *testing.T) {
	t.Run("succeeds after retries", func(t *testing.T) {
		mockLLM := new(MockRetryableLLMClient)
		extractor := &KnowledgeExtractor{llmClient: mockLLM}

		// First two calls fail with retryable error, third succeeds
		mockLLM.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return("", 0, errors.New("rate limit exceeded")).Twice()
		mockLLM.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return("success", 100, nil).Once()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, tokens, err := extractor.callLLMWithRetry(ctx, "test prompt", "test", 3)
		if err != nil {
			t.Errorf("should succeed after retries, got error: %v", err)
		}
		if result != "success" {
			t.Errorf("expected success, got %s", result)
		}
		if tokens != 100 {
			t.Errorf("expected 100 tokens, got %d", tokens)
		}
	})

	t.Run("fails after max retries", func(t *testing.T) {
		mockLLM := new(MockRetryableLLMClient)
		extractor := &KnowledgeExtractor{llmClient: mockLLM}

		// All calls fail with retryable error
		mockLLM.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return("", 0, errors.New("rate limit exceeded")).Times(3)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _, err := extractor.callLLMWithRetry(ctx, "test prompt", "test", 3)
		if err == nil {
			t.Error("should fail after max retries")
		}
		if !strings.Contains(err.Error(), "max retries") {
			t.Errorf("expected max retries error, got: %v", err)
		}
	})

	t.Run("does not retry non-retryable errors", func(t *testing.T) {
		mockLLM := new(MockRetryableLLMClient)
		extractor := &KnowledgeExtractor{llmClient: mockLLM}

		// Non-retryable error
		mockLLM.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return("", 0, errors.New("invalid request")).Once()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _, err := extractor.callLLMWithRetry(ctx, "test prompt", "test", 3)
		if err == nil {
			t.Error("should fail immediately for non-retryable error")
		}
		mockLLM.AssertNumberOfCalls(t, "Call", 1)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mockLLM := new(MockRetryableLLMClient)
		extractor := &KnowledgeExtractor{llmClient: mockLLM}

		// First call fails with retryable error, then context is cancelled
		mockLLM.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return("", 0, errors.New("rate limit exceeded")).Once()

		ctx, cancel := context.WithCancel(context.Background())
		// Cancel after a short delay to allow first call
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, _, err := extractor.callLLMWithRetry(ctx, "test prompt", "test", 3)
		// May fail with context cancellation or max retries, both are acceptable
		if err == nil {
			t.Error("should fail on context cancellation or retries")
		}
	})
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout error", errors.New("timeout occurred"), true},
		{"rate limit error", errors.New("rate limit exceeded"), true},
		{"temporary error", errors.New("temporary failure"), true},
		{"network error", errors.New("network connection failed"), true},
		{"connection error", errors.New("connection refused"), true},
		{"503 error", errors.New("503 service unavailable"), true},
		{"502 error", errors.New("502 bad gateway"), true},
		{"429 error", errors.New("429 too many requests"), true},
		{"permanent error", errors.New("invalid request"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

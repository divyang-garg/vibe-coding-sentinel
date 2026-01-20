// Package extraction provides LLM-powered knowledge extraction
package extraction

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBasicLogger struct {
	mock.Mock
}

func (m *MockBasicLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockBasicLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockBasicLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockBasicLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestStructuredLogger(t *testing.T) {
	t.Run("formats message with fields", func(t *testing.T) {
		mockLogger := &MockBasicLogger{}
		mockLogger.On("Info", mock.MatchedBy(func(msg string) bool {
			return msg != "" // Just check it's called with something
		}), mock.Anything).Return()

		sl := NewStructuredLogger(mockLogger)
		sl.Info("test message", map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		})

		mockLogger.AssertExpectations(t)
	})

	t.Run("handles empty fields", func(t *testing.T) {
		mockLogger := &MockBasicLogger{}
		mockLogger.On("Debug", "simple message", mock.Anything).Return()

		sl := NewStructuredLogger(mockLogger)
		sl.Debug("simple message", map[string]interface{}{})

		mockLogger.AssertExpectations(t)
	})

	t.Run("adds error to fields", func(t *testing.T) {
		mockLogger := &MockBasicLogger{}
		mockLogger.On("Error", mock.MatchedBy(func(msg string) bool {
			return msg != ""
		}), mock.Anything).Return()

		sl := NewStructuredLogger(mockLogger)
		fields := map[string]interface{}{"context": "test"}
		sl.Error("error occurred", fields, errors.New("test error"))

		// Verify error was added to fields
		assert.Equal(t, "test error", fields["error"])
		mockLogger.AssertExpectations(t)
	})
}

func TestFormatMessage(t *testing.T) {
	t.Run("returns message only when no fields", func(t *testing.T) {
		result := formatMessage("test", nil)
		assert.Equal(t, "test", result)
	})

	t.Run("appends fields to message", func(t *testing.T) {
		result := formatMessage("test", map[string]interface{}{
			"key": "value",
		})
		assert.Contains(t, result, "test")
		assert.Contains(t, result, "key=value")
	})
}

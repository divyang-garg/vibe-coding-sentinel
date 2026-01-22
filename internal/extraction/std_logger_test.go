// Package extraction provides tests for standard logger
package extraction

import (
	"os"
	"testing"
)

func TestNewStdLogger(t *testing.T) {
	t.Run("creates logger with default level", func(t *testing.T) {
		os.Unsetenv("EXTRACTION_LOG_LEVEL")
		logger := NewStdLogger()
		if logger == nil {
			t.Error("logger should not be nil")
		}
	})

	t.Run("creates logger with debug level", func(t *testing.T) {
		os.Setenv("EXTRACTION_LOG_LEVEL", "debug")
		defer os.Unsetenv("EXTRACTION_LOG_LEVEL")
		logger := NewStdLogger()
		if logger == nil {
			t.Error("logger should not be nil")
		}
	})
}

func TestStdLogger_LogMethods(t *testing.T) {
	logger := NewStdLogger().(*StdLogger)

	t.Run("debug logs when level is debug", func(t *testing.T) {
		logger.level = LogLevelDebug
		logger.Debug("test debug message")
		// Should not panic
	})

	t.Run("info logs when level is info or lower", func(t *testing.T) {
		logger.level = LogLevelInfo
		logger.Info("test info message")
		// Should not panic
	})

	t.Run("warn logs when level is warn or lower", func(t *testing.T) {
		logger.level = LogLevelWarn
		logger.Warn("test warn message")
		// Should not panic
	})

	t.Run("error logs when level is error or lower", func(t *testing.T) {
		logger.level = LogLevelError
		logger.Error("test error message")
		// Should not panic
	})

	t.Run("debug does not log when level is info", func(t *testing.T) {
		logger.level = LogLevelInfo
		logger.Debug("should not log")
		// Should not panic
	})
}

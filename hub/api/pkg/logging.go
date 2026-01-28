// Phase 12: Structured Logging
// Provides structured logging with levels and request ID tracking

package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

var currentLogLevel LogLevel = LogLevelInfo

// contextKey type for context keys
type contextKey string

// Context keys for request tracing
const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	UserIDKey    contextKey = "user_id"
	ProjectIDKey contextKey = "project_id"
)

// requestIDKey is the context key for request ID (deprecated, use RequestIDKey)
const requestIDKey contextKey = "requestID"

func init() {
	if level := os.Getenv("SENTINEL_LOG_LEVEL"); level != "" {
		currentLogLevel = LogLevel(level)
	}
}

func shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	return levels[level] >= levels[currentLogLevel]
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}

func logMessage(ctx context.Context, level LogLevel, msg string, args ...interface{}) {
	if !shouldLog(level) {
		return
	}
	requestID := getRequestID(ctx)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMsg := fmt.Sprintf(msg, args...)
	log.Printf("[%s] [%s] [%s] %s", timestamp, level, requestID, formattedMsg)
}

func LogDebug(ctx context.Context, msg string, args ...interface{}) {
	logMessage(ctx, LogLevelDebug, msg, args...)
}

func LogInfo(ctx context.Context, msg string, args ...interface{}) {
	logMessage(ctx, LogLevelInfo, msg, args...)
}

func LogWarn(ctx context.Context, msg string, args ...interface{}) {
	logMessage(ctx, LogLevelWarn, msg, args...)
}

func LogError(ctx context.Context, msg string, args ...interface{}) {
	logMessage(ctx, LogLevelError, msg, args...)
}

// LogErrorWithErr logs an error with an error object for structured logging
func LogErrorWithErr(ctx context.Context, err error, msg string, fields ...interface{}) {
	if !shouldLog(LogLevelError) {
		return
	}
	requestID := getRequestID(ctx)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Build structured log message
	logMsg := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, LogLevelError, requestID, msg)
	if err != nil {
		logMsg += fmt.Sprintf(" error=%v", err)
	}
	if len(fields) > 0 {
		logMsg += fmt.Sprintf(" %v", fields)
	}
	// Use constant format string to prevent format string injection
	log.Printf("%s", logMsg)
}

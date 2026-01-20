// Package pkg provides shared utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Error       *ErrorInfo             `json:"error,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Duration    *float64               `json:"duration_ms,omitempty"`
}

// ErrorInfo contains error details for logging
type ErrorInfo struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// JSONLogger implements structured JSON logging
type JSONLogger struct {
	writer      io.Writer
	level       LogLevel
	serviceName string
	version     string
	environment string
	mu          sync.Mutex
}

// JSONLoggerConfig configures the JSON logger
type JSONLoggerConfig struct {
	Writer      io.Writer
	Level       LogLevel
	ServiceName string
	Version     string
	Environment string
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(cfg JSONLoggerConfig) *JSONLogger {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = getEnvOrDefault("SERVICE_NAME", "sentinel-hub-api")
	}
	if cfg.Version == "" {
		cfg.Version = getEnvOrDefault("SERVICE_VERSION", "unknown")
	}
	if cfg.Environment == "" {
		cfg.Environment = getEnvOrDefault("ENVIRONMENT", "development")
	}
	if cfg.Level == "" {
		levelStr := getEnvOrDefault("LOG_LEVEL", "INFO")
		cfg.Level = LogLevel(levelStr)
	}
	return &JSONLogger{
		writer:      cfg.Writer,
		level:       cfg.Level,
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
		environment: cfg.Environment,
	}
}

// Log writes a structured log entry
func (l *JSONLogger) Log(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}) {
	if !shouldLogLevel(level, l.level) {
		return
	}

	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Level:       string(level),
		Message:     msg,
		Service:     l.serviceName,
		Version:     l.version,
		Environment: l.environment,
		Fields:      fields,
	}

	// Extract context values
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		entry.RequestID = requestID
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		entry.TraceID = traceID
	}
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
		entry.SpanID = spanID
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		entry.UserID = userID
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(entry)
	l.writer.Write(data)
	l.writer.Write([]byte("\n"))
}

// Info logs at INFO level
func (l *JSONLogger) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.Log(ctx, LogLevelInfo, msg, f)
}

// Error logs at ERROR level with error details
func (l *JSONLogger) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	if err != nil {
		stackTrace := getStackTrace()
		f["error"] = map[string]interface{}{
			"type":        fmt.Sprintf("%T", err),
			"message":     err.Error(),
			"stack_trace": stackTrace,
		}
		entry := LogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
			Level:       string(LogLevelError),
			Message:     msg,
			Service:     l.serviceName,
			Version:     l.version,
			Environment: l.environment,
			Fields:      f,
			Error: &ErrorInfo{
				Type:       fmt.Sprintf("%T", err),
				Message:    err.Error(),
				StackTrace: stackTrace,
			},
		}
		if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
			entry.RequestID = requestID
		}
		if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
			entry.TraceID = traceID
		}
		if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
			entry.SpanID = spanID
		}
		if userID, ok := ctx.Value(UserIDKey).(string); ok {
			entry.UserID = userID
		}
		l.mu.Lock()
		defer l.mu.Unlock()
		data, _ := json.Marshal(entry)
		l.writer.Write(data)
		l.writer.Write([]byte("\n"))
		return
	}
	l.Log(ctx, LogLevelError, msg, f)
}

// Debug logs at DEBUG level
func (l *JSONLogger) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.Log(ctx, LogLevelDebug, msg, f)
}

// Warn logs at WARN level
func (l *JSONLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, LogLevelWarn, msg, fields)
}

func mergeFields(fields []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func shouldLogLevel(level, currentLevel LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	return levels[level] >= levels[currentLevel]
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	if n > 0 {
		return string(buf[:n])
	}
	return ""
}

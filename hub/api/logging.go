// Phase 12: Structured Logging
// Provides structured logging with levels and request ID tracking

package main

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

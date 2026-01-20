// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"log"
	"os"
)

// StdLogger implements Logger using standard library log
type StdLogger struct {
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	level    LogLevel
}

// LogLevel represents logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// NewStdLogger creates a new standard logger
func NewStdLogger() Logger {
	level := LogLevelInfo
	if os.Getenv("EXTRACTION_LOG_LEVEL") == "debug" {
		level = LogLevelDebug
	}

	return &StdLogger{
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		infoLog:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warnLog:  log.New(os.Stderr, "[WARN] ", log.LstdFlags),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		level:    level,
	}
}

func (l *StdLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.debugLog.Printf(msg, args...)
	}
}

func (l *StdLogger) Info(msg string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.infoLog.Printf(msg, args...)
	}
}

func (l *StdLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.warnLog.Printf(msg, args...)
	}
}

func (l *StdLogger) Error(msg string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.errorLog.Printf(msg, args...)
	}
}

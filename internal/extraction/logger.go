// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import "fmt"

// StructuredLogger provides structured logging interface
type StructuredLogger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{}, err error)
}

// simpleStructuredLogger implements StructuredLogger with basic formatting
type simpleStructuredLogger struct {
	logger Logger
}

// NewStructuredLogger creates a structured logger from basic logger
func NewStructuredLogger(logger Logger) StructuredLogger {
	return &simpleStructuredLogger{logger: logger}
}

func (l *simpleStructuredLogger) Debug(msg string, fields map[string]interface{}) {
	l.logger.Debug(formatMessage(msg, fields))
}

func (l *simpleStructuredLogger) Info(msg string, fields map[string]interface{}) {
	l.logger.Info(formatMessage(msg, fields))
}

func (l *simpleStructuredLogger) Warn(msg string, fields map[string]interface{}) {
	l.logger.Warn(formatMessage(msg, fields))
}

func (l *simpleStructuredLogger) Error(msg string, fields map[string]interface{}, err error) {
	if err != nil {
		fields["error"] = err.Error()
	}
	l.logger.Error(formatMessage(msg, fields))
}

func formatMessage(msg string, fields map[string]interface{}) string {
	if len(fields) == 0 {
		return msg
	}
	formatted := msg
	for k, v := range fields {
		formatted += fmt.Sprintf(" %s=%v", k, v)
	}
	return formatted
}

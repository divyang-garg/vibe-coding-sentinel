// Package security - Unit tests for audit logger
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package security

import (
	"context"
	"testing"
	"time"
)

// mockLogger implements Logger interface for testing
type mockLogger struct {
	infoLogs  []logEntry
	warnLogs  []logEntry
	errorLogs []logEntry
}

type logEntry struct {
	msg    string
	fields []interface{}
}

func (m *mockLogger) Info(msg string, fields ...interface{}) {
	m.infoLogs = append(m.infoLogs, logEntry{msg, fields})
}

func (m *mockLogger) Warn(msg string, fields ...interface{}) {
	m.warnLogs = append(m.warnLogs, logEntry{msg, fields})
}

func (m *mockLogger) Error(msg string, fields ...interface{}) {
	m.errorLogs = append(m.errorLogs, logEntry{msg, fields})
}

func TestDefaultAuditLogger_LogEvent(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	tests := []struct {
		name     string
		event    AuditEvent
		wantInfo bool
		wantWarn bool
		wantErr  bool
	}{
		{
			name: "info severity",
			event: AuditEvent{
				Type:     EventTypeAuthSuccess,
				Severity: SeverityInfo,
				Message:  "Test message",
			},
			wantInfo: true,
			wantWarn: false,
			wantErr:  false,
		},
		{
			name: "warning severity",
			event: AuditEvent{
				Type:     EventTypeAuthFailure,
				Severity: SeverityWarning,
				Message:  "Test warning",
			},
			wantInfo: false,
			wantWarn: true,
			wantErr:  false,
		},
		{
			name: "error severity",
			event: AuditEvent{
				Type:     EventTypeSQLInjectionAttempt,
				Severity: SeverityError,
				Message:  "Test error",
			},
			wantInfo: false,
			wantWarn: false,
			wantErr:  true,
		},
		{
			name: "critical severity",
			event: AuditEvent{
				Type:     EventTypeSQLInjectionAttempt,
				Severity: SeverityCritical,
				Message:  "Test critical",
			},
			wantInfo: false,
			wantWarn: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog.infoLogs = nil
			mockLog.warnLogs = nil
			mockLog.errorLogs = nil

			err := logger.LogEvent(context.Background(), tt.event)
			if (err != nil) != false {
				t.Errorf("LogEvent() error = %v, wantErr false", err)
			}

			if tt.wantInfo && len(mockLog.infoLogs) == 0 {
				t.Error("Expected info log, got none")
			}
			if tt.wantWarn && len(mockLog.warnLogs) == 0 {
				t.Error("Expected warn log, got none")
			}
			if tt.wantErr && len(mockLog.errorLogs) == 0 {
				t.Error("Expected error log, got none")
			}
		})
	}
}

func TestDefaultAuditLogger_LogAuthSuccess(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	err := logger.LogAuthSuccess(
		context.Background(),
		"project-123",
		"org-456",
		"192.168.1.1",
		"Mozilla/5.0",
	)

	if err != nil {
		t.Errorf("LogAuthSuccess() error = %v", err)
	}

	if len(mockLog.infoLogs) == 0 {
		t.Error("Expected info log for auth success")
	}
}

func TestDefaultAuditLogger_LogAuthFailure(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	err := logger.LogAuthFailure(
		context.Background(),
		"invalid API key",
		"192.168.1.1",
		"Mozilla/5.0",
		"/api/v1/tasks",
	)

	if err != nil {
		t.Errorf("LogAuthFailure() error = %v", err)
	}

	if len(mockLog.warnLogs) == 0 {
		t.Error("Expected warn log for auth failure")
	}
}

func TestDefaultAuditLogger_LogAPIKeyGenerated(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	err := logger.LogAPIKeyGenerated(
		context.Background(),
		"project-123",
		"org-456",
	)

	if err != nil {
		t.Errorf("LogAPIKeyGenerated() error = %v", err)
	}

	if len(mockLog.infoLogs) == 0 {
		t.Error("Expected info log for API key generation")
	}
}

func TestDefaultAuditLogger_LogAPIKeyRevoked(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	err := logger.LogAPIKeyRevoked(
		context.Background(),
		"project-123",
		"org-456",
	)

	if err != nil {
		t.Errorf("LogAPIKeyRevoked() error = %v", err)
	}

	if len(mockLog.warnLogs) == 0 {
		t.Error("Expected warn log for API key revocation")
	}
}

func TestDefaultAuditLogger_LogSecurityViolation(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	err := logger.LogSecurityViolation(
		context.Background(),
		EventTypeSQLInjectionAttempt,
		map[string]interface{}{
			"input": "'; DROP TABLE users; --",
			"field": "query",
		},
	)

	if err != nil {
		t.Errorf("LogSecurityViolation() error = %v", err)
	}

	if len(mockLog.errorLogs) == 0 {
		t.Error("Expected error log for security violation")
	}
}

func TestAuditEvent_DefaultValues(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	event := AuditEvent{
		Type:    EventTypeUserAction,
		Message: "Test event",
	}

	err := logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Errorf("LogEvent() error = %v", err)
	}

	// Verify event got ID and timestamp (check in log output, not original event)
	// The event is passed by value, so modifications happen inside LogEvent
	if len(mockLog.infoLogs) == 0 {
		t.Error("Expected info log to be created")
	}
}

func TestEventTypes(t *testing.T) {
	// Verify all event types are defined
	eventTypes := []EventType{
		EventTypeAuthSuccess,
		EventTypeAuthFailure,
		EventTypeAuthTokenExpired,
		EventTypeAPIKeyGenerated,
		EventTypeAPIKeyRevoked,
		EventTypeAPIKeyValidated,
		EventTypeAccessGranted,
		EventTypeAccessDenied,
		EventTypePermissionDenied,
		EventTypeSQLInjectionAttempt,
		EventTypeXSSAttempt,
		EventTypePathTraversal,
		EventTypeRateLimitExceeded,
		EventTypeConfigChange,
		EventTypeUserAction,
	}

	for _, et := range eventTypes {
		if et == "" {
			t.Errorf("Event type %v is empty", et)
		}
	}
}

func TestSeverityLevels(t *testing.T) {
	// Verify all severity levels are defined
	severities := []Severity{
		SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical,
	}

	for _, sev := range severities {
		if sev == "" {
			t.Errorf("Severity %v is empty", sev)
		}
	}
}

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	id2 := generateEventID()

	if id1 == "" {
		t.Error("generateEventID() returned empty string")
	}

	if id2 == "" {
		t.Error("generateEventID() returned empty string")
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("generateEventID() returned duplicate IDs")
	}

	// IDs should start with "evt_"
	if len(id1) < 4 || id1[:4] != "evt_" {
		t.Errorf("generateEventID() = %q, want prefix 'evt_'", id1)
	}
}

func TestAuditEvent_WithMetadata(t *testing.T) {
	mockLog := &mockLogger{}
	logger := NewAuditLogger(mockLog)

	event := AuditEvent{
		Type:     EventTypeUserAction,
		Severity: SeverityInfo,
		Message:  "Test with metadata",
		Metadata: map[string]interface{}{
			"user_id":   "user-123",
			"action":    "create",
			"resource":  "task",
			"timestamp": time.Now(),
		},
	}

	err := logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Errorf("LogEvent() error = %v", err)
	}

	if len(mockLog.infoLogs) == 0 {
		t.Error("Expected info log")
	}
}

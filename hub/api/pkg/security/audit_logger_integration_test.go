// Package security - Integration tests for audit logger
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package security

import (
	"context"
	"testing"
	"time"
)

// mockLoggerIntegration implements Logger interface for integration testing
type mockLoggerIntegration struct {
	infoLogs  []logEntryIntegration
	warnLogs  []logEntryIntegration
	errorLogs []logEntryIntegration
}

type logEntryIntegration struct {
	msg    string
	fields []interface{}
}

func (m *mockLoggerIntegration) Info(msg string, fields ...interface{}) {
	m.infoLogs = append(m.infoLogs, logEntryIntegration{msg, fields})
}

func (m *mockLoggerIntegration) Warn(msg string, fields ...interface{}) {
	m.warnLogs = append(m.warnLogs, logEntryIntegration{msg, fields})
}

func (m *mockLoggerIntegration) Error(msg string, fields ...interface{}) {
	m.errorLogs = append(m.errorLogs, logEntryIntegration{msg, fields})
}

func TestAuditLogger_Integration_AuthFlow(t *testing.T) {
	mockLog := &mockLoggerIntegration{}
	logger := NewAuditLogger(mockLog)

	// Simulate authentication flow
	ctx := context.Background()

	// Test successful authentication
	err := logger.LogAuthSuccess(ctx, "project-123", "org-456", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("LogAuthSuccess() error = %v", err)
	}

	if len(mockLog.infoLogs) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(mockLog.infoLogs))
	}

	// Test failed authentication
	err = logger.LogAuthFailure(ctx, "invalid API key", "192.168.1.2", "Chrome/1.0", "/api/v1/tasks")
	if err != nil {
		t.Fatalf("LogAuthFailure() error = %v", err)
	}

	if len(mockLog.warnLogs) != 1 {
		t.Errorf("Expected 1 warn log, got %d", len(mockLog.warnLogs))
	}
}

func TestAuditLogger_Integration_APIKeyFlow(t *testing.T) {
	mockLog := &mockLoggerIntegration{}
	logger := NewAuditLogger(mockLog)

	ctx := context.Background()

	// Test API key generation
	err := logger.LogAPIKeyGenerated(ctx, "project-123", "org-456")
	if err != nil {
		t.Fatalf("LogAPIKeyGenerated() error = %v", err)
	}

	if len(mockLog.infoLogs) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(mockLog.infoLogs))
	}

	// Test API key revocation
	err = logger.LogAPIKeyRevoked(ctx, "project-123", "org-456")
	if err != nil {
		t.Fatalf("LogAPIKeyRevoked() error = %v", err)
	}

	if len(mockLog.warnLogs) != 1 {
		t.Errorf("Expected 1 warn log, got %d", len(mockLog.warnLogs))
	}
}

func TestAuditLogger_Integration_SecurityViolations(t *testing.T) {
	mockLog := &mockLoggerIntegration{}
	logger := NewAuditLogger(mockLog)

	ctx := context.Background()

	// Test SQL injection attempt
	err := logger.LogSecurityViolation(ctx, EventTypeSQLInjectionAttempt, map[string]interface{}{
		"input": "'; DROP TABLE users; --",
		"field": "query",
		"ip":    "192.168.1.100",
	})
	if err != nil {
		t.Fatalf("LogSecurityViolation() error = %v", err)
	}

	if len(mockLog.errorLogs) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(mockLog.errorLogs))
	}

	// Test XSS attempt
	err = logger.LogSecurityViolation(ctx, EventTypeXSSAttempt, map[string]interface{}{
		"input": "<script>alert('xss')</script>",
		"field": "comment",
	})
	if err != nil {
		t.Fatalf("LogSecurityViolation() error = %v", err)
	}

	if len(mockLog.errorLogs) != 2 {
		t.Errorf("Expected 2 error logs, got %d", len(mockLog.errorLogs))
	}
}

func TestAuditLogger_Integration_EventMetadata(t *testing.T) {
	mockLog := &mockLoggerIntegration{}
	logger := NewAuditLogger(mockLog)

	ctx := context.Background()

	event := AuditEvent{
		Type:     EventTypeUserAction,
		Severity: SeverityInfo,
		Message:  "User performed action",
		UserID:   "user-789",
		ProjectID: "project-123",
		OrgID:    "org-456",
		IPAddress: "10.0.0.1",
		UserAgent: "Test Agent",
		Path:     "/api/v1/tasks",
		Method:   "POST",
		Metadata: map[string]interface{}{
			"action":    "create",
			"resource":  "task",
			"timestamp": time.Now(),
		},
		Success: true,
	}

	err := logger.LogEvent(ctx, event)
	if err != nil {
		t.Fatalf("LogEvent() error = %v", err)
	}

	if len(mockLog.infoLogs) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(mockLog.infoLogs))
	}
}

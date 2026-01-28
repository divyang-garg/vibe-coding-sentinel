// Package security provides security event logging and audit trail
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package security

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// EventType represents the type of security event
type EventType string

const (
	// Authentication events
	EventTypeAuthSuccess      EventType = "auth_success"
	EventTypeAuthFailure      EventType = "auth_failure"
	EventTypeAuthTokenExpired EventType = "auth_token_expired"

	// API Key events
	EventTypeAPIKeyGenerated EventType = "api_key_generated"
	EventTypeAPIKeyRevoked   EventType = "api_key_revoked"
	EventTypeAPIKeyValidated EventType = "api_key_validated"

	// Authorization events
	EventTypeAccessGranted    EventType = "access_granted"
	EventTypeAccessDenied     EventType = "access_denied"
	EventTypePermissionDenied EventType = "permission_denied"

	// Security violations
	EventTypeSQLInjectionAttempt EventType = "sql_injection_attempt"
	EventTypeXSSAttempt          EventType = "xss_attempt"
	EventTypePathTraversal       EventType = "path_traversal"
	EventTypeRateLimitExceeded   EventType = "rate_limit_exceeded"

	// System events
	EventTypeConfigChange EventType = "config_change"
	EventTypeUserAction   EventType = "user_action"
)

// Severity represents the severity level of a security event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Severity  Severity               `json:"severity"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	ProjectID string                 `json:"project_id,omitempty"`
	OrgID     string                 `json:"org_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Success   bool                   `json:"success"`
}

// AuditLogger interface for security event logging
type AuditLogger interface {
	LogEvent(ctx context.Context, event AuditEvent) error
	LogAuthSuccess(ctx context.Context, projectID, orgID, ipAddress, userAgent string) error
	LogAuthFailure(ctx context.Context, reason, ipAddress, userAgent, path string) error
	LogAPIKeyGenerated(ctx context.Context, projectID, orgID string) error
	LogAPIKeyRevoked(ctx context.Context, projectID, orgID string) error
	LogSecurityViolation(ctx context.Context, eventType EventType, details map[string]interface{}) error
}

// DefaultAuditLogger implements AuditLogger with structured logging
type DefaultAuditLogger struct {
	logger Logger
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger Logger) AuditLogger {
	return &DefaultAuditLogger{
		logger: logger,
	}
}

// LogEvent logs a security audit event
func (l *DefaultAuditLogger) LogEvent(ctx context.Context, event AuditEvent) error {
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Convert to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Log based on severity
	logMsg := fmt.Sprintf("Security Event: %s - %s", event.Type, event.Message)
	fields := []interface{}{
		"event_id", event.ID,
		"event_type", string(event.Type),
		"severity", string(event.Severity),
		"timestamp", event.Timestamp.Format(time.RFC3339),
		"event_data", string(eventJSON),
	}

	if event.ProjectID != "" {
		fields = append(fields, "project_id", event.ProjectID)
	}
	if event.OrgID != "" {
		fields = append(fields, "org_id", event.OrgID)
	}
	if event.IPAddress != "" {
		fields = append(fields, "ip_address", event.IPAddress)
	}

	switch event.Severity {
	case SeverityCritical, SeverityError:
		l.logger.Error(logMsg, fields...)
	case SeverityWarning:
		l.logger.Warn(logMsg, fields...)
	default:
		l.logger.Info(logMsg, fields...)
	}

	return nil
}

// LogAuthSuccess logs successful authentication
func (l *DefaultAuditLogger) LogAuthSuccess(ctx context.Context, projectID, orgID, ipAddress, userAgent string) error {
	event := AuditEvent{
		Type:      EventTypeAuthSuccess,
		Severity:  SeverityInfo,
		ProjectID: projectID,
		OrgID:     orgID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Message:   "Authentication successful",
		Success:   true,
	}
	return l.LogEvent(ctx, event)
}

// LogAuthFailure logs failed authentication attempt
func (l *DefaultAuditLogger) LogAuthFailure(ctx context.Context, reason, ipAddress, userAgent, path string) error {
	event := AuditEvent{
		Type:      EventTypeAuthFailure,
		Severity:  SeverityWarning,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Path:      path,
		Message:   fmt.Sprintf("Authentication failed: %s", reason),
		Metadata: map[string]interface{}{
			"reason": reason,
		},
		Success: false,
	}
	return l.LogEvent(ctx, event)
}

// LogAPIKeyGenerated logs API key generation
func (l *DefaultAuditLogger) LogAPIKeyGenerated(ctx context.Context, projectID, orgID string) error {
	event := AuditEvent{
		Type:      EventTypeAPIKeyGenerated,
		Severity:  SeverityInfo,
		ProjectID: projectID,
		OrgID:     orgID,
		Message:   "API key generated",
		Success:   true,
	}
	return l.LogEvent(ctx, event)
}

// LogAPIKeyRevoked logs API key revocation
func (l *DefaultAuditLogger) LogAPIKeyRevoked(ctx context.Context, projectID, orgID string) error {
	event := AuditEvent{
		Type:      EventTypeAPIKeyRevoked,
		Severity:  SeverityWarning,
		ProjectID: projectID,
		OrgID:     orgID,
		Message:   "API key revoked",
		Success:   true,
	}
	return l.LogEvent(ctx, event)
}

// LogSecurityViolation logs security violation attempts
func (l *DefaultAuditLogger) LogSecurityViolation(ctx context.Context, eventType EventType, details map[string]interface{}) error {
	event := AuditEvent{
		Type:     eventType,
		Severity: SeverityError,
		Message:  fmt.Sprintf("Security violation detected: %s", eventType),
		Metadata: details,
		Success:  false,
	}
	return l.LogEvent(ctx, event)
}

var (
	// eventIDCounter provides atomic counter for unique event IDs
	eventIDCounter uint64
)

// generateEventID generates a unique event ID
// Uses timestamp + atomic counter to ensure uniqueness even when called in rapid succession
func generateEventID() string {
	now := time.Now()
	counter := atomic.AddUint64(&eventIDCounter, 1)
	// Use Unix timestamp, nanosecond precision, and atomic counter for guaranteed uniqueness
	return fmt.Sprintf("evt_%d_%d_%d", now.Unix(), now.UnixNano(), counter)
}

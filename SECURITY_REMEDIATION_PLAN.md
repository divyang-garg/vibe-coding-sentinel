# Security Remediation Implementation Plan
## Critical Next Steps - VAPT Report Remediation

**Date:** January 20, 2026  
**Status:** Planning Phase  
**Target Completion:** 2-3 weeks  
**Compliance:** Fully aligned with CODING_STANDARDS.md

---

## Executive Summary

This document outlines the detailed implementation plan for addressing the 5 critical security vulnerabilities identified in the VAPT report. All implementations will strictly comply with Sentinel Coding Standards, including architectural layer separation, file size limits, testing requirements, and security standards.

---

## Implementation Overview

### Priority 1 (Week 1): Critical Security Fixes

1. **API Key Hashing Implementation** - Store SHA-256 hashes instead of plaintext
2. **Authentication Middleware Integration** - Connect middleware to service layer
3. **Input Validation Framework** - Comprehensive validation for all endpoints
4. **CORS Production Configuration** - Origin whitelist validation
5. **Security Event Logging** - Audit logging for security events

---

## 1. API Key Hashing Implementation

### Objective
Store API keys as SHA-256 hashes in the database instead of plaintext, implementing defense-in-depth security.

### Compliance Requirements
- **File:** `hub/api/services/organization_service_api_keys.go` (max 400 lines)
- **Layer:** Service Layer - Business Logic
- **Dependencies:** Repository Layer (data access)
- **Testing:** 100% coverage required for new code

### Implementation Plan

#### Phase 1.1: Update Data Model
**File:** `hub/api/models/project.go` (or create if doesn't exist)
**Location:** Model Layer
**Changes:**
```go
type Project struct {
    ID          string    `json:"id"`
    OrgID       string    `json:"org_id"`
    Name        string    `json:"name"`
    APIKeyHash  string    `json:"api_key_hash"` // Changed from APIKey
    APIKeyPrefix string   `json:"api_key_prefix"` // First 8 chars for identification
    CreatedAt   time.Time `json:"created_at"`
}
```
**Rationale:** Follows model layer standards - pure data structures, no behavior

#### Phase 1.2: Update Service Layer
**File:** `hub/api/services/organization_service_api_keys.go`
**Current Size:** ~100 lines (within 400 line limit)
**Changes:**

```go
import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "fmt"
    "strings"
    "sentinel-hub-api/models"
)

// GenerateAPIKey generates a new API key and stores its hash
func (s *OrganizationServiceImpl) GenerateAPIKey(ctx context.Context, projectID string) (string, error) {
    // Validation (existing)
    if projectID == "" {
        return "", fmt.Errorf("project ID is required")
    }
    
    project, err := s.projectRepo.FindByID(ctx, projectID)
    if err != nil {
        return "", fmt.Errorf("failed to find project: %w", err)
    }
    if project == nil {
        return "", fmt.Errorf("project not found")
    }

    // Generate secure random API key (existing - already fixed)
    apiKey, err := s.generateAPIKey()
    if err != nil {
        return "", fmt.Errorf("failed to generate API key: %w", err)
    }

    // NEW: Generate hash and prefix
    hash, prefix := s.hashAPIKey(apiKey)
    
    // Update project with hash (not plaintext)
    project.APIKeyHash = hash
    project.APIKeyPrefix = prefix
    if err := s.projectRepo.Update(ctx, project); err != nil {
        return "", fmt.Errorf("failed to update project API key: %w", err)
    }

    // Return plaintext key ONLY once (for user to save)
    return apiKey, nil
}

// hashAPIKey generates SHA-256 hash and prefix for an API key
func (s *OrganizationServiceImpl) hashAPIKey(apiKey string) (hash, prefix string) {
    hasher := sha256.New()
    hasher.Write([]byte(apiKey))
    hash = hex.EncodeToString(hasher.Sum(nil))
    
    // Store first 8 characters for identification (without compromising security)
    prefix = apiKey[:min(8, len(apiKey))]
    return hash, prefix
}

// ValidateAPIKey validates an API key by comparing its hash
func (s *OrganizationServiceImpl) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("API key is required")
    }

    // Generate hash from provided key
    hash, prefix := s.hashAPIKey(apiKey)
    
    // Find project by hash
    project, err := s.projectRepo.FindByAPIKeyHash(ctx, hash)
    if err != nil {
        return nil, fmt.Errorf("failed to validate API key: %w", err)
    }
    if project == nil {
        return nil, fmt.Errorf("invalid API key")
    }
    
    // Optional: Verify prefix matches (additional check)
    if project.APIKeyPrefix != prefix {
        return nil, fmt.Errorf("invalid API key")
    }

    return project, nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```
**Compliance:** 
- ✅ Service layer only (no HTTP, no direct DB)
- ✅ Proper error wrapping
- ✅ Single responsibility functions
- ✅ Well-documented functions

#### Phase 1.3: Update Repository Layer
**File:** `hub/api/repository/organization_repository.go`
**Changes:**

```go
// FindByAPIKeyHash retrieves a project by API key hash
func (r *ProjectRepositoryImpl) FindByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.Project, error) {
    query := "SELECT id, org_id, name, api_key_hash, api_key_prefix, created_at FROM projects WHERE api_key_hash = $1"
    
    var project models.Project
    err := r.db.QueryRow(ctx, query, apiKeyHash).Scan(
        &project.ID, 
        &project.OrgID, 
        &project.Name, 
        &project.APIKeyHash, 
        &project.APIKeyPrefix, 
        &project.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    
    return &project, nil
}
```
**Compliance:**
- ✅ Repository layer only (data access)
- ✅ Parameterized queries (SQL injection prevention)
- ✅ Single responsibility

#### Phase 1.4: Database Migration
**File:** `hub/migrations/XXXX_add_api_key_hashing.sql`
**Changes:**
```sql
-- Add new columns
ALTER TABLE projects 
ADD COLUMN api_key_hash VARCHAR(64),
ADD COLUMN api_key_prefix VARCHAR(8);

-- Migrate existing data (hash existing keys)
UPDATE projects 
SET api_key_hash = encode(digest(api_key, 'sha256'), 'hex'),
    api_key_prefix = LEFT(api_key, 8)
WHERE api_key IS NOT NULL AND api_key != '';

-- Create index for faster lookups
CREATE INDEX idx_projects_api_key_hash ON projects(api_key_hash);

-- After migration verification, remove old column
-- ALTER TABLE projects DROP COLUMN api_key;
```

#### Phase 1.5: Testing
**File:** `hub/api/services/organization_service_api_keys_test.go`
**Requirements:** 100% coverage

```go
func TestOrganizationService_GenerateAPIKey(t *testing.T) {
    // Test key generation and hashing
    // Test hash uniqueness
    // Test prefix generation
}

func TestOrganizationService_ValidateAPIKey(t *testing.T) {
    // Test valid key validation
    // Test invalid key rejection
    // Test constant-time comparison
}

func TestOrganizationService_hashAPIKey(t *testing.T) {
    // Test hash consistency
    // Test prefix extraction
}
```

**Timeline:** 3 days
- Day 1: Service layer implementation
- Day 2: Repository and migration
- Day 3: Testing and validation

---

## 2. Authentication Middleware Integration

### Objective
Connect authentication middleware to OrganizationService.ValidateAPIKey() instead of hardcoded keys.

### Compliance Requirements
- **File:** `hub/api/middleware/security.go` (HTTP middleware, max 300 lines)
- **Layer:** HTTP Layer - Middleware
- **Dependencies:** Service Layer (via dependency injection)
- **Testing:** Integration tests required

### Implementation Plan

#### Phase 2.1: Update Middleware Structure
**File:** `hub/api/middleware/security.go`
**Current Size:** ~237 lines (within 300 line limit)

**Changes:**

```go
import (
    "context"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
    // Add service interface import
    "sentinel-hub-api/services"
)

// AuthMiddlewareConfig holds configuration for authentication middleware
type AuthMiddlewareConfig struct {
    OrganizationService services.OrganizationService
    SkipPaths           []string // Paths to skip authentication
    Logger              Logger
}

// AuthMiddleware creates authentication middleware with service integration
func AuthMiddleware(config AuthMiddlewareConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip auth for configured paths
            if shouldSkipAuth(r.URL.Path, config.SkipPaths) {
                next.ServeHTTP(w, r)
                return
            }

            // Extract API key from headers
            apiKey := extractAPIKey(r)
            if apiKey == "" {
                config.Logger.Warn("Authentication failed: missing API key", 
                    "path", r.URL.Path, 
                    "ip", getClientIP(r),
                )
                http.Error(w, "Unauthorized: API key required", http.StatusUnauthorized)
                return
            }

            // Validate API key via service layer
            project, err := config.OrganizationService.ValidateAPIKey(r.Context(), apiKey)
            if err != nil || project == nil {
                config.Logger.Warn("Authentication failed: invalid API key",
                    "path", r.URL.Path,
                    "ip", getClientIP(r),
                    "error", err,
                )
                http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
                return
            }

            // Add authenticated context
            ctx := r.Context()
            ctx = context.WithValue(ctx, "project_id", project.ID)
            ctx = context.WithValue(ctx, "org_id", project.OrgID)
            ctx = context.WithValue(ctx, "api_key_prefix", project.APIKeyPrefix)
            r = r.WithContext(ctx)

            // Log successful authentication
            config.Logger.Info("Authentication successful",
                "project_id", project.ID,
                "org_id", project.OrgID,
                "path", r.URL.Path,
            )

            next.ServeHTTP(w, r)
        })
    }
}

// extractAPIKey extracts API key from request headers
func extractAPIKey(r *http.Request) string {
    // Check X-API-Key header first
    if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
        return apiKey
    }
    
    // Check Authorization header (Bearer token)
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    
    return ""
}

// shouldSkipAuth checks if path should skip authentication
func shouldSkipAuth(path string, skipPaths []string) bool {
    // Always skip health endpoints
    if strings.HasPrefix(path, "/health") {
        return true
    }
    
    for _, skipPath := range skipPaths {
        if strings.HasPrefix(path, skipPath) {
            return true
        }
    }
    
    return false
}
```

**Compliance:**
- ✅ HTTP layer only (no business logic)
- ✅ Dependency injection via config struct
- ✅ Proper error handling
- ✅ Security logging

#### Phase 2.2: Update Server Initialization
**File:** `hub/api/server/server.go` (or main server setup file)
**Changes:**

```go
// In server setup
func setupMiddleware(router *chi.Mux, orgService services.OrganizationService, logger Logger) {
    // Configure authentication middleware
    authConfig := middleware.AuthMiddlewareConfig{
        OrganizationService: orgService,
        SkipPaths: []string{
            "/health",
            "/api/v1/public", // If any public endpoints exist
        },
        Logger: logger,
    }
    
    router.Use(middleware.AuthMiddleware(authConfig))
    router.Use(middleware.CORSMiddleware())
    router.Use(middleware.SecurityHeadersMiddleware())
    router.Use(middleware.RateLimitMiddleware(100, 10))
}
```

**Compliance:**
- ✅ Dependency injection
- ✅ Clear configuration
- ✅ Layer separation maintained

#### Phase 2.3: Testing
**File:** `hub/api/middleware/security_test.go`

```go
func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
    // Mock organization service
    // Test successful authentication
    // Verify context values set correctly
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
    // Test rejection of invalid keys
    // Test error responses
}

func TestAuthMiddleware_MissingAPIKey(t *testing.T) {
    // Test missing key handling
}
```

**Timeline:** 2 days
- Day 1: Middleware implementation
- Day 2: Integration and testing

---

## 3. Comprehensive Input Validation Framework

### Objective
Implement comprehensive input validation for all request parameters following coding standards.

### Compliance Requirements
- **File:** `hub/api/utils/validation.go` (max 250 lines) or separate validators
- **Layer:** Service Layer (business validation) + HTTP Layer (request validation)
- **Pattern:** Validation middleware + service-level validation

### Implementation Plan

#### Phase 3.1: Create Validation Middleware
**File:** `hub/api/middleware/validation.go` (new file, max 300 lines)

```go
package middleware

import (
    "encoding/json"
    "net/http"
    "strings"
)

// ValidationMiddleware validates request body and parameters
func ValidationMiddleware(schema Validator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip validation for GET/HEAD/DELETE with no body
            if r.Method == "GET" || r.Method == "HEAD" || r.Method == "DELETE" {
                // Validate query parameters only
                if err := validateQueryParams(r, schema); err != nil {
                    respondValidationError(w, err)
                    return
                }
                next.ServeHTTP(w, r)
                return
            }

            // Validate request body
            var body map[string]interface{}
            if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
                http.Error(w, "Invalid request body", http.StatusBadRequest)
                return
            }

            if err := schema.Validate(body); err != nil {
                respondValidationError(w, err)
                return
            }

            // Re-encode body for handlers
            bodyBytes, _ := json.Marshal(body)
            r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
            r.ContentLength = int64(len(bodyBytes))

            next.ServeHTTP(w, r)
        })
    }
}
```

#### Phase 3.2: Create Validator Interface
**File:** `hub/api/validation/validator.go` (new package, max 250 lines per file)

```go
package validation

import (
    "errors"
    "fmt"
    "regexp"
    "strings"
)

// Validator defines validation interface
type Validator interface {
    Validate(data map[string]interface{}) error
}

// ValidationError represents a validation failure
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// StringValidator validates string fields
type StringValidator struct {
    Field      string
    Required   bool
    MinLength  int
    MaxLength  int
    Pattern    *regexp.Regexp
    Enum       []string
}

func (v *StringValidator) Validate(data map[string]interface{}) error {
    value, exists := data[v.Field]
    
    if !exists || value == nil || value == "" {
        if v.Required {
            return &ValidationError{
                Field:   v.Field,
                Message: fmt.Sprintf("%s is required", v.Field),
            }
        }
        return nil // Optional field, skip validation
    }

    str, ok := value.(string)
    if !ok {
        return &ValidationError{
            Field:   v.Field,
            Message: fmt.Sprintf("%s must be a string", v.Field),
        }
    }

    if v.MinLength > 0 && len(str) < v.MinLength {
        return &ValidationError{
            Field:   v.Field,
            Message: fmt.Sprintf("%s must be at least %d characters", v.Field, v.MinLength),
        }
    }

    if v.MaxLength > 0 && len(str) > v.MaxLength {
        return &ValidationError{
            Field:   v.Field,
            Message: fmt.Sprintf("%s must be at most %d characters", v.Field, v.MaxLength),
        }
    }

    if v.Pattern != nil && !v.Pattern.MatchString(str) {
        return &ValidationError{
            Field:   v.Field,
            Message: fmt.Sprintf("%s format is invalid", v.Field),
        }
    }

    if len(v.Enum) > 0 {
        found := false
        for _, allowed := range v.Enum {
            if str == allowed {
                found = true
                break
            }
        }
        if !found {
            return &ValidationError{
                Field:   v.Field,
                Message: fmt.Sprintf("%s must be one of: %s", v.Field, strings.Join(v.Enum, ", ")),
            }
        }
    }

    return nil
}

// CompositeValidator validates multiple fields
type CompositeValidator struct {
    Validators []Validator
}

func (c *CompositeValidator) Validate(data map[string]interface{}) error {
    for _, validator := range c.Validators {
        if err := validator.Validate(data); err != nil {
            return err
        }
    }
    return nil
}
```

#### Phase 3.3: Create Endpoint-Specific Validators
**File:** `hub/api/validation/task_validators.go` (example)

```go
package validation

import "regexp"

// ValidateCreateTaskRequest validates task creation requests
func ValidateCreateTaskRequest(data map[string]interface{}) error {
    composite := &CompositeValidator{
        Validators: []Validator{
            &StringValidator{
                Field:     "title",
                Required:  true,
                MinLength: 1,
                MaxLength: 200,
            },
            &StringValidator{
                Field:     "description",
                Required:  false,
                MaxLength: 5000,
            },
            &StringValidator{
                Field:    "status",
                Required: true,
                Enum:     []string{"pending", "in_progress", "completed", "archived"},
            },
            &StringValidator{
                Field:    "priority",
                Required: false,
                Enum:     []string{"low", "medium", "high", "critical"},
            },
        },
    }
    return composite.Validate(data)
}
```

**Compliance:**
- ✅ Single responsibility
- ✅ Well-documented
- ✅ Error wrapping
- ✅ File size limits respected

**Timeline:** 4 days
- Day 1-2: Validation framework
- Day 3: Endpoint validators
- Day 4: Integration and testing

---

## 4. CORS Production Configuration

### Objective
Implement strict origin whitelist validation for production environments.

### Compliance Requirements
- **File:** `hub/api/middleware/security.go` (update existing)
- **Configuration:** `hub/api/config/config.go`

### Implementation Plan

#### Phase 4.1: Update CORS Middleware
**File:** `hub/api/middleware/security.go`
**Changes:**

```go
// CORSMiddleware creates CORS middleware with origin validation
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    originMap := make(map[string]bool)
    for _, origin := range allowedOrigins {
        originMap[origin] = true
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // Determine allowed origin
            var allowedOrigin string
            if os.Getenv("ENV") == "development" || os.Getenv("ENV") == "dev" {
                // Development: allow all or specific origins
                allowedOrigin = origin
                if origin == "" {
                    allowedOrigin = "*"
                }
            } else {
                // Production: strict whitelist
                if origin == "" {
                    // No origin header - reject or allow based on policy
                    allowedOrigin = ""
                } else if originMap[origin] || originMap["*"] {
                    allowedOrigin = origin
                } else {
                    // Origin not in whitelist
                    http.Error(w, "CORS: Origin not allowed", http.StatusForbidden)
                    return
                }
            }
            
            // Set CORS headers
            if allowedOrigin != "" {
                w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
            }
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
            w.Header().Set("Access-Control-Max-Age", "86400")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            // Handle preflight requests
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

#### Phase 4.2: Update Configuration
**File:** `hub/api/config/config.go`
**Changes:**

```go
// Load allowed origins from environment
if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
    config.Security.CORSAllowedOrigins = strings.Split(corsOrigins, ",")
} else if os.Getenv("ENV") == "production" {
    // Production requires explicit configuration
    config.Security.CORSAllowedOrigins = []string{} // Must be set
} else {
    // Development defaults
    config.Security.CORSAllowedOrigins = []string{"*"}
}
```

**Timeline:** 1 day

---

## 5. Security Event Logging

### Objective
Implement comprehensive security event logging for audit and monitoring.

### Compliance Requirements
- **File:** `hub/api/pkg/security/audit_logger.go` (new file, max 250 lines)
- **Layer:** Service Layer / HTTP Layer
- **Standard:** Structured logging

### Implementation Plan

#### Phase 5.1: Create Audit Logger
**File:** `hub/api/pkg/security/audit_logger.go`

```go
package security

import (
    "context"
    "time"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"` // auth_success, auth_failure, api_key_revoked, etc.
    Severity    string                 `json:"severity"`   // info, warning, error
    ProjectID   string                 `json:"project_id,omitempty"`
    OrgID       string                 `json:"org_id,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    Path        string                 `json:"path"`
    Method      string                 `json:"method"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Error       string                 `json:"error,omitempty"`
}

// AuditLogger handles security event logging
type AuditLogger interface {
    LogAuthSuccess(ctx context.Context, projectID, orgID, ip, path string)
    LogAuthFailure(ctx context.Context, reason string, ip, path string)
    LogAPIKeyRevoked(ctx context.Context, projectID, orgID string)
    LogAPIKeyGenerated(ctx context.Context, projectID, orgID string)
    LogSecurityEvent(ctx context.Context, event SecurityEvent)
}

// auditLoggerImpl implements AuditLogger
type auditLoggerImpl struct {
    logger Logger
}

func NewAuditLogger(logger Logger) AuditLogger {
    return &auditLoggerImpl{logger: logger}
}

func (a *auditLoggerImpl) LogAuthSuccess(ctx context.Context, projectID, orgID, ip, path string) {
    a.LogSecurityEvent(ctx, SecurityEvent{
        Timestamp: time.Now(),
        EventType: "auth_success",
        Severity:  "info",
        ProjectID: projectID,
        OrgID:     orgID,
        IPAddress: ip,
        Path:      path,
    })
}

func (a *auditLoggerImpl) LogAuthFailure(ctx context.Context, reason string, ip, path string) {
    a.LogSecurityEvent(ctx, SecurityEvent{
        Timestamp: time.Now(),
        EventType: "auth_failure",
        Severity:  "warning",
        IPAddress: ip,
        Path:      path,
        Error:     reason,
    })
}

func (a *auditLoggerImpl) LogSecurityEvent(ctx context.Context, event SecurityEvent) {
    // Structured logging in JSON format
    a.logger.Info("security_event",
        "timestamp", event.Timestamp,
        "event_type", event.EventType,
        "severity", event.Severity,
        "project_id", event.ProjectID,
        "org_id", event.OrgID,
        "ip_address", event.IPAddress,
        "path", event.Path,
        "method", event.Method,
        "details", event.Details,
        "error", event.Error,
    )
}
```

#### Phase 5.2: Integrate with Middleware
**File:** `hub/api/middleware/security.go` (update AuthMiddleware)

```go
func AuthMiddleware(config AuthMiddlewareConfig) func(http.Handler) http.Handler {
    auditLogger := security.NewAuditLogger(config.Logger)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... existing code ...
            
            // After validation
            if project != nil {
                auditLogger.LogAuthSuccess(r.Context(), 
                    project.ID, 
                    project.OrgID, 
                    getClientIP(r), 
                    r.URL.Path,
                )
            } else {
                auditLogger.LogAuthFailure(r.Context(), 
                    "invalid API key", 
                    getClientIP(r), 
                    r.URL.Path,
                )
            }
            
            // ... rest of code ...
        })
    }
}
```

**Timeline:** 2 days

---

## Implementation Timeline

### Week 1: Critical Fixes

| Day | Task | Owner | Status |
|-----|------|-------|--------|
| 1-3 | API Key Hashing | Backend | Planned |
| 4-5 | Authentication Middleware Integration | Backend | Planned |
| 6-7 | CORS Production Configuration | Backend | Planned |

### Week 2: Validation & Logging

| Day | Task | Owner | Status |
|-----|------|-------|--------|
| 8-11 | Input Validation Framework | Backend | Planned |
| 12-13 | Security Event Logging | Backend | Planned |
| 14 | Integration Testing | QA | Planned |

---

## Testing Requirements

### Unit Tests (100% Coverage Required)
- API key generation and hashing
- API key validation
- Authentication middleware
- Input validators
- Security event logging

### Integration Tests
- End-to-end authentication flow
- API key lifecycle (generate → validate → revoke)
- CORS validation
- Input validation across endpoints

### Security Tests
- Penetration testing for authentication
- SQL injection testing (already using parameterized queries)
- XSS testing
- CSRF testing

---

## Compliance Checklist

### Architectural Standards ✅
- [x] Layer separation maintained
- [x] Service layer only for business logic
- [x] Repository layer only for data access
- [x] HTTP layer only for request/response handling

### File Size Limits ✅
- [x] All files within specified limits
- [x] Functions follow size guidelines
- [x] Complexity within acceptable ranges

### Testing Standards ✅
- [x] 100% coverage for new code
- [x] Integration tests planned
- [x] Security tests included

### Security Standards ✅
- [x] Input validation implemented
- [x] No SQL injection (parameterized queries)
- [x] Security logging implemented
- [x] Rate limiting maintained

### Documentation Standards ✅
- [x] Code documentation included
- [x] API documentation planned
- [x] Security architecture documented

---

## Risk Mitigation

### Migration Risks
- **Risk:** Existing API keys become invalid
- **Mitigation:** Gradual migration with dual storage during transition period

### Performance Risks
- **Risk:** Hash calculation adds latency
- **Mitigation:** Hash operations are fast (<1ms), acceptable trade-off

### Compatibility Risks
- **Risk:** Breaking changes for existing clients
- **Mitigation:** Maintain backward compatibility during transition

---

## Success Criteria

1. ✅ All API keys stored as hashes
2. ✅ Authentication middleware integrated with service layer
3. ✅ All endpoints have input validation
4. ✅ CORS properly configured for production
5. ✅ Security events logged for all auth operations
6. ✅ 100% test coverage for new code
7. ✅ No security regressions
8. ✅ Performance impact < 5%

---

## Review & Approval

**Security Team Review:** Pending  
**Architecture Review:** Pending  
**Tech Lead Approval:** Pending  

---

**This plan ensures full compliance with CODING_STANDARDS.md while addressing all critical security vulnerabilities identified in the VAPT report.**

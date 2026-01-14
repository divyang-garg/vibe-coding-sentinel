# Sentinel Coding Standards & Development Guidelines

## Overview

This document establishes the coding standards, architectural patterns, and development processes that must be followed for all Sentinel project development. These standards ensure code quality, maintainability, and scalability.

**Status:** ENFORCED - All code changes must comply with these standards.

---

## 1. ARCHITECTURAL STANDARDS

### 1.1 Package Structure (ENFORCED)

```
sentinel/
├── cmd/                          # Application entry points
│   └── sentinel/                 # Main application
│       └── main.go              # Entry point only (< 50 lines)
├── internal/                     # Private application code
│   ├── api/                     # HTTP layer
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── middleware/           # HTTP middleware
│   │   ├── routes/              # Route definitions
│   │   └── server/              # Server setup
│   ├── services/                # Business logic layer
│   ├── models/                  # Data models & types
│   ├── repository/              # Data access layer
│   ├── config/                  # Configuration management
│   └── utils/                   # Shared utilities
├── pkg/                         # Public packages
├── docs/                        # Documentation
├── scripts/                     # Build/deployment scripts
└── tests/                       # Integration tests
```

### 1.2 Layer Separation (ENFORCED)

#### HTTP Layer (`internal/api/`)
- **Purpose:** HTTP request/response handling only
- **Responsibilities:**
  - Request parsing and validation
  - Response formatting
  - HTTP status codes
  - Middleware application
- **Restrictions:** No business logic, no database calls

#### Service Layer (`internal/services/`)
- **Purpose:** Business logic and domain rules
- **Responsibilities:**
  - Business rule validation
  - Domain logic execution
  - Transaction coordination
  - Error handling (business level)
- **Restrictions:** No HTTP concerns, no direct database access

#### Repository Layer (`internal/repository/`)
- **Purpose:** Data persistence and retrieval
- **Responsibilities:**
  - Database queries and commands
  - Data mapping (SQL ↔ Domain objects)
  - Connection management
  - Query optimization
- **Restrictions:** No business logic, no HTTP concerns

#### Model Layer (`internal/models/`)
- **Purpose:** Data structures and types
- **Responsibilities:**
  - Domain entity definitions
  - Data transfer objects
  - Value objects
  - Type definitions
- **Restrictions:** No behavior, no external dependencies

---

## 2. FILE SIZE LIMITS (ENFORCED)

| File Type | Max Lines | Max Functions | Max Complexity | Rationale |
|-----------|-----------|---------------|----------------|-----------|
| **Entry Points** (`main.go`) | 50 | 3 | 5 | Bootstrap only |
| **HTTP Handlers** | 300 | 10 | 8 | Single responsibility |
| **Business Services** | 400 | 15 | 10 | Focused business logic |
| **Repositories** | 350 | 12 | 8 | Data access patterns |
| **Data Models** | 200 | 0 | 0 | Pure data structures |
| **Utilities** | 250 | 8 | 6 | Helper functions |
| **Tests** | 500 | 15 | 12 | Comprehensive testing |

**Enforcement:** CI/CD pipeline will reject commits exceeding these limits.

---

## 3. FUNCTION DESIGN STANDARDS

### 3.1 Function Size & Complexity

```go
// ✅ GOOD: Single responsibility, clear purpose
func CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    if err := validateCreateUserRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    user := &User{
        ID:        generateID(),
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: time.Now(),
    }

    if err := s.userRepo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user: %w", err)
    }

    return user, nil
}

// ❌ BAD: Multiple responsibilities, too complex
func CreateUserAndSendEmail(email, name string) (*User, error) {
    // Validation, creation, email sending all mixed together
    // Too many concerns in one function
}
```

### 3.2 Parameter Limits

```go
// ✅ GOOD: Few, well-typed parameters
func CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error)

// ❌ BAD: Too many parameters
func CreateTask(ctx context.Context, title, description string, priority int,
                assigneeID, creatorID string, dueDate *time.Time, tags []string) (*Task, error)
```

### 3.3 Return Values

```go
// ✅ GOOD: Explicit error handling
func ProcessDocument(ctx context.Context, docID string) (*Document, error)

// ✅ GOOD: Multiple return values when appropriate
func ValidateUser(ctx context.Context, userID string) (bool, error)

// ❌ BAD: Using panics for expected errors
func ProcessDocument(docID string) *Document // Will panic on error
```

---

## 4. ERROR HANDLING STANDARDS

### 4.1 Error Wrapping (ENFORCED)

```go
// ✅ GOOD: Preserve error context
if err := validateInput(req); err != nil {
    return fmt.Errorf("failed to validate request: %w", err)
}

// ❌ BAD: Lose original error
if err := validateInput(req); err != nil {
    return fmt.Errorf("validation failed")
}
```

### 4.2 Structured Error Types

```go
// ✅ GOOD: Custom error types with context
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// Usage
return &ValidationError{
    Field:   "email",
    Value:   req.Email,
    Message: "invalid email format",
}
```

### 4.3 Logging Levels (ENFORCED)

```go
// DEBUG: Detailed diagnostic information
log.Debug("processing user request", "user_id", userID, "action", "login")

// INFO: Normal operational messages
log.Info("user authenticated successfully", "user_id", userID)

// WARN: Unexpected but recoverable events
log.Warn("external service timeout, using cached data", "service", "llm", "timeout", timeout)

// ERROR: Error conditions requiring attention
log.Error("database connection failed", "error", err, "attempt", attempt)
```

---

## 5. NAMING CONVENTIONS

### 5.1 Go Naming Standards

```go
// ✅ GOOD: Clear, descriptive names
type UserService interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
    GetUserByID(ctx context.Context, id string) (*User, error)
    UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
}

// ❌ BAD: Abbreviations and unclear names
type USvc interface {
    CreateU(ctx context.Context, req CreateUReq) (*U, error)
    GetUByID(ctx context.Context, id string) (*U, error)
    UpdateU(ctx context.Context, id string, req UpdateUReq) (*U, error)
}
```

### 5.2 Package Naming

```go
// ✅ GOOD: Clear package purposes
package handlers    // HTTP handlers
package services    // Business logic
package repository  // Data access
package models      // Data structures

// ❌ BAD: Unclear or generic names
package utils       // Too generic
package helpers     // Unclear purpose
package common      // Too vague
```

---

## 6. TESTING STANDARDS

### 6.1 Test Coverage Requirements (ENFORCED)

- **Minimum Coverage:** 80% overall
- **Critical Path:** 90% coverage for business logic
- **New Code:** 100% coverage required

### 6.2 Test Structure

```go
// ✅ GOOD: Clear test naming and structure
func TestUserService_CreateUser(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Given
        req := CreateUserRequest{Email: "test@example.com", Name: "Test User"}

        // When
        user, err := service.CreateUser(context.Background(), req)

        // Then
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "test@example.com", user.Email)
    })

    t.Run("validation_error", func(t *testing.T) {
        // Test error cases
        req := CreateUserRequest{Email: "", Name: "Test User"}
        user, err := service.CreateUser(context.Background(), req)

        assert.Error(t, err)
        assert.Nil(t, user)
    })
}
```

### 6.3 Mock Usage

```go
// ✅ GOOD: Proper mocking of dependencies
func TestUserService_GetUserByID(t *testing.T) {
    // Create mock repository
    mockRepo := &mocks.UserRepository{}
    service := NewUserService(mockRepo)

    // Setup expectations
    expectedUser := &User{ID: "123", Email: "test@example.com"}
    mockRepo.On("GetByID", mock.Anything, "123").Return(expectedUser, nil)

    // Test
    user, err := service.GetUserByID(context.Background(), "123")

    // Verify
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}
```

---

## 7. DEPENDENCY INJECTION

### 7.1 Constructor Injection (ENFORCED)

```go
// ✅ GOOD: Clear dependencies, testable
type UserService struct {
    repo    UserRepository
    logger  Logger
    metrics MetricsClient
}

func NewUserService(repo UserRepository, logger Logger, metrics MetricsClient) *UserService {
    return &UserService{
        repo:    repo,
        logger:  logger,
        metrics: metrics,
    }
}

// ❌ BAD: Hidden dependencies, hard to test
type UserService struct{}

func NewUserService() *UserService {
    return &UserService{}
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    db := getGlobalDatabase() // Hidden dependency
    return createUserInDB(db, req)
}
```

### 7.2 Interface-Based Design

```go
// ✅ GOOD: Interface-based design
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
}

// Implementation
type postgresUserRepository struct {
    db *sql.DB
}

func (r *postgresUserRepository) Save(ctx context.Context, user *User) error {
    // Implementation
}
```

---

## 8. CODE ORGANIZATION PATTERNS

### 8.1 Handler Organization

```go
// handlers/user_handler.go
type UserHandler struct {
    service UserService
    logger  Logger
}

func NewUserHandler(service UserService, logger Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // HTTP handling only - delegate to service
    req, err := parseCreateUserRequest(r)
    if err != nil {
        h.respondError(w, err)
        return
    }

    user, err := h.service.CreateUser(r.Context(), req)
    if err != nil {
        h.respondError(w, err)
        return
    }

    h.respondJSON(w, http.StatusCreated, user)
}
```

### 8.2 Service Organization

```go
// services/user_service.go
type UserService struct {
    repo       UserRepository
    validator  UserValidator
    publisher  EventPublisher
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Business logic only
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    user := s.createUserFromRequest(req)

    if err := s.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user: %w", err)
    }

    s.publisher.Publish(UserCreatedEvent{User: user})

    return user, nil
}
```

---

## 9. DEVELOPMENT PROCESS

### 9.1 Commit Standards (ENFORCED)

```
feat: add user authentication service
fix: resolve memory leak in document parser
refactor: extract common validation logic
docs: update API documentation
test: add integration tests for user service
style: format code according to standards
chore: update dependencies
```

### 9.2 Code Review Requirements

#### Mandatory Reviews:
- **All Changes:** Require 1+ reviewer approval
- **Architecture Changes:** Require tech lead approval
- **Security Changes:** Require security team approval
- **Database Changes:** Require DBA approval

#### Review Checklist:
- [ ] Code follows architectural standards
- [ ] Functions have single responsibility
- [ ] Error handling is appropriate
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No linting errors
- [ ] Performance impact assessed

### 9.3 CI/CD Pipeline Requirements

#### Quality Gates:
1. **Compilation:** Must compile without errors
2. **Linting:** Must pass golangci-lint
3. **Testing:** Must pass all tests (unit, integration)
4. **Coverage:** Must maintain 80%+ coverage
5. **Security:** Must pass security scan
6. **Performance:** Must not regress performance benchmarks

---

## 10. PERFORMANCE STANDARDS

### 10.1 Response Time Requirements

| Operation Type | Target Response Time | Max Response Time |
|----------------|---------------------|-------------------|
| API Health Check | < 50ms | < 200ms |
| Simple CRUD | < 100ms | < 500ms |
| Complex Queries | < 500ms | < 2s |
| File Processing | < 5s | < 30s |
| Report Generation | < 10s | < 60s |

### 10.2 Resource Usage Limits

- **Memory Usage:** < 512MB per service instance
- **CPU Usage:** < 80% sustained load
- **Database Connections:** Max 20 per service
- **Concurrent Requests:** Support 1000+ concurrent users

---

## 11. SECURITY STANDARDS

### 11.1 Input Validation (ENFORCED)

```go
// ✅ GOOD: Comprehensive validation
func validateCreateUserRequest(req CreateUserRequest) error {
    if req.Email == "" {
        return &ValidationError{Field: "email", Message: "email is required"}
    }
    if !isValidEmail(req.Email) {
        return &ValidationError{Field: "email", Message: "invalid email format"}
    }
    if len(req.Name) < 2 || len(req.Name) > 100 {
        return &ValidationError{Field: "name", Message: "name must be 2-100 characters"}
    }
    return nil
}
```

### 11.2 Secure Coding Practices

- **No SQL Injection:** Use parameterized queries
- **XSS Prevention:** Sanitize all user input
- **CSRF Protection:** Implement anti-CSRF tokens
- **Rate Limiting:** Apply to all public endpoints
- **Logging:** Never log sensitive data

---

## 12. DOCUMENTATION STANDARDS

### 12.1 Code Documentation

```go
// Package handlers provides HTTP request handlers for the Sentinel API.
//
// This package contains all HTTP handlers organized by domain area.
// Each handler follows the single responsibility principle and delegates
// business logic to the appropriate service layer.
package handlers

// UserHandler handles user-related HTTP requests.
// It provides endpoints for user management including CRUD operations
// and user authentication.
type UserHandler struct {
    service UserService
    logger  Logger
}

// NewUserHandler creates a new user handler with dependencies.
func NewUserHandler(service UserService, logger Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}

// CreateUser handles POST /api/v1/users
// Creates a new user account with validation and business rule enforcement.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Implementation with inline comments for complex logic
}
```

### 12.2 API Documentation

All API endpoints must be documented with:
- HTTP method and path
- Request/response schemas
- Authentication requirements
- Error responses
- Usage examples

---

## ENFORCEMENT & COMPLIANCE

### Automated Enforcement:
- **CI/CD Pipeline:** Rejects non-compliant code
- **Linting:** golangci-lint with custom rules
- **Testing:** Coverage and quality gates
- **Security:** Automated security scanning

### Manual Enforcement:
- **Code Reviews:** Mandatory for all changes
- **Architecture Reviews:** For major changes
- **Security Reviews:** For security-sensitive changes

### Compliance Reporting:
- **Weekly Reports:** Code quality metrics
- **Monthly Reviews:** Standards compliance assessment
- **Quarterly Audits:** Comprehensive code quality review

---

## MIGRATION & ADOPTION

### Phase 1: Standards Establishment (Week 1)
- Publish coding standards document
- Set up CI/CD quality gates
- Train development team

### Phase 2: Gradual Adoption (Weeks 2-4)
- Apply standards to new code only
- Refactor critical files as needed
- Establish code review processes

### Phase 3: Full Compliance (Weeks 5-8)
- Refactor all legacy code
- Achieve 100% standards compliance
- Implement monitoring and reporting

---

## EXCEPTIONS & WAIVERS

### Exception Process:
1. **Technical Justification:** Must provide clear technical reasons
2. **Risk Assessment:** Document potential risks and mitigations
3. **Approval Required:** Tech lead + architect approval
4. **Time-Limited:** Maximum 30 days, then must comply or remove
5. **Documentation:** All exceptions must be documented

### Prohibited Exceptions:
- File size limits
- Architectural layer violations
- Security standards
- Testing requirements

---

**This document is the authoritative source for Sentinel development standards. All team members are required to comply with these standards. Questions or clarifications should be raised through the architecture review process.**


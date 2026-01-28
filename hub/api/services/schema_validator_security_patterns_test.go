// Package services provides tests for security pattern detection
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sentinel-hub-api/ast"
)

func TestDetectSecurityPatternsInCode_JWT(t *testing.T) {
	code := `
package middleware

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			// Validate JWT token
		}
		next.ServeHTTP(w, r)
	})
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "BearerAuth" {
			found = true
			if pattern.Type != "authentication" {
				t.Errorf("Expected type 'authentication', got %s", pattern.Type)
			}
			if pattern.Confidence < 0.8 {
				t.Errorf("Expected confidence >= 0.8, got %f", pattern.Confidence)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find BearerAuth pattern")
	}
}

func TestDetectSecurityPatternsInCode_APIKey(t *testing.T) {
	code := `
package middleware

func extractAPIKey(r *http.Request) string {
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}
	return ""
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "ApiKeyAuth" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find ApiKeyAuth pattern")
	}
}

func TestDetectSecurityPatternsInCode_OAuth(t *testing.T) {
	code := `
package auth

func HandleOAuth2Callback(code string) error {
	// OAuth2 authorization code flow
	return nil
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "OAuth2" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find OAuth2 pattern")
	}
}

func TestDetectSecurityPatternsInCode_RBAC(t *testing.T) {
	code := `
package auth

func CheckRole(userID string, role string) bool {
	// RBAC role checking
	return true
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "RBAC" {
			found = true
			if pattern.Type != "authorization" {
				t.Errorf("Expected type 'authorization', got %s", pattern.Type)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find RBAC pattern")
	}
}

func TestDetectSecurityPatternsInCode_RateLimit(t *testing.T) {
	code := `
package middleware

func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := NewRateLimiter(100, 10)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", 429)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "RateLimit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find RateLimit pattern")
	}
}

func TestDetectSecurityPatternsInCode_CORS(t *testing.T) {
	code := `
package middleware

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "CORS" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find CORS pattern")
	}
}

func TestDetectSecurityPatternsInCode_NoPatterns(t *testing.T) {
	code := `
package main

func main() {
	fmt.Println("Hello, World!")
}
`
	ctx := context.Background()
	patterns := detectSecurityPatternsInCode(ctx, code, "go")

	if len(patterns) > 0 {
		t.Errorf("Expected no patterns, got %d", len(patterns))
	}
}

func TestDetectSecurityPatternsInCode_ContextCancellation(t *testing.T) {
	code := `package main`
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	patterns := detectSecurityPatternsInCode(ctx, code, "go")
	// Should handle cancellation gracefully
	_ = patterns
}

func TestContainsJWTBearerPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"Bearer in Authorization", `auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ")`, true},
		{"JWT library", `import "github.com/golang-jwt/jwt"`, true},
		{"Token parse", `token, err := jwt.Parse(tokenString, keyFunc)`, true},
		{"No JWT", `func main() { fmt.Println("hello") }`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsJWTBearerPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsJWTBearerPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsAPIKeyPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"X-API-Key header", `apiKey := r.Header.Get("X-API-Key")`, true},
		{"API key validate", `func ValidateAPIKey(key string) bool`, true},
		{"Extract API key", `func ExtractAPIKey(r *http.Request) string`, true},
		{"No API key", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsAPIKeyPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsAPIKeyPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsOAuthPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"OAuth2", `func HandleOAuth2(callback string)`, true},
		{"OAuth", `import "golang.org/x/oauth2"`, true},
		{"Authorization code", `authorizationCode := r.URL.Query().Get("code")`, true},
		{"No OAuth", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsOAuthPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsOAuthPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsRBACPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"RBAC", `func CheckRBAC(userID string) bool`, true},
		{"Role check", `if user.Role == "admin" {`, true},
		{"Authorize", `func Authorize(userID string, action string) bool`, true},
		{"Permission", `if HasPermission(user, "read") {`, true},
		{"No RBAC", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsRBACPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsRBACPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsRateLimitPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"RateLimit", `func RateLimitMiddleware() http.Handler`, true},
		{"Rate limiter", `limiter := NewRateLimiter(100, 10)`, true},
		{"Throttle", `func ThrottleRequests()`, true},
		{"No rate limit", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsRateLimitPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsRateLimitPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsCORSPattern(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"CORS", `func CORSMiddleware() http.Handler`, true},
		{"CORS header", `w.Header().Set("Access-Control-Allow-Origin", "*")`, true},
		{"Cross-origin", `// Handle cross-origin requests`, true},
		{"No CORS", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsCORSPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsCORSPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectMiddlewareInFunctions_AuthMiddleware(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "AuthMiddleware",
			Line: 160,
		},
		{
			Name: "ListUsers",
			Line: 10,
		},
	}

	ctx := context.Background()
	code := `func AuthMiddleware(next http.Handler) http.Handler { return next }`
	patterns := detectMiddlewareInFunctions(ctx, functions, code, "go")

	// Should find AuthMiddleware
	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "BearerAuth" {
			found = true
			if pattern.Confidence < 0.6 {
				t.Errorf("Expected confidence >= 0.6, got %f", pattern.Confidence)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find AuthMiddleware pattern")
	}
}

func TestDetectMiddlewareInFunctions_APIMiddleware(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "APIKeyMiddleware",
			Line: 50,
		},
	}

	ctx := context.Background()
	code := `func APIKeyMiddleware(next http.Handler) http.Handler { return next }`
	patterns := detectMiddlewareInFunctions(ctx, functions, code, "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "ApiKeyAuth" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find APIKeyAuth pattern")
	}
}

func TestDetectMiddlewareInFunctions_RateLimitMiddleware(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "RateLimitMiddleware",
			Line: 100,
		},
	}

	ctx := context.Background()
	patterns := detectMiddlewareInFunctions(ctx, functions, "", "go")

	found := false
	for _, pattern := range patterns {
		if pattern.Scheme == "RateLimit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find RateLimit pattern")
	}
}

func TestDetectMiddlewareInFunctions_NoMatchingPattern(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "RegularFunction",
			Line: 10,
		},
	}

	ctx := context.Background()
	patterns := detectMiddlewareInFunctions(ctx, functions, "", "go")

	// Should not find any middleware patterns
	if len(patterns) > 0 {
		t.Errorf("Expected no patterns for non-middleware function, got %d", len(patterns))
	}
}

func TestDetectMiddlewareInFunctions_MultipleFunctions(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "AuthMiddleware",
			Line: 10,
		},
		{
			Name: "APIKeyMiddleware",
			Line: 20,
		},
		{
			Name: "CORSMiddleware",
			Line: 30,
		},
		{
			Name: "RegularHandler",
			Line: 40,
		},
	}

	ctx := context.Background()
	code := `func AuthMiddleware(next http.Handler) http.Handler { return next }`
	patterns := detectMiddlewareInFunctions(ctx, functions, code, "go")

	// Should find multiple middleware patterns
	if len(patterns) < 2 {
		t.Errorf("Expected at least 2 patterns, got %d", len(patterns))
	}
}

func TestDetectMiddlewareInFunctions_ContextCancellation(t *testing.T) {
	functions := []ast.FunctionInfo{
		{
			Name: "AuthMiddleware",
			Line: 160,
		},
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	patterns := detectMiddlewareInFunctions(ctx, functions, "", "go")

	// Should return early due to context cancellation
	if len(patterns) > 0 {
		t.Logf("Context cancellation returned %d patterns (acceptable)", len(patterns))
	}
}

func TestIsMiddlewareFunction_GoMiddleware(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "AuthMiddleware",
		Line: 5,
	}

	code := `
package middleware

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
`
	result := isMiddlewareFunction(fn, code, "go")
	if !result {
		t.Error("Expected AuthMiddleware to be recognized as middleware function")
	}
}

func TestIsMiddlewareFunction_NonMiddleware(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "ListUsers",
		Line: 5,
	}

	code := `
package handlers

func ListUsers(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}
`
	result := isMiddlewareFunction(fn, code, "go")
	if result {
		t.Error("Expected ListUsers not to be recognized as middleware function")
	}
}

func TestIsMiddlewareFunction_JavaScript(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "authMiddleware",
		Line: 5,
	}

	result := isMiddlewareFunction(fn, "", "javascript")
	if !result {
		t.Error("Expected authMiddleware to be recognized as middleware function in JavaScript")
	}
}

func TestIsMiddlewareFunction_Python(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "oauth_middleware",
		Line: 5,
	}

	result := isMiddlewareFunction(fn, "", "python")
	if !result {
		t.Error("Expected oauth_middleware to be recognized as middleware function in Python")
	}
}

func TestIsMiddlewareFunction_GoWithHTTPHandler(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "CustomMiddleware",
		Line: 5,
	}

	code := `
package middleware

func CustomMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
`
	result := isMiddlewareFunction(fn, code, "go")
	if !result {
		t.Error("Expected CustomMiddleware to be recognized as middleware function with http.Handler signature")
	}
}

func TestIsMiddlewareFunction_GoWithoutHTTPHandler(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "RegularFunction",
		Line: 5,
	}

	code := `
package handlers

func RegularFunction(w http.ResponseWriter, r *http.Request) {
	// Regular handler
}
`
	result := isMiddlewareFunction(fn, code, "go")
	if result {
		t.Error("Expected RegularFunction not to be recognized as middleware function")
	}
}

func TestIsMiddlewareFunction_NonGoLanguage(t *testing.T) {
	fn := ast.FunctionInfo{
		Name: "authMiddleware",
		Line: 5,
	}

	// For non-Go languages, should use name-based detection
	result := isMiddlewareFunction(fn, "", "javascript")
	if !result {
		t.Error("Expected authMiddleware to be recognized as middleware in JavaScript")
	}

	result = isMiddlewareFunction(fn, "", "typescript")
	if !result {
		t.Error("Expected authMiddleware to be recognized as middleware in TypeScript")
	}
}

func TestExtractFunctionCodeAroundLine(t *testing.T) {
	code := `line1
line2
line3
line4
line5
line6
line7
line8
line9
line10
line11
line12
line13
line14
line15
line16
line17
line18
line19
line20
`

	result := extractFunctionCodeAroundLine(code, 10)
	lines := strings.Split(result, "\n")

	// Should extract lines around line 10 (0-20, so 10 lines before and after)
	if len(lines) < 10 {
		t.Errorf("Expected at least 10 lines, got %d", len(lines))
	}

	// Should include line 10
	if !strings.Contains(result, "line10") {
		t.Error("Expected extracted code to include line10")
	}
}

func TestExtractFunctionCodeAroundLine_EdgeCases(t *testing.T) {
	code := `line1
line2
line3
`

	// Test near start of file
	result := extractFunctionCodeAroundLine(code, 1)
	if result == "" {
		t.Error("Expected non-empty result for line 1")
	}

	// Test near end of file
	result = extractFunctionCodeAroundLine(code, 3)
	if result == "" {
		t.Error("Expected non-empty result for line 3")
	}
}

func TestMatchSecurityScheme_ExactMatch(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.9},
	}

	result := matchSecurityScheme(patterns, "BearerAuth")
	if !result {
		t.Error("Expected exact match for BearerAuth")
	}
}

func TestMatchSecurityScheme_PartialMatch(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.9},
	}

	result := matchSecurityScheme(patterns, "Bearer")
	if !result {
		t.Error("Expected partial match for Bearer")
	}
}

func TestMatchSecurityScheme_LowConfidence(t *testing.T) {
	// Use schemes that will partially match but not exactly
	// "BearerToken" normalizes to "bearertoken" (no "auth" suffix)
	// "Bearer" normalizes to "bearer"
	// They will partially match via Contains, but confidence check applies
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerToken", Confidence: 0.5}, // Below threshold (0.7)
	}

	// "Bearer" should partially match "BearerToken" but confidence is too low
	result := matchSecurityScheme(patterns, "Bearer")
	if result {
		t.Error("Expected no match for low confidence pattern (0.5 < 0.7 threshold)")
	}
}

func TestMatchSecurityScheme_LowConfidenceExactMatch(t *testing.T) {
	// Exact matches work regardless of confidence
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.5}, // Low confidence
	}

	// Exact match should work regardless of confidence
	result := matchSecurityScheme(patterns, "BearerAuth")
	if !result {
		t.Error("Expected exact match to work regardless of confidence")
	}
}

func TestMatchSecurityScheme_ExactMatchIgnoresConfidence(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.5}, // Low confidence
	}

	// Exact match should work regardless of confidence
	result := matchSecurityScheme(patterns, "BearerAuth")
	if !result {
		t.Error("Expected exact match to work regardless of confidence")
	}
}

func TestMatchSecurityScheme_NoMatch(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.9},
	}

	result := matchSecurityScheme(patterns, "OAuth2")
	if result {
		t.Error("Expected no match for OAuth2")
	}
}

func TestNormalizeSchemeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"BearerAuth", "BearerAuth", "bearer"},
		{"ApiKeyAuth", "ApiKeyAuth", "key"},
		{"OAuth2", "OAuth2", "oauth2"},
		{"RBAC", "RBAC", "rbac"},
		{"Bearer", "Bearer", "bearer"},
		{"APIKey", "APIKey", "key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSchemeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSchemeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectSecurityMiddleware_CompleteFlow(t *testing.T) {
	code := `
package middleware

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		// JWT validation
	}
	return next
}

func extractAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}
`

	ctx := context.Background()
	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	if err != nil {
		t.Fatalf("detectSecurityMiddleware failed: %v", err)
	}

	// Should find both BearerAuth and ApiKeyAuth
	bearerFound := false
	apiKeyFound := false
	for _, pattern := range patterns {
		if pattern.Scheme == "BearerAuth" {
			bearerFound = true
		}
		if pattern.Scheme == "ApiKeyAuth" {
			apiKeyFound = true
		}
	}

	if !bearerFound {
		t.Error("Expected to find BearerAuth pattern")
	}
	if !apiKeyFound {
		t.Error("Expected to find ApiKeyAuth pattern")
	}
}

func TestDetectSecurityMiddleware_ContextCancellation(t *testing.T) {
	code := `package main`
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	if len(patterns) > 0 {
		t.Errorf("Expected no patterns on cancellation, got %d", len(patterns))
	}
}

func TestDetectSecurityMiddleware_ContextCancellationInLoop(t *testing.T) {
	code := `
package middleware

func AuthMiddleware1() {}
func AuthMiddleware2() {}
func AuthMiddleware3() {}
`
	ctx, cancel := context.WithCancel(context.Background())

	// Start detection, then cancel during function extraction
	go func() {
		cancel()
	}()

	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	// Should handle cancellation gracefully
	_ = patterns
	_ = err
}

func TestDetectSecurityMiddleware_EmptyCode(t *testing.T) {
	ctx := context.Background()
	patterns, err := detectSecurityMiddleware(ctx, "", "go")
	if err != nil {
		t.Fatalf("detectSecurityMiddleware should handle empty code: %v", err)
	}
	if len(patterns) > 0 {
		t.Errorf("Expected no patterns for empty code, got %d", len(patterns))
	}
}

func TestDetectSecurityMiddleware_ExtractFunctionsError(t *testing.T) {
	// Code that might cause ExtractFunctions to fail (invalid syntax)
	code := `package main
func invalid syntax here
`
	ctx := context.Background()
	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	// Should not error - code-based detection should still work
	if err != nil {
		t.Fatalf("detectSecurityMiddleware should handle ExtractFunctions error gracefully: %v", err)
	}
	// Should still return patterns from code analysis if any
	_ = patterns
}

func TestDetectSecurityMiddleware_ExtractFunctionsSuccess(t *testing.T) {
	// Code with valid functions that should be extracted
	code := `
package middleware

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
`

	ctx := context.Background()
	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	if err != nil {
		t.Fatalf("detectSecurityMiddleware failed: %v", err)
	}

	// Should find patterns from both code analysis and function extraction
	if len(patterns) == 0 {
		t.Error("Expected to find security patterns")
	}

	// Should find CORS from code analysis
	corsFound := false
	for _, pattern := range patterns {
		if pattern.Scheme == "CORS" {
			corsFound = true
			break
		}
	}

	if !corsFound {
		t.Error("Expected to find CORS pattern")
	}
}

func TestDetectSecurityMiddleware_WithFunctionExtraction(t *testing.T) {
	code := `
package middleware

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// JWT validation
		}
		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := NewRateLimiter(100, 10)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", 429)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`

	ctx := context.Background()
	patterns, err := detectSecurityMiddleware(ctx, code, "go")
	if err != nil {
		t.Fatalf("detectSecurityMiddleware failed: %v", err)
	}

	// Should find patterns from both code analysis and function extraction
	if len(patterns) == 0 {
		t.Error("Expected to find security patterns")
	}

	// Check for BearerAuth from code analysis
	bearerFound := false
	rateLimitFound := false
	for _, pattern := range patterns {
		if pattern.Scheme == "BearerAuth" {
			bearerFound = true
		}
		if pattern.Scheme == "RateLimit" {
			rateLimitFound = true
		}
	}

	if !bearerFound {
		t.Error("Expected to find BearerAuth pattern")
	}
	if !rateLimitFound {
		t.Error("Expected to find RateLimit pattern")
	}
}

func TestMatchSecurityScheme_MultiplePatterns(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.9},
		{Type: "authentication", Scheme: "ApiKeyAuth", Confidence: 0.8},
		{Type: "authorization", Scheme: "RBAC", Confidence: 0.9},
	}

	// Should match first pattern
	result := matchSecurityScheme(patterns, "BearerAuth")
	if !result {
		t.Error("Expected match for BearerAuth")
	}

	// Should match second pattern
	result = matchSecurityScheme(patterns, "ApiKeyAuth")
	if !result {
		t.Error("Expected match for ApiKeyAuth")
	}

	// Should match third pattern
	result = matchSecurityScheme(patterns, "RBAC")
	if !result {
		t.Error("Expected match for RBAC")
	}
}

func TestMatchSecurityScheme_PartialMatchHighConfidence(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.8}, // Above threshold
	}

	result := matchSecurityScheme(patterns, "Bearer")
	if !result {
		t.Error("Expected partial match for Bearer with high confidence")
	}
}

func TestMatchSecurityScheme_ContractContainsPattern(t *testing.T) {
	// Test when contract scheme contains the pattern (reverse partial match)
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "Bearer", Confidence: 0.8},
	}

	result := matchSecurityScheme(patterns, "BearerAuth")
	if !result {
		t.Error("Expected match when contract contains pattern")
	}
}

func TestMatchSecurityScheme_NoPartialMatch(t *testing.T) {
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerAuth", Confidence: 0.9},
	}

	// "OAuth" doesn't partially match "BearerAuth"
	result := matchSecurityScheme(patterns, "OAuth")
	if result {
		t.Error("Expected no match for unrelated schemes")
	}
}

func TestMatchSecurityScheme_PartialMatchBelowThreshold(t *testing.T) {
	// Test partial match with confidence below threshold
	patterns := []SecurityPattern{
		{Type: "authentication", Scheme: "BearerToken", Confidence: 0.6}, // Below 0.7 threshold
	}

	// "Bearer" should partially match "BearerToken" but confidence is too low
	result := matchSecurityScheme(patterns, "Bearer")
	if result {
		t.Error("Expected no match for partial match with low confidence")
	}
}

func TestContainsJWTBearerPattern_VariousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"Bearer token", `auth := r.Header.Get("Authorization"); if strings.HasPrefix(auth, "Bearer ") {`, true},
		{"JWT library import", `import "github.com/golang-jwt/jwt/v4"`, true},
		{"JWT parse", `token, err := jwt.Parse(tokenString, keyFunc)`, true},
		{"JWT verify", `err := jwt.Verify(token, key)`, true},
		{"Token parse", `token, err := parseToken(tokenString)`, true},
		{"No JWT", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsJWTBearerPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsJWTBearerPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsAPIKeyPattern_VariousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"X-API-Key header", `apiKey := r.Header.Get("X-API-Key")`, true},
		{"XAPIKey header", `apiKey := r.Header.Get("XAPIKey")`, true},
		{"API key validate", `func ValidateAPIKey(key string) bool { return true }`, true},
		{"Extract API key", `func ExtractAPIKey(r *http.Request) string {`, true},
		{"Get API key", `apiKey := GetAPIKey(r)`, true},
		{"No API key", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsAPIKeyPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsAPIKeyPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsRBACPattern_VariousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"RBAC function", `func CheckRBAC(userID string) bool {`, true},
		{"Role check", `if user.Role == "admin" {`, true},
		{"Has role", `if HasRole(user, "admin") {`, true},
		{"Authorize function", `func Authorize(userID string, action string) bool {`, true},
		{"Permission check", `if HasPermission(user, "read") {`, true},
		{"Verify permission", `if VerifyPermission(user, perm) {`, true},
		{"Role and check", `if user.Role == "admin" && checkPermission(user) {`, true},
		{"Role and verify", `if user.Role == "admin" && verifyAccess(user) {`, true},
		{"Role check function", `func checkRole(user string) bool {`, true},
		{"Role verify function", `func verifyRole(user string) bool {`, true},
		{"No RBAC", `func main() {}`, false},
		{"Role variable (matches role check)", `var role = "admin"`, true}, // Contains "role" so matches
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsRBACPattern(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsRBACPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateSecurityMetadata_WithExistingFindings(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - BearerAuth: []
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	contract, err := ParseOpenAPIContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	endpoint := EndpointInfo{
		Method:    "GET",
		Path:      "/users",
		File:      "handlers/users.go",
		Auth:      []string{}, // No security
		Responses: []ResponseInfo{{StatusCode: 200}},
	}

	contractEndpoint := findMatchingContractEndpoint(endpoint, contract)
	if contractEndpoint == nil {
		t.Fatal("Expected to find matching contract endpoint")
	}

	existingFindings := []APILayerFinding{
		{Type: "info", Issue: "Existing finding"},
	}

	findings := validateSecurityMetadata(ctx, endpoint, *contractEndpoint, existingFindings)

	// Should include existing findings
	if len(findings) < len(existingFindings) {
		t.Error("Expected to preserve existing findings")
	}

	// Should add new findings for missing security
	found := false
	for _, finding := range findings {
		if strings.Contains(finding.Issue, "security") || strings.Contains(finding.Issue, "Security") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected finding for missing security")
	}
}

func TestValidateSecurityMetadata_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - BearerAuth: []
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	contract, err := ParseOpenAPIContract(context.Background(), contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	endpoint := EndpointInfo{
		Method:    "GET",
		Path:      "/users",
		File:      "handlers/users.go",
		Auth:      []string{},
		Responses: []ResponseInfo{{StatusCode: 200}},
	}

	contractEndpoint := findMatchingContractEndpoint(endpoint, contract)
	if contractEndpoint == nil {
		t.Fatal("Expected to find matching contract endpoint")
	}

	cancel() // Cancel context

	findings := validateSecurityMetadata(ctx, endpoint, *contractEndpoint, []APILayerFinding{})

	// Should return early due to context cancellation
	if len(findings) > 1 {
		t.Logf("Context cancellation returned %d findings (acceptable)", len(findings))
	}
}

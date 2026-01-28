// Package services provides security pattern detection for schema validation.
//
// This file contains AST-based security middleware pattern detection logic
// used by the schema validator to verify that security requirements defined
// in OpenAPI contracts are actually implemented in the code.
//
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"
	"strings"

	"sentinel-hub-api/ast"
	"sentinel-hub-api/pkg"
)

// SecurityPattern represents a detected security implementation pattern
// in the analyzed code.
type SecurityPattern struct {
	// Type indicates the category of security pattern (e.g., "authentication", "authorization")
	Type string

	// Scheme is the security scheme name (e.g., "BearerAuth", "ApiKeyAuth")
	Scheme string

	// Location is the file path and line number where the pattern was detected
	Location string

	// Confidence is a score from 0.0 to 1.0 indicating detection confidence
	Confidence float64
}

// detectSecurityMiddleware performs AST-based analysis to detect security
// middleware patterns in the provided code.
//
// It uses ast.AnalyzeAST with "security_middleware" analysis type to identify
// authentication, authorization, and other security-related patterns. Returns a
// slice of SecurityPattern structs representing detected security implementations.
//
// The function respects context cancellation and returns an error if analysis
// fails or context is cancelled.
func detectSecurityMiddleware(ctx context.Context, code string, language string) ([]SecurityPattern, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	patterns := []SecurityPattern{}

	// Use AST-based security middleware detection (primary method)
	analyses := []string{"security_middleware"}
	findings, _, err := ast.AnalyzeAST(code, language, analyses)
	if err != nil {
		// Log error and fall back to code-based detection
		pkg.LogWarn(ctx, "AST analysis failed for security middleware detection: %v", err)
		// Fall back to code-based pattern detection
		codePatterns := detectSecurityPatternsInCode(ctx, code, language)
		patterns = append(patterns, codePatterns...)
		return patterns, nil
	}

	// Convert AST findings to SecurityPatterns
	for _, finding := range findings {
		if ctx.Err() != nil {
			return patterns, ctx.Err()
		}

		pattern := extractSecurityPatternFromASTFinding(finding)
		if pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}

	// Also extract functions to find additional middleware patterns
	// This complements AST detection for edge cases
	functions, err := ast.ExtractFunctions(code, language, "")
	if err != nil {
		// Log but don't fail - AST detection is primary
		pkg.LogWarn(ctx, "Function extraction failed for security pattern detection: %v", err)
	} else {
		if ctx.Err() != nil {
			return patterns, ctx.Err()
		}
		// Analyze functions for middleware patterns (secondary method)
		middlewarePatterns := detectMiddlewareInFunctions(ctx, functions, code, language)
		patterns = append(patterns, middlewarePatterns...)
	}

	return patterns, nil
}

// extractSecurityPatternFromASTFinding converts AST finding to SecurityPattern
func extractSecurityPatternFromASTFinding(finding ast.ASTFinding) *SecurityPattern {
	findingType := strings.ToLower(finding.Type)

	switch {
	case strings.Contains(findingType, "jwt") || strings.Contains(findingType, "bearer"):
		return &SecurityPattern{
			Type:       "authentication",
			Scheme:     "BearerAuth",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	case strings.Contains(findingType, "apikey") || strings.Contains(findingType, "api_key"):
		return &SecurityPattern{
			Type:       "authentication",
			Scheme:     "ApiKeyAuth",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	case strings.Contains(findingType, "oauth"):
		return &SecurityPattern{
			Type:       "authentication",
			Scheme:     "OAuth2",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	case strings.Contains(findingType, "rbac") || strings.Contains(findingType, "authorize"):
		return &SecurityPattern{
			Type:       "authorization",
			Scheme:     "RBAC",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	case strings.Contains(findingType, "ratelimit"):
		return &SecurityPattern{
			Type:       "rate_limit",
			Scheme:     "RateLimit",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	case strings.Contains(findingType, "cors"):
		return &SecurityPattern{
			Type:       "cors",
			Scheme:     "CORS",
			Location:   fmt.Sprintf("line %d", finding.Line),
			Confidence: finding.Confidence,
		}
	default:
		return nil
	}
}

// detectSecurityPatternsInCode detects security patterns using code analysis.
// This is the primary detection method that works across all languages.
func detectSecurityPatternsInCode(ctx context.Context, code string, language string) []SecurityPattern {
	patterns := []SecurityPattern{}
	codeLower := strings.ToLower(code)

	// JWT/Bearer token detection
	if containsJWTBearerPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "authentication",
			Scheme:     "BearerAuth",
			Location:   "code analysis",
			Confidence: 0.85,
		})
	}

	// API key detection
	if containsAPIKeyPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "authentication",
			Scheme:     "ApiKeyAuth",
			Location:   "code analysis",
			Confidence: 0.80,
		})
	}

	// OAuth detection
	if containsOAuthPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "authentication",
			Scheme:     "OAuth2",
			Location:   "code analysis",
			Confidence: 0.80,
		})
	}

	// RBAC/Authorization detection
	if containsRBACPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "authorization",
			Scheme:     "RBAC",
			Location:   "code analysis",
			Confidence: 0.75,
		})
	}

	// Rate limiting detection
	if containsRateLimitPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "rate_limit",
			Scheme:     "RateLimit",
			Location:   "code analysis",
			Confidence: 0.75,
		})
	}

	// CORS detection
	if containsCORSPattern(code, codeLower) {
		patterns = append(patterns, SecurityPattern{
			Type:       "cors",
			Scheme:     "CORS",
			Location:   "code analysis",
			Confidence: 0.70,
		})
	}

	return patterns
}

// containsJWTBearerPattern checks for JWT/Bearer token authentication patterns.
func containsJWTBearerPattern(code, codeLower string) bool {
	// Check for Bearer token in Authorization header
	if strings.Contains(codeLower, "bearer") && strings.Contains(codeLower, "authorization") {
		return true
	}
	// Check for JWT library imports/usage
	if strings.Contains(codeLower, "jwt") || strings.Contains(codeLower, "jsonwebtoken") {
		return true
	}
	// Check for common JWT patterns
	if strings.Contains(codeLower, "token") && (strings.Contains(codeLower, "parse") || strings.Contains(codeLower, "verify")) {
		return true
	}
	return false
}

// containsAPIKeyPattern checks for API key authentication patterns.
func containsAPIKeyPattern(code, codeLower string) bool {
	// Check for X-API-Key header
	if strings.Contains(codeLower, "x-api-key") || strings.Contains(codeLower, "xapikey") {
		return true
	}
	// Check for API key validation functions
	if strings.Contains(codeLower, "apikey") && (strings.Contains(codeLower, "validate") || strings.Contains(codeLower, "check")) {
		return true
	}
	// Check for API key extraction
	if strings.Contains(codeLower, "extractapikey") || strings.Contains(codeLower, "getapikey") {
		return true
	}
	return false
}

// containsOAuthPattern checks for OAuth authentication patterns.
func containsOAuthPattern(code, codeLower string) bool {
	// Check for OAuth mentions
	if strings.Contains(codeLower, "oauth") || strings.Contains(codeLower, "oauth2") {
		return true
	}
	// Check for OAuth flow patterns
	if strings.Contains(codeLower, "authorization") && strings.Contains(codeLower, "code") {
		return true
	}
	return false
}

// containsRBACPattern checks for RBAC/authorization patterns.
func containsRBACPattern(code, codeLower string) bool {
	// Check for RBAC mentions
	if strings.Contains(codeLower, "rbac") || strings.Contains(codeLower, "role") {
		return true
	}
	// Check for authorization checks
	if strings.Contains(codeLower, "authorize") || strings.Contains(codeLower, "permission") {
		return true
	}
	// Check for role-based patterns
	if strings.Contains(codeLower, "role") && (strings.Contains(codeLower, "check") || strings.Contains(codeLower, "verify")) {
		return true
	}
	return false
}

// containsRateLimitPattern checks for rate limiting patterns.
func containsRateLimitPattern(code, codeLower string) bool {
	// Check for rate limit mentions
	if strings.Contains(codeLower, "ratelimit") || strings.Contains(codeLower, "rate_limit") {
		return true
	}
	// Check for rate limiting middleware
	if strings.Contains(codeLower, "ratelimiter") || strings.Contains(codeLower, "throttle") {
		return true
	}
	return false
}

// containsCORSPattern checks for CORS patterns.
func containsCORSPattern(code, codeLower string) bool {
	// Check for CORS mentions
	if strings.Contains(codeLower, "cors") || strings.Contains(codeLower, "cross-origin") {
		return true
	}
	// Check for CORS headers
	if strings.Contains(codeLower, "access-control-allow-origin") {
		return true
	}
	return false
}

// detectMiddlewareInFunctions detects middleware patterns in function definitions.
func detectMiddlewareInFunctions(ctx context.Context, functions []ast.FunctionInfo, code string, language string) []SecurityPattern {
	patterns := []SecurityPattern{}

	// Common middleware function name patterns
	middlewarePatterns := []struct {
		namePattern string
		scheme      string
		patternType string
	}{
		{"Auth", "BearerAuth", "authentication"},
		{"Authenticate", "BearerAuth", "authentication"},
		{"JWT", "BearerAuth", "authentication"},
		{"APIKey", "ApiKeyAuth", "authentication"},
		{"ApiKey", "ApiKeyAuth", "authentication"},
		{"OAuth", "OAuth2", "authentication"},
		{"Authorize", "RBAC", "authorization"},
		{"RBAC", "RBAC", "authorization"},
		{"RateLimit", "RateLimit", "rate_limit"},
		{"CORS", "CORS", "cors"},
	}

	for _, fn := range functions {
		if ctx.Err() != nil {
			return patterns
		}

		fnName := strings.ToLower(fn.Name)

		for _, pattern := range middlewarePatterns {
			if strings.Contains(fnName, strings.ToLower(pattern.namePattern)) {
				// Check if function has middleware signature
				if isMiddlewareFunction(fn, code, language) {
					patterns = append(patterns, SecurityPattern{
						Type:       pattern.patternType,
						Scheme:     pattern.scheme,
						Location:   fmt.Sprintf("function %s", fn.Name),
						Confidence: 0.7, // Lower confidence for name-based detection
					})
					break // Found a match, move to next function
				}
			}
		}
	}

	return patterns
}

// isMiddlewareFunction checks if a function has middleware signature or naming pattern.
// Uses name-based heuristics since FunctionInfo doesn't expose full signature details.
func isMiddlewareFunction(fn ast.FunctionInfo, code string, language string) bool {
	fnNameLower := strings.ToLower(fn.Name)

	// Common middleware naming patterns across languages
	middlewareSuffixes := []string{"middleware", "auth", "authmiddleware", "ratelimit", "cors", "authenticate", "authorize"}
	for _, suffix := range middlewareSuffixes {
		if strings.HasSuffix(fnNameLower, suffix) {
			return true
		}
	}

	// Check if name contains middleware keywords
	middlewareKeywords := []string{"middleware", "auth", "jwt", "oauth", "rbac", "ratelimit", "cors", "authenticate", "authorize"}
	for _, keyword := range middlewareKeywords {
		if strings.Contains(fnNameLower, keyword) {
			return true
		}
	}

	// For Go: Additional checks for common Go middleware patterns
	if language == "go" {
		// Check function code for middleware signature patterns
		if strings.Contains(code, "http.Handler") && strings.Contains(code, "func") {
			// Look for function definition in code around the function
			fnCode := extractFunctionCodeAroundLine(code, fn.Line)
			if strings.Contains(fnCode, "http.Handler") && strings.Contains(fnCode, "func") {
				return true
			}
		}
	}

	return false
}

// extractFunctionCodeAroundLine extracts code around a function line for analysis.
func extractFunctionCodeAroundLine(code string, lineNum int) string {
	lines := strings.Split(code, "\n")
	// Extract 10 lines before and after the function
	start := lineNum - 10
	if start < 0 {
		start = 0
	}
	end := lineNum + 10
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

// matchSecurityScheme matches detected patterns to contract security schemes.
func matchSecurityScheme(patterns []SecurityPattern, contractScheme string) bool {
	// Normalize scheme names for comparison
	normalizedContract := normalizeSchemeName(contractScheme)

	for _, pattern := range patterns {
		normalizedPattern := normalizeSchemeName(pattern.Scheme)

		// Exact match
		if normalizedPattern == normalizedContract {
			return true
		}

		// Partial match (e.g., "Bearer" matches "BearerAuth")
		if strings.Contains(normalizedContract, normalizedPattern) ||
			strings.Contains(normalizedPattern, normalizedContract) {
			// Check confidence threshold
			if pattern.Confidence >= 0.7 {
				return true
			}
		}
	}

	return false
}

// normalizeSchemeName normalizes security scheme names for comparison.
func normalizeSchemeName(scheme string) string {
	// Remove common suffixes/prefixes
	normalized := strings.ToLower(scheme)
	normalized = strings.TrimSuffix(normalized, "auth")
	normalized = strings.TrimSuffix(normalized, "authentication")
	normalized = strings.TrimPrefix(normalized, "api")
	normalized = strings.TrimSpace(normalized)
	return normalized
}

// Package ast provides security middleware detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectSecurityMiddleware finds security middleware patterns in code
// Uses language registry for supported languages, falls back to generic detection
func detectSecurityMiddleware(root *sitter.Node, code string, language string) []ASTFinding {
	// Try to get language-specific detector from registry
	detector := GetLanguageDetector(language)
	if detector != nil {
		// Use language-specific detection
		return detector.DetectSecurityMiddleware(root, code)
	}

	// Fallback to generic pattern detection for unsupported languages
	return detectSecurityMiddlewareGeneric(root, code, language)
}

// detectSecurityMiddlewareGo detects security middleware in Go code
func detectSecurityMiddlewareGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	codeLower := strings.ToLower(code)

	// Traverse AST to find function declarations
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			funcName, funcCode := extractFunctionInfo(node, code)
			if funcName == "" {
				return true // Continue traversal
			}

			funcNameLower := strings.ToLower(funcName)
			funcCodeLower := strings.ToLower(funcCode)

			// Check for JWT/Bearer token middleware
			if containsJWTBearerPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", funcName)
				findings = append(findings, finding)
			}

			// Check for API key middleware
			if containsAPIKeyPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", funcName)
				findings = append(findings, finding)
			}

			// Check for OAuth middleware
			if containsOAuthPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "oauth_middleware", "OAuth2", funcName)
				findings = append(findings, finding)
			}

			// Check for RBAC/Authorization middleware
			if containsRBACPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "rbac_middleware", "RBAC", funcName)
				findings = append(findings, finding)
			}

			// Check for rate limiting middleware
			if containsRateLimitPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "ratelimit_middleware", "RateLimit", funcName)
				findings = append(findings, finding)
			}

			// Check for CORS middleware
			if containsCORSPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "cors_middleware", "CORS", funcName)
				findings = append(findings, finding)
			}

			// Check for middleware signature pattern: func(http.Handler) http.Handler
			if isGoMiddlewareSignature(node, code) {
				// Additional check: if it's a middleware but not yet classified
				if !isClassifiedMiddleware(funcNameLower, funcCodeLower) {
					// Generic middleware finding
					finding := createMiddlewareFinding(node, code, "generic_middleware", "Middleware", funcName)
					findings = append(findings, finding)
				}
			}
		}
		return true // Continue traversal
	})

	// Also check for security patterns in the code (not just functions)
	// Check for JWT/Bearer in Authorization headers
	if strings.Contains(codeLower, "authorization") && strings.Contains(codeLower, "bearer") {
		finding := ASTFinding{
			Type:       "jwt_middleware",
			Severity:   "info",
			Line:       1,
			Column:     1,
			Message:    "JWT/Bearer token authentication detected",
			Code:       extractCodeSnippet(code, "bearer", "authorization"),
			Suggestion: "Security middleware pattern detected",
			Confidence: 0.85,
		}
		findings = append(findings, finding)
	}

	// Check for API key headers
	if strings.Contains(codeLower, "x-api-key") || strings.Contains(codeLower, "xapikey") {
		finding := ASTFinding{
			Type:       "apikey_middleware",
			Severity:   "info",
			Line:       1,
			Column:     1,
			Message:    "API key authentication detected",
			Code:       extractCodeSnippet(code, "api", "key"),
			Suggestion: "Security middleware pattern detected",
			Confidence: 0.80,
		}
		findings = append(findings, finding)
	}

	return findings
}

// detectSecurityMiddlewareJS detects security middleware in JavaScript/TypeScript code
func detectSecurityMiddlewareJS(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}

	// Traverse AST to find function declarations
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "function" ||
			node.Type() == "arrow_function" || node.Type() == "method_definition" {
			funcName, funcCode := extractFunctionInfo(node, code)
			funcNameLower := strings.ToLower(funcName)
			funcCodeLower := strings.ToLower(funcCode)

			// Check for middleware patterns
			if containsJWTBearerPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", funcName)
				findings = append(findings, finding)
			}
			if containsAPIKeyPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", funcName)
				findings = append(findings, finding)
			}
			if containsOAuthPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "oauth_middleware", "OAuth2", funcName)
				findings = append(findings, finding)
			}
		}
		return true
	})

	return findings
}

// detectSecurityMiddlewarePython detects security middleware in Python code
func detectSecurityMiddlewarePython(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}

	// Traverse AST to find function definitions
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_definition" {
			funcName, funcCode := extractFunctionInfo(node, code)
			funcNameLower := strings.ToLower(funcName)
			funcCodeLower := strings.ToLower(funcCode)

			// Check for middleware patterns
			if containsJWTBearerPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", funcName)
				findings = append(findings, finding)
			}
			if containsAPIKeyPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", funcName)
				findings = append(findings, finding)
			}
		}
		return true
	})

	return findings
}

// detectSecurityMiddlewareGeneric provides comprehensive generic pattern detection
// for unsupported languages. Uses the same pattern detection as code-based fallback
// to ensure consistent detection across all languages.
func detectSecurityMiddlewareGeneric(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}
	codeLower := strings.ToLower(code)

	// Try AST-based detection first if root is available
	if root != nil {
		astFindings := detectSecurityMiddlewareFromAST(root, code, language)
		if len(astFindings) > 0 {
			findings = append(findings, astFindings...)
		}
	}

	// JWT/Bearer token detection
	if containsJWTBearerPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "jwt_middleware", "BearerAuth", language, "bearer", "authorization")
		findings = append(findings, finding)
	}

	// API key detection
	if containsAPIKeyPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "apikey_middleware", "ApiKeyAuth", language, "api", "key")
		findings = append(findings, finding)
	}

	// OAuth detection
	if containsOAuthPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "oauth_middleware", "OAuth2", language, "oauth", "")
		findings = append(findings, finding)
	}

	// RBAC detection
	if containsRBACPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "rbac_middleware", "RBAC", language, "rbac", "role")
		findings = append(findings, finding)
	}

	// Rate limiting detection
	if containsRateLimitPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "ratelimit_middleware", "RateLimit", language, "ratelimit", "throttle")
		findings = append(findings, finding)
	}

	// CORS detection
	if containsCORSPatternGeneric(code, codeLower) {
		finding := createGenericMiddlewareFinding(code, "cors_middleware", "CORS", language, "cors", "")
		findings = append(findings, finding)
	}

	return findings
}

// detectSecurityMiddlewareFromAST performs AST-based detection for generic fallback
func detectSecurityMiddlewareFromAST(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Get language-specific function node types
	functionNodeTypes := getFunctionNodeTypes(language)

	// Traverse AST to find function declarations (language-aware approach)
	TraverseAST(root, func(node *sitter.Node) bool {
		// Check for function-like nodes using language-specific types
		nodeType := node.Type()
		isFunction := false
		for _, funcType := range functionNodeTypes {
			if nodeType == funcType {
				isFunction = true
				break
			}
		}
		// Fallback to generic check if no language-specific types match
		if !isFunction {
			isFunction = strings.Contains(nodeType, "function") || strings.Contains(nodeType, "method") ||
				strings.Contains(nodeType, "declaration") || strings.Contains(nodeType, "definition")
		}

		if isFunction {
			funcName, funcCode := extractFunctionInfo(node, code)
			if funcName == "" {
				return true // Continue traversal
			}

			funcNameLower := strings.ToLower(funcName)
			funcCodeLower := strings.ToLower(funcCode)

			// Check for middleware patterns
			if containsJWTBearerPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "jwt_middleware", "BearerAuth", funcName)
				// Add language context to finding
				finding.Reasoning = fmt.Sprintf("Function %s matches JWT middleware pattern in %s code", funcName, language)
				findings = append(findings, finding)
			}
			if containsAPIKeyPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "apikey_middleware", "ApiKeyAuth", funcName)
				finding.Reasoning = fmt.Sprintf("Function %s matches API key middleware pattern in %s code", funcName, language)
				findings = append(findings, finding)
			}
			if containsOAuthPattern(funcNameLower, funcCodeLower) {
				finding := createMiddlewareFinding(node, code, "oauth_middleware", "OAuth2", funcName)
				finding.Reasoning = fmt.Sprintf("Function %s matches OAuth middleware pattern in %s code", funcName, language)
				findings = append(findings, finding)
			}
		}
		return true
	})

	return findings
}

// Generic pattern detection functions (reused from code-based detection logic)

func containsJWTBearerPatternGeneric(code, codeLower string) bool {
	// Check for Bearer token in Authorization header (case-sensitive check in original code)
	if strings.Contains(code, "Authorization") && strings.Contains(codeLower, "bearer") {
		return true
	}
	// Check for Bearer token in Authorization header (lowercase)
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

func containsAPIKeyPatternGeneric(code, codeLower string) bool {
	// Check for X-API-Key header (case-sensitive check in original code)
	if strings.Contains(code, "X-API-Key") || strings.Contains(code, "X-Api-Key") {
		return true
	}
	// Check for X-API-Key header (lowercase)
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

func containsOAuthPatternGeneric(code, codeLower string) bool {
	// Check for OAuth mentions (case-sensitive check in original code)
	if strings.Contains(code, "OAuth") || strings.Contains(code, "OAuth2") {
		return true
	}
	// Check for OAuth mentions (lowercase)
	if strings.Contains(codeLower, "oauth") || strings.Contains(codeLower, "oauth2") {
		return true
	}
	// Check for OAuth flow patterns
	if strings.Contains(codeLower, "authorization") && strings.Contains(codeLower, "code") {
		return true
	}
	return false
}

func containsRBACPatternGeneric(code, codeLower string) bool {
	// Check for RBAC mentions (case-sensitive check in original code)
	if strings.Contains(code, "RBAC") || strings.Contains(code, "Role") {
		return true
	}
	// Check for RBAC mentions (lowercase)
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

func containsRateLimitPatternGeneric(code, codeLower string) bool {
	// Check for rate limit mentions (case-sensitive check in original code)
	if strings.Contains(code, "RateLimit") || strings.Contains(code, "Rate_Limit") {
		return true
	}
	// Check for rate limit mentions (lowercase)
	if strings.Contains(codeLower, "ratelimit") || strings.Contains(codeLower, "rate_limit") {
		return true
	}
	// Check for rate limiting middleware
	if strings.Contains(codeLower, "ratelimiter") || strings.Contains(codeLower, "throttle") {
		return true
	}
	return false
}

func containsCORSPatternGeneric(code, codeLower string) bool {
	// Check for CORS mentions (case-sensitive check in original code)
	if strings.Contains(code, "CORS") || strings.Contains(code, "Cross-Origin") {
		return true
	}
	// Check for CORS mentions (lowercase)
	if strings.Contains(codeLower, "cors") || strings.Contains(codeLower, "cross-origin") {
		return true
	}
	// Check for CORS headers (case-sensitive)
	if strings.Contains(code, "Access-Control-Allow-Origin") {
		return true
	}
	// Check for CORS headers (lowercase)
	if strings.Contains(codeLower, "access-control-allow-origin") {
		return true
	}
	return false
}

// createGenericMiddlewareFinding creates an ASTFinding for detected middleware in generic detection
func createGenericMiddlewareFinding(code, findingType, scheme, language, keyword1, keyword2 string) ASTFinding {
	// Find line number for the pattern
	lines := strings.Split(code, "\n")
	lineNum := 1
	for i, line := range lines {
		lineLower := strings.ToLower(line)
		if keyword2 != "" {
			if strings.Contains(lineLower, keyword1) && strings.Contains(lineLower, keyword2) {
				lineNum = i + 1
				break
			}
		} else if strings.Contains(lineLower, keyword1) {
			lineNum = i + 1
			break
		}
	}

	// Use language context in reasoning
	reasoningText := "Pattern detected via generic analysis"
	if language != "" {
		reasoningText = fmt.Sprintf("Pattern detected via generic analysis for %s language", language)
	}

	return ASTFinding{
		Type:       findingType,
		Severity:   "info",
		Line:       lineNum,
		Column:     1,
		Message:    fmt.Sprintf("Security middleware detected: %s", scheme),
		Code:       extractCodeSnippet(code, keyword1, keyword2),
		Suggestion: fmt.Sprintf("Middleware implements %s security scheme", scheme),
		Confidence: 0.75, // Lower confidence for generic detection
		Reasoning:  reasoningText,
	}
}

// Helper functions for pattern detection

func containsJWTBearerPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "jwt") || strings.Contains(funcName, "bearer") ||
		strings.Contains(funcName, "token") ||
		strings.Contains(funcCode, "bearer") && strings.Contains(funcCode, "authorization") ||
		strings.Contains(funcCode, "jwt") || strings.Contains(funcCode, "jsonwebtoken")
}

func containsAPIKeyPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "apikey") || strings.Contains(funcName, "api_key") ||
		strings.Contains(funcCode, "x-api-key") || strings.Contains(funcCode, "xapikey") ||
		strings.Contains(funcCode, "extractapikey") || strings.Contains(funcCode, "validateapikey")
}

func containsOAuthPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "oauth") || strings.Contains(funcCode, "oauth") ||
		strings.Contains(funcCode, "oauth2")
}

func containsRBACPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "rbac") || strings.Contains(funcName, "role") ||
		strings.Contains(funcName, "authorize") || strings.Contains(funcCode, "rbac") ||
		strings.Contains(funcCode, "role") && (strings.Contains(funcCode, "check") || strings.Contains(funcCode, "verify"))
}

func containsRateLimitPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "ratelimit") || strings.Contains(funcName, "rate_limit") ||
		strings.Contains(funcCode, "ratelimit") || strings.Contains(funcCode, "throttle")
}

func containsCORSPattern(funcName, funcCode string) bool {
	return strings.Contains(funcName, "cors") || strings.Contains(funcCode, "cors") ||
		strings.Contains(funcCode, "access-control-allow-origin")
}

func isClassifiedMiddleware(funcName, funcCode string) bool {
	return containsJWTBearerPattern(funcName, funcCode) ||
		containsAPIKeyPattern(funcName, funcCode) ||
		containsOAuthPattern(funcName, funcCode) ||
		containsRBACPattern(funcName, funcCode) ||
		containsRateLimitPattern(funcName, funcCode) ||
		containsCORSPattern(funcName, funcCode)
}

// isGoMiddlewareSignature checks if a function has the middleware signature: func(http.Handler) http.Handler
func isGoMiddlewareSignature(node *sitter.Node, code string) bool {
	// Check if function signature contains http.Handler
	funcCode := safeSlice(code, node.StartByte(), node.EndByte())
	return strings.Contains(funcCode, "http.Handler") && strings.Contains(funcCode, "func")
}

// extractFunctionInfo extracts function name and code from AST node
func extractFunctionInfo(node *sitter.Node, code string) (string, string) {
	funcName := ""
	funcCode := safeSlice(code, node.StartByte(), node.EndByte())

	// Extract function name from node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "identifier" {
			funcName = safeSlice(code, child.StartByte(), child.EndByte())
			break
		}
	}

	return funcName, funcCode
}

// createMiddlewareFinding creates an ASTFinding for detected middleware
func createMiddlewareFinding(node *sitter.Node, code string, findingType, scheme, funcName string) ASTFinding {
	return ASTFinding{
		Type:       findingType,
		Severity:   "info",
		Line:       int(node.StartPoint().Row) + 1,
		Column:     int(node.StartPoint().Column) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		EndColumn:  int(node.EndPoint().Column) + 1,
		Message:    fmt.Sprintf("Security middleware detected: %s (%s)", funcName, scheme),
		Code:       safeSlice(code, node.StartByte(), node.EndByte()),
		Suggestion: fmt.Sprintf("Middleware implements %s security scheme", scheme),
		Confidence: 0.85,
		Reasoning:  fmt.Sprintf("Function %s matches %s middleware pattern", funcName, scheme),
	}
}

// extractCodeSnippet extracts a code snippet containing the given keywords
func extractCodeSnippet(code, keyword1, keyword2 string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, keyword1) && strings.Contains(lineLower, keyword2) {
			// Return surrounding context
			start := i - 2
			if start < 0 {
				start = 0
			}
			end := i + 3
			if end > len(lines) {
				end = len(lines)
			}
			return strings.Join(lines[start:end], "\n")
		}
	}
	return ""
}

// getFunctionNodeTypes returns language-specific AST node types for function declarations
func getFunctionNodeTypes(language string) []string {
	switch language {
	case "go":
		return []string{"function_declaration", "method_declaration"}
	case "javascript", "typescript":
		return []string{"function_declaration", "function", "arrow_function", "method_definition"}
	case "python":
		return []string{"function_definition"}
	case "java":
		return []string{"method_declaration", "constructor_declaration"}
	default:
		// Return empty slice to use fallback generic detection
		return []string{}
	}
}

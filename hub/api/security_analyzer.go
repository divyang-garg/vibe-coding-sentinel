// Security Analyzer - Phase 8 Implementation
// AST-based security rule checking for common vulnerabilities

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// DataFlowAnalyzer tracks variable assignments and usages for data flow analysis
type DataFlowAnalyzer struct {
	variables map[string]*VariableInfo
	functions map[string]*FunctionInfo
}

// VariableInfo tracks a variable's assignments and usages
type VariableInfo struct {
	Name        string
	Type        string // "password", "hash", "other"
	Assignments []Assignment
	Usages      []Usage
	Line        int
}

// Assignment represents a variable assignment
type Assignment struct {
	Line    int
	Value   string
	Source  string // "user_input", "function_call", "variable", "literal"
	Context string // function or scope where assignment occurs
}

// Usage represents a variable usage
type Usage struct {
	Line    int
	Context string // function or scope where usage occurs
	Target  string // function or variable name
	Type    string // "function_call", "assignment", "return", "comparison"
}

// FunctionInfo tracks function definitions and calls
type FunctionInfo struct {
	Name       string
	Parameters []string
	Returns    []string
	Line       int
	Calls      []FunctionCall
}

// FunctionCall represents a function call
type FunctionCall struct {
	FunctionName string
	Arguments    []string
	Line         int
}

// SecurityRule represents a security rule definition
type SecurityRule struct {
	ID          string
	Name        string
	Type        string // authorization, authentication, injection, validation, cryptography, transport
	Severity    string // critical, high, medium, low
	Description string
	Detection   SecurityDetection
	ASTCheck    *ASTSecurityCheck
}

// SecurityDetection defines how to detect security issues
type SecurityDetection struct {
	Endpoints         []string // Route patterns to check
	PatternsForbidden []string // Regex patterns that indicate vulnerability
	PatternsRequired  []string // Regex patterns that must be present (safe patterns)
	RequiredChecks    []string // Required ownership/validation checks
}

// ASTSecurityCheck defines AST-based checks
type ASTSecurityCheck struct {
	FunctionContains []string // Functions that trigger this check
	MustHaveBefore   string   // Required check before response
	RouteMiddleware  []string // Required middleware names
}

// FrameworkType represents detected web framework
type FrameworkType string

const (
	FrameworkExpress FrameworkType = "express"
	FrameworkFastAPI FrameworkType = "fastapi"
	FrameworkGin     FrameworkType = "gin"
	FrameworkFlask   FrameworkType = "flask"
	FrameworkDjango  FrameworkType = "django"
	FrameworkRails   FrameworkType = "rails"
	FrameworkUnknown FrameworkType = "unknown"
)

// FrameworkDetection represents framework detection with confidence scoring
type FrameworkDetection struct {
	Framework  FrameworkType
	Confidence string   // "high", "medium", "low"
	Evidence   []string // List of evidence (imports, route patterns, etc.)
}

// Note: SecurityFinding is defined in main.go to match API response structure

// SecurityRules contains all security rule definitions
var SecurityRules = map[string]SecurityRule{
	"SEC-001": {
		ID:          "SEC-001",
		Name:        "Resource Ownership Verification",
		Type:        "authorization",
		Severity:    "critical",
		Description: "Ensure resource access is verified against user ownership",
		Detection: SecurityDetection{
			Endpoints:      []string{"/api/:resource/:id", "/api/users/:id", "/api/posts/:id"},
			RequiredChecks: []string{"req.user.id === resource.userId", "req.user.role === 'admin'"},
		},
		ASTCheck: &ASTSecurityCheck{
			FunctionContains: []string{"findById", "findOne", "getById", "getUser", "getPost"},
			MustHaveBefore:   "ownership_check",
		},
	},
	"SEC-002": {
		ID:          "SEC-002",
		Name:        "SQL Injection Prevention",
		Type:        "injection",
		Severity:    "critical",
		Description: "Ensure SQL queries use parameterized statements",
		Detection: SecurityDetection{
			PatternsForbidden: []string{
				"(?i)(SELECT|INSERT|UPDATE|DELETE).*\\+.*['\"]",
				"(?i)EXEC\\s*\\(@|EXECUTE\\s*\\(@|sp_executesql\\s+@",
				"query\\([^)]*\\+",
				"db\\.query\\([^)]*['\"]\\s*\\+",
			},
			PatternsRequired: []string{
				"\\$[0-9]+|\\?|:param|@param|\\$1|\\$2",
			},
		},
		ASTCheck: &ASTSecurityCheck{
			FunctionContains: []string{"query", "execute", "exec", "db.query", "db.execute"},
		},
	},
	"SEC-003": {
		ID:          "SEC-003",
		Name:        "Authentication Middleware",
		Type:        "authentication",
		Severity:    "critical",
		Description: "Ensure protected routes have authentication middleware",
		Detection: SecurityDetection{
			Endpoints: []string{"/api/*"},
		},
		ASTCheck: &ASTSecurityCheck{
			RouteMiddleware: []string{"auth", "authenticate", "requireAuth", "jwt", "passport"},
		},
	},
	"SEC-004": {
		ID:          "SEC-004",
		Name:        "Rate Limiting",
		Type:        "transport",
		Severity:    "high",
		Description: "Ensure API endpoints have rate limiting",
		Detection: SecurityDetection{
			Endpoints: []string{"/api/*"},
		},
		ASTCheck: &ASTSecurityCheck{
			RouteMiddleware: []string{"rateLimit", "rate-limit", "throttle", "limiter"},
		},
	},
	"SEC-005": {
		ID:          "SEC-005",
		Name:        "Password Hashing",
		Type:        "cryptography",
		Severity:    "critical",
		Description: "Ensure passwords are hashed using secure algorithms",
		Detection: SecurityDetection{
			PatternsForbidden: []string{
				"password\\s*=\\s*['\"][^'\"]+['\"]",
				"md5\\s*\\(",
				"sha1\\s*\\(",
				"crypto\\.createHash\\s*\\(['\"]md5['\"]",
				"hashlib\\.md5",
				"hashlib\\.sha1",
			},
			PatternsRequired: []string{
				"bcrypt|argon2|scrypt|pbkdf2|bcrypt\\.hash|bcrypt\\.hashSync|argon2\\.hash",
			},
		},
		ASTCheck: &ASTSecurityCheck{
			FunctionContains: []string{"createUser", "register", "signup", "password", "hashPassword"},
		},
	},
	"SEC-006": {
		ID:          "SEC-006",
		Name:        "Input Validation",
		Type:        "validation",
		Severity:    "high",
		Description: "Ensure user input is validated before processing",
		Detection: SecurityDetection{
			PatternsRequired: []string{
				"validate|validator|joi|yup|zod|express-validator|marshmallow|pydantic",
			},
		},
		ASTCheck: &ASTSecurityCheck{
			FunctionContains: []string{"create", "update", "post", "put", "patch"},
		},
	},
	"SEC-007": {
		ID:          "SEC-007",
		Name:        "Secure Headers",
		Type:        "transport",
		Severity:    "medium",
		Description: "Ensure secure HTTP headers are set",
		Detection: SecurityDetection{
			PatternsRequired: []string{
				"helmet|secure-headers|cors|X-Frame-Options|X-Content-Type-Options|Strict-Transport-Security",
			},
		},
		ASTCheck: &ASTSecurityCheck{
			RouteMiddleware: []string{"helmet", "secureHeaders", "cors"},
		},
	},
	"SEC-008": {
		ID:          "SEC-008",
		Name:        "CORS Configuration",
		Type:        "transport",
		Severity:    "high",
		Description: "Ensure CORS is properly configured (not wildcard for production)",
		Detection: SecurityDetection{
			PatternsForbidden: []string{
				"origin:\\s*['\"]\\*['\"]",
				"Access-Control-Allow-Origin:\\s*['\"]\\*['\"]",
				"cors\\(\\{[^}]*origin:\\s*['\"]\\*",
			},
		},
		ASTCheck: &ASTSecurityCheck{
			FunctionContains: []string{"cors", "CORS", "enableCors"},
		},
	},
}

// verifyFrameworkUsage checks actual framework usage in AST (not just imports)
func verifyFrameworkUsage(rootNode *sitter.Node, code string, language string, detectedFramework FrameworkType) FrameworkDetection {
	detection := FrameworkDetection{
		Framework:  detectedFramework,
		Confidence: "low",
		Evidence:   []string{},
	}

	if detectedFramework == FrameworkUnknown {
		return detection
	}

	// Check for route definitions (high confidence)
	hasRoutes := false
	traverseAST(rootNode, func(n *sitter.Node) bool {
		switch detectedFramework {
		case FrameworkExpress:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "app.get") || strings.Contains(nodeText, "app.post") ||
					strings.Contains(nodeText, "router.get") || strings.Contains(nodeText, "router.post") {
					hasRoutes = true
					detection.Evidence = append(detection.Evidence, "Route definitions found")
					return false
				}
			}
		case FrameworkGin:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "router.GET") || strings.Contains(nodeText, "router.POST") {
					hasRoutes = true
					detection.Evidence = append(detection.Evidence, "Route definitions found")
					return false
				}
			}
		case FrameworkFastAPI:
			if n.Type() == "decorator" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "@app.get") || strings.Contains(nodeText, "@app.post") {
					hasRoutes = true
					detection.Evidence = append(detection.Evidence, "Route decorators found")
					return false
				}
			}
		}
		return true
	})

	if hasRoutes {
		detection.Confidence = "high"
		return detection
	}

	// Check for middleware usage (medium confidence)
	hasMiddleware := false
	traverseAST(rootNode, func(n *sitter.Node) bool {
		switch detectedFramework {
		case FrameworkExpress:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "app.use") || strings.Contains(nodeText, "router.use") {
					hasMiddleware = true
					detection.Evidence = append(detection.Evidence, "Middleware usage found")
					return false
				}
			}
		case FrameworkGin:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, ".Use(") {
					hasMiddleware = true
					detection.Evidence = append(detection.Evidence, "Middleware usage found")
					return false
				}
			}
		case FrameworkFastAPI:
			if n.Type() == "decorator" || n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "@app.middleware") || strings.Contains(nodeText, "Depends(") {
					hasMiddleware = true
					detection.Evidence = append(detection.Evidence, "Middleware usage found")
					return false
				}
			}
		}
		return true
	})

	if hasMiddleware {
		detection.Confidence = "medium"
		return detection
	}

	// Only imports found (low confidence)
	detection.Evidence = append(detection.Evidence, "Only imports found, no route/middleware usage")
	return detection
}

// detectFramework detects the web framework from code (returns FrameworkDetection with confidence)
func detectFramework(code string, language string, rootNode *sitter.Node) FrameworkDetection {
	codeLower := strings.ToLower(code)
	var detectedFramework FrameworkType

	switch language {
	case "javascript", "typescript":
		if strings.Contains(codeLower, "express") || strings.Contains(codeLower, "require('express')") || strings.Contains(codeLower, "import express") {
			detectedFramework = FrameworkExpress
		} else if strings.Contains(codeLower, "fastify") {
			detectedFramework = FrameworkExpress // Similar pattern
		}
	case "python":
		if strings.Contains(codeLower, "from fastapi") || strings.Contains(codeLower, "import fastapi") {
			detectedFramework = FrameworkFastAPI
		} else if strings.Contains(codeLower, "from flask") || strings.Contains(codeLower, "import flask") {
			detectedFramework = FrameworkFlask
		} else if strings.Contains(codeLower, "from django") || strings.Contains(codeLower, "import django") {
			detectedFramework = FrameworkDjango
		}
	case "go":
		if strings.Contains(codeLower, "github.com/gin-gonic/gin") || strings.Contains(codeLower, "gin.") {
			detectedFramework = FrameworkGin
		}
	case "ruby":
		if strings.Contains(codeLower, "rails") || strings.Contains(codeLower, "actioncontroller") {
			detectedFramework = FrameworkRails
		}
	}

	if detectedFramework == FrameworkUnknown {
		return FrameworkDetection{
			Framework:  FrameworkUnknown,
			Confidence: "low",
			Evidence:   []string{"No framework imports detected"},
		}
	}

	// Verify actual usage if AST is available
	if rootNode != nil {
		return verifyFrameworkUsage(rootNode, code, language, detectedFramework)
	}

	// Fallback: only imports detected (low confidence)
	return FrameworkDetection{
		Framework:  detectedFramework,
		Confidence: "low",
		Evidence:   []string{"Framework import detected, but AST not available for verification"},
	}
}

// Security cache for performance
var (
	securityCache      = make(map[string]*securityCacheEntry)
	securityCacheMutex sync.RWMutex
	securityCacheTTL   = 5 * time.Minute
)

type securityCacheEntry struct {
	Findings []SecurityFinding
	Expires  time.Time
}

func getSecurityCacheKey(code string, language string, rulesToCheck []string) string {
	hash := sha256.Sum256([]byte(code + language + strings.Join(rulesToCheck, ",")))
	return hex.EncodeToString(hash[:])
}

// countASTNodes counts the number of nodes in the AST (for size validation)
func countASTNodes(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	count := 1
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			count += countASTNodes(child)
		}
	}
	return count
}

// analyzeSecurity performs security analysis on code
func analyzeSecurity(code string, language string, filename string, rulesToCheck []string) ([]SecurityFinding, error) {
	startTime := time.Now()
	log.Printf("Starting security analysis for file: %s (language: %s, rules: %v)", filename, language, rulesToCheck)

	// Edge case: Empty code
	if len(code) == 0 {
		log.Printf("Security analysis skipped: empty code for file %s", filename)
		return []SecurityFinding{}, nil
	}

	// Edge case: File size limit (10MB)
	const maxFileSize = 10 * 1024 * 1024
	if len(code) > maxFileSize {
		return []SecurityFinding{
			{
				RuleID:      "SEC-ERROR",
				RuleName:    "File Size Limit Exceeded",
				Severity:    "info",
				Line:        1,
				Code:        "",
				Issue:       fmt.Sprintf("File exceeds size limit (%d bytes). Security analysis skipped.", len(code)),
				Remediation: "Split large files into smaller modules for better analysis.",
				AutoFixable: false,
			},
		}, nil
	}

	// Check cache first
	cacheKey := getSecurityCacheKey(code, language, rulesToCheck)
	securityCacheMutex.RLock()
	if entry, ok := securityCache[cacheKey]; ok {
		if time.Now().Before(entry.Expires) {
			securityCacheMutex.RUnlock()
			return entry.Findings, nil
		}
		// Cache expired, remove it
		delete(securityCache, cacheKey)
	}
	securityCacheMutex.RUnlock()
	var findings []SecurityFinding

	// Get parser for AST analysis
	parser, err := getParser(language)
	if err != nil {
		log.Printf("AST parser not available for language %s, falling back to pattern-only analysis", language)
		// Fallback to pattern-only analysis if AST not available
		return analyzeSecurityPatterns(code, language, filename, rulesToCheck, nil), nil
	}

	// Parse code to AST with error handling
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil || tree == nil {
		log.Printf("AST parsing failed for file %s (language: %s), falling back to pattern-only analysis: %v", filename, language, err)
		// Graceful degradation: fallback to pattern-only analysis on parse errors
		// This handles malformed code gracefully
		return analyzeSecurityPatterns(code, language, filename, rulesToCheck, nil), nil
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		log.Printf("AST root node is nil for file %s, falling back to pattern-only analysis", filename)
		return analyzeSecurityPatterns(code, language, filename, rulesToCheck, nil), nil
	}

	// Edge case: Very large AST (>100k nodes) - use sampling to prevent performance issues
	const maxASTNodes = 100000
	nodeCount := countASTNodes(rootNode)
	if nodeCount > maxASTNodes {
		log.Printf("Warning: Very large AST for file %s (%d nodes), analysis may be slower", filename, nodeCount)
		// For very large ASTs, we'll still analyze but log a warning
		// In practice, this is rare but we handle it gracefully
		// The analysis will proceed but may be slower
	}

	// Detect framework with confidence scoring
	frameworkDetection := detectFramework(code, language, rootNode)
	framework := frameworkDetection.Framework
	log.Printf("Detected framework: %s (confidence: %s) for file %s", framework, frameworkDetection.Confidence, filename)

	// Determine which rules to check
	rules := rulesToCheck
	if len(rules) == 0 {
		// Check all rules
		for ruleID := range SecurityRules {
			rules = append(rules, ruleID)
		}
		log.Printf("Checking all security rules (%d rules) for file %s", len(rules), filename)
	} else {
		log.Printf("Checking %d specific security rules for file %s: %v", len(rules), filename, rules)
	}

	// Check each rule
	for _, ruleID := range rules {
		rule, exists := SecurityRules[ruleID]
		if !exists {
			log.Printf("Warning: Security rule %s not found, skipping", ruleID)
			continue
		}

		log.Printf("Checking security rule %s (%s) for file %s", ruleID, rule.Name, filename)
		ruleFindings := checkSecurityRule(rule, code, language, filename, rootNode, framework)
		if len(ruleFindings) > 0 {
			log.Printf("Found %d violations for rule %s in file %s", len(ruleFindings), ruleID, filename)
		}
		findings = append(findings, ruleFindings...)
	}

	// Store in cache
	securityCacheMutex.Lock()
	securityCache[cacheKey] = &securityCacheEntry{
		Findings: findings,
		Expires:  time.Now().Add(securityCacheTTL),
	}
	securityCacheMutex.Unlock()

	duration := time.Since(startTime)
	log.Printf("Security analysis completed for file %s: %d findings in %v", filename, len(findings), duration)
	return findings, nil
}

// checkSecurityRule checks a specific security rule
func checkSecurityRule(rule SecurityRule, code string, language string, filename string, rootNode *sitter.Node, framework FrameworkType) []SecurityFinding {
	log.Printf("Checking rule %s (%s) using AST-based detection", rule.ID, rule.Name)
	var findings []SecurityFinding
	lines := strings.Split(code, "\n")

	// Pattern-based checks
	if len(rule.Detection.PatternsForbidden) > 0 {
		for _, pattern := range rule.Detection.PatternsForbidden {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatchIndex(code, -1)
			for _, match := range matches {
				lineNum := getLineNumber(code, match[0])
				if lineNum > 0 && lineNum <= len(lines) {
					lineCode := strings.TrimSpace(lines[lineNum-1])
					findings = append(findings, SecurityFinding{
						RuleID:      rule.ID,
						RuleName:    rule.Name,
						Severity:    rule.Severity,
						Line:        lineNum,
						Code:        lineCode,
						Issue:       fmt.Sprintf("%s: Forbidden pattern detected", rule.Name),
						Remediation: getRemediation(rule.ID),
						AutoFixable: isAutoFixable(rule.ID),
					})
				}
			}
		}
	}

	// AST-based checks
	if rule.ASTCheck != nil {
		astFindings := checkASTSecurityRule(rule, code, language, rootNode, framework)
		findings = append(findings, astFindings...)
	}

	// Data flow analysis for SEC-005 (Password Hashing)
	if rule.ID == "SEC-005" {
		dataFlowFindings := verifyPasswordFlow(rootNode, code, language)
		findings = append(findings, dataFlowFindings...)
	}

	return findings
}

// checkASTSecurityRule performs AST-based security checks
func checkASTSecurityRule(rule SecurityRule, code string, language string, rootNode *sitter.Node, framework FrameworkType) []SecurityFinding {
	var findings []SecurityFinding
	lines := strings.Split(code, "\n")

	if rule.ASTCheck == nil {
		return findings
	}

	// Check for required functions
	if len(rule.ASTCheck.FunctionContains) > 0 {
		for _, funcName := range rule.ASTCheck.FunctionContains {
			hasFunction := findFunctionInAST(rootNode, code, funcName, language)
			if hasFunction {
				// Check if required checks are present within the function scope
				if rule.ASTCheck.MustHaveBefore != "" {
					// Get the function scope to verify check is within the function
					funcScope := getFunctionScope(rootNode, code, funcName, language)
					hasCheck := findSecurityCheckInAST(rootNode, code, rule.ASTCheck.MustHaveBefore, language, funcScope)
					if !hasCheck {
						// Find the function location
						funcLine := findFunctionLine(rootNode, code, funcName, language)
						if funcLine > 0 {
							findings = append(findings, SecurityFinding{
								RuleID:      rule.ID,
								RuleName:    rule.Name,
								Severity:    rule.Severity,
								Line:        funcLine,
								Code:        strings.TrimSpace(lines[funcLine-1]),
								Issue:       fmt.Sprintf("%s: Missing required security check '%s' within function scope", rule.Name, rule.ASTCheck.MustHaveBefore),
								Remediation: getRemediation(rule.ID),
								AutoFixable: isAutoFixable(rule.ID),
							})
						}
					}
				}
			}
		}
	}

	// Check for required middleware
	if len(rule.ASTCheck.RouteMiddleware) > 0 && framework != FrameworkUnknown {
		hasMiddleware := findMiddlewareInAST(rootNode, code, rule.ASTCheck.RouteMiddleware, language, framework)
		if !hasMiddleware {
			// Find route definitions
			routeLines := findRouteDefinitions(rootNode, code, language, framework)
			for _, lineNum := range routeLines {
				if lineNum > 0 && lineNum <= len(lines) {
					findings = append(findings, SecurityFinding{
						RuleID:      rule.ID,
						RuleName:    rule.Name,
						Severity:    rule.Severity,
						Line:        lineNum,
						Code:        strings.TrimSpace(lines[lineNum-1]),
						Issue:       fmt.Sprintf("%s: Missing required middleware", rule.Name),
						Remediation: getRemediation(rule.ID),
						AutoFixable: isAutoFixable(rule.ID),
					})
				}
			}
		}
	}

	return findings
}

// verifyPatternInContext verifies that a pattern exists within a specific function scope
func verifyPatternInContext(rootNode *sitter.Node, code string, language string, funcName string, pattern string) bool {
	if rootNode == nil {
		// No AST available, fallback to simple pattern matching
		return regexp.MustCompile(pattern).MatchString(code)
	}

	// Get function scope
	funcScope := getFunctionScope(rootNode, code, funcName, language)
	if funcScope == nil {
		// Function not found, pattern check not applicable
		return false
	}

	// Check if pattern exists within function scope
	found := false
	traverseAST(rootNode, func(n *sitter.Node) bool {
		if found {
			return false
		}

		// Only check nodes within the function scope
		if !isNodeWithinScope(n, funcScope) {
			return true
		}

		// Check if pattern matches in this node
		nodeText := code[n.StartByte():n.EndByte()]
		re := regexp.MustCompile(pattern)
		if re.MatchString(nodeText) {
			found = true
			return false
		}
		return true
	})

	return found
}

// analyzeSecurityPatterns performs pattern-only security analysis (fallback)
// rootNode is optional - if provided, enables context-aware pattern matching
func analyzeSecurityPatterns(code string, language string, filename string, rulesToCheck []string, rootNode *sitter.Node) []SecurityFinding {
	var findings []SecurityFinding
	lines := strings.Split(code, "\n")

	// Determine which rules to check
	rules := rulesToCheck
	if len(rules) == 0 {
		for ruleID := range SecurityRules {
			rules = append(rules, ruleID)
		}
	}

	for _, ruleID := range rules {
		rule, exists := SecurityRules[ruleID]
		if !exists {
			continue
		}

		// Check forbidden patterns
		for _, pattern := range rule.Detection.PatternsForbidden {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatchIndex(code, -1)
			for _, match := range matches {
				lineNum := getLineNumber(code, match[0])
				if lineNum > 0 && lineNum <= len(lines) {
					findings = append(findings, SecurityFinding{
						RuleID:      rule.ID,
						RuleName:    rule.Name,
						Severity:    rule.Severity,
						Line:        lineNum,
						Code:        strings.TrimSpace(lines[lineNum-1]),
						Issue:       fmt.Sprintf("%s: Forbidden pattern detected", rule.Name),
						Remediation: getRemediation(rule.ID),
						AutoFixable: isAutoFixable(rule.ID),
					})
				}
			}
		}

		// Check for required patterns (if function contains trigger)
		// For SEC-005 (password hashing), verify patterns are within function scope
		if rule.ASTCheck != nil && len(rule.ASTCheck.FunctionContains) > 0 {
			for _, funcName := range rule.ASTCheck.FunctionContains {
				funcPattern := regexp.MustCompile(fmt.Sprintf("\\b%s\\s*\\(", regexp.QuoteMeta(funcName)))
				if funcPattern.MatchString(code) {
					// Check if required patterns are present (with context verification if AST available)
					hasRequired := false
					for _, requiredPattern := range rule.Detection.PatternsRequired {
						if rootNode != nil && rule.ID == "SEC-005" {
							// Use context-aware verification for password hashing
							hasRequired = verifyPatternInContext(rootNode, code, language, funcName, requiredPattern)
						} else {
							// Simple pattern matching
							re := regexp.MustCompile(requiredPattern)
							hasRequired = re.MatchString(code)
						}
						if hasRequired {
							break
						}
					}

					if !hasRequired && len(rule.Detection.PatternsRequired) > 0 {
						funcMatches := funcPattern.FindAllStringSubmatchIndex(code, -1)
						for _, match := range funcMatches {
							lineNum := getLineNumber(code, match[0])
							if lineNum > 0 && lineNum <= len(lines) {
								findings = append(findings, SecurityFinding{
									RuleID:      rule.ID,
									RuleName:    rule.Name,
									Severity:    rule.Severity,
									Line:        lineNum,
									Code:        strings.TrimSpace(lines[lineNum-1]),
									Issue:       fmt.Sprintf("%s: Missing required security pattern", rule.Name),
									Remediation: getRemediation(rule.ID),
									AutoFixable: isAutoFixable(rule.ID),
								})
							}
						}
					}
				}
			}
		}
	}

	return findings
}

// Helper functions for AST traversal

// getFunctionNameNode extracts function name from AST node
func getFunctionNameNode(node *sitter.Node, code string, language string) string {
	switch language {
	case "go":
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" || child.Type() == "field_identifier" {
						return code[child.StartByte():child.EndByte()]
					}
				}
			}
		}
	case "javascript", "typescript":
		if node.Type() == "function_declaration" || node.Type() == "function" || node.Type() == "arrow_function" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" || child.Type() == "property_identifier" {
						return code[child.StartByte():child.EndByte()]
					}
				}
			}
		} else if node.Type() == "call_expression" {
			// For call expressions, get the function being called
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					return code[child.StartByte():child.EndByte()]
				}
			}
		}
	case "python":
		if node.Type() == "function_definition" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					return code[child.StartByte():child.EndByte()]
				}
			}
		}
	}
	return ""
}

func findFunctionInAST(node *sitter.Node, code string, funcName string, language string) bool {
	found := false
	traverseAST(node, func(n *sitter.Node) bool {
		if found {
			return false
		}

		// Use AST node type matching instead of string matching
		extractedName := getFunctionNameNode(n, code, language)
		if extractedName == funcName {
			found = true
			return false
		}

		// Also check call expressions for function calls
		if n.Type() == "call_expression" {
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child != nil && child.Type() == "identifier" {
					callName := code[child.StartByte():child.EndByte()]
					if callName == funcName {
						found = true
						return false
					}
				}
			}
		}

		return true
	})
	return found
}

// getFunctionScope finds a function node by name and returns it (or nil if not found)
func getFunctionScope(rootNode *sitter.Node, code string, funcName string, language string) *sitter.Node {
	var funcNode *sitter.Node
	traverseAST(rootNode, func(n *sitter.Node) bool {
		if funcNode != nil {
			return false
		}

		// Check if this is a function node matching the name
		extractedName := getFunctionNameNode(n, code, language)
		if extractedName == funcName {
			// Verify it's actually a function declaration/definition
			switch language {
			case "go":
				if n.Type() == "function_declaration" || n.Type() == "method_declaration" {
					funcNode = n
					return false
				}
			case "javascript", "typescript":
				if n.Type() == "function_declaration" || n.Type() == "function" || n.Type() == "arrow_function" {
					funcNode = n
					return false
				}
			case "python":
				if n.Type() == "function_definition" {
					funcNode = n
					return false
				}
			}
		}
		return true
	})
	return funcNode
}

// isNodeWithinScope checks if a node is within the given function scope
func isNodeWithinScope(node *sitter.Node, scopeNode *sitter.Node) bool {
	if scopeNode == nil {
		return true // No scope restriction
	}

	current := node
	for current != nil {
		if current == scopeNode {
			return true
		}
		current = current.Parent()
	}
	return false
}

func findSecurityCheckInAST(node *sitter.Node, code string, checkName string, language string, scopeNode *sitter.Node) bool {
	found := false
	traverseAST(node, func(n *sitter.Node) bool {
		if found {
			return false
		}

		// If scope is specified, only check nodes within that scope
		if scopeNode != nil && !isNodeWithinScope(n, scopeNode) {
			return true // Continue traversal but skip this branch
		}

		nodeText := code[n.StartByte():n.EndByte()]
		if strings.Contains(nodeText, checkName) {
			found = true
			return false
		}
		return true
	})
	return found
}

// findMiddlewareInAST verifies middleware is actually applied to routes, not just imported
func findMiddlewareInAST(node *sitter.Node, code string, middlewareNames []string, language string, framework FrameworkType) bool {
	// First, find all route definitions to verify middleware is applied before them
	routeNodes := []*sitter.Node{}
	traverseAST(node, func(n *sitter.Node) bool {
		switch framework {
		case FrameworkExpress:
			if n.Type() == "call_expression" {
				// Check for app.get, app.post, router.get, etc.
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "app.get") || strings.Contains(nodeText, "app.post") ||
					strings.Contains(nodeText, "app.put") || strings.Contains(nodeText, "app.delete") ||
					strings.Contains(nodeText, "router.get") || strings.Contains(nodeText, "router.post") {
					routeNodes = append(routeNodes, n)
				}
			}
		case FrameworkGin:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "router.GET") || strings.Contains(nodeText, "router.POST") ||
					strings.Contains(nodeText, "router.PUT") || strings.Contains(nodeText, "router.DELETE") {
					routeNodes = append(routeNodes, n)
				}
			}
		case FrameworkFastAPI:
			if n.Type() == "decorator" {
				nodeText := code[n.StartByte():n.EndByte()]
				if strings.Contains(nodeText, "@app.get") || strings.Contains(nodeText, "@app.post") ||
					strings.Contains(nodeText, "@app.put") || strings.Contains(nodeText, "@app.delete") {
					routeNodes = append(routeNodes, n)
				}
			}
		}
		return true
	})

	// If no routes found, middleware check is not applicable
	if len(routeNodes) == 0 {
		return false
	}

	// Now check if middleware is applied before routes
	middlewareApplied := false
	traverseAST(node, func(n *sitter.Node) bool {
		if middlewareApplied {
			return false
		}

		// Check if this is a middleware application (not just import)
		switch framework {
		case FrameworkExpress:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				// Check for app.use() or router.use() with middleware name
				if strings.Contains(nodeText, "app.use") || strings.Contains(nodeText, "router.use") {
					for _, mwName := range middlewareNames {
						if strings.Contains(nodeText, mwName) {
							// Verify this middleware call appears before any route definition
							middlewarePos := int(n.StartByte())
							allBeforeRoutes := true
							for _, routeNode := range routeNodes {
								if middlewarePos >= int(routeNode.StartByte()) {
									allBeforeRoutes = false
									break
								}
							}
							if allBeforeRoutes {
								middlewareApplied = true
								return false
							}
						}
					}
				}
			}
		case FrameworkGin:
			if n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				// Check for router.Use() or engine.Use() with middleware name
				if strings.Contains(nodeText, ".Use(") {
					for _, mwName := range middlewareNames {
						if strings.Contains(nodeText, mwName) {
							middlewarePos := int(n.StartByte())
							allBeforeRoutes := true
							for _, routeNode := range routeNodes {
								if middlewarePos >= int(routeNode.StartByte()) {
									allBeforeRoutes = false
									break
								}
							}
							if allBeforeRoutes {
								middlewareApplied = true
								return false
							}
						}
					}
				}
			}
		case FrameworkFastAPI:
			if n.Type() == "decorator" || n.Type() == "call_expression" {
				nodeText := code[n.StartByte():n.EndByte()]
				// Check for @app.middleware or Depends() with middleware
				if strings.Contains(nodeText, "@app.middleware") || strings.Contains(nodeText, "Depends(") {
					for _, mwName := range middlewareNames {
						if strings.Contains(nodeText, mwName) {
							middlewarePos := int(n.StartByte())
							allBeforeRoutes := true
							for _, routeNode := range routeNodes {
								if middlewarePos >= int(routeNode.StartByte()) {
									allBeforeRoutes = false
									break
								}
							}
							if allBeforeRoutes {
								middlewareApplied = true
								return false
							}
						}
					}
				}
			}
		}
		return true
	})

	return middlewareApplied
}

// findRouteDefinitions uses AST node types to find route definitions
func findRouteDefinitions(node *sitter.Node, code string, language string, framework FrameworkType) []int {
	var lines []int
	traverseAST(node, func(n *sitter.Node) bool {
		switch framework {
		case FrameworkExpress:
			// Look for call_expression nodes with app.get, app.post, router.get, etc.
			if n.Type() == "call_expression" {
				// Get the function being called
				if n.ChildCount() > 0 {
					firstChild := n.Child(0)
					if firstChild != nil {
						// Check for member_expression (app.get, router.post, etc.)
						if firstChild.Type() == "member_expression" {
							// Get the property name (get, post, put, delete)
							if firstChild.ChildCount() >= 2 {
								propertyNode := firstChild.Child(1)
								if propertyNode != nil {
									propertyName := code[propertyNode.StartByte():propertyNode.EndByte()]
									if propertyName == "get" || propertyName == "post" || propertyName == "put" || propertyName == "delete" {
										lineNum := getLineNumber(code, int(n.StartByte()))
										if lineNum > 0 {
											lines = append(lines, lineNum)
										}
									}
								}
							}
						}
					}
				}
			}
		case FrameworkGin:
			// Look for call_expression nodes with router.GET, router.POST, etc.
			if n.Type() == "call_expression" {
				if n.ChildCount() > 0 {
					firstChild := n.Child(0)
					if firstChild != nil {
						if firstChild.Type() == "member_expression" {
							if firstChild.ChildCount() >= 2 {
								propertyNode := firstChild.Child(1)
								if propertyNode != nil {
									propertyName := code[propertyNode.StartByte():propertyNode.EndByte()]
									if propertyName == "GET" || propertyName == "POST" || propertyName == "PUT" || propertyName == "DELETE" {
										lineNum := getLineNumber(code, int(n.StartByte()))
										if lineNum > 0 {
											lines = append(lines, lineNum)
										}
									}
								}
							}
						}
					}
				}
			}
		case FrameworkFastAPI:
			// Look for decorator nodes with @app.get, @app.post, etc.
			if n.Type() == "decorator" {
				nodeText := code[n.StartByte():n.EndByte()]
				// Check for @app.get, @app.post, etc. patterns
				if strings.Contains(nodeText, "@app.get") || strings.Contains(nodeText, "@app.post") ||
					strings.Contains(nodeText, "@app.put") || strings.Contains(nodeText, "@app.delete") {
					lineNum := getLineNumber(code, int(n.StartByte()))
					if lineNum > 0 {
						lines = append(lines, lineNum)
					}
				}
			}
		}
		return true
	})
	return lines
}

func findFunctionLine(node *sitter.Node, code string, funcName string, language string) int {
	var lineNum int
	traverseAST(node, func(n *sitter.Node) bool {
		if lineNum > 0 {
			return false
		}

		nodeText := code[n.StartByte():n.EndByte()]
		if strings.Contains(nodeText, funcName) {
			if n.Type() == "function_declaration" || n.Type() == "function" || n.Type() == "method_definition" {
				lineNum = getLineNumber(code, int(n.StartByte()))
				return false
			}
		}
		return true
	})
	return lineNum
}

func getLineNumber(code string, byteOffset int) int {
	line := 1
	for i := 0; i < byteOffset && i < len(code); i++ {
		if code[i] == '\n' {
			line++
		}
	}
	return line
}

func getRemediation(ruleID string) string {
	remediations := map[string]string{
		"SEC-001": "Add ownership verification: Check that req.user.id matches resource.userId before returning data",
		"SEC-002": "Use parameterized queries: Replace string concatenation with $1, $2 placeholders or prepared statements",
		"SEC-003": "Add authentication middleware: Use auth middleware on all protected routes",
		"SEC-004": "Add rate limiting: Implement rate limiting middleware on API endpoints",
		"SEC-005": "Use secure hashing: Replace MD5/SHA1 with bcrypt, argon2, or scrypt",
		"SEC-006": "Add input validation: Use validation library (joi, yup, zod, etc.) to validate user input",
		"SEC-007": "Add secure headers: Use helmet or similar middleware to set secure HTTP headers",
		"SEC-008": "Configure CORS properly: Set specific allowed origins instead of wildcard (*)",
	}
	return remediations[ruleID]
}

func isAutoFixable(ruleID string) bool {
	// SEC-002 (SQL injection) and SEC-008 (CORS) may be partially auto-fixable
	return ruleID == "SEC-002" || ruleID == "SEC-008"
}

// calculateSecurityScore calculates security score (0-100) and grade
func calculateSecurityScore(findings []SecurityFinding) (int, string) {
	if len(findings) == 0 {
		return 100, "A"
	}

	// Weight findings by severity
	totalPenalty := 0
	criticalCount := 0

	for _, finding := range findings {
		switch finding.Severity {
		case "critical":
			totalPenalty += 20
			criticalCount++
		case "high":
			totalPenalty += 10
		case "medium":
			totalPenalty += 5
		case "low":
			totalPenalty += 2
		}
	}

	// Cap penalty at 100
	if totalPenalty > 100 {
		totalPenalty = 100
	}

	score := 100 - totalPenalty
	if score < 0 {
		score = 0
	}

	// Determine grade
	var grade string
	switch {
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	default:
		grade = "F"
	}

	// Adjust grade down if critical issues exist
	if criticalCount > 0 && grade == "A" {
		grade = "B"
	}
	if criticalCount >= 3 {
		grade = "F"
	}

	return score, grade
}

// calculateSecuritySummary calculates summary statistics
func calculateSecuritySummary(findings []SecurityFinding) struct {
	TotalRules int `json:"totalRules"`
	Passed     int `json:"passed"`
	Failed     int `json:"failed"`
	Critical   int `json:"critical"`
	High       int `json:"high"`
	Medium     int `json:"medium"`
	Low        int `json:"low"`
} {
	totalRules := len(SecurityRules)

	// Count findings by severity
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	// Track which rules failed
	failedRules := make(map[string]bool)
	for _, finding := range findings {
		failedRules[finding.RuleID] = true

		switch finding.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		case "medium":
			mediumCount++
		case "low":
			lowCount++
		}
	}

	failedCount := len(failedRules)
	passedCount := totalRules - failedCount

	return struct {
		TotalRules int `json:"totalRules"`
		Passed     int `json:"passed"`
		Failed     int `json:"failed"`
		Critical   int `json:"critical"`
		High       int `json:"high"`
		Medium     int `json:"medium"`
		Low        int `json:"low"`
	}{
		TotalRules: totalRules,
		Passed:     passedCount,
		Failed:     failedCount,
		Critical:   criticalCount,
		High:       highCount,
		Medium:     mediumCount,
		Low:        lowCount,
	}
}

// buildDataFlowGraph builds a data flow graph from AST
// Maps variable names to their assignments and usages
func buildDataFlowGraph(rootNode *sitter.Node, code string, language string) *DataFlowAnalyzer {
	analyzer := &DataFlowAnalyzer{
		variables: make(map[string]*VariableInfo),
		functions: make(map[string]*FunctionInfo),
	}

	traverseAST(rootNode, func(node *sitter.Node) bool {
		switch language {
		case "javascript", "typescript":
			// Track variable declarations
			if node.Type() == "variable_declaration" || node.Type() == "lexical_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "variable_declarator" || child.Type() == "identifier") {
						varName := ""
						var lineNum int

						// Extract variable name
						for j := 0; j < int(child.ChildCount()); j++ {
							grandchild := child.Child(j)
							if grandchild != nil && grandchild.Type() == "identifier" {
								varName = code[grandchild.StartByte():grandchild.EndByte()]
								lineNum = getLineNumber(code, int(grandchild.StartByte()))
								break
							}
						}

						if varName != "" {
							// Check if it's a password-related variable
							varType := "other"
							varNameLower := strings.ToLower(varName)
							if strings.Contains(varNameLower, "password") || strings.Contains(varNameLower, "passwd") || strings.Contains(varNameLower, "pwd") {
								varType = "password"
							}

							if analyzer.variables[varName] == nil {
								analyzer.variables[varName] = &VariableInfo{
									Name:        varName,
									Type:        varType,
									Assignments: []Assignment{},
									Usages:      []Usage{},
									Line:        lineNum,
								}
							}

							// Track assignment
							analyzer.variables[varName].Assignments = append(analyzer.variables[varName].Assignments, Assignment{
								Line:    lineNum,
								Value:   code[child.StartByte():child.EndByte()],
								Source:  "variable_declaration",
								Context: "global", // Will be enhanced with scope tracking
							})
						}
					}
				}
			}

			// Track function calls that might be password hashing
			if node.Type() == "call_expression" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						funcName := code[child.StartByte():child.EndByte()]
						funcNameLower := strings.ToLower(funcName)

						// Check for secure hashing functions
						if strings.Contains(funcNameLower, "bcrypt") || strings.Contains(funcNameLower, "argon2") ||
							strings.Contains(funcNameLower, "scrypt") || strings.Contains(funcNameLower, "pbkdf2") {
							// Track this as a secure hashing usage
							// We'll link this to password variables in verifyPasswordFlow
							_ = getLineNumber(code, int(node.StartByte()))
						}

						// Check for insecure hashing functions
						if strings.Contains(funcNameLower, "md5") || strings.Contains(funcNameLower, "sha1") {
							// Will be used in verifyPasswordFlow
							_ = getLineNumber(code, int(node.StartByte()))
						}
					}
				}
			}

		case "python":
			// Track variable assignments
			if node.Type() == "assignment" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						varName := code[child.StartByte():child.EndByte()]
						lineNum := getLineNumber(code, int(child.StartByte()))

						varType := "other"
						varNameLower := strings.ToLower(varName)
						if strings.Contains(varNameLower, "password") || strings.Contains(varNameLower, "passwd") {
							varType = "password"
						}

						if analyzer.variables[varName] == nil {
							analyzer.variables[varName] = &VariableInfo{
								Name:        varName,
								Type:        varType,
								Assignments: []Assignment{},
								Usages:      []Usage{},
								Line:        lineNum,
							}
						}

						analyzer.variables[varName].Assignments = append(analyzer.variables[varName].Assignments, Assignment{
							Line:    lineNum,
							Value:   code[node.StartByte():node.EndByte()],
							Source:  "assignment",
							Context: "global",
						})
					}
				}
			}

		case "go":
			// Track variable declarations
			if node.Type() == "short_var_declaration" || node.Type() == "var_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						varName := code[child.StartByte():child.EndByte()]
						lineNum := getLineNumber(code, int(child.StartByte()))

						varType := "other"
						varNameLower := strings.ToLower(varName)
						if strings.Contains(varNameLower, "password") || strings.Contains(varNameLower, "passwd") {
							varType = "password"
						}

						if analyzer.variables[varName] == nil {
							analyzer.variables[varName] = &VariableInfo{
								Name:        varName,
								Type:        varType,
								Assignments: []Assignment{},
								Usages:      []Usage{},
								Line:        lineNum,
							}
						}
					}
				}
			}
		}

		return true
	})

	return analyzer
}

// trackVariableFlow tracks how a variable flows through the code
func trackVariableFlow(analyzer *DataFlowAnalyzer, varName string, code string, language string) []Usage {
	usages := []Usage{}

	if analyzer.variables[varName] == nil {
		return usages
	}

	// This is a simplified version - full implementation would track
	// all usages of the variable through assignments, function calls, etc.
	// For now, we'll rely on verifyPasswordFlow to do the actual checking

	return usages
}

// verifyPasswordFlow verifies that password variables flow to secure hashing functions
func verifyPasswordFlow(rootNode *sitter.Node, code string, language string) []SecurityFinding {
	log.Printf("Starting password flow analysis (SEC-005 data flow check)")
	var findings []SecurityFinding
	lines := strings.Split(code, "\n")

	// Build data flow graph
	log.Printf("Building data flow graph for password variables")
	analyzer := buildDataFlowGraph(rootNode, code, language)
	log.Printf("Data flow graph built: %d variables, %d functions tracked", len(analyzer.variables), len(analyzer.functions))

	// Find all password-related variables
	passwordVars := []string{}
	for varName, varInfo := range analyzer.variables {
		if varInfo.Type == "password" {
			passwordVars = append(passwordVars, varName)
		}
	}

	if len(passwordVars) == 0 {
		log.Printf("No password variables found in data flow graph")
		return findings // No password variables found
	}

	log.Printf("Found %d password variables to analyze: %v", len(passwordVars), passwordVars)

	// Check if password variables are used with insecure hashing
	// Look for patterns like: md5(password), sha1(password), etc.
	insecurePatterns := []string{
		"md5", "sha1", "crypto\\.createHash.*md5", "crypto\\.createHash.*sha1",
		"hashlib\\.md5", "hashlib\\.sha1",
	}

	securePatterns := []string{
		"bcrypt", "argon2", "scrypt", "pbkdf2",
		"bcrypt\\.hash", "bcrypt\\.hashSync", "argon2\\.hash",
	}

	// Check each password variable
	for _, varName := range passwordVars {
		varInfo := analyzer.variables[varName]

		// Check if password is used with insecure hashing
		for _, assignment := range varInfo.Assignments {
			lineCode := ""
			if assignment.Line > 0 && assignment.Line <= len(lines) {
				lineCode = lines[assignment.Line-1]
			}

			// Check for insecure patterns
			for _, pattern := range insecurePatterns {
				re := regexp.MustCompile("(?i)" + pattern)
				if re.MatchString(lineCode) {
					// Check if this password variable is in the same context
					if strings.Contains(lineCode, varName) {
						findings = append(findings, SecurityFinding{
							RuleID:      "SEC-005",
							RuleName:    "Password Hashing",
							Severity:    "critical",
							Line:        assignment.Line,
							Code:        strings.TrimSpace(lineCode),
							Issue:       fmt.Sprintf("Password variable '%s' is being hashed with insecure algorithm (%s)", varName, pattern),
							Remediation: getRemediation("SEC-005"),
							AutoFixable: false,
						})
					}
				}
			}
		}

		// Check if password is used with secure hashing (positive check)
		hasSecureHash := false
		for _, assignment := range varInfo.Assignments {
			lineCode := ""
			if assignment.Line > 0 && assignment.Line <= len(lines) {
				lineCode = lines[assignment.Line-1]
			}

			for _, pattern := range securePatterns {
				re := regexp.MustCompile("(?i)" + pattern)
				if re.MatchString(lineCode) && strings.Contains(lineCode, varName) {
					hasSecureHash = true
					break
				}
			}
		}

		// If password variable exists but no secure hashing is found, flag it
		if !hasSecureHash && len(varInfo.Assignments) > 0 {
			// Only flag if it's actually being used (not just declared)
			// Check if it's assigned from user input (req.body.password, etc.)
			userInputPatterns := []string{
				"req\\.body", "req\\.query", "req\\.params",
				"request\\.body", "request\\.query", "request\\.params",
				"form\\.get", "form\\[", "body\\[",
			}

			hasUserInput := false
			for _, assignment := range varInfo.Assignments {
				lineCode := ""
				if assignment.Line > 0 && assignment.Line <= len(lines) {
					lineCode = lines[assignment.Line-1]
				}

				for _, pattern := range userInputPatterns {
					re := regexp.MustCompile("(?i)" + pattern)
					if re.MatchString(lineCode) {
						hasUserInput = true
						break
					}
				}
			}

			if hasUserInput {
				// Password from user input but no secure hashing found
				firstAssignment := varInfo.Assignments[0]
				findings = append(findings, SecurityFinding{
					RuleID:      "SEC-005",
					RuleName:    "Password Hashing",
					Severity:    "critical",
					Line:        firstAssignment.Line,
					Code:        strings.TrimSpace(lines[firstAssignment.Line-1]),
					Issue:       fmt.Sprintf("Password variable '%s' from user input is not being hashed with secure algorithm (bcrypt, argon2, scrypt, or pbkdf2)", varName),
					Remediation: getRemediation("SEC-005"),
					AutoFixable: false,
				})
			}
		}
	}

	return findings
}

// DetectionMetrics tracks detection rate validation metrics
type DetectionMetrics struct {
	TruePositives  int     `json:"truePositives"`
	FalsePositives int     `json:"falsePositives"`
	FalseNegatives int     `json:"falseNegatives"`
	TrueNegatives  int     `json:"trueNegatives"`
	DetectionRate  float64 `json:"detectionRate"` // Percentage
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
}

// calculateDetectionRate calculates detection metrics from ground truth labels
func calculateDetectionRate(findings []SecurityFinding, expectedFindings map[string]bool) DetectionMetrics {
	metrics := DetectionMetrics{
		TruePositives:  0,
		FalsePositives: 0,
		FalseNegatives: 0,
		TrueNegatives:  0,
	}

	// Track which rules were detected
	detectedRules := make(map[string]bool)
	for _, finding := range findings {
		detectedRules[finding.RuleID] = true
	}

	// Compare against expected findings
	for ruleID, shouldDetect := range expectedFindings {
		wasDetected := detectedRules[ruleID]

		if shouldDetect && wasDetected {
			metrics.TruePositives++
		} else if !shouldDetect && wasDetected {
			metrics.FalsePositives++
		} else if shouldDetect && !wasDetected {
			metrics.FalseNegatives++
		} else {
			metrics.TrueNegatives++
		}
	}

	// Calculate detection rate
	total := metrics.TruePositives + metrics.FalsePositives + metrics.FalseNegatives + metrics.TrueNegatives
	if total > 0 {
		metrics.DetectionRate = float64(metrics.TruePositives+metrics.TrueNegatives) / float64(total) * 100.0
	}

	// Calculate precision
	if metrics.TruePositives+metrics.FalsePositives > 0 {
		metrics.Precision = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalsePositives) * 100.0
	}

	// Calculate recall
	if metrics.TruePositives+metrics.FalseNegatives > 0 {
		metrics.Recall = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalseNegatives) * 100.0
	}

	return metrics
}

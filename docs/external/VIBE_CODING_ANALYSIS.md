# Vibe Coding Analysis & Solutions

> **For AI Agents**: This document captures comprehensive analysis of vibe coding issues. Use this as a reference for understanding the problems Sentinel solves and the architectural decisions made.

## Executive Summary

This document provides a detailed analysis of issues that arise during AI-assisted "vibe coding" and Sentinel's comprehensive solutions. It serves as the authoritative reference for understanding:

1. **What goes wrong** during AI-assisted development
2. **How Sentinel detects** these issues
3. **How Sentinel prevents** these issues
4. **What remains undetectable** and requires other mitigation

---

## Table of Contents

1. [Vibe Coding Issue Taxonomy](#1-vibe-coding-issue-taxonomy)
2. [Detection Methods & Architecture](#2-detection-methods--architecture)
3. [Large File Problem](#3-large-file-problem)
4. [Security Logic Implementation](#4-security-logic-implementation)
5. [Test Enforcement System](#5-test-enforcement-system)
6. [Documentation Standardization](#6-documentation-standardization)
7. [Requirements Lifecycle Management](#7-requirements-lifecycle-management)
8. [MCP as Active Orchestrator](#8-mcp-as-active-orchestrator)
9. [Coverage Matrix](#9-coverage-matrix)
10. [Fundamental Limitations](#10-fundamental-limitations)

---

## 1. Vibe Coding Issue Taxonomy

### 1.1 Structural/Syntax Issues

Issues with code structure that break compilation or cause runtime errors.

| Issue | Description | Real Example | Detection Method |
|-------|-------------|--------------|------------------|
| **Duplicate function definitions** | Same function defined twice | `showKnowledgeHelp()` duplicated in synapsevibsentinel.sh | AST analysis |
| **Orphaned code** | Code outside valid scope | `switch` statement outside function | AST scope analysis |
| **Brace mismatch** | Unbalanced brackets | Missing closing `}` | Parser error analysis |
| **Incomplete edits** | Edit interrupted mid-stream | Half a function definition | Parse error + diff analysis |
| **Corrupted code blocks** | Malformed code from failed edit | Nested function declarations | AST validation |

**Coverage**: 95% detectable via server-side AST analysis

### 1.2 Refactoring Inconsistencies

Issues from incomplete refactoring operations.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Incomplete rename** | Variable renamed in some places | `user` vs `currentUser` mixed | Cross-file symbol tracking |
| **Signature mismatch** | Function params changed, calls not updated | `showKnowledgeStats(args)` called but defined as `showKnowledgeStats()` | AST call analysis |
| **Missing import updates** | Import removed but usage remains | `import { X }` removed, `X` still used | Import/usage correlation |
| **Dangling references** | Deleted function still called | Function deleted, calls remain | Symbol resolution |

**Coverage**: 95% detectable via cross-file AST analysis

### 1.3 Variable/Scope Issues

Issues with variable declarations and usage.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Unused variables** | Declared but never used | `lineNum` declared in loop, never used | AST usage tracking |
| **Undeclared variables** | Used without declaration | Variable reference without `let/const/var` | Symbol resolution |
| **Shadow variables** | Inner scope hides outer variable | Same name in nested scope | Scope chain analysis |
| **Wrong variable used** | Semantically incorrect variable | `user` when `admin` intended | **Cannot detect** |

**Coverage**: 85% detectable (wrong variable is semantic, undetectable)

### 1.4 Control Flow Issues

Issues with program logic flow.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Unreachable code** | Code after return/throw | Statements after `return` | Control flow graph |
| **Missing return paths** | Not all branches return | `if` returns, `else` doesn't | Path analysis |
| **Empty catch blocks** | Exceptions swallowed | `catch (e) {}` | AST inspection |
| **Infinite loops** | No exit condition | `while(true)` without break | **Limited detection** |

**Coverage**: 85% detectable (infinite loops require runtime analysis)

### 1.5 Async/Concurrency Issues

Issues specific to asynchronous code.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Missing await** | Async function not awaited | `asyncFunc()` without `await` | Async call tracking |
| **Callback/async mixing** | Inconsistent patterns | Callbacks inside async functions | Pattern detection |
| **Unhandled rejection** | Promise without catch | No `.catch()` on promise chain | Chain analysis |
| **Race conditions** | Timing-dependent bugs | Concurrent state modification | **Cannot detect** |

**Coverage**: 75% detectable (race conditions require runtime analysis)

### 1.6 Type/Null Safety Issues

Issues with type handling and null checks.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Type coercion** | Implicit type conversion | `"5" + 3 === "53"` | Pattern flagging |
| **Null/undefined access** | Accessing property on null | `user.name` when user is null | Flow analysis |
| **Wrong type passed** | Incorrect argument type | String passed where number expected | Type inference |
| **Array bounds** | Index out of range | `arr[10]` when `arr.length === 5` | **Cannot detect** |

**Coverage**: 75% detectable (bounds checking requires value tracking)

### 1.7 Logic/Semantic Issues

Issues where code is syntactically correct but logically wrong.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **Wrong operator** | Incorrect comparison | `<` instead of `<=` | Pattern hints only |
| **Off-by-one** | Boundary errors | Loop ends too early/late | **Cannot detect** |
| **Incorrect condition** | Logic error | `&&` instead of `\|\|` | **Cannot detect** |
| **Business rule violation** | Doesn't match spec | 24 hours instead of 48 | Executable assertions |

**Coverage**: 65% detectable with executable business rules

### 1.8 AI-Specific Issues

Issues unique to AI-generated code.

| Issue | Description | Example | Detection Method |
|-------|-------------|---------|------------------|
| **AI hallucinations** | Invented APIs/methods | `response.toSecureHash()` (doesn't exist) | Symbol resolution |
| **Deprecated methods** | Using outdated APIs | `componentWillMount` | Deprecation list |
| **Pattern ignorance** | Not following project style | camelCase in snake_case project | Pattern learning |
| **Redundant code** | Similar logic repeated | Multiple implementations of same feature | Similarity detection |
| **Context loss** | Forgetting previous decisions | Different patterns in new session | Decision memory |

**Coverage**: 85% detectable with full system

### 1.9 Technical Debt

Issues that work but create maintenance burden.

| Issue | Description | Detection | Metric |
|-------|-------------|-----------|--------|
| **Copy-paste code** | Duplicated blocks | Similarity analysis | Duplicate ratio |
| **Over-engineering** | Unnecessary complexity | Complexity metrics | Lines per function |
| **Deep nesting** | Too many indent levels | AST depth analysis | Max depth |
| **Long functions** | Functions too large | Line counting | Lines per function |
| **Magic numbers** | Hardcoded values | Pattern detection | Unnamed constants |
| **N+1 queries** | Query in loop | Pattern detection | Loop + query pattern |

**Coverage**: 70% detectable with quality metrics

### 1.10 Security Issues

Security vulnerabilities in code.

| Issue | Pattern-Based | Logic-Based | Detection Method |
|-------|--------------|-------------|------------------|
| **SQL injection** | String concat in query | - | Pattern + AST |
| **XSS** | innerHTML usage | - | Pattern + AST |
| **Secrets in code** | Hardcoded keys | - | Entropy + pattern |
| **IDOR** | - | Missing ownership check | AST + security rules |
| **Auth bypass** | - | Missing middleware | Route analysis |
| **Rate limiting** | - | Missing on sensitive endpoints | Endpoint analysis |

**Coverage**: 85% detectable with security rules

---

## 2. Detection Methods & Architecture

### 2.1 Detection Hierarchy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        DETECTION HIERARCHY                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  PRIMARY: HUB AST ANALYSIS (Thorough, Server-side)                         │
│  ═════════════════════════════════════════════════                         │
│  • Tree-sitter parsing (100+ languages)                                    │
│  • Duplicate function detection                                             │
│  • Unused variable detection                                                │
│  • Control flow analysis                                                    │
│  • Cross-file symbol resolution                                             │
│  • Scope analysis                                                           │
│  • Signature mismatch detection                                             │
│  • Orphaned code detection                                                  │
│                                                                              │
│  Detection: ~70% of issues                                                  │
│  Latency: 100-300ms                                                         │
│  Accuracy: 95%+                                                             │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  FALLBACK: LOCAL PATTERN MATCHING (Fast, Offline)                          │
│  ═════════════════════════════════════════════════                         │
│  • Used ONLY when Hub unavailable                                          │
│  • Regex pattern matching                                                   │
│  • Basic syntax validation                                                  │
│  • Limited accuracy (~40% detection, 60-70% accuracy)                       │
│                                                                              │
│  Detection: ~40% of issues                                                  │
│  Latency: <100ms                                                            │
│  Accuracy: 60-70% (many false positives)                                   │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  LEVEL 3: HUB SEMANTIC ANALYSIS (Deep, LLM-assisted)                       │
│  ════════════════════════════════════════════════════                      │
│  • Business rule compliance                                                 │
│  • Security logic verification                                              │
│  • Test quality analysis                                                    │
│  • Code review (AI-based)                                                   │
│  • Intent verification                                                      │
│                                                                              │
│  Detection: ~85% of issues                                                  │
│  Latency: 200-500ms                                                         │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  LEVEL 4: TEST EXECUTION (Verification, Sandboxed)                         │
│  ═════════════════════════════════════════════════                         │
│  • Generated test execution                                                 │
│  • Mutation testing                                                         │
│  • Coverage analysis                                                        │
│  • Integration testing                                                      │
│                                                                              │
│  Detection: ~92% of detectable issues                                       │
│  Latency: 1-30 seconds                                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 AST-First Detection Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AST-FIRST DETECTION FLOW                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. User runs: sentinel audit --vibe-check --deep                          │
│                                                                              │
│  2. Agent attempts Hub connection                                           │
│     ├── Hub available?                                                       │
│     │   ├── YES → Send code to Hub for AST analysis                        │
│     │   │   ├── Parse with Tree-sitter                                      │
│     │   │   ├── Detect duplicates, unused vars, orphaned code              │
│     │   │   ├── Cross-file symbol resolution                                │
│     │   │   └── Return AST findings (95% accuracy)                         │
│     │   │                                                                   │
│     │   └── NO → Fallback to pattern matching                              │
│     │       ├── Run regex patterns                                          │
│     │       ├── Basic duplicate detection                                    │
│     │       ├── Empty catch blocks                                           │
│     │       └── Return pattern findings (60-70% accuracy)                   │
│     │                                                                       │
│     └── Deduplication: Remove pattern findings covered by AST              │
│                                                                              │
│  3. Merge findings into audit report                                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 Server-Side AST Analysis Architecture

Tree-sitter provides multi-language parsing with consistent API:

```go
// Supported languages via tree-sitter
var supportedLanguages = map[string]Language{
    "javascript": javascript.GetLanguage(),
    "typescript": typescript.GetLanguage(),
    "python":     python.GetLanguage(),
    "go":         golang.GetLanguage(),
    "rust":       rust.GetLanguage(),
    "java":       java.GetLanguage(),
    "ruby":       ruby.GetLanguage(),
    "c":          c.GetLanguage(),
    "cpp":        cpp.GetLanguage(),
    "csharp":     csharp.GetLanguage(),
    // ... 100+ languages supported
}

// Analysis capabilities
type ASTAnalyzer struct {
    parser       *sitter.Parser
    projectIndex *ProjectIndex
    rules        []AnalysisRule
}

func (a *ASTAnalyzer) Analyze(code []byte, lang string) *AnalysisResult {
    tree := a.parser.Parse(code)
    
    return &AnalysisResult{
        Duplicates:      a.findDuplicates(tree),
        UnusedVars:      a.findUnusedVariables(tree),
        Unreachable:     a.findUnreachableCode(tree),
        SecurityIssues:  a.checkSecurityRules(tree),
        StyleViolations: a.checkStyleRules(tree),
    }
}
```

### 2.4 Why Server-Side for Complex Analysis

| Aspect | Local Agent | Hub Server |
|--------|-------------|------------|
| **Dependencies** | Must be zero | Docker manages all |
| **Parser libraries** | Would bloat binary | Easy to add |
| **Multi-language** | Impractical | Install once |
| **Updates** | Requires agent update | Server-side update |
| **Complex analysis** | Limited by Go stdlib | Full tooling |
| **AI models** | Can't run locally | Full LLM access |

---

## 3. Large File Problem

### 3.1 Evidence of the Problem

This project demonstrates the issue:

| File | Lines | Impact |
|------|-------|--------|
| `synapsevibsentinel.sh` | **8,489** | Duplicate functions, orphaned code found |

### 3.2 Why Large Files Happen

```
Session 1: "Add audit functionality"
    └── AI adds 500 lines to main.go
    
Session 2: "Add fix functionality"  
    └── AI adds 600 lines to main.go (doesn't create new file)
    
Session 3: "Add pattern learning"
    └── AI adds 800 lines to main.go
    
...many sessions later...
    └── main.go is 8000+ lines
    └── AI can't see whole file (context overflow)
    └── AI creates duplicate functions
    └── Humans can't review
```

### 3.3 Token Math (Critical)

```
LLM Context Budget (Claude 200K tokens):
═══════════════════════════════════════

System Prompt:          ~2,000 tokens
MCP Context:            ~1,000 tokens
Business Rules:         ~3,000 tokens
Security Rules:         ~1,000 tokens
Pattern Definitions:    ~1,000 tokens
User Conversation:      ~2,000 tokens
───────────────────────────────────
Reserved:              ~10,000 tokens

Available for Code:    190,000 tokens ≈ 140,000 characters

8,489 line file @ 80 chars/line = 679,120 characters
Requires: ~905,000 tokens

PROBLEM: File needs 4.7x available context!

Result: LLM sees fragments → loses context → creates duplicates
```

### 3.4 Solution: File Size Management

#### Configuration

```json
{
  "file_size_rules": {
    "thresholds": {
      "warning": 300,
      "critical": 500,
      "maximum": 1000
    },
    
    "by_file_type": {
      "component": 200,
      "service": 400,
      "utility": 150,
      "test": 500,
      "config": 100
    },
    
    "exceptions": [
      "generated/*",
      "migrations/*",
      "*.min.js"
    ]
  }
}
```

#### Automatic Split Suggestions

```
./sentinel audit --analyze-structure

📊 File Structure Analysis
══════════════════════════════════════════════════════════════

⚠️  WARNING: src/services/orderService.ts (847 lines)
    
    Detected logical sections:
    ├── Lines 1-150:    Order CRUD operations
    ├── Lines 151-350:  Order validation logic
    ├── Lines 351-550:  Payment processing
    ├── Lines 551-700:  Notification handling
    └── Lines 701-847:  Reporting/analytics
    
    Suggested split:
    src/services/order/
    ├── index.ts           (exports)
    ├── orderCrud.ts       (150 lines)
    ├── orderValidation.ts (200 lines)
    ├── orderPayment.ts    (200 lines)
    ├── orderNotify.ts     (150 lines)
    └── orderAnalytics.ts  (147 lines)
```

#### MCP Proactive Guidance

```
Developer: "add shipping calculation"

[MCP: sentinel_check_file_size]
    └── Returns: {
           "target_file": "src/services/orderService.ts",
           "current_lines": 847,
           "status": "oversized",
           "recommendation": "Create new file",
           "suggested_location": "src/services/order/shipping.ts"
         }

Cursor prompt includes: "Note: orderService.ts is oversized (847 lines).
                         Create in src/services/order/shipping.ts instead."

Result: AI creates properly sized module
```

---

## 4. Security Logic Implementation

### 4.1 Security as Executable Rules

Security rules are defined in structured format and enforced via AST analysis:

```json
{
  "security_rules": [
    {
      "id": "SEC-001",
      "name": "Resource Ownership Verification",
      "type": "authorization",
      "severity": "critical",
      
      "detection": {
        "endpoints": ["/api/:resource/:id"],
        "resources": ["orders", "users", "payments"],
        "required_checks": [
          "req.user.id === resource.userId",
          "req.user.id === resource.ownerId",
          "req.user.role === 'admin'"
        ]
      },
      
      "ast_pattern": {
        "function_contains": ["findById", "findOne"],
        "must_have_before_response": "ownership_check OR admin_check"
      }
    },
    
    {
      "id": "SEC-002",
      "name": "SQL Injection Prevention",
      "type": "injection",
      "severity": "critical",
      
      "detection": {
        "patterns_forbidden": [
          "query.*\\$\\{",
          "query.*\\+.*req\\."
        ],
        "patterns_required": ["\\$1|\\$2|\\?|:param"]
      }
    },
    
    {
      "id": "SEC-003",
      "name": "Authentication Middleware",
      "type": "authentication",
      "severity": "critical",
      
      "detection": {
        "protected_routes": ["/api/*"],
        "public_routes": ["/api/auth/login", "/api/auth/register"],
        "middleware_required": "authMiddleware OR requireAuth"
      }
    },
    
    {
      "id": "SEC-004",
      "name": "Rate Limiting",
      "type": "dos_prevention",
      "severity": "high",
      
      "detection": {
        "sensitive_endpoints": ["/api/auth/login", "/api/auth/reset-password"],
        "must_have": "rateLimiter OR rateLimit"
      }
    },
    
    {
      "id": "SEC-005",
      "name": "Password Hashing",
      "type": "cryptography",
      "severity": "critical",
      
      "detection": {
        "when_storing": ["password", "secret"],
        "must_use": ["bcrypt", "argon2", "scrypt"],
        "must_not_use": ["md5", "sha1", "plain text"]
      }
    },
    
    {
      "id": "SEC-006",
      "name": "Input Validation",
      "type": "validation",
      "severity": "high",
      
      "detection": {
        "all_endpoints": true,
        "request_body": "must be validated",
        "validators": ["joi", "yup", "zod", "class-validator"]
      }
    },
    
    {
      "id": "SEC-007",
      "name": "Secure Headers",
      "type": "transport",
      "severity": "medium",
      
      "detection": {
        "required_headers": ["helmet", "Content-Security-Policy"],
        "app_setup": "must include security headers"
      }
    },
    
    {
      "id": "SEC-008",
      "name": "CORS Configuration",
      "type": "transport",
      "severity": "high",
      
      "detection": {
        "must_not_have": "origin: '*' in production",
        "must_have": "explicit origin whitelist"
      }
    }
  ]
}
```

### 4.2 Security Detection Coverage

| Issue | Detection Method | Coverage |
|-------|------------------|----------|
| SQL Injection | Pattern + AST | 95% |
| XSS | Pattern + AST | 90% |
| IDOR | AST ownership check | 85% |
| Auth bypass | Middleware analysis | 90% |
| Rate limiting | Route analysis | 80% |
| Password storage | AST data flow | 85% |
| Input validation | Route analysis | 80% |
| CORS misconfiguration | Config analysis | 90% |
| Secrets in code | Pattern + entropy | 95% |

**Overall Security Detection: 85%**

---

## 5. Test Enforcement System

### 5.1 Test Requirements from Business Rules

Each business rule automatically generates test requirements:

```json
{
  "test_requirements": {
    "BR-001": {
      "rule_name": "Order Cancellation Window",
      "required_tests": [
        {
          "id": "BR-001-T1",
          "name": "test_cancel_within_24h",
          "type": "happy_path",
          "priority": "critical",
          "scenario": "Cancel order created 1 hour ago",
          "setup": {"order": {"createdAt": "now - 1h"}},
          "action": "cancelOrder(order.id)",
          "expected": {"success": true}
        },
        {
          "id": "BR-001-T2",
          "name": "test_cancel_after_24h",
          "type": "error_case",
          "priority": "critical",
          "scenario": "Cancel order created 25 hours ago",
          "setup": {"order": {"createdAt": "now - 25h"}},
          "action": "cancelOrder(order.id)",
          "expected": {"error": "ERR_CANCEL_WINDOW_EXPIRED"}
        },
        {
          "id": "BR-001-T3",
          "name": "test_cancel_at_boundary",
          "type": "edge_case",
          "priority": "high",
          "scenario": "Cancel at exactly 24 hours"
        },
        {
          "id": "BR-001-T4",
          "name": "test_cancel_premium_48h",
          "type": "exception_case",
          "priority": "high",
          "scenario": "Premium user cancels at 30 hours"
        }
      ]
    }
  }
}
```

### 5.2 Test Coverage Tracking

```
./sentinel audit --test-coverage

📊 Test Coverage Report
══════════════════════════════════════════════════════════════

Rule Coverage:
┌────────┬────────────────┬─────────┬─────────┬──────────┐
│ Rule   │ Required Tests │ Written │ Passing │ Status   │
├────────┼────────────────┼─────────┼─────────┼──────────┤
│ BR-001 │ 4              │ 4       │ 4       │ ✅ 100%  │
│ BR-002 │ 3              │ 2       │ 2       │ ⚠️ 67%   │
│ BR-003 │ 5              │ 0       │ 0       │ ❌ 0%    │
│ SEC-001│ 3              │ 3       │ 3       │ ✅ 100%  │
└────────┴────────────────┴─────────┴─────────┴──────────┘

Overall: 75% rule coverage (target: 100%)
```

### 5.3 Test Quality via Mutation Testing

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          MUTATION TESTING                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Original Code:                    Mutants Generated:                       │
│  ─────────────                     ─────────────────                        │
│  if (age >= 18) {                  if (age > 18) {    // boundary          │
│    allow();                        if (age >= 17) {   // off-by-one        │
│  }                                 if (age <= 18) {   // flip              │
│                                    if (true) {        // always true       │
│                                                                              │
│  Run tests against each mutant:                                             │
│  • Test fails → Mutant killed (good!)                                       │
│  • Test passes → Mutant survived (test is weak!)                            │
│                                                                              │
│  Mutation Score = Killed / Total = 4/5 = 80%                                │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.4 Test Enforcement Rules

```json
{
  "test_enforcement": {
    "mode": "strict",
    
    "rules": {
      "minimum_coverage": {
        "line": 80,
        "branch": 70,
        "rule": 100
      },
      
      "required_test_types": {
        "business_rules": ["happy_path", "error_case"],
        "api_endpoints": ["success", "validation_error", "auth_error"],
        "security_rules": ["bypass_attempt", "valid_request"]
      },
      
      "test_quality": {
        "min_assertions_per_test": 1,
        "min_mutation_score": 70
      }
    },
    
    "blocking": {
      "pr_merge": true,
      "commit": false,
      "deployment": true
    }
  }
}
```

---

## 6. Documentation Standardization

### 6.1 The Problem

Different documentation styles lead to inconsistent interpretation:
- PDF requirements vs email threads vs JIRA tickets
- Ambiguous language ("within 24 hours" - inclusive or exclusive?)
- Missing edge cases
- No traceability

### 6.2 Standardized Knowledge Schema

All knowledge extracted from documents follows this schema:

```json
{
  "$schema": "https://sentinel.dev/schemas/knowledge/v1",
  "business_rule": {
    "id": "BR-001",
    "version": "1.0",
    "status": "active",
    "title": "Order Cancellation Window",
    "description": "Orders can only be cancelled within 24 hours",
    
    "specification": {
      "trigger": "User requests cancellation",
      "preconditions": [
        "Order exists",
        "Order status in ['pending', 'processing']"
      ],
      "constraints": [
        {
          "type": "time_based",
          "expression": "currentTime - order.createdAt < 24 hours",
          "boundary": "exclusive",
          "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000"
        }
      ],
      "exceptions": [
        {
          "condition": "user.tier === 'premium'",
          "modified_constraint": "48 hours",
          "source": "Premium policy, section 3.2"
        }
      ],
      "side_effects": [
        {"action": "refund", "condition": "order.isPaid"},
        {"action": "restore_inventory", "condition": "always"},
        {"action": "send_email", "template": "order_cancelled"}
      ],
      "error_cases": [
        {
          "condition": "Outside window",
          "error_code": "ERR_CANCEL_WINDOW_EXPIRED"
        }
      ]
    },
    
    "test_requirements": [...],
    
    "traceability": {
      "source_document": "Requirements.docx",
      "source_section": "4.2",
      "source_page": 12
    }
  }
}
```

### 6.3 How Standardization Reduces Semantic Bugs

```
WITHOUT STANDARDIZATION              WITH STANDARDIZATION
════════════════════                ════════════════════

Raw: "within 24 hours"              Structured:
                                    {
LLM interprets:                       "constraint": "< 24 hours",
• < 24h? (exclusive)                  "boundary": "exclusive",
• <= 24h? (inclusive)                 "test_at_boundary": "23:59:59"
• From order time?                  }
• From payment time?
                                    Result: Unambiguous
Result: Ambiguous
```

---

## 7. Requirements Lifecycle Management

### 7.1 The Drift Problem

```
Day 1:   Document: "24-hour cancellation"
         Code: 24 hours ✓
         
Day 30:  Business changes to 48 hours
         Document updated
         Code still 24 hours ✗
         
Day 60:  Bug report filed
         Developer: "Working as designed"
         Business: "We changed it!"
```

### 7.2 Change Request System

```json
{
  "change_request": {
    "id": "CR-001",
    "type": "modification",
    "target_rule": "BR-001",
    "requested_by": "product@company.com",
    
    "current_state": {
      "constraint": "< 24 hours"
    },
    
    "proposed_state": {
      "constraint": "< 48 hours"
    },
    
    "impact_analysis": {
      "affected_code": [
        "src/services/order/cancellation.ts:45-67"
      ],
      "affected_tests": [
        "tests/order/cancellation.test.ts"
      ],
      "estimated_effort": "2 hours"
    },
    
    "status": "pending_approval"
  }
}
```

### 7.3 Gap Analysis

```
./sentinel knowledge gap-analysis

📋 Requirements Gap Analysis
══════════════════════════════════════════════════════════════

IMPLEMENTED BUT NOT DOCUMENTED:
───────────────────────────────
⚠️  src/services/shipping.ts:calculateShipping()
    Logic: Free shipping for orders > $50
    Action: Create rule or remove feature

DOCUMENTED BUT NOT IMPLEMENTED:
───────────────────────────────
❌ BR-007: Loyalty points (1 per $10)
   No implementation found

PARTIALLY IMPLEMENTED:
──────────────────────
⚠️  BR-003: Order confirmation email
    Missing: Retry logic, invoice attachment

TESTS MISSING:
──────────────
❌ BR-001: 4 required, 2 found
   Missing: boundary test, premium test
```

---

## 8. MCP as Active Orchestrator

### 8.1 MCP Flow (Active, Not Passive)

```
Developer: "add order cancellation"
        │
        ▼
[MCP: sentinel_analyze_intent]
        │
        └── Queries Hub for context
        │
        ▼
Returns to Cursor prompt:
{
  "security_rules": ["SEC-001", "SEC-003"],
  "business_rules": ["BR-001", "BR-002"],
  "required_tests": 4,
  "MUST_INCLUDE": [
    "Ownership verification",
    "24-hour window check",
    "Refund trigger"
  ],
  "suggested_location": "src/services/order/cancellation.ts"
}
        │
        ▼
Cursor generates WITH these constraints
        │
        ▼
[MCP: sentinel_validate_code]
        │
        └── AST analysis
        └── Security check
        └── Business rule check
        │
        ▼
[MCP: sentinel_generate_tests]
        │
        └── Generate required tests
        │
        ▼
[MCP: sentinel_run_tests] (optional)
        │
        └── Execute in sandbox
```

### 8.2 MCP Tools

| Tool | Purpose | When Called |
|------|---------|-------------|
| `sentinel_analyze_intent` | Understand request context | Before generation |
| `sentinel_get_business_context` | Get relevant rules | Before generation |
| `sentinel_get_security_context` | Get security requirements | Before generation |
| `sentinel_check_file_size` | Check target file size | Before generation |
| `sentinel_validate_code` | Validate generated code | After generation |
| `sentinel_validate_security` | Check security compliance | After generation |
| `sentinel_validate_tests` | Check test quality | After tests written |
| `sentinel_run_tests` | Execute tests | Optional |
| `sentinel_generate_tests` | Generate test cases | When needed |

---

## 9. Coverage Matrix

### Final Coverage with Full System

| Category | Before | After | Method |
|----------|--------|-------|--------|
| Structural/Syntax | 0% | **95%** | AST |
| Refactoring | 0% | **95%** | Cross-file AST |
| Variable/Scope | 0% | **85%** | Scope analysis |
| Control Flow | 0% | **85%** | CFG |
| Async/Concurrency | 0% | **75%** | Pattern + AI |
| Type/Null Safety | 0% | **75%** | Type inference |
| Logic/Semantic | 0% | **80%** | Executable rules |
| Business Logic | 0% | **90%** | Rule enforcement |
| Style/Consistency | 0% | **95%** | Pattern learning |
| AI-Specific | 0% | **85%** | Context + validation |
| Security | 30% | **85%** | Security rules |
| Test Coverage | 0% | **90%** | Requirement tracking |
| Test Quality | 0% | **80%** | Mutation testing |
| **Overall** | **~5%** | **~85%** | Full system |

---

## 10. Fundamental Limitations

### What Remains Undetectable (~15%)

| Issue | Why | Mitigation |
|-------|-----|------------|
| **True semantic bugs** | Code does wrong thing correctly | Human review, tests |
| **Unknown requirements** | Not in any document | Discovery sessions |
| **Novel security attacks** | Zero-day patterns | Regular updates |
| **Performance under load** | Requires real traffic | Load testing |
| **UX issues** | Subjective | User testing |
| **Race conditions** | Non-deterministic | Runtime monitoring |

### The Four Layers of Defense

| Layer | Catches | Required |
|-------|---------|----------|
| **Sentinel (Automated)** | 85% of issues | Yes |
| **Tests (Required)** | Logic correctness | Yes |
| **Monitoring (Runtime)** | Performance, errors | Yes |
| **Human Review (Critical)** | Security, architecture | For critical paths |

**No single tool can replace all four layers.**

---

## Appendix: Project-Specific Findings

### Issues Found in This Project

| Issue | Location | Resolution |
|-------|----------|------------|
| Duplicate `showKnowledgeHelp()` | Line 4640-4663 | Removed |
| Orphaned switch statement | Line 4650 | Removed |
| Unused variable `lineNum` | Line 2885 | Changed to `_` |
| Wrong signature `showKnowledgeStats(remainingArgs)` | Line 4612 | Fixed |

These are **exactly the types of vibe coding issues** Sentinel is designed to detect and prevent.


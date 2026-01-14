# Feature Specification

> **For AI Agents**: This document specifies all features of Sentinel. Refer to this when implementing or extending functionality.

## Existing Features (v24)

### Core Engine

| Feature | Description | Status |
|---------|-------------|--------|
| Go binary compilation | Self-compiling binary | ✅ Done |
| Source code deletion | Security through obscurity | ✅ Done |
| Cursor rules generation | 6 stack templates | ✅ Done |
| Security scanning | 17 vulnerability patterns | ✅ Done |
| Parallel scanning | Goroutine-based | ✅ Done |
| False positive detection | Comment/doc awareness | ✅ Done |
| Entropy-based secrets | High-entropy string detection | ✅ Done |
| Baseline system | Accept known issues | ✅ Done |
| Configuration | 3-tier JSON config | ✅ Done |
| Report generation | JSON/HTML/MD/Text | ✅ Done |
| Audit history | Trends and comparisons | ✅ Done |
| Git hooks | Pre-commit/push/msg | ✅ Done |
| Rules backup/rollback | Version control for rules | ✅ Done |
| Windows support | PowerShell/batch wrappers | ✅ Done |

### Commands

| Command | Purpose | Status |
|---------|---------|--------|
| `init` | Bootstrap project | ✅ Done |
| `audit` | Security scan | ✅ Done |
| `docs` | Generate file structure | ✅ Done |
| `baseline` | Manage accepted findings | ✅ Done |
| `history` | View audit trends | ✅ Done |
| `install-hooks` | Set up git hooks | ✅ Done |
| `validate-rules` | Check rule syntax | ✅ Done |
| `update-rules` | Fetch external rules | ✅ Done |
| `status` | Project health | ✅ Done |
| `review` | Knowledge review | ✅ Done |
| `knowledge` | Knowledge management | ✅ Done |
| `doc-sync` | Documentation-code sync | ✅ Done |

---

## Enhanced Features (New)

### Phase A: Vibe Coding Detection ✅ COMPLETE (100%)

**Status**: ✅ COMPLETE - AST-first detection fully functional. Pattern fallback works. All features including Phase 7C optional enhancements implemented.

**Implementation Status**:
- ✅ `--vibe-check` flag exists and works (AST + patterns)
- ✅ `--vibe-only` flag exists and filters correctly
- ✅ `--deep` flag exists and Hub integration fully functional
- ✅ AST-based detection complete (Phase 6 complete)
- ✅ `--offline` flag for pattern-only mode
- ✅ Progress indicators for Hub analysis
- ✅ Cancellation support (Ctrl+C handling)
- ✅ Metrics tracking (AST vs pattern usage)
- ✅ Empty catch/except blocks detection (AST-based)
- ✅ Enhanced code after return/throw/raise detection
- ✅ Missing await detection for async functions
- ✅ Brace/bracket mismatch detection from parser errors
- ✅ Semantic deduplication (AST vs patterns)

**Purpose**: Detect and prevent common issues from AI-assisted code generation.

> **Reference**: See [VIBE_CODING_ANALYSIS.md](./VIBE_CODING_ANALYSIS.md) for complete analysis.

**Commands**:
```bash
sentinel audit --vibe-check       # ✅ COMPLETE - AST-first detection with pattern fallback
sentinel audit --vibe-only        # ✅ COMPLETE - Filters to vibe issues only
sentinel audit --deep             # ✅ COMPLETE - Server-side AST analysis via Hub
sentinel audit --offline          # ✅ COMPLETE - Force pattern-only mode (skip Hub)
```

**Detected Issues**:
| Issue | Detection Method | Severity | Status |
|-------|------------------|----------|--------|
| Duplicate function definitions | AST (Hub) | Error | ✅ Implemented |
| Orphaned code (outside scope) | AST (Hub) | Error | ✅ Implemented |
| Unused variables | AST (Hub) | Warning | ✅ Implemented |
| Signature mismatches | AST cross-file | Error | ✅ Implemented (Phase 6F) |
| Empty catch/except blocks | Pattern + AST | Warning | ✅ Implemented |
| Code after return | Control flow | Warning | ✅ Implemented |
| Missing await | Async tracking | Warning | ✅ Implemented |
| Brace/bracket mismatch | Parser | Error | ✅ Implemented |

**Architecture**:
```
Agent (Local)                     Hub (Server)
─────────────                     ─────────────
AST-first detection          ───►   AST analysis (PRIMARY)
(if Hub unavailable)                Tree-sitter parsing
Pattern fallback                    Cross-file analysis ✅
(offline only)                       AI code review
```

> **Note**: Cross-file analysis fully implemented in Phase 6F. Detects signature mismatches and import/export issues across files.

**Detection Flow**:
1. **Primary**: Send code to Hub for AST analysis (when `--deep` flag used or Hub available)
2. **Fallback**: Use pattern-based detection only if:
   - Hub is unavailable/unreachable
   - `--deep` flag not used AND telemetry disabled
   - Network timeout/error occurs
3. **Deduplication**: Remove pattern findings that overlap with AST findings (AST takes precedence)

### Phase B: File Size Management ✅ COMPLETE

**Status**: ✅ COMPLETE - All core functionality implemented and tested. Phase 9 complete.

**Implementation Status**:
- ✅ `FileSizeConfig` struct defined
- ✅ Default thresholds configured (300/500/1000 lines)
- ✅ Config merging logic implemented
- ✅ File size checking integrated into audit process
- ✅ `checkFileSize()` function implemented
- ✅ `--analyze-structure` flag implemented
- ✅ Hub architecture analysis endpoint implemented
- ✅ Section detection implemented (AST-first with pattern fallback)
- ✅ Split suggestions implemented
- ✅ Agent-Hub integration implemented
- ✅ MCP tool preparation (ready for Phase 14)
- ✅ MCP integration complete (Phase 14B)
- ⚠️ Telemetry integration (deferred - can be added in Phase 5 enhancement)
- ✅ Tests implemented

**Note**: Phase 9 provides suggestions and migration instructions only. File splitting execution is deferred to Phase 9B (future phase).

**Purpose**: Prevent large monolithic files that cause context overflow and vibe coding issues.

**Evidence**: This project's `synapsevibsentinel.sh` at 8,489 lines demonstrates the problem.

**Commands**:
```bash
sentinel audit --analyze-structure    # ✅ IMPLEMENTED - Analyze file sizes and suggest splits
# Note: sentinel fix --split removed from Phase 9 scope (deferred to Phase 9B)
```

**Configuration**:
```json
{
  "fileSize": {
    "thresholds": {
      "warning": 300,
      "critical": 500,
      "maximum": 1000
    },
    "byFileType": {
      "component": 200,
      "service": 400,
      "utility": 150,
      "test": 500
    },
    "exceptions": []
  }
}
```

**Features**:
| Feature | Description |
|---------|-------------|
| Size monitoring | Track file line counts |
| Split suggestions | Analyze logical sections, suggest splits |
| MCP guidance | Warn before generating into oversized file |
| Architecture analysis | Detect module boundaries |

**MCP Integration**:
```
Developer: "add shipping calculation"

[MCP: sentinel_check_file_size]
Returns: {
  "target_file": "orderService.ts",
  "current_lines": 847,
  "status": "oversized",
  "recommendation": "Create new file",
  "suggested_location": "src/services/order/shipping.ts"
}

Cursor generates in new file instead of adding to oversized file.
```

### Phase 9.5: Interactive Git Hooks ✅ COMPLETE

**Status**: ✅ COMPLETE - Interactive hooks with telemetry, Hub integration, and policy enforcement implemented.

**Implementation Status**:
- ✅ Interactive hook handler (`runInteractiveHook()`)
- ✅ Severity-based handling (block critical, warn high/medium, auto-proceed info)
- ✅ Hook context tracking (`HookContext` struct)
- ✅ Hook telemetry (`sendHookTelemetry()`)
- ✅ Hub API endpoints (`/api/v1/telemetry/hook`, `/api/v1/hooks/metrics`, `/api/v1/hooks/policies`)
- ✅ Database schema (hook_executions, hook_baselines, hook_policies tables)
- ✅ Policy enforcement (`getHookPolicy()`, `checkHookPolicy()`)
- ✅ Baseline review workflow
- ✅ CI/CD integration (`--non-interactive` flag)
- ✅ MCP integration (Phase 14B)
- ✅ Comprehensive analysis integration (Phase 14A)

**Purpose**: Interactive git hooks that warn users and provide options, with comprehensive integration into the Sentinel ecosystem for organizational governance and reporting.

**Commands**:
```bash
sentinel hook pre-commit              # ✅ IMPLEMENTED - Interactive pre-commit hook
sentinel hook pre-push                # ✅ IMPLEMENTED - Interactive pre-push hook
sentinel hook pre-commit --non-interactive  # ✅ IMPLEMENTED - CI/CD mode
sentinel install-hooks                # ✅ IMPLEMENTED - Install interactive hooks
```

**Features**:

1. **Configurable Audit Scope**:
   - Hooks run configurable audit checks based on Hub policies
   - **Security Analysis**: AST-based security checks (Phase 8) - configurable via `audit_config.security`
   - **Vibe Coding Detection**: Duplicate code, orphaned code, unused variables (Phase 7) - configurable via `audit_config.vibe`
   - **Business Rules**: Knowledge-based compliance checks - configurable via `audit_config.business_rules`
   - **File Size Checks**: Large file detection and split suggestions (Phase 9) - configurable via `audit_config.file_size`
   - Default configuration: All checks enabled
   - Organizations can customize audit scope via Hub policy configuration
   - Fallback to default config if Hub unavailable

2. **Interactive Menu**:
   - View details of findings
   - Proceed anyway (override)
   - Add to baseline
   - Add file size exception
   - Quit (abort commit)

3. **Severity-Based Handling**:
   - **Critical**: Block commit, cannot override
   - **High/Medium**: Warn, can override with justification (if policy allows)
   - **Info**: Auto-proceed with notification

4. **Hook Telemetry**:
   - Hook execution events sent to Hub
   - Tracks: hook type, result, override reason, findings summary, user actions, duration
   - Aggregated metrics available via Hub API

5. **Policy Enforcement**:
   - **Audit Scope Configuration**: Per-organization control of which checks run in hooks
   - Override limits (configurable per organization, enforced via Hub tracking)
   - Baseline review requirements
   - Exception approval workflow
   - Policies stored in Hub, cached locally (5 minutes)
   - Policy limits enforced with Hub tracking (daily override counts, weekly baseline counts)

6. **Baseline Review Workflow**:
   - Hook-added baselines marked "pending_review"
   - Auto-approved after configured days (default: 7)
   - Admins can approve/reject immediately
   - Baseline entries automatically sent to Hub for tracking

7. **Hub Integration**:
   - Hook metrics API (`GET /api/v1/hooks/metrics`)
   - Policy configuration API (`GET /api/v1/hooks/policies`, `POST /api/v1/hooks/policies`)
   - Hook limits API (`GET /api/v1/hooks/limits`) - tracks override/baseline counts
   - Baseline API (`POST /api/v1/hooks/baselines`) - stores hook-added baselines
   - Team-level metrics
   - Trend analysis

**Usage Examples**:

**Interactive Hook (Pre-commit)**:
```bash
$ git commit -m "Add feature"
🔍 Sentinel PRE-COMMIT Hook
══════════════════════════════════════════════════════════════

⚠️  ISSUES DETECTED
══════════════════════════════════════════════════════════════
  ⚠️  3 warning(s)
  ℹ️  2 info issue(s)

Options:
  [v] View details
  [p] Proceed anyway
  [b] Add to baseline
  [e] Add file size exception (if applicable)
  [q] Quit (abort commit)

Choose an option: v

📋 DETAILED FINDINGS
══════════════════════════════════════════════════════════════
  ⚠️  src/utils.js:45 - console.log detected
      Context: console.log('Debug info')
  ⚠️  src/api.js:123 - Hardcoded absolute path detected
  ℹ️  src/components/Button.jsx:12 - Missing EOF newline

Choose an option [p/b/e/q]: b
✅ Findings added to baseline (pending review). Proceeding with commit...
```

**CI/CD Mode (Non-Interactive)**:
```bash
$ sentinel hook pre-commit --non-interactive
🔍 Sentinel PRE-COMMIT Hook
══════════════════════════════════════════════════════════════
⚠️  Issues detected but proceeding (non-interactive mode)...
✅ Proceeding with commit...
```

**Hub Metrics Query**:
```bash
curl -H "Authorization: Bearer $API_KEY" \
  "https://hub.example.com/api/v1/hooks/metrics?org_id=$ORG_ID&start_date=2024-12-01"
```

**Response**:
```json
{
  "total_executions": 150,
  "blocked_count": 12,
  "allowed_count": 120,
  "overridden_count": 18,
  "override_rate": 12.0,
  "avg_duration_ms": 2345.6
}
```

**Policy Configuration** (Hub Dashboard):
```json
{
  "override_policy": {
    "critical_requires_approval": true,
    "max_overrides_per_day": 5,
    "override_requires_justification": true
  },
  "baseline_policy": {
    "requires_review": true,
    "auto_approve_after_days": 7,
    "max_baselines_per_week": 10
  },
  "exception_policy": {
    "requires_approval": true,
    "temporary_exception_days": 30
  }
}
```

**Dependencies**:
- Phase 5 (Hub MVP) - Required for telemetry and policy storage
- Phase 6 (AST Analysis) - Required for deep analysis in hooks
- Phase 8 (Security Rules) - Required for security checks in hooks
- Phase 9 (File Size Management) - Required for file size checks in hooks

**Reference**: See [INTERACTIVE_HOOKS_ANALYSIS.md](./INTERACTIVE_HOOKS_ANALYSIS.md) and [TELEMETRY_GRANULARITY.md](./TELEMETRY_GRANULARITY.md) for detailed analysis.

### Phase 9.5.1: Reliability Improvements ✅ COMPLETE

**Status**: ✅ COMPLETE - Database timeouts, retry logic, cache improvements, and error recovery system implemented.

**Implementation Status**:
- ✅ Database query timeout helpers (`queryWithTimeout`, `queryRowWithTimeout`, `execWithTimeout`)
- ✅ HTTP retry logic with exponential backoff (`httpRequestWithRetry`)
- ✅ Cache improvements (RWMutex, per-entry expiration, time-based cleanup)
- ✅ Error recovery system (`CheckResults` in `AuditReport`)
- ✅ Panic recovery in wrapper functions
- ✅ Database connection pool health monitoring
- ✅ AST cache resource leak prevention

**Purpose**: Improve system reliability, prevent resource leaks, and ensure graceful error handling.

**Features**:

1. **Database Query Timeouts**:
   - All database queries use context-aware timeouts (10 seconds default)
   - Helper functions: `queryWithTimeout()`, `queryRowWithTimeout()`, `execWithTimeout()`
   - Prevents database connection pool exhaustion
   - Automatic context cancellation on timeout
   - Used in: `hub/api/hook_handler.go`, `hub/api/policy.go`

2. **HTTP Retry Logic**:
   - Exponential backoff retry for transient failures
   - Retries on network errors and 5xx server errors
   - No retry on 4xx client errors
   - Configurable max retries (default: 3)
   - Helper function: `httpRequestWithRetry()`
   - Used for: Hub communication, baseline submission, telemetry

3. **Cache Improvements**:
   - **Policy Cache**: RWMutex for thread-safe access, timestamp-based invalidation
   - **Limits Cache**: Per-entry expiration, thread-safe map with RWMutex
   - **AST Cache**: Time-based periodic cleanup to prevent resource leaks
   - Cache corruption detection and automatic cleanup
   - Cache invalidation based on Hub `updated_at` timestamps

4. **Error Recovery System**:
   - `CheckResult` struct tracks check status (enabled, success, error, findings count)
   - `CheckResults` map in `AuditReport` tracks all check types
   - Error wrapper functions: `performSecurityAnalysisWithError()`, `detectVibeIssuesWithError()`, `checkBusinessRulesComplianceWithError()`
   - Panic recovery with detailed logging and error state tracking
   - Finding count tracking (before/after) for accurate reporting

5. **Database Connection Pool Health**:
   - Background goroutine monitors connection pool health
   - Logs connection pool metrics (open, idle, in-use)
   - Alerts on potential pool exhaustion
   - Connection lifetime management (`SetConnMaxLifetime`)

**Usage Examples**:

**Database Timeout Helper**:
```go
// In hub/api/hook_handler.go
result, err := queryRowWithTimeout(ctx, query, args...)
if err != nil {
    log.Printf("Query timeout or error: %v", err)
    return
}
```

**HTTP Retry Logic**:
```go
// In synapsevibsentinel.sh
resp, err := httpRequestWithRetry(client, req, 3)
if err != nil {
    logWarn("Hub communication failed after retries: %v", err)
    return
}
```

**Error Recovery**:
```go
// CheckResults populated automatically in performAuditForHook()
report.CheckResults["security"] = CheckResult{
    Enabled:  true,
    Success:  true,
    Findings: 5,
}
```

**Cache Invalidation**:
```go
// Policy cache invalidated when Hub updated_at > cached timestamp
if policyUpdatedTime.After(cachedHookPolicy.UpdatedAt) {
    // Invalidate cache, fetch fresh policy
}
```

**Dependencies**:
- Phase 9.5 (Interactive Git Hooks) - Required for hook error handling
- Phase 5 (Hub MVP) - Required for Hub communication retry logic
- Phase 6 (AST Analysis) - Required for AST cache cleanup

**Reference**: See reliability improvements in `hub/api/hook_handler.go` (timeout helpers), `hub/api/policy.go`, `hub/api/validation.go` (validation helpers), `hub/api/main.go` (health monitoring), `synapsevibsentinel.sh` (retry logic, cache improvements), and `docs/external/ERROR_HANDLING_STANDARDS.md` (error handling standards).

---

### Phase C: Built-in Security Rules ✅ IMPLEMENTED

**Status**: ✅ COMPLETE - Full security rule checking with AST analysis (Phase 8).

**Implementation Status**:
- ✅ SEC-001 to SEC-008 rule definitions exist in Hub
- ✅ Security analysis endpoint exists (`/api/v1/analyze/security`)
- ✅ Full AST-based security checking implemented
- ✅ Security scoring algorithm (0-100 with grade A-F)
- ✅ Framework detection (Express, FastAPI, Gin, Flask, Django, Rails)
- ✅ Pattern + AST hybrid detection
- ✅ Agent `--security` flag integrated

**Purpose**: Enforce security patterns beyond simple regex matching.

**Security Rule Types**:
| Type | Description | Examples |
|------|-------------|----------|
| authorization | Resource ownership checks | IDOR prevention |
| authentication | Auth middleware presence | JWT verification |
| injection | Parameterized queries | SQL/NoSQL injection |
| validation | Input sanitization | Request body validation |
| cryptography | Secure algorithms | Password hashing |
| transport | Secure headers | CORS, CSP |

**Security Rules (Built-in)** - ✅ IMPLEMENTED (Phase 8):
| ID | Name | Severity | Detection | Status |
|----|------|----------|-----------|--------|
| SEC-001 | Resource Ownership | Critical | AST ownership check | ✅ Implemented |
| SEC-002 | SQL Injection | Critical | Pattern + AST | ✅ Implemented |
| SEC-003 | Auth Middleware | Critical | Route analysis | ✅ Implemented |
| SEC-004 | Rate Limiting | High | Endpoint analysis | ✅ Implemented |
| SEC-005 | Password Hashing | Critical | Pattern + Data flow | ✅ Implemented |
| SEC-006 | Input Validation | High | Handler analysis | ✅ Implemented |
| SEC-007 | Secure Headers | Medium | Middleware check | ✅ Implemented |
| SEC-008 | CORS Config | High | Config analysis | ✅ Implemented |

**Rule Definition Format**:
```json
{
  "id": "SEC-001",
  "name": "Resource Ownership Verification",
  "type": "authorization",
  "severity": "critical",
  "detection": {
    "endpoints": ["/api/:resource/:id"],
    "required_checks": [
      "req.user.id === resource.userId",
      "req.user.role === 'admin'"
    ]
  },
  "ast_check": {
    "function_contains": ["findById"],
    "must_have_before_response": "ownership_check"
  }
}
```

**Commands**:
```bash
sentinel audit --security          # ✅ IMPLEMENTED - Security-focused audit with scoring
sentinel audit --security-rules   # ✅ IMPLEMENTED - List all security rules
sentinel audit --business-rules   # ✅ IMPLEMENTED - Validate code against approved business rules
```

**Usage Examples**:

**1. Security Analysis with Scoring**:
```bash
$ sentinel audit --security
🔒 Performing security analysis...
📊 Analyzing 15 files for security issues...
   Processing batch 1/2
   Processing batch 2/2
🔒 Security analysis found 3 issues

Security Score: 75/100 (Grade: C)
Summary:
  Total Rules: 8
  Passed: 5
  Failed: 3
  Critical: 1
  High: 2
  Medium: 0
  Low: 0

Findings:
  SEC-001: Resource Ownership (Critical)
    File: src/routes/users.js:45
    Issue: Missing required security check 'ownership_check'
    Remediation: Verify user.id === resource.userId before access

  SEC-005: Password Hashing (Critical)
    File: src/auth/register.js:23
    Issue: Password variable flows to insecure MD5 hashing
    Remediation: Use bcrypt.hash() or argon2.hash() instead
```

**2. Security Rules Listing**:
```bash
$ sentinel audit --security-rules
🔒 Available Security Rules:

  🔴 SEC-001: Resource Ownership
     Type: authorization | Severity: critical
     Ensure resource access is verified against user ownership

  🔴 SEC-002: SQL Injection Prevention
     Type: injection | Severity: critical
     Ensure SQL queries use parameterized statements

  🔴 SEC-003: Authentication Middleware
     Type: authentication | Severity: critical
     Ensure protected routes have authentication middleware

  🟡 SEC-004: Rate Limiting
     Type: transport | Severity: high
     Ensure API endpoints have rate limiting

  🔴 SEC-005: Password Hashing
     Type: cryptography | Severity: critical
     Ensure passwords are hashed using secure algorithms (Pattern + Data flow)

  🟡 SEC-006: Input Validation
     Type: validation | Severity: high
     Ensure user input is validated before processing

  🟠 SEC-007: Secure Headers
     Type: transport | Severity: medium
     Ensure secure HTTP headers are set

  🟡 SEC-008: CORS Configuration
     Type: transport | Severity: high
     Ensure CORS is properly configured (not wildcard for production)
```

**3. Data Flow Analysis Example**:
The security analyzer tracks password variables through code paths:
```javascript
// ❌ Insecure: Password flows to MD5
function registerUser(req, res) {
  const password = req.body.password;  // User input
  const hashed = md5(password);        // SEC-005: Insecure hashing
  // ...
}

// ✅ Secure: Password flows to bcrypt
function registerUser(req, res) {
  const password = req.body.password;  // User input
  const hashed = await bcrypt.hash(password, 10);  // SEC-005: Secure
  // ...
}
```

**4. Framework Detection with Confidence**:
The analyzer detects frameworks with confidence levels:
- **High confidence**: Route definitions found in AST
- **Medium confidence**: Middleware usage patterns detected
- **Low confidence**: Only imports detected

**5. Batch Processing Performance**:
Security analysis processes files in batches with concurrent requests:
- Batch size: 10 files
- Max concurrent requests: 5
- Progress indicators show batch processing status
- Significantly faster than sequential processing

**6. Detection Rate Metrics (Validation Mode)**:
When ground truth is provided via `expectedFindings` in the request, security analysis includes detection rate metrics:
```json
{
  "score": 75,
  "grade": "C",
  "findings": [...],
  "metrics": {
    "truePositives": 8,
    "falsePositives": 2,
    "falseNegatives": 1,
    "trueNegatives": 5,
    "detectionRate": 87.5,
    "precision": 80.0,
    "recall": 88.9
  }
}
```

**Usage with Ground Truth**:
```bash
# Send request with expected findings for validation
curl -X POST http://localhost:8080/api/v1/analyze/security \
  -H "Content-Type: application/json" \
  -d '{
    "code": "...",
    "language": "javascript",
    "filename": "test.js",
    "rules": ["SEC-005"],
    "expectedFindings": {
      "SEC-005": true,
      "SEC-002": false
    }
  }'
```

**Metrics Explanation**:
- **True Positives**: Correctly detected vulnerabilities
- **False Positives**: Incorrectly flagged as vulnerabilities
- **False Negatives**: Missed vulnerabilities
- **True Negatives**: Correctly identified as safe
- **Detection Rate**: Overall accuracy (TP + TN) / Total * 100
- **Precision**: Accuracy of positive predictions TP / (TP + FP) * 100
- **Recall**: Coverage of actual vulnerabilities TP / (TP + FN) * 100

**Ground Truth Test Suite**:
Located at `tests/fixtures/security/ground_truth/` with labeled vulnerabilities for validation.
```

**Business Rules Compliance**:
- Validates code against approved business rules from knowledge store
- Uses pattern matching to detect rule violations in codebase
- Extracts validation patterns from rule content (time-based, amount limits, approval requirements)
- Checks for rule-specific violations (e.g., cancellation without time check, hardcoded limits)
- Reports violations as audit findings with appropriate severity
- Integrated with `--business-rules` flag
- Function: `checkBusinessRulesCompliance()` in Agent

**Usage Example**:
```bash
$ sentinel audit --business-rules
📋 Checking 3 business rules...
📋 Checking business rule: BR-001 - Order Cancellation Policy
📋 Checking business rule: BR-002 - Maximum Order Amount
📋 Checking business rule: BR-003 - Approval Required
✅ Business rules validation complete

Findings:
  ⚠️ [WARNING] src/orders/cancel.js:45
     Business rule violation: Order Cancellation Policy (Rule: BR-001)
     Code: function cancelOrder(orderId) { ... }
     Issue: Cancellation logic found without 24-hour time check
```

**Validation Patterns**:
- **Time-based rules**: Detects operations without required time constraints
- **Amount/limit rules**: Identifies hardcoded values that may violate limits
- **Approval rules**: Checks for operations missing approval/authorization checks
- **Validation rules**: Identifies input handling without proper validation

### Phase 10: Test Enforcement System ✅ COMPLETE

**Status**: ✅ COMPLETE - All Phase 10 features implemented and tested.

**Purpose**: Ensure business rules have corresponding tests.

**Commands**:
```bash
sentinel test --requirements    # ✅ Generate test requirements from business rules
sentinel test --coverage        # ✅ Analyze test coverage (sends test file content)
sentinel test --validate        # ✅ Validate test correctness
sentinel test --mutation        # ✅ Run mutation testing (requires --source and --test)
sentinel test --run             # ✅ Execute tests in sandbox (requires Docker)
```

**Test Requirements Generation**:
- Each business rule generates required test cases
- Test types: happy_path, error_case, edge_case, exception_case
- Minimum coverage enforced before merge

**Test Coverage Tracking**:
```
┌────────┬────────────────┬─────────┬─────────┬──────────┐
│ Rule   │ Required Tests │ Written │ Passing │ Status   │
├────────┼────────────────┼─────────┼─────────┼──────────┤
│ BR-001 │ 4              │ 4       │ 4       │ ✅ 100%  │
│ BR-002 │ 3              │ 2       │ 2       │ ⚠️ 67%   │
│ BR-003 │ 5              │ 0       │ 0       │ ❌ 0%    │
└────────┴────────────────┴─────────┴─────────┴──────────┘
```

**Test Quality (Mutation Testing)**:
- Generate code mutants (change operators, boundaries)
- Run tests against mutants
- Mutation score = killed / total
- Flag weak tests that don't catch mutations

**Enforcement Configuration**:
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
      "test_quality": {
        "min_mutation_score": 70
      }
    },
    "blocking": {
      "pr_merge": true,
      "deployment": true
    }
  }
}
```

### Phase E: Requirements Lifecycle Management ✅ IMPLEMENTED

**Status**: ✅ Fully implemented in Phase 12.

**Purpose**: Track requirements changes and ensure code stays in sync.

**Commands**:
```bash
sentinel knowledge gap-analysis     # ✅ IMPLEMENTED - Find gaps between rules and code
sentinel knowledge changes          # ✅ IMPLEMENTED - Show pending change requests
sentinel knowledge impact CR-001    # ✅ IMPLEMENTED - Impact analysis
sentinel knowledge approve CR-001   # ✅ IMPLEMENTED - Approve change request
sentinel knowledge reject CR-001    # ✅ IMPLEMENTED - Reject change request
sentinel knowledge track CR-001     # ✅ IMPLEMENTED - Track implementation status
```

**Implementation Details**:
- Gap analysis uses AST-based code analysis for accurate detection
- Change detection automatically triggers on document re-ingestion
- Change requests follow approval workflow with impact analysis
- Implementation tracking monitors status from start to completion
- **Cache**: Gap analysis results are cached (1 hour TTL) for improved performance
- **Persistence**: Gap reports are automatically stored in database with `report_id` for tracking

**Gap Analysis**:
| Gap Type | Description | Action |
|----------|-------------|--------|
| Implemented but not documented | Code exists, no rule | Document or remove |
| Documented but not implemented | Rule exists, no code | Implement |
| Partially implemented | Rule exists, incomplete code | Complete |
| Tests missing | Rule exists, no tests | Add tests |

**Change Detection**:
When updated documents are ingested:
1. Compare with existing knowledge
2. Identify: New / Modified / Removed rules
3. Generate change request
4. Analyze impact on code and tests
5. Track implementation status

**API Endpoints**:
- `POST /api/v1/knowledge/gap-analysis` - Run gap analysis (returns `report_id` for stored reports)
- `GET /api/v1/change-requests` - List change requests
- `GET /api/v1/change-requests/{id}` - Get change request details
- `PUT /api/v1/change-requests/{id}/approve` - Approve change request
- `PUT /api/v1/change-requests/{id}/reject` - Reject change request
- `GET /api/v1/change-requests/{id}/impact` - Get impact analysis
- `PUT /api/v1/change-requests/{id}/implementation/start` - Start implementation
- `PUT /api/v1/change-requests/{id}/implementation/complete` - Complete implementation
- `GET /api/v1/change-requests/{id}/implementation-status` - Get implementation status

**Gap Analysis Response**:
```json
{
  "success": true,
  "report_id": "550e8400-e29b-41d4-a716-446655440000",
  "report": {
    "project_id": "uuid",
    "gaps": [ ... ],
    "summary": { ... }
  }
}
```

**Change Request Schema**:
```json
{
  "id": "CR-001",
  "type": "modification",
  "target_rule": "BR-001",
  "current_state": { "constraint": "< 24 hours" },
  "proposed_state": { "constraint": "< 48 hours" },
  "impact_analysis": {
    "affected_code": ["src/services/order/cancellation.ts:45-67"],
    "affected_tests": ["tests/order/cancellation.test.ts"],
    "estimated_effort": "2 hours"
  },
  "status": "pending_approval"
}
```

---

## Phase 1: Document Ingestion (Server-Side)

**Purpose**: Convert raw project documents into actionable knowledge.

> **Architecture Decision**: Document processing runs on Sentinel Hub (server),
> not on developer machines. This eliminates dependency management issues
> (poppler, tesseract) and enables LLM-powered extraction.
> See [Architecture Decision](./ARCHITECTURE_DOCUMENT_PROCESSING.md).

**Commands**:
```bash
sentinel ingest /path/to/docs/    # Upload to Hub (recommended)
sentinel ingest --status          # Check processing status
sentinel ingest --sync            # Sync results to local
sentinel ingest --offline         # Local processing (limited)
sentinel ingest --skip-images     # Skip image processing
sentinel review                   # Review extracted knowledge
sentinel review --list            # List pending items
sentinel review --approve file    # Approve specific file
```

**Processing Modes**:
| Mode | Dependencies | LLM Extraction | Formats |
|------|--------------|----------------|---------|
| Server (Hub) | None on client | ✅ Yes | All |
| Offline (Local) | Optional | ❌ No | Basic only |

**Supported Formats**:
| Format | Server (Hub) | Local (Offline) |
|--------|--------------|-----------------|
| Text (.txt, .md) | ✅ Go native | ✅ Go native |
| Word (.docx) | ✅ XML parser | ✅ XML parser |
| Excel (.xlsx) | ✅ XML parser | ✅ XML parser |
| Email (.eml) | ✅ net/mail | ✅ net/mail |
| PDF | ✅ poppler (server) | ⚠️ Requires poppler |
| Images | ✅ tesseract (server) | ⚠️ Requires tesseract |
| LLM Extraction | ✅ Azure/Ollama | ❌ Not available |

**Extraction Output**:
| Document | Content |
|----------|---------|
| domain-glossary.draft.md | Business entities and definitions |
| business-rules.draft.md | BR-XXX rules with conditions |
| user-journeys.draft.md | User workflows and steps |
| objectives.draft.md | Project goals and KPIs |
| entities/*.draft.md | Detailed entity specs |

**Human Review Workflow**:
1. All extractions create `.draft.md` files
2. Each item has confidence score and source reference
3. Human reviews: Accept / Edit / Reject
4. Only approved docs used by Cursor
5. Skipped items flagged for later review

### Phase 2: Pattern Learning

**Purpose**: Automatically detect project conventions from existing code.

**Commands**:
```bash
sentinel learn                    # Full pattern extraction
sentinel learn --naming           # Naming conventions only
sentinel learn --imports          # Import patterns only
sentinel learn --structure        # Folder structure only
sentinel learn --output json      # Machine-readable output
```

**Detected Patterns**:
| Pattern | Detection Method | Example |
|---------|------------------|---------|
| Function naming | Regex + frequency | camelCase (92%) |
| Variable naming | Regex + frequency | camelCase (88%) |
| File naming | Directory scan | kebab-case |
| Import style | Parse statements | Absolute with @/ |
| Folder structure | Tree analysis | src/components/{name}/ |
| Code style | Sample analysis | 2-space indent, single quotes |

**Output**:
- `.sentinel/patterns.json` - Stored patterns with confidence
- `.cursor/rules/project-patterns.md` - Generated Cursor rules

### Phase 3: Safe Auto-Fix

**Purpose**: Automatically fix issues that are safe to change.

**Commands**:
```bash
sentinel fix                      # Interactive mode
sentinel fix --safe               # Only safe fixes
sentinel fix --dry-run            # Preview changes
sentinel fix --yes                # Auto-approve
sentinel fix rollback             # Undo last fix
sentinel fix --pattern "name"     # Specific pattern
```

**Safe Fixes (Auto-Apply)**:
| Fix | Action | Languages |
|-----|--------|-----------|
| console.log removal | Delete line | JS/TS |
| print() debug | Delete line | Python |
| Trailing whitespace | Trim | All |
| Missing EOF newline | Add | All |
| Import sorting | Reorder | JS/TS/Python |
| Shell variable quoting | Add quotes | Bash/Shell |
| Unused imports | Remove | JS/TS/Python |

**Prompted Fixes (Require Confirmation)**:
| Fix | Prompt | Risk Level |
|-----|--------|------------|
| Rename function | "Rename 'X' to 'Y'? [Y/n]" | Medium |
| Move file | "Move to new location? [Y/n]" | Medium |
| Security issue | "Apply fix? [Y/n]" | High |
| Refactor pattern | "Update all instances? [Y/n]" | High |

**Backup System**:
- Automatic backup before any fix
- Timestamped backup folders
- Single-command rollback
- History of all fixes

### Phase 4: MCP Integration (Enhanced) ✅ COMPLETE (Phase 14B)

**Status**: ✅ COMPLETE - MCP server fully functional with comprehensive analysis tool.

**Implementation Status**:
- ✅ `mcp-server` command registered
- ✅ `runMCPServer()` function implemented
- ✅ MCP protocol handler implemented (JSON-RPC 2.0 over stdio)
- ✅ `sentinel_analyze_feature_comprehensive` tool implemented
- ✅ Hub API integration complete
- ✅ Error handling and fallback implemented

**Purpose**: Real-time integration with Cursor IDE as **active orchestrator**.

**MCP as Active Orchestrator** (Not Just Validator):
```
BEFORE GENERATION:
├── sentinel_analyze_intent      → Understand request
├── sentinel_get_business_context → Get relevant rules
├── sentinel_get_security_context → Get security requirements
├── sentinel_get_test_requirements → Get required tests
└── sentinel_check_file_size     → Check target file

AFTER GENERATION:
├── sentinel_validate_code       → Structural + AST
├── sentinel_validate_security   → Security rule compliance
├── sentinel_validate_business   → Business rule compliance
├── sentinel_validate_tests      → Test quality + coverage
└── sentinel_run_tests           → Execute tests (optional)
```

**Tools Exposed**:
| Tool | Purpose | When Called |
|------|---------|-------------|
| `sentinel_analyze_intent` | Understand request context | Before generation |
| `sentinel_get_context` | Recent activity, errors, git status | Before generating code |
| `sentinel_get_patterns` | Project conventions for path | Before generating code |
| `sentinel_check_intent` | Clarify unclear requests | When prompt is vague |
| `sentinel_get_business_context` | Business rules, entities | For business logic |
| `sentinel_get_security_context` | Security requirements | For secure code |
| `sentinel_get_test_requirements` | Required tests for feature | Before implementation |
| `sentinel_check_file_size` | Target file size check | Before generation |
| `sentinel_validate_code` | Validate generated code | After generating code |
| `sentinel_validate_security` | Security compliance | After generating code |
| `sentinel_validate_tests` | Test quality check | After writing tests |
| `sentinel_apply_fix` | Fix issues in code | When issues found |
| `sentinel_generate_tests` | Generate test cases | When requested |
| `sentinel_run_tests` | Execute tests in sandbox | Optional verification |

**Workflow**:
```
Developer: "add order cancellation"
    │
    ▼
[sentinel_analyze_intent]
    │
    └── Returns context, rules, security, tests needed
    │
    ▼
[sentinel_check_file_size]
    │
    └── Warns if target file is oversized
    │
    ▼
Cursor generates code WITH constraints in prompt:
    - Business rules to implement
    - Security requirements
    - Test requirements
    - Target file recommendation
    │
    ▼
[sentinel_validate_code] + [sentinel_validate_security]
    │
    ├── Valid → Present to user
    │
    └── Issues → Fix or regenerate
```

### Phase 5: Intent Clarification

**Purpose**: Handle unclear prompts with simple questions.

**Design Principles**:
- Use simple words ("change" not "refactor")
- Offer numbered options (1, 2, 3)
- Show context (recent files, errors)
- Confirm understanding
- Support non-English speakers

**Simple Language Templates**:
| Scenario | Template |
|----------|----------|
| Unclear location | "Where should this go?\n1. {option1}\n2. {option2}\n3. Somewhere else" |
| Unclear entity | "Which {entity} do you mean?\n1. {option1}\n2. {option2}" |
| Confirm action | "I will {action}. Is this correct? [Y/n]" |
| Need more info | "I need more information. What should {thing} do?" |

**Context Gathering**:
| Context | Use |
|---------|-----|
| Recent files | Infer working area |
| Recent errors | Infer what to fix |
| Git status | Infer what changed |
| Terminal output | Infer current task |

### Phase 6: Business Knowledge (Enhanced)

**Purpose**: Make Cursor understand business logic, not just code.

> **Reference**: See [KNOWLEDGE_SCHEMA.md](./KNOWLEDGE_SCHEMA.md) for complete schema.

**Knowledge Structure**:
```
docs/knowledge/business/
├── domain-glossary.md      # Entity definitions
├── business-rules.md       # BR-XXX rules  
├── user-journeys.md        # User workflows
├── objectives.md           # Project goals
├── api-contracts.md        # API specifications
├── security-rules.md       # Security requirements
└── entities/
    ├── user.md             # User entity details
    ├── order.md            # Order entity details
    └── payment.md          # Payment entity details
```

**Executable Business Rules**:
```json
{
  "id": "BR-001",
  "title": "Order Cancellation Window",
  "specification": {
    "constraints": [{
      "type": "time_based",
      "expression": "< 24 hours",
      "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000",
      "boundary": "exclusive"
    }],
    "exceptions": [{
      "condition": "user.tier === 'premium'",
      "modified_constraint": "< 48 hours"
    }],
    "side_effects": [
      { "action": "refund", "condition": "order.isPaid" },
      { "action": "restore_inventory", "condition": "always" }
    ]
  },
  "test_requirements": [
    { "type": "happy_path", "name": "test_cancel_within_24h" },
    { "type": "error_case", "name": "test_cancel_after_24h" },
    { "type": "edge_case", "name": "test_cancel_at_boundary" }
  ]
}
```

**MCP Integration**:
```
sentinel_get_business_context("order", "cancellation")
→ {
    "rules": ["BR-001", "BR-002", "BR-003"],
    "constraints": ["24-hour window", "No cancel if shipped"],
    "side_effects": ["refund", "inventory", "email"],
    "security_rules": ["SEC-001", "SEC-003"],
    "test_requirements": 4
  }
```

### Phase 7: Telemetry & Central Hub (Enhanced)

**Purpose**: Organizational visibility with server-side analysis capabilities.

**Hub Capabilities**:
| Service | Function |
|---------|----------|
| API Server | REST endpoints for agents |
| AST Analysis | Tree-sitter parsing (100+ languages) |
| Security Scanner | Rule-based security analysis |
| Test Engine | Test generation, execution, mutation |
| Document Service | PDF, DOCX, image processing |
| LLM Integration | Azure AI / Ollama for extraction |
| Project Intelligence | Cross-file analysis, symbol index |

**Agent Sends**:
```json
{
  "event": "audit_complete",
  "agentId": "uuid",
  "orgId": "your-org",
  "timestamp": "2024-01-15T10:00:00Z",
  "metrics": {
    "findings": {"critical": 0, "warning": 3, "info": 5},
    "compliance": {"naming": 0.92, "imports": 0.88},
    "fixes": {"safe": 5, "prompted": 2},
    "documentation": {"coverage": 0.85, "drafts": 3},
    "test_coverage": {"rules": 0.75, "line": 0.82}
  }
}
```

**Agent Never Sends**:
- Source code
- File contents
- File names (unless opted in)
- Secrets
- Code snippets

**Dashboard Shows**:
| View | Content |
|------|---------|
| Organization Overview | Total compliance, trend |
| Team Breakdown | Per-team metrics |
| Trend Charts | Improvement over time |
| Common Issues | Top issues across org |
| Documentation Coverage | Business rule implementation |
| Test Coverage | Rule-to-test mapping |
| Security Score | Security rule compliance |
| Agent Health | Active agents, versions |

### Phase 8: Organization Features

**Purpose**: Team management, shared patterns.

**Features**:
| Feature | Description |
|---------|-------------|
| Team Management | Create, edit, delete teams |
| Pattern Distribution | Push org patterns to agents |
| Agent Registration | Track connected agents |
| Alerting | Notify on threshold breach |
| Role-Based Access | Admin, Lead, Developer roles |
| Security Policies | Org-wide security rules |
| Test Policies | Coverage requirements |

**Pattern Distribution**:
```
Org defines patterns → Push to Hub → Agents pull on startup
```

**Alerting Rules**:
| Trigger | Action |
|---------|--------|
| Critical finding | Immediate Slack/email |
| Compliance < 70% | Daily digest |
| Security score < 80% | Immediate notification |
| Agent offline > 24h | Admin notification |
| New draft pending > 7d | Reminder to reviewer |
| Test coverage drops | Block deployment |

### Phase 14A: Comprehensive Feature Analysis ✅ COMPLETE

**Status**: ✅ Complete (Phase 14A)

### Phase 14B: MCP Integration ✅ COMPLETE

**Status**: ✅ Complete (Phase 14B)

**Purpose**: End-to-end feature analysis across all layers (UI, API, Database, Logic, Integration, Tests) with business context validation to ensure comprehensive coverage beyond surface-level checks.

**Commands**:
- MCP tool: `sentinel_analyze_feature_comprehensive`
- Hub API: `POST /api/v1/analyze/comprehensive`
- Hub Dashboard: View results at `/validations/{id}`

### Phase 15: Intent & Simple Language ✅ COMPLETE

**Status**: ✅ Complete (Phase 15)

**Purpose**: Handle unclear prompts gracefully through intent analysis, simple language templates, context gathering, decision recording, and pattern refinement. Improves developer experience by reducing back-and-forth and supporting non-English speakers.

**Commands**:
- MCP tool: `sentinel_check_intent`
- Hub API: `POST /api/v1/analyze/intent`
- Hub API: `POST /api/v1/intent/decisions`
- Hub API: `GET /api/v1/intent/patterns`

**Features**:

1. **Feature Discovery**:
   - Auto-discovery across all layers (UI, API, Database, Logic, Integration, Tests)
   - Manual file specification option
   - Keyword-based component mapping

2. **7-Layer Analysis**:
   - **Business Context**: Rules, journeys, entities validation
   - **UI Layer**: Components, forms, validation, accessibility
   - **API Layer**: Endpoints, security, middleware, contracts
   - **Database Layer**: Schema, migrations, integrity, indexes
   - **Business Logic**: AST, cross-file, semantic analysis
   - **Integration Layer**: External APIs, contracts, side effects
   - **Test Layer**: Coverage, quality, edge cases

3. **End-to-End Flow Verification**:

**Phase 15 Features**:

1. **Intent Analysis**:
   - Detects unclear prompts (location_unclear, entity_unclear, action_confirm, ambiguous)
   - Rule-based quick check for clear prompts
   - LLM-based analysis for unclear prompts (with fallback)
   - Confidence scoring (0.0-1.0)

2. **Simple Language Templates**:
   - Pre-defined templates for common clarification scenarios
   - Extensible template system
   - Multi-choice and yes/no question formats

3. **Context Gathering**:
   - Recent files (git or filesystem)
   - Git status (branch, modified files)
   - Project structure (directories, file extensions)
   - Business rules from knowledge_items
   - Code patterns from recent files

4. **Decision Recording & Learning**:
   - Records user choices in `intent_decisions` table
   - Updates pattern frequency in `intent_patterns` table
   - Learns from past decisions to improve suggestions
   - Pattern refinement based on frequency

5. **MCP Integration**:
   - `sentinel_check_intent` tool available in Cursor IDE
   - Seamless integration with Hub API
   - Error handling and fallback messages

**Example Usage**:

**MCP Tool Call**:
```json
{
  "name": "sentinel_check_intent",
  "arguments": {
    "prompt": "add a new component",
    "codebasePath": "/path/to/project",
    "includeContext": true
  }
}
```

**Response**:
```json
{
  "requires_clarification": true,
  "intent_type": "location_unclear",
  "confidence": 0.8,
  "clarifying_question": "Where should this go?\n1. src/components/\n2. src/features/",
  "options": ["src/components/", "src/features/"]
}
```
   - Flow detection across layers
   - Breakpoint identification
   - Integration verification

4. **Business Context Integration**:
   - Validates against business rules (from knowledge base)
   - Checks user journey adherence
   - Verifies entity definitions
   - Ensures requirement coverage

5. **LLM Semantic Analysis**:
   - Business logic correctness
   - Requirement compliance
   - Edge case identification
   - Dual access model (Codex Pro + API)

6. **Hub Configuration Interface**:
   - API key management (user-provided or org-shared)
   - Provider selection (OpenAI, Anthropic, Azure)
   - Model selection (GPT-5.1-Codex-Max, GPT-5.1 Instant, etc.)
   - Cost optimization settings (caching, progressive depth)
   - Usage tracking (reporting only, not billing)

7. **Results and Reporting**:
   - Prioritized checklist (critical, high, medium, low)
   - Layer-specific findings
   - End-to-end flow status
   - Hub storage with URL access
   - Exportable reports

**Usage Examples**:

**MCP Tool (from Cursor)**:
```json
{
  "name": "sentinel_analyze_feature_comprehensive",
  "arguments": {
    "feature": "Order Cancellation",
    "mode": "auto",
    "depth": "deep",
    "includeBusinessContext": true
  }
}
```

**Response**:
```json
{
  "validationId": "val_abc123",
  "hubUrl": "https://hub.example.com/validations/val_abc123",
  "summary": {
    "totalFindings": 12,
    "critical": 2,
    "high": 5,
    "medium": 3,
    "low": 2
  },
  "checklist": [
    {
      "id": "chk_001",
      "severity": "critical",
      "title": "Missing authentication on DELETE /api/orders/:id",
      "location": "src/routes/orders.ts:45"
    }
  ]
}
```

**Important Notes**:
- **API Key Management**: Users/Organizations subscribe to LLM providers separately and provide API keys to Hub. Sentinel does NOT handle billing or payments.
- **Cost Tracking**: Sentinel tracks usage for reporting only, not billing. Users pay LLM providers directly.
- **Integration Analysis**: When analyzing features (e.g., "order cancellation"), Sentinel identifies payment gateway integrations as part of the FEATURE being analyzed. This is analysis of the FEATURE's integrations, not Sentinel's own functionality.

**Dependencies**:
- Phase 6: AST Analysis Engine ✅
- Phase 8: Security Rules System ✅
- Phase 4: Knowledge Base ✅

**Reference**: See [COMPREHENSIVE_ANALYSIS_SOLUTION.md](./COMPREHENSIVE_ANALYSIS_SOLUTION.md) for complete specification.

---

### Phase 14E: Task Dependency & Verification System ✅ COMPLETE

**Status**: ✅ Fully Implemented - Production Ready

**Purpose**: Track and verify Cursor-generated tasks with dependency management and completion verification to ensure tasks are completed and dependencies are managed.

**Implementation Status**:
- ✅ Database schema (tasks, task_dependencies, task_verifications) - COMPLETE
- ✅ Task storage API routes registered - COMPLETE
- ✅ Task detection algorithm (TODO comments, task markers, Cursor format) - COMPLETE
- ✅ Task CRUD handlers (create, list, get, update, delete) - COMPLETE
- ✅ Multi-factor verification engine - COMPLETE
- ✅ Dependency detection (explicit, implicit, integration, feature-level) - COMPLETE
- ✅ Integration with existing systems (Phase 11, 12, 14A, 10, 4) - COMPLETE
- ✅ CLI integration (`sentinel tasks` command) - COMPLETE
- ✅ MCP integration (sentinel_get_task_status, sentinel_verify_task, sentinel_list_tasks) - COMPLETE

**Commands**:
```bash
sentinel tasks scan              # Scan codebase for tasks
sentinel tasks list              # List all tasks
sentinel tasks verify TASK-123   # Verify specific task
sentinel tasks verify --all     # Verify all pending tasks
sentinel tasks dependencies      # Show dependency graph
sentinel tasks complete TASK-123 # Manually mark task complete
```

**Features**:

1. **Task Detection**:
   - Scans codebase for TODO comments, task markers, Cursor task format
   - Identifies task source (cursor, manual, change_request, comprehensive_analysis)
   - Extracts task metadata (title, description, file, line number)

2. **Multi-Factor Verification**:
   - **Code Existence**: AST search for function/class/feature
   - **Code Usage**: Cross-file reference analysis
   - **Test Coverage**: Test file existence and coverage
   - **Integration**: External API/service integration verification
   - Confidence scoring (0.0-1.0) for each factor

3. **Dependency Detection**:
   - **Explicit**: Parsed from task descriptions ("Depends on: TASK-123")
   - **Implicit**: Detected through code analysis (Task A calls Task B's code)
   - **Integration**: Detected through comprehensive analysis (external APIs/services)
   - **Feature-Level**: Detected through comprehensive feature analysis (Phase 14A)
   - Dependency graph building and cycle detection

4. **Auto-Completion**:
   - High confidence (>0.8) → Auto-mark as completed
   - Medium confidence (0.5-0.8) → Mark as in_progress, alert developer
   - Low confidence (<0.5) → Keep as pending, require manual verification

5. **Integration with Existing Systems**:
   - **Phase 11 (Doc-Sync)**: Reuse `detectBusinessRuleImplementation()` pattern
   - **Phase 12 (Change Requests)**: Link tasks to change requests, auto-create tasks
   - **Phase 14A (Comprehensive Analysis)**: Use feature discovery for dependencies
   - **Phase 10 (Test Enforcement)**: Verify test-related tasks, link to test requirements
   - **Phase 4 (Knowledge Base)**: Link tasks to business rules

6. **MCP Integration**:
   - `sentinel_get_task_status` - Get task completion status
   - `sentinel_verify_task` - Verify task completion
   - `sentinel_list_tasks` - List all tasks

**Usage Examples**:

**Task Scanning**:
```bash
$ sentinel tasks scan
🔍 Scanning codebase for tasks...
✅ Found 15 tasks:
  TASK-001: Implement user authentication (pending, high)
  TASK-002: Add JWT token refresh (pending, medium)
  TASK-003: Add payment processing (in_progress, critical)
  ...
```

**Task Verification**:
```bash
$ sentinel tasks verify TASK-001
🔍 Verifying task TASK-001: Implement user authentication
  ✓ Code existence: 0.95 (verified)
  ✓ Code usage: 0.88 (verified)
  ✓ Test coverage: 0.92 (verified)
  ✗ Integration: 0.0 (pending)
  
Overall confidence: 0.69 → Status: in_progress
⚠️  Task needs integration verification
```

**Dependency Graph**:
```bash
$ sentinel tasks dependencies TASK-003
📊 Dependency Graph for TASK-003: Add payment processing
  │
  ├── TASK-001: Implement user authentication [explicit]
  │   └── TASK-002: Add JWT token refresh [implicit]
  │
  └── TASK-004: Setup payment gateway [integration]
      └── TASK-005: Configure API keys [explicit]
```

**Auto-Completion**:
```bash
$ sentinel tasks verify --all
🔍 Verifying all pending tasks...
  TASK-001: 0.69 confidence → in_progress
  TASK-002: 0.92 confidence → ✅ auto-completed
  TASK-003: 0.45 confidence → pending
  ...
  
✅ 3 tasks auto-completed
⚠️  5 tasks need attention
```

**MCP Tool Usage**:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "task_id": "TASK-001"
  }
}
```

**Response**:
```json
{
  "task_id": "TASK-001",
  "status": "in_progress",
  "verification": {
    "overall_confidence": 0.69,
    "factors": {
      "code_existence": 0.95,
      "code_usage": 0.88,
      "test_coverage": 0.92,
      "integration": 0.0
    }
  },
  "dependencies": {
    "blocking": [],
    "blocked_by": ["TASK-002"]
  }
}
```

**Database Schema**:
- `tasks` - Task metadata and status
- `task_dependencies` - Task dependency relationships
- `task_verifications` - Verification results and evidence
- `task_links` - Links to other systems (change requests, knowledge items, etc.)

**API Endpoints**:
- `POST /api/v1/tasks` - Create or update task
- `GET /api/v1/tasks` - List tasks with filters
- `GET /api/v1/tasks/{id}` - Get task details
- `POST /api/v1/tasks/{id}/verify` - Verify task completion
- `GET /api/v1/tasks/{id}/dependencies` - Get task dependencies
- `POST /api/v1/tasks/{id}/dependencies` - Add dependency

**Dependencies**:
- Phase 6 (AST Analysis) ✅ - Required for code verification
- Phase 10 (Test Enforcement) ✅ - Required for test task verification
- Phase 11 (Doc-Sync) ✅ - Required for status tracking patterns
- Phase 12 (Change Requests) ✅ - Required for task-to-change-request linking
- Phase 4 (Knowledge Base) ✅ - Required for business rule linking
- Phase 14A (Comprehensive Feature Analysis) ⏳ - Required for feature-level dependencies
- Phase 14D (Cost Optimization) ⏳ - Required for efficient verification

**Reference**: See [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) for complete specification.

---

### Phase 11: Code-Documentation Comparison ✅ COMPLETE

**Status**: ✅ COMPLETE - Implementation status tracking and business rules comparison fully functional.

**Purpose**: Bidirectional validation between code and documentation to prevent documentation drift and ensure code matches documented business rules.

**Implementation Status**:
- ✅ Status marker parser - Extracts phase status from IMPLEMENTATION_ROADMAP.md
- ✅ Code implementation detector - Scans codebase for feature evidence with confidence scores
- ✅ Status comparison engine - Compares documentation status vs code evidence
- ✅ Feature flag validator - Validates flags match documentation
- ✅ API endpoint validator - Validates endpoints match documentation
- ✅ Command validator - Validates commands match documentation
- ✅ Test coverage validator - Validates tests exist for documented features
- ✅ Discrepancy report generator - Generates JSON and human-readable reports
- ✅ Auto-update capability - Generates suggested documentation updates with approval workflow
- ✅ Business rules comparison - Bidirectional comparison between business rules and code
- ✅ Review workflow - Human review queue with approval/rejection tracking
- ✅ HTTP client - Retry logic with exponential backoff
- ✅ Database schema - doc_sync_reports and doc_sync_updates tables

**Commands**:
```bash
# Implementation status tracking
sentinel doc-sync                      # Standalone check
sentinel doc-sync --fix               # Generate and store update suggestions
sentinel doc-sync --report            # Generate compliance report
sentinel doc-sync --output json       # JSON output format
sentinel audit --doc-sync             # Include doc-sync check in audit

# Business rules comparison
sentinel doc-sync business-rules      # Compare business rules vs code
```

**API Endpoints**:
- `POST /api/v1/analyze/doc-sync` - Main doc-sync analysis endpoint
- `POST /api/v1/analyze/business-rules` - Business rules comparison endpoint
- `GET /api/v1/doc-sync/review-queue` - Get pending updates for review
- `POST /api/v1/doc-sync/review/{id}` - Approve/reject update

**Detected Discrepancies**:
| Type | Description | Status |
|------|-------------|--------|
| Status mismatch | Documentation says PENDING but code is COMPLETE | ✅ Detected |
| Missing implementation | Documentation says COMPLETE but code missing | ✅ Detected |
| Missing documentation | Code exists but not documented | ✅ Detected |
| Partial match | Code partially implements documented feature | ✅ Detected |
| Tests missing | Feature marked COMPLETE but no tests found | ✅ Detected |

**Architecture**:
```
Agent (Local)                     Hub (Server)
─────────────                     ─────────────
doc-sync command          ───►   Status marker parser
--fix flag                        Code implementation detector
                                  Status comparison engine
                                  Validators (flags, endpoints, commands, tests)
                                  Report generator
                                  Update storage
                                  Business rules comparison
```

**Reference**: See [DOCUMENTATION_CODE_SYNC_ANALYSIS.md](./DOCUMENTATION_CODE_SYNC_ANALYSIS.md) for detailed analysis.

---

## Feature Interaction Matrix

| Feature | Doc Ingest | Patterns | Fixes | MCP | Hub | Business | Security | Tests |
|---------|------------|----------|-------|-----|-----|----------|----------|-------|
| Doc Ingest | - | - | - | - | Reports | Produces | Produces | Produces |
| Patterns | - | - | Uses | Provides | Reports | - | - | - |
| Fixes | - | Uses | - | Provides | Reports | - | - | - |
| MCP | - | Provides | Provides | - | Uses | Provides | Provides | Provides |
| Hub | Receives | Receives | Receives | Provides | - | Stores | Analyzes | Executes |
| Business | Uses | - | - | Provides | - | - | - | Produces |
| Security | Uses | - | - | Provides | Analyzes | - | - | Produces |
| Tests | - | - | - | Provides | Executes | Uses | Uses | - |

---

## Command Summary

| Command | Purpose | Key Flags |
|---------|---------|-----------|
| `init` | Bootstrap project | `--stack`, `--with-business-docs` |
| `audit` | Scan for issues | `--ci`, `--business-rules` ✅, `--vibe-check` ✅, `--deep` ✅, `--offline` ✅, `--security` ✅ |
| `hook` | Interactive git hooks | `pre-commit` ✅, `pre-push` ✅, `--non-interactive` ✅ |
| `learn` | Extract patterns | `--naming`, `--imports` |
| `fix` | Apply fixes | `--safe` ✅, `--dry-run` ✅, `--split` ⏳ |
| `ingest` | Process documents | `--local-only`, `--provider`, `--sync` |
| `review` | Validate knowledge | `--list`, `--approve` |
| `knowledge` | Knowledge management | `list` ✅, `review` ✅, `approve` ✅, `reject` ✅, `activate` ✅, `gap-analysis` ⏳, `changes` ⏳, `impact` ⏳, `generate-tests` ⏳ |
| `doc-sync` | Documentation-code sync | `--fix` ✅, `--report` ✅, `--output` ✅, `--type` ✅ |
| `test` | ✅ IMPLEMENTED - Test management CLI | `requirements` ✅, `coverage` ✅, `validate` ✅, `run` ✅, `mutation` ✅ |
| `tasks` | ✅ IMPLEMENTED - Task management CLI | `scan` ✅, `list` ✅, `verify` ✅, `dependencies` ✅ |
| `status` | ✅ IMPLEMENTED - Project health overview | - |
| `baseline` | ✅ IMPLEMENTED - Exception management | `list` ✅, `add` ✅, `remove` ✅, `review` ✅ |
| `install-hooks` | Install git hooks | - ✅ |
| `mcp-server` | ✅ IMPLEMENTED - Start MCP server for Cursor integration | Phase 14B |
| `sentinel_check_intent` | ✅ IMPLEMENTED - Analyze unclear prompts and generate clarifying questions | Phase 15 |

---

## Coverage Summary

With full system implementation:

| Category | Coverage | Method |
|----------|----------|--------|
| Structural Issues | 95% | AST (Hub) |
| Refactoring Issues | 95% | Cross-file AST ✅ |
| Security Issues | 85% | Security rules ✅ COMPLETE |
| Business Logic | 90% | Executable rules |
| Test Coverage | 90% | Requirement tracking (⏳ Phase 10) |
| Vibe Coding Issues | 85% | AST + patterns |
| **Overall** | **~85%** | Full system (when all phases complete) |

See [VIBE_CODING_ANALYSIS.md](./VIBE_CODING_ANALYSIS.md) for detailed breakdown.

---

## MCP Tools Status (Phase 14)

Sentinel provides MCP tools for Cursor IDE integration. Current status:

### ✅ IMPLEMENTED MCP Tools (18/18 - Production Ready)

| Tool | Status | Notes |
|------|--------|-------|
| sentinel_analyze_feature_comprehensive | ✅ Implemented | Fully functional - comprehensive feature analysis |
| sentinel_validate_code | ✅ Implemented | Code syntax and business rule validation with AST analysis |
| sentinel_apply_fix | ✅ Implemented | Apply security/style/performance fixes |
| sentinel_validate_security | ✅ Implemented | Security vulnerability analysis with severity levels |
| sentinel_get_business_context | ✅ Implemented | Business rules and entities retrieval |
| sentinel_validate_business | ✅ Implemented | Business rule compliance validation |
| sentinel_analyze_intent | ✅ Implemented | Intent clarification and analysis |
| sentinel_get_patterns | ✅ Implemented | Pattern recognition and retrieval |
| sentinel_get_context | ✅ Implemented | General project context retrieval |
| sentinel_get_security_context | ✅ Implemented | Security context and vulnerability info |
| sentinel_get_task_status | ✅ Implemented | Task status retrieval - Phase 2 |
| sentinel_verify_task | ✅ Implemented | Task verification - Phase 2 |
| sentinel_list_tasks | ✅ Implemented | Task listing - Phase 2 |
| sentinel_get_test_requirements | ✅ Implemented | Test requirements generation - Phase 2 |
| sentinel_validate_tests | ✅ Implemented | Test validation - Phase 2 |
| sentinel_generate_tests | ✅ Implemented | Test case generation - Phase 2 |
| sentinel_run_tests | ✅ Implemented | Test execution in sandbox - Phase 2 |
| sentinel_check_file_size | ✅ Implemented | File size analysis and recommendations - Phase 2 |
| sentinel_check_intent | ✅ Implemented | Intent analysis and clarification - Phase 2 |

**Summary**:
- **Implemented**: 19/19 tools (100% complete) - Production ready ✅
- **All Planned MCP Tools**: Successfully implemented in Phase 2
- **Feature Completeness**: 100% of documented MCP functionality

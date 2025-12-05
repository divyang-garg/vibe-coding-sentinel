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

### Phase D: Test Enforcement System ⏳ NOT YET IMPLEMENTED

**Status**: Documented but not implemented. Scheduled for Phase 10.

**Purpose**: Ensure business rules have corresponding tests.

**Commands**:
```bash
sentinel audit --test-coverage     # ⏳ NOT IMPLEMENTED - Check test coverage
sentinel knowledge generate-tests # ⏳ NOT IMPLEMENTED - Generate test cases
sentinel test validate             # ⏳ NOT IMPLEMENTED - Validate test quality
sentinel test run                  # ⏳ NOT IMPLEMENTED - Execute tests (Hub)
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

### Phase E: Requirements Lifecycle Management ⏳ NOT YET IMPLEMENTED

**Status**: Documented but not implemented. Scheduled for Phase 12.

**Purpose**: Track requirements changes and ensure code stays in sync.

**Commands**:
```bash
sentinel knowledge gap-analysis     # ⏳ NOT IMPLEMENTED - Find gaps
sentinel knowledge changes          # ⏳ NOT IMPLEMENTED - Show pending changes
sentinel knowledge impact BR-001   # ⏳ NOT IMPLEMENTED - Impact analysis
sentinel knowledge generate-tasks  # ⏳ NOT IMPLEMENTED - Generate migration tasks
```

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

### Phase 4: MCP Integration (Enhanced) ⚠️ STUB IMPLEMENTATION

**Status**: ⚠️ STUB - Command exists but not functional. Scheduled for Phase 14 (requires Phases 6-10 to be complete first).

**Implementation Status**:
- ✅ `mcp-server` command registered
- ✅ `runMCPServer()` function exists
- ⚠️ MCP protocol handler not implemented (exits immediately)
- ⚠️ All MCP tools pending (require foundation phases)

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

### Phase 14A: Comprehensive Feature Analysis ⏳ PENDING

**Status**: ⏳ Pending (Phase 14A)

**Purpose**: End-to-end feature analysis across all layers (UI, API, Database, Logic, Integration, Tests) with business context validation to ensure comprehensive coverage beyond surface-level checks.

**Commands**:
- MCP tool: `sentinel_analyze_feature_comprehensive`
- Hub API: `POST /api/v1/analyze/comprehensive`
- Hub Dashboard: View results at `/validations/{id}`

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
| `learn` | Extract patterns | `--naming`, `--imports` |
| `fix` | Apply fixes | `--safe` ✅, `--dry-run` ✅, `--split` ⏳ |
| `ingest` | Process documents | `--local-only`, `--provider`, `--sync` |
| `review` | Validate knowledge | `--list`, `--approve` |
| `knowledge` | Knowledge management | `list` ✅, `review` ✅, `approve` ✅, `reject` ✅, `activate` ✅, `gap-analysis` ⏳, `changes` ⏳, `impact` ⏳, `generate-tests` ⏳ |
| `test` | ⏳ NOT IMPLEMENTED - Test management | `validate` ⏳, `run` ⏳, `coverage` ⏳ |
| `status` | Project health | - |
| `baseline` | Manage exceptions | `add`, `remove`, `list` |
| `mcp-server` | ⏳ NOT IMPLEMENTED - Start MCP mode | - |

---

## Coverage Summary

With full system implementation:

| Category | Coverage | Method |
|----------|----------|--------|
| Structural Issues | 95% | AST (Hub) |
| Refactoring Issues | 95% | Cross-file AST ✅ |
| Security Issues | 85% | Security rules (⚠️ STUB - Phase 8) |
| Business Logic | 90% | Executable rules |
| Test Coverage | 90% | Requirement tracking (⏳ Phase 10) |
| Vibe Coding Issues | 85% | AST + patterns |
| **Overall** | **~85%** | Full system (when all phases complete) |

See [VIBE_CODING_ANALYSIS.md](./VIBE_CODING_ANALYSIS.md) for detailed breakdown.

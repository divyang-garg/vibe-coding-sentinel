# System Architecture

## Overview

Sentinel uses a distributed architecture with local agents on developer machines and an optional central hub for organizational visibility.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SENTINEL PLATFORM ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DOCUMENT INGESTION LAYER                                               │
│  ═════════════════════════                                              │
│                                                                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │  PDF    │ │  Word   │ │  Excel  │ │  Image  │ │  Email  │          │
│  │ Scope   │ │ Require-│ │ Data    │ │ Wire-   │ │ Client  │          │
│  │ Doc     │ │ ments   │ │ Models  │ │ frames  │ │ Comms   │          │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘          │
│       │          │          │          │          │                    │
│       └──────────┴──────────┴──────────┴──────────┘                    │
│                             │                                           │
│                             ▼                                           │
│                    Document Parser + LLM Extraction                     │
│                             │                                           │
│                             ▼                                           │
│                    Structured Knowledge (with human review)             │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DEVELOPER LAYER (Local - Code Never Leaves)                            │
│  ════════════════════════════════════════════                           │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                    SENTINEL AGENT (Go Binary)                   │    │
│  ├────────────────────────────────────────────────────────────────┤    │
│  │                                                                 │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │    │
│  │  │ CORE ENGINE │  │ MCP SERVER  │  │ TELEMETRY   │            │    │
│  │  ├─────────────┤  ├─────────────┤  ├─────────────┤            │    │
│  │  │ • Scanning  │  │ • Tools     │  │ • Metrics   │            │    │
│  │  │ • Patterns  │  │ • Resources │  │ • Events    │            │    │
│  │  │ • Fixing    │  │ • Protocol  │  │ • Queue     │            │    │
│  │  │ • Context   │  │             │  │             │            │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘            │    │
│  │         │                │                │                    │    │
│  │         └────────────────┼────────────────┘                    │    │
│  │                          │                                     │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │    │
│  │  │ CLI CMDS    │  │ GIT HOOKS   │  │ DOC INGEST  │            │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘            │    │
│  │                                                                 │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     │                                   │
│                                     │ HTTPS (Metrics Only)             │
│                                     ▼                                   │
│  ORGANIZATION LAYER (Central - Metrics & Management)                   │
│  ════════════════════════════════════════════════════                  │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                    SENTINEL HUB                                 │    │
│  ├────────────────────────────────────────────────────────────────┤    │
│  │                                                                 │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │    │
│  │  │ API SERVER  │  │ DATABASE    │  │ DASHBOARD   │            │    │
│  │  ├─────────────┤  ├─────────────┤  ├─────────────┤            │    │
│  │  │ • Ingest    │  │ • Metrics   │  │ • Overview  │            │    │
│  │  │ • Query     │  │ • Orgs      │  │ • Config    │            │    │
│  │  │ • Auth      │  │ • Teams     │  │ • Usage     │            │    │
│  │  │ • Patterns  │  │ • Patterns  │  │ • Optimize  │            │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘            │    │
│  │                                                                 │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  FUTURE SAAS LAYER                                                      │
│  ════════════════                                                       │
│  • Multi-tenancy  • Billing  • Public API  • Self-service              │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### Document Ingestion Module

Processes raw project documents and extracts structured knowledge.

**Supported Formats**:
| Format | Parser | Notes |
|--------|--------|-------|
| PDF | pdftotext/poppler | Text extraction |
| Word (.docx) | unioffice | Native Go |
| Word (.doc) | LibreOffice | Legacy format |
| Excel (.xlsx) | excelize | Native Go |
| Images | Tesseract + Vision API | OCR + diagram analysis |
| Email (.eml) | net/mail | Headers + body + attachments |
| Text | Direct read | Plain text files |

**Processing Flow**:
1. Parse document locally (extract text)
2. Send text to LLM for knowledge extraction
3. Generate draft documents with confidence scores
4. Human reviews and approves
5. Approved docs become active knowledge

### Sentinel Agent

The agent is a compiled Go binary that runs on each developer machine.

**Core Engine**
- Pattern learning from existing code
- Security and code scanning (17+ patterns)
- Safe auto-fix application
- Context gathering for LLM
- Business knowledge integration

**MCP Server**
- Model Context Protocol implementation
- Real-time Cursor integration
- Tool exposure for LLM calls
- Intent clarification
- Business context provision

**Telemetry Client**
- Metrics collection (no code)
- Offline queue with retry
- Data sanitization
- Hub communication

### Test Enforcement System (Phase 10) ✅ IMPLEMENTED

**Purpose**: Ensure business rules have corresponding tests that are systematically written, validated, and enforced.

**Architecture**:
```
Agent (Local)                     Hub (Server)
─────────────                     ─────────────
Test Commands                ───►   Test Requirement Generator
  ├─ --requirements                 Test Coverage Tracker
  ├─ --coverage                     Test Validator
  ├─ --validate                     Mutation Testing Engine
  ├─ --mutation                     Test Execution Sandbox (Docker)
  └─ --run                          Database (test_requirements, test_coverage, etc.)
```

**Components**:
1. **Test Requirement Generation** (Hub): Extracts business rules from knowledge base, generates test requirements
2. **Test Coverage Tracking** (Hub): Analyzes test files (content-based), maps to business rules, calculates coverage
3. **Test Validation** (Hub): Validates test structure, assertions, completeness
4. **Mutation Testing Engine** (Hub): Generates mutants, executes tests against mutants, calculates mutation score
5. **Test Execution Sandbox** (Hub): Docker-based isolated test execution with resource limits
6. **Agent Commands** (Local): CLI interface for all test enforcement features

**Database Tables**:
- `test_requirements` - Generated test requirements from business rules
- `test_coverage` - Coverage tracking per business rule
- `test_validations` - Test validation results
- `mutation_results` - Mutation testing results
- `test_executions` - Test execution records

**API Endpoints**:
- `POST /api/v1/test-requirements/generate` - Generate test requirements
- `POST /api/v1/test-coverage/analyze` - Analyze test coverage
- `GET /api/v1/test-coverage/{knowledge_item_id}` - Get coverage
- `POST /api/v1/test-validations/validate` - Validate tests
- `GET /api/v1/test-validations/{test_requirement_id}` - Get validation
- `POST /api/v1/mutation-test/run` - Run mutation testing
- `GET /api/v1/mutation-test/{test_requirement_id}` - Get mutation results
- `POST /api/v1/test-execution/run` - Execute tests in sandbox
- `GET /api/v1/test-execution/{execution_id}` - Get execution status

### Sentinel Hub

The hub is a central server for organizational visibility.

**API Server**
- Telemetry ingestion
- Metrics query API
- Organization management
- Pattern distribution

**Database**
- PostgreSQL for metrics storage
- Row-level security for multi-tenancy
- Audit logging

**Dashboard**
- LLM provider configuration
- Usage monitoring and analytics
- Cost optimization recommendations
- API key management

### Task Dependency & Verification Module (Phase 14E)

The task dependency and verification module tracks Cursor-generated tasks, verifies completion, and manages dependencies.

**Architecture**:
```
Agent (Local)                     Hub (Server)
─────────────                     ─────────────
tasks scan command         ───►   Task Detection Engine
tasks verify command              Dependency Analyzer
tasks list command                Verification Engine
                                  Auto-Completion System
                                  Alert System
                                  Database (tasks, task_dependencies, task_verifications)
```

**Components**:

1. **Task Detection Engine**:
   - Scans codebase for TODO comments, task markers, Cursor task format
   - Extracts task metadata (title, description, file, line number)
   - Identifies task source (cursor, manual, change_request, comprehensive_analysis)

2. **Dependency Analyzer**:
   - Explicit dependency parsing (from task descriptions)
   - Implicit dependency detection (code analysis)
   - Integration dependency detection (Phase 14A feature discovery)
   - Feature-level dependency detection (Phase 14A comprehensive analysis)
   - Dependency graph building and cycle detection

3. **Verification Engine**:
   - Multi-factor verification (code existence, usage, tests, integration)
   - AST-based code verification (Phase 6)
   - Test coverage verification (Phase 10)
   - Integration verification (external APIs/services)
   - Confidence scoring algorithm

4. **Auto-Completion System**:
   - High confidence (>0.8) → Auto-mark as completed
   - Medium confidence (0.5-0.8) → Mark as in_progress, alert developer
   - Low confidence (<0.5) → Keep as pending, require manual verification

5. **Integration Layer**:
   - Links tasks to change requests (Phase 12)
   - Links tasks to knowledge items (Phase 4)
   - Links tasks to comprehensive analysis results (Phase 14A)
   - Links tasks to test requirements (Phase 10)
   - Status synchronization

**Database Tables**:
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

**Integration Points**:
- Phase 11 (Doc-Sync): Reuse `detectBusinessRuleImplementation()` pattern
- Phase 12 (Change Requests): Link tasks to change requests, auto-create tasks
- Phase 14A (Comprehensive Analysis): Use feature discovery for dependencies
- Phase 10 (Test Enforcement): Verify test-related tasks, link to test requirements
- Phase 4 (Knowledge Base): Link tasks to business rules

**Reference**: See [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) for complete specification.

---

### Comprehensive Analysis Module

The comprehensive analysis module provides end-to-end feature analysis across all layers with business context validation.

**Architecture**:
```
┌─────────────────────────────────────────────────────────────────┐
│                    COMPREHENSIVE ANALYSIS FLOW                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Cursor (MCP Request)                                            │
│      │                                                           │
│      ▼                                                           │
│  Agent: sentinel_analyze_feature_comprehensive                  │
│      │                                                           │
│      ├── Feature Discovery (Auto/Manual)                        │
│      │   ├── UI Components                                      │
│      │   ├── API Endpoints                                      │
│      │   ├── Database Tables                                    │
│      │   ├── Business Logic Functions                           │
│      │   ├── Integration Points                                 │
│      │   └── Test Files                                         │
│      │                                                           │
│      ▼                                                           │
│  Hub: POST /api/v1/analyze/comprehensive                        │
│      │                                                           │
│      ├── Layer-Specific Analysis (7 Layers)                     │
│      │   ├── Business Context (Rules, Journeys, Entities)      │
│      │   ├── UI Layer (Components, Forms, Validation)          │
│      │   ├── API Layer (Endpoints, Security, Middleware)        │
│      │   ├── Database Layer (Schema, Migrations, Integrity)    │
│      │   ├── Business Logic (AST, Cross-File, Semantic)        │
│      │   ├── Integration Layer (External APIs, Contracts)      │
│      │   └── Test Layer (Coverage, Quality, Edge Cases)         │
│      │                                                           │
│      ├── End-to-End Flow Verification                            │
│      │   ├── Flow Detection Across Layers                        │
│      │   ├── Breakpoint Identification                           │
│      │   └── Integration Verification                            │
│      │                                                           │
│      ├── LLM Semantic Analysis                                  │
│      │   ├── Business Logic Correctness                         │
│      │   ├── Requirement Compliance                             │
│      │   └── Edge Case Identification                           │
│      │                                                           │
│      ▼                                                           │
│  Result Aggregation & Checklist Generation                      │
│      │                                                           │
│      ▼                                                           │
│  Hub Storage + URL Generation                                   │
│      │                                                           │
│      ▼                                                           │
│  Agent Response to Cursor                                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Components**:

1. **Feature Discovery Engine**
   - Auto-discovers features across all layers
   - Keyword-based component mapping
   - Manual file specification support

2. **Layer-Specific Analyzers** (7 analyzers)
   - Business Context Analyzer (validates against knowledge base)
   - UI Layer Analyzer (components, forms, validation)
   - API Layer Analyzer (endpoints, security, middleware)
   - Database Layer Analyzer (schema, migrations, integrity)
   - Business Logic Analyzer (AST, cross-file, semantic)
   - Integration Layer Analyzer (external APIs, contracts)
   - Test Layer Analyzer (coverage, quality, edge cases)

3. **End-to-End Flow Verifier**
   - Flow detection across layers
   - Breakpoint identification
   - Integration verification

4. **LLM Integration Layer**
   - API key management (user-provided or org-shared)
   - Model selection (GPT-5.1-Codex-Max, GPT-5.1 Instant, etc.)
   - Cost optimization (caching, progressive depth)
   - Token tracking (reporting only, not billing)

5. **Result Aggregator**
   - Combines findings into prioritized checklists
   - Generates summaries per layer
   - Creates end-to-end flow reports

**Integration Points**:
- **AST Analysis Engine** (Phase 6): Uses AST for business logic analysis
- **Security Rules System** (Phase 8): Uses security rules for API layer analysis
- **Knowledge Base** (Phase 4): Uses business rules, journeys, entities for validation
- **MCP Server**: Exposes `sentinel_analyze_feature_comprehensive` tool

**Data Flow**:
- Agent collects feature files → Hub
- Hub performs layer-specific analysis → Results
- Hub performs LLM semantic analysis → Findings
- Hub aggregates results → Checklist
- Hub stores results → Database
- Hub returns URL → Agent → Cursor

**API Endpoints**:
- `POST /api/v1/analyze/comprehensive` - Request comprehensive analysis
- `GET /api/v1/validations/{id}` - Get analysis results
- `GET /api/v1/validations?project={id}` - List analyses for project

**Configuration**:
- LLM provider selection (OpenAI, Anthropic, Azure)
- API key management (encrypted storage)
- Model selection (high-accuracy vs. cost-optimized)
- Cost optimization settings (caching, progressive depth)
- Usage tracking (reporting only, not billing)

**Important Notes**:
- **API Key Management**: Users/Organizations subscribe to LLM providers separately and provide API keys to Hub. Sentinel does NOT handle billing or payments.
- **Cost Tracking**: Sentinel tracks usage for reporting only, not billing. Users pay LLM providers directly.
- **Integration Analysis**: When analyzing features, Sentinel identifies external integrations (e.g., payment gateways) as part of the FEATURE being analyzed, not Sentinel's own functionality.

## Data Flow

### What Stays Local (Never Leaves)

| Data | Reason |
|------|--------|
| Source code | Proprietary |
| File contents | Security |
| Secrets detected | Sensitive |
| Full audit details | May contain code |
| Fix operations | Local modification |
| Original documents | May be confidential |

### What Goes to Hub (Metrics Only)

| Data | Example | Purpose |
|------|---------|---------|
| Finding counts | "3 critical, 5 warnings" | Track quality |
| Compliance % | "92% naming compliance" | Track consistency |
| Fix counts | "Applied 10 safe fixes" | Track automation |
| Error types | "SQL injection detected" | Track patterns |
| Usage stats | "50 audits today" | Track adoption |
| Doc coverage | "85% rules documented" | Track completeness |

## Knowledge Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    KNOWLEDGE FLOW                                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  RAW DOCUMENTS                                                          │
│  ─────────────                                                          │
│  Scope.pdf, Requirements.docx, client_emails.eml                        │
│                           │                                              │
│                           ▼                                              │
│  DOCUMENT PARSING (Local)                                               │
│  ────────────────────────                                               │
│  Extract text, tables, images                                           │
│                           │                                              │
│                           ▼                                              │
│  LLM EXTRACTION (Cloud/Local)                                           │
│  ────────────────────────────                                           │
│  Identify entities, rules, journeys                                     │
│                           │                                              │
│                           ▼                                              │
│  DRAFT DOCUMENTS                                                        │
│  ───────────────                                                        │
│  *.draft.md with confidence scores                                      │
│                           │                                              │
│                           ▼                                              │
│  HUMAN REVIEW                                                           │
│  ────────────                                                           │
│  Accept / Edit / Reject items                                           │
│                           │                                              │
│                           ▼                                              │
│  APPROVED KNOWLEDGE                                                     │
│  ──────────────────                                                     │
│  domain-glossary.md, business-rules.md, entities/*.md                   │
│                           │                                              │
│                           ▼                                              │
│  CURSOR / MCP                                                           │
│  ────────────                                                           │
│  Business context available during code generation                      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Security Architecture

### Agent Security

| Measure | Implementation |
|---------|----------------|
| Binary compilation | Source deleted after build |
| No external deps | Pure Go, no C libraries |
| Input validation | All user input sanitized |
| Path traversal protection | Validated before use |
| Concurrent execution lock | Prevent race conditions |

### Hub Security

| Measure | Implementation |
|---------|----------------|
| Authentication | OAuth/SSO, API keys |
| Authorization | RBAC with roles |
| Data isolation | Row-level security |
| Encryption | TLS in transit, AES at rest |
| Audit logging | All actions logged |

### Document Security

| Measure | Implementation |
|---------|----------------|
| Local parsing | Documents never leave machine |
| Text-only to LLM | Only extracted text sent |
| Local-only option | Full offline processing available |
| Source archival | Originals kept for reference |

### Binary Security

| Feature | Implementation |
|---------|----------------|
| Source deletion | `rm main.go` after compile |
| Symbol stripping | `-ldflags="-s -w"` |
| No debug info | Release builds only |
| Reproducible | Deterministic compilation |

## Integration Points

### Cursor IDE (via MCP)

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "./sentinel",
      "args": ["mcp-server"]
    }
  }
}
```

### Interactive Git Hooks Module (Phase 9.5)

**Purpose**: Interactive git hooks with user warnings, options, and comprehensive integration with Hub for organizational governance.

**Architecture**:
```
Git Hook (pre-commit/pre-push)
    │
    ▼
sentinel hook <type> [--non-interactive]
    │
    ▼
runInteractiveHook()
    │
    ├── performAuditForHook() → AuditReport
    │
    ├── Parse findings by severity
    │
    ├── Interactive Menu (if not --non-interactive)
    │   ├── [v] View details
    │   ├── [p] Proceed anyway (check policy)
    │   ├── [b] Add to baseline (check policy)
    │   ├── [e] Add exception (check policy)
    │   └── [q] Quit
    │
    ├── Save audit history (with HookContext)
    │
    └── sendHookTelemetry() → Hub
        │
        ▼
    Hub: POST /api/v1/telemetry/hook
        │
        ▼
    Store in hook_executions table
        │
        ▼
    Aggregate for metrics dashboard
```

**Components**:

1. **Interactive Hook Handler** (`runInteractiveHook()`):
   - Runs audit and captures report
   - Parses findings by severity
   - Displays interactive menu
   - Handles user input
   - Enforces policies
   - Tracks user actions

2. **Policy System**:
   - Policies stored in Hub (`hook_policies` table)
   - Cached locally (5 minutes)
   - Checked before allowing overrides/baselines
   - Configurable per organization

3. **Telemetry System**:
   - Hook execution events sent to Hub
   - Aggregated metrics available via API
   - Team-level breakdowns
   - Trend analysis

4. **Baseline Review Workflow**:
   - Hook-added baselines marked "pending_review"
   - Auto-approved after configured days
   - Admin approval/rejection via Hub

**Data Flow**:
- Hook execution → Audit → Interactive menu → User action → Policy check → Save history → Send telemetry → Hub storage → Metrics aggregation

**Hub Integration**:
- `POST /api/v1/telemetry/hook` - Ingest hook events
- `GET /api/v1/hooks/metrics` - Get aggregated metrics
- `GET /api/v1/hooks/policies` - Get policy configuration
- `POST /api/v1/hooks/policies` - Create/update policy

**Database Schema**:
- `hook_executions` - Hook execution events
- `hook_baselines` - Baseline entries from hooks
- `hook_policies` - Policy configurations

**Reference**: See [INTERACTIVE_HOOKS_ANALYSIS.md](./INTERACTIVE_HOOKS_ANALYSIS.md) for complete analysis.

---

### Reliability Layer (Phase 9.5.1) ✅ IMPLEMENTED

**Purpose**: Ensure system reliability, prevent resource leaks, and provide graceful error handling.

**Components**:

1. **Database Connection Management**:
   - Connection pool health monitoring (`monitorDBHealth()`)
   - Connection lifetime management (`SetConnMaxLifetime`)
   - Pool metrics logging (open, idle, in-use connections)
   - Exhaustion alerts
   - Background health check goroutine (30-second intervals)

2. **Database Query Timeouts**:
   - Context-aware timeout helpers (`queryWithTimeout`, `queryRowWithTimeout`, `execWithTimeout`)
   - 10-second default timeout
   - Automatic context cancellation on timeout
   - Prevents connection pool exhaustion
   - Used in all Hub API handlers

3. **HTTP Retry Logic**:
   - Exponential backoff retry (`httpRequestWithRetry`)
   - Retries on network errors and 5xx server errors
   - No retry on 4xx client errors
   - Configurable max retries (default: 3)
   - Used for Hub communication, baseline submission, telemetry

4. **Cache Management**:
   - **Policy Cache**: RWMutex for thread-safe access, timestamp-based invalidation
   - **Limits Cache**: Per-entry expiration, thread-safe map with RWMutex
   - **AST Cache**: Time-based periodic cleanup to prevent resource leaks
   - Cache corruption detection and automatic cleanup
   - Cache invalidation based on Hub `updated_at` timestamps

5. **Error Recovery System**:
   - `CheckResult` struct tracks check status (enabled, success, error, findings count)
   - `CheckResults` map in `AuditReport` tracks all check types
   - Error wrapper functions with panic recovery
   - Detailed panic logging with context
   - Finding count tracking (before/after) for accurate reporting

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│              RELIABILITY LAYER                           │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Database Layer                                         │
│  ┌──────────────────────────────────────────────┐     │
│  │ queryWithTimeout()                            │     │
│  │ queryRowWithTimeout()                         │     │
│  │ execWithTimeout()                             │     │
│  │ monitorDBHealth()                             │     │
│  └──────────────────────────────────────────────┘     │
│                                                          │
│  HTTP Layer                                             │
│  ┌──────────────────────────────────────────────┐     │
│  │ httpRequestWithRetry()                        │     │
│  │ - Exponential backoff                        │     │
│  │ - Retry on 5xx errors                        │     │
│  └──────────────────────────────────────────────┘     │
│                                                          │
│  Cache Layer                                            │
│  ┌──────────────────────────────────────────────┐     │
│  │ Policy Cache (RWMutex)                       │     │
│  │ Limits Cache (per-entry expiration)           │     │
│  │ AST Cache (time-based cleanup)                │     │
│  └──────────────────────────────────────────────┘     │
│                                                          │
│  Error Recovery Layer                                   │
│  ┌──────────────────────────────────────────────┐     │
│  │ CheckResults map                              │     │
│  │ Error wrapper functions                      │     │
│  │ Panic recovery                               │     │
│  └──────────────────────────────────────────────┘     │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Integration Points**:
- Hub API handlers use timeout helpers for all database queries
- Agent uses retry logic for all Hub communication
- Cache management integrated into policy and limits fetching
- Error recovery integrated into hook execution and audit workflows

### Git (via Hooks) ✅ ENHANCED (Phase 9.5)

```
.git/hooks/
├── pre-commit    # ✅ Interactive hook (sentinel hook pre-commit)
├── pre-push      # ✅ Interactive hook (sentinel hook pre-push)
└── commit-msg    # Message validation (non-interactive)
```

**Interactive Hooks** (Phase 9.5):
- User warnings with options
- Policy enforcement
- Telemetry tracking
- Baseline review workflow
- CI/CD mode (`--non-interactive` flag)

### CI/CD (via CLI)

```yaml
- name: Sentinel Audit
  run: ./sentinel audit --ci
```

### LLM Providers (for Document Ingestion)

| Provider | Use Case |
|----------|----------|
| OpenAI GPT-4 | High quality text extraction |
| OpenAI GPT-4V | Image/diagram analysis |
| Anthropic Claude | Alternative provider |
| Ollama (local) | Privacy-sensitive documents |

## File System Structure

```
project/
├── .sentinel/
│   ├── patterns.json       # Learned patterns
│   ├── decisions.json      # Developer decisions
│   ├── history.json        # Audit history
│   ├── context.json        # Current context
│   └── backups/            # Fix backups
│       └── {timestamp}/
│
├── .sentinelsrc            # Project config
├── .sentinel-baseline.json # Baselined findings
│
├── .cursor/
│   └── rules/
│       ├── 00-constitution.md
│       ├── 01-business-context.md
│       └── project-patterns.md
│
└── docs/
    └── knowledge/
        ├── source-documents/       # Original uploads (archived)
        │   ├── Scope_v2.pdf
        │   ├── Requirements.docx
        │   └── manifest.json
        │
        ├── extracted/              # Raw extraction (intermediate)
        │   ├── Scope_v2.txt
        │   └── Data_Model.json
        │
        ├── drafts/                 # Pending human review
        │   ├── domain-glossary.draft.md
        │   └── review-status.json
        │
        └── business/               # Approved (active)
            ├── domain-glossary.md
            ├── business-rules.md
            ├── user-journeys.md
            ├── objectives.md
            └── entities/
                ├── user.md
                ├── order.md
                └── payment.md
```

## Scalability

### Agent
- Parallel scanning with goroutines
- File size limits (10MB)
- Configurable scan directories
- Incremental processing

### Hub
- Horizontal API scaling (stateless)
- Database connection pooling
- Metric aggregation (hourly/daily)
- CDN for dashboard assets


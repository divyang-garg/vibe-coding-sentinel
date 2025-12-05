# Implementation Roadmap

## Timeline Overview

```
WEEK 1-2    Foundation & Testing                    ✅ DONE
WEEK 3      Pattern Learning                        ✅ DONE
WEEK 4      Safe Auto-Fix                           ✅ DONE
WEEK 5      Document Ingestion (Local Parsing)      ✅ DONE
WEEK 6      Hub Document Service (Server-Side)      ✅ DONE
WEEK 7-8    Sentinel Hub MVP + Doc Processing      ✅ DONE
WEEK 9      LLM Knowledge Extraction               ✅ DONE
WEEK 9+     Azure AI Foundry Integration            ✅ DONE
WEEK 10-11  AST Analysis Engine (Hub)               ✅ COMPLETE (Phase 6)
WEEK 12-13  Vibe Coding Detection                   ✅ COMPLETE (Phase 7)
WEEK 13-14  Security Rules System                   ✅ COMPLETE (Phase 8)
WEEK 15     File Size Management                    ✅ COMPLETE (Phase 9)
WEEK 20-21  Comprehensive Feature Analysis           ⏳ PENDING (Phase 14A-14D)
WEEK 22-23  MCP Integration                         🔴 STUB (Phase 14)
WEEK 15     Intent & Simple Language                ⏳ Pending
WEEK 16     Organization Features                   ⏳ Pending
WEEK 17     Hardening & Documentation               ⏳ Pending
```

> **Architecture Decision**: Document processing moved from local (Agent) to server (Hub).
> See [ARCHITECTURE_DOCUMENT_PROCESSING.md](./ARCHITECTURE_DOCUMENT_PROCESSING.md) for details.

---

## Phase 0: Foundation Hardening (Week 1-2) ✅ COMPLETED

**Goal**: Ensure existing code is stable and tested.

**Status**: Completed on 2024-12-04

### Tasks

| Task | Days | Deliverable | Status |
|------|------|-------------|--------|
| Create test fixtures | 1 | fixtures/ directory | ✅ Done |
| Unit tests for scanning | 1 | scanning_test.sh | ✅ Done |
| Integration tests | 1 | workflow_test.sh | ✅ Done |
| CI pipeline setup | 0.5 | GitHub Actions | ✅ Done |
| Implement `status` command | 0.5 | runStatus() | ✅ Done |
| Test runner | 0.5 | run_all_tests.sh | ✅ Done |

### Implemented Test Structure

```
tests/
├── fixtures/
│   ├── projects/
│   │   ├── javascript/        # camelCase patterns, React components
│   │   │   ├── src/utils/helpers.js
│   │   │   ├── src/services/userService.js
│   │   │   ├── src/components/Button.jsx
│   │   │   └── package.json
│   │   ├── python/            # snake_case patterns
│   │   │   ├── src/utils/helpers.py
│   │   │   ├── src/services/user_service.py
│   │   │   └── requirements.txt
│   │   └── shell/             # Shell script patterns
│   │       └── scripts/
│   │           ├── deploy.sh
│   │           └── utils.sh
│   ├── security/              # Vulnerable code samples
│   │   ├── secrets_vulnerable.js       # 8+ secrets, console.logs
│   │   ├── sql_injection_vulnerable.php # SQL injection, eval, XXE
│   │   ├── shell_vulnerable.sh         # Unquoted vars, eval
│   │   ├── nosql_vulnerable.js         # $where, NoSQL injection
│   │   └── clean_code.js               # Clean file (0 findings)
│   ├── config/
│   │   ├── valid_config.json
│   │   ├── minimal_config.json
│   │   └── invalid_config.json
│   └── documents/             # Placeholder for Phase 3
├── unit/
│   └── scanning_test.sh       # 11 tests, 100% pass
├── integration/
│   └── workflow_test.sh       # 11 tests, 100% pass
├── run_all_tests.sh           # Master test runner
└── README.md                  # Test documentation
```

### New `status` Command

```bash
$ ./sentinel status

📊 PROJECT HEALTH
══════════════════════════════════════════════════════════════

✅ Compliance:    92% (↑3% from last)
   Last audit:     2 hours ago
   Findings:       0 critical, 3 warning, 0 info

🔧 CONFIGURATION
──────────────────────────────────────────────────────────────
✅ Config:         .sentinelsrc found
✅ Cursor Rules:   3 files in .cursor/rules/
📋 Patterns:       Not learned yet (run: sentinel learn)
✅ Git Hooks:      Installed

📈 OVERALL HEALTH
──────────────────────────────────────────────────────────────
   Score: [████████░░] 80% - Good
```

### Exit Criteria

- ✅ All tests pass (22/22 = 100%)
- ✅ CI pipeline configured (.github/workflows/ci.yml)
- ✅ No regressions in existing features
- ✅ `status` command implemented

---

## Phase 1: Pattern Learning (Week 3) ✅ COMPLETED

**Goal**: Enable automatic pattern detection.

**Status**: Completed on 2024-12-04

### Tasks

| Task | Days | Status |
|------|------|--------|
| Pattern type definitions | 0.5 | ✅ Done |
| Naming detection | 1 | ✅ Done |
| Import detection | 1 | ✅ Done |
| Structure detection | 0.5 | ✅ Done |
| Code style detection | 0.5 | ✅ Done |
| Pattern storage | 0.5 | ✅ Done |
| Cursor rules generation | 0.5 | ✅ Done |
| Tests | 1 | ✅ Done (16 tests) |

### Implemented Functions

```go
func runLearn(args []string)           // Main learn command
func collectSourceFiles()               // Gather files to analyze
func detectPrimaryLanguage(files)       // JS, Python, Go, etc.
func detectFramework(files)             // React, FastAPI, etc.
func extractNamingPatterns(files)       // camelCase, snake_case, etc.
func extractImportPatterns(files)       // absolute, relative, prefixes
func extractStructurePatterns(root)     // folders, test patterns
func extractCodeStylePatterns(files)    // indent, quotes, semicolons
func savePatterns(patterns)             // .sentinel/patterns.json
func generateRulesFromPatterns(patterns) // .cursor/rules/project-patterns.md
```

### Detected Patterns

| Pattern Type | Detection |
|--------------|-----------|
| Language | JS, TS, Python, Go, Shell, etc. |
| Framework | React, Next.js, FastAPI, Django, Gin, etc. |
| Functions | camelCase, snake_case, PascalCase |
| Variables | camelCase, snake_case |
| Classes | PascalCase |
| Constants | SCREAMING_SNAKE_CASE |
| Files | kebab-case, camelCase, snake_case |
| Imports | absolute, relative, prefixes (@/, ~/) |
| Structure | src/, components/, services/, utils/ |
| Code Style | indent, quotes, semicolons |

### Command Usage

```bash
# Full learning
./sentinel learn

# Specific patterns
./sentinel learn --naming      # Naming only
./sentinel learn --imports     # Imports only  
./sentinel learn --structure   # Structure only

# Output options
./sentinel learn --output json # JSON output
./sentinel learn --no-rules    # Skip rule generation
```

### Exit Criteria

- ✅ `sentinel learn` works on any project
- ✅ Patterns correctly detected in test fixtures (16/16 tests pass)
- ✅ Generated rules valid for Cursor
- ✅ Confidence scores indicate detection reliability

---

## Phase 2: Safe Auto-Fix (Week 4) ✅ COMPLETED

**Goal**: Automatically fix safe issues.

**Status**: Completed on 2024-12-04

### Tasks

| Task | Days | Status |
|------|------|--------|
| Fix type definitions | 0.5 | ✅ Done |
| Safe fix implementations | 1.5 | ✅ Done |
| Backup system | 1 | ✅ Done |
| Fix application engine | 1 | ✅ Done |
| Dry-run mode | 0.5 | ✅ Done |
| Prompted fixes | 1 | ✅ Done |
| Rollback capability | 0.5 | ✅ Done |
| Tests | 1 | ✅ Done (8 tests) |

### Implemented Fixes

| Fix | Level | Languages |
|-----|-------|-----------|
| Remove console.log | safe | JS/TS |
| Remove console.debug | safe | JS/TS |
| Remove debugger | safe | JS/TS |
| Remove trailing whitespace | safe | All |
| Add EOF newline | safe | All |
| Remove print() debug | prompted | Python |
| Quote shell variables | prompted | Shell |

### Command Usage

```bash
# Interactive mode (prompts for risky fixes)
./sentinel fix

# Safe fixes only (no prompts)
./sentinel fix --safe

# Preview without changes
./sentinel fix --dry-run

# Auto-approve all
./sentinel fix --yes

# Specific pattern only
./sentinel fix --pattern "console.log"

# Rollback last fix session
./sentinel fix rollback
```

### Features

- **Backup System**: Creates timestamped backups before any changes
- **Dry-Run Mode**: Preview all fixes without modifying files
- **Interactive Prompts**: Asks for confirmation on risky fixes
- **Fix History**: Tracks all fix sessions in `.sentinel/fix-history.json`
- **Rollback**: Restore files from last backup

### Exit Criteria

- ✅ Safe fixes don't break code (tested)
- ✅ Backup always created before changes
- ✅ Rollback restores original state
- ✅ 8/8 tests passing

---

## Phase 3: Document Ingestion - Local Parsing (Week 5) ✅ COMPLETED

**Goal**: Parse multiple document formats locally (fallback/offline mode).

**Status**: Completed on 2024-12-04

> **Note**: This phase implemented local parsing as a foundation. Based on dependency
> management concerns (each developer needs poppler, tesseract), the primary workflow
> has been redesigned to use **server-side processing** (Phase 3B). Local parsing
> remains as offline fallback. See [Architecture Decision](./ARCHITECTURE_DOCUMENT_PROCESSING.md).

### Tasks

| Task | Days | Status |
|------|------|--------|
| Document types | 0.5 | ✅ Done |
| Text/Markdown parser | 0.5 | ✅ Done |
| PDF parser | 0.5 | ✅ Done |
| Word (.docx) parser | 0.5 | ✅ Done |
| Excel (.xlsx) parser | 0.5 | ✅ Done |
| Email (.eml) parser | 0.5 | ✅ Done |
| Image OCR | 0.5 | ✅ Done |
| Ingest command | 1 | ✅ Done |
| Tests | 0.5 | ✅ Done (10 tests) |

### Implementation Details

| Format | Parser | Dependencies |
|--------|--------|--------------|
| Text (.txt, .md) | Go native | None |
| PDF (.pdf) | pdftotext | poppler-utils |
| Word (.docx) | archive/zip + XML | None (Go stdlib) |
| Excel (.xlsx) | archive/zip + XML | None (Go stdlib) |
| Email (.eml) | net/mail | None (Go stdlib) |
| Images | tesseract | tesseract-ocr (optional) |

### Command Usage

```bash
# Ingest single document
./sentinel ingest /path/to/document.pdf

# Ingest directory
./sentinel ingest /path/to/docs/

# Skip images (no OCR)
./sentinel ingest /path/to/docs/ --skip-images

# Verbose output
./sentinel ingest /path/to/docs/ --verbose

# List ingested documents
./sentinel ingest --list
```

### Output Structure

```
docs/knowledge/
├── source-documents/       # Original uploads
│   ├── Scope_v2.pdf
│   ├── Requirements.docx
│   └── manifest.json       # Tracks all ingested docs
└── extracted/              # Parsed text content
    ├── Scope_v2.txt
    └── Requirements.txt
```

### Exit Criteria

- ✅ All supported formats parse correctly
- ✅ Text extraction working for PDF, DOCX, XLSX, EML
- ✅ Image OCR functional (requires tesseract)
- ✅ 10/10 tests passing

---

## Phase 3B: Hub Document Service (Week 6) ✅ COMPLETED

**Goal**: Server-side document processing to eliminate client-side dependencies.

**Status**: Completed on 2024-12-04

> **Why Server-Side?**: Each developer installing poppler/tesseract is impractical.
> Processing on Hub means zero dependencies on developer machines.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Hub directory structure | 0.5 | ✅ Done |
| Docker setup (poppler, tesseract) | 0.5 | ✅ Done |
| Database schema | 0.5 | ✅ Done |
| Hub API server (Go) | 1.5 | ✅ Done |
| Document processor worker | 1 | ✅ Done |
| Agent upload command | 0.5 | ✅ Done |
| Agent sync command | 0.5 | ✅ Done |
| Agent offline-info | 0.5 | ✅ Done |
| docker-compose.yml | 0.5 | ✅ Done |
| Hub README | 0.5 | ✅ Done |

### Deliverables

```
hub/
├── api/
│   ├── main.go           # API server
│   ├── Dockerfile
│   └── go.mod
├── processor/
│   ├── main.go           # Document worker
│   ├── Dockerfile        # With poppler, tesseract
│   └── go.mod
├── docker-compose.yml
├── scripts/
│   └── setup.sh          # One-command setup
└── README.md
```

### API Endpoints (Implemented)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/documents/ingest` | Upload document |
| GET | `/api/v1/documents/{id}/status` | Check status |
| GET | `/api/v1/documents/{id}/extracted` | Get text |
| GET | `/api/v1/documents/{id}/knowledge` | Get items |
| GET | `/api/v1/documents` | List documents |
| POST | `/api/v1/admin/organizations` | Create org |
| POST | `/api/v1/admin/projects` | Create project |

### Agent Commands (Implemented)

```bash
# Upload to Hub (default when configured)
./sentinel ingest /path/to/doc.pdf

# Check processing status
./sentinel ingest --status

# Sync results to local
./sentinel ingest --sync

# Show offline capabilities
./sentinel ingest --offline-info

# Force offline mode
./sentinel ingest /path/to/doc.txt --offline
```

### Exit Criteria

- ✅ Hub API server implemented
- ✅ Document processor with all dependencies
- ✅ Agent upload/sync commands working
- ✅ Offline fallback for basic formats
- ✅ docker-compose ready for deployment

---

## Phase 4: LLM Knowledge Extraction (Week 9) ✅ COMPLETED

**Goal**: Extract structured knowledge from documents using LLM.

**Status**: Completed on 2024-12-04, Enhanced with Azure AI Foundry on 2024-12-XX

### Tasks

| Task | Days | Status |
|------|------|--------|
| LLM types & interfaces | 0.5 | ✅ Done |
| Ollama integration | 1 | ✅ Done |
| Azure AI Foundry integration | 2 | ✅ Done |
| Provider abstraction & fallback | 1 | ✅ Done |
| Extraction prompts | 1 | ✅ Done |
| Knowledge schema | 0.5 | ✅ Done |
| Confidence scoring | 0.5 | ✅ Done |
| Knowledge review CLI | 1 | ✅ Done |
| Knowledge activation | 0.5 | ✅ Done |
| Knowledge sync (Agent ↔ Hub) | 1 | ✅ Done |
| Tests | 1 | ✅ Done (13 tests + Azure tests) |

### Knowledge Types Implemented

| Type | Description | Example |
|------|-------------|---------|
| `business_rule` | Conditional logic | "Orders cancelled within 24h" |
| `entity` | Domain objects | "User: id, email, role" |
| `glossary` | Term definitions | "SKU: Stock Keeping Unit" |
| `journey` | User workflows | "Checkout: cart → payment" |

### Knowledge Schema

```json
{
  "id": "ki_a1b2c3d4",
  "type": "business_rule",
  "title": "Order Cancellation Policy",
  "content": "Orders can only be cancelled within 24 hours...",
  "source": "requirements.pdf",
  "confidence": 0.92,
  "status": "pending",
  "approvedBy": null,
  "approvedAt": null,
  "createdAt": "2024-12-04T12:00:00Z"
}
```

### Commands Implemented

```bash
# List knowledge items
./sentinel knowledge list
./sentinel knowledge list --pending
./sentinel knowledge list --approved

# Statistics
./sentinel knowledge stats

# Interactive review
./sentinel knowledge review

# Approve/Reject items
./sentinel knowledge approve ki_001
./sentinel knowledge approve --all    # Auto-approve ≥90% confidence
./sentinel knowledge reject ki_002

# Generate Cursor rules from approved knowledge
./sentinel knowledge activate

# Extract from document (requires LLM)
./sentinel knowledge extract document.txt
```

### LLM Integration

| Provider | Location | Use Case | Status |
|----------|----------|----------|--------|
| Azure AI Foundry (Claude Opus 4.5) | Hub | Enterprise-grade, highest quality | ✅ Implemented |
| Ollama | Local or Hub | Self-hosted, privacy-focused, fallback | ✅ Implemented |

**Provider Fallback**: System automatically falls back from Azure → Ollama if Azure is unavailable.

### Human Review Workflow

```
1. Ingest documents     → sentinel ingest /docs/
2. Process on Hub       → Automatic LLM extraction
3. Sync to local        → sentinel ingest --sync
4. Review items         → sentinel knowledge review
5. Auto-approve high    → sentinel knowledge approve --all
6. Activate to Cursor   → sentinel knowledge activate
```

### Generated Cursor Rule

```markdown
---
description: Project Business Knowledge (Auto-Generated)
globs: ["**/*"]
alwaysApply: true
---

# Business Knowledge

## Business Rules
### Order Cancellation Policy
Orders can only be cancelled within 24 hours...

## Domain Entities
### User
Represents a registered customer with attributes...

## Glossary
| Term | Definition |
|------|------------|
| **SKU** | Stock Keeping Unit... |
```

### Exit Criteria

- ✅ LLM extraction works with Azure AI Foundry (Claude Opus 4.5) and Ollama
- ✅ Provider abstraction with automatic fallback (Azure → Ollama)
- ✅ Knowledge items have confidence scores (0.0-1.0)
- ✅ Human review workflow functional (review, approve, reject)
- ✅ Auto-approve for high confidence (≥90%)
- ✅ Approved knowledge generates Cursor rules
- ✅ Bidirectional knowledge sync (Agent ↔ Hub)
- ✅ 13/13 tests passing + Azure integration tests

---

## Phase 5B: Telemetry Client (Week 7-8) ✅ COMPLETED

**Goal**: Agent sends metrics to Hub (built alongside Hub MVP).

**Status**: Completed on 2024-12-04

### Tasks

| Task | Days | Status |
|------|------|--------|
| Telemetry protocol | 0.5 | ✅ Done |
| Telemetry client in Agent | 1 | ✅ Done |
| Data sanitization | 0.5 | ✅ Done |
| Offline queue | 0.5 | ✅ Done |
| Integration points | 0.5 | ✅ Done |

### Telemetry Events

| Event | Data Sent | Data NOT Sent |
|-------|-----------|---------------|
| audit_complete | Finding counts, compliance % | Code content |
| fix_applied | Fix counts by type | File contents |
| pattern_learned | Confidence scores | Actual patterns |
| doc_ingested | Item counts | Document text |

### Implemented Features

**Agent Telemetry Client**:
- `TelemetryClient` with queue management
- Automatic telemetry on `audit`, `fix`, and `learn` commands
- Offline queue in `.sentinel/telemetry-queue.json`
- Automatic flush when Hub available
- Client-side payload sanitization

**Integration Points**:
- `runAudit()` → `sendAuditTelemetry()`
- `runFix()` → `sendFixTelemetry()`
- `runLearn()` → `sendPatternTelemetry()`

### Exit Criteria

- ✅ Metrics sent to Hub successfully
- ✅ No sensitive data in payloads (client + server sanitization)
- ✅ Offline queue works when Hub unreachable

---

## Phase 5: Sentinel Hub MVP (Week 7-8) ✅ COMPLETED

**Goal**: Central server for metrics, document processing, and organization management.

**Status**: Completed on 2024-12-04

> **Note**: Hub now includes document processing service (Phase 3B merged).

### Tasks

| Task | Days | Status |
|------|------|--------|
| API server setup | 0.5 | ✅ Done |
| Database schema | 1 | ✅ Done |
| Authentication (API keys) | 1 | ✅ Done |
| Telemetry ingestion | 1 | ✅ Done |
| Document processing service | 2 | ✅ Done |
| Metrics query API | 1 | ✅ Done |
| Org/Project management | 1 | ✅ Done |
| Dashboard: Overview | 1.5 | ⏸️ Deferred (Frontend) |
| Dashboard: Documents | 1 | ⏸️ Deferred (Frontend) |
| Dashboard: Trends | 1 | ⏸️ Deferred (Frontend) |
| Docker deployment | 0.5 | ✅ Done |
| Tests | 1.5 | ✅ Done (8 telemetry tests) |

### Database Schema

```sql
-- Organizations
CREATE TABLE organizations (
  id UUID PRIMARY KEY,
  name VARCHAR(255),
  created_at TIMESTAMP
);

-- Projects
CREATE TABLE projects (
  id UUID PRIMARY KEY,
  org_id UUID REFERENCES organizations(id),
  name VARCHAR(255),
  api_key VARCHAR(64) UNIQUE
);

-- Documents
CREATE TABLE documents (
  id UUID PRIMARY KEY,
  project_id UUID REFERENCES projects(id),
  name VARCHAR(255),
  status VARCHAR(20),  -- queued, processing, completed, failed
  file_path VARCHAR(500),
  extracted_text TEXT,
  created_at TIMESTAMP,
  processed_at TIMESTAMP
);

-- Knowledge Items
CREATE TABLE knowledge_items (
  id UUID PRIMARY KEY,
  document_id UUID REFERENCES documents(id),
  type VARCHAR(50),  -- business_rule, entity, glossary, journey
  title VARCHAR(255),
  content TEXT,
  confidence FLOAT,
  status VARCHAR(20),  -- pending, approved, rejected
  approved_by VARCHAR(100),
  approved_at TIMESTAMP
);

-- Telemetry Events
CREATE TABLE telemetry_events (
  id UUID PRIMARY KEY,
  project_id UUID REFERENCES projects(id),
  event_type VARCHAR(50),
  payload JSONB,
  created_at TIMESTAMP
);
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| API | Go + Chi router |
| Database | PostgreSQL |
| Job Queue | Go channels + worker pool |
| Document Processing | poppler, tesseract (Docker) |
| LLM | Ollama (self-hosted) or OpenAI |
| Dashboard | React + TypeScript |
| Charts | Recharts |
| Styling | Tailwind CSS |
| Deployment | Docker Compose |

### Docker Compose

```yaml
services:
  hub:
    build: ./hub
    ports: ["8080:8080"]
    depends_on: [db, ollama]
    
  db:
    image: postgres:15-alpine
    
  ollama:
    image: ollama/ollama:latest
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              capabilities: [gpu]
```

### Implemented Features

**Telemetry Ingestion** (`POST /api/v1/telemetry`):
- Accepts batch telemetry events
- Validates event types (audit_complete, fix_applied, pattern_learned, doc_ingested)
- Sanitizes payloads (removes code content, only allows metrics)
- Stores in `telemetry_events` table

**Metrics Query API** (`GET /api/v1/metrics`):
- Query telemetry events by date range and event type
- Aggregated metrics calculation:
  - Total events, audit count, fix count
  - Average compliance percentage
  - Total findings (critical, warning, info)
  - Pattern and document counts
- Returns both raw events and aggregated metrics

**Security**:
- Payload sanitization ensures no code content is stored
- Only allowed fields are accepted (finding_count, compliance_percent, etc.)
- API key authentication required

### Exit Criteria

- ✅ Agents connect and authenticate
- ✅ Documents upload and process
- ✅ Telemetry ingestion working
- ✅ Metrics query API functional
- ✅ Multiple orgs isolated
- ✅ API keys scoped per project
- ⏸️ Dashboard (deferred to frontend phase)

---

## Phase 6: AST Analysis Engine (Week 10-11) ✅ COMPLETE

**Goal**: Server-side code analysis using Tree-sitter for vibe coding detection.

> **Critical**: This phase MUST be completed BEFORE Phase 7. AST is PRIMARY detection method. Pattern-based is FALLBACK only.

### Architecture: AST-First Detection

```
Detection Flow:
1. PRIMARY: Hub AST analysis (when available)
2. FALLBACK: Pattern matching (only if Hub unavailable)
3. Deduplication: AST findings take precedence
```

### Tasks (Reordered by Priority)

#### Phase 6A: Core AST Infrastructure (MUST COMPLETE FIRST)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Tree-sitter integration in Hub | 2 | ✅ Done | P0 |
| Language parser initialization (Go, JS, TS, Python) | 1 | ✅ Done | P0 |
| Hub AST API endpoint `/api/v1/analyze/ast` | 1 | ✅ Done | P0 |
| Hub Vibe API endpoint `/api/v1/analyze/vibe` | 1 | ✅ Done | P0 |
| Error handling and response formatting | 0.5 | ✅ Done | P0 |
| **Subtotal** | **5.5 days** | ✅ COMPLETE | |

#### Phase 6B: Core AST Detection Algorithms (REQUIRED FOR PHASE 7)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Duplicate function detection (AST-based) | 1 | ✅ Done | P0 |
| Orphaned code detection (AST scope analysis) | 1 | ✅ Done | P0 |
| Unused variable detection (AST symbol tracking) | 1 | ✅ Done | P0 |
| Signature mismatch detection (cross-file AST) | 1.5 | ⏳ Pending (Phase 6F) | P0 |
| Control flow analysis (unreachable code) | 1 | ✅ Done | P1 |
| **Subtotal** | **5.5 days** | 🟡 73% COMPLETE (4/5.5 tasks) | |

> **Note**: Signature mismatch detection requires cross-file analysis, which is planned for Phase 6F.

#### Phase 6C: Agent-Hub Integration (REQUIRED FOR PHASE 7)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Agent `--deep` flag integration | 0.5 | ✅ Done | P0 |
| Code collection and batching for Hub | 0.5 | ✅ Done | P0 |
| HTTP client for Hub communication | 0.5 | ✅ Done | P0 |
| AST response parsing in Agent | 0.5 | ✅ Done | P0 |
| Finding integration into audit report | 0.5 | ✅ Done | P0 |
| Error handling (Hub unavailable fallback) | 0.5 | ✅ Done | P0 |
| **ASTResult struct with Success/Error fields** | 0.5 | ✅ Done | P0 |
| **Timeout handling (10s for analysis, 10s for health)** | 0.5 | ✅ Done | P0 |
| **Retry logic for transient failures (2-3 retries)** | 0.5 | ✅ Done | P0 |
| **Batching logic for large codebases** | 1 | ✅ Done | P0 |
| **Subtotal** | **5.5 days** | ✅ COMPLETE | |

#### Phase 6E: Critical Reliability Fixes (REQUIRED BEFORE PHASE 7)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Fix AST failure detection (distinguish success vs failure) | 0.5 | ✅ Done | P0 |
| Fix fallback condition logic (only run patterns if AST failed) | 0.5 | ✅ Done | P0 |
| Improve health check timeout (2s → 10s) | 0.5 | ✅ Done | P0 |
| Add health check caching (60s TTL) | 0.5 | ✅ Done | P0 |
| Fix telemetry check consistency | 0.5 | ✅ Done | P0 |
| **Subtotal** | **2.5 days** | ✅ COMPLETE | |

#### Phase 6D: Performance & Polish (CAN BE DONE IN PARALLEL WITH PHASE 7)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| AST result caching (performance) | 1 | ✅ Done | P2 |
| TypeScript support (separate from JS) | 0.5 | ✅ Done | P2 |
| Test fixtures (vibe issue samples) | 0.5 | ✅ Done | P1 |
| Tests (unit + integration) | 1 | ✅ Done | P1 |
| **Subtotal** | **3 days** | ✅ COMPLETE | |

**Total Phase 6**: ~19.5 days (but can start Phase 7 after 6A+6B+6C+6E = 16.5 days)

> **Note**: Cross-file analysis (signature mismatch detection) is planned for Phase 6F and can be implemented after Phase 7.

### Hub API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/analyze/ast` | Single file AST analysis |
| POST | `/api/v1/analyze/vibe` | Vibe pattern detection |
| POST | `/api/v1/analyze/cross-file` | Multi-file analysis ✅ |

### Detection Algorithms

- **Duplicate Function Detection**: Same name, different implementations ✅
- **Orphaned Code Detection**: Unreachable functions ✅
- **Unused Variable Detection**: Declared but never used ✅
- **Cross-File Symbol Tracking**: Import/export analysis ✅

### Exit Criteria (MUST MEET BEFORE PHASE 7)

- ✅ AST analysis works for Go, JavaScript/TypeScript, Python
- ✅ Duplicate functions detected with 95% accuracy
- ✅ Orphaned code detected with 90% accuracy
- ✅ Unused variables detected with 90% accuracy
- ✅ Agent can send code to Hub and receive AST findings
- ✅ Fallback to pattern matching works when Hub unavailable
- ✅ Server-side analysis responds < 500ms
- ✅ AST failure detection distinguishes success vs failure
- ✅ Fallback logic only runs patterns if AST failed (not if AST succeeded with 0 findings)
- ✅ Health check timeout increased to 10s with caching
- ✅ Telemetry checks are consistent across functions
- ⚠️ Cross-file analysis (signature mismatches) deferred to Phase 6F

---

## Phase 7: Vibe Coding Detection (Week 12-13) ✅ COMPLETE (100%)

**Goal**: Complete vibe coding detection with AST-first architecture.

> **Dependencies**: 
> - Phase 6 MUST be complete (AST infrastructure required)
> - Pattern-based detection is FALLBACK only (not primary)

### Architecture: AST-First with Pattern Fallback

```
detectVibeIssues() flow:
1. Try Hub AST analysis (PRIMARY)
2. If Hub unavailable → Fallback to patterns
3. Deduplicate findings (AST takes precedence)
```

### Tasks (Reordered by Dependency)

#### Phase 7A: AST-First Integration (REQUIRES PHASE 6 COMPLETE)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Integrate AST findings from Hub (Phase 6) | 0.5 | ✅ Done | P0 |
| Implement AST-first detection flow | 0.5 | ✅ Done | P0 |
| **Fix AST failure vs success detection** | 0.5 | ✅ Done | P0 |
| **Fix fallback condition (only if AST failed)** | 0.5 | ✅ Done | P0 |
| Deduplication logic (AST vs pattern findings) | 0.5 | ✅ Done | P0 |
| **Improve deduplication (semantic matching)** | 0.5 | ✅ Done | P0 |
| Update `detectVibeIssues()` to use AST-first | 0.5 | ✅ Done | P0 |
| **Subtotal** | **3.5 days** | ✅ COMPLETE | |

#### Phase 7B: Pattern Fallback (FALLBACK ONLY)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Pattern: Empty catch/except blocks | 0.5 | ✅ Done | P1 |
| Pattern: Code after return | 0.5 | ✅ Done | P1 |
| Pattern: Missing await | 0.5 | ✅ Done | P1 |
| Pattern: Brace/bracket mismatch | 0.5 | ⏳ Pending (low priority) | P1 |
| Pattern: Basic duplicate detection (fallback) | 0.5 | ✅ Done | P1 |
| **Subtotal** | **2.5 days** | | |

> **Note**: Pattern detection is ONLY used when Hub unavailable. AST findings take precedence.

#### Phase 7C: Additional AST Detections (EXTENDS PHASE 6)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Empty catch/except blocks (AST-based) | 0.5 | ✅ Done | P1 |
| Code after return (AST control flow) | 0.5 | ✅ Done | P1 |
| Missing await (AST async tracking) | 0.5 | ✅ Done | P1 |
| Brace/bracket mismatch (AST parser errors) | 0.5 | ✅ Done | P1 |
| **Subtotal** | **2 days** | **✅ COMPLETE** | |

#### Phase 7D: Testing & Validation

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Test fixtures (known vibe issues) | 0.5 | ✅ Done | P1 |
| Accuracy measurement (85%+ target) | 0.5 | ✅ Done | P1 |
| AST vs pattern comparison tests | 0.5 | ✅ Done | P1 |
| Fallback behavior tests | 0.5 | ✅ Done | P1 |
| **Test AST success with 0 findings (should NOT run patterns)** | 0.5 | ✅ Done | P1 |
| **Test AST failure (should run patterns)** | 0.5 | ✅ Done | P1 |
| **Test semantic deduplication** | 0.5 | ✅ Done | P1 |
| Tests | 1 | ✅ Done | P1 |
| **Subtotal** | **4.5 days** | | |

#### Phase 7E: Real-World Reliability & UX

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Add `--offline` flag (force pattern-only mode) | 0.5 | ✅ Done | P2 |
| Add progress indicators for Hub analysis | 0.5 | ✅ Done | P2 |
| Add cancellation support (Ctrl+C handling) | 0.5 | ✅ Done | P2 |
| Add metrics tracking (AST vs pattern usage) | 0.5 | ✅ Done | P2 |
| Add error reporting (Hub failures visible to user) | 0.5 | ✅ Done | P1 |
| **Subtotal** | **2.5 days** | | |

**Total Phase 7**: ~14.5 days

### Commands

```bash
sentinel audit --vibe-check       # Include vibe coding issues
sentinel audit --vibe-only        # Only vibe coding issues
sentinel audit --deep             # Server-side AST analysis
```

### Detection Categories

| Category | Primary Method | Fallback Method | Coverage Target |
|----------|---------------|-----------------|-----------------|
| Structural issues | AST (Hub) | Pattern | 95% |
| Refactoring issues | Cross-file AST | None | 95% |
| Variable/scope issues | AST scope analysis | Pattern (limited) | 85% |
| Control flow issues | AST CFG | Pattern | 85% |

### Exit Criteria

- ✅ AST-first detection works (Hub available)
- ✅ Pattern fallback works (Hub unavailable)
- ✅ Deduplication prevents duplicate findings (line-based and semantic)
- ✅ AST success with 0 findings does NOT trigger pattern fallback
- ✅ AST failure properly triggers pattern fallback
- ✅ Vibe issues detected with 85%+ accuracy
- ✅ Findings integrated into audit report
- ✅ `--vibe-check` flag works correctly
- ✅ `--offline` flag forces pattern-only mode
- ✅ Progress indicators show Hub analysis status
- ✅ Error reporting makes Hub failures visible
- ✅ Cancellation support (Ctrl+C) implemented
- ✅ Metrics tracking (AST vs pattern usage) implemented
- ✅ Comprehensive test suite (accuracy, comparison, fallback, deduplication)
- ✅ Empty catch/except blocks detection (AST-based)
- ✅ Enhanced code after return/throw/raise detection
- ✅ Missing await detection for async functions
- ✅ Brace/bracket mismatch detection from parser errors

---

## Phase 6F: Cross-File Analysis Implementation (Optional Enhancement)

**Goal**: Implement functional cross-file AST analysis for signature mismatch detection.

> **Dependencies**: Phase 6A, 6B, 6C must be complete. Can be implemented after Phase 7.

**Status**: ✅ COMPLETE

### Tasks

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Symbol table building (collect imports/exports across files) | 1 | ✅ Done | P1 |
| Cross-file reference resolution | 1 | ✅ Done | P1 |
| Signature mismatch detection (compare function signatures) | 1.5 | ✅ Done | P1 |
| Import/export mismatch detection | 1 | ✅ Done | P1 |
| Integration with vibe detection flow | 0.5 | ✅ Done | P1 |
| Tests for cross-file scenarios | 1 | ✅ Done | P1 |
| **Subtotal** | **6 days** | ✅ COMPLETE | |

### Implementation Details

**Symbol Table Building**:
- Collect all function/class exports from project files
- Build import dependency graph
- Track symbol definitions and usages across files

**Cross-File Reference Resolution**:
- Resolve imports to actual definitions
- Track symbol visibility (public/private)
- Handle namespace/module boundaries

**Signature Mismatch Detection**:
- Compare function signatures across files
- Detect parameter count/type mismatches
- Detect return type mismatches
- Report call sites with incorrect signatures

**Integration**:
- Add cross-file findings to vibe detection results
- Ensure deduplication with single-file findings
- Add to audit report with appropriate severity

### Exit Criteria

- ✅ Symbol table built from project files
- ✅ Cross-file references resolved correctly
- ✅ Signature mismatches detected with 90%+ accuracy
- ✅ Import/export mismatches detected
- ✅ Findings integrated into audit reports
- ✅ Tests validate cross-file scenarios

### Implementation Notes

**Hub Implementation** (`hub/api/ast_analyzer.go`):
- `buildSymbolTable()`: Collects symbols from multiple files
- `extractSymbols()`: Extracts function/class definitions using Tree-sitter
- `extractImportsExports()`: Extracts import/export statements
- `resolveCrossFileReferences()`: Maps imports to definitions
- `detectSignatureMismatches()`: Compares function signatures across files
- `detectImportExportMismatches()`: Detects missing exports/imports
- `analyzeCrossFile()`: Main cross-file analysis function

**Agent Integration** (`synapsevibsentinel.sh`):
- `sendCrossFileAnalysis()`: Sends multiple files to Hub's `/api/v1/analyze/cross-file` endpoint
- `sendBatchToHub()`: Automatically uses cross-file analysis when batch size > 1
- Integrated with existing vibe detection flow

**Testing**:
- `tests/unit/cross_file_analysis_test.sh`: Test suite for cross-file scenarios
- Tests for JavaScript, TypeScript, Go, and Python signature mismatches
- Tests for import/export mismatch detection

---

## Phase 8: Security Rules System (Week 13-14) ✅ COMPLETE (100%)

**Goal**: Implement SEC-001 through SEC-008 with AST-based enforcement.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Security rule schema | 0.5 | ✅ Done |
| Framework detection (Express, FastAPI, Gin, etc.) | 1 | ✅ Done |
| SEC-001: Resource Ownership (AST ownership check) | 0.5 | ✅ Done |
| SEC-002: SQL Injection (Pattern + AST) | 0.5 | ✅ Done |
| SEC-003: Auth Middleware (Route analysis) | 0.5 | ✅ Done |
| SEC-004: Rate Limiting (Endpoint analysis) | 0.5 | ✅ Done |
| SEC-005: Password Hashing (Data flow analysis) | 1 | ✅ Done |
| SEC-006: Input Validation (Handler analysis) | 0.5 | ✅ Done |
| SEC-007: Secure Headers (Middleware check) | 0.5 | ✅ Done |
| SEC-008: CORS Config (Config analysis) | 0.5 | ✅ Done |
| AST-based security checks | 2 | ✅ Done |
| Route/middleware analysis | 1 | ✅ Done |
| Data flow analysis (for SEC-005) | 1 | ✅ Done |
| Security scoring algorithm | 0.5 | ✅ Done |
| Hub security API endpoint | 1 | ✅ Done |
| Agent `--security` flag | 0.5 | ✅ Done |
| Agent `--security-rules` command | 0.5 | ✅ Done |
| Agent-Hub integration (call security endpoint) | 0.5 | ✅ Done |
| Security test fixtures (vulnerable code samples) | 1 | ✅ Done |
| Security analysis caching | 0.5 | ✅ Done |
| Error handling improvements | 0.5 | ✅ Done |
| Progress indicators | 0.5 | ✅ Done |
| Detection rate validation | 1 | ✅ Done |
| Tests | 1 | ✅ Done |
| **Total** | **~16 days** | **✅ COMPLETE (100%)** |

> **Note**: Framework detection is required before route/middleware analysis. SEC-005 requires data flow analysis (Phase 6 dependency).

### Built-in Rules

| ID | Name | Severity | Detection |
|----|------|----------|-----------|
| SEC-001 | Resource Ownership | Critical | AST ownership check |
| SEC-002 | SQL Injection | Critical | Pattern + AST |
| SEC-003 | Auth Middleware | Critical | Route analysis |
| SEC-004 | Rate Limiting | High | Endpoint analysis |
| SEC-005 | Password Hashing | Critical | Data flow |
| SEC-006 | Input Validation | High | Handler analysis |
| SEC-007 | Secure Headers | Medium | Middleware check |
| SEC-008 | CORS Config | High | Config analysis |

### Hub API Endpoint

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/analyze/security` | Security analysis with scoring |

### Exit Criteria

- 85% detection rate for security issues (validated using ground truth test suite)
- Security score calculated per project
- `--security` flag integrated into audit
- Detection rate metrics available when ground truth provided
- Ground truth test suite located at `tests/fixtures/security/ground_truth/`
- Validation procedure:
  1. Run security analysis with `expectedFindings` in request
  2. Compare detected findings against ground truth labels
  3. Calculate metrics: detection rate, precision, recall
  4. Verify detection rate meets 85% threshold
  5. Metrics included in `SecurityAnalysisResponse.Metrics` field

---

## Phase 9: File Size Management (Week 15) ✅ COMPLETE

**Goal**: Prevent monolithic files that cause context overflow.

**Current Status**: ✅ COMPLETE - All core functionality implemented and tested.

**Implementation Status**:
- ✅ Configuration structure (`FileSizeConfig` struct, default thresholds, config merging)
- ✅ File size checking integrated into audit process
- ✅ `--analyze-structure` flag implemented
- ✅ Hub architecture analysis endpoint implemented
- ✅ Section detection (AST-first with pattern fallback)
- ✅ Split suggestions with migration instructions
- ✅ Agent-Hub integration
- ✅ Tests implemented

**Note**: Phase 9 provides suggestions and migration instructions only. File splitting execution is deferred to Phase 9B (future phase).

### Evidence

This project's `synapsevibsentinel.sh` at 8,489 lines demonstrates the problem.

### Tasks

| Task | Days | Status |
|------|------|--------|
| File size thresholds config | 0.5 | ✅ Done |
| Config integration into audit (`runAudit()`) | 0.5 | ✅ Done |
| File size checking during scan | 0.5 | ✅ Done |
| File size monitoring | 0.5 | ✅ Done |
| Architecture analysis (Hub) | 2 | ✅ Done |
| Section detection: Function boundaries | 0.5 | ✅ Done |
| Section detection: Class/module boundaries | 0.5 | ✅ Done |
| Section detection: Language-specific parsing | 0.5 | ✅ Done |
| Dependency analysis between sections | 0.5 | ✅ Done |
| Split point suggestion algorithm | 1 | ✅ Done |
| Split suggestions | 1 | ✅ Done |
| Agent `--analyze-structure` flag | 0.5 | ✅ Done |
| Agent-Hub integration (call architecture endpoint) | 0.5 | ✅ Done |
| Proactive warnings in audit output | 0.5 | ✅ Done |
| MCP tool: `sentinel_check_file_size` (Phase 14) | - | See Phase 14 |
| Tests | 0.5 | ✅ Done |
| **Total** | **~10 days** | **✅ COMPLETE** |

> **Note**: Section detection uses Phase 6 AST analysis. All functionality integrated and tested.

### Configuration

```json
{
  "fileSize": {
    "thresholds": {
      "warning": 300,
      "critical": 500,
      "maximum": 1000
    },
    "byFileType": {},
    "exceptions": []
  }
}
```

### Exit Criteria ✅ ALL MET

- ✅ Oversized files flagged with suggested splits
- ✅ `--analyze-structure` provides architectural insights
- ✅ Warnings appear in audit output
- ✅ Hub architecture analysis endpoint functional
- ✅ Agent-Hub integration working
- ✅ Tests implemented and passing

---

## Phase 10: Test Enforcement System (Week 16-17) 🔴 MOVED FROM 14

**Goal**: Ensure business rules have corresponding tests.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Test requirement generation | 1.5 | ⏳ Pending |
| Test coverage tracking | 1 | ⏳ Pending |
| Test validation | 1 | ⏳ Pending |
| Mutation testing engine | 2 | ⏳ Pending |
| Test execution sandbox | 2 | ⏳ Pending |
| Hub test API endpoints | 1 | ⏳ Pending |
| Tests | 1 | ⏳ Pending |
| **Total** | **~9.5 days** | |

### Exit Criteria

- Tests generated from business rules
- Coverage tracked per rule
- Mutation score calculated
- Sandbox execution working

---

## Phase 11: Code-Documentation Comparison (Week 18) 🆕 NEW

**Goal**: Bidirectional validation between code and documentation.

> **Critical Enhancement**: This phase should include **implementation status tracking** to prevent documentation drift (as discovered in Phase 6/7 analysis). See [DOCUMENTATION_CODE_SYNC_ANALYSIS.md](./DOCUMENTATION_CODE_SYNC_ANALYSIS.md) for detailed proposal.

### Tasks

#### Phase 11A: Implementation Status Tracking (NEW - CRITICAL)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Status marker parser (extract from docs) | 1 | ⏳ Pending | P0 |
| Code implementation detector | 1.5 | ⏳ Pending | P0 |
| Status comparison engine | 1 | ⏳ Pending | P0 |
| Feature flag validator | 0.5 | ⏳ Pending | P0 |
| API endpoint validator | 0.5 | ⏳ Pending | P0 |
| Command validator | 0.5 | ⏳ Pending | P0 |
| Discrepancy report generator | 1 | ⏳ Pending | P0 |
| Integration into audit command | 0.5 | ⏳ Pending | P0 |
| Tests | 1 | ⏳ Pending | P1 |
| **Subtotal** | **~7.5 days** | | |

#### Phase 11B: Business Rules Comparison (ORIGINAL SCOPE)

| Task | Days | Status | Priority |
|------|------|--------|----------|
| Code behavior extraction | 2 | ⏳ Pending | P1 |
| Bidirectional comparison | 1.5 | ⏳ Pending | P1 |
| Discrepancy detection | 1 | ⏳ Pending | P1 |
| Human review workflow | 1 | ⏳ Pending | P1 |
| Hub API endpoints | 1 | ⏳ Pending | P1 |
| Tests | 0.5 | ⏳ Pending | P1 |
| **Subtotal** | **~7 days** | | |

**Total Phase 11**: ~14.5 days

### Gap Types

| Gap Type | Detection |
|----------|-----------|
| Implemented but not documented | Code scan vs rules |
| Documented but not implemented | Rules vs code |
| Partially implemented | Side effects check |
| Tests missing | Rule vs test mapping |

### Exit Criteria

- Gap analysis identifies discrepancies
- Comparison report generated
- Human review workflow functional

---

## Phase 12: Requirements Lifecycle (Week 19-20) 🔴 MOVED FROM 15

**Goal**: Track requirement changes and ensure code stays in sync.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Gap analysis command | 1.5 | ⏳ Pending |
| Change detection on ingest | 1.5 | ⏳ Pending |
| Change request workflow | 1.5 | ⏳ Pending |
| Impact analysis | 1 | ⏳ Pending |
| Implementation tracking | 1 | ⏳ Pending |
| Hub API endpoints | 1 | ⏳ Pending |
| Tests | 0.5 | ⏳ Pending |
| **Total** | **~8 days** | |

### Exit Criteria

- Gap analysis identifies discrepancies
- Change requests track modifications
- Impact analysis shows affected code

---

## Phase 13: Knowledge Schema Standardization (Week 21) 🔴 MOVED FROM 16

**Goal**: Standardize all knowledge extraction for consistent interpretation.

> **Reference**: See [KNOWLEDGE_SCHEMA.md](./KNOWLEDGE_SCHEMA.md) for complete schema.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Schema validation | 0.5 | ⏳ Pending |
| Enhanced extraction prompts | 1 | ⏳ Pending |
| Boundary specification | 0.5 | ⏳ Pending |
| Ambiguity handling | 1 | ⏳ Pending |
| Test case generation | 1 | ⏳ Pending |
| Migration of existing knowledge | 1 | ⏳ Pending |
| Tests | 0.5 | ⏳ Pending |
| **Total** | **~5.5 days** | |

### Exit Criteria

- All knowledge follows standard schema
- Ambiguities flagged for clarification
- Boundary behavior explicitly defined

---

## Phase 14A: Comprehensive Feature Analysis Foundation (Week 20-21) ⏳ PENDING

**Goal**: Implement core feature discovery and layer-specific analyzers for comprehensive end-to-end feature analysis.

> **Dependencies**: 
> - Phase 6 (AST Analysis) ✅ - Required for business logic analysis
> - Phase 8 (Security Rules) ✅ - Required for API layer security analysis
> - Phase 4 (Knowledge Base) ✅ - Required for business context validation

### Tasks

| Task | Days | Status |
|------|------|--------|
| Feature discovery algorithm (UI, API, Database, Logic, Integration, Tests) | 2 | ⏳ Pending |
| Business context analyzer (rules, journeys, entities) | 0.5 | ⏳ Pending |
| UI layer analyzer (components, forms, validation) | 0.5 | ⏳ Pending |
| API layer analyzer (endpoints, security, middleware) | 0.5 | ⏳ Pending |
| Database layer analyzer (schema, migrations, integrity) | 0.5 | ⏳ Pending |
| Business logic analyzer (AST, cross-file, semantic) | 0.5 | ⏳ Pending |
| Integration layer analyzer (external APIs, contracts) | 0.5 | ⏳ Pending |
| Test layer analyzer (coverage, quality, edge cases) | 0.5 | ⏳ Pending |
| End-to-end flow verification (flow detection, breakpoints) | 2 | ⏳ Pending |
| Hub LLM integration (API key management, model selection) | 2 | ⏳ Pending |
| Result aggregation (checklist generation, prioritization) | 1 | ⏳ Pending |
| Database schema (comprehensive_validations, analysis_configurations) | 1 | ⏳ Pending |
| API endpoints (POST /api/v1/analyze/comprehensive, GET /api/v1/validations/{id}) | 1 | ⏳ Pending |
| Tests | 1 | ⏳ Pending |
| Documentation | 0.5 | ⏳ Pending |
| **Total** | **~12 days** | |

### Exit Criteria

- Feature discovery works for auto and manual modes
- All 7 layer analyzers functional
- End-to-end flow verification working
- Hub LLM integration complete (API key management, model selection)
- Results stored in Hub with URL access
- API endpoints functional

**Reference**: See [COMPREHENSIVE_ANALYSIS_SOLUTION.md](./COMPREHENSIVE_ANALYSIS_SOLUTION.md) for complete specification.

---

## Phase 14B: MCP Integration for Comprehensive Analysis (Week 21) ⏳ PENDING

**Goal**: Integrate comprehensive analysis into Cursor via MCP tool.

> **Dependencies**: 
> - Phase 14A ✅ - Foundation must be complete

### Tasks

| Task | Days | Status |
|------|------|--------|
| MCP tool: sentinel_analyze_feature_comprehensive | 2 | ⏳ Pending |
| Agent integration (command handler, Hub communication) | 1 | ⏳ Pending |
| Error handling and fallback (Cursor default auto mode) | 0.5 | ⏳ Pending |
| Tests (unit, integration, end-to-end) | 1.5 | ⏳ Pending |
| **Total** | **~5 days** | |

### Exit Criteria

- MCP tool functional from Cursor
- Agent correctly communicates with Hub
- Fallback to Cursor default auto mode works
- All tests passing

---

## Phase 14C: Hub Configuration Interface (Week 21-22) ⏳ PENDING

**Goal**: Build Hub UI for LLM configuration and cost tracking.

> **Dependencies**: 
> - Phase 14A ✅ - Foundation must be complete

### Tasks

| Task | Days | Status |
|------|------|--------|
| Configuration UI (provider selection, API key input, model selection) | 2 | ⏳ Pending |
| Cost optimization settings (caching, progressive depth) | 0.5 | ⏳ Pending |
| Usage tracking dashboard (token usage, cost reports) | 2 | ⏳ Pending |
| Tests (UI, integration) | 1 | ⏳ Pending |
| **Total** | **~5.5 days** | |

### Exit Criteria

- Hub UI allows API key configuration (encrypted storage)
- Provider and model selection working
- Cost optimization settings configurable
- Usage tracking dashboard functional (reporting only, not billing)

---

## Phase 14D: Cost Optimization (Week 22) ⏳ PENDING

**Goal**: Implement advanced cost optimization features.

> **Dependencies**: 
> - Phase 14A ✅ - Foundation must be complete
> - Phase 14C ✅ - Configuration interface must be complete

### Tasks

| Task | Days | Status |
|------|------|--------|
| Caching system (result caching, business context caching, LLM response caching) | 2 | ⏳ Pending |
| Progressive depth (Level 1: fast checks, Level 2: medium-depth, Level 3: deep analysis) | 1.5 | ⏳ Pending |
| Smart model selection (task classification, model routing, cost tracking) | 1.5 | ⏳ Pending |
| **Total** | **~5 days** | |

### Exit Criteria

- 70% cache hit rate achieved
- Progressive depth working (skip LLM when possible)
- Smart model selection routing correctly
- 40% cost reduction via optimization

---

## Phase 14: MCP Integration (Week 22-23) 🔴 MOVED FROM 7

**Goal**: Real-time Cursor integration with all foundation features ready.

> **Critical Dependencies**: 
> - Phase 6 (AST Analysis) ✅ - Required for `sentinel_validate_code`
> - Phase 8 (Security Rules) ✅ - Required for `sentinel_get_security_context`, `sentinel_validate_security`
> - Phase 9 (File Size) - Required for `sentinel_check_file_size`
> - Phase 10 (Test Enforcement) - Required for test-related tools
> - Phase 15 (Intent) - Required for `sentinel_check_intent` (can be added later)
> - Phase 14A-14D ✅ - Comprehensive analysis foundation complete
> 
> **Note**: Most tools can be implemented with conditional availability based on phase completion. `sentinel_check_intent` depends on Phase 15 which comes after Phase 14, so this tool will be added in Phase 15 or Phase 14 should be moved after Phase 15.

### Tasks

| Task | Days | Status |
|------|------|--------|
| MCP protocol handler | 1 | ⏳ Pending |
| Tool definitions & schemas | 0.5 | ⏳ Pending |
| Tool registration system (dynamic discovery) | 0.5 | ⏳ Pending |
| Conditional tool availability (based on phases) | 0.5 | ⏳ Pending |
| sentinel_analyze_intent | 1 | ⏳ Pending |
| sentinel_get_context | 0.5 | ⏳ Pending |
| sentinel_get_patterns | 0.5 | ⏳ Pending |
| sentinel_check_intent | 1 | ⏳ Pending |
| sentinel_get_business_context | 0.5 | ⏳ Pending |
| sentinel_get_security_context | 0.5 | ⏳ Pending |
| sentinel_get_test_requirements | 0.5 | ⏳ Pending |
| sentinel_check_file_size | 0.5 | ⏳ Pending |
| sentinel_validate_code | 1 | ⏳ Pending |
| sentinel_validate_security | 0.5 | ⏳ Pending |
| sentinel_validate_business | 0.5 | ⏳ Pending |
| sentinel_validate_tests | 1 | ⏳ Pending |
| sentinel_apply_fix | 1 | ⏳ Pending |
| sentinel_generate_tests | 1 | ⏳ Pending |
| sentinel_run_tests | 1 | ⏳ Pending |
| Error handling (tool failures, fallbacks) | 0.5 | ⏳ Pending |
| MCP server mode | 0.5 | ⏳ Pending |
| Tests | 1 | ⏳ Pending |
| Documentation | 0.5 | ⏳ Pending |
| **Total** | **~14 days** | |

### MCP Tools (15 Total)

| Tool | Purpose | Dependencies | Phase |
|------|---------|--------------|-------|
| sentinel_analyze_feature_comprehensive | Comprehensive feature analysis across all layers | Phase 14A ✅ | Phase 14B |
| sentinel_analyze_intent | Understand request context | None | Phase 14 |
| sentinel_get_context | Recent activity, errors, git status | None | Phase 14 |
| sentinel_get_patterns | Project conventions | Phase 1 (Patterns) ✅ | Phase 14 |
| sentinel_check_intent | Clarify unclear requests | Phase 15 (Intent) ⚠️ | Phase 14* |
| sentinel_get_business_context | Business rules, entities | Phase 4 (Knowledge) ✅ | Phase 14 |
| sentinel_get_security_context | Security requirements | Phase 8 (Security) ⚠️ | Phase 14 |
| sentinel_get_test_requirements | Required tests | Phase 10 (Tests) ⚠️ | Phase 14 |
| sentinel_check_file_size | File size warnings | Phase 9 (File Size) ⚠️ | Phase 14 |
| sentinel_validate_code | Validate generated code | Phase 6 (AST) ⚠️ | Phase 14 |
| sentinel_validate_security | Security compliance | Phase 8 (Security) ⚠️ | Phase 14 |
| sentinel_validate_business | Business rule compliance | Phase 4 (Knowledge) ✅ | Phase 14 |
| sentinel_validate_tests | Test quality check | Phase 10 (Tests) ⚠️ | Phase 14 |
| sentinel_apply_fix | Fix issues in code | Phase 2 (Fixes) ✅ | Phase 14 |
| sentinel_generate_tests | Generate test cases | Phase 10 (Tests) ⚠️ | Phase 14 |
| sentinel_run_tests | Execute tests in sandbox | Phase 10 (Tests) ⚠️ | Phase 14 |

> **Note**: Tools marked with ⚠️ require their dependency phases to be complete. `sentinel_check_intent` depends on Phase 15 (Intent), which comes after Phase 14. This tool will be added in Phase 15 or Phase 14 can be moved after Phase 15.

### Exit Criteria

- Cursor can call Sentinel tools
- Validation works in real-time
- All tools functional (dependencies met)

---

## Phase 15: Intent & Simple Language (Week 24) 🔴 MOVED FROM 8

**Goal**: Handle unclear prompts gracefully.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Simple language templates | 0.5 | ⏳ Pending |
| Context gathering | 1 | ⏳ Pending |
| Intent analysis | 1 | ⏳ Pending |
| Clarifying questions | 1 | ⏳ Pending |
| Decision recording | 0.5 | ⏳ Pending |
| Pattern refinement | 1 | ⏳ Pending |
| Tests | 1 | ⏳ Pending |
| **Total** | **~6 days** | |

### Simple Language Templates

| Scenario | Template |
|----------|----------|
| Location unclear | "Where should this go?\n1. {opt1}\n2. {opt2}" |
| Entity unclear | "Which {entity}?\n1. {opt1}\n2. {opt2}" |
| Confirm action | "I will {action}. Correct? [Y/n]" |

### Exit Criteria

- Vague prompts handled gracefully
- Decisions recorded for learning
- Non-English speakers can use

---

## Phase 16: Organization Features (Week 25) 🔴 MOVED FROM 9

**Goal**: Team management, shared patterns.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Team management | 1 | ⏳ Pending |
| Pattern distribution | 1 | ⏳ Pending |
| Agent registration | 0.5 | ⏳ Pending |
| Dashboard: Team admin | 1 | ⏳ Pending |
| Dashboard: Pattern editor | 1 | ⏳ Pending |
| Alerting | 1 | ⏳ Pending |
| Tests | 0.5 | ⏳ Pending |
| **Total** | **~6 days** | |

### Features

| Feature | Description |
|---------|-------------|
| Teams | Create, edit, delete teams |
| Patterns | Push org patterns to agents |
| Agents | Track connected agents |
| Alerts | Notify on thresholds |
| Roles | Admin, Lead, Developer |

### Exit Criteria

- Teams manageable in dashboard
- Patterns distributed to agents
- Alerts working

---

## Phase 17: Dashboard Frontend (Week 26) ⏸️ DEFERRED

**Goal**: Web-based dashboard for organization management.

> **Note**: Frontend development deferred to focus on core Agent/Hub features first.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Dashboard: Overview | 1.5 | ⏸️ Deferred |
| Dashboard: Documents | 1 | ⏸️ Deferred |
| Dashboard: Trends | 1 | ⏸️ Deferred |
| Dashboard: Team admin | 1 | ⏸️ Deferred |
| Dashboard: Pattern editor | 1 | ⏸️ Deferred |
| **Total** | **~5.5 days** | |

---

## Phase 18: Hardening & Documentation (Week 27) 🔴 MOVED FROM 10

**Goal**: Production-ready release.

### Tasks

| Task | Days | Status |
|------|------|--------|
| Security audit | 1 | ⏳ Pending |
| Performance testing | 1 | ⏳ Pending |
| Error handling review | 0.5 | ⏳ Pending |
| Logging improvements | 0.5 | ⏳ Pending |
| User documentation | 1 | ⏳ Pending |
| Admin documentation | 0.5 | ⏳ Pending |
| API documentation | 0.5 | ⏳ Pending |
| Deployment guide | 0.5 | ⏳ Pending |
| Final QA | 0.5 | ⏳ Pending |
| **Total** | **~6 days** | |

### Documentation Deliverables

| Document | Audience |
|----------|----------|
| User Guide | Developers |
| Admin Guide | Organization admins |
| API Reference | Integrators |
| Deployment Guide | DevOps |

### Exit Criteria

- Security audit passed
- Performance acceptable
- Documentation complete
- Ready for production

---

**This phase has been MOVED TO Phase 7** (see above).

---

---

## Updated Timeline Overview (CORRECTED ORDER)

```
COMPLETED PHASES:
WEEK 1-2    Phase 0: Foundation & Testing              ✅ DONE
WEEK 3      Phase 1: Pattern Learning                  ✅ DONE
WEEK 4      Phase 2: Safe Auto-Fix                     ✅ DONE
WEEK 5      Phase 3: Document Ingestion (Local)        ✅ DONE
WEEK 6      Phase 3B: Hub Document Service             ✅ DONE
WEEK 7-8    Phase 5: Sentinel Hub MVP                  ✅ DONE
WEEK 9      Phase 4: LLM Knowledge Extraction          ✅ DONE
WEEK 9+     Azure AI Foundry Integration               ✅ DONE

FOUNDATION LAYER (Must complete before MCP):
WEEK 10-11  Phase 6: AST Analysis Engine                🔴 NEW - P0
WEEK 12     Phase 7: Vibe Coding Detection             🔴 MOVED - P0
WEEK 13-14  Phase 8: Security Rules System             🔴 MOVED - P0
WEEK 15     Phase 9: File Size Management              🔴 MOVED - P0

ENHANCEMENT LAYER:
WEEK 16-17  Phase 10: Test Enforcement System           🔴 MOVED - P1
WEEK 18     Phase 11: Code-Doc Comparison              🆕 NEW - P0 (Enhanced with Status Tracking)
WEEK 19-20  Phase 12: Requirements Lifecycle           🔴 MOVED - P1
WEEK 21     Phase 13: Knowledge Schema                 🔴 MOVED - P1

INTEGRATION LAYER (Now has dependencies):
WEEK 22-23  Phase 14: MCP Integration                  🔴 MOVED - P2
WEEK 24     Phase 15: Intent & Simple Language          🔴 MOVED - P2

BUSINESS LAYER:
WEEK 25     Phase 16: Organization Features             🔴 MOVED - P3
WEEK 26     Phase 17: Dashboard Frontend                ⏸️ DEFERRED
WEEK 27     Phase 18: Hardening & Documentation        🔴 MOVED - P3
```

> **Critical Change**: Phases reordered to fix dependency chain. MCP Integration (Phase 14) now comes AFTER all foundation features (AST, Security, File Size, Vibe Detection) are complete.

---

## Success Metrics (Updated)

| Phase | Key Metric |
|-------|------------|
| 0 | Test coverage >80% |
| 1 | Pattern accuracy >85% |
| 2 | Zero regressions from fixes |
| 3 | Document parsing >95% accuracy |
| 4 | Knowledge extraction with confidence scores |
| 5 | No code in telemetry |
| 6 | Dashboard loads <3s |
| 7 | MCP tools <500ms response |
| 8 | Vague prompts handled |
| 9 | Alerts delivered <1min |
| 10 | All audits passed |
| **11** | **Vibe issue detection >85%** |
| **12** | **Oversized files flagged 100%** |
| **13** | **Security detection >85%** |
| **14** | **Test coverage tracking >90%** |
| **15** | **Gap analysis accuracy >90%** |
| **16** | **Knowledge schema compliance 100%** |

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| MCP protocol changes | High | Abstract MCP layer |
| Pattern detection inaccurate | Medium | Confidence scores + override |
| Auto-fix breaks code | High | Backup always, tests |
| Document parsing fails | Medium | Graceful fallback, partial results |
| Telemetry leak | Critical | Security audit, sanitization |
| Hub scalability | Medium | Design for horizontal scale |
| Team adoption | Medium | Show value early |

---

## Resource Requirements

### Development Team

| Role | Count | Focus |
|------|-------|-------|
| Go Developer | 2 | Agent, Hub API |
| Frontend Developer | 1 | Dashboard |
| DevOps | 0.5 | CI/CD, Deployment |

### Infrastructure

| Component | Specification |
|-----------|---------------|
| Hub Server | 2 CPU, 4GB RAM |
| Database | PostgreSQL 14+, 50GB |
| CI/CD | GitHub Actions |

### External Services

| Service | Purpose | Cost |
|---------|---------|------|
| OpenAI API | Document extraction | ~$0.50/doc |
| Optional: Vision API | Image analysis | ~$0.05/image |
| Optional: Ollama | Local LLM | Free |

---

## Future SaaS Expansion Path

After organization deployment is stable:

| Feature | Effort | Priority |
|---------|--------|----------|
| Multi-tenancy | 2 weeks | P0 |
| Self-service signup | 2 weeks | P0 |
| Billing (Stripe) | 2 weeks | P0 |
| Public API | 2 weeks | P1 |
| Multi-region | 1 week | P2 |
| Support portal | 1 week | P2 |


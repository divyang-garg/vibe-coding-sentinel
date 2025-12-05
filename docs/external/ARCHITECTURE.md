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
│  │  │ • Query     │  │ • Orgs      │  │ • Teams     │            │    │
│  │  │ • Auth      │  │ • Teams     │  │ • Trends    │            │    │
│  │  │ • Patterns  │  │ • Patterns  │  │ • Admin     │            │    │
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
- Organization overview
- Team breakdown
- Trend analysis
- Documentation coverage
- Administration
- Comprehensive analysis results
- LLM configuration interface

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

### Git (via Hooks)

```
.git/hooks/
├── pre-commit    # Runs audit + safe fixes
├── pre-push      # Full audit
└── commit-msg    # Message validation
```

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


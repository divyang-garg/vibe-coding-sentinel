# User Guide

## Quick Start

### Installation

```bash
# Clone or download Sentinel
cd /your/project

# Build the binary
chmod +x synapsevibsentinel.sh
./synapsevibsentinel.sh

# Verify installation
./sentinel --help
```

---

## User Journeys

### Journey 1: New Project (Greenfield)

Starting a new project from scratch with best practices from day one.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    NEW PROJECT JOURNEY                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 0: GATHER PROJECT DOCUMENTS                                        │
│  ═══════════════════════════════                                        │
│                                                                          │
│  Collect all project documents:                                         │
│  • Scope document (PDF)                                                 │
│  • Requirements (Word)                                                  │
│  • Data models (Excel)                                                  │
│  • Wireframes (Images)                                                  │
│  • Client communications (Emails)                                       │
│                                                                          │
│  Place in a folder: /project-docs/                                      │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 1: PROJECT SETUP                                                   │
│  ════════════════════                                                   │
│                                                                          │
│  Step 1: Create project and install Sentinel                            │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ mkdir new-project && cd new-project                          │    │
│  │ $ git init                                                      │    │
│  │ $ ./synapsevibsentinel.sh                                       │    │
│  │ ✅ Sentinel binary compiled                                     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 2: Initialize with standards AND business docs                    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel init --with-business-docs                         │    │
│  │                                                                 │    │
│  │ 🏗️ Sentinel: Initializing New Project                          │    │
│  │                                                                 │    │
│  │ --- Service Line ---                                           │    │
│  │ 1) 🌐 Web App                                                   │    │
│  │ 2) 📱 Mobile (Cross-Platform)                                   │    │
│  │ 3) 🍏 Mobile (Native)                                           │    │
│  │ 4) 🛍️  Commerce                                                 │    │
│  │ 5) 🧠 AI & Data                                                 │    │
│  │ 6) 🔧 Infrastructure/Shell Scripts                              │    │
│  │ Selection: 1                                                    │    │
│  │                                                                 │    │
│  │ --- Naming Convention ---                                       │    │
│  │ 1) camelCase (JavaScript/TypeScript)                            │    │
│  │ 2) snake_case (Python)                                          │    │
│  │ Selection: 1                                                    │    │
│  │                                                                 │    │
│  │ ✅ Created .cursor/rules/ with project standards               │    │
│  │ ✅ Created docs/knowledge/business/ templates                  │    │
│  │ ✅ Git hooks installed                                          │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 3: Ingest project documents (KEY STEP!)                           │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel ingest /project-docs/                             │    │
│  │                                                                 │    │
│  │ Found 6 documents:                                             │    │
│  │ ├── Scope_Document.pdf (2.3 MB)                                │    │
│  │ ├── Requirements.docx (156 KB)                                 │    │
│  │ ├── Data_Model.xlsx (89 KB)                                    │    │
│  │ ├── wireframe_login.png (340 KB)                               │    │
│  │ ├── wireframe_dashboard.png (520 KB)                           │    │
│  │ └── client_kickoff.eml (23 KB)                                 │    │
│  │                                                                 │    │
│  │ Processing mode:                                                │    │
│  │ 1. Hybrid (text local, structure via cloud) - Recommended      │    │
│  │ Selection: 1                                                    │    │
│  │                                                                 │    │
│  │ 🔍 Parsing documents locally...                                │    │
│  │ 🤖 Extracting knowledge with LLM (Azure Claude Opus 4.5 or Ollama)...                            │    │
│  │                                                                 │    │
│  │ EXTRACTED:                                                      │    │
│  │ ├── 15 entities (User, Order, Product, etc.)                   │    │
│  │ ├── 12 business rules                                          │    │
│  │ ├── 5 user journeys                                            │    │
│  │ └── 3 objectives                                               │    │
│  │                                                                 │    │
│  │ ✅ Drafts created - REVIEW REQUIRED                            │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 4: Review and approve extracted knowledge                         │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel review                                             │    │
│  │                                                                 │    │
│  │ REVIEWING: domain-glossary.draft.md                            │    │
│  │                                                                 │    │
│  │ 1. User                                                        │    │
│  │    Definition: A registered customer who can place orders      │    │
│  │    Source: Scope_Document.pdf, page 5                          │    │
│  │    Confidence: 95%                                              │    │
│  │    [A]ccept  [E]dit  [R]eject  [S]kip: A                       │    │
│  │                                                                 │    │
│  │ 2. Order                                                       │    │
│  │    Definition: A purchase request containing products          │    │
│  │    Confidence: 92%                                              │    │
│  │    [A]ccept  [E]dit  [R]eject  [S]kip: A                       │    │
│  │                                                                 │    │
│  │ ... (review all items)                                         │    │
│  │                                                                 │    │
│  │ ✅ domain-glossary.md APPROVED                                  │    │
│  │ ✅ business-rules.md APPROVED                                   │    │
│  │ ✅ user-journeys.md APPROVED                                    │    │
│  │                                                                 │    │
│  │ Knowledge is now active for Cursor! 🎉                          │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 5: Verify everything is ready                                     │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel status                                             │    │
│  │                                                                 │    │
│  │ PROJECT STATUS                                                  │    │
│  │ ══════════════                                                  │    │
│  │ Code: ✅ Clean (no code yet)                                    │    │
│  │ Patterns: ✅ Configured                                         │    │
│  │ Hooks: ✅ Installed                                             │    │
│  │ Documentation: ✅ Complete                                      │    │
│  │   ├── 15 entities defined                                      │    │
│  │   ├── 12 business rules documented                             │    │
│  │   └── 5 user journeys mapped                                   │    │
│  │                                                                 │    │
│  │ Ready to start coding! 🚀                                       │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 2+: CODING WITH BUSINESS CONTEXT                                   │
│  ════════════════════════════════════                                   │
│                                                                          │
│  Developer opens Cursor, starts first feature...                        │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ Developer: "add order cancellation feature"                    │    │
│  │                                                                 │    │
│  │ [MCP: sentinel_get_business_context("order", "cancellation")]  │    │
│  │                                                                 │    │
│  │ Cursor: "I found business rules for order cancellation:        │    │
│  │                                                                 │    │
│  │ RULES TO IMPLEMENT:                                            │    │
│  │ ├── BR-001: 24-hour cancellation window                        │    │
│  │ ├── BR-002: Side effects (refund, inventory, email)           │    │
│  │ └── BR-003: Premium users get 48-hour window                   │    │
│  │                                                                 │    │
│  │ I'll implement all rules. Should I proceed?"                   │    │
│  │                                                                 │    │
│  │ Developer: "yes"                                                │    │
│  │                                                                 │    │
│  │ [Cursor generates business-aware code]                         │    │
│  │                                                                 │    │
│  │ Cursor: "Here's the implementation following:                  │    │
│  │ ✅ camelCase naming                                            │    │
│  │ ✅ BR-001 (cancellation window check)                          │    │
│  │ ✅ BR-002 (all side effects)                                   │    │
│  │ ✅ BR-003 (premium user exception)"                            │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

### Journey 2: Existing Project (Brownfield)

Adopting Sentinel on a project that's already in development.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    EXISTING PROJECT JOURNEY                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 1: INSTALLATION & DISCOVERY                                        │
│  ════════════════════════════════                                       │
│                                                                          │
│  Developer: "I just joined, this codebase is chaos"                     │
│                                                                          │
│  Step 1: Install Sentinel                                               │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ cd /path/to/existing-project                                  │    │
│  │ $ ./synapsevibsentinel.sh                                       │    │
│  │ ✅ Sentinel binary compiled                                     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 2: Learn existing patterns (detects what's there)                 │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel learn                                              │    │
│  │                                                                 │    │
│  │ 🔍 Analyzing 847 files...                                       │    │
│  │                                                                 │    │
│  │ PATTERNS DETECTED:                                              │    │
│  │ ├── Naming: camelCase (73%), snake_case (27%) ⚠️ Mixed         │    │
│  │ ├── Imports: Relative paths (85%)                               │    │
│  │ ├── Structure: src/ with flat structure                         │    │
│  │ └── Tests: *.test.js pattern                                    │    │
│  │                                                                 │    │
│  │ ⚠️  Low confidence in naming - multiple styles detected         │    │
│  │                                                                 │    │
│  │ What should be the standard?                                    │    │
│  │ 1. camelCase (most common currently)                            │    │
│  │ 2. snake_case                                                   │    │
│  │ Selection: 1                                                    │    │
│  │                                                                 │    │
│  │ ✅ Patterns saved to .sentinel/patterns.json                    │    │
│  │ ✅ Cursor rules generated                                       │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 3: Initial audit (understand current state)                       │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel audit                                              │    │
│  │                                                                 │    │
│  │ FINDINGS:                                                       │    │
│  │ ├── 🔴 CRITICAL: 3                                              │    │
│  │ ├── 🟡 WARNING: 47                                              │    │
│  │ └── ℹ️  INFO: 12                                                │    │
│  │                                                                 │    │
│  │ COMPLIANCE: 62%                                                 │    │
│  │ ⛔ Audit FAILED (3 critical issues)                             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 4: Baseline known issues (can't fix everything today)             │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel baseline add src/api/config.js 42 "secret" \      │    │
│  │     "Known issue, JIRA-1234"                                    │    │
│  │                                                                 │    │
│  │ ✅ Finding baselined                                            │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 5: Apply safe fixes (quick wins)                                  │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel fix --safe                                         │    │
│  │                                                                 │    │
│  │ SAFE FIXES AVAILABLE:                                           │    │
│  │ ├── Remove 28 console.log statements                           │    │
│  │ ├── Fix 4 trailing whitespace issues                           │    │
│  │ └── Sort imports in 12 files                                   │    │
│  │                                                                 │    │
│  │ Apply all? [Y/n]: Y                                             │    │
│  │                                                                 │    │
│  │ 💾 Backup created                                               │    │
│  │ ✅ Applied 44 safe fixes                                        │    │
│  │                                                                 │    │
│  │ COMPLIANCE: 78% (was 62%)                                       │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 6: Install hooks (prevent new issues)                             │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel install-hooks                                      │    │
│  │ ✅ Git hooks installed                                          │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 2: ADD BUSINESS DOCUMENTATION                                      │
│  ═════════════════════════════════                                      │
│                                                                          │
│  Step 7: Gather existing project documents                              │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ Find and collect:                                              │    │
│  │ • Original scope document                                      │    │
│  │ • Requirements from emails                                      │    │
│  │ • Existing wiki pages (export as PDF)                          │    │
│  │ • Data model diagrams                                          │    │
│  │ • Client communications                                         │    │
│  │                                                                 │    │
│  │ Place in: /project-docs/                                       │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 8: Ingest documents                                               │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel ingest /project-docs/                             │    │
│  │                                                                 │    │
│  │ Found 4 documents:                                             │    │
│  │ ├── original_scope.pdf                                         │    │
│  │ ├── requirements_email_chain.eml                               │    │
│  │ ├── data_model_diagram.png                                     │    │
│  │ └── feature_requests.xlsx                                      │    │
│  │                                                                 │    │
│  │ 🔍 Processing...                                                │    │
│  │                                                                 │    │
│  │ EXTRACTED:                                                      │    │
│  │ ├── 23 entities                                                │    │
│  │ ├── 18 business rules                                          │    │
│  │ └── 7 user journeys                                            │    │
│  │                                                                 │    │
│  │ ✅ Drafts created - REVIEW REQUIRED                            │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 9: Review and approve                                             │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel review                                             │    │
│  │                                                                 │    │
│  │ [Review each item, accept/edit/reject]                         │    │
│  │                                                                 │    │
│  │ REVIEW SUMMARY:                                                │    │
│  │ ├── Accepted: 35 items                                         │    │
│  │ ├── Edited: 8 items (clarified definitions)                    │    │
│  │ ├── Rejected: 3 items (hallucinated)                           │    │
│  │ └── Skipped: 2 items (need team input)                         │    │
│  │                                                                 │    │
│  │ ✅ Knowledge approved and active                                │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DAY 3+: DAILY DEVELOPMENT WITH FULL CONTEXT                            │
│  ═══════════════════════════════════════════                            │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ Developer: "add user profile update"                           │    │
│  │                                                                 │    │
│  │ Cursor: "I see the User entity has these rules:               │    │
│  │ ├── BR-005: Email changes require verification                │    │
│  │ ├── BR-006: Username cannot be changed after 30 days          │    │
│  │ └── BR-007: Profile changes logged for audit                  │    │
│  │                                                                 │    │
│  │ I'll implement following your patterns:                        │    │
│  │ ├── camelCase naming                                          │    │
│  │ ├── Relative imports                                          │    │
│  │ └── Existing UserService pattern                              │    │
│  │                                                                 │    │
│  │ Should I proceed?"                                             │    │
│  │                                                                 │    │
│  │ Developer: "yes"                                                │    │
│  │                                                                 │    │
│  │ [Cursor generates code matching patterns + business rules]     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  WEEKLY: TRACK IMPROVEMENT                                              │
│  ═════════════════════════                                              │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel status                                             │    │
│  │                                                                 │    │
│  │ PROJECT HEALTH                                                  │    │
│  │ ══════════════                                                  │    │
│  │ Compliance: 78% → 89% (↑11% this month)                        │    │
│  │ Baselined: 2 → 0 (all fixed!)                                  │    │
│  │ New issues: 0 this week 🎉                                      │    │
│  │                                                                 │    │
│  │ DOCUMENTATION                                                   │    │
│  │ ═════════════                                                   │    │
│  │ Coverage: 85%                                                  │    │
│  │ Pending drafts: 2 (need review)                                │    │
│  │ Business rules: 18 documented, 15 implemented                  │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

### Journey Comparison

| Step | New Project | Existing Project |
|------|-------------|------------------|
| 1 | Gather documents | Install Sentinel |
| 2 | Install Sentinel | Learn patterns |
| 3 | Init with business docs | Audit current state |
| 4 | Ingest documents | Baseline known issues |
| 5 | Review & approve | Apply safe fixes |
| 6 | Start coding | Install hooks |
| 7 | - | Gather documents |
| 8 | - | Ingest documents |
| 9 | - | Review & approve |
| 10 | - | Continue coding |

**Key Insight**: New projects can set up everything on Day 1. Existing projects take 2-3 days but then have full protection.

---

## Command Reference

### `sentinel init`

Initialize Sentinel in a project.

```bash
# Interactive mode
./sentinel init

# Non-interactive mode
./sentinel init --stack web --db sql --non-interactive

# With business documentation templates
./sentinel init --with-business-docs
```

### `sentinel learn`

Extract patterns from existing code.

```bash
# Full learning
./sentinel learn

# Specific patterns
./sentinel learn --naming
./sentinel learn --imports
./sentinel learn --structure

# Output format
./sentinel learn --output json
```

### `sentinel audit`

Scan for issues.

```bash
# Basic scan
./sentinel audit

# With output file
./sentinel audit --output json --output-file report.json
./sentinel audit --output html --output-file report.html

# Business rule coverage
./sentinel audit --business-rules

# CI mode (exit code reflects status)
./sentinel audit --ci
```

### `sentinel status`

View project health dashboard.

```bash
# Show project health
./sentinel status
```

Output:
```
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

⚡ QUICK ACTIONS
──────────────────────────────────────────────────────────────
   [AUTO] 5 safe fixes available (run: sentinel fix --safe)

📈 OVERALL HEALTH
──────────────────────────────────────────────────────────────
   Score: [████████░░] 80% - Good
```

### `sentinel fix`

Apply fixes to code.

```bash
# Interactive mode
./sentinel fix

# Safe fixes only (no prompts)
./sentinel fix --safe

# Preview without applying
./sentinel fix --dry-run

# Auto-approve all
./sentinel fix --yes

# Specific pattern
./sentinel fix --pattern "console.log"

# Rollback last fix
./sentinel fix rollback
```

### `sentinel ingest`

Process project documents (server-side by default).

> **Architecture**: Documents are uploaded to Sentinel Hub for processing.
> This eliminates the need to install PDF/OCR tools on developer machines.
> See [Architecture Decision](./ARCHITECTURE_DOCUMENT_PROCESSING.md).

```bash
# Upload to Hub (default - recommended)
./sentinel ingest /path/to/docs/

# Check processing status
./sentinel ingest --status

# Sync completed results to local
./sentinel ingest --sync

# Skip image processing (faster)
./sentinel ingest /path/to/docs/ --skip-images

# Offline mode (limited formats, no LLM)
./sentinel ingest /path/to/docs/ --offline

# Check offline capabilities
./sentinel ingest --offline-info
```

**Supported Formats**:

| Format | Server (Hub) | Offline (Local) |
|--------|--------------|-----------------|
| .txt, .md | ✅ | ✅ |
| .docx | ✅ | ✅ |
| .xlsx | ✅ | ✅ |
| .eml | ✅ | ✅ |
| .pdf | ✅ | ⚠️ Requires poppler |
| .png, .jpg | ✅ | ⚠️ Requires tesseract |
| LLM extraction | ✅ | ❌ |

### `sentinel review`

Review extracted knowledge.

```bash
# Interactive review
./sentinel review

# List pending items
./sentinel review --list

# Approve specific file
./sentinel review --approve domain-glossary.draft.md

# Reject specific file
./sentinel review --reject user-journeys.draft.md
```

### `sentinel baseline`

Manage accepted findings.

```bash
# Add to baseline
./sentinel baseline add src/file.js 42 "pattern" "reason"

# List baselined items
./sentinel baseline list

# Remove from baseline
./sentinel baseline remove src/file.js 42
```

### `sentinel status`

View project health.

```bash
./sentinel status

# Output:
# PROJECT STATUS
# ══════════════
# Compliance: 92%
# Pending drafts: 3
# Last audit: 2 hours ago
# Documentation: 85% coverage
```

---

## Configuration

### Project Configuration (`.sentinelsrc`)

```json
{
  "scanDirs": ["src", "lib"],
  "excludePaths": [
    "node_modules",
    ".git",
    "dist",
    "*.test.js"
  ],
  "severityLevels": {
    "console.log": "warning",
    "eval": "critical"
  },
  "customPatterns": {
    "todo": "TODO:|FIXME:|HACK:"
  },
  "ingest": {
    "llmProvider": "openai",
    "localOnly": false,
    "visionEnabled": true
  },
  "hub": {
    "url": "https://hub.yourcompany.com",
    "apiKey": "sk_live_xxxxx",
    "projectId": "optional-project-id"
  }
}
```

**Telemetry Configuration**:

Telemetry is automatically enabled when Hub is configured. The Agent sends metrics to the Hub after each `audit`, `fix`, and `learn` command. If the Hub is unreachable, events are queued locally in `.sentinel/telemetry-queue.json` and sent automatically when the Hub becomes available.

**What is sent**:
- Audit results: finding counts, compliance percentage
- Fix statistics: number of fixes applied, fix types
- Pattern learning: confidence scores, pattern types
- Document ingestion: document counts

**What is NOT sent**:
- Source code content
- File contents
- Actual patterns or code snippets
- Any sensitive data

### Cursor MCP Configuration (`~/.cursor/mcp.json`)

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/path/to/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_PROJECT": "/path/to/project"
      }
    }
  }
}
```

---

## Troubleshooting

### "Draft documents not being used by Cursor"

This is intentional! Drafts must be reviewed and approved first.

```bash
./sentinel review
# Approve all items, then drafts become active
```

### "Pattern learning shows low confidence"

This is expected for existing projects with mixed styles. Choose the dominant pattern or configure manually.

### "Document ingest failed for PDF"

Ensure `pdftotext` is installed:
```bash
# macOS
brew install poppler

# Ubuntu
apt-get install poppler-utils
```

### "Vision API not working for images"

Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-key"
```

Or use local-only mode (OCR only, no diagram understanding):
```bash
./sentinel ingest /docs/ --local-only
```

---

## Best Practices

### For New Projects

1. **Document First**: Gather all project documents before starting code
2. **Review Carefully**: Take time to validate extracted knowledge
3. **Start Clean**: Initialize patterns before writing any code
4. **Keep Updated**: Add new documents as project evolves

### For Existing Projects

1. **Baseline Strategically**: Don't baseline everything, fix what you can
2. **Document Incrementally**: Add business docs over time
3. **Track Progress**: Use `sentinel status` to monitor improvement
4. **Celebrate Wins**: Compliance going up means less technical debt

### For Teams

1. **Share Patterns**: Use hub to distribute org-wide patterns
2. **Review Together**: Have team review extracted business rules
3. **Onboard with Status**: New devs run `sentinel status` first
4. **Standardize Documents**: Use consistent doc formats for ingestion


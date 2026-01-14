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
│  Step 3: Upload project documents to Hub (KEY STEP!)                 │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ # Via Hub API or Dashboard:                                   │    │
│  │ $ curl -X POST https://hub.example.com/api/v1/documents/ingest │    │
│  │   -H "Authorization: Bearer YOUR_API_KEY"                     │    │
│  │   -F "files=@Scope_Document.pdf"                              │    │
│  │   -F "files=@Requirements.docx"                               │    │
│  │   -F "files=@Data_Model.xlsx"                                 │    │
│  │                                                                 │    │
│  │ # Or use Hub Dashboard:                                        │    │
│  │ # 1. Login to https://hub.example.com                         │    │
│  │ # 2. Navigate to Documents section                            │    │
│  │ # 3. Upload project documents                                 │    │
│  │ # 4. Wait for processing                                      │    │
│  │                                                                 │    │
│  │ 📄 Documents uploaded and processing...                       │    │
│  │ 🤖 Extracting knowledge with LLM...                           │    │
│  │                                                                 │    │
│  │ EXTRACTED:                                                      │    │
│  │ ├── 15 entities (User, Order, Product, etc.)                   │    │
│  │ ├── 12 business rules                                          │    │
│  │ ├── 5 user journeys                                            │    │
│  │ └── 3 objectives                                               │    │
│  │                                                                 │    │
│  │ ✅ Knowledge extracted and available in Hub                    │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 4: Review and approve extracted knowledge                         │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ # Via Hub Dashboard:                                           │    │
│  │ # 1. Login to https://hub.example.com                          │    │
│  │ # 2. Navigate to Knowledge section                             │    │
│  │ # 3. Review extracted entities, rules, and journeys            │    │
│  │ # 4. Approve or edit each item                                 │    │
│  │                                                                 │    │
│  │ REVIEWING EXTRACTED KNOWLEDGE:                                  │    │
│  │                                                                 │    │
│  │ 1. User                                                        │    │
│  │    Definition: A registered customer who can place orders      │    │
│  │    Source: Scope_Document.pdf, page 5                          │    │
│  │    Confidence: 95%                                              │    │
│  │    [✓] Approved                                                 │    │
│  │                                                                 │    │
│  │ 2. Order                                                       │    │
│  │    Definition: A purchase request containing products          │    │
│  │    Confidence: 92%                                              │    │
│  │    [✓] Approved                                                 │    │
│  │                                                                 │    │
│  │ ... (review all items via Hub interface)                       │    │
│  │                                                                 │    │
│  │ ✅ Knowledge approved and synced to project                    │    │
│  │ ✅ Available for Cursor integration                             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 5: Verify project setup is complete                             │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel audit                                              │    │
│  │                                                                 │    │
│  │ PROJECT VERIFICATION                                            │    │
│  │ ═══════════════════                                             │    │
│  │ ✅ Rules configured (.cursor/rules/)                           │    │
│  │ ✅ Business docs created (docs/knowledge/)                     │    │
│  │ ✅ Hub connection configured                                   │    │
│  │ ✅ Knowledge uploaded and approved                             │    │
│  │                                                                 │    │
│  │ Ready to start coding with business context! 🚀                 │    │
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
│  Step 3.5: Check project status (get overview)                           │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ $ ./sentinel status                                             │    │
│  │                                                                 │    │
│  │ 📊 PROJECT STATUS:                                               │    │
│  │ ├── Compliance: 62%                                              │    │
│  │ ├── Critical Issues: 3                                           │    │
│  │ ├── Warning Issues: 47                                          │    │
│  │ ├── Test Coverage: 45%                                           │    │
│  │ ├── Business Rules: 0 documented                                │    │
│  │ └── Last Updated: 2026-01-08 13:42                               │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 4: Document known issues (can't fix everything today)             │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ # Document known issues for future resolution:                 │    │
│  │ # 1. Create issue in project tracker (JIRA, GitHub, etc.)      │    │
│  │ # 2. Add TODO comments in code for temporary workarounds       │    │
│  │ # 3. Update team documentation                                  │    │
│  │                                                                 │    │
│  │ ✅ Issues documented for future resolution                      │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  Step 5: Address critical issues manually                               │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │ # Manually fix critical issues:                                │    │
│  │ # 1. Remove hardcoded secrets from config.js                   │    │
│  │ # 2. Fix SQL injection vulnerabilities                         │    │
│  │ # 3. Add input validation for user data                        │    │
│  │ # 4. Update dependencies to fix known CVEs                     │    │
│  │                                                                 │    │
│  │ 💾 Create git commit with fixes                                │    │
│  │ ✅ Critical security issues resolved                           │    │
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
│  │ # Via Hub Dashboard:                                           │    │
│  │ # 1. Login to https://hub.example.com                          │    │
│  │ # 2. Navigate to Knowledge section                             │    │
│  │ # 3. Review extracted patterns and knowledge                   │    │
│  │ # 4. Accept, edit, or reject each item                         │    │
│  │                                                                 │    │
│  │ REVIEW SUMMARY:                                                │    │
│  │ ├── Accepted: 35 items                                         │    │
│  │ ├── Edited: 8 items (clarified definitions)                    │    │
│  │ ├── Rejected: 3 items (hallucinated)                           │    │
│  │ └── Skipped: 2 items (need team input)                         │    │
│  │                                                                 │    │
│  │ ✅ Knowledge approved and synced to project                    │    │
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

## Task Dependency & Verification (Phase 14E)

### Overview

The task dependency and verification system tracks Cursor-generated tasks, verifies completion, and manages dependencies. This ensures tasks are completed and dependencies are properly managed.

### Quick Start

```bash
# Scan codebase for tasks
sentinel tasks scan

# List all tasks
sentinel tasks list

# Verify a specific task
sentinel tasks verify TASK-123

# Show dependency graph
sentinel tasks dependencies
```

### Task Scanning

Scan your codebase to detect tasks from TODO comments, task markers, and Cursor task format:

```bash
# Scan entire codebase
sentinel tasks scan

# Scan specific directory
sentinel tasks scan --dir src/

# Scan with filters
sentinel tasks scan --source cursor --status pending
```

**Example Output**:
```
🔍 Scanning codebase for tasks...
✅ Found 15 tasks:
  TASK-001: Implement user authentication (pending, high)
    File: src/auth/middleware.js:45
    Source: cursor
  TASK-002: Add JWT token refresh (pending, medium)
    File: src/auth/token.js:23
    Source: cursor
  TASK-003: Add payment processing (in_progress, critical)
    File: src/payments/processor.js:67
    Source: change_request
  ...
```

### Task Listing

List tasks with various filters:

```bash
# List all tasks
sentinel tasks list

# List pending tasks
sentinel tasks list --status pending

# List high priority tasks
sentinel tasks list --priority high

# List with dependencies
sentinel tasks list --show-dependencies
```

**Example Output**:
```
📋 Tasks (15 total)
══════════════════════════════════════════════════════════════

PENDING (8):
  TASK-001: Implement user authentication [high]
    Depends on: TASK-002, TASK-003
    File: src/auth/middleware.js:45
  
  TASK-002: Add JWT token refresh [medium]
    File: src/auth/token.js:23

IN_PROGRESS (5):
  TASK-003: Add payment processing [critical]
    Verification: 0.69 confidence
    File: src/payments/processor.js:67

COMPLETED (2):
  TASK-004: Setup database schema [high]
    Completed: 2024-12-10 14:30:00
```

### Task Verification

Verify task completion using multi-factor verification:

```bash
# Verify specific task
sentinel tasks verify TASK-123

# Verify all pending tasks
sentinel tasks verify --all

# Verify with force (ignore cache)
sentinel tasks verify TASK-123 --force
```

**Example Output**:
```
🔍 Verifying task TASK-001: Implement user authentication
  ✓ Code existence: 0.95 (verified)
    Found: src/auth/middleware.js:45 (authenticateUser function)
  ✓ Code usage: 0.88 (verified)
    Call sites: src/routes/users.js:23, src/routes/orders.js:45
  ✓ Test coverage: 0.92 (verified)
    Test file: tests/auth/middleware.test.js
    Coverage: 95%
  ✗ Integration: 0.0 (pending)
    Missing: External service configuration
  
Overall confidence: 0.69 → Status: in_progress
⚠️  Task needs integration verification
```

### Dependency Management

View and manage task dependencies:

```bash
# Show dependency graph
sentinel tasks dependencies

# Show dependencies for specific task
sentinel tasks dependencies TASK-123

# Export dependency graph
sentinel tasks dependencies --export graph.json
```

**Example Output**:
```
📊 Dependency Graph for TASK-003: Add payment processing
  │
  ├── TASK-001: Implement user authentication [explicit]
  │   └── TASK-002: Add JWT token refresh [implicit]
  │       └── TASK-005: Add token validation [explicit]
  │
  └── TASK-004: Setup payment gateway [integration]
      └── TASK-006: Configure API keys [explicit]

⚠️  Circular dependency detected: TASK-007 ↔ TASK-008
```

### Task Completion

Manually mark tasks as complete or use auto-completion:

```bash
# Manually mark task complete
sentinel tasks complete TASK-123

# Mark with reason
sentinel tasks complete TASK-123 --reason "Implemented manually"

# Auto-complete verified tasks
sentinel tasks complete --auto
```

**Example Output**:
```
🔍 Verifying all pending tasks...
  TASK-001: 0.69 confidence → in_progress
  TASK-002: 0.92 confidence → ✅ auto-completed
  TASK-003: 0.45 confidence → pending
  TASK-004: 0.88 confidence → ✅ auto-completed
  TASK-005: 0.91 confidence → ✅ auto-completed
  
✅ 3 tasks auto-completed
⚠️  2 tasks need attention
```

### Integration with Other Commands

Task verification integrates with other Sentinel commands:

```bash
# Include task verification in audit
sentinel audit --tasks

# Link tasks to change requests
sentinel knowledge track CR-001 --create-tasks

# Verify tasks from comprehensive analysis
sentinel analyze feature "Order Cancellation" --create-tasks
```

### Troubleshooting

**Tasks not detected**:
- Ensure files are in scanned directories (check `.sentinelsrc`)
- Check task format matches supported patterns (TODO, FIXME, Cursor markers)
- Run with `--verbose` flag for detailed output

**Verification fails**:
- Check code exists in expected locations
- Verify test files match naming conventions
- Check integration configuration files exist

**Dependencies not detected**:
- Ensure tasks have explicit dependencies in descriptions
- Run comprehensive analysis (Phase 14A) for feature-level dependencies
- Check code analysis for implicit dependencies

**Auto-completion not working**:
- Check verification confidence scores (need >0.8)
- Verify all verification factors are checked
- Check for blocking dependencies

### Best Practices

1. **Task Format**: Use consistent task format for better detection
   ```javascript
   // TASK: TASK-123 - Description
   // DEPENDS: TASK-122, TASK-121
   ```

2. **Regular Verification**: Run `sentinel tasks verify --all` regularly
   - Before commits: Verify tasks are complete
   - Before releases: Ensure all critical tasks done
   - Weekly: Review pending tasks

3. **Dependency Management**: Keep dependencies explicit
   - Document dependencies in task descriptions
   - Review dependency graph regularly
   - Resolve circular dependencies quickly

4. **Integration**: Link tasks to related systems
   - Link to change requests (Phase 12)
   - Link to knowledge items (Phase 4)
   - Link to comprehensive analysis (Phase 14A)

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




### `sentinel test`

Run comprehensive test suites for your project.

```bash
# Generate and run all tests
./sentinel test

# Test requirements generation
./sentinel test requirements

# Test coverage analysis
./sentinel test coverage

# Test validation
./sentinel test validate

# Run tests
./sentinel test run

# Mutation testing
./sentinel test mutation
```

### `sentinel status`

Display project health and status information.

```bash
# Show project overview
./sentinel status

# Include detailed metrics
./sentinel status --detailed

# JSON output for CI/CD
./sentinel status --json
```

### `sentinel baseline`

Manage baseline exceptions for known issues.

```bash
# Create baseline from current issues
./sentinel baseline create

# Update existing baseline
./sentinel baseline update

# Show baseline contents
./sentinel baseline show

# Clear baseline (reset to no exceptions)
./sentinel baseline clear
```

### `sentinel tasks`

Manage development tasks and track progress.

```bash
# Scan codebase for tasks
./sentinel tasks scan

# List all tasks
./sentinel tasks list

# Verify task completion
./sentinel tasks verify <task-id>

# Analyze task dependencies
./sentinel tasks dependencies <task-id>
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

### MCP Integration (Phase 14B)

Sentinel provides MCP (Model Context Protocol) integration for Cursor IDE, enabling comprehensive feature analysis directly from your IDE.

**Status**: ✅ 15/19 MCP tools fully functional (79% complete)

#### Available MCP Tools

The following tools are fully functional:

- `sentinel_analyze_feature_comprehensive` - Comprehensive feature analysis across all layers
- `sentinel_check_intent` - Analyze unclear prompts and generate clarifying questions
- `sentinel_get_context` - Get recent activity context (git status, recent commits, errors)
- `sentinel_get_patterns` - Get learned patterns and project conventions
- `sentinel_get_business_context` - Get business rules, entities, and journeys
- `sentinel_get_security_context` - Get security rules, compliance status, and security score
- `sentinel_get_test_requirements` - Get test requirements and coverage status
- `sentinel_check_file_size` - Check file size and get warnings/split suggestions
- `sentinel_validate_security` - Validate code for security compliance
- `sentinel_validate_business` - Validate code against business rules
- `sentinel_validate_tests` - Validate test quality and coverage
- `sentinel_generate_tests` - Generate test cases for a feature
- `sentinel_run_tests` - Execute tests in sandbox

#### Available Tools (Complete)

- `sentinel_analyze_intent` - ✅ Analyze user intent and return context, rules, security, and test requirements
- `sentinel_validate_code` - ✅ Validate code using AST analysis
- `sentinel_apply_fix` - ✅ Apply security, style, or performance fixes to code

#### Known Limitations

- Task management tools (`sentinel_get_task_status`, `sentinel_verify_task`, `sentinel_list_tasks`) - Require Phase 14E completion

**Setup**:
1. Configure Cursor MCP settings in `~/.cursor/mcp.json`
2. Restart Cursor IDE
3. Use `sentinel_analyze_feature_comprehensive` tool in Cursor chat

**For detailed setup and usage, see [Phase 14B Guide](./PHASE_14B_GUIDE.md)**

### Intent Analysis (Phase 15)

Phase 15 adds intent analysis to handle unclear prompts gracefully. When you provide a vague request, Sentinel analyzes the intent and generates clarifying questions.

**Usage in Cursor**:
```
Use sentinel_check_intent to analyze: "add a button"
```

**Features**:
- Detects unclear prompts (location, entity, action confirmation)
- Generates clarifying questions with options
- Gathers context (recent files, git status, business rules)
- Learns from your choices to improve future suggestions

**For detailed setup and usage, see [Phase 15 Guide](./PHASE_15_GUIDE.md)**

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


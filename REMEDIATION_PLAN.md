# 🔧 Sentinel Remediation Plan: From Monolith to Modular

**Created:** January 17, 2026  
**Status:** ACTIVE - Single Source of Truth for Project Recovery  
**Compliance Target:** 100% CODING_STANDARDS.md Compliance  

---

## 📋 Executive Summary

This document consolidates all gap analyses and provides a **focused, actionable plan** to transform Sentinel from a non-functional monolithic codebase to a production-ready, modular application.

### Current State Assessment
- **Functional Completion:** ~35-40%
- **CODING_STANDARDS.md Compliance:** ~15%
- **Build Status:** Agent builds, Hub does not compile
- **Test Pass Rate:** 23-27% across integration tests

### Target State
- **Functional Completion:** 90%+
- **CODING_STANDARDS.md Compliance:** 100%
- **Build Status:** All components compile and run
- **Test Pass Rate:** 80%+ across all test suites

---

## 🚨 Critical Issues (Must Fix First)

### Issue 1: Monolithic Architecture

**Problem:** `synapsevibsentinel.sh` contains an embedded 18,178-line Go program that:
- Violates every CODING_STANDARDS.md principle
- Self-compiles at runtime
- Cannot be tested, linted, or maintained
- Contains stub implementations documented as "complete"

**Solution:** Disable monolithic generation, implement modular architecture.

### Issue 2: Two Parallel Implementations

**Problem:** Two incompatible entry points exist:
1. `cmd/sentinel/main.go` - Clean but only implements user management
2. `synapsevibsentinel.sh` - Monolithic but has feature implementations

**Solution:** Migrate features from monolith to modular packages.

### Issue 3: Hub API Cannot Compile

**Problem:** `hub/api/` has 72 files in `package main` with:
- Undefined type errors (`ASTFinding`)
- Duplicate function definitions
- Incomplete refactoring

**Solution:** Complete package refactoring or create minimal viable Hub.

### Issue 4: Documentation Inaccuracy

**Problem:** Multiple conflicting status documents claim:
- FINAL_STATUS.md: "98% complete" → Actually ~35%
- IMPLEMENTATION_ROADMAP.md: All phases "✅ COMPLETE" → Many are stubs
- FEATURES.md: Features documented as working → Many don't work

**Solution:** Archive conflicting docs, maintain single source of truth.

---

## 📁 Target Architecture (CODING_STANDARDS.md Compliant)

```
sentinel/
├── cmd/
│   └── sentinel/
│       └── main.go                 # Entry point (<50 lines) ✅
├── internal/
│   ├── cli/                        # CLI command handling
│   │   ├── commands.go             # Command registry
│   │   ├── init.go                 # init command
│   │   ├── audit.go                # audit command
│   │   ├── learn.go                # learn command
│   │   ├── fix.go                  # fix command
│   │   ├── status.go               # status command
│   │   └── baseline.go             # baseline command
│   ├── scanner/                    # Security scanning
│   │   ├── scanner.go              # Main scanner interface
│   │   ├── patterns.go             # Security patterns
│   │   └── findings.go             # Finding types
│   ├── patterns/                   # Pattern learning
│   │   ├── learner.go              # Pattern learner
│   │   ├── detector.go             # Pattern detection
│   │   └── types.go                # Pattern types
│   ├── mcp/                        # MCP server
│   │   ├── server.go               # MCP server
│   │   ├── handlers.go             # Tool handlers
│   │   └── protocol.go             # JSON-RPC protocol
│   ├── config/                     # Configuration ✅ EXISTS
│   │   └── config.go
│   ├── models/                     # Data models ✅ EXISTS
│   │   └── types.go
│   └── hub/                        # Hub client
│       ├── client.go               # HTTP client
│       └── types.go                # Hub types
├── hub/                            # Hub API (separate service)
│   └── api/
│       ├── cmd/
│       │   └── main.go             # Hub entry point
│       ├── internal/               # Hub internal packages
│       └── ...
├── docs/
│   ├── REMEDIATION_PLAN.md         # THIS FILE - Single source of truth
│   ├── USER_GUIDE.md               # User documentation
│   └── external/
│       ├── CODING_STANDARDS.md     # Development standards
│       ├── TECHNICAL_SPEC.md       # Technical specification
│       └── FEATURES.md             # Feature specification
├── tests/                          # Integration tests
└── scripts/                        # Build scripts
```

---

## 🎯 Phase 0: Emergency Stabilization (Week 1)

### Task 0.1: Disable Monolithic Generation ✅ PRIORITY

**File:** `synapsevibsentinel.sh`

**Action:** Modify the script to NOT generate the embedded Go code, instead build from modular sources.

```bash
# Change from:
cat <<'EOF' > main.go
# 18,000 lines of Go code
EOF
go build -o sentinel main.go

# Change to:
go build -o sentinel ./cmd/sentinel
```

### Task 0.2: Archive Conflicting Documentation

Move these to `docs/archive/` to prevent confusion:
- `COMPREHENSIVE_GAP_ANALYSIS.md`
- `CRITICAL_ANALYSIS_REPORT.md`  
- `CRITICAL_ANALYSIS.md`
- `CRITICAL_EVALUATION_REPORT.md`
- `CRITICAL_FIX_PLAN.md`
- `COMPREHENSIVE_FIX_PLAN.md`
- `CURRENT_STATE_REVIEW.md`
- `ISSUES_31_47_DETAILED_ANALYSIS.md`
- `REFACTOR_PROGRESS_SUMMARY.md`
- `STANDARDS_VIOLATIONS.md`

### Task 0.3: Update FINAL_STATUS.md

Replace with accurate status reflecting this remediation plan.

---

## 🔨 Phase 1: Core CLI Framework (Week 2)

### Task 1.1: Create CLI Package

**File:** `internal/cli/commands.go`

```go
package cli

import (
    "fmt"
    "os"
)

type Command struct {
    Name        string
    Description string
    Run         func(args []string) error
}

var commands = make(map[string]*Command)

func Register(cmd *Command) {
    commands[cmd.Name] = cmd
}

func Execute() error {
    if len(os.Args) < 2 {
        return printHelp()
    }
    
    cmdName := os.Args[1]
    cmd, ok := commands[cmdName]
    if !ok {
        return fmt.Errorf("unknown command: %s", cmdName)
    }
    
    return cmd.Run(os.Args[2:])
}
```

### Task 1.2: Implement Core Commands

Priority order:
1. `init` - Project initialization
2. `audit` - Security scanning
3. `status` - Project health
4. `learn` - Pattern learning (basic)
5. `fix` - Auto-fix (safe fixes only)
6. `baseline` - Exception management

### Task 1.3: Update Entry Point

**File:** `cmd/sentinel/main.go`

```go
package main

import (
    "log"
    "os"

    "github.com/divyang-garg/sentinel-hub-api/internal/cli"
    _ "github.com/divyang-garg/sentinel-hub-api/internal/cli/commands"
)

func main() {
    if err := cli.Execute(); err != nil {
        log.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

---

## 🔍 Phase 2: Security Scanner (Week 3)

### Task 2.1: Create Scanner Package

**File:** `internal/scanner/scanner.go`

Extract working patterns from monolith and implement:
- Secret detection
- SQL injection patterns
- XSS vulnerability patterns
- Insecure function detection

### Task 2.2: Implement Finding Types

**File:** `internal/scanner/findings.go`

```go
package scanner

type Severity string

const (
    Critical Severity = "critical"
    High     Severity = "high"
    Medium   Severity = "medium"
    Low      Severity = "low"
    Info     Severity = "info"
)

type Finding struct {
    File     string   `json:"file"`
    Line     int      `json:"line"`
    Severity Severity `json:"severity"`
    Message  string   `json:"message"`
    Pattern  string   `json:"pattern"`
    Code     string   `json:"code"`
}
```

---

## 🧠 Phase 3: Pattern Learning (Week 4)

### Task 3.1: Create Pattern Learner

**File:** `internal/patterns/learner.go`

Implement actual pattern detection:
- Naming conventions
- Import styles
- File structure analysis

### Task 3.2: Generate Pattern Files

Output to `.sentinel/patterns.json` and `.cursor/rules/project-patterns.md`

---

## 🔧 Phase 4: Auto-Fix System (Week 5)

### Task 4.1: Safe Fixes Only

**File:** `internal/fix/safe.go`

Implement only "safe" fixes initially:
- Console.log removal
- Trailing whitespace
- Import sorting
- Debug code removal

### Task 4.2: Backup System

**File:** `internal/fix/backup.go`

Implement backup before any modifications.

---

## 📡 Phase 5: MCP Server (Week 6)

### Task 5.1: Create MCP Package

**File:** `internal/mcp/server.go`

Implement JSON-RPC 2.0 server with working tool handlers.

### Task 5.2: Implement Tool Handlers

Priority tools:
1. `sentinel_get_context`
2. `sentinel_validate_code`
3. `sentinel_get_patterns`
4. `sentinel_check_file_size`

---

## 🌐 Phase 6: Hub Integration (Week 7-8)

### Option A: Fix Existing Hub (Preferred if time permits)

- Complete package refactoring
- Fix import/type errors
- Implement database migrations

### Option B: Create Minimal Hub (Faster)

- New minimal Hub with core endpoints only
- Telemetry, metrics, basic API
- Defer advanced features

---

## ✅ Success Criteria

### Phase 0 Complete When:
- [ ] `./synapsevibsentinel.sh` builds from modular sources
- [ ] Conflicting docs archived
- [ ] FINAL_STATUS.md accurate

### Phase 1 Complete When:
- [ ] `sentinel init` works
- [ ] `sentinel audit` produces findings
- [ ] `sentinel status` shows project health
- [ ] All commands follow CODING_STANDARDS.md

### Phase 2 Complete When:
- [ ] Security scanner detects known patterns
- [ ] 80%+ pass rate on security tests
- [ ] Findings match expected format

### Phase 3 Complete When:
- [ ] Pattern learning generates valid output
- [ ] Detected patterns match codebase
- [ ] 80%+ pass rate on pattern tests

### Phase 4 Complete When:
- [ ] Safe fixes apply correctly
- [ ] Backups created before changes
- [ ] Rollback functionality works

### Phase 5 Complete When:
- [ ] MCP server responds to Cursor
- [ ] Tool handlers return real data
- [ ] JSON-RPC 2.0 compliance

### Phase 6 Complete When:
- [ ] Hub compiles and runs
- [ ] Agent can connect to Hub
- [ ] Telemetry flows through

---

## 📊 Metrics & Tracking

### Weekly Metrics to Track:
- Test pass rate (target: +10% per week)
- CODING_STANDARDS.md compliance score
- Build success (Agent + Hub)
- Lines of code in compliant packages vs monolith

### Definition of Done:
- Code compiles without errors
- Tests pass at 80%+ rate
- CODING_STANDARDS.md fully compliant
- Documentation matches implementation

---

## 🗂️ Document Hierarchy (Single Source of Truth)

### Primary Documents (AUTHORITATIVE):
1. **REMEDIATION_PLAN.md** - This file, project recovery plan
2. **docs/external/CODING_STANDARDS.md** - Development standards
3. **docs/external/TECHNICAL_SPEC.md** - Technical specification
4. **docs/external/FEATURES.md** - Feature specification

### Secondary Documents (Reference):
- **docs/USER_GUIDE.md** - End-user documentation
- **docs/external/PROJECT_VISION.md** - Vision and goals
- **README.md** - Project overview

### Archived Documents (Historical only):
All gap analysis and status documents moved to `docs/archive/`

---

## 🚀 Getting Started

### Immediate Next Steps:

1. **Run:** Review and approve this remediation plan
2. **Execute:** Disable monolithic generation in synapsevibsentinel.sh
3. **Create:** `internal/cli/` package with command framework
4. **Test:** Verify `./sentinel init` works from modular sources
5. **Iterate:** Complete Phase 1 before moving to Phase 2

### Commands to Verify Progress:

```bash
# After Phase 0:
./synapsevibsentinel.sh  # Should build from ./cmd/sentinel
./sentinel --help        # Should show available commands

# After Phase 1:
./sentinel init          # Should bootstrap project
./sentinel audit         # Should scan for issues
./sentinel status        # Should show project health

# After all phases:
go test ./...            # Should pass 80%+
./sentinel audit --ci    # Should work in CI mode
./sentinel mcp-server    # Should respond to Cursor
```

---

**This document supersedes all previous gap analyses and status documents.**

**Last Updated:** January 17, 2026  
**Next Review:** After Phase 0 completion

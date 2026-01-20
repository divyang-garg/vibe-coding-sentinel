# Coverage Gap Analysis Report

**Generated:** $(date)  
**Overall Coverage:** 69.8%  
**Thresholds:** 
- Overall packages: ≥80%
- Critical packages (handlers, services, repository, models): ≥90%

---

## 🔴 Critical Packages Below 90% Threshold

Critical packages require **90% coverage** but are currently below this threshold:

| Package | Coverage | Functions | Gap | Priority |
|---------|----------|-----------|-----|----------|
| `internal/repository/database.go` | 0.0% | 1 | 90.0% | 🔴 **CRITICAL** |
| `internal/services/user_service.go` | 76.8% | 8 | 13.2% | 🔴 **HIGH** |
| `internal/mcp/tool_handlers.go` | 82.0% | 5 | 8.0% | 🟡 Medium |
| `internal/mcp/handlers.go` | 82.6% | 7 | 7.4% | 🟡 Medium |
| `internal/repository/user_repository.go` | 84.5% | 7 | 5.5% | 🟡 Medium |
| `internal/api/handlers/user_handler.go` | 87.9% | 8 | 2.1% | 🟢 Low |

**Total Critical Gaps:** 6 packages

---

## 🟡 Non-Critical Packages Below 80% Threshold

Non-critical packages require **80% coverage** but are currently below this threshold:

| Package | Coverage | Functions | Gap | Priority |
|---------|----------|-----------|-----|----------|
| `internal/api/server/server.go` | 0.0% | 4 | 80.0% | 🔴 **HIGH** |
| `internal/patterns/types.go` | 0.0% | 1 | 80.0% | 🔴 **HIGH** |
| `internal/cli/history.go` | 15.3% | 4 | 64.7% | 🔴 **HIGH** |
| `internal/mcp/server.go` | 33.3% | 6 | 46.7% | 🔴 **HIGH** |
| `internal/cli/audit.go` | 34.2% | 2 | 45.8% | 🔴 **HIGH** |
| `internal/patterns/learner_output.go` | 39.6% | 3 | 40.4% | 🟡 Medium |
| `internal/cli/learn.go` | 40.0% | 1 | 40.0% | 🟡 Medium |
| `internal/cli/hooks.go` | 41.6% | 2 | 38.4% | 🟡 Medium |
| `internal/fix/fixer_rollback.go` | 46.7% | 3 | 33.3% | 🟡 Medium |
| `internal/scanner/vibe.go` | 51.7% | 5 | 28.3% | 🟡 Medium |
| `internal/cli/review.go` | 56.0% | 2 | 24.0% | 🟡 Medium |
| `internal/cli/init.go` | 57.5% | 2 | 22.5% | 🟡 Medium |
| `internal/mcp/audit_helper.go` | 61.7% | 4 | 18.3% | 🟡 Medium |
| `internal/scanner/parallel_helpers.go` | 64.1% | 3 | 15.9% | 🟡 Medium |
| `internal/cli/fix.go` | 73.1% | 2 | 6.9% | 🟢 Low |
| `internal/cli/docsync.go` | 73.6% | 3 | 6.4% | 🟢 Low |
| `internal/cli/knowledge.go` | 77.4% | 13 | 2.6% | 🟢 Low |
| `internal/patterns/learner.go` | 78.0% | 1 | 2.0% | 🟢 Low |
| `internal/cli/update.go` | 78.1% | 2 | 1.9% | 🟢 Low |
| `internal/cli/status.go` | 79.6% | 1 | 0.4% | 🟢 Low |

**Total Non-Critical Gaps:** 20 packages

---

## 📊 Summary Statistics

- **Total Packages Analyzed:** 56
- **Packages Meeting Threshold:** 30 (53.6%)
- **Critical Packages Below 90%:** 6 (10.7%)
- **Non-Critical Packages Below 80%:** 20 (35.7%)
- **Overall Coverage:** 69.8% (below 80% threshold)

---

## 🎯 Recommended Action Plan

### Priority 1: Zero Coverage Files (Immediate Action Required)
1. **`internal/repository/database.go`** (0.0%) - Critical package
2. **`internal/api/server/server.go`** (0.0%) - Server initialization
3. **`internal/patterns/types.go`** (0.0%) - Type definitions

### Priority 2: Critical Packages Below 90%
1. **`internal/services/user_service.go`** (76.8% → 90%) - Needs 13.2% increase
2. **`internal/repository/user_repository.go`** (84.5% → 90%) - Needs 5.5% increase
3. **`internal/api/handlers/user_handler.go`** (87.9% → 90%) - Needs 2.1% increase

### Priority 3: High Gap Non-Critical Packages
1. **`internal/cli/history.go`** (15.3% → 80%) - Needs 64.7% increase
2. **`internal/mcp/server.go`** (33.3% → 80%) - Needs 46.7% increase
3. **`internal/cli/audit.go`** (34.2% → 80%) - Needs 45.8% increase

---

## 📝 Notes

- Critical packages are defined as those containing: `handlers`, `services`, `repository`, or `models` in their path
- Coverage is calculated as average function coverage per package file
- Overall project coverage of 69.8% is below the 80% minimum threshold
- Focus should be on critical packages first to meet the 90% threshold requirement

# Standards Violations Report (File Size Limits)

This report lists Go files exceeding CODING_STANDARDS.md size limits. Each entry includes a micro-task suggestion for splitting or refactoring. The list is derived from current line counts.

## Violations

| File | Lines | Likely Type | Limit | Suggested Micro-Task |
|---|---:|---|---:|---|
| hub/api/task_handler.go | 827 | Handler | 300 | Split into domain-specific handler files (create/list/update/delete + sub-features). |
| hub/api/feature_discovery/database_schema.go | 811 | Service/Analyzer | 400 | Split by concerns: schema parsing, relationship detection, reporting. |
| hub/api/feature_discovery/ui_components.go | 627 | Service/Analyzer | 400 | Split by framework/component detection modules. |
| hub/api/feature_discovery/api_endpoints.go | 621 | Service/Analyzer | 400 | Split endpoint discovery vs reporting. |
| hub/api/architecture_analyzer.go | 614 | Service/Analyzer | 400 | Split analysis phases (size detection, section mapping, recommendations). |
| hub/api/services/dependency_detector.go | 755 | Service | 400 | Split graph builder, analyzer, and report formatting. |
| hub/api/services/intent_analyzer.go | 704 | Service | 400 | Split parsing, classification, and response construction. |
| hub/api/services/mutation_engine.go | 668 | Service | 400 | Split mutation operators vs execution runner. |
| hub/api/services/gap_analyzer.go | 660 | Service | 400 | Split rules extraction vs code comparison. |
| hub/api/services/llm_cache.go | 638 | Service/Utils | 400 | Split cache core vs eviction policy vs metrics. |
| hub/api/utils/flow_verifier.go | 622 | Utility | 250 | Split verification helpers by domain. |

## Notes
- Some files appear duplicated between root `hub/api/*.go` and `hub/api/services/*.go`. These should be reviewed for consolidation and to avoid parallel implementations.
- For each file above, split into 2–4 files that match the handler/service/util size limits.

# Architecture Analyzer – Full Implementation Plan

## 1. End-to-End Analysis Summary

### 1.1 Current State

| Component | Location | Status | Notes |
|-----------|----------|--------|--------|
| **Entry point** | None | **No API** | `analyzeArchitecture` is never called by any handler or service |
| **Implementations** | `hub/api/architecture_analyzer.go` (package main), `hub/api/services/architecture_*.go` | **Duplicated** | Same flow in both; services uses `getParser` (ast_bridge) |
| **ModuleGraph** | `architecture_types.go` / `architecture_analyzer.go` | **Nodes only** | `Nodes` populated (path, lines, type); `Edges` always empty |
| **detectDependencyIssues** | `architecture_deps.go` / `architecture_analyzer.go` | **Stub** | Only “god_module” (file &gt; 1000 lines); `graph` param unused |
| **AST dependency extraction** | `ast/dependency_graph.go` | **Implemented** | `ExtractDependenciesFromFile`, `extractGoImport` / `extractJSImport` / `extractPythonImport`, `DependencyGraph`, `FindCircularDependencies` |
| **AST cross-file** | `ast/cross_file.go` | **Uses deps** | Builds `DependencyGraph`, runs `FindCircularDependencies`; returns `CircularDeps` |
| **Config** | Hardcoded | **No config** | Thresholds (300 / 500 / 1000 lines) and “config in full impl” only in comments |

### 1.2 Data Flow (Current)

```
[No HTTP route]
    ↓
analyzeArchitecture(files []FileContent)
    ↓
├── Per file: line count → oversized check, sections, split suggestion
├── Per file: ModuleNode(path, lines, type) → ModuleGraph.Nodes
├── ModuleGraph.Edges: never populated
├── detectDependencyIssues(files, graph) → only god_module from file size
└── generateRecommendations(oversized, issues)
    ↓
ArchitectureAnalysisResponse{ OversizedFiles, ModuleGraph, DependencyIssues, Recommendations }
```

### 1.3 Gaps for “Full” Implementation

1. **ModuleGraph edges**
   - Edges are never built. `ModuleEdge` has `From`, `To`, `Type` (import / extends / implements).
   - Need to extract dependencies (reuse AST) and populate `ModuleGraph.Edges`.

2. **Import path → file path resolution**
   - AST returns `Dependency{FromFile, ToFile, Type}`. `FromFile` = file path; `ToFile` = import target:
     - **Go:** package path (e.g. `sentinel-hub-api/services`), not file path.
     - **JS/TS:** string from `import` / `require` (often relative, e.g. `./utils`).
     - **Python:** `dotted_name` or `relative_import` (e.g. `package.sub` or `.local`).
   - Cycle detection and graph analysis require nodes to be **file paths** (or a stable module ID). We must **resolve** `ToFile` to files in the analyzed set (or explicitly exclude externals).

3. **Use of `graph` in `detectDependencyIssues`**
   - Today: only file-size-based “god_module”; `graph` ignored.
   - Target:
     - **circular:** Detect cycles from `ModuleGraph` (or equivalent).
     - **tight_coupling:** Use graph metrics (e.g. fan-out, layering).
     - **god_module:** Keep size check; optionally enrich using graph.

4. **Unification**
   - Single implementation (prefer **services**), shared by any future handler. Remove or thin duplication in `package main`.

5. **API and config**
   - Optional: expose architecture analysis via HTTP.
   - Config: file-size thresholds and coupling rules (e.g. max fan-out, layering).

---

## 2. Existing Building Blocks

### 2.1 AST Package (`hub/api/ast`)

- **Parsers:** `GetParser(lang)` / `createParserForLanguage(lang)` for `go`, `javascript`, `typescript`, `python` (and aliases).
- **Dependency extraction:**  
  `ExtractDependenciesFromFile(rootNode, code, filePath, language)` → `[]*Dependency`.  
  Uses `extractGoImport`, `extractJSImport`, `extractPythonImport` (tree-sitter).
- **DependencyGraph:**  
  `AddDependency`, `GetDependencies`, `GetDependents`, `FindCircularDependencies` (DFS cycle detection).
- **Cross-file:**  
  `AnalyzeCrossFile` builds symbol table + `DependencyGraph`, runs `FindCircularDependencies`, returns `CircularDeps` among others.

### 2.2 Architecture Package (`hub/api/services`)

- **Types:** `FileContent`, `ArchitectureAnalysisRequest/Response`, `ModuleGraph`, `ModuleNode`, `ModuleEdge`, `DependencyIssue`, etc.
- **Flow:** `analyzeArchitecture` → sections, split suggestions, `ModuleGraph` (nodes only), `detectDependencyIssues`, `generateRecommendations`.

### 2.3 Relation to AST Cross-File

- **AST cross-file:** Symbol-level (unused exports, undefined refs, circular deps). Uses `DependencyGraph` with `FromFile`/`ToFile` as stored.
- **Architecture:** Coarser, file/module-level (oversized files, module graph, dependency issues). `ModuleGraph` is the equivalent structure.
- **Overlap:** Both care about dependencies and cycles. Architecture should leverage the same extraction and cycle logic, but work at file/module granularity and feed `ModuleGraph` + `DependencyIssue`.

---

## 3. Import Path → File Path Resolution

### 3.1 Requirements

- **Input:** `files []FileContent` (path, content, language). No explicit “project root” today; we can infer from paths or add it to the request.
- **Output:** For each `Dependency(FromFile, ToFile, Type)`, decide:
  - `ToFile` resolves to one or more **files in `files`** → add edges **FromFile → resolved path(s)**.
  - `ToFile` is external (stdlib, third-party) → either skip or add “external” nodes, depending on product choice.

### 3.2 Per-Language Rules

| Language | `ToFile` from AST | Resolution strategy |
|----------|-------------------|----------------------|
| **Go** | Package path (e.g. `sentinel-hub-api/services`) | Map package path → package dir → all `*.go` in that dir present in `files`. If multiple files, we can add an edge From → package (e.g. first file or a virtual “package” node) to keep cycles meaningful. **Simpler:** Map package path → **directory path**; treat “package” as node; cycle = A’s package imports B’s package, B’s imports A’s. |
| **JS/TS** | String from `import`/`require` (e.g. `./utils`, `../lib`) | Resolve relative to `filepath.Dir(FromFile)`. Support `.js`/`.ts`/`/index` conventions. Match resolved path to `files[].Path` (normalize slashes, symlinks if needed). |
| **Python** | `dotted_name` or `relative_import` | Similar to JS: resolve relative to `FromFile`’s dir; map `package.sub` to `package/sub.py` or `package/sub/__init__.py` within `files`. |

### 3.3 Practical Scope for V1

- **Internal only:** Resolve only targets that appear in `files` (or their directories). Ignore externals for cycle detection.
- **Optional:** Add `CodebasePath` or `ProjectRoot` to `ArchitectureAnalysisRequest` to improve resolution (e.g. Go module root).

---

## 4. Detailed Implementation Plan

### Phase 1: Unify and Wire Architecture Pipeline (Week 1)

**Goals:** Single implementation, clear ownership, no dead code.

| Task | Description | Files |
|------|-------------|--------|
| 1.1 | **Choose single implementation** | Prefer **services** (`architecture_*.go`) as source of truth. |
| 1.2 | **Remove or redirect duplication** | `hub/api/architecture_analyzer.go`: either delete and have `package main` call services, or reduce to a thin wrapper around services. Ensure no two implementations of `analyzeArchitecture` / `detectDependencyIssues`. |
| 1.3 | **Clarify request shape** | Keep `ArchitectureAnalysisRequest.Files`. Optionally add `CodebasePath` / `ProjectRoot` (string) for resolution. |
| 1.4 | **Add architecture API (optional)** | If product wants it: `POST /api/v1/analyze/architecture` → decode `ArchitectureAnalysisRequest` → call `analyzeArchitecture` → return `ArchitectureAnalysisResponse`. Handler in `handlers/`, route in `router`. |

**Deliverables:** One clear architecture flow; optional HTTP API; no duplicate logic.

---

### Phase 2: Build ModuleGraph Edges (Week 2)

**Goals:** Populate `ModuleGraph.Edges` from AST dependency extraction and resolution.

| Task | Description | Files |
|------|-------------|--------|
| 2.1 | **Import extraction helper** | In services (or a shared `architecture_deps` helper): for each `FileContent`, get parser, parse, call `ast.ExtractDependenciesFromFile`, collect `[]*ast.Dependency`. Reuse existing AST helpers; stay within file-size / complexity limits. |
| 2.2 | **Resolver interface** | `type ImportResolver interface { Resolve(fromPath, toImport, language string) ([]string, error) }` → returns resolved file paths in the analyzed set (or package dirs if we use package-level nodes). |
| 2.3 | **Resolver implementations** | **Go:** package path → dir; match `files` under that dir; return those paths (or representative). **JS/TS:** resolve relative path from `fromPath`, normalize, match `files`. **Python:** same idea. Skip unresolvable (e.g. externals) and optionally log. |
| 2.4 | **Edge building** | After building `ModuleGraph.Nodes` (as today): for each file, run extraction → resolver → for each resolved target, append `ModuleEdge{From: file.Path, To: target, Type: "import"}` (or `"require"` if we distinguish). Deduplicate edges. |
| 2.5 | **Wire into `analyzeArchitecture`** | Current loop: build nodes. Add second pass (or extend loop): extract deps, resolve, append to `ModuleGraph.Edges`. |

**Deliverables:** `ModuleGraph.Edges` populated for internal imports; resolver clearly separated and testable.

---

### Phase 3: Use Graph in `detectDependencyIssues` (Week 3)

**Goals:** Implement **circular** and **tight_coupling**; keep **god_module**; use `graph` meaningfully.

| Task | Description | Files |
|------|-------------|--------|
| 3.1 | **Cycle detection on ModuleGraph** | Either (a) implement DFS cycle detection on `ModuleGraph` (Nodes + Edges), or (b) build `ast.DependencyGraph` from `ModuleGraph` and call `FindCircularDependencies`, then map cycles back to `DependencyIssue`. Prefer (a) to avoid maintaining two graphs. |
| 3.2 | **Emit `circular` issues** | For each cycle (list of file paths): append `DependencyIssue{Type: "circular", Severity: "high", Files: <cycle>, Description: "...", Suggestion: "Break cycle by introducing interface or moving code."}`. |
| 3.3 | **Tight-coupling heuristics** | Use `ModuleGraph.Edges`: e.g. fan-out &gt; N (e.g. 15) ⇒ “too many dependencies”; or simple layering (e.g. `detectModuleType`: service vs handler vs util) and flag lower layers importing higher layers. Emit `DependencyIssue{Type: "tight_coupling", ...}`. |
| 3.4 | **God-module** | Keep current size-based check (e.g. &gt; 1000 lines). Optionally use in-degree / out-degree from graph to strengthen “god module” detection. |
| 3.5 | **Remove stub** | Delete `_ = graph`; `detectDependencyIssues` now uses `graph` for circular and coupling. |

**Deliverables:** `detectDependencyIssues` fully implemented; all three issue types supported; `graph` required and used.

---

### Phase 4: Config and Cleanup (Week 4)

**Goals:** Configurable thresholds; codebase cleanup; tests.

| Task | Description | Files |
|------|-------------|--------|
| 4.1 | **Config for thresholds** | Add (e.g. in `pkg/config` or service-level config): `WarningLines`, `CriticalLines`, `MaxLines` (defaults 300 / 500 / 1000), `MaxFanOut` (tight_coupling), etc. `analyzeArchitecture` (or a config holder) reads these. |
| 4.2 | **Replace hardcoded values** | Use config in oversized checks and in tight-coupling logic. |
| 4.3 | **Tests** | Unit tests: extraction → resolution → edge building; cycle detection; `detectDependencyIssues` (circular, tight_coupling, god_module). Integration test: `analyzeArchitecture` with a small multi-file setup (e.g. Go) that has a cycle. |
| 4.4 | **CODING_STANDARDS** | Ensure file size, function count, error wrapping (`fmt.Errorf` / `%w`), context where relevant. No hardcoded secrets. |

**Deliverables:** Config-driven behavior; tests; standards compliance.

---

## 5. Technical Notes

### 5.1 Go Package Path → File Mapping

- Go imports use **package path**, not file path. Multiple files per package.
- Options:
  - **A.** One node per **file**; edge From→To means “file A imports package containing file B”. Resolution: package path → dir → include all files in `files` under that dir; add A→B for each such B (or A→“pkg” and treat “pkg” as logical node).
  - **B.** One node per **package** (dir); edges between packages. Cycle = package-level cycle.
- Recommendation: **B** for simplicity. Represent “package” by package dir path; aggregation of `ModuleGraph.Nodes` by dir yields package-level graph; resolution maps import path → dir.

### 5.2 JS/TS Relative Imports

- Resolve `./x`, `../y` relative to `filepath.Dir(FromFile)`.
- Consider `extension` fallbacks (`.js`, `.ts`, `index.js`) and `package.json` exports later; V1 can use straightforward path matching against `files[].Path`.

### 5.3 Python

- `from pkg import x` → `pkg` or `pkg.x`; map to `pkg/__init__.py` or `pkg/x.py` within `files`. Relative imports (`.`, `..`) same as JS.

### 5.4 Cycle Detection

- Input: `ModuleGraph` (Nodes + Edges). Run DFS from each node; track `recStack`; on back edge, emit cycle. Deduplicate cycles (same set of nodes). Reuse same conceptual algorithm as `ast.DependencyGraph.FindCircularDependencies`.

### 5.5 Layering (Optional)

- `detectModuleType` already tags nodes (component, service, utility, test, module). Optional rules: e.g. “handler must not import service” (if we add handler), “util must not import service”. Can be added in Phase 3 or 4.

---

## 6. File-Level Plan

| File | Action |
|------|--------|
| `hub/api/architecture_analyzer.go` | Remove duplicate logic; call services or delete. |
| `hub/api/services/architecture_analysis.go` | Extend to build edges (Phase 2); keep as main orchestrator. |
| `hub/api/services/architecture_deps.go` | Implement full `detectDependencyIssues` (Phase 3); add cycle + coupling. |
| `hub/api/services/architecture_types.go` | No structural change; keep `ModuleGraph` etc. |
| **New** `hub/api/services/architecture_resolver.go` | Import path → file path resolution (Phase 2). |
| **New** `hub/api/services/architecture_edges.go` | Edge building from AST deps + resolver (Phase 2). |
| **New** `hub/api/services/architecture_cycles.go` | Cycle detection on `ModuleGraph` (Phase 3). |
| `hub/api/ast/dependency_graph.go` | Keep as-is; use only `ExtractDependenciesFromFile` (and optionally `Dependency` type). |
| Config | Add architecture-related keys (Phase 4). |
| Handlers / router | Optional architecture endpoint (Phase 1). |

---

## 7. Success Criteria

- [ ] Single implementation path for architecture analysis (services).
- [ ] `ModuleGraph.Edges` populated from AST extraction + resolution.
- [ ] `detectDependencyIssues` uses `graph`; implements **circular**, **tight_coupling**, **god_module**.
- [ ] Configurable thresholds; no hardcoded magic numbers for sizing/coupling.
- [ ] Unit tests for resolver, edge building, cycle detection, `detectDependencyIssues`.
- [ ] Integration test with a known cycle.
- [ ] Compliance with `docs/external/CODING_STANDARDS.md` and existing patterns.

---

## 8. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Resolution too complex (especially Go) | Start with package-as-node and dir-based resolution; defer multi-module and vanity imports. |
| Cross-file vs architecture overlap | Keep AST cross-file for symbol-level; architecture for file/module-level. Share extraction, not necessarily the graph structure. |
| Perf with large `files` | Extract deps in a single pass; resolve in batches; optional limits on files per request. |

---

## 9. Dependency Graph Reference

```
ArchitectureAnalysisRequest (Files)
         ↓
analyzeArchitecture
         ↓
    ┌────┴────┐
    ↓         ↓
 Nodes     (new) Extract deps per file (ast.ExtractDependenciesFromFile)
    ↓         ↓
    │     Resolve ToFile → file paths (new Resolver)
    │         ↓
    │     Build ModuleGraph.Edges (new)
    │         ↓
    └────┬────┘
         ↓
detectDependencyIssues(graph)
         ↓
    ├── circular: cycle detection on graph
    ├── tight_coupling: fan-out / layering
    └── god_module: size (+ optional graph metrics)
         ↓
generateRecommendations
         ↓
ArchitectureAnalysisResponse
```

---

*Document generated from end-to-end analysis of the architecture analyzer and dependency detection implementation.*

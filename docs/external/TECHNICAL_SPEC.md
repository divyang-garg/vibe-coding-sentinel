# Technical Specification

> **For AI Agents**: This document provides detailed technical specifications for Sentinel implementation. Follow these specifications exactly when implementing or extending functionality.

## Agent Specification

### Binary Details

| Property | Value |
|----------|-------|
| Language | Go 1.21+ |
| Build | `go build -ldflags="-s -w"` |
| Output | `sentinel` (Unix) / `sentinel.exe` (Windows) |
| Size | ~10-15 MB |
| Dependencies | None (pure Go) |

### Command Structure

```go
func main() {
    switch os.Args[1] {
    case "init":       runInit(os.Args[2:])
    case "audit":      runAudit(os.Args[2:])
    case "learn":      runLearn(os.Args[2:])
    case "fix":        runFix(os.Args[2:])
    case "status":     runStatus(os.Args[2:])
    case "ingest":     runIngest(os.Args[2:])
    case "review":     runReview(os.Args[2:])
    case "baseline":   runBaseline(os.Args[2:])
    case "history":    runHistory(os.Args[2:])
    case "mcp-server": runMCPServer()
    // ... other commands
    }
}
```

---

## Data Types

### Core Types

```go
// Configuration
type Config struct {
    ScanDirs       []string            `json:"scanDirs"`
    ExcludePaths   []string            `json:"excludePaths"`
    SeverityLevels map[string]string   `json:"severityLevels"`
    CustomPatterns map[string]string   `json:"customPatterns"`
    RuleLocations  []string            `json:"ruleLocations"`
    Ingest         IngestConfig        `json:"ingest"`
    Telemetry      TelemetryConfig     `json:"telemetry"`
}

type IngestConfig struct {
    LLMProvider   string `json:"llmProvider"`
    LocalOnly     bool   `json:"localOnly"`
    VisionEnabled bool   `json:"visionEnabled"`
}

type TelemetryConfig struct {
    Enabled  bool   `json:"enabled"`
    Endpoint string `json:"endpoint"`
    OrgID    string `json:"orgId"`
    APIKey   string `json:"apiKey"`
}
```

### Audit Types

```go
type Finding struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Pattern  string `json:"pattern"`
    Code     string `json:"code"`
}

type HookContext struct {
    HookType      string   `json:"hookType"`      // "pre-commit" | "pre-push"
    UserActions   []string `json:"userActions"`   // ["viewed", "baselined", "proceeded"]
    OverrideReason string  `json:"overrideReason,omitempty"`
    DurationMs    int64    `json:"durationMs"`
}

type CheckResult struct {
    Enabled  bool   `json:"enabled"`
    Success  bool   `json:"success"`
    Error    string `json:"error,omitempty"`
    Findings int    `json:"findings"`
}

type AuditReport struct {
    Timestamp   string    `json:"timestamp"`
    Status      string    `json:"status"`
    Directories []string  `json:"directories"`
    Findings    []Finding `json:"findings"`
    Summary     struct {
        Total    int `json:"total"`
        Critical int `json:"critical"`
        Warning  int `json:"warning"`
        Info     int `json:"info"`
    } `json:"summary"`
    HookContext *HookContext          `json:"hookContext,omitempty"`
    CheckResults map[string]CheckResult `json:"checkResults,omitempty"`
}
```

### Pattern Types

```go
type ProjectPatterns struct {
    Naming     NamingPatterns    `json:"naming"`
    Imports    ImportPatterns    `json:"imports"`
    Structure  StructurePatterns `json:"structure"`
    LearnedAt  string            `json:"learnedAt"`
    FileCount  int               `json:"fileCount"`
}

type NamingPatterns struct {
    Functions  string  `json:"functions"`  // camelCase, snake_case, PascalCase
    Variables  string  `json:"variables"`
    Files      string  `json:"files"`
    Classes    string  `json:"classes"`
    Confidence float64 `json:"confidence"`
}

type ImportPatterns struct {
    Style    string   `json:"style"`    // absolute, relative
    Prefix   string   `json:"prefix"`   // @/, ~/, etc.
    Grouping []string `json:"grouping"` // ["external", "internal", "relative"]
}

type StructurePatterns struct {
    SourceRoot       string            `json:"sourceRoot"`
    TestPattern      string            `json:"testPattern"`
    ComponentPattern string            `json:"componentPattern"`
    FolderMap        map[string]string `json:"folderMap"`
}
```

### Fix Types

```go
type Fix struct {
    ID          string   `json:"id"`
    Pattern     string   `json:"pattern"`
    Replacement string   `json:"replacement"`
    Description string   `json:"description"`
    SafeLevel   string   `json:"safeLevel"` // safe, prompted, manual
    Languages   []string `json:"languages"`
}

type FixResult struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Original string `json:"original"`
    Fixed    string `json:"fixed"`
    Status   string `json:"status"` // applied, skipped, failed
    FixID    string `json:"fixId"`
}

type FixSession struct {
    Timestamp string      `json:"timestamp"`
    BackupDir string      `json:"backupDir"`
    Results   []FixResult `json:"results"`
}
```

### Document Ingestion Types

```go
type Document struct {
    Path      string    `json:"path"`
    Type      string    `json:"type"` // pdf, docx, xlsx, image, eml, txt
    Size      int64     `json:"size"`
    ParsedAt  time.Time `json:"parsedAt"`
    TextPath  string    `json:"textPath"`  // path to extracted text
    Checksum  string    `json:"checksum"`
}

type ExtractedKnowledge struct {
    Entities  []Entity       `json:"entities"`
    Rules     []BusinessRule `json:"rules"`
    Journeys  []UserJourney  `json:"journeys"`
    Objectives []Objective   `json:"objectives"`
    SourceDoc string         `json:"sourceDoc"`
}

type Entity struct {
    Name         string            `json:"name"`
    Definition   string            `json:"definition"`
    Attributes   []string          `json:"attributes"`
    Relationships []Relationship   `json:"relationships"`
    Source       string            `json:"source"`
    Confidence   float64           `json:"confidence"`
    Status       string            `json:"status"` // draft, approved, rejected
}

type BusinessRule struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    Entities     []string `json:"entities"`
    Conditions   []string `json:"conditions"`
    Exceptions   []string `json:"exceptions"`
    Consequences []string `json:"consequences"`
    Source       string   `json:"source"`
    Confidence   float64  `json:"confidence"`
    Status       string   `json:"status"`
}

type UserJourney struct {
    Name          string        `json:"name"`
    UserType      string        `json:"userType"`
    Goal          string        `json:"goal"`
    Preconditions []string      `json:"preconditions"`
    Steps         []JourneyStep `json:"steps"`
    Outcomes      []string      `json:"outcomes"`
    Source        string        `json:"source"`
    Confidence    float64       `json:"confidence"`
    Status        string        `json:"status"`
}

type ReviewStatus struct {
    File       string    `json:"file"`
    TotalItems int       `json:"totalItems"`
    Accepted   int       `json:"accepted"`
    Edited     int       `json:"edited"`
    Rejected   int       `json:"rejected"`
    Pending    int       `json:"pending"`
    ReviewedAt time.Time `json:"reviewedAt"`
    ReviewedBy string    `json:"reviewedBy"`
}
```

### Vibe Coding Analysis Types

```go
// AST Analysis Request (sent to Hub)
type ASTAnalysisRequest struct {
    Code       string   `json:"code"`
    Language   string   `json:"language"`
    Filename   string   `json:"filename"`
    ProjectID  string   `json:"projectId"`
    Analyses   []string `json:"analyses"` // duplicates, unused, unreachable, security
}

// AST Analysis Response
type ASTAnalysisResponse struct {
    Success  bool            `json:"success"`
    Findings []ASTFinding    `json:"findings"`
    Stats    AnalysisStats   `json:"stats"`
}

type ASTFinding struct {
    Type       string   `json:"type"`       // duplicate_function, unused_variable, etc.
    Severity   string   `json:"severity"`
    Line       int      `json:"line"`
    Column     int      `json:"column"`
    EndLine    int      `json:"endLine"`
    EndColumn  int      `json:"endColumn"`
    Message    string   `json:"message"`
    Code       string   `json:"code"`       // Code snippet
    Suggestion string   `json:"suggestion"`
    AutoFix    *AutoFix `json:"autoFix,omitempty"`
}

type AutoFix struct {
    Available bool   `json:"available"`
    Code      string `json:"code"`
    RiskLevel string `json:"riskLevel"` // safe, medium, high
}

type AnalysisStats struct {
    ParseTime    int64 `json:"parseTimeMs"`
    AnalysisTime int64 `json:"analysisTimeMs"`
    NodesVisited int   `json:"nodesVisited"`
}

// Vibe Issue Types
const (
    VibeDuplicateFunction   = "duplicate_function"
    VibeOrphanedCode        = "orphaned_code"
    VibeUnusedVariable      = "unused_variable"
    VibeSignatureMismatch   = "signature_mismatch"
    VibeEmptyCatch          = "empty_catch"
    VibeCodeAfterReturn     = "code_after_return"
    VibeMissingAwait        = "missing_await"
    VibeBraceMismatch       = "brace_mismatch"
)
```

### Security Rules Types

```go
// Security Rule Definition
type SecurityRule struct {
    ID          string         `json:"id"`          // SEC-XXX
    Version     string         `json:"version"`
    Status      string         `json:"status"`      // active, deprecated
    Name        string         `json:"name"`
    Type        string         `json:"type"`        // authorization, authentication, injection, etc.
    Severity    string         `json:"severity"`    // critical, high, medium, low
    Description string         `json:"description"`
    Detection   SecurityDetect `json:"detection"`
    ASTCheck    *ASTSecCheck   `json:"astCheck,omitempty"`
    AutoFix     *SecurityFix   `json:"autoFix,omitempty"`
    TestReqs    []TestReq      `json:"testRequirements"`
}

type SecurityDetect struct {
    Endpoints         []string `json:"endpoints,omitempty"`
    Resources         []string `json:"resources,omitempty"`
    RequiredChecks    []string `json:"requiredChecks,omitempty"`
    PatternsForbidden []string `json:"patternsForbidden,omitempty"`
    PatternsRequired  []string `json:"patternsRequired,omitempty"`
}

type ASTSecCheck struct {
    FunctionContains     []string `json:"functionContains,omitempty"`
    MustHaveBefore       string   `json:"mustHaveBeforeResponse,omitempty"`
    RouteMiddleware      []string `json:"routeMiddleware,omitempty"`
}

type SecurityFix struct {
    Available    bool   `json:"available"`
    InsertBefore string `json:"insertBefore,omitempty"`
    InsertAfter  string `json:"insertAfter,omitempty"`
    Replace      string `json:"replace,omitempty"`
}

// Security Analysis Request
type SecurityAnalysisRequest struct {
    Code            string            `json:"code"`
    Language        string            `json:"language"`
    Filename        string            `json:"filename"`
    ProjectID       string            `json:"projectId"`
    Rules           []string          `json:"rules,omitempty"`           // Specific rules to check (SEC-001, etc.)
    ExpectedFindings map[string]bool  `json:"expectedFindings,omitempty"` // Ground truth for detection rate validation (ruleID -> shouldDetect)
}

// Security Analysis Response
type SecurityAnalysisResponse struct {
    Score    int               `json:"score"`   // 0-100
    Grade    string            `json:"grade"`   // A, B, C, D, F
    Findings []SecurityFinding `json:"findings"`
    Summary  SecuritySummary   `json:"summary"`
    Metrics  *DetectionMetrics `json:"metrics,omitempty"` // Optional: only for validation runs with ground truth
}

// DetectionMetrics tracks detection rate validation metrics
// Only included when expectedFindings is provided in SecurityAnalysisRequest
type DetectionMetrics struct {
    TruePositives  int     `json:"truePositives"`  // Correctly detected vulnerabilities
    FalsePositives int     `json:"falsePositives"` // Incorrectly flagged as vulnerabilities
    FalseNegatives int     `json:"falseNegatives"` // Missed vulnerabilities
    TrueNegatives  int     `json:"trueNegatives"`  // Correctly identified as safe
    DetectionRate  float64 `json:"detectionRate"`  // Overall accuracy percentage: (TP + TN) / Total * 100
    Precision      float64 `json:"precision"`      // Accuracy of positive predictions: TP / (TP + FP) * 100
    Recall         float64 `json:"recall"`         // Coverage of actual vulnerabilities: TP / (TP + FN) * 100
}

type SecurityFinding struct {
    RuleID      string `json:"ruleId"`
    RuleName    string `json:"ruleName"`
    Severity    string `json:"severity"`
    Line        int    `json:"line"`
    Code        string `json:"code"`
    Issue       string `json:"issue"`
    Remediation string `json:"remediation"`
    AutoFixable bool   `json:"autoFixable"`
    AutoFix     string `json:"autoFix,omitempty"`
}

type SecuritySummary struct {
    TotalRules  int `json:"totalRules"`
    Passed      int `json:"passed"`
    Failed      int `json:"failed"`
    Critical    int `json:"critical"`
    High        int `json:"high"`
    Medium      int `json:"medium"`
    Low         int `json:"low"`
}
```

### Test Enforcement Types (Phase 10) ✅ IMPLEMENTED

**Status**: All Phase 10 features implemented and tested.

**API Endpoints**:
- `POST /api/v1/test-requirements/generate` - Generate test requirements from business rules
- `POST /api/v1/test-coverage/analyze` - Analyze test coverage (accepts test file content)
- `GET /api/v1/test-coverage/{knowledge_item_id}` - Get coverage for a knowledge item
- `POST /api/v1/test-validations/validate` - Validate test correctness
- `GET /api/v1/test-validations/{test_requirement_id}` - Get validation results
- `POST /api/v1/mutation-test/run` - Run mutation testing
- `GET /api/v1/mutation-test/{test_requirement_id}` - Get mutation test results
- `POST /api/v1/test-execution/run` - Execute tests in sandbox
- `GET /api/v1/test-execution/{execution_id}` - Get execution status

**Database Tables**:
- `test_requirements` - Generated test requirements
- `test_coverage` - Coverage tracking per business rule
- `test_validations` - Test validation results
- `mutation_results` - Mutation testing results
- `test_executions` - Test execution records

**Agent Commands**:
- `sentinel test --requirements` - Generate test requirements
- `sentinel test --coverage` - Analyze test coverage
- `sentinel test --validate` - Validate tests
- `sentinel test --mutation --source <file> --test <file>` - Run mutation testing
- `sentinel test --run --language <lang>` - Execute tests in sandbox

### Test Enforcement Types

```go
// Test Requirements from Business Rules
type TestRequirement struct {
    ID          string            `json:"id"`          // BR-001-T1
    RuleID      string            `json:"ruleId"`      // BR-001
    Name        string            `json:"name"`        // test_cancel_within_24h
    Type        string            `json:"type"`        // happy_path, error_case, edge_case, exception_case
    Priority    string            `json:"priority"`    // critical, high, medium, low
    Scenario    string            `json:"scenario"`
    Setup       TestSetup         `json:"setup"`
    Action      string            `json:"action"`
    Expected    TestExpected      `json:"expected"`
    Assertions  []string          `json:"assertionsRequired"`
}

type TestSetup struct {
    Entities map[string]interface{} `json:"entities"`
    State    map[string]interface{} `json:"state,omitempty"`
}

type TestExpected struct {
    Success     bool                   `json:"success,omitempty"`
    ReturnValue map[string]interface{} `json:"returnValue,omitempty"`
    SideEffects []string               `json:"sideEffects,omitempty"`
    Error       string                 `json:"error,omitempty"`
}

// Test Coverage Report
type TestCoverageReport struct {
    RuleCoverage   map[string]RuleCoverage `json:"ruleCoverage"`
    LineCoverage   float64                 `json:"lineCoverage"`
    BranchCoverage float64                 `json:"branchCoverage"`
    OverallScore   float64                 `json:"overallScore"`
}

type RuleCoverage struct {
    RuleID        string   `json:"ruleId"`
    RequiredTests int      `json:"requiredTests"`
    WrittenTests  int      `json:"writtenTests"`
    PassingTests  int      `json:"passingTests"`
    MissingTests  []string `json:"missingTests"`
    Coverage      float64  `json:"coverage"`
}

// Test Validation Response
type TestValidationResponse struct {
    Valid          bool            `json:"valid"`
    Coverage       float64         `json:"coverage"`
    MutationScore  float64         `json:"mutationScore"`
    Quality        TestQuality     `json:"quality"`
    WeakTests      []WeakTest      `json:"weakTests"`
    Suggestions    []string        `json:"suggestions"`
}

type TestQuality struct {
    HasSetup        bool    `json:"hasSetup"`
    HasTeardown     bool    `json:"hasTeardown"`
    AssertionCount  int     `json:"assertionCount"`
    MutationScore   float64 `json:"mutationScore"`
    EffectivenessScore float64 `json:"effectivenessScore"`
}

type WeakTest struct {
    TestName       string   `json:"testName"`
    Issue          string   `json:"issue"`
    SurvivedMutants []string `json:"survivedMutants"`
    Suggestion     string   `json:"suggestion"`
}

// Test Enforcement Config
type TestEnforcementConfig struct {
    Mode     string              `json:"mode"`     // strict, recommended, off
    Rules    TestEnforcementRules `json:"rules"`
    Blocking BlockingConfig      `json:"blocking"`
}

type TestEnforcementRules struct {
    MinimumCoverage MinCoverage `json:"minimumCoverage"`
    RequiredTypes   RequiredTestTypes `json:"requiredTestTypes"`
    TestQuality     TestQualityReqs `json:"testQuality"`
}

type MinCoverage struct {
    Line   int `json:"line"`
    Branch int `json:"branch"`
    Rule   int `json:"rule"`
}

type RequiredTestTypes struct {
    BusinessRules []string `json:"businessRules"` // happy_path, error_case
    APIEndpoints  []string `json:"apiEndpoints"`
    SecurityRules []string `json:"securityRules"`
}

type TestQualityReqs struct {
    MinAssertions   int `json:"minAssertionsPerTest"`
    MinMutationScore int `json:"minMutationScore"`
}

type BlockingConfig struct {
    PRMerge    bool `json:"prMerge"`
    Commit     bool `json:"commit"`
    Deployment bool `json:"deployment"`
}
```

### File Size Management Types

```go
// File Size Configuration
type FileSizeConfig struct {
    Thresholds   FileSizeThresholds       `json:"thresholds"`
    ByFileType   map[string]int           `json:"byFileType"`
    Exceptions   []string                 `json:"exceptions"`
}

type FileSizeThresholds struct {
    Warning  int `json:"warning"`  // Lines
    Critical int `json:"critical"`
    Maximum  int `json:"maximum"`
}

// File Analysis Result
type FileAnalysisResult struct {
    File           string           `json:"file"`
    Lines          int              `json:"lines"`
    Status         string           `json:"status"` // ok, warning, critical, oversized
    Sections       []FileSection    `json:"sections,omitempty"`
    SplitSuggestion *SplitSuggestion `json:"splitSuggestion,omitempty"`
}

type FileSection struct {
    StartLine   int    `json:"startLine"`
    EndLine     int    `json:"endLine"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Lines       int    `json:"lines"`
}

type SplitSuggestion struct {
    Reason                string         `json:"reason"`
    ProposedFiles         []ProposedFile `json:"proposedFiles"`
    MigrationInstructions []string       `json:"migrationInstructions"` // Text instructions only, not executable
    EstimatedEffort       string         `json:"estimatedEffort"`
}

type ProposedFile struct {
    Path     string   `json:"path"`
    Lines    int      `json:"lines"`
    Contents []string `json:"contents"` // Function/class names to move
}

// Architecture Analysis
type ArchitectureAnalysis struct {
    OversizedFiles   []FileAnalysisResult `json:"oversizedFiles"`
    ModuleGraph      ModuleGraph          `json:"moduleGraph"`
    DependencyIssues []DependencyIssue    `json:"dependencyIssues"`
    Recommendations  []string             `json:"recommendations"`
}

type ModuleGraph struct {
    Nodes []ModuleNode `json:"nodes"`
    Edges []ModuleEdge `json:"edges"`
}

type ModuleNode struct {
    Path  string `json:"path"`
    Lines int    `json:"lines"`
    Type  string `json:"type"` // component, service, utility, etc.
}

type ModuleEdge struct {
    From   string `json:"from"`
    To     string `json:"to"`
    Type   string `json:"type"` // import, extends, implements
}

type DependencyIssue struct {
    Type        string   `json:"type"`    // circular, tight_coupling, god_module
    Severity    string   `json:"severity"`
    Files       []string `json:"files"`
    Description string   `json:"description"`
    Suggestion  string   `json:"suggestion"`
}
```

### Comprehensive Analysis Types

```go
// Comprehensive Analysis Request
type ComprehensiveAnalysisRequest struct {
    Feature              string            `json:"feature"`
    Mode                 string            `json:"mode"` // "auto", "manual"
    Files                *FeatureFiles     `json:"files,omitempty"` // Required if mode="manual"
    Depth                string            `json:"depth"` // "surface", "medium", "deep"
    IncludeBusinessContext bool            `json:"includeBusinessContext"`
    ProjectID            string            `json:"projectId"`
    AgentID              string            `json:"agentId"`
}

type FeatureFiles struct {
    UI         []string `json:"ui,omitempty"`
    API        []string `json:"api,omitempty"`
    Database   []string `json:"database,omitempty"`
    Logic      []string `json:"logic,omitempty"`
    Integration []string `json:"integration,omitempty"`
    Tests      []string `json:"tests,omitempty"`
}

// Comprehensive Analysis Response
type ComprehensiveAnalysisResponse struct {
    ValidationID string            `json:"validationId"`
    Feature      string            `json:"feature"`
    Status       string            `json:"status"` // "completed", "failed", "pending"
    HubURL       string            `json:"hubUrl"`
    Summary      AnalysisSummary   `json:"summary"`
    Checklist    []ChecklistItem   `json:"checklist"`
    LayerAnalysis map[string]LayerFindings `json:"layerAnalysis"`
    EndToEndFlows []EndToEndFlow  `json:"endToEndFlows"`
    Error        *AnalysisError   `json:"error,omitempty"`
}

type AnalysisSummary struct {
    TotalFindings int `json:"totalFindings"`
    Critical      int `json:"critical"`
    High          int `json:"high"`
    Medium        int `json:"medium"`
    Low           int `json:"low"`
    LayersAnalyzed int `json:"layersAnalyzed"`
    FlowsVerified  int `json:"flowsVerified"`
}

type ChecklistItem struct {
    ID          string `json:"id"`
    Category    string `json:"category"` // "business", "ui", "api", "database", "logic", "integration", "tests"
    Severity    string `json:"severity"` // "critical", "high", "medium", "low"
    Title       string `json:"title"`
    Description string `json:"description"`
    Location    string `json:"location"` // "file:line"
    Remediation string `json:"remediation"`
    AutoFixable bool   `json:"autoFixable"`
}

type LayerFindings struct {
    Findings int `json:"findings"`
    Critical int `json:"critical"`
    High     int `json:"high"`
    Medium   int `json:"medium"`
    Low      int `json:"low"`
}

type EndToEndFlow struct {
    Flow       string        `json:"flow"`
    Status     string        `json:"status"` // "complete", "broken", "partial"
    Breakpoints []Breakpoint `json:"breakpoints,omitempty"`
}

type Breakpoint struct {
    Layer    string `json:"layer"`
    Location string `json:"location"`
    Issue    string `json:"issue"`
}

type AnalysisError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Fallback string `json:"fallback,omitempty"`
    Details string `json:"details,omitempty"`
}

// LLM Provider Configuration
type LLMProviderConfig struct {
    Type            string                 `json:"type"` // "user-provided", "organization-shared"
    Provider        string                 `json:"provider"` // "openai", "anthropic", "azure"
    APIKey          string                 `json:"apiKey"` // Encrypted in database
    Model           string                 `json:"model"`
    Endpoint        string                 `json:"endpoint"`
    CodexPro        CodexProConfig       `json:"codexPro,omitempty"`
    UsageTracking   UsageTrackingConfig   `json:"usageTracking"`
    CostOptimization CostOptimizationConfig `json:"costOptimization"`
}

type CodexProConfig struct {
    Enabled        bool `json:"enabled"`
    FallbackToAPI  bool `json:"fallbackToAPI"`
}

type UsageTrackingConfig struct {
    Enabled    bool   `json:"enabled"`
    Allocation string `json:"allocation"` // "per-project", "per-user", "none"
}

type CostOptimizationConfig struct {
    Caching          CachingConfig          `json:"caching"`
    ProgressiveDepth ProgressiveDepthConfig  `json:"progressiveDepth"`
    ModelSelection   ModelSelectionConfig   `json:"modelSelection"`
}

type CachingConfig struct {
    Enabled      bool    `json:"enabled"`
    TargetHitRate float64 `json:"targetHitRate"`
}

type ProgressiveDepthConfig struct {
    Enabled              bool `json:"enabled"`
    SkipLLMForPatternMatches bool `json:"skipLLMForPatternMatches"`
}

type ModelSelectionConfig struct {
    Enabled        bool     `json:"enabled"`
    CriticalTasks  []string `json:"criticalTasks"`
    NonCriticalTasks []string `json:"nonCriticalTasks"`
}
```

### Requirements Lifecycle Types

```go
// Change Request
type ChangeRequest struct {
    ID             string         `json:"id"`
    Type           string         `json:"type"`       // new, modification, deprecation
    Status         string         `json:"status"`     // draft, pending_approval, approved, rejected, implemented
    Priority       string         `json:"priority"`
    TargetRule     string         `json:"targetRule"` // BR-XXX
    RequestedBy    string         `json:"requestedBy"`
    RequestedAt    time.Time      `json:"requestedAt"`
    CurrentState   RuleState      `json:"currentState,omitempty"`
    ProposedState  RuleState      `json:"proposedState"`
    Justification  string         `json:"justification"`
    ImpactAnalysis ImpactAnalysis `json:"impactAnalysis"`
    Approval       ApprovalStatus `json:"approval"`
    Implementation ImplStatus     `json:"implementation"`
}

type RuleState struct {
    Summary     string                 `json:"summary"`
    Constraints map[string]interface{} `json:"constraints,omitempty"`
}

type ImpactAnalysis struct {
    AffectedCode  []string `json:"affectedCode"`
    AffectedTests []string `json:"affectedTests"`
    AffectedRules []string `json:"affectedRules"`
    EstEffort     string   `json:"estimatedEffort"`
    RiskLevel     string   `json:"riskLevel"`
}

type ApprovalStatus struct {
    Required  []string   `json:"requiredApprovers"`
    Approvals []Approval `json:"approvals"`
}

type Approval struct {
    Approver   string    `json:"approver"`
    ApprovedAt time.Time `json:"approvedAt"`
    Comments   string    `json:"comments,omitempty"`
}

type ImplStatus struct {
    Status        string    `json:"status"` // not_started, in_progress, completed
    ImplementedBy string    `json:"implementedBy,omitempty"`
    ImplementedAt time.Time `json:"implementedAt,omitempty"`
    Commits       []string  `json:"commits,omitempty"`
}

// Gap Analysis
type GapAnalysis struct {
    ImplementedNotDoc []CodeGap  `json:"implementedButNotDocumented"`
    DocumentedNotImpl []RuleGap  `json:"documentedButNotImplemented"`
    PartiallyImpl     []PartialGap `json:"partiallyImplemented"`
    TestsMissing      []TestGap  `json:"testsMissing"`
    Summary           GapSummary `json:"summary"`
}

type CodeGap struct {
    File        string `json:"file"`
    Function    string `json:"function"`
    Logic       string `json:"logic"`
    Suggestion  string `json:"suggestion"`
}

type RuleGap struct {
    RuleID      string `json:"ruleId"`
    Title       string `json:"title"`
    Priority    string `json:"priority"`
    Suggestion  string `json:"suggestion"`
}

type PartialGap struct {
    RuleID       string   `json:"ruleId"`
    Title        string   `json:"title"`
    Implemented  []string `json:"implemented"`
    Missing      []string `json:"missing"`
}

type TestGap struct {
    RuleID       string   `json:"ruleId"`
    RequiredTests int     `json:"requiredTests"`
    WrittenTests int      `json:"writtenTests"`
    MissingTests []string `json:"missingTests"`
}

type GapSummary struct {
    TotalRules           int     `json:"totalRules"`
    FullyImplemented     int     `json:"fullyImplemented"`
    PartiallyImplemented int     `json:"partiallyImplemented"`
    NotImplemented       int     `json:"notImplemented"`
    UndocumentedFeatures int     `json:"undocumentedFeatures"`
    ImplementationRate   float64 `json:"implementationRate"`
}
```

### Telemetry Types

```go
type TelemetryEvent struct {
    Event     string                 `json:"event"`
    AgentID   string                 `json:"agentId"`
    OrgID     string                 `json:"orgId"`
    TeamID    string                 `json:"teamId,omitempty"`
    Timestamp string                 `json:"timestamp"`
    Metrics   map[string]interface{} `json:"metrics"`
}

type TelemetryClient struct {
    config   TelemetryConfig
    queue    []TelemetryEvent
    queueMux sync.Mutex
}
```

### MCP Types

```go
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
}

type MCPResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type MCPTool struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"inputSchema"`
}
```

### Reliability Layer Types (Phase 9.5.1) ✅ 100% COMPLETE

```go
// Database Timeout Helpers (Hub API - located in hub/api/hook_handler.go)
func queryWithTimeout(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
func queryRowWithTimeout(ctx context.Context, query string, args ...interface{}) *sql.Row
func execWithTimeout(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

// Parameters:
//   - ctx: Request context (cancelled on timeout)
//   - query: SQL query string
//   - args: Query parameters
// Returns:
//   - queryWithTimeout: *sql.Rows, error
//   - queryRowWithTimeout: *sql.Row
//   - execWithTimeout: sql.Result, error
// Timeout: 10 seconds (default)
// Error Conditions:
//   - context.DeadlineExceeded: Query timed out
//   - sql.ErrNoRows: No rows found (queryRowWithTimeout only)
//   - Database connection errors

// HTTP Retry Logic (Agent)
func httpRequestWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error)

// Parameters:
//   - client: HTTP client
//   - maxRetries: Maximum retry attempts (default: 3)
// Returns:
//   - *http.Response: HTTP response
//   - error: Error if all retries failed
// Retry Strategy:
//   - Exponential backoff: 100ms * (2^attempt)
//   - Retries on: network errors, 5xx server errors
//   - No retry on: 4xx client errors
//   - Max retries: 3 (default)

// Validation Helpers (Hub API - located in hub/api/validation.go)
func validateUUID(id string) error
func validateRequired(field, value string) error
func validateRange(field string, value, min, max int) error
func validateHookType(hookType string) error
func validateResult(result string) error
func validateDate(dateStr string) error
func validateAction(action string) error

// Parameters:
//   - id: UUID string to validate
//   - field: Field name for error messages
//   - value: Value to validate
//   - hookType: Hook type ("pre-commit", "pre-push", "commit-msg")
//   - result: Result type ("allowed", "blocked", "overridden")
//   - dateStr: Date string in YYYY-MM-DD or RFC3339 format
//   - action: Action type ("approve", "reject")
// Returns:
//   - error: Validation error with clear message
// Error Conditions:
//   - Invalid UUID format
//   - Empty required field
//   - Value outside allowed range
//   - Invalid enum value

// Cache Structures
type cachedPolicy struct {
    Policy    HookPolicy
    UpdatedAt time.Time
    CachedAt  time.Time
    mu        sync.RWMutex
}

type limitsCacheEntry struct {
    Limits   HookLimits
    Expires  time.Time
    mu       sync.RWMutex
}

// Cache Behavior:
//   - Policy cache: Invalidated when Hub updated_at > cached UpdatedAt
//   - Limits cache: Per-entry expiration (5 minutes default)
//   - AST cache: Time-based cleanup (prevents resource leaks)
//   - Thread-safe: RWMutex for concurrent access

// Error Recovery
type CheckResult struct {
    Enabled  bool   `json:"enabled"`
    Success  bool   `json:"success"`
    Error    string `json:"error,omitempty"`
    Findings int    `json:"findings"`
}

// Usage:
//   - CheckResults map[string]CheckResult in AuditReport
//   - Tracks: file_size, security, vibe, business_rules
//   - Populated automatically in performAuditForHook()
//   - Error state set on panic or failure
//   - Finding counts tracked (before/after)

// Database Connection Pool Health
func monitorDBHealth(db *sql.DB)

// Behavior:
//   - Background goroutine monitors connection pool
//   - Logs metrics: OpenConnections, IdleConnections, InUseConnections
//   - Alerts on potential pool exhaustion
//   - Runs continuously (checks every 30 seconds)
//   - Connection lifetime: SetConnMaxLifetime(5 * time.Minute)
```
```

---

## Production-Ready Implementations

All handlers are fully implemented and production-ready.

### validateCodeHandler

**Location**: `hub/api/main.go:1658-1750`

**Status**: ✅ Complete

**Implementation**: 
- Calls `analyzeAST(req.Code, req.Language, []string{"duplicates", "unused", "unreachable"})` (line 1703)
- Converts `ASTFinding` results to violations format (lines 1709-1725)
- Returns actual code violations with line numbers, columns, and messages
- Includes analysis statistics in response

**Impact**: `sentinel_validate_code` MCP tool fully functional and reports actual code violations.

### applyFixHandler

**Location**: `hub/api/main.go:1855-1940`

**Status**: ✅ Complete

**Implementation**:
- Applies fixes based on `fixType` parameter (lines 1898-1904):
  - `security`: Calls `ApplySecurityFixes()` - applies security fixes (SQL injection, XSS, secret detection)
  - `style`: Calls `ApplyStyleFixes()` - applies style fixes (line endings, indentation)
  - `performance`: Calls `ApplyPerformanceFixes()` - applies performance fixes (caching, loop optimization)
- Returns modified code with detailed change descriptions
- Includes retry logic with exponential backoff
- Verifies fixes after application

**Impact**: `sentinel_apply_fix` MCP tool fully functional and applies actual code fixes.

### sentinel_analyze_intent Handler

**Location**: `synapsevibsentinel.sh:5200+` (handleAnalyzeIntent function)

**Status**: ✅ Complete

**Implementation**:
- Handler function `handleAnalyzeIntent` exists and is registered
- Calls Hub endpoint `/api/v1/analyze/intent` (which exists and works)
- Formats response as MCP response with proper error handling
- Includes timeout and retry logic

**Impact**: `sentinel_analyze_intent` MCP tool fully functional and available.

---

## File System Structure

```
project/
├── .sentinel/
│   ├── patterns.json           # Learned patterns
│   ├── decisions.json          # Developer decisions
│   ├── history.json            # Audit history
│   ├── context.json            # Current context
│   ├── telemetry-queue.json    # Offline telemetry queue
│   └── backups/                # Fix backups
│       └── {timestamp}/
│           ├── manifest.json
│           └── {files...}
│
├── .sentinelsrc                # Project config
├── .sentinel-baseline.json     # Baselined findings
│
├── .cursor/
│   └── rules/
│       ├── 00-constitution.md
│       ├── 01-business-context.md
│       └── project-patterns.md
│
└── docs/
    └── knowledge/
        ├── source-documents/       # Original uploads
        │   ├── Scope_v2.pdf
        │   ├── Requirements.docx
        │   └── manifest.json       # Tracks ingested docs
        │
        ├── extracted/              # Raw extraction
        │   ├── Scope_v2.txt
        │   ├── Requirements.txt
        │   └── Data_Model.json
        │
        ├── drafts/                 # Pending review
        │   ├── domain-glossary.draft.md
        │   ├── business-rules.draft.md
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

---

## Task Dependency & Verification Types (Phase 14E)

### Task Types

```go
type Task struct {
    ID          string    `json:"id"`
    ProjectID   string    `json:"project_id"`
    Source      string    `json:"source"` // 'cursor', 'manual', 'change_request', 'comprehensive_analysis'
    Title       string    `json:"title"`
    Description string    `json:"description"`
    FilePath    string    `json:"file_path"`
    LineNumber  int       `json:"line_number"`
    Status      string    `json:"status"` // 'pending', 'in_progress', 'completed', 'blocked'
    Priority    string    `json:"priority"` // 'low', 'medium', 'high', 'critical'
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    VerifiedAt  *time.Time `json:"verified_at,omitempty"`
    VerificationConfidence float64 `json:"verification_confidence"`
}

type TaskDependency struct {
    ID              string  `json:"id"`
    TaskID          string  `json:"task_id"`
    DependsOnTaskID string  `json:"depends_on_task_id"`
    DependencyType  string  `json:"dependency_type"` // 'explicit', 'implicit', 'integration', 'feature'
    Confidence      float64 `json:"confidence"`
    CreatedAt       time.Time `json:"created_at"`
}

type TaskVerification struct {
    ID               string                 `json:"id"`
    TaskID           string                 `json:"task_id"`
    VerificationType string                 `json:"verification_type"` // 'code_existence', 'code_usage', 'test_coverage', 'integration'
    Status           string                 `json:"status"` // 'pending', 'verified', 'failed'
    Confidence       float64                `json:"confidence"`
    Evidence         map[string]interface{} `json:"evidence"`
    VerifiedAt       *time.Time             `json:"verified_at,omitempty"`
    CreatedAt        time.Time              `json:"created_at"`
}

type TaskLink struct {
    ID       string    `json:"id"`
    TaskID   string    `json:"task_id"`
    LinkType string    `json:"link_type"` // 'change_request', 'knowledge_item', 'comprehensive_analysis', 'test_requirement'
    LinkedID string    `json:"linked_id"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Task Detection Types

```go
type TaskDetectionResult struct {
    Tasks      []Task `json:"tasks"`
    TotalFound int    `json:"total_found"`
    ScannedFiles int  `json:"scanned_files"`
    DurationMs int64  `json:"duration_ms"`
}

type TaskDetectionConfig struct {
    ScanDirs      []string `json:"scan_dirs"`
    ExcludePaths  []string `json:"exclude_paths"`
    Sources       []string `json:"sources"` // 'cursor', 'manual', 'change_request', 'comprehensive_analysis'
    MinConfidence float64  `json:"min_confidence"`
}
```

### Task Verification Types

```go
type TaskVerificationRequest struct {
    TaskID string `json:"task_id"`
    Force  bool   `json:"force"` // Ignore cache
}

type TaskVerificationResponse struct {
    TaskID           string                 `json:"task_id"`
    OverallConfidence float64               `json:"overall_confidence"`
    Status           string                 `json:"status"`
    Verifications    []TaskVerification     `json:"verifications"`
    AutoCompleted    bool                   `json:"auto_completed"`
}

type VerificationEvidence struct {
    Files       []string          `json:"files"`
    Functions   []string          `json:"functions"`
    LineNumbers map[string][]int  `json:"line_numbers"`
    CallSites   []string          `json:"call_sites"`
    TestFile    string            `json:"test_file,omitempty"`
    Coverage    float64           `json:"coverage,omitempty"`
    Integration map[string]string  `json:"integration,omitempty"`
}
```

### Dependency Detection Types

```go
type DependencyDetectionResult struct {
    TaskID      string            `json:"task_id"`
    Dependencies []TaskDependency `json:"dependencies"`
    Cycles      [][]string        `json:"cycles"`
    Graph       DependencyGraph   `json:"graph"`
}

type DependencyGraph struct {
    Nodes []DependencyNode `json:"nodes"`
    Edges []DependencyEdge `json:"edges"`
}

type DependencyNode struct {
    ID     string `json:"id"`
    TaskID string `json:"task_id"`
    Label  string `json:"label"`
}

type DependencyEdge struct {
    From   string `json:"from"`
    To     string `json:"to"`
    Type   string `json:"type"`
    Weight float64 `json:"weight"`
}
```

### Task API Request/Response Types

```go
type CreateTaskRequest struct {
    ProjectID   string `json:"project_id"`
    Source      string `json:"source"`
    Title       string `json:"title"`
    Description string `json:"description"`
    FilePath    string `json:"file_path,omitempty"`
    LineNumber  int    `json:"line_number,omitempty"`
    Priority    string `json:"priority,omitempty"`
}

type ListTasksRequest struct {
    ProjectID string   `json:"project_id"`
    Status    []string `json:"status,omitempty"`
    Priority  []string `json:"priority,omitempty"`
    Source    []string `json:"source,omitempty"`
    Limit     int      `json:"limit,omitempty"`
    Offset    int      `json:"offset,omitempty"`
}

type ListTasksResponse struct {
    Tasks      []Task `json:"tasks"`
    Total      int    `json:"total"`
    Limit      int    `json:"limit"`
    Offset     int    `json:"offset"`
}

type AddDependencyRequest struct {
    TaskID          string  `json:"task_id"`
    DependsOnTaskID string  `json:"depends_on_task_id"`
    DependencyType  string  `json:"dependency_type"`
    Confidence      float64 `json:"confidence,omitempty"`
}
```

### Task Detection Algorithm Specification

**Input**: Codebase files, detection configuration
**Output**: List of detected tasks

**Algorithm**:
1. Scan files matching patterns (`.js`, `.ts`, `.py`, `.go`, etc.)
2. For each file:
   a. Extract TODO comments: `// TODO:`, `# TODO:`, `<!-- TODO: -->`
   b. Extract FIXME comments: `// FIXME:`, `# FIXME:`
   c. Extract Cursor task markers: `- [ ] Task:`, `- [x] Task:`
   d. Extract explicit task format: `// TASK: TASK-123 - Description`
3. Parse task metadata:
   - Title: Extract from comment/marker
   - Description: Extract from following lines
   - File path: Current file
   - Line number: Line where task found
   - Source: Infer from context (cursor, manual, etc.)
4. Store tasks in database

**Performance Target**: < 5 seconds for 1000 files

### Task Verification Algorithm Specification

**Input**: Task ID, codebase path
**Output**: Verification results with confidence scores

**Algorithm**:
1. Load task from database
2. For each verification factor:
   a. **Code Existence**:
      - Extract keywords from task title/description
      - Search codebase using AST (Phase 6)
      - Match function/class names, patterns
      - Calculate confidence based on match quality
   b. **Code Usage**:
      - Find function/class references
      - Track cross-file imports/calls
      - Calculate confidence based on usage frequency
   c. **Test Coverage**:
      - Find test files matching task keywords
      - Check test coverage (Phase 10)
      - Calculate confidence based on coverage percentage
   d. **Integration**:
      - Check for external API/service integration
      - Verify configuration files
      - Calculate confidence based on integration completeness
3. Calculate overall confidence: Weighted average of factors
4. Determine status:
   - > 0.8: Auto-complete
   - 0.5-0.8: In-progress
   - < 0.5: Pending
5. Store verification results

**Performance Target**: < 2 seconds per task

### Dependency Detection Algorithm Specification

**Input**: Task ID, codebase path, comprehensive analysis results (Phase 14A)
**Output**: Dependency graph

**Algorithm**:
1. **Explicit Dependencies**:
   - Parse task description for "Depends on: TASK-XXX"
   - Extract task IDs
   - Create explicit dependencies (confidence: 1.0)
2. **Implicit Dependencies**:
   - Analyze code for function/class calls
   - Map calls to tasks (by function/class name)
   - Create implicit dependencies (confidence: 0.7-0.9)
3. **Integration Dependencies**:
   - Use comprehensive analysis (Phase 14A) for feature discovery
   - Identify external API/service requirements
   - Create integration dependencies (confidence: 0.8)
4. **Feature-Level Dependencies**:
   - Use comprehensive analysis (Phase 14A) for feature mapping
   - Identify tasks part of same feature
   - Create feature-level dependencies (confidence: 0.9)
5. Build dependency graph
6. Detect cycles using DFS
7. Return dependency graph

**Performance Target**: < 1 second for 100 tasks

---

## Hub Specification

### API Endpoints

```
# Telemetry (Phase 5) ✅ IMPLEMENTED
POST   /api/v1/telemetry           # Ingest telemetry event
GET    /api/v1/telemetry/recent    # Recent events

# Metrics (Phase 5) ✅ IMPLEMENTED
GET    /api/v1/metrics             # Aggregate metrics
GET    /api/v1/metrics/trends      # Trend data
GET    /api/v1/metrics/team/:id    # Team metrics

# Knowledge (Phase 4) ✅ IMPLEMENTED
GET    /api/v1/projects/knowledge  # ✅ IMPLEMENTED - List project knowledge - Handler at line 1285 in main.go
GET    /api/v1/knowledge/business  # ✅ IMPLEMENTED - Get business context for MCP tools - Handler at line 1369 in main.go
PUT    /api/v1/knowledge/{id}/status # ✅ IMPLEMENTED - Update knowledge status - Handler at line 1215 in main.go
POST   /api/v1/knowledge/sync      # ✅ IMPLEMENTED - Sync knowledge items - Handler at line 1736 in main.go
POST   /api/v1/knowledge/migrate   # ✅ IMPLEMENTED - Migrate knowledge items (Phase 13) - Handler at line 3710 in main.go

# Documents (Phase 3B) ✅ IMPLEMENTED
POST   /api/v1/documents/ingest    # ✅ IMPLEMENTED - Upload/ingest document - Handler at line 866 in main.go
GET    /api/v1/documents/{id}/status # ✅ IMPLEMENTED - Get document status - Handler at line 1005 in main.go
GET    /api/v1/documents/{id}/extracted # ✅ IMPLEMENTED - Get extracted text - Handler at line 1081 in main.go
GET    /api/v1/documents/{id}/knowledge # ✅ IMPLEMENTED - Get knowledge items from document - Handler at line 1119 in main.go
POST   /api/v1/documents/{id}/detect-changes # ✅ IMPLEMENTED - Detect changes in document - Handler at line 968 in main.go
GET    /api/v1/documents           # ✅ IMPLEMENTED - List documents - Handler at line 1168 in main.go

# AST Analysis (Phase 6-9) ✅ IMPLEMENTED - Full Tree-sitter AST analysis
POST   /api/v1/analyze/ast         # ✅ IMPLEMENTED - Full AST analysis (Phase 6) - Handlers at lines 3118 and 3734 in main.go
POST   /api/v1/analyze/vibe        # ✅ IMPLEMENTED - Vibe coding issues only (Phase 7) - Handler at line 3171 in main.go
POST   /api/v1/analyze/cross-file  # ✅ IMPLEMENTED - Cross-file analysis (Phase 6) - Handler at line 3228 in main.go
POST   /api/v1/analyze/security    # ✅ IMPLEMENTED - Security analysis with AST and data flow (Phase 8) - Handler at line 3362 in main.go
GET    /api/v1/security/context    # ✅ IMPLEMENTED - Get security context for MCP tools (Phase 8) - Handler at line 1463 in main.go
POST   /api/v1/analyze/architecture # ✅ IMPLEMENTED - File structure analysis (Phase 9) - Handler at line 3594 in main.go

# Hook Management (Phase 9.5) ✅ IMPLEMENTED
POST   /api/v1/telemetry/hook       # ✅ IMPLEMENTED - Ingest hook execution events (Phase 9.5) - Handler at line 3770 in main.go
GET    /api/v1/hooks/metrics        # ✅ IMPLEMENTED - Get hook metrics (blocks, overrides, trends) (Phase 9.5) - Handler at line 3771 in main.go
GET    /api/v1/hooks/metrics/team   # ✅ IMPLEMENTED - Get team-level hook metrics (Phase 9.5) - Handler at line 3772 in main.go
GET    /api/v1/hooks/policies       # ✅ IMPLEMENTED - Get hook policies (Phase 9.5) - Handler at line 3773 in main.go
POST   /api/v1/hooks/policies       # ✅ IMPLEMENTED - Create/update hook policies (Phase 9.5) - Handler at line 3774 in main.go
GET    /api/v1/hooks/limits         # ✅ IMPLEMENTED - Get hook limits (Phase 9.5) - Handler at line 3775 in main.go
POST   /api/v1/hooks/baselines      # ✅ IMPLEMENTED - Create hook baseline (Phase 9.5) - Handler at line 3776 in main.go
POST   /api/v1/hooks/baselines/{id}/review # ✅ IMPLEMENTED - Review hook baseline (Phase 9.5) - Handler at line 3777 in main.go

# Comprehensive Analysis (Phase 14A) ✅ COMPLETE

# MCP Integration (Phase 14B) ✅ COMPLETE

**Status**: MCP server fully implemented with `sentinel_analyze_feature_comprehensive` tool.

**Implementation**:
- JSON-RPC 2.0 protocol over stdio
- Initialize, tools/list, tools/call methods implemented
- Comprehensive analysis tool integrated with Hub API
- Error handling with fallback messages
- Parameter validation

**See**: [Phase 14B Guide](./PHASE_14B_GUIDE.md) for usage details.
POST   /api/v1/analyze/comprehensive # ✅ COMPLETE - Comprehensive feature analysis across all layers (Phase 14A)
GET    /api/v1/validations/:id      # ✅ COMPLETE - Get comprehensive analysis results (Phase 14A)
GET    /api/v1/validations          # ✅ COMPLETE - List analyses for project (Phase 14A)

# MCP Integration (Phase 14B) ✅ COMPLETE
# MCP server: ./sentinel mcp-server
# Tool: sentinel_analyze_feature_comprehensive
# See: docs/external/PHASE_14B_GUIDE.md

# Validation Endpoints (Phase B) ✅ IMPLEMENTED
POST   /api/v1/validate/code        # ✅ IMPLEMENTED - Code validation with AST analysis - Handler at line 1517 in main.go
POST   /api/v1/validate/business    # ✅ IMPLEMENTED - Business rule validation - Handler at line 1595 in main.go
POST   /api/v1/fixes/apply          # ✅ IMPLEMENTED - Apply security/style/performance fixes - Handler at line 1682 in main.go

# Doc-Sync (Phase 11) ✅ IMPLEMENTED
POST   /api/v1/analyze/doc-sync     # ✅ IMPLEMENTED - Code-documentation comparison - Handler at line 3404 in main.go
POST   /api/v1/analyze/business-rules # ✅ IMPLEMENTED - Business rules comparison - Handler at line 3446 in main.go
GET    /api/v1/doc-sync/review-queue # ✅ IMPLEMENTED - Get review queue - Handler at line 3480 in main.go
POST   /api/v1/doc-sync/review/{id}  # ✅ IMPLEMENTED - Review doc-sync result - Handler at line 3550 in main.go

# Intent & Simple Language (Phase 15) ✅ COMPLETE
POST   /api/v1/analyze/intent       # ✅ COMPLETE - Analyze unclear prompts and generate clarifying questions (Phase 15) - Handler at line 4660 in main.go
POST   /api/v1/intent/decisions     # ✅ COMPLETE - Record user decisions for learning (Phase 15) - Handler at line 4731 in main.go
GET    /api/v1/intent/patterns      # ✅ COMPLETE - Get learned patterns (Phase 15) - Handler at line 4805 in main.go
# MCP Tool: sentinel_check_intent
# See: docs/external/PHASE_15_GUIDE.md

# LLM Configuration (Phase 14C) ✅ IMPLEMENTED
POST   /api/v1/llm/config           # ✅ IMPLEMENTED - Create LLM configuration - Handler at line 4880 in main.go
GET    /api/v1/llm/config/{id}      # ✅ IMPLEMENTED - Get LLM configuration - Handler at line 4980 in main.go
PUT    /api/v1/llm/config/{id}      # ✅ IMPLEMENTED - Update LLM configuration - Handler at line 5057 in main.go
DELETE /api/v1/llm/config/{id}      # ✅ IMPLEMENTED - Delete LLM configuration - Handler at line 5170 in main.go
GET    /api/v1/llm/config/project/{projectId} # ✅ IMPLEMENTED - List LLM configs for project - Handler at line 5218 in main.go
GET    /api/v1/llm/providers         # ✅ IMPLEMENTED - Get supported providers - Handler at line 5252 in main.go
GET    /api/v1/llm/models/{provider} # ✅ IMPLEMENTED - Get supported models for provider - Handler at line 5262 in main.go
POST   /api/v1/llm/config/validate   # ✅ IMPLEMENTED - Validate LLM configuration - Handler at line 5279 in main.go
GET    /api/v1/llm/usage/report      # ✅ IMPLEMENTED - Get usage report - Handler at line 5332 in main.go
GET    /api/v1/llm/usage/stats       # ✅ IMPLEMENTED - Get usage statistics - Handler at line 5380 in main.go
GET    /api/v1/llm/usage/cost-breakdown # ✅ IMPLEMENTED - Get cost breakdown - Handler at line 5408 in main.go
GET    /api/v1/llm/usage/trends      # ✅ IMPLEMENTED - Get usage trends - Handler at line 5436 in main.go

# Cost Optimization (Phase 14D) ✅ IMPLEMENTED
GET    /api/v1/metrics/cache        # ✅ IMPLEMENTED - Get cache metrics - Handler at line 5602 in main.go
GET    /api/v1/metrics/cost         # ✅ IMPLEMENTED - Get cost metrics - Handler at line 5652 in main.go

# Task Dependency & Verification (Phase 14E) ✅ IMPLEMENTED
POST   /api/v1/tasks                 # ✅ IMPLEMENTED - Create task - Handler at line 3814 in main.go
GET    /api/v1/tasks                 # ✅ IMPLEMENTED - List tasks - Handler at line 3815 in main.go
GET    /api/v1/tasks/{id}            # ✅ IMPLEMENTED - Get task - Handler at line 3816 in main.go
PUT    /api/v1/tasks/{id}            # ✅ IMPLEMENTED - Update task - Handler at line 3817 in main.go
DELETE /api/v1/tasks/{id}           # ✅ IMPLEMENTED - Delete task - Handler at line 3818 in main.go
POST   /api/v1/tasks/scan            # ✅ IMPLEMENTED - Scan codebase for tasks - Handler at line 3819 in main.go
POST   /api/v1/tasks/{id}/verify     # ✅ IMPLEMENTED - Verify task completion - Handler at line 3820 in main.go
POST   /api/v1/tasks/verify-all      # ✅ IMPLEMENTED - Verify all pending tasks - Handler at line 3821 in main.go
GET    /api/v1/tasks/{id}/dependencies # ✅ IMPLEMENTED - Get task dependencies - Handler at line 3822 in main.go
POST   /api/v1/tasks/{id}/detect-dependencies # ✅ IMPLEMENTED - Detect dependencies for task - Handler at line 3823 in main.go
# See: docs/external/TASK_DEPENDENCY_SYSTEM.md and docs/external/MCP_TASK_TOOLS_GUIDE.md

# Test Engine (Phase 10) ✅ IMPLEMENTED
POST   /api/v1/test-requirements/generate  # ✅ IMPLEMENTED - Generate test requirements (Phase 10) - Handler at line 3759 in main.go
POST   /api/v1/test-coverage/analyze        # ✅ IMPLEMENTED - Analyze test coverage (Phase 10) - Handler at line 3760 in main.go
GET    /api/v1/test-coverage/{knowledge_item_id} # ✅ IMPLEMENTED - Get test coverage (Phase 10) - Handler at line 3761 in main.go
POST   /api/v1/test-validations/validate   # ✅ IMPLEMENTED - Validate test quality (Phase 10) - Handler at line 3762 in main.go
GET    /api/v1/test-validations/{test_requirement_id} # ✅ IMPLEMENTED - Get validation results (Phase 10) - Handler at line 3763 in main.go
POST   /api/v1/mutation-test/run            # ✅ IMPLEMENTED - Run mutation testing (Phase 10) - Handler at line 3766 in main.go
GET    /api/v1/mutation-test/{test_requirement_id} # ✅ IMPLEMENTED - Get mutation results (Phase 10) - Handler at line 3767 in main.go
POST   /api/v1/test-execution/run           # ✅ IMPLEMENTED - Execute tests in sandbox (Phase 10) - Handler at line 3764 in main.go
GET    /api/v1/test-execution/{execution_id} # ✅ IMPLEMENTED - Get execution status (Phase 10) - Handler at line 3765 in main.go

# Requirements Lifecycle (Phase 12) ✅ IMPLEMENTED
# Note: Endpoints use different paths than originally documented
POST   /api/v1/knowledge/gap-analysis   # ✅ IMPLEMENTED - Gap analysis (Phase 12) - Handler at line 1796 in main.go
GET    /api/v1/change-requests           # ✅ IMPLEMENTED - List change requests (Phase 12) - Handler at line 1859 in main.go
GET    /api/v1/change-requests/{id}      # ✅ IMPLEMENTED - Get change request (Phase 12) - Handler at line 1905 in main.go
POST   /api/v1/change-requests/{id}/impact # ✅ IMPLEMENTED - Impact analysis (Phase 12) - Handler at line 2031 in main.go
POST   /api/v1/change-requests/{id}/start # ✅ IMPLEMENTED - Start implementation (Phase 12) - Handler at line 2086 in main.go
POST   /api/v1/change-requests/{id}/complete # ✅ IMPLEMENTED - Complete implementation (Phase 12) - Handler at line 2129 in main.go
POST   /api/v1/change-requests/{id}/update # ✅ IMPLEMENTED - Update implementation (Phase 12) - Handler at line 2172 in main.go
POST   /api/v1/change-requests/{id}/approve # ✅ IMPLEMENTED - Approve change request (Phase 12) - Handler at line 1934 in main.go
POST   /api/v1/change-requests/{id}/reject # ✅ IMPLEMENTED - Reject change request (Phase 12) - Handler at line 1981 in main.go
GET    /api/v1/change-requests/dashboard # ✅ IMPLEMENTED - Change requests dashboard (Phase 12) - Handler at line 2220 in main.go

# Organizations (Phase 6 - Planned)
POST   /api/orgs                   # Create org
GET    /api/orgs/:id               # Get org
PUT    /api/orgs/:id               # Update org
DELETE /api/orgs/:id               # Delete org

# Teams
POST   /api/teams                  # Create team
GET    /api/teams/:id              # Get team
PUT    /api/teams/:id              # Update team
DELETE /api/teams/:id              # Delete team
GET    /api/teams/:id/agents       # Team's agents

# Patterns
GET    /api/patterns               # Get org patterns
PUT    /api/patterns               # Update patterns
POST   /api/patterns/distribute    # Push to agents

# Agents
GET    /api/agents                 # List agents
GET    /api/agents/:id             # Get agent
DELETE /api/agents/:id             # Remove agent

# Auth
POST   /api/auth/login             # Login
POST   /api/auth/logout            # Logout
GET    /api/auth/me                # Current user
```

### Database Schema

```sql
-- Organizations
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'developer',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    name VARCHAR(255),
    version VARCHAR(50),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Telemetry
CREATE TABLE telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    metrics JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Patterns
CREATE TABLE patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    patterns JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Comprehensive Validations (Phase 14A)
CREATE TABLE comprehensive_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    feature VARCHAR(255) NOT NULL,
    validation_id VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL, -- "completed", "failed", "pending"
    summary JSONB NOT NULL,
    checklist JSONB NOT NULL,
    layer_analysis JSONB NOT NULL,
    end_to_end_flows JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Analysis Configurations (Phase 14A)
CREATE TABLE analysis_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    llm_provider_config JSONB NOT NULL,
    cost_optimization_config JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Hook Executions (Phase 9.5) ✅ IMPLEMENTED
CREATE TABLE hook_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(64),
    org_id UUID,
    team_id UUID,
    hook_type VARCHAR(20) NOT NULL,
    result VARCHAR(20) NOT NULL,
    override_reason VARCHAR(255),
    findings_summary JSONB,
    user_actions JSONB,
    duration_ms BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Hook Baselines (Phase 9.5) ✅ IMPLEMENTED
CREATE TABLE hook_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(64),
    org_id UUID,
    baseline_entry JSONB NOT NULL,
    source VARCHAR(20) NOT NULL,
    hook_type VARCHAR(20),
    reviewed BOOLEAN DEFAULT false,
    reviewed_by VARCHAR(100),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Hook Policies (Phase 9.5) ✅ IMPLEMENTED
CREATE TABLE hook_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    policy_config JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_telemetry_agent ON telemetry(agent_id);
CREATE INDEX idx_telemetry_created ON telemetry(created_at);
CREATE INDEX idx_telemetry_type ON telemetry(event_type);
CREATE INDEX idx_agents_org ON agents(org_id);
CREATE INDEX idx_teams_org ON teams(org_id);
CREATE INDEX idx_validations_project ON comprehensive_validations(project_id);
CREATE INDEX idx_validations_validation_id ON comprehensive_validations(validation_id);
CREATE INDEX idx_validations_created ON comprehensive_validations(created_at);
CREATE INDEX idx_configurations_org ON analysis_configurations(org_id);
CREATE INDEX idx_hook_executions_agent ON hook_executions(agent_id);
CREATE INDEX idx_hook_executions_org ON hook_executions(org_id);
CREATE INDEX idx_hook_executions_created ON hook_executions(created_at);
CREATE INDEX idx_hook_baselines_org ON hook_baselines(org_id);
CREATE INDEX idx_hook_baselines_reviewed ON hook_baselines(reviewed);
CREATE INDEX idx_hook_policies_org ON hook_policies(org_id);

-- Row Level Security
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE telemetry ENABLE ROW LEVEL SECURITY;
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| API Server | Go 1.21 + Chi router |
| Database | PostgreSQL 14+ |
| Migrations | golang-migrate |
| Dashboard | React 18 + TypeScript |
| Charts | Recharts |
| Styling | Tailwind CSS |
| Build | Vite |
| Deployment | Docker + Docker Compose |
| Auth | OAuth 2.0 / OIDC |

---

## MCP Specification

### Protocol

MCP uses JSON-RPC 2.0 over stdio.

### Server Initialization

**Status**: ✅ IMPLEMENTED (Phase 14B)

MCP server is fully implemented in `synapsevibsentinel.sh`. Key functions:

- `runMCPServer()`: Main loop reading JSON-RPC 2.0 requests from stdin, writing responses to stdout
- `handleMCPRequest()`: Routes requests to appropriate handlers (initialize, tools/list, tools/call, notifications/initialized)
- `handleInitialize()`: Returns protocol version, capabilities, and server info
- `handleToolsList()`: Returns list of registered tools
- `handleToolsCall()`: Routes tool calls to specific handlers
- `handleComprehensiveAnalysis()`: Implements `sentinel_analyze_feature_comprehensive` tool

**See**: `synapsevibsentinel.sh` for full implementation and [Phase 14B Guide](./PHASE_14B_GUIDE.md) for usage.

### Phase 15: Intent & Simple Language ✅ COMPLETE

**Status**: ✅ IMPLEMENTED (Phase 15)

Intent analysis and simple language handling for unclear prompts. Key functions:

- `handleCheckIntent()`: Implements `sentinel_check_intent` MCP tool
- `formatIntentAnalysisResponse()`: Formats Hub response for Cursor display
- Hub API endpoints: `/api/v1/analyze/intent`, `/api/v1/intent/decisions`, `/api/v1/intent/patterns`
- Intent analyzer: `AnalyzeIntent()`, `GatherContext()`, `RecordDecision()`, `GetLearnedPatterns()`

**See**: `hub/api/intent_analyzer.go` for implementation and [Phase 15 Guide](./PHASE_15_GUIDE.md) for usage.

### Tool Definitions

```json
{
  "tools": [
    {
      "name": "sentinel_get_context",
      "description": "Get current project context including recent files, errors, and git status",
      "inputSchema": {
        "type": "object",
        "properties": {}
      }
    },
    {
      "name": "sentinel_get_patterns",
      "description": "Get project patterns for a specific directory or file",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {"type": "string", "description": "Path to get patterns for"}
        }
      }
    },
    {
      "name": "sentinel_check_intent",
      "description": "Check if user intent is clear, return clarifying questions if not",
      "inputSchema": {
        "type": "object",
        "properties": {
          "request": {"type": "string", "description": "User's request"},
          "context": {"type": "object", "description": "Additional context"}
        },
        "required": ["request"]
      }
    },
    {
      "name": "sentinel_analyze_feature_comprehensive",
      "description": "Perform comprehensive analysis of a feature across all layers (UI, API, Database, Logic, Integration, Tests) with business context validation",
      "inputSchema": {
        "type": "object",
        "properties": {
          "feature": {
            "type": "string",
            "description": "Feature name or description (e.g., 'Order Cancellation')"
          },
          "mode": {
            "type": "string",
            "enum": ["auto", "manual"],
            "description": "Auto-discover feature components or use manual file specification",
            "default": "auto"
          },
          "files": {
            "type": "object",
            "description": "Manual file specification (required if mode='manual')",
            "properties": {
              "ui": {"type": "array", "items": {"type": "string"}},
              "api": {"type": "array", "items": {"type": "string"}},
              "database": {"type": "array", "items": {"type": "string"}},
              "logic": {"type": "array", "items": {"type": "string"}},
              "integration": {"type": "array", "items": {"type": "string"}},
              "tests": {"type": "array", "items": {"type": "string"}}
            }
          },
          "depth": {
            "type": "string",
            "enum": ["surface", "medium", "deep"],
            "description": "Analysis depth (surface=fast, medium=balanced, deep=comprehensive)",
            "default": "medium"
          },
          "includeBusinessContext": {
            "type": "boolean",
            "description": "Include business rules, journeys, and entities validation",
            "default": true
          }
        },
        "required": ["feature"]
      }
    },
    {
      "name": "sentinel_validate_code",
      "description": "Validate code against project patterns and security rules",
      "status": "✅ Complete",
      "inputSchema": {
        "type": "object",
        "properties": {
          "code": {"type": "string", "description": "Code to validate"},
          "filePath": {"type": "string", "description": "Target file path"},
          "operation": {"type": "string", "description": "Type of operation"}
        },
        "required": ["code"]
      }
    },
    {
      "name": "sentinel_apply_fix",
      "description": "Apply a fix to code",
      "status": "✅ Complete",
      "inputSchema": {
        "type": "object",
        "properties": {
          "filePath": {"type": "string", "description": "Path to file to fix"},
          "fixType": {"type": "string", "description": "Type of fix: security, style, or performance"},
          "content": {"type": "string", "description": "Code content to fix"}
        },
        "required": ["filePath", "fixType", "content"]
      },
      "fixTypes": {
        "security": "Applies security fixes (removes hardcoded secrets, fixes SQL injection, XSS prevention)",
        "style": "Applies style fixes (removes trailing whitespace, fixes formatting)",
        "performance": "Applies performance fixes (optimizes loops, adds caching suggestions)"
      }
    },
    {
      "name": "sentinel_get_business_context",
      "description": "Get business rules and entity information for a domain area",
      "inputSchema": {
        "type": "object",
        "properties": {
          "entity": {"type": "string", "description": "Entity name"},
          "operation": {"type": "string", "description": "Operation type"}
        }
      }
    },
    {
      "name": "sentinel_analyze_intent",
      "description": "Analyze user intent and return context, rules, security, and test requirements",
      "status": "⚠️ Handler missing - Hub endpoint exists at /api/v1/analyze/intent",
      "inputSchema": {
        "type": "object",
        "properties": {
          "request": {"type": "string", "description": "User's request"},
          "recentFiles": {"type": "array", "description": "Recently edited files"},
          "gitStatus": {"type": "object", "description": "Current git status"}
        },
        "required": ["request"]
      },
      "note": "Handler needs to be implemented - Hub endpoint exists and works"
    },
    {
      "name": "sentinel_get_security_context",
      "description": "Get security requirements for a specific operation or endpoint",
      "inputSchema": {
        "type": "object",
        "properties": {
          "operation": {"type": "string", "description": "Operation type"},
          "endpoint": {"type": "string", "description": "API endpoint pattern"},
          "resources": {"type": "array", "description": "Resources being accessed"}
        }
      }
    },
    {
      "name": "sentinel_get_test_requirements",
      "description": "Get required tests for implementing a feature",
      "inputSchema": {
        "type": "object",
        "properties": {
          "feature": {"type": "string", "description": "Feature description"},
          "ruleIds": {"type": "array", "description": "Related business rule IDs"}
        }
      }
    },
    {
      "name": "sentinel_check_file_size",
      "description": "Check if target file is oversized and suggest alternatives",
      "inputSchema": {
        "type": "object",
        "properties": {
          "filePath": {"type": "string", "description": "Target file path"}
        },
        "required": ["filePath"]
      }
    },
    {
      "name": "sentinel_validate_security",
      "description": "Validate code against security rules",
      "inputSchema": {
        "type": "object",
        "properties": {
          "code": {"type": "string", "description": "Code to validate"},
          "filePath": {"type": "string", "description": "Target file path"},
          "securityRules": {"type": "array", "description": "Specific rules to check"}
        },
        "required": ["code"]
      }
    },
    {
      "name": "sentinel_validate_tests",
      "description": "Validate test quality and coverage",
      "inputSchema": {
        "type": "object",
        "properties": {
          "testCode": {"type": "string", "description": "Test code"},
          "sourceCode": {"type": "string", "description": "Source code being tested"},
          "ruleIds": {"type": "array", "description": "Business rules to verify"}
        },
        "required": ["testCode"]
      }
    },
    {
      "name": "sentinel_generate_tests",
      "description": "Generate test cases from business rules",
      "inputSchema": {
        "type": "object",
        "properties": {
          "ruleIds": {"type": "array", "description": "Business rule IDs"},
          "language": {"type": "string", "description": "Test language (jest, pytest, etc.)"},
          "style": {"type": "string", "description": "Test style (unit, integration)"}
        },
        "required": ["ruleIds"]
      }
    },
    {
      "name": "sentinel_run_tests",
      "description": "Execute tests in Hub sandbox",
      "inputSchema": {
        "type": "object",
        "properties": {
          "testCode": {"type": "string", "description": "Test code"},
          "sourceCode": {"type": "string", "description": "Source code"},
          "language": {"type": "string", "description": "Language (node, python, go)"}
        },
        "required": ["testCode", "sourceCode"]
      }
    }
  ]
}
```

---

## Security Specification

### Threat Model

| Threat | Vector | Mitigation |
|--------|--------|------------|
| Code exposure | Telemetry | Sanitization, no code in payloads |
| Secret leak | Config/logs | Never log sensitive data |
| Document leak | Ingestion | Local parsing, text-only to LLM |
| Unauthorized access | API | OAuth, API keys, RBAC |
| Data tampering | Transit | TLS 1.3, HMAC |
| SQL injection | API | Parameterized queries |
| Path traversal | File ops | Validation, sandboxing |

### Encryption

| Context | Algorithm |
|---------|-----------|
| Transit | TLS 1.3 |
| Database | AES-256-GCM |
| Passwords | bcrypt |
| API keys | SHA-256 |

### Authentication

| Flow | Method |
|------|--------|
| Dashboard | OAuth 2.0 / OIDC |
| Agent → Hub | API key + org ID |
| LLM Provider | Provider API key |

### Data Sanitization

```go
func sanitizeForTelemetry(report *AuditReport) TelemetryMetrics {
    // NEVER send code, file names, or finding details
    return TelemetryMetrics{
        "findings": map[string]int{
            "critical": report.Summary.Critical,
            "warning":  report.Summary.Warning,
            "info":     report.Summary.Info,
        },
        "compliance": calculateCompliance(report),
        "duration":   report.Duration,
        "fileCount":  len(report.Files),
    }
}
```

---

## Performance Specification

### Agent Performance

| Operation | Target | Max |
|-----------|--------|-----|
| Audit (1000 files) | <10s | 30s |
| Pattern learning | <30s | 60s |
| Safe fix (single file) | <100ms | 500ms |
| MCP tool call | <200ms | 500ms |
| Document parsing | <5s per doc | 30s |
| Knowledge extraction | <30s per doc | 120s |

### Hub Performance

| Operation | Target | Max |
|-----------|--------|-----|
| Telemetry ingest | <50ms | 200ms |
| Metrics query | <200ms | 1s |
| Dashboard load | <3s | 5s |
| Concurrent agents | 1000 | 10000 |

### Resource Limits

| Resource | Agent | Hub |
|----------|-------|-----|
| Memory | 256MB | 2GB |
| CPU | 1 core | 4 cores |
| Disk | 100MB | 50GB |
| Network | 1Mbps | 100Mbps |

---

## LLM Integration Specification

### Provider Abstraction

```go
type LLMProvider interface {
    ExtractKnowledge(text string) (*ExtractedKnowledge, error)
    AnalyzeImage(image []byte) (*ImageAnalysis, error)
    ClarifyIntent(request string, context map[string]interface{}) (*ClarificationResult, error)
}

type OpenAIProvider struct {
    apiKey string
    model  string
}

type OllamaProvider struct {
    endpoint string
    model    string
}
```

### Cost Estimation

| Provider | Model | Cost per 1K tokens | 10-page doc |
|----------|-------|-------------------|-------------|
| OpenAI | GPT-4 | $0.03 / $0.06 | ~$0.50 |
| OpenAI | GPT-4V | $0.01 per image | ~$0.05 |
| Anthropic | Claude 3 Sonnet | $0.003 / $0.015 | ~$0.15 |
| Ollama | Local | Free | Free |

### Extraction Prompts

```go
const EntityExtractionPrompt = `
You are analyzing project documentation to extract business entities.

DOCUMENT CONTENT:
{{.Text}}

TASK: Identify all business entities (nouns that represent core concepts).

For each entity, provide:
1. Name (singular, PascalCase)
2. Definition (1-2 sentences)
3. Key attributes (list)
4. Relationships to other entities
5. Source location in document
6. Confidence score (0-100%)

OUTPUT FORMAT: JSON
{
  "entities": [
    {
      "name": "User",
      "definition": "...",
      "attributes": [...],
      "relationships": [...],
      "source": "Page 5",
      "confidence": 95
    }
  ]
}

Flag entities with confidence < 70% for human review.
`
```

---

## Testing Specification

### Test Coverage Targets

| Component | Target |
|-----------|--------|
| Core scanning | >90% |
| Pattern detection | >85% |
| Fix application | >95% |
| Document parsing | >90% |
| Telemetry | >90% |
| MCP handlers | >80% |
| Hub API | >85% |

### Test Structure

```
tests/
├── fixtures/
│   ├── projects/           # Sample projects
│   ├── patterns/           # Known patterns
│   ├── security/           # Security test cases
│   ├── documents/          # Sample documents
│   └── knowledge/          # Expected extractions
├── unit/
│   ├── patterns_test.go
│   ├── scanning_test.go
│   ├── fix_test.go
│   ├── ingest_test.go
│   ├── telemetry_test.go
│   └── mcp_test.go
├── integration/
│   ├── workflow_test.go
│   ├── hub_test.go
│   └── hooks_test.go
└── security/
    ├── injection_test.go
    ├── sanitization_test.go
    └── auth_test.go
```

### CI Pipeline

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y poppler-utils tesseract-ocr
      - name: Run tests
        run: go test -v -coverprofile=coverage.out ./...
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi
```


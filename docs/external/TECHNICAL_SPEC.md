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
GET    /api/v1/projects/knowledge  # List project knowledge
PUT    /api/v1/knowledge/:id/status # Update knowledge status
POST   /api/v1/knowledge/:id/sync  # Sync knowledge item

# Documents (Phase 3B) ✅ IMPLEMENTED
POST   /api/v1/documents/upload    # Upload document
GET    /api/v1/documents/:id       # Get document status
GET    /api/v1/documents/:id/results # Get processing results

# AST Analysis (Phase 6-9) ⏳ NOT IMPLEMENTED - Foundation layer required before MCP
POST   /api/v1/analyze/ast         # ⏳ NOT IMPLEMENTED - Full AST analysis (Phase 6)
POST   /api/v1/analyze/vibe        # ⏳ NOT IMPLEMENTED - Vibe coding issues only (Phase 7)
POST   /api/v1/analyze/security    # ✅ IMPLEMENTED - Security analysis with AST and data flow (Phase 8)
POST   /api/v1/analyze/architecture # ✅ IMPLEMENTED - File structure analysis (Phase 9) - Provides suggestions only, not execution

# Comprehensive Analysis (Phase 14A) ⏳ PENDING
POST   /api/v1/analyze/comprehensive # ⏳ PENDING - Comprehensive feature analysis across all layers (Phase 14A)
GET    /api/v1/validations/:id      # ⏳ PENDING - Get comprehensive analysis results (Phase 14A)
GET    /api/v1/validations          # ⏳ PENDING - List analyses for project (Phase 14A)

# Test Engine (Phase 10) ⏳ NOT IMPLEMENTED
POST   /api/v1/tests/validate      # ⏳ NOT IMPLEMENTED - Validate test quality (Phase 10)
POST   /api/v1/tests/generate      # ⏳ NOT IMPLEMENTED - Generate tests from rules (Phase 10)
POST   /api/v1/tests/run           # ⏳ NOT IMPLEMENTED - Execute tests in sandbox (Phase 10)
GET    /api/v1/tests/coverage      # ⏳ NOT IMPLEMENTED - Get test coverage (Phase 10)

# Requirements Lifecycle (Phase 12) ⏳ NOT IMPLEMENTED
GET    /api/v1/requirements/gaps   # ⏳ NOT IMPLEMENTED - Gap analysis (Phase 12)
POST   /api/v1/requirements/changes # ⏳ NOT IMPLEMENTED - Create change request (Phase 12)
GET    /api/v1/requirements/changes/:id # ⏳ NOT IMPLEMENTED - Get change request (Phase 12)
PUT    /api/v1/requirements/changes/:id # ⏳ NOT IMPLEMENTED - Update change request (Phase 12)
GET    /api/v1/requirements/impact/:ruleId # ⏳ NOT IMPLEMENTED - Impact analysis (Phase 12)

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

```go
func runMCPServer() {
    scanner := bufio.NewScanner(os.Stdin)
    encoder := json.NewEncoder(os.Stdout)
    
    for scanner.Scan() {
        var req MCPRequest
        json.Unmarshal(scanner.Bytes(), &req)
        
        resp := handleRequest(req)
        encoder.Encode(resp)
    }
}

func handleRequest(req MCPRequest) MCPResponse {
    switch req.Method {
    case "initialize":
        return handleInitialize(req)
    case "tools/list":
        return listTools()
    case "tools/call":
        return callTool(req.Params)
    default:
        return MCPResponse{
            Error: &MCPError{Code: -32601, Message: "Method not found"},
        }
    }
}
```

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
      "inputSchema": {
        "type": "object",
        "properties": {
          "code": {"type": "string", "description": "Code to fix"},
          "fixId": {"type": "string", "description": "ID of fix to apply"}
        },
        "required": ["code", "fixId"]
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
      "inputSchema": {
        "type": "object",
        "properties": {
          "request": {"type": "string", "description": "User's request"},
          "recentFiles": {"type": "array", "description": "Recently edited files"},
          "gitStatus": {"type": "object", "description": "Current git status"}
        },
        "required": ["request"]
      }
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


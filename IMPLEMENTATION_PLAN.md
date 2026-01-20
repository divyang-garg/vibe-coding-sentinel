# Detailed Implementation Plan

> **Compliance Note**: All implementations must follow `CODING_STANDARDS.md` requirements:
> - Entry Points: max 50 lines
> - HTTP Handlers: max 300 lines
> - Business Services: max 400 lines
> - Repositories: max 350 lines
> - Utilities: max 250 lines
> - Tests: max 500 lines
> - Constructor injection for all dependencies (Section 7.1)
> - Request validation (Section 11.1)

---

## Priority 1: Test Fixes (2-4 hours)

### Task 1.1: Fix Mock Args in `TestTaskRepository_FindByID`
**File**: `hub/api/repository/task_repository_test.go`
**Estimated Time**: 1 hour
**Issue**: Mock `QueryRow` call arguments mismatch with actual implementation

#### Current Problem

The test fails with a type mismatch in `Scan` arguments:

**Actual call types (from repository):**
```
Scan(*string, *string, *string, *string, *string, *string, **int, *models.TaskStatus, 
     *models.TaskPriority, **string, **int, **int, *float64, *time.Time, *time.Time, 
     **time.Time, **time.Time, **time.Time, *int)
```

**Mock expectation types (from test):**
```
Scan(*string, *string, *string, *string, *string, *string, *int, *models.TaskStatus,
     *models.TaskPriority, **string, *int, *int, *float64, *time.Time, *time.Time,
     **time.Time, **time.Time, **time.Time, *int)
```

**Key differences at positions 6, 10, 11:**
- Position 6 (LineNumber): actual `**int`, expected `*int`
- Position 10 (EstimatedEffort): actual `**int`, expected `*int`  
- Position 11 (ActualEffort): actual `**int`, expected `*int`

This is because nullable integer fields in the model are `*int`, and when scanned into, you need a pointer to that pointer (`**int`).

#### Implementation Steps

1. **Fix the variable declarations** (lines 177-178) to use pointers for nullable fields:
```go
// Change from:
var completedAt, verifiedAt, archivedAt *time.Time
var lineNumber, estimatedEffort, actualEffort int

// To:
var completedAt, verifiedAt, archivedAt *time.Time
var lineNumber, estimatedEffort, actualEffort *int  // Nullable fields need *int
```

2. **Fix Scan mock expectations** (lines 179-215) to match the actual pointer types:
```go
mockRow.On("Scan",
    &expectedTask.ID,                    // *string
    &expectedTask.ProjectID,             // *string
    &expectedTask.Source,                // *string
    &expectedTask.Title,                 // *string
    &expectedTask.Description,           // *string
    &expectedTask.FilePath,              // *string
    &lineNumber,                         // **int (pointer to *int)
    &expectedTask.Status,                // *TaskStatus
    &expectedTask.Priority,              // *TaskPriority
    &expectedTask.AssignedTo,            // **string
    &estimatedEffort,                    // **int (pointer to *int)
    &actualEffort,                       // **int (pointer to *int)
    &expectedTask.VerificationConfidence,// *float64
    &expectedTask.CreatedAt,             // *time.Time
    &expectedTask.UpdatedAt,             // *time.Time
    &completedAt,                        // **time.Time
    &verifiedAt,                         // **time.Time
    &archivedAt,                         // **time.Time
    &expectedTask.Version,               // *int
).Return(nil)
```

3. **Alternative: Use a simpler approach with mock.Anything** for complex variadic scan:
```go
func TestTaskRepository_FindByID(t *testing.T) {
    mockDB := &MockDatabase{}
    mockRow := &MockRow{}
    repo := NewTaskRepository(mockDB)
    
    ctx := context.Background()
    
    mockDB.On("QueryRow", ctx, mock.AnythingOfType("string"), "task-123").
        Return(mockRow)
    
    // Use variadic mock.Anything - simpler and more maintainable
    mockRow.On("Scan", 
        mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
        mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
        mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
        mock.Anything, mock.Anything, mock.Anything, mock.Anything,
    ).Return(nil).Run(func(args mock.Arguments) {
        // Set values via the passed pointers
        if id, ok := args.Get(0).(*string); ok {
            *id = "task-123"
        }
        if projectID, ok := args.Get(1).(*string); ok {
            *projectID = "project-456"
        }
        if source, ok := args.Get(2).(*string); ok {
            *source = "cursor"
        }
        if title, ok := args.Get(3).(*string); ok {
            *title = "Test Task"
        }
        if desc, ok := args.Get(4).(*string); ok {
            *desc = "A test task"
        }
        if status, ok := args.Get(7).(*models.TaskStatus); ok {
            *status = "pending"
        }
        if priority, ok := args.Get(8).(*models.TaskPriority); ok {
            *priority = "medium"
        }
        if version, ok := args.Get(18).(*int); ok {
            *version = 1
        }
    })
    
    task, err := repo.FindByID(ctx, "task-123")
    
    assert.NoError(t, err)
    assert.NotNil(t, task)
    assert.Equal(t, "task-123", task.ID)
    mockDB.AssertExpectations(t)
    mockRow.AssertExpectations(t)
}
```

#### Verification
```bash
cd hub/api && go test -run TestTaskRepository_FindByID ./repository/... -v
```

---

### Task 1.2: Fix Mock Setup in `TestDocumentUploadProcessExtract`
**File**: `hub/api/services/integration_test.go` (Note: This is `document_integration_test.go`)
**Estimated Time**: 1 hour
**Issue**: Mock expectations not matching actual service calls

#### Current Problem
The mock setup has sequencing issues where:
1. `mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Document"))` returns nil but doesn't properly capture the document ID
2. Subsequent mocks reference `docID` before it's set by the service

#### Implementation Steps

1. **Fix mock document ID capture** (lines 220-223):
```go
// Replace the current mock setup:
var capturedDocID string
mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Document")).
    Return(nil).
    Run(func(args mock.Arguments) {
        savedDoc := args.Get(1).(*models.Document)
        capturedDocID = savedDoc.ID // Capture the generated ID
    })
```

2. **Use captured ID in subsequent mocks**:
```go
// After upload, use captured ID for processing mocks
mockRepo.On("FindByID", ctx, mock.MatchedBy(func(id string) bool {
    return id == capturedDocID || capturedDocID == ""
})).Return(doc, nil)
```

3. **Alternative: Use deferred mock setup**:
```go
// Setup upload mocks
mockValidator.On("ValidateFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
mockValidator.On("ValidateSize", mock.Anything, mock.Anything).Return(nil)
mockValidator.On("CheckSecurity", mock.Anything, mock.Anything).Return(nil)
mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

// Upload
uploadResp, err := service.UploadDocument(ctx, uploadReq, filePath, "text/plain")
docID := uploadResp.Document.ID

// Now setup processing mocks with actual docID
mockRepo.On("FindByID", ctx, docID).Return(&uploadResp.Document, nil)
mockRepo.On("UpdateStatus", ctx, docID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
mockExtractor.On("ExtractFromText", ctx, mock.Anything, docID).Return(knowledgeItems, nil)
// ... etc
```

4. **Fix the nil document check** in `t.Run("document not found")` (line 316):
```go
// Current:
mockRepo.On("FindByID", ctx, "nonexistent").Return(nil, nil)

// Should return error, not nil document:
mockRepo.On("FindByID", ctx, "nonexistent").Return((*models.Document)(nil), fmt.Errorf("not found"))
```

#### Verification
```bash
cd hub/api && go test -run TestDocumentUploadProcessExtract ./services/... -v
```

---

## Priority 2: Knowledge Extraction Enhancement (1-2 days)

### Task 2.1: Add LLM-Powered Knowledge Extraction
**Files**: 
- `internal/extraction/extractor.go` (existing - modify)
- `hub/api/repository/knowledge.go` (modify to use extraction package)
**Estimated Time**: 8 hours

#### Current State Analysis
- `internal/extraction/extractor.go` already has LLM infrastructure with:
  - `LLMClient` interface
  - Prompt building for different schema types
  - Retry logic with exponential backoff
  - Fallback to regex
- `hub/api/repository/knowledge.go` uses only regex patterns

#### Implementation Steps

**Step 1: Create LLM Client Implementation** (2 hours)
Create `internal/extraction/llm_client.go`:

```go
// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

// OllamaClient implements LLMClient for Ollama API
type OllamaClient struct {
    baseURL    string
    model      string
    httpClient *http.Client
}

// OllamaConfig configures the Ollama client
type OllamaConfig struct {
    BaseURL    string
    Model      string
    Timeout    time.Duration
}

// DefaultOllamaConfig returns default configuration
func DefaultOllamaConfig() OllamaConfig {
    return OllamaConfig{
        BaseURL: getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434"),
        Model:   getEnvOrDefault("OLLAMA_MODEL", "llama3.2"),
        Timeout: 120 * time.Second,
    }
}

// NewOllamaClient creates a new Ollama LLM client
func NewOllamaClient(cfg OllamaConfig) *OllamaClient {
    return &OllamaClient{
        baseURL: cfg.BaseURL,
        model:   cfg.Model,
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
        },
    }
}

// Call invokes the LLM with the given prompt
func (c *OllamaClient) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
    reqBody := map[string]interface{}{
        "model":  c.model,
        "prompt": prompt,
        "stream": false,
        "options": map[string]interface{}{
            "temperature": 0.2,
            "num_predict": 4096,
        },
    }
    
    body, err := json.Marshal(reqBody)
    if err != nil {
        return "", 0, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
    if err != nil {
        return "", 0, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", 0, fmt.Errorf("LLM request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", 0, fmt.Errorf("LLM returned status %d", resp.StatusCode)
    }
    
    var result struct {
        Response string `json:"response"`
        Context  []int  `json:"context"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", 0, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Estimate tokens (rough approximation)
    tokens := len(prompt)/4 + len(result.Response)/4
    
    return result.Response, tokens, nil
}

func getEnvOrDefault(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

**Step 2: Implement Proper Prompt Builder** (2 hours)
Create `internal/extraction/prompt_builder.go`:

```go
// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

// PromptBuilder interface for building extraction prompts
type PromptBuilder interface {
    BuildBusinessRulesPrompt(text string) string
    BuildEntitiesPrompt(text string) string
    BuildAPIContractsPrompt(text string) string
    BuildUserJourneysPrompt(text string) string
    BuildGlossaryPrompt(text string) string
}

// promptBuilder implements PromptBuilder
type promptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() PromptBuilder {
    return &promptBuilder{}
}

func (p *promptBuilder) BuildBusinessRulesPrompt(text string) string {
    return `You are a business analyst extracting business rules from documents.

Analyze the following text and extract all business rules in JSON format:

TEXT:
` + text + `

Return a JSON object with the following structure:
{
  "business_rules": [
    {
      "id": "BR-001",
      "version": "1.0",
      "status": "draft",
      "title": "Short descriptive title",
      "description": "Full description of the rule",
      "priority": "high|medium|low",
      "specification": {
        "trigger": "When/event that triggers the rule",
        "preconditions": ["List of conditions that must be true"],
        "constraints": [
          {
            "id": "C1",
            "type": "state_based|time_based|calculation",
            "expression": "The constraint in natural language",
            "pseudocode": "IF condition THEN action"
          }
        ],
        "exceptions": [
          {"id": "E1", "condition": "Exception condition", "modified_constraint": "Modified behavior"}
        ],
        "error_cases": [
          {"condition": "Error condition", "error_code": "ERR001", "error_message": "User message"}
        ]
      },
      "traceability": {
        "source_document": "Document name",
        "source_section": "Section reference",
        "source_quote": "Direct quote from source"
      }
    }
  ]
}

Extract ONLY valid business rules. Be precise and include source quotes for traceability.`
}

func (p *promptBuilder) BuildEntitiesPrompt(text string) string {
    return `Extract domain entities from this text as JSON:

TEXT:
` + text + `

Return:
{
  "entities": [
    {
      "id": "ENT-001",
      "name": "EntityName",
      "description": "Description",
      "category": "core|reference|transactional",
      "fields": [{"name": "field", "type": "string|int|bool|date", "required": true}],
      "relationships": [{"entity": "OtherEntity", "type": "one_to_many|many_to_one|many_to_many"}]
    }
  ]
}`
}

func (p *promptBuilder) BuildAPIContractsPrompt(text string) string {
    return `Extract API contracts from this text as JSON:

TEXT:
` + text + `

Return:
{
  "api_contracts": [
    {
      "id": "API-001",
      "endpoint": "/path",
      "method": "GET|POST|PUT|DELETE",
      "description": "Endpoint description",
      "request": {"params": {}, "query": {}, "body": {}},
      "response": {"status_codes": {"200": {"description": "Success", "body": {}}}}
    }
  ]
}`
}

func (p *promptBuilder) BuildUserJourneysPrompt(text string) string {
    return `Extract user journeys from this text as JSON:

TEXT:
` + text + `

Return:
{
  "user_journeys": [
    {
      "id": "UJ-001",
      "name": "Journey name",
      "actor": "User role",
      "goal": "What user wants to achieve",
      "steps": [{"step": 1, "actor_action": "User does X", "system_response": "System does Y"}]
    }
  ]
}`
}

func (p *promptBuilder) BuildGlossaryPrompt(text string) string {
    return `Extract glossary terms from this text as JSON:

TEXT:
` + text + `

Return:
{
  "glossary": [
    {
      "id": "TERM-001",
      "term": "Technical term",
      "definition": "Definition",
      "context": "Where this term is used",
      "synonyms": ["alternate names"],
      "examples": ["usage examples"]
    }
  ]
}`
}
```

**Step 3: Implement Response Parser** (1 hour)
Create `internal/extraction/response_parser.go`:

```go
// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strings"
)

// ResponseParser parses LLM responses into structured data
type ResponseParser interface {
    Parse(response string) (*ExtractResult, error)
}

// jsonParser implements ResponseParser
type jsonParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() ResponseParser {
    return &jsonParser{}
}

func (p *jsonParser) Parse(response string) (*ExtractResult, error) {
    // Extract JSON from response (handle markdown code blocks)
    jsonStr := extractJSON(response)
    if jsonStr == "" {
        return nil, fmt.Errorf("no valid JSON found in response")
    }
    
    var result ExtractResult
    if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }
    
    return &result, nil
}

func extractJSON(text string) string {
    // Try to find JSON in markdown code blocks
    codeBlockRegex := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
    matches := codeBlockRegex.FindStringSubmatch(text)
    if len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }
    
    // Try to find raw JSON object
    start := strings.Index(text, "{")
    end := strings.LastIndex(text, "}")
    if start >= 0 && end > start {
        return text[start : end+1]
    }
    
    return ""
}
```

**Step 4: Wire Up in Hub API** (2 hours)
Update `hub/api/repository/knowledge.go`:

```go
// ExtractFromFile extracts knowledge items from a file
func (k *KnowledgeExtractorImpl) ExtractFromFile(ctx context.Context, filePath string, mimeType string, docID string) ([]models.KnowledgeItem, error) {
    // Parse file content based on MIME type
    text, err := k.parseFileContent(filePath, mimeType)
    if err != nil {
        return nil, fmt.Errorf("failed to parse file: %w", err)
    }
    
    if text == "" {
        return []models.KnowledgeItem{}, nil
    }
    
    // Extract using text extraction
    return k.ExtractFromText(ctx, text, docID)
}

func (k *KnowledgeExtractorImpl) parseFileContent(filePath, mimeType string) (string, error) {
    switch mimeType {
    case "text/plain", "text/markdown":
        content, err := os.ReadFile(filePath)
        return string(content), err
    case "application/pdf":
        return k.parsePDF(filePath)
    case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
        return k.parseDOCX(filePath)
    default:
        return "", fmt.Errorf("unsupported MIME type: %s", mimeType)
    }
}
```

**Step 5: Create Integration Factory** (1 hour)
Create `internal/extraction/factory.go`:

```go
// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

// ExtractorFactory creates configured extractors
type ExtractorFactory struct{}

// NewExtractorFactory creates a new factory
func NewExtractorFactory() *ExtractorFactory {
    return &ExtractorFactory{}
}

// CreateDefault creates an extractor with default configuration
func (f *ExtractorFactory) CreateDefault() *KnowledgeExtractor {
    cfg := DefaultOllamaConfig()
    llmClient := NewOllamaClient(cfg)
    promptBuilder := NewPromptBuilder()
    parser := NewResponseParser()
    scorer := NewConfidenceScorer()
    fallback := NewFallbackExtractor()
    cache := NewInMemoryCache(1000) // Implement in cache package
    logger := NewStdLogger() // Simple stdout logger
    
    return NewKnowledgeExtractor(
        llmClient,
        promptBuilder,
        parser,
        scorer,
        fallback,
        cache,
        logger,
    )
}
```

#### Verification
```bash
# Unit tests
go test ./internal/extraction/... -v

# Integration test with Ollama
OLLAMA_HOST=http://localhost:11434 go test ./internal/extraction/... -tags=integration -v
```

---

### Task 2.2: Implement `ExtractFromFile` with Text Extraction
**Files**: `hub/api/repository/knowledge.go`
**Estimated Time**: 4 hours

#### Current State
Returns empty slice - needs actual file parsing.

#### Implementation Steps

1. **Add file parsing dependencies** to `hub/api/go.mod`:
```
github.com/ledongthuc/pdf
github.com/nguyenthenguyen/docx
github.com/xuri/excelize/v2  // For Excel support
```

2. **Implement file parsers**:
```go
func (k *KnowledgeExtractorImpl) parsePDF(filePath string) (string, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer f.Close()
    
    info, _ := f.Stat()
    reader, err := pdf.NewReader(f, info.Size())
    if err != nil {
        return "", err
    }
    
    var text strings.Builder
    for i := 1; i <= reader.NumPage(); i++ {
        page := reader.Page(i)
        if page.V.IsNull() {
            continue
        }
        fontMap := make(map[string]*pdf.Font)
        pageText, _ := page.GetPlainText(fontMap)
        text.WriteString(pageText)
        text.WriteString("\n")
    }
    
    return text.String(), nil
}

func (k *KnowledgeExtractorImpl) parseDOCX(filePath string) (string, error) {
    doc, err := docx.ReadDocxFile(filePath)
    if err != nil {
        return "", err
    }
    defer doc.Close()
    
    return doc.Editable().GetContent(), nil
}
```

---

### Task 2.3: Add Extraction Confidence Scoring
**Files**: `internal/extraction/scoring.go`
**Estimated Time**: 4 hours

#### Current State
Basic weight-based scoring exists but needs enhancement.

#### Implementation Steps

1. **Add semantic completeness scoring**:
```go
// EnhancedConfidenceScorer provides advanced confidence scoring
type EnhancedConfidenceScorer struct {
    weights ConfidenceWeights
}

// ScoreRule calculates multi-factor confidence
func (s *EnhancedConfidenceScorer) ScoreRule(rule BusinessRule) float64 {
    var score float64
    
    // Structural completeness (40%)
    structScore := s.scoreStructure(rule)
    
    // Semantic quality (30%)
    semanticScore := s.scoreSemantics(rule)
    
    // Traceability (20%)
    traceScore := s.scoreTraceability(rule)
    
    // Constraint quality (10%)
    constraintScore := s.scoreConstraints(rule)
    
    score = structScore*0.4 + semanticScore*0.3 + traceScore*0.2 + constraintScore*0.1
    
    return math.Min(score, 1.0)
}

func (s *EnhancedConfidenceScorer) scoreStructure(rule BusinessRule) float64 {
    score := 0.0
    if rule.ID != "" { score += 0.15 }
    if rule.Title != "" { score += 0.25 }
    if len(rule.Description) > 20 { score += 0.30 }
    if rule.Priority != "" { score += 0.15 }
    if rule.Status != "" { score += 0.15 }
    return score
}

func (s *EnhancedConfidenceScorer) scoreSemantics(rule BusinessRule) float64 {
    score := 0.0
    
    // Check for actionable language
    actionWords := []string{"must", "shall", "should", "will", "can", "may"}
    desc := strings.ToLower(rule.Description)
    for _, word := range actionWords {
        if strings.Contains(desc, word) {
            score += 0.3
            break
        }
    }
    
    // Check for measurable criteria
    if regexp.MustCompile(`\d+`).MatchString(desc) {
        score += 0.3
    }
    
    // Description length quality
    if len(rule.Description) > 50 && len(rule.Description) < 500 {
        score += 0.4
    }
    
    return math.Min(score, 1.0)
}
```

2. **Add confidence classification**:
```go
type ConfidenceLevel string

const (
    ConfidenceHigh   ConfidenceLevel = "high"    // >= 0.8
    ConfidenceMedium ConfidenceLevel = "medium"  // 0.5-0.8
    ConfidenceLow    ConfidenceLevel = "low"     // < 0.5
)

func ClassifyConfidence(score float64) ConfidenceLevel {
    switch {
    case score >= 0.8:
        return ConfidenceHigh
    case score >= 0.5:
        return ConfidenceMedium
    default:
        return ConfidenceLow
    }
}
```

---

## Priority 3: Excel Support (4 hours)

### Task 3.1: Add XLSX Text Extraction
**Files**: 
- `internal/extraction/document_parser.go`
- `hub/api/repository/knowledge.go`
**Estimated Time**: 4 hours

#### Implementation Steps

1. **Add excelize dependency**:
```bash
go get github.com/xuri/excelize/v2
```

2. **Create Excel parser** in `internal/extraction/document_parser.go`:
```go
// xlsxParser handles Excel XLSX files
type xlsxParser struct{}

func (p *xlsxParser) Supports(filePath string) bool {
    ext := strings.ToLower(filepath.Ext(filePath))
    return ext == ".xlsx" || ext == ".xls"
}

func (p *xlsxParser) Parse(filePath string) (string, error) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to open Excel file: %w", err)
    }
    defer f.Close()
    
    var text strings.Builder
    
    // Iterate all sheets
    for _, sheetName := range f.GetSheetList() {
        text.WriteString(fmt.Sprintf("## Sheet: %s\n\n", sheetName))
        
        rows, err := f.GetRows(sheetName)
        if err != nil {
            continue
        }
        
        for rowIdx, row := range rows {
            // Header row detection
            if rowIdx == 0 {
                text.WriteString("| ")
                for _, cell := range row {
                    text.WriteString(cell + " | ")
                }
                text.WriteString("\n|")
                for range row {
                    text.WriteString("---|")
                }
                text.WriteString("\n")
                continue
            }
            
            // Data rows
            text.WriteString("| ")
            for _, cell := range row {
                text.WriteString(cell + " | ")
            }
            text.WriteString("\n")
        }
        text.WriteString("\n")
    }
    
    return text.String(), nil
}
```

3. **Update NewDocumentParser** factory:
```go
func NewDocumentParser(filePath string) (DocumentParser, error) {
    ext := strings.ToLower(filepath.Ext(filePath))
    switch ext {
    case ".md", ".markdown":
        return &markdownParser{}, nil
    case ".txt":
        return &textParser{}, nil
    case ".docx":
        return &docxParser{}, nil
    case ".pdf":
        return &pdfParser{}, nil
    case ".xlsx", ".xls":
        return &xlsxParser{}, nil
    default:
        return nil, fmt.Errorf("unsupported file type: %s", ext)
    }
}
```

4. **Update MIME type validation** in `hub/api/repository/knowledge.go`:
```go
allowedTypes := []string{
    "application/pdf",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // XLSX
    "application/vnd.ms-excel", // XLS
    "text/plain",
    "text/markdown",
}
```

#### Verification
```bash
# Create test Excel file
go test -run TestExcelParser ./internal/extraction/... -v
```

---

## Priority 4: Production Hardening (2-3 days)

### Task 4.1: Structured Logging (JSON)
**Files**: 
- `hub/api/pkg/logging.go` (rewrite)
- `hub/api/logging.go` (remove, redirect to pkg)
**Estimated Time**: 4 hours

#### Current State
Basic `log.Printf` with format `[timestamp] [level] [requestID] message`

#### Implementation Steps

1. **Create structured JSON logger** - `hub/api/pkg/json_logger.go`:
```go
// Package pkg provides shared utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package pkg

import (
    "context"
    "encoding/json"
    "io"
    "os"
    "sync"
    "time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
    Timestamp   string                 `json:"timestamp"`
    Level       string                 `json:"level"`
    Message     string                 `json:"message"`
    RequestID   string                 `json:"request_id,omitempty"`
    UserID      string                 `json:"user_id,omitempty"`
    TraceID     string                 `json:"trace_id,omitempty"`
    SpanID      string                 `json:"span_id,omitempty"`
    Service     string                 `json:"service"`
    Version     string                 `json:"version,omitempty"`
    Environment string                 `json:"environment,omitempty"`
    Error       *ErrorInfo             `json:"error,omitempty"`
    Fields      map[string]interface{} `json:"fields,omitempty"`
    Duration    *float64               `json:"duration_ms,omitempty"`
}

// ErrorInfo contains error details for logging
type ErrorInfo struct {
    Type       string `json:"type"`
    Message    string `json:"message"`
    StackTrace string `json:"stack_trace,omitempty"`
}

// JSONLogger implements structured JSON logging
type JSONLogger struct {
    writer      io.Writer
    level       LogLevel
    serviceName string
    version     string
    environment string
    mu          sync.Mutex
}

// JSONLoggerConfig configures the JSON logger
type JSONLoggerConfig struct {
    Writer      io.Writer
    Level       LogLevel
    ServiceName string
    Version     string
    Environment string
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(cfg JSONLoggerConfig) *JSONLogger {
    if cfg.Writer == nil {
        cfg.Writer = os.Stdout
    }
    if cfg.ServiceName == "" {
        cfg.ServiceName = "sentinel-hub-api"
    }
    return &JSONLogger{
        writer:      cfg.Writer,
        level:       cfg.Level,
        serviceName: cfg.ServiceName,
        version:     cfg.Version,
        environment: cfg.Environment,
    }
}

// Log writes a structured log entry
func (l *JSONLogger) Log(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}) {
    if !shouldLog(level) {
        return
    }
    
    entry := LogEntry{
        Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
        Level:       string(level),
        Message:     msg,
        Service:     l.serviceName,
        Version:     l.version,
        Environment: l.environment,
        Fields:      fields,
    }
    
    // Extract context values
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
        entry.RequestID = requestID
    }
    if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
        entry.TraceID = traceID
    }
    if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
        entry.SpanID = spanID
    }
    if userID, ok := ctx.Value(UserIDKey).(string); ok {
        entry.UserID = userID
    }
    
    l.mu.Lock()
    defer l.mu.Unlock()
    
    data, _ := json.Marshal(entry)
    l.writer.Write(data)
    l.writer.Write([]byte("\n"))
}

// Info logs at INFO level
func (l *JSONLogger) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
    f := mergeFields(fields)
    l.Log(ctx, LogLevelInfo, msg, f)
}

// Error logs at ERROR level with error details
func (l *JSONLogger) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
    f := mergeFields(fields)
    if err != nil {
        f["error"] = map[string]interface{}{
            "type":    fmt.Sprintf("%T", err),
            "message": err.Error(),
        }
    }
    l.Log(ctx, LogLevelError, msg, f)
}

// Debug logs at DEBUG level
func (l *JSONLogger) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
    f := mergeFields(fields)
    l.Log(ctx, LogLevelDebug, msg, f)
}

// Warn logs at WARN level
func (l *JSONLogger) Warn(ctx context.Context, msg string, fields ...map[string]interface{}) {
    f := mergeFields(fields)
    l.Log(ctx, LogLevelWarn, msg, f)
}

func mergeFields(fields []map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for _, f := range fields {
        for k, v := range f {
            result[k] = v
        }
    }
    return result
}
```

2. **Add context keys**:
```go
type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    TraceIDKey   contextKey = "trace_id"
    SpanIDKey    contextKey = "span_id"
    UserIDKey    contextKey = "user_id"
)
```

---

### Task 4.2: Prometheus Metrics Endpoint
**Files**: 
- `hub/api/pkg/metrics/metrics.go` (new)
- `hub/api/handlers/metrics_handler.go` (new)
- `hub/api/middleware/metrics_middleware.go` (new)
**Estimated Time**: 8 hours

#### Implementation Steps

1. **Add Prometheus dependency**:
```bash
cd hub/api && go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

2. **Create metrics registry** - `hub/api/pkg/metrics/metrics.go`:
```go
// Package metrics provides Prometheus metrics
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
    // HTTP metrics
    HTTPRequestsTotal   *prometheus.CounterVec
    HTTPRequestDuration *prometheus.HistogramVec
    HTTPRequestSize     *prometheus.HistogramVec
    HTTPResponseSize    *prometheus.HistogramVec
    
    // Business metrics
    TasksCreated        prometheus.Counter
    TasksCompleted      prometheus.Counter
    DocumentsProcessed  prometheus.Counter
    ExtractionDuration  *prometheus.HistogramVec
    ExtractionConfidence *prometheus.HistogramVec
    
    // System metrics
    ActiveConnections   prometheus.Gauge
    GoroutineCount      prometheus.Gauge
    MemoryUsage         prometheus.Gauge
}

// NewMetrics creates and registers all metrics
func NewMetrics(namespace string) *Metrics {
    return &Metrics{
        HTTPRequestsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "http_requests_total",
                Help:      "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        ),
        HTTPRequestDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "http_request_duration_seconds",
                Help:      "HTTP request duration in seconds",
                Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
            },
            []string{"method", "path", "status"},
        ),
        TasksCreated: promauto.NewCounter(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "tasks_created_total",
                Help:      "Total number of tasks created",
            },
        ),
        TasksCompleted: promauto.NewCounter(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "tasks_completed_total",
                Help:      "Total number of tasks completed",
            },
        ),
        DocumentsProcessed: promauto.NewCounter(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "documents_processed_total",
                Help:      "Total number of documents processed",
            },
        ),
        ExtractionDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "extraction_duration_seconds",
                Help:      "Knowledge extraction duration in seconds",
                Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
            },
            []string{"type", "source"},
        ),
        ExtractionConfidence: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "extraction_confidence",
                Help:      "Knowledge extraction confidence scores",
                Buckets:   []float64{.1, .2, .3, .4, .5, .6, .7, .8, .9, 1.0},
            },
            []string{"type"},
        ),
        ActiveConnections: promauto.NewGauge(
            prometheus.GaugeOpts{
                Namespace: namespace,
                Name:      "active_connections",
                Help:      "Number of active HTTP connections",
            },
        ),
    }
}
```

3. **Create metrics middleware** - `hub/api/middleware/metrics_middleware.go`:
```go
// Package middleware provides HTTP middleware
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
    "net/http"
    "strconv"
    "time"
    
    "sentinel-hub-api/pkg/metrics"
)

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status
            wrapper := &responseWrapper{ResponseWriter: w, status: http.StatusOK}
            
            m.ActiveConnections.Inc()
            defer m.ActiveConnections.Dec()
            
            next.ServeHTTP(wrapper, r)
            
            duration := time.Since(start).Seconds()
            status := strconv.Itoa(wrapper.status)
            path := normalizePath(r.URL.Path)
            
            m.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
            m.HTTPRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
        })
    }
}

type responseWrapper struct {
    http.ResponseWriter
    status int
}

func (w *responseWrapper) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

func normalizePath(path string) string {
    // Normalize paths with IDs to reduce cardinality
    // e.g., /api/v1/tasks/abc123 -> /api/v1/tasks/:id
    // Implementation depends on router
    return path
}
```

4. **Add metrics endpoint** to router:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// In router setup:
r.Handle("/metrics", promhttp.Handler())
```

---

### Task 4.3: Request Tracing (Correlation IDs)
**Files**: 
- `hub/api/middleware/tracing.go` (new)
- Update existing logging to use trace context
**Estimated Time**: 4 hours

#### Implementation Steps

1. **Create tracing middleware** - `hub/api/middleware/tracing.go`:
```go
// Package middleware provides HTTP middleware
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
    "context"
    "net/http"
    
    "github.com/google/uuid"
)

type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    TraceIDKey   contextKey = "trace_id"
    SpanIDKey    contextKey = "span_id"
)

// TracingMiddleware adds correlation IDs to requests
func TracingMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract or generate trace ID
            traceID := r.Header.Get("X-Trace-ID")
            if traceID == "" {
                traceID = uuid.New().String()
            }
            
            // Extract or generate request ID
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = uuid.New().String()
            }
            
            // Generate span ID for this request
            spanID := uuid.New().String()[:8]
            
            // Add to context
            ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
            ctx = context.WithValue(ctx, RequestIDKey, requestID)
            ctx = context.WithValue(ctx, SpanIDKey, spanID)
            
            // Add to response headers for client correlation
            w.Header().Set("X-Trace-ID", traceID)
            w.Header().Set("X-Request-ID", requestID)
            w.Header().Set("X-Span-ID", spanID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
    if id, ok := ctx.Value(TraceIDKey).(string); ok {
        return id
    }
    return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}

// GetSpanID extracts span ID from context
func GetSpanID(ctx context.Context) string {
    if id, ok := ctx.Value(SpanIDKey).(string); ok {
        return id
    }
    return ""
}
```

---

### Task 4.4: Graceful Shutdown Enhancement
**Files**: `hub/api/main_minimal.go`
**Estimated Time**: 2 hours

#### Current State
Basic graceful shutdown exists but needs enhancement for:
- Connection draining
- Background job completion
- Health endpoint status

#### Implementation Steps

Update `hub/api/main_minimal.go`:
```go
// Package main - Entry point
// Complies with CODING_STANDARDS.md: Entry Points max 50 lines
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync/atomic"
    "syscall"
    "time"
    
    "sentinel-hub-api/config"
    "sentinel-hub-api/handlers"
    "sentinel-hub-api/router"
)

var isShuttingDown atomic.Bool

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    deps := handlers.NewDependencies(nil)
    r := router.NewRouter(deps, &isShuttingDown)
    
    server := &http.Server{
        Addr:         cfg.GetServerAddr(),
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }
    
    go func() {
        log.Printf("Server starting on %s", cfg.GetServerAddr())
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()
    
    gracefulShutdown(server, deps)
}

func gracefulShutdown(server *http.Server, deps *handlers.Dependencies) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Initiating graceful shutdown...")
    isShuttingDown.Store(true)
    
    // Phase 1: Stop accepting new requests (30s)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Phase 2: Complete in-flight requests
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Forced shutdown: %v", err)
    }
    
    // Phase 3: Cleanup resources
    deps.Cleanup()
    
    log.Println("Shutdown complete")
}
```

---

### Task 4.5: Kubernetes Manifests
**Files**: 
- `k8s/base/deployment.yaml`
- `k8s/base/service.yaml`
- `k8s/base/configmap.yaml`
- `k8s/base/secrets.yaml`
- `k8s/base/hpa.yaml`
- `k8s/base/kustomization.yaml`
- `k8s/overlays/production/kustomization.yaml`
**Estimated Time**: 8 hours

#### Implementation Steps

1. **Create directory structure**:
```bash
mkdir -p k8s/base k8s/overlays/{development,staging,production}
```

2. **Create base deployment** - `k8s/base/deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sentinel-hub-api
  labels:
    app: sentinel-hub-api
    version: v1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: sentinel-hub-api
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: sentinel-hub-api
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: sentinel-hub-api
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: sentinel-hub-api
        image: sentinel-hub-api:latest
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        env:
        - name: HOST
          value: "0.0.0.0"
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: sentinel-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: sentinel-secrets
              key: jwt-secret
        - name: OLLAMA_HOST
          valueFrom:
            configMapKeyRef:
              name: sentinel-config
              key: ollama-host
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      terminationGracePeriodSeconds: 30
```

3. **Create service** - `k8s/base/service.yaml`:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: sentinel-hub-api
  labels:
    app: sentinel-hub-api
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: sentinel-hub-api
```

4. **Create HPA** - `k8s/base/hpa.yaml`:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: sentinel-hub-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: sentinel-hub-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
```

5. **Create configmap** - `k8s/base/configmap.yaml`:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sentinel-config
data:
  ollama-host: "http://ollama:11434"
  log-level: "info"
  cors-origins: "https://sentinel.example.com"
  rate-limit-requests: "1000"
  rate-limit-window: "15m"
```

6. **Create kustomization** - `k8s/base/kustomization.yaml`:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: sentinel

resources:
- deployment.yaml
- service.yaml
- configmap.yaml
- hpa.yaml

commonLabels:
  app.kubernetes.io/name: sentinel-hub-api
  app.kubernetes.io/part-of: sentinel

images:
- name: sentinel-hub-api
  newTag: latest
```

7. **Create production overlay** - `k8s/overlays/production/kustomization.yaml`:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: sentinel-production

resources:
- ../../base

replicas:
- name: sentinel-hub-api
  count: 3

patches:
- patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/resources/requests/memory
      value: 512Mi
    - op: replace
      path: /spec/template/spec/containers/0/resources/limits/memory
      value: 1Gi
    - op: replace
      path: /spec/template/spec/containers/0/resources/requests/cpu
      value: 250m
    - op: replace
      path: /spec/template/spec/containers/0/resources/limits/cpu
      value: 1000m
  target:
    kind: Deployment
    name: sentinel-hub-api

images:
- name: sentinel-hub-api
  newName: registry.example.com/sentinel/sentinel-hub-api
  newTag: v1.0.0
```

---

## Summary Timeline

| Priority | Task | Estimated Hours | Dependencies |
|----------|------|-----------------|--------------|
| 1.1 | Fix TestTaskRepository_FindByID | 1 | None |
| 1.2 | Fix TestDocumentUploadProcessExtract | 1 | None |
| 2.1 | LLM-powered extraction | 8 | None |
| 2.2 | ExtractFromFile implementation | 4 | 2.1 |
| 2.3 | Confidence scoring | 4 | 2.1 |
| 3.1 | Excel support | 4 | 2.2 |
| 4.1 | Structured JSON logging | 4 | None |
| 4.2 | Prometheus metrics | 8 | 4.1 |
| 4.3 | Request tracing | 4 | 4.1 |
| 4.4 | Graceful shutdown | 2 | None |
| 4.5 | K8s manifests | 8 | 4.2, 4.4 |

**Total Estimated Time**: ~48 hours (6 working days)

---

## Verification Checklist

- [ ] All tests pass: `go test ./... -v`
- [ ] No linter errors: `golangci-lint run`
- [ ] File line counts within limits
- [ ] Constructor injection used for all dependencies
- [ ] Context propagation for tracing
- [ ] Metrics exposed at `/metrics`
- [ ] Health endpoints return correct status during shutdown
- [ ] K8s manifests validated: `kubectl apply --dry-run=client -k k8s/base/`

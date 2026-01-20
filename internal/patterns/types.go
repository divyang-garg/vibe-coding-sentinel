// Package patterns provides pattern learning types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package patterns

// PatternData represents learned patterns from codebase
type PatternData struct {
	Languages        map[string]int      `json:"languages"`
	Frameworks       map[string]int      `json:"frameworks"`
	NamingPatterns   map[string]int      `json:"namingPatterns"`
	FileExtensions   map[string]int      `json:"fileExtensions"`
	ProjectStructure map[string][]string `json:"projectStructure"`
	ImportPatterns   ImportPatternData   `json:"importPatterns,omitempty"`
	CodeStyle        CodeStyleData       `json:"codeStyle,omitempty"`
	BusinessRules    *BusinessRuleData   `json:"businessRules,omitempty"`
}

// ImportPatternData represents import pattern analysis
type ImportPatternData struct {
	Style          string         `json:"style"`          // "absolute", "relative", "mixed"
	Aliasing       map[string]int `json:"aliasing"`       // import aliasing patterns
	BarrelFiles    []string       `json:"barrelFiles"`    // index.ts, index.js files
	DefaultImports int            `json:"defaultImports"` // count
	NamedImports   int            `json:"namedImports"`   // count
	Examples       []string       `json:"examples"`       // sample import statements
}

// CodeStyleData represents code style patterns
type CodeStyleData struct {
	IndentStyle   string `json:"indentStyle"`   // "spaces", "tabs"
	IndentSize    int    `json:"indentSize"`    // 2, 4, etc.
	QuoteStyle    string `json:"quoteStyle"`    // "single", "double"
	Semicolons    string `json:"semicolons"`    // "always", "never", "optional"
	LineEnding    string `json:"lineEnding"`    // "lf", "crlf"
	TrailingComma string `json:"trailingComma"` // "always", "never", "es5"
}

// BusinessRuleData represents business rules fetched from Hub
type BusinessRuleData struct {
	Rules []BusinessRule `json:"rules"`
}

// BusinessRule represents a single business rule from Hub
type BusinessRule struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
	SourcePage *int    `json:"source_page,omitempty"`
}

// NewPatternData creates a new PatternData instance
func NewPatternData() *PatternData {
	return &PatternData{
		Languages:        make(map[string]int),
		Frameworks:       make(map[string]int),
		NamingPatterns:   make(map[string]int),
		FileExtensions:   make(map[string]int),
		ProjectStructure: make(map[string][]string),
		ImportPatterns: ImportPatternData{
			Aliasing: make(map[string]int),
			Examples: []string{},
		},
	}
}

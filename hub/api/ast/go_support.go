// Package ast provides Go language support implementation
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

// GoLanguageSupport provides complete Go language support
type GoLanguageSupport struct {
	*BaseLanguageSupport
}

// NewGoLanguageSupport creates a new Go language support instance
func NewGoLanguageSupport() *GoLanguageSupport {
	return &GoLanguageSupport{
		BaseLanguageSupport: &BaseLanguageSupport{
			Language:  "go",
			Detector:  &GoDetector{},
			Extractor: &GoExtractor{},
			NodeTypes: LanguageNodeTypes{
				FunctionDeclaration: []string{"function_declaration"},
				MethodDeclaration:   []string{"method_declaration"},
				VariableDeclaration: []string{"var_declaration", "short_var_declaration"},
				ClassDeclaration:    []string{}, // Go doesn't have classes
				ImportStatement:     []string{"import_declaration"},
			},
		},
	}
}

// init registers Go language support
func init() {
	goSupport := NewGoLanguageSupport()
	if err := RegisterLanguageSupport(goSupport); err != nil {
		// Registration should not fail for built-in language
		// Log error if logging is available
		_ = err
	}
}

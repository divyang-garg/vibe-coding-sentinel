// Package ast provides Python language support
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

// PythonLanguageSupport provides Python language support
type PythonLanguageSupport struct {
	*BaseLanguageSupport
}

// NewPythonLanguageSupport creates Python language support
func NewPythonLanguageSupport() *PythonLanguageSupport {
	return &PythonLanguageSupport{
		BaseLanguageSupport: &BaseLanguageSupport{
			Language:  "python",
			Detector:  &PythonDetector{},
			Extractor: &PythonExtractor{},
			NodeTypes: LanguageNodeTypes{
				FunctionDeclaration: []string{"function_definition"},
				MethodDeclaration:   []string{"function_definition"},
				VariableDeclaration: []string{},
				ClassDeclaration:    []string{"class_definition"},
				ImportStatement:     []string{"import_statement", "import_from_statement"},
			},
		},
	}
}

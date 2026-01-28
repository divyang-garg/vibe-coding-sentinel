// Package ast provides JavaScript/TypeScript language support
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

// JsLanguageSupport provides JavaScript language support
type JsLanguageSupport struct {
	*BaseLanguageSupport
}

// NewJsLanguageSupport creates JavaScript language support
func NewJsLanguageSupport() *JsLanguageSupport {
	return &JsLanguageSupport{
		BaseLanguageSupport: &BaseLanguageSupport{
			Language:  "javascript",
			Detector:  &JsDetector{},
			Extractor: &JsExtractor{Lang: "javascript"},
			NodeTypes: LanguageNodeTypes{
				FunctionDeclaration: []string{"function_declaration", "function"},
				MethodDeclaration:   []string{"method_definition"},
				VariableDeclaration: []string{"lexical_declaration", "variable_declaration"},
				ClassDeclaration:    []string{"class_declaration"},
				ImportStatement:     []string{"import_statement", "import_declaration"},
			},
		},
	}
}

// TsLanguageSupport provides TypeScript language support (same detector/extractor as JS)
type TsLanguageSupport struct {
	*BaseLanguageSupport
}

// NewTsLanguageSupport creates TypeScript language support
func NewTsLanguageSupport() *TsLanguageSupport {
	return &TsLanguageSupport{
		BaseLanguageSupport: &BaseLanguageSupport{
			Language:  "typescript",
			Detector:  &JsDetector{},
			Extractor: &JsExtractor{Lang: "typescript"},
			NodeTypes: LanguageNodeTypes{
				FunctionDeclaration: []string{"function_declaration", "function"},
				MethodDeclaration:   []string{"method_definition"},
				VariableDeclaration: []string{"lexical_declaration", "variable_declaration"},
				ClassDeclaration:    []string{"class_declaration"},
				ImportStatement:     []string{"import_statement", "import_declaration"},
			},
		},
	}
}

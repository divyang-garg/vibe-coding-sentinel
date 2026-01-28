// Package ast provides base implementation for language support
// Complies with CODING_STANDARDS.md: Base implementations max 200 lines
package ast

// BaseLanguageSupport provides default implementations that can be embedded
// in language-specific implementations
type BaseLanguageSupport struct {
	Language  string
	Detector  LanguageDetector
	Extractor LanguageExtractor
	NodeTypes LanguageNodeTypes
}

// GetLanguage returns the language name
func (b *BaseLanguageSupport) GetLanguage() string {
	return b.Language
}

// GetDetector returns the language detector
func (b *BaseLanguageSupport) GetDetector() LanguageDetector {
	return b.Detector
}

// GetExtractor returns the language extractor
func (b *BaseLanguageSupport) GetExtractor() LanguageExtractor {
	return b.Extractor
}

// GetNodeTypes returns the language-specific node types
func (b *BaseLanguageSupport) GetNodeTypes() LanguageNodeTypes {
	return b.NodeTypes
}

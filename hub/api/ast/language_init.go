// Package ast provides language initialization
// Complies with CODING_STANDARDS.md: Initialization modules max 200 lines
package ast

import "fmt"

// init registers all built-in language support
func init() {
	// Register Go language support
	goSupport := NewGoLanguageSupport()
	if err := RegisterLanguageSupport(goSupport); err != nil {
		panic(fmt.Sprintf("Failed to register Go language support: %v", err))
	}
	// Register JavaScript language support
	jsSupport := NewJsLanguageSupport()
	if err := RegisterLanguageSupport(jsSupport); err != nil {
		panic(fmt.Sprintf("Failed to register JavaScript language support: %v", err))
	}
	// Register TypeScript language support (same detector/extractor as JS)
	tsSupport := NewTsLanguageSupport()
	if err := RegisterLanguageSupport(tsSupport); err != nil {
		panic(fmt.Sprintf("Failed to register TypeScript language support: %v", err))
	}
	// Register Python language support
	pySupport := NewPythonLanguageSupport()
	if err := RegisterLanguageSupport(pySupport); err != nil {
		panic(fmt.Sprintf("Failed to register Python language support: %v", err))
	}
}

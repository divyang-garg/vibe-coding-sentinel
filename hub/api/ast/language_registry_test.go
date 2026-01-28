// Package ast provides tests for language registry
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package ast

import (
	"context"
	"testing"
)

func TestLanguageRegistry_Register(t *testing.T) {
	// Create a test language support
	support := &BaseLanguageSupport{
		Language:  "test",
		Detector:  &GoDetector{}, // Reuse for testing
		Extractor: &GoExtractor{},
		NodeTypes: LanguageNodeTypes{},
	}

	err := RegisterLanguageSupport(support)
	if err != nil {
		t.Fatalf("Failed to register language support: %v", err)
	}

	// Verify it's registered
	retrieved := GetLanguageSupport("test")
	if retrieved == nil {
		t.Error("Expected to retrieve registered language support")
	}

	if retrieved.GetLanguage() != "test" {
		t.Errorf("Expected language 'test', got %s", retrieved.GetLanguage())
	}
}

func TestLanguageRegistry_DuplicateRegistration(t *testing.T) {
	support1 := &BaseLanguageSupport{
		Language:  "duplicate",
		Detector:  &GoDetector{},
		Extractor: &GoExtractor{},
		NodeTypes: LanguageNodeTypes{},
	}

	support2 := &BaseLanguageSupport{
		Language:  "duplicate",
		Detector:  &GoDetector{},
		Extractor: &GoExtractor{},
		NodeTypes: LanguageNodeTypes{},
	}

	// First registration should succeed
	err := RegisterLanguageSupport(support1)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Second registration should fail
	err = RegisterLanguageSupport(support2)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestLanguageRegistry_GetDetector(t *testing.T) {
	// Go should be registered by init()
	detector := GetLanguageDetector("go")
	if detector == nil {
		t.Error("Expected Go detector to be available")
	}

	// Test unsupported language
	detector = GetLanguageDetector("nonexistent")
	if detector != nil {
		t.Error("Expected nil for unsupported language")
	}
}

func TestLanguageRegistry_GetExtractor(t *testing.T) {
	// Go should be registered by init()
	extractor := GetLanguageExtractor("go")
	if extractor == nil {
		t.Error("Expected Go extractor to be available")
	}

	// Test unsupported language
	extractor = GetLanguageExtractor("nonexistent")
	if extractor != nil {
		t.Error("Expected nil for unsupported language")
	}
}

func TestLanguageRegistry_GetSupportedLanguages(t *testing.T) {
	languages := GetSupportedLanguages()

	// Should include Go (registered by init())
	found := false
	for _, lang := range languages {
		if lang == "go" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'go' to be in supported languages")
	}
}

func TestLanguageRegistry_IsLanguageSupported(t *testing.T) {
	// Go should be supported
	if !IsLanguageSupported("go") {
		t.Error("Expected Go to be supported")
	}

	// Unsupported language
	if IsLanguageSupported("nonexistent") {
		t.Error("Expected nonexistent language to not be supported")
	}
}

func TestLanguageRegistry_NilSupport(t *testing.T) {
	err := RegisterLanguageSupport(nil)
	if err == nil {
		t.Error("Expected error for nil language support")
	}
}

func TestLanguageRegistry_EmptyLanguageName(t *testing.T) {
	support := &BaseLanguageSupport{
		Language:  "", // Empty name
		Detector:  &GoDetector{},
		Extractor: &GoExtractor{},
		NodeTypes: LanguageNodeTypes{},
	}

	err := RegisterLanguageSupport(support)
	if err == nil {
		t.Error("Expected error for empty language name")
	}
}

// TestDetectionFallback_UnsupportedLanguage verifies that detection entry points
// do not panic and return empty (or fallback) when language has no registered detector.
func TestDetectionFallback_UnsupportedLanguage(t *testing.T) {
	parser, err := GetParser("go")
	if err != nil {
		t.Fatalf("GetParser(go): %v", err)
	}
	code := "package main\nfunc main() {}"
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil || tree == nil {
		t.Fatalf("parse: %v", err)
	}
	defer tree.Close()
	root := tree.RootNode()
	if root == nil {
		t.Fatal("root is nil")
	}

	// Unsupported language: no detector registered, fallback switch has no case -> empty
	findings := detectUnusedVariables(root, code, "nonexistent")
	if findings == nil {
		t.Error("expected non-nil slice")
	}
	if len(findings) != 0 {
		t.Errorf("unsupported language should return empty findings, got %d", len(findings))
	}

	findingsDup := detectDuplicateFunctions(root, code, "nonexistent")
	if findingsDup == nil {
		t.Error("expected non-nil slice")
	}
	if len(findingsDup) != 0 {
		t.Errorf("unsupported language duplicates should be empty, got %d", len(findingsDup))
	}

	findingsUnreach := detectUnreachableCode(root, code, "nonexistent")
	if findingsUnreach == nil {
		t.Error("expected non-nil slice")
	}
	if len(findingsUnreach) != 0 {
		t.Errorf("unsupported language unreachable should be empty, got %d", len(findingsUnreach))
	}

	findingsAwait := detectMissingAwait(root, code, "nonexistent")
	if findingsAwait == nil {
		t.Error("expected non-nil slice")
	}
	if len(findingsAwait) != 0 {
		t.Errorf("unsupported language missing_await should be empty, got %d", len(findingsAwait))
	}

	vulns := detectSQLInjection(root, code, "nonexistent")
	if vulns == nil {
		t.Error("expected non-nil slice")
	}
	if len(vulns) != 0 {
		t.Errorf("unsupported language SQL injection should be empty, got %d", len(vulns))
	}

	vulnsXSS := detectXSS(root, code, "nonexistent")
	if vulnsXSS == nil {
		t.Error("expected non-nil slice")
	}
	if len(vulnsXSS) != 0 {
		t.Errorf("unsupported language XSS should be empty, got %d", len(vulnsXSS))
	}

	vulnsCmd := detectCommandInjection(root, code, "nonexistent")
	if vulnsCmd == nil {
		t.Error("expected non-nil slice")
	}
	if len(vulnsCmd) != 0 {
		t.Errorf("unsupported language command injection should be empty, got %d", len(vulnsCmd))
	}

	vulnsCrypto := detectInsecureCrypto(root, code, "nonexistent")
	if vulnsCrypto == nil {
		t.Error("expected non-nil slice")
	}
	if len(vulnsCrypto) != 0 {
		t.Errorf("unsupported language crypto should be empty, got %d", len(vulnsCrypto))
	}
}

// TestGetLanguageDetector_NilForUnsupported verifies GetLanguageDetector returns nil for unregistered language.
func TestGetLanguageDetector_NilForUnsupported(t *testing.T) {
	for _, lang := range []string{"java", "ruby", "nonexistent", ""} {
		d := GetLanguageDetector(lang)
		if d != nil {
			t.Errorf("GetLanguageDetector(%q) should be nil, got %T", lang, d)
		}
	}
}

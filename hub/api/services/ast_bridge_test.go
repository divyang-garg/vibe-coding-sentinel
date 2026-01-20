// Package services AST bridge tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestGetParser_SupportedLanguages(t *testing.T) {
	languages := []string{"go", "javascript", "typescript", "python"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			parser, err := getParser(lang)
			if err != nil {
				t.Fatalf("getParser(%q) failed: %v", lang, err)
			}
			if parser == nil {
				t.Fatalf("getParser(%q) returned nil parser", lang)
			}
		})
	}
}

func TestGetParser_LanguageAliases(t *testing.T) {
	testCases := []struct {
		alias string
		valid bool
	}{
		{"js", true},
		{"ts", true},
		{"py", true},
		{"golang", true},
		{"unsupported", false},
	}

	for _, tc := range testCases {
		t.Run(tc.alias, func(t *testing.T) {
			parser, err := getParser(tc.alias)
			if tc.valid {
				if err != nil {
					t.Fatalf("getParser(%q) should succeed but failed: %v", tc.alias, err)
				}
				if parser == nil {
					t.Fatalf("getParser(%q) returned nil parser", tc.alias)
				}
			} else {
				if err == nil {
					t.Errorf("getParser(%q) should fail but succeeded", tc.alias)
				}
			}
		})
	}
}

func TestGetParser_Cache(t *testing.T) {
	lang := "go"

	// First call
	parser1, err1 := getParser(lang)
	if err1 != nil {
		t.Fatalf("First getParser failed: %v", err1)
	}
	if parser1 == nil {
		t.Fatal("First getParser returned nil parser")
	}

	// Second call should return cached parser
	parser2, err2 := getParser(lang)
	if err2 != nil {
		t.Fatalf("Second getParser failed: %v", err2)
	}
	if parser2 == nil {
		t.Fatal("Second getParser returned nil parser")
	}

	// Should be the same parser instance (cached)
	if parser1 != parser2 {
		t.Error("Expected same parser instance from cache")
	}
}

func TestAnalyzeCode(t *testing.T) {
	code := `
package main

func test() {
    fmt.Println("test")
}
`

	findings, stats, err := AnalyzeCode(code, "go", []string{})
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	if stats.NodesVisited == 0 {
		t.Error("Expected NodesVisited > 0")
	}

	t.Logf("AnalyzeCode: %d findings, %d nodes", len(findings), stats.NodesVisited)
}

func TestTraverseAST(t *testing.T) {
	// Create a simple parser and parse code
	parser, err := getParser("go")
	if err != nil {
		t.Fatalf("getParser failed: %v", err)
	}
	if parser == nil {
		t.Fatal("getParser returned nil parser")
	}

	code := `func test() {}`
	tree, err := parser.ParseCtx(context.TODO(), nil, []byte(code))
	if err != nil {
		t.Fatalf("ParseCtx failed: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("RootNode is nil")
	}

	// Traverse and count nodes
	nodeCount := 0
	traverseAST(rootNode, func(node *sitter.Node) bool {
		nodeCount++
		return true
	})

	if nodeCount == 0 {
		t.Error("Expected at least one node")
	}

	t.Logf("Traversed %d nodes", nodeCount)
}

func TestGetLineColumn(t *testing.T) {
	testCases := []struct {
		name       string
		code       string
		offset     int
		wantLine   int
		wantColumn int
	}{
		{
			name:       "start of first line",
			code:       "hello",
			offset:     0,
			wantLine:   1,
			wantColumn: 1,
		},
		{
			name:       "middle of first line",
			code:       "hello world",
			offset:     6,
			wantLine:   1,
			wantColumn: 7,
		},
		{
			name:       "second line",
			code:       "line1\nline2",
			offset:     7,
			wantLine:   2,
			wantColumn: 2,
		},
		{
			name:       "offset at newline",
			code:       "line1\nline2",
			offset:     5,
			wantLine:   1,
			wantColumn: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line, col := getLineColumn(tc.code, tc.offset)
			if line != tc.wantLine || col != tc.wantColumn {
				t.Errorf("getLineColumn(%q, %d) = (%d, %d), want (%d, %d)",
					tc.code, tc.offset, line, col, tc.wantLine, tc.wantColumn)
			}
		})
	}
}

func TestGetLineColumn_EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		code   string
		offset int
	}{
		{"negative offset", "hello", -1},
		{"offset beyond code", "hello", 100},
		{"empty code", "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line, col := getLineColumn(tc.code, tc.offset)
			if line < 1 {
				t.Errorf("getLineColumn returned invalid line: %d", line)
			}
			if col < 1 {
				t.Errorf("getLineColumn returned invalid column: %d", col)
			}
			t.Logf("getLineColumn(%q, %d) = (%d, %d)", tc.code, tc.offset, line, col)
		})
	}
}

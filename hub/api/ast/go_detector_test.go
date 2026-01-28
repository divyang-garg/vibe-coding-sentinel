// Package ast provides shared test helpers for Go detector tests
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package ast

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

// Helper function to parse Go code and get root node
// Note: The tree must remain alive while using the root node
// This function returns both tree and root node - caller must defer tree.Close()
func parseGoCode(t *testing.T, code string) (*sitter.Tree, *sitter.Node) {
	parser, err := GetParser("go")
	if err != nil {
		t.Fatalf("Failed to get Go parser: %v", err)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		tree.Close()
		t.Fatalf("Failed to get root node")
	}

	return tree, rootNode
}

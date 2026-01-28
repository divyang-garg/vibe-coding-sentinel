// Package architecture_edges - Edge building for module dependency graph
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"context"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"
)

// buildModuleGraphEdges extracts dependencies from files, resolves import paths,
// and returns deduplicated ModuleEdges. codebasePath is reserved for future
// use (e.g. Go module root for package-path resolution); nodes are unused.
func buildModuleGraphEdges(files []FileContent, codebasePath string, nodes []ModuleNode) []ModuleEdge {
	_ = codebasePath
	_ = nodes
	return extractAndResolveEdges(files)
}

func extractAndResolveEdges(files []FileContent) []ModuleEdge {
	pathSet := make(pathSet)
	for _, f := range files {
		pathSet[filepath.Clean(f.Path)] = true
	}

	edgeMap := make(map[string]map[string]bool) // from -> { to -> true }
	addEdge := func(from, to, edgeType string) {
		if from == "" || to == "" || from == to {
			return
		}
		if edgeMap[from] == nil {
			edgeMap[from] = make(map[string]bool)
		}
		edgeMap[from][to] = true
		_ = edgeType
	}

	ctx := context.Background()
	for _, f := range files {
		parser, err := getParser(f.Language)
		if err != nil || parser == nil {
			continue
		}
		tree, err := parser.ParseCtx(ctx, nil, []byte(f.Content))
		if err != nil || tree == nil {
			continue
		}
		root := tree.RootNode()
		if root == nil {
			tree.Close()
			continue
		}

		lang := normalizeArchitectureLanguage(f.Language)
		deps, err := ast.ExtractDependenciesFromFile(root, f.Content, f.Path, lang)
		tree.Close()
		if err != nil {
			continue
		}

		fromClean := filepath.Clean(f.Path)
		for _, d := range deps {
			targets := resolveImport(f.Path, d.ToFile, lang, pathSet)
			for _, t := range targets {
				addEdge(fromClean, filepath.Clean(t), d.Type)
			}
		}
	}

	var out []ModuleEdge
	for from, toSet := range edgeMap {
		for to := range toSet {
			out = append(out, ModuleEdge{From: from, To: to, Type: "import"})
		}
	}
	return out
}

func normalizeArchitectureLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "golang":
		return "go"
	case "js", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "py":
		return "python"
	default:
		return lang
	}
}

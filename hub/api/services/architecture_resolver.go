// Package architecture_resolver - Import path to file path resolution for architecture analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"path/filepath"
	"strings"
)

// pathSet maps normalized file path -> true for fast lookup
type pathSet map[string]bool

// resolveImport resolves an import spec (ToFile from AST) to zero or more file paths
// in the analyzed set. Returns nil for externals or unresolvable imports.
func resolveImport(fromPath, toImport, language string, paths pathSet) []string {
	if toImport == "" || paths == nil {
		return nil
	}
	toImport = strings.TrimSpace(toImport)
	dir := filepath.Dir(fromPath)

	switch language {
	case "go":
		return resolveGoImport(toImport, paths)
	case "javascript", "typescript":
		return resolveJSImport(dir, toImport, paths)
	case "python":
		return resolvePythonImport(dir, toImport, paths)
	default:
		return nil
	}
}

// resolveGoImport resolves a Go package path to files in the analyzed set.
// Skips stdlib and external packages (e.g. "fmt", "strings", "github.com/...").
func resolveGoImport(pkgPath string, paths pathSet) []string {
	if skipGoImport(pkgPath) {
		return nil
	}
	// Use last segment of package path (e.g. a/b/services -> services)
	last := pkgPath
	if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
		last = pkgPath[idx+1:]
	}
	seen := make(map[string]bool)
	var out []string
	for p := range paths {
		if seen[p] {
			continue
		}
		norm := filepath.ToSlash(p)
		dir := filepath.ToSlash(filepath.Dir(norm))
		if strings.Contains(norm, "/"+last+"/") || strings.HasSuffix(dir, "/"+last) {
			out = append(out, p)
			seen[p] = true
		}
	}
	return out
}

func skipGoImport(p string) bool {
	// Standard library: no slashes or well-known prefixes
	if !strings.Contains(p, "/") {
		return true
	}
	// Third-party: assume we only care about module-internal paths
	// Skip vanity URLs etc. if they look external
	if strings.HasPrefix(p, "github.com/") || strings.HasPrefix(p, "golang.org/") ||
		strings.HasPrefix(p, "google.golang.org/") || strings.HasPrefix(p, "gopkg.in/") {
		return true
	}
	return false
}

// resolveJSImport resolves JS/TS import (relative or bare) to file paths.
func resolveJSImport(fromDir, imp string, paths pathSet) []string {
	if imp == "" {
		return nil
	}
	// Relative: ./x, ../y
	if strings.HasPrefix(imp, ".") {
		base := filepath.Clean(filepath.Join(fromDir, imp))
		return matchPaths(base, paths, []string{".js", ".ts", ".jsx", ".tsx", ""})
	}
	// Bare specifier (node_modules): skip for V1
	if !strings.Contains(imp, ".") && !strings.HasPrefix(imp, "/") {
		return nil
	}
	return nil
}

// resolvePythonImport resolves Python import to file paths.
func resolvePythonImport(fromDir, imp string, paths pathSet) []string {
	if imp == "" {
		return nil
	}
	if strings.HasPrefix(imp, ".") {
		switch imp {
		case ".":
			return matchPaths(filepath.Clean(fromDir), paths, []string{".py", ""})
		case "..":
			return matchPaths(filepath.Clean(filepath.Join(fromDir, "..")), paths, []string{".py", ""})
		}
		// .x, .x.y: strip leading dots, then dots → path sep
		rel := strings.TrimLeft(imp, ".")
		rel = strings.TrimPrefix(rel, "/")
		base := filepath.Join(fromDir, strings.ReplaceAll(rel, ".", string(filepath.Separator)))
		return matchPaths(filepath.Clean(base), paths, []string{".py", ""})
	}
	// Dotted: a.b.c -> a/b/c
	base := filepath.Join(fromDir, strings.ReplaceAll(imp, ".", string(filepath.Separator)))
	return matchPaths(filepath.Clean(base), paths, []string{".py", ""})
}

// matchPaths checks base + ext and base + "/index" + ext against pathSet; returns matches.
func matchPaths(base string, paths pathSet, extSuffixes []string) []string {
	var out []string
	baseSlash := filepath.ToSlash(base)
	added := make(map[string]bool)
	for p := range paths {
		if added[p] {
			continue
		}
		norm := filepath.ToSlash(p)
		for _, ext := range extSuffixes {
			if ext != "" {
				if norm == baseSlash+ext || norm == baseSlash+"/index"+ext {
					out = append(out, p)
					added[p] = true
					break
				}
			} else {
				if norm == baseSlash || strings.HasPrefix(norm, baseSlash+"/") {
					out = append(out, p)
					added[p] = true
					break
				}
			}
		}
	}
	return out
}

// Package ast provides codebase search utilities
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SearchResult represents a single search match in the codebase
type SearchResult struct {
	FilePath string
	Line     int
	Column   int
	Content  string
}

// ImportInfo represents import information for a file
type ImportInfo struct {
	Path      string
	Alias     string
	IsPackage bool
}

// cachedSearchResult stores cached search results
type cachedSearchResult struct {
	Results []SearchResult
	Expires time.Time
}

var searchCache = make(map[string]cachedSearchResult)
var searchCacheMutex sync.RWMutex
var searchCacheTTL = 5 * time.Minute

// SearchCodebase searches for a pattern across the codebase
// Uses grep/ripgrep for fast searching, excluding common directories
func SearchCodebase(pattern, projectRoot string, excludeDirs []string) ([]SearchResult, error) {
	// Default exclusions
	defaultExcludes := []string{".git", "node_modules", "vendor", "testdata", ".cursor"}
	excludeDirs = append(excludeDirs, defaultExcludes...)

	// Try ripgrep first (faster), fallback to grep
	var cmd *exec.Cmd
	if _, err := exec.LookPath("rg"); err == nil {
		// Use ripgrep
		args := []string{"-n", "--type", "go", "--type", "js", "--type", "ts", "--type", "py"}
		for _, exclude := range excludeDirs {
			args = append(args, "--glob", "!"+exclude+"/**")
		}
		args = append(args, pattern, projectRoot)
		cmd = exec.Command("rg", args...)
	} else {
		// Fallback to grep with -E for extended regex
		args := []string{"-rnE", "--include=*.go", "--include=*.js", "--include=*.ts", "--include=*.py"}
		for _, exclude := range excludeDirs {
			args = append(args, "--exclude-dir="+exclude)
		}
		args = append(args, "-e", pattern, projectRoot)
		cmd = exec.Command("grep", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		// grep returns non-zero exit code when no matches found - that's OK
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []SearchResult{}, nil
		}
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results := []SearchResult{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: filepath:line:column:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		lineNum := 0
		fmt.Sscanf(parts[1], "%d", &lineNum)
		content := parts[2]

		results = append(results, SearchResult{
			FilePath: filePath,
			Line:     lineNum,
			Content:  content,
		})
	}

	return results, nil
}

// SearchCodebaseCached searches with result caching
func SearchCodebaseCached(pattern, projectRoot string, cacheTTL time.Duration) ([]SearchResult, error) {
	if cacheTTL <= 0 {
		cacheTTL = searchCacheTTL
	}

	// Generate cache key
	key := generateSearchCacheKey(pattern, projectRoot)

	// Check cache
	searchCacheMutex.RLock()
	if cached, ok := searchCache[key]; ok {
		if time.Now().Before(cached.Expires) {
			searchCacheMutex.RUnlock()
			return cached.Results, nil
		}
		// Cache expired, remove it
		delete(searchCache, key)
	}
	searchCacheMutex.RUnlock()

	// Perform search
	results, err := SearchCodebase(pattern, projectRoot, nil)
	if err != nil {
		return nil, err
	}

	// Cache results
	searchCacheMutex.Lock()
	searchCache[key] = cachedSearchResult{
		Results: results,
		Expires: time.Now().Add(cacheTTL),
	}

	// Clean up expired entries if cache is too large
	if len(searchCache) > 1000 {
		cleanExpiredSearchCache()
	}
	searchCacheMutex.Unlock()

	return results, nil
}

// generateSearchCacheKey generates a cache key for search
func generateSearchCacheKey(pattern, projectRoot string) string {
	h := md5.New()
	h.Write([]byte(pattern))
	h.Write([]byte(projectRoot))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// cleanExpiredSearchCache removes expired cache entries
func cleanExpiredSearchCache() {
	now := time.Now()
	for key, entry := range searchCache {
		if now.After(entry.Expires) {
			delete(searchCache, key)
		}
	}
}

// CountReferences counts how many times an identifier appears in the codebase
func CountReferences(identifier, projectRoot string) (int, error) {
	// Search for identifier as a word boundary match
	pattern := fmt.Sprintf("\\b%s\\b", identifier)
	results, err := SearchCodebase(pattern, projectRoot, nil)
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// FindImports extracts import information from a Go file
func FindImports(filePath, projectRoot string) ([]ImportInfo, error) {
	fullPath := filepath.Join(projectRoot, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	imports := []ImportInfo{}
	scanner := bufio.NewScanner(file)
	inImportBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Start of import block
		if strings.HasPrefix(line, "import (") {
			inImportBlock = true
			continue
		}

		// End of import block
		if inImportBlock && line == ")" {
			inImportBlock = false
			continue
		}

		// Single import statement
		if strings.HasPrefix(line, "import ") {
			importStr := strings.TrimPrefix(line, "import ")
			importStr = strings.Trim(importStr, "\"")
			imports = append(imports, ImportInfo{
				Path:      importStr,
				IsPackage: true,
			})
			continue
		}

		// Import within block
		if inImportBlock {
			line = strings.Trim(line, "\"")
			if line != "" {
				parts := strings.Fields(line)
				info := ImportInfo{IsPackage: true}
				if len(parts) == 2 {
					info.Alias, info.Path = parts[0], strings.Trim(parts[1], "\"")
				} else if len(parts) == 1 {
					info.Path = strings.Trim(parts[0], "\"")
				}
				if info.Path != "" {
					imports = append(imports, info)
				}
			}
		}
	}

	return imports, scanner.Err()
}

// CheckIntentComment checks if there's an intent comment (TODO, FIXME, etc.) near a line
func CheckIntentComment(filePath string, line int, projectRoot string) bool {
	fullPath := filepath.Join(projectRoot, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 0
	intentKeywords := []string{"TODO", "FIXME", "HACK", "XXX", "NOTE", "intentional", "ignore"}

	// Check lines around the target (within 3 lines)
	for scanner.Scan() {
		currentLine++
		if currentLine < line-3 || currentLine > line+3 {
			continue
		}

		lineText := strings.ToUpper(scanner.Text())
		for _, keyword := range intentKeywords {
			if strings.Contains(lineText, keyword) {
				return true
			}
		}
	}

	return false
}

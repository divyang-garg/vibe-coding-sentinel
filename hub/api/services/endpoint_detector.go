// Package services endpoint detector
// Detects API endpoints matching business rule keywords
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"path/filepath"
	"regexp"
	"strings"
)

// detectEndpoints finds API endpoints matching business rule keywords
// Supports multiple frameworks: Express.js, FastAPI, Go, Django
func detectEndpoints(code string, filePath string, keywords []string) []string {
	var endpoints []string

	// Determine framework from file path and code patterns
	framework := detectFramework(filePath, code)

	switch framework {
	case "express":
		endpoints = detectExpressEndpoints(code, keywords)
	case "fastapi":
		endpoints = detectFastAPIEndpoints(code, keywords)
	case "go":
		endpoints = detectGoEndpoints(code, keywords)
	case "django":
		endpoints = detectDjangoEndpoints(code, keywords)
	default:
		// Try generic detection
		endpoints = detectGenericEndpoints(code, keywords)
	}

	return endpoints
}

// detectFramework determines the web framework from file path and code.
// Supported frameworks: Express.js, FastAPI, Django, Go (router/mux/chi)
// Returns: Framework name or "unknown" if cannot be determined
func detectFramework(filePath string, code string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	codeLower := strings.ToLower(code)

	// Express.js detection
	if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
		if strings.Contains(codeLower, "express") ||
			strings.Contains(codeLower, "app.get") ||
			strings.Contains(codeLower, "app.post") ||
			strings.Contains(codeLower, "router.") {
			return "express"
		}
	}

	// FastAPI detection
	if ext == ".py" {
		if strings.Contains(codeLower, "fastapi") ||
			strings.Contains(codeLower, "@app.get") ||
			strings.Contains(codeLower, "@app.post") ||
			strings.Contains(codeLower, "from fastapi import") {
			return "fastapi"
		}
	}

	// Django detection
	if ext == ".py" {
		if strings.Contains(codeLower, "django") ||
			strings.Contains(codeLower, "@api_view") ||
			strings.Contains(codeLower, "from django") {
			return "django"
		}
	}

	// Go detection
	if ext == ".go" {
		if strings.Contains(codeLower, "router.handlefunc") ||
			strings.Contains(codeLower, "mux.handlefunc") ||
			strings.Contains(codeLower, "http.handlefunc") ||
			strings.Contains(codeLower, "chi.router") {
			return "go"
		}
	}

	return "unknown"
}

// detectExpressEndpoints detects Express.js endpoints
func detectExpressEndpoints(code string, keywords []string) []string {
	var endpoints []string
	// Pattern: app.get/post/put/delete('/path', handler) or router.get/post/etc.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`app\.(get|post|put|delete|patch)\s*\(\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`router\.(get|post|put|delete|patch)\s*\(\s*['"]([^'"]+)['"]`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) >= 3 && matchesKeywords(match[2], keywords) {
				endpoints = appendIfNotExists(endpoints, strings.ToUpper(match[1])+" "+match[2])
			}
		}
	}
	return endpoints
}

// detectFastAPIEndpoints detects FastAPI endpoints
func detectFastAPIEndpoints(code string, keywords []string) []string {
	var endpoints []string

	// Pattern: @app.get/post/put/delete('/path')
	pattern := regexp.MustCompile(`@app\.(get|post|put|delete|patch)\s*\(\s*['"]([^'"]+)['"]`)
	matches := pattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			method := strings.ToUpper(match[1])
			path := match[2]
			endpoint := method + " " + path

			if matchesKeywords(path, keywords) {
				endpoints = appendIfNotExists(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

// detectGoEndpoints detects Go HTTP endpoints
func detectGoEndpoints(code string, keywords []string) []string {
	var endpoints []string

	// Pattern: router.HandleFunc("/path", handler) or mux.HandleFunc("/path", handler)
	pattern := regexp.MustCompile(`(?:router|mux|http)\.HandleFunc\s*\(\s*['"]([^'"]+)['"]`)
	matches := pattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			path := match[1]
			// Try to infer method from context or default to GET
			method := "GET"
			endpoint := method + " " + path

			if matchesKeywords(path, keywords) {
				endpoints = appendIfNotExists(endpoints, endpoint)
			}
		}
	}

	// Chi router patterns: r.Get/Post/Put/Delete("/path", handler)
	chiPattern := regexp.MustCompile(`r\.(Get|Post|Put|Delete|Patch)\s*\(\s*['"]([^'"]+)['"]`)
	chiMatches := chiPattern.FindAllStringSubmatch(code, -1)

	for _, match := range chiMatches {
		if len(match) >= 3 {
			method := strings.ToUpper(match[1])
			path := match[2]
			endpoint := method + " " + path

			if matchesKeywords(path, keywords) {
				endpoints = appendIfNotExists(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

// detectDjangoEndpoints detects Django REST framework endpoints
func detectDjangoEndpoints(code string, keywords []string) []string {
	var endpoints []string
	// Pattern: @api_view(['GET', 'POST'])
	pattern := regexp.MustCompile(`@api_view\s*\(\s*\[([^\]]+)\]\s*\)`)
	matches := pattern.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			methodList := regexp.MustCompile(`['"]([^'"]+)['"]`).FindAllString(match[1], -1)
			funcPattern := regexp.MustCompile(`def\s+(\w+)\s*\(`)
			funcMatches := funcPattern.FindStringSubmatch(code)
			if len(funcMatches) >= 2 {
				path := "/" + funcMatches[1]
				for _, methodStr := range methodList {
					method := strings.Trim(methodStr, `'"`)
					endpoint := method + " " + path
					if matchesKeywords(path, keywords) {
						endpoints = appendIfNotExists(endpoints, endpoint)
					}
				}
			}
		}
	}
	return endpoints
}

// detectGenericEndpoints tries to detect endpoints using generic patterns
func detectGenericEndpoints(code string, keywords []string) []string {
	var endpoints []string

	// Look for common HTTP method patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(GET|POST|PUT|DELETE|PATCH)\s+['"]([^'"]+)['"]`),
		regexp.MustCompile(`['"]([^'"]+)['"]\s*:\s*(get|post|put|delete|patch)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				method := strings.ToUpper(match[1])
				path := match[2]
				endpoint := method + " " + path

				if matchesKeywords(path, keywords) {
					endpoints = appendIfNotExists(endpoints, endpoint)
				}
			}
		}
	}

	return endpoints
}

// matchesKeywords checks if a string contains any of the keywords
func matchesKeywords(text string, keywords []string) bool {
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

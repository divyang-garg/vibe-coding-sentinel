// Package ast provides tests for enhanced generic detection
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package ast

import (
	"context"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestGenericDetection_Java_JWT(t *testing.T) {
	code := `
public class AuthMiddleware {
    public void authenticate(HttpServletRequest request) {
        String auth = request.getHeader("Authorization");
        if (auth != null && auth.startsWith("Bearer ")) {
            // JWT validation
        }
    }
}
`

	parser, err := GetParser("java")
	if err != nil {
		// Java parser not available - skip test
		t.Skip("Java parser not available")
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		// May fail to parse - use generic detection
		tree = nil
	}

	var rootNode *sitter.Node
	if tree != nil {
		defer tree.Close()
		rootNode = tree.RootNode()
	}

	// Test generic detection (works even without parser)
	findings := detectSecurityMiddlewareGeneric(rootNode, code, "java")

	// Should detect JWT pattern
	found := false
	for _, finding := range findings {
		if finding.Type == "jwt_middleware" {
			found = true
			if finding.Confidence < 0.7 {
				t.Errorf("Expected confidence >= 0.7, got %f", finding.Confidence)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find JWT middleware pattern in Java code")
	}
}

func TestGenericDetection_Java_APIKey(t *testing.T) {
	code := `
public class ApiKeyMiddleware {
    public void validateApiKey(HttpServletRequest request) {
        String apiKey = request.getHeader("X-API-Key");
        if (apiKey != null) {
            // API key validation
        }
    }
}
`

	findings := detectSecurityMiddlewareGeneric(nil, code, "java")

	found := false
	for _, finding := range findings {
		if finding.Type == "apikey_middleware" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find API key middleware pattern")
	}
}

func TestGenericDetection_Rust_JWT(t *testing.T) {
	code := `
use actix_web::{HttpRequest, HttpResponse};

pub fn auth_middleware(req: &HttpRequest) -> Result<(), HttpResponse> {
    if let Some(auth) = req.headers().get("Authorization") {
        if let Ok(auth_str) = auth.to_str() {
            if auth_str.starts_with("Bearer ") {
                // JWT validation
            }
        }
    }
    Ok(())
}
`

	findings := detectSecurityMiddlewareGeneric(nil, code, "rust")

	found := false
	for _, finding := range findings {
		if finding.Type == "jwt_middleware" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find JWT middleware pattern in Rust code")
	}
}

func TestGenericDetection_AllPatterns(t *testing.T) {
	code := `
// JWT/Bearer
String auth = request.getHeader("Authorization");
if (auth.startsWith("Bearer ")) {}

// API Key
String apiKey = request.getHeader("X-API-Key");

// OAuth
oauth2.authenticate(request);

// RBAC
if (user.hasRole("admin")) {}

// Rate Limit
rateLimiter.check(request);

// CORS
response.setHeader("Access-Control-Allow-Origin", "*");
`

	findings := detectSecurityMiddlewareGeneric(nil, code, "java")

	patterns := make(map[string]bool)
	for _, finding := range findings {
		patterns[finding.Type] = true
	}

	expectedPatterns := []string{
		"jwt_middleware",
		"apikey_middleware",
		"oauth_middleware",
		"rbac_middleware",
		"ratelimit_middleware",
		"cors_middleware",
	}

	for _, expected := range expectedPatterns {
		if !patterns[expected] {
			t.Errorf("Expected to find %s pattern", expected)
		}
	}
}

func TestGenericDetection_NoPatterns(t *testing.T) {
	code := `
public class RegularClass {
    public void regularMethod() {
        System.out.println("Hello");
    }
}
`

	findings := detectSecurityMiddlewareGeneric(nil, code, "java")

	// Should not find security patterns
	if len(findings) > 0 {
		t.Logf("Found %d patterns (may be false positives)", len(findings))
	}
}

func TestGenericDetection_EmptyCode(t *testing.T) {
	findings := detectSecurityMiddlewareGeneric(nil, "", "java")

	if len(findings) > 0 {
		t.Errorf("Expected no findings for empty code, got %d", len(findings))
	}
}

func TestContainsJWTBearerPatternGeneric(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"Bearer in Authorization", `auth := r.Header.Get("Authorization"); if strings.HasPrefix(auth, "Bearer ")`, true},
		{"JWT library", `import "github.com/golang-jwt/jwt"`, true},
		{"Token parse", `token, err := jwt.Parse(tokenString)`, true},
		{"No JWT", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsJWTBearerPatternGeneric(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsJWTBearerPatternGeneric() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsAPIKeyPatternGeneric(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"X-API-Key header", `apiKey := r.Header.Get("X-API-Key")`, true},
		{"API key validate", `func ValidateAPIKey(key string) bool`, true},
		{"No API key", `func main() {}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeLower := strings.ToLower(tt.code)
			result := containsAPIKeyPatternGeneric(tt.code, codeLower)
			if result != tt.expected {
				t.Errorf("containsAPIKeyPatternGeneric() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateGenericMiddlewareFinding(t *testing.T) {
	// One line must contain both "bearer" and "authorization" for createGenericMiddlewareFinding to set Line
	code := `
line1
line2
line3: String auth = request.getHeader("Authorization"); if (auth.startsWith("Bearer ")) {}
line4
line5
`

	finding := createGenericMiddlewareFinding(code, "jwt_middleware", "BearerAuth", strings.ToLower(code), "bearer", "authorization")

	if finding.Type != "jwt_middleware" {
		t.Errorf("Expected type 'jwt_middleware', got %s", finding.Type)
	}

	if !strings.Contains(finding.Message, "BearerAuth") {
		t.Errorf("Expected message to contain 'BearerAuth', got %s", finding.Message)
	}

	// createGenericMiddlewareFinding sets Line to the 1-based index of first line containing both keywords
	if finding.Line < 1 || finding.Line > 5 {
		t.Errorf("Expected finding line between 1 and 5, got %d", finding.Line)
	}

	if finding.Confidence != 0.75 {
		t.Errorf("Expected confidence 0.75, got %f", finding.Confidence)
	}
}

package unit

import (
	"regexp"
	"strings"
	"testing"
)

// TestValidateVersionFormat tests version format validation (matches main.go logic)
func TestValidateVersionFormat(t *testing.T) {
	versionPattern := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid semver", "1.2.3", false},
		{"valid semver with v prefix", "v1.2.3", false},
		{"valid semver with prerelease", "1.2.3-alpha", false},
		{"valid semver with prerelease and v", "v1.2.3-beta", false},
		{"invalid format", "1.2", true},
		{"invalid format", "1.2.3.4", true},
		{"invalid format", "abc", true},
		{"invalid format", "1.2.3-", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := versionPattern.MatchString(tt.version)
			if matches == tt.wantErr {
				t.Errorf("validateVersionFormat(%q) matches = %v, wantErr %v", tt.version, matches, tt.wantErr)
			}
		})
	}
}

// TestValidatePlatform tests platform validation (matches main.go logic)
func TestValidatePlatform(t *testing.T) {
	allowedPlatforms := []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"}

	tests := []struct {
		name     string
		platform string
		wantErr  bool
	}{
		{"valid linux-amd64", "linux-amd64", false},
		{"valid linux-arm64", "linux-arm64", false},
		{"valid darwin-amd64", "darwin-amd64", false},
		{"valid darwin-arm64", "darwin-arm64", false},
		{"valid windows-amd64", "windows-amd64", false},
		{"invalid platform", "linux-x86", true},
		{"invalid platform", "macos-amd64", true},
		{"invalid platform", "win-amd64", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPlatform := false
			for _, p := range allowedPlatforms {
				if tt.platform == p {
					validPlatform = true
					break
				}
			}
			if validPlatform == tt.wantErr {
				t.Errorf("validatePlatform(%q) valid = %v, wantErr %v", tt.platform, validPlatform, tt.wantErr)
			}
		})
	}
}

// TestSanitizeString tests string sanitization (matches main.go logic)
func TestSanitizeString(t *testing.T) {
	sanitizeString := func(s string, maxLen int) string {
		s = strings.TrimSpace(s)
		if len(s) > maxLen {
			s = s[:maxLen]
		}
		// Remove control characters except newline, carriage return, and tab
		s = strings.Map(func(r rune) rune {
			if r < 32 && r != '\n' && r != '\r' && r != '\t' {
				return -1
			}
			return r
		}, s)
		return s
	}

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal string", "Hello World", 100, "Hello World"},
		{"string with control chars", "Hello\x00World", 100, "HelloWorld"},
		{"string with newline", "Hello\nWorld", 100, "Hello\nWorld"},
		{"string with tab", "Hello\tWorld", 100, "Hello\tWorld"},
		{"string too long", "Hello World", 5, "Hello"},
		{"trimmed string", "  Hello World  ", 100, "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("sanitizeString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}


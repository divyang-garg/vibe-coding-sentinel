// Package ast provides insecure crypto detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectInsecureCrypto finds insecure cryptographic usage
func detectInsecureCrypto(root *sitter.Node, code string, language string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Language-specific detection
	switch language {
	case "go":
		vulnerabilities = append(vulnerabilities, detectInsecureCryptoGo(root, code)...)
	case "javascript", "typescript":
		vulnerabilities = append(vulnerabilities, detectInsecureCryptoJS(root, code)...)
	case "python":
		vulnerabilities = append(vulnerabilities, detectInsecureCryptoPython(root, code)...)
	}

	return vulnerabilities
}

// detectInsecureCryptoGo detects insecure crypto in Go code
func detectInsecureCryptoGo(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	traverseAST(root, func(node *sitter.Node) bool {
		codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
		
		// Check for weak hash algorithms
		if hasWeakHash(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "insecure_crypto",
				Severity:    "high",
				Line:        line,
				Column:      col,
				Message:     "Insecure cryptographic hash function detected (MD5 or SHA1)",
				Code:        codeSnippet,
				Description: "MD5 and SHA1 are cryptographically broken and should not be used",
				Remediation: "Use SHA-256 or SHA-512 from crypto/sha256 or crypto/sha512",
				Confidence:  0.95,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		// Check for hardcoded keys/secrets
		if hasHardcodedSecret(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "hardcoded_secret",
				Severity:    "critical",
				Line:        line,
				Column:      col,
				Message:     "Hardcoded secret or cryptographic key detected",
				Code:        codeSnippet,
				Description: "Secrets and keys should never be hardcoded in source code",
				Remediation: "Use environment variables or secure secret management",
				Confidence:  0.9,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		return true
	})

	return vulnerabilities
}

// detectInsecureCryptoJS detects insecure crypto in JavaScript/TypeScript code
func detectInsecureCryptoJS(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	traverseAST(root, func(node *sitter.Node) bool {
		codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
		
		// Check for weak hash algorithms
		if hasWeakHash(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "insecure_crypto",
				Severity:    "high",
				Line:        line,
				Column:      col,
				Message:     "Insecure cryptographic hash function detected (MD5 or SHA1)",
				Code:        codeSnippet,
				Description: "MD5 and SHA1 are cryptographically broken and should not be used",
				Remediation: "Use crypto.subtle.digest with SHA-256 or SHA-512",
				Confidence:  0.95,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		// Check for hardcoded secrets
		if hasHardcodedSecret(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "hardcoded_secret",
				Severity:    "critical",
				Line:        line,
				Column:      col,
				Message:     "Hardcoded secret or API key detected",
				Code:        codeSnippet,
				Description: "Secrets and keys should never be hardcoded in source code",
				Remediation: "Use environment variables or secure secret management",
				Confidence:  0.9,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		return true
	})

	return vulnerabilities
}

// detectInsecureCryptoPython detects insecure crypto in Python code
func detectInsecureCryptoPython(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	traverseAST(root, func(node *sitter.Node) bool {
		codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
		
		// Check for weak hash algorithms
		if hasWeakHash(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "insecure_crypto",
				Severity:    "high",
				Line:        line,
				Column:      col,
				Message:     "Insecure cryptographic hash function detected (MD5 or SHA1)",
				Code:        codeSnippet,
				Description: "MD5 and SHA1 are cryptographically broken and should not be used",
				Remediation: "Use hashlib.sha256() or hashlib.sha512()",
				Confidence:  0.95,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		// Check for hardcoded secrets
		if hasHardcodedSecret(codeSnippet) {
			line, col := getLineColumn(code, int(node.StartByte()))
			vuln := SecurityVulnerability{
				Type:        "hardcoded_secret",
				Severity:    "critical",
				Line:        line,
				Column:      col,
				Message:     "Hardcoded secret or API key detected",
				Code:        codeSnippet,
				Description: "Secrets and keys should never be hardcoded in source code",
				Remediation: "Use environment variables or secure secret management",
				Confidence:  0.9,
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}

		return true
	})

	return vulnerabilities
}

// Helper functions
func hasWeakHash(code string) bool {
	weakHashes := []string{
		"md5", "MD5", "sha1", "SHA1", "sha-1", "SHA-1",
		"crypto/md5", "crypto/sha1",
		"createHash('md5')", "createHash('sha1')",
		"hashlib.md5", "hashlib.sha1",
	}
	codeLower := strings.ToLower(code)
	for _, hash := range weakHashes {
		if strings.Contains(codeLower, strings.ToLower(hash)) {
			return true
		}
	}
	return false
}

func hasHardcodedSecret(code string) bool {
	secretPatterns := []string{
		"password", "secret", "api_key", "apikey", "token",
		"private_key", "privatekey", "access_key", "accesskey",
	}
	codeLower := strings.ToLower(code)
	
	// Look for assignment patterns with potential secrets
	for _, pattern := range secretPatterns {
		if strings.Contains(codeLower, pattern) {
			// Check if it's assigned to a string literal (hardcoded)
			if strings.Contains(code, "=") && 
			   (strings.Contains(code, "\"") || strings.Contains(code, "'")) {
				// Simple heuristic: if pattern appears near assignment
				return true
			}
		}
	}
	return false
}

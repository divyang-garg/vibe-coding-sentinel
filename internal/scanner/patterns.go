// Package scanner provides security pattern definitions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package scanner

import "regexp"

// Pattern represents a security pattern to scan for
type Pattern struct {
	Name     string
	Regex    *regexp.Regexp
	Severity Severity
	Message  string
}

// GetSecurityPatterns returns all security patterns to scan for
func GetSecurityPatterns() []Pattern {
	return []Pattern{
		{
			Name:     "secrets",
			Regex:    regexp.MustCompile(`(?i)(ey|api[_-]?key|secret|password|token)\s*[=:]\s*["']?[a-zA-Z0-9]{20,}`),
			Severity: SeverityCritical,
			Message:  "Potential secret or API key detected",
		},
		{
			Name:     "debug",
			Regex:    regexp.MustCompile(`console\.(log|debug|info|warn|error)`),
			Severity: SeverityWarning,
			Message:  "console.log statement detected",
		},
		{
			Name:     "sql_safety",
			Regex:    regexp.MustCompile(`(?i)NOLOCK`),
			Severity: SeverityCritical,
			Message:  "MSSQL NOLOCK detected (unsafe)",
		},
		{
			Name:     "sql_injection",
			Regex:    regexp.MustCompile(`\$_[A-Z]+\[`),
			Severity: SeverityCritical,
			Message:  "Potential SQL injection vulnerability - direct use of user input",
		},
		{
			Name:     "eval_usage",
			Regex:    regexp.MustCompile(`\beval\b`),
			Severity: SeverityCritical,
			Message:  "Code injection vulnerability - eval() usage detected",
		},
		{
			Name:     "nosql_injection",
			Regex:    regexp.MustCompile(`(?i)\$where`),
			Severity: SeverityCritical,
			Message:  "NoSQL injection vulnerability - $where usage detected",
		},
		{
			Name:     "unquoted_variables",
			Regex:    regexp.MustCompile(`\$\{[a-zA-Z_][a-zA-Z0-9_]*\}`),
			Severity: SeverityWarning,
			Message:  "Unquoted variable expansion detected",
		},
		{
			Name:     "hardcoded_ip",
			Regex:    regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`),
			Severity: SeverityWarning,
			Message:  "Hardcoded IP address detected",
		},
		{
			Name:     "weak_crypto",
			Regex:    regexp.MustCompile(`(?i)(md5|sha1)\s*\(`),
			Severity: SeverityHigh,
			Message:  "Weak cryptographic hash function detected",
		},
		{
			Name:     "insecure_random",
			Regex:    regexp.MustCompile(`(?i)(math\.random|random\(\))`),
			Severity: SeverityMedium,
			Message:  "Insecure random number generation detected",
		},
	}
}

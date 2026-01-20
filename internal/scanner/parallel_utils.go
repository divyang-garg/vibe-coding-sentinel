// Parallel Scanner - Utility Functions
// String manipulation and conversion utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package scanner

// itoa converts int to string without fmt dependency
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// Helper functions to avoid importing strings package in parallel.go
func splitLines(s string) []string {
	var lines []string
	var current []rune
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, string(current))
			current = nil
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// Package ast provides command injection detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectCommandInjection finds command injection vulnerabilities.
// Uses registry when a detector is registered; otherwise falls back to switch.
func detectCommandInjection(root *sitter.Node, code string, language string) []SecurityVulnerability {
	if d := GetLanguageDetector(language); d != nil {
		return d.DetectCommandInjection(root, code)
	}
	vulnerabilities := []SecurityVulnerability{}
	switch language {
	case "go":
		vulnerabilities = append(vulnerabilities, detectCommandInjectionGo(root, code)...)
	case "javascript", "typescript":
		vulnerabilities = append(vulnerabilities, detectCommandInjectionJS(root, code)...)
	case "python":
		vulnerabilities = append(vulnerabilities, detectCommandInjectionPython(root, code)...)
	}
	return vulnerabilities
}

// detectCommandInjectionGo detects command injection in Go code
func detectCommandInjectionGo(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for exec.Command, exec.CommandContext
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			if isCommandExecutionFunction(funcName) {
				// Check if user input is used in command
				if hasUserInputInCommand(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "command_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential command injection in %s: user input in shell command", funcName),
						Code:        safeSlice(code, node.StartByte(), node.EndByte()),
						Description: "User input used directly in shell command without validation",
						Remediation: "Use exec.Command with separate arguments, validate/sanitize input",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// detectCommandInjectionJS detects command injection in JavaScript/TypeScript code
func detectCommandInjectionJS(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for child_process.exec, child_process.spawn, execSync
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			if isCommandExecutionFunction(funcName) || isChildProcessCall(codeSnippet) {
				// Check if user input is used
				if hasUserInputInCommand(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "command_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     "Potential command injection: user input in shell command",
						Code:        codeSnippet,
						Description: "User input used directly in shell command without validation",
						Remediation: "Use child_process.spawn with array arguments, validate/sanitize input",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// detectCommandInjectionPython detects command injection in Python code
func detectCommandInjectionPython(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Enhanced pattern-based detection for Python command injection
	lines := strings.Split(code, "\n")

	// Pattern 1: os.system with user input
	for lineNum, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "os.system") || strings.Contains(lineLower, "system(") {
			// Check for user input patterns
			if hasUserInputInCommandString(line) {
				vuln := SecurityVulnerability{
					Type:        "command_injection",
					Severity:    "critical",
					Line:        lineNum + 1,
					Column:      1,
					Message:     "Potential command injection: os.system with user input",
					Code:        strings.TrimSpace(line),
					Description: "os.system called with potentially unsafe user input",
					Remediation: "Use subprocess.run with list arguments, validate/sanitize input",
					Confidence:  0.9,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
	}

	// Pattern 2: subprocess.call/Popen with shell=True
	for lineNum, line := range lines {
		lineLower := strings.ToLower(line)
		if (strings.Contains(lineLower, "subprocess.call") || strings.Contains(lineLower, "subprocess.popen") ||
			strings.Contains(lineLower, "subprocess.run")) && strings.Contains(lineLower, "shell=true") {
			// Check for user input
			if hasUserInputInCommandString(line) {
				vuln := SecurityVulnerability{
					Type:        "command_injection",
					Severity:    "critical",
					Line:        lineNum + 1,
					Column:      1,
					Message:     "Potential command injection: subprocess with shell=True and user input",
					Code:        strings.TrimSpace(line),
					Description: "subprocess called with shell=True and potentially unsafe user input",
					Remediation: "Use subprocess.run with list arguments (shell=False), validate/sanitize input",
					Confidence:  0.9,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
	}

	// Pattern 3: subprocess with string command (not list)
	for lineNum, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "subprocess.") {
			// Check if command is a string (not a list) and contains user input
			if strings.Contains(line, "[") == false && // Not a list
				strings.Contains(line, "\"") || strings.Contains(line, "'") { // Is a string
				if hasUserInputInCommandString(line) {
					vuln := SecurityVulnerability{
						Type:        "command_injection",
						Severity:    "critical",
						Line:        lineNum + 1,
						Column:      1,
						Message:     "Potential command injection: subprocess with string command",
						Code:        strings.TrimSpace(line),
						Description: "subprocess called with string command (should use list)",
						Remediation: "Use subprocess.run with list arguments, validate/sanitize input",
						Confidence:  0.85,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
	}

	// AST-based detection for complex patterns
	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for os.system, subprocess.call, subprocess.Popen
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			if isCommandExecutionFunction(funcName) || isSubprocessCall(codeSnippet) {
				// Check if user input is used
				if hasUserInputInCommand(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "command_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     "Potential command injection: user input in shell command",
						Code:        codeSnippet,
						Description: "User input used directly in shell command without validation",
						Remediation: "Use subprocess.run with list arguments, validate/sanitize input",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// Helper functions
func isCommandExecutionFunction(name string) bool {
	cmdFuncs := []string{
		"Command", "CommandContext", "exec", "system", "popen",
		"execSync", "execFile", "spawn", "call", "run",
	}
	nameLower := strings.ToLower(name)
	for _, funcName := range cmdFuncs {
		if strings.Contains(nameLower, strings.ToLower(funcName)) {
			return true
		}
	}
	return false
}

func isChildProcessCall(code string) bool {
	return strings.Contains(code, "child_process") ||
		strings.Contains(code, "require('child_process')")
}

func isSubprocessCall(code string) bool {
	return strings.Contains(code, "subprocess.") ||
		strings.Contains(code, "os.system")
}

func hasUserInputInCommand(node *sitter.Node, code string) bool {
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	return hasUserInputInCommandString(codeSnippet)
}

func hasUserInputInCommandString(code string) bool {
	userInputPatterns := []string{
		"req.", "request.", "params.", "query.", "body.",
		"form.", "input.", "user.", "data.", "argv",
		"sys.argv", "process.argv", "filename", "file_name",
		"user_input", "userinput", "cmd", "command",
	}
	codeLower := strings.ToLower(code)

	// Check for function parameters (common in command injection)
	if strings.Contains(code, "(") && strings.Contains(code, ")") {
		// Extract parameter name if it's a simple variable
		// Look for patterns like func(param) where param could be user input
		if !strings.Contains(codeLower, "validate") &&
			!strings.Contains(codeLower, "sanitize") &&
			!strings.Contains(codeLower, "whitelist") &&
			!strings.Contains(codeLower, "shlex.quote") {
			// Check if it's not a literal string
			if !strings.Contains(code, "\"") && !strings.Contains(code, "'") {
				return true
			}
		}
	}

	for _, pattern := range userInputPatterns {
		if strings.Contains(codeLower, strings.ToLower(pattern)) {
			// Check if it's validated/sanitized
			if !strings.Contains(codeLower, "validate") &&
				!strings.Contains(codeLower, "sanitize") &&
				!strings.Contains(codeLower, "whitelist") &&
				!strings.Contains(codeLower, "shlex.quote") {
				return true
			}
		}
	}
	return false
}

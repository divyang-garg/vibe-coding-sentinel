// Mutation Engine - Generation Functions
// Generates mutants for source code using various mutation operators
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"fmt"
	"regexp"
	"strings"
)

// generateMutants generates mutants for source code (file-level, limited per function)
func generateMutants(sourceCode string, language string, maxMutantsPerFunction int) []Mutant {
	var mutants []Mutant
	mutantID := 0

	lines := strings.Split(sourceCode, "\n")

	// Limit total mutants to prevent excessive execution time
	maxTotalMutants := 50
	if maxMutantsPerFunction > 0 {
		maxTotalMutants = maxMutantsPerFunction * 10 // Assume ~10 functions per file
	}

	// Track mutations per function to limit them
	mutationsPerFunction := make(map[string]int)
	currentFunction := "global"

	for lineNum, line := range lines {
		// Detect function boundaries (simplified)
		if isFunctionStart(line, language) {
			// Extract function name
			currentFunction = extractFunctionNameForMutation(line, language)
			mutationsPerFunction[currentFunction] = 0
		}

		// Skip if we've hit the limit for this function
		if mutationsPerFunction[currentFunction] >= maxMutantsPerFunction {
			continue
		}

		// Skip if we've hit total limit
		if len(mutants) >= maxTotalMutants {
			break
		}

		// Generate mutants for this line
		lineMutants := generateLineMutants(line, lineNum+1, language)

		for _, mutant := range lineMutants {
			if mutationsPerFunction[currentFunction] >= maxMutantsPerFunction {
				break
			}
			mutant.ID = fmt.Sprintf("mutant_%d", mutantID)
			mutantID++
			mutants = append(mutants, mutant)
			mutationsPerFunction[currentFunction]++
		}
	}

	return mutants
}

// isFunctionStart checks if a line starts a function definition
func isFunctionStart(line string, language string) bool {
	lineTrimmed := strings.TrimSpace(line)
	switch strings.ToLower(language) {
	case "go", "golang":
		return strings.HasPrefix(lineTrimmed, "func ")
	case "javascript", "js", "typescript", "ts":
		return strings.Contains(lineTrimmed, "function ") ||
			strings.Contains(lineTrimmed, "=>") ||
			regexp.MustCompile(`^\s*(const|let|var)\s+\w+\s*=\s*\(`).MatchString(lineTrimmed)
	case "python", "py":
		return strings.HasPrefix(lineTrimmed, "def ")
	default:
		return false
	}
}

// extractFunctionNameForMutation extracts function name from function definition line (for mutation testing)
func extractFunctionNameForMutation(line string, language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		// func FunctionName(...)
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "func" && i+1 < len(parts) {
				funcName := parts[i+1]
				// Remove receiver if present
				if strings.Contains(funcName, "(") {
					continue
				}
				return funcName
			}
		}
	case "javascript", "js", "typescript", "ts":
		// function name(...) or const name = (...)
		if match := regexp.MustCompile(`function\s+(\w+)`).FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
		if match := regexp.MustCompile(`(const|let|var)\s+(\w+)\s*=`).FindStringSubmatch(line); len(match) > 2 {
			return match[2]
		}
	case "python", "py":
		// def name(...)
		if match := regexp.MustCompile(`def\s+(\w+)`).FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
	}
	return "unknown"
}

// generateLineMutants generates mutants for a single line of code
func generateLineMutants(line string, lineNum int, language string) []Mutant {
	var mutants []Mutant

	// Arithmetic operator mutations: + → -, * → /, etc.
	arithmeticOps := map[string][]string{
		"+": {"-", "*"},
		"-": {"+", "*"},
		"*": {"/", "+"},
		"/": {"*", "-"},
	}

	for op, replacements := range arithmeticOps {
		if strings.Contains(line, op) {
			for _, replacement := range replacements {
				mutated := strings.Replace(line, op, replacement, 1)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("arithmetic_%s_to_%s", op, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Comparison operator mutations: == → !=, < → <=, etc.
	comparisonOps := map[string][]string{
		"==": {"!=", "<", ">"},
		"!=": {"==", "<", ">"},
		"<":  {"<=", ">", "=="},
		">":  {">=", "<", "=="},
		"<=": {"<", ">", "=="},
		">=": {">", "<", "=="},
	}

	for op, replacements := range comparisonOps {
		if strings.Contains(line, op) {
			for _, replacement := range replacements {
				mutated := strings.Replace(line, op, replacement, 1)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("comparison_%s_to_%s", op, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Boolean operator mutations: && → ||, ! → remove
	if strings.Contains(line, "&&") {
		mutated := strings.Replace(line, "&&", "||", 1)
		mutants = append(mutants, Mutant{
			Original: line,
			Mutated:  mutated,
			Operator: "boolean_and_to_or",
			Line:     lineNum,
		})
	}
	if strings.Contains(line, "||") {
		mutated := strings.Replace(line, "||", "&&", 1)
		mutants = append(mutants, Mutant{
			Original: line,
			Mutated:  mutated,
			Operator: "boolean_or_to_and",
			Line:     lineNum,
		})
	}

	// Constant mutations: 1 → 0, true → false, etc.
	constantMutations := map[string]string{
		"1":     "0",
		"0":     "1",
		"true":  "false",
		"false": "true",
		"nil":   "not_nil", // Special case - would need proper replacement
		"null":  "not_null",
	}

	for constant, replacement := range constantMutations {
		if strings.Contains(line, constant) {
			// Only replace whole-word matches
			re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(constant)))
			if re.MatchString(line) {
				mutated := re.ReplaceAllString(line, replacement)
				if mutated != line {
					mutants = append(mutants, Mutant{
						Original: line,
						Mutated:  mutated,
						Operator: fmt.Sprintf("constant_%s_to_%s", constant, replacement),
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Return value mutations: return x → return nil (for languages that support it)
	if strings.Contains(line, "return") && !strings.Contains(line, "return nil") && !strings.Contains(line, "return null") {
		// Extract return value
		if match := regexp.MustCompile(`return\s+(\S+)`).FindStringSubmatch(line); len(match) > 1 {
			returnValue := match[1]
			var nilValue string
			switch strings.ToLower(language) {
			case "go", "golang":
				nilValue = "nil"
			case "javascript", "js", "typescript", "ts":
				nilValue = "null"
			case "python", "py":
				nilValue = "None"
			default:
				nilValue = "nil"
			}
			mutated := strings.Replace(line, returnValue, nilValue, 1)
			mutants = append(mutants, Mutant{
				Original: line,
				Mutated:  mutated,
				Operator: fmt.Sprintf("return_%s_to_%s", returnValue, nilValue),
				Line:     lineNum,
			})
		}
	}

	return mutants
}

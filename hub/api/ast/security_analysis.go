// Package ast provides security-focused AST analysis
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package ast

import (
	"context"
	"fmt"
	"time"
)

// AnalyzeSecurity performs security-focused AST analysis
func AnalyzeSecurity(ctx context.Context, code, language, severity string) ([]SecurityVulnerability, []ASTFinding, AnalysisStats, error) {
	// Get parser for language
	parser, err := GetParser(language)
	if err != nil {
		return nil, nil, AnalysisStats{}, fmt.Errorf("parser error: %w", err)
	}

	// Parse code into AST
	parseStart := time.Now()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, nil, AnalysisStats{}, fmt.Errorf("parse error: %w", err)
	}
	parseTime := time.Since(parseStart).Milliseconds()

	if tree == nil {
		return nil, nil, AnalysisStats{}, fmt.Errorf("failed to parse code")
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, nil, AnalysisStats{}, fmt.Errorf("failed to get root node")
	}

	// Perform security analyses
	analysisStart := time.Now()
	vulnerabilities := []SecurityVulnerability{}
	findings := []ASTFinding{}

	// SQL Injection detection
	sqlVulns := detectSQLInjection(rootNode, code, language)
	vulnerabilities = append(vulnerabilities, sqlVulns...)

	// XSS detection
	xssVulns := detectXSS(rootNode, code, language)
	vulnerabilities = append(vulnerabilities, xssVulns...)

	// Command injection detection
	cmdVulns := detectCommandInjection(rootNode, code, language)
	vulnerabilities = append(vulnerabilities, cmdVulns...)

	// Insecure crypto detection
	cryptoVulns := detectInsecureCrypto(rootNode, code, language)
	vulnerabilities = append(vulnerabilities, cryptoVulns...)

	// Secrets detection
	secretsVulns := detectSecrets(rootNode, code, language)
	vulnerabilities = append(vulnerabilities, secretsVulns...)

	// Filter by severity if specified
	if severity != "" && severity != "all" {
		filtered := []SecurityVulnerability{}
		for _, vuln := range vulnerabilities {
			if vuln.Severity == severity {
				filtered = append(filtered, vuln)
			}
		}
		vulnerabilities = filtered
	}

	// Convert vulnerabilities to findings
	for _, vuln := range vulnerabilities {
		finding := ASTFinding{
			Type:        vuln.Type,
			Severity:    vuln.Severity,
			Line:        vuln.Line,
			Column:      vuln.Column,
			Message:     vuln.Message,
			Code:        vuln.Code,
			Suggestion:  vuln.Remediation,
			Confidence:  vuln.Confidence,
			AutoFixSafe: false, // Never auto-fix security issues
			FixType:     "manual",
			Reasoning:   vuln.Description,
		}
		findings = append(findings, finding)
	}

	analysisTime := time.Since(analysisStart).Milliseconds()

	stats := AnalysisStats{
		ParseTime:    parseTime,
		AnalysisTime: analysisTime,
		NodesVisited: countNodes(rootNode),
	}

	return vulnerabilities, findings, stats, nil
}

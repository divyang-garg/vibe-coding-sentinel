// Package security provides security analysis types and utilities
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package security

// SecurityFinding represents a single security vulnerability finding
type SecurityFinding struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Severity    string `json:"severity"`
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Issue       string `json:"issue"`
	Remediation string `json:"remediation"`
	AutoFixable bool   `json:"autoFixable"`
}

// SecurityAnalysisResponse represents the response from security analysis
type SecurityAnalysisResponse struct {
	Score    int               `json:"score"` // 0-100
	Grade    string            `json:"grade"` // A, B, C, D, F
	Findings []SecurityFinding `json:"findings"`
	Summary  struct {
		TotalRules int `json:"totalRules"`
		Passed     int `json:"passed"`
		Failed     int `json:"failed"`
		Critical   int `json:"critical"`
		High       int `json:"high"`
		Medium     int `json:"medium"`
		Low        int `json:"low"`
	} `json:"summary"`
	Metrics *DetectionMetrics `json:"metrics,omitempty"` // Optional: only for validation runs
}

// DetectionMetrics represents security detection performance metrics
type DetectionMetrics struct {
	TruePositives  int     `json:"truePositives"`
	FalsePositives int     `json:"falsePositives"`
	TrueNegatives  int     `json:"trueNegatives"`
	FalseNegatives int     `json:"falseNegatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1Score        float64 `json:"f1Score"`
}

// SecurityAnalysisRequest represents a request for security analysis
type SecurityAnalysisRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	FileName string `json:"fileName,omitempty"`
}

// Security Rules Definitions (SEC-001 to SEC-008)
var SecurityRules = map[string]struct {
	Name     string
	Severity string
	Type     string
}{
	"SEC-001": {"Resource Ownership", "critical", "authorization"},
	"SEC-002": {"SQL Injection", "critical", "injection"},
	"SEC-003": {"Auth Middleware", "critical", "authentication"},
	"SEC-004": {"Rate Limiting", "high", "transport"},
	"SEC-005": {"Password Hashing", "critical", "cryptography"},
	"SEC-006": {"Input Validation", "high", "validation"},
	"SEC-007": {"Secure Headers", "medium", "transport"},
	"SEC-008": {"CORS Config", "high", "transport"},
}

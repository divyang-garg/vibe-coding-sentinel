// Package utils - Flow verification (re-exports for backward compatibility)
// This file maintains backward compatibility by re-exporting functions from refactored modules.
// All implementation has been moved to:
//   - flow_verifier_core.go: Main orchestration and types
//   - flow_breakpoints.go: Breakpoint detection
//   - flow_matchers.go: Component/endpoint/function matching
//   - flow_validators.go: Validation checks
package utils

// Re-export types and functions for backward compatibility
// All types and functions are defined in flow_verifier_core.go, flow_breakpoints.go,
// flow_matchers.go, and flow_validators.go

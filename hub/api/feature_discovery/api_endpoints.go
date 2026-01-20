// Package feature_discovery - API endpoints (re-exports for backward compatibility)
// This file maintains backward compatibility by re-exporting functions from refactored modules.
// All implementation has been moved to:
//   - api_endpoints_core.go: Main discovery logic
//   - api_express.go: Express.js endpoint discovery
//   - api_fastapi.go: FastAPI endpoint discovery
//   - api_django.go: Django endpoint discovery
//   - api_go.go: Go (Gin/Chi) endpoint discovery
//   - api_utils.go: Shared utilities
//
// All types and functions are defined in the above files and are accessible
// from this package since they are all in package feature_discovery.

package feature_discovery

// Re-export types and functions for backward compatibility
// All types and functions are defined in api_endpoints_core.go, api_express.go,
// api_fastapi.go, api_django.go, api_go.go, and api_utils.go

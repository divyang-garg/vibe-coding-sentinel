// Package feature_discovery - UI components (re-exports for backward compatibility)
// This file maintains backward compatibility by re-exporting functions from refactored modules.
// All implementation has been moved to:
//   - ui_components_core.go: Main discovery logic
//   - ui_react.go: React-specific extraction
//   - ui_vue.go: Vue-specific extraction
//   - ui_angular.go: Angular-specific extraction
//   - ui_styling.go: Styling framework detection
//   - ui_hierarchy.go: Component hierarchy building
//
// All types and functions are defined in the above files and are accessible
// from this package since they are all in package feature_discovery.

package feature_discovery

// Re-export types and functions for backward compatibility
// All types and functions are defined in ui_components_core.go, ui_react.go,
// ui_vue.go, ui_angular.go, ui_styling.go, and ui_hierarchy.go

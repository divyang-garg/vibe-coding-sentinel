// Package feature_discovery provides UI component analysis types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package feature_discovery

// UILayerComponents represents discovered UI components
type UILayerComponents struct {
	Components []ComponentInfo `json:"components"`
	Framework  string          `json:"framework"`
	Styling    []string        `json:"styling,omitempty"`   // CSS frameworks used
	Hierarchy  ComponentTree   `json:"hierarchy,omitempty"` // Component hierarchy
}

// ComponentInfo contains information about a UI component
type ComponentInfo struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Type         string            `json:"type"`      // "component", "form", "page", "layout"
	Framework    string            `json:"framework"` // "react", "vue", "angular"
	Props        []PropInfo        `json:"props,omitempty"`
	State        []StateInfo       `json:"state,omitempty"`
	Methods      []string          `json:"methods,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PropInfo contains information about component props
type PropInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
}

// StateInfo contains information about component state
type StateInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// ComponentTree represents the component hierarchy
type ComponentTree struct {
	Root    *ComponentNode      `json:"root,omitempty"`
	Imports map[string][]string `json:"imports,omitempty"` // file -> imported components
	Exports map[string][]string `json:"exports,omitempty"` // file -> exported components
}

// ComponentNode represents a node in the component tree
type ComponentNode struct {
	Component ComponentInfo    `json:"component"`
	Children  []*ComponentNode `json:"children,omitempty"`
	Parent    *ComponentNode   `json:"-"`
}

// StylingFramework represents a detected styling framework
type StylingFramework struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type"`  // "css", "scss", "less", "tailwind", "styled-components"
	Usage   string `json:"usage"` // "utility", "component", "global"
}

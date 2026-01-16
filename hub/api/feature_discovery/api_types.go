// Package feature_discovery provides API endpoint analysis types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package feature_discovery

// APILayerEndpoints represents discovered API endpoints
type APILayerEndpoints struct {
	Endpoints  []EndpointInfo   `json:"endpoints"`
	Framework  string           `json:"framework"`
	Middleware []MiddlewareInfo `json:"middleware,omitempty"` // Framework middleware
	Auth       []AuthInfo       `json:"auth,omitempty"`       // Authentication methods
}

// EndpointInfo contains information about an API endpoint
type EndpointInfo struct {
	Method     string            `json:"method"`               // GET, POST, PUT, DELETE, PATCH
	Path       string            `json:"path"`                 // Route path
	Handler    string            `json:"handler,omitempty"`    // Handler function name
	File       string            `json:"file"`                 // Source file path
	Parameters []ParameterInfo   `json:"parameters,omitempty"` // Path/query parameters
	Responses  []ResponseInfo    `json:"responses,omitempty"`  // Response schemas
	Middleware []string          `json:"middleware,omitempty"` // Endpoint-specific middleware
	Auth       []string          `json:"auth,omitempty"`       // Authentication requirements
	Metadata   map[string]string `json:"metadata,omitempty"`   // Additional metadata
}

// ParameterInfo contains information about endpoint parameters
type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`      // path, query, body, header
	DataType    string `json:"data_type,omitempty"` // string, int, bool, object
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

// ResponseInfo contains information about endpoint responses
type ResponseInfo struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema,omitempty"` // Response schema reference
}

// MiddlewareInfo contains information about middleware
type MiddlewareInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // global, route, error
	Description string `json:"description,omitempty"`
	File        string `json:"file,omitempty"`
}

// AuthInfo contains information about authentication methods
type AuthInfo struct {
	Method      string `json:"method"` // jwt, basic, oauth, api_key
	Description string `json:"description,omitempty"`
	File        string `json:"file,omitempty"`
}

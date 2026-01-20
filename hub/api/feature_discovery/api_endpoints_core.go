// Package feature_discovery - API endpoint discovery core
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"context"
)

// discoverAPIEndpoints discovers API endpoints in the codebase
// Supports Express, FastAPI, Django, Gin, Chi frameworks
func discoverAPIEndpoints(ctx context.Context, codebasePath string, featureName string, framework string) (*APILayerEndpoints, error) {
	endpoints := []EndpointInfo{}

	switch framework {
	case "express":
		endpoints = discoverExpressEndpoints(codebasePath, featureName)
	case "fastapi":
		endpoints = discoverFastAPIEndpointsInCodebase(codebasePath, featureName)
	case "django":
		endpoints = discoverDjangoEndpoints(codebasePath, featureName)
	case "gin", "chi":
		endpoints = discoverGoEndpoints(codebasePath, featureName, framework)
	default:
		// Try to auto-detect framework
		endpoints = autoDetectAPIEndpoints(codebasePath, featureName)
	}

	// Discover middleware and authentication
	middleware := discoverMiddleware(codebasePath, framework)
	auth := discoverAuthentication(codebasePath, framework)

	return &APILayerEndpoints{
		Endpoints:  endpoints,
		Framework:  framework,
		Middleware: middleware,
		Auth:       auth,
	}, nil
}

// autoDetectAPIEndpoints attempts to auto-detect API endpoints
func autoDetectAPIEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Try Express patterns
	expressEndpoints := discoverExpressEndpoints(codebasePath, featureName)
	endpoints = append(endpoints, expressEndpoints...)

	// Try FastAPI patterns
	fastapiEndpoints := discoverFastAPIEndpointsInCodebase(codebasePath, featureName)
	endpoints = append(endpoints, fastapiEndpoints...)

	// Try Django patterns
	djangoEndpoints := discoverDjangoEndpoints(codebasePath, featureName)
	endpoints = append(endpoints, djangoEndpoints...)

	// Try Go patterns (both Gin and Chi)
	goEndpointsGin := discoverGoEndpoints(codebasePath, featureName, "gin")
	goEndpointsChi := discoverGoEndpoints(codebasePath, featureName, "chi")
	endpoints = append(endpoints, goEndpointsGin...)
	endpoints = append(endpoints, goEndpointsChi...)

	return endpoints
}

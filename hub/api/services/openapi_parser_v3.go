// Package services provides OpenAPI 3.x parsing support
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// extractOpenAPI3Endpoints extracts endpoints from OpenAPI 3.x specification
func extractOpenAPI3Endpoints(ctx context.Context, document libopenapi.Document) ([]ContractEndpoint, error) {
	model, err := document.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build OpenAPI 3.x model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("OpenAPI 3.x model is nil")
	}

	var endpoints []ContractEndpoint
	paths := model.Model.Paths

	if paths == nil {
		return endpoints, nil
	}

	for pathPair := paths.PathItems.First(); pathPair != nil; pathPair = pathPair.Next() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		path := pathPair.Key()
		pathItem := pathPair.Value()

		if pathItem == nil {
			continue
		}

		// Extract operations (GET, POST, PUT, DELETE, PATCH, etc.)
		operations := []struct {
			method string
			op     *v3.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
			{"HEAD", pathItem.Head},
			{"OPTIONS", pathItem.Options},
			{"TRACE", pathItem.Trace},
		}

		for _, opData := range operations {
			if opData.op == nil {
				continue
			}

			endpoint := ContractEndpoint{
				Method:      opData.method,
				Path:        path,
				OperationID: opData.op.OperationId,
				Responses:   make(map[string]ContractResponse),
				Security:    []ContractSecurity{},
			}

			// Extract parameters
			if opData.op.Parameters != nil {
				endpoint.Parameters = extractV3Parameters(opData.op.Parameters)
			}

			// Extract request body
			if opData.op.RequestBody != nil {
				endpoint.RequestBody = extractV3RequestBody(opData.op.RequestBody)
			}

			// Extract responses
			if opData.op.Responses != nil {
				endpoint.Responses = extractV3Responses(opData.op.Responses)
			}

			// Extract security requirements
			if opData.op.Security != nil {
				endpoint.Security = extractV3Security(opData.op.Security)
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints, nil
}

// extractV3Parameters extracts parameters from OpenAPI 3.x operation
func extractV3Parameters(params []*v3.Parameter) []ContractParameter {
	var result []ContractParameter
	for _, param := range params {
		if param == nil {
			continue
		}

		cp := ContractParameter{
			Name:        param.Name,
			In:          param.In,
			Required:    param.Required != nil && *param.Required,
			Description: param.Description,
		}

		// Extract schema (with $ref resolution)
		if param.Schema != nil {
			cp.Schema = param.Schema.Schema()
		}

		result = append(result, cp)
	}
	return result
}

// extractV3RequestBody extracts request body from OpenAPI 3.x operation
func extractV3RequestBody(reqBody *v3.RequestBody) *ContractRequestBody {
	if reqBody == nil {
		return nil
	}

	result := &ContractRequestBody{
		Required:     reqBody.Required != nil && *reqBody.Required,
		ContentTypes: make(map[string]*base.Schema),
	}

	if reqBody.Content != nil {
		for contentTypePair := reqBody.Content.First(); contentTypePair != nil; contentTypePair = contentTypePair.Next() {
			contentType := contentTypePair.Key()
			mediaType := contentTypePair.Value()
			if mediaType != nil && mediaType.Schema != nil {
				result.ContentTypes[contentType] = mediaType.Schema.Schema()
			}
		}
	}

	return result
}

// extractV3Responses extracts responses from OpenAPI 3.x operation
func extractV3Responses(responses *v3.Responses) map[string]ContractResponse {
	result := make(map[string]ContractResponse)

	if responses == nil {
		return result
	}

	// Extract status code responses
	if responses.Codes != nil {
		for codePair := responses.Codes.First(); codePair != nil; codePair = codePair.Next() {
			statusCode := codePair.Key()
			response := codePair.Value()
			if response == nil {
				continue
			}

			cr := ContractResponse{
				Description:  response.Description,
				ContentTypes: make(map[string]*base.Schema),
				Headers:      make(map[string]*base.Schema),
			}

			// Extract content
			if response.Content != nil {
				for contentTypePair := response.Content.First(); contentTypePair != nil; contentTypePair = contentTypePair.Next() {
					contentType := contentTypePair.Key()
					mediaType := contentTypePair.Value()
					if mediaType != nil && mediaType.Schema != nil {
						cr.ContentTypes[contentType] = mediaType.Schema.Schema()
					}
				}
			}

			// Extract headers
			if response.Headers != nil {
				for headerPair := response.Headers.First(); headerPair != nil; headerPair = headerPair.Next() {
					headerName := headerPair.Key()
					header := headerPair.Value()
					if header != nil && header.Schema != nil {
						cr.Headers[headerName] = header.Schema.Schema()
					}
				}
			}

			result[statusCode] = cr
		}
	}

	return result
}

// extractV3Security extracts security requirements from OpenAPI 3.x operation
func extractV3Security(security []*base.SecurityRequirement) []ContractSecurity {
	var result []ContractSecurity
	for _, sec := range security {
		if sec == nil {
			continue
		}

		cs := ContractSecurity{
			Schemes: []string{},
			Scopes:  []string{},
		}

		// Extract schemes and scopes
		if sec.Requirements != nil {
			for reqPair := sec.Requirements.First(); reqPair != nil; reqPair = reqPair.Next() {
				schemeName := reqPair.Key()
				scopes := reqPair.Value()
				cs.Schemes = append(cs.Schemes, schemeName)
				if scopes != nil {
					cs.Scopes = append(cs.Scopes, scopes...)
				}
			}
		}

		result = append(result, cs)
	}
	return result
}

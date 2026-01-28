// Package services provides Swagger 2.0 parsing support
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
)

// extractSwagger2Endpoints extracts endpoints from Swagger 2.0 specification
func extractSwagger2Endpoints(ctx context.Context, document libopenapi.Document) ([]ContractEndpoint, error) {
	model, err := document.BuildV2Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build Swagger 2.0 model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("Swagger 2.0 model is nil")
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

		// Extract operations
		operations := []struct {
			method string
			op     *v2.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
			{"HEAD", pathItem.Head},
			{"OPTIONS", pathItem.Options},
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
				endpoint.Parameters = extractV2Parameters(opData.op.Parameters)
			}

			// Extract responses
			if opData.op.Responses != nil {
				endpoint.Responses = extractV2Responses(opData.op.Responses)
			}

			// Extract security requirements
			if opData.op.Security != nil {
				endpoint.Security = extractV2Security(opData.op.Security)
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints, nil
}

// extractV2Parameters extracts parameters from Swagger 2.0 operation
func extractV2Parameters(params []*v2.Parameter) []ContractParameter {
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

		// Extract schema
		if param.Schema != nil {
			cp.Schema = param.Schema.Schema()
		}

		result = append(result, cp)
	}
	return result
}

// extractV2Responses extracts responses from Swagger 2.0 operation
func extractV2Responses(responses *v2.Responses) map[string]ContractResponse {
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

			// Extract schema
			if response.Schema != nil {
				// Swagger 2.0 uses a single schema, default to application/json
				cr.ContentTypes["application/json"] = response.Schema.Schema()
			}

			// Extract headers
			// Note: Swagger 2.0 headers use Items which doesn't have Schema() method
			// Headers in Swagger 2.0 are simpler and don't use full schema objects
			// We skip header schema extraction for Swagger 2.0 as it's not critical for validation
			if response.Headers != nil {
				// Headers exist but we don't extract schema for Swagger 2.0
				// This is acceptable as header validation is less critical
			}

			result[statusCode] = cr
		}
	}

	return result
}

// extractV2Security extracts security requirements from Swagger 2.0 operation
func extractV2Security(security []*base.SecurityRequirement) []ContractSecurity {
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

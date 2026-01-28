// Package services provides OpenAPI/Swagger contract parsing and validation tests
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPIContract_OpenAPI3YAML(t *testing.T) {
	// Create a temporary OpenAPI 3.0 YAML file
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: integer
                    name:
                      type: string
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	contract, err := ParseOpenAPIContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("ParseOpenAPIContract returned nil contract")
	}

	if contract.Version != "3.0.0" {
		t.Errorf("Expected version 3.0.0, got %s", contract.Version)
	}

	if len(contract.Endpoints) == 0 {
		t.Fatal("Expected at least one endpoint, got none")
	}

	endpoint := contract.Endpoints[0]
	if endpoint.Method != "GET" {
		t.Errorf("Expected method GET, got %s", endpoint.Method)
	}

	if endpoint.Path != "/users" {
		t.Errorf("Expected path /users, got %s", endpoint.Path)
	}

	if endpoint.OperationID != "getUsers" {
		t.Errorf("Expected operationId getUsers, got %s", endpoint.OperationID)
	}
}

func TestParseOpenAPIContract_OpenAPI3JSON(t *testing.T) {
	// Create a temporary OpenAPI 3.0 JSON file
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.json")
	contractContent := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/users/{id}": {
      "get": {
        "operationId": "getUser",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "id": {"type": "integer"},
                    "name": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	contract, err := ParseOpenAPIContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("ParseOpenAPIContract returned nil contract")
	}

	if contract.Version != "3.0.0" {
		t.Errorf("Expected version 3.0.0, got %s", contract.Version)
	}

	if len(contract.Endpoints) == 0 {
		t.Fatal("Expected at least one endpoint, got none")
	}

	endpoint := contract.Endpoints[0]
	if endpoint.Method != "GET" {
		t.Errorf("Expected method GET, got %s", endpoint.Method)
	}

	if endpoint.Path != "/users/{id}" {
		t.Errorf("Expected path /users/{id}, got %s", endpoint.Path)
	}

	if len(endpoint.Parameters) == 0 {
		t.Fatal("Expected at least one parameter, got none")
	}

	param := endpoint.Parameters[0]
	if param.Name != "id" {
		t.Errorf("Expected parameter name id, got %s", param.Name)
	}

	if param.In != "path" {
		t.Errorf("Expected parameter in path, got %s", param.In)
	}

	if !param.Required {
		t.Error("Expected parameter to be required")
	}
}

func TestParseOpenAPIContract_Swagger2YAML(t *testing.T) {
	// Create a temporary Swagger 2.0 YAML file
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "swagger.yaml")
	contractContent := `swagger: "2.0"
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      parameters:
        - name: body
          in: body
          required: true
          schema:
            type: object
            properties:
              name:
                type: string
      responses:
        '201':
          description: Created
          schema:
            type: object
            properties:
              id:
                type: integer
              name:
                type: string
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	contract, err := ParseOpenAPIContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("ParseOpenAPIContract returned nil contract")
	}

	if contract.Version != "2.0" {
		t.Errorf("Expected version 2.0, got %s", contract.Version)
	}

	if len(contract.Endpoints) == 0 {
		t.Fatal("Expected at least one endpoint, got none")
	}

	endpoint := contract.Endpoints[0]
	if endpoint.Method != "POST" {
		t.Errorf("Expected method POST, got %s", endpoint.Method)
	}

	if endpoint.Path != "/users" {
		t.Errorf("Expected path /users, got %s", endpoint.Path)
	}
}

func TestParseOpenAPIContract_Swagger2JSON(t *testing.T) {
	// Create a temporary Swagger 2.0 JSON file
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "swagger.json")
	contractContent := `{
  "swagger": "2.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/users/{id}": {
      "delete": {
        "operationId": "deleteUser",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "type": "integer"
          }
        ],
        "responses": {
          "204": {
            "description": "No Content"
          }
        }
      }
    }
  }
}`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	contract, err := ParseOpenAPIContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("ParseOpenAPIContract returned nil contract")
	}

	if contract.Version != "2.0" {
		t.Errorf("Expected version 2.0, got %s", contract.Version)
	}

	if len(contract.Endpoints) == 0 {
		t.Fatal("Expected at least one endpoint, got none")
	}

	endpoint := contract.Endpoints[0]
	if endpoint.Method != "DELETE" {
		t.Errorf("Expected method DELETE, got %s", endpoint.Method)
	}
}

func TestParseOpenAPIContract_InvalidFile(t *testing.T) {
	ctx := context.Background()
	_, err := ParseOpenAPIContract(ctx, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestParseOpenAPIContract_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "invalid.yaml")
	contractContent := `this is not valid yaml: [`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	_, err := ParseOpenAPIContract(ctx, contractFile)
	if err == nil {
		t.Fatal("Expected error for invalid contract content, got nil")
	}
}

func TestParseOpenAPIContract_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := ParseOpenAPIContract(ctx, contractFile)
	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Express.js style path parameter",
			input:    "/users/:id",
			expected: "/users/{id}",
		},
		{
			name:     "OpenAPI style path parameter",
			input:    "/users/{id}",
			expected: "/users/{id}",
		},
		{
			name:     "Multiple path parameters",
			input:    "/users/:userId/posts/:postId",
			expected: "/users/{userId}/posts/{postId}",
		},
		{
			name:     "No path parameters",
			input:    "/users",
			expected: "/users",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Path with query string",
			input:    "/users/:id",
			expected: "/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindMatchingContractEndpoint(t *testing.T) {
	contract := &OpenAPIContract{
		Endpoints: []ContractEndpoint{
			{
				Method: "GET",
				Path:   "/users/{id}",
			},
			{
				Method: "POST",
				Path:   "/users",
			},
			{
				Method: "GET",
				Path:   "/posts/{id}",
			},
		},
	}

	tests := []struct {
		name           string
		endpoint       EndpointInfo
		expectedFound  bool
		expectedMethod string
		expectedPath   string
	}{
		{
			name: "Match OpenAPI style path",
			endpoint: EndpointInfo{
				Method: "GET",
				Path:   "/users/{id}",
			},
			expectedFound:  true,
			expectedMethod: "GET",
			expectedPath:   "/users/{id}",
		},
		{
			name: "Match Express.js style path",
			endpoint: EndpointInfo{
				Method: "GET",
				Path:   "/users/:id",
			},
			expectedFound:  true,
			expectedMethod: "GET",
			expectedPath:   "/users/{id}",
		},
		{
			name: "Match POST endpoint",
			endpoint: EndpointInfo{
				Method: "POST",
				Path:   "/users",
			},
			expectedFound:  true,
			expectedMethod: "POST",
			expectedPath:   "/users",
		},
		{
			name: "No match - different method",
			endpoint: EndpointInfo{
				Method: "PUT",
				Path:   "/users/{id}",
			},
			expectedFound: false,
		},
		{
			name: "No match - different path",
			endpoint: EndpointInfo{
				Method: "GET",
				Path:   "/comments/{id}",
			},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMatchingContractEndpoint(tt.endpoint, contract)
			if tt.expectedFound {
				if result == nil {
					t.Fatal("Expected to find matching endpoint, got nil")
				}
				if result.Method != tt.expectedMethod {
					t.Errorf("Expected method %s, got %s", tt.expectedMethod, result.Method)
				}
				if result.Path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, result.Path)
				}
			} else {
				if result != nil {
					t.Errorf("Expected no match, got endpoint %s %s", result.Method, result.Path)
				}
			}
		})
	}
}

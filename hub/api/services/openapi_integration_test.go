// Package services provides integration tests for OpenAPI contract validation
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAPIContracts_RealWorldOpenAPI3(t *testing.T) {
	// Create a real-world OpenAPI 3.0 spec with $refs
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: User Management API
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email
      properties:
        id:
          type: integer
          minimum: 1
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
    Error:
      type: object
      required:
        - message
      properties:
        message:
          type: string
        code:
          type: integer
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - BearerAuth: []
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
          required: false
        - name: offset
          in: query
          schema:
            type: integer
            minimum: 0
          required: false
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    post:
      operationId: createUser
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
              properties:
                name:
                  type: string
                  minLength: 1
                email:
                  type: string
                  format: email
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /users/{id}:
    get:
      operationId: getUser
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            minimum: 1
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not Found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    delete:
      operationId: deleteUser
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            minimum: 1
      responses:
        '204':
          description: No Content
        '404':
          description: Not Found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	// Test endpoints
	endpoints := []EndpointInfo{
		{
			Method: "GET",
			Path:   "/users",
			File:   "handlers/users.go",
			Parameters: []ParameterInfo{
				{Name: "limit", Type: "query", DataType: "int"},
				{Name: "offset", Type: "query", DataType: "int"},
			},
			Responses: []ResponseInfo{
				{StatusCode: 200},
				{StatusCode: 401},
			},
			Auth: []string{"BearerAuth"},
		},
		{
			Method: "POST",
			Path:   "/users",
			File:   "handlers/users.go",
			Parameters: []ParameterInfo{
				{Name: "body", Type: "body", DataType: "object"},
			},
			Responses: []ResponseInfo{
				{StatusCode: 201},
				{StatusCode: 400},
			},
			Auth: []string{"BearerAuth"},
		},
		{
			Method: "GET",
			Path:   "/users/:id",
			File:   "handlers/users.go",
			Parameters: []ParameterInfo{
				{Name: "id", Type: "path", DataType: "int", Required: true},
			},
			Responses: []ResponseInfo{
				{StatusCode: 200},
				{StatusCode: 404},
			},
			Auth: []string{"BearerAuth"},
		},
		{
			Method: "DELETE",
			Path:   "/users/:id",
			File:   "handlers/users.go",
			Parameters: []ParameterInfo{
				{Name: "id", Type: "path", DataType: "int", Required: true},
			},
			Responses: []ResponseInfo{
				{StatusCode: 204},
				{StatusCode: 404},
			},
			Auth: []string{"BearerAuth"},
		},
	}

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	// Should have minimal findings for well-matched endpoints
	if len(findings) > 0 {
		t.Logf("Found %d validation findings:", len(findings))
		for _, finding := range findings {
			t.Logf("  - %s: %s (severity: %s)", finding.Type, finding.Issue, finding.Severity)
		}
	}
}

func TestValidateAPIContracts_MissingRequiredParameter(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	// Endpoint missing required parameter
	endpoints := []EndpointInfo{
		{
			Method: "GET",
			Path:   "/users/:id",
			File:   "handlers/users.go",
			// Missing id parameter
			Parameters: []ParameterInfo{},
			Responses:  []ResponseInfo{{StatusCode: 200}},
		},
	}

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	// Should find missing required parameter
	found := false
	for _, finding := range findings {
		if finding.Type == "contract_mismatch" && 
		   (strings.Contains(finding.Issue, "required") || strings.Contains(finding.Issue, "missing")) {
			found = true
			if finding.Severity != "high" && finding.Severity != "critical" {
				t.Errorf("Expected high/critical severity for missing required parameter, got %s", finding.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected finding for missing required parameter, but none found")
	}
}

func TestValidateAPIContracts_MissingSecurity(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - BearerAuth: []
      responses:
        '200':
          description: Success
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	// Endpoint without security
	endpoints := []EndpointInfo{
		{
			Method: "GET",
			Path:   "/users",
			File:   "handlers/users.go",
			// No Auth field
			Auth:     []string{},
			Responses: []ResponseInfo{{StatusCode: 200}},
		},
	}

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	// Should find missing security
	found := false
	for _, finding := range findings {
		if finding.Type == "contract_mismatch" && 
		   (strings.Contains(finding.Issue, "security") || strings.Contains(finding.Issue, "Security")) {
			found = true
			if finding.Severity != "critical" {
				t.Errorf("Expected critical severity for missing security, got %s", finding.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected finding for missing security, but none found")
	}
}

func TestValidateAPIContracts_ResponseMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        '200':
          description: Success
        '400':
          description: Bad Request
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	// Endpoint with response not in contract
	endpoints := []EndpointInfo{
		{
			Method: "GET",
			Path:   "/users",
			File:   "handlers/users.go",
			Responses: []ResponseInfo{
				{StatusCode: 200},
				{StatusCode: 500}, // Not in contract
			},
		},
	}

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	// Should find response mismatch
	found := false
	for _, finding := range findings {
		if finding.Type == "contract_mismatch" && 
		   strings.Contains(finding.Issue, "500") {
			found = true
			if finding.Severity != "high" {
				t.Errorf("Expected high severity for response mismatch, got %s", finding.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected finding for response mismatch, but none found")
	}
}

func TestValidateAPIContracts_Swagger2RealWorld(t *testing.T) {
	// Create a real-world Swagger 2.0 spec
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "swagger.yaml")
	contractContent := `swagger: "2.0"
info:
  title: Pet Store API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      parameters:
        - name: limit
          in: query
          type: integer
          minimum: 1
          maximum: 100
        - name: offset
          in: query
          type: integer
          minimum: 0
      responses:
        '200':
          description: Success
          schema:
            type: array
            items:
              type: object
              properties:
                id:
                  type: integer
                name:
                  type: string
    post:
      operationId: createPet
      parameters:
        - name: body
          in: body
          required: true
          schema:
            type: object
            required:
              - name
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
  /pets/{id}:
    get:
      operationId: getPet
      parameters:
        - name: id
          in: path
          required: true
          type: integer
      responses:
        '200':
          description: Success
          schema:
            type: object
            properties:
              id:
                type: integer
              name:
                type: string
        '404':
          description: Not Found
`

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	endpoints := []EndpointInfo{
		{
			Method: "GET",
			Path:   "/pets",
			File:   "handlers/pets.go",
			Parameters: []ParameterInfo{
				{Name: "limit", Type: "query", DataType: "int"},
				{Name: "offset", Type: "query", DataType: "int"},
			},
			Responses: []ResponseInfo{{StatusCode: 200}},
		},
		{
			Method: "POST",
			Path:   "/pets",
			File:   "handlers/pets.go",
			Parameters: []ParameterInfo{
				{Name: "body", Type: "body", DataType: "object"},
			},
			Responses: []ResponseInfo{{StatusCode: 201}},
		},
		{
			Method: "GET",
			Path:   "/pets/:id",
			File:   "handlers/pets.go",
			Parameters: []ParameterInfo{
				{Name: "id", Type: "path", DataType: "int", Required: true},
			},
			Responses: []ResponseInfo{
				{StatusCode: 200},
				{StatusCode: 404},
			},
		},
	}

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	// Should have minimal findings for well-matched endpoints
	if len(findings) > 0 {
		t.Logf("Found %d validation findings:", len(findings))
		for _, finding := range findings {
			t.Logf("  - %s: %s (severity: %s)", finding.Type, finding.Issue, finding.Severity)
		}
	}
}

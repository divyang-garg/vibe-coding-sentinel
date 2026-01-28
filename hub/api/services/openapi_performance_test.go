// Package services provides performance benchmarks for OpenAPI validation
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkParseOpenAPIContract_LargeContract(b *testing.B) {
	// Create a large OpenAPI contract with many endpoints
	tmpDir := b.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")

	// Generate contract with 100 endpoints
	contractContent := `openapi: 3.0.0
info:
  title: Large API
  version: 1.0.0
paths:
`
	for i := 0; i < 100; i++ {
		contractContent += fmt.Sprintf(`  /resource%d/{id}:
    get:
      operationId: getResource%d
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
`, i, i)
	}

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		b.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseOpenAPIContract(ctx, contractFile)
		if err != nil {
			b.Fatalf("ParseOpenAPIContract failed: %v", err)
		}
	}
}

func BenchmarkValidateAPIContracts_ManyEndpoints(b *testing.B) {
	tmpDir := b.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")
	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
`

	// Create contract with 50 endpoints (unique paths to avoid YAML duplicate-key errors)
	for i := 0; i < 50; i++ {
		contractContent += fmt.Sprintf(`  /endpoint%d:
    get:
      operationId: getEndpoint%d
      responses:
        '200':
          description: Success
`, i, i)
	}

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		b.Fatalf("Failed to write test contract file: %v", err)
	}

	// Create 50 endpoints to validate
	endpoints := make([]EndpointInfo, 50)
	for i := 0; i < 50; i++ {
		endpoints[i] = EndpointInfo{
			Method: "GET",
			Path:   fmt.Sprintf("/endpoint%d", i),
			File:   "handlers/endpoint.go",
			Responses: []ResponseInfo{
				{StatusCode: 200},
			},
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := validateAPIContracts(ctx, tmpDir, endpoints)
		if err != nil {
			b.Fatalf("validateAPIContracts failed: %v", err)
		}
	}
}

func BenchmarkContractCache_Hit(b *testing.B) {
	tmpDir := b.TempDir()
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
		b.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()

	// Prime the cache
	_, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		b.Fatalf("Failed to prime cache: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := GetCachedContract(ctx, contractFile)
		if err != nil {
			b.Fatalf("GetCachedContract failed: %v", err)
		}
	}
}

func TestPerformance_Parse1000Endpoints(t *testing.T) {
	// Performance test: Parse contract with 1000 endpoints in < 1 second
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")

	contractContent := `openapi: 3.0.0
info:
  title: Large API
  version: 1.0.0
paths:
`

	// Generate contract with 1000 endpoints (unique paths to avoid YAML duplicate-key errors)
	for i := 0; i < 1000; i++ {
		contractContent += fmt.Sprintf(`  /resource%d/{id}:
    get:
      operationId: getResource%d
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Success
`, i, i)
	}

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	ctx := context.Background()
	start := time.Now()

	contract, err := ParseOpenAPIContract(ctx, contractFile)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ParseOpenAPIContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("ParseOpenAPIContract returned nil contract")
	}

	if elapsed > 1*time.Second {
		t.Errorf("ParseOpenAPIContract took %v, expected < 1 second", elapsed)
	}

	t.Logf("Parsed 1000-endpoint contract in %v", elapsed)
}

func TestPerformance_Validate100Endpoints(t *testing.T) {
	// Performance test: Validate 100 endpoints in < 500ms
	tmpDir := t.TempDir()
	contractFile := filepath.Join(tmpDir, "openapi.yaml")

	contractContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
`

	// Create contract with 100 endpoints (unique paths to avoid YAML duplicate-key errors)
	for i := 0; i < 100; i++ {
		contractContent += fmt.Sprintf(`  /endpoint%d:
    get:
      operationId: getEndpoint%d
      responses:
        '200':
          description: Success
`, i, i)
	}

	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("Failed to write test contract file: %v", err)
	}

	// Create 100 endpoints to validate
	endpoints := make([]EndpointInfo, 100)
	for i := 0; i < 100; i++ {
		endpoints[i] = EndpointInfo{
			Method: "GET",
			Path:   fmt.Sprintf("/endpoint%d", i),
			File:   "handlers/endpoint.go",
			Responses: []ResponseInfo{
				{StatusCode: 200},
			},
		}
	}

	ctx := context.Background()
	start := time.Now()

	findings, err := validateAPIContracts(ctx, tmpDir, endpoints)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("validateAPIContracts failed: %v", err)
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("validateAPIContracts took %v, expected < 500ms", elapsed)
	}

	t.Logf("Validated 100 endpoints in %v, found %d issues", elapsed, len(findings))
}

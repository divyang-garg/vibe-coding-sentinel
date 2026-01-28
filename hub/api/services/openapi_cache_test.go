// Package services provides tests for OpenAPI contract caching
// Complies with CODING_STANDARDS.md: Test coverage 90%+
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetCachedContract_FirstCall(t *testing.T) {
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

	ctx := context.Background()
	contract, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("GetCachedContract failed: %v", err)
	}

	if contract == nil {
		t.Fatal("GetCachedContract returned nil contract")
	}

	if contract.Version != "3.0.0" {
		t.Errorf("Expected version 3.0.0, got %s", contract.Version)
	}
}

func TestGetCachedContract_CacheHit(t *testing.T) {
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

	ctx := context.Background()

	// First call - should parse
	contract1, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("GetCachedContract (first call) failed: %v", err)
	}

	// Second call - should use cache
	start := time.Now()
	contract2, err := GetCachedContract(ctx, contractFile)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetCachedContract (second call) failed: %v", err)
	}

	// Should be the same contract instance (cached)
	if contract1 != contract2 {
		t.Error("Expected cached contract to be the same instance")
	}

	// Should be much faster (cache hit)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Cache hit took %v, expected < 100ms", elapsed)
	}
}

func TestGetCachedContract_FileModified(t *testing.T) {
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

	ctx := context.Background()

	// First call - should parse
	contract1, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("GetCachedContract (first call) failed: %v", err)
	}

	// Modify file
	newContent := contractContent + "\n# Modified"
	if err := os.WriteFile(contractFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to modify test contract file: %v", err)
	}

	// Wait a bit to ensure file modification time is different
	time.Sleep(10 * time.Millisecond)

	// Second call - should re-parse due to file modification
	contract2, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("GetCachedContract (second call) failed: %v", err)
	}

	// Should be different instances (cache invalidated)
	if contract1 == contract2 {
		t.Error("Expected contract to be re-parsed after file modification")
	}
}

func TestContractCache_SetTTL(t *testing.T) {
	cache := GetContractCache()
	cache.SetTTL(10 * time.Minute)

	if cache.ttl != 10*time.Minute {
		t.Errorf("Expected TTL 10 minutes, got %v", cache.ttl)
	}
}

func TestContractCache_ClearCache(t *testing.T) {
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

	ctx := context.Background()

	// Prime cache
	_, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("Failed to prime cache: %v", err)
	}

	cache := GetContractCache()
	stats := cache.GetCacheStats()
	if stats["total_contracts"].(int) == 0 {
		t.Fatal("Expected cache to have contracts")
	}

	// Clear cache
	cache.ClearCache()

	stats = cache.GetCacheStats()
	if stats["total_contracts"].(int) != 0 {
		t.Error("Expected cache to be empty after ClearCache")
	}
}

func TestContractCache_RemoveCache(t *testing.T) {
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

	ctx := context.Background()

	// Prime cache
	_, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("Failed to prime cache: %v", err)
	}

	cache := GetContractCache()
	cache.RemoveCache(contractFile)

	stats := cache.GetCacheStats()
	if stats["total_contracts"].(int) != 0 {
		t.Error("Expected cache to be empty after RemoveCache")
	}
}

func TestContractCache_GetCacheStats(t *testing.T) {
	cache := GetContractCache()
	stats := cache.GetCacheStats()

	if stats["total_contracts"] == nil {
		t.Error("Expected total_contracts in stats")
	}
	if stats["valid_contracts"] == nil {
		t.Error("Expected valid_contracts in stats")
	}
	if stats["expired_contracts"] == nil {
		t.Error("Expected expired_contracts in stats")
	}
	if stats["ttl_seconds"] == nil {
		t.Error("Expected ttl_seconds in stats")
	}
}

func TestContractCache_CleanupExpired(t *testing.T) {
	cache := GetContractCache()
	cache.SetTTL(1 * time.Millisecond) // Very short TTL

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

	ctx := context.Background()

	// Prime cache
	_, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		t.Fatalf("Failed to prime cache: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Cleanup expired
	cache.CleanupExpired()

	stats := cache.GetCacheStats()
	if stats["total_contracts"].(int) != 0 {
		t.Error("Expected expired contracts to be cleaned up")
	}
}

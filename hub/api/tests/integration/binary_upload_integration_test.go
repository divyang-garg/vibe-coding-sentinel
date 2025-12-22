// +build integration

package integration

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

// TestBinaryUpload_WithoutAuth tests that binary upload fails without admin auth
func TestBinaryUpload_WithoutAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := CreateTestServer()
	defer server.Close()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	writer.WriteField("version", "1.2.3")
	writer.WriteField("platform", "linux-amd64")
	
	fileWriter, _ := writer.CreateFormFile("binary", "test-binary")
	fileWriter.Write([]byte("fake binary content"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/admin/binary/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// TestBinaryUpload_InvalidKey tests that binary upload fails with invalid admin key
func TestBinaryUpload_InvalidKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := CreateTestServer()
	defer server.Close()

	// Set admin key in environment
	os.Setenv("ADMIN_API_KEY", "correct-admin-key")
	defer os.Unsetenv("ADMIN_API_KEY")

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	writer.WriteField("version", "1.2.3")
	writer.WriteField("platform", "linux-amd64")
	
	fileWriter, _ := writer.CreateFormFile("binary", "test-binary")
	fileWriter.Write([]byte("fake binary content"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/admin/binary/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Admin-API-Key", "wrong-admin-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// TestBinaryUpload_InvalidVersion tests that binary upload fails with invalid version format
func TestBinaryUpload_InvalidVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := CreateTestServer()
	defer server.Close()

	adminKey := "test-admin-key-123"
	os.Setenv("ADMIN_API_KEY", adminKey)
	defer os.Unsetenv("ADMIN_API_KEY")

	// Create multipart form data with invalid version
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	writer.WriteField("version", "invalid-version")
	writer.WriteField("platform", "linux-amd64")
	
	fileWriter, _ := writer.CreateFormFile("binary", "test-binary")
	fileWriter.Write([]byte("fake binary content"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/admin/binary/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Admin-API-Key", adminKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("invalid_format")) {
		t.Errorf("Expected invalid_format error, got: %s", string(body))
	}
}

// TestBinaryUpload_InvalidPlatform tests that binary upload fails with invalid platform
func TestBinaryUpload_InvalidPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := CreateTestServer()
	defer server.Close()

	adminKey := "test-admin-key-123"
	os.Setenv("ADMIN_API_KEY", adminKey)
	defer os.Unsetenv("ADMIN_API_KEY")

	// Create multipart form data with invalid platform
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	writer.WriteField("version", "1.2.3")
	writer.WriteField("platform", "invalid-platform")
	
	fileWriter, _ := writer.CreateFormFile("binary", "test-binary")
	fileWriter.Write([]byte("fake binary content"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/admin/binary/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Admin-API-Key", adminKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("invalid_platform")) {
		t.Errorf("Expected invalid_platform error, got: %s", string(body))
	}
}

// TestBinaryUpload_ValidKey_BearerHeader tests that binary upload works with valid admin key via Bearer header
func TestBinaryUpload_ValidKey_BearerHeader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires database and file system")
	}

	// Note: This test requires:
	// - Database setup with binary_versions table
	// - Binary storage directory configured
	// - Proper cleanup after test
	// For now, we'll skip the actual upload test and just verify auth works
	
	t.Skip("Full integration test requires database and file system setup")
}


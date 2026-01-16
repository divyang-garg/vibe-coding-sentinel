// Package feature_discovery provides tests for API endpoint discovery
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestDiscoverAPIEndpoints_Express tests Express endpoint discovery
func TestDiscoverAPIEndpoints_Express(t *testing.T) {
	tempDir := t.TempDir()

	// Create Express route file
	routeContent := `const express = require('express');
const app = express();

app.get('/users/:id', (req, res) => {
  res.json({ user: req.params.id });
});

app.post('/users', (req, res) => {
  res.status(201).json({ message: 'User created' });
});

module.exports = app;`

	// Create routes directory structure
	routesDir := filepath.Join(tempDir, "routes")
	err := os.MkdirAll(routesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create routes directory: %v", err)
	}

	routePath := filepath.Join(routesDir, "userRoutes.js")
	err = os.WriteFile(routePath, []byte(routeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create route file: %v", err)
	}

	// Test direct file parsing to debug
	data, err := os.ReadFile(routePath)
	if err != nil {
		t.Fatalf("Failed to read route file: %v", err)
	}

	content := string(data)
	t.Logf("File content:\n%s", content)

	parsedEndpoints := parseExpressRoutes(content, routePath, "user")
	t.Logf("Parsed endpoints: %d", len(parsedEndpoints))

	// Check if parsing worked
	if len(parsedEndpoints) == 0 {
		t.Errorf("parseExpressRoutes returned no endpoints")
		// Debug: check if file contains expected patterns
		if !strings.Contains(content, "app.get") {
			t.Errorf("File does not contain app.get")
		}
		if !strings.Contains(content, "app.post") {
			t.Errorf("File does not contain app.post")
		}
		// Test regex directly
		re := regexp.MustCompile(`(?i)app\.get\(\s*['"]([^'"]+)['"]`)
		matches := re.FindAllStringSubmatch(content, -1)
		t.Logf("Regex matches for GET: %v", matches)

		re2 := regexp.MustCompile(`(?i)app\.post\(\s*['"]([^'"]+)['"]`)
		matches2 := re2.FindAllStringSubmatch(content, -1)
		t.Logf("Regex matches for POST: %v", matches2)
		return
	}

	// Test endpoint discovery
	endpoints, err := discoverAPIEndpoints(context.Background(), tempDir, "user", "express")
	if err != nil {
		t.Fatalf("discoverAPIEndpoints returned error: %v", err)
	}

	if endpoints.Framework != "express" {
		t.Errorf("Expected framework 'express', got '%s'", endpoints.Framework)
	}

	if len(endpoints.Endpoints) == 0 {
		t.Errorf("Expected to find endpoints, but found none")
	}

	// Check for GET /users/:id endpoint
	foundGet := false
	foundPost := false
	for _, endpoint := range endpoints.Endpoints {
		if endpoint.Method == "GET" && endpoint.Path == "/users/:id" {
			foundGet = true
			if len(endpoint.Parameters) == 0 {
				t.Errorf("Expected parameters for GET /users/:id")
			}
		}
		if endpoint.Method == "POST" && endpoint.Path == "/users" {
			foundPost = true
		}
	}

	if !foundGet {
		t.Errorf("Expected to find GET /users/:id endpoint")
	}
	if !foundPost {
		t.Errorf("Expected to find POST /users endpoint")
	}
}

// TestDiscoverAPIEndpoints_FastAPI tests FastAPI endpoint discovery
func TestDiscoverAPIEndpoints_FastAPI(t *testing.T) {
	tempDir := t.TempDir()

	// Create FastAPI route file
	apiContent := `from fastapi import FastAPI
from typing import Optional

app = FastAPI()

@app.get("/items/{item_id}")
async def read_item(item_id: int, q: Optional[str] = None):
    return {"item_id": item_id, "q": q}

@app.post("/items", status_code=201)
async def create_item(name: str, price: float):
    return {"name": name, "price": price}`

	apiPath := filepath.Join(tempDir, "main.py")
	err := os.WriteFile(apiPath, []byte(apiContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create API file: %v", err)
	}

	// Test endpoint discovery
	endpoints, err := discoverAPIEndpoints(context.Background(), tempDir, "item", "fastapi")
	if err != nil {
		t.Fatalf("discoverAPIEndpoints returned error: %v", err)
	}

	if len(endpoints.Endpoints) == 0 {
		t.Errorf("Expected to find endpoints, but found none")
	}
}

// TestDiscoverAPIEndpoints_Django tests Django endpoint discovery
func TestDiscoverAPIEndpoints_Django(t *testing.T) {
	tempDir := t.TempDir()

	// Create Django urls.py file
	urlContent := `from django.urls import path
from . import views

urlpatterns = [
    path('articles/<int:year>/', views.year_archive),
    path('articles/<int:year>/<int:month>/', views.month_archive),
]`

	urlPath := filepath.Join(tempDir, "urls.py")
	err := os.WriteFile(urlPath, []byte(urlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create urls file: %v", err)
	}

	// Test endpoint discovery
	endpoints, err := discoverAPIEndpoints(context.Background(), tempDir, "article", "django")
	if err != nil {
		t.Fatalf("discoverAPIEndpoints returned error: %v", err)
	}

	if len(endpoints.Endpoints) == 0 {
		t.Errorf("Expected to find endpoints, but found none")
	}
}

// TestDiscoverAPIEndpoints_Gin tests Gin endpoint discovery
func TestDiscoverAPIEndpoints_Gin(t *testing.T) {
	tempDir := t.TempDir()

	// Create Gin route file
	ginContent := `package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/users/:id", getUser)
	r.POST("/users", createUser)

	return r
}`

	ginPath := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(ginPath, []byte(ginContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create Gin file: %v", err)
	}

	// Debug: test direct parsing
	data, err := os.ReadFile(ginPath)
	if err != nil {
		t.Fatalf("Failed to read Gin file: %v", err)
	}

	parsedEndpoints := parseGoRoutes(string(data), ginPath, "user", "gin")
	t.Logf("Parsed Gin endpoints: %d", len(parsedEndpoints))

	// Test endpoint discovery
	endpoints, err := discoverAPIEndpoints(context.Background(), tempDir, "user", "gin")
	if err != nil {
		t.Fatalf("discoverAPIEndpoints returned error: %v", err)
	}

	if len(endpoints.Endpoints) == 0 {
		t.Errorf("Expected to find endpoints, but found none")
		// Debug: check if files are found
		files, _ := findFilesRecursively(tempDir, "*.go")
		t.Logf("Found Go files: %v", files)
	}
}

// TestDiscoverAPIEndpoints_NoEndpoints tests case with no matching endpoints
func TestDiscoverAPIEndpoints_NoEndpoints(t *testing.T) {
	tempDir := t.TempDir()

	// Create a non-API file
	randomFile := filepath.Join(tempDir, "utils.js")
	utilsContent := `export const formatDate = (date) => {
  return date.toISOString();
};`

	err := os.WriteFile(randomFile, []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create utils file: %v", err)
	}

	// Test endpoint discovery with non-matching feature name
	endpoints, err := discoverAPIEndpoints(context.Background(), tempDir, "nonexistent", "express")
	if err != nil {
		t.Fatalf("discoverAPIEndpoints returned error: %v", err)
	}

	if len(endpoints.Endpoints) != 0 {
		t.Errorf("Expected to find no endpoints, but found %d", len(endpoints.Endpoints))
	}
}

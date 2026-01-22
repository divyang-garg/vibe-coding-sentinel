// Package services tests for endpoint detector
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"testing"
)

func TestDetectEndpoints_Express(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
const express = require('express');
const app = express();

app.get('/api/users', (req, res) => {});
app.post('/api/users', (req, res) => {});
router.put('/api/products', (req, res) => {});
`
		keywords := []string{"users", "products"}

		// When
		endpoints := detectEndpoints(code, "server.js", keywords)

		// Then
		if len(endpoints) < 2 {
			t.Errorf("Expected at least 2 endpoints, got %d", len(endpoints))
		}

		// Check for expected endpoints
		foundUsers := false
		foundProducts := false
		for _, ep := range endpoints {
			if ep == "GET /api/users" || ep == "POST /api/users" {
				foundUsers = true
			}
			if ep == "PUT /api/products" {
				foundProducts = true
			}
		}

		if !foundUsers {
			t.Error("Expected to find /api/users endpoint")
		}
		if !foundProducts {
			t.Error("Expected to find /api/products endpoint")
		}
	})
}

func TestDetectEndpoints_FastAPI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
from fastapi import FastAPI
app = FastAPI()

@app.get("/api/users")
async def get_users():
    pass

@app.post("/api/orders")
async def create_order():
    pass
`
		keywords := []string{"users", "orders"}

		// When
		endpoints := detectEndpoints(code, "main.py", keywords)

		// Then
		if len(endpoints) < 2 {
			t.Errorf("Expected at least 2 endpoints, got %d", len(endpoints))
		}
	})
}

func TestDetectEndpoints_Go(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
package main

import "net/http"

func setupRoutes() {
    router.HandleFunc("/api/users", handleUsers)
    r.Get("/api/products", getProducts)
    r.Post("/api/orders", createOrder)
}
`
		keywords := []string{"users", "products", "orders"}

		// When
		endpoints := detectEndpoints(code, "routes.go", keywords)

		// Then
		if len(endpoints) < 2 {
			t.Errorf("Expected at least 2 endpoints, got %d", len(endpoints))
		}
	})
}

func TestDetectFramework(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		code     string
		want     string
	}{
		{"Express.js", "server.js", "const app = express();", "express"},
		{"FastAPI", "main.py", "from fastapi import FastAPI", "fastapi"},
		{"Django", "views.py", "from django import views", "django"},
		{"Go", "routes.go", "router.HandleFunc", "go"},
		{"Unknown", "unknown.txt", "some code", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFramework(tt.filePath, tt.code)
			if got != tt.want {
				t.Errorf("detectFramework() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesKeywords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keywords []string
		want     bool
	}{
		{"matches", "/api/users", []string{"users"}, true},
		{"no match", "/api/products", []string{"users"}, false},
		{"case insensitive", "/API/USERS", []string{"users"}, true},
		{"empty keywords", "/api/users", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesKeywords(tt.text, tt.keywords)
			if got != tt.want {
				t.Errorf("matchesKeywords() = %v, want %v", got, tt.want)
			}
		})
	}
}

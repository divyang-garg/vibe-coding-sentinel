//go:build integration
// +build integration

package main

import "net/http"

// ExportTestHandlerCaller exports handler caller constructor for integration tests
// This file is only compiled when building with -tags integration
// It allows integration tests to access test handler functionality

// TestHandlerCallerInterface defines the interface for test handler caller
// This allows integration tests to work with handlers without importing main package
type TestHandlerCallerInterface interface {
	CallValidateCodeHandler(w http.ResponseWriter, r *http.Request) error
	CallApplyFixHandler(w http.ResponseWriter, r *http.Request) error
	CallValidateLLMConfigHandler(w http.ResponseWriter, r *http.Request) error
	CallGetCacheMetricsHandler(w http.ResponseWriter, r *http.Request) error
	CallGetCostMetricsHandler(w http.ResponseWriter, r *http.Request) error
}

// Note: The actual TestHandlerCaller type is defined in test_handlers.go
// This file provides build-tag based access for integration tests

// Package mcp provides MCP server implementation
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Server represents an MCP server instance
type Server struct {
	tools     []Tool
	stats     *ServerStats
	startTime time.Time
}

// ServerStats tracks server statistics
type ServerStats struct {
	RequestsTotal   int64
	ErrorsTotal     int64
	ActiveRequests  int64
	LastRequestTime time.Time
}

// NewServer creates a new MCP server
func NewServer() *Server {
	return &Server{
		tools:     GetRegisteredTools(),
		stats:     &ServerStats{},
		startTime: time.Now(),
	}
}

// Start starts the MCP server, reading from stdin and writing to stdout
func (s *Server) Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Configure scanner with message size limits (10MB max)
	const maxMessageSize = 10 * 1024 * 1024
	scanner.Buffer(make([]byte, 1024), maxMessageSize)

	// Note: Do not reassign os.Stdout as it can cause file descriptor issues
	// in test environments. The encoder already uses os.Stdout directly.

	for scanner.Scan() {
		messageBytes := scanner.Bytes()

		// Validate message size
		if len(messageBytes) > maxMessageSize {
			resp := s.createErrorResponse(nil, InvalidRequestCode, "Message too large",
				fmt.Sprintf("Message size %d exceeds maximum allowed size %d", len(messageBytes), maxMessageSize))
			encoder.Encode(resp)
			continue
		}

		// Update server stats
		atomic.AddInt64(&s.stats.RequestsTotal, 1)
		s.stats.LastRequestTime = time.Now()

		// Process message
		resp := s.processMessage(messageBytes)
		if resp != nil {
			if err := encoder.Encode(resp); err != nil {
				// Log to stderr to avoid breaking JSON-RPC protocol
				fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
				atomic.AddInt64(&s.stats.ErrorsTotal, 1)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// processMessage processes a single JSON-RPC message
func (s *Server) processMessage(messageBytes []byte) *Response {
	// Validate JSON
	if !json.Valid(messageBytes) {
		return s.createErrorResponse(nil, ParseErrorCode, "Invalid JSON", "Message is not valid JSON")
	}

	// Parse into generic map for validation
	var rawMessage map[string]interface{}
	if err := json.Unmarshal(messageBytes, &rawMessage); err != nil {
		return s.createErrorResponse(nil, ParseErrorCode, "Parse error", err.Error())
	}

	// Validate JSON-RPC 2.0 format
	if err := s.validateJSONRPCFormat(rawMessage); err != nil {
		return s.createErrorResponse(rawMessage["id"], err.Code, err.Message, err.Data)
	}

	// Parse into structured request
	var req Request
	if err := json.Unmarshal(messageBytes, &req); err != nil {
		return s.createErrorResponse(rawMessage["id"], ParseErrorCode, "Parse error", err.Error())
	}

	// Handle request
	return s.handleRequest(req)
}

// validateJSONRPCFormat validates JSON-RPC 2.0 message structure
func (s *Server) validateJSONRPCFormat(msg map[string]interface{}) *Error {
	// Check jsonrpc version
	if jsonrpc, ok := msg["jsonrpc"]; !ok || jsonrpc != "2.0" {
		return &Error{
			Code:    InvalidRequestCode,
			Message: "Invalid JSON-RPC version",
			Data:    "jsonrpc field must be present and equal to '2.0'",
		}
	}

	// Check method field for requests
	if method, exists := msg["method"]; exists {
		if methodStr, ok := method.(string); !ok || methodStr == "" {
			return &Error{
				Code:    InvalidRequestCode,
				Message: "Invalid method",
				Data:    "method field must be a non-empty string",
			}
		}
	}

	return nil
}

// handleRequest routes requests to appropriate handlers
func (s *Server) handleRequest(req Request) *Response {
	atomic.AddInt64(&s.stats.ActiveRequests, 1)
	defer atomic.AddInt64(&s.stats.ActiveRequests, -1)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "ping":
		return s.handlePing(req)
	default:
		return s.createErrorResponse(req.ID, MethodNotFoundCode, "Method not found",
			fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// createErrorResponse creates an error response
func (s *Server) createErrorResponse(id interface{}, code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

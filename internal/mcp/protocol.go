// Package mcp provides MCP (Model Context Protocol) server implementation
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package mcp

import "encoding/json"

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc"` // "2.0"
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string      `json:"jsonrpc"` // "2.0"
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Error code constants
const (
	// JSON-RPC 2.0 Standard Errors
	ParseErrorCode     = -32700 // Invalid JSON was received
	InvalidRequestCode = -32600 // The JSON sent is not a valid Request object
	MethodNotFoundCode = -32601 // The method does not exist or is not available
	InvalidParamsCode  = -32602 // Invalid method parameter(s)
	InternalErrorCode  = -32603 // Internal JSON-RPC error

	// Custom MCP Errors
	HubUnavailableCode   = -32000 // Hub service is not available
	HubTimeoutCode       = -32001 // Hub request timed out
	ConfigErrorCode      = -32002 // Configuration error
	ServerOverloadedCode = -32003 // Server is overloaded
	RequestTimeoutCode   = -32004 // Request processing timed out
)

// InitializeParams represents MCP initialize request parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]string      `json:"clientInfo,omitempty"`
}

// InitializeResult represents MCP initialize response
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]string      `json:"serverInfo"`
}

// ToolCallParams represents MCP tool call parameters
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

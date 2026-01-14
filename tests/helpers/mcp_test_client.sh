#!/bin/bash
# MCP Test Client
# Helper script for sending MCP requests to Sentinel MCP server via stdio
# Usage: 
#   send_mcp_request <method> <params_json> [id]
#   send_mcp_request_and_wait <method> <params_json> [id] [timeout]

# Send MCP request and return response
send_mcp_request() {
    local method=$1
    local params=${2:-"{}"}
    local id=${3:-1}
    
    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $id,
  "method": "$method",
  "params": $params
}
EOF
)
    
    echo "$request" | ./sentinel mcp-server 2>/dev/null | head -1
}

# Send MCP request with timeout
send_mcp_request_and_wait() {
    local method=$1
    local params=${2:-"{}"}
    local id=${3:-1}
    local timeout=${4:-2}
    
    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $id,
  "method": "$method",
  "params": $params
}
EOF
)
    
    echo "$request" | timeout "$timeout" ./sentinel mcp-server 2>/dev/null | head -1
}

# Send initialize request
send_initialize() {
    local protocol_version=${1:-"2024-11-05"}
    local id=${2:-1}
    
    local params=$(cat <<EOF
{
  "protocolVersion": "$protocol_version",
  "capabilities": {},
  "clientInfo": {
    "name": "test-client",
    "version": "1.0"
  }
}
EOF
)
    
    send_mcp_request "initialize" "$params" "$id"
}

# Send tools/list request
send_tools_list() {
    local id=${1:-2}
    send_mcp_request "tools/list" "{}" "$id"
}

# Send tools/call request
send_tools_call() {
    local tool_name=$1
    local arguments=$2
    local id=${3:-3}
    
    local params=$(cat <<EOF
{
  "name": "$tool_name",
  "arguments": $arguments
}
EOF
)
    
    send_mcp_request "tools/call" "$params" "$id"
}

# Parse MCP response and extract fields
parse_mcp_response() {
    local response=$1
    local field=$2
    
    echo "$response" | grep -o "\"$field\":[^,}]*" | cut -d: -f2 | tr -d '"' | tr -d ' '
}

# Send task status request
send_task_status_request() {
    local task_id=$1
    local codebase_path=${2:-""}
    local id=${3:-1}
    
    local arguments="{\"taskId\": \"$task_id\""
    if [ -n "$codebase_path" ]; then
        arguments="$arguments, \"codebasePath\": \"$codebase_path\""
    fi
    arguments="$arguments}"
    
    send_tools_call "sentinel_get_task_status" "$arguments" "$id"
}

# Send verify task request
send_verify_task_request() {
    local task_id=$1
    local force=${2:-false}
    local codebase_path=${3:-""}
    local id=${4:-1}
    
    local arguments="{\"taskId\": \"$task_id\", \"force\": $force"
    if [ -n "$codebase_path" ]; then
        arguments="$arguments, \"codebasePath\": \"$codebase_path\""
    fi
    arguments="$arguments}"
    
    send_tools_call "sentinel_verify_task" "$arguments" "$id"
}

# Send list tasks request
send_list_tasks_request() {
    local filters=$1
    local id=${2:-1}
    
    send_tools_call "sentinel_list_tasks" "$filters" "$id"
}

# Assert task response has expected fields
assert_task_response() {
    local response=$1
    local field=$2
    local expected_value=${3:-""}
    
    if echo "$response" | grep -q "\"$field\""; then
        if [ -n "$expected_value" ]; then
            local actual_value=$(echo "$response" | grep -o "\"$field\":[^,}]*" | cut -d: -f2 | tr -d '"' | tr -d ' ')
            if [ "$actual_value" = "$expected_value" ]; then
                return 0
            else
                return 1
            fi
        fi
        return 0
    else
        return 1
    fi
}

# Assert error code in response
assert_error_code() {
    local response=$1
    local expected_code=$2
    
    local actual_code=$(get_error_code "$response")
    if [ "$actual_code" = "$expected_code" ]; then
        return 0
    else
        return 1
    fi
}

# Check if MCP response has error
has_mcp_error() {
    local response=$1
    echo "$response" | grep -q '"error"'
}

# Get error code from MCP response
get_mcp_error_code() {
    local response=$1
    echo "$response" | grep -o '"code":-[0-9]*' | cut -d: -f2
}

# Get error message from MCP response
get_mcp_error_message() {
    local response=$1
    echo "$response" | grep -o '"message":"[^"]*"' | cut -d: -f2 | tr -d '"'
}

# Check if MCP response has result
has_mcp_result() {
    local response=$1
    echo "$response" | grep -q '"result"'
}

# Example usage:
# response=$(send_initialize)
# if has_mcp_error "$response"; then
#     echo "Error: $(get_mcp_error_message "$response")"
# else
#     echo "Success: $(parse_mcp_response "$response" "protocolVersion")"
# fi



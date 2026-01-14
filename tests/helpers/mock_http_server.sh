#!/bin/bash
# Mock HTTP Server for Testing Retry Logic
# Usage: source mock_http_server.sh; start_mock_server <port> <response_code> <fail_count>

MOCK_SERVER_PID=""
MOCK_SERVER_PORT=""
MOCK_SERVER_RESPONSE_CODE=""
MOCK_SERVER_FAIL_COUNT=""
MOCK_SERVER_CALL_COUNT=0

start_mock_server() {
    local port=$1
    local response_code=${2:-200}
    local fail_count=${3:-0}
    
    MOCK_SERVER_PORT=$port
    MOCK_SERVER_RESPONSE_CODE=$response_code
    MOCK_SERVER_FAIL_COUNT=$fail_count
    MOCK_SERVER_CALL_COUNT=0
    
    # Create a simple HTTP server that fails N times then succeeds
    cat > /tmp/mock_server_${port}.py <<EOF
import http.server
import socketserver
import sys

call_count = 0
fail_count = ${fail_count}
response_code = ${response_code}

class MockHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.handle_request()
    
    def do_POST(self):
        self.handle_request()
    
    def handle_request(self):
        global call_count
        call_count += 1
        
        if call_count <= fail_count:
            self.send_response(500)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"error": "Internal Server Error"}')
        else:
            self.send_response(response_code)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status": "ok"}')
    
    def log_message(self, format, *args):
        pass

with socketserver.TCPServer(("", ${port}), MockHandler) as httpd:
    httpd.serve_forever()
EOF
    
    python3 /tmp/mock_server_${port}.py > /dev/null 2>&1 &
    MOCK_SERVER_PID=$!
    
    # Wait for server to start
    sleep 1
    
    if ! kill -0 $MOCK_SERVER_PID 2>/dev/null; then
        echo "Failed to start mock server on port $port"
        return 1
    fi
    
    echo "Mock server started on port $port (PID: $MOCK_SERVER_PID)"
    return 0
}

stop_mock_server() {
    if [ -n "$MOCK_SERVER_PID" ] && kill -0 $MOCK_SERVER_PID 2>/dev/null; then
        kill $MOCK_SERVER_PID 2>/dev/null
        wait $MOCK_SERVER_PID 2>/dev/null
        MOCK_SERVER_PID=""
    fi
    rm -f /tmp/mock_server_${MOCK_SERVER_PORT}.py
}

get_mock_server_url() {
    echo "http://localhost:${MOCK_SERVER_PORT}"
}

cleanup_mock_server() {
    stop_mock_server
}

# Start mock server with task endpoint support
start_mock_task_server() {
    local port=$1
    local task_response_file=${2:-""}
    local verification_response_file=${3:-""}
    local list_response_file=${4:-""}
    local dependency_response_file=${5:-""}
    local response_code=${6:-200}
    local endpoint_error=${7:-""}  # Format: "endpoint:code" e.g., "tasks/123:404"
    
    MOCK_SERVER_PORT=$port
    MOCK_SERVER_RESPONSE_CODE=$response_code
    
    # Read fixture files if provided
    local task_response="{}"
    local verification_response="{}"
    local list_response="{}"
    local dependency_response="{}"
    
    if [ -n "$task_response_file" ] && [ -f "$task_response_file" ]; then
        task_response=$(cat "$task_response_file")
    fi
    
    if [ -n "$verification_response_file" ] && [ -f "$verification_response_file" ]; then
        verification_response=$(cat "$verification_response_file")
    fi
    
    if [ -n "$list_response_file" ] && [ -f "$list_response_file" ]; then
        list_response=$(cat "$list_response_file")
    fi
    
    if [ -n "$dependency_response_file" ] && [ -f "$dependency_response_file" ]; then
        dependency_response=$(cat "$dependency_response_file")
    fi
    
    # Escape JSON for Python
    task_response=$(echo "$task_response" | sed 's/"/\\"/g')
    verification_response=$(echo "$verification_response" | sed 's/"/\\"/g')
    list_response=$(echo "$list_response" | sed 's/"/\\"/g')
    dependency_response=$(echo "$dependency_response" | sed 's/"/\\"/g')
    
    cat > /tmp/mock_task_server_${port}.py <<EOF
import http.server
import socketserver
import json
import urllib.parse

task_response = """$task_response"""
verification_response = """$verification_response"""
list_response = """$list_response"""
dependency_response = """$dependency_response"""
endpoint_error = """$endpoint_error"""
default_response_code = $response_code

class TaskMockHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.handle_request()
    
    def do_POST(self):
        self.handle_request()
    
    def handle_request(self):
        path = urllib.parse.urlparse(self.path).path
        method = self.command
        
        # Check for endpoint-specific errors
        error_code = None
        if endpoint_error:
            parts = endpoint_error.split(":")
            if len(parts) == 2 and parts[0] in path:
                error_code = int(parts[1])
        
        if error_code:
            self.send_response(error_code)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"error": "Requested error"}')
            return
        
        # Route to appropriate response
        if path.startswith('/api/v1/tasks/') and path.endswith('/dependencies'):
            response_data = dependency_response
        elif path.startswith('/api/v1/tasks/') and method == 'GET':
            response_data = task_response
        elif path.startswith('/api/v1/tasks') and method == 'GET':
            response_data = list_response
        else:
            response_data = '{"status": "ok"}'
        
        self.send_response(default_response_code)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(response_data.encode('utf-8'))
    
    def log_message(self, format, *args):
        pass

with socketserver.TCPServer(("", ${port}), TaskMockHandler) as httpd:
    httpd.serve_forever()
EOF
    
    python3 /tmp/mock_task_server_${port}.py > /dev/null 2>&1 &
    MOCK_SERVER_PID=$!
    
    # Wait for server to start
    sleep 1
    
    if ! kill -0 $MOCK_SERVER_PID 2>/dev/null; then
        echo "Failed to start mock task server on port $port"
        return 1
    fi
    
    echo "Mock task server started on port $port (PID: $MOCK_SERVER_PID)"
    return 0
}





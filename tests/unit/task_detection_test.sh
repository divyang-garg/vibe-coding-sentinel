#!/bin/bash
# Phase 14E: Task Detection Unit Tests

set -e

TEST_DIR=$(dirname "$0")
FIXTURES_DIR="$TEST_DIR/../fixtures/tasks"

echo "🧪 Testing Task Detection..."

# Test 1: TODO comment detection
echo "Test 1: TODO comment detection"
cat > "$FIXTURES_DIR/test_todo.js" << 'EOF'
// TODO: Implement user authentication
function login() {
  // FIXME: Add error handling
  console.log("login");
}
EOF

# Test 2: Cursor task marker detection
echo "Test 2: Cursor task marker detection"
cat > "$FIXTURES_DIR/test_cursor.md" << 'EOF'
- [ ] Task: Add payment processing
- [x] Task: Add user authentication (completed)
EOF

# Test 3: Explicit task format
echo "Test 3: Explicit task format"
cat > "$FIXTURES_DIR/test_explicit.go" << 'EOF'
// TASK: TASK-123 - Add order cancellation
// DEPENDS: TASK-122, TASK-121
func cancelOrder() {
  // Implementation
}
EOF

# Test 4: Priority detection
echo "Test 4: Priority detection"
cat > "$FIXTURES_DIR/test_priority.py" << 'EOF'
# TODO: CRITICAL - Fix security vulnerability
# TODO: HIGH - Add rate limiting
# TODO: LOW - Update documentation
EOF

# Test 5: Tag extraction
echo "Test 5: Tag extraction"
cat > "$FIXTURES_DIR/test_tags.js" << 'EOF'
// TODO: #auth #security Implement JWT middleware
// FIXME: #bug #critical Fix memory leak
EOF

echo "✅ Task detection test fixtures created"
echo "Run: ./sentinel tasks scan --dir $FIXTURES_DIR"










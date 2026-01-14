#!/bin/bash
# Performance Benchmarking Suite
# Tests system performance under various conditions

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

SENTINEL="./sentinel"

# Create test data
setup_test_data() {
    local test_dir="$1"
    local file_count="$2"

    mkdir -p "$test_dir/src" "$test_dir/tests"

    # Create package.json
    cat > "$test_dir/package.json" << 'EOF'
{
  "name": "performance-test",
  "version": "1.0.0",
  "dependencies": {"react": "^18.0.0", "lodash": "^4.17.0"}
}
EOF

    # Create test files
    for i in $(seq 1 "$file_count"); do
        cat > "$test_dir/src/component_$i.js" << EOF
import React, { useState } from 'react';
import _ from 'lodash';

function Component$i({ data }) {
    const [count, setCount] = useState(0);
    console.log("Component $i rendered with data:", data);

    const handleClick = _.debounce(() => {
        setCount(c => c + 1);
    }, 300);

    return (
        <div>
            <h1>Component $i</h1>
            <p>Count: {count}</p>
            <button onClick={handleClick}>Increment</button>
            <p>Data length: {data?.length || 0}</p>
        </div>
    );
}

export default Component$i;
EOF

        cat > "$test_dir/tests/component_$i.test.js" << EOF
import React from 'react';
import { render } from '@testing-library/react';
import Component$i from '../src/component_$i';

test('Component$i renders correctly', () => {
    const { getByText } = render(<Component$i data={[1,2,3]} />);
    expect(getByText('Component $i')).toBeInTheDocument();
});
EOF
    done
}

# Time execution and check performance
benchmark_command() {
    local cmd="$1"
    local expected_max_time="$2"
    local description="$3"

    log_info "Benchmarking: $description"

    start_time=$(date +%s.%3N)
    if eval "$cmd" >/dev/null 2>&1; then
        end_time=$(date +%s.%3N)
        execution_time=$(echo "$end_time - $start_time" | bc)

        if (( $(echo "$execution_time <= $expected_max_time" | bc -l) )); then
            log_success "$description completed in ${execution_time}s (≤ ${expected_max_time}s)"
            return 0
        else
            log_warning "$description took ${execution_time}s (> ${expected_max_time}s expected)"
            return 1
        fi
    else
        log_error "$description failed to execute"
        return 2
    fi
}

echo "⚡ Performance Benchmarking Suite"
echo "================================"

# Test 1: Small project performance
echo ""
log_info "Test 1: Small Project Performance (10 files)"
TEST_DIR_1="/tmp/perf_test_small_$(date +%s)"
setup_test_data "$TEST_DIR_1" 10
cd "$TEST_DIR_1"

benchmark_command "$SENTINEL audit --offline" 5 "Small project audit" || true
benchmark_command "$SENTINEL learn" 3 "Small project pattern learning" || true
benchmark_command "$SENTINEL fix --safe" 3 "Small project auto-fix" || true

cd /
rm -rf "$TEST_DIR_1"

# Test 2: Medium project performance
echo ""
log_info "Test 2: Medium Project Performance (50 files)"
TEST_DIR_2="/tmp/perf_test_medium_$(date +%s)"
setup_test_data "$TEST_DIR_2" 50
cd "$TEST_DIR_2"

benchmark_command "$SENTINEL audit --offline" 15 "Medium project audit" || true
benchmark_command "$SENTINEL learn" 8 "Medium project pattern learning" || true

cd /
rm -rf "$TEST_DIR_2"

# Test 3: Memory usage check
echo ""
log_info "Test 3: Memory Usage Analysis"
TEST_DIR_3="/tmp/perf_test_memory_$(date +%s)"
setup_test_data "$TEST_DIR_3" 25
cd "$TEST_DIR_3"

log_info "Running memory-intensive audit..."
/usr/bin/time -v $SENTINEL audit --offline >/dev/null 2>&1 || true

cd /
rm -rf "$TEST_DIR_3"

# Test 4: Concurrent operations
echo ""
log_info "Test 4: Concurrent Operations Test"
TEST_DIR_4="/tmp/perf_test_concurrent_$(date +%s)"
setup_test_data "$TEST_DIR_4" 20
cd "$TEST_DIR_4"

log_info "Testing concurrent audit operations..."
start_time=$(date +%s.%3N)

# Run multiple audits concurrently
$SENTINEL audit --offline >/dev/null 2>&1 &
PID1=$!
$SENTINEL audit --offline >/dev/null 2>&1 &
PID2=$!

wait $PID1 $PID2
end_time=$(date +%s.%3N)
concurrent_time=$(echo "$end_time - $start_time" | bc)

if (( $(echo "$concurrent_time < 10" | bc -l) )); then
    log_success "Concurrent operations completed in ${concurrent_time}s"
else
    log_warning "Concurrent operations took ${concurrent_time}s (may indicate issues)"
fi

cd /
rm -rf "$TEST_DIR_4"

# Test 5: Startup time
echo ""
log_info "Test 5: Application Startup Performance"
start_time=$(date +%s.%3N)
$SENTINEL --help >/dev/null 2>&1
end_time=$(date +%s.%3N)
startup_time=$(echo "$end_time - $start_time" | bc)

if (( $(echo "$startup_time < 0.5" | bc -l) )); then
    log_success "Application startup: ${startup_time}s (< 0.5s)"
else
    log_warning "Application startup: ${startup_time}s (≥ 0.5s)"
fi

# Test 6: Large file handling
echo ""
log_info "Test 6: Large File Handling"
TEST_DIR_6="/tmp/perf_test_large_$(date +%s)"
mkdir -p "$TEST_DIR_6"
cd "$TEST_DIR_6"

# Create a large file (1MB)
dd if=/dev/zero of=large_file.js bs=1024 count=1024 >/dev/null 2>&1
echo 'console.log("large file");' >> large_file.js

benchmark_command "$SENTINEL audit --offline" 10 "Large file audit" || true

cd /
rm -rf "$TEST_DIR_6"

echo ""
echo "📊 Performance Benchmarking Complete"
echo "===================================="
log_info "Performance tests help identify bottlenecks and optimization opportunities"
log_info "All benchmarks completed - review results above for any performance issues"

exit 0




#!/bin/bash
# Production Load Testing Suite
# Validates system performance under production-like conditions

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

# Performance thresholds (in seconds)
THRESHOLD_STARTUP=0.5
THRESHOLD_SMALL_AUDIT=5
THRESHOLD_MEDIUM_AUDIT=15
THRESHOLD_LARGE_AUDIT=30
THRESHOLD_PATTERN_LEARN=10
THRESHOLD_FIX=5

PASSED=0
FAILED=0

# Generate test data of various sizes
generate_test_data() {
    local dir="$1"
    local size="$2"

    mkdir -p "$dir/src" "$dir/tests"

    # Create package.json
    cat > "$dir/package.json" << 'EOF'
{
  "name": "load-test-project",
  "version": "1.0.0",
  "dependencies": {"react": "^18.0.0", "lodash": "^4.17.0", "express": "^4.18.0"}
}
EOF

    # Generate files based on size
    case "$size" in
        "small")
            file_count=5
            ;;
        "medium")
            file_count=25
            ;;
        "large")
            file_count=100
            ;;
        "xlarge")
            file_count=500
            ;;
    esac

    log_info "Generating $size test project with $file_count files..."

    for i in $(seq 1 "$file_count"); do
        # Create source file
        cat > "$dir/src/component_$i.js" << EOF
import React, { useState, useEffect } from 'react';
import _ from 'lodash';
import express from 'express';

const API_KEY = process.env.API_KEY || 'test-key-123';
const DB_PASSWORD = 'super_secret_password';

function Component$i({ data, onUpdate }) {
    const [state, setState] = useState({
        loading: false,
        error: null,
        items: []
    });

    useEffect(() => {
        console.log('Component $i mounted with data:', data);
        fetchData();
    }, [data]);

    const fetchData = async () => {
        try {
            setState(prev => ({ ...prev, loading: true }));

            const response = await fetch('/api/data', {
                headers: {
                    'Authorization': \`Bearer \${API_KEY}\`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error('API request failed');
            }

            const result = await response.json();
            setState(prev => ({
                ...prev,
                loading: false,
                items: result.data
            }));

            onUpdate && onUpdate(result);
        } catch (error) {
            console.error('Error fetching data:', error);
            setState(prev => ({
                ...prev,
                loading: false,
                error: error.message
            }));
        }
    };

    const handleSearch = _.debounce((query) => {
        const filtered = state.items.filter(item =>
            item.name.toLowerCase().includes(query.toLowerCase())
        );
        setState(prev => ({ ...prev, filteredItems: filtered }));
    }, 300);

    const dangerousQuery = "SELECT * FROM users WHERE id = " + data.userId;
    const safeQuery = "SELECT * FROM users WHERE id = ?";

    return (
        <div className="component-$i">
            <h2>Component $i</h2>
            {state.loading && <p>Loading...</p>}
            {state.error && <p>Error: {state.error}</p>}

            <input
                type="text"
                placeholder="Search..."
                onChange={(e) => handleSearch(e.target.value)}
            />

            <ul>
                {(state.filteredItems || state.items).map((item, index) => (
                    <li key={index}>
                        {item.name} - {item.value}
                    </li>
                ))}
            </ul>

            <button onClick={() => setState(prev => ({ ...prev, items: [] }))}>
                Clear Data
            </button>
        </div>
    );
}

export default Component$i;
EOF

        # Create test file
        cat > "$dir/tests/component_$i.test.js" << EOF
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Component$i from '../src/component_$i';

global.fetch = jest.fn();

describe('Component$i', () => {
    beforeEach(() => {
        fetch.mockClear();
    });

    test('renders component correctly', () => {
        render(<Component$i data={{ userId: 123 }} />);
        expect(screen.getByText('Component $i')).toBeInTheDocument();
    });

    test('loads data on mount', async () => {
        const mockData = { data: [{ name: 'Test', value: '123' }] };
        fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockData)
        });

        render(<Component$i data={{ userId: 123 }} />);

        await waitFor(() => {
            expect(screen.getByText('Test - 123')).toBeInTheDocument();
        });
    });

    test('handles API errors', async () => {
        fetch.mockRejectedValueOnce(new Error('Network error'));

        render(<Component$i data={{ userId: 123 }} />);

        await waitFor(() => {
            expect(screen.getByText(/Error/)).toBeInTheDocument();
        });
    });

    test('search functionality works', async () => {
        const mockData = {
            data: [
                { name: 'Apple', value: 'fruit' },
                { name: 'Car', value: 'vehicle' }
            ]
        };
        fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockData)
        });

        const user = userEvent.setup();
        render(<Component$i data={{ userId: 123 }} />);

        await waitFor(() => {
            expect(screen.getByText('Apple - fruit')).toBeInTheDocument();
        });

        const searchInput = screen.getByPlaceholderText('Search...');
        await user.type(searchInput, 'car');

        await waitFor(() => {
            expect(screen.getByText('Car - vehicle')).toBeInTheDocument();
            expect(screen.queryByText('Apple - fruit')).not.toBeInTheDocument();
        });
    });
});
EOF
    done

    log_success "Generated $file_count files in $dir"
}

# Run performance test
run_performance_test() {
    local test_name="$1"
    local setup_cmd="$2"
    local test_cmd="$3"
    local threshold="$4"
    local description="$5"

    log_info "Running $test_name..."

    # Setup
    eval "$setup_cmd"

    # Time the operation
    start_time=$(date +%s.%3N)
    if eval "$test_cmd" >/dev/null 2>&1; then
        end_time=$(date +%s.%3N)
        duration=$(echo "$end_time - $start_time" | bc 2>/dev/null || echo "0")

        if (( $(echo "$duration <= $threshold" | bc -l 2>/dev/null || echo "1") )); then
            log_success "$description completed in ${duration}s (≤ ${threshold}s)"
            ((PASSED++))
        else
            log_warning "$description took ${duration}s (> ${threshold}s threshold)"
            ((FAILED++))
        fi
    else
        log_error "$description failed to execute"
        ((FAILED++))
    fi
}

echo "🔥 PRODUCTION LOAD TESTING SUITE"
echo "==============================="

# Test 1: Startup Performance
echo ""
log_info "PHASE 1: Startup Performance Tests"

run_performance_test \
    "Cold Start" \
    "true" \
    "$SENTINEL --help >/dev/null 2>&1" \
    "$THRESHOLD_STARTUP" \
    "Application cold start"

run_performance_test \
    "Warm Start" \
    "true" \
    "$SENTINEL --help >/dev/null 2>&1 && $SENTINEL --help >/dev/null 2>&1" \
    "$THRESHOLD_STARTUP" \
    "Application warm start"

# Test 2: Small Project Performance
echo ""
log_info "PHASE 2: Small Project Performance"

TEST_DIR_SMALL="/tmp/load_test_small_$(date +%s)"
run_performance_test \
    "Small Project Audit" \
    "generate_test_data '$TEST_DIR_SMALL' 'small' && cd '$TEST_DIR_SMALL'" \
    "$SENTINEL audit --offline" \
    "$THRESHOLD_SMALL_AUDIT" \
    "Small project security audit"

run_performance_test \
    "Small Project Pattern Learning" \
    "cd '$TEST_DIR_SMALL'" \
    "$SENTINEL learn" \
    "$THRESHOLD_PATTERN_LEARN" \
    "Small project pattern learning"

run_performance_test \
    "Small Project Auto-Fix" \
    "cd '$TEST_DIR_SMALL'" \
    "$SENTINEL fix --safe" \
    "$THRESHOLD_FIX" \
    "Small project auto-fix"

rm -rf "$TEST_DIR_SMALL"

# Test 3: Medium Project Performance
echo ""
log_info "PHASE 3: Medium Project Performance"

TEST_DIR_MEDIUM="/tmp/load_test_medium_$(date +%s)"
run_performance_test \
    "Medium Project Audit" \
    "generate_test_data '$TEST_DIR_MEDIUM' 'medium' && cd '$TEST_DIR_MEDIUM'" \
    "$SENTINEL audit --offline" \
    "$THRESHOLD_MEDIUM_AUDIT" \
    "Medium project security audit"

rm -rf "$TEST_DIR_MEDIUM"

# Test 4: Large Project Performance
echo ""
log_info "PHASE 4: Large Project Performance"

TEST_DIR_LARGE="/tmp/load_test_large_$(date +%s)"
run_performance_test \
    "Large Project Audit" \
    "generate_test_data '$TEST_DIR_LARGE' 'large' && cd '$TEST_DIR_LARGE'" \
    "$SENTINEL audit --offline" \
    "$THRESHOLD_LARGE_AUDIT" \
    "Large project security audit"

rm -rf "$TEST_DIR_LARGE"

# Test 5: Memory and Resource Usage
echo ""
log_info "PHASE 5: Resource Usage Analysis"

TEST_DIR_RESOURCE="/tmp/load_test_resource_$(date +%s)"
generate_test_data "$TEST_DIR_RESOURCE" "medium"
cd "$TEST_DIR_RESOURCE"

log_info "Testing memory usage during audit..."
if command -v /usr/bin/time >/dev/null 2>&1; then
    AUDIT_OUTPUT=$(/usr/bin/time -v $SENTINEL audit --offline 2>&1)
    MAX_MEMORY=$(echo "$AUDIT_OUTPUT" | grep "Maximum resident set size" | awk '{print $NF}' || echo "unknown")
    if [[ "$MAX_MEMORY" != "unknown" && "$MAX_MEMORY" -lt 500000 ]]; then  # Less than 500MB
        log_success "Memory usage: ${MAX_MEMORY} KB (< 500MB limit)"
        ((PASSED++))
    else
        log_warning "Memory usage: ${MAX_MEMORY} KB (≥ 500MB)"
        ((FAILED++))
    fi
else
    log_warning "Memory testing not available (missing /usr/bin/time)"
fi

rm -rf "$TEST_DIR_RESOURCE"

# Test 6: Concurrent Operations
echo ""
log_info "PHASE 6: Concurrent Load Testing"

TEST_DIR_CONCURRENT="/tmp/load_test_concurrent_$(date +%s)"
generate_test_data "$TEST_DIR_CONCURRENT" "small"
cd "$TEST_DIR_CONCURRENT"

log_info "Testing concurrent audit operations..."
start_time=$(date +%s.%3N)

# Run multiple audits concurrently
$SENTINEL audit --offline >/dev/null 2>&1 &
PID1=$!
$SENTINEL audit --offline >/dev/null 2>&1 &
PID2=$!
$SENTINEL audit --offline >/dev/null 2>&1 &
PID3=$!

# Wait for all to complete
wait $PID1 $PID2 $PID3 2>/dev/null || true
end_time=$(date +%s.%3N)
concurrent_time=$(echo "$end_time - $start_time" | bc 2>/dev/null || echo "0")

if (( $(echo "$concurrent_time < 15" | bc -l 2>/dev/null || echo "1") )); then
    log_success "Concurrent operations completed in ${concurrent_time}s"
    ((PASSED++))
else
    log_warning "Concurrent operations took ${concurrent_time}s (may indicate issues)"
    ((FAILED++))
fi

rm -rf "$TEST_DIR_CONCURRENT"

# Test 7: Stability Under Load
echo ""
log_info "PHASE 7: Stability Testing"

TEST_DIR_STABILITY="/tmp/load_test_stability_$(date +%s)"
generate_test_data "$TEST_DIR_STABILITY" "small"
cd "$TEST_DIR_STABILITY"

log_info "Running repeated operations for stability..."
stable_operations=0
for i in {1..5}; do
    if $SENTINEL audit --offline >/dev/null 2>&1; then
        ((stable_operations++))
    fi
done

if [[ $stable_operations -eq 5 ]]; then
    log_success "Stability test: $stable_operations/5 operations successful"
    ((PASSED++))
else
    log_error "Stability test: $stable_operations/5 operations successful"
    ((FAILED++))
fi

rm -rf "$TEST_DIR_STABILITY"

# Final Results
echo ""
echo "📊 LOAD TESTING RESULTS"
echo "======================"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
TOTAL=$((PASSED + FAILED))

if [[ $TOTAL -gt 0 ]]; then
    SUCCESS_RATE=$((PASSED * 100 / TOTAL))

    if [[ $SUCCESS_RATE -ge 85 ]]; then
        echo -e "${GREEN}🎉 LOAD TEST SUCCESS: ${SUCCESS_RATE}%${NC}"
        echo "System performance meets production requirements"
        exit 0
    else
        echo -e "${RED}❌ LOAD TEST ISSUES: ${SUCCESS_RATE}%${NC}"
        echo "Performance optimization may be needed"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  No tests were executed${NC}"
    exit 1
fi




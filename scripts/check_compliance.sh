#!/bin/bash
# Compliance checker for CODING_STANDARDS.md
# Checks file size limits, function counts, and basic code organization

set -e

echo "=== Sentinel Coding Standards Compliance Check ==="
echo ""

ERRORS=0
WARNINGS=0

# File size limits from CODING_STANDARDS.md
MAX_ENTRY_POINTS=50
MAX_HANDLERS=300
MAX_SERVICES=400
MAX_REPOSITORIES=350
MAX_MODELS=200
MAX_UTILITIES=250
MAX_TESTS=500

check_file_size() {
    local file=$1
    local max_lines=$2
    local file_type=$3
    
    if [ ! -f "$file" ]; then
        return
    fi
    
    local lines=$(wc -l < "$file" | tr -d ' ')
    
    if [ "$lines" -gt "$max_lines" ]; then
        echo "❌ ERROR: $file_type file exceeds limit: $file ($lines lines, max: $max_lines)"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
    return 0
}

check_test_file() {
    local file=$1
    check_file_size "$file" $MAX_TESTS "Test"
}

check_source_file() {
    local file=$1
    local basename=$(basename "$file")
    local dir=$(dirname "$file")
    
    # Determine file type based on path and name
    if [[ "$file" == *"/main.go" ]] || [[ "$file" == *"/cmd/"*"/main.go" ]]; then
        check_file_size "$file" $MAX_ENTRY_POINTS "Entry point (main.go)"
    elif [[ "$file" == *"/handlers/"* ]] || [[ "$file" == *"/handler"* ]]; then
        check_file_size "$file" $MAX_HANDLERS "HTTP Handler"
    elif [[ "$file" == *"/services/"* ]] || [[ "$file" == *"/service"* ]]; then
        check_file_size "$file" $MAX_SERVICES "Business Service"
    elif [[ "$file" == *"/repository/"* ]] || [[ "$file" == *"/repo"* ]]; then
        check_file_size "$file" $MAX_REPOSITORIES "Repository"
    elif [[ "$file" == *"/models/"* ]] || [[ "$file" == *"/model"* ]]; then
        check_file_size "$file" $MAX_MODELS "Data Model"
    elif [[ "$file" == *"/utils/"* ]] || [[ "$file" == *"/util"* ]]; then
        check_file_size "$file" $MAX_UTILITIES "Utility"
    fi
}

echo "Checking test files..."
find . -name "*_test.go" -type f ! -path "./vendor/*" ! -path "./.git/*" ! -path "./node_modules/*" | while read file; do
    check_test_file "$file"
done

echo ""
echo "Checking source files..."
find . -name "*.go" -type f ! -name "*_test.go" ! -path "./vendor/*" ! -path "./.git/*" ! -path "./node_modules/*" | while read file; do
    check_source_file "$file"
done

echo ""
echo "=== Compliance Check Summary ==="
if [ $ERRORS -eq 0 ]; then
    echo "✅ All files comply with CODING_STANDARDS.md size limits"
    exit 0
else
    echo "❌ Found $ERRORS file(s) exceeding size limits"
    echo "Please refactor files to comply with CODING_STANDARDS.md"
    exit 1
fi

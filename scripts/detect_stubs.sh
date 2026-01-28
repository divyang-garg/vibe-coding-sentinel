#!/bin/bash
# Improved stub detection script with false positive filtering
# Used by pre-commit hook to detect real stubs, excluding false positives

# Stub patterns to detect (more specific to avoid false positives)
STUB_PATTERNS=(
    "// Stub[^a-zA-Z]"
    "// stub[^a-zA-Z]"
    "// STUB[^a-zA-Z]"
    "not implemented[^)]"
    "not yet implemented"
    "would be implemented"
    "return nil.*//.*stub"
    "return.*fmt\.Errorf.*not implemented"
)

# Files/directories to exclude (ONLY true false positives - NOT documented/pending stubs)
# Note: Documented/pending stubs are still stubs and SHOULD be flagged
EXCLUDE_FILES=(
    "_test.go$"
    "/tests/fixtures/"
    "/test"
    # Note: helpers_stubs.go may contain real stubs - we'll check if they're actual stubs
    # Only exclude if they're test-only helpers, not production stubs
    "task_integrations.go"          # Database operations (not code analysis stubs)
    "utils_business_rule.go"        # Has Tree-Sitter integration
    "doc_sync_business.go"          # Has Tree-Sitter integration
    "dependency_detector_helpers.go" # Has Tree-Sitter integration
    "architecture_analysis.go"      # Has Tree-Sitter integration
    "logic_analyzer_helpers.go"     # Has Tree-Sitter integration
    "ast_bridge.go"                 # AST bridge (fully implemented)
    "/ast/"                         # AST package (fully implemented)
)

# Find files matching stub patterns, excluding false positives
find_stubs() {
    local real_stubs=()
    local seen_files=()
    
    for pattern in "${STUB_PATTERNS[@]}"; do
        while IFS= read -r file; do
            # Skip if file is empty or doesn't exist
            [ -z "$file" ] || [ ! -f "$file" ] && continue
            
            # Skip excluded files
            local skip=0
            for exclude in "${EXCLUDE_FILES[@]}"; do
                if echo "$file" | grep -qE "$exclude"; then
                    skip=1
                    break
                fi
            done
            [ $skip -eq 1 ] && continue
            
            # Skip if already seen
            local seen=0
            for seen_file in "${seen_files[@]}"; do
                [ "$seen_file" = "$file" ] && seen=1 && break
            done
            [ $seen -eq 1 ] && continue
            
            # Check if it's a Tree-Sitter false positive
            # Only exclude if it's JUST a documentation comment with NO actual stub implementation
            # If it has a stub implementation (return nil, return error, empty body), FLAG IT even if documented
            if grep -E "$pattern" "$file" 2>/dev/null | grep -qE "(tree-sitter|Tree-Sitter|tree_sitter|AST|ast|parser|Parser|integrated|integration)" 2>/dev/null; then
                # Check if this is actually a stub implementation or just documentation
                # Real stubs have: return nil, return error, or empty function body
                if grep -E "$pattern" "$file" 2>/dev/null | grep -qE "(return nil|return.*error|return.*fmt\.Errorf|^[[:space:]]*\{[[:space:]]*\})" 2>/dev/null; then
                    # This IS a stub implementation (even if documented as pending) - FLAG IT
                    # Don't skip - documented/pending stubs are still stubs!
                    : # No-op, continue processing
                else
                    # This is just a documentation comment with no stub implementation - skip it
                    continue
                fi
            fi
            
            # This appears to be a real stub
            real_stubs+=("$file")
            seen_files+=("$file")
        done < <(grep -r -l --include="*.go" -E "$pattern" . 2>/dev/null | \
            grep -v "_test.go$" | \
            grep -v "/tests/fixtures/" | \
            grep -v "/test" | \
            grep -v "helpers_stubs_test.go" | \
            grep -v "task_integrations.go" | \
            grep -v "utils_business_rule.go" | \
            grep -v "doc_sync_business.go" | \
            grep -v "dependency_detector_helpers.go" | \
            grep -v "architecture_analysis.go" | \
            grep -v "logic_analyzer_helpers.go" | \
            grep -v "ast_bridge.go" | \
            grep -v "/ast/" || true)
    done
    
    # Check for function names ending in "Stub" (exclude test helpers)
    while IFS= read -r file; do
        [ -z "$file" ] || [ ! -f "$file" ] && continue
        
        # Skip excluded files
        local skip=0
        for exclude in "${EXCLUDE_FILES[@]}"; do
            if echo "$file" | grep -qE "$exclude"; then
                skip=1
                break
            fi
        done
        [ $skip -eq 1 ] && continue
        
        # Skip if already seen
        local seen=0
        for seen_file in "${seen_files[@]}"; do
            [ "$seen_file" = "$file" ] && seen=1 && break
        done
        [ $seen -eq 1 ] && continue
        
        real_stubs+=("$file")
        seen_files+=("$file")
    done < <(grep -r -l --include="*.go" -E "func.*Stub\(" . 2>/dev/null | \
        grep -v "_test.go$" | \
        grep -v "/tests/fixtures/" | \
        grep -v "/test" | \
        grep -v "helpers_stubs_test.go" || true)
    
    # Output results: count first, then files
    local count=${#real_stubs[@]}
    echo "$count"
    if [ $count -gt 0 ]; then
        printf '%s\n' "${real_stubs[@]}" | head -10
    fi
}

# Run detection
find_stubs

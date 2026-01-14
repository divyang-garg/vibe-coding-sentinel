#!/bin/bash
# Utility functions for shell scripts
# Follows shell script best practices

set -euo pipefail

# =============================================================================
# Color Codes
# =============================================================================

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# =============================================================================
# Logging
# =============================================================================

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# =============================================================================
# Validation Functions
# =============================================================================

is_command_available() {
    local cmd="$1"
    command -v "$cmd" &> /dev/null
}

validate_file_exists() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        print_error "File not found: $file"
        return 1
    fi
    return 0
}

validate_directory_exists() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        print_error "Directory not found: $dir"
        return 1
    fi
    return 0
}

# =============================================================================
# File Operations
# =============================================================================

safe_copy() {
    local source="$1"
    local dest="$2"
    
    validate_file_exists "$source" || return 1
    
    # Create destination directory if needed
    mkdir -p "$(dirname "$dest")"
    
    cp "$source" "$dest"
    print_success "Copied $source to $dest"
}

safe_move() {
    local source="$1"
    local dest="$2"
    
    validate_file_exists "$source" || return 1
    
    # Create destination directory if needed
    mkdir -p "$(dirname "$dest")"
    
    mv "$source" "$dest"
    print_success "Moved $source to $dest"
}

create_temp_file() {
    local prefix="${1:-tmp}"
    mktemp "/tmp/${prefix}.XXXXXX"
}

create_temp_dir() {
    local prefix="${1:-tmpdir}"
    mktemp -d "/tmp/${prefix}.XXXXXX"
}

# =============================================================================
# String Functions
# =============================================================================

trim() {
    local str="$1"
    # Remove leading whitespace
    str="${str#"${str%%[![:space:]]*}"}"
    # Remove trailing whitespace
    str="${str%"${str##*[![:space:]]}"}"
    echo "$str"
}

to_lowercase() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}

to_uppercase() {
    echo "$1" | tr '[:lower:]' '[:upper:]'
}

# =============================================================================
# Array Functions
# =============================================================================

array_contains() {
    local needle="$1"
    shift
    local arr=("$@")
    
    for item in "${arr[@]}"; do
        if [[ "$item" == "$needle" ]]; then
            return 0
        fi
    done
    return 1
}

# =============================================================================
# Date Functions
# =============================================================================

get_timestamp() {
    date '+%Y%m%d_%H%M%S'
}

get_date() {
    date '+%Y-%m-%d'
}

get_datetime() {
    date '+%Y-%m-%d %H:%M:%S'
}













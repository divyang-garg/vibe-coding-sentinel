#!/bin/bash
# Sample deployment script following shell best practices
# Used for pattern detection and security tests

set -e          # Exit on error
set -u          # Exit on undefined variable
set -o pipefail # Exit on pipe failure

# =============================================================================
# Configuration
# =============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly LOG_FILE="${PROJECT_ROOT}/logs/deploy.log"
readonly BACKUP_DIR="${PROJECT_ROOT}/backups"
readonly TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

# Default values
ENVIRONMENT="${ENVIRONMENT:-staging}"
DRY_RUN="${DRY_RUN:-false}"
VERBOSE="${VERBOSE:-false}"

# =============================================================================
# Logging Functions
# =============================================================================

log_info() {
    local message="$1"
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $message" | tee -a "$LOG_FILE"
}

log_error() {
    local message="$1"
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $message" | tee -a "$LOG_FILE" >&2
}

log_debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        local message="$1"
        echo "[DEBUG] $(date '+%Y-%m-%d %H:%M:%S') - $message" | tee -a "$LOG_FILE"
    fi
}

# =============================================================================
# Utility Functions
# =============================================================================

cleanup() {
    local exit_code=$?
    log_info "Cleaning up temporary files..."
    
    # Remove temporary files safely
    if [[ -n "${TEMP_DIR:-}" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
    
    if [[ $exit_code -ne 0 ]]; then
        log_error "Deployment failed with exit code $exit_code"
    fi
    
    exit $exit_code
}

trap cleanup EXIT

create_backup() {
    local source_dir="$1"
    local backup_name="${2:-backup_${TIMESTAMP}}"
    
    if [[ ! -d "$source_dir" ]]; then
        log_error "Source directory does not exist: $source_dir"
        return 1
    fi
    
    mkdir -p "$BACKUP_DIR"
    
    local backup_path="${BACKUP_DIR}/${backup_name}.tar.gz"
    log_info "Creating backup: $backup_path"
    
    tar -czf "$backup_path" -C "$(dirname "$source_dir")" "$(basename "$source_dir")"
    
    log_info "Backup created successfully"
    echo "$backup_path"
}

validate_environment() {
    local env="$1"
    
    case "$env" in
        development|staging|production)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env"
            log_error "Valid environments: development, staging, production"
            return 1
            ;;
    esac
}

# =============================================================================
# Deployment Functions
# =============================================================================

pre_deploy_checks() {
    log_info "Running pre-deployment checks..."
    
    # Check required commands
    local required_commands=(git docker docker-compose)
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "Required command not found: $cmd"
            return 1
        fi
    done
    
    # Check git status
    if [[ -n "$(git status --porcelain)" ]]; then
        log_error "Working directory is not clean. Commit or stash changes first."
        return 1
    fi
    
    log_info "Pre-deployment checks passed"
    return 0
}

deploy() {
    local environment="$1"
    
    log_info "Starting deployment to $environment..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would deploy to $environment"
        return 0
    fi
    
    # Create backup before deployment
    create_backup "${PROJECT_ROOT}/app" "pre_deploy_${environment}_${TIMESTAMP}"
    
    # Pull latest changes
    log_info "Pulling latest changes..."
    git pull origin main
    
    # Build and deploy
    log_info "Building application..."
    docker-compose -f "docker-compose.${environment}.yml" build
    
    log_info "Deploying application..."
    docker-compose -f "docker-compose.${environment}.yml" up -d
    
    # Health check
    log_info "Running health check..."
    sleep 10
    
    if curl -sf "http://localhost:8080/health" > /dev/null; then
        log_info "Health check passed"
    else
        log_error "Health check failed"
        return 1
    fi
    
    log_info "Deployment to $environment completed successfully"
    return 0
}

# =============================================================================
# Main
# =============================================================================

usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

Deploy the application to specified environment.

Options:
    -e, --environment ENV   Target environment (default: staging)
    -d, --dry-run          Run without making changes
    -v, --verbose          Enable verbose output
    -h, --help             Show this help message

Environments:
    development    Local development
    staging        Staging environment
    production     Production environment

Examples:
    $(basename "$0") -e staging
    $(basename "$0") -e production --dry-run
    $(basename "$0") -v -e development

EOF
}

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -v|--verbose)
                VERBOSE="true"
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Validate environment
    validate_environment "$ENVIRONMENT" || exit 1
    
    # Create log directory
    mkdir -p "$(dirname "$LOG_FILE")"
    
    log_info "=== Deployment Started ==="
    log_info "Environment: $ENVIRONMENT"
    log_info "Dry Run: $DRY_RUN"
    
    # Run deployment
    pre_deploy_checks || exit 1
    deploy "$ENVIRONMENT" || exit 1
    
    log_info "=== Deployment Completed ==="
}

main "$@"













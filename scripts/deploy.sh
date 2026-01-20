#!/bin/bash
# Sentinel Hub API Deployment Script
# Complies with CODING_STANDARDS.md: Deployment automation standards

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
ENVIRONMENT=${1:-development}
PROJECT_NAME="sentinel-hub-api"
DOCKER_REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate environment
validate_environment() {
    case $ENVIRONMENT in
        development|staging|production)
            log_info "Deploying to $ENVIRONMENT environment"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT"
            log_info "Valid environments: development, staging, production"
            exit 1
            ;;
    esac
}

# Pre-deployment checks
pre_deployment_checks() {
    log_info "Running pre-deployment checks..."

    # Check if required tools are installed
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { log_error "Docker Compose is required but not installed"; exit 1; }

    # Check if required files exist
    [[ -f "docker-compose.yml" ]] || { log_error "docker-compose.yml not found"; exit 1; }
    [[ -f "docker-compose.prod.yml" ]] || { log_error "docker-compose.prod.yml not found"; exit 1; }
    [[ -f "Dockerfile" ]] || { log_error "Dockerfile not found"; exit 1; }

    # Run tests
    log_info "Running tests..."
    if go test ./... -short; then
        log_success "All tests passed"
    else
        log_error "Tests failed - aborting deployment"
        exit 1
    fi

    # Run quality checks
    if [[ -x ".githooks/pre-commit" ]]; then
        log_info "Running quality checks..."
        if ./.githooks/pre-commit >/dev/null 2>&1; then
            log_success "Quality checks passed"
        else
            log_error "Quality checks failed - aborting deployment"
            exit 1
        fi
    fi

    log_success "Pre-deployment checks completed"
}

# Build application
build_application() {
    log_info "Building application..."

    # Build Docker image
    local image_tag="${DOCKER_REGISTRY}/${PROJECT_NAME}:${ENVIRONMENT}-$(date +%Y%m%d-%H%M%S)"
    local latest_tag="${DOCKER_REGISTRY}/${PROJECT_NAME}:${ENVIRONMENT}-latest"

    docker build -t "$image_tag" -t "$latest_tag" .

    # Tag as latest for the environment
    docker tag "$image_tag" "${DOCKER_REGISTRY}/${PROJECT_NAME}:${ENVIRONMENT}"

    log_success "Application built successfully"
    echo "$image_tag" > .last_build_tag
}

# Deploy to environment
deploy_to_environment() {
    log_info "Deploying to $ENVIRONMENT environment..."

    case $ENVIRONMENT in
        development)
            # Development deployment
            docker-compose down || true
            docker-compose up -d --build
            ;;

        staging|production)
            # Production deployment
            local compose_file="docker-compose.prod.yml"
            local project_name="${PROJECT_NAME}-${ENVIRONMENT}"

            # Pull latest images
            docker-compose -f "$compose_file" -p "$project_name" pull || true

            # Deploy with zero-downtime
            docker-compose -f "$compose_file" -p "$project_name" up -d

            # Wait for health checks
            log_info "Waiting for services to be healthy..."
            sleep 30

            # Run post-deployment tests
            if [[ $ENVIRONMENT == "staging" ]]; then
                run_post_deployment_tests
            fi

            # Clean up old images (keep last 3)
            log_info "Cleaning up old Docker images..."
            docker image prune -f
            ;;
    esac

    log_success "Deployment to $ENVIRONMENT completed"
}

# Run post-deployment tests
run_post_deployment_tests() {
    log_info "Running post-deployment tests..."

    # Wait for services to be ready
    local max_attempts=30
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        if curl -f -s http://localhost:8080/health >/dev/null 2>&1; then
            log_success "API health check passed"
            break
        fi

        log_info "Waiting for API to be ready (attempt $attempt/$max_attempts)..."
        sleep 10
        ((attempt++))
    done

    if [[ $attempt -gt $max_attempts ]]; then
        log_error "API failed to become ready after deployment"
        exit 1
    fi

    # Run smoke tests
    log_info "Running smoke tests..."
    # Add your smoke tests here
    # Example: curl -f http://localhost:8080/api/v1/health

    log_success "Post-deployment tests passed"
}

# Rollback deployment
rollback_deployment() {
    log_error "Deployment failed - initiating rollback..."

    case $ENVIRONMENT in
        development)
            docker-compose down
            # Restore previous image if available
            if [[ -f ".last_successful_tag" ]]; then
                local previous_tag=$(cat .last_successful_tag)
                log_info "Rolling back to $previous_tag"
                docker tag "$previous_tag" "${DOCKER_REGISTRY}/${PROJECT_NAME}:${ENVIRONMENT}"
                docker-compose up -d
            fi
            ;;

        staging|production)
            local compose_file="docker-compose.prod.yml"
            local project_name="${PROJECT_NAME}-${ENVIRONMENT}"

            # Rollback to previous version
            if [[ -f ".last_successful_tag" ]]; then
                local previous_tag=$(cat .last_successful_tag)
                log_info "Rolling back to $previous_tag"
                docker tag "$previous_tag" "${DOCKER_REGISTRY}/${PROJECT_NAME}:${ENVIRONMENT}"
                docker-compose -f "$compose_file" -p "$project_name" up -d
            else
                log_warning "No previous version found for rollback"
            fi
            ;;
    esac
}

# Main deployment process
main() {
    log_info "🚀 Starting Sentinel Hub API deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Timestamp: $(date)"

    validate_environment

    if ! pre_deployment_checks; then
        log_error "Pre-deployment checks failed"
        exit 1
    fi

    if ! build_application; then
        log_error "Build failed"
        exit 1
    fi

    if deploy_to_environment; then
        log_success "🎉 Deployment completed successfully!"

        # Save successful deployment tag
        if [[ -f ".last_build_tag" ]]; then
            cp .last_build_tag .last_successful_tag
        fi

        # Print deployment information
        echo ""
        echo "Deployment Summary:"
        echo "==================="
        echo "Environment: $ENVIRONMENT"
        echo "Services:"
        case $ENVIRONMENT in
            development)
                echo "  - API: http://localhost:8080"
                echo "  - PostgreSQL: localhost:5432"
                echo "  - Redis: localhost:6379"
                echo "  - pgAdmin: http://localhost:5050"
                ;;
            staging|production)
                echo "  - API: https://your-domain.com"
                echo "  - Monitoring: https://your-domain.com:9090"
                ;;
        esac
    else
        log_error "Deployment failed"
        rollback_deployment
        exit 1
    fi
}

# Handle command line arguments
case "${2:-}" in
    --rollback)
        rollback_deployment
        exit 0
        ;;
    --status)
        log_info "Checking deployment status..."
        docker-compose ps
        exit 0
        ;;
    --logs)
        log_info "Showing deployment logs..."
        docker-compose logs -f --tail=100
        exit 0
        ;;
    *)
        main
        ;;
esac
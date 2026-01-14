#!/bin/bash

# Sentinel Hub Local Deployment Script
# This script deploys the Sentinel Hub API locally using Docker Compose

set -e

echo "🚀 Sentinel Hub Local Deployment"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    print_status "Checking Docker availability..."
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker Desktop and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Check if .env file exists
check_env() {
    print_status "Checking environment configuration..."
    if [ ! -f ".env" ]; then
        print_error ".env file not found. Please copy .env.example to .env and configure it."
        exit 1
    fi
    print_success "Environment configuration found"
}

# Clean up existing containers
cleanup() {
    print_status "Cleaning up existing containers..."
    docker-compose down -v 2>/dev/null || true
    docker system prune -f >/dev/null 2>&1 || true
    print_success "Cleanup completed"
}

# Build and start services
deploy() {
    print_status "Building and starting services..."
    print_status "This may take several minutes on first run..."

    # Start services with timeout
    timeout 600 docker-compose up --build -d

    if [ $? -eq 124 ]; then
        print_error "Deployment timed out after 10 minutes"
        print_status "Checking container status..."
        docker-compose ps
        exit 1
    fi

    print_success "Services started successfully"
}

# Wait for services to be healthy
wait_for_health() {
    print_status "Waiting for services to become healthy..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        print_status "Health check attempt $attempt/$max_attempts..."

        # Check database health
        if docker-compose exec -T db pg_isready -U sentinel -d sentinel >/dev/null 2>&1; then
            print_success "Database is healthy"
        else
            print_warning "Database not ready yet, waiting..."
            sleep 5
            ((attempt++))
            continue
        fi

        # Check API health
        if curl -f -s http://localhost:8080/health >/dev/null 2>&1; then
            print_success "API is healthy"
            break
        else
            print_warning "API not ready yet, waiting..."
            sleep 5
            ((attempt++))
            continue
        fi
    done

    if [ $attempt -gt $max_attempts ]; then
        print_error "Services failed to become healthy within timeout"
        print_status "Checking container logs..."
        docker-compose logs --tail=50
        exit 1
    fi

    print_success "All services are healthy!"
}

# Test API endpoints
test_endpoints() {
    print_status "Testing API endpoints..."

    # Test health endpoints
    echo "Testing /health endpoint..."
    if curl -s http://localhost:8080/health | grep -q "ok"; then
        print_success "Health endpoint working"
    else
        print_warning "Health endpoint returned unexpected response"
    fi

    echo "Testing /health/db endpoint..."
    if curl -s http://localhost:8080/health/db | grep -q "ok\|connected"; then
        print_success "Database health endpoint working"
    else
        print_warning "Database health endpoint returned unexpected response"
    fi

    # Test protected endpoints (should return 401 without auth)
    echo "Testing authentication on protected endpoints..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/tasks | grep -q "401"; then
        print_success "Authentication working (401 for unauthenticated request)"
    else
        print_warning "Authentication check returned unexpected status"
    fi

    # Test with authentication
    echo "Testing authenticated request..."
    response=$(curl -s -H "Authorization: Bearer dev-api-key-123" http://localhost:8080/api/v1/tasks)
    if echo "$response" | grep -q "tasks\|\[\]"; then
        print_success "Authenticated API request working"
    else
        print_warning "Authenticated request returned unexpected response: $response"
    fi
}

# Show deployment status
show_status() {
    print_success "Deployment completed successfully!"
    echo ""
    echo "📊 Deployment Status:"
    docker-compose ps
    echo ""
    echo "🌐 Service URLs:"
    echo "  • API: http://localhost:8080"
    echo "  • Health: http://localhost:8080/health"
    echo "  • Database: localhost:5432 (internal only)"
    echo ""
    echo "🔑 API Keys (from .env):"
    echo "  • dev-api-key-123"
    echo "  • test-api-key-456"
    echo "  • prod-api-key-789"
    echo ""
    echo "📝 Useful Commands:"
    echo "  • View logs: docker-compose logs -f"
    echo "  • Stop services: docker-compose down"
    echo "  • Restart API: docker-compose restart api"
    echo "  • Test API: curl -H 'Authorization: Bearer dev-api-key-123' http://localhost:8080/api/v1/tasks"
}

# Main deployment flow
main() {
    echo "Starting Sentinel Hub local deployment..."

    check_docker
    check_env
    cleanup
    deploy
    wait_for_health
    test_endpoints
    show_status

    print_success "🎉 Sentinel Hub is now running locally!"
    echo ""
    echo "Next steps:"
    echo "1. Open http://localhost:8080 in your browser"
    echo "2. Use the API keys above for authentication"
    echo "3. Check the logs with: docker-compose logs -f"
    echo "4. Stop with: docker-compose down"
}

# Handle command line arguments
case "${1:-}" in
    "cleanup")
        cleanup
        print_success "Cleanup completed"
        ;;
    "logs")
        docker-compose logs -f
        ;;
    "stop")
        docker-compose down
        print_success "Services stopped"
        ;;
    "restart")
        docker-compose restart
        print_success "Services restarted"
        ;;
    "status")
        docker-compose ps
        ;;
    *)
        main
        ;;
esac
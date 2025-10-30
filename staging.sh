#!/bin/bash

# Unburdy Server Staging Environment Management Script
# Usage: ./staging.sh [start|stop|restart|logs|build|clean|status]

set -e

COMPOSE_FILE="docker-compose-staging.yml"
PROJECT_NAME="unburdy-staging"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
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

# Check if docker and docker-compose are available
check_dependencies() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker compose &> /dev/null; then
        log_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
}

# Build the application
build() {
    log_info "Building Unburdy Server staging container..."
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME build --no-cache unburdy-server-staging
    log_success "Build completed successfully"
}

# Start services
start() {
    log_info "Starting Unburdy Server staging environment..."
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
    
    log_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check if services are running
    if docker compose -f $COMPOSE_FILE -p $PROJECT_NAME ps | grep -q "Up"; then
        log_success "Staging environment started successfully!"
        echo ""
        log_info "🌐 Unburdy Server: http://localhost:8080"
        log_info "📚 API Documentation: http://localhost:8080/swagger/index.html"
        log_info "🔍 Health Check: http://localhost:8080/api/v1/health"
        log_info "🗃️  Database: postgres://unburdy_user:unburdy_staging_password@localhost:5433/unburdy_staging"
        echo ""
        log_info "To enable additional tools:"
        log_info "  pgAdmin: docker compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile tools up -d"
        log_info "  Redis: docker compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile cache up -d"
        log_info "  Nginx: docker compose -f $COMPOSE_FILE -p $PROJECT_NAME --profile proxy up -d"
    else
        log_error "Failed to start some services. Check logs with: ./staging.sh logs"
        exit 1
    fi
}

# Stop services
stop() {
    log_info "Stopping Unburdy Server staging environment..."
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME down
    log_success "Staging environment stopped"
}

# Restart services
restart() {
    log_info "Restarting Unburdy Server staging environment..."
    stop
    sleep 2
    start
}

# Show logs
logs() {
    if [ -z "$2" ]; then
        docker compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
    else
        docker compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f "$2"
    fi
}

# Clean up everything
clean() {
    log_warning "This will remove all containers, volumes, and images for the staging environment!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up staging environment..."
        docker compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v --remove-orphans
        docker system prune -f
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Show status
status() {
    log_info "Staging environment status:"
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
    
    echo ""
    log_info "Container health:"
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME ps --format "table {{.Service}}\t{{.Status}}\t{{.Ports}}"
    
    # Check if services are responding
    echo ""
    log_info "Service health checks:"
    
    if curl -sf http://localhost:8080/api/v1/health > /dev/null 2>&1; then
        log_success "Unburdy Server is healthy"
    else
        log_error "Unburdy Server is not responding"
    fi
    
    if pg_isready -h localhost -p 5433 -U unburdy_user > /dev/null 2>&1; then
        log_success "PostgreSQL is ready"
    else
        log_error "PostgreSQL is not ready"
    fi
}

# Show help
help() {
    echo "Unburdy Server Staging Environment Management"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build     Build the staging container"
    echo "  start     Start the staging environment"
    echo "  stop      Stop the staging environment"
    echo "  restart   Restart the staging environment"
    echo "  logs      Show logs (optionally specify service name)"
    echo "  status    Show current status of services"
    echo "  clean     Remove all containers and volumes (destructive!)"
    echo "  help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start                    # Start all services"
    echo "  $0 logs                     # Show all logs"
    echo "  $0 logs unburdy-server-staging  # Show only app logs"
    echo "  $0 status                   # Check service status"
    echo ""
}

# Main script logic
check_dependencies

case "$1" in
    build)
        build
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    logs)
        logs "$@"
        ;;
    status)
        status
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        help
        ;;
    "")
        help
        ;;
    *)
        log_error "Unknown command: $1"
        help
        exit 1
        ;;
esac
#!/bin/bash

# Staging Environment Management Script
# Manages the complete staging environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

show_help() {
    echo -e "${BLUE}🚀 Staging Environment Manager${NC}"
    echo "==============================="
    echo ""
    echo -e "${YELLOW}Usage: $0 <command>${NC}"
    echo ""
    echo -e "${GREEN}Commands:${NC}"
    echo "  start         Start staging environment"
    echo "  stop          Stop staging environment"
    echo "  restart       Restart staging environment"
    echo "  build         Build staging containers"
    echo "  logs          Show logs"
    echo "  status        Show environment status"
    echo "  copy-db       Copy development database to staging"
    echo "  fresh-db      Create fresh empty database"
    echo "  shell         Open server container shell"
    echo "  db-shell      Open database shell"
    echo "  clean         Remove all containers and volumes"
    echo ""
    echo -e "${YELLOW}Advanced (uses staging.sh):${NC}"
    echo "  ./staging.sh copy-db    # Advanced database copy with options"
    echo "  ./staging.sh fresh-db   # Advanced database reset"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 start        # Start staging environment"
    echo "  $0 copy-db      # Copy dev data to staging"
    echo "  $0 logs         # Show all logs"
}

start_staging() {
    log_info "Starting staging environment..."
    docker compose up -d
    log_success "Staging environment started!"
    show_urls
}

stop_staging() {
    log_info "Stopping staging environment..."
    docker compose down
    log_success "Staging environment stopped"
}

restart_staging() {
    stop_staging
    sleep 2
    start_staging
}

build_staging() {
    log_info "Building staging containers..."
    docker compose build --no-cache
    log_success "Build completed"
}

show_logs() {
    if [ -z "$2" ]; then
        log_info "Showing all logs (press Ctrl+C to exit)..."
        docker compose logs -f
    else
        log_info "Showing logs for $2..."
        docker compose logs -f "$2"
    fi
}

show_status() {
    log_info "Staging environment status:"
    docker compose ps
    
    if docker compose ps | grep -q "Up"; then
        show_urls
    else
        log_warning "Staging environment is not running"
    fi
}

copy_database() {
    log_info "Copying development database to staging..."
    ./staging.sh copy-db
}

fresh_database() {
    log_warning "This will create a fresh empty database!"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker compose stop unburdy-server-staging || true
        docker compose down unburdy-db-staging
        docker volume rm staging_unburdy_staging_data 2>/dev/null || true
        docker compose up -d unburdy-db-staging
        log_info "Waiting for database..."
        sleep 10
        docker compose up -d unburdy-server-staging
        log_success "Fresh database created"
    fi
}

open_shell() {
    log_info "Opening server container shell..."
    docker compose exec unburdy-server-staging /bin/ash || docker compose exec unburdy-server-staging /bin/bash
}

open_db_shell() {
    log_info "Opening database shell..."
    docker compose exec unburdy-db-staging psql -U unburdy_user -d ae_saas_basic_test
}

clean_staging() {
    log_warning "This will remove all staging containers and volumes!"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker compose down -v --remove-orphans
        log_success "Staging environment cleaned"
    fi
}

show_urls() {
    echo ""
    log_info "🔗 Service URLs:"
    echo "   • Staging API: http://localhost:8080"
    echo "   • API Health: http://localhost:8080/api/v1/health"
    echo "   • Swagger UI: http://localhost:8080/swagger/index.html"
    echo "   • Database: localhost:5433 (user: unburdy_user, db: ae_saas_basic_test)"
}

case "${1:-help}" in
    start)
        start_staging
        ;;
    stop)
        stop_staging
        ;;
    restart)
        restart_staging
        ;;
    build)
        build_staging
        ;;
    logs)
        show_logs "$@"
        ;;
    status)
        show_status
        ;;
    copy-db)
        copy_database
        ;;
    fresh-db)
        fresh_database
        ;;
    shell)
        open_shell
        ;;
    db-shell)
        open_db_shell
        ;;
    clean)
        clean_staging
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
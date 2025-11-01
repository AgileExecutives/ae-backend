#!/bin/bash

# Development Database Management Script
# Starts PostgreSQL for development

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
    echo -e "${BLUE}🗄️  Development Database Manager${NC}"
    echo "=================================="
    echo ""
    echo -e "${YELLOW}Usage: $0 <command>${NC}"
    echo ""
    echo -e "${GREEN}Commands:${NC}"
    echo "  start     Start development database"
    echo "  stop      Stop development database"
    echo "  restart   Restart development database"
    echo "  logs      Show database logs"
    echo "  status    Show database status"
    echo "  shell     Open database shell"
    echo "  clean     Remove database volume (destructive!)"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 start        # Start PostgreSQL"
    echo "  $0 shell        # Open psql shell"
    echo "  $0 logs         # Show database logs"
}

start_db() {
    log_info "Starting development database..."
    docker compose up -d
    log_success "Development database started!"
    log_info "Database available at: postgres://postgres:pass@localhost:5432/ae_base_server"
}

stop_db() {
    log_info "Stopping development database..."
    docker compose down
    log_success "Development database stopped"
}

restart_db() {
    log_info "Restarting development database..."
    docker compose restart
    log_success "Development database restarted"
}

show_logs() {
    log_info "Showing database logs (press Ctrl+C to exit)..."
    docker compose logs -f postgres
}

show_status() {
    log_info "Development database status:"
    docker compose ps
    
    if docker compose ps | grep -q "Up"; then
        log_success "Database is running"
        log_info "Connection: postgres://postgres:pass@localhost:5432/ae_base_server"
    else
        log_warning "Database is not running"
    fi
}

open_shell() {
    log_info "Opening database shell..."
    docker compose exec postgres psql -U postgres -d ae_base_server
}

clean_db() {
    log_warning "This will remove the database volume and all data!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning database..."
        docker compose down -v
        log_success "Database volume removed"
    else
        log_info "Operation cancelled"
    fi
}

case "${1:-help}" in
    start)
        start_db
        ;;
    stop)
        stop_db
        ;;
    restart)
        restart_db
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    shell)
        open_shell
        ;;
    clean)
        clean_db
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
# Minimal Server

A minimal example server demonstrating the ae-base-server modular architecture. This server includes only the core base functionality without any additional modules.

## Features

This minimal server provides the complete base SaaS functionality plus a demonstration module:

### Base SaaS Features:
- **Authentication**: JWT-based user registration, login, and session management
- **User Management**: User profiles, settings, and account management  
- **Customer Management**: Customer relationship management with tenant isolation
- **Contact Management**: Contact lists with newsletter subscription support
- **Email Services**: Transactional email sending and template management
- **Health Monitoring**: System health checks and status endpoints

### Demo Module:
- **Ping Module**: A simple ping->pong module demonstrating how to add custom functionality

## Quick Start

1. **Setup Database**:
   ```bash
   # Create PostgreSQL database (or it will be auto-created)
   createdb ae_minimal
   ```

2. **Configure Environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

3. **Run Server**:
   ```bash
   go run main.go
   ```

4. **Test the API**:
   ```bash
   # Health check
   curl http://localhost:8080/health

   # Register a user
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "password123", "tenant_name": "My Company"}'

   # Login
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "password123"}'

   # Test the ping module
   curl http://localhost:8080/api/v1/ping/ping
   ```

## Available Endpoints

### Public Endpoints
- `GET /health` - Health check
- `POST /api/v1/auth/register` - User registration  
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/forgot-password` - Password reset request

### Protected Endpoints (require Bearer token)
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `GET /api/v1/contacts` - List contacts  
- `POST /api/v1/contacts` - Create contact
- `GET /api/v1/emails` - List emails
- `POST /api/v1/emails/send` - Send email

## Adding Modules

To extend this server with additional modules (like calendar, inventory, etc.), simply import the module and add it to the modules slice:

```go
import (
    "github.com/your-org/calendar-module/calendar"
)

func main() {
    // ... database setup ...
    
    // Add modules
    var modules []api.ModuleRouteProvider
    
    // Add calendar module
    calendarModule := calendar.NewModule(db)
    modules = append(modules, calendarModule)
    
    // Setup router with modules
    router := api.SetupModularRouter(db, modules)
    
    // ... start server ...
}
```

## Architecture

This minimal server demonstrates the ae-base-server modular architecture:

- **Base Server**: Provides core SaaS functionality (auth, users, customers, etc.)
- **Module System**: Clean interfaces for adding domain-specific features
- **Tenant Isolation**: Multi-tenant database design with automatic tenant context
- **Authentication**: JWT-based auth with automatic user/tenant extraction
- **Database**: PostgreSQL with GORM for migrations and seeding

## Database Schema

The server automatically creates and seeds the following tables:
- `users` - User accounts and profiles
- `tenants` - Tenant/organization data  
- `customers` - Customer relationship data
- `contacts` - Contact lists and newsletter subscriptions
- `emails` - Email sending history and templates
- Supporting tables for authentication and settings

## Configuration

All configuration is handled through environment variables. See `.env.example` for available options.

## Development

```bash
# Install dependencies
go mod tidy

# Run with live reload (if you have air installed)
air

# Build binary
go build -o bin/minimal-server main.go

# Run binary
./bin/minimal-server
```

## Production Deployment

1. Set production environment variables
2. Use a production-grade PostgreSQL database
3. Configure proper JWT secrets
4. Set up SSL/TLS termination
5. Use a process manager (systemd, supervisor, etc.)

This minimal server demonstrates how easily you can build a complete SaaS backend using the ae-base-server foundation, with the flexibility to add domain-specific modules as needed.
# Unburdy Extended API

A Go-based API server that extends the AE SaaS Basic API with client management functionality. This project imports the base `ae-saas/server-api` as a module and adds new endpoints for managing clients.

## Features

### Base Features (from AE SaaS Basic)
- 🔐 **Authentication & Authorization** - JWT-based auth system
- 👥 **User Management** - User registration, login, profile management
- 🏢 **Customer Management** - Customer CRUD operations
- 📧 **Email System** - Email sending and management
- 📋 **Contact Management** - Contact information handling
- 📄 **PDF Generation** - Document generation capabilities
- 🔍 **Search & Filtering** - Advanced search functionality

### Extended Features (Client Management)
- 👤 **Client Management** - Complete CRUD operations for clients
- 📝 **Client Fields**:
  - First Name
  - Last Name  
  - Date of Birth
- 🔍 **Client Search** - Search clients by name
- 📊 **Pagination Support** - Efficient data retrieval
- 🗑️ **Soft Delete** - Safe client removal

## API Endpoints

### Base Endpoints
All endpoints from the AE SaaS Basic API are available:
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/users` - Get all users
- `GET /api/v1/customers` - Get all customers
- `POST /api/v1/customers` - Create customer
- And many more...

### Extended Client Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/clients` | Get all clients (paginated) |
| `POST` | `/api/v1/clients` | Create new client |
| `GET` | `/api/v1/clients/search?q={query}` | Search clients by name |
| `GET` | `/api/v1/clients/{id}` | Get client by ID |
| `PUT` | `/api/v1/clients/{id}` | Update client |
| `DELETE` | `/api/v1/clients/{id}` | Delete client (soft delete) |

## Quick Start

### Prerequisites
- Go 1.24 or later
- The AE SaaS Basic API in `../../ae-saas/server-api`

### Installation

1. **Clone and navigate to the project**
   ```bash
   cd /Users/alex/src/AgileExecutives/unburdy/server-api
   ```

2. **Initialize the project**
   ```bash
   make init
   ```

3. **Start the server**
   ```bash
   make run
   ```

4. **Access the API**
   - Server: http://localhost:8080
   - Swagger docs: http://localhost:8080/swagger/index.html

## Development

### Available Make Commands

```bash
# Development
make run          # Start the server
make dev          # Start with hot reload
make build        # Build the application
make install      # Install dependencies

# Code Quality
make fmt          # Format code
make lint         # Run linter
make test         # Run tests

# Documentation
make swagger      # Generate Swagger docs
make endpoints    # List all API endpoints

# Utilities
make status       # Show project status
make check-base   # Verify base API availability
make clean        # Clean build artifacts
```

### Project Structure

```
unburdy/server-api/
├── main.go                     # Application entry point
├── go.mod                      # Go module definition
├── Makefile                    # Development commands
├── internal/
│   ├── models/
│   │   └── client.go          # Client data models
│   ├── services/
│   │   └── client_service.go  # Client business logic
│   ├── handlers/
│   │   └── client_handler.go  # HTTP handlers for clients
│   ├── router/
│   │   └── router.go          # Extended router setup
│   └── database/
│       └── database.go        # Extended database setup
└── README.md                  # This file
```

## API Usage Examples

### Create a Client
```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-15T00:00:00Z"
  }'
```

### Get All Clients
```bash
curl -X GET "http://localhost:8080/api/v1/clients?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Search Clients
```bash
curl -X GET "http://localhost:8080/api/v1/clients/search?q=John" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update a Client
```bash
curl -X PUT http://localhost:8080/api/v1/clients/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith"
  }'
```

## Database Schema

The client table is automatically created with the following structure:

```sql
CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

## Authentication

This API uses the same JWT-based authentication system as the base AE SaaS API. All client endpoints require a valid Bearer token in the Authorization header.

## Environment Variables

The same environment variables used by the base AE SaaS API apply here:

- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `JWT_SECRET` - JWT signing secret

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and test them
4. Commit your changes: `git commit -am 'Add feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the base AE SaaS Basic project for details.

## Architecture Notes

This project demonstrates how to extend an existing Go API by:
- Importing the base project as a Go module
- Adding new models, services, and handlers
- Extending the router with additional endpoints
- Maintaining compatibility with the base system
- Preserving all original functionality while adding new features

The modular approach allows for clean separation of concerns while leveraging the robust foundation provided by the AE SaaS Basic API.
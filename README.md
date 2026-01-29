# AE Backend

**Last Updated:** January 29, 2026

## Overview

**AE Backend** is a modular Go monorepo that powers scalable SaaS applications.  
It provides a solid foundation for authentication, multi-tenancy, and reusable business logic modules like Calendar, Billing, Invoicing, and Audit Logging.

> 📖 **Complete Documentation:** See [DOCUMENTATION.md](DOCUMENTATION.md)

---

## Quick Start

### Prerequisites
- Go 1.24.5+
- PostgreSQL 14+
- Redis 7+
- MinIO (for document storage)
- Docker & Docker Compose

### 1. Clone & Setup

```bash
git clone <repository-url>
cd ae-backend

# Initialize Go workspace
go work init
go work use base-server modules/* unburdy_server
```

### 2. Start Services

```bash
cd environments/dev
docker-compose up -d
```

### 3. Run Application

```bash
# Base server (port 8081)
cd base-server && go run main.go

# Unburdy server (port 8080)
cd unburdy_server && go run main.go
```

### 4. Access

- API: http://localhost:8080/api/v1
- Swagger: http://localhost:8080/swagger/index.html

---

## Project Structure

```
ae-backend/
├── base-server/           # Core SaaS foundation
│   ├── modules/           # Built-in modules (base, customer, email, pdf)
│   ├── pkg/               # Shared packages (core, settings, utils)
│   └── api/               # Public API exports
├── modules/               # Shared business modules
│   ├── audit/             # Audit trail logging
│   ├── booking/           # Appointment scheduling
│   ├── calendar/          # Calendar management
│   ├── documents/         # Document storage (MinIO)
│   ├── invoice/           # Invoice generation
│   └── invoice_number/    # Sequential numbering
├── unburdy_server/        # Production application
│   └── modules/           # App-specific modules
├── documentation/         # Detailed docs
└── environments/          # Docker configs
```

---

## Key Features

### ✅ Modular Architecture
- Plug-and-play modules with clean boundaries
- Independent testing and development
- Easy integration into new services

### ✅ Multi-Tenant Ready
- Built-in tenant isolation at database and application layers
- Automatic tenant scoping via middleware
- Ideal for B2B/multi-organization platforms

### ✅ Production Ready
- JWT authentication
- Comprehensive API documentation (Swagger)
- Audit logging
- Document management (MinIO)
- PDF generation (ChromeDP)
- Email services (SMTP)

### ✅ Monorepo Benefits
- Unified dependency management
- Atomic commits across modules
- Consistent testing and deployment

---

## Core Modules

**Base Server Modules:**
- **base** - Authentication, users, tenants
- **customer** - Customer & subscription management
- **email** - Email delivery (SMTP)
- **pdf** - PDF generation (ChromeDP)
- **templates** - Document templates

**Shared Modules** (in `/modules`):
- **audit** - Audit trail logging
- **booking** - Appointment scheduling
- **calendar** - Calendar & events
- **documents** - Document storage
- **invoice** - Invoice generation & management
- **invoice_number** - Sequential numbering

---

## Documentation

- **[DOCUMENTATION.md](DOCUMENTATION.md)** - Complete system documentation
- **[documentation/](documentation/)** - Detailed feature docs
  - [Architecture.md](documentation/Architecture.md)
  - [MODULE_DEVELOPMENT_GUIDE.md](documentation/MODULE_DEVELOPMENT_GUIDE.md)
  - [INVOICE_CANCELLATION.md](documentation/INVOICE_CANCELLATION.md)
  - [AUDIT_TRAIL_README.md](documentation/AUDIT_TRAIL_README.md)
  - And many more...

---

## Recent Changes

### January 29, 2026
- ✅ **Audit Module** - Moved to shared modules
- ✅ **Invoice Module** - Refactored to use PDF module
- ✅ **Settings System** - Fixed API mismatches
- ✅ **Documentation** - Consolidated and validated
- ✅ **Email Module** - Simplified architecture

### January 26, 2026
- ✅ **Invoice Cancellation** - GoBD-compliant storno feature

---

## Development

### Adding a Module

See [DOCUMENTATION.md#module-system](DOCUMENTATION.md#module-system) for complete guide.

Quick example:

```bash
# 1. Create module
mkdir modules/your-module
cd modules/your-module
go mod init github.com/unburdy/your-module

# 2. Implement Module interface
# See documentation/MODULE_DEVELOPMENT_GUIDE.md

# 3. Add to application
# Import and register in main.go
```

---

## API Documentation

Generate Swagger docs:

```bash
cd unburdy_server
make swag
```

Access at: http://localhost:8080/swagger/index.html

---

## Testing

```bash
# Run all tests
go test ./...

# Run specific module tests
cd modules/booking && go test ./...

# Generate coverage
go test -cover ./...
```

---

## Deployment

See [DOCUMENTATION.md#deployment](DOCUMENTATION.md#deployment) for production checklist.

Docker build:
```bash
cd unburdy_server
docker build -t unburdy-server:latest .
```

---

## Support

- **Documentation:** [DOCUMENTATION.md](DOCUMENTATION.md)
- **Module Guides:** [documentation/](documentation/)
- **Issues:** GitHub Issues

---

**Built with ❤️ using Go, Gin, GORM, PostgreSQL, Redis, and MinIO**
   ```

---

## What I love
 
- **Clean separation**: base logic vs. business logic vs. services  
- **Reliable auth & tenancy**: built in from day one  
- **Consistent development flow**: Go workspace + monorepo simplicity  
- **Future-proof**: modules can evolve into microservices later  

---


## License
This project is licensed under the MIT License — see the `LICENSE` file for details.

````

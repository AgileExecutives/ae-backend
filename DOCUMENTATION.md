# AE Backend - Complete Documentation

**Last Updated:** January 29, 2026  
**Version:** 2.0

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Quick Start](#quick-start)
4. [Module System](#module-system)
5. [Core Modules](#core-modules)
6. [Shared Modules](#shared-modules)
7. [Development Guide](#development-guide)
8. [API Documentation](#api-documentation)
9. [Deployment](#deployment)

---

## Overview

AE Backend is a **modular Go monorepo** for building scalable, multi-tenant SaaS applications. It provides:

- ✅ **Modular Architecture** - Plug-and-play modules with clean boundaries
- ✅ **Multi-Tenant Ready** - Built-in tenant isolation at database and application layers
- ✅ **Monorepo Benefits** - Unified dependency management and atomic commits
- ✅ **Production Ready** - JWT authentication, audit logging, and comprehensive API documentation

### Project Structure

```
ae-backend/
├── base-server/              # Core SaaS foundation
│   ├── modules/              # Built-in modules (base, customer, email, pdf, etc.)
│   ├── pkg/                  # Shared packages (core, settings, utils)
│   └── api/                  # Public API exports
├── modules/                  # Shared business modules
│   ├── audit/                # Audit trail logging ✨ RECENTLY MOVED
│   ├── booking/              # Appointment scheduling
│   ├── calendar/             # Calendar management
│   ├── documents/            # Document storage (MinIO)
│   ├── invoice/              # Invoice generation
│   └── invoice_number/       # Sequential invoice numbering
├── unburdy_server/           # Production application
│   └── modules/              # App-specific modules
│       ├── client_management/  # Client & cost provider management
│       └── settings_api/       # Settings API wrapper
├── minimal-server/           # Minimal example server
├── documentation/            # Detailed documentation
└── environments/             # Docker compose configs

```

---

## Architecture

### Core Concepts

#### 1. **Base Server**
The foundation providing:
- Authentication & Authorization (JWT)
- User & Tenant Management
- Customer & Plan Management
- Email Services (SMTP)
- PDF Generation (ChromeDP)
- Settings System
- Health Checks

#### 2. **Module System**
Each module is a self-contained unit implementing:

```go
type Module interface {
    Name() string
    Version() string
    Dependencies() []string
    Initialize(ctx ModuleContext) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Entities() []Entity
    Routes() []RouteProvider
    EventHandlers() []EventHandler
    Services() []ServiceProvider
    Middleware() []MiddlewareProvider
}
```

#### 3. **Multi-Tenancy**
All entities include:
```go
type BaseEntity struct {
    UserID         uint   `gorm:"not null;index:idx_user_tenant"`
    TenantID       uint   `gorm:"not null;index:idx_user_tenant"`
    OrganizationID string `gorm:"type:varchar(255);index"`
}
```

Automatic tenant isolation via middleware.

#### 4. **Dependency Management**
Go workspace with replace directives:
```go
replace github.com/ae-base-server => ./base-server
replace github.com/unburdy/audit-module => ./modules/audit
```

---

## Quick Start

### Prerequisites
- Go 1.24.5+
- PostgreSQL 14+
- Redis 7+ (for caching, invoice numbers)
- MinIO (for document storage)
- Docker & Docker Compose (recommended)

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

# Verify services
psql -h localhost -U postgres -d ae_dev
redis-cli -h localhost -p 6379 -a redis123 ping
```

### 3. Run Application

```bash
# Base server (port 8081)
cd base-server
go run main.go

# Unburdy server (port 8080)
cd unburdy_server
go run main.go
```

### 4. Access API

- Base Server: http://localhost:8081/api/v1
- Unburdy Server: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html

---

## Module System

### Module Lifecycle

1. **Registration** - Modules registered in `main.go`
2. **Initialization** - `Initialize()` called with context (DB, logger, event bus)
3. **Start** - `Start()` called after all modules initialized
4. **Running** - Routes active, event handlers listening
5. **Stop** - `Stop()` called on shutdown

### Creating a New Module

#### Step 1: Create Module Structure

```bash
mkdir modules/your-module
cd modules/your-module
go mod init github.com/unburdy/your-module
```

#### Step 2: Define go.mod

```go
module github.com/unburdy/your-module

go 1.24.5

require (
    github.com/ae-base-server v0.0.0
    github.com/gin-gonic/gin v1.10.1
    gorm.io/gorm v1.30.0
)

replace github.com/ae-base-server => ../../base-server
```

#### Step 3: Implement Module

```go
// module.go
package yourmodule

import (
    "context"
    "github.com/ae-base-server/pkg/core"
)

type YourModule struct {
    // dependencies
}

func NewModule() core.Module {
    return &YourModule{}
}

func (m *YourModule) Name() string { return "your-module" }
func (m *YourModule) Version() string { return "1.0.0" }
func (m *YourModule) Dependencies() []string { return []string{"base"} }

func (m *YourModule) Initialize(ctx core.ModuleContext) error {
    // Setup handlers, services, etc.
    return nil
}

// Implement other Module interface methods...
```

#### Step 4: Use in Application

```go
// main.go
import yourmodule "github.com/unburdy/your-module"

registry := core.NewModuleRegistry(logger, eventBus, db)
registry.Register(yourmodule.NewModule())
```

---

## Core Modules

### Base Module
**Location:** `base-server/modules/base`  
**Purpose:** Core authentication & user management

**Features:**
- JWT authentication
- User registration & login
- Password reset
- Tenant management
- Contact forms
- Newsletter subscriptions

**Endpoints:**
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/reset-password`
- `GET /users/me`

### Customer Module
**Location:** `base-server/modules/customer`  
**Purpose:** Customer & subscription management

**Features:**
- Customer CRUD
- Subscription plans
- Plan assignments
- Customer events

### Email Module
**Location:** `base-server/modules/email`  
**Purpose:** Email delivery

**Features:**
- SMTP integration
- Template email sending
- Verification emails
- Password reset emails

**Architecture:** ✅ **Simplified** - No template rendering (delegates to calling service)

### PDF Module
**Location:** `base-server/modules/pdf`  
**Purpose:** PDF generation from HTML

**Features:**
- HTML to PDF conversion via ChromeDP
- Template-based PDF generation
- A4 format support

**Recent Update:** ✅ Added `ConvertHtmlStringToPDF()` for direct HTML→PDF conversion

### Templates Module
**Location:** `base-server/modules/templates`  
**Purpose:** Document template management

**Features:**
- Template CRUD with versioning
- Contract-based rendering
- Multi-channel support (EMAIL, DOCUMENT)
- MinIO storage for template assets

---

## Shared Modules

### Audit Module ✨
**Location:** `modules/audit` (moved from `unburdy_server/modules/audit`)  
**Purpose:** Audit trail logging

**Features:**
- User action logging
- Entity change tracking
- Filtering by user, entity type, action, date range
- Pagination support

**Recent Changes:**
- ✅ Moved to shared modules directory (Jan 29, 2026)
- ✅ Created independent module with go.mod
- ✅ Updated to use `baseAPI` instead of internal models

**Dependencies:** `base` module

### Booking Module
**Location:** `modules/booking`  
**Purpose:** Appointment scheduling

**Features:**
- Booking link generation with tokens
- Token validation (bearer + organization)
- Time slot management
- Email confirmations

### Calendar Module
**Location:** `modules/calendar`  
**Purpose:** Calendar & availability management

**Features:**
- Calendar CRUD
- Recurring events
- Availability windows
- Integration tests

### Documents Module
**Location:** `modules/documents`  
**Purpose:** Document storage & retrieval

**Features:**
- Multi-tenant document storage (MinIO)
- SHA256 checksum validation
- Pre-signed download URLs (1-hour expiry)
- Template management with versioning
- Metadata & tagging (JSONB)

**Dependencies:** `base`, `templates`

### Invoice Module
**Location:** `modules/invoice`  
**Purpose:** Invoice generation & management

**Features:**
- Invoice CRUD with draft/finalized states
- PDF generation via template module
- VAT handling (German tax law compliant)
- Credit notes
- XRechnung export (UBL 2.1)
- GoBD-compliant cancellation

**Architecture:** ✅ **Refactored** - Uses PDF module, not direct ChromeDP

**Recent Changes:**
- ✅ Removed direct chromedp dependency (Jan 29, 2026)
- ✅ Uses `pdfGenerator.ConvertHtmlStringToPDF()`
- ✅ Template rendering via template module

**Dependencies:** `base`, `templates`, PDF service

### Invoice Number Module
**Location:** `modules/invoice_number`  
**Purpose:** Sequential invoice number generation

**Features:**
- Redis-backed sequential numbering
- PostgreSQL persistence
- Distributed locking
- Custom prefix support
- Zero-padding

---

## Development Guide

### Settings System

The Advanced Settings System provides module-driven configuration:

```go
// Define settings in your module
type YourSettingsProvider struct{}

func (p *YourSettingsProvider) GetModuleName() string {
    return "your-module"
}

func (p *YourSettingsProvider) GetDomain() string {
    return entities.DomainYourModule
}

func (p *YourSettingsProvider) GetSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "api_key": map[string]interface{}{
                "type": "string",
                "description": "API key for service",
            },
        },
    }
}
```

**Current Status:** ✅ Settings service fixed (Jan 29, 2026)
- Uses `Data` field (JSONB) instead of `Value`/`Type`
- Removed organizationID from repository calls
- Added `parseSettingData()` and `serializeData()` methods

### Error Handling

Always use base API response types:

```go
import baseAPI "github.com/ae-base-server/api"

// Success
c.JSON(http.StatusOK, baseAPI.SuccessResponse(data))

// Error
c.JSON(http.StatusBadRequest, 
    baseAPI.ErrorResponseFunc("Invalid input", err.Error()))
```

### Event System

Emit events for cross-module communication:

```go
ctx.EventBus.Emit("user.created", user)
ctx.EventBus.Emit("invoice.finalized", invoice)
```

### Database Migrations

Entities are automatically migrated via `Entities()` method:

```go
func (m *YourModule) Entities() []core.Entity {
    return []core.Entity{
        entities.NewYourEntity(),
    }
}
```

---

## API Documentation

### Swagger Generation

```bash
# Base server
cd base-server
make swag

# Unburdy server (includes all modules)
cd unburdy_server
make swag
```

Access at: `http://localhost:8080/swagger/index.html`

### Authentication

All protected endpoints require JWT:

```bash
Authorization: Bearer <jwt_token>
```

Get token via:
```bash
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password"
}
```

---

## Deployment

### Docker Build

```bash
cd unburdy_server
docker build -t unburdy-server:latest .
```

### Environment Variables

Required:
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection  
- `MINIO_ENDPOINT` - MinIO endpoint
- `MINIO_ACCESS_KEY` - MinIO access key
- `MINIO_SECRET_KEY` - MinIO secret key
- `JWT_SECRET` - JWT signing secret
- `SMTP_HOST` - Email server host
- `SMTP_PORT` - Email server port
- `SMTP_USER` - SMTP username
- `SMTP_PASSWORD` - SMTP password

### Production Checklist

- [ ] Set strong `JWT_SECRET`
- [ ] Configure TLS for database
- [ ] Set up SSL for MinIO
- [ ] Configure proper CORS settings
- [ ] Enable rate limiting
- [ ] Set up monitoring (health checks)
- [ ] Configure backup strategy (PostgreSQL + MinIO)
- [ ] Review and set log levels
- [ ] Test disaster recovery

---

## Recent Changes & Migrations

### January 29, 2026

#### Audit Module Migration
✅ Moved from `unburdy_server/modules/audit` to `modules/audit`
- Created independent module with `go.mod`
- Changed imports: `github.com/unburdy/unburdy-server-api/modules/audit/*` → `github.com/unburdy/audit-module/*`
- Updated to use `baseAPI` instead of internal models
- Added to unburdy_server dependencies

#### Invoice Module Refactoring
✅ Removed direct dependency on ChromeDP/MinIO
- Now uses PDF module's `ConvertHtmlStringToPDF()`
- Uses template module for rendering
- Will use documents module for storage (TODO)
- Cleaner separation of concerns

#### Settings System Fix
✅ Fixed API mismatches
- Updated to use `Data` field (JSONB) vs old `Value`/`Type`
- Fixed repository method signatures
- Removed unused `organizationID` parameter
- Added JSONB parsing methods

#### Email Module Simplification
✅ Simplified architecture
- Removed template rendering responsibility
- Calling services now handle template rendering
- Email service focuses on delivery only

### January 26, 2026

#### Invoice Cancellation Feature
✅ GoBD-compliant invoice cancellation
- Storno records with immutable audit trail
- Cancellation only for unsent invoices
- Complete API documentation

---

## Documentation Index

Detailed documentation in `/documentation`:

### Core
- [Architecture.md](documentation/Architecture.md) - System architecture
- [DevPrinciples.md](documentation/DevPrinciples.md) - Coding standards
- [MODULE_DEVELOPMENT_GUIDE.md](documentation/MODULE_DEVELOPMENT_GUIDE.md) - Module creation

### Features
- [ADVANCED_SETTINGS_SYSTEM.md](documentation/ADVANCED_SETTINGS_SYSTEM.md) - Settings system
- [AUDIT_TRAIL_README.md](documentation/AUDIT_TRAIL_README.md) - Audit logging
- [TEMPLATE_SYSTEM_ARCHITECTURE.md](documentation/TEMPLATE_SYSTEM_ARCHITECTURE.md) - Templates

### Invoice System
- [INVOICE_CANCELLATION.md](documentation/INVOICE_CANCELLATION.md) - Cancellation feature
- [INVOICE_PDF_GENERATION.md](documentation/INVOICE_PDF_GENERATION.md) - PDF generation
- [INVOICE_VAT_HANDLING.md](documentation/INVOICE_VAT_HANDLING.md) - VAT calculations
- [XRECHNUNG_README.md](documentation/XRECHNUNG_README.md) - German e-invoicing

### API
- [FRONTEND_INTEGRATION_GUIDE.md](documentation/FRONTEND_INTEGRATION_GUIDE.md) - Frontend integration
- [SWAGGER_DOCUMENTATION.md](documentation/SWAGGER_DOCUMENTATION.md) - API specs

---

## Support

- **Issues:** GitHub Issues
- **Email:** support@ae-backend.com
- **Documentation:** `/documentation`

---

**Built with ❤️ using Go, Gin, GORM, PostgreSQL, Redis, and MinIO**

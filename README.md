# AE Backend

## Overview

**AE Backend** is a modular Go monorepo that powers scalable SaaS applications.  
It provides a solid foundation for authentication, multi-tenancy, and reusable business logic modules like Calendar, Billing, or Notifications.

This project aims to make backend development **faster, cleaner, and more maintainable** — ideal for building multiple related services that share common functionality.

---

## Purpose

- Provide a **base-server** with shared functionality (auth, org management, user handling)
- Enable fast development through **plug-and-play modules**
- Ensure **tenant isolation** and consistent **authentication patterns**
- Support a **monorepo** setup for easy dependency management and unified CI/CD

You can think of AE Backend as a **Go-based application framework for SaaS backends**, with modular extensibility and simple integration.

---

## Why This Approach Works

### ✅ Modular by Design
Each feature (e.g., calendar, billing) lives in its own module:
- Fully testable and independent
- Easy to integrate into new services
- Promotes reusability and clean boundaries

### ✅ Single Source of Truth
- Core auth, user, and tenant logic live in **base-server**
- Shared models and utilities are consistent across modules

### ✅ Multi-Tenant Ready
- Built-in tenant isolation
- Every query automatically filters by `organization_id`
- Simplifies building B2B or multi-organization platforms

### ✅ Monorepo Benefits
- One place for all modules and services
- Atomic commits and version alignment
- Easier testing and deployment across projects

---

## Development Environment Setup

### Prerequisites
- Go 1.22+
- PostgreSQL (local or Docker)
- `make` or simple bash scripts for setup
- Optional: Docker & Docker Compose

### Clone & Configure
```bash
git clone https://github.com/your-org/ae-backend.git
cd ae-backend
```

### Initialize Go Workspace
```bash
go work init
go work use base-server modules/* unburdy_server
go mod tidy
```

### Database Setup
```bash
createdb ae_backend_dev
```

### Run the Base Server
```bash
cd base-server
go run main.go
```

### Run the Example Service
```bash
cd unburdy_server
go run main.go
```

---

## Adding a Module

1. Create a new folder under `modules/`
2. Initialize Go module:
   ```bash
   go mod init github.com/ae-backend/your-module
   ```
3. Implement `models.go`, `handlers.go`, `routes.go`
4. Integrate into your service with:
   ```go
   import "github.com/ae-backend/your-module"

   yourmodule.Migrate(db)
   yourmodule.RegisterRoutes(protectedGroup, db)
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

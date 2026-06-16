# Environment Management

This directory contains Docker Compose configurations and management scripts for different environments.

## 📁 Directory Structure

```
environments/
├── dev-db/                 # Development database only
│   ├── docker-compose.yml  # PostgreSQL for development
│   └── manage.sh           # Database management script
├── staging/                # Complete staging environment
│   ├── docker-compose.yml  # Full staging setup with app + database
│   ├── Dockerfile.staging  # Staging container build configuration
│   ├── manage.sh          # Simple staging environment management
│   └── staging.sh         # Advanced staging operations (database copy, etc.)
└── README.md              # This file
```

## 🗄️ Development Database (`dev-db/`)

Provides a standalone PostgreSQL database for development work.

### Quick Start:
```bash
cd environments/dev-db/
./manage.sh start          # Start PostgreSQL
./manage.sh shell          # Open database shell
./manage.sh stop           # Stop database
```

### Available Commands:
- `start` - Start development database
- `stop` - Stop development database  
- `restart` - Restart database
- `logs` - Show database logs
- `status` - Show database status
- `shell` - Open psql shell
- `clean` - Remove database volume (destructive!)

### Connection Details:
- **Host**: localhost:5432
- **Database**: ae_base_server
- **User**: postgres
- **Password**: pass

## 🚀 Staging Environment (`staging/`)

Complete staging environment with application server and database.

### Quick Start:
```bash
cd environments/staging/
./manage.sh start          # Start complete staging environment
./manage.sh status         # Check status and get URLs
./manage.sh copy-db        # Copy dev data (run from main directory)
```

### Available Commands:

**Simple Management (manage.sh):**
- `start` - Start complete staging environment
- `stop` - Stop staging environment
- `restart` - Restart staging environment
- `build` - Build staging containers
- `logs [service]` - Show logs (all or specific service)
- `status` - Show environment status with URLs
- `copy-db` - Copy development database to staging
- `fresh-db` - Create fresh empty database
- `shell` - Open server container shell
- `db-shell` - Open database shell
- `clean` - Remove all containers and volumes

**Advanced Operations (staging.sh):**
- `copy-db` - Advanced database copy with validation and sanitization
- `fresh-db` - Advanced database reset with options
- `start` - Start with detailed logging
- `stop` - Stop with cleanup
- `restart` - Restart with validation
- `build` - Build with progress tracking
- `status` - Detailed status with health checks
- `logs` - Advanced log filtering
- `clean` - Complete environment cleanup

### Service URLs:
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/api/v1/health
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Database**: localhost:5433 (user: unburdy_user, db: ae_saas_basic_test)

## 🔄 Database Copying Workflow

To copy development data to staging:

1. **Start development database:**
   ```bash
   cd environments/dev-db/
   ./manage.sh start
   ```

2. **Start staging environment:**
   ```bash
   cd environments/staging/
   ./manage.sh start
   ```

3. **Copy database:**
   ```bash
   cd environments/staging/
   ./manage.sh copy-db        # Simple copy
   # OR
   ./staging.sh copy-db       # Advanced copy with validation
   ```

## 📋 Migration Notes

The environments have been reorganized from:
- `base-server/docker-compose.yml` → `environments/dev-db/docker-compose.yml`
- `docker-compose-staging.yml` → `environments/staging/docker-compose.yml`
- `Dockerfile.staging` → `environments/staging/Dockerfile.staging`
- `staging.sh` → `environments/staging/staging.sh`

### Updated Scripts:
- All staging files moved to `environments/staging/`
- Database copy operations look in `../dev-db/`
- Docker build context updated for new Dockerfile location
- All paths updated for the new structure

## 🛠️ Troubleshooting

### Development Database Issues:
```bash
cd environments/dev-db/
./manage.sh status    # Check if running
./manage.sh logs      # Check for errors
./manage.sh clean     # Reset if corrupted
./manage.sh start     # Start fresh
```

### Staging Environment Issues:
```bash
cd environments/staging/
./manage.sh status    # Check service status
./manage.sh logs      # Check for errors
./manage.sh build     # Rebuild if needed
./manage.sh restart   # Restart services
```

### Port Conflicts:
- Dev database: 5432
- Staging database: 5433
- Staging API: 8080

If you have port conflicts, stop other services or modify the ports in the respective `docker-compose.yml` files.

## 🔧 Customization

### Environment Variables
Both environments support customization through Docker Compose environment variables. Check the respective `docker-compose.yml` files for available options.

### Volume Persistence
- Development: `postgres_data` volume
- Staging: `unburdy_staging_data` volume

Data persists between container restarts but can be removed with the `clean` commands.
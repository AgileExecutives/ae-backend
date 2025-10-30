# Unburdy Server Staging Environment

This directory contains the Docker configuration for running the Unburdy Server in a staging environment with its own dedicated database.

## 🚀 Quick Start

1. **Start the staging environment:**
   ```bash
   ./staging.sh start
   ```

2. **Access the services:**
   - **Unburdy API**: http://localhost:8080
   - **API Documentation**: http://localhost:8080/swagger/index.html
   - **Health Check**: http://localhost:8080/api/v1/health
   - **Database**: postgres://unburdy_user:unburdy_staging_password@localhost:5433/unburdy_staging

3. **Stop the staging environment:**
   ```bash
   ./staging.sh stop
   ```

## 📁 Files Overview

### Core Files
- **`Dockerfile.staging`**: Multi-stage Docker build for the Unburdy Server
- **`docker-compose-staging.yml`**: Complete Docker Compose setup with PostgreSQL database
- **`staging.sh`**: Management script for easy operations

### Configuration Files
- **`unburdy_server/scripts/init-staging-db.sql`**: Database initialization script
- **`.env`**: Environment variables (auto-created in container)

## 🔧 Available Commands

```bash
# Management script commands
./staging.sh build     # Build the staging container
./staging.sh start     # Start all services
./staging.sh stop      # Stop all services
./staging.sh restart   # Restart all services
./staging.sh logs      # Show all logs
./staging.sh status    # Show service status
./staging.sh clean     # Remove all containers and volumes (destructive!)

# Show logs for specific service
./staging.sh logs unburdy-server-staging
./staging.sh logs unburdy-db-staging
```

## 🗃️ Database

The staging environment uses a dedicated PostgreSQL 15 database with the following configuration:

- **Database Name**: `unburdy_staging`
- **Username**: `unburdy_user`
- **Password**: `unburdy_staging_password`
- **Port**: `5433` (exposed on host to avoid conflicts)

### Database Features
- **Automatic Initialization**: Database is created and initialized on first startup
- **Automatic Seeding**: Application handles migration and seeding of test data
- **Performance Tuning**: Optimized PostgreSQL settings for development/staging
- **Persistent Data**: Data is preserved in Docker volumes

## 🔧 Optional Services

The Docker Compose setup includes optional services that can be enabled with profiles:

### pgAdmin (Database Management)
```bash
# Start with pgAdmin
docker compose -f docker-compose-staging.yml --profile tools up -d

# Access pgAdmin at http://localhost:5050
# Login: admin@unburdy-staging.com / pgadmin_staging_password
```

### Redis (Caching)
```bash
# Start with Redis
docker compose -f docker-compose-staging.yml --profile cache up -d

# Redis available at localhost:6379
# Password: redis_staging_password
```

### Nginx (Reverse Proxy)
```bash
# Start with Nginx
docker compose -f docker-compose-staging.yml --profile proxy up -d

# HTTP proxy available at localhost:80
```

## 🌍 Environment Variables

The staging environment uses the following key configuration:

### Application Settings
- `GIN_MODE=release`
- `SERVER_HOST=0.0.0.0`
- `SERVER_PORT=8080`

### Database Settings
- `DB_HOST=unburdy-db-staging`
- `DB_PORT=5432`
- `DB_NAME=unburdy_staging`
- `DB_USER=unburdy_user`
- `DB_PASSWORD=unburdy_staging_password`

### Security Settings
- `JWT_SECRET=unburdy-staging-jwt-secret-key-change-in-production`
- `JWT_EXPIRY=24h`

### Email Settings (Mailtrap)
- `SMTP_HOST=smtp.mailtrap.io`
- `SMTP_PORT=587`
- Configure `SMTP_USERNAME` and `SMTP_PASSWORD` for your Mailtrap account

## 📊 Health Monitoring

The setup includes comprehensive health monitoring:

### Application Health Checks
- **Endpoint**: `/api/v1/health`
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Retries**: 3

### Database Health Checks
- **Command**: `pg_isready`
- **Interval**: 10 seconds
- **Timeout**: 5 seconds
- **Retries**: 5

### Service Status Monitoring
```bash
./staging.sh status  # Check all service health
```

## 🔒 Security Features

### Container Security
- **Non-root user**: Application runs as `unburdy` user
- **Read-only filesystem**: Static files are read-only
- **Resource limits**: Memory and CPU constraints
- **Network isolation**: Services communicate via dedicated network

### Database Security
- **Isolated network**: Database only accessible from application container
- **Custom credentials**: Non-default username and password
- **Connection limits**: Configured connection pooling

## 📝 Logs

### View Logs
```bash
# All services
./staging.sh logs

# Specific service
./staging.sh logs unburdy-server-staging
./staging.sh logs unburdy-db-staging

# Follow logs in real-time
docker compose -f docker-compose-staging.yml logs -f
```

### Log Locations
- **Application logs**: `/app/logs/` (mounted to `./unburdy_server/logs/`)
- **Database logs**: `/var/log/postgresql/` (in database container)

## 🚨 Troubleshooting

### Common Issues

1. **Port conflicts:**
   ```bash
   # Check what's using the ports
   lsof -i :8080  # Unburdy Server
   lsof -i :5433  # PostgreSQL
   ```

2. **Database connection issues:**
   ```bash
   # Check database health
   docker compose -f docker-compose-staging.yml exec unburdy-db-staging pg_isready -U unburdy_user
   
   # Check logs
   ./staging.sh logs unburdy-db-staging
   ```

3. **Application startup issues:**
   ```bash
   # Check application logs
   ./staging.sh logs unburdy-server-staging
   
   # Rebuild container
   ./staging.sh build
   ```

4. **Clean restart:**
   ```bash
   ./staging.sh clean  # Warning: This removes all data!
   ./staging.sh start
   ```

### Reset Everything
```bash
# Complete cleanup (removes all data and containers)
./staging.sh clean

# Fresh start
./staging.sh build
./staging.sh start
```

## 🔄 Development Workflow

1. **Make code changes** in the unburdy_server directory
2. **Rebuild the container**: `./staging.sh build`
3. **Restart services**: `./staging.sh restart`
4. **Test your changes**: Visit http://localhost:8080/swagger/index.html
5. **Check logs**: `./staging.sh logs unburdy-server-staging`

## 📚 API Documentation

Once the staging environment is running, you can access:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json
- **Health Check**: http://localhost:8080/api/v1/health

## 🔗 Related Documentation

- [Main Project README](../README.md)
- [Base Server Documentation](../base-server/README.md)
- [Unburdy Server Documentation](../unburdy_server/README.md)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

---

**Note**: This is a staging environment. Do not use the default passwords and secrets in production!
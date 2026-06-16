# Development Environment

Docker Compose setup for local development with PostgreSQL and MinIO (S3-compatible storage).

## Services

### PostgreSQL
Database for application data.

- **Container Name**: `ae-base-server-postgres`
- **Port**: `5432`
- **Database**: `ae_base_server`
- **User**: `postgres`
- **Password**: `pass`
- **Connection String**: `postgresql://postgres:pass@localhost:5432/ae_base_server`

**Connect with psql:**
```bash
psql -h localhost -U postgres -d ae_base_server
```

### MinIO (S3-Compatible Object Storage)
Document and file storage with S3-compatible API.

- **Container Name**: `ae-minio`
- **API Port**: `9000`
- **Web Console Port**: `9001`
- **Root User**: `minioadmin`
- **Root Password**: `minioadmin123`
- **Console URL**: http://localhost:9001
- **API Endpoint**: http://localhost:9000

**Buckets to create:**
- `invoices` - Invoice PDF storage
- `templates` - Invoice and document templates
- `documents` - General document storage

**Access via MinIO Client (mc):**
```bash
mc alias set local http://localhost:9000 minioadmin minioadmin123
mc ls local
mc mb local/invoices
mc mb local/templates
mc mb local/documents
```

**S3 SDK Configuration (Go):**
```go
endpoint := "localhost:9000"
accessKeyID := "minioadmin"
secretAccessKey := "minioadmin123"
useSSL := false
```

## Usage

### Start Services
```bash
cd /Users/alex/src/ae/backend/environments/dev
docker-compose up -d
```

### Check Status
```bash
docker-compose ps
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f postgres
docker-compose logs -f minio
```

### Stop Services
```bash
docker-compose down
```

### Stop and Remove Volumes (Delete All Data)
```bash
docker-compose down -v
```

### Restart a Service
```bash
docker-compose restart postgres
docker-compose restart minio
docker-compose restart redis
```

## Health Checks

All services have health checks configured:

- **PostgreSQL**: `pg_isready` every 10s
- **MinIO**: HTTP health endpoint every 3
Check health status:
```bash
docker inspect ae-base-server-postgres | grep -A 5 Health
docker inspect ae-minio | grep -A 5 Health
docker inspect ae-redis | grep -A 5 Health
```
```

## Data Persistence

Data is persisted in Docker volumes:

- `postgres_data` - PostgreSQL database files
- `minio_data` - MinIO object storag
```bash
docker volume ls | grep ae
```

**Backup volumes:**
```bash
# Postgres
docker exec ae-base-server-postgres pg_dump -U postgres ae_base_server > backup.sql

# MinIO (using mc)
mc mirror local/invoices ./backup/invoices

# Redis
docker exec ae-redis redis-cli -a redis123 SAVE
docker cp ae-redis:/data/dump.rdb ./backup/
```

## Network

All services are connected via the `ae-network` bridge network, allowing inter-service communication using container names:

- From app: `postgres:5432`, `minio:9000`
- From host: `localhost:5432`, `localhost:9000

```env
# PostgreSQL
DATABASE_URL=postgresql://postgres:pass@localhost:5432/ae_base_server
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=pass
DB_NAME=ae_base_server

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_REGION=us-east-1

# Redis
REDIS_URL=redis://:redis123@localhost:6379/0
REDIS_PASSWORD=redis123
REDIS_DB=0
```

## Troubleshooting
```

## Troubleshooting

### Port Already in Use
If ports 5432top conflicting service
brew services stop postgresql
brew services stop redis
```

### MinIO Console Not Accessible
Check if container is running and ports are mapped:
```bash
docker ps9000

# Kill process or stop conflicting service
brew services stop postgresqlpassword `redis123`:
```bash
redis-cli -a redis123 ping
# Should return: PONG
```

### PostgreSQL Connection Refused
Wait for health check to pass:
```bash
docker-compose logs postgres
# Look for: "database system is ready to accept connections"
```
stall MinIO client (if not installed)
brew install minio/stable/mc

# Configure alias
mc alias set local http://localhost:9000 minioadmin minioadmin123

# Create buckets
mc mb local/invoices
mc mb local/templates
mc mb local/documents

# Set public policy for templates (optional)
mc anonymous set download local/templates
```

2. **Run Database Migrations:**
```bash
cd /Users/alex/src/ae/backend/unburdy_server
go run main.go migrate
```

3. **Verify Redis:**
```bash
redis-cli -a redis123 ping
# Expected: PONG
```

## Production Differences

This dev environment differs from production:

- **Passwords**: Use strong, unique passwords in production
- **SSL/TLS**: Enable SSL for MinIO and PostgreSQL
- **Resources**: No resource limits set (add in production)
- **Backups**: Automated backup strategy required
- **Monitoring**: Add Prometheus, Grafana, or similar
- **Security**: Use secrets management, network policies

## Clean Slate

# Remove images (optional)
docker rmi postgres:15-alpine minio/minio:latest redis:7-alpine

# Start fresh
docker-compose up -d
```

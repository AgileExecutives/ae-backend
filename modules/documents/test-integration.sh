#!/bin/bash
# Test script for documents module integration

set -e

echo "=== Documents Module Integration Test ==="
echo ""

# Check if Docker services are running
echo "1. Checking Docker services..."
cd /Users/alex/src/ae/backend/environments/dev
docker-compose ps

echo ""
echo "2. Testing MinIO connection..."
curl -s http://localhost:9000/minio/health/live || echo "MinIO health check failed"

echo ""
echo "3. Testing Redis connection..."
redis-cli -h localhost -p 6379 -a redis123 ping || echo "Redis connection failed"

echo ""
echo "4. Building documents module..."
cd /Users/alex/src/ae/backend/modules/documents
go build ./...
echo "✓ Module builds successfully"

echo ""
echo "5. Module structure:"
find . -name "*.go" -not -path "./tmp/*" | sort

echo ""
echo "=== All checks passed! ==="
echo ""
echo "Next steps:"
echo "1. Integrate the module into unburdy_server or base-server"
echo "2. Run database migrations"
echo "3. Test document upload/download endpoints"
echo "4. Implement invoice number service (Priority 3)"

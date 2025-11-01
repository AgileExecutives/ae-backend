#!/bin/bash

# Script to copy dev database content to staging environment
# This creates a complete copy of the development database to staging

set -e  # Exit on any error

# Configuration
DEV_CONTAINER="unburdy-db-dev"
STAGING_CONTAINER="unburdy-db-staging" 
DEV_DB="unburdy_dev"
STAGING_DB="unburdy_staging"
DB_USER="unburdy_user"
BACKUP_FILE="/tmp/dev_backup_$(date +%Y%m%d_%H%M%S).sql"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Copying Development Database to Staging Environment${NC}"
echo "=================================================="

# Function to check if container is running
check_container() {
    local container_name=$1
    if ! docker ps --format "table {{.Names}}" | grep -q "^${container_name}$"; then
        echo -e "${RED}❌ Error: Container ${container_name} is not running${NC}"
        echo -e "${YELLOW}💡 Please start the container first:${NC}"
        if [[ $container_name == *"staging"* ]]; then
            echo "   docker-compose -f docker-compose-staging.yml up -d"
        else
            echo "   docker-compose up -d"
        fi
        exit 1
    fi
    echo -e "${GREEN}✅ Container ${container_name} is running${NC}"
}

# Function to wait for database to be ready
wait_for_db() {
    local container_name=$1
    local db_name=$2
    echo -e "${YELLOW}⏳ Waiting for database ${db_name} to be ready...${NC}"
    
    for i in {1..30}; do
        if docker exec $container_name pg_isready -U $DB_USER -d $db_name >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Database ${db_name} is ready${NC}"
            return 0
        fi
        sleep 2
    done
    
    echo -e "${RED}❌ Database ${db_name} is not ready after 60 seconds${NC}"
    exit 1
}

# Check if containers are running
echo -e "${BLUE}1. Checking containers...${NC}"
check_container $DEV_CONTAINER
check_container $STAGING_CONTAINER

# Wait for databases to be ready
wait_for_db $DEV_CONTAINER $DEV_DB
wait_for_db $STAGING_CONTAINER $STAGING_DB

# Create backup of development database
echo -e "${BLUE}2. Creating backup of development database...${NC}"
docker exec $DEV_CONTAINER pg_dump -U $DB_USER -d $DEV_DB --clean --if-exists --create > $BACKUP_FILE

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Development database backup created: $BACKUP_FILE${NC}"
    backup_size=$(ls -lh $BACKUP_FILE | awk '{print $5}')
    echo -e "${BLUE}📊 Backup size: $backup_size${NC}"
else
    echo -e "${RED}❌ Failed to create development database backup${NC}"
    exit 1
fi

# Ask for confirmation before proceeding
echo -e "\n${YELLOW}⚠️  WARNING: This will completely replace the staging database content!${NC}"
echo -e "${YELLOW}   Current staging database will be dropped and recreated with dev data.${NC}"
read -p "Do you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}❌ Operation cancelled${NC}"
    rm -f $BACKUP_FILE
    exit 0
fi

# Stop staging server to prevent connections during restore
echo -e "${BLUE}3. Stopping staging server...${NC}"
docker-compose -f docker-compose-staging.yml stop unburdy-server-staging || true

# Restore backup to staging database
echo -e "${BLUE}4. Restoring backup to staging database...${NC}"
docker exec -i $STAGING_CONTAINER psql -U $DB_USER -d postgres < $BACKUP_FILE

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Development database successfully copied to staging${NC}"
else
    echo -e "${RED}❌ Failed to restore backup to staging database${NC}"
    exit 1
fi

# Update staging-specific configurations
echo -e "${BLUE}5. Updating staging-specific configurations...${NC}"
docker exec $STAGING_CONTAINER psql -U $DB_USER -d $STAGING_DB -c "
    -- Update any environment-specific settings
    UPDATE users SET email = REPLACE(email, '@unburdy.de', '@staging.unburdy.de') WHERE email LIKE '%@unburdy.de';
    
    -- Reset password reset tokens
    UPDATE users SET password_reset_token = NULL, password_reset_expires = NULL;
    
    -- Clear any production-specific data that shouldn't be in staging
    -- Add more updates as needed for your specific use case
"

echo -e "${GREEN}✅ Staging-specific configurations updated${NC}"

# Restart staging server
echo -e "${BLUE}6. Starting staging server...${NC}"
docker-compose -f docker-compose-staging.yml up -d unburdy-server-staging

# Wait for staging server to be ready
echo -e "${YELLOW}⏳ Waiting for staging server to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:8080/api/v1/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Staging server is ready${NC}"
        break
    fi
    sleep 2
done

# Clean up backup file
rm -f $BACKUP_FILE
echo -e "${GREEN}🧹 Backup file cleaned up${NC}"

# Success message
echo -e "\n${GREEN}🎉 SUCCESS: Development database copied to staging!${NC}"
echo "=================================================="
echo -e "${BLUE}📋 Next steps:${NC}"
echo -e "   • Staging server: http://localhost:8080"
echo -e "   • Staging database: localhost:5433"
echo -e "   • pgAdmin (if enabled): http://localhost:5050"
echo -e "\n${YELLOW}💡 Note: All users now have '@staging.unburdy.de' email addresses${NC}"
echo -e "${YELLOW}   and password reset tokens have been cleared for security.${NC}"
#!/bin/bash

# Generic Docker Volume Duplication Script
# Usage: ./scripts/duplicate-volume.sh SOURCE_VOLUME TARGET_VOLUME

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check arguments
if [ $# -ne 2 ]; then
    echo -e "${RED}❌ Usage: $0 SOURCE_VOLUME TARGET_VOLUME${NC}"
    echo ""
    echo "Examples:"
    echo "  $0 unburdy_dev_data unburdy_staging_data"
    echo "  $0 my_app_volume my_app_backup_volume"
    exit 1
fi

SOURCE_VOLUME="$1"
TARGET_VOLUME="$2"

echo -e "${BLUE}🔄 Docker Volume Duplication${NC}"
echo "=============================="
echo -e "${YELLOW}Source: $SOURCE_VOLUME${NC}"
echo -e "${YELLOW}Target: $TARGET_VOLUME${NC}"
echo ""

# Check if source volume exists
if ! docker volume ls --format "{{.Name}}" | grep -q "^${SOURCE_VOLUME}$"; then
    echo -e "${RED}❌ Source volume '$SOURCE_VOLUME' does not exist${NC}"
    echo ""
    echo "Available volumes:"
    docker volume ls --format "table {{.Name}}\t{{.Driver}}\t{{.Size}}"
    exit 1
fi

# Check if target volume already exists
if docker volume ls --format "{{.Name}}" | grep -q "^${TARGET_VOLUME}$"; then
    echo -e "${YELLOW}⚠️  Target volume '$TARGET_VOLUME' already exists${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}❌ Operation cancelled${NC}"
        exit 0
    fi
    echo -e "${BLUE}🗑️  Removing existing target volume...${NC}"
    docker volume rm "$TARGET_VOLUME"
fi

# Create target volume
echo -e "${BLUE}📦 Creating target volume '$TARGET_VOLUME'...${NC}"
docker volume create "$TARGET_VOLUME"

# Copy data using alpine container with tar pipe
echo -e "${BLUE}📋 Copying data from '$SOURCE_VOLUME' to '$TARGET_VOLUME'...${NC}"
docker run --rm \
    -v "$SOURCE_VOLUME":/source:ro \
    -v "$TARGET_VOLUME":/target \
    alpine ash -c "cd /source && tar cf - . | (cd /target && tar xf -)"

# Verify the copy
echo -e "${BLUE}🔍 Verifying copy...${NC}"

# Get file counts from both volumes
SOURCE_COUNT=$(docker run --rm -v "$SOURCE_VOLUME":/data alpine find /data -type f | wc -l)
TARGET_COUNT=$(docker run --rm -v "$TARGET_VOLUME":/data alpine find /data -type f | wc -l)

if [ "$SOURCE_COUNT" -eq "$TARGET_COUNT" ]; then
    echo -e "${GREEN}✅ Volume duplication successful!${NC}"
    echo -e "${GREEN}📊 Copied $SOURCE_COUNT files${NC}"
    
    # Show size information
    echo -e "\n${BLUE}📏 Volume sizes:${NC}"
    echo "Source volume:"
    docker run --rm -v "$SOURCE_VOLUME":/data alpine du -sh /data
    echo "Target volume:"
    docker run --rm -v "$TARGET_VOLUME":/data alpine du -sh /data
else
    echo -e "${RED}❌ File count mismatch!${NC}"
    echo -e "${RED}Source: $SOURCE_COUNT files, Target: $TARGET_COUNT files${NC}"
    exit 1
fi

echo -e "\n${GREEN}🎉 Volume '$SOURCE_VOLUME' successfully duplicated to '$TARGET_VOLUME'${NC}"
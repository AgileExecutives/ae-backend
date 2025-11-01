#!/bin/bash

# generate-tokens.sh - Generate bearer tokens for users from seed data
# This script reads seed-data.json from unburdy_server and generates tokens for each user

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEED_FILE="${SCRIPT_DIR}/unburdy_server/seed-data.json"
HOST="${HOST:-http://localhost:8080}"
LOGIN_ENDPOINT="/api/v1/auth/login"
OUTPUT_FILE="${SCRIPT_DIR}/bearer-tokens.json"

# Check if seed file exists
if [ ! -f "$SEED_FILE" ]; then
    echo -e "${RED}Error: Seed file not found at $SEED_FILE${NC}"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed. Please install jq first.${NC}"
    echo "  macOS: brew install jq"
    echo "  Ubuntu/Debian: sudo apt-get install jq"
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed.${NC}"
    exit 1
fi

# Function to check if server is running
check_server() {
    echo -e "${BLUE}Checking if server is running at $HOST...${NC}"
    if curl -s -f "$HOST/api/v1/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running or not responding at $HOST${NC}"
        echo -e "${YELLOW}Please start the unburdy server first:${NC}"
        echo "  cd unburdy_server && make run"
        exit 1
    fi
}

# Function to generate token for a user
generate_token() {
    local username="$1"
    local password="$2"
    local role="$3"
    local tenant_slug="$4"
    
    echo -e "${BLUE}Generating token for $username ($role @ $tenant_slug)...${NC}" >&2
    
    # Prepare login request
    local login_data=$(cat <<EOF
{
  "username": "$username",
  "password": "$password"
}
EOF
)
    
    # Make login request
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$login_data" \
        "$HOST$LOGIN_ENDPOINT" 2>/dev/null)
    
    # Check if request was successful
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Failed to connect to server for user $username${NC}"
        return 1
    fi
    
    # Parse response
    local success=$(echo "$response" | jq -r '.success // false')
    
    if [ "$success" = "true" ]; then
        local token=$(echo "$response" | jq -r '.data.token')
        local user_id=$(echo "$response" | jq -r '.data.user.id')
        local tenant_id=$(echo "$response" | jq -r '.data.user.tenant_id')
        
        echo -e "${GREEN}✅ Token generated for $username${NC}" >&2
        
        # Return structured data
        cat <<EOF
{
  "username": "$username",
  "email": "$username",
  "role": "$role",
  "tenant_slug": "$tenant_slug",
  "user_id": $user_id,
  "tenant_id": $tenant_id,
  "token": "$token",
  "curl_example": "curl -H 'Authorization: Bearer $token' $HOST/api/v1/health"
}
EOF
    else
        local error_message=$(echo "$response" | jq -r '.message // "Unknown error"')
        echo -e "${RED}❌ Login failed for $username: $error_message${NC}" >&2
        echo -e "${YELLOW}Response: $response${NC}" >&2
        return 1
    fi
}

# Main function
main() {
    echo -e "${BLUE}🔐 Bearer Token Generator${NC}"
    echo -e "${BLUE}=========================${NC}"
    echo
    
    # Check server availability
    check_server
    echo
    
    # Read users from seed file
    echo -e "${BLUE}Reading users from seed file...${NC}"
    local users=$(jq -c '.users[]' "$SEED_FILE")
    
    if [ -z "$users" ]; then
        echo -e "${RED}Error: No users found in seed file${NC}"
        exit 1
    fi
    
    # Initialize output
    echo "{" > "$OUTPUT_FILE"
    echo '  "generated_at": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",' >> "$OUTPUT_FILE"
    echo '  "server_host": "'$HOST'",' >> "$OUTPUT_FILE"
    echo '  "tokens": [' >> "$OUTPUT_FILE"
    
    local first=true
    local success_count=0
    local total_count=0
    
    # Process each user
    while IFS= read -r user; do
        total_count=$((total_count + 1))
        
        local username=$(echo "$user" | jq -r '.email')
        local password=$(echo "$user" | jq -r '.password')
        local role=$(echo "$user" | jq -r '.role')
        local tenant_slug=$(echo "$user" | jq -r '.tenant_slug')
        
        # Add comma separator for JSON array
        if [ "$first" = false ]; then
            echo "," >> "$OUTPUT_FILE"
        fi
        first=false
        
        # Generate token (redirect stderr to avoid color codes in JSON)
        local token_data=$(generate_token "$username" "$password" "$role" "$tenant_slug" 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            success_count=$((success_count + 1))
            # Indent and append to output file
            echo "$token_data" | sed 's/^/    /' >> "$OUTPUT_FILE"
        else
            # Add error entry
            cat <<EOF >> "$OUTPUT_FILE"
    {
      "username": "$username",
      "email": "$username", 
      "role": "$role",
      "tenant_slug": "$tenant_slug",
      "error": "Login failed",
      "token": null
    }
EOF
        fi
        
        echo
    done <<< "$users"
    
    # Close JSON structure
    echo >> "$OUTPUT_FILE"
    echo "  ]" >> "$OUTPUT_FILE"
    echo "}" >> "$OUTPUT_FILE"
    
    # Summary
    echo -e "${BLUE}Summary:${NC}"
    echo -e "  ${GREEN}✅ Successful: $success_count/$total_count${NC}"
    echo -e "  ${BLUE}📄 Output file: $OUTPUT_FILE${NC}"
    echo
    
    # Show generated tokens
    if [ $success_count -gt 0 ]; then
        echo -e "${GREEN}Generated Tokens:${NC}"
        echo "=================="
        
        jq -r '.tokens[] | select(.token != null) | "
\(.username) (\(.role) @ \(.tenant_slug)):
  Token: \(.token)
  User ID: \(.user_id)
  Tenant ID: \(.tenant_id)
  Test: \(.curl_example)
"' "$OUTPUT_FILE"
        
        echo -e "${YELLOW}💡 Tip: You can now use these tokens for API testing!${NC}"
        echo -e "${YELLOW}   Example: export TOKEN=\$(jq -r '.tokens[0].token' $OUTPUT_FILE)${NC}"
        echo -e "${YELLOW}            curl -H \"Authorization: Bearer \$TOKEN\" $HOST/api/v1/health${NC}"
        echo
        
        # Print bearer tokens in requested format
        echo -e "${GREEN}Bearer Tokens:${NC}"
        echo "=============="
        jq -r '.tokens[] | select(.token != null) | "User \(.user_id) Bearer \(.token)"' "$OUTPUT_FILE"
    fi
    
    if [ $success_count -lt $total_count ]; then
        echo -e "${YELLOW}⚠️  Some tokens could not be generated. Check the output above for details.${NC}"
        exit 1
    fi
}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Generate bearer tokens by logging in users from unburdy seed data"
    echo
    echo "OPTIONS:"
    echo "  -h, --help     Show this help message"
    echo "  -H, --host     Set server host (default: http://localhost:8080)"
    echo
    echo "ENVIRONMENT VARIABLES:"
    echo "  HOST           Server host URL (default: http://localhost:8080)"
    echo
    echo "EXAMPLES:"
    echo "  $0                                    # Generate tokens for localhost"
    echo "  $0 -H http://localhost:3000          # Generate tokens for custom host"
    echo "  HOST=http://staging.example.com $0   # Use environment variable"
    echo
    echo "OUTPUT:"
    echo "  Tokens are saved to: bearer-tokens.json"
    echo
    echo "REQUIREMENTS:"
    echo "  - jq (JSON processor)"
    echo "  - curl (HTTP client)"
    echo "  - Running unburdy server"
    echo "  - Seed data file at: unburdy_server/seed-data.json"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -H|--host)
            HOST="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}Error: Unknown option $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main
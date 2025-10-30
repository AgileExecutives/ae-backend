# Enhanced HURL Test Runner - Environment Detection Functions
# Add these functions to your run-hurl-tests.sh script

# Function to get server environment configuration
get_server_config() {
    local health_response
    echo -e "${YELLOW}🔧 Detecting server configuration...${NC}"
    
    health_response=$(curl -s --max-time 5 "${HOST}/api/v1/health" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$health_response" ]; then
        # Extract environment variables
        MOCK_EMAIL=$(echo "$health_response" | jq -r '.environment.mock_email // false')
        RATE_LIMIT_ENABLED=$(echo "$health_response" | jq -r '.environment.rate_limit_enabled // true')
        EMAIL_VERIFICATION=$(echo "$health_response" | jq -r '.environment.email_verification // true')
        GIN_MODE=$(echo "$health_response" | jq -r '.environment.gin_mode // "debug"')
        
        echo -e "${BLUE}📋 Server Configuration:${NC}"
        echo -e "${BLUE}  Mock Email: ${MOCK_EMAIL}${NC}"
        echo -e "${BLUE}  Rate Limiting: ${RATE_LIMIT_ENABLED}${NC}"
        echo -e "${BLUE}  Email Verification: ${EMAIL_VERIFICATION}${NC}"
        echo -e "${BLUE}  Mode: ${GIN_MODE}${NC}"
        
        return 0
    else
        echo -e "${RED}❌ Failed to get server configuration${NC}"
        # Set safe defaults
        MOCK_EMAIL="false"
        RATE_LIMIT_ENABLED="true"
        EMAIL_VERIFICATION="true"
        GIN_MODE="debug"
        return 1
    fi
}

# Function to adapt test execution based on server config
adapt_test_execution() {
    echo -e "${YELLOW}🎯 Adapting test execution...${NC}"
    
    # Rate limiting adaptations
    if [ "$RATE_LIMIT_ENABLED" = "true" ]; then
        echo -e "${YELLOW}⚠️  Rate limiting enabled - adding delays between tests${NC}"
        TEST_DELAY=2  # seconds between tests
        AUTH_DELAY=5  # extra delay for auth tests
    else
        echo -e "${GREEN}✅ Rate limiting disabled - full speed testing${NC}"
        TEST_DELAY=0
        AUTH_DELAY=0
    fi
    
    # Email testing adaptations
    if [ "$MOCK_EMAIL" = "true" ]; then
        echo -e "${GREEN}✅ Mock email enabled - safe for email testing${NC}"
        SKIP_EMAIL_TESTS=false
    else
        echo -e "${YELLOW}⚠️  Mock email disabled - using test email addresses${NC}"
        SKIP_EMAIL_TESTS=false
        # Could set SKIP_EMAIL_TESTS=true to skip email tests in production
    fi
    
    # Email verification adaptations
    if [ "$EMAIL_VERIFICATION" = "true" ]; then
        echo -e "${YELLOW}⚠️  Email verification required - may affect user tests${NC}"
        SKIP_UNVERIFIED_TESTS=true
    else
        echo -e "${GREEN}✅ Email verification disabled - all users active${NC}"
        SKIP_UNVERIFIED_TESTS=false
    fi
}

# Enhanced run_test function with environment-aware delays
run_test_with_config() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .hurl)
    
    # Apply delays based on test type and server config
    if [[ "$test_name" == *"auth"* ]] && [ "$AUTH_DELAY" -gt 0 ]; then
        echo -e "${YELLOW}⏱️  Applying auth delay (${AUTH_DELAY}s) for rate limiting...${NC}"
        sleep "$AUTH_DELAY"
    elif [ "$TEST_DELAY" -gt 0 ]; then
        sleep "$TEST_DELAY"
    fi
    
    # Skip tests based on configuration
    if [ "$SKIP_EMAIL_TESTS" = "true" ] && [[ "$test_name" == *"email"* ]]; then
        echo -e "${YELLOW}⏭️  Skipping ${test_name}.hurl (email testing disabled)${NC}"
        return 0
    fi
    
    # Run the actual test (your existing run_test function logic)
    echo -e "${BLUE}🧪 Running ${test_name}.hurl...${NC}"
    
    # ... rest of your test execution logic ...
}

# Example usage in your main script:
# 1. Check server and get config
# check_server
# get_server_config
# adapt_test_execution

# 2. Use environment-aware test runner
# for template in "$TEMPLATES_DIR"/*.hurl; do
#     if [ -f "$template" ]; then
#         if run_test_with_config "$template"; then
#             ((passed_tests++))
#         else
#             ((failed_tests++))
#         fi
#     fi
# done

echo "📝 Integration Example:"
echo "====================="
echo "Add these functions to your run-hurl-tests.sh to:"
echo "1. Detect server configuration automatically"
echo "2. Adapt test execution based on environment settings"  
echo "3. Add appropriate delays for rate limiting"
echo "4. Skip tests that aren't compatible with current config"
echo "5. Provide better feedback about test environment"
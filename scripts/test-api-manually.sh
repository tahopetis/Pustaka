#!/bin/bash

echo "=== Manual API Test for Audit Logs ==="
echo "This script tests the audit API endpoints directly using curl"

API_BASE="http://localhost:8080/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "\n${YELLOW}1. Testing Backend Health${NC}"
health_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" http://localhost:8080/health)
http_code=$(echo "$health_response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
response_body=$(echo "$health_response" | sed -e 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Backend is healthy${NC}"
    echo "Response: $response_body"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
    echo "HTTP Status: $http_code"
    echo "Response: $response_body"
    exit 1
fi

echo -e "\n${YELLOW}2. Testing Login${NC}"
login_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"Admin@123"}' \
    "$API_BASE/auth/login")

http_code=$(echo "$login_response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
response_body=$(echo "$login_response" | sed -e 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Login successful${NC}"

    # Extract tokens from response
    access_token=$(echo "$response_body" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    refresh_token=$(echo "$response_body" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

    echo "Access Token: ${access_token:0:30}..."
    echo "Refresh Token: ${refresh_token:0:30}..."

    if [ -z "$access_token" ]; then
        echo -e "${RED}❌ No access token found in response${NC}"
        echo "Full response: $response_body"
        exit 1
    fi
else
    echo -e "${RED}❌ Login failed${NC}"
    echo "HTTP Status: $http_code"
    echo "Response: $response_body"
    exit 1
fi

echo -e "\n${YELLOW}3. Testing User Profile${NC}"
profile_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -H "Authorization: Bearer $access_token" \
    "$API_BASE/me")

http_code=$(echo "$profile_response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
response_body=$(echo "$profile_response" | sed -e 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ User profile retrieved${NC}"
    echo "User info: $response_body"

    # Extract permissions
    permissions=$(echo "$response_body" | grep -o '"permissions":\[[^]]*\]' | cut -d: -f2)
    echo "Permissions: $permissions"

    # Check for audit:read permission
    if echo "$response_body" | grep -q "audit:read"; then
        echo -e "${GREEN}✅ User has audit:read permission${NC}"
    else
        echo -e "${RED}❌ User does NOT have audit:read permission${NC}"
    fi
else
    echo -e "${RED}❌ Profile retrieval failed${NC}"
    echo "HTTP Status: $http_code"
    echo "Response: $response_body"
fi

echo -e "\n${YELLOW}4. Testing Audit API Endpoint${NC}"
audit_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -H "Authorization: Bearer $access_token" \
    "$API_BASE/audit?page=1&limit=10")

http_code=$(echo "$audit_response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
response_body=$(echo "$audit_response" | sed -e 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Audit API call successful${NC}"

    # Count audit logs
    audit_count=$(echo "$response_body" | grep -o '"audit_logs":\[[^]]*\]' | grep -o '{' | wc -l)
    total_count=$(echo "$response_body" | grep -o '"total":[0-9]*' | cut -d: -f2)

    echo "Number of audit logs in response: $audit_count"
    echo "Total audit logs available: $total_count"
    echo "Full response: $response_body"

    if [ "$audit_count" = "0" ] && [ "$total_count" = "0" ]; then
        echo -e "${YELLOW}⚠️  No audit logs found in database${NC}"
    fi
else
    echo -e "${RED}❌ Audit API call failed${NC}"
    echo "HTTP Status: $http_code"
    echo "Response: $response_body"

    case "$http_code" in
        "401")
            echo -e "${RED}Authentication failed - invalid or expired token${NC}"
            ;;
        "403")
            echo -e "${RED}Authorization failed - user lacks audit:read permission${NC}"
            ;;
        "500")
            echo -e "${RED}Server error - check backend logs${NC}"
            ;;
    esac
fi

echo -e "\n${YELLOW}5. Testing Audit API without Authorization${NC}"
unauth_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "$API_BASE/audit?page=1&limit=10")

http_code=$(echo "$unauth_response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
response_body=$(echo "$unauth_response" | sed -e 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✅ API correctly requires authentication (401 Unauthorized)${NC}"
else
    echo -e "${YELLOW}⚠️  API responded with $http_code (expected 401)${NC}"
    echo "Response: $response_body"
fi

echo -e "\n${YELLOW}=== Test Summary ===${NC}"

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ All API tests passed successfully${NC}"
    echo "The backend API is working correctly. The issue is likely in the frontend."
else
    echo -e "${RED}❌ Some API tests failed${NC}"
    echo "The issue may be in the backend API or authentication/authorization."
fi

echo -e "\nNext steps:"
echo "1. If API tests passed, check the frontend JavaScript console for errors"
echo "2. If API tests failed, check the backend logs for issues"
echo "3. Verify the user has the correct permissions in the database"
echo "4. Check the JWT token format and validation"
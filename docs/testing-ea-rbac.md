# EA RBAC Testing Guide

This document provides manual testing steps for verifying Role-Based Access Control (RBAC) enforcement on EA (Enterprise Architecture) entity endpoints.

## Overview

The EA module implements RBAC using four permissions:
- `ea:read` - View EA entities
- `ea:create` - Create new EA entities
- `ea:update` - Modify existing EA entities
- `ea:delete` - Delete EA entities

## Role Permissions Matrix

| Role     | ea:read | ea:create | ea:update | ea:delete |
|----------|---------|-----------|-----------|-----------|
| Admin    | ✓       | ✓         | ✓         | ✓         |
| Editor   | ✓       | ✓         | ✓         | ✗         |
| Viewer   | ✓       | ✗         | ✗         | ✗         |

## Prerequisites

1. Running Pustaka API server (http://localhost:8080)
2. PostgreSQL database with EA permissions seeded (migration 010_ea_permissions.sql)
3. Test users with different roles:
   - Admin user (has all EA permissions)
   - Editor user (has read, create, update permissions)
   - Viewer user (has only read permission)

## Setup Test Users

If test users don't exist, create them via the API:

### Create Admin User

```bash
# Login as default admin to get token
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}' \
  | jq -r '.access_token')

# Create viewer user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "viewer",
    "email": "viewer@example.com",
    "password": "Viewer@123",
    "roles": ["viewer"]
  }'

# Create editor user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "editor",
    "email": "editor@example.com",
    "password": "Editor@123",
    "roles": ["editor"]
  }'
```

## Test Cases

### Test 1: Viewer Role (Read-Only Access)

#### 1.1 Login as Viewer

```bash
VIEWER_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer","password":"Viewer@123"}' \
  | jq -r '.access_token')

echo "Viewer token: $VIEWER_TOKEN"
```

#### 1.2 List EA Entities (Should Succeed - 200)

```bash
curl -X GET http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 1.3 Get Single EA Entity (Should Succeed - 200)

```bash
# First get an entity ID from list, then:
curl -X GET "http://localhost:8080/api/v1/ea/entities/{entity_id}" \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 1.4 Create EA Entity (Should Fail - 403)

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Application",
    "ci_type": "EA.Application-BusinessApp",
    "owner": "Platform Team",
    "attributes": {
      "description": "Test application",
      "criticality": "high"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 403 Forbidden

#### 1.5 Update EA Entity (Should Fail - 403)

```bash
curl -X PUT "http://localhost:8080/api/v1/ea/entities/{entity_id}" \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Application Name"
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 403 Forbidden

#### 1.6 Delete EA Entity (Should Fail - 403)

```bash
curl -X DELETE "http://localhost:8080/api/v1/ea/entities/{entity_id}" \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 403 Forbidden

---

### Test 2: Editor Role (Read, Create, Update Access)

#### 2.1 Login as Editor

```bash
EDITOR_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"editor","password":"Editor@123"}' \
  | jq -r '.access_token')

echo "Editor token: $EDITOR_TOKEN"
```

#### 2.2 List EA Entities (Should Succeed - 200)

```bash
curl -X GET http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 2.3 Create EA Entity (Should Succeed - 201)

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Editor Test Application",
    "ci_type": "EA.Application-BusinessApp",
    "owner": "Platform Team",
    "attributes": {
      "description": "Created by editor",
      "criticality": "medium"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 201 Created

#### 2.4 Update EA Entity (Should Succeed - 200)

```bash
# Use the entity ID from the create response
ENTITY_ID="<entity_id_from_create_response>"

curl -X PUT "http://localhost:8080/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated by Editor"
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 2.5 Delete EA Entity (Should Fail - 403)

```bash
curl -X DELETE "http://localhost:8080/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 403 Forbidden

---

### Test 3: Admin Role (Full Access)

#### 3.1 Login as Admin

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}' \
  | jq -r '.access_token')

echo "Admin token: $ADMIN_TOKEN"
```

#### 3.2 List EA Entities (Should Succeed - 200)

```bash
curl -X GET http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 3.3 Create EA Entity (Should Succeed - 201)

```bash
curl -X POST http://localhost:8080/api/v1/ea/entities \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Test Application",
    "ci_type": "EA.Application-BusinessApp",
    "owner": "Platform Team",
    "attributes": {
      "description": "Created by admin",
      "criticality": "high"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 201 Created

#### 3.4 Update EA Entity (Should Succeed - 200)

```bash
ENTITY_ID="<entity_id_from_create_response>"

curl -X PUT "http://localhost:8080/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated by Admin"
  }' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

#### 3.5 Delete EA Entity (Should Succeed - 204)

```bash
curl -X DELETE "http://localhost:8080/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 204 No Content

---

### Test 4: Validate Endpoint Permission Check

#### 4.1 Viewer Can Access Validation (Should Succeed - 200)

```bash
curl -X GET "http://localhost:8080/api/v1/ea/entities/{entity_id}/validate" \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result:** HTTP Status: 200

---

## Quick Test Script

Save this as `test_ea_rbac.sh` and run it:

```bash
#!/bin/bash

API_URL="http://localhost:8080"

echo "=== EA RBAC Testing ==="
echo ""

# Get tokens
echo "Getting user tokens..."
VIEWER_TOKEN=$(curl -s -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer","password":"Viewer@123"}' \
  | jq -r '.access_token')

EDITOR_TOKEN=$(curl -s -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"editor","password":"Editor@123"}' \
  | jq -r '.access_token')

ADMIN_TOKEN=$(curl -s -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}' \
  | jq -r '.access_token')

# Test 1: Viewer - Read Only
echo "=== Test 1: Viewer (Read Only) ==="
echo "1. List entities (expect 200):"
curl -s -X GET $API_URL/api/v1/ea/entities \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -w "\nStatus: %{http_code}\n" | tail -1

echo "2. Create entity (expect 403):"
curl -s -X POST $API_URL/api/v1/ea/entities \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","ci_type":"EA.Application-BusinessApp","owner":"Team"}' \
  -w "\nStatus: %{http_code}\n" | tail -1

# Test 2: Editor - Create/Update
echo ""
echo "=== Test 2: Editor (Create/Update) ==="
echo "1. Create entity (expect 201):"
RESPONSE=$(curl -s -X POST $API_URL/api/v1/ea/entities \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Editor Test","ci_type":"EA.Application-BusinessApp","owner":"Team"}' \
  -w "\n%{http_code}")
ENTITY_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Status: $(echo $RESPONSE | tail -1)"

echo "2. Delete entity (expect 403):"
curl -s -X DELETE "$API_URL/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -w "\nStatus: %{http_code}\n" | tail -1

# Test 3: Admin - Full Access
echo ""
echo "=== Test 3: Admin (Full Access) ==="
echo "1. Create entity (expect 201):"
RESPONSE=$(curl -s -X POST $API_URL/api/v1/ea/entities \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin Test","ci_type":"EA.Application-BusinessApp","owner":"Team"}' \
  -w "\n%{http_code}")
ENTITY_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Status: $(echo $RESPONSE | tail -1)"

echo "2. Delete entity (expect 204):"
curl -s -X DELETE "$API_URL/api/v1/ea/entities/$ENTITY_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -w "\nStatus: %{http_code}\n" | tail -1

echo ""
echo "=== Testing Complete ==="
```

Run the script:

```bash
chmod +x test_ea_rbac.sh
./test_ea_rbac.sh
```

## Verification Database Queries

After testing, verify permissions are correctly assigned:

```sql
-- Check EA permissions exist
SELECT name, description FROM permissions WHERE name LIKE 'ea:%'
ORDER BY name;

-- Check role permissions
SELECT r.name as role_name, p.name as permission_name
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE p.name LIKE 'ea:%'
ORDER BY r.name, p.name;

-- Expected results:
-- admin:  ea:create, ea:delete, ea:read, ea:update (4 rows)
-- editor: ea:create, ea:read, ea:update (3 rows)
-- viewer: ea:read (1 row)
```

## Troubleshooting

### Issue: 401 Unauthorized

**Cause:** Invalid or expired JWT token

**Solution:**
- Ensure you're logged in and have a valid token
- Check token expiration (24 hours for access tokens)
- Re-login to get a fresh token

### Issue: 403 Forbidden on Authorized Operation

**Cause:** User role doesn't have the required permission

**Solution:**
- Verify user's role assignment in database
- Check role_permissions table has correct grants
- Ensure migration 010_ea_permissions.sql was applied

### Issue: Permission Not Found

**Cause:** EA permissions not seeded in database

**Solution:**
- Run migration: `docker compose exec -T postgres psql -U pustaka -d pustaka < cmd/migrations/010_ea_permissions.sql`
- Verify permissions exist: `SELECT * FROM permissions WHERE name LIKE 'ea:%';`

## Summary

This testing guide verifies:
- ✓ Viewer role can only read (GET requests)
- ✓ Editor role can read, create, and update (GET, POST, PUT)
- ✓ Admin role has full access (GET, POST, PUT, DELETE)
- ✓ All unauthorized operations return 403 Forbidden
- ✓ Middleware chain correctly enforces RBAC before handler execution

# Pustaka API Guide

This guide provides comprehensive documentation for using the Pustaka CMDB API, including authentication, usage examples, and best practices.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Rate Limiting](#rate-limiting)
4. [Error Handling](#error-handling)
5. [Pagination](#pagination)
6. [Search and Filtering](#search-and-filtering)
7. [Working with CI Types](#working-with-ci-types)
8. [Configuration Items](#configuration-items)
9. [Relationships](#relationships)
10. [Graph API](#graph-api)
11. [Audit Logs](#audit-logs)
12. [SDKs and Tools](#sdks-and-tools)
13. [Best Practices](#best-practices)

## Getting Started

### Base URL
```
Production: https://api.pustaka.dev/v1
Staging: https://staging-api.pustaka.dev/v1
Development: http://localhost:8080/api/v1
```

### Quick Start Example

```bash
# 1. Authenticate
curl -X POST https://api.pustaka.dev/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@pustaka.dev",
    "password": "your-password"
  }'

# Response contains access_token and refresh_token

# 2. Use the access token to make authenticated requests
curl -X GET https://api.pustaka.dev/v1/ci \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

## Authentication

Pustaka uses JWT (JSON Web Token) authentication. You'll receive two tokens when logging in:

- **Access Token**: Short-lived (15 minutes by default) token for API requests
- **Refresh Token**: Long-lived (7 days by default) token for obtaining new access tokens

### Login Endpoint

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

### Response

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "is_active": true
  }
}
```

### Using the Access Token

Include the access token in the Authorization header for all authenticated requests:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Refreshing Tokens

When your access token expires, use the refresh token to get a new one:

```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Default limit**: 1000 requests per hour per user
- **Rate limit headers** are included in responses:
  - `X-RateLimit-Limit`: Total requests allowed per hour
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Unix timestamp when rate limit resets

### Rate Limit Response (429)

```json
{
  "error": "Rate limit exceeded",
  "details": {
    "limit": 1000,
    "remaining": 0,
    "reset_time": 1640995200
  }
}
```

## Error Handling

The API uses standard HTTP status codes and provides detailed error information:

### Standard Error Response Format

```json
{
  "error": "Human-readable error message",
  "details": {
    "field": "Additional error details",
    "code": "ERROR_CODE"
  }
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or invalid
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists
- `422 Unprocessable Entity`: Validation failed
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Pagination

List endpoints support pagination using the following parameters:

### Parameters

- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)

### Example

```http
GET /ci?page=2&limit=50
```

### Response Format

```json
{
  "cis": [...],
  "page": 2,
  "limit": 50,
  "total": 150,
  "total_pages": 3
}
```

## Search and Filtering

### Basic Search

Most list endpoints support a `search` parameter for basic text search:

```http
GET /ci?search=web-server
```

### Advanced Filtering

#### CI List Filtering

```http
GET /ci?ci_type=Server&tags=production,critical&sort=created_at&order=desc
```

#### Attribute-based Search

For searching within CI attributes, use the `attributes` parameter with JSON:

```http
GET /ci?attributes={"cpu_cores":{"min":8,"max":16},"os_type":"Ubuntu"}
```

#### Audit Log Filtering

```http
GET /audit/logs?action=create&resource_type=ci&start_date=2023-01-01T00:00:00Z&end_date=2023-01-31T23:59:59Z
```

## Working with CI Types

CI Types define the schema for Configuration Items, including required and optional attributes with validation rules.

### Creating a CI Type

```http
POST /ci-types
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "name": "Server",
  "description": "Physical or virtual server",
  "required_attributes": [
    {
      "name": "hostname",
      "type": "string",
      "description": "Server hostname",
      "validation": {
        "min_length": 1,
        "max_length": 255,
        "pattern": "^[a-zA-Z0-9.-]+$"
      }
    },
    {
      "name": "ip_address",
      "type": "string",
      "description": "Primary IP address",
      "validation": {
        "format": "ipv4"
      }
    }
  ],
  "optional_attributes": [
    {
      "name": "cpu_cores",
      "type": "integer",
      "description": "Number of CPU cores",
      "validation": {
        "min": 1,
        "max": 128
      }
    },
    {
      "name": "environment",
      "type": "string",
      "description": "Environment type",
      "validation": {
        "enum": ["development", "staging", "production"]
      }
    }
  ]
}
```

### Attribute Validation Types

#### String Attributes
- `pattern`: Regular expression pattern
- `min_length`, `max_length`: Length constraints
- `format`: Predefined formats (email, url, ipv4, date, datetime)
- `enum`: List of allowed values

#### Integer Attributes
- `min`, `max`: Value range constraints

#### Boolean Attributes
- Simple true/false values

#### Array Attributes
- JSON arrays with optional length validation

#### Object Attributes
- JSON objects with optional property count validation

## Configuration Items

Configuration Items (CIs) are instances of CI Types with specific attribute values.

### Creating a CI

```http
POST /ci
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "name": "web-server-01",
  "ci_type": "Server",
  "attributes": {
    "hostname": "web01.example.com",
    "ip_address": "192.168.1.10",
    "cpu_cores": 8,
    "memory_gb": 32,
    "environment": "production"
  },
  "tags": ["production", "web", "critical"]
}
```

### Updating a CI

```http
PUT /ci/550e8400-e29b-41d4-a716-446655440002
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "attributes": {
    "cpu_cores": 16,
    "memory_gb": 64
  },
  "tags": ["production", "web", "critical", "upgraded"]
}
```

### Advanced CI Search

```http
GET /ci?ci_type=Server&attributes={"cpu_cores":{"min":8},"environment":"production"}
```

## Relationships

Relationships define connections between Configuration Items.

### Creating a Relationship

```http
POST /relationships
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "source_id": "550e8400-e29b-41d4-a716-446655440002",
  "target_id": "550e8400-e29b-41d4-a716-446655440003",
  "relationship_type": "depends_on",
  "description": "Web server depends on database server"
}
```

### Relationship Types

Common relationship types include:
- `depends_on`: One CI depends on another
- `hosts`: One CI hosts another
- `connects_to`: Network connectivity
- `manages`: Management relationship
- `contains`: Physical containment
- Custom types as needed

## Graph API

The Graph API provides network visualization data for understanding CI relationships.

### Getting Full Graph Data

```http
GET /graph?limit=100
Authorization: Bearer YOUR_TOKEN
```

### Interactive Graph Exploration

```http
GET /graph/explore?start_id=550e8400-e29b-41d4-a716-446655440002&depth=3&relationship_types=depends_on,hosts
Authorization: Bearer YOUR_TOKEN
```

### Graph Data Format

```json
{
  "nodes": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "label": "web-server-01",
      "type": "Server",
      "group": "production",
      "metadata": {
        "cpu_cores": 8,
        "environment": "production"
      }
    }
  ],
  "edges": [
    {
      "id": "rel-001",
      "source": "550e8400-e29b-41d4-a716-446655440002",
      "target": "550e8400-e29b-41d4-a716-446655440003",
      "label": "depends_on",
      "type": "depends_on"
    }
  ],
  "metadata": {
    "total_nodes": 25,
    "total_edges": 40,
    "layout": "force"
  }
}
```

## Audit Logs

Audit logs provide a complete history of all changes made in the system.

### Retrieving Audit Logs

```http
GET /audit/logs?page=1&limit=50&action=create&resource_type=ci&start_date=2023-01-01T00:00:00Z
Authorization: Bearer YOUR_TOKEN
```

### Audit Log Entry

```json
{
  "id": "audit-001",
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "action": "create",
  "resource_type": "ci",
  "resource_id": "550e8400-e29b-41d4-a716-446655440002",
  "details": {
    "name": "web-server-01",
    "ci_type": "Server",
    "changed_attributes": ["hostname", "ip_address"]
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "created_at": "2023-01-01T12:00:00Z"
}
```

### Audit Statistics

```http
GET /audit/stats?period=month
Authorization: Bearer YOUR_TOKEN
```

### Exporting Audit Logs

```http
GET /audit/export?start_date=2023-01-01T00:00:00Z&end_date=2023-01-31T23:59:59Z&format=csv
Authorization: Bearer YOUR_TOKEN
```

## SDKs and Tools

### Official SDKs

- **JavaScript/TypeScript**: `npm install @pustaka/api-client`
- **Python**: `pip install pustaka-sdk`
- **Go**: `go get github.com/pustaka/pustaka-go`

### Third-Party Tools

- **Terraform Provider**: Manage CIs as infrastructure
- **Ansible Modules**: Automate CI management
- **Grafana Plugin**: Visualize CMDB data

### Example: JavaScript SDK

```javascript
import { PustakaAPI } from '@pustaka/api-client';

const client = new PustakaAPI({
  baseURL: 'https://api.pustaka.dev/v1',
  apiKey: 'your-access-token'
});

// List configuration items
const cis = await client.ci.list({
  ci_type: 'Server',
  attributes: {
    environment: 'production'
  }
});

// Create a new CI
const newCI = await client.ci.create({
  name: 'app-server-01',
  ci_type: 'Server',
  attributes: {
    hostname: 'app01.example.com',
    cpu_cores: 4
  }
});
```

## Best Practices

### 1. Authentication
- Store refresh tokens securely
- Implement automatic token refresh in your client
- Use short-lived access tokens
- Revoke tokens when user sessions end

### 2. Error Handling
- Always check HTTP status codes
- Implement exponential backoff for rate limits
- Log error responses for debugging
- Handle network timeouts gracefully

### 3. Pagination
- Use reasonable page sizes (20-100 items)
- Implement pagination controls in UI
- Cache pagination metadata
- Handle empty result sets gracefully

### 4. Performance
- Use specific filters instead of broad searches
- Cache frequently accessed CI types
- Implement client-side caching for static data
- Use conditional requests when possible

### 5. Data Integrity
- Validate data client-side before sending
- Use transaction patterns for related changes
- Implement conflict resolution for concurrent updates
- Always include meaningful descriptions in relationships

### 6. Security
- Never expose tokens in client-side code
- Use HTTPS for all API calls
- Implement proper role-based access control
- Audit access to sensitive operations

### 7. Monitoring
- Monitor API usage and rate limits
- Track error rates and response times
- Set up alerts for unusual activity
- Use audit logs for compliance

### 8. CI Type Design
- Keep attribute names consistent
- Use appropriate data types and validation
- Document attribute purposes and constraints
- Plan for future extensibility

### 9. Relationship Management
- Use meaningful relationship types
- Avoid circular dependencies when possible
- Document relationship purposes
- Regularly review and clean up outdated relationships

### 10. Bulk Operations
- Use batching for large imports
- Implement progress tracking for long operations
- Handle partial failures gracefully
- Provide rollback capabilities for critical operations

## Support and Resources

- **Documentation**: https://docs.pustaka.dev
- **API Reference**: https://api.pustaka.dev/docs
- **Community**: https://community.pustaka.dev
- **Support**: support@pustaka.dev
- **Status Page**: https://status.pustaka.dev

For additional help, check the [OpenAPI specification](./openapi.yaml) for detailed endpoint documentation.
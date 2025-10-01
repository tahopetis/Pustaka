# Pustaka API Examples

This document provides practical examples for using the Pustaka CMDB API, including curl commands, JavaScript SDK usage, and common integration patterns.

## Table of Contents

1. [Authentication Examples](#authentication-examples)
2. [User Management Examples](#user-management-examples)
3. [CI Type Examples](#ci-type-examples)
4. [Configuration Item Examples](#configuration-item-examples)
5. [Relationship Examples](#relationship-examples)
6. [Search and Filtering Examples](#search-and-filtering-examples)
7. [Graph API Examples](#graph-api-examples)
8. [Audit Examples](#audit-examples)
9. [SDK Integration Examples](#sdk-integration-examples)
10. [Common Workflows](#common-workflows)

## Authentication Examples

### Basic Login

```bash
# Login and store tokens
response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@pustaka.dev",
    "password": "securePassword123"
  }')

# Extract tokens using jq
access_token=$(echo $response | jq -r '.access_token')
refresh_token=$(echo $response | jq -r '.refresh_token')

echo "Access Token: $access_token"
echo "Refresh Token: $refresh_token"
```

### Token Refresh

```bash
# Refresh access token
response=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$refresh_token\"}")

new_access_token=$(echo $response | jq -r '.access_token')
echo "New Access Token: $new_access_token"
```

### JavaScript SDK Authentication

```javascript
import { PustakaAPI } from '@pustaka/api-client';

// Initialize the API client
const api = new PustakaAPI({
  baseURL: 'http://localhost:8080/api/v1',
  apiKey: 'your-access-token'
});

// Or authenticate with credentials
const user = await api.auth.login({
  email: 'admin@pustaka.dev',
  password: 'securePassword123'
});

// Automatically refresh tokens
api.setAuthToken(user.access_token);
api.setRefreshToken(user.refresh_token);
```

## User Management Examples

### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@pustaka.dev",
    "password": "securePassword123",
    "name": "Jane Smith",
    "role": "user"
  }'
```

### List Users with Pagination

```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&search=jane" \
  -H "Authorization: Bearer $access_token"
```

### Update User Role

```bash
curl -X PUT http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin",
    "is_active": true
  }'
```

## CI Type Examples

### Create Server CI Type

```bash
curl -X POST http://localhost:8080/api/v1/ci-types \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
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
          "max": 256
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
  }'
```

### Create Application CI Type

```bash
curl -X POST http://localhost:8080/api/v1/ci-types \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Application",
    "description": "Software application",
    "required_attributes": [
      {
        "name": "name",
        "type": "string",
        "validation": {
          "min_length": 1,
          "max_length": 100
        }
      },
      {
        "name": "version",
        "type": "string",
        "validation": {
          "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$"
        }
      }
    ],
    "optional_attributes": [
      {
        "name": "dependencies",
        "type": "array",
        "description": "Application dependencies",
        "validation": {
          "min_length": 0,
          "max_length": 50
        }
      },
      {
        "name": "configuration",
        "type": "object",
        "description": "Application configuration"
      }
    ]
  }'
```

### List CI Types

```bash
curl -X GET "http://localhost:8080/api/v1/ci-types?limit=50" \
  -H "Authorization: Bearer $access_token"
```

## Configuration Item Examples

### Create a Server CI

```bash
curl -X POST http://localhost:8080/api/v1/ci \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

### Create an Application CI

```bash
curl -X POST http://localhost:8080/api/v1/ci \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "customer-portal",
    "ci_type": "Application",
    "attributes": {
      "name": "Customer Portal",
      "version": "2.3.1",
      "dependencies": ["postgresql", "redis", "nginx"],
      "configuration": {
        "port": 8080,
        "max_connections": 1000,
        "cache_ttl": 3600
      }
    },
    "tags": ["production", "web", "customer-facing"]
  }'
```

### Update CI Attributes

```bash
curl -X PUT http://localhost:8080/api/v1/ci/550e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "cpu_cores": 16,
      "memory_gb": 64,
      "environment": "production"
    },
    "tags": ["production", "web", "critical", "upgraded"]
  }'
```

### Get CI Details

```bash
curl -X GET http://localhost:8080/api/v1/ci/550e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer $access_token"
```

## Relationship Examples

### Create Dependencies

```bash
# Application depends on database
curl -X POST http://localhost:8080/api/v1/relationships \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "550e8400-e29b-41d4-a716-446655440001",
    "target_id": "550e8400-e29b-41d4-a716-446655440000",
    "relationship_type": "depends_on",
    "description": "Web application depends on database server"
  }'

# Server hosts application
curl -X POST http://localhost:8080/api/v1/relationships \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "550e8400-e29b-41d4-a716-446655440003",
    "target_id": "550e8400-e29b-41d4-a716-446655440001",
    "relationship_type": "hosts",
    "description": "Web server hosts customer portal application"
  }'
```

### Get CI Relationships

```bash
curl -X GET http://localhost:8080/api/v1/ci/550e8400-e29b-41d4-a716-446655440001/relationships \
  -H "Authorization: Bearer $access_token"
```

## Search and Filtering Examples

### Basic Search

```bash
# Search by name
curl -X GET "http://localhost:8080/api/v1/ci?search=web" \
  -H "Authorization: Bearer $access_token"

# Filter by CI type
curl -X GET "http://localhost:8080/api/v1/ci?ci_type=Server" \
  -H "Authorization: Bearer $access_token"

# Filter by tags
curl -X GET "http://localhost:8080/api/v1/ci?tags=production,critical" \
  -H "Authorization: Bearer $access_token"
```

### Advanced Attribute Search

```bash
# Search servers with specific CPU range
curl -X GET "http://localhost:8080/api/v1/ci?attributes=%7B%22cpu_cores%22%3A%7B%22min%22%3A8%2C%22max%22%3A16%7D%7D" \
  -H "Authorization: Bearer $access_token"

# The URL-encoded attributes JSON: {"cpu_cores":{"min":8,"max":16}}

# Search by environment and minimum memory
curl -X GET "http://localhost:8080/api/v1/ci?attributes=%7B%22environment%22%3A%22production%22%2C%22memory_gb%22%3A%7B%22min%22%3A32%7D%7D" \
  -H "Authorization: Bearer $access_token"

# The URL-encoded attributes JSON: {"environment":"production","memory_gb":{"min":32}}
```

### Complex Search Examples

```bash
# Find production servers with specific CPU and memory
attributes='{
  "ci_type": "Server",
  "environment": "production",
  "cpu_cores": {"min": 8},
  "memory_gb": {"min": 32}
}'

curl -X GET "http://localhost:8080/api/v1/ci?attributes=$(echo $attributes | jq -sRr @uri)" \
  -H "Authorization: Bearer $access_token"

# Find applications with specific dependencies
attributes='{
  "ci_type": "Application",
  "dependencies": {"contains": "postgresql"}
}'

curl -X GET "http://localhost:8080/api/v1/ci?attributes=$(echo $attributes | jq -sRr @uri)" \
  -H "Authorization: Bearer $access_token"
```

## Graph API Examples

### Get Full Graph

```bash
curl -X GET "http://localhost:8080/api/v1/graph?limit=50" \
  -H "Authorization: Bearer $access_token"
```

### Explore Graph from Specific CI

```bash
# Explore 2 levels deep from a specific CI
curl -X GET "http://localhost:8080/api/v1/graph/explore?start_id=550e8400-e29b-41d4-a716-446655440001&depth=2" \
  -H "Authorization: Bearer $access_token"

# Explore specific relationship types
curl -X GET "http://localhost:8080/api/v1/graph/explore?start_id=550e8400-e29b-41d4-a716-446655440001&depth=3&relationship_types=depends_on,hosts" \
  -H "Authorization: Bearer $access_token"
```

## Audit Examples

### Get Recent Activity

```bash
curl -X GET "http://localhost:8080/api/v1/audit/logs?page=1&limit=20" \
  -H "Authorization: Bearer $access_token"
```

### Filter Audit Logs

```bash
# Get creation events only
curl -X GET "http://localhost:8080/api/v1/audit/logs?action=create&limit=50" \
  -H "Authorization: Bearer $access_token"

# Get CI-related events
curl -X GET "http://localhost:8080/api/v1/audit/logs?resource_type=ci&limit=50" \
  -H "Authorization: Bearer $access_token"

# Get events by date range
curl -X GET "http://localhost:8080/api/v1/audit/logs?start_date=2023-01-01T00:00:00Z&end_date=2023-01-31T23:59:59Z" \
  -H "Authorization: Bearer $access_token"
```

### Get Audit Statistics

```bash
curl -X GET "http://localhost:8080/api/v1/audit/stats?period=month" \
  -H "Authorization: Bearer $access_token"
```

### Export Audit Logs

```bash
# Export as CSV
curl -X GET "http://localhost:8080/api/v1/audit/export?start_date=2023-01-01T00:00:00Z&end_date=2023-01-31T23:59:59Z&format=csv" \
  -H "Authorization: Bearer $access_token" \
  -o audit-logs-jan-2023.csv

# Export as JSON
curl -X GET "http://localhost:8080/api/v1/audit/export?start_date=2023-01-01T00:00:00Z&end_date=2023-01-31T23:59:59Z&format=json" \
  -H "Authorization: Bearer $access_token" \
  -o audit-logs-jan-2023.json
```

## SDK Integration Examples

### JavaScript SDK Basic Usage

```javascript
import { PustakaAPI } from '@pustaka/api-client';

const api = new PustakaAPI({
  baseURL: 'http://localhost:8080/api/v1',
  apiKey: 'your-access-token'
});

// List CIs with filtering
const servers = await api.ci.list({
  ci_type: 'Server',
  attributes: {
    environment: 'production'
  },
  limit: 50
});

// Create a new CI
const newServer = await api.ci.create({
  name: 'app-server-02',
  ci_type: 'Server',
  attributes: {
    hostname: 'app02.example.com',
    cpu_cores: 4,
    memory_gb: 16
  },
  tags: ['production', 'app']
});

// Create relationships
await api.relationships.create({
  source_id: newServer.id,
  target_id: 'database-server-01',
  relationship_type: 'depends_on',
  description: 'Application server depends on database'
});

// Get graph data
const graph = await api.graph.get({ limit: 100 });
console.log(`Found ${graph.nodes.length} nodes and ${graph.edges.length} edges`);
```

### Python SDK Integration

```python
import pustaka_sdk
import json

# Initialize client
client = pustaka_sdk.Client(
    base_url='http://localhost:8080/api/v1',
    api_key='your-access-token'
)

# Create CI Type
server_type = client.ci_types.create(
    name="Database Server",
    description="Database server configuration",
    required_attributes=[
        {
            "name": "hostname",
            "type": "string",
            "validation": {"min_length": 1}
        },
        {
            "name": "database_type",
            "type": "string",
            "validation": {"enum": ["postgresql", "mysql", "mongodb"]}
        }
    ]
)

# Create multiple CIs
servers = []
for i in range(1, 4):
    server = client.ci.create(
        name=f"db-server-{i:02d}",
        ci_type="Database Server",
        attributes={
            "hostname": f"db{i:02d}.example.com",
            "database_type": "postgresql"
        },
        tags=["database", "production"]
    )
    servers.append(server)

# Search for specific CIs
production_dbs = client.ci.list(
    attributes={
        "ci_type": "Database Server",
        "environment": "production"
    }
)

print(f"Found {len(production_dbs)} production database servers")
```

### Go SDK Integration

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/pustaka/pustaka-go"
    "github.com/pustaka/pustaka-go/models"
)

func main() {
    // Initialize client
    client := pustaka.NewClient(&pustaka.ClientConfig{
        BaseURL: "http://localhost:8080/api/v1",
        APIKey:  "your-access-token",
    })

    ctx := context.Background()

    // Create CI Type
    ciType := &models.CIType{
        Name:        "Network Device",
        Description: "Network infrastructure device",
        RequiredAttributes: []*models.AttributeDefinition{
            {
                Name: "device_name",
                Type: "string",
                Validation: &models.ValidationRules{
                    MinLength: 1,
                    MaxLength: 100,
                },
            },
            {
                Name: "ip_address",
                Type: "string",
                Validation: &models.ValidationRules{
                    Format: "ipv4",
                },
            },
        },
    }

    createdType, err := client.CITypes.Create(ctx, ciType)
    if err != nil {
        log.Fatalf("Failed to create CI type: %v", err)
    }

    // Create CI
    ci := &models.ConfigurationItem{
        Name:     "switch-01",
        CIType:   "Network Device",
        Attributes: map[string]interface{}{
            "device_name": "core-switch-01",
            "ip_address": "192.168.1.1",
        },
        Tags: []string{"network", "core", "production"},
    }

    createdCI, err := client.CI.Create(ctx, ci)
    if err != nil {
        log.Fatalf("Failed to create CI: %v", err)
    }

    fmt.Printf("Created CI: %s (ID: %s)\n", createdCI.Name, createdCI.ID)
}
```

## Common Workflows

### 1. Onboarding New Application Infrastructure

```bash
#!/bin/bash

# Create CI types
server_type_id=$(curl -s -X POST http://localhost:8080/api/v1/ci-types \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Application Server",
    "description": "Application hosting server",
    "required_attributes": [{"name": "hostname", "type": "string"}],
    "optional_attributes": [{"name": "environment", "type": "string"}]
  }' | jq -r '.id')

app_type_id=$(curl -s -X POST http://localhost:8080/api/v1/ci-types \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web Application",
    "description": "Web application service",
    "required_attributes": [{"name": "name", "type": "string"}]
  }' | jq -r '.id')

# Create servers
server1_id=$(curl -s -X POST http://localhost:8080/api/v1/ci \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"app-server-01\", \"ci_type\": \"Application Server\", \"attributes\": {\"hostname\": \"app01.example.com\"}, \"tags\": [\"application\", \"server\"]}" | jq -r '.id')

server2_id=$(curl -s -X POST http://localhost:8080/api/v1/ci \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"app-server-02\", \"ci_type\": \"Application Server\", \"attributes\": {\"hostname\": \"app02.example.com\"}, \"tags\": [\"application\", \"server\"]}" | jq -r '.id')

# Create applications
app_id=$(curl -s -X POST http://localhost:8080/api/v1/ci \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"customer-portal\", \"ci_type\": \"Web Application\", \"attributes\": {\"name\": \"Customer Portal\"}, \"tags\": [\"application\", \"web\"]}" | jq -r '.id')

# Create relationships
curl -s -X POST http://localhost:8080/api/v1/relationships \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d "{\"source_id\": \"$app_id\", \"target_id\": \"$server1_id\", \"relationship_type\": \"hosts_on\"}"

curl -s -X POST http://localhost:8080/api/v1/relationships \
  -H "Authorization: Bearer $access_token" \
  -H "Content-Type: application/json" \
  -d "{\"source_id\": \"$app_id\", \"target_id\": \"$server2_id\", \"relationship_type\": \"hosts_on\"}"

echo "Infrastructure onboarded successfully!"
echo "Application ID: $app_id"
echo "Server IDs: $server1_id, $server2_id"
```

### 2. Impact Analysis

```bash
#!/bin/bash

# Get relationships for a specific CI
ci_id="550e8400-e29b-41d4-a716-446655440001"

echo "=== Impact Analysis for CI: $ci_id ==="

# Get all relationships
relationships=$(curl -s -X GET "http://localhost:8080/api/v1/ci/$ci_id/relationships" \
  -H "Authorization: Bearer $access_token")

# Analyze dependencies
echo "Dependencies:"
echo "$relationships" | jq -r '.relationships[] | select(.relationship_type == "depends_on") | "  \(.source_ci.name) depends on \(.target_ci.name)"'

echo ""
echo "Dependents:"
echo "$relationships" | jq -r '.relationships[] | select(.relationship_type == "depends_on") | "  \(.target_ci.name) is depended upon by \(.source_ci.name)"'

# Get graph for deeper analysis
graph=$(curl -s -X GET "http://localhost:8080/api/v1/graph/explore?start_id=$ci_id&depth=3" \
  -H "Authorization: Bearer $access_token")

echo ""
echo "=== Full Impact Graph (3 levels deep) ==="
echo "Total nodes: $(echo $graph | jq '.metadata.total_nodes')"
echo "Total edges: $(echo $graph | jq '.metadata.total_edges')"

# Export detailed impact report
echo "$graph" > impact-analysis-$(date +%Y%m%d).json
echo "Impact analysis saved to impact-analysis-$(date +%Y%m%d).json"
```

### 3. Daily Compliance Check

```bash
#!/bin/bash

# Check for compliance issues

echo "=== Daily Compliance Check ==="
date=$(date +%Y-%m-%d)

# Get all production servers
servers=$(curl -s -X GET "http://localhost:8080/api/v1/ci?attributes=%7B%22environment%22%3A%22production%22%2C%22ci_type%22%3A%22Server%22%7D" \
  -H "Authorization: Bearer $access_token")

echo "Production Servers Found:"
echo "$servers" | jq -r '.cis[].name'

# Check for servers without monitoring tag
echo ""
echo "=== Servers Without Monitoring Tag ==="
echo "$servers" | jq -r '.cis[] | select(.tags | index("monitoring") | not) | .name'

# Get audit statistics
stats=$(curl -s -X GET "http://localhost:8080/api/v1/audit/stats?period=day" \
  -H "Authorization: Bearer $access_token")

echo ""
echo "=== Daily Activity Summary ==="
echo "Total actions: $(echo $stats | jq '.total_actions')"
echo "Actions by type:"
echo "$stats" | jq -r '.actions_by_type | to_entries[] | "  \(.key): \(.value)"'

# Generate compliance report
{
  echo "Compliance Report - $date"
  echo "=================================="
  echo ""
  echo "Production Servers: $(echo $servers | jq '.cis | length')"
  echo "Servers Without Monitoring: $(echo "$servers" | jq '.cis[] | select(.tags | index("monitoring") | not) | length')"
  echo ""
  echo "Daily Actions: $(echo $stats | jq '.total_actions')"
  echo "Creations: $(echo $stats | jq '.actions_by_type.create // 0')"
  echo "Updates: $(echo $stats | jq '.actions_by_type.update // 0')"
  echo "Deletions: $(echo $stats | jq '.actions_by_type.delete // 0')"
} > compliance-report-$date.txt

echo "Compliance report saved to compliance-report-$date.txt"
```

These examples provide a comprehensive foundation for integrating with the Pustaka CMDB API. Remember to:

1. Always handle authentication securely
2. Implement proper error handling
3. Use pagination for large datasets
4. Cache frequently accessed data
5. Monitor API usage and rate limits
6. Follow the best practices outlined in the API guide

For more detailed information, refer to the [OpenAPI specification](./openapi.yaml) and [API guide](./api-guide.md).
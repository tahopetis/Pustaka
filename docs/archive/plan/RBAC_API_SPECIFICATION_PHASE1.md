# 📘 API & RBAC Specification – Phase 1

**Project**: Pustaka CMDB
**Version**: 1.0
**Date**: 2025-10-01
**Scope**: Phase 1

---

## 🎯 Purpose
This document defines the API contracts, RBAC rules, and validation requirements for **Phase 1** of Pustaka CMDB.
It removes ambiguity for implementation by both human developers and AI coding agents.

---

## 🔹 Standard HTTP Status Codes

**Success Responses**
* `200 OK` – Successful operation (GET, PUT, DELETE with response body)
* `201 Created` – Resource created successfully (POST)
* `204 No Content` – Successful operation with no response body (DELETE)

**Error Responses**
* `400 Bad Request` – Invalid request body, missing required fields, validation failed
* `401 Unauthorized` – Invalid credentials, missing or invalid token
* `403 Forbidden` – Valid credentials but insufficient permissions
* `404 Not Found` – Resource not found
* `409 Conflict` – Resource already exists or constraint violation
* `422 Unprocessable Entity` – Validation failed (detailed error messages)

**Standard Error Response Format**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "field": "hostname",
      "reason": "required field missing"
    }
  }
}
```

---

## 🔹 Authentication API

### `POST /auth/login`
Authenticate a user and return a JWT token.

**Request**
```json
{
  "username": "admin",
  "password": "secret123"
}
```

**Response**

```json
{
  "access_token": "<jwt_token>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "roles": ["admin"],
    "permissions": ["ci:create", "ci:read", "ci:update", "ci:delete"]
  }
}
```

**Errors**
* `400 Bad Request` – invalid request format
* `401 Unauthorized` – invalid credentials

---

## 🔹 CI Type Schema Management APIs

### `POST /ci-types` – Create CI Type Schema

**Request**
```json
{
  "name": "Server",
  "description": "Physical or virtual server",
  "required_attributes": [
    {
      "name": "hostname",
      "type": "string",
      "description": "Server hostname"
    },
    {
      "name": "ip_address",
      "type": "string",
      "description": "Primary IP address"
    },
    {
      "name": "os",
      "type": "string",
      "description": "Operating system"
    }
  ],
  "optional_attributes": [
    {
      "name": "vendor",
      "type": "string",
      "description": "Hardware vendor"
    },
    {
      "name": "rack_location",
      "type": "string",
      "description": "Data center rack location"
    }
  ]
}
```

**Response**
```json
{
  "id": "uuid",
  "name": "Server",
  "description": "Physical or virtual server",
  "required_attributes": [...],
  "optional_attributes": [...],
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `400 Bad Request` – invalid schema definition
* `403 Forbidden` – insufficient permissions
* `409 Conflict` – CI type already exists

---

### `GET /ci-types` – List CI Type Schemas

**Query Parameters**
* `page` – Page number (default: 1)
* `limit` – Items per page (default: 20, max: 100)
* `search` – Search by name or description

**Response**
```json
{
  "ci_types": [
    {
      "id": "uuid",
      "name": "Server",
      "description": "Physical or virtual server",
      "required_attributes": [...],
      "optional_attributes": [...],
      "created_by": "admin",
      "created_at": "2025-10-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### `GET /ci-types/{id}` – Read CI Type Schema

**Response**
```json
{
  "id": "uuid",
  "name": "Server",
  "description": "Physical or virtual server",
  "required_attributes": [...],
  "optional_attributes": [...],
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `404 Not Found` – CI type not found

---

### `PUT /ci-types/{id}` – Update CI Type Schema

**Request**
```json
{
  "description": "Updated description",
  "optional_attributes": [
    {
      "name": "vendor",
      "type": "string",
      "description": "Hardware vendor"
    },
    {
      "name": "cpu_cores",
      "type": "integer",
      "description": "Number of CPU cores"
    }
  ]
}
```

**Response**
```json
{
  "id": "uuid",
  "name": "Server",
  "description": "Updated description",
  "required_attributes": [...],
  "optional_attributes": [...],
  "updated_by": "admin",
  "updated_at": "2025-10-01T11:00:00Z"
}
```

**Errors**
* `400 Bad Request` – invalid schema updates
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – CI type not found

---

### `DELETE /ci-types/{id}` – Delete CI Type Schema

**Response**
```json
{
  "status": "deleted",
  "message": "CI type schema deleted successfully"
}
```

**Errors**
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – CI type not found
* `409 Conflict` – Cannot delete CI type with existing CIs

---

## 🔹 Configuration Item (CI) APIs

### `POST /ci` – Create CI

**Request**

```json
{
  "ci_type": "Server",
  "attributes": {
    "hostname": "srv01",
    "ip_address": "10.0.0.5",
    "os": "RHEL 9.5"
  },
  "tags": ["production", "web-tier"]
}
```

**Response**

```json
{
  "id": "uuid",
  "ci_type": "Server",
  "attributes": {
    "hostname": "srv01",
    "ip_address": "10.0.0.5",
    "os": "RHEL 9.5"
  },
  "tags": ["production", "web-tier"],
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `400 Bad Request` – missing required attributes, validation failed
* `403 Forbidden` – insufficient permissions
* `409 Conflict` – CI with same unique attributes already exists

---

### `GET /ci/{id}` – Read CI

**Response**

```json
{
  "id": "uuid",
  "ci_type": "Server",
  "attributes": {
    "hostname": "srv01",
    "ip_address": "10.0.0.5",
    "os": "RHEL 9.5"
  },
  "tags": ["production", "web-tier"],
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `404 Not Found` – CI not found

---

### `GET /ci` – List CIs

**Query Parameters**
* `page` – Page number (default: 1)
* `limit` – Items per page (default: 20, max: 100)
* `ci_type` – Filter by CI type
* `search` – Search in attributes
* `tags` – Filter by tags (comma-separated)
* `created_by` – Filter by creator
* `sort` – Sort field (default: created_at)
* `order` – Sort order (asc/desc, default: desc)

**Response**
```json
{
  "cis": [
    {
      "id": "uuid",
      "ci_type": "Server",
      "attributes": {
        "hostname": "srv01",
        "ip_address": "10.0.0.5",
        "os": "RHEL 9.5"
      },
      "tags": ["production", "web-tier"],
      "created_by": "admin",
      "created_at": "2025-10-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### `PUT /ci/{id}` – Update CI

**Request**

```json
{
  "attributes": {
    "ip_address": "10.0.0.10"
  },
  "tags": ["production", "web-tier", "critical"]
}
```

**Response**

```json
{
  "id": "uuid",
  "ci_type": "Server",
  "attributes": {
    "hostname": "srv01",
    "ip_address": "10.0.0.10",
    "os": "RHEL 9.5"
  },
  "tags": ["production", "web-tier", "critical"],
  "updated_by": "admin",
  "updated_at": "2025-10-01T10:30:00Z"
}
```

**Errors**
* `400 Bad Request` – validation failed
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – CI not found

---

### `DELETE /ci/{id}` – Delete CI

**Response**
```json
{
  "status": "deleted",
  "message": "CI deleted successfully"
}
```

**Errors**
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – CI not found

---

## 🔹 Relationship APIs

### `POST /relationship` – Create Relationship

**Request**

```json
{
  "source_id": "uuid1",
  "target_id": "uuid2",
  "relationship_type": "HOSTED_ON",
  "attributes": {
    "environment": "production",
    "sla_impact": "high"
  }
}
```

**Response**

```json
{
  "id": "uuid_rel",
  "source_id": "uuid1",
  "target_id": "uuid2",
  "relationship_type": "HOSTED_ON",
  "attributes": {
    "environment": "production",
    "sla_impact": "high"
  },
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `400 Bad Request` – invalid relationship data, circular dependency
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – source or target CI not found
* `409 Conflict` – relationship already exists

---

### `GET /relationship/{id}` – Read Relationship

**Response**

```json
{
  "id": "uuid_rel",
  "source_id": "uuid1",
  "target_id": "uuid2",
  "relationship_type": "HOSTED_ON",
  "attributes": {
    "environment": "production",
    "sla_impact": "high"
  },
  "created_by": "admin",
  "created_at": "2025-10-01T10:00:00Z",
  "updated_at": "2025-10-01T10:00:00Z"
}
```

**Errors**
* `404 Not Found` – relationship not found

---

### `GET /relationship` – List Relationships

**Query Parameters**
* `page` – Page number (default: 1)
* `limit` – Items per page (default: 20, max: 100)
* `source_id` – Filter by source CI ID
* `target_id` – Filter by target CI ID
* `relationship_type` – Filter by relationship type
* `search` – Search in attributes

**Response**
```json
{
  "relationships": [
    {
      "id": "uuid_rel",
      "source_id": "uuid1",
      "target_id": "uuid2",
      "relationship_type": "HOSTED_ON",
      "attributes": {
        "environment": "production",
        "sla_impact": "high"
      },
      "created_by": "admin",
      "created_at": "2025-10-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### `PUT /relationship/{id}` – Update Relationship

**Request**

```json
{
  "attributes": {
    "sla_impact": "critical"
  }
}
```

**Response**

```json
{
  "id": "uuid_rel",
  "source_id": "uuid1",
  "target_id": "uuid2",
  "relationship_type": "HOSTED_ON",
  "attributes": {
    "environment": "production",
    "sla_impact": "critical"
  },
  "updated_by": "admin",
  "updated_at": "2025-10-01T10:30:00Z"
}
```

**Errors**
* `400 Bad Request` – validation failed
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – relationship not found

---

### `DELETE /relationship/{id}` – Delete Relationship

**Response**
```json
{
  "status": "deleted",
  "message": "Relationship deleted successfully"
}
```

**Errors**
* `403 Forbidden` – insufficient permissions
* `404 Not Found` – relationship not found

---

## 🔹 Audit API

### `GET /audit` – Retrieve Audit Logs

**Query Parameters**
* `page` – Page number (default: 1)
* `limit` – Items per page (default: 20, max: 100)
* `entity` – Filter by entity type (ci, relationship, ci_type)
* `entity_id` – Filter by specific entity ID
* `action` – Filter by action (create, read, update, delete)
* `performed_by` – Filter by user who performed action
* `from_date` – Start date filter (ISO 8601)
* `to_date` – End date filter (ISO 8601)

**Response**

```json
{
  "audit_logs": [
    {
      "id": "audit_uuid",
      "entity": "ci",
      "entity_id": "uuid",
      "action": "create",
      "performed_by": "admin",
      "timestamp": "2025-10-01T10:00:00Z",
      "before": null,
      "after": {
        "ci_type": "Server",
        "attributes": {
          "hostname": "srv01",
          "ip_address": "10.0.0.5",
          "os": "RHEL 9.5"
        },
        "tags": ["production"]
      }
    },
    {
      "id": "audit_uuid2",
      "entity": "ci",
      "entity_id": "uuid",
      "action": "update",
      "performed_by": "admin",
      "timestamp": "2025-10-01T10:30:00Z",
      "before": {
        "attributes": {
          "ip_address": "10.0.0.5"
        }
      },
      "after": {
        "attributes": {
          "ip_address": "10.0.0.10"
        }
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 2,
    "total_pages": 1
  }
}
```

**Errors**
* `400 Bad Request` – invalid query parameters
* `403 Forbidden` – insufficient permissions

---

### `GET /audit/{id}` – Read Specific Audit Log

**Response**

```json
{
  "id": "audit_uuid",
  "entity": "ci",
  "entity_id": "uuid",
  "action": "create",
  "performed_by": "admin",
  "timestamp": "2025-10-01T10:00:00Z",
  "before": null,
  "after": {
    "ci_type": "Server",
    "attributes": {
      "hostname": "srv01",
      "ip_address": "10.0.0.5",
      "os": "RHEL 9.5"
    },
    "tags": ["production"]
  },
  "metadata": {
    "ip_address": "192.168.1.100",
    "user_agent": "Pustaka-API/1.0"
  }
}
```

**Errors**
* `404 Not Found` – audit log not found
* `403 Forbidden` – insufficient permissions

---

## 🔹 RBAC Model

### Roles (Phase 1)

* **Admin** → full CRUD on CIs, CI types, and relationships, manage users, view audit logs
* **Editor** → CRUD on CIs and relationships, read CI types, view audit logs
* **Viewer** → read-only access to CIs, relationships, CI types, and audit logs

### Permissions Matrix

| Action                     | Admin | Editor | Viewer |
| -------------------------- | :---: | :----: | :----: |
| Create CI                  |   ✅   |    ✅   |    ❌   |
| Read CI                    |   ✅   |    ✅   |    ✅   |
| Update CI                  |   ✅   |    ✅   |    ❌   |
| Delete CI                  |   ✅   |    ✅   |    ❌   |
| List CIs                   |   ✅   |    ✅   |    ✅   |
| Search CIs                 |   ✅   |    ✅   |    ✅   |
| Create CI Type Schema      |   ✅   |    ❌   |    ❌   |
| Read CI Type Schema        |   ✅   |    ✅   |    ✅   |
| Update CI Type Schema      |   ✅   |    ❌   |    ❌   |
| Delete CI Type Schema      |   ✅   |    ❌   |    ❌   |
| List CI Type Schemas       |   ✅   |    ✅   |    ✅   |
| Create Relationship        |   ✅   |    ✅   |    ❌   |
| Read Relationship          |   ✅   |    ✅   |    ✅   |
| Update Relationship        |   ✅   |    ✅   |    ❌   |
| Delete Relationship        |   ✅   |    ✅   |    ❌   |
| List Relationships         |   ✅   |    ✅   |    ✅   |
| View Audit Logs            |   ✅   |    ✅   |    ✅   |
| Search Audit Logs          |   ✅   |    ✅   |    ✅   |

---

## 🔹 Validation Rules (Schema-Based)

### CI Type Schema Definition

```json
{
  "ci_type": "Server",
  "required_attributes": [
    {
      "name": "hostname",
      "type": "string",
      "validation": {
        "pattern": "^[a-zA-Z0-9-]+$",
        "min_length": 1,
        "max_length": 253
      }
    },
    {
      "name": "ip_address",
      "type": "string",
      "validation": {
        "pattern": "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
      }
    },
    {
      "name": "os",
      "type": "string",
      "validation": {
        "enum": ["RHEL", "Ubuntu", "CentOS", "Debian", "Windows Server", "Other"]
      }
    }
  ],
  "optional_attributes": [
    {
      "name": "vendor",
      "type": "string",
      "validation": {
        "max_length": 100
      }
    },
    {
      "name": "rack_location",
      "type": "string",
      "validation": {
        "max_length": 50
      }
    }
  ]
}
```

### Relationship Type Schema Definition

```json
{
  "relationship_type": "HOSTED_ON",
  "required_attributes": [
    {
      "name": "environment",
      "type": "string",
      "validation": {
        "enum": ["development", "staging", "production", "testing"]
      }
    }
  ],
  "optional_attributes": [
    {
      "name": "sla_impact",
      "type": "string",
      "validation": {
        "enum": ["low", "medium", "high", "critical"]
      }
    },
    {
      "name": "start_date",
      "type": "datetime",
      "validation": {
        "format": "ISO8601"
      }
    }
  ]
}
```

---

## 🔹 Request/Response Headers

**Required Headers**
* `Authorization: Bearer <jwt_token>` – For all protected endpoints
* `Content-Type: application/json` – For POST/PUT requests

**Response Headers**
* `X-Request-ID: <uuid>` – Unique request identifier for debugging
* `X-Rate-Limit-Remaining: <number>` – Remaining API calls (if rate limited)

---

## 🔹 Rate Limiting (Phase 1)

* **Authenticated requests**: 1000 requests/hour per user
* **Authentication endpoint**: 60 requests/hour per IP
* **Rate limit headers**: Included in all responses

---

## ✅ Summary

### APIs Included in Phase 1:
* **Authentication API** – JWT-based authentication with user/permissions info
* **CI Type Schema Management APIs** – Full CRUD for CI type schemas with validation rules
* **Configuration Item APIs** – Full CRUD for CIs with tags and flexible attributes
* **Relationship APIs** – Full CRUD for CI relationships with attributes
* **Audit APIs** – Comprehensive audit logging with search and filtering

### Key Features:
* **RBAC enforces Admin, Editor, Viewer roles** with granular permissions
* **Schema-based validation** ensures JSONB attributes match defined rules
* **Pagination and search** on all list endpoints
* **Comprehensive error handling** with standardized error format
* **Audit trail** for all operations with before/after state
* **Tags support** for CI categorization and filtering
* **Creator tracking** (created_by, updated_by) for accountability

### Implementation Notes:
* All UUIDs follow RFC 4122 format
* All timestamps are ISO 8601 (UTC)
* This spec provides complete details needed for **Phase 1 implementation**
* Validation errors include specific field-level details for better UX
* Relationships support circular dependency detection


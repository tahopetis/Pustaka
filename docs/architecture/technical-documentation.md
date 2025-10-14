# Pustaka CMDB - Comprehensive Technical Documentation

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [System Architecture Overview](#system-architecture-overview)
3. [Backend Architecture](#backend-architecture)
4. [Frontend Architecture](#frontend-architecture)
5. [Database Architecture](#database-architecture)
6. [API Design and Patterns](#api-design-and-patterns)
7. [Security Architecture](#security-architecture)
8. [Data Models and Schemas](#data-models-and-schemas)
9. [Development Patterns and Conventions](#development-patterns-and-conventions)
10. [Deployment Architecture](#deployment-architecture)
11. [Testing Strategy](#testing-strategy)
12. [Performance Considerations](#performance-considerations)
13. [Monitoring and Observability](#monitoring-and-observability)

## Executive Summary

Pustaka is a modern Configuration Management Database (CMDB) built with a microservices-oriented architecture. It features a hierarchical taxonomy system for IT assets, comprehensive relationship mapping, role-based access control, audit logging, and graph visualization capabilities. The system supports flexible schema definitions (FSD-compliant) and provides both REST APIs and a web interface for managing configuration items and their relationships.

### Key Characteristics
- **Multi-database Architecture**: PostgreSQL for structured data, Neo4j for relationships, Redis for caching
- **Modern Tech Stack**: Go (Chi v5) backend, Vue 3 + TypeScript frontend
- **Security-first Design**: JWT authentication, RBAC, comprehensive audit trails
- **Scalable Architecture**: Containerized deployment with Docker Compose
- **Flexible Schema**: JSONB-based attribute system with validation
- **Graph-based Relationships**: Native support for complex dependency mapping

## System Architecture Overview

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   Mobile App    │    │   External APIs │
│   (Vue 3 + TS)  │    │   (Future)      │    │   (Future)      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      API Gateway          │
                    │   (Chi v5 Router)         │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼──────────┐ ┌─────────▼──────────┐ ┌─────────▼──────────┐
│  PostgreSQL        │ │      Neo4j         │ │       Redis        │
│  (Primary Data)    │ │  (Relationships)   │ │    (Caching)       │
└────────────────────┘ └────────────────────┘ └────────────────────┘
```

### Core Services

1. **Authentication Service**: JWT-based authentication with refresh tokens
2. **Authorization Service**: Role-based access control (RBAC)
3. **CI Management Service**: Configuration item CRUD operations
4. **Relationship Service**: Graph-based relationship management
5. **Audit Service**: Comprehensive audit logging
6. **Graph Service**: Network analysis and visualization

## Backend Architecture

### Technology Stack
- **Language**: Go 1.22
- **Web Framework**: Chi v5 (lightweight, fast HTTP router)
- **Database Drivers**:
  - PostgreSQL: pgx/v5
  - Neo4j: neo4j-go-driver/v5
  - Redis: go-redis/v9
- **Authentication**: golang-jwt/jwt/v5
- **Password Hashing**: alexedwards/argon2id
- **Logging**: rs/zerolog (structured logging)
- **Configuration**: spf13/viper + joho/godotenv

### Project Structure

```
cmd/
├── api/
│   └── main.go              # Application entrypoint
└── migrations/              # Database migration files

internal/
├── api/                     # HTTP handlers and routing
│   ├── handlers/            # Request handlers
│   ├── middleware/          # Custom middleware
│   └── *_test.go           # Integration tests
├── auth/                    # Authentication & authorization
│   ├── jwt.go              # JWT token service
│   ├── rbac.go             # Role-based access control
│   ├── password.go         # Password hashing
│   └── *_test.go          # Unit tests
├── ci/                      # Core CMDB business logic
│   ├── models.go           # Data models
│   ├── service.go          # Business logic layer
│   ├── repository.go       # PostgreSQL data access
│   ├── neo4j_service.go   # Neo4j relationship management
│   ├── audit_*.go         # Audit logging system
│   └── *_test.go          # Unit tests
├── database/               # Database connection management
└── config/                 # Configuration loading
```

### Architecture Patterns

#### 1. Layered Architecture
```
┌─────────────────────────────┐
│        HTTP Layer           │  (Handlers, Middleware)
├─────────────────────────────┤
│        Service Layer        │  (Business Logic)
├─────────────────────────────┤
│       Repository Layer      │  (Data Access)
├─────────────────────────────┤
│       Database Layer        │  (PostgreSQL, Neo4j, Redis)
└─────────────────────────────┘
```

#### 2. Dependency Injection
- Services are injected into handlers
- Repositories are injected into services
- Database connections are injected into repositories
- All dependencies are managed in main.go

#### 3. Middleware Chain
```
Request → RequestID → RealIP → Recovery → Timeout →
          CORS → Logger → JWT Auth → Activity Tracker →
          Audit Logging → RBAC → Handler → Response
```

### Key Backend Components

#### 1. HTTP Router (`cmd/api/main.go`)
- Chi v5 router with nested routes
- Middleware stack for cross-cutting concerns
- Graceful shutdown with signal handling
- Health check endpoint for monitoring

#### 2. Authentication System (`internal/auth/`)
- JWT tokens with access/refresh mechanism
- Argon2ID password hashing
- User session management
- Token validation middleware

#### 3. RBAC System (`internal/auth/rbac.go`)
- Role-based permissions
- Granular access control (resource:action format)
- Dynamic permission checking
- User role assignment

#### 4. CI Management (`internal/ci/`)
- Flexible schema with JSONB attributes
- Attribute validation with custom rules
- CI type definitions with required/optional fields
- Full-text search capabilities

#### 5. Relationship Management (`internal/ci/neo4j_service.go`)
- Bidirectional relationships in Neo4j
- Graph traversal algorithms
- Impact analysis
- Circular dependency detection

#### 6. Audit System (`internal/ci/audit_*.go`)
- Comprehensive audit trails
- User action tracking
- IP address and user agent logging
- Configurable retention policies

## Frontend Architecture

### Technology Stack
- **Framework**: Vue 3 with Composition API
- **Language**: TypeScript
- **State Management**: Pinia
- **Routing**: Vue Router 4
- **UI Framework**: Tailwind CSS + Headless UI
- **HTTP Client**: Axios
- **Graph Visualization**: vis-network
- **Build Tool**: Vite
- **Testing**: Vitest (unit), Cypress (E2E)

### Project Structure

```
web/src/
├── components/              # Reusable Vue components
│   ├── base/               # Base UI components
│   ├── ci/                 # CI-specific components
│   ├── graph/              # Graph visualization
│   └── layout/             # Layout components
├── views/                  # Page components
│   ├── auth/               # Authentication pages
│   ├── ci/                 # CI management pages
│   ├── graph/              # Graph visualization
│   └── audit/              # Audit logs
├── stores/                 # Pinia state management
│   ├── auth.ts             # Authentication state
│   ├── ci.ts               # CI data state
│   └── notification.ts     # Global notifications
├── services/               # API communication
│   └── api.ts              # Axios HTTP client
├── router/                 # Vue Router configuration
└── utils/                  # Utility functions
```

### Key Frontend Patterns

#### 1. Component Architecture
- **Composition API**: Modern Vue 3 pattern for reusability
- **Props/Emits**: Clear parent-child communication
- **Slots**: Flexible component composition
- **Provide/Inject**: Cross-component dependency injection

#### 2. State Management (Pinia)
- **Stores**: Domain-specific state containers
- **Actions**: Async operations and state mutations
- **Getters**: Computed derived state
- **Persistence**: LocalStorage for auth state

#### 3. Route Protection
- **Navigation Guards**: Authentication and authorization checks
- **Permission-based Routing**: RBAC integration
- **Lazy Loading**: Code splitting for performance

#### 4. API Integration
- **Axios Interceptors**: Request/response handling
- **Token Refresh**: Automatic token renewal
- **Error Handling**: Centralized error processing
- **Request Cancellation**: Prevent race conditions

### UI/UX Patterns

#### 1. Responsive Design
- **Mobile-first**: Tailwind CSS responsive utilities
- **Progressive Enhancement**: Core functionality everywhere
- **Adaptive Layouts**: Flexible grid system

#### 2. User Feedback
- **Toast Notifications**: Non-intrusive alerts
- **Loading States**: Skeleton screens and spinners
- **Error Boundaries**: Graceful error handling
- **Form Validation**: Real-time validation feedback

## Database Architecture

### Multi-Database Strategy

#### 1. PostgreSQL (Primary Database)
**Purpose**: Structured data, ACID transactions, relational integrity
**Tables**:
- `users`: User accounts and authentication
- `roles`: Role definitions
- `permissions`: Permission definitions
- `user_roles`: User-role assignments
- `role_permissions`: Role-permission assignments
- `ci_type_definitions`: CI type schemas
- `configuration_items`: Main CI data with JSONB attributes
- `relationships`: Relationship metadata
- `audit_logs`: Comprehensive audit trail

**Key Features**:
- JSONB for flexible CI attributes
- GIN indexes for full-text search
- Foreign key constraints for data integrity
- Triggers for automatic timestamp updates

#### 2. Neo4j (Graph Database)
**Purpose**: Complex relationships, graph traversals, network analysis
**Node Types**:
- `ConfigurationItem`: Represents CIs in the graph

**Relationship Types**:
- Dynamic based on relationship type definitions
- Bidirectional relationships
- Attributes on relationships

**Key Features**:
- Cypher query language for graph traversals
- Path finding algorithms
- Impact analysis capabilities
- Circular dependency detection

#### 3. Redis (Cache & Session Store)
**Purpose**: Caching, session storage, rate limiting
**Use Cases**:
- Session data storage
- API response caching
- Rate limiting counters
- Background job queues (future)

### Database Schema Design

#### 1. Flexible Schema Architecture
```sql
-- CI Type Definitions (Schema as Data)
CREATE TABLE ci_type_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    required_attributes JSONB NOT NULL DEFAULT '[]',
    optional_attributes JSONB NOT NULL DEFAULT '[]',
    -- ...
);

-- Configuration Items (Flexible Attributes)
CREATE TABLE configuration_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    ci_type VARCHAR(100) NOT NULL REFERENCES ci_type_definitions(name),
    attributes JSONB NOT NULL DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    -- ...
);
```

#### 2. Relationship Modeling
```sql
-- Relationships (Relational + Graph)
CREATE TABLE relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES configuration_items(id),
    target_id UUID NOT NULL REFERENCES configuration_items(id),
    relationship_type VARCHAR(50) NOT NULL,
    attributes JSONB DEFAULT '{}',
    -- ...
);
```

### Data Migration Strategy
- **Versioned Migrations**: Sequential migration files
- **Rollback Support**: Down migration scripts
- **Data Validation**: Post-migration integrity checks
- **Zero Downtime**: Online schema changes where possible

## API Design and Patterns

### RESTful API Design

#### 1. Resource Modeling
```
/api/v1/
├── auth/                   # Authentication endpoints
│   ├── POST /login
│   ├── POST /refresh
│   └── GET /me
├── users/                  # User management
│   ├── GET /users
│   ├── POST /users
│   ├── GET /users/{id}
│   ├── PUT /users/{id}
│   └── DELETE /users/{id}
├── ci-types/              # CI type definitions
│   ├── GET /ci-types
│   ├── POST /ci-types
│   ├── GET /ci-types/{id}
│   ├── PUT /ci-types/{id}
│   └── DELETE /ci-types/{id}
├── ci/                    # Configuration items
│   ├── GET /ci
│   ├── POST /ci
│   ├── GET /ci/{id}
│   ├── PUT /ci/{id}
│   ├── DELETE /ci/{id}
│   └── GET /ci/search
├── relationships/         # CI relationships
│   ├── GET /relationships
│   ├── POST /relationships
│   ├── GET /relationships/{id}
│   ├── PUT /relationships/{id}
│   └── DELETE /relationships/{id}
├── relationship-types/    # Relationship type definitions
│   ├── GET /relationship-types
│   ├── POST /relationship-types
│   ├── GET /relationship-types/{id}
│   ├── PUT /relationship-types/{id}
│   ├── DELETE /relationship-types/{id}
│   └── POST /relationship-types/validate
├── graph/                 # Graph visualization
│   ├── GET /graph
│   └── GET /graph/explore
└── audit/                 # Audit logs
    ├── GET /audit
    ├── GET /audit/{id}
    ├── GET /audit/stats
    └── DELETE /audit/{id}
```

#### 2. HTTP Status Codes
- `200 OK`: Successful GET, PUT
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Validation errors
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflicts
- `422 Unprocessable Entity`: Business logic violations
- `500 Internal Server Error`: Server errors

#### 3. Response Format
```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  },
  "errors": [
    {
      "field": "name",
      "message": "Name is required"
    }
  ]
}
```

### API Patterns

#### 1. Pagination
```go
// Query parameters
GET /ci?page=1&limit=20&sort=name&order=asc

// Response structure
type CIListResponse struct {
    CIs        []ConfigurationItem `json:"cis"`
    Page       int                 `json:"page"`
    Limit      int                 `json:"limit"`
    Total      int64               `json:"total"`
    TotalPages int                 `json:"total_pages"`
}
```

#### 2. Filtering and Search
```go
// Search parameters
GET /ci?search=server&ci_type=Server&tags=production,web

// Filter structure
type ListCIFilters struct {
    CIType   string   `json:"ci_type,omitempty"`
    Search   string   `json:"search,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Sort     string   `json:"sort,omitempty"`
    Order    string   `json:"order,omitempty"`
}
```

#### 3. Validation
```go
// Request validation with detailed errors
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Attribute validation against CI type definitions
func (ciType *CITypeDefinition) ValidateAttributes(attributes map[string]interface{}) []ValidationError
```

## Security Architecture

### Authentication & Authorization

#### 1. JWT Authentication
```go
type JWTClaims struct {
    UserID      uuid.UUID   `json:"user_id"`
    Username    string      `json:"username"`
    Roles       []string    `json:"roles"`
    Permissions []string    `json:"permissions"`
    jwt.RegisteredClaims
}
```

**Features**:
- Access tokens (24h TTL)
- Refresh tokens (7d TTL)
- Token rotation
- Automatic refresh

#### 2. RBAC System
```go
// Permission format: resource:action
// Examples: ci:read, user:create, relationship:delete

// Role-based permission checking
func (u *User) HasPermission(permission string) bool
func (u *User) HasRole(role string) bool
func (u *User) HasAnyPermission(permissions ...string) bool
```

**Default Roles**:
- `admin`: Full system access
- `editor`: CI and relationship management
- `viewer`: Read-only access

#### 3. Security Middleware
```go
// Middleware chain for security
router.Use(middleware.JWTAuth(jwtService))        // Authentication
router.Use(middleware.RBAC("ci:read"))            // Authorization
router.Use(middleware.AuditLogging(rbacService))   // Audit logging
router.Use(middleware.ActivityTracker(rbacService)) // Activity tracking
```

### Data Security

#### 1. Password Security
- **Hashing**: Argon2ID (memory-hard, GPU-resistant)
- **Salting**: Automatic salt generation
- **Complexity Requirements**: Configurable minimum length

#### 2. Data Protection
- **Encryption**: TLS 1.3 for all communications
- **Input Validation**: Comprehensive input sanitization
- **SQL Injection Prevention**: Parameterized queries
- **XSS Prevention**: Output encoding in frontend

#### 3. Audit Trail
```go
type AuditLog struct {
    ID         uuid.UUID            `json:"id"`
    EntityType string               `json:"entity_type"`
    EntityID   *uuid.UUID           `json:"entity_id"`
    Action     string               `json:"action"`
    PerformedBy uuid.UUID           `json:"performed_by"`
    Timestamp  time.Time            `json:"timestamp"`
    Details    map[string]interface{} `json:"details"`
    IPAddress  *string              `json:"ip_address"`
    UserAgent  *string              `json:"user_agent"`
}
```

## Data Models and Schemas

### Core Data Models

#### 1. Configuration Item
```go
type ConfigurationItem struct {
    ID        uuid.UUID            `json:"id"`
    Name      string               `json:"name"`
    CIType    string               `json:"ci_type"`
    Attributes map[string]interface{} `json:"attributes"`
    Tags      []string             `json:"tags"`
    CreatedAt time.Time            `json:"created_at"`
    UpdatedAt time.Time            `json:"updated_at"`
    CreatedBy uuid.UUID            `json:"created_by"`
    UpdatedBy *uuid.UUID           `json:"updated_by"`
}
```

#### 2. CI Type Definition
```go
type CITypeDefinition struct {
    ID                 uuid.UUID              `json:"id"`
    Name               string                 `json:"name"`
    Description        *string                `json:"description"`
    RequiredAttributes []AttributeDefinition `json:"required_attributes"`
    OptionalAttributes []AttributeDefinition `json:"optional_attributes"`
    CreatedBy         uuid.UUID              `json:"created_by"`
    CreatedAt         time.Time              `json:"created_at"`
    UpdatedAt         time.Time              `json:"updated_at"`
}

type AttributeDefinition struct {
    Name        string               `json:"name"`
    Type        string               `json:"type"`        // string, integer, boolean, array, object
    Description string               `json:"description"`
    Validation  *AttributeValidation `json:"validation,omitempty"`
}
```

#### 3. Relationship
```go
type Relationship struct {
    ID               uuid.UUID            `json:"id"`
    SourceID         uuid.UUID            `json:"source_id"`
    TargetID         uuid.UUID            `json:"target_id"`
    RelationshipType string               `json:"relationship_type"`
    Attributes       map[string]interface{} `json:"attributes"`
    CreatedAt        time.Time            `json:"created_at"`
    UpdatedAt        *time.Time           `json:"updated_at"`
    CreatedBy        uuid.UUID            `json:"created_by"`
    UpdatedBy        *uuid.UUID           `json:"updated_by"`
}
```

#### 4. Relationship Type
```go
type RelationshipType struct {
    ID                 uuid.UUID            `json:"id"`
    Name               string               `json:"name"`
    DisplayName        *string              `json:"display_name"`
    Description        *string              `json:"description"`
    ForwardLabel       string               `json:"forward_label"`
    ReverseLabel       string               `json:"reverse_label"`
    Color              *string              `json:"color"`
    Icon               *string              `json:"icon"`
    Category           *string              `json:"category"`
    IsActive           bool                 `json:"is_active"`
    IsSystem           bool                 `json:"is_system"`
    AllowedSourceTypes []string             `json:"allowed_source_types"`
    AllowedTargetTypes []string             `json:"allowed_target_types"`
    CardinalitySource  string               `json:"cardinality_source"`  // one, many
    CardinalityTarget  string               `json:"cardinality_target"`  // one, many
    Bidirectional      bool                 `json:"bidirectional"`
    Attributes         map[string]interface{} `json:"attributes"`
    CreatedAt          time.Time            `json:"created_at"`
    UpdatedAt          *time.Time           `json:"updated_at"`
    CreatedBy          uuid.UUID            `json:"created_by"`
    UpdatedBy          *uuid.UUID           `json:"updated_by"`
}
```

### Validation Rules

#### 1. Attribute Validation
```go
type AttributeValidation struct {
    Pattern    string      `json:"pattern,omitempty"`    // Regex pattern
    MinLength  *int        `json:"min_length,omitempty"`  // Min length for strings/arrays
    MaxLength  *int        `json:"max_length,omitempty"`  // Max length for strings/arrays
    Min        *int        `json:"min,omitempty"`         // Min value for numbers
    Max        *int        `json:"max,omitempty"`         // Max value for numbers
    Enum       []string    `json:"enum,omitempty"`        // Allowed values
    Format     string      `json:"format,omitempty"`      // email, url, ipv4, date, datetime
}
```

#### 2. Built-in Validators
- **String**: Pattern matching, length constraints
- **Integer**: Range validation
- **Boolean**: Type checking
- **Array**: Length constraints, enum validation
- **Object**: Property count constraints
- **Format**: Email, URL, IPv4, Date, DateTime

## Development Patterns and Conventions

### Backend Patterns

#### 1. Error Handling
```go
// Standardized error response
type APIError struct {
    Error struct {
        Code    string                 `json:"code"`
        Message string                 `json:"message"`
        Details map[string]interface{} `json:"details,omitempty"`
    } `json:"error"`
}

// Consistent error handling in handlers
if err != nil {
    logger.Error().Err(err).Msg("Failed to create CI")
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
```

#### 2. Repository Pattern
```go
type CIRepository interface {
    Create(ctx context.Context, ci *ConfigurationItem) error
    GetByID(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error)
    List(ctx context.Context, filters ListCIFilters) (*CIListResponse, error)
    Update(ctx context.Context, ci *ConfigurationItem) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type postgresCIRepository struct {
    db     *pgxpool.Pool
    logger *pustakaLogger.Logger
}
```

#### 3. Service Layer Pattern
```go
type CIService struct {
    repo          CIRepository
    neo4jService  *Neo4jService
    redisClient   *redis.Client
    auditService  *AuditService
    logger        *pustakaLogger.Logger
}

func (s *CIService) CreateCI(ctx context.Context, req *CreateCIRequest, userID uuid.UUID) (*ConfigurationItem, error) {
    // Business logic
    // Validation
    // Repository calls
    // Audit logging
    // Cache updates
}
```

#### 4. Context Usage
```go
// Proper context propagation
func (s *CIService) CreateCI(ctx context.Context, req *CreateCIRequest, userID uuid.UUID) (*ConfigurationItem, error) {
    // Use context for all external calls
    ci, err := s.repo.Create(ctx, ci)
    if err != nil {
        return nil, fmt.Errorf("failed to create CI: %w", err)
    }

    // Audit with context
    err = s.auditService.CreateAuditLog(ctx, "ci", &ci.ID, "create", userID, details, "", "")
    return ci, err
}
```

### Frontend Patterns

#### 1. Composition API Pattern
```typescript
// Composable function for CI management
export function useCI() {
  const store = useCIStore()
  const { isLoading, error } = storeToRefs(store)

  const createCI = async (ciData: CreateCIData) => {
    try {
      await store.createCI(ciData)
      showToast('CI created successfully', 'success')
    } catch (error) {
      showToast('Failed to create CI', 'error')
      throw error
    }
  }

  return {
    createCI,
    isLoading,
    error
  }
}
```

#### 2. Reactive State Management
```typescript
// Pinia store with reactivity
export const useCIStore = defineStore('ci', () => {
  const cis = ref<ConfigurationItem[]>([])
  const currentCI = ref<ConfigurationItem | null>(null)
  const loading = ref(false)

  const fetchCIs = async (filters?: CIFilters) => {
    loading.value = true
    try {
      const response = await ciAPI.list(filters)
      cis.value = response.data.cis
    } finally {
      loading.value = false
    }
  }

  return {
    cis: readonly(cis),
    currentCI: readonly(currentCI),
    loading: readonly(loading),
    fetchCIs
  }
})
```

#### 3. Error Boundaries
```vue
<template>
  <div v-if="error" class="error-boundary">
    <h2>Something went wrong</h2>
    <p>{{ error.message }}</p>
    <button @click="retry">Retry</button>
  </div>
  <slot v-else />
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'

const error = ref<Error | null>(null)

onErrorCaptured((err) => {
  error.value = err
  return false
})

const retry = () => {
  error.value = null
}
</script>
```

### Code Quality Standards

#### 1. Go Code Style
- **gofmt**: Consistent formatting
- **golangci-lint**: Comprehensive linting
- **Package Naming**: Lowercase, short, descriptive
- **Interface Naming**: -er suffix (Reader, Writer, Service)
- **Error Handling**: Explicit error checking and propagation

#### 2. TypeScript Code Style
- **ESLint**: Code quality and consistency
- **Prettier**: Code formatting
- **Type Safety**: Strict TypeScript configuration
- **Component Naming**: PascalCase for components
- **File Organization**: Feature-based grouping

#### 3. Testing Standards
- **Unit Tests**: 80%+ code coverage
- **Integration Tests**: API endpoint testing
- **E2E Tests**: Critical user workflows
- **Test Naming**: Descriptive test names
- **Test Organization**: Table-driven tests for Go

## Deployment Architecture

### Container Strategy

#### 1. Docker Compose Services
```yaml
services:
  postgres:    # Primary database
  neo4j:       # Graph database
  redis:       # Cache and session store
  api:         # Go backend service
  frontend:    # Vue.js frontend
```

#### 2. Service Configuration
- **Environment Variables**: Configuration via .env files
- **Health Checks**: Service readiness checks
- **Resource Limits**: Memory and CPU constraints
- **Volume Mounts**: Persistent data storage
- **Network Isolation**: Internal service communication

#### 3. Development vs Production
```yaml
# Development
docker-compose.yml
# Production overrides
docker-compose.prod.yml
```

### Infrastructure Considerations

#### 1. Scalability
- **Horizontal Scaling**: Multiple API instances
- **Load Balancing**: Nginx or cloud load balancer
- **Database Pooling**: Connection pool management
- **Caching Strategy**: Multi-level caching

#### 2. High Availability
- **Database Replication**: PostgreSQL streaming replication
- **Service Redundancy**: Multiple instances of each service
- **Health Monitoring**: Service health checks
- **Graceful Degradation**: Fallback mechanisms

#### 3. Security
- **Network Security**: Internal service networks
- **Secret Management**: Environment variable encryption
- **SSL/TLS**: Encrypted communication
- **Access Control**: Limited service exposure

## Testing Strategy

### Testing Pyramid

#### 1. Unit Tests (70%)
- **Backend**: Go testing package with testify
- **Frontend**: Vitest with Vue Test Utils
- **Coverage Target**: 80%+ code coverage
- **Focus**: Business logic, data models, utilities

```go
// Go unit test example
func TestCIService_CreateCI(t *testing.T) {
    tests := []struct {
        name    string
        request *CreateCIRequest
        wantErr bool
    }{
        {
            name: "valid CI creation",
            request: &CreateCIRequest{
                Name:   "test-server",
                CIType: "Server",
                Attributes: map[string]interface{}{
                    "hostname": "test-server.example.com",
                    "ip_address": "192.168.1.100",
                },
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 2. Integration Tests (20%)
- **API Testing**: HTTP endpoint testing
- **Database Testing**: Repository layer testing
- **Service Integration**: Cross-service communication
- **Test Containers**: Docker-based test databases

```go
// Integration test example
func TestCIHandler_CreateCI(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Setup test server
    router := setupTestRouter(db)

    // Test request
    req := httptest.NewRequest("POST", "/api/v1/ci", bytes.NewReader(ciJSON))
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assertions
    assert.Equal(t, http.StatusCreated, w.Code)
    // ... more assertions
}
```

#### 3. E2E Tests (10%)
- **User Workflows**: Critical user journeys
- **Cross-browser Testing**: Multiple browser support
- **Visual Regression**: UI consistency testing
- **Performance Testing**: Load and stress testing

```typescript
// Cypress E2E test example
describe('CI Management', () => {
  beforeEach(() => {
    cy.login('admin', 'password')
    cy.visit('/ci')
  })

  it('should create a new CI', () => {
    cy.get('[data-testid="create-ci-btn"]').click()
    cy.get('[data-testid="ci-name-input"]').type('Test Server')
    cy.get('[data-testid="ci-type-select"]').select('Server')
    cy.get('[data-testid="ci-form-submit"]').click()

    cy.get('[data-testid="notification"]').should('contain', 'CI created successfully')
    cy.get('[data-testid="ci-list"]').should('contain', 'Test Server')
  })
})
```

### Testing Infrastructure

#### 1. Test Database
- **Docker Containers**: Isolated test databases
- **Migrations**: Automated schema setup
- **Fixtures**: Test data management
- **Cleanup**: Automatic test cleanup

#### 2. Mock Services
- **External APIs**: Mock external dependencies
- **Database Mocks**: Repository layer mocking
- **HTTP Mocks**: Service-to-service communication

#### 3. CI/CD Integration
- **GitHub Actions**: Automated test execution
- **Parallel Testing**: Test suite parallelization
- **Test Reporting**: Coverage and result reporting
- **Quality Gates**: Test failure blocking deployment

## Performance Considerations

### Database Optimization

#### 1. PostgreSQL Performance
- **Indexing Strategy**: GIN indexes for JSONB, B-tree for common queries
- **Query Optimization**: EXPLAIN ANALYZE for slow queries
- **Connection Pooling**: pgxpool for efficient connections
- **Read Replicas**: Query distribution for read-heavy workloads

```sql
-- Example indexes
CREATE INDEX idx_cis_type ON configuration_items(ci_type);
CREATE INDEX idx_cis_attributes ON configuration_items USING GIN(attributes);
CREATE INDEX idx_cis_name_fulltext ON configuration_items USING GIN(to_tsvector('english', name));
```

#### 2. Neo4j Performance
- **Index Management**: Automatic and manual indexes
- **Query Optimization**: Cypher query profiling
- **Memory Configuration**: Heap and page cache tuning
- **Connection Pooling**: Driver connection pooling

```cypher
-- Create indexes for performance
CREATE INDEX ci_id_index FOR (ci:ConfigurationItem) ON (ci.id);
CREATE INDEX ci_type_index FOR (ci:ConfigurationItem) ON (ci.type);
```

#### 3. Caching Strategy
- **Application-level Cache**: In-memory caching with Redis
- **Database Query Cache**: PostgreSQL query cache
- **HTTP Cache**: ETag and Cache-Control headers
- **CDN Integration**: Static asset distribution

### Application Performance

#### 1. Go Performance
- **Goroutine Management**: Proper goroutine lifecycle
- **Memory Management**: Reduce allocations, use object pools
- **Profiling**: pprof for performance analysis
- **Concurrency**: Channel-based communication

#### 2. Frontend Performance
- **Code Splitting**: Lazy loading routes and components
- **Tree Shaking**: Eliminate unused code
- **Asset Optimization**: Image compression, minification
- **Browser Caching**: Service worker implementation

#### 3. Network Performance
- **HTTP/2**: Multiplexed requests
- **Compression**: Gzip response compression
- **Keep-alive**: Persistent connections
- **CDN**: Edge content delivery

### Monitoring and Metrics

#### 1. Application Metrics
- **Request Metrics**: Response time, error rate, throughput
- **Business Metrics**: CI creation rate, user activity
- **System Metrics**: CPU, memory, disk usage
- **Database Metrics**: Connection pool, query performance

#### 2. Alerting
- **Threshold Alerts**: Performance degradation notifications
- **Error Rate Alerts**: High error rate detection
- **Availability Alerts**: Service downtime monitoring
- **Capacity Alerts**: Resource utilization warnings

## Monitoring and Observability

### Logging Strategy

#### 1. Structured Logging
```go
// Zerolog structured logging
logger.Info().
    Str("user_id", userID.String()).
    Str("action", "create_ci").
    Str("ci_id", ciID.String()).
    Dur("duration", time.Since(start)).
    Msg("CI created successfully")
```

#### 2. Log Levels
- **TRACE**: Detailed debugging information
- **DEBUG**: Development debugging
- **INFO**: General information (default)
- **WARN**: Warning conditions
- **ERROR**: Error conditions
- **FATAL**: Critical errors

#### 3. Log Aggregation
- **Centralized Logging**: ELK stack or similar
- **Log Rotation**: Automatic log file management
- **Log Retention**: Configurable retention policies
- **Log Analysis**: Search and visualization

### Distributed Tracing

#### 1. Request Tracing
- **Request IDs**: Unique request identifiers
- **Span Tracking**: Operation timing
- **Service Communication**: Cross-service tracing
- **Performance Analysis**: Bottleneck identification

#### 2. Error Tracking
- **Error Context**: Detailed error information
- **Stack Traces**: Debugging information
- **User Impact**: Affected user tracking
- **Recovery Patterns**: Error resolution tracking

### Health Monitoring

#### 1. Health Checks
```go
// Health check endpoint
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().UTC(),
        "version": "1.0.0",
        "checks": map[string]interface{}{
            "database": h.checkDatabase(),
            "neo4j": h.checkNeo4j(),
            "redis": h.checkRedis(),
        },
    }

    json.NewEncoder(w).Encode(status)
}
```

#### 2. Metrics Collection
- **Prometheus Integration**: Metrics exposition
- **Custom Metrics**: Business-specific metrics
- **System Metrics**: Infrastructure monitoring
- **Application Metrics**: Performance indicators

This comprehensive technical documentation provides a thorough understanding of the Pustaka CMDB architecture, patterns, and implementation details. It serves as a reference for developers, architects, and operations teams working with the system.
# Architecture

**Analysis Date:** 2026-02-20

## Pattern Overview

**Overall:** Clean Architecture with Domain-Driven Design principles

**Key Characteristics:**
- Layered architecture with clear separation of concerns
- Multi-database approach: PostgreSQL + Neo4j + Redis
- API-first design with RESTful endpoints
- Event-driven audit logging system
- Frontend-backend separation with standard API contract
- RBAC-based security with JWT authentication

## Layers

**Presentation Layer (Backend):**
- Purpose: HTTP request handling and response formatting
- Location: `/internal/api/`
- Contains: HTTP handlers, middleware, request/response models
- Depends on: Service layer, Auth layer
- Used by: External HTTP clients, Frontend

**Service Layer:**
- Purpose: Business logic and domain rules
- Location: `/internal/ci/`, `/internal/amortization/`, `/internal/auth/`
- Contains: Domain services, business rules, validation
- Depends on: Repository layer, Database layer
- Used by: Presentation layer

**Repository Layer:**
- Purpose: Data access abstraction
- Location: `/internal/ci/repository.go`, `/internal/database/`
- Contains: Data access interfaces and implementations
- Depends on: Database drivers
- Used by: Service layer

**Domain Layer:**
- Purpose: Core business entities and models
- Location: `/internal/ci/models.go`
- Contains: Configuration Items, Relationships, CI Types
- Depends on: None (core domain)
- Used by: All layers above

**Infrastructure Layer:**
- Purpose: External integrations and technical concerns
- Location: `/internal/database/`, `/internal/config/`
- Contains: Database connections, external service clients
- Depends on: External libraries and services
- Used by: Repository layer

## Data Flow

**Request Processing Flow:**

1. **HTTP Request** → Chi router → Middleware stack
2. **Authentication** → JWT validation → RBAC checks
3. **Validation** → Request body parsing → Business rules
4. **Service Call** → Business logic execution
5. **Data Access** → Repository → Database operations
6. **Response Formatting** → JSON marshaling → HTTP response

**State Management:**
- **Frontend:** Vue 3 + Pinia for reactive state management
- **Backend:** Service instances with dependency injection
- **Database:** PostgreSQL ACID transactions, Neo4j ACID transactions, Redis caching
- **Session Management:** JWT access tokens (24h) + refresh tokens (7d)

## Key Abstractions

**ConfigurationItem:**
- Purpose: Core entity representing any IT asset
- Examples: `[internal/ci/models.go:11]`
- Pattern: Entity with attributes map for flexibility

**Service:**
- Purpose: Business logic coordination
- Examples: `[internal/ci/service.go:15]`, `[internal/amortization/service.go]`
- Pattern: Repository pattern with dependencies injected

**Repository:**
- Purpose: Data access abstraction
- Examples: `[internal/ci/repository.go]`, `[internal/database/postgres.go]`
- Pattern: Interface-based with concrete implementations

**Middleware:**
- Purpose: Cross-cutting concerns
- Examples: `[internal/api/middleware/jwt.go]`, `[internal/api/middleware/rbac.go]`
- Pattern: Chain of responsibility

## Entry Points

**Backend API Server:**
- Location: `[cmd/api/main.go:101]`
- Triggers: HTTP requests, system startup
- Responsibilities: Server initialization, database connections, service setup, routing

**Frontend Application:**
- Location: `[web/src/main.ts:1]`
- Triggers: User navigation, page loads
- Responsibilities: UI rendering, state management, API communication

**Database Migrations:**
- Location: `cmd/migrations/`
- Triggers: Schema changes, deployment
- Responsibilities: Database schema evolution

**Scheduled Jobs:**
- Location: `[internal/amortization/service.go:250]`
- Triggers: Daily scheduler at 00:00
- Responsibilities: Automated amortization calculations, cleanup tasks

## Error Handling

**Strategy:** Structured error responses with proper HTTP status codes

**Patterns:**
- Service validation errors with detailed field information
- Database errors wrapped with context
- Authentication/authorization errors with appropriate HTTP 401/403
- Global error middleware for consistent error formatting

## Cross-Cutting Concerns

**Logging:** Zerolog structured logging with request context
**Validation:** Attribute-based validation with CI type schemas
**Authentication:** JWT tokens with refresh mechanism
**Authorization:** RBAC with granular permissions (e.g., `ci:read`, `user:create`)
**Audit Trail:** Comprehensive logging of all changes with user context
**Caching:** Redis for frequently accessed data (CI types, relationship types)

---

*Architecture analysis: 2026-02-20*
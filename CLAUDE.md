# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pustaka is a modern, open-source Configuration Management Database (CMDB) built for enterprise IT asset management. It features a hierarchical taxonomy system (Domains → Categories → Subcategories → Configuration Items), relationship mapping, RBAC, audit logging, and graph visualization.

**Architecture:**
- **Backend**: Go with Chi v5 framework, PostgreSQL for structured data, Neo4j for relationships, Redis for caching
- **Frontend**: Vue 3 + TypeScript with Pinia state management, vis-network for graph visualization
- **Deployment**: Docker Compose with multi-service orchestration

## Development Commands

### Backend (Go)
```bash
# Build and run
make build                    # Build API binary to bin/api
make run                      # Run API server directly
go run cmd/api/main.go       # Alternative way to run API

# Testing
make test                     # Run all Go tests
go test ./...                 # Run tests directly
go test ./internal/ci/        # Run specific package tests
go test -v ./internal/ci/     # Run tests with verbose output
go test -cover ./internal/ci/ # Run tests with coverage report

# Development
make dev                      # Start full dev environment (Docker + migrations)
make docker-up                # Start only Docker services
make docker-down              # Stop Docker services

# Code quality
make fmt                      # Format Go code
make lint                     # Run golangci-lint
make security                 # Run gosec security scan

# Database
make migrate                  # Run database migrations (requires golang-migrate)
make create-admin             # Create initial admin user (curl command)

# Note on migrations: Install golang-migrate from https://github.com/golang-migrate/migrate
# Migration files are located in cmd/migrations/

# Dependencies
make deps                     # Install/update Go modules
go mod tidy                   # Clean up modules
```

### Frontend (Vue.js)
```bash
cd web/
npm run dev                   # Start development server (Vite)
npm run build                 # Build for production
npm run preview               # Preview production build

# Testing
npm run test                  # Run Vitest unit tests
npm run test:ui               # Run tests with UI
npm run test:e2e              # Run Cypress E2E tests
npm run test:e2e:playwright   # Run Playwright E2E tests

# Code quality
npm run lint                  # Run ESLint with auto-fix
npm run format                # Format code with Prettier
```

### Docker Services
- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **Neo4j Browser**: http://localhost:7474
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

Default credentials (from `.env.example`):
- PostgreSQL: pustaka/password
- Neo4j: neo4j/password
- Admin user: admin/Admin@123 (created automatically on startup)

## Key Architecture Components

### Backend Structure
```
cmd/api/main.go               # Main entrypoint with server initialization
internal/
├── api/                      # HTTP handlers and routing
│   ├── handlers/             # Specific request handlers (auth, users, CI)
│   ├── middleware/           # Custom middleware (JWT, RBAC, CORS, logging)
│   └── *_test.go             # Integration tests for API endpoints
├── auth/                     # Authentication & authorization
│   ├── jwt.go               # JWT token service
│   ├── rbac.go              # Role-based access control
│   ├── password.go          # Password hashing (Argon2ID)
│   └── *_test.go            # Tests for auth and RBAC logic
├── ci/                       # Core CMDB business logic
│   ├── models.go            # CI types and relationships
│   ├── service.go           # Business logic layer
│   ├── repository.go        # PostgreSQL data access
│   ├── neo4j_repository.go  # Neo4j relationship management
│   ├── audit_*.go           # Audit logging system
│   └── *_test.go            # Unit tests for CI services
├── database/                 # Database connection management
│   ├── postgres.go          # PostgreSQL connection pool
│   ├── neo4j.go             # Neo4j driver setup
│   └── redis.go             # Redis client setup
└── config/                   # Configuration loading
```

### Frontend Structure
```
web/src/
├── views/                    # Page components
│   ├── auth/                # Authentication pages
│   ├── ci/                  # CI management pages
│   ├── dashboard/           # Dashboard views
│   ├── users/               # User management
│   ├── audit/               # Audit log viewer
│   ├── graph/               # Graph visualization
│   ├── relationships/       # Relationship management
│   └── relationship-types/  # Relationship type configuration
├── stores/                   # Pinia state management
│   ├── auth.ts              # Authentication state
│   ├── ciTypes.ts           # CI types data
│   ├── relationshipTypes.ts # Relationship types data
│   └── notification.ts      # Global notifications
├── services/                 # API communication layer
│   └── api.ts               # Axios HTTP client setup
├── components/               # Reusable Vue components
├── router/                   # Vue Router configuration
└── types/                    # TypeScript type definitions
```

### Database Schema
- **PostgreSQL**: Users, roles, permissions, CI types, CIs, relationships, audit logs
- **Neo4j**: CI relationships with bidirectional graph connections and traversal
- **Redis**: Session storage, caching, rate limiting

### API Architecture
- **Base path**: `/api/v1`
- **Authentication**: JWT tokens with refresh mechanism (24h access, 7d refresh)
- **Authorization**: RBAC with granular permissions (e.g., `ci:read`, `user:create`, `system:admin`)
- **Middleware stack**: Request ID → Logging → CORS → JWT Auth → Activity Tracking → RBAC → Audit Logging
- **Key endpoints**:
  - `/auth/*` - Authentication (login, refresh)
  - `/ci/*` - Configuration Items CRUD
  - `/ci-types/*` - CI type management
  - `/relationships/*` - CI relationships CRUD
  - `/relationship-types/*` - Relationship type definitions
  - `/graph/*` - Graph visualization data
  - `/audit/*` - Audit logs and compliance
  - `/users/*` - User management (admin only)

## Development Patterns

### Authentication Flow
1. User logs in via `/auth/login` → JWT access + refresh tokens
2. Access token (24h) used for API requests via Authorization header
3. Refresh token (7d) used to get new access tokens via `/auth/refresh`
4. JWT middleware validates tokens and extracts user context
5. RBAC middleware checks permissions based on user roles

### CI Management
1. **Taxonomy**: Domain → Category → Subcategory → CI Type
2. **Relationships**: Stored in Neo4j with bidirectional connections
3. **Audit Trail**: All changes logged with user context and timestamps
4. **Caching**: Redis caching for frequently accessed data (CI types, relationship types)

### Error Handling
- Structured error responses with proper HTTP status codes
- Zerolog structured logging with request context
- Graceful degradation for external service dependencies
- Service-level error wrapping for better debugging

### Testing Strategy
- Unit tests for service layer business logic (internal/ci/, internal/auth/)
- Integration tests for database operations (internal/api/)
- E2E tests for critical user workflows (web/tests/)
- Test files follow Go convention: *_test.go
- Frontend tests in web/tests/ using Vitest

## Configuration

### Environment Setup
1. Copy `.env.example` to `.env` and update values
2. Minimum required: DATABASE_URL, NEO4J_URI, REDIS_URL, JWT_SECRET
3. Run `make dev` to start full environment
4. Admin user is created automatically on startup from config

### Key Configuration Files
- `.env` / `.env.example` - Environment variables
- `docker-compose.yml` - Multi-service orchestration
- `go.mod` - Go dependencies
- `web/package.json` - Frontend dependencies and scripts
- `Makefile` - Development commands automation

### Production Considerations
- Use strong JWT secrets (32+ chars)
- Configure proper CORS origins
- Enable production logging levels
- Set up database connection pooling
- Configure Redis persistence
- Migration files: cmd/migrations/001_initial_schema.sql

## Troubleshooting

### Common Issues
- **Database connection**: Ensure Docker services are running (`make docker-up`)
- **Migration failures**: Check PostgreSQL connection and user permissions
- **Neo4j connection**: Verify Neo4j auth credentials and APOC plugin configuration
- **CORS errors**: Update `CORS_ALLOWED_ORIGINS` in `.env`
- **JWT validation**: Check token expiration and secret configuration

### Debug Commands
```bash
# Check service status
docker-compose ps
docker-compose logs postgres
docker-compose logs neo4j
docker-compose logs api
docker-compose logs frontend

# Test API health
curl http://localhost:8080/health

# Verify database connections
docker exec -it pustaka-postgres psql -U pustaka -d pustaka -c "SELECT 1;"
docker exec -it pustaka-neo4j cypher-shell -u neo4j -p password "RETURN 1;"

# Run specific tests with coverage
go test -cover ./internal/ci/
go test -v -run TestSpecificFunction ./internal/ci/
```
- use docker compose down && docker compose up --build -d to rebuild the app after modifying the code
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Porta is a modern, open-source Configuration Management Database (CMDB) built for enterprise IT asset management. It features a hierarchical taxonomy system (Domains → Categories → Subcategories → Configuration Items), relationship mapping, RBAC, audit logging, and graph visualization.

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
make test-integration         # DB-backed suites (needs TEST_DATABASE_URL, see below)
go test ./...                 # Run tests directly
go test ./internal/ci/        # Run specific package tests
go test -v ./internal/ci/     # Run tests with verbose output
go test -cover ./internal/ci/ # Run tests with coverage report

# Development
make pull                     # Pull published multi-arch images from GHCR
make dev                      # Start full dev environment (pulls images + migrations)
make docker-up                # Start only Docker services (pulls missing images)
make images-build             # Rebuild images locally (escape hatch; CI normally builds)
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

Dependency bumps run in CI, not locally (the dev host is a low-power arm64
machine; no local Go toolchain is needed):

1. Actions → **go-deps** → *Run workflow* with `module@version` specifiers
   (space-separated; `@latest` works).
2. If the run reports "No dependency changes", stop here — there is no artifact.
3. `gh run download <run-id> -n go-deps-patch`
4. `git checkout -b chore/go-deps && git apply go-deps.patch`, commit, push the
   branch and **open a pull request**. Opening the PR triggers the required
   checks (build/test, CodeQL, dependency review, image builds); a plain branch
   push triggers nothing under this repo's workflow triggers.

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
- **Frontend**: http://localhost:3001
- **API**: http://localhost:8081
- **API Health**: http://localhost:8081/health
- **Neo4j Browser**: http://localhost:7475
- **PostgreSQL**: localhost:5433
- **Redis**: localhost:6380

Default credentials (from `.env.example`):
- PostgreSQL: porta/password
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
docker compose ps
docker compose logs postgres
docker compose logs neo4j
docker compose logs api
docker compose logs frontend

# Test API health
curl http://localhost:8081/health

# Verify database connections
docker exec -it porta-postgres psql -U porta -d porta -c "SELECT 1;"
docker exec -it porta-neo4j cypher-shell -u neo4j -p password "RETURN 1;"

# Database-backed suites (internal/auth, internal/ci) skip unless the target
# database name contains "test"; the helper creates tables and deletes rows:
#   TEST_DATABASE_URL=postgres://porta:password@localhost:5433/porta_test?sslmode=disable make test-integration

# Run specific tests with coverage
go test -cover ./internal/ci/
go test -v -run TestSpecificFunction ./internal/ci/
```
- images are built and published by GitHub Actions (multi-arch, GHCR); the host only pulls: `docker compose pull && docker compose up -d`
- use `make images-build` (`docker compose -f docker-compose.yml -f docker-compose.build.yml build`) only when you intentionally rebuild locally — it is slow on low-memory hosts
- use agents if necessary
- always assume app runs on docker, so if you want to know the state of the app use docker ps
- use agents to help your tasks if necessary

## Pending Feature: IT Asset Amortization Module

**Status:** Planned, ready for implementation

**Documentation:**
- PRD: `docs/prd-amortization-module.md`
- Implementation Plan: `docs/implementation-plan-amortization-module.md`

**Summary:** Automated depreciation tracking for IT assets with status-aware financial calculations, ledger-based audit trail, and monthly scheduler.

**Phase 1 Scope (MVP):**
1. Database migrations:
   - Add `is_amortizable` to `ci_type_definitions`
   - Add `amortization_behavior` (pending/active/terminal) to `lifecycle_statuses`
   - Add financial columns to `configuration_items` (purchase_cost, salvage_value, amort_start_date, useful_life_months, current_book_value)
   - Create `amortization_ledger` and `amortization_runs` tables

2. Backend:
   - New `internal/amortization/` package (models, repository, service, scheduler)
   - New handlers at `internal/api/handlers/amortization_handlers.go`
   - Modify CI service to detect terminal status changes and trigger write-offs

3. Frontend:
   - CI Type form: Add "Is Amortizable" checkbox
   - Lifecycle Status form: Add "Amortization Behavior" dropdown
   - CI Edit: Add "Financials" tab (visible only for amortizable CI types)

**Key Design Decisions:**
- Single currency via `AMORTIZATION_CURRENCY` env variable
- Daily scheduler at 00:00 with checkpoint mechanism (not month-end)
- Prospective recalculation when useful life changes (history preserved)
- Adjustment entries for corrections (never delete ledger records)
- Write-offs and adjustments logged to main audit_logs table

**To start implementation:** Run `implement amortization module phase 1` or follow the implementation plan in docs/

## Agent skills

### Issue tracker

GitHub Issues in `tahopetis/porta` (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-role vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` at the repo root and ADRs in `docs/adr/`. See `docs/agents/domain.md`.

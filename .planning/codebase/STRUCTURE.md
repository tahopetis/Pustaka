# Codebase Structure

**Analysis Date:** 2026-02-20

## Directory Layout

```
pustaka/
├── cmd/                    # Application entry points
│   └── api/               # Main API server
│       └── main.go       # Server initialization and routing
├── internal/              # Private application code
│   ├── api/              # HTTP layer
│   │   ├── handlers/     # Request handlers
│   │   └── middleware/   # HTTP middleware
│   ├── auth/             # Authentication & authorization
│   ├── ci/               # Core CMDB business logic
│   ├── amortization/      # Financial amortization module
│   ├── config/           # Configuration management
│   ├── database/         # Database connections
│   └── testutils/        # Testing utilities
├── web/                  # Frontend Vue.js application
│   ├── src/
│   │   ├── views/        # Page components
│   │   ├── stores/       # Pinia state management
│   │   ├── services/     # API communication
│   │   ├── components/   # Reusable UI components
│   │   ├── router/       # Vue Router configuration
│   │   └── types/        # TypeScript definitions
│   ├── tests/           # Frontend tests
│   └── package.json      # Frontend dependencies
├── docs/                # Documentation
├── docker/              # Docker configuration
├── scripts/             # Build and deployment scripts
├── tests/               # Integration tests
└── .planning/           # Generated analysis documents
```

## Directory Purposes

**cmd/api/:**
- Purpose: Main application entry point
- Contains: Server initialization, routing, service setup
- Key files: `[cmd/api/main.go]`

**internal/api/:**
- Purpose: HTTP layer implementation
- Contains: Request handlers, middleware, HTTP utilities
- Key files: `[internal/api/handlers/]`, `[internal/api/middleware/]`

**internal/ci/:**
- Purpose: Core CMDB business logic
- Contains: CI models, services, repositories, audit logging
- Key files: `[internal/ci/models.go]`, `[internal/ci/service.go]`, `[internal/ci/repository.go]`

**internal/amortization/:**
- Purpose: Financial amortization module
- Contains: Amortization models, services, scheduling
- Key files: `[internal/amortization/service.go]`, `[internal/amortization/models.go]`

**internal/auth/:**
- Purpose: Authentication and authorization
- Contains: JWT handling, RBAC, password management
- Key files: `[internal/auth/jwt.go]`, `[internal/auth/rbac.go]`

**internal/config/:**
- Purpose: Configuration management
- Contains: Environment-based configuration loading
- Key files: `[internal/config/config.go]`

**internal/database/:**
- Purpose: Database connection management
- Contains: PostgreSQL, Neo4j, Redis connections
- Key files: `[internal/database/postgres.go]`, `[internal/database/neo4j.go]`

**web/src/:**
- Purpose: Frontend application source
- Contains: Vue components, views, stores, services
- Key files: `[web/src/main.ts]`, `[web/src/router/index.ts]`

**web/src/views/:**
- Purpose: Page-level components
- Contains: Dashboard, CI management, relationships, etc.
- Key files: `[web/src/views/DashboardView.vue]`, `[web/src/views/ci/CIListView.vue]`

**web/src/stores/:**
- Purpose: Pinia state management
- Contains: Authentication, CI data, notifications
- Key files: `[web/src/stores/auth.ts]`, `[web/src/stores/ciTypes.ts]`

**web/src/services/:**
- Purpose: API communication layer
- Contains: HTTP client setup, API calls
- Key files: `[web/src/services/api.ts]`

## Key File Locations

**Entry Points:**
- `[cmd/api/main.go]`: Main API server with Chi router
- `[web/src/main.ts]`: Vue.js application entry point
- `[web/src/router/index.ts]`: Frontend routing configuration

**Configuration:**
- `[internal/config/config.go]`: Configuration loading and validation
- `.env.example`: Environment variable template

**Core Logic:**
- `[internal/ci/models.go]`: Core domain models
- `[internal/ci/service.go]`: Business logic services
- `[internal/ci/repository.go]`: Data access layer

**HTTP Layer:**
- `[internal/api/handlers/]`: Route handlers for all endpoints
- `[internal/api/middleware/]`: Authentication, CORS, logging, RBAC

**Frontend Views:**
- `[web/src/views/DashboardView.vue]`: Main dashboard
- `[web/src/views/ci/]:` Configuration Item management
- `[web/src/views/amortization/]:` Financial module UI

**Testing:**
- `[internal/ci/*_test.go]`: Unit tests for core logic
- `[web/tests/]:` Frontend component and E2E tests

## Naming Conventions

**Files:**
- Go: `snake_case.go` (e.g., `service.go`, `repository.go`)
- Vue: `PascalCase.vue` (e.g., `DashboardView.vue`, `CIListView.vue`)
- TypeScript: `camelCase.ts` (e.g., `auth.ts`, `api.ts`)

**Directories:**
- Go: `snake_case/` (e.g., `internal/api/`, `internal/auth/`)
- Vue: PascalCase/ (e.g., `web/src/views/`, `web/src/components/`)

**Variables and Functions:**
- Go: `camelCase` (e.g., `CreateCI`, `GetCITypeByName`)
- Vue/TypeScript: `camelCase` (e.g., `loadCIs`, `updateCI`)

**Components:**
- Vue: `PascalCase` with descriptive names (e.g., `CIList`, `AmortizationWidget`)
- Reusable components: Base category (e.g., `BaseButton`, `LayoutSidebar`)

## Where to Add New Code

**New Backend Feature:**
- API handlers: `/internal/api/handlers/`
- Business logic: `/internal/ci/` or new `/internal/[feature]/`
- Database models: Update `/internal/ci/models.go`
- Tests: Add `*_test.go` files

**New Frontend Feature:**
- Views: `/web/src/views/[feature]/`
- Components: `/web/src/components/[feature]/` or `/web/src/components/common/`
- Stores: `/web/src/stores/`
- Routes: Add to `/web/src/router/index.ts`

**New API Endpoint:**
1. Handler in `/internal/api/handlers/`
2. Route registration in `cmd/api/main.go`
3. Add middleware as needed (auth, RBAC, validation)
4. Add tests in handler file

**New Database Migration:**
- Create files in `/cmd/migrations/`
- Apply with `make migrate` command

**New Configuration Option:**
- Add to `/internal/config/config.go`
- Update defaults in `setDefaults()`
- Bind environment variable in `Load()`

## Special Directories

**.planning/:**
- Purpose: Generated analysis documents
- Contains: Architecture, structure, conventions docs
- Generated: Auto-created by codebase analysis

**internal/testutils/:**
- Purpose: Shared testing utilities
- Contains: Test helpers, fixtures, mock generators

**scripts/:**
- Purpose: Build and deployment automation
- Contains: Makefile targets, deployment scripts

---

*Structure analysis: 2026-02-20*
# Pustaka CMDB - Source Tree Analysis

## Annotated Directory Structure

```
pustaka/                                   # Project root (monorepo)
│
├── 📁 web/                               # Vue 3 Frontend SPA (Part: frontend)
│   ├── 📁 src/
│   │   ├── 📁 components/                 # Reusable UI components
│   │   ├── 📁 views/                     # Route-based page components
│   │   ├── 📁 stores/                    # Pinia state management
│   │   │   ├── auth.ts                   # Authentication state
│   │   │   ├── ciTypes.ts                # CI types data
│   │   │   └── notification.ts           # Global notifications
│   │   ├── 📁 services/                  # API client layer
│   │   │   └── api.ts                    # Axios HTTP client setup
│   │   ├── 📁 tests/                     # Frontend unit tests (Vitest)
│   │   └── 📁 router/                    # Vue Router configuration
│   ├── 📄 package.json                   # Frontend dependencies & scripts
│   ├── 📄 vite.config.ts                # Vite build configuration
│   └── 📄 tailwind.config.js            # Tailwind CSS configuration
│
├── 📁 internal/                          # Go Backend API (Part: backend)
│   ├── 📁 api/                           # HTTP handlers and routing
│   │   ├── 📁 handlers/                  # Specific request handlers
│   │   │   ├── auth.go                  # Authentication endpoints
│   │   │   ├── users.go                 # User management
│   │   │   └── ci.go                    # CI management endpoints
│   │   ├── 📁 middleware/                # Custom middleware
│   │   │   ├── jwt.go                   # JWT authentication
│   │   │   ├── rbac.go                  # Role-based access control
│   │   │   └── cors.go                  # CORS handling
│   │   └── 📄 *_test.go                  # API integration tests
│   ├── 📁 auth/                          # Authentication & authorization
│   │   ├── 📄 jwt.go                    # JWT token service
│   │   ├── 📄 rbac.go                   # RBAC implementation
│   │   └── 📄 password.go               # Password hashing (Argon2ID)
│   ├── 📁 ci/                            # Core CMDB business logic
│   │   ├── 📄 models.go                 # CI types and relationships
│   │   ├── 📄 service.go                # Business logic layer
│   │   ├── 📄 repository.go             # PostgreSQL data access
│   │   ├── 📄 neo4j_service.go          # Neo4j relationship management
│   │   └── 📄 audit_*.go                # Audit logging system
│   └── 📁 database/                     # Database connection management
│
├── 📁 cmd/                               # Application entry points
│   ├── 📁 api/                           # API server main application
│   │   └── 📄 main.go                   # Server initialization
│   └── 📁 migrations/                    # Database migration files
│       └── 📄 001_initial_schema.sql     # Initial database schema
│
├── 📁 pkg/                               # Shared libraries and utilities
│   └── 📁 testutils/                     # Testing utilities
│       └── 📄 database.go               # Database test helpers
│
├── 📁 docker/                            # Docker configuration files
│   ├── 📄 Dockerfile.api                 # Backend API container
│   └── 📄 Dockerfile.frontend          # Frontend container
│
├── 📁 plan/                              # Implementation plans (8 files)
│   ├── 📄 README_SETUP.md                # Environment setup guide
│   ├── 📄 RBAC_API_SPECIFICATION_PHASE1.md # RBAC API spec
│   ├── 📄 IMPLEMENTATION_PLAN.md        # Implementation roadmap
│   ├── 📄 FSD.md                         # Feature-driven design
│   ├── 📄 CODING_STANDARDS.md            # Development standards
│   ├── 📄 TSD.md                         # Technical specifications
│   ├── 📄 SECURITY_CHECKLIST.md          # Security requirements
│   └── 📄 TEST_PLAN.md                   # Testing strategy
│
├── 📁 docs/                              # Generated documentation (11+ files)
│   ├── 📁 architecture/                  # Architecture documentation
│   │   └── 📄 technical-documentation.md # Technical architecture
│   ├── 📁 planning/                      # Planning documents
│   │   ├── 📄 prd.md                     # Product requirements document
│   │   ├── 📄 gdd.md                     # Game design document
│   │   ├── 📄 epics-user-stories.md     # Epics and user stories
│   │   ├── 📄 ux-ui-specifications.md    # UX/UI specifications
│   │   └── 📄 implementation-roadmap.md  # Implementation roadmap
│   ├── 📄 api-guide.md                   # API usage guide
│   ├── 📄 api-examples.md                # API examples
│   ├── 📄 relationship_types_implementation.md # Relationship types
│   ├── 📄 technical-decisions-template.md # Decision template
│   └── 📄 bmm-workflow-status.md         # BMAD workflow status
│
├── 📁 tests/                             # Test suites
│   ├── 📁 e2e/                           # End-to-end tests
│   └── 📁 integration/                   # Integration tests
│
├── 📁 bmad/                              # BMAD workflow framework
│   └── 📁 bmm/                           # BMAD module configuration
│
├── 📁 configs/                           # Configuration files
├── 📁 scripts/                           # Utility scripts
│
├── 📄 go.mod                             # Go module dependencies
├── 📄 go.sum                             # Go dependency checksums
├── 📄 package.json                       # Root package.json (scripts)
├── 📄 docker-compose.yml                 # Multi-service orchestration
├── 📄 Makefile                           # Development commands automation
├── 📄 README.md                          # Project overview
├── 📄 CLAUDE.md                          # Claude development instructions
└── 📄 .env.example                       # Environment variables template
```

## Critical Folders Explained

### Frontend (web/)
- **components/**: Reusable Vue components with TypeScript support
- **views/**: Page-level components connected to router
- **stores/**: Pinia state management for global application state
- **services/**: API client abstraction layer with Axios
- **router/**: Vue Router configuration for SPA navigation

### Backend (internal/)
- **api/**: HTTP layer with handlers for REST API endpoints
- **auth/**: Authentication and authorization logic (JWT + RBAC)
- **ci/**: Core CMDB business logic and data models
- **database/**: Database connection and configuration management

### Infrastructure
- **docker/**: Container definitions for each service
- **cmd/**: Application entry points and main functions
- **configs/**: Environment-specific configurations

### Documentation & Planning
- **plan/**: Implementation specifications and requirements (8 files)
- **docs/**: Generated documentation and planning artifacts (11+ files)
- **bmad/**: BMAD workflow and project management framework

## Integration Points

### API Communication
```
Frontend (web/src/services/api.ts)
    ↓ HTTP/REST API calls
Backend API (internal/api/handlers/)
    ↓ Business logic
CMDB Services (internal/ci/service.go)
    ↓ Data persistence
[PostgreSQL, Neo4j, Redis]
```

### Service Dependencies
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │◄──►│     API     │◄──►│ Databases   │
│  Vue 3 SPA  │    │ Go REST API │    │ PostgreSQL  │
│  Port 3000  │    │ Port 8080   │    │ Neo4j       │
└─────────────┘    └─────────────┘    │ Redis       │
                                      └─────────────┘
```

## Entry Points

### Frontend Entry Point
- **File**: `web/src/main.ts` (inferred from Vue 3 + Vite setup)
- **Purpose**: Vue application initialization, router setup, Pinia store creation

### Backend Entry Point
- **File**: `cmd/api/main.go`
- **Purpose**: Server initialization, middleware setup, database connections, route registration

## Technology Integration

### Frontend Stack Integration
```
Vue 3 + TypeScript
├── Vite (build tool)
├── Pinia (state management)
├── Vue Router (routing)
├── Tailwind CSS (styling)
├── Axios (HTTP client)
├── vis-network (graph visualization)
└── @headlessui/vue (UI components)
```

### Backend Stack Integration
```
Go 1.22
├── Chi v5 (web framework)
├── pgx/v5 (PostgreSQL driver)
├── neo4j-go-driver/v5 (Neo4j driver)
├── go-redis/v9 (Redis driver)
├── golang-jwt/jwt/v5 (JWT authentication)
├── argon2id (password hashing)
└── zerolog (structured logging)
```

## Development Patterns

### Frontend Patterns
- **Composition API**: Modern Vue 3 reactive patterns
- **TypeScript**: Type-safe component development
- **Pinia**: Centralized state management with TypeScript support
- **Component Architecture**: Reusable, composable UI components

### Backend Patterns
- **Repository Pattern**: Clean data access abstraction
- **Service Layer**: Business logic separation from HTTP handlers
- **Middleware Chain**: Composable request processing pipeline
- **Configuration Management**: Environment-based configuration with Viper

### Database Patterns
- **Multi-Database Architecture**: PostgreSQL for structured data, Neo4j for relationships
- **Migration Management**: Version-controlled schema changes
- **Connection Pooling**: Efficient database connection management
- **Health Checking**: Service availability monitoring

## Key Architectural Decisions

1. **Separate Frontend/Backend**: Clear separation enables independent development and deployment
2. **Multi-Database Strategy**: PostgreSQL for relational data, Neo4j for graph relationships
3. **Docker-First Development**: Reproducible development environment
4. **JWT + RBAC**: Secure authentication with granular permissions
5. **TypeScript Everywhere**: Type safety across frontend and backend validation
6. **Comprehensive Documentation**: Extensive planning and technical documentation

---

**Analysis Date**: 2025-10-14
**Scan Level**: Quick (pattern-based with existing documentation integration)
**Focus Areas**: Flexible taxonomy system, graph clustering visualization
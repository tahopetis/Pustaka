# Pustaka CMDB - Project Overview

## Executive Summary

Pustaka is a modern, open-source Configuration Management Database (CMDB) built for enterprise IT asset management. It features a hierarchical taxonomy system, relationship mapping, role-based access control (RBAC), audit logging, and graph visualization capabilities. The system is designed as a brownfield project with comprehensive documentation and a flexible architecture supporting infinite-level taxonomies.

## Quick Reference

| Attribute | Details |
|-----------|---------|
| **Project Type** | Multi-part Web Application (Vue 3 + Go API) |
| **Primary Languages** | TypeScript (Frontend), Go (Backend) |
| **Architecture Pattern** | SPA with REST API + Multi-database |
| **Repository Type** | Monorepo with separate frontend/backend |
| **Graph Technology** | Neo4j with vis-network visualization |
| **Deployment** | Docker Compose multi-service |

## Technology Stack

### Frontend (Vue 3 SPA)
- **Framework**: Vue 3.4.0 with Composition API
- **Language**: TypeScript 5.2.2
- **Build Tool**: Vite 5.0.0
- **State Management**: Pinia 2.1.7
- **Routing**: Vue Router 4.2.5
- **UI Framework**: Tailwind CSS 3.3.6
- **Headless Components**: @headlessui/vue 1.7.19
- **Icons**: @heroicons/vue 2.0.18
- **Graph Visualization**: vis-network 9.1.6
- **HTTP Client**: Axios 1.6.0
- **Notifications**: vue-toastification 2.0.0
- **Virtual Scrolling**: vue-virtual-scroller 2.0.0

### Backend (Go API)
- **Language**: Go 1.22
- **Web Framework**: Chi v5.0.10
- **Router**: Gorilla Mux 1.8.0
- **Database Drivers**:
  - PostgreSQL: pgx/v5 5.5.0
  - Neo4j: neo4j-go-driver/v5 5.16.0
  - Redis: go-redis/v9 9.4.0
- **Authentication**: golang-jwt/jwt/v5 5.2.0
- **Password Hashing**: argon2id 1.0.0
- **CORS**: go-chi/cors 1.2.1
- **Logging**: zerolog 1.32.0
- **Configuration**: viper 1.18.2
- **Environment**: godotenv 1.5.1
- **UUID**: google/uuid 1.5.0
- **Testing**: testify 1.10.0

### Database & Storage
- **Primary Database**: PostgreSQL 15 (structured data)
- **Graph Database**: Neo4j 5 Community (relationships)
- **Cache/Session**: Redis 7 Alpine
- **File Storage**: Local filesystem

### Infrastructure & DevOps
- **Containerization**: Docker + Docker Compose
- **Service Orchestration**: Docker Compose multi-service
- **Health Checks**: Comprehensive health checks for all services
- **Networking**: Custom bridge network (172.20.0.0/16)
- **Volume Management**: Persistent data volumes

### Testing & Quality Assurance
- **Frontend Testing**: Vitest + Vue Test Utils
- **E2E Testing**: Playwright + Cypress
- **Backend Testing**: Go testing + testify
- **Database Testing**: dockertest for integration tests
- **Code Quality**: ESLint + Prettier (frontend), gofmt (backend)

## Architecture Overview

### Multi-Service Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue 3 SPA     │    │   Go API        │    │   Data Layer    │
│   (Port 3000)   │◄──►│   (Port 8080)   │◄──►│   (Multi-DB)    │
│                 │    │                 │    │                 │
│ • UI Components │    │ • REST API      │    │ • PostgreSQL    │
│ • State Mgmt    │    │ • JWT Auth      │    │ • Neo4j         │
│ • Graph Vis     │    │ • RBAC          │    │ • Redis         │
│ • Router        │    │ • Audit Log     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Features
1. **Hierarchical Taxonomy System**
   - Flexible taxonomy with infinite nesting levels
   - Similar flexibility as CI types and relationships
   - Future consideration: Cluster/grouping visualization

2. **Configuration Items (CIs) Management**
   - CRUD operations for all CI types
   - Type-safe CI definitions
   - Relationship mapping between CIs

3. **Graph Visualization**
   - Interactive network graphs using vis-network
   - Bidirectional relationship display
   - Cluster/grouping exploration for complex taxonomies

4. **Authentication & Authorization**
   - JWT-based authentication with refresh tokens
   - Role-based access control (RBAC)
   - Granular permissions system

5. **Audit & Compliance**
   - Comprehensive audit logging
   - User action tracking
   - Compliance reporting capabilities

## Repository Structure

This is a **monorepo** with clear separation of concerns:

```
pustaka/
├── web/                    # Vue 3 Frontend SPA
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── views/         # Page components
│   │   ├── stores/        # Pinia state management
│   │   ├── services/      # API client layer
│   │   └── router/        # Vue Router config
│   └── package.json
├── internal/               # Go Backend API
│   ├── api/              # HTTP handlers and routing
│   ├── auth/             # Authentication & authorization
│   ├── ci/               # CMDB business logic
│   └── database/         # Database connections
├── cmd/                   # Application entry points
├── pkg/                   # Shared libraries
├── docker/                # Docker configuration files
├── plan/                  # Implementation plans (8 files)
├── docs/                  # Documentation (11+ files)
├── tests/                 # Test suites
└── docker-compose.yml     # Multi-service orchestration
```

## Service Architecture

### Docker Compose Services
1. **postgres**: PostgreSQL 15 database
2. **neo4j**: Neo4j 5 Community with APOC plugins
3. **redis**: Redis 7 for caching and sessions
4. **api**: Go REST API service
5. **frontend**: Vue 3 SPA development server

### Service Dependencies
```
Frontend → API → [PostgreSQL, Neo4j, Redis]
```

All services include comprehensive health checks and graceful startup/shutdown handling.

## Development Workflow

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Node.js 18+
- Make (for build automation)

### Quick Start
```bash
# Start full development environment
make dev

# Or start services manually
docker-compose up -d
cd web/ && npm run dev
cd / && go run cmd/api/main.go
```

### Service Endpoints
- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **Neo4j Browser**: http://localhost:7474

## Current Status

### Completed Features ✅
- Multi-service Docker environment
- Go API with authentication and RBAC
- Vue 3 frontend with modern toolchain
- CI types and relationships management
- Basic graph visualization
- Comprehensive audit logging
- Pagination for audit logs
- Relationship types management system

### In Progress 🚧
- Flexible taxonomy system (infinite levels)
- Enhanced graph clustering visualization
- Advanced relationship mapping

### Documentation 📚
This project has extensive documentation:
- **Plans**: 8 files in `plan/` directory
- **Architecture**: Technical documentation in `docs/architecture/`
- **Planning**: PRD, GDD, epics, UX specs in `docs/planning/`
- **API**: API guides and examples
- **BMAD Workflow**: Active development planning in `bmad/`

## Next Steps for Development

1. **Taxonomy System Enhancement**
   - Implement infinite-level taxonomy hierarchy
   - Design flexible schema for dynamic taxonomies
   - Add taxonomy management UI components

2. **Graph Visualization Improvements**
   - Explore cluster/grouping mechanisms
   - Implement performance optimizations for large graphs
   - Add interactive taxonomy exploration features

3. **Integration & Testing**
   - Comprehensive E2E testing for taxonomy workflows
   - Performance testing with large CI datasets
   - Security testing for complex relationship permissions

---

**Last Updated**: 2025-10-14
**Documentation Integration**: Built upon existing comprehensive planning documents
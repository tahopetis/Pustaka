# Pustaka CMDB - Documentation Index

## Project Overview

**Type:** Monorepo with 2 parts
**Primary Language:** TypeScript (Frontend), Go (Backend)
**Architecture:** SPA with REST API + Multi-database
**Repository Root:** /home/syam/dev/pustaka

## Quick Reference

### Technology Stack Summary
- **Frontend:** Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS
- **Backend:** Go 1.22 + Chi v5 + PostgreSQL + Neo4j + Redis
- **Graph:** vis-network for interactive relationship visualization
- **Infrastructure:** Docker Compose multi-service architecture
- **Authentication:** JWT + RBAC with granular permissions

### Architecture Pattern
**Web Application (Vue 3 + Go API Separate)**
- Frontend SPA on port 3000
- Backend REST API on port 8080
- Multi-database persistence layer
- Containerized development environment

## Generated Documentation

### Core Architecture Documentation
- [Project Overview](./project-overview.md)
- [Source Tree Analysis](./source-tree-analysis.md)
- [Component Inventory - Frontend](./component-inventory-frontend.md) *(To be generated)*
- [Component Inventory - Backend](./component-inventory-backend.md) *(To be generated)*
- [Development Guide - Frontend](./development-guide-frontend.md) *(To be generated)*
- [Development Guide - Backend](./development-guide-backend.md) *(To be generated)*

### API & Data Documentation
- [API Contracts - Backend](./api-contracts-backend.md) *(To be generated)*
- [Data Models - Schema](./data-models-schema.md) *(To be generated)*
- [Integration Architecture](./integration-architecture.md) *(To be generated)*

### Existing Documentation

#### Implementation Plans (plan/ directory)
- [Setup Instructions](../plan/README_SETUP.md) - Environment setup guide
- [RBAC API Specification](../plan/RBAC_API_SPECIFICATION_PHASE1.md) - Authentication & authorization
- [Implementation Plan](../plan/IMPLEMENTATION_PLAN.md) - Development roadmap
- [Feature-Driven Design](../plan/FSD.md) - FSD methodology
- [Coding Standards](../plan/CODING_STANDARDS.md) - Development guidelines
- [Technical Specifications](../plan/TSD.md) - Technical requirements
- [Security Checklist](../plan/SECURITY_CHECKLIST.md) - Security requirements
- [Test Plan](../plan/TEST_PLAN.md) - Testing strategy

#### Architecture & Planning (docs/ directory)
- [Technical Architecture](./architecture/technical-documentation.md) - Technical architecture
- [Product Requirements](./planning/prd.md) - Product requirements document
- [Game Design Document](./planning/gdd.md) - Design specifications
- [Epics & User Stories](./planning/epics-user-stories.md) - Feature breakdown
- [UX/UI Specifications](./planning/ux-ui-specifications.md) - Design specifications
- [Implementation Roadmap](./planning/implementation-roadmap.md) - Development timeline
- [API Guide](./api-guide.md) - API usage guide
- [API Examples](./api-examples.md) - API usage examples
- [Relationship Types Implementation](./relationship_types_implementation.md) - Relationship system
- [Technical Decisions Template](./technical-decisions-template.md) - Decision tracking

#### Project Documentation
- [BMAD Workflow Status](./bmm-workflow-status.md) - Current workflow status
- [CLAUDE.md](../CLAUDE.md) - Claude development instructions

## Part-Based Navigation

### Frontend Part (web/)
**Type:** Web Application (Vue 3 SPA)
**Tech Stack:** Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS
**Root:** /home/syam/dev/pustaka/web/

**Key Components:**
- State management with Pinia stores
- Router-based navigation with Vue Router
- vis-network integration for graph visualization
- Axios HTTP client for API communication
- Tailwind CSS + Headless UI components

### Backend Part (internal/)
**Type:** Backend/API Service (Go REST API)
**Tech Stack:** Go 1.22 + Chi v5 + PostgreSQL + Neo4j + Redis
**Root:** /home/syam/dev/pustaka/internal/

**Key Components:**
- REST API handlers with Chi routing
- JWT authentication + RBAC authorization
- CMDB business logic in ci/ package
- Multi-database integration (PostgreSQL + Neo4j + Redis)
- Comprehensive audit logging system

## Current Development Focus

### 🚧 In Progress Features
1. **Flexible Taxonomy System**
   - Infinite-level taxonomy hierarchy implementation
   - Dynamic schema design for flexible taxonomies
   - Management UI components for taxonomy administration

2. **Graph Visualization Enhancements**
   - Cluster/grouping mechanisms for complex relationship visualization
   - Performance optimizations for large graph datasets
   - Interactive taxonomy exploration features

### ✅ Completed Features
- Multi-service Docker environment with health checks
- Go API with JWT authentication and RBAC system
- Vue 3 frontend with modern development toolchain
- CI types and relationships management
- Basic graph visualization using vis-network
- Comprehensive audit logging with pagination
- Relationship types management system

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local development)
- Node.js 18+ (for frontend development)
- Make (for build automation)

### Quick Development Setup
```bash
# Start full development environment
make dev

# Alternative: Manual startup
docker-compose up -d                    # Start services
cd web/ && npm run dev                  # Start frontend dev server
go run cmd/api/main.go                  # Start backend API server
```

### Service Endpoints
- **Frontend Application:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **API Health Check:** http://localhost:8080/health
- **Neo4j Browser:** http://localhost:7474
- **PostgreSQL:** localhost:5432
- **Redis:** localhost:6379

### Default Credentials
- **Admin User:** admin / Admin@123
- **PostgreSQL:** pustaka / password
- **Neo4j:** neo4j / password

## Development Commands

### Backend (Go)
```bash
make build           # Build API binary
make run             # Run API server
make test            # Run Go tests
make fmt             # Format Go code
make lint            # Run golangci-lint
make migrate         # Run database migrations
make create-admin    # Create initial admin user
```

### Frontend (Vue.js)
```bash
cd web/
npm run dev          # Start development server
npm run build        # Build for production
npm run test         # Run Vitest unit tests
npm run test:e2e     # Run Cypress E2E tests
npm run lint         # Run ESLint with auto-fix
npm run format       # Format code with Prettier
```

### Docker Services
```bash
make docker-up        # Start all Docker services
make docker-down      # Stop all Docker services
docker-compose ps     # Check service status
docker-compose logs   # View service logs
```

## Architecture Insights

### Flexible Taxonomy System Design
Based on user requirements, the taxonomy system should be:
- **Flexible**: Similar to current CI types and relationships implementation
- **Infinite Levels**: Support unlimited nesting depth
- **Dynamic**: Schema should accommodate future taxonomy needs
- **Cluster-Friendly**: Designed for graph clustering visualization

### Graph Visualization Strategy
- **Current**: vis-network for basic relationship display
- **Future**: Consider clustering/grouping mechanisms for complex taxonomies
- **Performance**: Optimize for large relationship datasets
- **Interactivity**: Enable exploration of hierarchical taxonomy structures

## Key Integration Points

### Frontend ↔ Backend Communication
```
Vue 3 SPA (Port 3000)
    ↓ HTTP/REST API (Axios)
Go API (Port 8080)
    ↓ Business Logic
CMDB Services (internal/ci/)
    ↓ Data Persistence
[PostgreSQL, Neo4j, Redis]
```

### Multi-Database Architecture
- **PostgreSQL**: Structured data (users, roles, CI types, CIs, audit logs)
- **Neo4j**: Relationship graph with bidirectional connections
- **Redis**: Session storage, caching, rate limiting

### Authentication Flow
1. JWT-based authentication with refresh mechanism
2. Access token (24h) for API requests
3. Refresh token (7d) for token renewal
4. RBAC middleware for granular permission checking

## Testing Strategy

### Frontend Testing
- **Unit Tests**: Vitest + Vue Test Utils
- **Component Testing**: Vue component isolation testing
- **E2E Testing**: Playwright + Cypress for user workflows

### Backend Testing
- **Unit Tests**: Go testing + testify for business logic
- **Integration Tests**: API endpoint testing with test database
- **Database Testing**: dockertest for database interaction testing

---

**Documentation Generated:** 2025-10-14
**Scan Mode:** Quick Scan (pattern-based with existing documentation integration)
**Focus Areas:** Flexible taxonomy system, graph clustering visualization, comprehensive architecture documentation
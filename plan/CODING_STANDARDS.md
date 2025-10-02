# 📘 Coding Standards & Conventions Document

**Project**: Pustaka CMDB
**Version**: 1.0
**Date**: 2025-10-01
**Author**: Tahopetis  

---

## 🎯 Purpose
This document defines coding standards and conventions to ensure consistency, maintainability, and quality across the entire Pustaka project. It applies to all backend, frontend, database, and infrastructure code generated for the system.

---

## 🏗️ General Principles
1. **Clarity First**: Code must be readable and self-explanatory. Favor clarity over cleverness.  
2. **Consistency**: Follow defined naming, formatting, and structure conventions across all services.  
3. **Error Handling**: All errors must be properly caught, logged, and returned in a consistent JSON format.  
4. **Security by Default**: Input validation, RBAC, and audit logging are mandatory for all endpoints.  
5. **Testability**: All code must include unit tests with meaningful coverage.  

---

## 🔹 Backend (Go API)

### Language & Framework
- Use **Go 1.22+**.  
- Use **Chi router** for HTTP routing.  
- Use **sqlx** for PostgreSQL and **neo4j-go-driver** for Neo4j.  
- Use **zerolog** for structured logging.

### Naming Conventions
- **Packages**: lowercase, no underscores (e.g., `repository`, `service`, `handler`).  
- **Functions**: `CamelCase` (exported) / `camelCase` (private).  
- **Variables**: `camelCase`.  
- **Constants**: `ALL_CAPS`.  
- **Database tables**: `snake_case` plural (e.g., `configuration_items`).  
- **JSON fields**: `snake_case`.  

### Error Handling
- All API errors must return JSON in this format:
```json
{
  "error": {
    "code": "ERR_CODE",
    "message": "Human readable message",
    "details": {}
  }
}
```
- Use standard HTTP status codes.  
- Never expose internal errors (stack traces) in API responses.  

### Logging
- Use **zerolog** with structured logs:
```go
log.Error().Str("user_id", userID.String()).Err(err).Msg("failed to create CI")
```
- All logs must include correlation IDs for tracing requests.  

### Authentication & Authorization
- All API endpoints require **JWT validation middleware**.  
- All CRUD endpoints must enforce **RBAC permissions** at the handler level.  

---

## 🔹 Frontend (Vue 3)

### Framework & Tools
- Vue 3 with **Composition API**.  
- State management via **Pinia**.  
- Styling via **TailwindCSS**.  
- Graph visualization via **D3.js** + **Vis.js**.  
- Build tool: **Vite**.  

### Naming Conventions
- **Components**: `PascalCase.vue` (e.g., `CIList.vue`, `GraphViewer.vue`).  
- **Files**: `kebab-case` (e.g., `ci-list.vue`).  
- **Store modules**: `camelCase` (e.g., `ciStore`).  
- **CSS classes**: Tailwind first, custom classes in `kebab-case`.  

### Component Structure
Each Vue component must have the following order:
```vue
<script setup lang="ts">
// imports
// props
// state (ref, reactive)
// computed
// methods
</script>

<template>
  <!-- Template content -->
</template>

<style scoped>
/* Optional scoped styles */
</style>
```

### API Calls
- All API calls go through a centralized **API service layer** (`/src/services/api.ts`).  
- API responses must be validated before being used in components.  
- Errors from API calls must display user-friendly messages.  

---

## 🔹 Database (PostgreSQL & Neo4j)

### PostgreSQL
- Use **snake_case plural table names** (`configuration_items`, `relationships`).  
- Primary keys: UUID with `gen_random_uuid()`.  
- Use **JSONB** for flexible attributes.  
- Always add **GIN indexes** for JSONB and tag arrays.  
- Foreign keys must include `ON DELETE CASCADE` where relationships require cleanup.  

### Neo4j
- Node labels: `PascalCase` (e.g., `ConfigurationItem`, `User`).  
- Relationship types: `UPPER_SNAKE_CASE` (e.g., `DEPENDS_ON`, `HOSTS`).  
- All queries must use parameters (no string concatenation).  
- Always create indexes for frequently queried fields (`id`, `type`, `name`).  

---

## 🔹 API Standards

### General
- RESTful API design.  
- Base URL: `/api/v1/`.  
- All responses must use `application/json`.  

### Pagination
- All list endpoints must support:
  - `?page=1&limit=20`  
  - Response must include:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

### Versioning
- API must be versioned (`/api/v1/`).  
- Breaking changes require incrementing version (`/api/v2/`).  

---

## 🔹 Testing

### Backend
- Unit tests for all repositories and services.  
- Integration tests for database + API endpoints.  
- Minimum coverage per phase:  
  - Phase 1: 70%  
  - Phase 2: 85%  
  - Phase 3+: 95%  

### Frontend
- Use **Vitest** for unit tests.  
- Use **Cypress** for E2E tests.  
- Each component must have:  
  - Render test  
  - Event test (e.g., button click)  
  - API integration mock test  

---

## 🔹 Infrastructure & Deployment
- Use **Docker Compose** for local dev.  
- Use **Kubernetes manifests** for production deployment.  
- Config via **environment variables**, never hardcoded secrets.  
- Health checks required for all services (`/healthz`).  
- CI/CD pipelines must run tests, linting, and security scans before deployment.  

---

## 🔹 File Size and Structure Standards
- **500 lines maximum per code file** - strictly enforced upper limit
- Keep files focused and modular with single responsibility
- Split large files into logical components
- Use clear directory structure
- If a file exceeds 500 lines, refactor into smaller, focused modules
- Exception: Architectural review required for files 500-600 lines (only for truly cohesive services)

---

## ✅ Summary
- Keep naming **consistent** across backend, frontend, DB, and APIs.
- Use **structured logging, error handling, and RBAC enforcement** in all code.
- Apply **JSONB + indexing** in Postgres and parameterized queries in Neo4j.
- Follow **REST API standards** with versioning and pagination.
- Enforce **testing coverage targets** at every phase.
- **Maintain 500-line limit per code file** for better maintainability and architecture.  

# Technology Stack

**Analysis Date:** 2026-02-20

## Languages

**Primary:**
- Go 1.22 - Backend API server and business logic
- TypeScript 5.2.2 - Frontend application development

**Secondary:**
- Vue 3.4.0 - Frontend framework
- JavaScript - Browser runtime

## Runtime

**Environment:**
- Docker containers for all services
- Development: Local machine with Docker Compose
- Production: Docker containers with health checks and restart policies

**Package Managers:**
- Go modules (go.mod)
- npm for frontend dependencies
- No lockfiles: Go uses go.sum, npm uses package-lock.json (not checked in)

## Frameworks

**Core:**
- Go Chi v5 - Web framework for REST API
- Vue 3 - Progressive JavaScript framework
- Vite 5.0.0 - Frontend build tool and dev server

**Testing:**
- Vitest 0.34.6 - Frontend unit testing
- Playwright 1.55.1 - E2E testing
- Cypress 13.6.0 - Alternative E2E testing
- Go testing package - Backend unit testing

**Build/Dev:**
- Docker - Containerization and orchestration
- Make - Development automation
- ESLint 8.54.0 - JavaScript linting
- Prettier 3.1.0 - Code formatting
- TailwindCSS 3.3.6 - Utility-first CSS framework

## Key Dependencies

**Critical:**
- Argon2ID v1.0.0 - Password hashing
- JWT v5.2.0 - JSON Web Token authentication
- Neo4j v5.16.0 - Graph database driver
- Redis v9.4.0 - Caching and session storage
- PostgreSQL v5.5.0 - Primary database driver

**Infrastructure:**
- Zerolog v1.32.0 - Structured logging
- Viper v1.18.2 - Configuration management
- UUID v1.5.0 - Unique identifier generation

## Configuration

**Environment:**
- Environment variables via .env file
- Configuration loaded with Viper
- Separate configs for development/production

**Build:**
- Vite configuration with path aliases (@ → ./src)
- Dockerfile for API service
- Dockerfile for frontend service
- Makefile for development commands

## Platform Requirements

**Development:**
- Docker Compose
- Go 1.22+
- Node.js (for frontend)
- Docker CLI

**Production:**
- Docker runtime
- Database: PostgreSQL 15
- Graph Database: Neo4j 5 (with APOC plugin)
- Cache: Redis 7
- Reverse proxy (not configured)

---

*Stack analysis: 2026-02-20*

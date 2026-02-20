# External Integrations

**Analysis Date:** 2026-02-20

## APIs & External Services

**Database APIs:**
- PostgreSQL - Primary relational database
  - Connection: DATABASE_URL environment variable
  - Client: pgx/v5 driver
- Neo4j - Graph database for relationships
  - Connection: NEO4J_URI environment variable
  - Client: Neo4j Go driver
  - APOC plugin enabled for enhanced graph operations

**Cache:**
- Redis - Session storage and caching
  - Connection: REDIS_URL environment variable
  - Client: go-redis/v9
  - Password authentication required

**Authentication:**
- Custom JWT implementation - No external auth providers
  - Access tokens: 24h TTL
  - Refresh tokens: 7d TTL
  - Secret: JWT_SECRET environment variable

## Data Storage

**Databases:**
- PostgreSQL 15 - Structured data (users, CIs, audit logs, CI types)
- Neo4j 5 - Relationship mapping with bidirectional graph connections
- Redis 7 - Session management, caching, rate limiting

**File Storage:**
- Local filesystem only - No external file storage service
- Docker volumes for persistent data

**Caching:**
- Redis for API responses and frequently accessed data
- No external CDN or distributed caching

## Authentication & Identity

**Auth Provider:**
- Custom JWT implementation - No external identity providers
  - Implementation: internal/auth/ package with JWT generation/validation
  - Password hashing: Argon2ID
  - RBAC: Custom role-based access control system

**Session Management:**
- Redis-based session storage
- Refresh token mechanism for long-lived sessions

## Monitoring & Observability

**Error Tracking:**
- None - No external error tracking service integrated
- Structured logging with Zerolog

**Logs:**
- Zerolog with JSON formatting
- No external log aggregation or monitoring
- Application logs in Docker containers

**Metrics:**
- Prometheus endpoint available (/metrics)
- No external metrics collection configured
- Built-in Go metrics exposed via /metrics

## CI/CD & Deployment

**Hosting:**
- Local Docker Compose for development
- Production deployment: Docker containers (no external hosting specified)

**CI Pipeline:**
- None detected - No external CI/CD service integration
- Manual deployment via Docker Compose

**Version Control:**
- Git (GitHub repository structure)

## Environment Configuration

**Required env vars:**
- DATABASE_URL - PostgreSQL connection
- NEO4J_URI - Neo4j connection
- REDIS_URL - Redis connection
- JWT_SECRET - JWT signing secret
- CORS_ALLOWED_ORIGINS - Frontend origins
- SERVER_PORT - API server port

**Secrets location:**
- .env file (not checked in to git)
- Docker Compose environment variables

## Webhooks & Callbacks

**Incoming:**
- None - No webhook endpoints implemented

**Outgoing:**
- None - No outbound webhooks to external services

---

*Integration audit: 2026-02-20*

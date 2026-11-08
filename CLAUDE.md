# CLAUDE.md

Practical context for coding agents working in this repository. Read
[docs/README.md](docs/README.md) for the documentation map, [GLOSSARY.md](GLOSSARY.md)
for domain vocabulary, and [docs/status.md](docs/status.md) for what is actually
implemented. Do not present Partial or Planned work as finished.

## Project summary

Porta is an open-source Configuration Management Database (CMDB) with an
Enterprise Architecture layer and an IT asset amortization module. The backend is
a Go API (Chi v5) backed by PostgreSQL, Neo4j, and Redis; the frontend is a Vue 3 +
TypeScript SPA. The single Go module is `github.com/tahopetis/porta`.

## Repository layout

```
cmd/api/main.go          API entrypoint: config, databases, migrations, admin bootstrap, router
cmd/migrations/          numbered SQL migrations (001-011), applied automatically at startup
internal/
  amortization/          financial models, straight-line calculator, ledger, scheduler
  api/                   router, middleware, CI/CI-type/relationship/audit/QR handlers
  api/handlers/          auth, user, EA, import, lifecycle, relationship-type, amortization handlers
  api/middleware/        JWT, RBAC, CORS, logging, activity tracking, audit logging
  auth/                  JWT service, Argon2id password service, RBAC service
  ci/                    CMDB core models, repositories, services, audit, lifecycle, relationship types
  config/                viper configuration with unprefixed and PORTA_-prefixed env bindings
  database/              PostgreSQL, Neo4j, Redis connections and the migration runner
  ea/                    Enterprise Architecture services, validation, CSV import
  repository/            EA data-quality queries
  testutils/             database-backed test helper (TEST_DATABASE_URL)
pkg/logger/              zerolog wrapper with database and audit helpers
web/                     Vue 3 SPA (Vite, Pinia, vue-router, Tailwind)
docker/                  Dockerfiles for the API and frontend, nginx config
docs/                    canonical documentation; start at docs/README.md
```

## Commands

### Backend

```sh
make build                 # go build -o bin/api cmd/api/main.go
make test                  # go test -v ./...
make test-integration      # DB-backed suites; requires TEST_DATABASE_URL
make fmt lint security     # go fmt, golangci-lint, gosec (tools must be installed)
```

Migrations are applied automatically when the API starts (the migrations
directory is copied into the API image). `make migrate` is a manual alternative
and needs the `migrate` CLI installed.

Database-backed tests skip themselves unless `TEST_DATABASE_URL` is set, and the
target database name must contain `test` because the helper creates tables and
deletes rows:

```sh
TEST_DATABASE_URL=postgres://porta:password@localhost:5433/porta_test?sslmode=disable make test-integration
```

Dependency bumps run through the `go-deps` GitHub Actions workflow on a hosted
runner; the job uploads a patch, which you apply on a branch and open as a pull
request. No local Go toolchain is needed for a deployment host.

### Frontend

```sh
cd web
npm run dev                # Vite dev server on port 3000
npm run build              # production bundle
npm run test               # Vitest
npm run test:e2e:playwright
```

`npm run test:e2e` is a Cypress script with no config or specs; use Playwright.
`npm run lint` and `npm run format` run ESLint and Prettier.

### Docker

```sh
make pull                  # pull published GHCR images
make dev                   # docker compose up -d, then the migrate target
make docker-up             # start services only (pulls missing images)
make docker-down           # stop services
make images-build          # rebuild images locally; slow, only when needed
```

Images are built and published by GitHub Actions (multi-arch, GHCR); a deployment
host only pulls them. `make dev` starts the stack and then runs the host-side
migrate target, and the API also applies migrations on startup, so the stack is
self-migrating.

### Services and credentials

| Service | Address | Defaults |
| --- | --- | --- |
| Frontend | http://localhost:3001 | - |
| API | http://localhost:8081 | health at `/health` |
| Neo4j Browser | http://localhost:7475 | `neo4j` / `password` |
| PostgreSQL | localhost:5433 | `porta` / `password`, database `porta` |
| Redis | localhost:6380 | password `redispassword` |

The admin user (`admin` / `Admin@123`) is created or updated at API startup.

## Configuration

- `internal/config/config.go` binds its settings twice: unprefixed
  (`DATABASE_URL`, `JWT_SECRET`, ...) and with the `PORTA_` prefix
  (`PORTA_DATABASE_URL`, ...). `.env.example` uses the unprefixed names.
- The Neo4j URL variable is `NEO4J_URL` unprefixed and `PORTA_NEO4J_URI` with the
  prefix; `NEO4J_URI` is not read.
- Defaults cover every required setting, including the admin bootstrap. Production
  validates the JWT secret and requires database and Neo4j URLs and a positive port.
- `.env.example` lists `METRICS_*`, `RATE_LIMIT_*`, `MAX_UPLOAD_SIZE`, and
  `SERVER_BASE_URL`, but the config does not bind them and no middleware enforces
  rate limiting or serves `/metrics` yet. Do not advertise them as working knobs.
- Redis is a cache only (CI types, relationship types, amortization cache).
  Authentication uses stateless JWTs, not Redis sessions.

## Architecture notes

- API base path is `/api/v1`. Public routes are `/health`, `/auth/login`,
  `/auth/refresh`, and `/public/ci/{id}`; everything else requires a JWT.
- Global middleware order: Chi `RequestID`, `RealIP`, `Recoverer`, `Timeout`,
  `AllowContentType`, `CleanPath`, then the custom logger and CORS, then
  `JWTAuth`, `ActivityTracker`, and `AuditLogging`; RBAC is applied per route
  group.
- RBAC seeds three roles (`admin`, `editor`, `viewer`) and granular permissions.
  EA permissions are added by migration 010; amortization permissions by
  migration 006.
- CIs are flat records with JSONB attributes and tags. There are no domain,
  category, or subcategory tables: grouping is by CI type, and EA entities are CIs
  whose type starts with `EA.`.
- Relationships are written to PostgreSQL and mirrored in Neo4j; the graph
  explorer reads from Neo4j.
- Audit rows are written by services (CI, CI-type, relationship, relationship-type,
  lifecycle, EA), not by the HTTP middleware, which only emits structured logs.
  Amortization events are application log lines, not `audit_logs` rows.
- Background jobs are limited: the amortization scheduler's daily entry point only
  logs, and `HandleTerminalStatusChange` has no caller. Only manual amortization
  runs touch assets today.

## Testing

Seven Go test files exist:

| File | Kind |
| --- | --- |
| `internal/amortization/calculator_simple_test.go` | unit |
| `internal/api/handlers/amortization_handlers_test.go` | handler/unit |
| `internal/api/handlers/auth_db_test.go` | database-backed |
| `internal/auth/rbac_test.go` | database-backed |
| `internal/ci/service_db_test.go` | database-backed |
| `internal/ea/import_service_test.go` | unit (CSV parsing and validation) |
| `internal/testutils/database_test.go` | unit (database-URL guard) |

`make test` runs everything; database-backed suites skip when
`TEST_DATABASE_URL` is unset. `make test-integration` runs only
`internal/auth` and `internal/ci` serially against the test database. The
frontend has Vitest suites under `web/tests` plus
`web/src/components/dashboard/DonutChart.test.ts`, and Playwright configs at
`playwright.config.ts` (testDir `tests/e2e`) and `web/playwright.config.ts`
(testDir `web/tests`).

CI (`.github/workflows/ci.yml`) runs `go build`, `go vet`, and
`go test -p 1 ./...` with a PostgreSQL/Redis service, plus frontend `vitest run`
and `npm run build`. Workflows run on pull requests; a feature-branch push
triggers nothing, though `ci.yml` also runs on pushes to `main`.

## Change workflow

- Commits follow the conventional-commit style used in the history (`feat(...)`,
  `fix(...)`, `docs(...)`, `test(...)`, `build(...)`, `ci(...)`).
- A behavior change should update the matching docs: `docs/status.md` when the
  status matrix moves, `docs/architecture/api.md` and `docs/api/openapi.yaml` when
  routes or payloads change, and the relevant `docs/features/` page.
- New migrations go in `cmd/migrations/` as a numbered `.up.sql` / `.down.sql`
  pair; the API applies them at startup.
- New HTTP routes need a handler, registration in `setupRouter` in
  `cmd/api/main.go`, an RBAC middleware on the group, and API docs.
- Frontend API calls live in `web/src/services/`; shared state lives in Pinia
  stores under `web/src/stores/`.

## Agent skills

### Issue tracker

GitHub Issues in `tahopetis/porta`, managed with the `gh` CLI. See
[docs/agents/issue-tracker.md](docs/agents/issue-tracker.md).

### Triage labels

Five canonical roles (`needs-triage`, `needs-info`, `ready-for-agent`,
`ready-for-human`, `wontfix`). See
[docs/agents/triage-labels.md](docs/agents/triage-labels.md).

### Domain docs

Single-context: `GLOSSARY.md` at the repository root and ADRs in `docs/adr/`.
See [docs/agents/domain.md](docs/agents/domain.md).

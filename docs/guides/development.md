# Development

This guide covers working on the source tree; see [installation](installation.md)
for a released deployment and [testing](testing.md) for the suites.

## Toolchain

| Tool | Version and purpose |
| --- | --- |
| Go | 1.25 (`go.mod`) for the API, migrations, and Go tests |
| Node.js and npm | 24 (as used by CI) for the Vue frontend and its tests |
| Docker Engine and Compose v2 | PostgreSQL, Neo4j, Redis, and the full stack |
| Make | Convenience targets; each wraps a command you can run directly |
| `golang-migrate`, `golangci-lint`, `gosec` | Optional CLIs used by the `migrate`, `lint`, and `security` targets |

The published containers need no local Go or Node toolchain; CI is authoritative
when the host lacks them.

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/api` | API entrypoint, router wiring, administrator bootstrap |
| `cmd/migrations` | Numbered SQL migrations (001-011) |
| `internal/api` | Handlers and middleware |
| `internal/auth` | JWT, RBAC, and password hashing |
| `internal/ci` | CMDB service, repositories, Neo4j relationships, audit |
| `internal/amortization` | Depreciation calculator, ledger service, scheduler |
| `internal/ea` | Enterprise Architecture entities, validation, CSV import |
| `internal/config` | Environment loading, defaults, validation |
| `internal/database` | PostgreSQL, Neo4j, and Redis clients; migration runner |
| `internal/repository`, `internal/testutils` | Data-quality queries and test helpers |
| `web` | Vue 3 + TypeScript frontend, Vite, Vitest, Playwright |
| `tests/e2e` | Root Playwright specs (the root config uses this directory) |
| `docker` | API and frontend Dockerfiles, nginx config |
| `scripts`, `docs` | Helper scripts; architecture, features, API, guides, ADRs, agent guidance |

## Backend

```sh
make build          # go build -o bin/api cmd/api/main.go
make run            # go run cmd/api/main.go
make fmt            # go fmt ./...
make lint           # golangci-lint run
make security       # gosec ./...
make deps           # go mod download && go mod tidy
make test           # go test -v ./...
make test-integration
```

Migrations are applied automatically when the API starts: `internal/database`
runs every pending file from `cmd/migrations` against `DATABASE_URL`, and the API
image ships that directory. The host-side `make migrate` target is only needed to
run `golang-migrate` yourself; it targets PostgreSQL port 5433 or an exported
`DATABASE_URL`.

For `make run`, copy `.env.example` to `.env` so the database, Neo4j, and Redis
URLs point at the published host ports. Without `.env`, the built-in defaults
assume the container ports (5432, 7687, 6379) and cannot reach the Compose
stack. The administrator is created or updated from `ADMIN_*` configuration at
startup, so no seed step is needed.

Database-backed Go tests skip unless `TEST_DATABASE_URL` is set. That variable
must name a database containing `test`; the helper applies migrations and
truncates tables. The CMDB suite also flushes Redis via `TEST_REDIS_URL`; see
[testing](testing.md) for exact commands.

## Frontend

```sh
cd web
npm ci
npm run dev
```

The Vite dev server listens on port 3000 and proxies `/api` to
`http://localhost:8080` (`web/vite.config.ts`), so a host-run API must use
`SERVER_PORT=8080`. The containerised frontend is served on port 3001 instead.

| Command | Purpose |
| --- | --- |
| `npm run dev` | Vite development server with the API proxy |
| `npm run build` | Production bundle; also run by CI |
| `npm run test` | Vitest in watch mode |
| `npm run test:ui` | Vitest UI |
| `npm run lint` | ESLint with auto-fix |
| `npm run format` | Prettier over `src/` |

`npm run test:e2e` opens Cypress, but there is no Cypress configuration or spec
directory, so it is not a working suite. `npm run test:e2e:playwright` uses
`web/playwright.config.ts` and starts Vite on port 3000; the root config runs
`tests/e2e` and expects a frontend on 3000. Neither is wired into CI.

## Docker workflow

| Command | Effect |
| --- | --- |
| `make pull` | Pull the published API and frontend images from GHCR |
| `make docker-up` | Start the stack, pulling only images that are missing locally |
| `make docker-down` | Stop the stack and keep its named volumes |
| `make dev` | `docker-up` plus the host-side migration target |
| `make images-build` | Build both images locally from `docker-compose.build.yml` |
| `make prod` | Start the stack with `--env-file .env` |

Images are built and published by GitHub Actions; local builds are the escape
hatch for Dockerfile or dependency changes and are slow on small hosts.

## Agent and documentation guidance

- `CLAUDE.md` at the repository root is the instruction file for coding agents;
  `docs/agents/` holds the issue-tracker convention, triage labels, and domain
  documentation rules.
- `GLOSSARY.md` defines the domain vocabulary; `docs/adr/` records architecture
  decisions.
- `docs/README.md` is the documentation hub and `docs/status.md` is the source of
  truth for what is implemented, partial, or planned; `docs/archive/README.md`
  explains the archived historical material.

Before sending a change for review, run `make fmt`, `make lint`, the relevant
tests, and `make docs-check`; see [contributing](contributing.md).

# Testing

Porta has Go unit and database-backed tests, Vitest suites for the frontend, and
two Playwright configurations. Only the Go and Vitest suites run in CI; the
end-to-end specs are local tools, and the suite set is smaller than older
phase-based plans describe.

## Go test files

| File | Type | Covers |
| --- | --- | --- |
| `internal/amortization/calculator_simple_test.go` | Unit | Straight-line monthly depreciation, salvage-value clamping, invalid configuration |
| `internal/api/handlers/amortization_handlers_test.go` | Unit, fake service | Amortization read/write routes, authentication requirement, malformed IDs and bodies, adjustments |
| `internal/api/handlers/auth_db_test.go` | Database | Real login handler against PostgreSQL: password verification and token response |
| `internal/auth/rbac_test.go` | Database | RBAC service suite and permission/role helper tests |
| `internal/ci/service_db_test.go` | Database and Redis | CI type and CI reads plus pagination from PostgreSQL; flushes test Redis |
| `internal/ea/import_service_test.go` | Unit | EA CSV parsing, row validation, error CSV generation, mock flow; its integration and template cases are permanently skipped |
| `internal/testutils/database_test.go` | Unit | Test-helper DSN guard and skip behavior |

The three database-backed files skip when `TEST_DATABASE_URL` is unset. They
require a database whose name contains `test` because the helper applies
migrations and truncates tables; `TEST_DATABASE_URL_ALLOW_ANY=1` disables the
guard and should only be used against a throwaway instance. `internal/ci` reads
`TEST_REDIS_URL`, defaulting to `redis://localhost:6379`.

## Running Go tests

```sh
make test        # go test -v ./...; database-backed suites skip without TEST_DATABASE_URL
go test ./...    # equivalent, less verbose
```

To run the complete suite locally, start the Compose services, create a test
database, and pass both URLs. This requires Go 1.25 on the host:

```sh
docker exec porta-postgres psql -U porta -d postgres -c 'CREATE DATABASE porta_test'

TEST_DATABASE_URL='postgres://porta:password@localhost:5433/porta_test?sslmode=disable' \
TEST_REDIS_URL='redis://:redispassword@localhost:6380' \
go test -p 1 ./... -count=1
```

`-p 1` is required because the packages share one PostgreSQL and Redis instance
and truncate or flush it between tests. `make test-integration` is narrower: it
requires `TEST_DATABASE_URL` and runs only `internal/auth` and `internal/ci` with
`-p 1`. The `internal/ci` suite also needs Redis through `TEST_REDIS_URL` (its
default port 6379 is not the published 6380), and the target skips the handler,
EA, and amortization suites.

## CI

The backend job in `.github/workflows/ci.yml` provisions PostgreSQL 16 and Redis
7, sets `TEST_DATABASE_URL` and `TEST_REDIS_URL`, then runs:

```sh
go build ./...
go vet ./...
go test -p 1 ./... -count=1
```

The frontend job runs on Node.js 24 from `web/`:

```sh
npx vitest run
npm run build
```

## Frontend unit tests

Vitest runs six files under jsdom:

| File | Covers |
| --- | --- |
| `web/tests/AttributeEditor.test.ts` | Dynamic attribute editor behavior |
| `web/tests/BaseSelect.test.ts` | Base select component |
| `web/tests/SearchableCISelect.test.ts` | Searchable CI picker |
| `web/tests/auth.test.ts` | Auth store and token handling |
| `web/tests/chunks.test.ts` | Production chunk grouping rules |
| `web/src/components/dashboard/DonutChart.test.ts` | Donut chart rendering and props |

`npm run test` starts Vitest in watch mode; `npx vitest run` is the one-shot CI
form. The configuration in `web/vite.config.ts` excludes `tests/e2e/**` and
`tests/**/*.spec.ts` so Playwright specs are not discovered.

## End-to-end tests

Two Playwright configurations exist and neither runs in CI. `npm run test:e2e`
opens Cypress, but there is no Cypress configuration or spec directory, so it is
not a working suite.

| Config | Test directory | Specs | Dev server |
| --- | --- | --- | --- |
| `playwright.config.ts` (root) | `tests/e2e` | 22 files declaring 72 tests | None; expects a frontend on port 3000 |
| `web/playwright.config.ts` | `web/tests` | 10 spec files (the `web/tests/e2e` suite plus `debug-audit-logs.spec.ts`), 37 tests | Starts `npm run dev` on port 3000 |

Both configurations use `http://localhost:3000` and assume the Vite proxy points
at an API on port 8080. The Compose stack publishes the frontend on 3001 and the
API on 8081, so running against Compose requires changing the base URL, or
running Vite on 3000 with an API on 8080. Many specs are debugging exercises
rather than maintained regression tests.

## Coverage gaps

- No coverage threshold is enforced; the old 70/85/95 percent targets are
  aspirational.
- Untested areas include the graph, audit, analytics, and dashboard handlers,
  middleware, most EA import execution paths, and the amortization scheduler.
- Frontend tests cover five components/utilities plus `DonutChart`; most views
  have no unit tests, and no visual or accessibility tests exist.
- End-to-end specs are not in CI and are not a reliable regression gate. There is
  also no automated migration, load, performance, chaos, penetration, or
  disaster-recovery test.

The phase-based plan that older documents describe (unit, integration, and
regression everywhere; later performance, security, chaos, and acceptance gates)
is preserved in [the archive](../archive/README.md). Its performance,
security-hardening, chaos, and disaster-recovery sections are Planned; only the
Go and Vitest suites are implemented and gated.

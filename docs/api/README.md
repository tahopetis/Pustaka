# API Documentation

This directory holds the HTTP API documentation for Porta. The single machine-readable
contract is [`openapi.yaml`](./openapi.yaml); it is **explicitly partial**.

## Contents

- [`openapi.yaml`](./openapi.yaml) - partial OpenAPI 3.0.3 contract, verified against the
  Go handlers in `cmd/api/main.go`, `internal/api/`, and `internal/api/handlers/`.
  Base server: `http://localhost:8081` (docker compose host port); routes under
  `/api/v1`. Bearer JWT security scheme.
- [`api-guide.md`](./api-guide.md) - auth flow (24h access / 168h refresh tokens), RBAC
  permission list, pagination/filtering conventions, error formats, curl quickstart.
- [`api-examples.md`](./api-examples.md) - runnable curl recipes for CI types, CIs,
  relationships, lifecycle statuses, EA entities/import/data quality, and amortization.

## Coverage (deliberately partial)

Documented: health, auth (login/refresh/me), users, CI types, CIs (CRUD and
`/ci/{id}/relationships`), relationships (CRUD plus bulk-sources/bulk-matrix),
relationship types (including active/usage/statistics/categories/validate), lifecycle
statuses (including active/usage/distribution), EA entities/CI types/teams/import/data
quality, and amortization (including settings, restructuring and reports).

Not documented in this pass - the routes exist but are intentionally out of scope and
incomplete as a contract: graph, QR, audit, analytics, dashboard, and the public CI page.
Status per operation is marked with `x-status` (`implemented` / `partial` / `planned`).

## Using the contract

Load `openapi.yaml` into Swagger UI, Redoc, or a client generator. Validate it locally
(from the repository root):

```bash
node -e "const Y=require('./web/node_modules/yaml');Y.parse(require('fs').readFileSync('docs/api/openapi.yaml','utf8'));console.log('OK')"
```

Related docs: [`docs/architecture/api.md`](../architecture/api.md) for the server architecture and [`docs/features/amortization.md`](../features/amortization.md) with the [amortization design](../features/amortization/design.md) for the financial model.

# API Reference

The Porta HTTP API is a JSON REST API served by the Go binary. This reference
lists every route registered by the live router (`setupRouter`,
`cmd/api/main.go:340-715`) with the permission enforced by the route's RBAC
middleware.

## Base URL and versioning

- Direct API: `http://<host>:8081` in Docker Compose (container port 8080);
  through the frontend: `http://<host>:3001/api/...`, where nginx proxies `/api/`
  to the API (`docker/nginx.conf`).
- Versioned routes live under `/api/v1`; `/health` and `/public/ci/{id}` are
  served at the server root. Collection routes accept an optional trailing slash.

## Authentication

Protected routes require `Authorization: Bearer <access_token>`. The token is an
HS256 JWT containing `user_id`, `username`, `email`, `roles`, and `permissions`;
issuer `porta`, audience `porta`. Default lifetimes are 24h (access) and 168h
(refresh), configurable with `JWT_ACCESS_TOKEN_TTL` /
`PORTA_JWT_ACCESS_TOKEN_TTL` and the refresh equivalents.

| Step | Request | Response fields |
|---|---|---|
| Login | `POST /api/v1/auth/login` with `{"username","password"}` | `access_token`, `refresh_token`, `token_type: "Bearer"`, `expires_in`, `user` (roles and permissions) |
| Refresh | `POST /api/v1/auth/refresh` with `{"refresh_token"}` | New `access_token`, new `refresh_token`, `token_type`, `expires_in` (rotation) |
| Current user | `GET /api/v1/me` | `{"user": {...}}` |

There is no revocation list and no server-side session: an access token stays
valid until it expires, so deactivating a user blocks new logins and refreshes
but does not invalidate an already-issued access token. A missing or malformed
`Authorization` header, or an invalid/expired token, returns **401**; a valid
token missing a required permission returns **403**. A non-empty body whose
content type is not `application/json` is rejected with **415**.

## Response and error format

Successful responses return the resource or list envelope directly as JSON (no
`{"data": ...}` wrapper). Most handler errors are `{"error": "message"}`. The
amortization handlers use a nested `error.message` / `error.code` /
`error.details` variant, and auth/middleware errors are plain-text `http.Error`
bodies, so clients should treat any non-2xx response as an error without assuming
one error shape.

Status codes: `200` read/update, `201` create, `202` accepted (manual
amortization run), `204` delete, `400` invalid input or ID, `401`
unauthenticated, `403` permission denied, `404` not found, `409` conflict,
`415` unsupported media type, `500` server or storage error.

## Pagination and filtering

List endpoints use `page` (default 1) and `limit` (default 20, capped at 100),
returning `{ <collection>, page, limit, total, total_pages }`; amortization lists
add resource-specific filters (`ci_id`, `ci_type_id`, `entry_types`,
`ci_name_search`, date ranges), and dashboard/analytics are aggregate-only.

| Endpoint | Extra filters | Envelope key |
|---|---|---|
| `GET /ci` | `ci_type`, `search`, `tags` (repeatable), `created_by`, `sort`, `order` | `cis` |
| `GET /ci-types` | `search` | `ci_types` |
| `GET /relationships` | `source_id`, `target_id`, `relationship_type`, `search` | `relationships` |
| `GET /relationship-types` | `search`, `category`, `is_active`, `is_system`, `sort`, `order` | `relationship_types` |
| `GET /lifecycle-status` | `search`, `is_active`, `is_system` | `lifecycle_statuses` |
| `GET /users` | none effective: `search`, `page`, and `limit` are parsed but ignored (all users are returned) | `users` |
| `GET /audit` | frontend variant: `entity`, `from_date`, `to_date`, `action`, `entity_id`, `performed_by`, `search`, `sort`, `order`; limit cap 200 | `audit_logs` |
| `GET /audit/logs` | `entity_type`, `start_date`, `end_date`, plus `action`, `entity_id`, `performed_by`, `search`, `sort`, `order`; limit cap 200 | `audit_logs` |
| `GET /ea/entities` | `domain`, `ci_type`, `search`, `tags`; uses `page_size` (default 25) | `entities` |

## Public routes

| Methods | Path | Auth | Notes |
|---|---|---|---|
| GET | `/health` | none | Static liveness JSON (`status`, `timestamp`); does not probe databases |
| GET | `/public/ci/{id}` | none | Whitelisted CI fields for QR scans; no audit metadata |
| POST | `/api/v1/auth/login` | none | Token pair from username/password |
| POST | `/api/v1/auth/refresh` | none | Rotates access and refresh tokens |

## Protected routes

All paths are relative to `/api/v1` and pass through JWT auth, activity tracking,
and audit logging. Permissions shown are the exact RBAC checks; group-level
RBAC is cumulative, so a route lists the group's base permission plus its own.
For example `POST /users` requires both `user:read` and `user:create`.

### Core CMDB

| Methods | Path | Permission | Notes |
|---|---|---|---|
| GET, POST | `/users` | GET `user:read`; POST `user:read` + `user:create` | 201 on create; 409 on duplicate |
| GET, PUT, DELETE | `/users/{id}` | GET `user:read`; PUT `user:read` + `user:update`; DELETE `user:read` + `user:delete` | **Partial**: PUT returns the current user and DELETE returns 204 without mutating (TODOs) |
| GET, POST | `/ci-types` | GET `ci_type:read`; POST `ci_type:read` + `ci_type:create` | 201 on create |
| GET, PUT, DELETE | `/ci-types/{id}` | GET `ci_type:read`; PUT `ci_type:read` + `ci_type:update`; DELETE `ci_type:read` + `ci_type:delete` | |
| GET, POST | `/ci` | GET `ci:read`; POST `ci:read` + `ci:create` | Attributes validated against the CI type |
| GET, PUT, DELETE | `/ci/{id}` | GET `ci:read`; PUT `ci:read` + `ci:update`; DELETE `ci:read` + `ci:delete` | Cached reads; deletes cascade relationships and ledger rows |
| GET, POST | `/relationships` | GET `relationship:read`; POST `relationship:read` + `relationship:create` | Mirrored to Neo4j after commit |
| GET, PUT, DELETE | `/relationships/{id}` | GET `relationship:read`; PUT `relationship:read` + `relationship:update`; DELETE `relationship:read` + `relationship:delete` | |
| POST | `/relationships/bulk-sources` | `relationship:read` + `relationship:create` | Many sources to one target, max 50 |
| POST | `/relationships/bulk-matrix` | `relationship:read` + `relationship:create` | Cartesian product, max 50 x 50 |
| GET, POST | `/relationship-types` | GET `relationship_type:read`; POST `relationship_type:read` + `relationship_type:create` | 201 on create |
| GET | `/relationship-types/active` | `relationship_type:read` | Cached active list |
| GET | `/relationship-types/usage` | `relationship_type:read` | Usage counts |
| GET | `/relationship-types/statistics` | `relationship_type:read` | Aggregate statistics |
| GET | `/relationship-types/categories` | `relationship_type:read` | Distinct categories |
| POST | `/relationship-types/validate` | `relationship_type:read` | Validates a proposed relationship |
| GET, PUT, DELETE | `/relationship-types/{id}` | GET `relationship_type:read`; PUT `relationship_type:read` + `relationship_type:update`; DELETE `relationship_type:read` + `relationship_type:delete` | |
| GET, POST | `/lifecycle-status` | GET `lifecycle_status:read`; POST `lifecycle_status:read` + `lifecycle_status:create` | 201 on create |
| GET | `/lifecycle-status/active` | `lifecycle_status:read` | |
| GET | `/lifecycle-status/usage` | `lifecycle_status:read` | CIs per status |
| GET | `/lifecycle-status/distribution` | `lifecycle_status:read` | Distribution for charts |
| GET, PUT, DELETE | `/lifecycle-status/{id}` | GET `lifecycle_status:read`; PUT `lifecycle_status:read` + `lifecycle_status:update`; DELETE `lifecycle_status:read` + `lifecycle_status:delete` | |

### Amortization

Group base permission: `amortization:read`. Calculation and manual runs are
implemented; the scheduled daily run is a log-only stub.

| Methods | Path | Permission | Notes |
|---|---|---|---|
| GET | `/amortization/configuration-items` | `amortization:read` | Amortizable CIs with calculated fields |
| GET, PUT | `/amortization/configuration-items/{ciId}` | GET `amortization:read`; PUT `amortization:read` + `amortization:configure` | Financial configuration update |
| GET | `/amortization/ledger` | `amortization:read` | Filterable entries |
| GET | `/amortization/ledger/{entryId}` | `amortization:read` | |
| GET | `/amortization/runs` | `amortization:read` | Run history |
| POST | `/amortization/runs` | `amortization:read` + `amortization:run` | Triggers a manual run; 202 with `run_id` |
| GET | `/amortization/runs/{runId}` | `amortization:read` | |
| GET | `/amortization/summaries` | `amortization:read` | Reporting snapshots |
| GET | `/amortization/reports/depreciation-schedule` | `amortization:read` | JSON by default; `format=csv` returns CSV |
| GET, PUT | `/amortization/settings` | GET `amortization:read`; PUT `amortization:read` + `amortization:configure` | **Broken**: queries the non-existent `amortization_settings` table |
| POST | `/amortization/adjustments` | `amortization:read` + `amortization:configure` | Manual adjustment entry |
| POST | `/amortization/restructuring/preview` | `amortization:read` + `amortization:configure` | Prospective recalculation preview |
| POST | `/amortization/restructuring` | `amortization:read` + `amortization:configure` | Applies restructuring entry |

### Platform, analytics, and EA

| Methods | Path | Permission | Notes |
|---|---|---|---|
| GET | `/graph` | `ci:read` | Graph data for vis-network |
| GET | `/graph/explore` | `ci:read` | Filtered exploration |
| GET | `/qr/ci/{id}` | `ci:read` | QR payload; URL is `<base URL>/public/ci/{id}` |
| GET | `/qr/ci/{id}/image` | `ci:read` | PNG; `size` 100-1000, default 256 |
| GET | `/ci/{id}/relationships` | `relationship:read` | **Broken**: ID resolution fallback cannot match a path ending in `relationships`, so returns 400 |
| GET | `/audit` | `audit:read` | Frontend-compatible list |
| GET | `/audit/logs` | `audit:read` | Audit log list |
| GET, DELETE | `/audit/logs/{id}` | GET `audit:read`; DELETE `audit:read` + `audit:delete` | Delete is **unreachable**: `audit:delete` is never seeded |
| GET | `/audit/stats` | `audit:read` | Aggregate statistics |
| GET | `/audit/export` | `audit:read` | CSV download (`text/csv`) |
| DELETE | `/audit/cleanup` | `audit:read` + `audit:delete` | **Unreachable** with seeded roles |
| GET | `/analytics/ci-types/usage` | `ci:read` | CI counts per type |
| GET | `/analytics/cycles` | `ci:read` | Relationship cycles (Neo4j) |
| GET | `/analytics/most-connected` | `ci:read` | Most connected CIs |
| GET | `/analytics/ci-growth` | `ci:read` | CI growth over time |
| GET | `/analytics/relationship-types/usage` | `ci:read` | Usage per relationship type |
| GET | `/me` | authenticated token only | Full user record with roles and permissions |
| GET, POST | `/ea/entities` | GET `ea:read`; POST `ea:read` + `ea:create` | Lists CIs whose type starts with `EA.`; 201 on create |
| GET, PUT, DELETE | `/ea/entities/{id}` | GET `ea:read`; PUT `ea:read` + `ea:update`; DELETE `ea:read` + `ea:delete` | |
| GET | `/ea/entities/{id}/validate` | `ea:read` | Metamodel validation result |
| GET | `/ea/entities/{id}/audit` | `ea:read` | Entity audit history |
| GET | `/ea/ci-types` | `ea:read` | EA CI types (`EA.%`) |
| GET | `/ea/teams` | `ea:read` | EA team list |
| POST | `/ea/import/template` | `ea:create` | CSV template download |
| POST | `/ea/import/validate` | `ea:create` | Validate a CSV upload |
| POST | `/ea/import/execute` | `ea:create` | Execute an import |
| POST | `/ea/import/errors/download` | `ea:create` | Error CSV download |
| GET | `/ea/import/status/{batch_id}` | `ea:create` | Import status |
| GET | `/ea/data-quality` | `ea:read` | Overall quality metrics; optional `domain` |
| GET | `/ea/data-quality/stale` | `ea:read` | Stale entities; `days_threshold` |
| GET | `/ea/data-quality/errors` | `ea:read` | Entities with validation errors |
| GET | `/ea/data-quality/lifecycle` | `ea:read` | Lifecycle breakdown |
| GET | `/dashboard/stats` | `ci:read` | CI, type, relationship counts |
| GET | `/dashboard/health-score` | `ci:read` | Computes a score and appends `health_score_history` |
| GET | `/dashboard/data-quality` | `ci:read` | Completeness/correctness summary |
| GET | `/dashboard/asset-aging` | `ci:read` | Aging buckets |
| GET | `/dashboard/risk-metrics` | `ci:read` | Risk indicators |

## Route counts

Public routes: 4 (including `/health` and `/public/ci/{id}`). Protected
`/api/v1` routes: 96. Total registrations in `setupRouter`: 100.

## Known gaps and caveats

- `GET /api/v1/ci/{id}/relationships` is registered but cannot resolve its `id`;
  `GET`/`PUT /api/v1/amortization/settings` fail because `amortization_settings`
  is never created by a migration.
- `PUT`/`DELETE /api/v1/users/{id}` are stubs that do not mutate the user, and
  the two `audit:delete` routes cannot be called by any seeded role.
- The daily amortization scheduler never fires; only manual runs execute. The API
  serves no metrics or OpenAPI route, and `docs/api/openapi.yaml` is a manually
  maintained specification, not generated from the router.
- The dead gorilla/mux router (`internal/api/router.go`) registers paths such as
  `/ci/{id}/network` and `/ci/{id}/impact` that the live server does **not**
  serve.

## Related documentation

- [Architecture overview](overview.md) and [data model](data-model.md)
- [API guide](../api/api-guide.md) and [examples](../api/api-examples.md)
- [Project status](../status.md) and [configuration guide](../guides/configuration.md)

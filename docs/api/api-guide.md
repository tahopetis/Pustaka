# Porta API Guide

Practical guide to the Porta CMDB HTTP API. The machine-readable contract is
[`openapi.yaml`](./openapi.yaml); this page covers the authentication flow, RBAC
permissions, pagination/filtering conventions and error shapes. For endpoint-by-endpoint
recipes see [`api-examples.md`](./api-examples.md), and for the server architecture
(middleware stack, service wiring) see [`docs/architecture/api.md`](../architecture/api.md).

> **The contract is explicitly partial.** It documents only the groups verified against
> the handlers in this repository. Graph, QR, audit, analytics, dashboard and the public
> CI page are intentionally out of scope for this pass. Do not treat the OpenAPI file as
> full API coverage.

## Base URL and transport

The API process listens on port **8080 inside its container**. `docker compose`
publishes it on the host as **http://localhost:8081**. All versioned routes live under
`/api/v1`; `GET /health` is the only route outside that prefix.

- `GET http://localhost:8081/health` - unauthenticated liveness probe
- `http://localhost:8081/api/v1/...` - everything else

Every request with a body must send `Content-Type: application/json`. The router applies
`AllowContentType("application/json")` globally; other media types are rejected with
`415` before the handler runs.

## Authentication

Authentication is JWT-based. The only public API endpoints are `POST /auth/login`,
`POST /auth/refresh` and `GET /public/ci/{id}` (the last one is not part of the documented
contract). Everything else needs `Authorization: Bearer <access_token>`.

1. **Log in** with `username` and `password`. The field is `username`, not `email`:

   ```bash
   curl -s -X POST http://localhost:8081/api/v1/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"admin","password":"Admin@123"}'
   ```

   The response contains `access_token`, `refresh_token`, `token_type: "Bearer"`,
   `expires_in`, and the full `user` object (roles and permissions).

2. **Use** the access token on subsequent calls:

   ```bash
   curl -s http://localhost:8081/api/v1/me -H "Authorization: Bearer $TOKEN"
   ```

3. **Refresh** before or after expiry. The refresh endpoint validates the refresh token
   and returns a **new access token and a new refresh token** (both rotate):

   ```bash
   curl -s -X POST http://localhost:8081/api/v1/auth/refresh \
     -H 'Content-Type: application/json' \
     -d "{\"refresh_token\":\"$REFRESH\"}"
   ```

Default lifetimes come from configuration: access token `24h`
(`JWT_ACCESS_TOKEN_TTL`), refresh token `168h` (`JWT_REFRESH_TOKEN_TTL`). There is no
server-side token revocation endpoint; discard tokens client-side.

## Authorization (RBAC)

Every protected route also requires one or more permission strings, checked by middleware
against the permissions embedded in the JWT. Permissions are seeded by migrations and
grouped by resource:

| Resource | Permissions |
| --- | --- |
| CI | `ci:read`, `ci:create`, `ci:update`, `ci:delete` |
| CI types | `ci_type:read`, `ci_type:create`, `ci_type:update`, `ci_type:delete` |
| Relationships | `relationship:read`, `relationship:create`, `relationship:update`, `relationship:delete` |
| Relationship types | `relationship_type:read`, `relationship_type:create`, `relationship_type:update`, `relationship_type:delete` |
| Lifecycle statuses | `lifecycle_status:read`, `lifecycle_status:create`, `lifecycle_status:update`, `lifecycle_status:delete` |
| Users | `user:read`, `user:create`, `user:update`, `user:delete` |
| Audit | `audit:read` (deletes also need `audit:delete`) |
| Amortization | `amortization:read`, `amortization:configure`, `amortization:run` (migrations also seed `amortization:adjust`/`amortization:admin`, but no route checks them) |
| Enterprise architecture | `ea:read`, `ea:create`, `ea:update`, `ea:delete` |
| System | `system:admin` |

The seeded roles are `admin` (all permissions), `editor` (read on CI, CI types,
relationships, relationship types, lifecycle statuses, EA and audit; CI and relationship
create/update/delete; `ea:create`/`ea:update`; `amortization:read/configure`) and
`viewer` (read-only). Read routes also require their read permission, so a mutation needs
both the read and the write permission - for example `POST /ci` requires `ci:read`
**and** `ci:create`. A missing permission produces `403` with a plain-text body.

## Pagination, filtering and sorting

Most list endpoints accept `page` (default 1) and `limit` (default 20, maximum 100);
out-of-range values are clamped by the handler. Pagination response keys vary by group:

- CMDB lists (`/ci`, `/ci-types`, `/relationships`, `/relationship-types`,
  `/lifecycle-status`): `page`, `limit`, `total`, `total_pages`
- EA entities: `page`, `page_size`, `total`, `total_pages`
- Amortization: `page`, `page_size`, `total_count`, `total_pages`

Common filters: `search` on list endpoints, `ci_type` and repeatable `tags` on `/ci`,
`source_id`/`target_id`/`relationship_type` on `/relationships`, `category`/`is_active`/
`is_system` on `/relationship-types`, `domain` on EA queries. `/ci` supports
`sort=name|type|created_at|updated_at` plus `order=asc|desc`; other list endpoints sort by
`sort`/`order` with group-specific fields. When in doubt, the operation parameters in
`openapi.yaml` are authoritative.

## Errors

Check the HTTP status first; error bodies are not uniform:

- Most handlers return `{"error": "<message>"}`.
- Authentication handlers and the JWT/RBAC middleware return `text/plain`.
- Amortization returns a nested object:
  `{"error": {"message": "...", "code": "...", "details": "..."}}`.

Common statuses: `400` invalid input, `401` missing/invalid token, `403` missing
permission, `404` unknown resource, `409` conflict (duplicate CI-type name, CI still
referenced), `422` critical EA validation failure, `500` server-side failure. Some
handlers map "not found" conditions to `500`; where that is the case the operation
description in `openapi.yaml` says so.

## Implementation status

The contract marks each operation with `x-status`:

- **Implemented** - route and handler match the documented behaviour.
- **Partial** - route works but has a stated limitation (for example, a handler that
  resolves path parameters with the wrong router library, or an import route blocked by
  the global JSON content-type middleware).
- **Planned** - the documented behaviour is not wired up yet (user update/delete, async
  import status).

Not implemented or not present: rate limiting, the Prometheus `/metrics` endpoint,
Redis-backed sessions, `/ci/search`, and one-off audit resources such as `/audit/{id}`.
Redis is used for caching only. The documented list of paths is a subset of the router;
do not infer unwritten endpoints from it.

## Related documentation

- [`openapi.yaml`](./openapi.yaml) - partial OpenAPI 3.0.3 contract
- [`api-examples.md`](./api-examples.md) - runnable curl recipes
- [`docs/architecture/api.md`](../architecture/api.md) - server structure and middleware
- [`docs/features/amortization.md`](../features/amortization.md) - amortization module overview, with the [design](../features/amortization/design.md) and [data validation](../features/amortization/data-validation.md) detail docs next to it

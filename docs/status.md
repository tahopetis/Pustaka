# Feature Status

What actually works in Porta today, checked against the code rather than against
the planning documents. Use this page as the source of truth when a guide, PRD,
or README disagrees with it.

## Status vocabulary

| Status | Meaning |
| --- | --- |
| **Implemented** | Works end-to-end through the API and UI (or through the API where no UI exists). No stub in the main path. |
| **Partial** | Shipped in some form but stubbed, limited to one surface, or not wired into the main path. Described honestly below. |
| **Planned** | No implementation. At most a configuration key, interface, or placeholder exists. |

## Matrix

| Area | Capability | Status | Notes |
| --- | --- | --- | --- |
| Access | JWT login and refresh | Implemented | `POST /api/v1/auth/login` and `/auth/refresh`; Argon2id password hashing; access token 24h, refresh 168h. |
| Access | RBAC with per-route permissions | Implemented | Roles `admin`, `editor`, `viewer`; protected route groups require granular permissions such as `ci:read`, `ea:create`, `amortization:run`. |
| Access | User management | Partial | Create, list, and read under `/api/v1/users` work; update and delete are stubs (PUT returns the current user, DELETE returns 204 without deleting), and list search/pagination are not implemented. |
| CMDB | CI types with attribute schemas | Implemented | `ci_type_definitions` stores required/optional attribute definitions as JSONB, including `is_amortizable`. |
| CMDB | Configuration items | Implemented | `configuration_items` with JSONB attributes, tags, and an optional lifecycle status. |
| CMDB | Lifecycle status catalogue | Implemented | `lifecycle_statuses`, including `amortization_behavior` (`pending`/`active`/`terminal`). |
| CMDB | Relationship types | Implemented | Forward/reverse labels, category, allowed source/target CI types, cardinality, bidirectional flag, usage statistics. |
| CMDB | Relationships and bulk creation | Implemented | Single create plus `bulk-sources` and `bulk-matrix`; mirrored in Neo4j for graph queries. |
| CMDB | Graph explorer | Implemented | `/api/v1/graph` and `/graph/explore` back the vis-network UI. |
| CMDB | Search, filtering, pagination | Implemented | `search`, filter, page, and limit query parameters on the CI, CI type, relationship, EA, and audit list endpoints. |
| CMDB | QR codes and public CI read endpoint | Implemented | `/qr/ci/{id}` and `/qr/ci/{id}/image` generate codes; `GET /public/ci/{id}` returns a read-only JSON view without authentication. |
| EA | Eight-domain metamodel and entity CRUD | Implemented | Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance; entities are CIs with `EA.<domain>-<type>` CI types. |
| EA | EA teams and CI-type lookups | Implemented | `ea_teams`; `/api/v1/ea/teams` and `/api/v1/ea/ci-types`. |
| EA | Entity validation and data-quality score | Implemented | Required-attribute and cross-field validation blocks the save with HTTP 422; optional-attribute problems are recorded as warnings and do not block. Each entity stores a 0–100 `data_quality_score`, and `GET /ea/entities/{id}/validate` reports attribute and cross-field errors. |
| EA | CSV import for EA entities | Partial | Synchronous import only, and `ci_type` must start with `EA.`; `/api/v1/ea/import/status/{batch_id}` is a placeholder, and there is no general CMDB/CI import path. |
| EA | Data-quality dashboard and APIs | Implemented | `/api/v1/ea/data-quality`, `/stale`, `/errors`, `/lifecycle` back the EA data-quality dashboard. |
| Amortization | Per-CI financial configuration | Implemented | Purchase cost, salvage value, amortization start date, useful life, and current book value columns; accumulated depreciation and monthly depreciation are derived. |
| Amortization | Straight-line depreciation calculator | Implemented | Monthly, catch-up, write-off, projection, and restructuring calculations. The model has a `depreciation_method` field, but only straight-line is implemented. |
| Amortization | Append-only ledger | Implemented | `amortization_ledger` rows are appended, never updated. The schema (migrations 006 and 007) allows seven `entry_type` values: `depreciation`, `monthly_depreciation`, `catch_up_depreciation`, `write_off`, `adjustment`, `correction`, `restructuring`. The service currently writes five of them: `monthly_depreciation`, `catch_up_depreciation`, `write_off`, `adjustment`, and `restructuring`. |
| Amortization | Manual runs | Implemented | `POST /api/v1/amortization/runs` supports dry-run, date override, and a subset of CIs. |
| Amortization | Reports and summaries | Implemented | Depreciation schedule and summaries expose OCC, GVB, NBV, AD, and salvage totals. |
| Amortization | Adjustment entries | Partial | `POST /api/v1/amortization/adjustments` appends a ledger entry and updates the CI book value, but neither the handler nor the service validates the reason or rejects negative/oversized amounts. |
| Amortization | Restructuring | Implemented | `POST /api/v1/amortization/restructuring/preview` and `/restructuring` preview and apply a prospective useful-life change. |
| Amortization | Global amortization settings | Partial | API handlers, service, and settings UI exist, but the repository queries an `amortization_settings` table that no migration creates, so the settings endpoints do not work end-to-end. |
| Amortization | Scheduled daily automation | Partial | The scheduler's daily entry point only logs; no timer or cron calls the run executor. Scheduled runs must be triggered manually today. |
| Amortization | Automatic write-off on terminal status | Partial | `HandleTerminalStatusChange` exists but has no caller from the CI service; a terminal-status asset is written off only when a run processes it. |
| Analytics | Dashboard widgets | Implemented | `/api/v1/dashboard/stats`, `/health-score`, `/data-quality`, `/asset-aging`, `/risk-metrics`; health-score snapshots provide the trend. |
| Analytics | Analytics endpoints | Implemented | `/api/v1/analytics/cycles`, `/most-connected`, `/ci-growth`, and CI-type/relationship-type usage. |
| Audit | Audit logging | Implemented | Service-layer writes to `audit_logs` for CI, CI-type, relationship, relationship-type, lifecycle, and EA changes. Amortization and request-level events go to the application log only, and the HTTP middleware does not write rows itself. |
| Audit | Audit viewer and export | Implemented | List/filter, single entry, statistics, and CSV export (capped at 10,000 rows). |
| Audit | Audit retention cleanup and deletion | Partial | The delete and cleanup routes require `audit:delete`, a permission no seeded role is granted, so they are unreachable; the default retention is one year. |
| Operations | Docker Compose deployment | Implemented | Five services (`postgres`, `neo4j`, `redis`, `api`, `frontend`); images built by GitHub Actions and published to GHCR. |
| Operations | Database migrations | Implemented | 11 numbered migrations (001–011) in `cmd/migrations`; the API applies pending migrations at startup and ships them in its image. |
| Operations | Health endpoint | Implemented | `GET /health`. |
| Operations | In-app notifications | Partial | Client-side toast store only. There is no server-side notification model, email, webhook, or push delivery. |
| Operations | Rate limiting | Planned | Rate-limit defaults exist in configuration, but no middleware enforces them and the `RATE_LIMIT_*` environment variables are not read. |
| Operations | Prometheus metrics | Planned | A metrics flag exists, but the `/metrics` handler is still a TODO in `cmd/api/main.go`. |
| Operations | Redis-backed sessions | Planned | Authentication is stateless JWT. Redis is used only as a cache. |

## Notes

- There are no domain, category, or subcategory tables. Pustaka-era documents describe a four-level taxonomy, but grouping today is by CI type (and by EA domain for Enterprise Architecture content).
- Feature docs under `docs/features/` describe how each implemented area behaves; this page is the quick reference for what is safe to rely on.

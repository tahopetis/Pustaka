# Porta API Examples

Runnable `curl` recipes for the Porta CMDB API. The base URL is
**http://localhost:8081/api/v1** (the docker compose host port; the API listens on 8080
inside the container). Every example assumes a shell with `curl` (and optionally `jq`).

> The documented contract is explicitly partial. Graph, QR, audit, analytics, dashboard
> and the public CI page are out of scope here; see [`openapi.yaml`](./openapi.yaml) for
> the verified paths and [`api-guide.md`](./api-guide.md) for auth and conventions.

## 0. Set up a token

```bash
BASE=http://localhost:8081/api/v1

TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' | jq -r '.access_token')

REFRESH=$(curl -s -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' | jq -r '.refresh_token')

echo "token: ${TOKEN:0:20}..."
```

`docker compose` seeds the admin user from `PORTA_ADMIN_USERNAME` / `PORTA_ADMIN_PASSWORD`
(`admin` / `Admin@123` in the provided compose file). Login uses `username`, not email.

### Refresh the token pair

```bash
curl -s -X POST "$BASE/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```

The response rotates **both** tokens: `access_token`, `refresh_token`, `token_type`,
`expires_in`. No user object is returned.

### Current user

```bash
curl -s "$BASE/me" -H "Authorization: Bearer $TOKEN"
```

Returns `{"user": {...}}` with `roles` and `permissions`.

## 1. Users

Requires `user:read` / `user:create` (admin by default).

```bash
# Create
curl -s -X POST "$BASE/users" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"username":"jdoe","email":"jdoe@example.com","password":"ChangeMe123","roles":["viewer"]}'

# List (search is accepted but currently ignored)
curl -s "$BASE/users?page=1&limit=20&search=jdoe" -H "Authorization: Bearer $TOKEN"
```

`PUT /users/{id}` and `DELETE /users/{id}` are registered but **not implemented**: the
first returns the unchanged user, the second returns `204` without deleting.

## 2. CI types

A CI type defines the attribute schema that every CI of that type must satisfy. The
migrations seed `Server`, `Application`, `Database`, and the EA types; you can add your
own. `is_amortizable` opts a type into the amortization module.

```bash
# Create an amortizable Laptop type
curl -s -X POST "$BASE/ci-types" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "name": "Laptop",
    "description": "Employee laptop",
    "is_amortizable": true,
    "required_attributes": [
      {"name": "serial_number", "type": "string", "description": "Manufacturer serial number",
       "validation": {"min_length": 3, "max_length": 64}}
    ],
    "optional_attributes": [
      {"name": "model", "type": "string", "description": "Hardware model"},
      {"name": "purchase_year", "type": "integer", "description": "Year purchased",
       "validation": {"min": 2000, "max": 2100}}
    ]
  }'

# List / search
curl -s "$BASE/ci-types?search=Laptop&page=1&limit=20" -H "Authorization: Bearer $TOKEN"

# Get one
CT_ID=$(curl -s "$BASE/ci-types?search=Laptop" -H "Authorization: Bearer $TOKEN" | jq -r '.ci_types[0].id')
curl -s "$BASE/ci-types/$CT_ID" -H "Authorization: Bearer $TOKEN"
```

`PUT /ci-types/{id}` cannot rename the type; it updates the description, the
`is_amortizable` flag and the attribute lists. `DELETE` returns `409` while CIs of the
type exist.

## 3. Configuration items

For the seeded `Server` type the required attributes are `hostname`, `ip_address`, and
`os`; unknown attributes are rejected by schema validation.

```bash
# Create a Server CI
CI_ID=$(curl -s -X POST "$BASE/ci" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "name": "web-server-01",
    "ci_type": "Server",
    "attributes": {
      "hostname": "web01",
      "ip_address": "192.168.1.10",
      "os": "Ubuntu",
      "cpu_cores": 8,
      "memory_gb": 32
    },
    "tags": ["production", "web"]
  }' | jq -r '.id')

# Create a Laptop CI (uses the type created above)
LAPTOP_ID=$(curl -s -X POST "$BASE/ci" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "name": "lt-0042",
    "ci_type": "Laptop",
    "attributes": {"serial_number": "SN-44213", "model": "T14 Gen 4", "purchase_year": 2025},
    "tags": ["laptop"]
  }' | jq -r '.id')

# Get
curl -s "$BASE/ci/$CI_ID" -H "Authorization: Bearer $TOKEN"

# List with filters: ci_type, search, repeatable tags, sort/order, page/limit
curl -s "$BASE/ci?ci_type=Server&search=web&tags=production&sort=name&order=asc&page=1&limit=50" \
  -H "Authorization: Bearer $TOKEN"

# Update attributes/tags/lifecycle_status_id (name and ci_type are immutable)
curl -s -X PUT "$BASE/ci/$CI_ID" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"attributes":{"hostname":"web01","ip_address":"192.168.1.11","os":"Ubuntu","cpu_cores":16,"memory_gb":64},"tags":["production","web","upgraded"]}'

# Delete (409 while relationships exist)
curl -s -X DELETE "$BASE/ci/$CI_ID" -H "Authorization: Bearer $TOKEN"
```

## 4. Relationships

```bash
# Create a dependency
curl -s -X POST "$BASE/relationships" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"source_id\":\"$LAPTOP_ID\",\"target_id\":\"$CI_ID\",\"relationship_type\":\"depends_on\"}"

# List with filters
curl -s "$BASE/relationships?source_id=$LAPTOP_ID&page=1&limit=20" -H "Authorization: Bearer $TOKEN"

# Bulk: many sources -> one target (max 50 sources, no duplicates)
curl -s -X POST "$BASE/relationships/bulk-sources" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"source_ids\":[\"$LAPTOP_ID\"],\"target_id\":\"$CI_ID\",\"relationship_type\":\"depends_on\"}"

# Bulk: cartesian product many sources x many targets (max 50 each)
curl -s -X POST "$BASE/relationships/bulk-matrix" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"source_ids\":[\"$LAPTOP_ID\"],\"target_ids\":[\"$CI_ID\"],\"relationship_type\":\"depends_on\"}"

# Update only attributes
curl -s -X PUT "$BASE/relationships/$REL_ID" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"attributes":{"criticality":"high"}}'
```

`GET /ci/{id}/relationships` is registered but currently returns `400 Invalid CI ID`
(the handler reads path parameters with the wrong router library). Use
`GET /relationships?source_id=...` or `?target_id=...` instead.

## 5. Relationship types

```bash
curl -s "$BASE/relationship-types?page=1&limit=50" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/relationship-types/active" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/relationship-types/categories" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/relationship-types/usage" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/relationship-types/statistics" -H "Authorization: Bearer $TOKEN"

# Create a custom type
curl -s -X POST "$BASE/relationship-types" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "name": "hosts",
    "display_name": "Hosts",
    "description": "Source hosts the target",
    "forward_label": "hosts",
    "reverse_label": "hosted on",
    "category": "infrastructure",
    "cardinality_source": "one",
    "cardinality_target": "many",
    "bidirectional": false
  }'

# Validate a combination before creating an edge
curl -s -X POST "$BASE/relationship-types/validate" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"relationship_type":"hosts","source_type":"Server","target_type":"Application"}'
```

Note: create/update validation failures are reported as `500`, and the boolean list
filters `is_active`/`is_system` only apply when set to `true`.

## 6. Lifecycle statuses

The migrations seed ten statuses (`planned`, `on_order`, `in_stock`, `pending_install`,
`operational`, `in_maintenance`, `defective_repair`, `retired`, `disposed`,
`missing_stolen`). `amortization_behavior` on a status is `pending`, `active`, or
`terminal` and drives the amortization module.

```bash
curl -s "$BASE/lifecycle-status?page=1&limit=20&is_active=true" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/lifecycle-status/active" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/lifecycle-status/usage" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/lifecycle-status/distribution" -H "Authorization: Bearer $TOKEN"

# Create a custom status (system statuses cannot be deleted)
curl -s -X POST "$BASE/lifecycle-status" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"in_use","display_name":"In Use","color":"#22c55e","sort_order":55,"amortization_behavior":"active"}'

# Attach a status to a CI
curl -s -X PUT "$BASE/ci/$CI_ID" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"lifecycle_status_id\":\"$STATUS_ID\"}"
```

## 7. Enterprise architecture

EA entities are CI records of the seeded `EA.<Domain>-<Type>` CI types, owned by an EA
team. List the catalogue first:

```bash
curl -s "$BASE/ea/teams" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/ci-types" -H "Authorization: Bearer $TOKEN"
```

```bash
# Create an EA entity (owner must be a seeded team; lifecycle_status_id is required by the model)
EA_ID=$(curl -s -X POST "$BASE/ea/entities" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "name": "customer-portal",
    "ci_type": "EA.Application-BusinessApplication",
    "owner": "enterprise-architecture",
    "lifecycle_status_id": "'"$STATUS_ID"'",
    "attributes": {"criticality": "high", "version": "2.3.1"},
    "tags": ["customer-facing"]
  }' | jq -r '.id')

# List/filter
curl -s "$BASE/ea/entities?domain=Application&search=portal&page=1&page_size=25" \
  -H "Authorization: Bearer $TOKEN"

# Get / validate / audit
curl -s "$BASE/ea/entities/$EA_ID" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/entities/$EA_ID/validate" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/entities/$EA_ID/audit?page=1&page_size=50" -H "Authorization: Bearer $TOKEN"

# Delete (400 + relationship_count when relationships exist; force bypasses the check)
curl -s -X DELETE "$BASE/ea/entities/$EA_ID?force=true" -H "Authorization: Bearer $TOKEN"
```

### EA data quality

```bash
curl -s "$BASE/ea/data-quality" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/data-quality?domain=Application" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/data-quality/stale?days_threshold=90&include_incomplete=true" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/data-quality/errors?domain=Application" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/ea/data-quality/lifecycle?domain=Application" -H "Authorization: Bearer $TOKEN"
```

### EA CSV import

Only the template download works today (JSON request, CSV response):

```bash
curl -s -X POST "$BASE/ea/import/template" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"ci_type":"EA.Application-BusinessApplication"}' \
  -o ea-application-template.csv
```

`POST /ea/import/validate` and `POST /ea/import/execute` are designed for
`multipart/form-data` (`file` + `ci_type`), but the global JSON content-type middleware
rejects them with `415` before the handler runs. Do not rely on them yet.

## 8. Amortization

Every amortization route requires `amortization:read`; changes to configuration,
adjustments, restructuring and settings additionally require
`amortization:configure`; triggering runs requires `amortization:run`.

```bash
CI=$LAPTOP_ID

# Configure the financials of an amortizable CI
curl -s -X PUT "$BASE/amortization/configuration-items/$CI" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"purchase_cost":2400.00,"salvage_value":200.00,"amort_start_date":"2025-01-01T00:00:00Z","useful_life_months":36}'

# Details (list, filter, details, ledger)
curl -s "$BASE/amortization/configuration-items?search=lt-0042&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/amortization/configuration-items/$CI" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/amortization/ledger?ci_id=$CI&page=1&limit=20" -H "Authorization: Bearer $TOKEN"

# Manual adjustment
curl -s -X POST "$BASE/amortization/adjustments" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"ci_id\":\"$CI\",\"amount\":50.00,\"description\":\"Correction after inventory audit\"}"

# Manual run (202; then poll the run)
RUN_ID=$(curl -s -X POST "$BASE/amortization/runs" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"dry_run":false}' | jq -r '.run_id')
curl -s "$BASE/amortization/runs/$RUN_ID" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/amortization/runs?status=completed,failed&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"

# Reports: summary, depreciation schedule (JSON or CSV)
curl -s "$BASE/amortization/summaries?group_by=ci_type" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/amortization/reports/depreciation-schedule?date_from=2025-01-01&date_to=2025-12-31&format=csv" \
  -H "Authorization: Bearer $TOKEN" -o depreciation-2025.csv

# Settings: the handlers exist but query an amortization_settings table that no
# migration creates, so both calls currently return 500 on a migrated database.
curl -s "$BASE/amortization/settings" -H "Authorization: Bearer $TOKEN"
curl -s -X PUT "$BASE/amortization/settings" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"currency":"EUR","default_useful_life_months":48}'

# Preview then execute a useful-life restructuring
curl -s -X POST "$BASE/amortization/restructuring/preview" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"ci_id\":\"$CI\",\"new_useful_life_months\":48}"
curl -s -X POST "$BASE/amortization/restructuring" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"ci_id\":\"$CI\",\"new_useful_life_months\":48,\"reason\":\"Extended support contract\"}"
```

Amortization failures (including unknown IDs) are returned as `500` with the nested body
`{"error": {"message": "...", "code": "...", "details": "..."}}`; only malformed UUIDs
and query parameters produce `400`.

## 9. Error handling

```bash
# 401 - missing/expired token (plain text from middleware)
curl -s -i "$BASE/ci" | head -5

# 403 - authenticated but missing permission (plain text)
curl -s -i "$BASE/users" -H "Authorization: Bearer $VIEWER_TOKEN" | head -5

# 400 - schema validation failure (JSON)
curl -s -i -X POST "$BASE/ci" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"broken","ci_type":"Server","attributes":{}}' | tail -3

# 409 - deleting a CI type that still has CIs
curl -s -i -X DELETE "$BASE/ci-types/$CT_ID" -H "Authorization: Bearer $TOKEN" | tail -3
```

Recommended client behaviour: branch on HTTP status first; parse
`{"error": ...}` when the response is JSON, otherwise show the text body. There is no
rate limiting, `/metrics`, or Redis session store in this codebase, so no
retry-after/rate-limit headers or server-side session invalidation should be assumed.

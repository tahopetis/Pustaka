# Upgrading from Pustaka to Porta

Porta is the same application previously shipped as **Pustaka**. The rename is a
**breaking change**: there is no compatibility shim, by [design](../adr/0001-rename-pustaka-to-porta.md).

If you are starting fresh, ignore this guide — `docker compose down -v` and pull the
new images.

## 1. Environment variables

Every `PUSTAKA_*` variable became `PORTA_*`. The old prefix is **not** read anymore, and
the application falls back to defaults (localhost database, development JWT secret)
if it is missing — so update these before starting, or you will silently run on
defaults.

| Old | New |
| --- | --- |
| `PUSTAKA_DATABASE_URL` | `PORTA_DATABASE_URL` |
| `PUSTAKA_NEO4J_URI` | `PORTA_NEO4J_URI` |
| `PUSTAKA_NEO4J_USERNAME` | `PORTA_NEO4J_USERNAME` |
| `PUSTAKA_NEO4J_PASSWORD` | `PORTA_NEO4J_PASSWORD` |
| `PUSTAKA_REDIS_URL` | `PORTA_REDIS_URL` |
| `PUSTAKA_JWT_SECRET` | `PORTA_JWT_SECRET` |
| `PUSTAKA_SERVER_PORT` | `PORTA_SERVER_PORT` |
| `PUSTAKA_ADMIN_EMAIL` | `PORTA_ADMIN_EMAIL` |

The same applies to every other key: `DATABASE_*`, `REDIS_*`, `JWT_*`, `SERVER_*`,
`CORS_*`, `LOGGING_*`, `ADMIN_*`, `ENVIRONMENT`.

Unprefixed names (`DATABASE_URL`, `JWT_SECRET`, …) are still accepted, and
`.env.example` uses them.

## 2. Database and volumes

The default role and database are now `porta` (was `pustaka`), and because the compose
project name is now explicit the old volumes are **not** reused.

**Fresh development environment** — recreate everything:

```sh
docker compose down -v
docker compose pull
docker compose up -d
```

**Data you must keep** — do this while the *old* stack is still running, then copy the
volumes across before starting the new stack:

```sh
# 1. Rename the database and role. The old stack's POSTGRES_USER is 'pustaka'
#    (a superuser); there is no 'postgres' role, so connect as the app user.
docker exec -it pustaka-ci-fix-postgres \
  psql -U pustaka -d postgres -c 'ALTER DATABASE pustaka RENAME TO porta;'
docker exec -it pustaka-ci-fix-postgres \
  psql -U pustaka -d postgres -c 'ALTER ROLE pustaka RENAME TO porta;'

# 2. Stop the old stack; its volumes (pustaka_*) survive.
docker compose down

# 3. Copy the data volumes to the new project prefix.
for v in postgres_data neo4j_data neo4j_logs neo4j_import redis_data; do
  docker volume create "porta_$v"
  docker run --rm -v "pustaka_$v:/from:ro" -v "porta_$v:/to" \
    alpine sh -c 'cp -a /from/. /to/'
done

# 4. Start the new stack.
docker compose pull
docker compose up -d
```

The administrator email also moved. Migrations insert `admin@porta.local`, and on Docker
installs `PORTA_ADMIN_EMAIL` (compose sets `admin@porta.dev`) is applied at startup and
wins — so the effective address is `admin@porta.dev` unless you override it. Existing
rows keep the address they already have.

## 3. Containers, network, volumes

| Before | After |
| --- | --- |
| containers `pustaka-ci-fix-*` | `porta-*` (`porta-api`, `porta-postgres`, `porta-neo4j`, `porta-redis`, `porta-frontend`) |
| network `pustaka_pustaka-network` | `porta_porta-network` (project-prefixed; the network has no explicit `name:`) |
| compose project derived from directory | explicit `name: porta` |
| volumes `pustaka_*` | volumes `porta_*` (`porta_postgres_data`, `porta_neo4j_data`, …) |
| images built on the host | `ghcr.io/tahopetis/porta-{api,frontend}` pulled from GHCR |

## 4. Authentication

The JWT issuer and audiences changed (`pustaka` → `porta`, `pustaka-refresh` →
`porta-refresh`). All existing access and refresh tokens are invalid: users must log
in again.

## 5. Go module path

`github.com/pustaka/pustaka` → `github.com/tahopetis/porta`. Any out-of-tree importer
must update its imports and `go.mod`.

## 6. GitHub

The repository is now `tahopetis/porta`. Old URLs (including
`tahopetis/Pustaka`) redirect, and open issues/PRs are preserved.

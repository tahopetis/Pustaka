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

## 2. Database

The default role and database are now `porta` (was `pustaka`).

**Fresh development environment** — the volumes are recreated:

```sh
docker compose down -v
docker compose pull
docker compose up -d
```

**Data you must keep** — either dump/restore, or rename in place:

```sh
# PostgreSQL, inside the postgres container
psql -U postgres -c 'ALTER DATABASE pustaka RENAME TO porta;'
psql -U postgres -c 'ALTER ROLE pustaka RENAME TO porta;'
```

The seeded administrator email also changed: `admin@pustaka.local` → `admin@porta.local`
(fresh installs only; existing rows are untouched).

## 3. Containers, network, volumes

| Before | After |
| --- | --- |
| `pustaka-ci-fix-*` containers | `porta-*` (`porta-api`, `porta-postgres`, …) |
| `pustaka-network` | `porta-network` |
| compose project derived from directory | explicit `name: porta` |
| images built on the host | `ghcr.io/tahopetis/porta-{api,frontend}` pulled from GHCR |

Because the compose project name is now explicit, volumes are named
`porta_postgres_data`, `porta_neo4j_data`, etc. Old `pustaka_*` volumes are **not**
reused; export your data before switching if it matters.

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

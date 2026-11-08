# Installation

Porta runs as five Docker Compose services: `api`, `frontend`, `postgres`,
`neo4j`, and `redis`. GitHub Actions publishes multi-architecture API and
frontend images to GHCR, so a normal installation pulls images; see
[development](development.md) to build from source.

## Prerequisites

| Requirement | Needed for |
| --- | --- |
| Docker Engine with the Compose v2 plugin; Git | Running the stack and cloning the repository |
| Make | Convenience targets such as `make dev`; each has an equivalent `docker compose` command |
| Go 1.25+ and Node.js 24+ | Only for local image builds or running the services from source |

## Quick start

```sh
git clone https://github.com/tahopetis/porta.git
cd porta
cp .env.example .env
make dev
```

`make dev` runs `docker compose up -d`, which pulls any image not already
present locally, then runs the host-side migration target. The API container
also applies pending migrations at startup, so the stack starts even when the
`golang-migrate` CLI is missing from the host; that step then only prints a
notice. `.env.example` targets host-run tools; the Compose API service receives
its own `PORTA_*` variables and does not read `.env` (see
[configuration](configuration.md)). The Compose stack needs no environment
variables at all; a host-run API minimally needs `DATABASE_URL`, `NEO4J_URL`,
`REDIS_URL`, and a JWT secret of at least 32 characters.

To refresh images after a release:

```sh
make pull
make docker-up
```

## URLs and first login

| Service | Host URL or address | Compose mapping |
| --- | --- | --- |
| Frontend | http://localhost:3001 | 3001 to 3000 |
| API | http://localhost:8081 | 8081 to 8080 |
| Neo4j Browser | http://localhost:7475 | 7475 to 7474 |
| Neo4j Bolt | localhost:7688 | 7688 to 7687 |
| PostgreSQL | localhost:5433 | 5433 to 5432 |
| Redis | localhost:6380 | 6380 to 6379 |

Open the frontend and sign in with the bootstrapped administrator:

| Field | Default |
| --- | --- |
| Username | `admin` |
| Password | `Admin@123` |

The API creates the administrator on startup if missing and resets its email and
password when it exists, using `ADMIN_USERNAME`, `ADMIN_EMAIL`, and
`ADMIN_PASSWORD` (Compose sets `admin`, `admin@porta.dev`, `Admin@123`). Change
these before exposure: a restart reapplies them, and the default password must
never stay active on a reachable host.

`make create-admin` is optional and only verifies those credentials; it creates nothing.

## Health and logs

```sh
docker compose ps
curl -fsS http://localhost:8081/health
docker compose logs -f api
```

`docker compose ps` shows a `healthy` status once each healthcheck passes. The API
waits for PostgreSQL, Neo4j, and Redis; Neo4j's first start can take five minutes
on slow hosts.

## Local image build (optional)

Use this only when you change a Dockerfile, a dependency, or need an offline
build. It is the escape hatch, not the normal path:

```sh
make images-build
make docker-up
```

`make images-build` uses the `docker-compose.build.yml` overlay and builds the
two images sequentially to avoid exhausting memory on small hosts. `make pull`
followed by `make docker-up` restores the published images.

## Stop, reset, and upgrade

```sh
make docker-down          # stop the stack, keep data volumes
docker compose down -v    # stop and delete named volumes (data loss)
```

Application data lives in the `porta_*` named volumes (`postgres_data`,
`neo4j_data`, `redis_data`), so `make docker-down` followed by `make docker-up`
keeps it. To upgrade, pull and restart; migrations run at API startup:

```sh
make pull
make docker-up
```

Existing Pustaka installations must follow
[upgrading from Pustaka](upgrading-from-pustaka.md) instead.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| A published port is already in use | Stop the conflicting process or change the host side of the mapping in `docker-compose.yml` |
| API container restarts or is unhealthy | `docker compose logs api`; confirm PostgreSQL, Neo4j, and Redis are healthy |
| Frontend loads but API calls fail | The API must be healthy on 8081; CORS origins must include the frontend URL being used |


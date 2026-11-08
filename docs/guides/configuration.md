# Configuration

Porta is configured entirely through environment variables. Defaults live in
`internal/config/config.go`; `.env` is optional.

## How configuration is loaded

1. Defaults are registered first (`setDefaults` in `internal/config/config.go`).
2. If a `.env` file exists in the working directory, `godotenv.Load()` reads it
   without overwriting variables that are already set in the process
   environment.
3. Each setting is then bound to an explicit environment variable. A bound
   variable wins over the default.

Every bound setting accepts two names: an unprefixed name (used by
`.env.example`) and a `PORTA_*` alias (used by `docker-compose.yml`). When both
are set, the unprefixed name wins because it is bound first. Variables that are
not bound are never read, even if they appear in `.env.example`.

Precedence: process environment, then `.env`, then the built-in default.

## Host and container addresses

The Compose file publishes host ports that differ from the ports used inside the
container network, so the same setting has two correct values depending on where
the process runs:

| Service | Host address | Container address |
| --- | --- | --- |
| PostgreSQL | localhost:5433 | postgres:5432 |
| Neo4j Bolt | localhost:7688 | neo4j:7687 |
| Redis | localhost:6380 | redis:6379 |
| API | localhost:8081 | api:8080 |

`.env.example` is written for host-run tools and points at the published ports.
The Compose API service does not read `.env`; it receives explicit `PORTA_*`
variables that use the container addresses.

## Database and Neo4j

| Variable (alias `PORTA_...`) | Default | Meaning |
| --- | --- | --- |
| `DATABASE_URL` | `postgres://porta:password@localhost:5432/porta?sslmode=disable` | PostgreSQL DSN; required |
| `DATABASE_MAX_OPEN_CONNS` | `25` | Maximum open connections in the pool |
| `DATABASE_MAX_IDLE_CONNS` | `25` | Maximum idle connections in the pool |
| `DATABASE_CONN_MAX_LIFETIME` | `5m` | Go duration after which a connection is recycled |
| `NEO4J_URL` | `bolt://localhost:7687` | Neo4j connection URL; required |
| `NEO4J_USERNAME` | `neo4j` | Neo4j user |
| `NEO4J_PASSWORD` | `password` | Neo4j password |
| `NEO4J_DATABASE` | `neo4j` | Neo4j database name |
| `NEO4J_MAX_POOL` | `50` | Maximum Neo4j driver pool size |

The unprefixed Neo4j variable is `NEO4J_URL`, not `NEO4J_URI`. Its `PORTA_*`
alias is spelled `PORTA_NEO4J_URI`; setting only `NEO4J_URI` has no effect.

## Redis

| Variable (alias `PORTA_...`) | Default | Meaning |
| --- | --- | --- |
| `REDIS_URL` | `redis://localhost:6379` | Redis URL, including credentials when required |
| `REDIS_PASSWORD` | empty | Password passed to the client separately from the URL |
| `REDIS_POOL_SIZE` | `10` | Connection pool size |

Redis is used as a cache. Authentication is stateless JWT, not Redis sessions.

## Server and JWT

| Variable (alias `PORTA_...`) | Default | Meaning |
| --- | --- | --- |
| `SERVER_HOST` | `0.0.0.0` | Listen address |
| `SERVER_PORT` | `8080` | Listen port; Compose publishes it as 8081 on the host |
| `SERVER_READ_TIMEOUT` | `30s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | HTTP keep-alive idle timeout |
| `JWT_SECRET` | `change-this-secret-key` | HS256 signing secret; set a long random value |
| `JWT_ACCESS_TOKEN_TTL` | `24h` | Access-token lifetime |
| `JWT_REFRESH_TOKEN_TTL` | `168h` | Refresh-token lifetime (7 days) |

The API does not enforce a secret length. Use at least 32 random characters.
When `ENVIRONMENT=production`, startup fails if the secret still equals the
built-in default literal `change-this-secret-key`; the placeholder in
`.env.example` is already different, so it passes that check but must still be
replaced.

## CORS

| Variable (alias `PORTA_...`) | Default | Meaning |
| --- | --- | --- |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated list of allowed browser origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization` | Allowed request headers |

## Administrator bootstrap

| Variable (alias `PORTA_...`) | Default | Meaning |
| --- | --- | --- |
| `ADMIN_USERNAME` | `admin` | Username created on startup when missing |
| `ADMIN_EMAIL` | `admin@porta.dev` | Email assigned to that user |
| `ADMIN_PASSWORD` | `Admin@123` | Password hashed with Argon2id and applied on every startup |

The bootstrap runs on each API start: it creates the user if it is missing and
otherwise resets its email and password to the configured values. Change these
values before exposing an installation.

## Logging and environment

| Variable (alias) | Default | Meaning |
| --- | --- | --- |
| `LOG_LEVEL` (`PORTA_LOGGING_LEVEL`) | `info` | Log level, for example `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` (`PORTA_LOGGING_FORMAT`) | `json` | Log encoding: `json` or `console` |
| `ENVIRONMENT` (`PORTA_ENVIRONMENT`) | `development` | Free-form environment name; only `production` adds startup checks |

Note that the prefixed aliases for logging are `PORTA_LOGGING_LEVEL` and
`PORTA_LOGGING_FORMAT`, not `PORTA_LOG_LEVEL` and `PORTA_LOG_FORMAT`.

## Configuration keys that are not wired

These names appear in `.env.example` or in the configuration struct, but the
corresponding environment variables are not bound and no code path consumes
them. Setting them has no effect.

| Key or variable | Default | Status |
| --- | --- | --- |
| `SERVER_BASE_URL` | none | Not bound and unused; no base-URL setting is read |
| `METRICS_ENABLED`, `METRICS_PATH` | `true`, `/metrics` | Defaults only; the `/metrics` handler is a TODO, so no endpoint exists (Planned) |
| `RATE_LIMIT_REQUESTS_PER_MINUTE`, `RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE` | `1000`, `60` | Defaults only; no rate-limit middleware exists (Planned) |
| `MAX_UPLOAD_SIZE` | `10MB` | Default only; no upload size or MIME check reads it |
| `security.password_min_length`, `security.account_lockout_attempts`, `security.session_timeout` | defaults for some fields | Not read anywhere; the password length check is hardcoded to 8-128 in `internal/auth/password.go` |

## Example configurations

Compose deployment: no variable is mandatory because `docker-compose.yml` sets
every `PORTA_*` value it needs. To change a value, edit the `api` service
environment or use a Compose override file.

Host-run API (for source work, see [development](development.md)):

```sh
cp .env.example .env
# edit JWT_SECRET and any credential you changed
make run
```

With `.env.example` loaded, `DATABASE_URL`, `NEO4J_URL`, and `REDIS_URL` point at
the published host ports, so an already-running Compose stack is enough. The
API listens on `SERVER_PORT=8080`, while the Compose API is published on 8081.

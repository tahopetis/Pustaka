# Contributing

Use the GitHub issue tracker for bugs and feature requests, keep changes focused,
and update the documentation that depends on the change. Issues live in
`tahopetis/porta` and are managed with the `gh` CLI.

## Before you start

- Search existing issues with `gh issue list` and read the conventions in
  `docs/agents/issue-tracker.md`.
- Read [status.md](../status.md) so a change does not describe Partial or
  Planned behavior as finished.
- Check `GLOSSARY.md` for the project's terms, and the relevant page under
  `docs/architecture/` or `docs/features/` for the area you are changing.

## Workflow

1. Create a branch from `main`; use a descriptive name such as
   `docs/refresh-current-state` or `chore/go-deps`.
2. Make the change with tests. Behavior changes need a test or an explicit note
   in the pull request explaining why one is not practical.
3. Run the local gates: `make fmt`, `make lint`, the test commands from
   [testing](testing.md), and `make docs-check` when documentation changed.
4. Commit with a conventional-commit message.
5. Push the branch and open a pull request. Workflows trigger on pull requests
   and on pushes to `main`; a plain branch push runs nothing.

## Commit style

Use `type(scope): summary`, for example `feat(amortization): add write-off
preview` or `build(deps): bump viper`. Common types are `feat`, `fix`, `docs`,
`test`, `refactor`, `build`, `ci`, and `chore`. Use the imperative mood, keep the
subject short, and explain the motivation in the body. The dependency workflow
already commits in this style.

## Pull request checks

| Workflow | Trigger | Checks |
| --- | --- | --- |
| `ci.yml` | Pull request, push to `main` | `go build`, `go vet`, `go test -p 1 ./...` with PostgreSQL and Redis services; `npx vitest run` and `npm run build` for the frontend |
| `codeql.yml` | Pull request, push to `main`, weekly schedule | CodeQL analysis for Go and JavaScript/TypeScript |
| `security.yml` | Pull request, push to `main`, weekly schedule | Blocking dependency review on PRs, `govulncheck`, `npm audit --audit-level=high`; gosec and Trivy are report-only, and Trivy is skipped on PRs |
| `docker.yml` | Pull request, push to `main` or a version tag | Builds both images for amd64 and arm64; pushes only outside pull requests |

Pull requests from forks get a read-only token: images build but are not
published, which is expected.

## Coding standards

General expectations: readable code, consistent naming, explicit error
handling, security by default, and tests for new logic.

| Area | Conventions |
| --- | --- |
| Go | Lowercase package names; exported CamelCase identifiers; `snake_case` JSON fields; wrap errors with context; `zerolog` structured logs that never include secrets or tokens |
| HTTP | `/api/v1` base path; JSON request and response bodies; `page` and `limit` for list endpoints; consistent JSON error payloads with a code and message, never a stack trace; RBAC middleware on every protected route group |
| PostgreSQL | `snake_case` plural table names; UUID primary keys; JSONB for flexible attributes with GIN indexes where queried; `ON DELETE CASCADE` where cleanup is required |
| Neo4j | PascalCase node labels; `UPPER_SNAKE_CASE` relationship types; parameterised queries only; indexes for frequently queried fields |
| Vue | Composition API, Pinia for shared state, Tailwind for styling; API calls go through `web/src/services/api.ts`; PascalCase component files |
| Files | Stay below 500 lines per file; split proactively rather than adding a large file |

Documentation and comments are in English. The phase-based coverage targets in
the archived plan (70, 85, 95 percent) are aspirational; CI does not enforce a
coverage threshold, so do not claim a number that is not measured.

## Security checklist for changes

- Enforce JWT authentication and per-route RBAC on new endpoints.
- Hash passwords with Argon2id only. The shared password service enforces an
  8-128 character length.
- Read secrets from environment variables; never commit a credential or a
  working production secret.
- Use parameterised queries for PostgreSQL and Neo4j.
- Keep CORS origins explicit through `CORS_ALLOWED_ORIGINS`; do not add a
  wildcard for a deployment that handles real data.
- Do not log passwords, tokens, or personal data. Return sanitised errors.
- Write audit rows for changes to CIs, CI types, relationship types, lifecycle
  statuses, and EA entities. Amortization changes are recorded in the
  amortization ledger and structured logs, not in `audit_logs`.
- Rate limiting, upload size or MIME enforcement, and account lockout are not
  implemented. Do not describe them as controls that exist; treating them as
  Planned is correct (see [status.md](../status.md)).
- TLS termination, encrypted backups, and external penetration testing are
  deployment responsibilities; Compose serves plain HTTP and no backup job ships
  with the repository.
- Dependency scanning runs in CI: dependency review and `npm audit` are
  blocking, `govulncheck` blocks on reachable findings, and gosec and Trivy are
  report-only.

## Issues, triage, and dependencies

Issues and specs live as GitHub issues; pull requests are not a request
surface. When an issue depends on another, use GitHub's native dependency link
or a `Blocked by: #<n>` line in the body.

| Label | Meaning |
| --- | --- |
| `needs-triage` | A maintainer must evaluate the issue |
| `needs-info` | Waiting on the reporter for more information |
| `ready-for-agent` | Fully specified and ready for an autonomous agent |
| `ready-for-human` | Requires human implementation |
| `wontfix` | Will not be actioned |

## Documentation changes

Update the page that owns the behavior: [status.md](../status.md) when the
matrix moves, `docs/architecture/api.md` and `docs/api/openapi.yaml` when routes
or payloads change, and the matching `docs/features/` page. New docs belong under
`docs/` and are linked from [the documentation hub](../README.md); historical
material belongs under `docs/archive/`. Run `make docs-check` to verify links.

## Dependency bumps

Go dependency bumps run through the `go-deps` workflow (Actions, then Run
workflow) with `module@version` inputs; the job uploads a patch instead of
opening a pull request because commits made with the workflow token cannot
trigger checks. Download the artifact, apply it on a branch, commit
`build(deps): ...`, push, and open a pull request. Frontend dependencies are
bumped in `web/package.json` and `web/package-lock.json`; CI's `npm audit` gate
must pass.

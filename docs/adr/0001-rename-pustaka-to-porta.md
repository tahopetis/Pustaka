# ADR-0001: Rename the project from Pustaka to Porta

- **Status:** Accepted
- **Date:** 2026-11-07
- **Deciders:** project owner

## Context

The project shipped as **Pustaka** — a name derived from *pustaka* (Sanskrit/Javanese:
"manuscript, book, knowledge repository"), chosen for its "library of IT assets"
resonance. The owner decided the product should be called **Porta** instead: Latin/Italian
for *door, gate*. The intended meaning is a **gateway** — the entry point to an
organisation's IT asset knowledge, which is what a CMDB actually is.

At decision time:

- There are **no deployed installations with data worth preserving** — only a single
  development environment, so no backwards-compatibility shim is warranted.
- The name was embedded in nearly every layer: Go module path, imports, logger alias,
  29 environment-variable bindings, Docker container/network/volume names, database
  role and name, JWT issuer and audiences, admin seed emails, OpenAPI documents,
  container image names, and docs.
- The GitHub repository is **public**, and runtime delivery was moved in the same
  change from host-built images to images built by GitHub Actions and published to
  GHCR (the deployment host is a low-memory tablet that should not compile the stack).

## Decision

Rename every identifier and user-facing surface, with these forms:

| Context | Form |
| --- | --- |
| User-facing text, docs, page title | `Porta` |
| Machine identifiers (module, repo, containers, DB, npm package) | `porta` |
| Environment variables | `PORTA_` prefix (plain unprefixed names remain supported) |
| Go module | `github.com/tahopetis/porta` |
| Repository | `tahopetis/porta` |
| Images | `ghcr.io/tahopetis/porta-api`, `ghcr.io/tahopetis/porta-frontend` |

**No compatibility shim.** The old `PUSTAKA_*` environment prefix is deleted rather
than deprecated: a third accepted spelling is a permanent "which name wins?" trap, and
there are no consumers to protect.

## Consequences

- Anyone with an existing checkout or environment must rename `PUSTAKA_*` variables,
  recreate the database/role, and re-authenticate (the JWT issuer and audiences moved
  from `pustaka`/`pustaka-refresh` to `porta`/`porta-refresh`, invalidating tokens).
  See `docs/guides/upgrading-from-pustaka.md`.
- The GitHub repository rename auto-redirects old URLs, but `.beads` issue IDs
  (`pustaka-*`) are immutable references and are **not** renamed.
- Historical records are not rewritten (see Exemptions) so past decisions stay
  truthful; `git grep -i pustaka` will therefore still return hits in archived files.
- The name `porta` is generic and already used by other public software (for example
  Red Hat's 3scale `porta`); the repository description disambiguates. No trademark is
  claimed or defended.
- Because the working directory path cannot be moved by a commit, the local clone is
  renamed last, after the GitHub repository (the redirect absorbs stale URLs meanwhile).

## Exemptions (intentionally still contain "Pustaka")

These are historical or machine-generated records; rewriting them would falsify history
or break immutable identifiers:

- `.planning/**` — agent-generated project state (phase summaries, research)
- `plan/**` — founding specifications, including the FSD/TSD etymology of the old name
- `.beads/**` — issue-tracker state; IDs such as `pustaka-pdq` are immutable references
- `docs/implementation/**` — records of completed work
- `docs/planning/prd.md` — dated duplicate PRD marked "Final"
- `docs/amortization/AMORTIZATION_TESTS_OVERVIEW.md`
- `docs/guides/bmm-workflow-status.md`
- `docs/analysis/project-scan-report.json`
- `docs/planning/source-tree-analysis.md`

In addition, a few **rename-record** files necessarily name the old product so that
humans and agents can find it: this ADR, `docs/guides/upgrading-from-pustaka.md`, the
single "formerly Pustaka" line in `README.md`, and the one-line link to that guide in
`docs/guides/README.md`.

## Not decided here

- A new logo/wordmark/favicon (the project currently ships none; branding is text-only)
- The `porta.dev` domain that appears in OpenAPI examples is placeholder fiction
- Product version: no git tags or releases exist, so `1.0.0` is unchanged
- Host port remapping and CORS-origin cleanup (deliberately out of scope)

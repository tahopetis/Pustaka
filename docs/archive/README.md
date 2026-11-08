# Archived Documentation

This directory holds **historical documents kept for provenance**. They describe
earlier plans, designs, and implementation attempts. They are **not maintained**
and are **not** a statement of how Porta works today.

Archived documents may be stale or simply wrong:

- old **ports** (for example `8080` where the API now runs on `8081`),
- old **paths** and repository layouts,
- old **commands** and `make` targets that no longer exist,
- old project names, including **Pustaka** before the rename to **Porta**,
- features that were described here but never shipped, or shipped differently.

Do not update these files to match the current code, and do not cite them as
current behavior. If an archived document contradicts the code or the live docs,
the code and the live docs win.

## Current documentation

Start here instead:

| Document | What it covers |
| --- | --- |
| [docs/README.md](../README.md) | Documentation map and entry point |
| [docs/status.md](../status.md) | What is implemented, partial, or planned |
| [docs/architecture/overview.md](../architecture/overview.md) | System context, stack, topology |
| [docs/features/](../features/) | CMDB, Enterprise Architecture, amortization, dashboards, audit |
| [docs/guides/](../guides/) | Installation, configuration, development, testing, contributing |

## What is in here

| Folder | Provenance | Contents |
| --- | --- | --- |
| `plan/` | Repository-root `plan/` directory | Early FSD, TSD, implementation plan, RBAC API specification, and the setup README. |
| `planning/` | `docs/planning/` (Pustaka era) | GDD, lower-case PRD, epics, implementation roadmap, source-tree analysis, UX/UI specifications. |
| `implementation/` | `docs/implementation/` | Per-change implementation notes and progress summaries (TrendChart, bidirectional relationships, CI lifecycle, dashboard, relationship types). |
| `testing/` | `docs/` root | CSV import, EA RBAC, and warn-but-allow testing notes. |
| `misc/` | Repository root, `docs/`, `docs/analysis/`, `docs/amortization/`, `docs/guides/` | One-off analyses, verification notes, amortization backup/metrics plans, and legacy root instructions and plan documents. |

# Porta Documentation

This is the entry point for Porta's documentation. Start with the path that
matches what you are trying to do. For what is actually implemented, partial, or
planned, read [status.md](status.md) first.

## Using Porta

| I want to... | Read |
| --- | --- |
| Install and run Porta for the first time | [guides/installation.md](guides/installation.md) |
| Understand the environment variables and defaults | [guides/configuration.md](guides/configuration.md) |
| Migrate an existing Pustaka install | [guides/upgrading-from-pustaka.md](guides/upgrading-from-pustaka.md) |
| Manage CIs, CI types, relationships, and the graph | [features/cmdb.md](features/cmdb.md) |
| Use the Enterprise Architecture layer | [features/enterprise-architecture.md](features/enterprise-architecture.md) |
| Configure and run asset amortization | [features/amortization.md](features/amortization.md) |
| Understand dashboards and analytics | [features/dashboards-analytics.md](features/dashboards-analytics.md) |
| Review audit logging and compliance | [features/audit-compliance.md](features/audit-compliance.md) |

## Developing Porta

| I want to... | Read |
| --- | --- |
| Get the system context, stack, and topology | [architecture/overview.md](architecture/overview.md) |
| Understand the database schema | [architecture/data-model.md](architecture/data-model.md) |
| Look up HTTP routes and the auth model | [architecture/api.md](architecture/api.md) |
| Call the API | [api/api-guide.md](api/api-guide.md) and [api/api-examples.md](api/api-examples.md) |
| Set up a local workflow and run commands | [guides/development.md](guides/development.md) |
| Run the test suites | [guides/testing.md](guides/testing.md) |
| Contribute a change | [guides/contributing.md](guides/contributing.md) |
| Read the product requirements and the historical backlog | [product/prd.md](product/prd.md), [product/epics-user-stories.md](product/epics-user-stories.md) |
| Learn why the project was renamed | [adr/0001-rename-pustaka-to-porta.md](adr/0001-rename-pustaka-to-porta.md) |

## Agent guidance

| I want to... | Read |
| --- | --- |
| Give a coding agent repository context | [../CLAUDE.md](../CLAUDE.md) |
| Resolve domain vocabulary | [../GLOSSARY.md](../GLOSSARY.md) |
| File and triage issues | [agents/issue-tracker.md](agents/issue-tracker.md) and [agents/triage-labels.md](agents/triage-labels.md) |
| Consume the repo's domain documentation | [agents/domain.md](agents/domain.md) |
| Check what is real before claiming a feature | [status.md](status.md) |

## Where things live

| Path | Contents |
| --- | --- |
| `docs/status.md` | Feature status matrix (Implemented / Partial / Planned). |
| `docs/architecture/` | System overview, data model, and API route reference. |
| `docs/features/` | CMDB, Enterprise Architecture, amortization, dashboards, and audit. |
| `docs/features/amortization/` | Amortization PRD, design, and data-validation rules. |
| `docs/guides/` | Installation, configuration, development, testing, contributing, upgrade. |
| `docs/api/` | API guide and worked examples. |
| `docs/product/` | Product requirements and the aspirational backlog (intent, not implemented status). |
| `docs/adr/` | Architecture decision records. |
| `docs/agents/` | Issue tracker, triage labels, and domain-doc conventions. |
| `docs/archive/` | Superseded documents kept for provenance; unmaintained. |

## Conventions

- Status terms are fixed: **Implemented** works end-to-end, **Partial** is shipped
  but stubbed or limited, **Planned** has no implementation. See
  [status.md](status.md) for the definitions and the matrix.
- Documents under [archive/README.md](archive/README.md) are historical and are
  not maintained; do not treat them as current behavior.
- If a document contradicts [status.md](status.md) or the code, the code and
  `status.md` win.

# Glossary

Canonical vocabulary for Porta, used across the codebase, the API, and these docs.
Where a plain-language term maps to a database table or type, the identifier is
shown in the definition. See [docs/status.md](docs/status.md) for which of these
capabilities are implemented today.

| Term | Definition |
| --- | --- |
| Porta | The project: an open-source Configuration Management Database with an Enterprise Architecture layer and an IT asset amortization module. Formerly named Pustaka; see [ADR-0001](docs/adr/0001-rename-pustaka-to-porta.md). |
| CMDB | Configuration Management Database — the system of record for configuration items and the relationships between them. |
| Configuration item (CI) | One managed asset, stored in `configuration_items`: a name, a CI type, a JSONB `attributes` object, `tags`, and an optional lifecycle status. |
| CI type | A reusable schema for CIs, stored in `ci_type_definitions`. It declares `required_attributes` and `optional_attributes` (each an attribute definition), and whether the type is amortizable (`is_amortizable`). CI types are flat: there is no separate domain/category/subcategory table. |
| Attribute definition | One named field inside a CI type or relationship type schema: name, type, description, and optional validation (pattern, length, numeric range, enum, or format). |
| Domain / category / subcategory | Pustaka-era taxonomy vocabulary (Domain → Category → Subcategory → CI type). It is **not** implemented as schema; assets are grouped by CI type and, for EA content, by EA domain. |
| Lifecycle status | A named state an asset can be in, stored in `lifecycle_statuses`. Each status carries an `amortization_behavior` of `pending`, `active`, or `terminal`. |
| Relationship | A directed edge between two CIs, stored in `relationships` and mirrored in Neo4j for graph traversal. |
| Relationship type | A reusable definition in `relationship_types`: forward/reverse labels, category, allowed source and target CI types, cardinality, and whether it is bidirectional. |
| Cardinality | The multiplicity the relationship type permits on each side: `one` or `many`, combined as source-to-target (for example one-to-many). |
| RBAC | Role-based access control. Seeded roles are `admin`, `editor`, and `viewer`; route access is granted through granular permissions. |
| Permission | A named capability such as `ci:read`, `user:create`, `ea:create`, or `amortization:configure`, attached to roles and checked by routing middleware. |
| Access token / refresh token | JWTs issued at login. The access token (24h default) authenticates API calls; the refresh token (168h default) obtains a new access token. |
| Audit log | A record of a state change, stored in `audit_logs` with entity type, action, performer, timestamp, IP address, and user agent. The API can list, filter, and export entries; pruning exists but is gated behind a permission no seeded role has. |
| EA domain | One of the eight fixed Enterprise Architecture groupings: Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance. |
| EA entity | An Enterprise Architecture record modeled as a CI whose type name begins with `EA.` (for example `EA.Application-BusinessApplication`). Managed through `/api/v1/ea/*` and the Enterprise Architecture UI. |
| EA team | An organizational owner record in `ea_teams`, used when assigning EA entities. |
| Warn-but-allow | The EA validation pattern: required-attribute and cross-field errors block the save (HTTP 422), while optional-attribute problems are recorded as warnings and do not block. An explicit override with a justification bypasses EA-specific validation. |
| Data quality score | A 0–100 score on an EA entity reflecting how completely its CI type's attributes are populated. Surfaced in the EA data-quality dashboard. |
| Health score | A 0–100 indicator of CMDB quality calculated from completeness, correctness, and compliance metrics. Snapshots are stored in `health_score_history` so the dashboard can show a trend. |
| Purchase cost | The original acquisition cost recorded on an amortizable CI (`purchase_cost`). Also called original capitalized cost (OCC) in amortization reports. |
| Salvage value | The residual value an asset is expected to retain at the end of its useful life (`salvage_value`). Depreciation stops at this floor. |
| Useful life | The depreciation period in months (`useful_life_months`) over which the cost of an asset is allocated. |
| Accumulated depreciation (AD) | The total depreciation charged against an asset since amortization started. It is derived from the amortization ledger rather than stored on the CI. |
| Net book value (NBV) | The asset's current value after accumulated depreciation, stored as `current_book_value`. |
| Gross book value (GVB) | Original capitalized cost plus net ledger adjustments, used as the cost basis for reports. |
| Amortization | Straight-line allocation of an asset's cost (purchase cost minus salvage value) over its useful life. |
| Amortization behavior | The `pending`/`active`/`terminal` setting on a lifecycle status that tells the amortization service whether an asset should be depreciated or written off. |
| Amortization run | One batch execution recorded in `amortization_runs`; it can be manual (`POST /api/v1/amortization/runs`) or scheduled. |
| Scheduled run | A run intended to execute on a daily timer. The entry point exists but is not wired to a timer today, so runs are started manually. |
| Ledger entry | One append-only row in `amortization_ledger`. The schema allows seven entry types (`depreciation`, `monthly_depreciation`, `catch_up_depreciation`, `write_off`, `adjustment`, `correction`, `restructuring`); the service currently writes five of them. |
| Adjustment entry | A ledger entry that adjusts an asset's book value up or down, leaving earlier entries intact. No reason or sign validation is enforced today. |
| Restructuring | A prospective change to an asset's useful life, recorded as a `restructuring` ledger entry so the remaining depreciable value is recalculated without rewriting history. |
| Notification | An in-app toast shown by the web UI. There is no server-side notification store or delivery channel. |
| Implemented / Partial / Planned | The status vocabulary used in [docs/status.md](docs/status.md): works end-to-end / shipped but stubbed or limited / no implementation. |

# Data Model

Porta stores its authoritative data in PostgreSQL; the schema is created at API
startup from the numbered SQL migrations in `cmd/migrations` (runner:
`internal/database/migrate.go`). Neo4j mirrors configuration items and
relationships for graph queries but is not the source of truth; Redis holds only
cache entries. CIs use a *flexible schema*: a CI type declares required/optional
attribute definitions as JSONB and each CI stores its values as JSONB, so there
is no fixed column per attribute.

## Tables at a glance

| Table | Purpose | Key columns |
|---|---|---|
| `users` | Login identities | `id`, `username`, `email`, `password_hash`, `is_active`, `last_active` |
| `roles` | Named roles | `id`, `name` (admin/editor/viewer), `description` |
| `permissions` | Permission catalogue | `id`, `name` (`resource:action`), `resource_type` |
| `role_permissions` | Role -> permission mapping | `role_id`, `permission_id` |
| `user_roles` | User -> role mapping | `user_id`, `role_id`, `assigned_by`, `assigned_at` |
| `ci_type_definitions` | CI type schemas | `name` (unique), `required_attributes`, `optional_attributes`, `is_amortizable` |
| `configuration_items` | CI records | `name`, `ci_type` (FK to type name), `attributes`, `tags`, `lifecycle_status_id`, financial columns |
| `relationships` | CI-to-CI edges (relational copy) | `source_id`, `target_id`, `relationship_type`, `attributes` |
| `relationship_types` | Relationship type catalogue | `name` (unique), labels, allowed types, cardinality, direction flags |
| `lifecycle_statuses` | CI status catalogue | `name` (unique), `display_name`, `amortization_behavior`, `sort_order` |
| `audit_logs` | Audit trail | `entity_type`, `entity_id`, `action`, `performed_by`, `details`, `ip_address` |
| `amortization_runs` | Depreciation run tracking | `run_date`, `status`, counters, checkpoint data, `is_manual` |
| `amortization_ledger` | Append-only financial ledger | `ci_id`, `entry_date`, `entry_type`, `amount`, book values, `sequence_number` |
| `amortization_summaries` | Per-CI reporting snapshots | `ci_id`, `reporting_date`, `current_book_value`, `amortization_status` |
| `health_score_history` | CMDB health trend samples | `calculated_at`, `overall_score`, completeness/correctness/compliance scores |
| `ea_teams` | EA team ownership list | `name` (unique), `description` |

## Identity and access

`users` holds `id`, unique `username`/`email`, an Argon2id `password_hash`,
`is_active`, `last_active`, and timestamps; the admin user is bootstrapped at
startup (`cmd/api/main.go:34-115`). `role_permissions` and `user_roles` are join
tables; `user_roles` also records `assigned_by` and `assigned_at`.

Migration 001 seeds 3 roles and 18 permissions and grants the admin role every
permission existing then. Later migrations add 4 relationship-type, 4
lifecycle-status, 5 amortization, and 4 EA permissions, for **35 seeded
`resource:action` permission names**. Login loads them into the JWT and RBAC
middleware compares them by exact match. Known gaps: `audit:delete` is required
by two audit routes but never inserted, so no seeded role can call them;
`system:admin` and `amortization:adjust` are seeded but unused by the backend.

## CI types and configuration items

`ci_type_definitions` (`cmd/migrations/001_initial_schema.up.sql:54-63`,
extended by 006) stores `id`, unique `name`, `description`, the JSONB arrays
`required_attributes`/`optional_attributes`, `is_amortizable` (default false),
and audit columns. An attribute definition is a JSON object:

```json
{ "name": "hostname", "type": "string", "description": "Server hostname",
  "validation": { "pattern": "^[a-zA-Z0-9-]+$", "min_length": 1, "max_length": 253 } }
```

`type` is `string`, `integer`, `boolean`, `array`, or `object`; `validation` may
carry `pattern`, `min_length`, `max_length`, `min`, `max`, `enum`, and `format`
(`email`, `url`, `ipv4`, `date`, `datetime`), implemented in
`internal/ci/models.go` (`ValidateAttributes`). `configuration_items`
(`001:66-79`, extended by 005 and 006) stores `id`, `name` (unique together with
`ci_type`), `ci_type` (FK to the type name), JSONB `attributes`, GIN-indexed
`tags` (`TEXT[]`), `lifecycle_status_id` (005), `purchase_cost`,
`salvage_value`, `current_book_value` (`DECIMAL(15,2)`), `amort_start_date`
(`DATE`), and `useful_life_months` (`INTEGER`) from 006, plus audit columns.

Check constraints enforce non-negative amounts, `salvage_value <=
purchase_cost`, positive useful life, and consistent dates. A `BEFORE INSERT OR
UPDATE` trigger (`update_ci_book_value_trigger`) initialises `current_book_value`
from `purchase_cost` when amortization has not started.

## Relationships and relationship types

`relationship_types` (created by 002, reshaped by 003 and 004) holds unique
`name`, `display_name`, `description`, `forward_label`, `reverse_label`,
`color`, `icon`, `category`, `source_types`/`target_types` (original arrays),
`allowed_source_types`/`allowed_target_types` (used by the service),
`allow_same_type`, `allow_self_reference`, `cardinality_source` and
`cardinality_target` (`one` or `many`), `bidirectional`, `is_active`,
`is_system`, `attributes` JSONB, and audit columns.

`relationships` (`001:82-96`) stores `source_id`, `target_id`,
`relationship_type`, `attributes` JSONB, and audit columns, with
`no_self_relationship` and `unique_relationship` on
`(source_id, target_id, relationship_type)`. `relationship_type` and the
`allowed_*_types` arrays are plain values, not foreign keys, so type existence is
not enforced by the database; the CI type link is
(`configuration_items.ci_type` references `ci_type_definitions.name`). Rows are
written to PostgreSQL first and mirrored to Neo4j; graph sync failures are logged,
not fatal (`internal/ci/service.go:402-413`). Migrations 001 and 009 seed 35 CI
types (3 base — `Server`, `Application`, `Database` — plus 32
`EA.<domain>-<type>` types). Migration 002 seeds 5 relationship types and 009
supplies 33 further EA edge definitions. Because `relationship_types.name` is
unique and inserts use `ON CONFLICT (name) DO NOTHING`, repeated EA names collapse
(five `contains` variants become one) and the shared `depends_on` is skipped,
leaving **30 relationship type rows**.

## Lifecycle statuses

`lifecycle_statuses` (`005:7-24`, extended by 006 and 011) stores unique `name`
(snake_case), `display_name`, `description`, `color`, `icon`, `sort_order`,
`is_active`, `is_system`, `amortization_behavior`, and audit columns.
`amortization_behavior` is constrained to `pending`, `active`, or `terminal`
(`006:9-12`): `pending` assets are not processed, `active` assets depreciate,
`terminal` assets are written off.

Migration 005 inserts 10 inventory statuses (`planned`, `on_order`, `in_stock`,
`pending_install`, `operational`, `in_maintenance`, `defective_repair`,
`retired`, `disposed`, `missing_stolen`); 011 inserts 5 EA statuses (`proposed`,
`active`, `deprecated`, `retired`, `archived`). Since `name` is unique and 011
uses `ON CONFLICT (name) DO NOTHING`, the EA `retired` row is skipped and the
catalogue holds **14 rows**. 006 maps `planned`/`on_order` to `pending`,
`disposed`/`missing_stolen` to `terminal`, and the rest to `active`.

## Audit trail

`audit_logs` (`001:99-109`): `entity_type`, `entity_id`, `action`,
`performed_by` (FK users), `timestamp`, `details` JSONB, `ip_address` (INET),
`user_agent`. Rows are written by **domain services** (CI, CI type, relationship,
relationship type, lifecycle, EA) through `internal/ci/audit_service.go`; the HTTP
audit middleware only emits structured log entries, and amortization events are
application log lines, not `audit_logs` rows. Rows are read through
`/api/v1/audit/*`; retention is operator-driven through the cleanup endpoint.

## Amortization tables

Three tables exist; a fourth table referenced by the repository does not.

`amortization_runs` (`006:40-76`): one row per execution with `run_date`,
`status` (`pending`, `started`, `running`, `completed`, `partial`, `failed`,
`cancelled`), timing columns, counters, `last_processed_ci_id` +
`checkpoint_data` for restart, error fields, `run_config`, and `is_manual`. A
partial unique index permits many manual runs but one scheduled run per date.

`amortization_ledger` (`006:79-118`): append-only entries with `ci_id`,
`entry_date`, `entry_type`, `amount`, book values, `accumulated_depreciation`,
period/adjustment fields, `corrects_entry_id` (self-reference),
`batch_run_id`/`amortization_run_id`, `is_system_generated`, `created_by`, and a
per-CI-per-date `sequence_number` assigned by a trigger; uniqueness is
`(ci_id, entry_date, sequence_number)`. The `entry_type` check accepts
`depreciation`, `monthly_depreciation`, `catch_up_depreciation`, `write_off`,
`adjustment`, `correction` (006) and `restructuring` (007); the current
calculator and services produce `monthly_depreciation`, `catch_up_depreciation`,
`write_off`, `adjustment`, and `restructuring`, but not `depreciation` or
`correction`.

`amortization_summaries` (`006:121-143`): per-CI `reporting_date` snapshots of
`current_book_value`, `accumulated_depreciation`, period amounts, and
`amortization_status` (`pending`, `active`, `completed`, `written_off`), unique
on `(ci_id, reporting_date)`, maintained by an `AFTER INSERT` trigger on the
ledger. Two views exist: `active_amortization_assets` (assets still being
depreciated) and `amortization_report` (latest summary per CI with year-to-date
depreciation and adjustment totals).

There is **no `amortization_settings` table**: `internal/amortization/repository.go`
queries one, but no migration creates it, so both settings endpoints fail at the
database layer although the read path has a default-settings branch.

## Health scores and EA storage

`health_score_history` (008) stores one row per calculation: `calculated_at`,
`overall_score`, `completeness_score`/`correctness_score`/`compliance_score`
(0-100 checks), and CI counts. `/api/v1/dashboard/health-score` computes a score
on demand and appends a row (`internal/ci/repository.go:1140-1163`).

EA entities have **no dedicated table**: they are `configuration_items` whose
`ci_type` starts with `EA.`, filtered with `ci_type LIKE 'EA.%'`
(`internal/ea/repository.go:307-309,429-435`). EA domains (`Strategy`,
`Business`, `Application`, `Data`, `Technology`, `Infrastructure`, `Security`,
`Governance`) are hard-coded in `internal/ea/models.go:10-33`; team ownership is
the `ea_teams` table (`009:63-70`, 8 seeded teams). There are no `ea_entities`,
`ea_domains`, or `ea_relationships` tables.

## What does not exist

- No `domains`, `categories`, or `subcategories` tables; hierarchy is CI type
  plus `EA.<domain>-<type>` naming and `parent_*_id` JSONB attributes for some
  EA types.
- No `ea_entities` or `amortization_settings` table, and no tenancy,
  soft-delete, row-version history, or sessions table.

## Keys, triggers, and views

| Object | Detail |
|---|---|
| Indexes | B-tree on foreign keys/filters; GIN on `configuration_items.attributes`/`.tags` and `relationship_types` arrays; GIN full-text on CI names; partial indexes on financial columns; ledger composite on `(ci_id, entry_date, entry_type)` |
| Update triggers | `update_updated_at_column()` on users, CI types, CIs, relationships, relationship types, lifecycle statuses, EA teams |
| Amortization triggers | `update_ci_book_value_trigger`, `set_ledger_sequence_trigger`, `update_amortization_summary_trigger` |
| SQL helpers | `calculate_monthly_depreciation`, `calculate_write_off_amount`, `get_next_amortization_sequence` |

Migrations 001-011: identity and core CMDB (001), relationship types (002-004),
lifecycle statuses (005), amortization (006-007), health, admin and EA (008-011).

## Related documentation

- [Architecture overview](overview.md) and [API reference](api.md)
- [Project status](../status.md) and [configuration guide](../guides/configuration.md)
- [ADR-0001: rename Pustaka to Porta](../adr/0001-rename-pustaka-to-porta.md)

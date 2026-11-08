# Amortization

The amortization module tracks depreciation of IT assets: CI types can be marked as amortizable, lifecycle
statuses carry an amortization behaviour, configuration items get financial columns, and every financial event is
appended to an immutable ledger. The reporting layer exposes book value, accumulated depreciation, and
accounting summary metrics.

**Status: Implemented core, Partial automation.** Financial data, the ledger, balance maintenance, manual runs,
adjustments, restructuring, and reporting work. The daily scheduler is not wired: the startup hook calls
`ScheduleDailyRun`, which only writes a log line, and the processing entry points
`ProcessDailyAmortization` and `HandleTerminalStatusChange` have no caller outside tests. Manual runs must be
triggered through the API or the scheduler screen.

## Configuration

| Setting | Where it lives | Notes |
| --- | --- | --- |
| Amortizable CI type | `ci_type_definitions.is_amortizable` | Exposed as a checkbox in CI type management |
| Amortization behaviour | `lifecycle_statuses.amortization_behavior` | `pending`, `active`, or `terminal` |
| Currency and default useful life | `amortization_settings` (repository) | No migration creates this table; settings endpoints are Partial |
| Financial fields | `configuration_items` columns | Purchase cost, salvage value, start date, useful life, current book value; accumulated depreciation is derived as cost minus book value |

The financial columns and the ledger, run, and summary tables are created by the amortization module migration.
The global settings endpoints (`GET`/`PUT /amortization/settings`) and their screen exist, but the backing table
is absent from the migration set, so they are not usable end to end today.

## Data Model

| Object | Purpose |
| --- | --- |
| Financial CI columns | Cost basis, salvage value, amortization start date, useful life in months, current book value |
| `amortization_ledger` | Append-only entries with amount, book value before/after, accumulated depreciation, run reference, and sequence number |
| `amortization_runs` | Execution records with status, run date, counters, checkpoint fields, and manual/scheduled flag |
| `amortization_summaries` | Per-CI reporting snapshots maintained by a database trigger on ledger inserts |
| Views | `active_amortization_assets` and `amortization_report` |

Ledger entry types accepted by the database constraint are `monthly_depreciation`, `catch_up_depreciation`,
`write_off`, `adjustment`, `correction`, and `restructuring` (a legacy `depreciation` value is also accepted).
Current code writes monthly depreciation, catch-up depreciation, write-offs, adjustments, and restructuring
entries; correction entries are accepted by the schema but not produced.

## Calculation

The implemented calculator is straight-line only:

- monthly depreciation is `(purchase_cost - salvage_value) / useful_life_months`;
- the first processing of an asset writes a catch-up entry for the elapsed portion of its life;
- the book value never falls below the salvage value;
- a terminal lifecycle behaviour writes off the remaining book value down to the salvage value;
- useful-life changes use prospective recalculation: history is preserved, and the current book value is spread
  over the remaining months with a restructuring entry;
- corrections are additive adjustment entries; ledger rows are never deleted.

The schema also contains `period_start_date`, `period_end_date`, and `days_in_period` columns, but the current
scheduler does not populate them, so mid-month proration as described in older design material is not active.

## API

All routes are under `/api/v1/amortization`. Read endpoints require `amortization:read`; configuration, adjustments,
and restructuring require `amortization:configure`; triggering a run requires `amortization:run`.

| Endpoint | Purpose | Permission |
| --- | --- | --- |
| `GET /configuration-items` | List amortizable CIs with financial data | `amortization:read` |
| `GET /configuration-items/{ciId}` | Amortization details and schedule for one CI | `amortization:read` |
| `PUT /configuration-items/{ciId}` | Update purchase cost, salvage value, start date, useful life, method | `amortization:configure` |
| `GET /ledger` | Paginated ledger with filters | `amortization:read` |
| `GET /ledger/{entryId}` | Single ledger entry | `amortization:read` |
| `GET /runs` | Paginated run history | `amortization:read` |
| `GET /runs/{runId}` | Single run | `amortization:read` |
| `POST /runs` | Trigger a manual run (optional date override, CI subset, dry run) | `amortization:run` |
| `GET /summaries` | Grouped accounting summary | `amortization:read` |
| `GET /reports/depreciation-schedule` | Depreciation schedule as JSON, or CSV with `format=csv` | `amortization:read` |
| `GET /settings` | Global settings | `amortization:read` |
| `PUT /settings` | Update global settings | `amortization:configure` |
| `POST /adjustments` | Create an adjustment entry | `amortization:configure` |
| `POST /restructuring/preview` | Preview a useful-life change | `amortization:configure` |
| `POST /restructuring` | Execute a useful-life change | `amortization:configure` |

## Frontend

| Route | Screen | Permission |
| --- | --- | --- |
| `/amortization` | Dashboard with accounting cards, charts, recent ledger activity | `amortization:read` |
| `/amortization/ledger` | Ledger list with entry-type and date filters | `amortization:read` |
| `/amortization/reports` | Asset schedule, depreciation chart, journal and export | `amortization:read` |
| `/amortization/settings` | Global settings form | `amortization:admin` |
| `/amortization/scheduler` | Run history and manual run trigger | `amortization:admin` |

The CI edit form shows the financial fields only for amortizable CI types. The reports screen displays OCC (original
capitalized cost), GVB (gross book value), NBV (net book value), AD (accumulated depreciation), and salvage value
summary cards, OCC/GVB reference lines on the depreciation chart, and a monthly table; CSV export is available.

## Automation Limits

| Item | Status | Detail |
| --- | --- | --- |
| Daily scheduled depreciation | Partial | Startup calls a scheduler method that only logs |
| Automated run execution | Partial | Scheduler processing code exists but is not invoked by any production caller |
| Terminal status write-off trigger | Partial | Handler method exists but is not called when a CI changes lifecycle status |
| Manual run | Implemented | `POST /amortization/runs` and the scheduler screen |
| Settings table | Planned | No migration; settings are read with a fallback but the query fails at the database layer |

The lifecycle terminal behaviour is still honoured when an asset is processed by a run; the missing piece is the
automatic trigger on status change.

## Related Documentation

- [Amortization PRD](amortization/prd.md)
- [Amortization design](amortization/design.md)
- [Data validation rules](amortization/data-validation.md)
- [Project status](../status.md)

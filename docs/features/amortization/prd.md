# Amortization PRD

This is the product requirements document for IT asset amortization and lifecycle financials. The feature is
implemented in the current codebase except for automation, which is Partial. Requirements below are annotated with
their current implementation status: **Implemented**, **Partial**, or **Planned**.

Status of the module overall, and links to the design and validation detail docs:

- [Amortization feature overview](../amortization.md)
- [Design](design.md)
- [Data validation](data-validation.md)
- [Project status](../../status.md)

## 1. Executive Summary

**Objective:** provide a status-aware financial engine for IT assets inside the CMDB. The module tracks
depreciation over time, handles lifecycle events (repair, retirement, disposal, loss), and produces monthly
financial reporting.

**Key value:**

- **Automation:** removes manual spreadsheet tracking for planned depreciation runs. *(Partial: the schedule is
  defined but not wired to a running job; manual runs work.)*
- **Accuracy:** ledger-based book value with append-only history and adjustments instead of destructive edits.
- **Compliance:** assets reach their terminal value through explicit write-offs with an audit-friendly ledger.

## 2. Configuration

### 2.1 Global settings

**Status: Partial.** The settings are modelled as a single global record with a currency and a default useful life
in months, edited through `/amortization/settings`. The settings screen and API endpoints exist, but the
`amortization_settings` table is not created by any migration, so reads and writes fail at the database layer in
the current schema. There are no `AMORTIZATION_*` environment variables; earlier drafts of this PRD assumed them.

| Setting | Default | Meaning |
| --- | --- | --- |
| Currency | USD (fallback) | Single display currency for the module |
| Default useful life | 36 months (fallback) | Pre-filled when configuring an asset |

### 2.2 CI type configuration

**Status: Implemented.** A CI type carries an `is_amortizable` flag, edited as a checkbox in CI type management.
When true, the financial fields and the Financials section appear on the CI edit form; when false they are
hidden.

### 2.3 Lifecycle status configuration

**Status: Implemented.** Each lifecycle status carries `amortization_behavior`, selected in the lifecycle status
form, one of:

| Behaviour | Meaning | Example statuses |
| --- | --- | --- |
| `pending` | No depreciation yet | Planned, On Order |
| `active` | Normal depreciation | In Stock, Operational, In Maintenance |
| `terminal` | Write off remaining value | Disposed, Missing/Stolen |

The database default is `pending`. The migration sets behaviours for the seeded inventory statuses.

## 3. Asset Data Management

### 3.1 Creation phase

**Status: Implemented.** Financial fields are not required when an asset is first created; they are configured
after procurement through the edit flow.

### 3.2 Edit phase (Financials)

**Status: Implemented.** For amortizable CI types the edit form shows:

- **Purchase cost** - original cost basis, non-negative, must be at least the salvage value;
- **Salvage value** - expected residual value, defaulted from settings when available;
- **Amortization start date** - the in-service date; depreciation is not calculated before it;
- **Useful life** - months, must be positive (the form accepts 1-600);
- **Current book value** - read-only, maintained by the module;
- **Accumulated depreciation** - read-only, derived as cost minus current book value.

## 4. Lifecycle Logic

The module reads the asset's lifecycle status and its `amortization_behavior`:

| Behaviour | Amortization action |
| --- | --- |
| `pending` | Do not process |
| `active` | Depreciate monthly while book value is above salvage value |
| `terminal` | Write off the remaining book value when the asset is processed |

### 4.1 Disposal (write-off)

**Status: Partial.** The write-off calculation and ledger entry are implemented, and a run processes terminal
assets. The automatic trigger on lifecycle status change is not wired: `HandleTerminalStatusChange` has no caller
outside tests, so a status change alone does not create a write-off until a run processes the asset.

When an asset is written off:

1. the write-off amount is `current_book_value - salvage_value`;
2. the closing book value becomes the salvage value (the residual is retained, not zeroed);
3. a `write_off` ledger entry records the before/after values and the run reference;
4. the asset is no longer depreciated once processing is complete.

This differs from earlier drafts that zeroed the asset; the implemented engine preserves the salvage value.

## 5. Calculation Formulas

**Implemented, straight-line only.**

### 5.1 Monthly depreciation

```
monthly_depreciation = (purchase_cost - salvage_value) / useful_life_months
```

The book value is floored at the salvage value. A value below salvage is never produced.

### 5.2 Catch-up depreciation

The first processing of an asset with an amortization start date in the past writes a single catch-up entry for the
elapsed portion, instead of one entry per missed month. Subsequent runs write monthly depreciation entries.

### 5.3 Prospective recalculation (useful life change)

**Status: Implemented.** Changing useful life is a restructuring operation: the new monthly amount is
`(current_book_value - salvage_value) / remaining_months` using the new remaining life. Historical ledger entries
are unchanged; a `restructuring` entry (amount zero) records the old and new useful life, the old and new monthly
amounts, the change, and the reason. Preview and execute are separate endpoints.

### 5.4 Mid-month proration

**Status: Planned.** The ledger has period start/end and days-in-period columns, but the current processor does not
populate them and does not prorate the first month by days. Monthly depreciation is a flat amount.

## 6. Automation and Triggers

### 6.1 Scheduler

**Status: Partial.** The intended design is a daily run that processes all amortizable assets whose behaviour is
active and whose book value is above salvage. The codebase contains this processing logic, run records, and
checkpoint columns, but the startup hook calls a scheduler method that only logs, and the actual processing entry
point has no production caller. In practice depreciation is applied by manual runs, not by a background job.

### 6.2 Status change trigger

**Status: Partial.** The service exposes a terminal-status handler and the write-off calculation it needs, but no
caller invokes it from the CI lifecycle path. Automated write-off on disposal is therefore not active.

### 6.3 Manual run

**Status: Implemented.** `POST /amortization/runs` triggers a run, optionally with a date override, a subset of
CI IDs, and a dry-run flag. The scheduler screen in the frontend starts a run and lists recent runs.

## 7. Data Schema

**Status: Implemented.** The module adds:

- `is_amortizable` on CI types;
- `amortization_behavior` on lifecycle statuses;
- financial columns on configuration items (purchase cost, salvage value, start date, useful life, current book
  value);
- `amortization_runs`, `amortization_ledger`, and `amortization_summaries` tables;
- reporting views, triggers, and check constraints.

The exact schema is documented in [design](design.md), and the enforced and intended validation rules in
[data validation](data-validation.md).

## 8. Corrections and Adjustments

**Status: Implemented.** Ledger entries are never deleted. Corrections are additive `adjustment` entries created
through `POST /amortization/adjustments` under the `amortization:configure` permission:

```json
{
  "ci_id": "uuid",
  "amount": 100.00,
  "description": "Correction for calculation error",
  "effective_date": "optional RFC3339 date"
}
```

A positive amount increases book value; a negative amount decreases it. The entry records the book value before and
after and the accumulated depreciation at the time of the adjustment.

## 9. Audit Trail

**Status: Partial.** The ledger is the financial record: every financial event is appended with amounts, balances,
entry type, user, and run reference, and it cannot be rewritten through the API.

The intended integration with the main audit log (configuration changes, write-offs, adjustments) is not in place:
amortization events are written to the structured application log, and no rows are inserted into `audit_logs` by
the amortization module. See [audit and compliance](../audit-compliance.md).

## 10. Dashboard and Reporting

**Status: Implemented.**

- **Dashboard:** accounting cards for OCC (original capitalized cost), GVB (gross book value), NBV (net book
  value), and AD (accumulated depreciation), recent ledger activity, and charts.
- **Reports:** asset schedule with OCC/GVB reference lines, monthly depreciation table, journal entries, and CSV
  export of the depreciation schedule.
- **Ledger:** filterable, paginated view with entry type and date filters.

## 11. UI Flows

| Flow | Status | Notes |
| --- | --- | --- |
| CI type amortizable checkbox | Implemented | CI type form |
| Lifecycle status behaviour dropdown | Implemented | Lifecycle status form |
| Financial fields on asset edit | Implemented | Visible only for amortizable CI types |
| Amortization dashboard, ledger, reports | Implemented | Under `/amortization` |
| Manual run screen | Implemented | Run history and trigger |
| Settings screen | Partial | Backing table missing |
| Disposal wizard with write-off confirmation | Planned | No modal flow exists |

## 12. API Endpoints

| Method | Endpoint | Description | Permission |
| --- | --- | --- | --- |
| GET | `/api/v1/amortization/configuration-items` | List amortizable CIs | `amortization:read` |
| GET | `/api/v1/amortization/configuration-items/{ciId}` | Asset amortization details | `amortization:read` |
| PUT | `/api/v1/amortization/configuration-items/{ciId}` | Update financial configuration | `amortization:configure` |
| GET | `/api/v1/amortization/ledger` | Ledger entries | `amortization:read` |
| GET | `/api/v1/amortization/ledger/{entryId}` | Single ledger entry | `amortization:read` |
| GET | `/api/v1/amortization/runs` | Run history | `amortization:read` |
| GET | `/api/v1/amortization/runs/{runId}` | Single run | `amortization:read` |
| POST | `/api/v1/amortization/runs` | Trigger manual run | `amortization:run` |
| GET | `/api/v1/amortization/summaries` | Grouped summary | `amortization:read` |
| GET | `/api/v1/amortization/reports/depreciation-schedule` | Schedule (JSON or CSV) | `amortization:read` |
| GET | `/api/v1/amortization/settings` | Global settings | `amortization:read` |
| PUT | `/api/v1/amortization/settings` | Update global settings | `amortization:configure` |
| POST | `/api/v1/amortization/adjustments` | Create adjustment | `amortization:configure` |
| POST | `/api/v1/amortization/restructuring/preview` | Preview useful-life change | `amortization:configure` |
| POST | `/api/v1/amortization/restructuring` | Execute useful-life change | `amortization:configure` |

Note that adjustment and restructuring operations are gated by `amortization:configure`, not by the separate
`amortization:adjust` permission created by the schema migration; the latter is not referenced by any route.

## 13. Implementation Status by Phase

| Phase | Scope | Status |
| --- | --- | --- |
| Foundation | Schema, calculator, ledger, handlers, CI type flag, lifecycle behaviour, CI financial fields, adjustments | Implemented |
| UX enhancements | Disposal wizard, basic monthly report | Partial (report implemented; wizard not) |
| Advanced | Dashboard widgets, CSV export, useful-life preview, filtering and drill-down | Implemented |
| Automation | Daily scheduled run and status-change write-off trigger | Partial (code present, not wired) |

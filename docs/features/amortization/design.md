# Amortization Design

This document describes the database design and design decisions of the IT asset amortization module as
implemented in the migrations and Go packages. Where an original design intention is not implemented, the section
says so and marks it Planned. For the feature summary see [amortization](../amortization.md), for requirements see
the [PRD](prd.md), and for validation rules see [data validation](data-validation.md).

## Architecture Summary

| Component | Storage | Purpose |
| --- | --- | --- |
| Amortizable CI types | `ci_type_definitions.is_amortizable` | Opt-in flag for financial tracking |
| Status-aware processing | `lifecycle_statuses.amortization_behavior` | `pending`, `active`, or `terminal` |
| Financial CI data | `configuration_items` columns | Cost basis, salvage, start date, useful life, current book value |
| Immutable ledger | `amortization_ledger` | Append-only financial event trail |
| Run tracking | `amortization_runs` | Execution records for manual runs and the expected scheduled runs |
| Reporting snapshots | `amortization_summaries` | Per-CI, per-date snapshots maintained by a trigger |
| Reporting views | `active_amortization_assets`, `amortization_report` | Current assets and financial reporting queries |

Design principles retained from the original design:

- financial records are never deleted; corrections are new ledger entries;
- every ledger entry stores book value before and after plus accumulated depreciation;
- the ledger references the run that produced it and a sequence number orders multiple entries per day;
- downstream reporting can be regenerated from the ledger.

## Schema

### CI type flag

```sql
ALTER TABLE ci_type_definitions
    ADD COLUMN is_amortizable BOOLEAN DEFAULT false NOT NULL;
```

### Lifecycle behaviour

```sql
ALTER TABLE lifecycle_statuses
    ADD COLUMN amortization_behavior VARCHAR(20) DEFAULT 'pending' NOT NULL
    CONSTRAINT valid_amortization_behavior
    CHECK (amortization_behavior IN ('pending', 'active', 'terminal'));
```

| Behaviour | Meaning |
| --- | --- |
| `pending` | Not yet in amortization |
| `active` | Depreciated by runs while book value exceeds salvage |
| `terminal` | Written off the next time the asset is processed |

### Financial columns

```sql
ALTER TABLE configuration_items
    ADD COLUMN purchase_cost DECIMAL(15,2) NULL,
    ADD COLUMN salvage_value DECIMAL(15,2) NULL,
    ADD COLUMN amort_start_date DATE NULL,
    ADD COLUMN useful_life_months INTEGER NULL,
    ADD COLUMN current_book_value DECIMAL(15,2) NULL;
```

Constraints enforce non-negative cost and salvage, salvage not exceeding purchase cost, positive useful life,
non-negative book value, and a consistency rule tying an amortization start date to a purchase cost and useful
life. There is no stored accumulated-depreciation column on `configuration_items`; the module derives it from cost
minus current book value.

### Ledger

```sql
CREATE TABLE amortization_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    entry_type VARCHAR(30) NOT NULL,
    entry_date DATE NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    book_value_before DECIMAL(15,2) NOT NULL,
    book_value_after DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    description TEXT,
    period_start_date DATE,
    period_end_date DATE,
    days_in_period INTEGER,
    adjustment_reason TEXT,
    adjustment_reference VARCHAR(100),
    corrects_entry_id UUID REFERENCES amortization_ledger(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    is_system_generated BOOLEAN DEFAULT false,
    batch_run_id UUID REFERENCES amortization_runs(id),
    amortization_run_id UUID REFERENCES amortization_runs(id),
    sequence_number INTEGER NOT NULL,
    CONSTRAINT unique_ledger_entry UNIQUE (ci_id, entry_date, sequence_number)
);
```

**Entry types.** The database constraint accepts six meaningful entry types plus one legacy value:

| Entry type | Written by current code | Meaning |
| --- | --- | --- |
| `monthly_depreciation` | Yes | Regular monthly depreciation produced by a run |
| `catch_up_depreciation` | Yes | First-processing catch-up for a start date in the past |
| `write_off` | Yes | Terminal write-off of remaining value above salvage |
| `adjustment` | Yes | Manual correction that changes book value |
| `correction` | No | Schema-reserved correction type referencing the corrected row |
| `restructuring` | Yes | Useful-life change record, amount zero, metadata in description |
| `depreciation` (legacy) | No | Accepted by the constraint for backward compatibility |

### Runs

```sql
CREATE TABLE amortization_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'started', 'running', 'completed', 'partial', 'failed', 'cancelled')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    total_cis_processed INTEGER DEFAULT 0,
    successful_depreciations INTEGER DEFAULT 0,
    write_offs_generated INTEGER DEFAULT 0,
    errors_encountered INTEGER DEFAULT 0,
    last_processed_ci_id UUID,
    last_processed_at TIMESTAMPTZ,
    checkpoint_data JSONB DEFAULT '{}',
    error_summary TEXT,
    error_details JSONB DEFAULT '{}',
    run_config JSONB DEFAULT '{}',
    is_manual BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

A partial unique index allows at most one scheduled run per date while permitting multiple manual runs:

```sql
CREATE UNIQUE INDEX unique_scheduled_run_per_date
    ON amortization_runs (run_date)
    WHERE is_manual = false;
```

The checkpoint columns exist, but the current run processor writes empty checkpoint and config documents and does
not implement resume-from-checkpoint. That stays **Planned**.

### Summaries

```sql
CREATE TABLE amortization_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    reporting_date DATE NOT NULL,
    current_book_value DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL,
    period_depreciation DECIMAL(15,2) DEFAULT 0,
    period_adjustments DECIMAL(15,2) DEFAULT 0,
    amortization_status VARCHAR(20) NOT NULL
        CHECK (amortization_status IN ('pending', 'active', 'completed', 'written_off')),
    last_updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_summary_per_ci_date UNIQUE (ci_id, reporting_date)
);
```

### Triggers and views

| Object | Behaviour |
| --- | --- |
| `update_ci_book_value_trigger` | Sets initial book value from purchase cost before a CI is written |
| `update_amortization_summary_trigger` | Upserts the summary row for a CI and ledger date on ledger insert |
| `set_ledger_sequence_trigger` | Assigns the next sequence number per CI and date |
| `active_amortization_assets` view | Amortizable CIs with a future or running schedule and positive book value, plus remaining months and monthly depreciation |
| `amortization_report` view | Latest summary per CI with year-to-date depreciation and total adjustments |

Functions `calculate_monthly_depreciation`, `calculate_write_off_amount`, and `get_next_amortization_sequence` are
also defined.

## Calculation Model

The Go calculator is straight-line only:

- monthly amount: `(purchase_cost - salvage_value) / useful_life_months`;
- catch-up: elapsed months from the start date to the processing date multiplied by the monthly amount, capped at
  the depreciable base;
- write-off: `current_book_value - salvage_value`, leaving the salvage value on the books;
- restructuring: prospective recalculation of the monthly amount over the new remaining life; historical entries
  remain unchanged.

Planned but not implemented: declining-balance and other methods, mid-month proration, multi-currency, and asset
revaluation. The schema has period columns for proration but the processor leaves them empty.

## Access Control

The migration seeds five amortization permissions and assigns them to the built-in roles:

| Permission | Admin | Editor | Viewer | Used by routes |
| --- | --- | --- | --- | --- |
| `amortization:read` | Yes | Yes | Yes | All read routes |
| `amortization:configure` | Yes | Yes | No | Configuration updates, adjustments, restructuring, settings update |
| `amortization:adjust` | Yes | Yes | No | Not referenced by a route |
| `amortization:run` | Yes | No | No | Trigger manual run |
| `amortization:admin` | Yes | No | No | Referenced by the frontend settings and scheduler routes |

## Integration Points

- **CI management:** the CI edit form renders financial fields only for amortizable types; financial changes go
  through the amortization configuration endpoint.
- **Lifecycle statuses:** behaviour is configured on status records; terminal behaviour drives write-offs when an
  asset is processed.
- **Reporting:** the reports screen reads summaries and the depreciation schedule; the dashboard shows OCC, GVB,
  NBV, and AD.
- **Audit:** the ledger is the financial audit trail. The module writes structured application log entries for
  configuration changes, write-offs, adjustments, and restructurings, but does not insert rows into `audit_logs`.

## Completed History: Accounting Metrics

The original metrics enhancement plan proposed OCC, GVB, NBV, AD, and salvage-value totals and marked itself as a
planning document. Those metrics are now implemented: they are part of the summary and schedule response models
(`total_original_cost`, `total_gross_book_value`, `total_net_book_value`, `total_accumulated_depreciation`,
`total_salvage_value`), the shared TypeScript types, and the amortization dashboard and reports screens, which
render the summary cards, OCC/GVB chart reference lines, and the monthly table. The plan is therefore historical;
the implemented behaviour is what this document and the feature documentation describe.

## Operational Notes

- The scheduler is not wired: `ScheduleDailyRun` only logs, and the processing entry points are not called by the
  server. Manual runs are the only path that mutates assets today.
- The global settings table is missing from the migrations while the repository queries it, so settings endpoints
  fail at the database layer.
- Retention and archival, partitioning, read replicas, and ERP/procurement integrations from the original design
  are **Planned** and have no implementation.

## Related Documentation

- [Amortization feature](../amortization.md)
- [PRD](prd.md)
- [Data validation](data-validation.md)
- [Data model](../../architecture/data-model.md)

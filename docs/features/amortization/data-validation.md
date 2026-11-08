# Amortization Data Validation

This document lists the rules that protect amortization data. Rules are split into three groups:

1. **Enforced by the database** - constraints and triggers created by the amortization migrations;
2. **Enforced in application code** - checks inside the calculator, service, and handlers;
3. **Intended but not enforced** - rules from the original design that have no implementation in the current
   schema or code. These are marked Planned so they are not mistaken for active protections.

Feature summary: [amortization](../amortization.md). Schema design: [design](design.md). Requirements:
[PRD](prd.md).

## Database-Level Constraints

### Configuration items (financial data)

These constraints are created by the amortization migration and are active:

| Constraint | Rule |
| --- | --- |
| `valid_purchase_cost` | `purchase_cost IS NULL OR purchase_cost >= 0` |
| `valid_salvage_value` | `salvage_value IS NULL OR salvage_value >= 0` |
| `valid_salvage_vs_purchase` | Salvage may not exceed purchase cost when both are present |
| `valid_useful_life_months` | `useful_life_months IS NULL OR useful_life_months > 0` |
| `valid_book_value` | `current_book_value IS NULL OR current_book_value >= 0` |
| `amortization_dates_consistency` | An amortization start date requires both purchase cost and useful life |

### Amortization ledger

| Constraint | Rule |
| --- | --- |
| `valid_entry_type` | Entry type must be one of `depreciation`, `monthly_depreciation`, `catch_up_depreciation`, `write_off`, `adjustment`, `correction`, `restructuring` |
| `unique_ledger_entry` | `(ci_id, entry_date, sequence_number)` is unique |
| `set_ledger_sequence_trigger` | Sets `sequence_number` per CI and date on insert |
| Required columns | `entry_type`, `entry_date`, `amount`, `book_value_before`, `book_value_after` are `NOT NULL` |

The ledger is append-only by convention and by API surface: no update or delete route exists for ledger entries.
The database itself would allow updates; immutability is an application guarantee.

### Amortization runs

| Constraint | Rule |
| --- | --- |
| `valid_run_status` | Status must be one of `pending`, `started`, `running`, `completed`, `partial`, `failed`, `cancelled` |
| Unique scheduled run per date | Partial unique index on `run_date WHERE is_manual = false`; multiple manual runs per day are allowed |

### Amortization summaries

| Constraint | Rule |
| --- | --- |
| `valid_summary_status` | Status must be one of `pending`, `active`, `completed`, `written_off` |
| `unique_summary_per_ci_date` | `(ci_id, reporting_date)` is unique; the trigger upserts on conflict |

The summary row is maintained by the `update_amortization_summary_trigger` after each ledger insert. The
accumulated figure in that trigger only sums entries whose type is `depreciation`, so summaries are approximate
for assets that use `monthly_depreciation` or `catch_up_depreciation`; reports that matter read the ledger.

### Lifecycle statuses

`valid_amortization_behavior` restricts the value to `pending`, `active`, or `terminal`. The column is `NOT NULL`
with a default of `pending`.

### Intended constraints that do not exist

The following constraints appear in earlier design material but are **not** present in the migrations, so they
must not be relied on:

| Intended rule | Current state |
| --- | --- |
| Ledger amount non-negative | Not enforced; `amount` is `NOT NULL` but has no sign check |
| Book value progression (`after <= before` or a signed rule) | Not enforced; adjustments can increase book value |
| Period date consistency (`book_value_before` / `book_value_after` / days in period) | Not enforced |
| Run metric consistency (`successful + write-offs <= total`) | Not enforced |
| Plain unique run per date | Replaced by the partial unique index described above |
| Automatic correction of book value mismatches | No such function exists |

## Application-Level Validation

### Financial configuration

- Purchase cost must be non-negative and no lower than the salvage value; useful life must be positive. The
  database checks enforce the bounds, so an invalid update fails and the service returns an error.
- Changing purchase cost resets the current book value to the new cost and zeroes accumulated depreciation; the
  update is wrapped in a transaction.
- Setting an amortization start date without cost and useful life is rejected by the database consistency
  constraint.

### Monthly depreciation

- Useful life of zero or less is rejected by the calculator.
- Monthly amount is `(purchase_cost - salvage_value) / useful_life_months` with two-decimal rounding.
- If the calculated amount would push the book value below salvage, the amount is reduced so the book value ends
  exactly at salvage.

### Processing eligibility

An asset is processed by a run only when all of these hold:

- it has an amortization start date on or before the processing date;
- its current book value is above its salvage value;
- its lifecycle behaviour is not `pending`.

For a `terminal` behaviour, a run writes the asset off. For `active`, a run depreciates it. Assets with no
lifecycle status fall back to normal processing.

### Catch-up and regular entries

- The first processing of an asset with no ledger history writes a `catch_up_depreciation` entry covering the
  elapsed period.
- Subsequent runs write a `monthly_depreciation` entry.
- There is no per-month idempotency check in the processor: a second manual run for the same asset can create a
  second monthly entry. The only scheduling guard is the unique index for scheduled runs. Operators should treat
  duplicate manual runs as an operational risk.

### Write-offs

- A write-off requires the asset to be amortizable, to have a book value above salvage, and to be at a terminal
  behaviour.
- The write-off amount is `current_book_value - salvage_value`; the resulting book value is the salvage value.
- A write-off with no remaining value above salvage is rejected with "no remaining book value to write off".
- Write-off entries record the before and after book value, the accumulated depreciation, and the reason.

### Adjustments

- Adjustments are additive `adjustment` entries; a positive amount increases book value and a negative amount
  decreases it.
- The handler does not validate the sign, the description, or the resulting book value. A large negative
  adjustment can drive the book value below zero; the database then rejects the book value update, which the
  service logs as a warning after the ledger entry has been written. This is a known gap.

### Restructuring

- A new useful life must be positive and a reason is required; both are checked in the handler.
- The calculator validates the preview and returns a non-success result with a message when the change is not
  valid; the service does not then write the restructuring entry.
- Historical ledger entries are untouched; the restructuring entry has amount zero and records the old and new
  useful life, the old and new monthly amounts, the change, and the reason.

### Lifecycle transitions

- Inventory CIs do not have a transition state machine; any lifecycle status can be assigned, and the amortization
  behaviour is read when a run processes the asset.
- EA entities do have a state machine: Proposed to Active or Deprecated, Active to Deprecated or Retired,
  Deprecated to Retired, and Retired terminal. Invalid transitions are rejected by the EA service.
- The automatic terminal write-off on a CI status change is not wired, so a status change does not itself create a
  write-off. Wait for a run, or create an adjustment or manual write-off path.

## Integrity Checks and Error Handling

The ledger is the financial source of truth; the CI book value is maintained alongside it. To audit consistency:

```sql
-- CIs whose book value disagrees with their latest ledger entry
SELECT ci.id, ci.name, ci.current_book_value, latest.book_value_after
FROM configuration_items ci
LEFT JOIN LATERAL (
    SELECT book_value_after
    FROM amortization_ledger
    WHERE ci_id = ci.id
    ORDER BY entry_date DESC, sequence_number DESC
    LIMIT 1
) latest ON true
WHERE ci.current_book_value IS DISTINCT FROM latest.book_value_after;
```

Validation failures surface as API errors (400 for malformed input, 422 or 500 for domain errors, depending on
the handler) or as warnings in the structured application log when a follow-up update fails. Critical failures stop
the operation; non-critical inconsistencies are logged and processing continues.

## Related Documentation

- [Amortization feature](../amortization.md)
- [Design](design.md)
- [PRD](prd.md)
- [Amortization audit trail](../../features/audit-compliance.md)

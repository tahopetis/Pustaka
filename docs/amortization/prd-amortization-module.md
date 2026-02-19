---

# PRD: IT Asset Amortization & Lifecycle Financials

## 1. Executive Summary

**Objective:** To implement an automated, status-aware financial engine within the IT Asset Management system. This feature tracks the depreciation of assets over time, handles complex lifecycle events (repairs, disposal, theft), and provides accurate monthly financial reporting.

**Key Value:**

* **Automation:** Eliminates manual spreadsheet tracking.
* **Accuracy:** Handles mid-month proration and exact day calculations.
* **Compliance:** Ensures assets are written off correctly when disposed or stolen.

---

## 2. Configuration & System Setup

### 2.1 Global Environment Variables

System-wide defaults configurable via `.env`:

| Variable | Default | Description |
| --- | --- | --- |
| `AMORTIZATION_DEFAULT_SALVAGE` | `1` | Default residual value (e.g., $1) to keep asset on books until disposal. |
| `AMORTIZATION_CURRENCY` | `USD` | Currency symbol for display (single currency system). |
| `AMORTIZATION_TIMEZONE` | `UTC` | Reference timezone for scheduler execution. |

### 2.2 CI Type Configuration

The system must allow administrators to define *which* assets are financial in nature.

* **UI Requirement:** Add a checkbox `Is Amortizable?` to the **CI Type** creation/edit form.
* **Logic:**
  * If `True`: The "Financials" tab is visible in the CI Edit view.
  * If `False`: Financial fields are hidden (e.g., for Keyboards, Cables).

### 2.3 Lifecycle Status Configuration

Each lifecycle status must define its amortization behavior.

* **UI Requirement:** Add a dropdown `Amortization Behavior` to the **Lifecycle Status** creation/edit form.
* **Options:**
  * `Pending` - No depreciation occurs (e.g., Planned, On Order, Pending Install)
  * `Active` - Standard monthly depreciation (e.g., In Stock, Operational, In Maintenance)
  * `Terminal` - Immediate write-off to $0 (e.g., Disposed, Retired, Missing/Stolen)

---

## 3. Asset Data Management

### 3.1 Creation Phase (The "Blind" Phase)

* **Constraint:** Financial/Amortization fields are **HIDDEN** during the initial "Create Asset" wizard.
* **Rationale:** Speed up operational onboarding; financial data is typically added post-procurement.

### 3.2 Edit Phase (The "Financials" Tab)

Users configure amortization only when **Editing** an existing CI.

* **Purchase Cost:** Input field (Decimal). The original cost basis.
* **Salvage Value:** Pre-filled with `AMORTIZATION_DEFAULT_SALVAGE` ($1). Editable.
* **Amortization Start Date:** Date asset was placed in service.
* **Useful Life (Duration):** Dropdown mapping to months:
  * 1 Year (12 Months)
  * 3 Years (36 Months)
  * 5 Years (60 Months)
  * Custom (User inputs integer months)

* **End Date:** Read-only. Calculated automatically: `Start Date + Useful Life`.
* **Current Book Value:** Read-only. Calculated from ledger entries.

---

## 4. Lifecycle Status Logic

The amortization engine is strictly coupled to the Asset Lifecycle Status via the `amortization_behavior` field.

| Amortization Behavior | Example Statuses | Amortization Action |
| --- | --- | --- |
| `pending` | Planned, On Order, Pending Install | No depreciation occurs. |
| `active` | In Stock, Operational, In Maintenance, Defective / Repair | Standard monthly depreciation. |
| `terminal` | Retired, Disposed, Missing / Stolen | Immediate write-off to **$0.00**. |

### 4.1 The Disposal Logic (Write-Off)

If an asset is moved to a status with `amortization_behavior = 'terminal'`:

1. **Salvage Ignored:** The system disregards the salvage value.
2. **Immediate Action:** The remaining book value is written off entirely.
3. **Result:** Book Value becomes **$0.00**.
4. **Audit Trail:** Entry created in both `amortization_ledger` and `audit_logs`.

---

## 5. Calculation Formulas

### 5.1 Standard Monthly Depreciation (Straight-Line)

Used for full months where the asset is active from Day 1 to Day 30/31.

```
monthly_depreciation = (purchase_cost - salvage_value) / useful_life_months
```

### 5.2 Mid-Month Proration (Onboarding)

Used for the first month if `Start Date` is not the 1st.

```
days_in_month = total days in the start month
days_active = days_in_month - start_day + 1
prorated_amount = monthly_depreciation * (days_active / days_in_month)
```

### 5.3 Prospective Recalculation (Useful Life Change)

When useful life is modified on an asset with existing depreciation:

```
new_monthly = (current_book_value - salvage_value) / remaining_months
```

* Historical ledger entries remain unchanged.
* New depreciation rate applies from current period forward.

### 5.4 Write-Off Event

Used when Status changes to a terminal status.

```
write_off_amount = current_book_value
closing_balance = 0
```

*(The entire remaining balance is recorded as a loss in the ledger).*

---

## 6. Automation & Triggers

### 6.1 The Scheduler (Cron Job)

* **Schedule:** Daily at 00:00:00 (Server Time / Configured Timezone).
* **Logic:** Check if previous month has been processed. If not, run depreciation.
* **Target:** All CIs where:
  * CI type has `is_amortizable = true`
  * Lifecycle status has `amortization_behavior = 'active'`
  * `current_book_value > salvage_value`
* **Action:** Calculates expense, writes to Ledger, updates CI Book Value.
* **Checkpoint:** Uses `amortization_runs` table to track processed periods (prevents duplicate runs, handles server downtime).

### 6.2 Status Change Trigger

When a CI's lifecycle status changes:

1. Check if new status has `amortization_behavior = 'terminal'`
2. If yes AND `current_book_value > 0`, trigger write-off flow

---

## 7. Data Schema (PostgreSQL)

### 7.1 Schema Changes

```sql
-- 1. CI Type Definitions (add amortization flag)
ALTER TABLE ci_type_definitions
ADD COLUMN is_amortizable BOOLEAN DEFAULT FALSE;

-- 2. Lifecycle Statuses (add amortization behavior)
ALTER TABLE lifecycle_statuses
ADD COLUMN amortization_behavior VARCHAR(20)
CHECK (amortization_behavior IN ('pending', 'active', 'terminal'))
DEFAULT 'active';

-- 3. Configuration Items (add financial columns - dedicated, not JSONB)
ALTER TABLE configuration_items ADD COLUMN purchase_cost DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN salvage_value DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN amort_start_date DATE;
ALTER TABLE configuration_items ADD COLUMN useful_life_months INT;
ALTER TABLE configuration_items ADD COLUMN current_book_value DECIMAL(19,4);

-- Index for amortizable assets
CREATE INDEX idx_cis_book_value ON configuration_items(current_book_value)
WHERE current_book_value IS NOT NULL;
```

### 7.2 New Tables

```sql
-- Amortization Ledger (History/Reporting)
CREATE TABLE amortization_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    period_date DATE NOT NULL,  -- Last day of month (e.g., '2024-01-31')

    opening_balance DECIMAL(19,4) NOT NULL,
    amount DECIMAL(19,4) NOT NULL,
    closing_balance DECIMAL(19,4) NOT NULL,

    transaction_type VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('DEPRECIATION', 'WRITE_OFF', 'ADJUSTMENT')),
    notes TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    CONSTRAINT unique_ci_period_type UNIQUE (ci_id, period_date, transaction_type)
);

CREATE INDEX idx_ledger_ci ON amortization_ledger(ci_id);
CREATE INDEX idx_ledger_period ON amortization_ledger(period_date);
CREATE INDEX idx_ledger_type ON amortization_ledger(transaction_type);

-- Amortization Runs (Scheduler Checkpoint)
CREATE TABLE amortization_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_month DATE NOT NULL UNIQUE,  -- First day of month (e.g., '2024-01-01')
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    assets_processed INT DEFAULT 0,
    total_depreciation DECIMAL(19,4) DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## 8. Corrections & Adjustments

### 8.1 Adjustment Entries

To correct erroneous depreciation:

* **Never delete** ledger entries (audit trail integrity).
* Create `ADJUSTMENT` type entries to correct errors.
* Net effect corrects the book value.

**API Endpoint:** `POST /api/v1/amortization/adjustment`

**Request:**
```json
{
  "ci_id": "uuid",
  "amount": 100.00,
  "reason": "Correction for calculation error in Jan 2024",
  "reference_period": "2024-01"
}
```

* Positive amount = increase book value
* Negative amount = decrease book value
* **Permission Required:** `amortization:adjust` (admin only)

---

## 9. Audit Trail Integration

### 9.1 Amortization Ledger

Financial record for reporting. Contains: amounts, balances, transaction types.

Used for: monthly reports, asset valuation, compliance.

### 9.2 Main Audit Log

The following events are logged to the main `audit_logs` table:

| Action | Trigger |
| --- | --- |
| `AMORTIZATION_CONFIG_UPDATED` | Financial fields set/changed on CI |
| `AMORTIZATION_WRITEOFF` | Terminal status triggers write-off |
| `AMORTIZATION_ADJUSTMENT` | Manual correction created |

**Note:** Automated monthly depreciation does NOT create audit log entries (too noisy; ledger is the record).

---

## 10. Dashboard & Reporting Requirements

### 10.1 Phase 2: Basic Report

**Monthly Expense View:** Table with monthly breakdown.

* Sum of `amount` grouped by Month
* Columns: Month, Depreciation Total, Write-off Total, Adjustment Total, Asset Count

### 10.2 Phase 3: Advanced Dashboards

1. **Monthly Expense Chart:** Bar chart visualization
2. **Operational vs. Loss:**
   * Stack A: Type `DEPRECIATION` (Standard Cost)
   * Stack B: Type `WRITE_OFF` (Unexpected Loss)
3. **Asset Class Breakdown:** Depreciation grouped by CI type
4. **Upcoming Depletions:** Assets nearing $0 book value

---

## 11. UI/UX Flows

### 11.1 CI Type Form

* Add checkbox: "Is Amortizable?"
* When checked, CIs of this type will show Financials tab

### 11.2 Lifecycle Status Form

* Add dropdown: "Amortization Behavior"
* Options: Pending, Active, Terminal
* Required field with default "Active"

### 11.3 Asset Edit Form (Financials Tab)

* Visible only if CI type `is_amortizable = true`
* Fields: Purchase Cost, Salvage Value, Start Date, Useful Life
* Read-only displays: Current Book Value, End Date

### 11.4 Disposal Wizard (Phase 2)

When User changes status to a `terminal` status:

1. **Check:** Is `current_book_value > 0`?
2. **If Yes:** Trigger Modal.

> **Warning: Confirm Write-off**
>
> This asset has a remaining value of **$450.00**.
>
> Changing status to **Disposed** will immediately write this off to **$0.00**.
>
> This action will be logged and cannot be undone.
>
> **Reason:** [ Text Input - Required ]
>
> [ Cancel ] [ Confirm & Write Off ]

---

## 12. API Endpoints

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/api/v1/ci/{id}/amortization` | Get CI amortization details & schedule | `ci:read` |
| PUT | `/api/v1/ci/{id}/amortization` | Update CI financial fields | `ci:update` |
| GET | `/api/v1/ci/{id}/amortization/ledger` | Get ledger entries for CI | `ci:read` |
| POST | `/api/v1/amortization/adjustment` | Create manual adjustment | `amortization:adjust` |
| GET | `/api/v1/amortization/report/monthly` | Monthly expense summary | `audit:read` |

---

## 13. Phased Implementation

### Phase 1: Foundation (MVP)

* Database schema changes (migrations)
* Backend service for depreciation calculation
* Scheduler with checkpoint mechanism
* CI Type: `is_amortizable` checkbox
* Lifecycle Status: `amortization_behavior` dropdown
* CI Edit: Financials tab
* API endpoints for CRUD and adjustments

### Phase 2: UX Enhancements

* Disposal wizard with confirmation modal
* Basic monthly expense report (table view)

### Phase 3: Advanced Features

* Dashboard widgets and visualizations
* Advanced reporting with export (CSV/Excel)
* Useful life change preview UI
* Filter and drill-down capabilities

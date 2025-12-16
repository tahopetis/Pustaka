# Implementation Plan: IT Asset Amortization Module

## Overview

This document outlines the implementation plan for the IT Asset Amortization & Lifecycle Financials feature in Pustaka CMDB.

---

## Phase 1: Foundation (MVP)

### 1.1 Database Schema Changes

**Migration: Add amortization_behavior to lifecycle_statuses**

```sql
ALTER TABLE lifecycle_statuses
ADD COLUMN amortization_behavior VARCHAR(20)
CHECK (amortization_behavior IN ('pending', 'active', 'terminal'))
DEFAULT 'active';
```

Update existing statuses with appropriate values.

**Migration: Add is_amortizable to ci_type_definitions**

```sql
ALTER TABLE ci_type_definitions
ADD COLUMN is_amortizable BOOLEAN DEFAULT FALSE;
```

**Migration: Add financial columns to configuration_items**

```sql
ALTER TABLE configuration_items ADD COLUMN purchase_cost DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN salvage_value DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN amort_start_date DATE;
ALTER TABLE configuration_items ADD COLUMN useful_life_months INT;
ALTER TABLE configuration_items ADD COLUMN current_book_value DECIMAL(19,4);

CREATE INDEX idx_cis_amortizable ON configuration_items(current_book_value)
WHERE current_book_value IS NOT NULL;
```

**Migration: Create amortization_ledger table**

```sql
CREATE TABLE amortization_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    period_date DATE NOT NULL,
    opening_balance DECIMAL(19,4) NOT NULL,
    amount DECIMAL(19,4) NOT NULL,
    closing_balance DECIMAL(19,4) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('DEPRECIATION', 'WRITE_OFF', 'ADJUSTMENT')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    CONSTRAINT unique_ci_period UNIQUE (ci_id, period_date, transaction_type)
);

CREATE INDEX idx_ledger_ci ON amortization_ledger(ci_id);
CREATE INDEX idx_ledger_period ON amortization_ledger(period_date);
CREATE INDEX idx_ledger_type ON amortization_ledger(transaction_type);
```

**Migration: Create amortization_runs table (checkpoint)**

```sql
CREATE TABLE amortization_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_month DATE NOT NULL UNIQUE,  -- First day of month (e.g., 2024-01-01)
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    assets_processed INT DEFAULT 0,
    total_depreciation DECIMAL(19,4) DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 1.2 Environment Configuration

Add to `.env.example`:

```env
# Amortization Settings
AMORTIZATION_DEFAULT_SALVAGE=1
AMORTIZATION_CURRENCY=USD
AMORTIZATION_TIMEZONE=UTC
```

### 1.3 Backend Implementation

**Files to create:**

| File | Purpose |
|------|---------|
| `internal/amortization/models.go` | Data models and DTOs |
| `internal/amortization/repository.go` | Database operations |
| `internal/amortization/service.go` | Business logic & calculations |
| `internal/amortization/scheduler.go` | Cron job for monthly depreciation |
| `internal/api/handlers/amortization_handlers.go` | HTTP handlers |

**Core Service Methods:**

```go
// service.go
type AmortizationService interface {
    // Calculate depreciation for a single CI
    CalculateDepreciation(ci *ConfigurationItem) (*DepreciationResult, error)

    // Process monthly depreciation for all eligible CIs
    RunMonthlyDepreciation(period time.Time) (*RunResult, error)

    // Handle status change to terminal (write-off)
    ProcessWriteOff(ciID uuid.UUID, reason string, userID uuid.UUID) error

    // Create manual adjustment entry
    CreateAdjustment(req *AdjustmentRequest) error

    // Get amortization schedule for a CI
    GetSchedule(ciID uuid.UUID) (*AmortizationSchedule, error)

    // Get ledger entries for a CI
    GetLedger(ciID uuid.UUID, filters *LedgerFilters) (*LedgerResponse, error)
}
```

**Calculation Logic:**

```go
// Monthly depreciation (straight-line)
func calculateMonthlyDepreciation(cost, salvage float64, usefulLifeMonths int) float64 {
    return (cost - salvage) / float64(usefulLifeMonths)
}

// Prorated first month
func calculateProratedDepreciation(monthlyAmount float64, startDate time.Time) float64 {
    daysInMonth := daysInMonth(startDate)
    daysActive := daysInMonth - startDate.Day() + 1
    return monthlyAmount * (float64(daysActive) / float64(daysInMonth))
}

// Prospective recalculation (when useful life changes)
func recalculateProspective(currentBookValue, salvage float64, remainingMonths int) float64 {
    if remainingMonths <= 0 {
        return 0
    }
    return (currentBookValue - salvage) / float64(remainingMonths)
}
```

**Scheduler Implementation:**

```go
// scheduler.go
// Runs daily at 00:00:00, checks if month-end processing is needed
func (s *Scheduler) Start() {
    c := cron.New()
    c.AddFunc("0 0 * * *", s.checkAndProcess)  // Daily at midnight
    c.Start()
}

func (s *Scheduler) checkAndProcess() {
    now := time.Now().In(s.timezone)

    // Check if we need to process the previous month
    lastRun, _ := s.repo.GetLastCompletedRun()
    targetMonth := getLastDayOfPreviousMonth(now)

    if lastRun == nil || lastRun.PeriodMonth.Before(targetMonth) {
        s.service.RunMonthlyDepreciation(targetMonth)
    }
}
```

### 1.4 API Endpoints

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/api/v1/ci/{id}/amortization` | Get CI amortization details & schedule | `ci:read` |
| PUT | `/api/v1/ci/{id}/amortization` | Update CI financial fields | `ci:update` |
| GET | `/api/v1/ci/{id}/amortization/ledger` | Get ledger entries for CI | `ci:read` |
| POST | `/api/v1/amortization/adjustment` | Create manual adjustment | `amortization:adjust` |
| GET | `/api/v1/amortization/report/monthly` | Monthly expense summary | `audit:read` |

### 1.5 Frontend Changes

**CI Type Form (`web/src/views/ci-types/`):**
- Add "Is Amortizable" checkbox

**Lifecycle Status Form (`web/src/views/lifecycle-statuses/`):**
- Add "Amortization Behavior" dropdown (Pending/Active/Terminal)

**CI Edit Form (`web/src/views/ci/`):**
- Add "Financials" tab (visible only if CI type is amortizable)
- Fields: Purchase Cost, Salvage Value, Start Date, Useful Life dropdown
- Display: Current Book Value (read-only), End Date (calculated)

### 1.6 Status Change Hook

Modify CI update logic to detect terminal status changes:

```go
// In CI service update method
func (s *CIService) Update(id uuid.UUID, req *UpdateCIRequest, userID uuid.UUID) (*CI, error) {
    oldCI, _ := s.repo.GetByID(id)

    // ... existing update logic ...

    // Check for terminal status change
    if req.LifecycleStatusID != nil {
        newStatus, _ := s.lifecycleRepo.GetByID(*req.LifecycleStatusID)
        if newStatus.AmortizationBehavior == "terminal" && oldCI.CurrentBookValue > 0 {
            // Trigger write-off
            s.amortService.ProcessWriteOff(id, "Status changed to "+newStatus.Name, userID)
        }
    }

    return updatedCI, nil
}
```

---

## Phase 2: UX Enhancements

### 2.1 Disposal Wizard

When user changes CI status to a terminal status and `current_book_value > 0`:

1. Show confirmation modal:
   - Display current book value
   - Warning about write-off
   - Require reason input
   - Confirm/Cancel buttons

2. On confirm:
   - Update status
   - Create WRITE_OFF ledger entry
   - Create audit log entry
   - Set book value to $0

### 2.2 Basic Monthly Report

**Endpoint:** `GET /api/v1/amortization/report/monthly?year=2024`

**Response:**
```json
{
  "currency": "USD",
  "year": 2024,
  "months": [
    {
      "month": "2024-01",
      "depreciation_total": 15000.00,
      "write_off_total": 500.00,
      "adjustment_total": 0.00,
      "assets_count": 45
    }
  ],
  "year_total": {
    "depreciation": 180000.00,
    "write_off": 2500.00,
    "adjustment": -100.00
  }
}
```

**Frontend:** Simple table view with monthly breakdown.

---

## Phase 3: Advanced Features

### 3.1 Dashboard Widgets

- Monthly expense chart (bar chart)
- Depreciation vs Write-off comparison
- Asset class breakdown (by CI type)
- Upcoming fully-depreciated assets

### 3.2 Advanced Reporting

- Export to CSV/Excel
- Filter by CI type, date range, transaction type
- Drill-down from summary to individual assets

### 3.3 Useful Life Change Handling

When useful life is modified on a CI with existing depreciation:
1. Calculate remaining months
2. Recalculate new monthly amount (prospective)
3. Show preview to user before saving
4. Store recalculation event in ledger notes

---

## File Summary

### Backend (Go)

| File | Action |
|------|--------|
| `cmd/migrations/XXX_amortization.sql` | Create |
| `internal/amortization/models.go` | Create |
| `internal/amortization/repository.go` | Create |
| `internal/amortization/service.go` | Create |
| `internal/amortization/scheduler.go` | Create |
| `internal/api/handlers/amortization_handlers.go` | Create |
| `internal/ci/models.go` | Modify (add financial fields to CI struct) |
| `internal/ci/lifecycle_status.go` | Modify (add AmortizationBehavior field) |
| `internal/ci/service.go` | Modify (hook for terminal status detection) |
| `internal/config/config.go` | Modify (add amortization config) |
| `cmd/api/main.go` | Modify (register routes, start scheduler) |

### Frontend (Vue)

| File | Action |
|------|--------|
| `web/src/views/ci-types/CITypeForm.vue` | Modify (add is_amortizable checkbox) |
| `web/src/views/lifecycle-statuses/LifecycleStatusForm.vue` | Modify (add amortization_behavior dropdown) |
| `web/src/views/ci/CIEdit.vue` | Modify (add Financials tab) |
| `web/src/components/ci/FinancialsTab.vue` | Create |
| `web/src/components/ci/WriteOffModal.vue` | Create (Phase 2) |
| `web/src/services/api.ts` | Modify (add amortization endpoints) |
| `web/src/types/ci.ts` | Modify (add financial field types) |

---

## Permissions

Add new permission:

```sql
INSERT INTO permissions (name, description, resource_type) VALUES
('amortization:adjust', 'Create amortization adjustments', 'amortization');

-- Assign to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name = 'amortization:adjust';
```

---

## Testing Strategy

1. **Unit Tests:**
   - Depreciation calculation (standard, prorated, prospective)
   - Write-off logic
   - Adjustment creation

2. **Integration Tests:**
   - Monthly run with mixed scenarios
   - Status change triggering write-off
   - Ledger entry creation

3. **E2E Tests:**
   - Create amortizable CI type
   - Add CI with financial data
   - Verify depreciation schedule
   - Change status to terminal, verify write-off

---

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `AMORTIZATION_DEFAULT_SALVAGE` | `1` | Default salvage value for new assets |
| `AMORTIZATION_CURRENCY` | `USD` | Currency symbol for display |
| `AMORTIZATION_TIMEZONE` | `UTC` | Timezone for scheduler execution |

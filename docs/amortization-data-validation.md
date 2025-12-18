# Amortization Module Data Validation Rules

## Overview

This document outlines comprehensive data validation rules, constraints, and business logic for the IT Asset Amortization Module. These rules ensure data integrity, financial accuracy, and regulatory compliance.

## Database-Level Constraints

### 1. Configuration Items Financial Constraints

#### Cost Validations
```sql
-- Non-negative costs
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_purchase_cost
    CHECK (purchase_cost IS NULL OR purchase_cost >= 0),

    ADD CONSTRAINT valid_salvage_value
    CHECK (salvage_value IS NULL OR salvage_value >= 0),

    ADD CONSTRAINT valid_book_value
    CHECK (current_book_value IS NULL OR current_book_value >= 0);
```

#### Logical Cost Relationships
```sql
-- Salvage value cannot exceed purchase cost
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_salvage_vs_purchase
    CHECK (
        purchase_cost IS NULL OR
        salvage_value IS NULL OR
        salvage_value <= purchase_cost
    );

-- Useful life must be positive if specified
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_useful_life_months
    CHECK (useful_life_months IS NULL OR useful_life_months > 0);
```

#### Date Consistency
```sql
-- Amortization dates must be logical
ALTER TABLE configuration_items
    ADD CONSTRAINT amortization_dates_consistency
    CHECK (
        amort_start_date IS NULL OR
        useful_life_months IS NULL OR
        (purchase_cost IS NOT NULL AND useful_life_months IS NOT NULL)
    );
```

### 2. Amortization Ledger Constraints

#### Entry Type Validation
```sql
ALTER TABLE amortization_ledger
    ADD CONSTRAINT valid_entry_type
    CHECK (entry_type IN ('depreciation', 'write_off', 'adjustment', 'correction'));
```

#### Amount Validations
```sql
ALTER TABLE amortization_ledger
    ADD CONSTRAINT valid_ledger_amount
    CHECK (amount >= 0),

    ADD CONSTRAINT valid_book_value_progression
    CHECK (book_value_after <= book_value_before);
```

#### Period Validations
```sql
ALTER TABLE amortization_ledger
    ADD CONSTRAINT valid_period_dates
    CHECK (
        (period_start_date IS NULL AND period_end_date IS NULL AND days_in_period IS NULL) OR
        (period_start_date IS NOT NULL AND period_end_date IS NOT NULL AND days_in_period IS NOT NULL AND
         period_end_date >= period_start_date AND days_in_period > 0)
    );
```

#### Sequence Integrity
```sql
ALTER TABLE amortization_ledger
    ADD CONSTRAINT unique_ledger_entry
    UNIQUE (ci_id, entry_date, sequence_number);
```

### 3. Lifecycle Status Constraints

#### Amortization Behavior Validation
```sql
ALTER TABLE lifecycle_statuses
    ADD CONSTRAINT valid_amortization_behavior
    CHECK (amortization_behavior IN ('pending', 'active', 'terminal'));
```

### 4. Amortization Runs Constraints

#### Status Validation
```sql
ALTER TABLE amortization_runs
    ADD CONSTRAINT valid_run_status
    CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'));
```

#### Run Uniqueness
```sql
ALTER TABLE amortization_runs
    ADD CONSTRAINT unique_run_per_date
    UNIQUE (run_date);
```

#### Metric Validations
```sql
ALTER TABLE amortization_runs
    ADD CONSTRAINT valid_run_metrics
    CHECK (
        total_cis_processed >= 0 AND
        successful_depreciations >= 0 AND
        write_offs_generated >= 0 AND
        errors_encountered >= 0 AND
        successful_depreciations + write_offs_generated <= total_cis_processed
    );
```

## Application-Level Validation Rules

### 1. Asset Configuration Validation

#### Amortization Eligibility
```go
func ValidateAmortizationEligibility(ci *ConfigurationItem) error {
    // Check if CI type is amortizable
    if !ci.CIType.IsAmortizable {
        return errors.New("CI type is not amortizable")
    }

    // Validate required financial fields
    if ci.PurchaseCost == nil || ci.PurchaseCost.LessThanOrEqual(decimal.Zero) {
        return errors.New("purchase cost must be positive")
    }

    if ci.UsefulLifeMonths == nil || *ci.UsefulLifeMonths <= 0 {
        return errors.New("useful life must be positive")
    }

    // Validate salvage value
    if ci.SalvageValue != nil && ci.SalvageValue.GreaterThan(*ci.PurchaseCost) {
        return errors.New("salvage value cannot exceed purchase cost")
    }

    return nil
}
```

#### Amortization Start Date Validation
```go
func ValidateAmortizationStartDate(startDate time.Time, ci *ConfigurationItem) error {
    // Cannot start before asset creation
    if startDate.Before(ci.CreatedAt.Time) {
        return errors.New("amortization start date cannot precede asset creation")
    }

    // Cannot start in the future (unless explicitly allowed)
    if startDate.After(time.Now().AddDate(0, 0, 7)) { // 7-day grace period
        return errors.New("amortization start date too far in the future")
    }

    return nil
}
```

### 2. Financial Calculation Validation

#### Monthly Depreciation Calculation
```go
func CalculateMonthlyDepreciation(purchaseCost, salvageValue decimal.Decimal, usefulLifeMonths int) (decimal.Decimal, error) {
    if usefulLifeMonths <= 0 {
        return decimal.Zero, errors.New("useful life must be positive")
    }

    if purchaseCost.LessThanOrEqual(decimal.Zero) {
        return decimal.Zero, errors.New("purchase cost must be positive")
    }

    depreciableAmount := purchaseCost.Sub(salvageValue)
    if depreciableAmount.LessThan(decimal.Zero) {
        return decimal.Zero, errors.New("salvage value cannot exceed purchase cost")
    }

    monthlyDepreciation := depreciableAmount.Div(decimal.NewFromInt(int64(usefulLifeMonths)))

    // Round to 2 decimal places for currency
    return monthlyDepreciation.Round(2), nil
}
```

#### Book Value Validation
```go
func ValidateBookValueTransition(before, after decimal.Decimal, amount decimal.Decimal) error {
    // Check arithmetic consistency
    expectedAfter := before.Sub(amount)
    if !expectedAfter.Equal(after) {
        return errors.New("book value after transaction is inconsistent")
    }

    // Book value cannot go negative
    if after.LessThan(decimal.Zero) {
        return errors.New("book value cannot be negative")
    }

    // Amount must be positive
    if amount.LessThanOrEqual(decimal.Zero) {
        return errors.New("transaction amount must be positive")
    }

    return nil
}
```

### 3. Ledger Entry Validation

#### Depreciation Entry Rules
```go
func ValidateDepreciationEntry(entry *AmortizationLedgerEntry) error {
    if entry.EntryType != "depreciation" {
        return nil // Not a depreciation entry
    }

    // Must have period information
    if entry.PeriodStartDate == nil || entry.PeriodEndDate == nil {
        return errors.New("depreciation entries must have period dates")
    }

    if entry.DaysInPeriod == nil || *entry.DaysInPeriod <= 0 {
        return errors.New("depreciation entries must have valid period days")
    }

    // Period dates must be within the same month
    if !isSameMonth(*entry.PeriodStartDate, *entry.PeriodEndDate) {
        return errors.New("depreciation period must be within a single month")
    }

    // Entry date should be end of period
    if !entry.EntryDate.Equal(*entry.PeriodEndDate) {
        return errors.New("depreciation entry date must match period end date")
    }

    return nil
}
```

#### Write-Off Entry Rules
```go
func ValidateWriteOffEntry(entry *AmortizationLedgerEntry, ci *ConfigurationItem) error {
    if entry.EntryType != "write_off" {
        return nil // Not a write-off entry
    }

    // Write-off amount should bring book value to zero or salvage value
    expectedWriteOff := entry.BookValueBefore
    if ci.SalvageValue != nil {
        expectedWriteOff = expectedWriteOff.Sub(*ci.SalvageValue)
    }

    if !entry.Amount.Equal(expectedWriteOff) {
        return errors.New("write-off amount is incorrect")
    }

    // Final book value should be zero or salvage value
    finalValue := decimal.Zero
    if ci.SalvageValue != nil {
        finalValue = *ci.SalvageValue
    }

    if !entry.BookValueAfter.Equal(finalValue) {
        return errors.New("final book value after write-off is incorrect")
    }

    return nil
}
```

#### Adjustment Entry Rules
```go
func ValidateAdjustmentEntry(entry *AmortizationLedgerEntry) error {
    if entry.EntryType != "adjustment" && entry.EntryType != "correction" {
        return nil // Not an adjustment entry
    }

    // Must have adjustment reason
    if entry.AdjustmentReason == nil || strings.TrimSpace(*entry.AdjustmentReason) == "" {
        return errors.New("adjustment entries must include a reason")
    }

    // Adjustments should be made by users, not system-generated
    if entry.IsSystemGenerated {
        return errors.New("system-generated entries cannot be adjustments")
    }

    // Correction entries must reference the entry being corrected
    if entry.EntryType == "correction" && entry.CorrectsEntryID == nil {
        return errors.New("correction entries must reference the entry being corrected")
    }

    return nil
}
```

### 4. Status Transition Validation

#### Lifecycle Status Validation
```go
func ValidateStatusTransition(current, new *LifecycleStatus) error {
    // Cannot transition from terminal back to active
    if current.AmortizationBehavior == "terminal" && new.AmortizationBehavior == "active" {
        return errors.New("cannot transition from terminal to active amortization status")
    }

    // Write-offs should be generated for terminal status transitions
    if current.AmortizationBehavior != "terminal" && new.AmortizationBehavior == "terminal" {
        // This should trigger a write-off in the business logic
        log.Info("Terminal status transition detected - write-off should be generated")
    }

    return nil
}
```

## Business Logic Validation

### 1. Amortization Schedule Validation

#### Monthly Processing Rules
```go
func ValidateMonthlyDepreciation(ci *ConfigurationItem, processDate time.Time) error {
    // Check if amortization has started
    if ci.AmortStartDate == nil {
        return errors.New("amortization has not started")
    }

    // Check if amortization is complete
    if ci.CurrentBookValue != nil && ci.CurrentBookValue.LessThanOrEqual(decimal.Zero) {
        return errors.New("amortization is already complete")
    }

    // Check if depreciation already processed for this month
    lastDepreciationDate := getLastDepreciationDate(ci.ID)
    if isSameMonth(lastDepreciationDate, processDate) {
        return errors.New("depreciation already processed for this month")
    }

    // Calculate remaining life
    monthsElapsed := calculateMonthsElapsed(*ci.AmortStartDate, processDate)
    if monthsElapsed >= *ci.UsefulLifeMonths {
        return errors.New("amortization period has expired")
    }

    return nil
}
```

#### Proportional Depreciation Calculation
```go
func CalculateProportionalDepreciation(ci *ConfigurationItem, startDate, endDate time.Time) (decimal.Decimal, error) {
    monthlyDepreciation, err := CalculateMonthlyDepreciation(
        *ci.PurchaseCost,
        getSalvageValueOrZero(ci.SalvageValue),
        *ci.UsefulLifeMonths,
    )
    if err != nil {
        return decimal.Zero, err
    }

    // Calculate days in period
    daysInPeriod := endDate.Sub(startDate).Hours() / 24
    daysInMonth := daysInMonth(startDate)

    // Proportional depreciation
    proportionalAmount := monthlyDepreciation.Mul(
        decimal.NewFromFloat(daysInPeriod / daysInMonth),
    )

    return proportionalAmount.Round(2), nil
}
```

### 2. Asset Write-Off Validation

#### Write-Off Eligibility
```go
func ValidateWriteOffEligibility(ci *ConfigurationItem) error {
    // Asset must be amortizable
    if !ci.CIType.IsAmortizable {
        return errors.New("asset is not amortizable")
    }

    // Must have current book value
    if ci.CurrentBookValue == nil || ci.CurrentBookValue.LessThanOrEqual(decimal.Zero) {
        return errors.New("asset has no book value to write off")
    }

    // Check lifecycle status
    if ci.LifecycleStatus != nil && ci.LifecycleStatus.AmortizationBehavior != "terminal" {
        return errors.New("asset lifecycle status does not permit write-off")
    }

    return nil
}
```

### 3. Adjustment and Correction Validation

#### Adjustment Limits
```go
func ValidateAdjustmentLimits(ci *ConfigurationItem, adjustmentAmount decimal.Decimal) error {
    // Adjustments cannot be negative
    if adjustmentAmount.LessThan(decimal.Zero) {
        return errors.New("adjustment amount cannot be negative")
    }

    // Adjustments cannot reduce book value below salvage value
    minBookValue := getSalvageValueOrZero(ci.SalvageValue)
    maxAdjustment := ci.CurrentBookValue.Sub(minBookValue)

    if adjustmentAmount.GreaterThan(maxAdjustment) {
        return fmt.Errorf("adjustment amount %.2f exceeds maximum allowed %.2f",
            adjustmentAmount, maxAdjustment)
    }

    return nil
}
```

#### Correction Chain Validation
```go
func ValidateCorrectionChain(originalEntry, correctionEntry *AmortizationLedgerEntry) error {
    // Correction must be for the same CI
    if originalEntry.CIID != correctionEntry.CIID {
        return errors.New("correction must be for the same CI")
    }

    // Correction must be after original entry
    if correctionEntry.EntryDate.Before(originalEntry.EntryDate) {
        return errors.New("correction date cannot precede original entry date")
    }

    // Cannot correct a correction (create new correction instead)
    if originalEntry.EntryType == "correction" {
        return errors.New("cannot correct a correction entry")
    }

    // Correction amount should reverse and restate
    if !correctionEntry.Amount.Equal(originalEntry.Amount) {
        return errors.New("correction amount must match original entry amount")
    }

    return nil
}
```

## Data Integrity Checks

### 1. Consistency Validation Queries

#### Ledger Balance Verification
```sql
-- Verify CI book values match latest ledger entries
WITH latest_ledger AS (
    SELECT DISTINCT ON (ci_id)
        ci_id,
        book_value_after,
        entry_date
    FROM amortization_ledger
    ORDER BY ci_id, entry_date DESC, sequence_number DESC
)
SELECT
    ci.id as ci_id,
    ci.name as ci_name,
    ci.current_book_value as ci_book_value,
    ll.book_value_after as ledger_book_value,
    CASE
        WHEN ci.current_book_value IS NULL AND ll.book_value_after IS NULL THEN 'OK'
        WHEN ci.current_book_value = ll.book_value_after THEN 'OK'
        ELSE 'MISMATCH'
    END as status
FROM configuration_items ci
LEFT JOIN latest_ledger ll ON ci.id = ll.ci_id
WHERE ci.current_book_value IS NOT NULL
AND ll.book_value_after IS NOT NULL
AND ci.current_book_value != ll.book_value_after;
```

#### Duplicate Entry Detection
```sql
-- Find potential duplicate depreciation entries
SELECT
    ci_id,
    entry_date,
    COUNT(*) as duplicate_count,
    STRING_AGG(sequence_number::text, ', ') as sequences
FROM amortization_ledger
WHERE entry_type = 'depreciation'
GROUP BY ci_id, entry_date
HAVING COUNT(*) > 1;
```

#### Negative Book Value Detection
```sql
-- Find assets with negative book values
SELECT
    ci.id,
    ci.name,
    ci.current_book_value,
    al.entry_date,
    al.entry_type,
    al.amount
FROM configuration_items ci
JOIN amortization_ledger al ON ci.id = al.ci_id
WHERE ci.current_book_value < 0
ORDER BY ci.current_book_value;
```

### 2. Automated Validation Functions

#### Ledger Integrity Function
```sql
CREATE OR REPLACE FUNCTION validate_ledger_integrity(p_ci_id UUID)
RETURNS TABLE(
    check_name TEXT,
    status TEXT,
    details TEXT
) AS $$
BEGIN
    -- Check 1: Book value consistency
    RETURN QUERY
    SELECT
        'Book Value Consistency'::TEXT,
        CASE
            WHEN ci.current_book_value = latest.book_value_after THEN 'PASS'
            ELSE 'FAIL'
        END::TEXT,
        CASE
            WHEN ci.current_book_value = latest.book_value_after THEN 'Book values match'
            ELSE 'Book value mismatch: ' || ci.current_book_value || ' vs ' || latest.book_value_after
        END::TEXT
    FROM configuration_items ci
    LEFT JOIN LATERAL (
        SELECT book_value_after
        FROM amortization_ledger
        WHERE ci_id = p_ci_id
        ORDER BY entry_date DESC, sequence_number DESC
        LIMIT 1
    ) latest ON true
    WHERE ci.id = p_ci_id;

    -- Check 2: No negative book values
    RETURN QUERY
    SELECT
        'No Negative Book Values'::TEXT,
        CASE WHEN MIN(book_value_after) >= 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        CASE
            WHEN MIN(book_value_after) >= 0 THEN 'All book values are non-negative'
            ELSE 'Negative book value found: ' || MIN(book_value_after)
        END::TEXT
    FROM amortization_ledger
    WHERE ci_id = p_ci_id;

    -- Check 3: Sequential date integrity
    RETURN QUERY
    SELECT
        'Sequential Dates'::TEXT,
        CASE WHEN has_gaps THEN 'FAIL' ELSE 'PASS' END::TEXT,
        CASE
            WHEN has_gaps THEN 'Gaps found in depreciation sequence'
            ELSE 'Depreciation dates are sequential'
        END::TEXT
    FROM (
        SELECT
            bool_or(prev_date IS NOT NULL AND entry_date > prev_date + INTERVAL '1 month') as has_gaps
        FROM (
            SELECT
                entry_date,
                LAG(entry_date) OVER (ORDER BY entry_date) as prev_date
            FROM amortization_ledger
            WHERE ci_id = p_ci_id AND entry_type = 'depreciation'
        ) seq
    ) gaps;
END;
$$ LANGUAGE plpgsql;
```

## Error Handling and Recovery

### 1. Validation Error Categories

#### Critical Errors (Stop Processing)
- Negative book values
- Invalid financial amounts
- Missing required amortization parameters
- Constraint violations

#### Warning Conditions (Log and Continue)
- Rounding differences in calculations
- Non-critical data inconsistencies
- Missing optional reference data

#### Informational Messages (Log Only)
- Successful validations
- Scheduled maintenance events
- Performance statistics

### 2. Error Recovery Procedures

#### Data Correction Process
1. **Identify Issue**: Run validation queries to locate problems
2. **Assess Impact**: Determine affected assets and financial impact
3. **Create Correction**: Generate adjustment or correction entries
4. **Validate Fix**: Re-run validation to confirm resolution
5. **Document Change**: Record in audit logs with justification

#### Automated Recovery
```sql
-- Function to fix book value inconsistencies
CREATE OR REPLACE FUNCTION fix_book_value_inconsistency(p_ci_id UUID, p_user_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    ci_book_value DECIMAL(15,2);
    ledger_book_value DECIMAL(15,2);
    correction_amount DECIMAL(15,2);
BEGIN
    -- Get current values
    SELECT current_book_value INTO ci_book_value
    FROM configuration_items WHERE id = p_ci_id;

    SELECT book_value_after INTO ledger_book_value
    FROM amortization_ledger
    WHERE ci_id = p_ci_id
    ORDER BY entry_date DESC, sequence_number DESC
    LIMIT 1;

    -- If values match, no correction needed
    IF ci_book_value = ledger_book_value THEN
        RETURN true;
    END IF;

    -- Create correction entry
    correction_amount := ABS(ci_book_value - ledger_book_value);

    INSERT INTO amortization_ledger (
        ci_id, entry_date, entry_type, amount,
        book_value_before, book_value_after,
        adjustment_reason, created_by, is_system_generated
    ) VALUES (
        p_ci_id, CURRENT_DATE, 'correction', correction_amount,
        GREATEST(ci_book_value, ledger_book_value),
        LEAST(ci_book_value, ledger_book_value),
        'Automated correction of book value inconsistency',
        p_user_id, true
    );

    -- Update CI book value
    UPDATE configuration_items
    SET current_book_value = ledger_book_value
    WHERE id = p_ci_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql;
```

This comprehensive validation framework ensures data integrity, financial accuracy, and regulatory compliance for the amortization module while providing mechanisms for error detection, correction, and recovery.
-- Add 'restructuring' to valid entry types for amortization_ledger
-- This allows recording when an asset's useful life is changed

-- Drop the old constraint
ALTER TABLE amortization_ledger DROP CONSTRAINT valid_entry_type;

-- Add the new constraint with 'restructuring' included
ALTER TABLE amortization_ledger
ADD CONSTRAINT valid_entry_type
CHECK (entry_type IN ('depreciation', 'monthly_depreciation', 'catch_up_depreciation', 'write_off', 'adjustment', 'correction', 'restructuring'));

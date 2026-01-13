-- Migration: Add amortization support to CMDB
-- This migration adds financial tracking and amortization capabilities to the system

-- 1. Add amortization_behavior to lifecycle_statuses
ALTER TABLE lifecycle_statuses
ADD COLUMN amortization_behavior VARCHAR(20)
CHECK (amortization_behavior IN ('pending', 'active', 'terminal'))
DEFAULT 'active';

-- Set default behavior for existing statuses
UPDATE lifecycle_statuses SET amortization_behavior = 'active' WHERE amortization_behavior IS NULL;

-- 2. Add is_amortizable flag to ci_type_definitions
ALTER TABLE ci_type_definitions
ADD COLUMN is_amortizable BOOLEAN DEFAULT FALSE;

-- 3. Add financial columns to configuration_items
ALTER TABLE configuration_items ADD COLUMN purchase_cost DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN salvage_value DECIMAL(19,4);
ALTER TABLE configuration_items ADD COLUMN amort_start_date DATE;
ALTER TABLE configuration_items ADD COLUMN useful_life_months INT;
ALTER TABLE configuration_items ADD COLUMN current_book_value DECIMAL(19,4);

-- Index for amortizable assets (partial index for performance)
CREATE INDEX idx_cis_amortizable ON configuration_items(current_book_value)
WHERE current_book_value IS NOT NULL;

-- 4. Create amortization_ledger table (financial history/reporting)
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

-- 5. Create amortization_runs table (scheduler checkpoint)
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

-- 6. Add new permission for amortization adjustments
INSERT INTO permissions (name, description, resource_type)
VALUES ('amortization:adjust', 'Create amortization adjustments', 'amortization')
ON CONFLICT (name) DO NOTHING;

-- Assign amortization:adjust permission to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name = 'amortization:adjust'
ON CONFLICT DO NOTHING;

-- Migration: Add IT Asset Amortization Module
-- This migration adds comprehensive amortization tracking for IT assets

-- Add amortization support to CI type definitions
ALTER TABLE ci_type_definitions
    ADD COLUMN is_amortizable BOOLEAN DEFAULT false NOT NULL;

-- Add amortization behavior to lifecycle statuses
ALTER TABLE lifecycle_statuses
    ADD COLUMN amortization_behavior VARCHAR(20) DEFAULT 'pending' NOT NULL
    CONSTRAINT valid_amortization_behavior
    CHECK (amortization_behavior IN ('pending', 'active', 'terminal'));

-- Add financial columns to configuration items
ALTER TABLE configuration_items
    ADD COLUMN purchase_cost DECIMAL(15,2) NULL,
    ADD COLUMN salvage_value DECIMAL(15,2) NULL,
    ADD COLUMN amort_start_date DATE NULL,
    ADD COLUMN useful_life_months INTEGER NULL,
    ADD COLUMN current_book_value DECIMAL(15,2) NULL;

-- Add constraints for financial data
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_purchase_cost CHECK (purchase_cost IS NULL OR purchase_cost >= 0),
    ADD CONSTRAINT valid_salvage_value CHECK (salvage_value IS NULL OR salvage_value >= 0),
    ADD CONSTRAINT valid_salvage_vs_purchase CHECK (
        purchase_cost IS NULL OR
        salvage_value IS NULL OR
        salvage_value <= purchase_cost
    ),
    ADD CONSTRAINT valid_useful_life_months CHECK (useful_life_months IS NULL OR useful_life_months > 0),
    ADD CONSTRAINT valid_book_value CHECK (current_book_value IS NULL OR current_book_value >= 0),
    ADD CONSTRAINT amortization_dates_consistency CHECK (
        amort_start_date IS NULL OR
        useful_life_months IS NULL OR
        (purchase_cost IS NOT NULL AND useful_life_months IS NOT NULL)
    );

-- Create amortization runs table (scheduler execution tracking)
CREATE TABLE amortization_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
    CONSTRAINT valid_run_status
    CHECK (status IN ('pending', 'started', 'running', 'completed', 'partial', 'failed', 'cancelled')),

    -- Execution metrics
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_cis_processed INTEGER DEFAULT 0,
    successful_depreciations INTEGER DEFAULT 0,
    write_offs_generated INTEGER DEFAULT 0,
    errors_encountered INTEGER DEFAULT 0,

    -- Checkpoint mechanism for restarts
    last_processed_ci_id UUID,
    last_processed_at TIMESTAMP WITH TIME ZONE,
    checkpoint_data JSONB DEFAULT '{}',

    -- Error handling
    error_summary TEXT,
    error_details JSONB DEFAULT '{}',

    -- Configuration snapshot
    run_config JSONB DEFAULT '{}',

    -- Manual vs scheduled run identification
    is_manual BOOLEAN DEFAULT false NOT NULL,

    -- System fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    -- Unique constraint for one run per day
    CONSTRAINT unique_run_per_date UNIQUE (run_date)
);

-- Create amortization ledger table (append-only audit trail)
CREATE TABLE amortization_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    entry_date DATE NOT NULL,
    entry_type VARCHAR(20) NOT NULL
    CONSTRAINT valid_entry_type
    CHECK (entry_type IN ('depreciation', 'write_off', 'adjustment', 'correction')),

    -- Financial amounts
    amount DECIMAL(15,2) NOT NULL,
    book_value_before DECIMAL(15,2) NOT NULL,
    book_value_after DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    -- Description field for all entry types
    description TEXT,

    -- Period information for depreciation entries
    period_start_date DATE,
    period_end_date DATE,
    days_in_period INTEGER,

    -- Adjustment and correction metadata
    adjustment_reason TEXT,
    adjustment_reference VARCHAR(100),
    corrects_entry_id UUID REFERENCES amortization_ledger(id),

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    -- System-generated fields
    is_system_generated BOOLEAN DEFAULT false,
    batch_run_id UUID REFERENCES amortization_runs(id),
    amortization_run_id UUID REFERENCES amortization_runs(id),
    sequence_number INTEGER NOT NULL,

    -- Natural ordering for audit trail
    CONSTRAINT unique_ledger_entry UNIQUE (ci_id, entry_date, sequence_number)
);

-- Create amortization summaries table for optimized reporting
CREATE TABLE amortization_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    reporting_date DATE NOT NULL,

    -- Current financial state
    current_book_value DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL,

    -- Period calculations
    period_depreciation DECIMAL(15,2) DEFAULT 0,
    period_adjustments DECIMAL(15,2) DEFAULT 0,

    -- Status and metadata
    amortization_status VARCHAR(20) NOT NULL
    CONSTRAINT valid_summary_status
    CHECK (amortization_status IN ('pending', 'active', 'completed', 'written_off')),

    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Unique constraint for CI and date
    CONSTRAINT unique_summary_per_ci_date UNIQUE (ci_id, reporting_date)
);

-- Create indexes for amortization performance
-- CI type definitions indexes
CREATE INDEX idx_ci_types_amortizable ON ci_type_definitions(is_amortizable);

-- Lifecycle statuses indexes for amortization behavior queries
CREATE INDEX idx_lifecycle_amortization_behavior ON lifecycle_statuses(amortization_behavior);

-- Configuration items financial indexes
CREATE INDEX idx_cis_purchase_cost ON configuration_items(purchase_cost) WHERE purchase_cost IS NOT NULL;
CREATE INDEX idx_cis_amort_start_date ON configuration_items(amort_start_date) WHERE amort_start_date IS NOT NULL;
CREATE INDEX idx_cis_useful_life ON configuration_items(useful_life_months) WHERE useful_life_months IS NOT NULL;
CREATE INDEX idx_cis_book_value ON configuration_items(current_book_value) WHERE current_book_value IS NOT NULL;

-- Composite index for finding active amortization assets
CREATE INDEX idx_cis_amortization_active ON configuration_items(amort_start_date, useful_life_months, current_book_value)
WHERE amort_start_date IS NOT NULL
AND useful_life_months IS NOT NULL
AND current_book_value > 0;

-- Amortization ledger indexes
CREATE INDEX idx_ledger_ci_id ON amortization_ledger(ci_id);
CREATE INDEX idx_ledger_entry_date ON amortization_ledger(entry_date);
CREATE INDEX idx_ledger_entry_type ON amortization_ledger(entry_type);
CREATE INDEX idx_ledger_created_at ON amortization_ledger(created_at);
CREATE INDEX idx_ledger_batch_run ON amortization_ledger(batch_run_id) WHERE batch_run_id IS NOT NULL;
CREATE INDEX idx_ledger_amortization_run ON amortization_ledger(amortization_run_id) WHERE amortization_run_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX idx_ledger_ci_date_type ON amortization_ledger(ci_id, entry_date, entry_type);
CREATE INDEX idx_ledger_corrections ON amortization_ledger(corrects_entry_id) WHERE corrects_entry_id IS NOT NULL;

-- Amortization runs indexes
CREATE INDEX idx_runs_status ON amortization_runs(status);
CREATE INDEX idx_runs_date ON amortization_runs(run_date);
CREATE INDEX idx_runs_started_at ON amortization_runs(started_at) WHERE started_at IS NOT NULL;

-- Amortization summaries indexes
CREATE INDEX idx_summaries_ci_id ON amortization_summaries(ci_id);
CREATE INDEX idx_summaries_date ON amortization_summaries(reporting_date);
CREATE INDEX idx_summaries_status ON amortization_summaries(amortization_status);

-- Composite index for reporting queries
CREATE INDEX idx_summaries_ci_date ON amortization_summaries(ci_id, reporting_date);

-- Create triggers for amortization data consistency
-- Function to update current_book_value when CI financial data changes
CREATE OR REPLACE FUNCTION update_ci_book_value()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate initial book value if purchase cost is set and no amortization has started
    IF NEW.purchase_cost IS NOT NULL AND NEW.amort_start_date IS NULL THEN
        NEW.current_book_value = NEW.purchase_cost;
    ELSIF NEW.purchase_cost IS NULL THEN
        NEW.current_book_value = NULL;
    END IF;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update book value on CI financial changes
CREATE TRIGGER update_ci_book_value_trigger
    BEFORE INSERT OR UPDATE ON configuration_items
    FOR EACH ROW EXECUTE FUNCTION update_ci_book_value();

-- Function to update amortization summary
CREATE OR REPLACE FUNCTION update_amortization_summary()
RETURNS TRIGGER AS $$
BEGIN
    -- Update summary for current date when ledger entry is created
    INSERT INTO amortization_summaries (
        ci_id,
        reporting_date,
        current_book_value,
        accumulated_depreciation,
        period_depreciation,
        amortization_status,
        last_updated_at
    )
    VALUES (
        NEW.ci_id,
        NEW.entry_date,
        NEW.book_value_after,
        (
            SELECT COALESCE(SUM(CASE WHEN entry_type = 'depreciation' THEN amount ELSE 0 END), 0)
            FROM amortization_ledger
            WHERE ci_id = NEW.ci_id
            AND entry_date <= NEW.entry_date
        ),
        CASE WHEN NEW.entry_type = 'depreciation' THEN NEW.amount ELSE 0 END,
        CASE
            WHEN NEW.entry_type = 'write_off' THEN 'written_off'
            WHEN NEW.book_value_after <= 0 THEN 'completed'
            ELSE 'active'
        END,
        NOW()
    )
    ON CONFLICT (ci_id, reporting_date)
    DO UPDATE SET
        current_book_value = EXCLUDED.current_book_value,
        accumulated_depreciation = EXCLUDED.accumulated_depreciation,
        period_depreciation = EXCLUDED.period_depreciation,
        amortization_status = EXCLUDED.amortization_status,
        last_updated_at = EXCLUDED.last_updated_at;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update summary on ledger changes
CREATE TRIGGER update_amortization_summary_trigger
    AFTER INSERT ON amortization_ledger
    FOR EACH ROW EXECUTE FUNCTION update_amortization_summary();

-- Function to maintain ledger sequence numbers
CREATE OR REPLACE FUNCTION set_ledger_sequence_number()
RETURNS TRIGGER AS $$
DECLARE
    max_sequence INTEGER;
BEGIN
    -- Get the next sequence number for this CI and date
    SELECT COALESCE(MAX(sequence_number), 0) + 1
    INTO max_sequence
    FROM amortization_ledger
    WHERE ci_id = NEW.ci_id AND entry_date = NEW.entry_date;

    NEW.sequence_number = max_sequence;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to set sequence numbers before insert
CREATE TRIGGER set_ledger_sequence_trigger
    BEFORE INSERT ON amortization_ledger
    FOR EACH ROW EXECUTE FUNCTION set_ledger_sequence_number();

-- Update lifecycle statuses with amortization behavior
UPDATE lifecycle_statuses SET amortization_behavior = 'pending' WHERE name IN (
    'planned', 'on_order', 'in_stock', 'pending_install'
);

UPDATE lifecycle_statuses SET amortization_behavior = 'active' WHERE name IN (
    'operational', 'in_maintenance', 'defective_repair'
);

UPDATE lifecycle_statuses SET amortization_behavior = 'terminal' WHERE name IN (
    'retired', 'disposed', 'missing_stolen'
);

-- Add RBAC permissions for amortization management
INSERT INTO permissions (name, description, resource_type) VALUES
    ('amortization:read', 'Read amortization data and reports', 'amortization'),
    ('amortization:configure', 'Configure amortization settings and CI types', 'amortization'),
    ('amortization:adjust', 'Make manual adjustments to amortization entries', 'amortization'),
    ('amortization:run', 'Trigger amortization calculations', 'amortization'),
    ('amortization:admin', 'Full amortization system administration', 'amortization')
ON CONFLICT (name) DO NOTHING;

-- Assign amortization permissions to roles
-- Admin gets all amortization permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
AND p.name LIKE 'amortization:%'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Editor gets read, configure, and adjust permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'editor'
AND p.name IN ('amortization:read', 'amortization:configure', 'amortization:adjust')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Viewer gets read permission only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'viewer'
AND p.name = 'amortization:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create view for active amortization assets
CREATE VIEW active_amortization_assets AS
SELECT
    ci.id,
    ci.name,
    ci.ci_type,
    ci.purchase_cost,
    ci.salvage_value,
    ci.amort_start_date,
    ci.useful_life_months,
    ci.current_book_value,
    ls.name as lifecycle_status,
    ls.amortization_behavior,
    ctd.is_amortizable,
    -- Calculated fields
    CASE
        WHEN ci.amort_start_date IS NOT NULL AND ci.useful_life_months IS NOT NULL THEN
            GREATEST(0, ci.useful_life_months -
                EXTRACT(MONTHS FROM AGE(CURRENT_DATE, ci.amort_start_date)))
        ELSE NULL
    END as remaining_months,
    CASE
        WHEN ci.amort_start_date IS NOT NULL AND ci.useful_life_months IS NOT NULL THEN
            ROUND((ci.purchase_cost - COALESCE(ci.salvage_value, 0)) / ci.useful_life_months, 2)
        ELSE NULL
    END as monthly_depreciation
FROM configuration_items ci
JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
WHERE ctd.is_amortizable = true
AND ci.amort_start_date IS NOT NULL
AND ci.useful_life_months IS NOT NULL
AND ci.current_book_value > 0
AND (ci.current_book_value > COALESCE(ci.salvage_value, 0) OR ci.salvage_value IS NULL);

-- Create view for amortization reporting
CREATE VIEW amortization_report AS
SELECT
    ci.id as ci_id,
    ci.name as ci_name,
    ci.ci_type,
    ci.purchase_cost,
    ci.salvage_value,
    ci.amort_start_date,
    ci.useful_life_months,
    al.accumulated_depreciation,
    al.current_book_value,
    al.amortization_status,
    al.reporting_date,
    -- Year-to-date depreciation
    (
        SELECT COALESCE(SUM(amount), 0)
        FROM amortization_ledger al2
        WHERE al2.ci_id = ci.id
        AND al2.entry_type = 'depreciation'
        AND al2.entry_date >= DATE_TRUNC('year', al.reporting_date)
        AND al2.entry_date <= al.reporting_date
    ) as ytd_depreciation,
    -- Total adjustments
    (
        SELECT COALESCE(SUM(amount), 0)
        FROM amortization_ledger al3
        WHERE al3.ci_id = ci.id
        AND al3.entry_type IN ('adjustment', 'correction')
    ) as total_adjustments
FROM configuration_items ci
JOIN amortization_summaries al ON ci.id = al.ci_id
WHERE ci.purchase_cost IS NOT NULL
AND al.reporting_date = (
    SELECT MAX(reporting_date)
    FROM amortization_summaries al2
    WHERE al2.ci_id = ci.id
);

-- Add comments for documentation
COMMENT ON TABLE amortization_ledger IS 'Append-only ledger tracking all amortization transactions including depreciation, write-offs, and adjustments';
COMMENT ON TABLE amortization_runs IS 'Scheduler execution tracking with checkpoint mechanism for restart capability';
COMMENT ON TABLE amortization_summaries IS 'Pre-calculated summaries for optimized amortization reporting';
COMMENT ON COLUMN amortization_ledger.sequence_number IS 'Sequence number for ordering multiple entries on the same date';
COMMENT ON COLUMN amortization_ledger.corrects_entry_id IS 'Reference to the entry being corrected (for audit trail)';
COMMENT ON COLUMN amortization_runs.checkpoint_data IS 'JSON data for scheduler restart capability';
COMMENT ON COLUMN amortization_runs.error_details IS 'Detailed error information for failed runs';
COMMENT ON VIEW active_amortization_assets IS 'Current assets actively being amortized with calculated fields';
COMMENT ON VIEW amortization_report IS 'Financial reporting view with year-to-date calculations and totals';

-- Create function for amortization calculation (will be used by backend service)
CREATE OR REPLACE FUNCTION calculate_monthly_depreciation(
    p_purchase_cost DECIMAL(15,2),
    p_salvage_value DECIMAL(15,2),
    p_useful_life_months INTEGER
) RETURNS DECIMAL(15,2) AS $$
BEGIN
    RETURN ROUND((p_purchase_cost - COALESCE(p_salvage_value, 0)) / p_useful_life_months, 2);
END;
$$ language 'plpgsql';

-- Create function for asset write-off amount calculation
CREATE OR REPLACE FUNCTION calculate_write_off_amount(
    p_current_book_value DECIMAL(15,2),
    p_salvage_value DECIMAL(15,2)
) RETURNS DECIMAL(15,2) AS $$
BEGIN
    RETURN GREATEST(0, p_current_book_value - COALESCE(p_salvage_value, 0));
END;
$$ language 'plpgsql';

-- Create function to get next amortization sequence for a CI
CREATE OR REPLACE FUNCTION get_next_amortization_sequence(
    p_ci_id UUID,
    p_entry_date DATE
) RETURNS INTEGER AS $$
DECLARE
    next_seq INTEGER;
BEGIN
    SELECT COALESCE(MAX(sequence_number), 0) + 1
    INTO next_seq
    FROM amortization_ledger
    WHERE ci_id = p_ci_id AND entry_date = p_entry_date;

    RETURN next_seq;
END;
$$ language 'plpgsql';

-- Fix amortization runs constraint to allow multiple manual runs
ALTER TABLE amortization_runs
    DROP CONSTRAINT IF EXISTS unique_run_per_date;

-- Add the new constraint that allows multiple manual runs but only one scheduled run per day
ALTER TABLE amortization_runs
    ADD CONSTRAINT unique_scheduled_run_per_date
    UNIQUE (run_date, is_manual)
    DEFERRABLE INITIALLY DEFERRED;
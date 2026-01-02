-- Add health_score_history table for tracking CMDB health over time
-- This enables trend calculation and historical analysis

CREATE TABLE IF NOT EXISTS health_score_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Overall health score (0-100)
    overall_score NUMERIC(5,2) NOT NULL CHECK (overall_score >= 0 AND overall_score <= 100),

    -- Sub-scores
    completeness_score NUMERIC(5,2) NOT NULL CHECK (completeness_score >= 0 AND completeness_score <= 100),
    correctness_score NUMERIC(5,2) NOT NULL CHECK (correctness_score >= 0 AND correctness_score <= 100),
    compliance_score NUMERIC(5,2) NOT NULL CHECK (compliance_score >= 0 AND compliance_score <= 100),

    -- Metrics used for calculation (useful for debugging and drill-down)
    total_cis INTEGER NOT NULL DEFAULT 0,
    complete_cis INTEGER NOT NULL DEFAULT 0,
    current_cis INTEGER NOT NULL DEFAULT 0, -- updated in 90 days
    compliant_cis INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient time-series queries
CREATE INDEX idx_health_score_history_calculated_at ON health_score_history(calculated_at DESC);

-- Index for retrieving historical data within date ranges
CREATE INDEX idx_health_score_history_date_range ON health_score_history(calculated_at DESC);

-- Add a comment to document the table's purpose
COMMENT ON TABLE health_score_history IS 'Stores daily CMDB health score snapshots for trend analysis and historical reporting';

COMMENT ON COLUMN health_score_history.overall_score IS 'Overall health score (average of completeness, correctness, and compliance)';

COMMENT ON COLUMN health_score_history.completeness_score IS 'Percentage of CIs with all required attributes filled';

COMMENT ON COLUMN health_score_history.correctness_score IS 'Percentage of CIs updated in the last 90 days';

COMMENT ON COLUMN health_score_history.compliance_score IS 'Percentage of CIs following naming standards';

COMMENT ON COLUMN health_score_history.total_cis IS 'Total number of CIs in the system';

COMMENT ON COLUMN health_score_history.complete_cis IS 'Number of CIs with all required attributes filled';

COMMENT ON COLUMN health_score_history.current_cis IS 'Number of CIs updated in the last 90 days';

COMMENT ON COLUMN health_score_history.compliant_cis IS 'Number of CIs following naming standards';

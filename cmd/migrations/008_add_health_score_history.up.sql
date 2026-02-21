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

-- ============================================================================
-- Create Admin User (needed by migration 009)
-- ============================================================================
-- This creates the admin user that migration 009 (EA metamodel) requires.
-- The admin role is already created by migration 001.
-- The password hash is for "Admin@123"

-- Create admin user if not exists (using generated UUID)
INSERT INTO users (username, email, password_hash, is_active, created_at, updated_at)
VALUES (
    'admin',
    'admin@pustaka.local',
    '$argon2id$v=19$m=65536,t=3,p=4$y5XhcR4fSJK6YqL8cJZx6g$W9wKjXKmYLqKo4XhNXpGvQXLvLHqRqJqLZhBKDqYQYE',
    true,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    updated_at = NOW();

-- Assign admin role to admin user (role created in migration 001)
INSERT INTO user_roles (user_id, role_id, assigned_at, created_at)
SELECT
    (SELECT id FROM users WHERE username = 'admin'),
    (SELECT id FROM roles WHERE name = 'admin'),
    NOW(),
    NOW()
WHERE EXISTS (SELECT 1 FROM users WHERE username = 'admin')
  AND EXISTS (SELECT 1 FROM roles WHERE name = 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM user_roles
    WHERE user_id = (SELECT id FROM users WHERE username = 'admin')
      AND role_id = (SELECT id FROM roles WHERE name = 'admin')
  );

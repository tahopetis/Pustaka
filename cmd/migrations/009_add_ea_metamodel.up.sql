-- ============================================================================
-- Migration: 009_add_ea_metamodel
-- Created: 2026-02-20
-- Purpose: Seed EA metamodel (CI types, relationship types, teams, permissions)
-- ============================================================================

-- ============================================================================
-- SECTION 1: EA Teams Table
-- ============================================================================

-- Create ea_teams table for team-based ownership
CREATE TABLE IF NOT EXISTS ea_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Create index on name for lookups
CREATE INDEX idx_ea_teams_name ON ea_teams(name);

-- Add comment
COMMENT ON TABLE ea_teams IS 'EA teams for team-based ownership of EA entities';
COMMENT ON COLUMN ea_teams.created_by IS 'User who created the team (admin)';

-- Seed EA teams (8 teams, one per EA domain)
INSERT INTO ea_teams (name, description, created_by)
VALUES
    ('enterprise-architecture', 'Enterprise Architecture team responsible for strategic EA oversight', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('business-architecture', 'Business Architecture team responsible for business capability modeling', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('application-architecture', 'Application Architecture team responsible for application portfolio management', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('data-architecture', 'Data Architecture team responsible for data modeling and governance', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('technology-architecture', 'Technology Architecture team responsible for technology stack standards', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('infrastructure-architecture', 'Infrastructure Architecture team responsible for infrastructure planning', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('security-architecture', 'Security Architecture team responsible for security controls and policies', (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('governance', 'IT Governance team responsible for governance policies and compliance', (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Add update trigger for ea_teams
CREATE TRIGGER update_ea_teams_updated_at BEFORE UPDATE ON ea_teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

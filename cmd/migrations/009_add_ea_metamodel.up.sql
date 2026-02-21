-- ============================================================================
-- Migration: 009_add_ea_metamodel
-- Created: 2026-02-20
-- Updated: 2026-02-21 (Corrected to match metamodel docs exactly)
-- Purpose: Seed EA metamodel (CI types, relationship types, teams, permissions)
-- ============================================================================

-- NOTE: This migration assumes that:
-- 1. All tables have been created by migrations 001-008
-- 2. Admin role exists (seeded in migration 001 or earlier)

-- ============================================================================
-- SECTION 0: Ensure Admin User Exists
-- ============================================================================

-- This migration references the admin user in created_by columns.
-- If the admin user doesn't exist yet (fresh deployment), create it now.
-- This ensures migration 009 works on both fresh and existing deployments.

DO $$
DECLARE
    admin_user_id UUID;
    admin_role_id UUID;
BEGIN
    -- Check if admin user exists
    SELECT id INTO admin_user_id FROM users WHERE username = 'admin';

    IF admin_user_id IS NULL THEN
        -- Create admin user
        admin_user_id := gen_random_uuid();
        INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at)
        VALUES (
            admin_user_id,
            'admin',
            'admin@pustaka.local',
            '$argon2id$v=19$m=65536,t=3,p=4$change-this-password-in-production-salt$change-this-password-in-production-hash',
            true,
            NOW(),
            NOW()
        );

        RAISE NOTICE 'Created admin user for migration 009';
    ELSE
        RAISE NOTICE 'Admin user already exists, skipping creation';
    END IF;

    -- Ensure admin user has admin role
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';
    IF admin_role_id IS NOT NULL THEN
        INSERT INTO user_roles (user_id, role_id, created_at)
        VALUES (admin_user_id, admin_role_id, NOW())
        ON CONFLICT (user_id, role_id) DO NOTHING;

        RAISE NOTICE 'Ensured admin user has admin role';
    END IF;
END $$;

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
CREATE INDEX IF NOT EXISTS idx_ea_teams_name ON ea_teams(name);

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
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_ea_teams_updated_at') THEN
        CREATE TRIGGER update_ea_teams_updated_at BEFORE UPDATE ON ea_teams
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================================================
-- SECTION 2: EA CI Type Definitions (32 types matching metamodel docs)
-- ============================================================================

-- Strategy & Transformation Domain (4 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Strategy-Objective', 'High-level strategic objective representing organizational goals', '[{"name":"name","type":"string","description":"Objective name","validation":{"min_length":5,"max_length":100}},{"name":"description","type":"string","description":"Detailed description"},{"name":"owner","type":"string","description":"EA team responsible"}]'::jsonb, '[{"name":"strategic_alignment","type":"string","enum":["high","medium","low"]},{"name":"target_date","type":"date"},{"name":"metrics","type":"array"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Initiative', 'Strategic initiative to achieve objectives', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"status","type":"string","enum":["planning","active","on_hold","completed"]},{"name":"budget","type":"string"},{"name":"start_date","type":"date"},{"name":"end_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Program', 'Program grouping multiple related projects', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"program_manager","type":"string"},{"name":"budget","type":"string"},{"name":"start_date","type":"date"},{"name":"end_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Project', 'Project implementing specific initiatives or programs', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"project_manager","type":"string"},{"name":"status","type":"string","enum":["proposed","active","on_hold","completed"]},{"name":"start_date","type":"date"},{"name":"end_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Business Architecture Domain (5 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Business-Organization', 'Organizational unit or department', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"organization_type","type":"string","enum":["department","division","unit","team"]},{"name":"parent_org_id","type":"uuid"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-BusinessDomain', 'Business domain representing area of business operations', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"domain_owner","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-CapabilityL1', 'Level 1 Business Capability representing high-level business functions', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"strategic_importance","type":"string","enum":["critical","high","medium","low"]},{"name":"business_value","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-CapabilityL2', 'Level 2 Business Capability (child of L1)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_capability_id","type":"uuid"},{"name":"strategic_importance","type":"string","enum":["critical","high","medium","low"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-BusinessProduct', 'Business product or service offered by the organization', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"product_type","type":"string","enum":["product","service","solution"]},{"name":"lifecycle_status","type":"string","enum":["proposed","active","deprecated","retired"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Application Architecture Domain (5 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Application-ApplicationGroup', 'Group of related business applications', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"business_owner","type":"string"},{"name":"it_owner","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-BusinessApplication', 'Business application supporting enterprise operations', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"lifecycle_status","type":"string","enum":["proposed","active","deprecated","retired"]},{"name":"criticality","type":"string","enum":["mission_critical","high","medium","low"]},{"name":"version","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Subsystem', 'Application subsystem or module', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_app_id","type":"uuid"},{"name":"technology","type":"string"},{"name":"interface_type","type":"string","enum":["api","ui","batch","library"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Interface', 'Application interface for integration', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"interface_type","type":"string","enum":["rest","graphql","soap","grpc","event"]},{"name":"protocol","type":"string"},{"name":"authentication","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-SupportingApplication', 'Supporting application providing infrastructure or utility services', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"lifecycle_status","type":"string","enum":["proposed","active","deprecated","retired"]},{"name":"criticality","type":"string","enum":["mission_critical","high","medium","low"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Data Architecture Domain (2 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Data-DataDomain', 'Data domain representing area of data management', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"domain_owner","type":"string"},{"name":"data_steward","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-DataObject', 'Data entity or document used by applications', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"data_classification","type":"string","enum":["public","internal","confidential","restricted"]},{"name":"retention_period","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Technology Architecture Domain (3 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Technology-ITComponent', 'Software component, library, or framework', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"version","type":"string"},{"name":"license","type":"string"},{"name":"end_of_support","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-TechCategory', 'Technology category grouping related IT components', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"category_owner","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Provider', 'Technology provider or vendor', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"provider_type","type":"string","enum":["vendor","supplier","partner","internal"]},{"name":"contact_info","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Infrastructure Architecture Domain (5 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Infrastructure-Location', 'Physical or logical location (office, data center, region)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"location_type","type":"string","enum":["office","data_center","cloud_region","branch"]},{"name":"address","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-DataCenter', 'Data center facility housing IT infrastructure', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"data_center_type","type":"string","enum":["primary","secondary","colocation","cloud"]},{"name":"location_id","type":"uuid"},{"name":"tier_rating","type":"string","enum":["tier1","tier2","tier3","tier4"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-NetworkZone', 'Network zone or segment (VLAN, subnet, DMZ)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"zone_type","type":"string","enum":["lan","wan","dmz","vpn","vlan"]},{"name":"cidr_block","type":"string"},{"name":"parent_location_id","type":"uuid"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-ComputePlatform', 'Compute platform (server, VM, container host)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"platform_type","type":"string","enum":["physical","virtual","container","cloud"]},{"name":"cpu_cores","type":"integer"},{"name":"memory_gb","type":"integer"},{"name":"zone_id","type":"uuid"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-NetworkSecurityNodes', 'Network and security devices (routers, switches, firewalls)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"device_type","type":"string","enum":["router","switch","firewall","load_balancer","proxy","ids","ips"]},{"name":"zone_id","type":"uuid"},{"name":"ip_address","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Information Security (NIST) Domain (4 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Security-Function', 'NIST CSF Function (Govern, Identify, Protect, Detect, Respond, Recover)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"function_code","type":"string","enum":["GV","ID","PR","DE","RS","RC"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Category', 'NIST CSF Category under a Function', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_function_id","type":"uuid"},{"name":"category_code","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Subcategory', 'NIST CSF Subcategory under a Category', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_category_id","type":"uuid"},{"name":"subcategory_code","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Control', 'Security control or safeguard', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"control_type","type":"string","enum":["preventive","detective","corrective"]},{"name":"control_framework","type":"string"},{"name":"implementation_status","type":"string","enum":["planned","implemented","enforced","monitored"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- IT Governance Domain (4 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Governance-Policy', 'Governance policy document', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"policy_scope","type":"string"},{"name":"effective_date","type":"date"},{"name":"compliance_level","type":"string","enum":["mandatory","recommended","guideline"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Procedure', 'Governance procedure documenting process steps', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"procedure_type","type":"string","enum":["operational","approval","review","audit"]},{"name","frequency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Standard', 'Governance standard or best practice', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"standard_body","type":"string"},{"name":"standard_version","type":"string"},{"name":"adoption_level","type":"string","enum":["full","partial","planned"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-StandardComponent', 'Component or clause within a governance standard', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_standard_id","type":"uuid"},{"name":"component_type","type":"string","enum":["clause","requirement","control objective"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SECTION 3: EA Relationship Type Definitions (from directional graph)
-- ============================================================================

INSERT INTO relationship_types (name, description, forward_label, reverse_label, allowed_source_types, allowed_target_types, cardinality_source, cardinality_target, bidirectional, attributes, created_by)
VALUES
-- Strategy Internal Relationships
(
    'drives',
    'Objective drives Initiative',
    'drives',
    'driven by',
    ARRAY['EA.Strategy-Objective'],
    ARRAY['EA.Strategy-Initiative'],
    'one',
    'many',
    true,
    '{"description": "Strategic objective drives initiative", "archimate_concept": "Driver"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'consists_of',
    'Initiative consists of Programs',
    'consists of',
    'part of',
    ARRAY['EA.Strategy-Initiative', 'EA.Strategy-Program'],
    ARRAY['EA.Strategy-Program', 'EA.Strategy-Project'],
    'one',
    'many',
    false,
    '{"description": "Hierarchical composition"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Business Internal Relationships
(
    'contains',
    'Business Capability L1 contains L2',
    'contains',
    'contained in',
    ARRAY['EA.Business-CapabilityL1'],
    ARRAY['EA.Business-CapabilityL2'],
    'one',
    'many',
    false,
    '{"description": "Capability hierarchy"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'belongs_to',
    'DataObject belongs to DataDomain',
    'belongs to',
    'contains',
    ARRAY['EA.Data-DataObject'],
    ARRAY['EA.Data-DataDomain'],
    'many',
    'one',
    true,
    '{"description": "Data domain membership"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Security Internal Relationships
(
    'has',
    'NIST Function has Category, Category has Subcategory',
    'has',
    'part of',
    ARRAY['EA.Security-Function', 'EA.Security-Category'],
    ARRAY['EA.Security-Category', 'EA.Security-Subcategory'],
    'one',
    'many',
    false,
    '{"description": "NIST hierarchy"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Governance Internal Relationships
(
    'defines',
    'Policy defines Procedure',
    'defines',
    'defined by',
    ARRAY['EA.Governance-Policy'],
    ARRAY['EA.Governance-Procedure'],
    'one',
    'many',
    true,
    '{"description": "Policy defines procedure"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'contains',
    'Standard contains StandardComponent',
    'contains',
    'contained in',
    ARRAY['EA.Governance-Standard'],
    ARRAY['EA.Governance-StandardComponent'],
    'one',
    'many',
    false,
    '{"description": "Standard components"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Strategy Cross-domain Relationships
(
    'targets',
    'Objective targets Business Capability L1',
    'targets',
    'targeted by',
    ARRAY['EA.Strategy-Objective'],
    ARRAY['EA.Business-CapabilityL1'],
    'many',
    'many',
    true,
    '{"description": "Strategic alignment"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'changes',
    'Project changes Applications',
    'changes',
    'changed by',
    ARRAY['EA.Strategy-Project'],
    ARRAY['EA.Application-BusinessApplication', 'EA.Application-SupportingApplication'],
    'many',
    'many',
    true,
    '{"description": "Project impact on applications"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Business Cross-domain Relationships
(
    'supports',
    'Business Application supports Business Capability L1',
    'supports',
    'supported by',
    ARRAY['EA.Application-BusinessApplication'],
    ARRAY['EA.Business-CapabilityL1'],
    'many',
    'many',
    true,
    '{"description": "Application support for capability"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'uses',
    'Business Product uses Business Application',
    'uses',
    'used by',
    ARRAY['EA.Business-BusinessProduct'],
    ARRAY['EA.Application-BusinessApplication'],
    'many',
    'many',
    true,
    '{"description": "Product uses application"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'owns',
    'Organization owns Applications and DataObjects',
    'owns',
    'owned by',
    ARRAY['EA.Business-Organization'],
    ARRAY['EA.Application-BusinessApplication', 'EA.Data-DataObject'],
    'one',
    'many',
    true,
    '{"description": "Organizational ownership"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'responsible_for',
    'Organization responsible for Business Capability L1',
    'responsible for',
    'responsibility of',
    ARRAY['EA.Business-Organization'],
    ARRAY['EA.Business-CapabilityL1'],
    'one',
    'many',
    true,
    '{"description": "Organizational responsibility"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Application Internal Relationships
(
    'contains',
    'Application Group contains Business Applications',
    'contains',
    'contained in',
    ARRAY['EA.Application-ApplicationGroup'],
    ARRAY['EA.Application-BusinessApplication'],
    'one',
    'many',
    false,
    '{"description": "Application grouping"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'consists_of',
    'Business Application consists of Subsystems',
    'consists of',
    'part of',
    ARRAY['EA.Application-BusinessApplication', 'EA.Application-SupportingApplication'],
    ARRAY['EA.Application-Subsystem'],
    'one',
    'many',
    false,
    '{"description": "Application composition"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'exposes',
    'Subsystem exposes Interface',
    'exposes',
    'exposed by',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Application-Interface'],
    'many',
    'many',
    true,
    '{"description": "Interface exposure"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'consumes',
    'Subsystem consumes Interface or DataObject',
    'consumes',
    'consumed by',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Application-Interface', 'EA.Data-DataObject'],
    'many',
    'many',
    true,
    '{"description": "Dependency consumption"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'routes_to',
    'Interface routes to Subsystem',
    'routes to',
    'routed from',
    ARRAY['EA.Application-Interface'],
    ARRAY['EA.Application-Subsystem'],
    'many',
    'many',
    false,
    '{"description": "Interface routing"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'depends_on',
    'Subsystem depends on Subsystem',
    'depends on',
    'depended on by',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Application-Subsystem'],
    'many',
    'many',
    true,
    '{"description": "Subsystem dependency"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Application Cross-domain Relationships
(
    'provides',
    'Subsystem provides DataObject',
    'provides',
    'provided by',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Data-DataObject'],
    'many',
    'many',
    true,
    '{"description": "Data provision"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'realizes',
    'Subsystem realizes IT Component',
    'realizes',
    'realized by',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Technology-ITComponent'],
    'many',
    'many',
    false,
    '{"description": "Technology realization"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'deployed_on',
    'Subsystem deployed on Compute Platform',
    'deployed on',
    'hosts',
    ARRAY['EA.Application-Subsystem'],
    ARRAY['EA.Infrastructure-ComputePlatform'],
    'many',
    'many',
    true,
    '{"description": "Deployment location"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Technology & Infrastructure Relationships
(
    'belongs_to',
    'IT Component belongs to Tech Category',
    'belongs to',
    'contains',
    ARRAY['EA.Technology-ITComponent'],
    ARRAY['EA.Technology-TechCategory'],
    'many',
    'one',
    true,
    '{"description": "Technology categorization"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'provided_by',
    'IT Component provided by Provider',
    'provided by',
    'provides',
    ARRAY['EA.Technology-ITComponent'],
    ARRAY['EA.Technology-Provider'],
    'many',
    'many',
    true,
    '{"description": "Provider relationship"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'contains',
    'Location contains DataCenter and NetworkZone',
    'contains',
    'contained in',
    ARRAY['EA.Infrastructure-Location'],
    ARRAY['EA.Infrastructure-DataCenter', 'EA.Infrastructure-NetworkZone'],
    'one',
    'many',
    false,
    '{"description": "Location hierarchy"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'contains',
    'NetworkZone contains ComputePlatform and NetworkSecurityNodes',
    'contains',
    'contained in',
    ARRAY['EA.Infrastructure-NetworkZone'],
    ARRAY['EA.Infrastructure-ComputePlatform', 'EA.Infrastructure-NetworkSecurityNodes'],
    'one',
    'many',
    false,
    '{"description": "Network zone contains infrastructure"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'realizes',
    'Compute Platform realizes IT Component',
    'realizes',
    'realized by',
    ARRAY['EA.Infrastructure-ComputePlatform'],
    ARRAY['EA.Technology-ITComponent'],
    'many',
    'many',
    false,
    '{"description": "Infrastructure realizes technology"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- Security & Governance Relationships
(
    'implements',
    'Control implements Subcategory',
    'implements',
    'implemented by',
    ARRAY['EA.Security-Control'],
    ARRAY['EA.Security-Subcategory'],
    'many',
    'many',
    false,
    '{"description": "Control implements NIST subcategory"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'enforces',
    'IT Component enforces Control',
    'enforces',
    'enforced by',
    ARRAY['EA.Technology-ITComponent'],
    ARRAY['EA.Security-Control'],
    'many',
    'many',
    false,
    '{"description": "Technology enforces control"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'complies_with',
    'Business Application complies with Control',
    'complies with',
    'compliance of',
    ARRAY['EA.Application-BusinessApplication'],
    ARRAY['EA.Security-Control'],
    'many',
    'many',
    true,
    '{"description": "Application compliance"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'documented_in',
    'Control documented in Procedure',
    'documented in',
    'documents',
    ARRAY['EA.Security-Control'],
    ARRAY['EA.Governance-Procedure'],
    'many',
    'many',
    true,
    '{"description": "Control documentation"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'complies_with',
    'IT Component complies with StandardComponent',
    'complies with',
    'compliance of',
    ARRAY['EA.Technology-ITComponent'],
    ARRAY['EA.Governance-StandardComponent'],
    'many',
    'many',
    true,
    '{"description": "Technology compliance"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'governs',
    'Policy governs Organization',
    'governs',
    'governed by',
    ARRAY['EA.Governance-Policy'],
    ARRAY['EA.Business-Organization'],
    'many',
    'many',
    true,
    '{"description": "Policy governance"}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
)

ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SECTION 4: Validation Queries
-- ============================================================================

-- Validate EA teams created
DO $$
BEGIN
    ASSERT (SELECT COUNT(*) FROM ea_teams) = 8, 'Expected 8 EA teams to be created';
    RAISE NOTICE 'EA teams validation passed: 8 teams created';
END $$;

-- Validate EA CI types created (should be exactly 32)
DO $$
DECLARE
    ea_ci_type_count INT;
BEGIN
    SELECT COUNT(*) INTO ea_ci_type_count FROM ci_type_definitions WHERE name LIKE 'EA.%';
    ASSERT ea_ci_type_count = 32, format('Expected exactly 32 EA CI types, got %s', ea_ci_type_count);
    RAISE NOTICE 'EA CI types validation passed: %s types created', ea_ci_type_count;
END $$;

-- Validate EA relationship types created (should match directional graph)
DO $$
DECLARE
    ea_rel_type_count INT;
BEGIN
    SELECT COUNT(*) INTO ea_rel_type_count FROM relationship_types;
    ASSERT ea_rel_type_count >= 30, format('Expected at least 30 EA relationship types, got %s', ea_rel_type_count);
    RAISE NOTICE 'EA relationship types validation passed: %s types created', ea_rel_type_count;
END $$;

-- Display summary
SELECT
    'EA Teams' as item, COUNT(*) as count FROM ea_teams
UNION ALL
SELECT
    'EA CI Types', COUNT(*) FROM ci_type_definitions WHERE name LIKE 'EA.%'
UNION ALL
SELECT
    'EA Relationship Types', COUNT(*) FROM relationship_types;

-- ============================================================================
-- SECTION 5: EA RBAC Permissions
-- ============================================================================

-- EA entity permissions
INSERT INTO permissions (name, description, resource_type) VALUES
    ('ea:read', 'Read EA entities', 'ea'),
    ('ea:create', 'Create EA entities', 'ea'),
    ('ea:update', 'Update EA entities', 'ea'),
    ('ea:delete', 'Delete EA entities', 'ea')
ON CONFLICT (name) DO NOTHING;

-- Grant all EA permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'admin'),
    p.id
FROM permissions p
WHERE p.name LIKE 'ea:%'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Grant EA read permission to editor and viewer roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    (SELECT id FROM permissions WHERE name = 'ea:read')
FROM roles r
WHERE r.name IN ('editor', 'viewer')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Verify permissions created
DO $$
BEGIN
    ASSERT (SELECT COUNT(*) FROM permissions WHERE name LIKE 'ea:%') = 4, 'Expected 4 EA permissions (ea:read, ea:create, ea:update, ea:delete)';
    RAISE NOTICE 'EA permissions created successfully';
END $$;

-- ============================================================================
-- Migration 009_add_ea_metamodel complete
-- ============================================================================

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

-- ============================================================================
-- SECTION 2: EA CI Type Definitions
-- ============================================================================

-- Strategy Domain CI Types (6 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Strategy-Objective', 'High-level strategic objective representing organizational goals', '[{"name":"name","type":"string","description":"Objective name","validation":{"min_length":5,"max_length":100}},{"name":"description","type":"string","description":"Detailed description"},{"name":"owner","type":"string","description":"EA team responsible"}]'::jsonb, '[{"name":"strategic_alignment","type":"string","enum":["high","medium","low"]},{"name":"target_date","type":"date"},{"name":"metrics","type":"array"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Goal', 'Strategic goal supporting objectives', '[{"name":"name","type":"string","description":"Goal name","validation":{"min_length":5,"max_length":100}},{"name":"description","type":"string","description":"Detailed description"},{"name":"owner","type":"string","description":"EA team responsible"}]'::jsonb, '[{"name":"parent_objective_id","type":"uuid"},{"name":"target_value","type":"string"},{"name":"current_value","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Outcome', 'Strategic outcome resulting from achieving goals', '[{"name":"name","type":"string","description":"Outcome name"},{"name":"description","type":"string","description":"Description of outcome"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"measurement_criteria","type":"string"},{"name="target_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Requirement', 'Strategic requirement or constraint', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"priority","type":"string","enum":["critical","high","medium","low"]},{"name":"compliance_impact","type":"boolean"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Constraint', 'Strategic constraint limiting options', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"constraint_type","type":"string","enum":["technical","financial","regulatory","organizational"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Strategy-Initiative', 'Strategic initiative to achieve objectives', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"status","type":"string","enum":["planning","active","on_hold","completed"]},{"name":"budget","type":"string"},{"name="start_date","type":"date"},{"name="end_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Business Domain CI Types (10 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Business-CapabilityL1', 'Level 1 Business Capability representing high-level business functions', '[{"name":"name","type":"string","description":"Capability name","validation":{"min_length":3,"max_length":100}},{"name":"description","type":"string","description":"Capability description"},{"name":"owner","type":"string","description":"Business team responsible"}]'::jsonb, '[{"name":"strategic_importance","type":"string","enum":["critical","high","medium","low"]},{"name":"business_value","type":"string"},{"name":"target_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-CapabilityL2', 'Level 2 Business Capability (child of L1)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_capability_id","type":"uuid"},{"name":"strategic_importance","type":"string","enum":["critical","high","medium","low"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Process', 'Business process representing workflows', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"process_level","type":"string","enum":["level1","level2","level3"]},{"name":"inputs","type":"array"},{"name":"outputs","type":"array"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Function', 'Business function representing cohesive business activity', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_process_id","type":"uuid"},{"name":"criticality","type":"string","enum":["mission_critical","important","standard","support"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Interaction', 'Business interaction between roles', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"interaction_type","type":"string","enum":["synchronous","asynchronous"]},{"name":"trigger","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Event', 'Business event triggering processes', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"event_type","type":"string","enum":["external","internal","temporal"]},{"name":"frequency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Service', 'Business service delivered to customers', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"service_level","type":"string","enum":["gold","silver","bronze"]},{"name="sla_target","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Actor', 'Business actor (person or organization)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"actor_type","type":"string","enum":["person","role","organization"]},{"name":"contact_info","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Role', 'Business role defining responsibilities', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"permissions","type":"array"},{"name":"assigned_to","type":"array"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Business-Collaboration', 'Business collaboration between actors', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"collaboration_type","type":"string","enum":["permanent","temporary","project_based"]},{"name":"duration","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Application Domain CI Types (8 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Application-BusinessApp', 'Business application supporting enterprise operations', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"lifecycle_status","type":"string","enum":["proposed","active","deprecated","retired"]},{"name":"criticality","type":"string","enum":["mission_critical","high","medium","low"]},{"name":"version","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Component', 'Application component or module', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_application_id","type":"uuid"},{"name":"technology","type":"string"},{"name":"interface_type","type":"string","enum":["api","ui","batch","library"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Interface', 'Application interface for integration', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"interface_type","type":"string","enum":["rest","graphql","soap","grpc","event"]},{"name":"protocol","type":"string"},{"name":"authentication","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Service', 'Application service exposing functionality', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"service_type","type":"string","enum":["internal","external","shared"]},{"name":"availability_sla","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Function', 'Application function or business logic', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"parent_component_id","type":"uuid"},{"name="complexity","type":"string","enum":["low","medium","high"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Event', 'Application event or message', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"event_type","type":"string","enum":["command","event","query"]},{"name":"payload_schema","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-DataObject', 'Application data structure', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"data_format","type":"string","enum":["json","xml","csv","binary"]},{"name":"sensitivity","type":"string","enum":["public","internal","confidential","restricted"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Application-Collaboration', 'Application collaboration or interaction', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"collaboration_type","type":"string","enum":["synchronous","asynchronous","batch"]},{"name":"protocol","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Data Domain CI Types (7 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Data-DataObject', 'Data entity or document used by applications', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"data_classification","type":"string","enum":["public","internal","confidential","restricted"]},{"name":"retention_period","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-DataSet', 'Collection of related data objects', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"data_objects","type":"array"},{"name":"data_source","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-Repository', 'Data repository or database system', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"repository_type","type":"string","enum":["relational","document","key_value","graph","time_series"]},{"name":"capacity_gb","type":"integer"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-Structure', 'Data structure or schema definition', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"schema_version","type":"string"},{"name":"validation_rules","type":"array"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-Artifact', 'Data artifact or report', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"artifact_type","type":"string","enum":["report","dashboard","export","feed"]},{"name":"generation_frequency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-Representation', 'Data representation or format', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"format_type","type":"string","enum":["json","xml","csv","parquet","avro"]},{"name":"encoding","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Data-Metadata', 'Metadata describing data assets', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"metadata_type","type":"string","enum":["technical","business","operational"]},{"name":"data_element_id","type":"uuid"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Technology Domain CI Types (8 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Technology-ITComponent', 'Software component, library, or framework', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"version","type":"string"},{"name":"license","type":"string"},{"name":"end_of_support","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Platform', 'Technology platform or framework', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"components","type":"array"},{"name":"version","type":"string"},{"name":"platform_type","type":"string","enum":["application","data","integration","mobile"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Artifact', 'Technology artifact or deployment unit', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"artifact_type","type":"string","enum":["jar","war","dll","exe","docker_image","npm_package"]},{"name":"checksum","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Resource', 'Technology resource or capability', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"resource_type","type":"string","enum":["compute","storage","network","license"]},{"name":"capacity","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Capability', 'Technology capability or feature', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"maturity_level","type":"string","enum":["emerging","growing","mature","declining"]},{"name":"adoption_status","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Function', 'Technology function or API', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"function_type","type":"string","enum":["api","library","service","utility"]},{"name="signature","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Service', 'Technology service or utility', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"service_model","type":"string","enum":["saas","paas","iaas","on_premise"]},{"name="availability_sla","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Technology-Path', 'Technology path or communication channel', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"path_type","type":"string","enum":["network","bus","queue","topic"]},{"name":"bandwidth","type":"string"},{"name":"latency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Infrastructure Domain CI Types (8 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Infrastructure-Node', 'Infrastructure node (server, VM, container)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"node_type","type":"string","enum":["physical","virtual","container"]},{"name":"cpu_cores","type":"integer"},{"name":"memory_gb","type":"integer"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-Network', 'Network segment or VLAN', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"cidr_block","type":"string"},{"name":"network_type","type":"string","enum":["lan","wan","vlan","vpn"]},{"name":"bandwidth","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-Device', 'Infrastructure device (router, switch, firewall)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"device_type","type":"string","enum":["router","switch","firewall","load_balancer","proxy"]},{"name":"ip_address","type":"string"},{"name":"mac_address","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-Storage', 'Storage system or device', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"storage_type","type":"string","enum":["san","nas","object","block"]},{"name":"capacity_gb","type":"integer"},{"name":"throughput_mb_s","type":"integer"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-Cluster', 'Infrastructure cluster or group', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"cluster_type","type":"string","enum":["kubernetes","vmware","database"]},{"name":"node_count","type":"integer"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-SystemSoftware', 'System software (OS, hypervisor)', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"software_type","type":"string","enum":["os","hypervisor","container_runtime","firmware"]},{"name":"version","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-CommunicationPath', 'Communication path between infrastructure elements', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"protocol","type":"string"},{"name":"port","type":"string"},{"name":"encryption","type":"boolean"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Infrastructure-Capability', 'Infrastructure capability or feature', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"capability_type","type":"string","enum":["compute","storage","network","security"]},{"name":"provisioning","type":"string","enum":["manual","automated","dynamic"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Security Domain CI Types (6 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Security-Control', 'Security control or safeguard', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"control_type","type":"string","enum":["preventive","detective","corrective"]},{"name":"control_framework","type":"string"},{"name":"implementation_status","type":"string","enum":["planned","implemented","enforced","monitored"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Policy', 'Security policy document', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"policy_type","type":"string","enum":["access_control","data_protection","incident_response","compliance"]},{"name":"approval_date","type":"date"},{"name":"review_frequency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Risk', 'Security risk or threat', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"risk_level","type":"string","enum":["critical","high","medium","low"]},{"name":"likelihood","type":"string","enum":["rare","unlikely","possible","likely","certain"]},{"name":"impact","type":"string","enum":["negligible","minor","moderate","major","catastrophic"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Vulnerability', 'Security vulnerability', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"severity","type":"string","enum":["critical","high","medium","low"}]},{"name":"cve_id","type":"string"},{"name":"patch_status","type":"string","enum":["open","in_progress","patched","mitigated"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Assessment', 'Security assessment or audit', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"assessment_type","type":"string","enum":["vulnerability_scan","penetration_test","compliance_audit","risk_assessment"]},{"name":"assessment_date","type":"date"},{"name":"status","type":"string","enum":["planned","in_progress","completed"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Security-Requirement', 'Security requirement or standard', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"requirement_type","type":"string","enum":["functional","non_functional","compliance"]},{"name":"source_standard","type":"string"},{"name":"priority","type":"string","enum":["mandatory","recommended","optional"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- Governance Domain CI Types (7 types)
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes, created_by)
VALUES
    ('EA.Governance-Policy', 'Governance policy document', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"policy_scope","type":"string"},{"name":"effective_date","type":"date"},{"name":"compliance_level","type":"string","enum":["mandatory","recommended","guideline"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Compliance', 'Compliance requirement or regulation', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"regulation","type":"string"},{"name":"compliance_status","type":"string","enum":["compliant","non_compliant","in_progress","not_applicable"]},{"name":"last_audit_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Standard', 'Governance standard or best practice', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"standard_body","type":"string"},{"name":"standard_version","type":"string"},{"name":"adoption_level","type":"string","enum":["full","partial","planned"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Process', 'Governance process or procedure', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"process_type","type":"string","enum":["approval","review","audit","assessment"]},{"name":"frequency","type":"string"},{"name":"automated","type":"boolean"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Audit', 'Governance audit or review', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"audit_type","type":"string","enum":["internal","external","regulatory"]},{"name":"scheduled_date","type":"date"},{"name":"status","type":"string","enum":["planned","in_progress","completed"]}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Metric', 'Governance metric or KPI', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"metric_type","type":"string","enum":["quantitative","qualitative"]},{"name":"target_value","type":"string"},{"name":"current_value","type":"string"},{"name":"measurement_frequency","type":"string"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)),
    ('EA.Governance-Exception', 'Governance exception or waiver', '[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"owner","type":"string"}]'::jsonb, '[{"name":"exception_type","type":"string","enum":["temporary","permanent","conditional"]},{"name":"justification","type":"string"},{"name":"approval_date","type":"date"},{"name":"expiry_date","type":"date"}]'::jsonb, (SELECT id FROM users WHERE username = 'admin' LIMIT 1))
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SECTION 3: EA Relationship Type Definitions
-- ============================================================================

INSERT INTO relationship_types (name, description, forward_label, backward_label, source_types, target_types, cardinality_source, cardinality_target, bidirectional, attributes, created_by)
VALUES
-- Core ArchiMate relationships
(
    'supports',
    'Source supports or enables target functionality',
    'supports',
    'supported by',
    ARRAY['EA.Application-*', 'EA.Technology-*', 'EA.Infrastructure-*'],
    ARRAY['EA.Business-*', 'EA.Application-*'],
    'many',
    'many',
    true,
    '{"description": "Support relationship indicating source provides capability to target", "archimate_concept": "Serving", "examples": ["CRM Application supports Customer Management", "Java Platform supports Order Processing"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'depends_on',
    'Source depends on target for functionality or data',
    'depends on',
    'is depended on by',
    ARRAY['EA.Application-*', 'EA.Technology-*', 'EA.Business-*'],
    ARRAY['EA.Application-*', 'EA.Data-*', 'EA.Technology-*', 'EA.Infrastructure-*'],
    'many',
    'many',
    true,
    '{"description": "Dependency relationship where source requires target to function", "archimate_concept": "Dependency", "examples": ["Application depends on Database", "Service depends on API"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'realizes',
    'Source implements or realizes target',
    'realizes',
    'realized by',
    ARRAY['EA.Application-*', 'EA.Technology-*'],
    ARRAY['EA.Business-*', 'EA.Strategy-*'],
    'many',
    'many',
    false,
    '{"description": "Realization relationship where source provides implementation of target", "archimate_concept": "Realization", "examples": ["Application realizes Business Service", "Component realizes Business Function"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'flows_to',
    'Data or information flows from source to target',
    'flows to',
    'receives flow from',
    ARRAY['EA.Application-*', 'EA.Business-*', 'EA.Data-*'],
    ARRAY['EA.Application-*', 'EA.Data-*', 'EA.Business-*'],
    'many',
    'many',
    true,
    '{"description": "Flow relationship representing information exchange", "archimate_concept": "Flow", "examples": ["Data flows from Application to Database", "Information flows to Business Process"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'assigned_to',
    'Source is assigned to target (e.g., Role assigned to Actor)',
    'assigned to',
    'has assigned',
    ARRAY['EA.Business-*', 'EA.Application-*'],
    ARRAY['EA.Business-*', 'EA.Infrastructure-*'],
    'many',
    'many',
    false,
    '{"description": "Assignment relationship associating responsibility", "archimate_concept": "Assignment", "examples": ["Business Role assigned to Business Actor", "Application assigned to Business Unit"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'aggregates',
    'Source aggregates target (collection relationship)',
    'aggregates',
    'aggregated by',
    ARRAY['EA.Application-*', 'EA.Business-*', 'EA.Infrastructure-*'],
    ARRAY['EA.Application-*', 'EA.Business-*', 'EA.Infrastructure-*'],
    'one',
    'many',
    false,
    '{"description": "Aggregation relationship where source is a collection of target", "archimate_concept": "Aggregation", "examples": ["Application aggregates Modules", "Business Domain aggregates Capabilities"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'composes',
    'Source composes target (strong composition)',
    'composes',
    'composed of',
    ARRAY['EA.Application-*', 'EA.Technology-*'],
    ARRAY['EA.Application-*', 'EA.Technology-*'],
    'one',
    'many',
    false,
    '{"description": "Composition relationship with strong ownership", "archimate_concept": "Composition", "examples": ["Application composes Components", "Node composes Devices"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'accesses',
    'Source accesses target (typically data)',
    'accesses',
    'is accessed by',
    ARRAY['EA.Application-*', 'EA.Business-*'],
    ARRAY['EA.Data-*'],
    'many',
    'many',
    false,
    '{"description": "Access relationship for data/object access", "archimate_concept": "Access", "examples": ["Application accesses Database Table", "Process accesses Data Object"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'associated_with',
    'General association relationship',
    'associated with',
    'associated with',
    ARRAY['EA.*'],
    ARRAY['EA.*'],
    'many',
    'many',
    true,
    '{"description": "General association for relationships not covered by specific types", "archimate_concept": "Association", "examples": ["Risk associated with Control", "Policy associated with Process"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),

-- EA-specific relationships
(
    'deployed_on',
    'Source is deployed on target infrastructure',
    'deployed on',
    'hosts',
    ARRAY['EA.Application-*'],
    ARRAY['EA.Infrastructure-*', 'EA.Technology-*'],
    'many',
    'many',
    true,
    '{"description": "Deployment relationship showing where applications run", "examples": ["CRM Application deployed on Production Server", "Microservice deployed on Kubernetes Cluster"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'runs_on',
    'Source executes on target',
    'runs on',
    'executes',
    ARRAY['EA.Application-*', 'EA.Technology-*'],
    ARRAY['EA.Infrastructure-*', 'EA.Technology-*'],
    'many',
    'many',
    true,
    '{"description": "Execution relationship", "examples": ["Application runs on Server", "Component runs on Platform"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'uses',
    'Source uses target',
    'uses',
    'is used by',
    ARRAY['EA.Application-*', 'EA.Business-*'],
    ARRAY['EA.Technology-*', 'EA.Data-*', 'EA.Application-*'],
    'many',
    'many',
    true,
    '{"description": "Usage relationship", "examples": ["Application uses Library", "Process uses Service"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'implements',
    'Source implements target',
    'implements',
    'is implemented by',
    ARRAY['EA.Technology-*'],
    ARRAY['EA.Application-*', 'EA.Business-*'],
    'many',
    'many',
    false,
    '{"description": "Implementation relationship", "examples": ["Component implements Service", "Technology implements Capability"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'validates',
    'Source validates or checks target',
    'validates',
    'is validated by',
    ARRAY['EA.Security-*', 'EA.Governance-*'],
    ARRAY['EA.Security-*', 'EA.Governance-*'],
    'many',
    'many',
    false,
    '{"description": "Validation relationship for controls and assessments", "examples": ["Control validates Risk", "Assessment validates Control"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'mitigates',
    'Source mitigates target risk',
    'mitigates',
    'is mitigated by',
    ARRAY['EA.Security-*'],
    ARRAY['EA.Security-*'],
    'many',
    'many',
    false,
    '{"description": "Risk mitigation relationship", "examples": ["Control mitigates Risk", "Safeguard mitigates Threat"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'enforces',
    'Source enforces target policy',
    'enforces',
    'is enforced by',
    ARRAY['EA.Security-*', 'EA.Governance-*'],
    ARRAY['EA.Governance-*', 'EA.Security-*'],
    'many',
    'many',
    false,
    '{"description": "Policy enforcement relationship", "examples": ["Control enforces Policy", "Process enforces Standard"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'assesses',
    'Source assesses target',
    'assesses',
    'is assessed by',
    ARRAY['EA.Governance-*', 'EA.Security-*'],
    ARRAY['EA.Governance-*', 'EA.Security-*'],
    'many',
    'many',
    false,
    '{"description": "Assessment relationship", "examples": ["Audit assesses Control", "Review assesses Process"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'governs',
    'Source governs target',
    'governs',
    'is governed by',
    ARRAY['EA.Governance-*'],
    ARRAY['EA.Business-*', 'EA.Application-*', 'EA.Data-*'],
    'many',
    'many',
    false,
    '{"description": "Governance relationship", "examples": ["Policy governs Process", "Standard governs Application"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'aligned_with',
    'Source is aligned with target strategy',
    'aligned with',
    'is aligned with',
    ARRAY['EA.Business-*', 'EA.Application-*'],
    ARRAY['EA.Strategy-*'],
    'many',
    'many',
    true,
    '{"description": "Strategic alignment relationship", "examples": ["Capability aligned with Objective", "Application aligned with Goal"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'conforms_to',
    'Source conforms to target standard/policy',
    'conforms to',
    'is conformed to by',
    ARRAY['EA.Application-*', 'EA.Data-*', 'EA.Technology-*'],
    ARRAY['EA.Governance-*', 'EA.Security-*'],
    'many',
    'many',
    false,
    '{"description": "Compliance relationship", "examples": ["Application conforms to Policy", "Data conforms to Standard"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'derived_from',
    'Source is derived from target',
    'derived from',
    'is source of',
    ARRAY['EA.Data-*'],
    ARRAY['EA.Data-*'],
    'many',
    'many',
    false,
    '{"description": "Data derivation relationship", "examples": ["DataSet derived from DataObject", "Report derived from Transaction"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'decomposes',
    'Source decomposes into target components',
    'decomposes',
    'is component of',
    ARRAY['EA.Business-*', 'EA.Application-*'],
    ARRAY['EA.Business-*', 'EA.Application-*'],
    'one',
    'many',
    false,
    '{"description": "Decomposition relationship (e.g., Capability L1 decomposes into Capability L2)", "examples": ["Business Domain decomposes into Capabilities", "Capability L1 decomposes into Capability L2"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'triggers',
    'Source triggers target action',
    'triggers',
    'is triggered by',
    ARRAY['EA.Business-*', 'EA.Application-*'],
    ARRAY['EA.Business-*', 'EA.Application-*'],
    'many',
    'many',
    false,
    '{"description": "Event triggering relationship", "examples": ["Event triggers Process", "Message triggers Service"]}'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
)

ON CONFLICT (name) DO NOTHING;

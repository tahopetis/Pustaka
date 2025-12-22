-- Add relationship_types table and permissions

-- Relationship Types table
CREATE TABLE relationship_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    forward_label VARCHAR(100) NOT NULL,
    backward_label VARCHAR(100) NOT NULL,
    source_types TEXT[] DEFAULT '{}',
    target_types TEXT[] DEFAULT '{}',
    allow_same_type BOOLEAN DEFAULT false,
    attributes JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Indexes for relationship_types
CREATE INDEX idx_relationship_types_name ON relationship_types(name);
CREATE INDEX idx_relationship_types_source_types ON relationship_types USING GIN(source_types);
CREATE INDEX idx_relationship_types_target_types ON relationship_types USING GIN(target_types);
CREATE INDEX idx_relationship_types_attributes ON relationship_types USING GIN(attributes);

-- Update timestamp trigger for relationship_types
CREATE TRIGGER update_relationship_types_updated_at BEFORE UPDATE ON relationship_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add relationship_type permissions
INSERT INTO permissions (name, description, resource_type) VALUES
('relationship_type:create', 'Create relationship types', 'relationship_type'),
('relationship_type:read', 'Read relationship types', 'relationship_type'),
('relationship_type:update', 'Update relationship types', 'relationship_type'),
('relationship_type:delete', 'Delete relationship types', 'relationship_type');

-- Grant relationship_type permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name LIKE 'relationship_type:%';

-- Grant relationship_type read permissions to editor role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'editor' AND p.name = 'relationship_type:read';

-- Grant relationship_type read permissions to viewer role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'relationship_type:read';

-- Insert default relationship types
INSERT INTO relationship_types (name, description, forward_label, backward_label, source_types, target_types, allow_same_type, attributes, created_by) VALUES
(
    'depends_on',
    'Dependency relationship where source depends on target',
    'depends on',
    'is depended on by',
    ARRAY['Server', 'Application', 'Database'],
    ARRAY['Server', 'Application', 'Database'],
    false,
    '{
        "description": "Indicates that the source CI requires the target CI to function properly",
        "examples": ["Application depends on Database", "Service depends on Server"]
    }'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'runs_on',
    'Hosted on relationship where source runs on target',
    'runs on',
    'hosts',
    ARRAY['Application', 'Database'],
    ARRAY['Server'],
    false,
    '{
        "description": "Indicates that the source CI is deployed and running on the target CI",
        "examples": ["Application runs on Server", "Database runs on Server"]
    }'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'connects_to',
    'Network connection relationship',
    'connects to',
    'is connected from',
    ARRAY['Server', 'Application', 'Database'],
    ARRAY['Server', 'Application', 'Database'],
    true,
    '{
        "description": "Indicates network connectivity between components",
        "examples": ["Web Server connects to Database", "API connects to Cache"]
    }'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'manages',
    'Management relationship where source manages target',
    'manages',
    'is managed by',
    ARRAY['Application', 'Server'],
    ARRAY['Server', 'Application', 'Database'],
    false,
    '{
        "description": "Indicates that the source CI has administrative control over the target CI",
        "examples": ["Management Server manages Application", "Monitoring Tool manages Server"]
    }'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
),
(
    'backed_up_by',
    'Backup relationship where source is backed up by target',
    'is backed up by',
    'backs up',
    ARRAY['Server', 'Application', 'Database'],
    ARRAY['Server', 'Application'],
    false,
    '{
        "description": "Indicates backup relationship between components",
        "examples": ["Database is backed up by Backup Server", "Application is backed up by Storage System"]
    }'::jsonb,
    (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
);
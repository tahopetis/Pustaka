-- Initial schema for Pustaka CMDB
-- Supports FSD-compliant flexible JSONB attributes

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Role-Permission mapping
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role mapping
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- CI Type Definitions (FSD-compliant schema management)
CREATE TABLE ci_type_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    required_attributes JSONB NOT NULL DEFAULT '[]',
    optional_attributes JSONB NOT NULL DEFAULT '[]',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Configuration Items (FSD-compliant with flexible JSONB attributes)
CREATE TABLE configuration_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    ci_type VARCHAR(100) NOT NULL REFERENCES ci_type_definitions(name),
    attributes JSONB NOT NULL DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Constraints
    CONSTRAINT unique_name_per_type UNIQUE (name, ci_type)
);

-- Relationships table (FSD-compliant with flexible JSONB attributes)
CREATE TABLE relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL,
    attributes JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Constraints
    CONSTRAINT no_self_relationship CHECK (source_id != target_id),
    CONSTRAINT unique_relationship UNIQUE (source_id, target_id, relationship_type)
);

-- Audit Logs (comprehensive audit trail)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    action VARCHAR(50) NOT NULL,
    performed_by UUID REFERENCES users(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    details JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT
);

-- Indexes for performance
-- Users indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);

-- Configuration Items indexes
CREATE INDEX idx_cis_type ON configuration_items(ci_type);
CREATE INDEX idx_cis_name ON configuration_items(name);
CREATE INDEX idx_cis_tags ON configuration_items USING GIN(tags);
CREATE INDEX idx_cis_attributes ON configuration_items USING GIN(attributes);
CREATE INDEX idx_cis_created_at ON configuration_items(created_at);
CREATE INDEX idx_cis_created_by ON configuration_items(created_by);

-- Relationships indexes
CREATE INDEX idx_relationships_source ON relationships(source_id);
CREATE INDEX idx_relationships_target ON relationships(target_id);
CREATE INDEX idx_relationships_type ON relationships(relationship_type);
CREATE INDEX idx_relationships_created_at ON relationships(created_at);

-- Audit Logs indexes
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_performed_by ON audit_logs(performed_by);
CREATE INDEX idx_audit_action ON audit_logs(action);

-- CI Type Definitions indexes
CREATE INDEX idx_ci_type_names ON ci_type_definitions(name);

-- Full-text search index for CI names
CREATE INDEX idx_cis_name_fulltext ON configuration_items USING GIN(to_tsvector('english', name));

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Update timestamp triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_type_definitions_updated_at BEFORE UPDATE ON ci_type_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_configuration_items_updated_at BEFORE UPDATE ON configuration_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_relationships_updated_at BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles and permissions
INSERT INTO roles (name, description) VALUES
('admin', 'Full system access'),
('editor', 'CI and relationship management'),
('viewer', 'Read-only access');

INSERT INTO permissions (name, description, resource_type) VALUES
-- CI permissions
('ci:create', 'Create configuration items', 'ci'),
('ci:read', 'Read configuration items', 'ci'),
('ci:update', 'Update configuration items', 'ci'),
('ci:delete', 'Delete configuration items', 'ci'),

-- CI Type permissions
('ci_type:create', 'Create CI type definitions', 'ci_type'),
('ci_type:read', 'Read CI type definitions', 'ci_type'),
('ci_type:update', 'Update CI type definitions', 'ci_type'),
('ci_type:delete', 'Delete CI type definitions', 'ci_type'),

-- Relationship permissions
('relationship:create', 'Create relationships', 'relationship'),
('relationship:read', 'Read relationships', 'relationship'),
('relationship:update', 'Update relationships', 'relationship'),
('relationship:delete', 'Delete relationships', 'relationship'),

-- User management permissions
('user:create', 'Create users', 'user'),
('user:read', 'Read users', 'user'),
('user:update', 'Update users', 'user'),
('user:delete', 'Delete users', 'user'),

-- Audit permissions
('audit:read', 'Read audit logs', 'audit'),

-- System permissions
('system:admin', 'Full system administration', 'system');

-- Assign permissions to roles
-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';

-- Editor gets CI, relationship, and audit permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'editor' AND p.resource_type IN ('ci', 'ci_type', 'relationship', 'audit')
AND p.name LIKE '%:read%';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'editor' AND p.name IN ('ci:create', 'ci:update', 'ci:delete', 'relationship:create', 'relationship:update', 'relationship:delete');

-- Viewer gets read permissions only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name LIKE '%:read%';

-- Insert default CI type definitions
INSERT INTO ci_type_definitions (name, description, required_attributes, optional_attributes) VALUES
(
    'Server',
    'Physical or virtual server',
    '[
        {
            "name": "hostname",
            "type": "string",
            "description": "Server hostname",
            "validation": {
                "pattern": "^[a-zA-Z0-9-]+$",
                "min_length": 1,
                "max_length": 253
            }
        },
        {
            "name": "ip_address",
            "type": "string",
            "description": "Primary IP address",
            "validation": {
                "pattern": "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
            }
        },
        {
            "name": "os",
            "type": "string",
            "description": "Operating system",
            "validation": {
                "enum": ["RHEL", "Ubuntu", "CentOS", "Debian", "Windows Server", "Other"]
            }
        }
    ]'::jsonb,
    '[
        {
            "name": "vendor",
            "type": "string",
            "description": "Hardware vendor",
            "validation": {
                "max_length": 100
            }
        },
        {
            "name": "rack_location",
            "type": "string",
            "description": "Data center rack location",
            "validation": {
                "max_length": 50
            }
        },
        {
            "name": "cpu_cores",
            "type": "integer",
            "description": "Number of CPU cores",
            "validation": {
                "min": 1,
                "max": 256
            }
        },
        {
            "name": "memory_gb",
            "type": "integer",
            "description": "Memory in GB",
            "validation": {
                "min": 1,
                "max": 1024
            }
        }
    ]'::jsonb
),
(
    'Application',
    'Software application or service',
    '[
        {
            "name": "name",
            "type": "string",
            "description": "Application name",
            "validation": {
                "min_length": 1,
                "max_length": 100
            }
        },
        {
            "name": "version",
            "type": "string",
            "description": "Application version",
            "validation": {
                "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$"
            }
        }
    ]'::jsonb,
    '[
        {
            "name": "repository_url",
            "type": "string",
            "description": "Source code repository URL",
            "validation": {
                "format": "url"
            }
        },
        {
            "name": "language",
            "type": "string",
            "description": "Programming language",
            "validation": {
                "enum": ["Go", "Python", "Java", "JavaScript", "TypeScript", "C#", "Ruby", "PHP", "Other"]
            }
        }
    ]'::jsonb
),
(
    'Database',
    'Database instance or cluster',
    '[
        {
            "name": "name",
            "type": "string",
            "description": "Database name",
            "validation": {
                "min_length": 1,
                "max_length": 100
            }
        },
        {
            "name": "type",
            "type": "string",
            "description": "Database type",
            "validation": {
                "enum": ["PostgreSQL", "MySQL", "MongoDB", "Redis", "Neo4j", "Other"]
            }
        }
    ]'::jsonb,
    '[
        {
            "name": "version",
            "type": "string",
            "description": "Database version",
            "validation": {
                "max_length": 50
            }
        },
        {
            "name": "size_gb",
            "type": "integer",
            "description": "Database size in GB",
            "validation": {
                "min": 0
            }
        }
    ]'::jsonb
);
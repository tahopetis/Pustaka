-- Migration: Add lifecycle statuses for CI management
-- This migration adds lifecycle status management to the CMDB system

-- Create lifecycle_statuses table
CREATE TABLE lifecycle_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,              -- Internal identifier (snake_case)
    display_name VARCHAR(100) NOT NULL,             -- User-friendly label
    description TEXT,
    color VARCHAR(7),                               -- Hex color code
    icon VARCHAR(50),                               -- Icon name for UI
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,                 -- Soft delete
    is_system BOOLEAN DEFAULT false,                -- Prevent deletion of defaults
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

-- Create indexes for better performance
CREATE INDEX idx_lifecycle_statuses_name ON lifecycle_statuses(name);
CREATE INDEX idx_lifecycle_statuses_active ON lifecycle_statuses(is_active);
CREATE INDEX idx_lifecycle_statuses_sort_order ON lifecycle_statuses(sort_order);

-- Create trigger for updated_at
CREATE TRIGGER update_lifecycle_statuses_updated_at
    BEFORE UPDATE ON lifecycle_statuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add lifecycle_status_id to configuration_items table
ALTER TABLE configuration_items
    ADD COLUMN lifecycle_status_id UUID REFERENCES lifecycle_statuses(id);

-- Create index for CI queries
CREATE INDEX idx_cis_lifecycle_status ON configuration_items(lifecycle_status_id);

-- Seed default lifecycle statuses
-- Note: created_by will be set later when admin user is created via application startup
INSERT INTO lifecycle_statuses (name, display_name, description, color, icon, sort_order, is_system) VALUES
('planned', 'Planned', 'CI is planned but not yet acquired', '#94a3b8', 'calendar', 10, true),
('on_order', 'On Order', 'CI has been ordered and is awaiting delivery', '#3b82f6', 'package', 20, true),
('in_stock', 'In Stock', 'CI is available in inventory/stock', '#10b981', 'archive', 30, true),
('pending_install', 'Pending Install', 'CI is ready and scheduled for installation', '#f59e0b', 'clock', 40, true),
('operational', 'Operational', 'CI is in normal operation', '#22c55e', 'check-circle', 50, true),
('in_maintenance', 'In Maintenance', 'CI is temporarily out of service for maintenance', '#f97316', 'wrench', 60, true),
('defective_repair', 'Defective/Repair', 'CI is defective and requires repair', '#ef4444', 'alert-triangle', 70, true),
('retired', 'Retired', 'CI is no longer in service but preserved', '#6b7280', 'power-off', 80, true),
('disposed', 'Disposed', 'CI has been permanently disposed', '#4b5563', 'trash-2', 90, true),
('missing_stolen', 'Missing/Stolen', 'CI is missing or has been stolen', '#991b1b', 'x-circle', 100, true);

-- Set existing CIs to 'operational' status if they don't have a status
UPDATE configuration_items
SET lifecycle_status_id = (
    SELECT id FROM lifecycle_statuses WHERE name = 'operational'
)
WHERE lifecycle_status_id IS NULL;

-- Add RBAC permissions for lifecycle status management
INSERT INTO permissions (name, description)
VALUES
    ('lifecycle_status:create', 'Create lifecycle statuses'),
    ('lifecycle_status:read', 'Read lifecycle statuses'),
    ('lifecycle_status:update', 'Update lifecycle statuses'),
    ('lifecycle_status:delete', 'Delete lifecycle statuses')
ON CONFLICT (name) DO NOTHING;

-- Assign all lifecycle status permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
AND p.name LIKE 'lifecycle_status:%'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign read permission to editor and viewer roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name IN ('editor', 'viewer')
AND p.name = 'lifecycle_status:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
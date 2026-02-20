-- EA Permissions Migration
-- Adds EA-specific permissions for entity management

-- Insert EA permissions
INSERT INTO permissions (name, description, resource_type) VALUES
('ea:read', 'Read EA entities', 'ea'),
('ea:create', 'Create EA entities', 'ea'),
('ea:update', 'Update EA entities', 'ea'),
('ea:delete', 'Delete EA entities', 'ea')
ON CONFLICT (name) DO NOTHING;

-- Grant EA permissions to admin role (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN ('ea:read', 'ea:create', 'ea:update', 'ea:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Grant EA read permission to viewer role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'viewer'
  AND p.name = 'ea:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Grant EA create and update permissions to editor role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'editor'
  AND p.name IN ('ea:read', 'ea:create', 'ea:update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

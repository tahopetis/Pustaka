-- ============================================================================
-- Migration Rollback: 009_add_ea_metamodel
-- Purpose: Remove EA metamodel (CI types, relationship types, teams, permissions)
-- Updated: 2026-02-21 (Corrected for 32 CI types)
-- ============================================================================

-- Remove EA RBAC permissions
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name LIKE 'ea:%'
);

DELETE FROM permissions
WHERE name LIKE 'ea:%';

-- Remove EA relationship types (all relationship types created by this migration)
DELETE FROM relationship_types
WHERE name IN (
    'drives', 'consists_of', 'contains', 'belongs_to', 'has',
    'defines', 'targets', 'changes', 'supports', 'uses',
    'owns', 'responsible_for', 'exposes', 'consumes', 'routes_to',
    'depends_on', 'provides', 'realizes', 'deployed_on', 'provided_by',
    'implements', 'enforces', 'complies_with', 'documented_in', 'governs'
);

-- Remove EA CI types (all 32 EA types from metamodel)
DELETE FROM ci_type_definitions
WHERE name LIKE 'EA.%';

-- Drop ea_teams table (CASCADE will remove relationships and constraints)
DROP TABLE IF EXISTS ea_teams CASCADE;

-- ============================================================================
-- Rollback complete
-- ============================================================================

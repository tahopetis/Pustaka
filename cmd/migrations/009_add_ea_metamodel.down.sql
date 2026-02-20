-- ============================================================================
-- Migration Rollback: 009_add_ea_metamodel
-- Purpose: Remove EA metamodel (CI types, relationship types, teams, permissions)
-- ============================================================================

-- Remove EA RBAC permissions
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name LIKE 'ea:%'
);

DELETE FROM permissions
WHERE name LIKE 'ea:%';

-- Remove EA relationship types
DELETE FROM relationship_types
WHERE name IN (
    'supports', 'depends_on', 'realizes', 'flows_to', 'assigned_to',
    'aggregates', 'composes', 'accesses', 'associated_with',
    'deployed_on', 'runs_on', 'uses', 'implements', 'validates',
    'mitigates', 'enforces', 'assesses', 'governs', 'aligned_with',
    'conforms_to', 'derived_from', 'decomposes', 'triggers'
);

-- Remove EA CI types
DELETE FROM ci_type_definitions
WHERE name LIKE 'EA.%';

-- Drop ea_teams table (CASCADE will remove relationships and constraints)
DROP TABLE IF EXISTS ea_teams CASCADE;

-- ============================================================================
-- Rollback complete
-- ============================================================================

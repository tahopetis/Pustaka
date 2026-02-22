-- ============================================================================
-- Migration Rollback: 011_add_ea_lifecycle_statuses
-- Purpose: Remove EA-specific lifecycle statuses
-- ============================================================================

-- Remove EA-specific lifecycle statuses
DELETE FROM lifecycle_statuses WHERE name IN ('proposed', 'active', 'deprecated', 'archived');

-- Note: We keep 'retired' as it's used by both EA and inventory CIs
-- If you need to remove it entirely, uncomment the line below:
-- DELETE FROM lifecycle_statuses WHERE name = 'retired';

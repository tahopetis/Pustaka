-- ============================================================================
-- Migration: 011_add_ea_lifecycle_statuses
-- Created: 2026-02-22
-- Purpose: Add EA-specific lifecycle statuses for EA entity management
-- ============================================================================

-- Add EA-appropriate lifecycle statuses
-- These statuses are designed for Enterprise Architecture entities rather than inventory CIs
INSERT INTO lifecycle_statuses (name, display_name, description, color, icon, sort_order, is_system, is_active) VALUES
('proposed', 'Proposed', 'EA entity is proposed and under review', '#94a3b8', 'lightbulb', 10, true, true),
('active', 'Active', 'EA entity is active and in use', '#22c55e', 'check-circle', 20, true, true),
('deprecated', 'Deprecated', 'EA entity is deprecated but still in use', '#f59e0b', 'alert-triangle', 30, true, true),
('retired', 'Retired', 'EA entity is retired and no longer in use', '#6b7280', 'power-off', 40, true, true),
('archived', 'Archived', 'EA entity is archived for historical reference', '#4b5563', 'archive', 50, true, true)
ON CONFLICT (name) DO NOTHING;

-- Add comment to document the purpose of these statuses
COMMENT ON TABLE lifecycle_statuses IS 'Lifecycle statuses for both inventory CIs and EA entities. EA statuses: proposed, active, deprecated, retired, archived. Inventory statuses: planned, on_order, in_stock, pending_install, operational, in_maintenance, defective_repair, retired, disposed, missing_stolen';

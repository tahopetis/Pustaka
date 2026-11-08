CI Lifecycle Management Implementation Plan

 Overview

 Add lifecycle status management to Pustaka CMDB with:
 - Universal lifecycle statuses (10 default statuses: Planned, On Order, In Stock, Pending Install, Operational, In Maintenance,
 Defective/Repair, Retired, Disposed, Missing/Stolen)
 - Admin CRUD interface for managing status definitions
 - Status field on all CIs (optional, no transition rules required)
 - Audit logging via existing audit_logs table

 Pattern: Follow relationship_types implementation as reference.

 ---
 Phase 1: Database Schema (Migration 002)

 File: /home/tahopetis/dev/pustaka/cmd/migrations/002_add_lifecycle_statuses.sql

 Create lifecycle_statuses table:
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

 CREATE INDEX idx_lifecycle_statuses_name ON lifecycle_statuses(name);
 CREATE INDEX idx_lifecycle_statuses_active ON lifecycle_statuses(is_active);
 CREATE INDEX idx_lifecycle_statuses_sort_order ON lifecycle_statuses(sort_order);

 CREATE TRIGGER update_lifecycle_statuses_updated_at
     BEFORE UPDATE ON lifecycle_statuses
     FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

 Add status field to CIs:
 ALTER TABLE configuration_items
     ADD COLUMN lifecycle_status_id UUID REFERENCES lifecycle_statuses(id);

 CREATE INDEX idx_cis_lifecycle_status ON configuration_items(lifecycle_status_id);

 Seed default statuses (10 statuses with appropriate colors/icons/sort order)

 Add RBAC permissions:
 - lifecycle_status:create, lifecycle_status:read, lifecycle_status:update, lifecycle_status:delete
 - Assign all to admin role, read to editor/viewer roles

 Set default for existing CIs: Update to 'operational' status

 ---
 Phase 2: Backend Implementation

 2.1 Models

 File: /home/tahopetis/dev/pustaka/internal/ci/lifecycle_status.go (NEW)

 Define structs:
 - LifecycleStatus (ID, Name, DisplayName, Description, Color, Icon, SortOrder, IsActive, IsSystem, timestamps,
 created_by/updated_by)
 - CreateLifecycleStatusRequest
 - UpdateLifecycleStatusRequest
 - LifecycleStatusListResponse
 - ListLifecycleStatusFilters
 - LifecycleStatusUsage

 2.2 Repository

 File: /home/tahopetis/dev/pustaka/internal/ci/lifecycle_status_repository.go (NEW)

 Implement methods:
 - Create, GetByID, GetByName, List, Update, Delete
 - GetActive (for CI forms)
 - GetUsageStats (usage counts per status)
 - CountCIsWithStatus (check before delete)

 Pattern: Follow internal/ci/relationship_types_repository.go

 2.3 Service

 File: /home/tahopetis/dev/pustaka/internal/ci/lifecycle_status_service.go (NEW)

 Dependencies: repo, ciRepo, redis, auditService, logger

 Methods:
 - CreateLifecycleStatus - validate uniqueness, log audit, cache
 - GetLifecycleStatus - check cache first, fallback to DB
 - ListLifecycleStatuses - with filters (search, is_active, is_system)
 - UpdateLifecycleStatus - prevent name change for system statuses, log audit, invalidate cache
 - DeleteLifecycleStatus - check is_system, check CI usage, log audit, invalidate cache
 - GetActiveLifecycleStatuses - cached list for CI forms
 - GetLifecycleStatusUsage - usage statistics

 Caching:
 - Individual: lifecycle_status:{id} (5 min TTL)
 - Active list: lifecycle_statuses:active (5 min TTL)

 Validation:
 - Cannot delete system status
 - Cannot delete if in use by CIs

 Pattern: Follow internal/ci/relationship_types_service.go

 2.4 HTTP Handlers

 File: /home/tahopetis/dev/pustaka/internal/api/handlers/lifecycle_status_handlers.go (NEW)

 Implement handlers:
 - CreateLifecycleStatus (POST /api/v1/lifecycle-statuses)
 - GetLifecycleStatus (GET /api/v1/lifecycle-statuses/{id})
 - ListLifecycleStatuses (GET /api/v1/lifecycle-statuses)
 - UpdateLifecycleStatus (PUT /api/v1/lifecycle-statuses/{id})
 - DeleteLifecycleStatus (DELETE /api/v1/lifecycle-statuses/{id})
 - GetActiveLifecycleStatuses (GET /api/v1/lifecycle-statuses/active)
 - GetLifecycleStatusUsage (GET /api/v1/lifecycle-statuses/usage)

 Error handling: 400 (validation), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict/in use), 500 (server error)

 Pattern: Follow internal/api/handlers/relationship_types.go

 2.5 Update CI Models

 File: /home/tahopetis/dev/pustaka/internal/ci/models.go (MODIFY)

 Add to ConfigurationItem:
 - LifecycleStatusID *uuid.UUID
 - LifecycleStatus *LifecycleStatus (from JOIN, not stored)

 Add to CreateCIRequest:
 - LifecycleStatusID *uuid.UUID (optional)

 Add to UpdateCIRequest:
 - LifecycleStatusID *uuid.UUID (optional)

 Add to ListCIFilters:
 - LifecycleStatusID *uuid.UUID (filter by status)

 2.6 Update CI Repository

 File: /home/tahopetis/dev/pustaka/internal/ci/repository.go (MODIFY)

 Changes:
 1. CreateCI - add lifecycle_status_id to INSERT
 2. GetCI - LEFT JOIN with lifecycle_statuses table
 3. ListCIs - add lifecycle_status_id filter support
 4. UpdateCI - support updating lifecycle_status_id
 5. Add CountCIsWithStatus(ctx, statusID) method

 2.7 Update CI Service for Audit

 File: /home/tahopetis/dev/pustaka/internal/ci/service.go (MODIFY)

 In UpdateCI method, track status changes in audit details:
 if req.LifecycleStatusID != nil && (current.LifecycleStatusID == nil || *current.LifecycleStatusID != *req.LifecycleStatusID) {
     details["old_lifecycle_status_id"] = current.LifecycleStatusID
     details["new_lifecycle_status_id"] = req.LifecycleStatusID
 }

 2.8 Add Routing

 File: /home/tahopetis/dev/pustaka/cmd/api/main.go (MODIFY)

 Initialize service in main():
 lifecycleStatusRepo := ci.NewLifecycleStatusRepository(postgresDB.Pool)
 lifecycleStatusService := ci.NewLifecycleStatusService(lifecycleStatusRepo, ciRepo, redisDB.Client, auditService, logger)
 lifecycleStatusHandlers := handlers.NewLifecycleStatusHandler(lifecycleStatusService, rbacService, logger)

 Add routes in setupRouter():
 r.Route("/lifecycle-statuses", func(r chi.Router) {
     r.Use(middleware.RBAC("lifecycle_status:read"))
     r.Get("/", lifecycleStatusHandlers.ListLifecycleStatuses)
     r.Get("/active", lifecycleStatusHandlers.GetActiveLifecycleStatuses)
     r.Get("/usage", lifecycleStatusHandlers.GetLifecycleStatusUsage)
     r.Get("/{id}", lifecycleStatusHandlers.GetLifecycleStatus)

     r.With(middleware.RBAC("lifecycle_status:create")).Post("/", lifecycleStatusHandlers.CreateLifecycleStatus)
     r.With(middleware.RBAC("lifecycle_status:update")).Put("/{id}", lifecycleStatusHandlers.UpdateLifecycleStatus)
     r.With(middleware.RBAC("lifecycle_status:delete")).Delete("/{id}", lifecycleStatusHandlers.DeleteLifecycleStatus)
 })

 ---
 Phase 3: Frontend Implementation

 3.1 TypeScript Types

 File: /home/tahopetis/dev/pustaka/web/src/types/lifecycle.ts (NEW)

 Define interfaces:
 - LifecycleStatus
 - CreateLifecycleStatusRequest
 - UpdateLifecycleStatusRequest
 - LifecycleStatusListResponse
 - LifecycleStatusUsage

 3.2 API Service

 File: /home/tahopetis/dev/pustaka/web/src/services/api.ts (MODIFY)

 Add lifecycleStatusAPI with methods:
 - list(params), get(id), create(data), update(id, data), delete(id)
 - getActive(), getUsage()

 3.3 Pinia Store

 File: /home/tahopetis/dev/pustaka/web/src/stores/lifecycleStatus.ts (NEW)

 State:
 - lifecycleStatuses, activeLifecycleStatuses, currentLifecycleStatus, loading, error

 Actions:
 - listLifecycleStatuses, getLifecycleStatus, createLifecycleStatus
 - updateLifecycleStatus, deleteLifecycleStatus
 - getActiveLifecycleStatuses, getLifecycleStatusUsage
 - clearError, clearCurrent

 Error handling: 409 (conflict), 404 (not found), 422 (validation), 500 (server error)

 Pattern: Follow web/src/stores/relationshipTypes.ts

 3.4 Admin Management View

 File: /home/tahopetis/dev/pustaka/web/src/views/lifecycle/LifecycleStatusManagementView.vue (NEW)

 Features:
 - Search bar and Create button
 - Table with columns:
   - Status (badge with color/icon)
   - Name (internal identifier)
   - Description
   - Usage Count
   - Active/Inactive badge
   - System badge
   - Actions (Edit, Delete - disabled for system statuses or in-use statuses)
 - Pagination controls
 - Create/Edit modal
 - Delete confirmation

 Pattern: Follow web/src/views/ci/CITypeManagementView.vue

 3.5 Modal Component

 File: /home/tahopetis/dev/pustaka/web/src/components/lifecycle/LifecycleStatusModal.vue (NEW)

 Form fields:
 - Name (required, unique, disabled on edit)
 - Display Name (required)
 - Description (optional, textarea)
 - Color (color picker, hex validation)
 - Icon (text input)
 - Sort Order (number, min 0)
 - Active (checkbox, edit only)

 Validation: name format (lowercase, numbers, underscores), color format (#RRGGBB)

 3.6 Update CI Forms

 Files: CI create/edit forms in /home/tahopetis/dev/pustaka/web/src/views/ci/ (MODIFY)

 Add lifecycle status dropdown:
 - Load active statuses on mount
 - Optional field (nullable)
 - Display with color/icon if available

 3.7 Update CI Detail/List Views

 Files: CI detail/list views (MODIFY)

 Display status:
 - Badge with color and icon
 - Show status in CI list table
 - Add status filter to CI list page

 3.8 Update Router

 File: /home/tahopetis/dev/pustaka/web/src/router/index.ts (MODIFY)

 Add route:
 {
   path: '/lifecycle-statuses',
   name: 'lifecycle-status-management',
   component: () => import('@/views/lifecycle/LifecycleStatusManagementView.vue'),
   meta: {
     requiresAuth: true,
     requiredPermission: 'lifecycle_status:read',
     title: 'Lifecycle Status Management'
   }
 }

 3.9 Update Navigation

 Add menu item for "Lifecycle Statuses" with permission check (lifecycle_status:read)

 ---
 Key Design Decisions

 1. Universal statuses: All CI types share same lifecycle statuses (simpler than per-type)
 2. No transition rules: Admins can set any status to any value (no state machine)
 3. Optional status field: CIs can exist without a status (nullable lifecycle_status_id)
 4. System status protection: Default 10 statuses have is_system=true to prevent deletion
 5. Soft delete: is_active flag hides statuses from forms but preserves data
 6. Usage protection: Cannot delete status in use by CIs
 7. Audit via existing logs: Use existing audit_logs table (no separate history table)

 ---
 Critical Files to Create/Modify

 New Files (12):

 1. cmd/migrations/002_add_lifecycle_statuses.sql
 2. internal/ci/lifecycle_status.go
 3. internal/ci/lifecycle_status_repository.go
 4. internal/ci/lifecycle_status_service.go
 5. internal/api/handlers/lifecycle_status_handlers.go
 6. web/src/types/lifecycle.ts
 7. web/src/stores/lifecycleStatus.ts
 8. web/src/views/lifecycle/LifecycleStatusManagementView.vue
 9. web/src/components/lifecycle/LifecycleStatusModal.vue

 Modified Files (7):

 1. internal/ci/models.go - Add LifecycleStatus fields to CI structs
 2. internal/ci/repository.go - Update queries with JOIN and filter
 3. internal/ci/service.go - Track status changes in audit
 4. cmd/api/main.go - Initialize service and routes
 5. web/src/services/api.ts - Add lifecycleStatusAPI
 6. web/src/router/index.ts - Add route
 7. CI form/detail views - Add status dropdown and display

 ---
 Implementation Sequence

 1. Database first: Run migration, verify schema and seed data
 2. Backend: Models → Repository → Service → Handlers → Routes
 3. Frontend: Types → API service → Store → Components → Views
 4. Integration: Test full flow (create status → assign to CI → filter by status)
 5. Testing: Unit tests (service), integration tests (API), E2E tests

 ---
 Testing Checklist

 - Create lifecycle status (admin)
 - Update lifecycle status (admin)
 - Delete unused lifecycle status (admin)
 - Cannot delete system status
 - Cannot delete status in use
 - Soft delete (is_active=false) hides from CI forms
 - Create CI with lifecycle status
 - Update CI lifecycle status
 - Filter CIs by lifecycle status
 - View audit log for status changes
 - Pagination in admin view
 - Search in admin view
 - RBAC permissions enforced

 ---
 Notes

 - Migration file must handle existing CIs (set to 'operational' or leave NULL)
 - Use Redis caching for performance (individual status + active list)
 - Follow relationship_types pattern for consistency
 - All operations must log to audit_logs
 - Frontend should handle 409 errors gracefully (duplicate name, in-use deletion)

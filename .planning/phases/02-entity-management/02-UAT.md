---
status: complete
phase: 02-entity-management
source: 02-01-SUMMARY.md, 02-02-SUMMARY.md, 02-03-SUMMARY.md, 02-04-SUMMARY.md, 02-05-SUMMARY.md, 02-06-SUMMARY.md, 02-07-SUMMARY.md
started: 2026-02-22T11:43:00Z
updated: 2026-02-22T13:50:00Z
---

## Current Test

[testing complete]

## Tests

### 1. EA Entity CRUD Endpoints
expected: Navigate to http://localhost:8080/api/v1/ea/entities health check. All 5 EA entity CRUD endpoints should be accessible: POST /api/v1/ea/entities (create), GET /api/v1/ea/entities (list), GET /api/v1/ea/entities/{id} (get single), PUT /api/v1/ea/entities/{id} (update), DELETE /api/v1/ea/entities/{id} (delete). Each endpoint should return proper HTTP status codes (201, 200, 404, 422, 204) and be protected by RBAC middleware (ea:read, ea:create, ea:update, ea:delete).
result: issue
reported: "API endpoints exist and are accessible but have issues: 1) GET /api/v1/ea/entities returns empty list correctly, 2) GET /api/v1/ea/ci-types returns all 32 EA CI types correctly, 3) POST /api/v1/ea/entities fails because 'owner' field must match EA team name exactly (e.g., 'business-architecture' not 'Business Architecture'), 4) Lifecycle status names in database are 'planned', 'on_order', 'in_stock' - no 'Active' status exists for EA entities, 5) PUT/DELETE untested due to no entity created successfully"
severity: major

### 2. EA Entity List View with AG Grid
expected: Navigate to http://localhost:3000/entities/business. Should see ag-grid table with domain sidebar showing 8 EA domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance). AG Grid should load without error #272 (modules registered via ModuleRegistry.registerModules([AllCommunityModule])). Table should have server-side pagination (25/50/100), sortable columns, filters, search box, and bulk actions bar when rows selected.
result: pass

### 3. EA CI Types API Endpoint
expected: Navigate to http://localhost:8080/api/v1/ea/ci-types with valid JWT token. Should return array of 32 EA CI types with EA. prefix (EA.Application-*, EA.Business-*, etc.). Response should include CI type definitions with required_attributes and optional_attributes schemas. All 8 EA domains should be represented. Endpoint should be protected by ea:read RBAC permission.
result: pass

### 4. EA Entity Create with Validation
expected: Navigate to http://localhost:3000/entities/business/create. Form should load with CI Type dropdown populated with 32 EA CI types (from /api/v1/ea/ci-types). Fill in required fields (Name, CI Type, Domain), select Lifecycle Status, add Owner/Team. Submit form. Entity should be created with data_quality_score calculated. Validation errors should display if required fields missing or attribute values don't match CI type schema (string, integer, boolean, date, array, object types enforced).
result: issue
reported: "CI Type dropdown shows only 'Select CI Type' placeholder - the 32 EA CI types from /api/v1/ea/ci-types are not loading into dropdown. Lifecycle Status dropdown displays properly. No Owner/Team field visible for EA team selection (should show EA teams like 'business-architecture', 'enterprise-architecture', etc.). Cannot proceed with entity creation test."
severity: blocker

### 5. EA Entity Details View with Tabs
expected: Click on any entity in list view. Should navigate to entity details page with tabbed interface: Overview tab (entity info card with name, CI type, domain, lifecycle status, owner, team, data_quality_score color-coded green>=80%, yellow>=60%, red<60%), Attributes tab (flexible attribute display using FlexibleAttributeDisplay component), Relationships tab (placeholder for future), Audit History tab (placeholder). Edit and Delete buttons should show based on ea:update and ea:delete permissions.
result: skipped
reason: "Blocked by Test #4 - cannot create entity to view details"

### 6. EA Entity Update with Lifecycle Transition
expected: From entity details, click Edit button. Form should populate with existing entity data. Change lifecycle status or update attributes. Valid transitions work (Proposed→Active, Active→Deprecated). Invalid transitions (Retired→Active) show error: "invalid lifecycle transition: Retired → Active". Update should refresh ea_last_updated_by timestamp. CI type field should be disabled (cannot change CI type after creation).
result: skipped
reason: "Blocked by Test #4 - cannot create entity to update"

### 7. EA Entity Delete with Relationship Check
expected: From entity details, click Delete button. If entity has relationships in Neo4j, show confirmation: "This entity has N relationships. Deleting will affect all connected entities. Delete anyway?" If confirmed, entity is deleted. If no relationships, deletes immediately without confirmation. Delete should be protected by ea:delete RBAC permission. Audit log entry created for deletion.
result: skipped
reason: "Blocked by Test #4 - cannot create entity to delete"

### 8. EA CSV Import Template Download
expected: Navigate to http://localhost:3000/entities/import. Select CI Type from dropdown (e.g., "EA.Application-BusinessApplication"). Click "Download Template" button. CSV file should download with columns: Name*, CI_Type*, Domain*, Lifecycle_Status, Owner, Team, Tags, and CI type-specific attributes. Example row should be populated with valid sample data showing expected format.
result: pending

### 9. EA CSV Import Upload and Preview
expected: On import wizard step 1, select CI Type and drag-and-drop CSV file (or use file picker). Click Next. Step 2 should show preview table with first 10 rows of parsed data using PapaParse client-side parsing. Message should show: "Previewing first 10 rows of {total} rows". Table should have striped rows with hover states and sticky header. Back button returns to upload step.
result: pending

### 10. EA CSV Import Validation with Errors
expected: Click Validate on step 2. Step 3 should show validation results. If errors exist, see error count with icon. Two-tab interface: Error Table tab (Row, Field, Error, Expected, Actual columns) and Download CSV tab (click button to download error CSV for offline correction). All validation errors aggregated before import (fail-fast). Back button returns to upload step to fix errors.
result: pending

### 11. EA CSV Import Execute
expected: After validation passes (or fix errors), click Import button. Step 4 should show success: green checkmark, success/error counts. "View Imported Entities" button navigates to filtered entity list. "Import More" button resets wizard for new import. All entities created in database transaction with audit logs. Import protected by ea:create RBAC permission.
result: pending

### 12. EA Data Quality Dashboard
expected: Navigate to http://localhost:3000/ea/data-quality. Dashboard should show 4 metric cards: Total Entities, Completeness %, Stale Entities, Entities with Errors. Two donut charts: Lifecycle Status Breakdown, Errors by Domain (rendered using D3.js). Detail tables: Recent Stale Entities (top 10), Entities with Most Errors (top 10). Click metric cards to drill down to filtered entity lists. Dashboard requires ea:read permission.
result: pending

### 13. EA Domain Navigation Sidebar
expected: In EntityListView (http://localhost:3000/entities/business), left sidebar should show 8 EA domains with icons: Strategy (LightBulb), Business (Briefcase), Application (Chip), Data (Database), Technology (Server), Infrastructure (Cloud), Security (Shield), Governance (Scale). Clicking domain name should switch entity list to that domain with proper filtering. Active domain should be highlighted.
result: pass

### 14. EA Permissions in Database
expected: Query database: SELECT * FROM permissions WHERE name LIKE 'ea:%'. Should return 4 EA-specific permissions: ea:read, ea:create, ea:update, ea:delete. Query role_permissions: Should show admin role has all 4 permissions, editor has 3 (ea:read, ea:create, ea:update), viewer has 1 (ea:read). Permissions seeded by migration 010_ea_permissions.sql.
result: pending

### 15. EA Metamodel Seeded in Database
expected: Query database: SELECT COUNT(*) FROM ci_type_definitions WHERE name LIKE 'EA.%'. Should return 32 EA CI types. Query should show all 8 domains represented: Application (5), Business (5), Data (2), Governance (4), Infrastructure (5), Security (4), Strategy (4), Technology (3). CI types should have JSONB attribute schemas with required_attributes and optional_attributes arrays. Seeded by migration 009_add_ea_metamodel.up.sql.
result: pending

## Summary

total: 15
passed: 3
issues: 2
pending: 0
skipped: 10

## Gaps

- truth: "EA Entity CRUD endpoints work correctly with proper validation"
  status: failed
  reason: "User reported: API endpoints exist but have data issues - EA owner field requires exact team name match (kebab-case like 'business-architecture'), lifecycle status table missing 'Active' status for EA entities (has 'planned', 'on_order', 'in_stock' instead)"
  severity: major
  test: 1
  root_cause: "Migration 009 seeded EA teams with kebab-case names (e.g., 'business-architecture') but EA team field likely expects display names. Lifecycle statuses seeded from migration 003 are inventory-focused (planned, on_order, in_stock) not EA-appropriate (Proposed, Active, Deprecated, Retired)."
  artifacts:
    - path: cmd/migrations/009_add_ea_metamodel.up.sql
      issue: "EA teams inserted with kebab-case names but no corresponding display_name field"
    - path: cmd/migrations/003_lifecycle_statuses.sql
      issue: "Lifecycle statuses are inventory-focused, not EA-appropriate"
    - path: internal/ea/service.go
      issue: "EA entity validation expects exact team name match from database"
  missing:
    - "EA-appropriate lifecycle statuses (Proposed, Active, Deprecated, Retired, Archived)"
    - "EA team display names vs database name mapping or use consistent naming"
    - "Migration to add EA-specific lifecycle statuses"
  debug_session: ""

- truth: "CI Type dropdown loads 32 EA CI types from API"
  status: failed
  reason: "User reported: CI Type dropdown shows only 'Select CI Type' placeholder - the 32 EA CI types from /api/v1/ea/ci-types are not loading into dropdown. Lifecycle Status displays properly. No Owner/Team field visible for EA team selection."
  severity: blocker
  test: 4
  root_cause: "Frontend CI type dropdown not populated from /api/v1/ea/ci-types endpoint. The API returns data correctly (verified via curl), but the frontend EntityFormView or DynamicFormBuilder component not fetching or displaying CI types. Missing Owner/Team field suggests EA teams not loaded into form."
  artifacts:
    - path: web/src/views/ea/EntityFormView.vue
      issue: "Component not fetching EA CI types or not passing them to DynamicFormBuilder"
    - path: web/src/components/ea/DynamicFormBuilder.vue
      issue: "CI Type dropdown not populated with data from eaTypes store or API"
    - path: web/src/stores/eaTypes.ts
      issue: "Store may not be fetching CI types on component mount"
  missing:
    - "CI Type fetch call in EntityFormView onMounted"
    - "Owner/Team dropdown field in EA entity form"
    - "EA teams fetch and display in form"
  debug_session: ""

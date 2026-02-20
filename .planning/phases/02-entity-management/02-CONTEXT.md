# Phase 2: Entity Management - Context

**Gathered:** 2026-02-21
**Status:** Ready for planning

## Phase Boundary

Users can create, edit, search, and import EA entities across 8 EA domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance) with governance and data quality controls. This phase builds CRUD operations, bulk import, RBAC extensions, and a data quality dashboard for EA entities, leveraging the metamodel and service layer from Phase 1.

## Implementation Decisions

### Entity Form Approach

- **Dynamic form builder**: Single component that reads CI type schema and generates fields dynamically. Handles 60+ CI types across 8 domains automatically.
- **Extended field types**: Standard set (text, number, date, dropdown, textarea, checkbox) plus rich text editor, multi-select tags, date ranges, JSON/YAML editors for technical configs.
- **Accordion/sections layout**: Collapsible sections within forms for complex CI types with 20+ attributes. Better than single scrollable form or tabs for large forms.
- **Hybrid validation**: Frontend enforces basic validation (required, format), backend handles cross-field and business rule validation. Balanced approach.
- **Field inheritance**: Claude's discretion — choose approach based on Phase 1 CI type schema structure (base CI template or field groups).
- **Optional drafts**: "Save as draft" button for user-controlled draft saving. Not auto-save, gives users explicit control.
- **Relationship linking**: Both autocomplete search AND entity picker modal. Typeahead for quick selection, picker for complex filtering.
- **Validation errors**: Summary + inline display. Error summary at top ("3 errors need fixing") plus inline messages below each invalid field.
- **Field help**: Collapsible guides. "What's this?" links expand to show detailed explanations. On-demand help without clutter.
- **Mobile-optimized**: Simplified mobile form view with critical fields only, advanced fields in "More options" section. Not just responsive stacking.
- **Save feedback**: Button state only. Save button shows spinner and disables during save. Minimal feedback pattern.
- **Unsaved changes**: Warning prompt. "You have unsaved changes. Stay or leave?" browser-beforeunload-style prompt.
- **Conditional field display**: Yes, simple cases. Support basic visibility toggles (e.g., lifecycle-specific fields), not full rules engine.

### List View Interaction

- **Row density**: Comfortable rows (48px medium row height). Balanced readability and density, standard enterprise app feel.
- **Default columns**: Core fields only. Name, CI Type, Domain, Lifecycle Status, Owner, Last Updated. User adds more as needed.
- **Page organization**: Per-domain pages. Separate pages like `/entities/business-capabilities`, `/entities/applications`, not one unified list.
- **Page structure**: Domain + CI type pages. `/entities/business-capabilities` shows all, `/entities/business-capabilities/capability` filters to that type. Granular control.
- **Domain navigation**: Sidebar navigation. Left sidebar shows "Business", "Application", "Data", etc. Clicking switches domain page.
- **Breadcrumbs**: Full hierarchy. EA → Business Capabilities → Business Capability (type) → [entity]. Clear where you are.
- **Column configuration**: Tailored per CI type. Each CI type page has domain-specific columns and filters (e.g., Business Capabilities show "Criticality", Applications show "Version").
- **Filter types**: Basic filters. Text contains, number range, date range, dropdown for enums. Covers 80% of filtering needs.
- **Saved views**: No saved views. Grid resets to default on each visit. Simpler, no per-user view storage.
- **Row actions**: View + Edit + Delete. Three action icons in each row for basic CRUD operations.
- **Inline editing**: No, form only. All edits open full entity form for consistent validation.
- **Export**: CSV export only. Export current filtered/sorted view as CSV. Simple, universal format.
- **Pagination**: Server-side pagination. Traditional prev/next with page size selector (25/50/100). Predictable, standard enterprise pattern.
- **Bulk actions**: Full bulk operations. Delete, status change, owner reassignment, tag add/remove when multiple rows selected.
- **Search**: Global text search. Single search box searches across name, description, key fields. Simple, familiar.
- **Row selection**: Both methods. Checkbox multi-select (explicit, with "Select All") AND click-to-select (cleaner UI).
- **Detail view**: Separate page. Clicking row or "View" navigates to entity detail page. Full-screen details.

### Bulk Import Workflow

- **Upload initiation**: Separate import page. Dedicated `/import` route with full-page wizard for complex imports, not just a modal from list.
- **Import wizard flow**: Upload → Template → Fill → Validate → Import. Guided flow with template download ensures correct format.
- **Validation errors**: Both options. Show error summary ("45 errors") with both error CSV download AND inline error table viewing. Most flexible.
- **Preview step**: Yes, required preview. Show first 10 rows of parsed data in table. User confirms "This looks right" before importing.
- **Import undo**: No undo, re-import. No undo button — user must delete entities manually or re-import with corrections. Simpler, less risk.

### Data Quality Dashboard

- **Dashboard type**: Single-page widgets. All metrics visible as cards/charts on one scrollable page. Executive overview style.
- **Core metrics**: Essential metrics. Completeness % (based on required attributes of each CI type for EA), stale entities count, entities with errors, entities by lifecycle status. Covers basics.
- **Completeness calculation**: Based on required attributes of each CI type for EA entities. Each CI type has its own required field set, completeness calculated accordingly.
- **Visualization**: Stat cards + simple charts. Big number cards (e.g., "87% Complete") + pie charts for breakdowns. Clean, executive-friendly.
- **Drill-down**: Filtered entity list. Clicking "45 stale entities" opens entity list filtered to show those entities. Direct action.
- **Staleness definition**: Lifecycle-aware. Active entities not updated in X days, OR proposed/active entities with missing required fields. Context-aware, not just time-based.
- **Data freshness**: Manual refresh. User clicks "Refresh" button, data updates on-demand. Efficient, no surprise updates.
- **Historical trends**: Current state only. Snapshot of now, no line charts or historical tracking. Simpler, no historical data storage needed.

### Claude's Discretion

- **Field inheritance model**: Choose between base CI template (all types inherit common fields) or field groups (reusable field groups composed by CI types) based on Phase 1 CI type schema structure that works best.
- **Dynamic form architecture**: Component structure, state management, and schema reading approach for the dynamic form builder.
- **ag-grid configuration**: Specific configuration for comfortable rows, filters, and pagination that balances performance with UX.
- **Staleness threshold**: Define specific "X days" threshold for stale entity detection based on EA governance requirements.
- **Quality metric calculations**: Specific algorithms for completeness %, staleness detection, and error aggregation that align with EA data model.

## Specific Ideas

- **Dynamic form builder needs schema reading**: Component should read CI type definitions from API (Phase 1 metamodel) to generate fields dynamically.
- **Domain-specific columns are real**: Different CI types genuinely have different attributes (Business Capabilities have "criticality", Applications have "version"). Need flexible grid that adapts.
- **Per-domain page structure matters**: Users think in terms of "I need to see all Business Capabilities", not "I need to see all EA entities filtered by Business domain". Dedicated pages per domain reflect mental model.
- **Import wizard should guide users**: EA admins importing spreadsheets need guidance. Template download ensures correct format, preview prevents mistakes, both error display options support different workflows (fix offline vs fix in-browser).
- **Dashboard is monitoring, not action**: Data quality dashboard is for visibility ("How healthy is our EA data?"), not for bulk fixes. Clicking metrics takes you to entity list for actions.

## Deferred Ideas

- **Advanced search with query builder**: Noted for future phase. Current global text search covers basic needs.
- **Scheduled/automated imports**: Noted for future phase. Current manual import workflow handles MVP use case.
- **Data quality alerting**: Noted for future phase. Current manual refresh dashboard is monitoring-only.
- **Historical trend analysis**: Noted for future phase. Current state-only dashboard is simpler for MVP.
- **Advanced field rules engine**: Noted for future phase. Current simple conditional visibility covers 80% of needs.

---

*Phase: 02-entity-management*
*Context gathered: 2026-02-21*

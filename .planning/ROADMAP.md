# Roadmap: Enterprise Architecture Module

## Overview

This roadmap transforms Pustaka's CMDB into a comprehensive Enterprise Architecture platform by extending the existing CI taxonomy with EA entities across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance). The journey begins with foundational metamodel and service layer construction, progresses to entity management with full CRUD and bulk import capabilities, and culminates in relationship management and impact analysis visualization. EA entities are modeled as CI Types within the unified CMDB taxonomy, leveraging existing PostgreSQL, Neo4j, and Redis infrastructure while adding domain-specific validation and governance.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - Define EA metamodel, seed database, establish service layer
- [ ] **Phase 2: Entity Management** - CRUD operations, search, bulk import, and governance
- [ ] **Phase 3: Relationships & Impact** - Relationship management, visualization, and 1-hop analysis

## Phase Details

### Phase 1: Foundation
**Goal**: EA entities can be modeled as CI Types with validated metamodel and service infrastructure
**Depends on**: Nothing (first phase)
**Requirements**: META-01, META-02, META-03, META-04, META-05, INT-01, INT-05
**Success Criteria** (what must be TRUE):
  1. All 60+ EA CI type definitions exist in database across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance)
  2. All 20-25 core relationship types exist in database with bidirectional support (supports, depends_on, deployed_on, flows_to, assigned_to, etc.)
  3. EA service layer skeleton exists and wraps existing CI service with composition pattern
  4. Database migration successfully seeds EA types without breaking existing CMDB functionality
  5. EA entities are queryable through existing CI infrastructure with domain-specific validation framework
**Plans**: 3 plans (Wave 1: 01-01, 01-02 | Wave 2: 01-03)

Plans:
- [x] 01-01-PLAN.md — Define EA metamodel specifications (60+ CI types, 20-25 relationship types, cross-domain validation rules, data ownership model) ✅ COMPLETED 2026-02-20
- [ ] 01-02-PLAN.md — Create database migration for EA CI type and relationship type seeding (ea_teams table, validation queries, RBAC permissions)
- [ ] 01-03-PLAN.md — Build EA service layer skeleton with CI service composition (base service, repository, validation, 8 domain services)

### Phase 2: Entity Management
**Goal**: Users can create, edit, search, and import EA entities with governance and data quality controls
**Depends on**: Phase 1
**Requirements**: ENT-01, ENT-02, ENT-03, ENT-04, ENT-05, ENT-06, ENT-07, ENT-08, ENT-09, GOV-01, GOV-02, GOV-03, GOV-04, GOV-05, GOV-06, GOV-07
**Success Criteria** (what must be TRUE):
  1. User can create EA entities manually through forms for all 8 domains with validation
  2. User can edit and delete existing EA entities with relationship dependency checking
  3. User can search, filter, and paginate EA entity lists (handles 10K+ entities)
  4. User can import EA entities in bulk from CSV files with validation error reporting
  5. System enforces EA-specific RBAC permissions (ea:read, ea:create, ea:update, ea:delete) and tracks all changes in audit log
  6. System provides data quality dashboard showing completeness, staleness, and validation errors
  7. EA entities maintain lifecycle status (proposed, active, deprecated, retired)
**Plans**: TBD

Plans:
- [ ] 02-01: Build EA entity CRUD HTTP handlers and endpoints
- [ ] 02-02: Create frontend entity management views for each EA domain
- [ ] 02-03: Implement search, filtering, and pagination with ag-grid-vue3
- [ ] 02-04: Build bulk CSV import with validation (gocsv backend, PapaParse frontend)
- [ ] 02-05: Extend RBAC with EA permissions and enforce access control
- [ ] 02-06: Implement data quality dashboard and lifecycle status management

### Phase 3: Relationships & Impact
**Goal**: Users can model cross-domain relationships and visualize 1-hop impact analysis
**Depends on**: Phase 2
**Requirements**: REL-01, REL-02, REL-03, REL-04, REL-05, REL-06, REL-07, REL-08, IMP-01, IMP-02, IMP-03, IMP-04, IMP-05, IMP-06, INT-02, INT-03, INT-04
**Success Criteria** (what must be TRUE):
  1. User can create, edit, and delete relationships between EA entities across domains
  2. System automatically creates bidirectional relationships in Neo4j (forward + reverse edges)
  3. User can visualize EA entities and relationships in an interactive graph with domain filtering
  4. User can navigate graph by clicking entities to expand/collapse relationships
  5. User can perform 1-hop impact analysis (e.g., "This Application affects these Business Capabilities")
  6. System highlights impact paths in graph visualization
  7. User can search and filter relationships by type and domain
  8. EA entities link bidirectionally with existing infrastructure CIs (Application ↔ Server)
  9. User can export impact analysis results
**Plans**: TBD

Plans:
- [ ] 03-01: Build relationship CRUD endpoints with bidirectional Neo4j synchronization
- [ ] 03-02: Implement relationship type validation to prevent incorrect cross-domain connections
- [ ] 03-03: Create relationship management UI with autocomplete and bulk operations
- [ ] 03-04: Implement 1-hop impact analysis queries in Neo4j repository
- [ ] 03-05: Extend vis-network graph visualization for EA entities with domain filtering
- [ ] 03-06: Build CMDB-EA bidirectional integration (Application ↔ Server)
- [ ] 03-07: Implement impact analysis export functionality

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 1/3 | In progress | 2026-02-20 |
| 2. Entity Management | 0/6 | Not started | - |
| 3. Relationships & Impact | 0/7 | Not started | - |

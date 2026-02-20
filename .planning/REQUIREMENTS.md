# Requirements: Pustaka Enterprise Architecture Module

**Defined:** 2026-02-20
**Core Value:** Architects and stakeholders can trace relationships across domains to understand impact

## v1 Requirements

Requirements for initial EA module release. Each maps to roadmap phases.

### Metamodel & Data Foundation

- [x] **META-01**: Define 60+ EA CI types across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance) following ArchiMate 3.x patterns
- [x] **META-02**: Define 20-25 core relationship types supporting critical use cases (supports, depends_on, deployed_on, flows_to, assigned_to, etc.)
- [x] **META-03**: Create database migration seeding EA CI type definitions and relationship type definitions
- [x] **META-04**: Establish EA service layer skeleton with CI service composition pattern
- [x] **META-05**: Define data ownership rules and validation framework for EA entities

### Entity Management

- [x] **ENT-01**: User can create EA entities manually through forms for all 8 domains
- [x] **ENT-02**: User can edit existing EA entities with validation
- [ ] **ENT-03**: User can delete EA entities with relationship dependency checking
- [x] **ENT-04**: User can view EA entity details with all attributes and relationships
- [x] **ENT-05**: User can search EA entities by domain, type, name, and attributes
- [x] **ENT-06**: User can filter and paginate EA entity lists (handles 10K+ entities)
- [ ] **ENT-07**: User can import EA entities in bulk from CSV files with validation
- [ ] **ENT-08**: System validates EA entity data against type-specific rules before saving
- [ ] **ENT-09**: System tracks all EA entity changes in audit log (reuse existing audit_logs table)

### Relationship Management

- [ ] **REL-01**: User can create relationships between EA entities across domains
- [ ] **REL-02**: User can edit existing relationships with validation
- [ ] **REL-03**: User can delete relationships with dependency checking
- [ ] **REL-04**: System automatically creates bidirectional relationships in Neo4j (forward + reverse edges)
- [ ] **REL-05**: User can view all relationships for an EA entity in a list view
- [ ] **REL-06**: User can search and filter relationships by type and domain
- [ ] **REL-07**: System validates relationship types to prevent incorrect cross-domain connections
- [ ] **REL-08**: User can perform bulk relationship operations (import, delete)

### Impact Analysis & Visualization

- [ ] **IMP-01**: User can visualize EA entities and relationships in an interactive graph
- [ ] **IMP-02**: User can filter graph view by domain (show only business+application layers)
- [ ] **IMP-03**: User can navigate graph by clicking entities to expand/collapse relationships
- [ ] **IMP-04**: User can perform 1-hop impact analysis (e.g., "This Application affects these Business Capabilities")
- [ ] **IMP-05**: System highlights impact paths in graph visualization
- [ ] **IMP-06**: User can export impact analysis results

### Governance & Security

- [ ] **GOV-01**: EA entities respect existing RBAC system with extended permissions
- [x] **GOV-02**: System enforces `ea:read` permission for viewing EA entities
- [ ] **GOV-03**: System enforces `ea:create` permission for creating EA entities
- [ ] **GOV-04**: System enforces `ea:update` permission for editing EA entities
- [ ] **GOV-05**: System enforces `ea:delete` permission for deleting EA entities
- [ ] **GOV-06**: EA entities maintain lifecycle status (proposed, active, deprecated, retired)
- [ ] **GOV-07**: System provides data quality dashboard showing completeness, staleness, and validation errors

### CMDB Integration

- [x] **INT-01**: EA entities modeled as CI Types within existing CMDB taxonomy (unified data model)
- [ ] **INT-02**: EA entities link to existing infrastructure CIs (e.g., Application → Server)
- [ ] **INT-03**: Infrastructure CIs link to EA entities (e.g., Server → deployed Applications)
- [ ] **INT-04**: EA relationship queries return both EA and CMDB entities in unified graph
- [x] **INT-05**: EA entities leverage existing CI infrastructure (PostgreSQL, Neo4j, Redis, audit logging)

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Analytics

- **ANALYT-01**: Multi-hop impact analysis (2-5 hop Neo4j traversals)
- **ANALYT-02**: Application Portfolio Management (6R framework: Retain, Rehost, Refactor, etc.)
- **ANALYT-03**: Business Impact Analysis with criticality scoring
- **ANALYT-04**: Technology Stack Consolidation recommendations
- **ANALYT-05**: What-if scenario planning and simulation

### Advanced Visualization

- **VIZ-01**: Heat Map Visualization (color-coded maturity/risk/cost views)
- **VIZ-02**: Gap Analysis (current vs target state comparison)
- **VIZ-03**: Business Capability Mapping with explicit hierarchies
- **VIZ-04**: Custom Views/Perspectives by stakeholder type
- **VIZ-05**: Advanced graph clustering for large repositories (1000+ nodes)

### Advanced Governance

- **GOV-ADV-01**: Stakeholder Portal (read-only dashboards for PMs/execs)
- **GOV-ADV-02**: EA governance workflows with approval chains
- **GOV-ADV-03**: Automated data quality monitoring and alerts
- **GOV-ADV-04**: EA governance board decision tracking

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Automated application discovery | Unreliable; manual entry + bulk import preferred |
| Real-time architecture dashboards | EA is strategic, not operational; daily/weekly reports sufficient |
| Full TOGAF ADM automation | TOGAF is process framework, not software; support artifacts, don't automate process |
| Multi-tenant EA repositories | Breaks cross-organizational visibility; use domains/tags within single repo |
| Native Jira/Confluence sync | Integration maintenance nightmare; export via CSV/API instead |
| AI-driven architecture recommendations | Black-box decisions architects can't explain; provide analytics, let architects decide |
| Custom modeling language designer | Eats implementation time, prevents standardization; use ArchiMate 3.x + custom attributes |
| What-if scenario branching | Doubles data model complexity; use lifecycle status + versioning instead |
| EA-specific role types | Role proliferation; add EA permissions to existing roles, don't create new roles |
| Separate EA data model | EA entities are modeled as CI Types, not a separate parallel system |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| META-01 | Phase 1 | Complete |
| META-02 | Phase 1 | Complete |
| META-03 | Phase 1 | Complete |
| META-04 | Phase 1 | Complete |
| META-05 | Phase 1 | Complete |
| ENT-01 | Phase 2 | Complete |
| ENT-02 | Phase 2 | Complete |
| ENT-03 | Phase 2 | Pending |
| ENT-04 | Phase 2 | Complete |
| ENT-05 | Phase 2 | Complete |
| ENT-06 | Phase 2 | Complete |
| ENT-07 | Phase 2 | Pending |
| ENT-08 | Phase 2 | Pending |
| ENT-09 | Phase 2 | Pending |
| REL-01 | Phase 3 | Pending |
| REL-02 | Phase 3 | Pending |
| REL-03 | Phase 3 | Pending |
| REL-04 | Phase 3 | Pending |
| REL-05 | Phase 3 | Pending |
| REL-06 | Phase 3 | Pending |
| REL-07 | Phase 3 | Pending |
| REL-08 | Phase 3 | Pending |
| IMP-01 | Phase 3 | Pending |
| IMP-02 | Phase 3 | Pending |
| IMP-03 | Phase 3 | Pending |
| IMP-04 | Phase 3 | Pending |
| IMP-05 | Phase 3 | Pending |
| IMP-06 | Phase 3 | Pending |
| GOV-01 | Phase 2 | Pending |
| GOV-02 | Phase 2 | Complete |
| GOV-03 | Phase 2 | Pending |
| GOV-04 | Phase 2 | Pending |
| GOV-05 | Phase 2 | Pending |
| GOV-06 | Phase 2 | Pending |
| GOV-07 | Phase 2 | Pending |
| INT-01 | Phase 1 | Complete |
| INT-02 | Phase 3 | Pending |
| INT-03 | Phase 3 | Pending |
| INT-04 | Phase 3 | Pending |
| INT-05 | Phase 1 | Complete |

**Coverage:**
- v1 requirements: 46 total
- Mapped to phases: 46
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-20*
*Last updated: 2026-02-20 after roadmap creation*

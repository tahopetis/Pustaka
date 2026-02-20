# Pustaka Enterprise Architecture Module

## What This Is

An Enterprise Architecture (EA) module extending Pustaka's CMDB with an 8-domain metamodel (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance). EA entities are modeled as Configuration Items within the existing CMDB taxonomy, enabling unified relationship mapping, impact analysis, and visualization across all architectural layers.

## Core Value

Architects and stakeholders can trace relationships across domains to understand impact (e.g., "If this application changes, what business capabilities are affected?").

## Requirements

### Validated

*Shipped and confirmed valuable in the existing CMDB:*

- ✓ **Hierarchical CI taxonomy** — Domain → Category → Subcategory → CI Type with full CRUD operations (existing)
- ✓ **Relationship management** — Bidirectional relationships stored in Neo4j with graph traversal (existing)
- ✓ **RBAC with granular permissions** — Role-based access control with resource-level permissions (existing)
- ✓ **Audit logging** — All changes tracked with user context and timestamps (existing)
- ✓ **Vue 3 + TypeScript frontend** — Interactive UI with Pinia state management (existing)
- ✓ **Graph visualization** — vis-network integration for relationship rendering (existing)

### Active

*Current scope - building the EA module:*

- [ ] **8-domain EA metamodel** — All CI types defined for Strategy, Business, Application, Data, Technology, Infrastructure, Security, and Governance domains
- [ ] **Domain entity CRUD** — Create, read, update, delete operations for all EA entity types
- [ ] **Cross-domain relationships** — Model all relationships from the metamodel (app supports capability, project changes app, subsystem deployed on compute, etc.)
- [ ] **Relationship visualization** — Interactive graph showing EA entities and their relationships across domains
- [ ] **Basic impact analysis** — 1-hop relationship traversal (e.g., "This application affects these business capabilities")
- [ ] **Governance capabilities** — Lifecycle management and audit trail for EA entities
- [ ] **Manual data entry** — Forms for creating and editing EA entities
- [ ] **Bulk data import** — Import EA entities from CSV/spreadsheets
- [ ] **Extended RBAC** — EA-specific permissions added to existing roles (architects, strategists, PMs)
- [ ] **Bi-directional CMDB integration** — Link EA entities (applications) to existing infrastructure CIs (servers, databases)

### Out of Scope

*Explicitly excluded from v1:*

- **Analytical engines** (6R framework, Business Impact Analysis, Application Criticality) — Defer to v2; v1 focuses on foundation and data modeling
- **Advanced multi-hop impact analysis** — v1 supports direct relationships only; deep traversal algorithms in v2+
- **What-if scenario planning** — Requires analytical engines as foundation
- **EA-specific roles** — Extend existing RBAC roles rather than creating new Enterprise Architect, Business Architect roles
- **Separate EA data model** — EA entities are modeled as CI Types, not a separate parallel system

## Context

**Existing Pustaka Architecture:**
- **Backend:** Go with Chi v5 framework, PostgreSQL for structured data, Neo4j for relationships, Redis for caching
- **Frontend:** Vue 3 + TypeScript with Pinia state management, vis-network for graph visualization
- **Relationship Engine:** Bidirectional Neo4j graph with rich traversal capabilities
- **Taxonomy System:** Flexible Domain → Category → Subcategory → CI Type hierarchy
- **Maturity:** Production CMDB with proven infrastructure management, audit logging, and RBAC

**EA Metamodel Reference:**
Based on comprehensive metamodel with 8 domains and rich cross-domain relationships (see /home/tahopetis/dev/archer/docs for detailed structure). Includes 60+ entity types and 80+ relationship types across Strategy, Business, Application, Data, Technology, Infrastructure, Security (NIST 2.0), and Governance domains.

**Primary Use Cases:**
1. **Impact Analysis** — Trace relationships to understand cross-domain dependencies
2. **Visualization** — Interactive graphs showing architectural relationships
3. **Governance** — Lifecycle management, change tracking, and audit trail

## Constraints

- **Technology Stack** — Must use existing stack (Go backend, Vue 3 frontend, PostgreSQL + Neo4j + Redis)
- **Integration Model** — EA entities as CI Types within existing CMDB taxonomy, not separate system
- **Relationship Direction** — Bidirectional relationships in Neo4j matching current CMDB pattern
- **Performance** — Graph queries must remain responsive as EA dataset grows
- **Permissions** — Extend existing RBAC; no new EA-specific role types
- **Data Entry** — Support both manual forms and bulk import workflows

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| EA as CI Types (Option 1) | Leverages existing Neo4j relationship graph, RBAC, audit logging; avoids duplicate infrastructure; enables cross-domain queries in unified graph | — Pending |
| Full 8-domain metamodel in v1 | Greenfield EA requires complete foundation; analytical engines need all domains populated; avoids piecemeal iterations | — Pending |
| Defer analytical engines to v2 | Foundation first (data modeling, relationships, governance); analytics require quality data and user feedback; reduces v1 complexity | — Pending |
| Bi-directional CMDB integration | Unified data model; Infrastructure CIs link to applications (apps run on servers); Applications link to infrastructure (deployment dependencies) | — Pending |
| Extend existing RBAC roles | Avoids role proliferation; Architects/PMs already exist; add EA permissions (ea:create, ea:update) rather than new roles | — Pending |

---
*Last updated: 2026-02-20 after project initialization*

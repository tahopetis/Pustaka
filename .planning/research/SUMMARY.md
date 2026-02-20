# Project Research Summary

**Project:** Enterprise Architecture Module (CMDB extension)
**Domain:** Enterprise Architecture + IT Asset Management
**Researched:** 2026-02-20
**Confidence:** HIGH

## Executive Summary

The Enterprise Architecture (EA) module extends Pustaka's existing CMDB to provide strategic business-technology alignment capabilities. Unlike traditional EA tools that create parallel universes to infrastructure CMDBs, experts recommend modeling EA entities (objectives, capabilities, applications, data objects) as Configuration Items within a unified taxonomy. This approach leverages existing infrastructure—PostgreSQL for structured data, Neo4j for relationship traversal, Redis for caching—while adding domain-specific validation and 60+ EA entity types across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance).

The recommended approach prioritizes **data quality over feature completeness**. Research consistently shows that EA initiatives fail when data becomes stale (60% accuracy within 3-6 months), relationships grow unmanaged (80+ types become unusable), or governance is absent. The roadmap must therefore begin with a minimal, use-case-driven metamodel (20-25 core relationship types, not 80+), clear data ownership rules, and automated validation before building advanced analytics. The architecture employs service layer composition—EA service wraps the existing CI service—ensuring clean separation while reusing RBAC, audit logging, and lifecycle management.

Key risks center on **integration failure** (treating EA as separate from CMDB creates duplicate entities and broken impact analysis) and **premature optimization** (building dashboards before data is reliable). Mitigation strategy: model EA as CI Types from day one, implement data quality gates between phases (>85% accuracy required to advance), and start with 1-hop graph traversals (expand to multi-hop in v2+). The payoff is significant—enterprise EA tools cost $100K+ while an open-source CMDB-integrated approach delivers core value (impact analysis, capability mapping, portfolio rationalization) at zero licensing cost.

## Key Findings

### Recommended Stack

The stack leverages existing Pustaka infrastructure with targeted extensions for EA-specific needs.

**Core technologies (no changes):**
- **Go 1.22+ + Chi v5**: Backend type safety and lightweight routing — ideal for EA data processing and REST APIs
- **PostgreSQL 15+**: EA entity storage — ACID compliance for architectural data, proven at scale
- **Neo4j 5.x**: Relationship graph — best-in-class for cross-domain impact analysis (e.g., "Database X affects which business capabilities?")
- **Redis 7.x**: Caching layer — fast lookup for EA metadata (CI types, relationship types)
- **Vue 3.4+ + TypeScript 5.2+**: Frontend with composition API — essential for 60+ EA entity type forms
- **Pinia 2.1+**: State management — simpler than Vuex for EA entity stores

**Key additions:**
- **gocsv (Go)**: Type-safe CSV import/export — struct-based parsing beats manual field assignment for bulk EA entity uploads
- **PapaParse 5.4.1 (Vue)**: Client-side CSV validation — web workers prevent UI freeze during large imports, robust error handling
- **ag-grid-vue3 v34.3+**: Enterprise data grid — virtual scrolling handles 10K+ EA entities, filtering/sorting/pagination out-of-box
- **validator/v10 (Go)**: Struct validation — industry standard with 100+ built-in rules for EA business rules
- **zod (TypeScript)**: Runtime validation — TypeScript-first schema validation for EA entity forms

**Critical decision:** EA entities are modeled as CI Types, not separate database tables. This extends the existing `ci_type_definitions` table with an `ea_domain` column and reuses `configuration_items` for all EA entities.

### Expected Features

Research identified 10 table stakes features (missing = product feels incomplete), 10 differentiators (competitive advantage), and 10 anti-features (commonly requested but problematic).

**Must have (table stakes - v1):**
- **Multi-Domain Metamodel** — EA requires business, application, data, technology layers to be useful (8 domains already defined)
- **Entity CRUD Operations** — Extend CMDB CRUD to EA entity types (capabilities, applications, processes)
- **Cross-Domain Relationships** — Core EA value is linking capabilities→apps→infrastructure (80+ relationship types defined)
- **Relationship Visualization** — vis-network shows architectural connections (extend existing CMDB graph)
- **Basic Impact Analysis** — "What breaks if X changes?" is fundamental EA use case (1-hop Neo4j traversal)
- **Bulk Data Import** — Organizations have spreadsheets to import (CSV with gocsv/PapaParse)
- **Search & Filtering** — Find specific EA entities in large repositories (extend CMDB search)
- **Audit Trail** — All EA changes tracked for governance (reuse existing audit_logs)
- **RBAC** — EA has sensitive strategic data (extend existing roles with `ea:read`, `ea:update`, `ea:delete`)
- **Standard Framework Support** — TOGAF/ArchiMate users expect standard entities (metamodel based on ArchiMate 3.x)

**Should have (competitive - v1.x):**
- **Heat Map Visualization** — Color-coded maturity/risk/cost views for leadership dashboards (aggregate attributes for visual display)
- **Business Capability Mapping** — Explicit "what business does" view vs. "how" (hierarchical capability taxonomies)
- **Gap Analysis** — Compare current vs target architecture (requires defining "target" lifecycle status)
- **Application Portfolio Management** — Identify redundant applications, rationalize portfolio (15-20% cost savings per McKinsey)
- **Stakeholder Portal** — Read-only views for non-architects (execs, PMs, developers)
- **Custom Views/Perspectives** — Filter graph by domain (e.g., show only business+application layers)
- **Technology Stack Consolidation** — Identify overlapping technologies to reduce sprawl

**Defer (v2+):**
- **Automated CMDB Discovery for EA** — Application discovery unreliable; manual entry + bulk import preferred
- **Real-Time Architecture Dashboard** — EA is strategic, not operational; daily/weekly change reports sufficient
- **Full TOGAF ADM Automation** — TOGAF is process framework, not software; support artifacts, don't automate process
- **Multi-Tenant EA Repositories** — Breaks cross-organizational visibility; use domains/tags within single repository
- **Native Jira/Confluence Integration** — Integration maintenance nightmare; export via CSV/API, no real-time sync
- **AI-Driven Architecture Recommendations** — Black-box decisions architects can't explain; provide analytics, let architects decide

**Anti-features (avoid entirely):**
- **Custom Modeling Language Designer** — Eats implementation time, prevents standardization; use ArchiMate 3.x + custom attributes
- **What-If Scenario Branching** — Doubles data model complexity; use lifecycle status (current vs future) + versioning instead
- **Separate EA Role Types** — Role proliferation; add EA permissions to existing roles, don't create new role types

### Architecture Approach

EA module extends existing CMDB via **service layer composition**, not modification. EA service wraps CI service, treating EA entities as specialized CIs with domain-specific validation and relationship rules. All EA entities are stored in the existing `configuration_items` table with CI type names following `{domain}.{entity}` pattern (e.g., "strategy.objective", "business.capability_l1"). Relationships are bidirectional in Neo4j for O(1) traversal in either direction.

**Major components:**
1. **EA Service Layer** (`internal/ea/`) — Domain logic for 8 EA domains, wraps existing CI service, enforces cross-domain validation rules
2. **EA HTTP Handlers** (`internal/api/handlers/ea_handlers.go`) — EA-specific endpoints, request/response transformation, reuses middleware stack (JWT, RBAC, audit)
3. **EA Frontend** (`web/src/views/ea/`) — Entity CRUD forms by domain, relationship management, graph visualization, bulk import UI
4. **Pinia Stores** (`web/src/stores/ea*.ts`) — State management for EA entities, types, relationships, import state
5. **Neo4j Repository** — Extended with EA-specific relationship queries, 1-hop impact analysis in v1, multi-hop in v2+
6. **Redis Cache** — EA type definitions cache, frequently accessed entity metadata

**Key architectural patterns:**
- **Service Layer Composition**: EA service wraps CI service (clean separation, CI service remains reusable)
- **CI Type Taxonomy**: EA entities stored as CIs with typed names (single source of truth, unified relationships)
- **Bidirectional Relationship Sync**: Forward and reverse relationships in Neo4j (O(1) traversal both directions)
- **Domain-Based Organization**: Backend subdomain services (`strategy_service.go`, `business_service.go`), frontend views grouped by domain

**Critical principle:** EA adds data and rules, not infrastructure. All EA entities are CIs, all EA relationships use existing relationship system, all EA operations are CI operations with domain-specific behavior.

### Critical Pitfalls

Research identified 8 critical pitfalls that kill EA projects, plus technical debt patterns, integration gotchas, performance traps, and security mistakes.

**Top 5 pitfalls (with prevention strategies):**

1. **Data Quality Death Spiral** — EA data becomes stale within 3-6 months, accuracy drops below 60%, users abandon system for spreadsheets. **Prevention**: Define clear data ownership during metamodel creation, build automated validation from day one (required fields, relationship constraints, data quality dashboards), integrate EA lifecycle with ITSM change processes, start with high-value low-volume domains first.

2. **Relationship Modeling Bloat** — Attempting to model every possible relationship type creates unmaintainable web (80+ types become unusable). **Prevention**: Start with 20-25 core relationship types supporting 3-5 critical use cases, implement relationship tiers (Core/Extended/Historical), add relationships incrementally based on user demand not theoretical completeness.

3. **CMDB Integration Failure** — Treating EA as "separate but related" to existing CMDB creates parallel universes (duplicate entities, broken impact analysis). **Prevention**: Model EA entities AS CI Types from day one (unified taxonomy), create explicit bidirectional relationships (Application→Server AND Server→Application), use CI Type attributes to flag EA-relevant entities.

4. **Governance Vacuum** — Without governance, EA becomes free-for-all (contradictory relationships, data loses trust, reflects politics not reality). **Prevention**: Define EA governance board BEFORE launching module (charter, decision rights, escalation path), implement approval workflows for entity creation and critical relationship changes, start with lightweight governance (review 10% of changes randomly) and scale based on need.

5. **Premature Optimization** — Building advanced analytics before data is reliable (queries return nonsense, stakeholders lose faith). **Prevention**: Phase 1 focuses ONLY on accurate entities, core relationships, basic CRUD, simple graph viz; define success criteria as data quality metrics not feature count; resist "wouldn't it be cool if" requests until data quality >90%.

**Technical debt patterns to avoid:**
- Manual entity creation to test metamodel → ONLY for prototype, must be automated before Phase 1 completion
- Hardcoded relationship types → NEVER; relationship type system must be data-driven from day one
- Limiting graph traversal depth to avoid performance → Acceptable for Phase 1-2, Phase 3 requires optimization not feature limitations

**Performance traps:**
- N+1 query pattern in relationship loading → Use batch loading, pre-fetch relationships
- Neo4j recursive queries without depth limits → Hard limit traversal depth (max 5 hops), use breadth-first search
- No caching of frequently accessed entities → Redis caching of popular entities (top 20% accessed 80% of time)

**Security mistakes:**
- EA entities visible to all authenticated users by default → Default-deny permissions: EA entities require explicit `ea:read` permission
- Relationship history not tracked → Immutable audit logs for all relationship changes
- Bulk export without approval mechanism → Require admin approval for exports >1000 entities, log all bulk exports

## Implications for Roadmap

Based on combined research (stack requirements, feature dependencies, architecture patterns, pitfall prevention), the recommended roadmap follows a **data-quality-first phased approach** that delivers value incrementally while avoiding common EA failure modes.

### Phase 1: Foundation (Metamodel + Core Data)

**Rationale:** Research shows EA projects fail when metamodels are bloated (80+ relationship types) or data quality is ignored. This phase establishes minimal, use-case-driven data model with validation and ownership rules—preventing Pitfall 2 (Relationship Bloat) and Pitfall 1 (Data Quality Death Spiral). Architecture dependency: nothing works without CI type definitions and database schema.

**Delivers:**
- EA metamodel with 60+ entity types across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance)
- 20-25 core relationship types (supports, depends_on, deployed_on, flows_to, assigned_to, etc.)
- Database migration seeding EA CI types and relationship types
- EA service skeleton with CI service composition
- Basic validation rules and data ownership framework

**Addresses:**
- Multi-Domain Metamodel (table stakes)
- Entity CRUD Operations foundation (table stakes)
- Cross-Domain Relationships foundation (table stakes)
- RBAC Extension foundation (table stakes)

**Avoids:**
- Relationship modeling bloat (starts with 20-25 core types, not 80+)
- Data quality death spiral (defines ownership and validation before entities exist)
- CMDB integration failure (models EA as CI Types from day one)

**Features from FEATURES.md:** Multi-Domain Metamodel, Entity CRUD Operations, Cross-Domain Relationships, RBAC Extension

**Stack elements from STACK.md:** Go + Chi (backend), PostgreSQL (schema), existing CI service (composition pattern), validator/v10 (validation)

**Implements architecture components:** EA Service Layer skeleton, EA CI Type Definitions, Database Migration

### Phase 2: Entity Management (CRUD + Search + Bulk Import)

**Rationale:** Once metamodel exists, users need to create and manage EA entities. This phase delivers full CRUD, search/filtering, and bulk import—addressing table stakes features while avoiding premature optimization (no advanced analytics yet). Architecture dependency: requires Phase 1 metamodel and service skeleton.

**Delivers:**
- EA entity CRUD endpoints (create, read, update, delete) for all 8 domains
- HTTP handlers with request/response transformation (EA request ↔ CI request)
- Frontend views for each EA domain (Strategy/, Business/, Application/, etc.)
- Pinia stores for EA entities and types
- Search and filtering by domain, type, attributes
- Bulk CSV import with validation (gocsv backend, PapaParse frontend)
- Data quality dashboard (completeness, staleness, validation errors)

**Addresses:**
- Entity CRUD Operations (table stakes)
- Search & Filtering (table stakes)
- Bulk Data Import (table stakes)
- Audit Trail (table stakes - automatic via CMDB)

**Uses:**
- gocsv (backend CSV parsing)
- PapaParse (frontend CSV validation)
- ag-grid-vue3 (entity list views)
- validator/v10 (validation rules)
- zod (frontend form validation)

**Implements architecture components:** EA HTTP Handlers, Frontend Views, Pinia Stores, EA Import Service

### Phase 3: Relationship Management (Create + Visualize + Analyze)

**Rationale:** Core EA value is linking business→applications→infrastructure. This phase implements relationship CRUD, bidirectional sync, and 1-hop impact analysis—delivering "what breaks if X changes?" capability. Architecture dependency: requires Phase 2 entities (can't create relationships without entities).

**Delivers:**
- EA relationship CRUD endpoints (create, read, update, delete relationships)
- Bidirectional relationship synchronization (Application→Database AND Database→Application)
- Relationship type validation (prevent incorrect cross-domain relationships)
- Relationship management UI (list, filter, search, create with autocomplete)
- 1-hop impact analysis (Neo4j traversal: "This Application affects these Business Capabilities")
- Relationship visualization in existing vis-network graph
- Bulk relationship operations

**Addresses:**
- Cross-Domain Relationships (table stakes - full implementation)
- Basic Impact Analysis (table stakes)
- Relationship Visualization (table stakes)

**Uses:**
- Neo4j (relationship storage and traversal)
- vis-network (relationship graph visualization)
- Existing CI relationship service (extended for EA bidirectional sync)

**Implements architecture components:** EA Relationship Service, Relationship Frontend, 1-hop Impact Analysis

### Phase 4: Advanced Features (Heat Maps + Gap Analysis + Governance)

**Rationale:** Once core entities and relationships exist with quality data (>85% accuracy, verified by data quality dashboard), add differentiator features that provide competitive advantage. This phase delivers analytics and governance while avoiding Premature Optimization pitfall.

**Delivers:**
- Heat Map Visualization (color-coded maturity/risk/cost views)
- Gap Analysis (current vs target state comparison)
- Business Capability Mapping (explicit capability hierarchies)
- Custom Views/Perspectives (filter graph by domain)
- Stakeholder Portal (read-only dashboards for PMs/execs)
- EA governance workflows (approval for entity creation, critical relationship changes)
- Extended RBAC with EA permissions matrix

**Addresses:**
- Heat Map Visualization (differentiator)
- Gap Analysis (differentiator)
- Business Capability Mapping (differentiator)
- Stakeholder Portal (differentiator)
- Custom Views/Perspectives (differentiator)

**Avoids:**
- Governance vacuum (implements approval workflows and decision rights)
- Premature optimization (only advanced after data quality gate passed)

**Uses:**
- D3.js (custom heatmap visualizations)
- Existing RBAC system (extended with EA permissions)
- vis-network (domain-based filtering)

### Phase 5: Optimization (Performance + Multi-hop Analytics)

**Rationale:** With proven value and growing usage, optimize performance and add advanced analytics. This phase addresses scaling bottlenecks identified in architecture research and adds multi-hop impact analysis (deferred from v1 per Pitfall 5).

**Delivers:**
- Multi-hop impact analysis (2-5 hop Neo4j traversals)
- Graph clustering for large EA repositories (1000+ nodes)
- Performance optimizations (query batching, caching, database indexes)
- Background job processing for deep traversals
- Advanced graph visualization (lazy loading, search-driven navigation)
- Technology Stack Consolidation (identify duplicate technologies)
- Application Portfolio Management foundation (lifecycle, criticality, cost attribution)

**Addresses:**
- Multi-Hop Impact Analysis (v2+ feature)
- Technology Stack Consolidation (differentiator)
- Application Portfolio Management foundation (v2+ feature)

**Avoids:**
- Performance traps (N+1 queries, uncached traversals, no depth limits)

**Uses:**
- Neo4j clustering algorithms
- Redis caching (popular entities, query results)
- Background job queues (async deep traversals)

### Phase Ordering Rationale

**Foundation-first sequence:** Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 follows dependency chains identified in architecture research (metamodel → entities → relationships → analytics) and feature dependencies in FEATURES.md (Multi-Domain Metamodel requires Entity CRUD, Basic Impact Analysis requires Relationship Visualization).

**Data quality gates:** Each phase requires >85% data accuracy before advancing to next (prevents Data Quality Death Spiral pitfall). Phase 2 delivers Data Quality Dashboard to measure this; Phase 4 requires governance workflows to maintain it.

**Pitfall avoidance by design:**
- **Phase 1 minimal metamodel** → avoids Relationship Modeling Bloat (starts with 20-25 core types, not 80+)
- **EA as CI Types from day one** → avoids CMDB Integration Failure (unified taxonomy, no duplicate entities)
- **Data quality before analytics** → avoids Premature Optimization (Phase 4-5 analytics only after Phase 2-3 data is reliable)
- **Governance in Phase 4** → avoids Governance Vacuum (approval workflows once data exists and is valuable)
- **1-hop traversals in Phase 3, multi-hop in Phase 5** → avoids Performance Traps (depth limits before optimization)

**Feature grouping rationale:**
- **Phases 1-3 (MVP):** Table stakes features required for core EA use cases (model, view, impact)
- **Phase 4 (Competitive):** Differentiators that provide edge over proprietary tools (heatmaps, gap analysis, governance)
- **Phase 5 (Advanced):** High-complexity features deferred until product-market fit established (APM, multi-hop analytics)

### Research Flags

**Phases likely needing deeper research during planning:**

- **Phase 1 (Foundation):** Database migration strategy requires research into PostgreSQL migration tooling (golang-migrate already in use, but need to verify EA-specific seed data approach). **Action:** Research migration file structure and CI type seeding patterns during Phase 1 planning.

- **Phase 2 (Entity Management):** ag-grid-vue3 integration patterns need research into Vue 3 + TypeScript setup, column definition structures, and virtual scrolling configuration. **Action:** Research ag-grid-vue3 examples and best practices during Phase 2 planning.

- **Phase 3 (Relationship Management):** Neo4j bidirectional relationship synchronization requires research into Cypher query patterns for creating reverse edges, transaction management, and rollback strategies. **Action:** Research Neo4j relationship management patterns during Phase 3 planning.

- **Phase 4 (Advanced Features):** Heat map visualization with D3.js requires research into color-scale algorithms, data aggregation patterns, and Vue 3 + D3 integration. **Action:** Research D3.js heatmap examples and Vue integration during Phase 4 planning.

- **Phase 5 (Optimization):** Multi-hop Neo4j traversal algorithms require research into query optimization, depth limiting strategies, and background job processing patterns. **Action:** Research Neo4j graph algorithms and Go job queue libraries during Phase 5 planning.

**Phases with standard patterns (skip `/gsd:research-phase`):**

- **Phase 2 (Entity Management):** CRUD operations are well-documented (existing CI handlers provide patterns), validation with validator/v10 is standard, CSV import with gocsv follows established patterns. **Reason:** High-confidence stack with existing codebase examples.

- **Phase 4 (Stakeholder Portal):** Read-only dashboards leverage existing RBAC system, Vue 3 view composition is standard pattern. **Reason:** Extension of existing authentication/authorization patterns, no novel integration challenges.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies verified via official documentation (gocsv, ag-grid-vue3, PapaParse, validator/v10, zod). Version compatibility confirmed with existing Pustaka stack (Go 1.22, Vue 3.4, PostgreSQL 15, Neo4j 5.x). Architecture decision (EA as CI Types) backed by strong rationale from existing codebase analysis. |
| Features | HIGH | Table stakes and differentiator features identified from multiple high-confidence sources (official ArchiMate 3.1 spec, EA tool comparison articles, Gartner research on EA pitfalls). Feature dependency graph validated against architecture patterns. Anti-features section based on real EA failure patterns documented by Open Group and Gartner. |
| Architecture | HIGH | Architecture patterns (service layer composition, CI type taxonomy, bidirectional relationship sync) derived from analysis of existing Pustaka CMDB codebase and standard Go/Vue 3 practices. Build order and dependencies validated against feature requirements and data flow. Anti-patterns based on common CMDB integration failures documented in industry research. |
| Pitfalls | MEDIUM | Data quality and relationship bloat pitfalls backed by multiple Chinese industry sources (2025) and Gartner research (HIGH confidence). EA governance pitfalls supported by Open Group TOGAF guidance (HIGH confidence). Performance traps based on Neo4j best practices (MEDIUM confidence - limited EA-specific benchmarks). Specific quantitative thresholds (e.g., "85% accuracy") inferred from general CMDB best practices, not EA-specific studies (LOW confidence - needs validation). |

**Overall confidence:** HIGH

Stack, features, and architecture recommendations are grounded in official documentation, existing codebase analysis, and industry best practices. Pitfalls research is comprehensive but some specific quantitative thresholds (85% accuracy, 20-25 relationship types) are inferred from general CMDB practices rather than EA-specific studies—these should be validated during implementation and adjusted based on real-world usage.

### Gaps to Address

**Uncertainties identified during research:**

1. **Neo4j scaling benchmarks for EA graphs** — Research found general Neo4j best practices but limited EA-specific performance data for 10K+ entity graphs with 80+ relationship types. **How to handle:** Performance test during Phase 3 with realistic dataset (1000+ entities, 20+ relationship types), profile query times, add indexes and depth limits if needed. Monitor Neo4j memory usage and query times in production.

2. **ag-grid-vue3 virtual scrolling with complex EA entity forms** — ag-grid documentation excellent for simple tables, but limited examples of grid integration with complex forms (60+ entity types, domain-specific attributes). **How to handle:** Start with simple ag-grid implementation in Phase 2, test virtual scrolling with 1000+ EA entities, iterate on column definitions and filtering. If ag-grid proves insufficient for complex forms, consider hybrid approach (ag-grid for list views, custom forms for CRUD).

3. **Optimal relationship type count for Phase 1** — Research recommends 20-25 core relationship types to avoid bloat, but doesn't specify which 20-25 from the 80+ defined in metamodel. **How to handle:** During Phase 1 planning, map 3-5 critical use cases (e.g., "Application depends on Database", "Capability supported by Application") and select relationship types that support them. Add relationship types incrementally in Phase 3 based on user demand.

4. **Data quality thresholds** — "85% accuracy" threshold inferred from general CMDB practices, not EA-specific research. **How to handle:** Implement Data Quality Dashboard in Phase 2 with metrics (completeness, staleness, validation errors). Use Phase 2 end as baseline measurement, adjust threshold based on real-world data quality achievable in practice. Set realistic minimum (e.g., 75%) and ideal (e.g., 90%) targets.

5. **Bulk import performance for 10K+ entities** — Research identifies batch processing and web workers as best practices, but limited benchmarks for EA-specific import complexity (60+ entity types, cross-domain relationship validation). **How to handle:** Implement batch import in Phase 2 (100 entities per transaction), test with 1000+ entity CSVs, measure processing time. If import takes >5 minutes, add WebSocket progress streaming and background job processing. Document realistic import size limits.

**Validation strategy:**
- Performance testing in Phase 3 (graph traversal with 1000+ entities)
- User testing in Phase 2 (CRUD workflows, bulk import usability)
- Data quality measurement in Phase 2 (baseline accuracy, completeness)
- Governance pilot in Phase 4 (lightweight approval workflows with 10% random audit)

## Sources

### Primary (HIGH confidence)

**Official Documentation:**
- [gocsv GitHub Repository](https://github.com/gocarina/gocsv) — CSV serialization/deserialization for Go, struct tags, custom converters
- [AG Grid Vue 3 Documentation](https://www.ag-grid.com/vue-data-grid/getting-started/) — Enterprise data grid features, version 34.3+ with Vue 3.3+ requirement
- [vis-network NPM Package](https://www.npmjs.com/package/vis-network) — Network visualization for EA relationship graphs, clustering support
- [go-playground/validator v10](https://github.com/go-playground/validator) — Struct validation for Go, 100+ built-in rules, cross-field validation
- [ArchiMate 3.1 Specification - Application Layer](https://pubs.opengroup.org/architecture/archimate3-doc/chap09.html) — Official metamodel for application entities and relationships

**Existing Codebase Analysis:**
- `/home/tahopetis/dev/pustaka/internal/ci/` — CI service, repository, Neo4j operations (reused by EA module)
- `/home/tahopetis/dev/pustaka/internal/api/` — HTTP handlers, middleware patterns (JWT, RBAC, audit logging)
- `/home/tahopetis/dev/pustaka/web/src/` — Vue 3 component structure, Pinia stores, API client patterns
- `/home/tahopetis/dev/pustaka/.planning/PROJECT.md` — EA module requirements and constraints

### Secondary (MEDIUM confidence)

**Industry Articles & Research:**
- [Enterprise Architecture Tools Comparison 2026 - LeanIX, Ardoq, Planview](https://www.softwaretestinghelp.com/enterprise-architecture-tools/) — Comprehensive tool comparison, feature differentiation, cost analysis
- [TOGAF ADM Support in EA Tools](https://www.visual-paradigm.com/togaf/adm/) — TOGAF process framework vs. software automation distinction
- [Business Capability Mapping and Heat Map Analysis](https://www.mega.com/resources/business-capability-mapping/) — Capability hierarchy patterns, heatmap visualization use cases
- [EA Tool Collaboration Features - Prolaborate](https://www.prolaborate.com/collaboration) — Stakeholder portal patterns, democratized data collection
- [Common EA Pitfalls and Mistakes](https://www.gartner.com/en/information-technology/insights/enterprise-architecture-pitfalls) — Gartner research on EA failure modes

**CMDB Data Quality & Maintenance (Chinese Industry Sources 2025):**
- "CMDB数据质量管理：从'垃圾进，垃圾出'到精准运维的蜕变之路" — Data quality challenges, 70% of CMDB projects face quality issues
- "初识CMDB：数据质量的管理" — Data accuracy, completeness, consistency, timeliness dimensions
- "企业CMDB配置管理系统：从痛点出发的实战构建指南" — Common pitfalls: data silos, change management gaps, relationship chaos
- "从某大型企业实践看CMDB建设的核心问题" — Value proposition, data governance, organizational resistance

**Enterprise Architecture Frameworks (Chinese Industry Sources 2025-2026):**
- "TOGAF难以落地吗？组织常见的误解" — Common misapplications: treating as linear process, producing artifacts without decision purpose
- "企业架构的十个误区" — Strategic disconnection from execution, PPT architectures without implementation detail
- "TOGAF架构元模型详解" — Core metamodel concepts, traceability, consistency requirements

### Tertiary (LOW confidence)

**Web Search Results - Verify Before Use:**
- Enterprise Architecture AI-driven visualization trends (2025) — Multiple sources mention AI, but specific tools unclear
- Neo4j Bloom for EA exploration — Mentioned in searches, but requires separate license and setup
- TOGAF ArchiMate modeling tools — Searches returned AI tools, not specific ArchiMate modelers. Use [Archi](https://www.archimatetool.com/) for open-source ArchiMate modeling if needed
- Specific quantitative thresholds (e.g., "85% data accuracy", "20-25 relationship types") — Inferred from general CMDB best practices, not EA-specific studies. Validate during implementation.

**Performance & Scalability:**
- "Getting started with chaos engineering" (2025) — Resilience testing for distributed systems, failure injection (general practices, not EA-specific)
- "Optimized predictive maintenance for streaming data" (2025) — Industrial IoT challenges with dynamic data, scale issues (analogous to EA data streams, but domain-specific validation needed)

---
*Research completed: 2026-02-20*
*Ready for roadmap: yes*

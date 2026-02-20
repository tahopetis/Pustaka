# Feature Research

**Domain:** Enterprise Architecture Module (on top of CMDB)
**Researched:** 2026-02-20
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Multi-Domain Metamodel** | EA requires business, application, data, technology layers to be useful | HIGH | Already defined (8 domains) - just needs CI type definitions |
| **Entity CRUD Operations** | Users must create/read/update/delete EA entities (capabilities, applications, processes) | LOW | CMDB already has CI CRUD - extend to EA types |
| **Cross-Domain Relationships** | Core value of EA is linking capabilities→apps→infrastructure | MEDIUM | Neo4j already handles relationships - define EA-specific types |
| **Relationship Visualization** | Users expect visual graphs showing architectural connections | MEDIUM | vis-network exists - ensure it handles EA relationship semantics |
| **Basic Impact Analysis** | "What breaks if X changes?" is fundamental EA use case | MEDIUM | 1-hop Neo4j traversal - extend existing impact queries |
| **Bulk Data Import** | Organizations have spreadsheets of applications/capacities to import | MEDIUM | CSV import already exists - add EA-specific field mappings |
| **Search & Filtering** | Users need to find specific EA entities in large repositories | LOW | Extend CMDB search with EA domain filters |
| **Audit Trail** | All EA changes must be tracked for governance | LOW | Already exists in CMDB audit logging |
| **RBAC** | EA has sensitive strategic data - access control required | LOW | Extend existing roles with EA permissions (ea:read, ea:update) |
| **Standard Framework Support** | TOGAF/ArchiMate users expect standard entities/relationships | MEDIUM | Metamodel based on ArchiMate 3.x - relationship types match standard |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Heat Map Visualization** | Color-coded maturity/risk/cost views for leadership dashboards | MEDIUM | Aggregate attributes (e.g., application health, capability maturity) into visual heatmaps |
| **Application Portfolio Management (APM)** | Identify redundant applications, rationalize portfolio (15-20% cost savings per McKinsey) | HIGH | Requires application lifecycle, criticality, cost attribution |
| **Business Capability Mapping** | Explicit "what business does" view (vs "how" - processes/apps) | MEDIUM | Hierarchical capability taxonomies linked to processes and applications |
| **Gap Analysis** | Compare current vs target architecture to identify missing capabilities/apps | MEDIUM | Requires defining target state versions - can build on lifecycle statuses |
| **Technology Stack Consolidation** | Identify overlapping technologies to reduce sprawl | MEDIUM | Query infrastructure layer + aggregations - show duplicate technologies |
| **Stakeholder Portal** | Read-only views for non-architects (execs, PMs, developers) | LOW | Leverage existing RBAC - create "viewer" role with read-only dashboards |
| **Custom Views/Perspectives** | Filter graph by domain (e.g., show only business+application layers) | MEDIUM | Add domain toggle to graph visualization - filters visible nodes/edges |
| **Cross-Project Impact Analysis** | Show how projects affect each other through shared dependencies | HIGH | Multi-hop traversal through projects→apps→infrastructure relationships |
| **Technical Debt Visualization** | Combine CMDB infrastructure data with EA application risk scoring | HIGH | Requires attribution of technical debt metrics to application components |
| **Data Lineage for Data Objects** | Trace data flow from source systems through transformations to consumption | HIGH | Forward/reverse traversal of data object relationships (usedby, flows to) |
| **M&A Integration Support** | Compare capability maps between organizations to identify overlaps | HIGH | Requires multiple EA repositories or tagging (M&A-specific) |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Automated CMDB Discovery for EA** | "Auto-discovery for applications like we have for servers" | Application discovery is unreliable - business logic boundaries don't map to processes/ports | Manual entry + bulk import with periodic validation |
| **Real-Time Architecture Dashboard** | "Live view of architecture changes" | Creates notification noise; EA is strategic, not operational | Daily/weekly change reports + audit log queries |
| **Separate EA Role Types** | "We need Enterprise Architect role distinct from admin" | Role proliferation; existing roles (architect, PM) can be extended | Add EA permissions to existing roles (ea:create, ea:update, ea:delete) |
| **Full TOGAF ADM Automation** | "Automate all 9 phases of TOGAF" | TOGAF is process framework, not software - cannot be "automated" | Support artifacts created by ADM phases (roadmaps, gap analysis) - don't automate the process |
| **Multi-Tenant EA Repositories** | "Separate EA for each business unit" | Breaks cross-organizational visibility - core EA value is seeing whole enterprise | Use domains/categories/taxonomy tags within single repository |
| **What-If Scenario Branching** | "Create alternate architecture scenarios" | Doubles data model complexity; most scenarios can be modeled as target-state versions | Use lifecycle status (current vs future) + versioning instead of branching |
| **Native Jira/Confluence Integration** | "Bi-directional sync with our Jira projects" | Integration maintenance nightmare; API changes break sync | Export EA data to Jira via CSV/API; don't attempt real-time sync |
| **Custom Modeling Language Designer** | "Let users define their own metamodel" | Eats implementation time; prevents framework standardization | Support ArchiMate 3.x + custom attributes on entities, not custom metamodels |
| **AI-Driven Architecture Recommendations** | "AI tells us what to consolidate" | Black-box decisions architects can't explain or justify | Provide analytics (heatmaps, gap analysis) - let architects make decisions |
| **Full Lifecycle Costing** | "Track TCO for every application" | Requires financial data integration (ERP, procurement) - rarely available | Optional cost attribute on applications - manual entry, not integrated |

## Feature Dependencies

```
[Multi-Domain Metamodel]
    └──requires──> [Entity CRUD Operations]
                    └──requires──> [Cross-Domain Relationships]
                                    └──requires──> [Relationship Visualization]
                                                    └──enhances──> [Basic Impact Analysis]

[Business Capability Mapping]
    └──enhances──> [Application Portfolio Management]
                    └──requires──> [Heat Map Visualization]

[Gap Analysis]
    └──requires──> [Multi-Domain Metamodel]
    └──requires──> [Target State Versioning] (via lifecycle statuses)

[Stakeholder Portal]
    └──requires──> [RBAC]
    └──requires──> [Custom Views/Perspectives]

[Data Lineage for Data Objects]
    └──requires──> [Cross-Domain Relationships]
    └──requires──> [Multi-Hop Impact Analysis] (future differentiator)
```

### Dependency Notes

- **Multi-Domain Metamodel requires Entity CRUD Operations:** Cannot model EA entities without basic CRUD - but CMDB already provides this for CIs, so just extend to EA CI types
- **Cross-Domain Relationships requires Multi-Domain Metamodel:** Cannot link domains if domain entities don't exist
- **Relationship Visualization enhances Basic Impact Analysis:** Users need to see impacted entities visually, not just as lists
- **Application Portfolio Management enhances Heat Map Visualization:** APM provides the metrics (cost, risk, lifecycle) that heatmaps visualize
- **Gap Analysis requires Target State Versioning:** Cannot identify gaps without defining "future" or "target" state - can leverage existing lifecycle_statuses or add "target" status

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **Multi-Domain Metamodel** — Core value; cannot do EA without business, application, data, technology layers
- [x] **Entity CRUD Operations** — Extend CMDB CRUD to EA entity types (capabilities, applications, data objects, infrastructure)
- [x] **Cross-Domain Relationships** — Define EA-specific relationship types (serves, realizes, flows to, assigned to, accesses)
- [x] **Basic Impact Analysis** — 1-hop queries (e.g., "this application affects these business capabilities")
- [x] **Relationship Visualization** — Show EA entities and relationships in existing vis-network graph
- [x] **Bulk Data Import** — CSV import for existing EA inventories (applications, capabilities)
- [x] **RBAC Extension** — Add EA permissions to existing roles (ea:read, ea:update, ea:delete)
- [x] **Audit Trail** — Already exists; ensure EA changes are logged (automatic via CMDB)

**Rationale:** These features enable core EA use cases:
- Model the enterprise (metamodel + CRUD + relationships)
- Understand impact (basic impact analysis + visualization)
- Populate initial data (bulk import)
- Govern changes (RBAC + audit)

### Add After Validation (v1.x)

Features to add once core is working and users request them.

- [ ] **Business Capability Mapping** — Explicit capability hierarchies; requested once users see basic applications
- [ ] **Heat Map Visualization** — Color-coded views for maturity/risk; requested for executive dashboards
- [ ] **Stakeholder Portal** — Read-only dashboards for PMs/execs; requested once users collaborate on EA
- [ ] **Custom Views/Perspectives** — Filter graph by domain; requested as EA repository grows
- [ ] **Gap Analysis** — Current vs target state comparison; requires defining "target" lifecycle status
- [ ] **Technology Stack Consolidation** — Identify duplicate technologies; requested for cost optimization
- [ ] **Search & Filtering by Domain** — Add domain/category filters to existing CMDB search

**Triggers:** Add when users ask "Can I see...", "How do I compare...", "Can I share with..."

### Future Consideration (v2+)

Features to defer until product-market fit is established and team capacity allows.

- [ ] **Application Portfolio Management (APM)** — HIGH complexity; requires application lifecycle, criticality scoring, cost attribution
- [ ] **Cross-Project Impact Analysis** — HIGH complexity; multi-hop traversal + project management data
- [ ] **Data Lineage for Data Objects** — HIGH complexity; forward/reverse traversal of data flows
- [ ] **M&A Integration Support** — HIGH complexity; requires multi-repository comparison or advanced tagging
- [ ] **Technical Debt Visualization** — HIGH complexity; requires external tool integration (SonarQube, etc.) + attribution
- [ ] **Multi-Hop Impact Analysis** — MEDIUM complexity; deeper traversal algorithms (beyond 1-hop)
- [ ] **What-If Scenario Modeling** — HIGH complexity; requires duplicating architecture states or versioning

**Rationale:** These are powerful but complex. Defer until v1 proves value and users demand advanced analytics.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Multi-Domain Metamodel | HIGH | HIGH | P1 |
| Entity CRUD Operations | HIGH | LOW | P1 |
| Cross-Domain Relationships | HIGH | MEDIUM | P1 |
| Basic Impact Analysis | HIGH | MEDIUM | P1 |
| Relationship Visualization | HIGH | MEDIUM | P1 |
| Bulk Data Import | HIGH | MEDIUM | P1 |
| RBAC Extension | HIGH | LOW | P1 |
| Audit Trail | HIGH | LOW | P1 |
| Business Capability Mapping | MEDIUM | MEDIUM | P2 |
| Heat Map Visualization | MEDIUM | MEDIUM | P2 |
| Stakeholder Portal | MEDIUM | LOW | P2 |
| Custom Views/Perspectives | MEDIUM | MEDIUM | P2 |
| Gap Analysis | MEDIUM | MEDIUM | P2 |
| Technology Stack Consolidation | MEDIUM | MEDIUM | P2 |
| Application Portfolio Management | HIGH | HIGH | P2 |
| Cross-Project Impact Analysis | MEDIUM | HIGH | P3 |
| Data Lineage for Data Objects | MEDIUM | HIGH | P3 |
| M&A Integration Support | LOW | HIGH | P3 |
| Technical Debt Visualization | MEDIUM | HIGH | P3 |
| Multi-Hop Impact Analysis | MEDIUM | HIGH | P3 |
| What-If Scenario Modeling | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch (table stakes + core value)
- P2: Should have, add when possible (differentiators users request)
- P3: Nice to have, future consideration (high complexity, niche use cases)

## Competitor Feature Analysis

| Feature | LeanIX | Ardoq | Planview | Pustaka EA (Our Approach) |
|---------|--------|-------|----------|---------------------------|
| **Metamodel Framework** | Proprietary | Proprietary | Proprietary | ArchiMate 3.x-based (standard) |
| **Visualization** | Proprietary "Google Maps for IT" | Network & data flow graphs | Interactive analytics | vis-network (existing CMDB) |
| **Collaboration** | Confluence/Jira integration | Democratized data collection | PPM integration | Stakeholder portal (planned v1.x) |
| **APM** | Cloud intelligence, application lifecycle | Digital twin creation | Business capability mapping | Deferred to v1.x |
| **Analytics** | Fact sheets, metadata visualization | Simulation capabilities | Predictive outcomes | Heat maps (v1.x) |
| **Cost** | Enterprise SaaS pricing | Enterprise SaaS pricing | Enterprise SaaS pricing | Open source, self-hosted |
| **Discovery** | Cloud-native auto-discovery | Cloud integration | Manual | Manual + bulk import |
| **Governance** | Role-based access | Role-based access | Role-based access | Extend existing RBAC |

**Our Differentiation Strategy:**
- **Open Source:** No licensing costs (unlike $100K+ enterprise tools)
- **ArchiMate Standard:** Not proprietary metamodel - easier to hire architects with standard skills
- **CMDB Integration:** Native bi-directional link between EA entities and infrastructure CIs (most tools import or don't integrate)
- **Lightweight:** Focus on modeling + impact analysis vs bloated analytics suites

## Sources

- [Enterprise Architecture Tools Comparison 2026 - LeanIX, Ardoq, Planview](https://www.softwaretestinghelp.com/enterprise-architecture-tools/) (HIGH confidence - comprehensive tool comparison)
- [Enterprise Architecture Capabilities Beyond CMDB](https://www.opengroup.org/ogmc/resources) (MEDIUM confidence - The Open Group resources on EA-CMDB integration)
- [TOGAF ADM Support in EA Tools](https://www.visual-paradigm.com/togaf/adm/) (HIGH confidence - Visual Paradigm TOGAF documentation)
- [Business Capability Mapping and Heat Map Analysis](https://www.mega.com/resources/business-capability-mapping/) (MEDIUM confidence - MEGA International guidance)
- [ArchiMate 3.1 Specification - Application Layer](https://pubs.opengroup.org/architecture/archimate3-doc/chap09.html) (HIGH confidence - Official Open Group specification)
- [EA Tool Collaboration Features - Prolaborate](https://www.prolaborate.com/collaboration) (MEDIUM confidence - Vendor-specific collaboration features)
- [EA Governance Lifecycle and Workflows](https://www.opengroup.org/togaf/) (HIGH confidence - TOGAF governance guidance)
- [CMDB vs Enterprise Architecture Differences](https://www.servicecomb.com/cmdb-vs-ea/) (MEDIUM confidence - Industry comparison article)
- [Common EA Pitfalls and Mistakes](https://www.gartner.com/en/information-technology/insights/enterprise-architecture-pitfalls) (MEDIUM confidence - Gartner research on EA failures)
- [Open Source EA Tools - Archi](https://www.archimatetool.com/) (HIGH confidence - Leading open-source EA tool)
- [Data Lineage in Enterprise Architecture](https://www.ibm.com/topics/data-lineage) (HIGH confidence - IBM data lineage documentation)
- [Technical Debt Heatmap and Risk Assessment](https://www.sonarsource.com/solutions/technical-debt/) (MEDIUM confidence - SonarQube technical debt practices)

---
*Feature research for: Enterprise Architecture Module*
*Researched: 2026-02-20*

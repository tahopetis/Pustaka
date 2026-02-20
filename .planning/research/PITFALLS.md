# Pitfalls Research

**Domain:** Enterprise Architecture module extending CMDB
**Researched:** 2026-02-20
**Confidence:** MEDIUM

## Critical Pitfalls

### Pitfall 1: Data Quality Death Spiral

**What goes wrong:**
EA module data becomes stale within 3-6 months, accuracy drops below 60%, users abandon the system for spreadsheets, and the CMDB reverts to being a simple asset registry. The "garbage in, garbage out" problem compounds as relationship data rots faster than asset data.

**Why it happens:**
- No clear ownership for EA entity updates (responsibility assigned to "everyone" = no one)
- Manual data entry without automated discovery/validation
- Lack of integration with change management processes
- Treating EA data as one-time setup rather than living asset
- Resistance from teams who see EA as "additional paperwork" without immediate value

**How to avoid:**
- Define clear data ownership during metamodel creation (each entity type has an owner role)
- Build automated validation from day one: required fields, relationship constraints, data quality dashboards
- Integrate EA entity lifecycle with ITSM change processes (changes update EA automatically)
- Start with high-value, low-volume domains first (prove value before expanding)
- Implement data quality KPIs with executive visibility

**Warning signs:**
- Relationship query returns different results on successive runs
- Users keeping "shadow spreadsheets" alongside the system
- Automated discovery showing >30% discrepancy from manual records
- Declining API call volume from other systems (they're bypassing your data)

**Phase to address:**
Phase 1 (Foundation) — Data quality mechanisms must be built into the core, not bolted on later. The metamodel design phase is when ownership and validation rules are defined.

---

### Pitfall 2: Relationship Modeling Bloat

**What goes wrong:**
Attempting to model every possible relationship type from day one creates an unmaintainable web. Users face 80+ relationship types, can't distinguish between critical and trivial connections, and the graph becomes unusable for impact analysis. Performance degrades as Neo4j traversals explode combinatorially.

**Why it happens:**
- Metamodel designed by committee (every stakeholder adds "their" relationships)
- Following TOGAF/ArchiMate too literally without pragmatic simplification
- Belief that "more relationships = better architecture"
- Failure to prioritize based on actual use cases (impact analysis vs. documentation)

**How to avoid:**
- Start with 20-25 core relationship types that support 3-5 critical use cases
- Use the "5-hop rule": if a relationship isn't used in >5-hop impact analysis, make it optional
- Implement relationship tiers: Core (always visible), Extended (opt-in), Historical (audit only)
- Add relationships incrementally based on user demand, not theoretical completeness
- Create relationship style guides with clear examples of when to use each type

**Warning signs:**
- Graph visualization shows "hairball" with >500 nodes and >2000 edges
- Users frequently asking "which relationship type should I use for X?"
- Impact analysis queries timing out (>10 seconds for 3-hop traversal)
- Relationship selection dropdown taking >5 seconds to load

**Phase to address:**
Phase 1 (Foundation) — Core metamodel must be minimal. Phase 3 (Advanced Analytics) can add extended relationships once the core is stable and valuable.

---

### Pitfall 3: CMDB Integration Through, Not Alongside

**What goes wrong:**
Treating EA entities as "separate but related" to existing CMDB CIs creates parallel universes. Applications in EA don't match applications in infrastructure CMDB, relationships conflict, and users must reconcile two systems. Impact analysis fails because it only sees EA entities, not the actual infrastructure CIs they depend on.

**Why it happens:**
- Fear of "polluting" the pure CMDB with EA abstractions
- Organizational silos (EA team vs. Operations team)
- Treating EA as a separate capability rather than an extension
- Lack of bidirectional relationship design in the metamodel

**How to avoid:**
- Design EA entities AS CI Types from day one (unified taxonomy)
- Create explicit bidirectional relationships: Application→Server (deployment) AND Server→Application (hosts)
- Use CI Type attributes to flag EA-relevant entities (is_architectural=true)
- Implement "architectural views" as filtered queries, not separate data stores
- Reuse existing CMDB RBAC, audit logging, and relationship infrastructure

**Warning signs:**
- Duplicate entities (same application in EA module and infrastructure CMDB)
- Users manually maintaining "mapping tables" between EA and CMDB
- Impact analysis missing critical infrastructure dependencies
- Separate permission systems for EA vs. CMDB data

**Phase to address:**
Phase 1 (Foundation) — The integration model is decided before metamodel design. EA entities are CI Types, not parallel entities.

---

### Pitfall 4: Governance Vacuum

**What goes wrong:**
Without architectural governance, the EA module becomes a free-for-all. Teams create contradictory relationships, entities proliferate without standards, and the data loses trust. The system reflects organizational politics rather than architectural reality. "My application depends on your database" becomes a negotiation tool rather than factual dependency tracking.

**Why it happens:**
- Implementing the tool without the process (classic TOGAF mistake)
- Governance viewed as "slowing down innovation" rather than "preventing chaos"
- No executive mandate for architectural standards
- Unclear decision rights: who approves relationship changes? who arbitrates conflicts?

**How to avoid:**
- Define EA governance board BEFORE launching the module (charter, decision rights, escalation path)
- Implement approval workflows for entity creation and critical relationship changes
- Create architectural principles that guide modeling (e.g., "no direct dependency between applications in different domains")
- Start with lightweight governance (review 10% of changes randomly) and scale based on need
- Make governance data-driven (dashboards show modeling compliance, not just counts)

**Warning signs:**
- Same dependency modeled differently by different teams
- Entities created without following naming conventions
- "My app depends on everything" anti-pattern (uninformative relationships)
- Executive questions about system architecture going unanswered despite EA module existing

**Phase to address:**
Phase 1 (Foundation) — Governance structure is defined alongside the metamodel. Phase 2 (Operations) implements the governance workflows.

---

### Pitfall 5: Premature Optimization

**What goes wrong:**
Building advanced analytics, what-if scenarios, and predictive capabilities before the data is reliable. Queries return nonsense, stakeholders lose faith, and the project is cancelled before delivering value. The foundation is crumbling but we're building a penthouse.

**Why it happens:**
- Desire to show "strategic value" to executives quickly
- Confusing "cool features" with "useful features"
- Underestimating data quality challenges
- Vendor demos showing mature capabilities that assume 5 years of clean data

**How to avoid:**
- Phase 1 focuses ONLY on: accurate entities, core relationships, basic CRUD, simple graph viz
- Define success criteria as data quality metrics, not feature count
- Resist "wouldn't it be cool if" requests until data quality >90%
- Pilot advanced features with a small trusted group before broad rollout
- Celebrate "boring" wins: "All application dependencies are now accurate" > "We have AI-powered impact prediction"

**Warning signs:**
- Stakeholders requesting advanced analytics while basic queries return inconsistent results
- Demo environments showing impressive capabilities that don't work in production
- Development velocity focused on new features rather than data quality improvements
- Executive excitement about dashboards that no one trusts

**Phase to address:**
All phases — Each phase has a "data quality gate" that must be passed before moving to the next. Phase 1 requires >85% accuracy on core entities before Phase 2 begins.

---

### Pitfall 6: The Framework Trap

**What goes wrong:**
Implementing TOGAF/ArchiMate too rigidly creates a compliant but useless system. Entities and relationships exist because "the framework says so" rather than because they solve real problems. Architects spend hours debating metamodel nuances while operational teams can't find the answers they need.

**Why it happens:**
- Fear of deviating from "industry standards"
- Consultants selling framework implementation rather than problem-solving
- Mistaking framework completeness for practical completeness
- Lack of courage to simplify: "But TOGAF says we need X"

**How to avoid:**
- Treat TOGAF/ArchiMate as input, not requirements
- Start with 5-10 real questions the EA module must answer (e.g., "If this database goes down, which business capabilities are affected?")
- Design metamodel backward from answers, not forward from framework
- Create "pragmatic deviations" log where you intentionally deviate from framework with rationale
- Review and simplify every 6 months: remove entities/relationships that aren't used

**Warning signs:**
- Diagrams that look "impressive" to architects but confuse everyone else
- Metamodel review meetings that devolve into framework theology debates
- Entities/relationships with no documented use case
- "We need this because ArchiMate 3.1 includes it"

**Phase to address:**
Phase 1 (Foundation) — Use-case driven metamodel design, not framework-driven. Phase 4 (Optimization) is when framework alignment is assessed, not Phase 1.

---

### Pitfall 7: Visualization Without Navigation

**What goes wrong:**
Beautiful graph visualizations that can't answer questions. Users see colorful node-link diagrams but can't find specific applications, can't filter meaningfully, and can't export actionable reports. The "wow factor" of the demo doesn't translate to daily utility.

**Why it happens:**
- Over-investing in vis-network configuration at the expense of query APIs
- Designing for executive presentations rather than daily workflows
- Treating visualization as an end rather than a means
- Lack of user testing: demos impress but real users struggle

**How to avoid:**
- Design UI workflows first: "How does an architect find all dependencies of Application X?" then add visualization
- Implement powerful search/filter before fancy visual layouts
- Create multiple visualization modes: detailed (for debugging) vs. simplified (for presentations)
- Test with real users on real tasks, not demo scenarios
- Provide exportable reports (CSV, PDF) alongside interactive graphs

**Warning signs:**
- Demo screenshots look great but daily users complain about "finding anything"
- Visualization taking >10 seconds to render for >100 nodes
- No search functionality in graph view
- Users taking screenshots and manually annotating them because the tool doesn't support their workflow

**Phase to address:**
Phase 2 (User Experience) — Navigation and search capabilities are prioritized over advanced visualization features.

---

### Pitfall 8: Permission Proliferation

**What goes wrong:**
Creating granular EA-specific permissions (ea:strategy:read, ea:application:write, ea:relationship:delete) that don't align with existing RBAC. Users have 15+ permissions across CMDB and EA, no one understands what they can access, and permission requests become a full-time job.

**Why it happens:**
- Treating EA as a separate security domain
- "Least privilege" taken to an extreme without pragmatism
- Lack of trust in existing CMDB RBAC structure
- Designing permissions based on theoretical roles rather than actual teams

**How to avoid:**
- Extend existing CMDB roles rather than create new EA-specific roles
- Use 3-4 coarse-grained permissions: ea:read, ea:write, ea:admin, ea:govern
- Leverage existing CI Type permissions (if user can ci:write, they can ea:write for architectural entities)
- Implement role templates for common personas (architect, PM, operations) that bundle permissions
- Document permission model in one page, not twenty

**Warning signs:**
- New EA permissions requested weekly
- Users unable to explain what they're allowed to do
- Permission-related support tickets >20% of total
- Duplicate permission checks in code (ea:read AND ci:read for the same resource)

**Phase to address:**
Phase 1 (Foundation) — Permissions are designed as extensions of existing CMDB RBAC, not a new system.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Manual entity creation to test metamodel | Quick validation of data model | Ongoing manual maintenance, data quality rot | ONLY for prototype, must be automated before Phase 1 completion |
| Hardcoded relationship types in backend | Faster initial development, no migration system | Cannot add relationship types without code changes, limits extensibility | NEVER - relationship type system must be data-driven from day one |
| Skipping relationship validation during data import | Faster bulk load, no constraint errors | Cascading data quality issues, impossible to debug later | NEVER - validation must be synchronous, fail fast on bad data |
| Using CMDB's generic audit logs for EA events | No new development needed | Cannot distinguish EA changes from infrastructure changes, poor governance reporting | Acceptable for Phase 1 (MVP), must add EA-specific event types in Phase 2 |
| Limiting graph traversal depth to 2 hops to avoid performance issues | Acceptable query performance in early phases | Cannot support complex impact analysis, limited value | Acceptable for Phase 1-2, Phase 3 requires performance optimization not feature limitations |
| Storing relationship attributes as JSON strings | Flexible schema, easy to add metadata | Cannot query on relationship attributes, no type safety | Acceptable for Phase 1, Phase 2 requires structured attributes for filtering/search |

## Integration Gotchas

Common mistakes when connecting EA to external systems.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| ServiceNow CMDB | Creating parallel EA entities in ServiceNow, then trying to sync | Model EA entities as extended CMDB classes in ServiceNow, single source of truth |
| ITSM (ServiceNow/Jira) | Treating EA entity changes as regular changes, no governance review | Create EA-specific change types with mandatory architectural review before approval |
| Monitoring (Prometheus/DataDog) | Importing all monitored services as EA applications without curation | Define mapping rules: only production services with SLAs become EA applications |
| Asset Management | Importing all purchased software as EA applications | EA applications = business-facing capabilities, not installed software |
| Cloud Providers (AWS/Azure) | Auto-discovery creates thousands of temporary entities (lambda functions, auto-scaling groups) | Discovery rules filter for persistent architectural entities only (long-lived services, databases) |
| Project Portfolio Management | Creating EA entities for every project regardless of maturity | Only create EA entities for approved projects in execution/implementation phase |
| Compliance Tools (Drata/Vanta) | Marking every control as a governance requirement in EA | Map controls to EA entities only where there's architectural impact (data storage, authentication services) |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| N+1 query pattern in relationship loading | Graph visualization takes >30 seconds, 1000s of database queries | Use batch loading, pre-fetch relationships, implement GraphQL-style query planning | 500+ entities with 10+ relationships each |
| Neo4j recursive queries without depth limits | Database hangs, memory exhaustion on complex traversals | Hard limit traversal depth (max 5 hops), use breadth-first search with early termination | >10,000 relationships, deeply nested hierarchies |
| No caching of frequently accessed entities | Same entity queried 100s of times per dashboard load | Redis caching of popular entities (top 20% accessed 80% of time), TTL 5 minutes | >100 concurrent users, dashboard-heavy usage |
| Relationship validation during bulk import | Importing 10,000 entities takes >1 hour | Defer validation to async job, show progress, validate in batches | >1000 entities per import |
| Full-text search on entity names without indexing | Search takes >10 seconds, database CPU spikes | Add PostgreSQL full-text indexes, implement autocomplete with prefix search | >5000 entities |
| Graph visualization recalculating layout on every node addition | Adding nodes freezes browser for >5 seconds | Use incremental layout updates, web workers for layout calculation, render >500 nodes in simplified mode | >200 nodes in graph view |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| EA entities visible to all authenticated users by default | Unauthorized access to strategic initiatives, M&A plans, security architecture | Default-deny permissions: EA entities require explicit ea:read permission, not authenticated access |
| Relationship history not tracked | tampering with dependency graphs to hide issues or cover up outages | Immutable audit logs for all relationship changes, legal hold retention for critical relationships |
| No data masking for exported reports | Sensitive architecture (security controls, encryption details) leaked in exported CSVs | Apply RBAC to exports, mask sensitive attributes (security classifications, key names) in non-admin exports |
| GraphQL-like query API without depth limits | Denial of service via intentionally complex nested queries | Hard limit query depth, cost-based query planning, rate limiting per user |
| Bulk export without approval mechanism | Data exfiltration of entire architecture by departing employees | Require admin approval for exports >1000 entities, log all bulk exports, quarantine for review |
| Authentication tokens valid for EA API but not CMDB API | Privilege escalation if EA permissions are more permissive | Unified auth tokens, EA permissions inherit from CMDB base permissions |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| "Show me everything" graph views | Users overwhelmed, cannot find relevant information, abandon tool | Context-aware filtering: when viewing Application X, only show direct dependencies (1 hop) by default, expand on demand |
| Technical entity names (app-order-mgmt-prod-v2) | Business stakeholders cannot understand architecture diagrams | Support display name aliases, show business-friendly names in architecture views, technical names as detail |
| Separate "create" and "edit" forms with different layouts | User confusion about which form to use, inconsistent data entry | Unified entity form with conditional fields based on creation vs. edit, consistent layout |
| No search in graph visualization | Users visually hunt for nodes in 500+ node graphs | Global search bar highlights matching nodes, auto-center on result, filter to show only matches |
| Relationship types as cryptic codes (REL_001, REL_002) | Users don't understand relationships, choose wrong types | Human-readable names with descriptions ("Application depends on Database"), show example of correct usage in tooltip |
| Saving graph layouts per-user without sharing | Team collaboration requires screenshot sharing, cannot build on each other's work | Shared graph layouts (team, org, public levels), auto-save curated views for common questions |
| No undo for relationship deletion | Accidental deletion cascades, hours of manual restoration | Soft delete (archive), restore relationship within 30 days, show "recently deleted" panel |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Relationship validation**: Often missing circular dependency detection — verify "App A depends on App B depends on App A" is flagged or blocked
- [ ] **Entity lifecycle management**: Often missing draft/published/retired states — verify entities aren't visible until approved, have retirement workflows
- [ ] **Bulk relationship editing**: Often missing — verify architect can select 10 applications and assign them to a new business capability in one action
- [ ] **Impact analysis export**: Often missing — verify "show me all business capabilities affected by Database X" can be exported to PDF/CSV
- [ ] **Change history rollback**: Often missing — verify relationship changes can be reverted to previous state, not just viewed in audit log
- [ ] **Permission inheritance**: Often missing — verify if user can edit Applications, they can also edit Application-to-Database relationships (not separate permissions)
- [ ] **Data quality dashboard**: Often missing — verify admins can see % completeness, % stale, validation errors by domain
- [ ] **Entity merge**: Often missing — verify duplicate entities (two "Customer Service App" entries) can be merged with relationship consolidation
- [ ] **API rate limiting**: Often missing — verify bulk imports don't starve interactive queries, implement query prioritization
- [ ] **Graph layout persistence**: Often missing — verify users can save "Application Dependency Map" layout and return to it later, positions remembered

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Data quality below 60% | HIGH | 1. Declare "data amnesty" — pause governance for 30 days, 2. Automated discovery to identify discrepancies, 3. Stakeholder workshops to manually verify top 100 critical entities, 4. Delete or archive entities without owners, 5. Re-launch with mandatory ownership requirement |
| Relationship bloat (>100 types) | MEDIUM | 1. Analyze usage metrics, identify unused relationship types, 2. Migrate low-value relationships to "extended" category, 3. Deprecate unused types (read-only, warn on use), 4. Remove after 6-month grace period, 5. Document relationship type addition criteria |
| CMDB/EA duplicate entities | HIGH | 1. Entity reconciliation service to detect duplicates (fuzzy name match, attribute similarity), 2. Manual review of merge candidates, 3. Merge operation consolidates relationships, 4. Audit trail of merged entities for rollback, 5. Prevent future duplicates via mandatory type classification |
| Governance failure | MEDIUM | 1. Executive mandate re-affirming EA governance authority, 2. Pause entity creation to non-architects, 3. Implement mandatory review workflow for critical changes, 4. Create escalation path for conflicts, 5. Publish governance metrics (compliance, review time) |
| Performance degradation (query timeouts) | MEDIUM | 1. Add query depth limits (max 3 hops for interactive queries), 2. Implement caching layer for popular entities, 3. Add database indexes on frequently-filtered attributes, 4. Background job to pre-calculate common traversals, 5. User education on query complexity |
| Permission confusion | LOW | 1. Audit current permissions, document actual usage vs. design, 2. Consolidate granular permissions into role templates, 3. Migrate users to new simplified roles, 4. Remove unused permissions, 5. Update documentation with examples |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Data quality death spiral | Phase 1: Define ownership, validation rules, automated discovery | Phase 1 signoff: >85% data accuracy on core entities, owners assigned for all entity types |
| Relationship modeling bloat | Phase 1: Minimal metamodel (20-25 core relationships) | Phase 1 signoff: Metamodel reviewed against use cases, no framework-driven entities |
| CMDB integration failure | Phase 1: EA entities as CI Types, bidirectional relationships | Phase 1 signoff: Single unified taxonomy, no duplicate entities in test data |
| Governance vacuum | Phase 1: Governance board charter, Phase 2: Workflow implementation | Phase 2 signoff: 10% random audit shows <5% non-compliant relationships |
| Premature optimization | All phases: Data quality gates between phases | Phase gates blocked if quality <85%, regardless of feature completeness |
| Framework trap | Phase 1: Use-case driven metamodel, Phase 4: Framework alignment review | Phase 1 signoff: All entities/relationships map to answering 5-10 critical questions |
| Visualization without navigation | Phase 2: UX workflows tested before advanced visualization | Phase 2 signoff: User testing shows navigation tasks <30 seconds, <3 clicks |
| Permission proliferation | Phase 1: Extend existing RBAC, 3-4 coarse-grained EA permissions | Phase 1 signoff: Permission model fits on one page, understood by pilot users |
| N+1 query performance trap | Phase 1: Performance testing with 1000+ entities, 10+ relationships | Phase 1 signoff: Graph queries <3 seconds for typical workloads |
| No data quality dashboard | Phase 2: Admin observability, quality metrics | Phase 2 signoff: Quality dashboard shows completeness, staleness, errors |

## Sources

### CMDB Data Quality & Maintenance
- "CMDB数据质量管理：从'垃圾进，垃圾出'到精准运维的蜕变之路" (2025) - Data quality challenges, 70% of CMDB projects face quality issues
- "初识CMDB：数据质量的管理" (2025) - Data accuracy, completeness, consistency, timeliness dimensions
- "企业CMDB配置管理系统：从痛点出发的实战构建指南" (2025) - Common pitfalls: data silos, change management gaps, relationship chaos
- "重构CMDB，避免运维之耻" (2019, archived) - Organizational design issues, Excel dependencies, application-centric vs. infrastructure-centric
- "从某大型企业实践看CMDB建设的核心问题" (2025) - Value proposition, data governance, organizational resistance

### Enterprise Architecture Frameworks
- "TOGAF难以落地吗？组织常见的误解" (2026) - Common misapplications: treating as linear process, producing artifacts without decision purpose
- "企业架构的十个误区" (2025) - Strategic disconnection from execution, PPT architectures without implementation detail
- "TOGAF架构元模型详解" (2025) - Core metamodel concepts, traceability, consistency requirements
- "Salesforce反模式" (2025) - Governance anti-patterns, stakeholder management, architectural governance integration

### Relationship & Data Modeling
- "指标建模有哪些常见误区" (2025-2026) - Modeling errors: inconsistency, complexity without value, disconnect from business scenarios
- "数据建模有哪些常见错误" (2025) - Requirements misalignment, 60% of data modeling failures from poor requirements understanding
- "ER模型设计概念梳理" (2025) - Industry ER modeling pitfalls: entity confusion, incorrect cardinality, relationship design errors

### Performance & Scalability
- "Getting started with chaos engineering" (2025) - Resilience testing for distributed systems, failure injection
- "Optimized predictive maintenance for streaming data" (2025) - Industrial IoT challenges with dynamic data, scale issues

### General CMDB & EA Integration
- "2025年值得挑选的几款CMDB软件推荐" (2025) - CMDB selection criteria, data quality sustainability, integration capabilities
- "Patterns of Enterprise Application Architecture" (Martin Fowler) - Enterprise architecture patterns, though dated (2002), principles remain relevant

### Confidence Assessment
- **HIGH**: Data quality challenges (multiple Chinese industry sources 2025), TOGAF implementation pitfalls (official Open Group guidance)
- **MEDIUM**: EA-CMDB integration patterns (limited documented case studies, inferred from first principles), performance scalability (Neo4j best practices but limited EA-specific benchmarks)
- **LOW**: Specific quantitative thresholds (e.g., "85% accuracy" based on general CMDB best practices, not EA-specific studies)

---
*Pitfalls research for: Enterprise Architecture module extending CMDB*
*Researched: 2026-02-20*

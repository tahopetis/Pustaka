# Architecture Research

**Domain:** Enterprise Architecture Module (extending existing CMDB)
**Researched:** 2026-02-20
**Confidence:** HIGH

## Standard Architecture

### System Overview

The Enterprise Architecture (EA) module extends Pustaka's existing CMDB architecture without creating a separate parallel system. EA entities are modeled as Configuration Items within the established CI taxonomy, leveraging all existing infrastructure (PostgreSQL, Neo4j, Redis, RBAC, audit logging).

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Frontend Layer (Vue 3)                           │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │ EA Entities │  │ EA Relation │  │ Graph View  │  │  Import UI  │  │
│  │   (CRUD)    │  │ Management  │  │             │  │             │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │
│         │                │                │                │          │
├─────────┼────────────────┼────────────────┼────────────────┼──────────┤
│         │      ┌─────────┴─────────┐      │                │          │
│         │      │ Pinia Stores      │      │                │          │
│         │      │ - eaTypes.ts      │      │                │          │
│         │      │ - eaEntities.ts   │      │                │          │
│         │      │ - eaImport.ts     │      │                │          │
│         │      └─────────┬─────────┘      │                │          │
├─────────┼────────────────┼────────────────┼────────────────┼──────────┤
│         │      ┌─────────┴─────────┐      │                │          │
│         │      │ API Service Layer │      │                │          │
│         │      │ - eaAPI.ts        │      │                │          │
│         │      └─────────┬─────────┘      │                │          │
├─────────┼────────────────┼────────────────┼────────────────┼──────────┤
│         │                │                │                │          │
│         ▼                ▼                ▼                ▼          │
│  ┌──────────────────────────────────────────────────────────────────┐ │
│  │                    HTTP API Layer (Chi v5)                       │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │ │
│  │  │  EA Entity  │  │  EA Import  │  │   EA Graph  │              │ │
│  │  │  Handlers   │  │  Handlers   │  │  Handlers   │              │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │ │
│  └──────────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────┤
│                        Service Layer (Go)                              │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                     EA Service Package                          │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │  │
│  │  │ EA Entity    │  │ EA Import    │  │ EA Graph     │         │  │
│  │  │ Service      │  │ Service      │  │ Service      │         │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │  │
│  └─────────┼─────────────────┼─────────────────┼───────────────────┘  │
│            │                 │                 │                      │
├────────────┼─────────────────┼─────────────────┼──────────────────────┤
│            │      ┌──────────┴──────────┐      │                      │
│            │      │ CI Service (Reuse)  │      │                      │
│            │      │ - CI CRUD           │      │                      │
│            │      │ - Validation        │      │                      │
│            │      │ - Audit Logging     │      │                      │
│            │      └──────────┬──────────┘      │                      │
├────────────┼─────────────────┼─────────────────┼──────────────────────┤
│            │                 │                 │                      │
│  ┌─────────▼─────────┐ ┌────▼──────────┐ ┌────▼──────────┐           │
│  │ PostgreSQL Repo  │ │  Neo4j Repo    │ │  Redis Cache  │           │
│  │  (EA CI Types)   │ │  (Relations)   │ │  (EACache)    │           │
│  └───────────────────┘ └───────────────┘ └───────────────┘           │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Architectural Insight:** EA module uses **extension by composition**, not modification. The EA service layer wraps and extends the existing CI service, treating EA entities as specialized CI types with domain-specific validation and relationship rules.

### Component Responsibilities

| Component | Responsibility | Communicates With |
|-----------|----------------|-------------------|
| **Frontend Views** | Entity CRUD forms, relationship editors, graph visualization, bulk import UI | Pinia stores, API services |
| **Pinia Stores** | State management for EA entities, types, relationships, import state | Frontend views, API service layer |
| **API Service Layer** | HTTP client for EA endpoints, request/response transformation | Backend HTTP handlers |
| **HTTP Handlers** | Request validation, JWT auth, RBAC enforcement, response formatting | EA service layer, CI service layer |
| **EA Service Layer** | Domain logic for EA entities, import processing, graph queries | CI service layer, repositories |
| **CI Service Layer** | Generic CI CRUD, validation, audit logging (reused) | PostgreSQL repository, Neo4j service, Redis |
| **PostgreSQL Repository** | EA CI type storage, EA entity persistence (via CI tables) | Service layer |
| **Neo4j Repository** | EA relationship storage and traversal | Service layer, graph visualization |
| **Redis Cache** | EA type definitions cache, frequently accessed entities | Service layer |

## Recommended Project Structure

```
internal/
├── api/
│   ├── handlers/
│   │   ├── ea_handlers.go           # EA entity CRUD endpoints
│   │   ├── ea_import_handlers.go    # Bulk import endpoints
│   │   └── ea_graph_handlers.go     # EA-specific graph queries
│   └── middleware/
│       └── rbac.go                  # Extended with EA permissions
│
├── ea/                               # NEW: EA-specific business logic
│   ├── models.go                    # EA-specific types, domain enums
│   ├── service.go                   # EA entity service (wraps CI service)
│   ├── validation.go                # EA cross-domain validation rules
│   ├── import.go                    # Bulk import logic (CSV parsing)
│   ├── graph_service.go             # EA-specific graph queries
│   └── domain_service.go            # Per-domain business logic
│
├── ci/                               # EXISTING: CMDB core (no changes)
│   ├── models.go                    # ConfigurationItem, Relationship
│   ├── service.go                   # CI CRUD (reused by EA)
│   ├── repository.go                # PostgreSQL operations
│   ├── neo4j_repository.go          # Neo4j operations
│   └── audit_*.go                   # Audit logging (reused)
│
├── database/                         # EXISTING: Database connections
│   ├── postgres.go
│   ├── neo4j.go
│   └── redis.go
│
└── auth/                             # EXISTING: Authentication/authorization
    ├── rbac.go                      # Extended with EA permissions
    └── jwt.go

web/src/
├── views/ea/                         # NEW: EA-specific views
│   ├── entities/                    # Entity CRUD by domain
│   │   ├── strategy/
│   │   │   ├── ObjectiveListView.vue
│   │   │   ├── ObjectiveFormView.vue
│   │   │   └── ...
│   │   ├── business/
│   │   ├── application/
│   │   ├── data/
│   │   ├── technology/
│   │   ├── infrastructure/
│   │   ├── security/
│   │   └── governance/
│   ├── relationships/               # Relationship management
│   │   ├── RelationshipListView.vue
│   │   ├── RelationshipFormView.vue
│   │   └── RelationshipDetailsView.vue
│   ├── graph/                       # EA graph visualization
│   │   └── EAGraphView.vue
│   └── import/                      # Bulk import UI
│       └── ImportView.vue
│
├── stores/                           # NEW: EA state management
│   ├── eaTypes.ts                   # EA type definitions
│   ├── eaEntities.ts                # EA entity state
│   ├── eaRelationships.ts           # EA relationships
│   └── eaImport.ts                  # Import state
│
├── services/
│   └── eaAPI.ts                     # NEW: EA API client
│
└── router/                           # EXISTING: Vue Router
    └── index.ts                     # Extended with EA routes
```

### Structure Rationale

- **`internal/ea/` package**: Isolates EA domain logic from generic CMDB logic. Wraps CI service rather than modifying it, maintaining separation of concerns.
- **Handler extension**: New `ea_handlers.go` for EA-specific endpoints, reusing middleware stack (JWT, RBAC, audit logging).
- **Frontend domain organization**: Views grouped by EA domain (strategy/, business/, application/) for clear navigation and maintainability.
- **Store separation**: EA-specific stores (`eaTypes.ts`, `eaEntities.ts`) separate from existing CI stores to avoid pollution while maintaining same patterns.

## Architectural Patterns

### Pattern 1: Service Layer Composition

**What:** EA service wraps and extends CI service rather than modifying it. EA entities are CIs with domain-specific validation and relationship rules.

**When to use:** When extending a generic system with domain-specific requirements without breaking existing functionality.

**Trade-offs:**
- **Pros:** Clean separation, CI service remains reusable, EA logic is isolated, easier to test.
- **Cons:** Additional abstraction layer, potential for wrapper overhead (minimal in Go).

**Example:**
```go
// internal/ea/service.go
type Service struct {
    ciService    *ci.Service          // Composed CI service
    neo4j        *ci.Neo4jService     // Reuse Neo4j operations
    validator    *DomainValidator     // EA-specific validation
    logger       *pustakaLogger.Logger
}

// EA entity creation delegates to CI service with extra validation
func (s *Service) CreateObjective(ctx context.Context, req *CreateObjectiveRequest, userID uuid.UUID) (*ConfigurationItem, error) {
    // 1. EA-specific validation
    if err := s.validator.ValidateObjective(req); err != nil {
        return nil, err
    }

    // 2. Transform to CI request
    ciReq := &ci.CreateCIRequest{
        Name:              req.Name,
        CIType:            "strategy.objective",   // EA CI type
        Attributes:        req.Attributes,         // Domain attributes
        Tags:              req.Tags,
        LifecycleStatusID: req.LifecycleStatusID,
    }

    // 3. Delegate to CI service
    ci, err := s.ciService.CreateCI(ctx, ciReq, userID)
    if err != nil {
        return nil, err
    }

    // 4. EA-specific post-processing (e.g., auto-relate to parent entities)
    if req.ParentObjectiveID != nil {
        s.createParentChildRelationship(ctx, ci.ID, *req.ParentObjectiveID, userID)
    }

    return ci, nil
}
```

### Pattern 2: CI Type Taxonomy for Domain Separation

**What:** EA entities are stored as CIs with typed names following the pattern `{domain}.{entity}` (e.g., "strategy.objective", "business.capability_l1").

**When to use:** When multiple domains coexist in a unified data store with shared infrastructure.

**Trade-offs:**
- **Pros:** Single source of truth, unified relationships, reuses all CMDB features (audit, RBAC, search).
- **Cons:** Requires naming convention discipline, queries must filter by CI type prefix.

**Example:**
```go
// internal/ea/models.go
package ea

type EADomain string

const (
    StrategyDomain        EADomain = "strategy"
    BusinessDomain        EADomain = "business"
    ApplicationDomain     EADomain = "application"
    DataDomain            EADomain = "data"
    TechnologyDomain      EADomain = "technology"
    InfrastructureDomain  EADomain = "infrastructure"
    SecurityDomain        EADomain = "security"
    GovernanceDomain      EADomain = "governance"
)

type EAEntityType string

const (
    // Strategy entities
    ObjectiveEA       EAEntityType = "objective"
    InitiativeEA      EAEntityType = "initiative"
    ProgramEA         EAEntityType = "program"
    ProjectEA         EAEntityType = "project"

    // Business entities
    OrganizationEA    EAEntityType = "organization"
    BusinessDomainEA  EAEntityType = "business_domain"
    CapabilityL1EA    EAEntityType = "capability_l1"
    CapabilityL2EA    EAEntityType = "capability_l2"
    ProductEA         EAEntityType = "product"

    // Application entities
    AppGroupEA        EAEntityType = "app_group"
    BusinessAppEA     EAEntityType = "business_app"
    SupportingAppEA   EAEntityType = "supporting_app"
    SubsystemEA       EAEntityType = "subsystem"
    InterfaceEA       EAEntityType = "interface"

    // Data entities
    DataDomainEA      EAEntityType = "data_domain"
    DataObjectEA      EAEntityType = "data_object"

    // Technology entities
    ITComponentEA     EAEntityType = "it_component"
    TechCategoryEA    EAEntityType = "tech_category"
    ProviderEA        EAEntityType = "provider"

    // Infrastructure entities
    LocationEA        EAEntityType = "location"
    DataCenterEA      EAEntityType = "data_center"
    NetworkZoneEA     EAEntityType = "network_zone"
    ComputePlatformEA EAEntityType = "compute_platform"
    NetworkNodeEA     EAEntityType = "network_node"

    // Security entities
    SecurityFunctionEA   EAEntityType = "security_function"
    SecurityCategoryEA   EAEntityType = "security_category"
    SecuritySubcategoryEA EAEntityType = "security_subcategory"
    ControlEA            EAEntityType = "control"

    // Governance entities
    PolicyEA          EAEntityType = "policy"
    ProcedureEA       EAEntityType = "procedure"
    StandardEA        EAEntityType = "standard"
    StandardComponentEA EAEntityType = "standard_component"
)

// CIType returns the full CI type name for an EA entity
func (e EAEntityType) CIType(domain EADomain) string {
    return fmt.Sprintf("%s.%s", domain, e)
}

// Example: "strategy.objective", "business.capability_l1"
```

### Pattern 3: Bidirectional Relationship Sync

**What:** Relationships stored bidirectionally in Neo4j for fast traversal regardless of query direction. EA service creates both forward and reverse relationships automatically.

**When to use:** When graph queries need to traverse relationships in both directions efficiently.

**Trade-offs:**
- **Pros:** O(1) traversal in either direction, no runtime direction calculation, simpler Cypher queries.
- **Cons:** Double storage overhead, need to keep both directions in sync.

**Example:**
```go
// internal/ea/graph_service.go
func (s *Service) CreateEARelationship(ctx context.Context, sourceID, targetID uuid.UUID, relType EARelationshipType, userID uuid.UUID) error {
    // 1. Create forward relationship (Source -> Target)
    forwardReq := &ci.CreateRelationshipRequest{
        SourceID:        sourceID,
        TargetID:        targetID,
        RelationshipType: string(relType.ForwardType),
        Attributes:      map[string]interface{}{
            "direction": "forward",
        },
    }

    if _, err := s.ciService.CreateRelationship(ctx, forwardReq, userID); err != nil {
        return fmt.Errorf("failed to create forward relationship: %w", err)
    }

    // 2. Create reverse relationship (Target -> Source)
    reverseReq := &ci.CreateRelationshipRequest{
        SourceID:        targetID,
        TargetID:        sourceID,
        RelationshipType: string(relType.ReverseType),
        Attributes:      map[string]interface{}{
            "direction": "reverse",
            "forward_relationship_id": forwardRel.ID,
        },
    }

    if _, err := s.ciService.CreateRelationship(ctx, reverseReq, userID); err != nil {
        return fmt.Errorf("failed to create reverse relationship: %w", err)
    }

    return nil
}

// Relationship type definitions
type EARelationshipType struct {
    ForwardType  EARelationshipKind
    ReverseType  EARelationshipKind
    Description  string
}

type EARelationshipKind string

const (
    // Strategy relationships
    RelDrives      EARelationshipKind = "drives"        // Objective -> Initiative
    RelConsistsOf  EARelationshipKind = "consists_of"   // Program -> Project
    RelDrivenBy    EARelationshipKind = "driven_by"     // Initiative -> Objective

    // Business relationships
    RelSupports    EARelationshipKind = "supports"      // App -> Capability
    RelSupportedBy EARelationshipKind = "supported_by"  // Capability -> App

    // Application relationships
    RelDependsOn   EARelationshipKind = "depends_on"    // Subsystem -> Subsystem
    RelRequiredBy  EARelationshipKind = "required_by"   // Subsystem -> Subsystem

    // Infrastructure relationships
    RelDeployedOn  EARelationshipKind = "deployed_on"   // Subsystem -> Compute
    RelHosts       EARelationshipKind = "hosts"         // Compute -> Subsystem
)

var RelationshipTypeRegistry = map[string]EARelationshipType{
    "objective_initiative": {
        ForwardType: RelDrives,
        ReverseType: RelDrivenBy,
        Description: "Objective drives Initiative",
    },
    "app_capability": {
        ForwardType: RelSupports,
        ReverseType: RelSupportedBy,
        Description: "Application supports Business Capability",
    },
    // ... 80+ relationship types defined
}
```

## Data Flow

### Request Flow (EA Entity Creation)

```
[User submits Objective form]
    ↓
[Vue Component: ObjectiveFormView.vue]
    ↓ (calls store action)
[Pinia Store: eaEntities.createObjective()]
    ↓ (HTTP POST /api/v1/ea/strategy/objectives)
[API Handler: ea_handlers.CreateObjective()]
    ↓ (validates JWT, extracts userID)
[Middleware: RBAC check (ea:objective:create)]
    ↓
[EA Service: eaService.CreateObjective()]
    ├→ [Domain Validator: ValidateObjective attributes]
    ├→ [Transform to CI request with CIType="strategy.objective"]
    ├→ [CI Service: ciService.CreateCI()] → PostgreSQL (INSERT configuration_items)
    ├→ [Neo4j Service: neo4j.SyncCI()] → Neo4j (MERGE ConfigurationItem node)
    ├→ [EA Service: createParentChildRelationship()]
    │   └→ [CI Service: CreateRelationship()] → Neo4j (CREATE bidirectional edges)
    ├→ [Redis: Invalidate cache for objective list]
    └→ [Audit Service: Log "create objective" event]
    ↓
[Response: Objective entity with ID]
    ↓
[Pinia Store: Update local state, invalidate cache]
    ↓
[Vue Component: Display success, redirect to details view]
```

### Graph Traversal Flow (Impact Analysis)

```
[User clicks "Show Impact" on Application]
    ↓
[Vue Component: ApplicationDetailsView.vue]
    ↓ (calls store action)
[Pinia Store: eaRelationships.getImpactAnalysis(appID)]
    ↓ (HTTP GET /api/v1/ea/graph/impact/{id})
[API Handler: ea_graph_handlers.GetImpactAnalysis()]
    ↓
[EA Graph Service: eaGraphService.GetImpactAnalysis()]
    ├→ [Neo4j Query: 1-hop upstream and downstream]
    │   │
    │   ▼
    │   MATCH path = (ci:ConfigurationItem {id: $id})<-[:RELATES_TO*1..1]-(related)
    │   RETURN related, relationships(path)
    │   │
    │   └→ Neo4j Graph DB
    │
    ├→ [Filter results by EA domain: "business.capability_l1", "business.product"]
    ├→ [Enrich with CI attributes from PostgreSQL cache]
    └→ [Format as ImpactAnalysis response]
    ↓
[Response: { upstream: [...], downstream: [...] }]
    ↓
[Vue Component: vis-network renders impact graph]
```

### Key Data Flows

1. **EA Entity CRUD Flow:** EA service → CI service → PostgreSQL (entity) + Neo4j (node) → Redis (cache invalidation) → Audit log
2. **EA Relationship Creation Flow:** EA service → CI service → PostgreSQL (relationship record) + Neo4j (bidirectional edges) → Audit log
3. **Graph Visualization Flow:** Frontend request → EA graph service → Neo4j Cypher query → Enrich with PostgreSQL → vis-network rendering
4. **Bulk Import Flow:** CSV upload → EA import service → Parse & validate → Batch CI creation → Batch relationship creation → Progress tracking → Result summary

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-1k EA entities | Single PostgreSQL instance, single Neo4j instance, Redis caching sufficient |
| 1k-10k EA entities | Add database indexes on `ci_type` prefix, Neo4j relationship indexes, Redis caching for EA type definitions, pagination on all list views |
| 10k-100k EA entities | Consider read replicas for PostgreSQL (analytics queries), Neo4j cluster for high-availability, background job processing for large graph traversals, query result caching |
| 100k+ EA entities | Database sharding by domain (strategy/business/app/data separate tables), Neo4j federation by domain, materialized views for common queries, Elasticsearch for full-text search across EA entities |

### Scaling Priorities

1. **First bottleneck: Graph traversal performance** (Neo4j)
   - **Fix:** Add indexes on frequently traversed relationship types (supports, depends_on, deployed_on)
   - **Implementation:** Cypher queries use `USING INDEX` hints, create composite indexes on `(ci_type, relationship_type)`

2. **Second bottleneck: Large relationship queries** (impact analysis)
   - **Fix:** Pagination, depth limiting (max 3 hops), background job processing for deep traversals
   - **Implementation:** Async job queue (e.g., PostgreSQL LISTEN/NOTIFY or dedicated job runner), results cached in Redis

3. **Third bottleneck: Import performance** (bulk data loading)
   - **Fix:** Batch inserts (100 CIs per transaction), parallel relationship creation, progress streaming
   - **Implementation:** `COPY` command for PostgreSQL data loading, Neo4j batch transaction API, WebSocket for progress updates

## Anti-Patterns

### Anti-Pattern 1: Separate EA Data Model

**What people do:** Create new database tables for EA entities (e.g., `ea_objectives`, `ea_capabilities`) separate from CMDB CIs.

**Why it's wrong:** Duplicates infrastructure (CRUD handlers, audit logging, RBAC), breaks unified relationship graph, prevents cross-domain queries between EA and infrastructure CIs, violates DRY principle.

**Do this instead:** Model EA entities as CI types in the existing `configuration_items` table with `ci_type` following the `{domain}.{entity}` pattern. Reuse all CMDB infrastructure.

### Anti-Pattern 2: Circular Domain Dependencies

**What people do:** Create tight coupling between EA domains (e.g., Application service directly calling Business service for validation).

**Why it's wrong:** Creates circular dependencies, makes testing difficult, violates single responsibility principle, prevents independent domain evolution.

**Do this instead:** Use relationship-based validation. EA service validates that required relationships exist (e.g., "Application must support at least one Business Capability") but doesn't call other domain services. Validation happens at relationship creation time, not entity creation time.

### Anti-Pattern 3: Synchronous Deep Graph Traversals

**What people do:** Perform multi-hop graph traversals (5+ hops) synchronously in HTTP request handlers.

**Why it's wrong:** Causes request timeouts, blocks HTTP workers, creates poor user experience, doesn't scale.

**Do this instead:** For deep traversals (3+ hops), use background jobs with:
- Immediate response: "Analysis started, job ID: xyz"
- Background processing: Job worker performs traversal, caches results
- Polling/WebSocket: Client checks job status, retrieves results when ready
- Result TTL: Cached results expire after 1 hour

### Anti-Pattern 4: Monolithic EA Service

**What people do:** Create single 5000-line EA service with all domain logic mixed together.

**Why it's wrong:** Impossible to test, difficult to maintain, violates single responsibility principle, changes to one domain risk breaking others.

**Do this instead:** Subdomain services within `internal/ea/`:
```
internal/ea/
├── service.go              # Generic EA entity operations
├── strategy_service.go     # Strategy domain logic
├── business_service.go     # Business domain logic
├── application_service.go  # Application domain logic
├── data_service.go         # Data domain logic
├── technology_service.go   # Technology domain logic
├── infrastructure_service.go # Infrastructure domain logic
├── security_service.go     # Security domain logic
└── governance_service.go   # Governance domain logic
```

Each subdomain service handles domain-specific validation and business rules while sharing common operations through the base EA service.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| NIST CSF Controls API | HTTP client (async) | Security controls imported during initialization, cached in Redis |
| Technology Provider APIs | HTTP client (scheduled) | IT component provider data synced nightly via background job |
| LDAP/Active Directory | existing auth integration | Organizational hierarchy synced nightly to `ea_organizations` table |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| EA Service ↔ CI Service | Direct function calls | EA service composes CI service (no HTTP overhead) |
| EA Service ↔ Neo4j Service | Direct function calls | Leverages existing `internal/ci/neo4j_repository.go` |
| EA Service ↔ Audit Service | Direct function calls | Reuses existing `internal/ci/audit_service.go` |
| Frontend Stores ↔ Backend API | HTTP (axios) | Follows existing `/api/v1/` pattern, adds `/ea/` prefix |
| EA Views ↔ EA Stores | Pinia actions/getters | Reactive state management, same pattern as existing CI stores |

### EA Module Integration with Existing CMDB

```
┌──────────────────────────────────────────────────────────────────┐
│                     Existing CMDB Infrastructure                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  PostgreSQL (configuration_items, relationships, audit)    │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │ Infrastructure CIs (servers, databases, networks)    │ │  │
│  │  │ + EA CIs (objectives, capabilities, applications)     │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Neo4j (relationships)                                     │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │ Infrastructure ↔ Infrastructure relationships        │ │  │
│  │  │ + Infrastructure ↔ EA relationships (app → server)   │ │  │
│  │  │ + EA ↔ EA relationships (app → capability)           │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Services (CI, Auth, Audit) - Reused by EA module         │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                           ↑
                           │ Integrated via
                           │ composition
                           │
┌──────────────────────────────────────────────────────────────────┐
│                     EA Module (New Code)                         │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  EA Service Layer (wraps CI service)                      │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │  │
│  │  │ Strategy │  │ Business │  │   App    │  │  Data    │  │  │
│  │  │ Service  │  │ Service  │  │ Service  │  │ Service  │  │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

**Key Integration Principle:** EA module extends CMDB by **adding data and rules, not infrastructure**. All EA entities are CIs, all EA relationships are relationships, all EA operations are CI operations with domain-specific behavior.

## Build Order & Dependencies

### Phase 1: Foundation (Required for all subsequent work)

1. **EA CI Type Definitions** (internal/ea/models.go)
   - Define 8 domains, 60+ entity types
   - Define 80+ relationship types with bidirectional mappings
   - Create CI type naming convention registry

2. **Database Migration** (cmd/migrations/)
   - No schema changes required (reuse existing tables)
   - Create EA CI type seed data (60+ INSERT into `ci_type_definitions`)
   - Create relationship type seed data (80+ INSERT into `relationship_types`)

3. **EA Service Skeleton** (internal/ea/service.go)
   - Service struct with CI service composition
   - Basic CRUD methods that delegate to CI service
   - Domain-specific validation hooks (empty implementations)

### Phase 2: Core Entity Management (Depends on: Phase 1)

4. **Validation Logic** (internal/ea/validation.go)
   - Per-entity attribute validation rules
   - Cross-domain relationship validation
   - Business rule enforcement (e.g., "Project must change at least one Application")

5. **HTTP Handlers** (internal/api/handlers/ea_handlers.go)
   - Generic EA entity CRUD endpoints (`/api/v1/ea/{domain}/{entity}`)
   - Request transformation (EA request → CI request)
   - Response transformation (CI response → EA response)

6. **Frontend Stores & API Client** (web/src/stores/ea*.ts, services/eaAPI.ts)
   - Generic EA entity store (works for all entity types)
   - API client with type-safe methods
   - Cache management

### Phase 3: Relationship Management (Depends on: Phase 2)

7. **EA Relationship Service** (internal/ea/relationship_service.go)
   - Bidirectional relationship creation
   - Relationship type validation
   - Bulk relationship operations

8. **Relationship Frontend** (web/src/views/ea/relationships/)
   - Relationship list/filter/search
   - Relationship creation forms (with autocomplete)
   - Relationship details view

### Phase 4: Graph Visualization (Depends on: Phase 3)

9. **EA Graph Service** (internal/ea/graph_service.go)
   - EA-specific graph queries (filter by domain, relationship type)
   - 1-hop impact analysis
   - Path finding between entities

10. **Graph Visualization** (web/src/views/ea/graph/EAGraphView.vue)
    - vis-network integration
    - Domain-based node coloring
    - Interactive exploration

### Phase 5: Advanced Features (Depends on: Phase 2, 3, 4)

11. **Bulk Import** (internal/ea/import.go, web/src/views/ea/import/)
    - CSV parsing and validation
    - Batch entity creation
    - Progress tracking (WebSocket)
    - Error reporting

12. **Extended RBAC** (internal/auth/rbac.go modifications)
    - Add EA permissions (`ea:strategy:create`, `ea:business:update`, etc.)
    - Domain-level permissions matrix
    - Role-permissions mapping updates

### Dependency Graph

```
Phase 1 (Foundation)
    ↓
Phase 2 (Entity Management) ←───┐
    ↓                          │
Phase 3 (Relationships) ────────┤ (Import needs relationships)
    ↓                          │
Phase 4 (Graph) ←──────────────┘
    ↓
Phase 5 (Advanced Features)
```

## Sources

- **Existing Pustaka Architecture:** Analysis of `/home/tahopetis/dev/pustaka/internal/ci/`, `/home/tahopetis/dev/pustaka/internal/api/`, `/home/tahopetis/dev/pustaka/web/src/`
- **EA Metamodel Reference:** `/home/tahopetis/dev/archer/docs/01-metamodel-structure.md` and `/home/tahopetis/dev/archer/docs/02-metamodel-relationships.md`
- **Project Requirements:** `/home/tahopetis/dev/pustaka/.planning/PROJECT.md`

---
*Architecture research for: Enterprise Architecture Module extending Pustaka CMDB*
*Researched: 2026-02-20*

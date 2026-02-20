# Phase 1: Foundation - Context

**Gathered:** 2026-02-20
**Status:** Ready for planning

## Phase Boundary

Define EA metamodel with 60+ CI types across 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance), establish 20-25 core relationship types, create database migration seeding EA types into existing CMDB, and build EA service layer skeleton. EA entities are modeled as CI Types within the unified CMDB taxonomy, leveraging existing PostgreSQL, Neo4j, and Redis infrastructure.

## Implementation Decisions

### Metamodel Organization

**CI Type Naming Pattern:**
- Format: `EA.Strategy-Objective`, `EA.Business-CapabilityL1`, `EA.Application-BusinessApp`
- Structure: `EA.` prefix + domain-entity with dash separator
- Example: `EA.Strategy-Objective`, `EA.Business-CapabilityL1`, `EA.Application-BusinessApp`, `EA.Data-DataObject`, `EA.Technology-ITComponent`

**CI Type Attributes:**
- Shared base attributes: name, description, owner
- Domain-specific extensions per CI type
- Each EA CI type has base attributes plus its own custom fields
- Example: Strategy Objective has (name, description, owner) + (strategic_alignment, target_date)

**Relationship Type Organization:**
- Unified namespace for all 20-25 relationship types
- All relationships defined together, domain is an attribute
- Examples: supports, depends_on, deployed_on, flows_to, assigned_to, realizes

**Domain Boundaries:**
- Validation with overrides, but only for EA metamodel relationships
- Enforce cross-domain relationship rules
- Admin can override with justification logged to audit trail

**CI Type Identifiers:**
- Numeric autonumber IDs (consistent with existing CIs: 1, 2, 3...)
- No semantic codes or UUIDs

**Hierarchical Relationships:**
- Yes, strict hierarchy in CI taxonomy
- Example: Capability L2 modeled as child of Capability L1 in taxonomy structure
- Explicit parent-child relationships where applicable

**Metadata Attributes:**
- Yes, track metadata for all EA entities
- Fields: source, last_updated_by, data_quality_score
- Helps with data quality governance

**CI Type Definition Versioning:**
- Yes, versioned CI type definitions
- Allow EA CI type definitions to evolve with versioning
- Support schema evolution over time

**Lifecycle Statuses:**
- Reuse existing CMDB lifecycle_statuses
- Values: proposed, active, deprecated, retired
- No EA-specific lifecycle states in v1

**Ownership Model:**
- Create new `ea_teams` table
- Admin can add teams via UI
- EA entities reference team via dropdown (team ownership)
- Separate from individual user ownership

**Attribute Extensibility:**
- Controlled extension for custom attributes
- Custom key-value attributes allowed with validation rules on attribute names
- Balance flexibility with data governance

**Documentation References:**
- Optional `documentation_url` field on EA entities
- Link to external docs (Confluence, Google Drive)
- Not required but available when needed

**CMDB Taxonomy Integration:**
- Separate `ea_domain` field for EA entities
- Distinct from existing Domain/Category/Subcategory taxonomy
- EA domain = one of 8 domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance)

**Criticality Scoring:**
- No criticality/priority field in v1
- Defer to v2 analytics
- Focus on data modeling first

**Relationship Cardinality:**
- Yes, define cardinality rules for relationship types
- Examples: many-to-many (App supports many Capabilities), one-to-many (Capability has many L2 children)
- Enforced in validation layer

**Status Change Tracking:**
- Audit log is sufficient for status changes
- No separate status history table
- Leverage existing audit_logs infrastructure

### Migration Strategy

**Seed File Organization:**
- Single monolithic migration file
- Contains all 60+ CI types and 20-25 relationship types
- One file to execute: `YYYYMMDD_add_ea_metamodel.up.sql`

**Insertion Order:**
- Dependencies first approach
- Order: EA teams → CI types → Relationship types → Validation rules
- Ensures referential integrity

**Rollback Support:**
- Full rollback support with DOWN migrations
- Cleanly remove EA types if needed
- `YYYYMMDD_add_ea_metamodel.down.sql` included

**Data Validation:**
- Include validation queries in migrations
- Verify EA types loaded correctly
- Check counts (60 CI types, 20-25 relationships, teams created)
- Validate referential integrity after load

### Service Layer Design

**Integration with CI Service:**
- Inheritance-based approach
- EA service extends CI service
- Reuse CI service functionality while adding EA-specific behavior

**Domain Service Organization:**
- 8 separate services, one per domain
- Files: strategy_service.go, business_service.go, application_service.go, data_service.go, technology_service.go, infrastructure_service.go, security_service.go, governance_service.go
- Each handles domain-specific logic

**Validation Logic Location:**
- Service layer validation
- EA business rules live in EA services
- Handlers do HTTP validation + auth, services do EA domain validation
- Clean separation, self-contained services, consistent enforcement across all callers

**Error Handling:**
- Hybrid approach
- EA-specific errors for unique EA cases (validation errors, governance issues)
- Reuse CI errors for common cases (not found, duplicate, invalid input)
- Balance clarity with consistency

### Validation Framework

**Validation Strictness:**
- Warn but allow approach
- Validation errors logged, entity saved with `validation_errors` flag
- Data quality dashboard shows validation status
- Balance data quality with usability during initial data population

**Cross-Domain Relationship Validation:**
- Yes, enforce cross-domain relationship rules
- Validate that relationships only connect allowed domain pairs
- Prevent invalid cross-domain connections
- Examples: Application supports Business Capability (valid), Application deployed-on Data Object (invalid)

**Attribute Validation Implementation:**
- Hybrid approach
- Standard validation (required fields, string lengths, email format, numeric ranges) → use `validator/v10` struct tags
- EA-specific validation (cross-domain rules, parent-child relationships, cardinality, business logic) → custom validation functions
- Extensible: easy to add new CI types and validation rules

**Override Mechanism:**
- Yes, override with reason
- Admin users can bypass validation with explicit flag
- Requires justification/comment in audit log
- Enables flexibility while maintaining governance

## Specific Ideas

- Teams table ownership model enables team-based governance rather than individual ownership
- Hybrid validation (struct tags + custom functions) balances declarative simplicity with EA business complexity
- Strict hierarchy for Capability L1/L2 reflects standard EA practice
- Metadata tracking (source, quality score) supports data quality initiatives from day one
- Versioned CI types support EA metamodel evolution without breaking existing entities

## Deferred Ideas

None — discussion stayed within Phase 1 scope (metamodel, migrations, service layer, validation).

---

*Phase: 01-foundation*
*Context gathered: 2026-02-20*

# EA Metamodel Specifications

**Version:** 1.0
**Created:** 2026-02-20
**Purpose:** Single source of truth for EA entity modeling in Pustaka CMDB

## Section 1: EA Domain Overview

The Enterprise Architecture (EA) metamodel is organized into 8 domains following ArchiMate 3.x patterns with extensions for Security, Governance, and Infrastructure. Each EA entity is modeled as a Configuration Item (CI) type within the existing CMDB taxonomy, enabling unified relationship mapping, impact analysis, and visualization across all architectural layers.

### EA Domains

1. **Strategy** - Strategic goals, objectives, and outcomes guiding organizational direction
2. **Business** - Business capabilities, processes, services, and actors
3. **Application** - Software applications, components, services, and interfaces
4. **Data** - Data entities, datasets, stores, and data services
5. **Technology** - Technology components, platforms, artifacts, and systems
6. **Infrastructure** - Physical and virtual infrastructure (nodes, networks, storage, facilities)
7. **Security** - Security controls, policies, risks, and assessments (NIST 2.0 aligned)
8. **Governance** - Governance policies, standards, compliance, and procedures

### CI Type Naming Convention

All EA CI types follow the pattern: `EA.Domain-EntityType`

**Examples:**
- `EA.Strategy-Objective` - Strategic objective
- `EA.Business-CapabilityL1` - Level 1 Business Capability
- `EA.Application-BusinessApp` - Business Application
- `EA.Data-DataObject` - Data Entity/Object
- `EA.Technology-ITComponent` - IT Software Component

**Structure:**
- `EA.` prefix identifies all Enterprise Architecture entities
- Domain name (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance)
- Hyphen separator (`-`)
- Entity type name (camelCase)

### CI Type Inheritance from Base CMDB

All EA CI types inherit from the base `configuration_items` table structure:

**Base Attributes (all EA types):**
- `name` (string) - Entity name
- `description` (text) - Detailed description
- `owner` (string) - EA team responsible (references `ea_teams.name`)
- `attributes` (JSONB) - Domain-specific extensions
- `tags` (string array) - Tags for filtering and search
- `lifecycle_status_id` (UUID) - Lifecycle status (proposed, active, deprecated, retired)
- `created_at`, `updated_at` - Timestamps
- `created_by`, `updated_by` - User references

**Domain-Specific Extensions:**
Each EA CI type has optional attributes stored in the `attributes` JSONB field, defined per type in this specification.

## Section 2: CI Type Definition Schema

Each EA CI type is defined with the following structure:

### Schema Fields

**Core Identity:**
- `name` - EA.Domain-EntityType format (e.g., EA.Strategy-Objective)
- `description` - Human-readable description (1-2 sentences)
- `domain` - One of 8 EA domains

**Required Attributes** (shared by all EA types, stored in `ci_type_definitions.required_attributes` JSONB):
```json
[
  {
    "name": "name",
    "type": "string",
    "description": "Entity name",
    "validation": {"min_length": 3, "max_length": 255}
  },
  {
    "name": "description",
    "type": "string",
    "description": "Detailed description"
  },
  {
    "name": "owner",
    "type": "string",
    "description": "EA team responsible (references ea_teams.name)"
  }
]
```

**Optional Attributes** (domain-specific, stored in `ci_type_definitions.optional_attributes` JSONB):
Each field includes:
- `name` - Attribute name
- `type` - Data type: string, integer, date, enum, uuid, array, boolean
- `validation` - Validation rules (min/max length, enum values, regex patterns)
- `description` - Purpose of the attribute

## Section 3: EA CI Type Catalog (60+ types)

### Strategy Domain (6 types)

#### EA.Strategy-Objective
- **Domain**: Strategy
- **Description**: High-level strategic objective representing organizational goals and desired outcomes with measurable targets and timeframes
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Objective name
  - description: text - Detailed description of the objective
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - strategic_alignment: enum (high, medium, low) - Strategic importance level
  - target_date: date - Target achievement date
  - metrics: array of strings - Key performance indicators
  - parent_objective_id: uuid (self-reference) - Parent objective if this is a sub-objective
  - outcome: text - Expected outcome or result

#### EA.Strategy-Goal
- **Domain**: Strategy
- **Description**: Intermediate goal that supports strategic objectives with specific measurable criteria
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Goal name
  - description: text - Detailed description
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - objective_id: uuid (references EA.Strategy-Objective) - Associated objective
  - target_value: string - Target measurable value
  - current_value: string - Current progress value
  - measurement_unit: string - Unit of measurement
  - due_date: date - Target completion date

#### EA.Strategy-Requirement
- **Domain**: Strategy
- **Description**: Strategic requirement derived from objectives and goals that constrains or guides solution design
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Requirement name
  - description: text - Detailed requirement description
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - priority: enum (critical, high, medium, low) - Requirement priority
  - category: enum (functional, non-functional, regulatory, technical) - Requirement category
  - source: string - Requirement source (stakeholder, regulation, business need)
  - acceptance_criteria: text - Criteria for requirement fulfillment

#### EA.Strategy-Constraint
- **Domain**: Strategy
- **Description**: Strategic limitation or restriction that affects solution options and architectural decisions
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Constraint name
  - description: text - Detailed constraint description
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - constraint_type: enum (technical, financial, regulatory, organizational, temporal) - Type of constraint
  - impact: text - Description of constraint impact on architecture
  - expiry_date: date - Date when constraint expires (if applicable)

#### EA.Strategy-Principle
- **Domain**: Strategy
- **Description**: Guiding principle that informs architectural decision-making and design choices
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Principle name
  - description: text - Principle statement
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - rationale: text - Rationale for the principle
  - implications: text - Implications for architecture and design
  - priority: enum (must, should, could) - Principle priority level

#### EA.Strategy-Outcome
- **Domain**: Strategy
- **Description**: Expected outcome or result from achieving strategic objectives and goals
- **Required Attributes**:
  - name: string (min: 5, max: 100) - Outcome name
  - description: text - Detailed outcome description
  - owner: string (ea_teams reference) - EA team responsible
- **Optional Attributes**:
  - objective_ids: array of uuid - Associated objectives
  - measurement_criteria: text - How outcome achievement is measured
  - target_date: date - Expected realization date
  - value_type: enum (qualitative, quantitative, financial, operational) - Type of value

### Business Domain (10 types)

#### EA.Business-CapabilityL1
- **Domain**: Business
- **Description**: Level 1 Business Capability representing high-level business functions that define what the business does
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Capability name
  - description: text - Capability description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - strategic_importance: enum (critical, high, medium, low) - Strategic importance
  - business_value: text - Business value proposition
  - parent_capability_id: uuid (self-reference) - Parent capability for hierarchical capabilities
  - maturity_level: enum (initial, developing, defined, managed, optimized) - Capability maturity
  - target_maturity: enum (initial, developing, defined, managed, optimized) - Target maturity state

#### EA.Business-CapabilityL2
- **Domain**: Business
- **Description**: Level 2 Business Capability providing detailed breakdown of L1 capabilities into specific sub-capabilities
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Capability name
  - description: text - Capability description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - parent_capability_id: uuid (references EA.Business-CapabilityL1) - Parent L1 capability (required)
  - strategic_importance: enum (critical, high, medium, low) - Strategic importance
  - business_value: text - Business value proposition
  - maturity_level: enum (initial, developing, defined, managed, optimized) - Capability maturity

#### EA.Business-Process
- **Domain**: Business
- **Description**: Business process defining a set of coordinated activities performed to achieve specific business outcomes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Process name
  - description: text - Process description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - capability_id: uuid (references EA.Business-CapabilityL2) - Supported capability
  - inputs: array of strings - Process inputs
  - outputs: array of strings - Process outputs
  - process_type: enum (primary, support, management) - Process classification
  - frequency: enum (real-time, hourly, daily, weekly, monthly, quarterly, annually, ad-hoc) - Execution frequency
  - automation_level: enum (manual, semi-automated, fully-automated) - Automation level

#### EA.Business-Function
- **Domain**: Business
- **Description**: Business function representing a cohesive collection of business processes and activities
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Function name
  - description: text - Function description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - capability_id: uuid (references EA.Business-CapabilityL2) - Associated capability
  - parent_function_id: uuid (self-reference) - Parent function if nested
  - responsibility: string - Primary responsibility or purpose

#### EA.Business-Service
- **Domain**: Business
- **Description**: Business service delivered to customers or internal stakeholders, representing value provided by the organization
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Service name
  - description: text - Service description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - service_type: enum (external, internal, supporting) - Service type
  - capability_id: uuid (references EA.Business-CapabilityL2) - Supported capability
  - sla: text - Service level agreement description
  - criticality: enum (mission-critical, critical, important, standard) - Service criticality
  - customer_segments: array of strings - Target customer segments

#### EA.Business-Event
- **Domain**: Business
- **Description**: Business event that triggers or initiates business processes or functions
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Event name
  - description: text - Event description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - event_type: enum (external, internal, scheduled, ad-hoc) - Event classification
  - trigger_condition: text - Condition that triggers the event
  - triggered_process_ids: array of uuid - Processes triggered by this event

#### EA.Business-Role
- **Domain**: Business
- **Description**: Business role defining a set of responsibilities and skills within the organization
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Role name
  - description: text - Role description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - responsibilities: array of strings - Role responsibilities
  - skills_required: array of strings - Required skills and competencies
  - parent_role_id: uuid (self-reference) - Parent role in role hierarchy

#### EA.Business-Actor
- **Domain**: Business
- **Description**: Business actor representing an individual, team, or organization that performs business functions and processes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Actor name
  - description: text - Actor description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - actor_type: enum (person, team, department, organization, external) - Actor classification
  - role_ids: array of uuid - Roles assigned to this actor
  - location: string - Physical or organizational location
  - contact_info: text - Contact information

#### EA.Business-Collaboration
- **Domain**: Business
- **Description**: Business collaboration representing interaction and cooperation between actors to achieve common goals
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Collaboration name
  - description: text - Collaboration description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - collaboration_type: enum (long-term, temporary, project-based, partnership) - Type of collaboration
  - actor_ids: array of uuid - Actors involved in collaboration
  - purpose: text - Purpose and objectives of collaboration
  - start_date: date - Collaboration start date
  - end_date: date - Collaboration end date (if applicable)

#### EA.Business-Interaction
- **Domain**: Business
- **Description**: Business interaction representing the exchange of information or services between business actors, roles, or services
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Interaction name
  - description: text - Interaction description
  - owner: string (ea_teams reference) - Business team responsible
- **Optional Attributes**:
  - interaction_type: enum (synchronous, asynchronous) - Interaction mode
  - source_actor_id: uuid - Initiating actor
  - target_actor_id: uuid - Receiving actor
  - service_id: uuid (references EA.Business-Service) - Associated service
  - frequency: enum (real-time, hourly, daily, weekly, monthly, ad-hoc) - Interaction frequency

### Application Domain (9 types)

#### EA.Application-BusinessApp
- **Domain**: Application
- **Description**: Business application representing a complete software application that supports business capabilities and processes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Application name
  - description: text - Application description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - lifecycle_status: enum (proposed, development, testing, production, deprecated, retired) - Application lifecycle
  - criticality: enum (mission-critical, critical, important, standard, low) - Business criticality
  - technology_stack: array of strings - Technology stack components
  - business_owner: string - Business owner (person or team)
  - it_owner: string - IT owner (person or team)
  - vendor: string - Software vendor (if commercial)
  - version: string - Current version
  - first_deployed_date: date - Initial production deployment date
  - end_of_life_date: date - Expected end of life date
  - license_type: enum (commercial, open-source, custom, saas) - License type

#### EA.Application-Component
- **Domain**: Application
- **Description**: Application component representing a modular part of an application with defined functionality and interfaces
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Component name
  - description: text - Component description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - parent_app_id: uuid (references EA.Application-BusinessApp) - Parent application
  - component_type: enum (module, library, microservice, plugin, extension) - Component type
  - technology: string - Technology or framework used
  - interface_type: enum (api, ui, cli, library, message-based) - Interface type
  - is_public_facing: boolean - Whether component is externally accessible

#### EA.Application-Service
- **Domain**: Application
- **Description**: Application service exposing functionality through well-defined interfaces (APIs, web services, message endpoints)
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Service name
  - description: text - Service description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - service_type: enum (rest, graphql, soap, grpc, message-queue, event-stream) - Service type
  - component_id: uuid (references EA.Application-Component) - Parent component
  - protocol: string - Protocol (HTTP/HTTPS, AMQP, Kafka, etc.)
  - endpoint: string - Service endpoint URL or address
  - authentication: enum (none, api-key, oauth2, jwt, mutual-tls, basic-auth) - Authentication method
  - version: string - Service version
  - sla: text - Service level agreement description
  - availability: string (percentage) - Service availability target (e.g., "99.9%")

#### EA.Application-Interface
- **Domain**: Application
- **Description**: Application interface defining a contract for interaction between applications or components
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Interface name
  - description: text - Interface description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - interface_type: enum (api, sdk, cli, file, database) - Interface type
  - protocol: string - Communication protocol
  - data_format: enum (json, xml, csv, binary, custom) - Data format
  - version: string - Interface version
  - service_id: uuid (references EA.Application-Service) - Associated service
  - is_internal: boolean - Whether interface is internal or external

#### EA.Application-Function
- **Domain**: Application
- **Description**: Application function representing a unit of application behavior or business logic
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Function name
  - description: text - Function description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - component_id: uuid (references EA.Application-Component) - Parent component
  - business_process_id: uuid (references EA.Business-Process) - Supported business process
  - function_type: enum (business-logic, data-processing, integration, utility) - Function type

#### EA.Application-Event
- **Domain**: Application
- **Description**: Application event representing a significant state change or occurrence within an application
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Event name
  - description: text - Event description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - event_type: enum (state-change, error, warning, info, business-event) - Event classification
  - component_id: uuid (references EA.Application-Component) - Source component
  - severity: enum (critical, high, medium, low, info) - Event severity
  - is_audit_event: boolean - Whether event is logged for audit purposes

#### EA.Application-DataObject
- **Domain**: Application (shared with Data domain)
- **Description**: Application data object representing data structure or entity used or produced by applications
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Data object name
  - description: text - Data object description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - data_type: enum (structured, semi-structured, unstructured) - Data structure type
  - format: string - Data format (JSON, XML, CSV, binary, etc.)
  - sensitivity: enum (public, internal, confidential, restricted) - Data sensitivity level
  - is_pii: boolean - Whether data contains personally identifiable information
  - retention_period: string - Data retention period
  - application_id: uuid (references EA.Application-BusinessApp) - Primary application

#### EA.Application-Contract
- **Domain**: Application
- **Description**: Application contract defining the specification or agreement between application components or services
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Contract name
  - description: text - Contract description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - contract_type: enum (api-contract, sla, data-contract, integration-agreement) - Contract type
  - party_ids: array of uuid - Parties involved (applications or services)
  - terms: text - Contract terms and conditions
  - effective_date: date - Contract effective date
  - review_date: date - Contract review date

#### EA.Application-Representation
- **Domain**: Application
- **Description**: Application representation defining how application functionality or data is presented to users
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Representation name
  - description: text - Representation description
  - owner: string (ea_teams reference) - Application team responsible
- **Optional Attributes**:
  - representation_type: enum (ui, api-documentation, report, dashboard, mobile) - Type of representation
  - channel: enum (web, mobile, desktop, voice, chatbot) - Delivery channel
  - service_id: uuid (references EA.Application-Service) - Underlying service
  - audience: enum (internal, external, partner, public) - Target audience

### Data Domain (7 types)

#### EA.Data-DataObject
- **Domain**: Data (shared with Application domain)
- **Description**: Data object representing a structured data entity or concept used within applications and business processes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Data object name
  - description: text - Data object description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - data_type: enum (structured, semi-structured, unstructured) - Data structure type
  - format: string - Data format (JSON, XML, CSV, etc.)
  - sensitivity: enum (public, internal, confidential, restricted) - Data sensitivity level
  - is_pii: boolean - Whether data contains personally identifiable information
  - retention_period: string - Data retention period
  - classification: enum (public, internal, confidential, restricted, secret) - Data classification
  - compliance_requirements: array of strings - Applicable compliance requirements (GDPR, HIPAA, etc.)

#### EA.Data-DataSet
- **Domain**: Data
- **Description**: Data set representing a collection of related data objects grouped for a specific purpose or domain
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Data set name
  - description: text - Data set description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - data_object_ids: array of uuid - Included data objects
  - domain: string - Business or technical domain
  - data_volume: string - Estimated data volume (e.g., "100GB", "1M records")
  - update_frequency: enum (real-time, hourly, daily, weekly, monthly) - Update frequency
  - retention_period: string - Data retention period

#### EA.Data-DataStore
- **Domain**: Data
- **Description**: Data store representing a database, data warehouse, or data lake where data is persisted and managed
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Data store name
  - description: text - Data store description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - store_type: enum (relational, document, key-value, column-family, graph, time-series, data-lake, data-warehouse) - Store type
  - technology: string - Technology platform (e.g., "PostgreSQL", "MongoDB", "S3")
  - version: string - Database version
  - deployment_type: enum (on-premise, cloud, hybrid) - Deployment model
  - is_primary: boolean - Whether this is the primary store for data objects
  - backup_frequency: enum (real-time, hourly, daily, weekly) - Backup frequency
  - disaster_recovery: boolean - Whether disaster recovery is configured

#### EA.Data-DataService
- **Domain**: Data
- **Description**: Data service providing access, manipulation, or processing of data through well-defined interfaces
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Service name
  - description: text - Service description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - service_type: enum (crud, query, etl, streaming, analytics, search) - Service type
  - protocol: string - Protocol (SQL, REST, GraphQL, etc.)
  - endpoint: string - Service endpoint
  - data_store_id: uuid (references EA.Data-DataStore) - Underlying data store
  - authentication: enum (none, api-key, oauth2, jwt, mutual-tls) - Authentication method
  - is_public_facing: boolean - Whether service is externally accessible

#### EA.Data-DataRule
- **Domain**: Data
- **Description**: Data rule defining business rules, validation constraints, or transformation logic applied to data
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Rule name
  - description: text - Rule description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - rule_type: enum (validation, transformation, business-rule, data-quality) - Rule type
  - scope: enum (attribute, record, dataset, cross-dataset) - Rule scope
  - severity: enum (critical, high, medium, low, warning) - Rule violation severity
  - logic: text - Rule logic or expression
  - data_object_id: uuid (references EA.Data-DataObject) - Associated data object
  - is_enforced: boolean - Whether rule is enforced or advisory

#### EA.Data-DataContract
- **Domain**: Data
- **Description**: Data contract defining the schema, structure, and semantics of data exchanged between systems or components
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Contract name
  - description: text - Contract description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - contract_version: string - Contract version
  - schema: text - Data schema definition (JSON schema, XSD, etc.)
  - data_format: enum (json, xml, csv, avro, protobuf) - Data format
  - provider_id: uuid - Data provider (application or service)
  - consumer_ids: array of uuid - Data consumers
  - governance: enum (strict, flexible, advisory) - Contract governance level
  - review_date: date - Next contract review date

#### EA.Data-DataLineage
- **Domain**: Data
- **Description**: Data lineage tracking the origin, transformation, and flow of data from source to destination
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Lineage name
  - description: text - Lineage description
  - owner: string (ea_teams reference) - Data team responsible
- **Optional Attributes**:
  - source_id: uuid - Source data object or store
  - target_id: uuid - Target data object or store
  - transformation: text - Transformation logic applied
  - frequency: enum (real-time, batch, streaming, scheduled) - Data transfer frequency
  - pipeline_id: string - Associated data pipeline or ETL job

### Technology Domain (7 types)

#### EA.Technology-ITComponent
- **Domain**: Technology
- **Description**: IT component representing a software component, library, framework, or runtime that supports applications
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Component name
  - description: text - Component description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - component_type: enum (library, framework, runtime, middleware, container) - Component type
  - technology: string - Technology name (e.g., "React", "Spring Boot", "Node.js")
  - version: string - Current version
  - license_type: enum (open-source, commercial, lgpl, gpl, apache, mit, bsd) - License type
  - end_of_support_date: date - End of support or maintenance date
  - is_deprecated: boolean - Whether component is deprecated
  - security_risk: enum (none, low, medium, high, critical) - Known security risk level

#### EA.Technology-Platform
- **Domain**: Technology
- **Description**: Technology platform providing a foundation for developing, deploying, or running applications and services
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Platform name
  - description: text - Platform description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - platform_type: enum (development, runtime, integration, data, mobile, analytics) - Platform category
  - technology: string - Underlying technology stack
  - version: string - Platform version
  - vendor: string - Platform vendor
  - capabilities: array of strings - Platform capabilities and features
  - sla: text - Service level agreement
  - scaling_type: enum (vertical, horizontal, auto) - Scaling model

#### EA.Technology-Artifact
- **Domain**: Technology
- **Description**: Technology artifact representing a deployable unit of software (container image, jar file, package, binary)
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Artifact name
  - description: text - Artifact description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - artifact_type: enum (docker-image, jar, war, npm-package, python-wheel, executable, library) - Artifact type
  - version: string - Artifact version
  - registry: string - Artifact registry or repository
  - checksum: string - Artifact checksum or hash
  - build_date: date - Artifact build date
  - size_bytes: integer - Artifact size in bytes
  - component_id: uuid (references EA.Technology-ITComponent) - Associated component

#### EA.Technology-Device
- **Domain**: Technology
- **Description**: Technology device representing hardware devices such as laptops, mobile devices, IoT devices, or endpoints
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Device name
  - description: text - Device description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - device_type: enum (laptop, desktop, mobile, tablet, iot-device, server, printer, other) - Device type
  - manufacturer: string - Device manufacturer
  - model: string - Device model
  - serial_number: string - Device serial number
  - os: string - Operating system
  - os_version: string - OS version
  - purchase_date: date - Device purchase date
  - warranty_expiry: date - Warranty expiration date
  - status: enum (active, inactive, lost, stolen, retired) - Device status

#### EA.Technology-System
- **Domain**: Technology
- **Description**: Technology system representing a software system with multiple components working together
- **Required Attributes**:
  - name: string (min: 3, max: 100) - System name
  - description: text - System description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - system_type: enum (information-system, control-system, embedded-system, operating-system) - System type
  - architecture: enum (monolithic, layered, microservices, event-driven, serverless) - System architecture
  - component_ids: array of uuid - System components
  - technology_stack: array of strings - Technology stack
  - complexity: enum (low, medium, high) - System complexity

#### EA.Technology-Network
- **Domain**: Technology
- **Description**: Technology network representing logical network protocols and communication patterns
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Network name
  - description: text - Network description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - network_type: enum (lan, wan, vpn, internet, intranet, extranet) - Network type
  - protocol: string - Network protocol (TCP/IP, HTTP, etc.)
  - bandwidth: string - Network bandwidth (e.g., "1Gbps")
  - is_encrypted: boolean - Whether network traffic is encrypted
  - encryption_type: string - Encryption type (TLS, VPN, etc.)
  - firewall_rules: text - Firewall or security rules

#### EA.Technology-Path
- **Domain**: Technology
- **Description**: Technology path representing a communication path or link between nodes or systems
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Path name
  - description: text - Path description
  - owner: string (ea_teams reference) - Technology team responsible
- **Optional Attributes**:
  - path_type: enum (wired, wireless, virtual, logical) - Path type
  - source_id: uuid - Source node or system
  - target_id: uuid - Target node or system
  - protocol: string - Communication protocol
  - bandwidth: string - Path bandwidth
  - latency: string - Typical latency (e.g., "10ms")

### Infrastructure Domain (8 types)

#### EA.Infrastructure-Node
- **Domain**: Infrastructure
- **Description**: Infrastructure node representing a physical or virtual computing resource (server, VM, container host)
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Node name
  - description: text - Node description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - node_type: enum (physical-server, virtual-machine, container-host, bare-metal, cloud-instance) - Node type
  - hostname: string - Node hostname
  - ip_address: string - Primary IP address
  - os: string - Operating system
  - os_version: string - OS version
  - cpu_cores: integer - Number of CPU cores
  - ram_gb: integer - RAM in GB
  - storage_gb: integer - Storage in GB
  - deployment_type: enum (on-premise, cloud, hybrid, edge) - Deployment model
  - environment: enum (production, staging, development, test) - Environment

#### EA.Infrastructure-Device
- **Domain**: Infrastructure
- **Description**: Infrastructure device representing physical hardware devices (servers, storage, network equipment)
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Device name
  - description: text - Device description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - device_type: enum (server, storage-array, switch, router, firewall, load-balancer, other) - Device type
  - manufacturer: string - Device manufacturer
  - model: string - Device model
  - serial_number: string - Device serial number
  - rack_location: string - Rack and unit location (e.g., "Rack A, U12")
  - purchase_date: date - Device purchase date
  - warranty_expiry: date - Warranty expiration date
  - status: enum (active, inactive, maintenance, retired) - Device status
  - ip_address: string - Management IP address

#### EA.Infrastructure-Network
- **Domain**: Infrastructure
- **Description**: Infrastructure network representing physical or logical network segments (VLANs, subnets, zones)
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Network name
  - description: text - Network description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - network_type: enum (vlan, subnet, vnet, zone, segment) - Network type
  - cidr: string - Network CIDR block (e.g., "10.0.1.0/24")
  - gateway: string - Default gateway IP
  - dns_servers: array of strings - DNS server IPs
  - vlan_id: integer - VLAN ID (if applicable)
  - is_public: boolean - Whether network is publicly accessible
  - security_level: enum (public, private, protected, isolated) - Security classification

#### EA.Infrastructure-Storage
- **Domain**: Infrastructure
- **Description**: Infrastructure storage representing storage systems, arrays, or volumes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Storage name
  - description: text - Storage description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - storage_type: enum (san, nas, object-storage, block-storage, local, cloud) - Storage type
  - technology: string - Storage technology (e.g., "EMC VNX", "AWS S3")
  - capacity_gb: integer - Total capacity in GB
  - used_gb: integer - Used storage in GB
  - is_encrypted: boolean - Whether storage is encrypted
  - redundancy: enum (raid-0, raid-1, raid-5, raid-6, raid-10, erasure-coding) - Redundancy type
  - backup_enabled: boolean - Whether backup is configured
  - snapshot_enabled: boolean - Whether snapshots are enabled

#### EA.Infrastructure-Facility
- **Domain**: Infrastructure
- **Description**: Infrastructure facility representing physical data centers, offices, or locations
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Facility name
  - description: text - Facility description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - facility_type: enum (data-center, office, colocation, cloud-region, edge-site) - Facility type
  - address: text - Physical address
  - city: string - City
  - country: string - Country
  - region: string - Region or state
  - tier: enum (tier-1, tier-2, tier-3, tier-4) - Data center tier rating
  - certification: array of strings - Certifications (ISO 27001, SOC 2, etc.)
  - disaster_recovery_site: boolean - Whether this is a DR site

#### EA.Infrastructure-Path
- **Domain**: Infrastructure
- **Description**: Infrastructure path representing physical or logical communication links between infrastructure nodes
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Path name
  - description: text - Path description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - path_type: enum (fiber, copper, wireless, vpn, internet) - Path type
  - source_id: uuid (references EA.Infrastructure-Node or Device) - Source node
  - target_id: uuid (references EA.Infrastructure-Node or Device) - Target node
  - bandwidth: string - Bandwidth capacity (e.g., "10Gbps")
  - latency_ms: integer - Typical latency in milliseconds
  - is_redundant: boolean - Whether path has redundancy
  - carrier: string - Network carrier or provider (if applicable)

#### EA.Infrastructure-DistributionNetwork
- **Domain**: Infrastructure
- **Description**: Infrastructure distribution network representing energy, cooling, or utilities infrastructure
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Distribution network name
  - description: text - Distribution network description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - utility_type: enum (power, cooling, water, fire-suppression) - Utility type
  - capacity: string - Capacity or rating
  - redundancy: enum (n, n+1, n+2, 2n, 2n+1) - Redundancy level
  - facility_id: uuid (references EA.Infrastructure-Facility) - Associated facility

#### EA.Infrastructure-MobileNetwork
- **Domain**: Infrastructure
- **Description**: Infrastructure mobile network representing cellular or mobile communication infrastructure
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Mobile network name
  - description: text - Mobile network description
  - owner: string (ea_teams reference) - Infrastructure team responsible
- **Optional Attributes**:
  - network_generation: enum (3g, 4g, 5g, lte, wi-fi) - Network generation
  - carrier: string - Mobile carrier or provider
  - coverage_area: string - Geographic coverage area
  - bandwidth: string - Available bandwidth
  - facility_id: uuid (references EA.Infrastructure-Facility) - Base station facility

### Security Domain (7 types)

#### EA.Security-Control
- **Domain**: Security
- **Description**: Security control representing a safeguard, countermeasure, or measure implemented to manage security risk
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Control name
  - description: text - Control description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - control_type: enum (preventive, detective, corrective, compensating, directive) - Control type (NIST 2.0)
  - control_category: enum (technical, management, operational) - Control category
  - nist_function: enum (identify, protect, detect, respond, recover) - NIST CSF function
  - implementation_status: enum (implemented, partial, planned, not-implemented) - Implementation status
  - effectiveness: enum (high, medium, low, unknown) - Control effectiveness
  - testing_frequency: enum (monthly, quarterly, semi-annually, annually) - Testing frequency
  - last_tested_date: date - Last control testing date

#### EA.Security-Policy
- **Domain**: Security
- **Description**: Security policy defining rules, requirements, and guidelines for security practices
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Policy name
  - description: text - Policy description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - policy_type: enum (technical, administrative, physical) - Policy type
  - scope: text - Policy scope and applicability
  - compliance_framework: array of strings - Applicable frameworks (NIST, ISO 27001, HIPAA, PCI DSS, SOC 2)
  - version: string - Policy version
  - approval_date: date - Policy approval date
  - review_frequency: enum (annually, bi-annually, tri-annually) - Review frequency
  - next_review_date: date - Next scheduled review date
  - violation_consequence: text - Consequences for policy violation

#### EA.Security-Risk
- **Domain**: Security
- **Description**: Security risk representing a potential threat, vulnerability, or exposure that could impact the organization
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Risk name
  - description: text - Risk description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - risk_type: enum (threat, vulnerability, exposure) - Risk type
  - likelihood: enum (rare, unlikely, possible, likely, almost-certain) - Risk likelihood
  - impact: enum (negligible, minor, moderate, major, catastrophic) - Risk impact
  - risk_level: enum (critical, high, medium, low) - Calculated risk level
  - threat_actor: enum (nation-state, cybercriminal, insider, hacktivist, script-kiddie) - Threat actor type
  - vulnerability_id: string - CVE or vulnerability ID (if applicable)
  - cvss_score: number - CVSS score (0-10)
  - mitigation_ids: array of uuid - Mitigating controls (references EA.Security-Control)
  - acceptance_justification: text - Justification if risk is accepted
  - review_date: date - Next risk review date

#### EA.Security-Vulnerability
- **Domain**: Security
- **Description**: Security vulnerability representing a weakness or flaw in systems, software, or processes that can be exploited
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Vulnerability name
  - description: text - Vulnerability description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - vulnerability_id: string - CVE or unique identifier
  - severity: enum (critical, high, medium, low) - Vulnerability severity
  - cvss_score: number - CVSS base score (0-10)
  - cvss_vector: string - CVSS vector string
  - affected_component_id: uuid - Affected component or system
  - discovered_date: date - Vulnerability discovery date
  - published_date: date - Vulnerability publication date
  - patch_available: boolean - Whether a patch is available
  - patch_id: string - Patch or fix version
  - exploit_available: boolean - Whether public exploit exists
  - remediation_status: enum (open, in-progress, mitigated, resolved, accepted) - Remediation status
  - target_resolution_date: date - Target resolution date

#### EA.Security-Incident
- **Domain**: Security
- **Description**: Security incident representing an actual or attempted breach of security policies or controls
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Incident name
  - description: text - Incident description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - incident_type: enum (malware, phishing, dos, data-breach, unauthorized-access, insider-threat, social-engineering) - Incident type
  - severity: enum (critical, high, medium, low) - Incident severity
  - status: enum (open, investigating, contained, eradicated, recovered, closed) - Incident status
  - discovery_date: date - Incident discovery date
  - reported_date: date - Incident report date
  - containment_date: date - Incident containment date
  - resolution_date: date - Incident resolution date
  - affected_assets: array of uuid - Affected systems or assets
  - root_cause: text - Root cause analysis
  - lessons_learned: text - Lessons learned from incident

#### EA.Security-Assessment
- **Domain**: Security
- **Description**: Security assessment representing a systematic evaluation of security controls, vulnerabilities, or compliance
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Assessment name
  - description: text - Assessment description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - assessment_type: enum (vulnerability-scan, penetration-test, risk-assessment, compliance-audit, security-review) - Assessment type
  - scope: text - Assessment scope
  - methodology: string - Assessment methodology or framework
  - performed_by: string - Who performed the assessment (internal, external vendor)
  - start_date: date - Assessment start date
  - end_date: date - Assessment end date
  - findings_count: integer - Number of findings
  - critical_findings: integer - Number of critical findings
  - high_findings: integer - Number of high findings
  - overall_rating: enum (satisfactory, satisfactory-with-improvements-needed, unsatisfactory) - Overall rating
  - report_url: string - Link to assessment report
  - follow_up_required: boolean - Whether follow-up actions are required

#### EA.Security-ThreatIntelligence
- **Domain**: Security
- **Description**: Security threat intelligence representing information about potential or current threats facing the organization
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Threat intelligence name
  - description: text - Threat intelligence description
  - owner: string (ea_teams reference) - Security team responsible
- **Optional Attributes**:
  - threat_type: enum (apt, malware, campaign, threat-actor, indicator-of-compromise) - Threat type
  - confidence_level: enum (high, medium, low) - Confidence in intelligence
  - source: string - Intelligence source (commercial, open-source, government, industry)
  - relevant_since: date - Date threat became relevant
  - relevant_until: date - Date threat relevance expires
  - ioc_ids: array of strings - Indicators of compromise
  - affected_systems: array of uuid - Potentially affected systems
  - mitigation_recommendations: text - Recommended mitigations

### Governance Domain (7 types)

#### EA.Governance-Policy
- **Domain**: Governance
- **Description**: Governance policy defining organizational rules, guidelines, and requirements for IT and business practices
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Policy name
  - description: text - Policy description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - policy_type: enum (it-policy, business-policy, financial-policy, hr-policy) - Policy type
  - category: enum (strategy, architecture, security, data, procurement, compliance) - Policy category
  - scope: text - Policy scope and applicability
  - version: string - Policy version
  - approval_date: date - Policy approval date
  - effective_date: date - Policy effective date
  - review_frequency: enum (annually, bi-annually, tri-annually, on-request) - Review frequency
  - next_review_date: date - Next scheduled review date
  - compliance_mandatory: boolean - Whether compliance is mandatory or advisory
  - enforcement_body: string - Body responsible for enforcement
  - exceptions_allowed: boolean - Whether policy exceptions are permitted

#### EA.Governance-Standard
- **Domain**: Governance
- **Description**: Governance standard defining mandatory or voluntary technical, process, or quality standards
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Standard name
  - description: text - Standard description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - standard_type: enum (technical, process, quality, security, industry, regulatory) - Standard type
  - source: string - Standard source (ISO, IEEE, NIST, ITIL, internal)
  - standard_id: string - Standard identifier (e.g., "ISO 27001:2022")
  - version: string - Standard version
  - adoption_status: enum (adopted, partial, under-review, not-adopted) - Adoption status
  - compliance_required: boolean - Whether compliance is required
  - certification_available: boolean - Whether certification is available
  - scope: text - Standard scope within organization

#### EA.Governance-Procedure
- **Domain**: Governance
- **Description**: Governance procedure defining step-by-step processes for implementing policies or standards
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Procedure name
  - description: text - Procedure description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - procedure_type: enum (operational, approval, review, change-management, incident-response) - Procedure type
  - policy_id: uuid - Governing policy this procedure implements
  - scope: text - Procedure scope
  - steps: text - Procedure steps or workflow
  - roles_involved: array of strings - Roles involved in procedure
  - approval_required: boolean - Whether approval is required
  - approver_role: string - Role that must approve
  - sla: string - Service level agreement for procedure completion
  - version: string - Procedure version
  - last_updated_date: date - Last procedure update date

#### EA.Governance-Compliance
- **Domain**: Governance
- **Description**: Governance compliance representing regulatory, legal, or contractual compliance requirements
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Compliance requirement name
  - description: text - Compliance requirement description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - compliance_type: enum (regulatory, contractual, industry, internal) - Compliance type
  - authority: string - Regulatory authority or governing body (e.g., "GDPR", "HIPAA", "SOC 2")
  - requirement_id: string - Requirement identifier or section
  - category: enum (data-privacy, security, financial, environmental, accessibility) - Compliance category
  - risk_level: enum (critical, high, medium, low) - Non-compliance risk level
  - penalty: text - Penalties for non-compliance
  - compliance_status: enum (compliant, non-compliant, partial, under-review, not-applicable) - Compliance status
  - last_audit_date: date - Last compliance audit date
  - next_audit_date: date - Next scheduled audit date
  - evidence_url: string - Link to compliance evidence or documentation
  - control_ids: array of uuid - Controls addressing this requirement (references EA.Security-Control)

#### EA.Governance-Assessment
- **Domain**: Governance
- **Description**: Governance assessment representing evaluations of governance, risk, or compliance posture
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Assessment name
  - description: text - Assessment description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - assessment_type: enum (internal-audit, external-audit, self-assessment, maturity-assessment, gap-analysis) - Assessment type
  - scope: text - Assessment scope
  - framework: string - Framework or methodology used
  - performed_by: string - Who performed assessment (internal, external firm)
  - start_date: date - Assessment start date
  - end_date: date - Assessment end date
  - maturity_level: enum (initial, repeatable, defined, managed, optimized) - Maturity level achieved
  - gap_count: integer - Number of identified gaps
  - risk_count: integer - Number of identified risks
  - recommendation_count: integer - Number of recommendations
  - overall_rating: enum (excellent, good, satisfactory, needs-improvement, unsatisfactory) - Overall rating
  - report_url: string - Link to assessment report
  - action_plan_required: boolean - Whether action plan is required

#### EA.Governance-Exception
- **Domain**: Governance
- **Description**: Governance exception representing a documented exception to a policy, standard, or compliance requirement
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Exception name
  - description: text - Exception description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - exception_type: enum (policy-exception, standard-exception, compliance-waiver) - Exception type
  - policy_id: uuid - Policy or standard being excepted
  - justification: text - Business justification for exception
  - risk_level: enum (critical, high, medium, low) - Risk level from exception
  - mitigation_plan: text - Mitigation plan to manage risk
  - requested_by: string - Person or role requesting exception
  - approved_by: string - Person or role approving exception
  - approval_date: date - Exception approval date
  - expiry_date: date - Exception expiration date (if temporary)
  - status: enum (pending, approved, denied, expired, revoked) - Exception status
  - review_required: boolean - Whether exception requires periodic review

#### EA.Governance-Decision
- **Domain**: Governance
- **Description**: Governance decision representing a documented architectural, investment, or policy decision
- **Required Attributes**:
  - name: string (min: 3, max: 100) - Decision name
  - description: text - Decision description
  - owner: string (ea_teams reference) - Governance team responsible
- **Optional Attributes**:
  - decision_type: enum (architectural, investment, policy, strategic, tactical) - Decision type
  - category: enum (technology, vendor, platform, data, security, governance) - Decision category
  - context: text - Background and context for decision
  - options_considered: array of strings - Alternative options evaluated
  - decision: text - Final decision made
  - rationale: text - Rationale for the decision
  - impact: text - Impact of the decision
  - alternatives: array of strings - Alternatives if primary option fails
  - made_by: string - Decision maker or governance body
  - decision_date: date - Date decision was made
  - effective_date: date - Date decision becomes effective
  - review_date: date - Date decision should be reviewed
  - status: enum (proposed, approved, implemented, deprecated, overturned) - Decision status

## CI Type Count Summary

**Total CI Types:** 61

**By Domain:**
- Strategy: 6
- Business: 10
- Application: 9
- Data: 7
- Technology: 7
- Infrastructure: 8
- Security: 7
- Governance: 7

## Section 4: EA Domain Separation

### EA Domain vs CMDB Taxonomy

**CMDB Taxonomy (existing):**
- Domain → Category → Subcategory → CI Type
- Used for organizing all CI types in the CMDB
- Examples: Infrastructure → Server → Virtual Machine, Application → Database → Relational DB

**EA Domain (new):**
- Separate `ea_domain` field in CI attributes or filtering mechanism
- One of 8 EA domains (Strategy, Business, Application, Data, Technology, Infrastructure, Security, Governance)
- Used for EA-specific filtering and reporting
- Distinct from CMDB Domain/Category/Subcategory

### Relationship Between EA and CMDB

All EA entities are:
1. **CI Types** in the `ci_type_definitions` table with `EA.` prefix
2. **Configuration Items** in the `configuration_items` table
3. **Tagged** with EA domain for filtering
4. **Validated** against EA-specific business rules

**Example:**
- CI Type: `EA.Application-BusinessApp`
- CMDB Taxonomy: Application → Business Application → Enterprise Application
- EA Domain: Application
- Filtering: `ci_type LIKE 'EA.Application-%'` OR tag includes 'Application'

### Cross-Domain Filtering

Query EA entities by domain:
```sql
-- Filter by EA domain using CI type pattern
SELECT * FROM configuration_items
WHERE ci_type LIKE 'EA.Application-%';

-- Alternative: Using tags (if EA domain is added as a tag)
SELECT * FROM configuration_items
WHERE ci_type LIKE 'EA.%' AND 'Application' = ANY(tags);
```

## Section 5: EA Relationship Type Catalog (23 types)

### Core ArchiMate Relationships

#### supports
- **Description**: Source supports or enables target functionality
- **Forward Label**: supports
- **Backward Label**: supported by
- **Bidirectional**: true
- **Source Types**: EA.Application-*, EA.Technology-*, EA.Infrastructure-*
- **Target Types**: EA.Business-*, EA.Application-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Serving
- **Examples**:
  - "CRM Application supports Customer Management Business Capability"
  - "Java Platform supports Order Processing Application"

#### depends_on
- **Description**: Source depends on target for functionality or data
- **Forward Label**: depends on
- **Backward Label**: is depended on by
- **Bidirectional**: true
- **Source Types**: EA.Application-*, EA.Technology-*, EA.Business-*
- **Target Types**: EA.Application-*, EA.Data-*, EA.Technology-*, EA.Infrastructure-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Dependency
- **Examples**:
  - "Application depends on Database"
  - "Service depends on API"

#### realizes
- **Description**: Source implements or realizes target
- **Forward Label**: realizes
- **Backward Label**: realized by
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Technology-*
- **Target Types**: EA.Business-*, EA.Strategy-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Realization
- **Examples**:
  - "Application realizes Business Service"
  - "Component realizes Business Function"

#### flows_to
- **Description**: Data or information flows from source to target
- **Forward Label**: flows to
- **Backward Label**: receives flow from
- **Bidirectional**: true
- **Source Types**: EA.Application-*, EA.Business-*, EA.Data-*
- **Target Types**: EA.Application-*, EA.Data-*, EA.Business-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Flow
- **Examples**:
  - "Data flows from Application to Database"
  - "Information flows to Business Process"

#### assigned_to
- **Description**: Source is assigned to target (responsibility association)
- **Forward Label**: assigned to
- **Backward Label**: has assigned
- **Bidirectional**: false
- **Source Types**: EA.Business-*, EA.Application-*
- **Target Types**: EA.Business-*, EA.Infrastructure-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Assignment
- **Examples**:
  - "Business Role assigned to Business Actor"
  - "Application assigned to Business Unit"

#### aggregates
- **Description**: Source aggregates target (collection relationship)
- **Forward Label**: aggregates
- **Backward Label**: aggregated by
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Business-*, EA.Infrastructure-*
- **Target Types**: EA.Application-*, EA.Business-*, EA.Infrastructure-*
- **Cardinality Source**: one
- **Cardinality Target**: many
- **ArchiMate Concept**: Aggregation
- **Examples**:
  - "Application aggregates Modules"
  - "Business Domain aggregates Capabilities"

#### composes
- **Description**: Source composes target (strong composition, target ceases to exist without source)
- **Forward Label**: composes
- **Backward Label**: composed of
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Technology-*
- **Target Types**: EA.Application-*, EA.Technology-*
- **Cardinality Source**: one
- **Cardinality Target**: many
- **ArchiMate Concept**: Composition
- **Examples**:
  - "Application composes Components"
  - "Node composes Devices"

#### accesses
- **Description**: Source accesses target (typically application accessing data)
- **Forward Label**: accesses
- **Backward Label**: accessed by
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Business-*
- **Target Types**: EA.Data-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Access
- **Examples**:
  - "Application accesses Data Object"
  - "Business Service accesses Data Store"

#### associated_with
- **Description**: General association relationship between entities
- **Forward Label**: associated with
- **Backward Label**: associated with
- **Bidirectional**: true
- **Source Types**: EA.* (all types)
- **Target Types**: EA.* (all types)
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Association
- **Examples**:
  - "Policy associated with Control"
  - "Risk associated with Vulnerability"

### EA-Specific Relationships

#### deployed_on
- **Description**: Deployment relationship where source is deployed on target infrastructure
- **Forward Label**: deployed on
- **Backward Label**: hosts
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Technology-*
- **Target Types**: EA.Infrastructure-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (deployment)
- **Examples**:
  - "Application deployed on Server"
  - "Container deployed on VM"

#### runs_on
- **Description**: Execution relationship where source executes on target
- **Forward Label**: runs on
- **Backward Label**: executes
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Technology-*
- **Target Types**: EA.Technology-*, EA.Infrastructure-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (execution)
- **Examples**:
  - "Application runs on Platform"
  - "Service runs on Server"

#### uses
- **Description**: Usage relationship where source uses or consumes target
- **Forward Label**: uses
- **Backward Label**: used by
- **Bidirectional**: true
- **Source Types**: EA.Application-*, EA.Business-*
- **Target Types**: EA.Technology-*, EA.Data-*, EA.Application-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (usage)
- **Examples**:
  - "Application uses IT Component"
  - "Business Process uses Application Service"

#### implements
- **Description**: Implementation relationship where source provides implementation of target
- **Forward Label**: implements
- **Backward Label**: implemented by
- **Bidirectional**: false
- **Source Types**: EA.Technology-*
- **Target Types**: EA.Application-*, EA.Business-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (implementation)
- **Examples**:
  - "Component implements Service Interface"
  - "Platform implements Application Architecture"

#### validates
- **Description**: Validation relationship where source validates or checks target
- **Forward Label**: validates
- **Backward Label**: validated by
- **Bidirectional**: false
- **Source Types**: EA.Security-Control, EA.Governance-Assessment
- **Target Types**: EA.Security-Risk, EA.Security-Vulnerability, EA.Governance-Compliance
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (validation)
- **Examples**:
  - "Control validates Risk"
  - "Assessment validates Compliance"

#### mitigates
- **Description**: Risk mitigation relationship where source reduces or mitigates target risk
- **Forward Label**: mitigates
- **Backward Label**: mitigated by
- **Bidirectional**: false
- **Source Types**: EA.Security-Control
- **Target Types**: EA.Security-Risk, EA.Security-Vulnerability
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (risk management)
- **Examples**:
  - "Control mitigates Risk"
  - "Security Patch mitigates Vulnerability"

#### enforces
- **Description**: Policy enforcement relationship where source enforces target policy
- **Forward Label**: enforces
- **Backward Label**: enforced by
- **Bidirectional**: false
- **Source Types**: EA.Security-Control, EA.Governance-Procedure
- **Target Types**: EA.Security-Policy, EA.Governance-Policy
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (enforcement)
- **Examples**:
  - "Control enforces Policy"
  - "Procedure enforces Standard"

#### assesses
- **Description**: Assessment relationship where source evaluates or assesses target
- **Forward Label**: assesses
- **Backward Label**: assessed by
- **Bidirectional**: false
- **Source Types**: EA.Governance-Assessment, EA.Security-Assessment
- **Target Types**: EA.* (any entity type)
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (assessment)
- **Examples**:
  - "Assessment assesses Control"
  - "Audit assesses Compliance"

#### governs
- **Description**: Governance relationship where source governs or regulates target
- **Forward Label**: governs
- **Backward Label**: governed by
- **Bidirectional**: false
- **Source Types**: EA.Governance-Policy, EA.Governance-Standard
- **Target Types**: EA.Business-*, EA.Application-*, EA.Technology-*, EA.Data-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (governance)
- **Examples**:
  - "Policy governs Process"
  - "Standard governs Application"

#### aligned_with
- **Description**: Strategic alignment relationship between strategy and business or IT
- **Forward Label**: aligned with
- **Backward Label**: aligned with
- **Bidirectional**: true
- **Source Types**: EA.Strategy-*, EA.Business-*
- **Target Types**: EA.Business-*, EA.Application-*, EA.Strategy-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (alignment)
- **Examples**:
  - "Objective aligned with Capability"
  - "Application aligned with Strategy"

#### conforms_to
- **Description**: Compliance relationship where source conforms to target policy or standard
- **Forward Label**: conforms to
- **Backward Label**: conformance of
- **Bidirectional**: false
- **Source Types**: EA.Application-*, EA.Data-*, EA.Technology-*, EA.Business-*
- **Target Types**: EA.Governance-Policy, EA.Governance-Standard, EA.Security-Policy, EA.Governance-Compliance
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (compliance)
- **Examples**:
  - "Application conforms to Security Policy"
  - "Data Store conforms to GDPR Standard"

#### derived_from
- **Description**: Data derivation relationship where source data is derived from target
- **Forward Label**: derived from
- **Backward Label**: source for
- **Bidirectional**: false
- **Source Types**: EA.Data-*
- **Target Types**: EA.Data-*
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (data lineage)
- **Examples**:
  - "Data Mart derived from Data Warehouse"
  - "Report derived from Transaction Data"

#### decomposes
- **Description**: Decomposition relationship breaking down entities into component parts
- **Forward Label**: decomposes
- **Backward Label**: decomposed into
- **Bidirectional**: false
- **Source Types**: EA.Business-CapabilityL1, EA.Application-BusinessApp
- **Target Types**: EA.Business-CapabilityL2, EA.Application-Component
- **Cardinality Source**: one
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (decomposition)
- **Examples**:
  - "Capability L1 decomposes into Capability L2"
  - "Application decomposes into Components"

#### triggers
- **Description**: Event triggering relationship where source initiates target
- **Forward Label**: triggers
- **Backward Label**: triggered by
- **Bidirectional**: false
- **Source Types**: EA.Business-Event, EA.Application-Event
- **Target Types**: EA.Business-Process, EA.Business-Function, EA.Application-Function
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Trigger (event)
- **Examples**:
  - "Event triggers Process"
  - "Application Event triggers Function"

#### specialized_from
- **Description**: Specialization relationship where source is a specialized version of target
- **Forward Label**: specialized from
- **Backward Label**: generalized by
- **Bidirectional**: false
- **Source Types**: EA.* (any entity type)
- **Target Types**: EA.* (any entity type of same domain)
- **Cardinality Source**: many
- **Cardinality Target**: one
- **ArchiMate Concept**: Specialization
- **Examples**:
  - "Business Process specialized from Generic Process"
  - "Application Component specialized from Reusable Component"

#### governed_by
- **Description**: Governance relationship where source is governed by target policy or decision
- **Forward Label**: governed by
- **Backward Label**: governs
- **Bidirectional**: false
- **Source Types**: EA.* (any entity type)
- **Target Types**: EA.Governance-Policy, EA.Governance-Decision, EA.Governance-Procedure
- **Cardinality Source**: many
- **Cardinality Target**: many
- **ArchiMate Concept**: Derived (governance)
- **Examples**:
  - "Application governed by Architecture Decision"
  - "Process governed by IT Policy"

## Section 6: Cross-Domain Relationship Validation Matrix

### Allowed Cross-Domain Connections

| Source Domain | Target Domain | Allowed Relationship Types |
|---------------|---------------|----------------------------|
| **Strategy** | Business | supports, aligned_with, associated_with |
| **Strategy** | Application | supports, aligned_with, associated_with |
| **Strategy** | Data | aligned_with, associated_with |
| **Strategy** | Technology | aligned_with, associated_with |
| **Strategy** | Infrastructure | aligned_with, associated_with |
| **Strategy** | Security | associated_with |
| **Strategy** | Governance | associated_with, governed_by |
| **Business** | Strategy | supports, aligned_with, associated_with |
| **Business** | Application | uses, depends_on, associated_with |
| **Business** | Data | accesses, uses, associated_with |
| **Business** | Technology | uses, depends_on, associated_with |
| **Business** | Infrastructure | uses, runs_on, associated_with |
| **Business** | Security | associated_with, governed_by |
| **Business** | Governance | associated_with, governed_by, conforms_to |
| **Application** | Strategy | realizes, aligned_with, associated_with |
| **Application** | Business | supports, realizes, associated_with |
| **Application** | Data | accesses, flows_to, associated_with |
| **Application** | Technology | uses, depends_on, runs_on, associated_with |
| **Application** | Infrastructure | deployed_on, runs_on, associated_with |
| **Application** | Security | conforms_to, associated_with, governed_by |
| **Application** | Governance | conforms_to, associated_with, governed_by |
| **Data** | Application | flows_to, accessed_by, associated_with |
| **Data** | Business | flows_to, accessed_by, associated_with |
| **Data** | Technology | runs_on, stored_on, associated_with |
| **Data** | Infrastructure | stored_on, associated_with |
| **Data** | Security | conforms_to, associated_with, governed_by |
| **Data** | Governance | conforms_to, associated_with, governed_by |
| **Technology** | Application | supports, implements, associated_with |
| **Technology** | Business | supports, associated_with |
| **Technology** | Data | hosts, associated_with |
| **Technology** | Infrastructure | runs_on, associated_with |
| **Technology** | Security | conforms_to, associated_with, governed_by |
| **Technology** | Governance | conforms_to, associated_with, governed_by |
| **Infrastructure** | Application | hosts, associated_with |
| **Infrastructure** | Business | hosts, associated_with |
| **Infrastructure** | Data | stores, associated_with |
| **Infrastructure** | Technology | hosts, associated_with |
| **Infrastructure** | Security | conforms_to, associated_with, governed_by |
| **Infrastructure** | Governance | conforms_to, associated_with, governed_by |
| **Security** | Strategy | validates, associated_with |
| **Security** | Business | validates, governs, associated_with |
| **Security** | Application | validates, governs, associated_with |
| **Security** | Data | validates, governs, associated_with |
| **Security** | Technology | validates, governs, associated_with |
| **Security** | Infrastructure | validates, governs, associated_with |
| **Security** | Governance | conforms_to, associated_with, governed_by |
| **Governance** | Strategy | governs, associated_with |
| **Governance** | Business | governs, associated_with |
| **Governance** | Application | governs, associated_with |
| **Governance** | Data | governs, associated_with |
| **Governance** | Technology | governs, associated_with |
| **Governance** | Infrastructure | governs, associated_with |
| **Governance** | Security | governs, associated_with |
| **Governance** | Governance | associated_with |

**Notes:**
- All domains can use `associated_with` as a general-purpose relationship
- `governed_by` is the reverse of `governs`
- `conforms_to` is one-way (entity conforms to policy, not vice versa)
- Same-domain relationships (e.g., Business to Business) use domain-specific types like `aggregates`, `composes`, `decomposes`

## Section 7: EA Teams and Data Ownership

### EA Teams Table Schema

**Purpose**: Team-based ownership model for EA entities (separate from individual user ownership)

**Table Definition:**
```sql
CREATE TABLE ea_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

**Seed Data:**
- `enterprise-architecture` - Enterprise Architecture team
- `business-architecture` - Business Architecture team
- `application-architecture` - Application Architecture team
- `data-architecture` - Data Architecture team
- `technology-architecture` - Technology Architecture team
- `infrastructure-architecture` - Infrastructure Architecture team
- `security-architecture` - Security Architecture team
- `governance` - IT Governance team

**Ownership Rules:**
- EA entities reference `ea_teams.name` via `owner` attribute (team ownership)
- Separate from individual `created_by` / `updated_by` user tracking in `configuration_items` table
- One EA entity owned by one EA team (many-to-one from entity to team)
- One EA team can own multiple EA entities (one-to-many from team to entities)
- `owner` field is required for all EA types (in `required_attributes` JSONB schema)

**EA Team to Domain Mapping:**
- `enterprise-architecture` → Strategy domain
- `business-architecture` → Business domain
- `application-architecture` → Application domain
- `data-architecture` → Data domain
- `technology-architecture` → Technology domain
- `infrastructure-architecture` → Infrastructure domain
- `security-architecture` → Security domain
- `governance` → Governance domain

## Section 8: Validation Framework Specification

### Approach: Hybrid Validation (Struct Tags + Custom Functions)

#### Layer 1: Standard Validation (validator/v10 struct tags)

Handles 80% of validation cases declaratively:

- Required fields
- String length (min, max)
- Numeric ranges
- Email format
- UUID format
- Enum values

Applied to request DTOs in handlers:

```go
type CreateEACIRequest struct {
    Name              string `json:"name" validate:"required,min=3,max=100"`
    CIType            string `json:"ci_type" validate:"required"`
    Description       string `json:"description" validate:"max=1000"`
    Owner             string `json:"owner" validate:"required"`
    StrategicAlignment string `json:"strategic_alignment" validate:"omitempty,oneof=high medium low"`
    TargetDate        string `json:"target_date" validate:"omitempty,datetime=2006-01-02"`
}
```

#### Layer 2: EA-Specific Validation (custom functions)

Handles EA business logic:

- Cross-domain relationship validation
- Parent-child hierarchy validation
- Domain-specific attribute rules
- Data quality scoring
- Override mechanism

Implemented in `internal/ea/validation.go`:

```go
func ValidateCrossDomainRelationship(sourceType, targetType, relationshipType string) error
func ValidateBusinessCapabilityAttributes(attributes map[string]interface{}) error
func ValidateApplicationServiceAttributes(attributes map[string]interface{}) error
// ... other domain validators
```

### Validation Strictness: Warn-But-Allow

#### Approach:
1. Validate all attributes against CI type definition
2. Log validation errors to `validation_errors` JSONB field in CI `attributes`
3. Set `data_quality_score` based on percentage of valid attributes
4. Save entity even with validation errors (don't block creation)
5. Display validation status in data quality dashboard
6. Admin users can override validation with justification

#### Data Quality Score Calculation:
```
score = (valid_attributes / total_validatable_attributes) * 100
```

Where:
- `valid_attributes` = Number of attributes that pass validation
- `total_validatable_attributes` = Number of attributes with validation rules defined
- Score range: 0-100

#### Override Mechanism:
- Admin-only feature (check user has admin role)
- Request fields: `override_validation: boolean`, `override_justification: string`
- Justification required when `override=true`
- Logged to `audit_logs` table
- Triggers data quality alert in dashboard

## Section 9: Metadata Attributes Specification

### EA Metadata Attributes (All EA Types)

**Required Metadata:**
- `source` - string - Data source (manual, import, discovery)
- `last_updated_by` - string (user reference) - User who last updated
- `data_quality_score` - integer (0-100) - Validation score percentage
- `validation_errors` - array of strings - List of validation errors (empty if none)

**Optional Metadata:**
- `documentation_url` - string - Link to external documentation (Confluence, Google Drive)
- `stale_since` - date - Date entity marked as stale (no updates for 6+ months)
- `review_date` - date - Next review date for governance

**Metadata Storage:**
Metadata fields are stored in the `attributes` JSONB column of the `configuration_items` table alongside domain-specific attributes. This allows flexible evolution of metadata requirements without schema changes.

## Section 10: CI Type Definition Versioning

### Version Column:
- Add `version INTEGER DEFAULT 1` column to `ci_type_definitions` table
- Initial version: 1 for all EA CI types
- Increment version when attributes change

### Versioning Strategy (Future - Phase 2):
- Existing EA entities retain old version (backwards compatibility)
- New EA entities use latest version
- Migration path: update script to transform old entities to new schema

### Phase 1 Implementation:
- Add `version INTEGER DEFAULT 1` column to `ci_type_definitions` table
- All EA types start at version 1
- No versioning logic implemented yet (placeholder for future)

---

## Appendix: Summary Statistics

### CI Type Statistics
- **Total CI Types Defined:** 61
- **Total Domains:** 8
- **Average Types per Domain:** 7.6
- **Min Types per Domain:** 6 (Strategy)
- **Max Types per Domain:** 10 (Business)

### Relationship Type Statistics
- **Total Relationship Types Defined:** 23
- **Core ArchiMate Relationships:** 9
- **EA-Specific Relationships:** 14
- **Bidirectional Relationships:** 9 (39%)
- **Unidirectional Relationships:** 14 (61%)

### Domain Coverage
- **CI Type Domain Coverage:** 100% (8/8 domains represented)
- **Relationship Domain Coverage:** 100% (all domains participate as sources and targets)
- **Cross-Domain Matrix Size:** 64 domain pair combinations (8×8)

---

**End of EA Metamodel Specifications v1.0**

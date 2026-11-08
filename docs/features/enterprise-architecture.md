# Enterprise Architecture

The Enterprise Architecture (EA) module adds an EA metamodel on top of the CMDB. EA entities are ordinary
configuration items whose CI types carry the `EA.` prefix; each CI type names one of eight EA domains, and the
domain is derived from the CI type name. This keeps EA inventories inside the same attribute, relationship,
lifecycle, and audit machinery as the rest of the CMDB.

## Status

| Capability | Status | Notes |
| --- | --- | --- |
| EA metamodel and seeded CI types | Implemented | 32 EA CI types across eight domains |
| EA entity CRUD and validation | Implemented | Warn-but-allow data quality scoring; eight seeded teams |
| Lifecycle transition validation | Implemented | Proposed/Active/Deprecated/Retired state machine |
| CSV import wizard | Partial | EA CI types only; import status endpoint is a placeholder |
| Cross-domain relationship validation | Partial | Service-layer rule; no EA-specific relationship route |

## Domains

The domain is taken from the first segment of the CI type name, for example `EA.Application-BusinessApplication`.
Valid domains are `Strategy`, `Business`, `Application`, `Data`, `Technology`, `Infrastructure`, `Security`, and
`Governance`; any other prefix is rejected.

| Domain | Example CI types |
| --- | --- |
| Strategy | Objective, Initiative, Program, Project |
| Business | Organization, BusinessDomain, CapabilityL1, CapabilityL2, BusinessProduct |
| Application | ApplicationGroup, BusinessApplication, Subsystem, Interface, SupportingApplication |
| Data | DataDomain, DataObject |
| Technology | ITComponent, TechCategory, Provider |
| Infrastructure | Location, DataCenter, NetworkZone, ComputePlatform, NetworkSecurityNodes |
| Security | Function, Category, Subcategory, Control |
| Governance | Policy, Procedure, Standard, StandardComponent |

The metamodel migration seeds these types with required attributes (`name`, `description`, `owner`) and optional
domain-specific attributes, plus EA relationship types such as capability decomposition, application-to-capability,
and governance composition.

## Entities

An EA entity is a row in `configuration_items` with an `EA.*` CI type. Creating one stores the supplied attributes
and adds EA metadata: `ea_domain`, `ea_owner` (the owning EA team name), and `ea_team_id`. The domain is appended
to the tags. Entities also carry a lifecycle status and a `data_quality_score`.

| Endpoint | Purpose |
| --- | --- |
| `GET /ea/entities` | List by domain, CI type, lifecycle status, tags, and search |
| `GET /ea/entities/{id}` | Read one entity |
| `POST /ea/entities` | Create an entity |
| `PUT /ea/entities/{id}` | Update an entity |
| `DELETE /ea/entities/{id}` | Delete; entities with relationships require a forced delete |
| `GET /ea/entities/{id}/validate` | Re-run attribute validation |
| `GET /ea/entities/{id}/audit` | Paginated audit history |
| `GET /ea/ci-types`, `GET /ea/teams` | EA CI type and team lookups |

All entity routes require `ea:read`; writes require `ea:create`, `ea:update`, or `ea:delete`.

### Warn-but-allow validation

Missing or invalid required attributes and cross-field rule violations are critical: the request returns 422 and
nothing is stored. Invalid optional attributes are warnings: the entity saves with a `data_quality_score` below 100
equal to the share of valid required attributes, and the response includes a `validation_warnings` list such as
"Data quality score is 75.0% (recommended: 100%)". An admin can pass `override_validation` with an
`override_justification` to save despite critical errors. The frontend form shows warnings without blocking save.

### Lifecycle transitions

EA entities use the lifecycle statuses seeded for EA: Proposed, Active, Deprecated, Retired, and Archived. Entity
updates pass through a transition state machine: Proposed can move to Active or Deprecated; Active to Deprecated or
Retired; Deprecated to Retired; Retired is terminal. Invalid transitions are rejected by the EA service.

EA teams are named ownership groups (`ea_teams`) seeded as enterprise-architecture, business-architecture,
application-architecture, data-architecture, technology-architecture, infrastructure-architecture,
security-architecture, and governance. Entity creation requires the `owner` value to match an existing team name;
an unknown team is rejected with the list of valid teams.

## CSV Import

Bulk onboarding is a four-step wizard in the frontend (`/entities/import`): upload, preview, validate, complete.
The API behind it is EA-specific and requires `ea:create`:

| Endpoint | Behaviour |
| --- | --- |
| `POST /ea/import/template` | Returns a CSV template for one EA CI type, with an example row |
| `POST /ea/import/validate` | Parses and validates a multipart CSV upload (up to 32 MB) |
| `POST /ea/import/execute` | Creates entities from a valid CSV |
| `POST /ea/import/errors/download` | Converts returned row errors into a CSV download |
| `GET /ea/import/status/{batch_id}` | Placeholder; asynchronous imports are not implemented |

CSV columns are `Name`, `CI_Type`, `Domain`, `Lifecycle_Status`, `Owner`, `Team`, `Attributes` (a JSON document),
and `Tags` (comma separated), plus a column per required attribute of the chosen CI type. Validation reports
per-row, per-field errors with expected format and actual value. This is the only bulk import path in the product:
there is no generic CI import, and the status endpoint does not reflect real progress.

## Data Quality

The dashboard (`/ea/data-quality`) is backed by `/api/v1/ea/data-quality` under `ea:read`; all endpoints accept an
optional `domain` filter.

| Endpoint | Returns |
| --- | --- |
| `GET /ea/data-quality` | Total entities, completeness percentage, stale count, entities with errors, lifecycle breakdown, error breakdown by domain |
| `GET /ea/data-quality/stale` | Entities not updated within a threshold (default 90 days) |
| `GET /ea/data-quality/errors` | Entities with non-empty validation errors |
| `GET /ea/data-quality/lifecycle` | Entity counts per lifecycle status |

## Permissions and Routes

| Permission | Admin | Editor | Viewer |
| --- | --- | --- | --- |
| `ea:read` | Yes | Yes | Yes |
| `ea:create` | Yes | Yes | No |
| `ea:update` | Yes | Yes | No |
| `ea:delete` | Yes | No | No |

Frontend routes: `/entities/:domain` (list), `/entities/:domain/create` and
`/entities/:domain/:ciType/create` (create), `/entities/:domain/:id` (details), `/entities/:domain/:id/edit`
(edit), `/entities/import` (wizard), and `/ea/data-quality` (dashboard). The list uses an ag-grid table with row
selection; the form renders attribute fields dynamically from the CI type definition.

See also [CMDB core](cmdb.md), [amortization](amortization.md), [data model](../architecture/data-model.md), and
[project status](../status.md).

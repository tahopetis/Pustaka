# Configuration Management Database

The CMDB core stores configuration items (CIs) with flexible JSONB attributes, organizes them with a flat CI-type
model, records lifecycle status, and maps relationships in Neo4j. This document describes the behaviour that is
implemented today. For table definitions see [data model](../architecture/data-model.md), and for request/response
shapes see the [API reference](../architecture/api.md).

## Status

| Capability | Status | Notes |
| --- | --- | --- |
| CI type definitions | Implemented | Flat model; required and optional attributes stored as JSONB |
| Configuration items | Implemented | JSONB attributes, tags, lifecycle status |
| Lifecycle status administration | Implemented | Includes amortization behaviour |
| Relationship types and relationships | Implemented | Postgres records mirrored to Neo4j |
| Bulk relationship creation | Implemented | From-sources and matrix creation |
| Graph exploration | Implemented | Filtered graph data endpoint |
| QR codes and public CI page | Implemented | Public page is unauthenticated and read-only |
| Hierarchy (domains/categories/subcategories) | Planned | No such tables exist; the taxonomy is flat |

## CI Type Definitions

A CI type is a named schema for CI attributes. There is no domain, category, or subcategory layer: every type is
defined directly in `ci_type_definitions` with a `name`, `description`, an `is_amortizable` flag, and two JSONB
arrays of attribute definitions.

Each attribute carries a `name`, a `type` (`string`, `integer`, `boolean`, `array`, `object`, `date`, and related
types), an optional `description`, and optional validation settings: `pattern`, `min_length`, `max_length`, `min`,
`max`, `enum`, and `format` (email, URL, IPv4, date, and similar). Validation on write checks that required
attributes are present and valid, validates optional attributes only when supplied, and rejects attributes not
declared on the type. CI types are managed through `/api/v1/ci-types` under `ci_type:*` permissions; the
management screen is `/ci-types`.

## Configuration Items

A configuration item is a row in `configuration_items` with a `name` (unique per CI type), a `ci_type` reference,
a JSONB `attributes` document, a text `tags` array, an optional lifecycle status reference, and audit fields. The
list endpoint supports pagination and filters for CI type, free-text search, tags, lifecycle status, and sorting.
CIs are managed through `/api/v1/ci` under `ci:*` permissions, with details, create, and edit screens under `/ci`.

## Lifecycle Statuses

Lifecycle statuses are first-class records (`/api/v1/lifecycle-status`) with name, display name, description,
colour, icon, sort order, active flag, and a system flag. System statuses cannot be modified or deleted. Each
status also carries `amortization_behavior` (`pending`, `active`, `terminal`), which the amortization module uses to
decide whether an asset is depreciated or written off (see [amortization](amortization.md)). The API also exposes
status usage (`/lifecycle-status/usage`) and the CI status distribution (`/lifecycle-status/distribution`) for
dashboard widgets. The frontend provides list, details, and form views under `/lifecycle-status`.

## Relationship Types and Relationships

Relationship types describe how CIs may be connected: machine name, display name, forward and reverse labels,
optional colour and icon, a category for grouping, active and system flags, allowed source and target CI types,
source and target cardinality (`one` or `many`), a bidirectional flag, and an attributes schema. Management is
under `/api/v1/relationship-types` (`relationship_type:*` permissions), including active types for pickers, per-type
usage, statistics, distinct categories, and `POST /validate` for checking a source-type/target-type/type
combination.

Relationship records are written to the `relationships` table and mirrored into Neo4j: creation syncs graph nodes
and the edge, updates and deletes are applied to both stores, and the per-CI relationship list at
`GET /ci/{id}/relationships` is served from the graph service. Relationships are managed through
`/api/v1/relationships` under `relationship:*` permissions. Beyond single creation, two bulk helpers exist:

- `POST /relationships/bulk-sources` creates one relationship type from up to 50 sources to one target;
- `POST /relationships/bulk-matrix` creates the cartesian product of up to 50 sources and 50 targets.

Both return the created records and a created count.

## Graph Exploration

`GET /api/v1/graph` returns nodes and edges for the graph view, while `GET /api/v1/graph/explore` accepts CI type,
search, and limit filters (limit clamped to 500). The frontend `/graph` view renders the returned network with
vis-network and a tooltip component. Cycle detection and most-connected CI analytics are exposed under `/analytics`
(see [dashboards and analytics](dashboards-analytics.md)).

## QR Codes and the Public CI Page

A CI can be identified in the physical world with a QR code:

- `GET /api/v1/qr/ci/{id}` returns the public URL and a base64 PNG data URL;
- `GET /api/v1/qr/ci/{id}/image` returns the PNG image;
- `GET /public/ci/{id}` is unauthenticated and returns a read-only subset: id, name, CI type, attributes, tags,
  lifecycle status, and timestamps.

The QR code encodes the public URL built from the server base URL. The frontend exposes QR download and display
components from the CI details screen.

## Validation Behaviour

CMDB validation is strict: creating or updating a CI validates the attribute document against its CI type, and an
unknown attribute or a missing required attribute is rejected with a validation error rather than stored.

Enterprise Architecture entities add a warn-but-allow layer on top of this: required-attribute and cross-field
failures are still critical and block the save, but an entity that is merely incomplete can be saved with a
`data_quality_score` below 100 and a `validation_warnings` list instead of being rejected. An admin can override
even critical failures by supplying an override flag and justification. See the validation rules in
[enterprise architecture](enterprise-architecture.md).

## Permissions

| Permission | Covers |
| --- | --- |
| `ci:read`, `ci:create`, `ci:update`, `ci:delete` | Configuration item operations |
| `ci_type:read`, `ci_type:create`, `ci_type:update`, `ci_type:delete` | CI type administration |
| `relationship:read`, `relationship:create`, `relationship:update`, `relationship:delete` | Relationship operations |
| `relationship_type:read`, `relationship_type:create`, `relationship_type:update`, `relationship_type:delete` | Relationship type administration |
| `lifecycle_status:read`, `lifecycle_status:create`, `lifecycle_status:update`, `lifecycle_status:delete` | Lifecycle status administration |

Graph, analytics, and QR routes reuse `ci:read`. EA entities are CIs too but use the separate `ea:*` permission
family described in [enterprise architecture](enterprise-architecture.md).

## Related Documentation

- [Enterprise architecture](enterprise-architecture.md)
- [Amortization](amortization.md)
- [Data model](../architecture/data-model.md)
- [Project status](../status.md)

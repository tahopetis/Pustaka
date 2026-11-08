# Audit and Compliance

The audit module records mutations of CMDB objects and exposes them for review, filtering, statistics, and
export. It is implemented as an append-only table plus a read/cleanup API and a frontend viewer. A separate
request-level structured log also exists, but it is not the audit trail described here.

## Status

| Capability | Status | Notes |
| --- | --- | --- |
| Audit trail storage | Implemented | Rows are written by domain services on CI, CI type, relationship, relationship type, lifecycle status, and EA mutations |
| List, filter, and detail API | Implemented | Pagination, filters, search, sorting |
| Statistics | Implemented | Totals, type/action/user breakdowns, recent and daily activity |
| CSV export | Implemented | Limited to 10,000 rows per export |
| Retention cleanup | Partial | Route requires `audit:delete`, a permission no built-in role is granted |
| Single-entry delete | Partial | Same permission gap |
| Request-level audit logging | Partial | Successful writes are logged as structured application log lines, not `audit_logs` rows |
| Amortization financial events | Partial | Amortization writes application log lines; it does not insert `audit_logs` rows |

## What Is Recorded

`audit_logs` rows have this shape:

| Field | Description |
| --- | --- |
| `entity_type` | Domain object type, for example `ci`, `ci_type`, `relationship`, `relationship_type`, `lifecycle_status`, `ea`, `ea_entity`, `ea_relationship` |
| `entity_id` | Identifier of the affected object |
| `action` | `create`, `update`, or `delete` |
| `performed_by` | User who performed the change |
| `timestamp` | When the row was written |
| `details` | JSON document with the change context (entity fields, before/after values where the service records them) |
| `ip_address`, `user_agent` | Optional request metadata; domain services currently pass empty values |

Rows are inserted by the service layer after a successful operation, through each domain's audit service
(CMDB, relationship type, lifecycle status, and EA services). This means the audit trail follows business
transactions, not HTTP plumbing:

- creating, updating, and deleting a CI or CI type;
- creating, updating, and deleting a relationship or relationship type;
- creating, updating, and deleting a lifecycle status;
- creating, updating, and deleting an EA entity or EA relationship.

Reads are not recorded. User administration, login, and logout are not written to `audit_logs` either; login
activity is visible only in the structured application log.

## Request-Level Logging

The audit logging middleware runs after JWT authentication and writes structured application log entries for
successful non-GET requests. It captures method, path, status code, client IP, user agent, the user, and a derived
entity type (the first path segment) and action (`create`, `update`, `delete`). It does not insert database rows
and does not record the entity id. Failed requests and GET requests are skipped. Treat this as operational
observability; the database `audit_logs` table remains the compliance record.

## API

All routes are under `/api/v1/audit` and require `audit:read` unless noted. Admin, editor, and viewer roles all
receive `audit:read` through the default permission seed.

| Endpoint | Purpose |
| --- | --- |
| `GET /audit/logs` | Paginated audit log list |
| `GET /audit/logs/{id}` | Single audit entry |
| `GET /audit/stats` | Aggregate statistics |
| `GET /audit/export` | CSV export of the current filter set |
| `GET /audit` | Frontend-compatible alias with the same filters and a flattened response |
| `DELETE /audit/logs/{id}` | Delete one entry; requires `audit:delete` |
| `DELETE /audit/cleanup` | Delete entries older than a cutoff; requires `audit:delete` |

Filters: `entity_type`, `entity_id`, `action`, `performed_by`, `start_date`, `end_date`, `search`, `sort`,
`order`, `page`, and `limit`. The frontend alias also accepts `entity`, `from_date`, and `to_date` parameter
names and returns `audit_logs` with `total`, `page`, `limit`, and `total_pages`.

The statistics response includes the total event count, counts grouped by entity type, action, and user, a list of
recent activity, and a per-day activity map.

`GET /audit/export` returns CSV with the columns ID, Entity Type, Entity ID, Action, Performed By, Timestamp, IP
Address, and User Agent. The export applies the same filters as the list endpoint and caps at 10,000 rows.

`DELETE /audit/cleanup` accepts an optional `retention_days` query parameter. With no parameter it uses the
service default of 365 days; a value of zero also falls back to that default. It returns the number of deleted
rows, the computed cutoff date, and the retention period used. Both delete routes are gated by `audit:delete`,
which is referenced by the route configuration but has not been inserted as a permission in the migrations, so
no built-in role can call them without a custom role assignment.

## Frontend

The audit viewer lives at `/audit` and requires `audit:read`. It provides a search field, entity type and action
selectors, a performer selector, date range filters, sortable columns, a detail drawer for the selected entry, and
empty and error states. The viewer reads the frontend-compatible alias route.

## Consistency and Retention

- Audit rows are written in the same service call path as the mutation they describe; a failed mutation does not
  produce an audit row.
- The table is append-only from the API's perspective: there is no update endpoint, and deletion is restricted to
  the two `audit:delete` routes.
- The only retention policy in code is the 365-day default used by the cleanup endpoint; there is no scheduled
  cleanup job.
- Archived audit data is out of scope; there is no archive or cold-storage pipeline.

## Related Documentation

- [CMDB core](cmdb.md)
- [Enterprise architecture](enterprise-architecture.md)
- [Dashboards and analytics](dashboards-analytics.md)
- [Data model](../architecture/data-model.md)

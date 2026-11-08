# Dashboards and Analytics

Porta exposes two related read-only surfaces over the same data. Dashboard endpoints aggregate operational health,
data quality, asset age, and risk for the landing dashboard. Analytics endpoints answer graph questions: which
types are used, where cycles exist, which CIs are hubs, how fast the inventory grows, and how relationship types
are used. Both surfaces require `ci:read`; the amortization-specific numbers are read through the amortization
module.

## Status

Implemented. All endpoints below are routed and backed by queries against PostgreSQL, Neo4j, or both. The frontend
dashboard consumes them; there is no separate standalone analytics page.

## Dashboard API

| Endpoint | Returns |
| --- | --- |
| `GET /api/v1/dashboard/stats` | Total CIs, CI types, relationships, and users |
| `GET /api/v1/dashboard/health-score` | Overall score with completeness, correctness, compliance sub-scores, trend, and calculation time |
| `GET /api/v1/dashboard/data-quality` | Missing attributes, orphaned CIs, CIs without lifecycle status, CIs without tags, stale counts for 30/60/90 days, and duplicate count |
| `GET /api/v1/dashboard/asset-aging` | Age distribution buckets, assets approaching end of life, average age, oldest asset |
| `GET /api/v1/dashboard/risk-metrics` | Risk score, single points of failure, critical assets, assets without redundancy, compliance violations, and the highest-risk CIs |

### Health score

The health score has three components:

- **Completeness**: share of CIs with all required attributes;
- **Correctness**: share of CIs updated within the last 90 days;
- **Compliance**: share of CIs following naming standards.

The overall score is the average of the three. The trend compares the current score with a snapshot from about 30
days ago. Snapshots are stored in `health_score_history`, so the trend becomes meaningful as history accumulates.

### Data quality

`GET /dashboard/data-quality` accepts `include_details=true` and `limit` to include affected CIs per issue
category; without details it returns counts and percentages only. Issues covered: missing required attributes,
orphaned CIs (no relationships), CIs without a lifecycle status, CIs without tags, stale CIs at 30/60/90 days, and
duplicate candidates.

### Asset aging

The endpoint accepts `eol_threshold_months` (default 6) and `limit` (default 10). Age buckets are under one year,
one to three years, three to five years, and over five years; the response also includes approaching-end-of-life
assets with days until EOL, the average age in months, and the oldest asset.

### Risk metrics

Risk metrics derive a 0-100 risk score from single points of failure, assets without redundancy, obsolete assets,
and compliance violations. The response includes the top high-risk CIs with their individual risk score, whether
the CI is amortizable, redundancy state, criticality, age in months, and whether it has tags. `limit` controls how
many high-risk CIs are listed (default 10).

## Analytics API

| Endpoint | Returns |
| --- | --- |
| `GET /api/v1/analytics/ci-types/usage` | CI type name and CI count, ordered by usage |
| `GET /api/v1/analytics/cycles` | Relationship cycles, each a list of CI identifiers |
| `GET /api/v1/analytics/most-connected` | CIs with the highest distinct relationship counts; accepts `limit` |
| `GET /api/v1/analytics/ci-growth` | Daily CI and relationship creation counts; accepts `from_date` and `to_date` (`YYYY-MM-DD`) |
| `GET /api/v1/analytics/relationship-types/usage` | Relationship type and relationship count |

The graph-based analytics (usage, cycles, most-connected) are Cypher queries against Neo4j, so they reflect the
relationship mirror, not the PostgreSQL relationship table. CI growth is a PostgreSQL aggregate over the CI and
relationship history tables.

## Frontend

The dashboard at `/dashboard` is a widget grid. Implemented widgets include:

| Widget | Data source |
| --- | --- |
| Summary counters (CIs, CI types, relationships, users) | `/dashboard/stats` |
| Amortization summary (book value, monthly depreciation) | Amortization summaries; shown only with `amortization:read` |
| CI status distribution | Lifecycle status distribution |
| Activity heatmap | Relationship creation counts over recent weeks |
| Health score | `/dashboard/health-score` |
| Data quality | `/dashboard/data-quality` |
| Asset aging | `/dashboard/asset-aging` |
| Financial summary | Amortization summaries and ledger aggregates |
| Risk metrics | `/dashboard/risk-metrics` |
| CI type usage, most connected CIs, relationship type usage | `/analytics/*` |
| Trend chart with a 7/30/90-day range filter | CI growth and time-series data |

Widgets degrade independently: a widget whose endpoint fails shows an empty state without blocking the page.
Permission-gated widgets are hidden when the user lacks the required permission.

## Notes and Limits

- Analytics responses are not paginated; graph-derived lists are bounded by an explicit `limit` where the endpoint
  accepts one.
- `GET /analytics/cycles` returns raw identifier lists; it is intended for diagnostics rather than display.
- A dedicated network/impact-analysis API exists in the handler layer but is not routed; the dashboard's network
  analytics card is a user-facing summary of the analytics endpoints above.
- The dashboard has no persisted widget configuration; the layout is fixed in the view.

## Related Documentation

- [CMDB core](cmdb.md)
- [Amortization](amortization.md)
- [Audit and compliance](audit-compliance.md)
- [API reference](../architecture/api.md)

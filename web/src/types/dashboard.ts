/**
 * Dashboard Type Definitions
 *
 * Comprehensive TypeScript interfaces for the dashboard enhancement project.
 * These types correspond to API responses from multiple backend endpoints.
 */

// ============================================================================
// Time Range and Filtering
// ============================================================================

/**
 * Time range configuration for dashboard filtering
 * Used by TimeRangeFilter component to manage date selections
 */
export interface TimeRange {
  /** Start date in ISO format (YYYY-MM-DD) or null for no start filter */
  startDate: string | null
  /** End date in ISO format (YYYY-MM-DD) or null for no end filter */
  endDate: string | null
  /** Preset value for quick selection or 'custom' for custom range */
  preset: '7' | '30' | '90' | 'custom'
}

// ============================================================================
// Dashboard Statistics - GET /api/v1/dashboard/stats
// ============================================================================

/**
 * Dashboard statistics response containing counts of all major entities
 * Returned by GET /api/v1/dashboard/stats endpoint
 */
export interface DashboardStats {
  /** Total number of Configuration Items */
  total_cis: number
  /** Total number of CI Types defined */
  total_ci_types: number
  /** Total number of relationships between CIs */
  total_relationships: number
  /** Total number of users in the system */
  total_users: number
}

// ============================================================================
// Audit Statistics - GET /api/v1/audit/stats
// ============================================================================

/**
 * Comprehensive audit log statistics with filtering support
 * Returned by GET /api/v1/audit/stats?from_date={date}&to_date={date}
 */
export interface AuditStats {
  /** Total count of audit events in the filtered period */
  total_events: number

  /**
   * Event counts grouped by entity type (e.g., 'ci', 'relationship', 'user')
   * Key: entity_type, Value: count
   */
  events_by_type: Record<string, number>

  /**
   * Event counts grouped by action (e.g., 'create', 'update', 'delete')
   * Key: action, Value: count
   */
  events_by_action: Record<string, number>

  /**
   * Event counts grouped by username
   * Key: username, Value: count
   */
  events_by_user: Record<string, number>

  /**
   * Daily activity counts for time-series charts
   * Key: date (YYYY-MM-DD), Value: event count for that day
   * Used for trend line charts and activity heatmaps
   */
  daily_activity: Record<string, number>

  /** Recent audit log entries (typically last 10-20 events) */
  recent_activity: AuditLogEntry[]
}

/**
 * Individual audit log entry
 */
export interface AuditLogEntry {
  /** Unique identifier for the audit log entry */
  id: string
  /** Type of entity that was modified (e.g., 'ci', 'relationship', 'user') */
  entity_type: string
  /** ID of the entity that was modified */
  entity_id: string
  /** Action performed ('create', 'update', 'delete') */
  action: string
  /** ID of the user who performed the action */
  performed_by: string
  /** Username of the user who performed the action (may be included via join) */
  username?: string
  /** Timestamp of when the action was performed (ISO 8601 format) */
  timestamp: string
  /** Additional details about the change (old/new values, etc.) */
  details?: Record<string, any>
  /** IP address of the user who performed the action */
  ip_address?: string
}

// ============================================================================
// CI Type Analytics - GET /api/v1/analytics/ci-types/usage
// ============================================================================

/**
 * CI type usage statistics showing distribution of CIs by type
 * Returned by GET /api/v1/analytics/ci-types/usage endpoint
 */
export interface CITypeUsage {
  /** CI type name (e.g., 'Server', 'Database', 'Application') */
  type: string
  /** Number of CIs of this type */
  count: number
}

// ============================================================================
// Network Analytics - GET /api/v1/analytics/most-connected
// ============================================================================

/**
 * Configuration Item with connectivity metrics
 * Returned by GET /api/v1/analytics/most-connected?limit=10 endpoint
 * Used for network analytics visualizations
 */
export interface MostConnectedCI {
  /** UUID of the Configuration Item */
  id: string
  /** Display name of the CI */
  name: string
  /** CI type (e.g., 'Server', 'Database') */
  type: string
  /** Number of relationships (connections) this CI has */
  connection_count: number
}

// ============================================================================
// Relationship Type Analytics - GET /api/v1/analytics/relationship-types/usage
// ============================================================================

/**
 * Relationship type with basic metadata
 */
export interface RelationshipType {
  /** UUID of the relationship type */
  id: string
  /** Internal name identifier */
  name: string
  /** User-friendly display name */
  display_name?: string
  /** Description of what this relationship type represents */
  description?: string
  /** Label for forward direction (e.g., 'depends on') */
  forward_label: string
  /** Label for reverse direction (e.g., 'required by') */
  reverse_label: string
  /** Hex color code for visualization (e.g., '#FF5733') */
  color?: string
  /** Icon identifier for UI display */
  icon?: string
  /** Category grouping (e.g., 'dependency', 'containment') */
  category?: string
  /** Whether this type is currently active/usable */
  is_active: boolean
  /** Whether this is a system-defined type (non-deletable) */
  is_system: boolean
  /** Allowed CI types for the source end */
  allowed_source_types?: string[]
  /** Allowed CI types for the target end */
  allowed_target_types?: string[]
  /** Cardinality constraint for source ('one' or 'many') */
  cardinality_source: string
  /** Cardinality constraint for target ('one' or 'many') */
  cardinality_target: string
  /** Whether the relationship is bidirectional */
  bidirectional: boolean
  /** Custom attributes schema */
  attributes?: Record<string, any>
  /** Timestamp when this type was created (ISO 8601) */
  created_at: string
  /** Timestamp when this type was last updated (ISO 8601) */
  updated_at?: string
  /** UUID of user who created this type */
  created_by: string
  /** UUID of user who last updated this type */
  updated_by?: string
}

/**
 * Relationship type usage statistics
 * Returned by GET /api/v1/analytics/relationship-types/usage endpoint
 */
export interface RelationshipTypeUsage {
  /** The relationship type with full metadata */
  relationship_type: RelationshipType
  /** Number of relationships using this type */
  usage_count: number
}

// ============================================================================
// Growth Analytics - GET /api/v1/analytics/ci-growth (TO BE IMPLEMENTED)
// ============================================================================

/**
 * Time-series growth data for CIs and relationships
 * This endpoint will be implemented in Phase 4 of the dashboard enhancement
 */
export interface GrowthData {
  /** CI creation counts over time */
  ci_growth: DailyCount[]
  /** Relationship creation counts over time */
  relationship_growth: DailyCount[]
}

/**
 * Daily count data point for time-series charts
 */
export interface DailyCount {
  /** Date in YYYY-MM-DD format */
  date: string
  /** Count of items created/modified on this date */
  count: number
}

// ============================================================================
// Chart Data Structures for D3.js
// ============================================================================

/**
 * Generic chart data point for D3.js visualizations
 * Used for trend charts, bar charts, and other D3 visualizations
 */
export interface ChartDataPoint {
  /** X-axis value (typically date string or category label) */
  x: string | number
  /** Y-axis value (typically a count or metric) */
  y: number
  /** Optional label for the data point */
  label?: string
  /** Optional color override for this specific data point */
  color?: string
  /** Optional metadata for tooltips or interactions */
  metadata?: Record<string, any>
}

/**
 * Multi-series chart data structure
 * Used when displaying multiple trend lines or comparison charts
 */
export interface ChartSeries {
  /** Series identifier */
  id: string
  /** Display name for the series */
  name: string
  /** Color for this series (hex code) */
  color: string
  /** Data points in this series */
  data: ChartDataPoint[]
}

/**
 * Donut/Pie chart data structure
 */
export interface DonutChartData {
  /** Category/segment label */
  label: string
  /** Numeric value */
  value: number
  /** Color for this segment (hex code) */
  color: string
  /** Percentage of total (calculated in component) */
  percentage?: number
}

/**
 * Heatmap data structure for activity calendar
 */
export interface HeatmapData {
  /** Date in YYYY-MM-DD format */
  date: string
  /** Activity count for this date */
  value: number
  /** Intensity level (calculated based on value, e.g., 0-4) */
  intensity?: number
}

// ============================================================================
// Component State Interfaces
// ============================================================================

/**
 * Loading state for dashboard widgets
 */
export interface DashboardLoadingState {
  /** Whether stats are currently loading */
  stats: boolean
  /** Whether audit data is loading */
  audit: boolean
  /** Whether CI type usage is loading */
  ciTypes: boolean
  /** Whether network analytics is loading */
  network: boolean
  /** Whether relationship type data is loading */
  relationshipTypes: boolean
  /** Whether growth data is loading */
  growth: boolean
}

/**
 * Error state for dashboard widgets
 */
export interface DashboardErrorState {
  /** Error message for stats, null if no error */
  stats: string | null
  /** Error message for audit data, null if no error */
  audit: string | null
  /** Error message for CI types, null if no error */
  ciTypes: string | null
  /** Error message for network analytics, null if no error */
  network: string | null
  /** Error message for relationship types, null if no error */
  relationshipTypes: string | null
  /** Error message for growth data, null if no error */
  growth: string | null
}

/**
 * Complete dashboard data state
 * Used by useDashboardData composable
 */
export interface DashboardData {
  /** Basic dashboard statistics */
  stats: DashboardStats | null
  /** Audit statistics with time filtering */
  auditStats: AuditStats | null
  /** CI type distribution data */
  ciTypeUsage: CITypeUsage[]
  /** Most connected CIs for network analytics */
  mostConnected: MostConnectedCI[]
  /** Relationship type usage statistics */
  relationshipTypeUsage: RelationshipTypeUsage[]
  /** Growth trend data (when implemented) */
  growthData: GrowthData | null
  /** Current time range filter */
  timeRange: TimeRange
  /** Loading states for each data source */
  loading: DashboardLoadingState
  /** Error states for each data source */
  errors: DashboardErrorState
  /** Timestamp of last data refresh */
  lastRefresh: Date | null
}

// ============================================================================
// API Response Wrappers (if needed for error handling)
// ============================================================================

/**
 * Generic API response wrapper
 */
export interface ApiResponse<T> {
  /** Response data */
  data?: T
  /** Error message if request failed */
  error?: string
  /** HTTP status code */
  status: number
  /** Success flag */
  success: boolean
}

// ============================================================================
// Widget Configuration Types
// ============================================================================

/**
 * Configuration for dashboard widget appearance and behavior
 */
export interface WidgetConfig {
  /** Widget title */
  title: string
  /** Widget description for tooltips */
  description?: string
  /** Whether to show the refresh button */
  showRefresh?: boolean
  /** Whether the widget is collapsible */
  collapsible?: boolean
  /** Default collapsed state */
  defaultCollapsed?: boolean
  /** Custom icon for the widget header */
  icon?: string
  /** Custom CSS classes */
  customClass?: string
}

// ============================================================================
// Type Guards and Utility Types
// ============================================================================

/**
 * Type guard to check if a value is a valid TimeRange preset
 */
export function isTimeRangePreset(value: string): value is '7' | '30' | '90' | 'custom' {
  return ['7', '30', '90', 'custom'].includes(value)
}

/**
 * Type guard to check if an object has the structure of DashboardStats
 */
export function isDashboardStats(obj: any): obj is DashboardStats {
  return (
    obj &&
    typeof obj.total_cis === 'number' &&
    typeof obj.total_ci_types === 'number' &&
    typeof obj.total_relationships === 'number' &&
    typeof obj.total_users === 'number'
  )
}

/**
 * Type guard to check if an object has the structure of AuditStats
 */
export function isAuditStats(obj: any): obj is AuditStats {
  return (
    obj &&
    typeof obj.total_events === 'number' &&
    typeof obj.events_by_type === 'object' &&
    typeof obj.events_by_action === 'object' &&
    typeof obj.daily_activity === 'object'
  )
}

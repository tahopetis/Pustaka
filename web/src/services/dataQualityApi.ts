import { api } from './api'

// ============================================================================
// Types
// ============================================================================

export interface DataQualityMetrics {
  total_entities: number
  completeness_pct: number
  stale_entities_count: number
  entities_with_errors_count: number
  lifecycle_breakdown: Record<string, number>
  error_breakdown_by_domain: Record<string, number>
  generated_at: string
}

export interface StaleEntityCriteria {
  days_threshold: number
  include_incomplete: boolean
}

export interface EAEntitySummary {
  id: string
  name: string
  ci_type: string
  ea_domain: string
  data_quality_score: number
  updated_at: string
}

export interface StaleEntitiesResponse {
  entities: EAEntitySummary[] | null
  total: number
  query: StaleEntityCriteria
}

export interface EntitiesWithErrorsResponse {
  entities: EAEntitySummary[] | null
  total: number
  domain: string | null
}

export interface LifecycleBreakdownResponse {
  breakdown: Record<string, number>
  domain: string | null
  generated_at: string
}

// ============================================================================
// API Service
// ============================================================================

/**
 * Get EA data quality metrics
 * @param domain Optional domain filter
 */
export async function getMetrics(domain?: string): Promise<DataQualityMetrics> {
  const params = domain ? { domain } : undefined
  const response = await api.get<DataQualityMetrics>('/ea/data-quality', { params })
  return response.data
}

/**
 * Get stale EA entities
 */
export async function getStaleEntities(
  criteria?: Partial<StaleEntityCriteria>
): Promise<StaleEntitiesResponse> {
  const params = criteria ? criteria : undefined
  const response = await api.get<StaleEntitiesResponse>('/ea/data-quality/stale', { params })
  return response.data
}

/**
 * Get EA entities with errors
 * @param domain Optional domain filter
 */
export async function getEntitiesWithErrors(domain?: string): Promise<EntitiesWithErrorsResponse> {
  const params = domain ? { domain } : undefined
  const response = await api.get<EntitiesWithErrorsResponse>('/ea/data-quality/errors', { params })
  return response.data
}

/**
 * Get lifecycle status breakdown
 * @param domain Optional domain filter
 */
export async function getLifecycleBreakdown(domain?: string): Promise<LifecycleBreakdownResponse> {
  const params = domain ? { domain } : undefined
  const response = await api.get<LifecycleBreakdownResponse>('/ea/data-quality/lifecycle', { params })
  return response.data
}

// Export as default object for convenience
const dataQualityApi = {
  getMetrics,
  getStaleEntities,
  getEntitiesWithErrors,
  getLifecycleBreakdown,
}

export default dataQualityApi

import { api } from './api'
import type {
  EAEntity,
  EACreateRequest,
  EAUpdateRequest,
  EAFilter,
  PaginationMeta,
  AuditLogsResponse
} from '@/types/ea'

export interface EntityListResponse {
  entities: EAEntity[]
  meta: PaginationMeta
}

export const eaApi = {
  /**
   * Create a new EA entity
   * POST /api/v1/ea/entities
   */
  createEntity: (data: EACreateRequest) =>
    api.post<EAEntity>('/ea/entities', data),

  /**
   * Get a single EA entity by ID
   * GET /api/v1/ea/entities/{id}
   */
  getEntity: (id: string) =>
    api.get<EAEntity>(`/ea/entities/${id}`),

  /**
   * Update an EA entity
   * PUT /api/v1/ea/entities/{id}
   */
  updateEntity: (id: string, data: EAUpdateRequest) =>
    api.put<EAEntity>(`/ea/entities/${id}`, data),

  /**
   * Delete an EA entity
   * DELETE /api/v1/ea/entities/{id}
   * @param force - Set to true to bypass relationship check after confirmation
   */
  deleteEntity: (id: string, force: boolean = false) =>
    api.delete(`/ea/entities/${id}${force ? '?force=true' : ''}`),

  /**
   * List EA entities with filtering and pagination
   * GET /api/v1/ea/entities
   */
  listEntities: (filter: EAFilter) => {
    const params: Record<string, any> = {}

    if (filter.domain) params.domain = filter.domain
    if (filter.ci_type) params.ci_type = filter.ci_type
    if (filter.lifecycle_status_id) params.lifecycle_status_id = filter.lifecycle_status_id
    if (filter.search) params.search = filter.search
    if (filter.page) params.page = filter.page
    if (filter.page_size) params.page_size = filter.page_size

    return api.get<EntityListResponse>('/ea/entities', { params })
  },

  /**
   * Validate an EA entity
   * GET /api/v1/ea/entities/{id}/validate
   */
  validateEntity: (id: string) =>
    api.get(`/ea/entities/${id}/validate`),

  /**
   * List EA CI types
   * GET /api/v1/ea/ci-types
   */
  listCiTypes: (params?: { page?: number; limit?: number; search?: string }) =>
    api.get('/ea/ci-types', { params }),

  /**
   * Get a specific EA CI type
   * GET /api/v1/ea/ci-types/{name}
   */
  getCiType: (name: string) =>
    api.get(`/ea/ci-types/${name}`),

  /**
   * Get audit logs for an EA entity
   * GET /api/v1/ea/entities/{id}/audit
   */
  getEntityAuditLogs: (id: string, params?: { page?: number; page_size?: number }) => {
    const queryParams: Record<string, any> = {}
    if (params?.page) queryParams.page = params.page
    if (params?.page_size) queryParams.page_size = params.page_size

    return api.get<AuditLogsResponse>(`/ea/entities/${id}/audit`, { params: queryParams })
  },

  /**
   * List EA teams
   * GET /api/v1/ea/teams
   */
  listTeams: () =>
    api.get('/ea/teams')
}

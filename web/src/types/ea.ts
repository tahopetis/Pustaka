// EA (Enterprise Architecture) Type Definitions

export interface EAEntity {
  id: string
  name: string
  ci_type: string
  ci_type_display: string
  domain: string
  lifecycle_status_id: string | null
  lifecycle_status_display: string
  attributes: Record<string, any>
  tags: string[]
  owner_id: string | null
  owner_name: string | null
  team_id: string | null
  team_name: string | null
  data_quality_score: number
  created_at: string
  updated_at: string
}

export interface EACreateRequest {
  name: string
  ci_type: string
  lifecycle_status_id?: string
  attributes?: Record<string, any>
  tags?: string[]
}

export interface EAUpdateRequest {
  name?: string
  lifecycle_status_id?: string
  attributes?: Record<string, any>
  tags?: string[]
}

export interface EAFilter {
  domain?: string
  ci_type?: string
  lifecycle_status_id?: string
  search?: string
  page?: number
  page_size?: number
}

export interface ValidationError {
  field: string
  message: string
  code: string
}

export interface PaginationMeta {
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface FieldGroup {
  name: string
  title: string
  collapsed: boolean
  attributes: AttributeSchema[]
}

export interface AttributeSchema {
  name: string
  type: string
  description?: string
  required: boolean
  validation?: {
    pattern?: string
    min_length?: number
    max_length?: number
    min?: number
    max?: number
    enum?: string[]
    format?: string
  }
}

export interface CITypeDefinition {
  id: string
  name: string
  description?: string
  required_attributes: AttributeSchema[]
  optional_attributes: AttributeSchema[]
}

export interface AuditLog {
  id: string
  timestamp: string
  action: string
  user_id: string
  user_name?: string
  details: Record<string, any>
  ip_address?: string
  user_agent?: string
}

export interface AuditLogsResponse {
  audit_logs: AuditLog[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

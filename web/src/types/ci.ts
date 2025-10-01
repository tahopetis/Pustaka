export interface CI {
  id: string
  name: string
  ci_type: string
  attributes: Record<string, any>
  tags: string[]
  created_at: string
  updated_at: string
  created_by: string
  updated_by?: string
}

export interface CIType {
  id: string
  name: string
  description?: string
  required_attributes: any[]
  optional_attributes: any[]
  created_at: string
  updated_at: string
  created_by: string
}

export interface Relationship {
  id: string
  source_id: string
  target_id: string
  relationship_type: string
  attributes: Record<string, any>
  created_at: string
  created_by: string
  source_ci?: CI
  target_ci?: CI
}

export interface CreateCIData {
  name: string
  ci_type: string
  attributes: Record<string, any>
  tags?: string[]
}

export interface UpdateCIData {
  attributes?: Record<string, any>
  tags?: string[]
}

export interface CIListResponse {
  cis: CI[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface CITypeListResponse {
  ci_types: CIType[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface RelationshipListResponse {
  relationships: Relationship[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface GraphNode {
  id: string
  name: string
  type: string
  attributes: Record<string, any>
  x?: number
  y?: number
  color?: string
}

export interface GraphEdge {
  source: string
  target: string
  type: string
  attributes?: Record<string, any>
  label?: string
}

export interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
}
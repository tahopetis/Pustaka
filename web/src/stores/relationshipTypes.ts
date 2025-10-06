import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { relationshipTypeAPI } from '@/services/api'

export interface RelationshipType {
  id: string
  name: string
  display_name?: string
  description?: string
  forward_label: string
  reverse_label: string
  color?: string
  icon?: string
  category?: string
  is_active: boolean
  is_system: boolean
  allowed_source_types?: string[]
  allowed_target_types?: string[]
  cardinality_source?: string
  cardinality_target?: string
  bidirectional?: boolean
  created_at: string
  updated_at: string
  created_by?: string
  updated_by?: string
}

export interface RelationshipTypeCreateData {
  name: string
  display_name?: string
  description?: string
  forward_label: string
  reverse_label: string
  color?: string
  icon?: string
  category?: string
  allowed_source_types?: string[]
  allowed_target_types?: string[]
  cardinality_source?: string
  cardinality_target?: string
  bidirectional?: boolean
}

export interface RelationshipTypeUpdateData {
  display_name?: string
  description?: string
  forward_label?: string
  reverse_label?: string
  color?: string
  icon?: string
  is_active?: boolean
  category?: string
  allowed_source_types?: string[]
  allowed_target_types?: string[]
  cardinality_source?: string
  cardinality_target?: string
  bidirectional?: boolean
}

export interface RelationshipTypeStats {
  total_types: number
  active_types: number
  system_types: number
  custom_types: number
  total_relationships: number
  usage_by_category: Record<string, {
    type_count: number
    relationship_count: number
  }>
  most_used_types: Array<{
    type: RelationshipType
    relationship_count: number
  }>
}

export const useRelationshipTypeStore = defineStore('relationshipType', () => {
  const relationshipTypes = ref<RelationshipType[]>([])
  const categories = ref<string[]>([])
  const stats = ref<RelationshipTypeStats | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  // Get all active relationship types
  const activeRelationshipTypes = computed(() =>
    relationshipTypes.value.filter(type => type.is_active)
  )

  // Get relationship types by ID
  const getRelationshipTypeById = (id: string) =>
    relationshipTypes.value.find(type => type.id === id)

  // Get relationship types by name
  const getRelationshipTypeByName = (name: string) =>
    relationshipTypes.value.find(type => type.name === name)

  // Get relationship types by category
  const getRelationshipTypesByCategory = (category: string) =>
    relationshipTypes.value.filter(type => type.category === category)

  // Load relationship types
  const loadRelationshipTypes = async (params?: {
    page?: number
    limit?: number
    search?: string
    is_active?: boolean
    category?: string
  }) => {
    loading.value = true
    error.value = null

    try {
      const response = await relationshipTypeAPI.list(params)
      relationshipTypes.value = response.data.relationship_types || []
      total.value = response.data.total || 0
      return response.data
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to load relationship types'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Load categories
  const loadCategories = async () => {
    try {
      const response = await relationshipTypeAPI.getCategories()
      categories.value = response.data.categories || []
      return response.data
    } catch (err: any) {
      console.error('Failed to load categories:', err)
      throw err
    }
  }

  // Load stats
  const loadStats = async () => {
    try {
      const response = await relationshipTypeAPI.getStats()
      stats.value = response.data
      return response.data
    } catch (err: any) {
      console.error('Failed to load stats:', err)
      throw err
    }
  }

  // Create relationship type
  const createRelationshipType = async (data: RelationshipTypeCreateData) => {
    loading.value = true
    error.value = null

    try {
      const response = await relationshipTypeAPI.create(data)
      relationshipTypes.value.push(response.data)
      return response.data
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create relationship type'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Update relationship type
  const updateRelationshipType = async (id: string, data: RelationshipTypeUpdateData) => {
    loading.value = true
    error.value = null

    try {
      const response = await relationshipTypeAPI.update(id, data)
      const index = relationshipTypes.value.findIndex(type => type.id === id)
      if (index !== -1) {
        relationshipTypes.value[index] = response.data
      }
      return response.data
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update relationship type'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Delete relationship type
  const deleteRelationshipType = async (id: string) => {
    loading.value = true
    error.value = null

    try {
      await relationshipTypeAPI.delete(id)
      relationshipTypes.value = relationshipTypes.value.filter(type => type.id !== id)
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete relationship type'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Validate relationship compatibility
  const validateCompatibility = async (data: {
    relationship_type: string
    source_type: string
    target_type: string
  }) => {
    try {
      const response = await relationshipTypeAPI.validateCompatibility(data)
      return response.data
    } catch (err: any) {
      console.error('Failed to validate compatibility:', err)
      throw err
    }
  }

  // Clear error
  const clearError = () => {
    error.value = null
  }

  return {
    relationshipTypes,
    categories,
    stats,
    activeRelationshipTypes,
    loading,
    error,
    total,
    getRelationshipTypeById,
    getRelationshipTypeByName,
    getRelationshipTypesByCategory,
    loadRelationshipTypes,
    loadCategories,
    loadStats,
    createRelationshipType,
    updateRelationshipType,
    deleteRelationshipType,
    validateCompatibility,
    clearError
  }
})
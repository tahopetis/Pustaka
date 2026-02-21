import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { eaApi } from '@/services/eaApi'
import type {
  EAEntity,
  EACreateRequest,
  EAUpdateRequest,
  EAFilter,
  PaginationMeta,
  ValidationError
} from '@/types/ea'

export const useEaStore = defineStore('ea', () => {
  // State
  const entities = ref<EAEntity[]>([])
  const currentEntity = ref<EAEntity | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const pagination = ref<PaginationMeta>({
    total: 0,
    page: 1,
    page_size: 25,
    total_pages: 0
  })
  const filter = ref<EAFilter>({})

  // Getters
  const entitiesByDomain = computed(() => {
    return (domain: string) => {
      return entities.value.filter(e => e.domain === domain)
    }
  })

  const entityById = computed(() => {
    return (id: string) => {
      return entities.value.find(e => e.id === id)
    }
  })

  const currentFilter = computed(() => filter.value)

  // Actions
  const fetchEntities = async (newFilter?: EAFilter) => {
    try {
      loading.value = true
      error.value = null

      if (newFilter) {
        filter.value = { ...filter.value, ...newFilter }
      }

      const response = await eaApi.listEntities(filter.value)
      const data = response.data

      entities.value = data.entities
      pagination.value = data.meta

      return { entities: data.entities, meta: data.meta }
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || 'Failed to load entities'
      error.value = errorMessage
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchEntity = async (id: string) => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.getEntity(id)
      currentEntity.value = response.data
      return currentEntity.value
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || 'Failed to load entity'
      error.value = errorMessage
      throw err
    } finally {
      loading.value = false
    }
  }

  const createEntity = async (data: EACreateRequest) => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.createEntity(data)
      const newEntity = response.data

      // Add to local state
      entities.value.unshift(newEntity)

      return newEntity
    } catch (err: any) {
      // Handle validation errors
      if (err.response?.status === 422) {
        const errorDetails = err.response.data?.error?.details
        if (errorDetails) {
          const validationErrors: ValidationError[] = Object.entries(errorDetails).map(
            ([field, messages]) => ({
              field,
              message: Array.isArray(messages) ? messages.join('. ') : String(messages),
              code: 'validation_error'
            })
          )
          error.value = `Validation errors: ${validationErrors.map(e => e.message).join('. ')}`
        } else {
          error.value = 'Validation failed. Please check your input.'
        }
      } else {
        const errorMessage = err.response?.data?.error?.message || err.message || 'Failed to create entity'
        error.value = errorMessage
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateEntity = async (id: string, data: EAUpdateRequest) => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.updateEntity(id, data)
      const updatedEntity = response.data

      // Update in local state
      const index = entities.value.findIndex(e => e.id === id)
      if (index !== -1) {
        entities.value[index] = updatedEntity
      }

      if (currentEntity.value?.id === id) {
        currentEntity.value = updatedEntity
      }

      return updatedEntity
    } catch (err: any) {
      // Handle validation errors
      if (err.response?.status === 422) {
        const errorDetails = err.response.data?.error?.details
        if (errorDetails) {
          const validationErrors: ValidationError[] = Object.entries(errorDetails).map(
            ([field, messages]) => ({
              field,
              message: Array.isArray(messages) ? messages.join('. ') : String(messages),
              code: 'validation_error'
            })
          )
          error.value = `Validation errors: ${validationErrors.map(e => e.message).join('. ')}`
        } else {
          error.value = 'Validation failed. Please check your input.'
        }
      } else {
        const errorMessage = err.response?.data?.error?.message || err.message || 'Failed to update entity'
        error.value = errorMessage
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteEntity = async (id: string, force: boolean = false) => {
    try {
      loading.value = true
      error.value = null

      await eaApi.deleteEntity(id, force)

      // Remove from local state
      entities.value = entities.value.filter(e => e.id !== id)

      if (currentEntity.value?.id === id) {
        currentEntity.value = null
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || 'Failed to delete entity'
      error.value = errorMessage
      throw err
    } finally {
      loading.value = false
    }
  }

  const setFilter = (newFilter: Partial<EAFilter>) => {
    filter.value = { ...filter.value, ...newFilter }
  }

  const clearFilter = () => {
    filter.value = {}
  }

  const clearError = () => {
    error.value = null
  }

  const clearCurrentEntity = () => {
    currentEntity.value = null
  }

  return {
    // State
    entities,
    currentEntity,
    loading,
    error,
    pagination,
    filter,

    // Getters
    entitiesByDomain,
    entityById,
    currentFilter,

    // Actions
    fetchEntities,
    fetchEntity,
    createEntity,
    updateEntity,
    deleteEntity,
    setFilter,
    clearFilter,
    clearError,
    clearCurrentEntity
  }
})

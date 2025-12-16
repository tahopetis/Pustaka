import { defineStore } from 'pinia'
import { ref } from 'vue'
import { lifecycleStatusAPI } from '@/services/api'

export interface LifecycleStatus {
  id: string
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  sort_order: number
  is_active: boolean
  is_system: boolean
  created_at: string
  updated_at: string
  created_by: string
}

export interface LifecycleStatusListResponse {
  lifecycle_statuses: LifecycleStatus[]
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface LifecycleStatusUsage {
  lifecycle_status: LifecycleStatus
  usage_count: number
}

export interface LifecycleStatusUsageResponse {
  total_cis: number
  cis_with_status: number
  cis_without_status: number
  status_usage: LifecycleStatusUsage[]
  status_distribution: Record<string, number>
}

export interface CIStatusDistribution {
  status_name: string
  display_name: string
  color: string
  icon: string
  count: number
  percentage: number
}

export interface CreateLifecycleStatusRequest {
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
}

export interface UpdateLifecycleStatusRequest {
  display_name?: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
  is_active?: boolean
}

export const useLifecycleStatusStore = defineStore('lifecycleStatus', () => {
  // State
  const lifecycleStatuses = ref<LifecycleStatus[]>([])
  const activeLifecycleStatuses = ref<LifecycleStatus[]>([])
  const currentLifecycleStatus = ref<LifecycleStatus | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  const listLifecycleStatuses = async (
    page: number = 1,
    limit: number = 20,
    filters?: {
      search?: string
      is_active?: boolean
      is_system?: boolean
      sort?: string
      order?: string
    }
  ): Promise<LifecycleStatusListResponse> => {
    try {
      loading.value = true
      error.value = null

      const params = new URLSearchParams({
        page: page.toString(),
        limit: limit.toString()
      })

      if (filters) {
        if (filters.search) {
          params.append('search', filters.search)
        }
        if (filters.is_active !== undefined) {
          params.append('is_active', filters.is_active.toString())
        }
        if (filters.is_system !== undefined) {
          params.append('is_system', filters.is_system.toString())
        }
        if (filters.sort) {
          params.append('sort', filters.sort)
        }
        if (filters.order) {
          params.append('order', filters.order)
        }
      }

      const response = await lifecycleStatusAPI.list(params)
      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to load lifecycle statuses'
      throw err
    } finally {
      loading.value = false
    }
  }

  const getLifecycleStatus = async (id: string): Promise<LifecycleStatus> => {
    try {
      loading.value = true
      error.value = null

      const response = await lifecycleStatusAPI.get(id)
      currentLifecycleStatus.value = response.data
      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to load lifecycle status'
      throw err
    } finally {
      loading.value = false
    }
  }

  const getActiveLifecycleStatuses = async (): Promise<LifecycleStatus[]> => {
    try {
      loading.value = true
      error.value = null

      const response = await lifecycleStatusAPI.getActive()
      activeLifecycleStatuses.value = response.data
      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to load active lifecycle statuses'
      throw err
    } finally {
      loading.value = false
    }
  }

  const createLifecycleStatus = async (data: CreateLifecycleStatusRequest): Promise<LifecycleStatus> => {
    try {
      loading.value = true
      error.value = null

      console.log('Creating Lifecycle Status with data:', data)

      const response = await lifecycleStatusAPI.create(data)
      const newLifecycleStatus = response.data

      console.log('Lifecycle Status created successfully:', newLifecycleStatus)

      // Add to local state
      lifecycleStatuses.value.push(newLifecycleStatus)

      // Update active statuses if it's active
      if (newLifecycleStatus.is_active) {
        activeLifecycleStatuses.value.push(newLifecycleStatus)
        // Sort by sort_order
        activeLifecycleStatuses.value.sort((a, b) => a.sort_order - b.sort_order)
      }

      return newLifecycleStatus
    } catch (err: any) {
      console.error('Failed to create Lifecycle Status:', err)

      // Enhanced error handling
      if (err.response?.status === 422) {
        const errorDetails = err.response.data?.error?.details
        if (errorDetails) {
          const errorMessages = Object.values(errorDetails).flat()
          error.value = `Validation errors: ${errorMessages.join('. ')}`
        } else {
          error.value = 'Validation failed. Please check your input.'
        }
      } else if (err.response?.status === 409) {
        error.value = 'A lifecycle status with this name already exists'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to create lifecycle status'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const updateLifecycleStatus = async (id: string, data: UpdateLifecycleStatusRequest): Promise<LifecycleStatus> => {
    try {
      loading.value = true
      error.value = null

      console.log('Updating Lifecycle Status:', { id, data })

      // Filter out undefined values to avoid sending unnecessary data
      const updateData: any = {}
      if (data.display_name !== undefined) {
        updateData.display_name = data.display_name
      }
      if (data.description !== undefined) {
        updateData.description = data.description
      }
      if (data.color !== undefined) {
        updateData.color = data.color
      }
      if (data.icon !== undefined) {
        updateData.icon = data.icon
      }
      if (data.sort_order !== undefined) {
        updateData.sort_order = data.sort_order
      }
      if (data.is_active !== undefined) {
        updateData.is_active = data.is_active
      }

      console.log('Sending update request:', { id, data: updateData })

      const response = await lifecycleStatusAPI.update(id, updateData)
      const updatedLifecycleStatus = response.data

      console.log('Lifecycle Status updated successfully:', updatedLifecycleStatus)

      // Update local state
      const index = lifecycleStatuses.value.findIndex(ls => ls.id === id)
      if (index !== -1) {
        lifecycleStatuses.value[index] = updatedLifecycleStatus
      }

      // Update active statuses
      const activeIndex = activeLifecycleStatuses.value.findIndex(ls => ls.id === id)
      if (activeIndex !== -1) {
        if (updatedLifecycleStatus.is_active) {
          activeLifecycleStatuses.value[activeIndex] = updatedLifecycleStatus
          // Sort by sort_order
          activeLifecycleStatuses.value.sort((a, b) => a.sort_order - b.sort_order)
        } else {
          // Remove from active if deactivated
          activeLifecycleStatuses.value.splice(activeIndex, 1)
        }
      } else if (updatedLifecycleStatus.is_active) {
        // Add to active if newly activated
        activeLifecycleStatuses.value.push(updatedLifecycleStatus)
        activeLifecycleStatuses.value.sort((a, b) => a.sort_order - b.sort_order)
      }

      if (currentLifecycleStatus.value?.id === id) {
        currentLifecycleStatus.value = updatedLifecycleStatus
      }

      return updatedLifecycleStatus
    } catch (err: any) {
      console.error('Failed to update Lifecycle Status:', err)

      // Enhanced error handling
      if (err.response?.status === 422) {
        const errorDetails = err.response.data?.error?.details
        if (errorDetails) {
          const errorMessages = Object.values(errorDetails).flat()
          error.value = `Validation errors: ${errorMessages.join('. ')}`
        } else {
          error.value = 'Validation failed. Please check your input.'
        }
      } else if (err.response?.status === 404) {
        error.value = 'Lifecycle status not found'
      } else if (err.response?.status === 403) {
        error.value = 'Cannot modify system lifecycle statuses'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to update lifecycle status'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteLifecycleStatus = async (id: string): Promise<void> => {
    try {
      loading.value = true
      error.value = null

      await lifecycleStatusAPI.delete(id)

      // Remove from local state
      lifecycleStatuses.value = lifecycleStatuses.value.filter(ls => ls.id !== id)
      activeLifecycleStatuses.value = activeLifecycleStatuses.value.filter(ls => ls.id !== id)

      if (currentLifecycleStatus.value?.id === id) {
        currentLifecycleStatus.value = null
      }
    } catch (err: any) {
      console.error('Failed to delete Lifecycle Status:', err)

      // Enhanced error handling
      if (err.response?.status === 404) {
        error.value = 'Lifecycle status not found'
      } else if (err.response?.status === 403) {
        error.value = 'Cannot delete system lifecycle statuses'
      } else if (err.response?.status === 409) {
        error.value = 'Cannot delete lifecycle status that is in use'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to delete lifecycle status'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const getLifecycleStatusUsage = async (): Promise<LifecycleStatusUsageResponse> => {
    try {
      const response = await lifecycleStatusAPI.getUsage()
      return response.data
    } catch (err: any) {
      console.error('Failed to load lifecycle status usage statistics:', err)
      error.value = err.message || 'Failed to load lifecycle status usage statistics'
      throw err
    }
  }

  const getCIStatusDistribution = async (): Promise<CIStatusDistribution[]> => {
    try {
      const response = await lifecycleStatusAPI.getDistribution()
      return response.data
    } catch (err: any) {
      console.error('Failed to load CI status distribution:', err)
      error.value = err.message || 'Failed to load CI status distribution'
      throw err
    }
  }

  const clearError = () => {
    error.value = null
  }

  const clearCurrentLifecycleStatus = () => {
    currentLifecycleStatus.value = null
  }

  return {
    // State
    lifecycleStatuses,
    activeLifecycleStatuses,
    currentLifecycleStatus,
    loading,
    error,

    // Actions
    listLifecycleStatuses,
    getLifecycleStatus,
    getActiveLifecycleStatuses,
    createLifecycleStatus,
    updateLifecycleStatus,
    deleteLifecycleStatus,
    getLifecycleStatusUsage,
    getCIStatusDistribution,
    clearError,
    clearCurrentLifecycleStatus
  }
})
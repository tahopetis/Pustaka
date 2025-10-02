import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ciTypeAPI } from '@/services/api'

export interface AttributeDefinition {
  name: string
  type: string
  description: string
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

export interface CIType {
  id: string
  name: string
  description?: string
  required_attributes: AttributeDefinition[]
  optional_attributes: AttributeDefinition[]
  created_by: string
  created_at: string
  updated_at: string
}

export interface CITypeListResponse {
  ci_types: CIType[]
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface CITypeUsage {
  type: string
  count: number
}

export interface CreateCITypeRequest {
  name: string
  description?: string
  required_attributes: AttributeDefinition[]
  optional_attributes: AttributeDefinition[]
}

export interface UpdateCITypeRequest {
  description?: string
  required_attributes?: AttributeDefinition[]
  optional_attributes?: AttributeDefinition[]
}

export const useCITypesStore = defineStore('ciTypes', () => {
  // State
  const ciTypes = ref<CIType[]>([])
  const currentCIType = ref<CIType | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  const listCITypes = async (page: number = 1, limit: number = 20, search?: string): Promise<CITypeListResponse> => {
    try {
      loading.value = true
      error.value = null

      const params = new URLSearchParams({
        page: page.toString(),
        limit: limit.toString()
      })

      if (search) {
        params.append('search', search)
      }

      const response = await ciTypeAPI.list(params)
      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to load CI types'
      throw err
    } finally {
      loading.value = false
    }
  }

  const getCIType = async (id: string): Promise<CIType> => {
    try {
      loading.value = true
      error.value = null

      const response = await ciTypeAPI.get(id)
      currentCIType.value = response.data
      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to load CI type'
      throw err
    } finally {
      loading.value = false
    }
  }

  const createCIType = async (data: CreateCITypeRequest): Promise<CIType> => {
    try {
      loading.value = true
      error.value = null

      console.log('Creating CI Type with data:', data)

      const response = await ciTypeAPI.create(data)
      const newCIType = response.data

      console.log('CI Type created successfully:', newCIType)

      // Add to local state
      ciTypes.value.push(newCIType)

      return newCIType
    } catch (err: any) {
      console.error('Failed to create CI Type:', err)

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
        error.value = 'A CI type with this name already exists'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to create CI type'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const updateCIType = async (id: string, data: UpdateCITypeRequest): Promise<CIType> => {
    try {
      loading.value = true
      error.value = null

      console.log('Updating CI Type:', { id, data })

      // Filter out undefined values to avoid sending unnecessary data
      const updateData: any = {}
      if (data.description !== undefined) {
        updateData.description = data.description
      }
      if (data.required_attributes !== undefined) {
        updateData.required_attributes = data.required_attributes
      }
      if (data.optional_attributes !== undefined) {
        updateData.optional_attributes = data.optional_attributes
      }

      console.log('Sending update request:', { id, data: updateData })

      const response = await ciTypeAPI.update(id, updateData)
      const updatedCIType = response.data

      console.log('CI Type updated successfully:', updatedCIType)

      // Update local state
      const index = ciTypes.value.findIndex(ct => ct.id === id)
      if (index !== -1) {
        ciTypes.value[index] = updatedCIType
      }

      if (currentCIType.value?.id === id) {
        currentCIType.value = updatedCIType
      }

      return updatedCIType
    } catch (err: any) {
      console.error('Failed to update CI Type:', err)

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
        error.value = 'CI Type not found'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to update CI type'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteCIType = async (id: string): Promise<void> => {
    try {
      loading.value = true
      error.value = null

      await ciTypeAPI.delete(id)

      // Remove from local state
      ciTypes.value = ciTypes.value.filter(ct => ct.id !== id)

      if (currentCIType.value?.id === id) {
        currentCIType.value = null
      }
    } catch (err: any) {
      console.error('Failed to delete CI Type:', err)

      // Enhanced error handling
      if (err.response?.status === 404) {
        error.value = 'CI Type not found'
      } else if (err.response?.status === 409) {
        error.value = 'Cannot delete CI Type that is in use'
      } else if (err.response?.status >= 500) {
        const serverMessage = err.response.data?.error?.message || 'Server error occurred'
        error.value = `${serverMessage}. Please try again later.`
      } else {
        error.value = err.message || 'Failed to delete CI type'
      }

      throw err
    } finally {
      loading.value = false
    }
  }

  const getCITypesByUsage = async (): Promise<CITypeUsage[]> => {
    try {
      const response = await ciTypeAPI.getUsage()
      return response.data
    } catch (err: any) {
      console.error('Failed to load CI type usage statistics:', err)
      error.value = err.message || 'Failed to load CI type usage statistics'
      throw err
    }
  }

  const clearError = () => {
    error.value = null
  }

  const clearCurrentCIType = () => {
    currentCIType.value = null
  }

  return {
    // State
    ciTypes,
    currentCIType,
    loading,
    error,

    // Actions
    listCITypes,
    getCIType,
    createCIType,
    updateCIType,
    deleteCIType,
    getCITypesByUsage,
    clearError,
    clearCurrentCIType
  }
})
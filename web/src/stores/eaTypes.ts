import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { eaApi } from '@/services/eaApi'
import type { CITypeDefinition } from '@/types/ea'

export interface EATeam {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  created_by: string
}

export const useEaTypesStore = defineStore('eaTypes', () => {
  // State
  const ciTypes = ref<CITypeDefinition[]>([])
  const teams = ref<EATeam[]>([])
  const currentCIType = ref<CITypeDefinition | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const getCiTypeByName = computed(() => {
    return (name: string) => ciTypes.value.find(ct => ct.name === name)
  })

  const getCiTypesByDomain = computed(() => {
    return (domain: string) => {
      const domainPrefix = `EA.${domain}`
      return ciTypes.value.filter(ct => ct.name.startsWith(domainPrefix))
    }
  })

  const getTeamByName = computed(() => {
    return (name: string) => teams.value.find(t => t.name === name)
  })

  // Actions
  const fetchCiTypes = async (page: number = 1, limit: number = 100) => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.listCiTypes({ page, limit })
      // Assuming response format matches CI types pattern
      const data = response.data as any

      if (data.ci_types) {
        ciTypes.value = data.ci_types
      } else if (Array.isArray(data)) {
        ciTypes.value = data
      }

      return ciTypes.value
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Failed to load EA CI types'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchCiTypeByName = async (name: string) => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.getCiType(name)
      currentCIType.value = response.data
      return currentCIType.value
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Failed to load EA CI type'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchTeams = async () => {
    try {
      loading.value = true
      error.value = null

      const response = await eaApi.listTeams()
      const data = response.data as any

      if (data.data) {
        teams.value = data.data
      } else if (Array.isArray(data)) {
        teams.value = data
      }

      return teams.value
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Failed to load EA teams'
      throw err
    } finally {
      loading.value = false
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
    teams,
    currentCIType,
    loading,
    error,

    // Getters
    getCiTypeByName,
    getCiTypesByDomain,
    getTeamByName,

    // Actions
    fetchCiTypes,
    fetchCiTypeByName,
    fetchTeams,
    clearError,
    clearCurrentCIType
  }
})

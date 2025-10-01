<template>
  <div class="px-4 py-6 sm:px-0">
    <div class="max-w-7xl mx-auto">
      <!-- Page header -->
      <div class="mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Configuration Items</h1>
          <p class="mt-2 text-gray-600">Manage your configuration items and their relationships</p>
        </div>
        <router-link
          v-if="hasPermission('ci:create')"
          to="/ci/new"
          class="btn btn-primary"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
          New Configuration Item
        </router-link>
      </div>

      <!-- Advanced Search -->
      <AdvancedSearch
        :ci-types="ciTypes"
        :is-loading="loading"
        :total-results="response?.total"
        @search="onSearch"
        @clear="onClearSearch"
      />

      <!-- CI List -->
      <div class="bg-white shadow rounded-lg">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">
            Configuration Items ({{ response?.total || 0 }})
          </h3>
        </div>
        <div class="card-body p-0">
          <!-- Loading state -->
          <div v-if="loading" class="text-center py-12">
            <div class="spinner w-8 h-8 mx-auto mb-4"></div>
            <p class="text-gray-500">Loading configuration items...</p>
          </div>

          <!-- Empty state -->
          <div v-else-if="!loading && (!response?.cis || response.cis.length === 0)" class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No configuration items</h3>
            <p class="mt-1 text-sm text-gray-500">
              Get started by creating your first configuration item.
            </p>
            <div class="mt-6" v-if="hasPermission('ci:create')">
              <router-link to="/ci/new" class="btn btn-primary">
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                </svg>
                New Configuration Item
              </router-link>
            </div>
          </div>

          <!-- CI Table -->
          <div v-else class="overflow-x-auto">
            <table class="table">
              <thead class="table-header">
                <tr>
                  <th class="table-header-cell">Name</th>
                  <th class="table-header-cell">Type</th>
                  <th class="table-header-cell">Tags</th>
                  <th class="table-header-cell">Created</th>
                  <th class="table-header-cell">Updated</th>
                  <th class="table-header-cell">Actions</th>
                </tr>
              </thead>
              <tbody class="table-body">
                <tr v-for="ci in response?.cis" :key="ci.id">
                  <td class="table-cell">
                    <router-link :to="`/ci/${ci.id}`" class="text-blue-600 hover:text-blue-900 font-medium">
                      {{ ci.name }}
                    </router-link>
                  </td>
                  <td class="table-cell">
                    <span class="badge badge-info">{{ ci.ci_type }}</span>
                  </td>
                  <td class="table-cell">
                    <div class="flex flex-wrap gap-1">
                      <span v-for="tag in ci.tags" :key="tag" class="badge badge-success">
                        {{ tag }}
                      </span>
                    </div>
                  </td>
                  <td class="table-cell">
                    {{ formatDate(ci.created_at) }}
                  </td>
                  <td class="table-cell">
                    {{ formatDate(ci.updated_at) }}
                  </td>
                  <td class="table-cell">
                    <div class="flex space-x-2">
                      <router-link
                        :to="`/ci/${ci.id}`"
                        class="text-blue-600 hover:text-blue-900"
                        title="View"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                        </svg>
                      </router-link>
                      <router-link
                        v-if="hasPermission('ci:update')"
                        :to="`/ci/${ci.id}/edit`"
                        class="text-indigo-600 hover:text-indigo-900"
                        title="Edit"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                        </svg>
                      </router-link>
                      <button
                        v-if="hasPermission('ci:delete')"
                        @click="confirmDelete(ci)"
                        class="text-red-600 hover:text-red-900"
                        title="Delete"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div v-if="response && response.total > response.limit" class="px-6 py-3 bg-gray-50 border-t border-gray-200">
            <div class="flex items-center justify-between">
              <div class="text-sm text-gray-700">
                Showing {{ ((response.page - 1) * response.limit) + 1 }} to {{ Math.min(response.page * response.limit, response.total) }} of {{ response.total }} results
              </div>
              <div class="flex space-x-2">
                <button
                  @click="goToPage(response.page - 1)"
                  :disabled="response.page <= 1"
                  class="btn btn-outline"
                  :class="{ 'opacity-50 cursor-not-allowed': response.page <= 1 }"
                >
                  Previous
                </button>
                <button
                  @click="goToPage(response.page + 1)"
                  :disabled="response.page >= response.total_pages"
                  class="btn btn-outline"
                  :class="{ 'opacity-50 cursor-not-allowed': response.page >= response.total_pages }"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { ciAPI, ciTypeAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'
import AdvancedSearch from '@/components/ci/AdvancedSearch.vue'
import type { CI, CIType, CIListResponse } from '@/types/ci'

const authStore = useAuthStore()

const loading = ref(false)
const response = ref<CIListResponse | null>(null)
const ciTypes = ref<CIType[]>([])

const filters = reactive({
  search: '',
  ci_type: '',
  sort: 'name',
  order: 'asc',
  tags: [] as string[],
  attributes: {} as Record<string, any>
})

const pagination = reactive({
  page: 1,
  limit: 20,
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const onSearch = (searchFilters: any) => {
  // Update filters with search data
  Object.assign(filters, searchFilters)
  pagination.page = 1
  loadCIs()
}

const onClearSearch = () => {
  // Reset filters to defaults
  filters.search = ''
  filters.ci_type = ''
  filters.sort = 'name'
  filters.order = 'asc'
  filters.tags = []
  filters.attributes = {}
  pagination.page = 1
  loadCIs()
}

const loadCIs = async () => {
  loading.value = true
  try {
    const params = {
      ...filters,
      ...pagination,
    }

    // Clean up empty values from attributes
    if (params.attributes) {
      const cleanedAttributes: Record<string, any> = {}
      Object.keys(params.attributes).forEach(key => {
        const value = params.attributes[key]
        if (value !== undefined && value !== null && value !== '') {
          cleanedAttributes[key] = value
        }
      })
      params.attributes = cleanedAttributes
    }

    response.value = await ciAPI.list(params)
  } catch (error) {
    console.error('Failed to load CIs:', error)
    showErrorToast('Failed to load configuration items')
  } finally {
    loading.value = false
  }
}

const loadCITypes = async () => {
  try {
    const typesResponse = await ciTypeAPI.list()
    ciTypes.value = typesResponse.data.ci_types || []
  } catch (error) {
    console.error('Failed to load CI types:', error)
  }
}

const goToPage = (page: number) => {
  pagination.page = page
  loadCIs()
}

const confirmDelete = async (ci: CI) => {
  if (confirm(`Are you sure you want to delete "${ci.name}"? This action cannot be undone.`)) {
    try {
      await ciAPI.delete(ci.id)
      showSuccessToast('Configuration item deleted successfully')
      await loadCIs()
    } catch (error: any) {
      console.error('Failed to delete CI:', error)
      const message = error.response?.data?.error || 'Failed to delete configuration item'
      showErrorToast(message)
    }
  }
}

onMounted(() => {
  loadCIs()
  loadCITypes()
})
</script>
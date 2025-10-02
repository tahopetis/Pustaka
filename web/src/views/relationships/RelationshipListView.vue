<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">Relationships</h1>
        <p class="page-subtitle">Manage relationships between configuration items</p>
      </div>
      <router-link
        v-if="hasPermission('relationship:create')"
        to="/relationships/new"
        class="btn btn-primary"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        Add Relationship
      </router-link>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="md:col-span-2">
          <label class="form-label">Search</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search by CI name or type..."
            class="form-input"
            @input="debouncedSearch"
          />
        </div>
        <div>
          <label class="form-label">Relationship Type</label>
          <select v-model="filters.relationship_type" class="form-input" @change="loadRelationships">
            <option value="">All Types</option>
            <option value="depends_on">Depends On</option>
            <option value="connected_to">Connected To</option>
            <option value="runs_on">Runs On</option>
            <option value="contains">Contains</option>
          </select>
        </div>
        <div class="flex items-end">
          <button @click="loadRelationships" :disabled="loading" class="btn btn-primary w-full">
            <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
            Search
          </button>
        </div>
      </div>
    </div>

    <!-- Relationships List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Relationships ({{ response?.total || 0 }})
        </h3>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">Loading relationships...</p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && (!response?.relationships || response.relationships.length === 0)" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No relationships found</h3>
          <p class="mt-1 text-sm text-gray-500">
            Get started by creating a relationship between configuration items.
          </p>
          <div class="mt-6" v-if="hasPermission('relationship:create')">
            <router-link to="/relationships/new" class="btn btn-primary">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              Add Relationship
            </router-link>
          </div>
        </div>

        <!-- Relationships Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell">Source CI</th>
                <th class="table-header-cell">Relationship</th>
                <th class="table-header-cell">Target CI</th>
                <th class="table-header-cell">Created</th>
                <th class="table-header-cell">Updated</th>
                <th class="table-header-cell">Actions</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="relationship in response?.relationships" :key="relationship.id">
                <td class="table-cell">
                  <div>
                    <router-link
                      :to="`/ci/${relationship.source_id}`"
                      class="text-blue-600 hover:text-blue-900 font-medium"
                    >
                      {{ getSourceCIName(relationship) }}
                    </router-link>
                    <div class="text-sm text-gray-500">{{ getSourceCIType(relationship) }}</div>
                  </div>
                </td>
                <td class="table-cell">
                  <span class="badge badge-info">{{ relationship.relationship_type }}</span>
                </td>
                <td class="table-cell">
                  <div>
                    <router-link
                      :to="`/ci/${relationship.target_id}`"
                      class="text-blue-600 hover:text-blue-900 font-medium"
                    >
                      {{ getTargetCIName(relationship) }}
                    </router-link>
                    <div class="text-sm text-gray-500">{{ getTargetCIType(relationship) }}</div>
                  </div>
                </td>
                <td class="table-cell">
                  {{ formatDate(relationship.created_at) }}
                </td>
                <td class="table-cell">
                  {{ formatDate(relationship.updated_at) }}
                </td>
                <td class="table-cell">
                  <div class="flex space-x-2">
                    <router-link
                      :to="`/relationships/${relationship.id}`"
                      class="text-blue-600 hover:text-blue-900"
                      title="View"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                      </svg>
                    </router-link>
                    <router-link
                      v-if="hasPermission('relationship:update')"
                      :to="`/relationships/${relationship.id}/edit`"
                      class="text-indigo-600 hover:text-indigo-900"
                      title="Edit"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                      </svg>
                    </router-link>
                    <button
                      v-if="hasPermission('relationship:delete')"
                      @click="confirmDelete(relationship)"
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
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { relationshipAPI, ciAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface Relationship {
  id: string
  source_id: string
  target_id: string
  relationship_type: string
  attributes?: Record<string, any>
  created_at: string
  updated_at: string
}

interface RelationshipListResponse {
  relationships: Relationship[]
  total: number
  page: number
  limit: number
  total_pages: number
}

interface CI {
  id: string
  name: string
  ci_type: string
}

const authStore = useAuthStore()

const loading = ref(false)
const response = ref<RelationshipListResponse | null>(null)
const ciCache = ref<Record<string, CI>>({})

const filters = reactive({
  search: '',
  relationship_type: '',
  sort: 'created_at',
  order: 'desc',
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

const debouncedSearch = debounce(() => {
  pagination.page = 1
  loadRelationships()
}, 500)

function debounce(func: Function, wait: number) {
  let timeout: NodeJS.Timeout
  return function executedFunction(...args: any[]) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

const loadRelationships = async () => {
  if (!hasPermission('relationship:read')) return

  loading.value = true
  try {
    const params = {
      ...filters,
      ...pagination,
    }

    // Clean up empty values
    Object.keys(params).forEach(key => {
      if (params[key] === '') {
        delete params[key]
      }
    })

    response.value = await relationshipAPI.list(params)

    // Load CI details for display
    await loadCIDetails(response.value.relationships)
  } catch (error) {
    console.error('Failed to load relationships:', error)
    showErrorToast('Failed to load relationships')
  } finally {
    loading.value = false
  }
}

const loadCIDetails = async (relationships: Relationship[]) => {
  const ciIds = new Set<string>()

  relationships.forEach(rel => {
    ciIds.add(rel.source_id)
    ciIds.add(rel.target_id)
  })

  // Load details for CIs we don't have cached
  const uncachedIds = Array.from(ciIds).filter(id => !ciCache.value[id])

  for (const id of uncachedIds) {
    try {
      const ciResponse = await ciAPI.get(id)
      ciCache.value[id] = ciResponse.data
    } catch (error) {
      console.error(`Failed to load CI details for ${id}:`, error)
    }
  }
}

const getSourceCIName = (relationship: Relationship) => {
  return ciCache.value[relationship.source_id]?.name || `CI: ${relationship.source_id.substring(0, 8)}...`
}

const getSourceCIType = (relationship: Relationship) => {
  return ciCache.value[relationship.source_id]?.ci_type || 'Unknown'
}

const getTargetCIName = (relationship: Relationship) => {
  return ciCache.value[relationship.target_id]?.name || `CI: ${relationship.target_id.substring(0, 8)}...`
}

const getTargetCIType = (relationship: Relationship) => {
  return ciCache.value[relationship.target_id]?.ci_type || 'Unknown'
}

const goToPage = (page: number) => {
  pagination.page = page
  loadRelationships()
}

const confirmDelete = async (relationship: Relationship) => {
  if (confirm(`Are you sure you want to delete this relationship? This action cannot be undone.`)) {
    try {
      await relationshipAPI.delete(relationship.id)
      showSuccessToast('Relationship deleted successfully')
      await loadRelationships()
    } catch (error: any) {
      console.error('Failed to delete relationship:', error)
      const message = error.response?.data?.error || 'Failed to delete relationship'
      showErrorToast(message)
    }
  }
}

onMounted(() => {
  loadRelationships()
})
</script>
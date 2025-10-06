<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">Relationship Types</h1>
        <p class="page-subtitle">Manage relationship types for configuration items</p>
      </div>
      <router-link
        v-if="hasPermission('relationship_type:create')"
        to="/relationship-types/new"
        class="btn btn-primary"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        Add Relationship Type
      </router-link>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="md:col-span-2 relative">
          <label class="form-label">Search</label>
          <div class="relative">
            <input
              v-model="filters.search"
              type="text"
              placeholder="Search by name, display name, or description..."
              class="form-input pr-10"
              @input="debouncedSearch"
            />
            <button
              v-if="filters.search"
              @click="clearSearch"
              class="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
              title="Clear search"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
        </div>
        <div>
          <label class="form-label">Category</label>
          <select v-model="filters.category" class="form-input" @change="loadRelationshipTypes">
            <option value="">All Categories</option>
            <option v-for="category in availableCategories" :key="category" :value="category">
              {{ category }}
            </option>
          </select>
        </div>
        <div>
          <label class="form-label">Status</label>
          <select v-model="filters.is_active" class="form-input" @change="loadRelationshipTypes">
            <option value="">All Types</option>
            <option value="true">Active</option>
            <option value="false">Inactive</option>
          </select>
        </div>
      </div>
      <div class="mt-4 flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <label class="flex items-center">
            <input
              v-model="filters.include_system"
              type="checkbox"
              class="form-checkbox"
              @change="loadRelationshipTypes"
            />
            <span class="ml-2 text-sm text-gray-700">Include system types</span>
          </label>
        </div>
        <div class="flex items-center space-x-2">
          <button @click="loadRelationshipTypes" :disabled="loading" class="btn btn-primary">
            <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
            Search
          </button>
          <button @click="resetFilters" class="btn btn-outline">
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div v-if="stats" class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Total Types</p>
            <p class="text-lg font-semibold text-gray-900">{{ stats.total_types }}</p>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-green-100 rounded-lg">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Active</p>
            <p class="text-lg font-semibold text-gray-900">{{ stats.active_types }}</p>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-gray-100 rounded-lg">
            <svg class="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">System</p>
            <p class="text-lg font-semibold text-gray-900">{{ stats.system_types }}</p>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Custom</p>
            <p class="text-lg font-semibold text-gray-900">{{ stats.custom_types }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Relationship Types List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Relationship Types ({{ relationshipTypeStore.total || 0 }})
        </h3>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">Loading relationship types...</p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && filteredRelationshipTypes.length === 0" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No relationship types found</h3>
          <p class="mt-1 text-sm text-gray-500">
            {{ hasSearchOrFilter ? 'Try adjusting your search or filters' : 'Get started by creating relationship types for your configuration items.' }}
          </p>
          <div class="mt-6">
            <button v-if="hasSearchOrFilter" @click="resetFilters" class="btn btn-outline">
              Clear Filters
            </button>
            <router-link v-else-if="hasPermission('relationship_type:create')" to="/relationship-types/new" class="btn btn-primary">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              Add Relationship Type
            </router-link>
          </div>
        </div>

        <!-- Relationship Types Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell">Name</th>
                <th class="table-header-cell">Category</th>
                <th class="table-header-cell">Forward Label</th>
                <th class="table-header-cell">Reverse Label</th>
                <th class="table-header-cell">Description</th>
                <th class="table-header-cell">Status</th>
                <th class="table-header-cell">Created</th>
                <th class="table-header-cell">Actions</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="type in filteredRelationshipTypes" :key="type.id">
                <td class="table-cell">
                  <div class="flex items-center">
                    <div
                      v-if="type.color"
                      class="w-3 h-3 rounded-full mr-2"
                      :style="{ backgroundColor: type.color }"
                    ></div>
                    <div>
                      <span class="font-medium">{{ type.display_name || type.name }}</span>
                      <span v-if="type.display_name" class="text-xs text-gray-500 ml-1">({{ type.name }})</span>
                      <div class="flex items-center mt-1">
                        <span v-if="type.is_system" class="badge badge-gray text-xs mr-1">System</span>
                        <span v-if="type.bidirectional" class="badge badge-outline text-xs">Bidirectional</span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="table-cell">
                  <span v-if="type.category" class="badge badge-outline text-xs">{{ type.category }}</span>
                  <span v-else class="text-gray-400">-</span>
                </td>
                <td class="table-cell">
                  <span
                    v-if="type.color"
                    class="badge"
                    :style="{ backgroundColor: type.color + '20', color: type.color, borderColor: type.color }"
                  >
                    {{ type.forward_label }}
                  </span>
                  <span v-else class="badge badge-info">
                    {{ type.forward_label }}
                  </span>
                </td>
                <td class="table-cell">
                  <span
                    v-if="type.color"
                    class="badge"
                    :style="{ backgroundColor: type.color + '20', color: type.color, borderColor: type.color }"
                  >
                    {{ type.reverse_label }}
                  </span>
                  <span v-else class="badge badge-secondary">
                    {{ type.reverse_label }}
                  </span>
                </td>
                <td class="table-cell">
                  <p class="text-sm text-gray-600 max-w-xs truncate" :title="type.description">
                    {{ type.description || '-' }}
                  </p>
                </td>
                <td class="table-cell">
                  <span :class="type.is_active ? 'badge badge-success' : 'badge badge-warning'">
                    {{ type.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="table-cell">
                  {{ formatDate(type.created_at) }}
                </td>
                <td class="table-cell">
                  <div class="flex space-x-2">
                    <router-link
                      :to="`/relationship-types/${type.id}`"
                      class="text-blue-600 hover:text-blue-900"
                      title="View"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                      </svg>
                    </router-link>
                    <router-link
                      v-if="hasPermission('relationship_type:update') && !type.is_system"
                      :to="`/relationship-types/${type.id}/edit`"
                      class="text-indigo-600 hover:text-indigo-900"
                      title="Edit"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                      </svg>
                    </router-link>
                    <button
                      v-if="hasPermission('relationship_type:delete') && !type.is_system"
                      @click="confirmDelete(type)"
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
        <div v-if="relationshipTypeStore.total > filters.limit" class="px-6 py-3 bg-gray-50 border-t border-gray-200">
          <div class="flex items-center justify-between">
            <div class="text-sm text-gray-700">
              Showing {{ ((filters.page - 1) * filters.limit) + 1 }} to {{ Math.min(filters.page * filters.limit, relationshipTypeStore.total) }} of {{ relationshipTypeStore.total }} results
            </div>
            <div class="flex space-x-2">
              <button
                @click="goToPage(filters.page - 1)"
                :disabled="filters.page <= 1"
                class="btn btn-outline"
                :class="{ 'opacity-50 cursor-not-allowed': filters.page <= 1 }"
              >
                Previous
              </button>
              <button
                @click="goToPage(filters.page + 1)"
                :disabled="filters.page >= totalPages"
                class="btn btn-outline"
                :class="{ 'opacity-50 cursor-not-allowed': filters.page >= totalPages }"
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRelationshipTypeStore } from '@/stores/relationshipTypes'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

const authStore = useAuthStore()
const relationshipTypeStore = useRelationshipTypeStore()

const loading = ref(false)
const availableCategories = ref<string[]>([])
const stats = ref<any>(null)

const filters = reactive({
  page: 1,
  limit: 20,
  search: '',
  is_active: '',
  category: '',
  include_system: true  // Fixed: Changed from false to true to include system types by default
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const relationshipTypes = computed(() => relationshipTypeStore.relationshipTypes)
const totalPages = computed(() => Math.ceil(relationshipTypeStore.total / filters.limit))

const hasSearchOrFilter = computed(() => {
  return filters.search || filters.is_active !== '' || filters.category || !filters.include_system
})

const filteredRelationshipTypes = computed(() => {
  let filtered = relationshipTypes.value

  // Filter out system types unless explicitly included
  if (!filters.include_system) {
    filtered = filtered.filter(type => !type.is_system)
  }

  return filtered
})

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const debouncedSearch = debounce(() => {
  filters.page = 1
  loadRelationshipTypes()
}, 500)

const clearSearch = () => {
  filters.search = ''
  filters.page = 1
  loadRelationshipTypes()
}

const resetFilters = () => {
  filters.search = ''
  filters.is_active = ''
  filters.category = ''
  filters.include_system = true  // Fixed: Changed to true to maintain consistency
  filters.page = 1
  loadRelationshipTypes()
}

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

const loadCategories = async () => {
  try {
    await relationshipTypeStore.loadCategories()
    availableCategories.value = relationshipTypeStore.categories
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

const loadStats = async () => {
  try {
    await relationshipTypeStore.loadStats()
    stats.value = relationshipTypeStore.stats
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
}

const loadRelationshipTypes = async () => {
  if (!hasPermission('relationship_type:read')) return

  loading.value = true
  try {
    const params = { ...filters }

    // Clean up empty values
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    // Don't pass include_system to API (we filter client-side)
    delete params.include_system

    await relationshipTypeStore.loadRelationshipTypes(params)
  } catch (error) {
    console.error('Failed to load relationship types:', error)
    showErrorToast('Failed to load relationship types')
  } finally {
    loading.value = false
  }
}

const goToPage = (page: number) => {
  filters.page = page
  loadRelationshipTypes()
}

const confirmDelete = async (type: any) => {
  if (confirm(`Are you sure you want to delete the relationship type "${type.name}"? This action cannot be undone and may affect existing relationships.`)) {
    try {
      await relationshipTypeStore.deleteRelationshipType(type.id)
      showSuccessToast('Relationship type deleted successfully')
      await loadRelationshipTypes()
      await loadStats()
    } catch (error: any) {
      console.error('Failed to delete relationship type:', error)
      const message = error.response?.data?.error || 'Failed to delete relationship type'
      showErrorToast(message)
    }
  }
}

onMounted(async () => {
  await Promise.all([
    loadCategories(),
    loadStats(),
    loadRelationshipTypes()
  ])
})
</script>
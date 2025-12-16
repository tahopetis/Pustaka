<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">Lifecycle Statuses</h1>
        <p class="page-subtitle">Manage lifecycle statuses for configuration items</p>
      </div>
      <router-link
        v-if="hasPermission('lifecycle_status:create')"
        to="/lifecycle-statuses/new"
        class="btn btn-primary"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        Add Lifecycle Status
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
          <label class="form-label">Status</label>
          <select v-model="filters.is_active" class="form-input" @change="loadLifecycleStatuses">
            <option value="">All Statuses</option>
            <option value="true">Active</option>
            <option value="false">Inactive</option>
          </select>
        </div>
        <div>
          <label class="form-label">Type</label>
          <select v-model="filters.is_system" class="form-input" @change="loadLifecycleStatuses">
            <option value="">All Types</option>
            <option value="false">Custom</option>
            <option value="true">System</option>
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
              @change="loadLifecycleStatuses"
            />
            <span class="ml-2 text-sm text-gray-700">Include system statuses</span>
          </label>
        </div>
        <div class="flex items-center space-x-2">
          <button @click="loadLifecycleStatuses" :disabled="loading" class="btn btn-primary">
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
    <div v-if="usageStats" class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Total Statuses</p>
            <p class="text-lg font-semibold text-gray-900">{{ usageStats.total_cis }}</p>
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
            <p class="text-sm font-medium text-gray-600">With Status</p>
            <p class="text-lg font-semibold text-gray-900">{{ usageStats.cis_with_status }}</p>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 rounded-lg">
            <svg class="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.314 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Without Status</p>
            <p class="text-lg font-semibold text-gray-900">{{ usageStats.cis_without_status }}</p>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">Active Types</p>
            <p class="text-lg font-semibold text-gray-900">{{ activeStatusCount }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Lifecycle Statuses List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Lifecycle Statuses ({{ totalItems }})
        </h3>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">Loading lifecycle statuses...</p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && filteredLifecycleStatuses.length === 0" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No lifecycle statuses found</h3>
          <p class="mt-1 text-sm text-gray-500">
            {{ hasSearchOrFilter ? 'Try adjusting your search or filters' : 'Get started by creating lifecycle statuses for your configuration items.' }}
          </p>
          <div class="mt-6">
            <button v-if="hasSearchOrFilter" @click="resetFilters" class="btn btn-outline">
              Clear Filters
            </button>
            <router-link v-else-if="hasPermission('lifecycle_status:create')" to="/lifecycle-statuses/new" class="btn btn-primary">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              Add Lifecycle Status
            </router-link>
          </div>
        </div>

        <!-- Lifecycle Statuses Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell">Status</th>
                <th class="table-header-cell">Description</th>
                <th class="table-header-cell">Sort Order</th>
                <th class="table-header-cell">Status</th>
                <th class="table-header-cell">Created</th>
                <th class="table-header-cell">Actions</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="status in filteredLifecycleStatuses" :key="status.id">
                <td class="table-cell">
                  <div class="flex items-center">
                    <div
                      v-if="status.color"
                      class="w-3 h-3 rounded-full mr-2"
                      :style="{ backgroundColor: status.color }"
                    ></div>
                    <div>
                      <span class="font-medium">{{ status.display_name || status.name }}</span>
                      <span v-if="status.display_name" class="text-xs text-gray-500 ml-1">({{ status.name }})</span>
                      <div class="flex items-center mt-1">
                        <span v-if="status.is_system" class="badge badge-gray text-xs mr-1">System</span>
                        <div v-if="status.icon" class="flex items-center text-xs text-gray-500">
                          <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="getIconPath(status.icon)"></path>
                          </svg>
                          {{ status.icon }}
                        </div>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="table-cell">
                  <p class="text-sm text-gray-600 max-w-xs truncate" :title="status.description">
                    {{ status.description || '-' }}
                  </p>
                </td>
                <td class="table-cell">
                  <span class="text-sm">{{ status.sort_order }}</span>
                </td>
                <td class="table-cell">
                  <span :class="status.is_active ? 'badge badge-success' : 'badge badge-warning'">
                    {{ status.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="table-cell">
                  {{ formatDate(status.created_at) }}
                </td>
                <td class="table-cell">
                  <div class="flex space-x-2">
                    <router-link
                      :to="`/lifecycle-statuses/${status.id}`"
                      class="text-blue-600 hover:text-blue-900"
                      title="View"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                      </svg>
                    </router-link>
                    <router-link
                      v-if="hasPermission('lifecycle_status:update') && !status.is_system"
                      :to="`/lifecycle-statuses/${status.id}/edit`"
                      class="text-indigo-600 hover:text-indigo-900"
                      title="Edit"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                      </svg>
                    </router-link>
                    <button
                      v-if="hasPermission('lifecycle_status:delete') && !status.is_system"
                      @click="confirmDelete(status)"
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
        <div v-if="totalItems > filters.limit" class="px-6 py-3 bg-gray-50 border-t border-gray-200">
          <div class="flex items-center justify-between">
            <div class="text-sm text-gray-700">
              Showing {{ ((filters.page - 1) * filters.limit) + 1 }} to {{ Math.min(filters.page * filters.limit, totalItems) }} of {{ totalItems }} results
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
import { useLifecycleStatusStore } from '@/stores/lifecycleStatus'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

const authStore = useAuthStore()
const lifecycleStatusStore = useLifecycleStatusStore()

const loading = ref(false)
const usageStats = ref<any>(null)
const lifecycleStatusList = ref<any>(null)

const filters = reactive({
  page: 1,
  limit: 20,
  search: '',
  is_active: '',
  is_system: '',
  include_system: true
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const totalPages = computed(() => {
  if (!lifecycleStatusList.value) return 1
  return Math.ceil(lifecycleStatusList.value.total / filters.limit)
})

const totalItems = computed(() => {
  return lifecycleStatusList.value?.total || 0
})

const hasSearchOrFilter = computed(() => {
  return filters.search || filters.is_active !== '' || filters.is_system !== '' || !filters.include_system
})

const filteredLifecycleStatuses = computed(() => {
  if (!lifecycleStatusList.value?.lifecycle_statuses) return []

  let filtered = lifecycleStatusList.value.lifecycle_statuses

  // Filter out system types unless explicitly included
  if (!filters.include_system) {
    filtered = filtered.filter(status => !status.is_system)
  }

  return filtered
})

const activeStatusCount = computed(() => {
  return filteredLifecycleStatuses.value.filter(status => status.is_active).length
})

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const getIconPath = (iconName: string) => {
  const iconPaths: Record<string, string> = {
    'calendar': 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
    'package': 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
    'archive': 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4',
    'clock': 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    'check-circle': 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    'wrench': 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z',
    'alert-triangle': 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.314 16.5c-.77.833.192 2.5 1.732 2.5z',
    'power-off': 'M18.364 5.636l-1.414 1.414M9.172 9.172L5.636 5.636M12 2v4m0 4v4m0 8a7 7 0 110-14 7 7 0 010 14z',
    'trash-2': 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
    'x-circle': 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
  }
  return iconPaths[iconName] || 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1'
}

const debouncedSearch = debounce(() => {
  filters.page = 1
  loadLifecycleStatuses()
}, 500)

const clearSearch = () => {
  filters.search = ''
  filters.page = 1
  loadLifecycleStatuses()
}

const resetFilters = () => {
  filters.search = ''
  filters.is_active = ''
  filters.is_system = ''
  filters.include_system = true
  filters.page = 1
  loadLifecycleStatuses()
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

const loadUsageStats = async () => {
  try {
    const stats = await lifecycleStatusStore.getLifecycleStatusUsage()
    usageStats.value = stats
  } catch (err) {
    console.error('Failed to load usage stats:', err)
  }
}

const loadLifecycleStatuses = async () => {
  if (!hasPermission('lifecycle_status:read')) return

  loading.value = true
  try {
    const params: any = {
      page: filters.page,
      limit: filters.limit
    }

    if (filters.search) {
      params.search = filters.search
    }
    if (filters.is_active !== '') {
      params.is_active = filters.is_active === 'true'
    }
    if (filters.is_system !== '') {
      params.is_system = filters.is_system === 'true'
    }

    const response = await lifecycleStatusStore.listLifecycleStatuses(filters.page, filters.limit, params)
    lifecycleStatusList.value = response
  } catch (error) {
    console.error('Failed to load lifecycle statuses:', error)
    showErrorToast('Failed to load lifecycle statuses')
  } finally {
    loading.value = false
  }
}

const goToPage = (page: number) => {
  filters.page = page
  loadLifecycleStatuses()
}

const confirmDelete = async (status: any) => {
  const usageCount = usageStats.value?.status_usage?.find((usage: any) => usage.lifecycle_status.id === status.id)?.usage_count || 0

  let message = `Are you sure you want to delete the lifecycle status "${status.display_name || status.name}"? This action cannot be undone.`

  if (usageCount > 0) {
    message += ` This status is currently used by ${usageCount} configuration item(s).`
  }

  if (confirm(message)) {
    try {
      await lifecycleStatusStore.deleteLifecycleStatus(status.id)
      showSuccessToast('Lifecycle status deleted successfully')
      await Promise.all([
        loadLifecycleStatuses(),
        loadUsageStats()
      ])
    } catch (error: any) {
      console.error('Failed to delete lifecycle status:', error)
      const message = error.response?.data?.error || 'Failed to delete lifecycle status'
      showErrorToast(message)
    }
  }
}

onMounted(async () => {
  await Promise.all([
    loadUsageStats(),
    loadLifecycleStatuses()
  ])
})
</script>
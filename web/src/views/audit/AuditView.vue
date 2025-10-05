<template>
  <div class="page-container page-content">
    <div class="page-header">
      <h1 class="page-title">Audit Logs</h1>
      <p class="page-subtitle">Track system activities and changes</p>
    </div>

    <!-- Filters -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Search</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search logs..."
            class="form-input"
            @input="debouncedLoadAuditLogs"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Entity</label>
          <select v-model="filters.entity" class="form-select" @change="loadAuditLogs">
            <option value="">All Entities</option>
            <option value="user">User</option>
            <option value="ci">Configuration Item</option>
            <option value="ci_type">CI Type</option>
            <option value="relationship">Relationship</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Action</label>
          <select v-model="filters.action" class="form-select" @change="loadAuditLogs">
            <option value="">All Actions</option>
            <option value="create">Create</option>
            <option value="update">Update</option>
            <option value="delete">Delete</option>
            <option value="login">Login</option>
            <option value="logout">Logout</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Performed By</label>
          <input
            v-model="filters.performed_by"
            type="text"
            placeholder="User ID or username..."
            class="form-input"
            @input="debouncedLoadAuditLogs"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">From Date</label>
          <input
            v-model="filters.from_date"
            type="date"
            class="form-input"
            @change="loadAuditLogs"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">To Date</label>
          <input
            v-model="filters.to_date"
            type="date"
            class="form-input"
            @change="loadAuditLogs"
          />
        </div>
        <div class="flex items-end">
          <button @click="resetFilters" class="btn btn-secondary w-full">
            Reset Filters
          </button>
        </div>
        <div class="flex items-end">
          <button @click="loadAuditLogs" class="btn btn-primary w-full">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Refresh
          </button>
        </div>
      </div>
    </div>

    <!-- Audit Logs List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Audit Logs ({{ response?.total || 0 }})
        </h3>
        <div class="flex items-center space-x-2">
          <span class="text-sm text-gray-500">
            Page {{ pagination.page }} of {{ response?.total_pages || 1 }}
          </span>
        </div>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading || authStore.isLoading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">
            {{ authStore.isLoading ? 'Initializing authentication...' : 'Loading audit logs...' }}
          </p>
        </div>

        <!-- Permission denied state -->
        <div v-else-if="!loading && authStore.isInitialized && !hasPermission('audit:read')" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">Access Denied</h3>
          <p class="mt-1 text-sm text-gray-500">
            You don't have permission to view audit logs.
          </p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && authStore.isInitialized && (!response?.audit_logs || response.audit_logs.length === 0)" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No audit logs found</h3>
          <p class="mt-1 text-sm text-gray-500">
            No audit logs found in the last 30 days. Try adjusting the date range or reset filters to see all logs.
          </p>
          <div class="mt-4 flex space-x-3 justify-center">
            <button @click="resetFilters" class="btn btn-secondary">
              View All Logs
            </button>
            <button @click="clearDateFilters" class="btn btn-outline">
              Clear Date Range
            </button>
          </div>
        </div>

        <!-- Audit Logs Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell cursor-pointer hover:bg-gray-50" @click="sortBy('timestamp')">
                  <div class="flex items-center">
                    Timestamp
                    <svg v-if="filters.sort === 'timestamp'" class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path v-if="filters.order === 'asc'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                      <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </th>
                <th class="table-header-cell">User</th>
                <th class="table-header-cell cursor-pointer hover:bg-gray-50" @click="sortBy('action')">
                  <div class="flex items-center">
                    Action
                    <svg v-if="filters.sort === 'action'" class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path v-if="filters.order === 'asc'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                      <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </th>
                <th class="table-header-cell">Entity</th>
                <th class="table-header-cell">Details</th>
                <th class="table-header-cell">IP Address</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="log in response?.audit_logs" :key="log.id" class="hover:bg-gray-50">
                <td class="table-cell">
                  <div>
                    <div class="text-sm text-gray-900">{{ formatDateTime(log.timestamp) }}</div>
                    <div class="text-xs text-gray-500">{{ formatRelativeTime(log.timestamp) }}</div>
                  </div>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ formatUser(log.performed_by) }}</div>
                </td>
                <td class="table-cell">
                  <span :class="getActionBadgeClass(log.action)">
                    {{ formatAction(log.action) }}
                  </span>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ log.entity_type || 'N/A' }}</div>
                  <div v-if="log.entity_id" class="text-xs text-gray-500">
                    ID: {{ log.entity_id.substring(0, 8) }}...
                  </div>
                </td>
                <td class="table-cell">
                  <div class="max-w-xs">
                    <div class="text-sm text-gray-900 truncate">
                      {{ getLogDetails(log) }}
                    </div>
                    <div v-if="hasMoreDetails(log)" @click="showLogDetails(log)" class="text-xs text-blue-600 hover:text-blue-900 cursor-pointer">
                      View details
                    </div>
                  </div>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ log.ip_address || 'N/A' }}</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Details Modal -->
        <div v-if="selectedLog" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50" @click="closeDetailsModal">
          <div class="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-md bg-white" @click.stop>
            <div class="flex justify-between items-center pb-3 border-b">
              <h3 class="text-lg font-medium text-gray-900">Audit Log Details</h3>
              <button @click="closeDetailsModal" class="text-gray-400 hover:text-gray-500">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div class="mt-4 space-y-4">
              <!-- Basic Information -->
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <h4 class="text-sm font-medium text-gray-500">Timestamp</h4>
                  <p class="text-sm text-gray-900">{{ formatDateTime(selectedLog.timestamp) }}</p>
                </div>
                <div>
                  <h4 class="text-sm font-medium text-gray-500">Action</h4>
                  <p class="text-sm text-gray-900">
                    <span :class="getActionBadgeClass(selectedLog.action)">
                      {{ formatAction(selectedLog.action) }}
                    </span>
                  </p>
                </div>
                <div>
                  <h4 class="text-sm font-medium text-gray-500">Entity Type</h4>
                  <p class="text-sm text-gray-900">{{ selectedLog.entity_type || 'N/A' }}</p>
                </div>
                <div>
                  <h4 class="text-sm font-medium text-gray-500">Entity ID</h4>
                  <p class="text-sm text-gray-900 font-mono">{{ selectedLog.entity_id || 'N/A' }}</p>
                </div>
                <div>
                  <h4 class="text-sm font-medium text-gray-500">Performed By</h4>
                  <p class="text-sm text-gray-900">{{ formatUser(selectedLog.performed_by) }}</p>
                </div>
                <div>
                  <h4 class="text-sm font-medium text-gray-500">IP Address</h4>
                  <p class="text-sm text-gray-900">{{ selectedLog.ip_address || 'N/A' }}</p>
                </div>
              </div>

              <!-- User Agent -->
              <div v-if="selectedLog.user_agent">
                <h4 class="text-sm font-medium text-gray-500">User Agent</h4>
                <p class="text-sm text-gray-900 bg-gray-50 p-2 rounded break-all">{{ selectedLog.user_agent }}</p>
              </div>

              <!-- Detailed Information -->
              <div v-if="selectedLog.details && Object.keys(selectedLog.details).length > 0">
                <h4 class="text-sm font-medium text-gray-500 mb-2">Additional Details</h4>
                <div class="bg-gray-50 p-3 rounded-md">
                  <pre class="text-sm text-gray-800 whitespace-pre-wrap">{{ JSON.stringify(selectedLog.details, null, 2) }}</pre>
                </div>
              </div>
            </div>

            <div class="mt-6 flex justify-end">
              <button @click="closeDetailsModal" class="btn btn-secondary">
                Close
              </button>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="response && response.total_pages > 1" class="px-6 py-4 bg-gray-50 border-t border-gray-200">
          <div class="flex items-center justify-between">
            <div class="text-sm text-gray-700">
              Showing {{ ((pagination.page - 1) * pagination.limit) + 1 }} to
              {{ Math.min(pagination.page * pagination.limit, response.total) }} of
              {{ response.total }} results
            </div>
            <div class="flex items-center space-x-2">
              <button
                @click="changePage(pagination.page - 1)"
                :disabled="pagination.page <= 1"
                class="btn btn-secondary"
                :class="{ 'opacity-50 cursor-not-allowed': pagination.page <= 1 }"
              >
                Previous
              </button>
              <span class="px-3 py-1 text-sm text-gray-700">
                Page {{ pagination.page }} of {{ response.total_pages }}
              </span>
              <button
                @click="changePage(pagination.page + 1)"
                :disabled="pagination.page >= response.total_pages"
                class="btn btn-secondary"
                :class="{ 'opacity-50 cursor-not-allowed': pagination.page >= response.total_pages }"
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
import { ref, reactive, onMounted, watch, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { auditAPI } from '@/services/api'
import { showErrorToast } from '@/utils/toast'

interface AuditLog {
  id: string
  timestamp: string
  performed_by: string
  action: string
  entity_type: string
  entity_id?: string
  details?: Record<string, any>
  ip_address?: string
  user_agent?: string
}

interface AuditListResponse {
  audit_logs: AuditLog[]
  total: number
  page: number
  limit: number
  total_pages: number
}

const authStore = useAuthStore()

const loading = ref(false)
const response = ref<AuditListResponse | null>(null)
const selectedLog = ref<AuditLog | null>(null)

const filters = reactive({
  search: '',
  entity: '',
  action: '',
  performed_by: '',
  from_date: '',
  to_date: '',
  sort: 'timestamp',
  order: 'desc',
})

const pagination = reactive({
  page: 1,
  limit: 50,
})

// Simple debounce function implementation
const debounce = (func: Function, wait: number) => {
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

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatRelativeTime = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60))

  if (diffInHours < 1) {
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))
    return `${diffInMinutes} minutes ago`
  } else if (diffInHours < 24) {
    return `${diffInHours} hours ago`
  } else {
    const diffInDays = Math.floor(diffInHours / 24)
    return `${diffInDays} days ago`
  }
}

const formatAction = (action: string) => {
  return action
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

const formatUser = (performedBy: string) => {
  // For now, just show "Admin" for the admin user ID, otherwise show the UUID
  if (performedBy === '1a8ede90-8ca3-4976-8310-9002c9484024') {
    return 'Admin'
  }
  return performedBy.substring(0, 8) + '...'
}

const getActionBadgeClass = (action: string) => {
  switch (action) {
    case 'create':
      return 'badge-success'
    case 'update':
      return 'badge-warning'
    case 'delete':
      return 'badge-danger'
    case 'login':
      return 'badge-info'
    case 'logout':
      return 'badge-secondary'
    default:
      return 'badge'
  }
}

const getLogDetails = (log: AuditLog) => {
  if (log.action === 'create' && log.details) {
    const name = log.details.ci_name || log.details.name || log.details.username || 'Item'
    const type = log.details.ci_type || log.entity_type || 'item'
    return `Created ${name} (${type})`
  } else if (log.action === 'update' && log.details) {
    const name = log.details.ci_name || log.details.name || 'Item'
    const type = log.details.ci_type || log.entity_type || 'item'
    return `Updated ${name} (${type})`
  } else if (log.action === 'delete' && log.details) {
    const name = log.details.ci_name || log.details.name || 'Item'
    const type = log.details.ci_type || log.entity_type || 'item'
    return `Deleted ${name} (${type})`
  } else if (log.action === 'login') {
    return 'User logged in'
  } else if (log.action === 'logout') {
    return 'User logged out'
  }
  return log.action
}

const hasMoreDetails = (log: AuditLog) => {
  return (log.details && Object.keys(log.details).length > 0)
}

const resetFilters = () => {
  filters.search = ''
  filters.entity = ''
  filters.action = ''
  filters.performed_by = ''
  filters.from_date = ''  // Clear date filters to show all logs
  filters.to_date = ''    // Clear date filters to show all logs
  filters.sort = 'timestamp'
  filters.order = 'desc'

  pagination.page = 1
  loadAuditLogs()
}

const clearDateFilters = () => {
  filters.from_date = ''  // Clear only date filters
  filters.to_date = ''    // Clear only date filters
  pagination.page = 1
  loadAuditLogs()
}

const showLogDetails = (log: AuditLog) => {
  selectedLog.value = log
}

const closeDetailsModal = () => {
  selectedLog.value = null
}

// Handle escape key to close modal
const handleEscapeKey = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && selectedLog.value) {
    closeDetailsModal()
  }
}

// Add and remove event listener for escape key
watch(selectedLog, (newValue) => {
  if (newValue) {
    document.addEventListener('keydown', handleEscapeKey)
  } else {
    document.removeEventListener('keydown', handleEscapeKey)
  }
})

const sortBy = (column: string) => {
  if (filters.sort === column) {
    filters.order = filters.order === 'asc' ? 'desc' : 'asc'
  } else {
    filters.sort = column
    filters.order = 'desc'
  }
  loadAuditLogs()
}

const changePage = (page: number) => {
  if (page >= 1 && response.value && page <= response.value.total_pages) {
    pagination.page = page
    loadAuditLogs()
  }
}

const loadAuditLogs = async () => {
  // Wait for auth to be initialized
  if (!authStore.isInitialized) {
    console.log('Audit logs: Auth not initialized, skipping load')
    return
  }

  // Check if authenticated
  if (!authStore.isAuthenticated) {
    console.log('Audit logs: User not authenticated, skipping load')
    return
  }

  // Check permissions
  if (!hasPermission('audit:read')) {
    console.log('Audit logs: User does not have audit:read permission')
    return
  }

  console.log('Audit logs: Loading with filters:', filters)
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

    console.log('Audit logs: API call params:', params)
    const apiResponse = await auditAPI.list(params)
    console.log('Audit logs: API response:', apiResponse)

    // Handle wrapped API response
    if (apiResponse.data) {
      console.log('Audit logs: Extracted data from response.data:', apiResponse.data)
      response.value = apiResponse.data
    } else {
      console.log('Audit logs: Using response directly')
      response.value = apiResponse
    }

  } catch (error) {
    console.error('Failed to load audit logs:', error)

    // More specific error messages
    if (error.response?.status === 403) {
      showErrorToast('You do not have permission to view audit logs')
    } else if (error.response?.status === 401) {
      showErrorToast('Authentication required. Please log in again.')
    } else {
      showErrorToast(`Failed to load audit logs: ${error.message || 'Unknown error'}`)
    }
  } finally {
    loading.value = false
  }
}

// Debounced version for search inputs
const debouncedLoadAuditLogs = debounce(() => {
  pagination.page = 1
  loadAuditLogs()
}, 500)

// Watch for auth initialization to complete and then load audit logs
watch(
  () => authStore.isInitialized,
  async (isInitialized) => {
    if (isInitialized && authStore.isAuthenticated && hasPermission('audit:read')) {
      await loadAuditLogs()
    }
  },
  { immediate: true }
)

onMounted(async () => {
  // Set default date range to last 30 days
  const today = new Date()
  const thirtyDaysAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)

  filters.to_date = today.toISOString().split('T')[0]
  filters.from_date = thirtyDaysAgo.toISOString().split('T')[0]

  // Ensure auth is initialized before trying to load audit logs
  if (!authStore.isInitialized) {
    try {
      await authStore.checkAuth()
    } catch (error) {
      console.error('Auth check failed:', error)
    }
  }

  // Try to load audit logs if auth is ready
  if (authStore.isInitialized && authStore.isAuthenticated) {
    await loadAuditLogs()
  }
})
</script>
<template>
  <div class="page-container page-content">
    <div class="page-header">
      <h1 class="page-title">Audit Logs</h1>
      <p class="page-subtitle">View system audit logs and activity history</p>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div>
          <label class="form-label">Search</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search logs..."
            class="form-input"
            @input="debouncedSearch"
          />
        </div>
        <div>
          <label class="form-label">Entity Type</label>
          <select v-model="filters.entity" class="form-input" @change="loadAuditLogs">
            <option value="">All Entities</option>
            <option value="ci">Configuration Item</option>
            <option value="ci_type">CI Type</option>
            <option value="user">User</option>
            <option value="relationship">Relationship</option>
          </select>
        </div>
        <div>
          <label class="form-label">Action</label>
          <select v-model="filters.action" class="form-input" @change="loadAuditLogs">
            <option value="">All Actions</option>
            <option value="create">Create</option>
            <option value="update">Update</option>
            <option value="delete">Delete</option>
            <option value="login">Login</option>
            <option value="logout">Logout</option>
          </select>
        </div>
        <div>
          <label class="form-label">From Date</label>
          <input
            v-model="filters.from_date"
            type="date"
            class="form-input"
            @change="loadAuditLogs"
          />
        </div>
        <div>
          <label class="form-label">To Date</label>
          <input
            v-model="filters.to_date"
            type="date"
            class="form-input"
            @change="loadAuditLogs"
          />
        </div>
      </div>
      <div class="mt-4 flex items-end">
        <button @click="loadAuditLogs" :disabled="loading" class="btn btn-primary">
          <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
          Search Logs
        </button>
        <button @click="clearFilters" class="btn btn-outline ml-2">
          Clear Filters
        </button>
      </div>
    </div>

    <!-- Audit Logs List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Audit Logs ({{ response?.total || 0 }})
        </h3>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">Loading audit logs...</p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && (!response?.audit_logs || response.audit_logs.length === 0)" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No audit logs found</h3>
          <p class="mt-1 text-sm text-gray-500">
            Try adjusting your search criteria or date range.
          </p>
        </div>

        <!-- Audit Logs Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell">Timestamp</th>
                <th class="table-header-cell">User</th>
                <th class="table-header-cell">Action</th>
                <th class="table-header-cell">Entity</th>
                <th class="table-header-cell">Details</th>
                <th class="table-header-cell">IP Address</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="log in response?.audit_logs" :key="log.id">
                <td class="table-cell">
                  <div>
                    <div class="text-sm text-gray-900">{{ formatDateTime(log.timestamp) }}</div>
                    <div class="text-xs text-gray-500">{{ formatRelativeTime(log.timestamp) }}</div>
                  </div>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ log.performed_by || 'System' }}</div>
                </td>
                <td class="table-cell">
                  <span :class="getActionBadgeClass(log.action)">
                    {{ formatAction(log.action) }}
                  </span>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ log.entity || 'N/A' }}</div>
                  <div v-if="log.entity_id" class="text-xs text-gray-500">
                    ID: {{ log.entity_id.substring(0, 8) }}...
                  </div>
                </td>
                <td class="table-cell">
                  <div class="max-w-xs">
                    <div class="text-sm text-gray-900 truncate">
                      {{ getLogDetails(log) }}
                    </div>
                    <div v-if="hasMoreDetails(log)" class="text-xs text-blue-600 hover:text-blue-900 cursor-pointer">
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
import { auditAPI } from '@/services/api'
import { showErrorToast } from '@/utils/toast'

interface AuditLog {
  id: string
  timestamp: string
  performed_by: string
  action: string
  entity: string
  entity_id?: string
  old_values?: Record<string, any>
  new_values?: Record<string, any>
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
  if (log.action === 'create' && log.new_values) {
    const name = log.new_values.name || log.new_values.username || 'Item'
    return `Created ${name}`
  } else if (log.action === 'update' && log.old_values && log.new_values) {
    const name = log.new_values.name || log.old_values.name || 'Item'
    const changedFields = Object.keys(log.new_values).filter(key =>
      log.old_values[key] !== log.new_values[key]
    ).length
    return `Updated ${name} (${changedFields} fields changed)`
  } else if (log.action === 'delete' && log.old_values) {
    const name = log.old_values.name || log.old_values.username || 'Item'
    return `Deleted ${name}`
  } else if (log.action === 'login') {
    return 'User logged in'
  } else if (log.action === 'logout') {
    return 'User logged out'
  }
  return log.action
}

const hasMoreDetails = (log: AuditLog) => {
  return (log.old_values && Object.keys(log.old_values).length > 0) ||
         (log.new_values && Object.keys(log.new_values).length > 0)
}

const debouncedSearch = debounce(() => {
  pagination.page = 1
  loadAuditLogs()
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

const loadAuditLogs = async () => {
  if (!hasPermission('audit:read')) return

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

    response.value = await auditAPI.list(params)
  } catch (error) {
    console.error('Failed to load audit logs:', error)
    showErrorToast('Failed to load audit logs')
  } finally {
    loading.value = false
  }
}

const clearFilters = () => {
  filters.search = ''
  filters.entity = ''
  filters.action = ''
  filters.performed_by = ''
  filters.from_date = ''
  filters.to_date = ''
  pagination.page = 1
  loadAuditLogs()
}

const goToPage = (page: number) => {
  pagination.page = page
  loadAuditLogs()
}

onMounted(() => {
  // Set default date range to last 30 days
  const today = new Date()
  const thirtyDaysAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)

  filters.to_date = today.toISOString().split('T')[0]
  filters.from_date = thirtyDaysAgo.toISOString().split('T')[0]

  loadAuditLogs()
})
</script>
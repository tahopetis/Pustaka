<template>
  <div class="page-container page-content">
    <div class="page-header">
      <h1 class="page-title">Audit Logs (Debug Version)</h1>
      <p class="page-subtitle">Debug version to identify issues</p>
    </div>

    <!-- Debug Information -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <h3>Debug Information</h3>
      <div>
        <p><strong>Is Authenticated:</strong> {{ authStore.isAuthenticated }}</p>
        <p><strong>Is Initialized:</strong> {{ authStore.isInitialized }}</p>
        <p><strong>User:</strong> {{ authStore.user?.username || 'No user' }}</p>
        <p><strong>Permissions:</strong> {{ authStore.user?.permissions?.join(', ') || 'No permissions' }}</p>
        <p><strong>Has audit:read permission:</strong> {{ hasPermission('audit:read') }}</p>
        <p><strong>Access Token:</strong> {{ authStore.accessToken ? `${authStore.accessToken.substring(0, 20)}...` : 'No token' }}</p>
        <p><strong>Loading:</strong> {{ loading }}</p>
        <p><strong>Response:</strong> {{ response ? 'Has data' : 'No data' }}</p>
        <p><strong>API Base URL:</strong> {{ apiBaseURL }}</p>
      </div>
      <div class="mt-4">
        <button @click="testAuth" class="btn btn-primary mr-2">Test Auth</button>
        <button @click="testAPICall" class="btn btn-primary mr-2">Test API Call</button>
        <button @click="debugLoadAuditLogs" class="btn btn-primary">Debug Load Audit Logs</button>
      </div>
    </div>

    <!-- Original Audit Logs List -->
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
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { auditAPI, api } from '@/services/api'
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

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const hasPermission = (permission: string) => {
  console.log(`Checking permission: ${permission}`)
  console.log('User permissions:', authStore.user?.permissions)
  const result = authStore.hasPermission(permission)
  console.log(`Permission ${permission}: ${result}`)
  return result
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

const testAuth = async () => {
  console.log('=== Testing Auth ===')
  console.log('Is authenticated:', authStore.isAuthenticated)
  console.log('User:', authStore.user)
  console.log('Access token:', authStore.accessToken ? `${authStore.accessToken.substring(0, 20)}...` : 'No token')
  console.log('Permissions:', authStore.user?.permissions)

  // Try to refresh user data
  try {
    const user = await authStore.getUserProfile()
    console.log('Refreshed user data:', user)
  } catch (error) {
    console.error('Failed to refresh user data:', error)
  }
}

const testAPICall = async () => {
  console.log('=== Testing API Call ===')
  try {
    const response = await api.get('/audit')
    console.log('Raw API response:', response.data)
    console.log('Response status:', response.status)
    console.log('Response headers:', response.headers)
  } catch (error) {
    console.error('API call failed:', error)
    console.error('Error response:', error.response?.data)
    console.error('Error status:', error.response?.status)
  }
}

const debugLoadAuditLogs = async () => {
  console.log('=== Debug Load Audit Logs ===')

  console.log('1. Checking permission...')
  const hasAuditPermission = hasPermission('audit:read')
  console.log('Has audit:read permission:', hasAuditPermission)

  if (!hasAuditPermission) {
    console.log('❌ Permission check failed - aborting')
    return
  }

  console.log('2. Starting API call...')
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

    console.log('3. API call params:', params)
    console.log('4. Making API call to:', `/audit`)

    const apiResponse = await auditAPI.list(params)
    console.log('5. API response received:', apiResponse)

    response.value = apiResponse
    console.log('6. Response set in component')

  } catch (error) {
    console.error('7. API call failed:', error)
    console.error('Error details:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status
    })
    showErrorToast('Failed to load audit logs')
  } finally {
    loading.value = false
    console.log('8. Loading state set to false')
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

onMounted(() => {
  console.log('=== AuditViewDebug mounted ===')
  console.log('Auth state:', {
    isAuthenticated: authStore.isAuthenticated,
    isInitialized: authStore.isInitialized,
    user: authStore.user,
    accessToken: authStore.accessToken ? `${authStore.accessToken.substring(0, 20)}...` : 'No token'
  })

  // Set default date range to last 30 days
  const today = new Date()
  const thirtyDaysAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)

  filters.to_date = today.toISOString().split('T')[0]
  filters.from_date = thirtyDaysAgo.toISOString().split('T')[0]

  console.log('Calling loadAuditLogs...')
  loadAuditLogs()
})
</script>
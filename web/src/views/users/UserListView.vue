<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">Users</h1>
        <p class="page-subtitle">Manage user accounts and permissions</p>
      </div>
      <router-link
        v-if="hasPermission('user:create')"
        to="/users/new"
        class="btn btn-primary"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        Add User
      </router-link>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="md:col-span-2">
          <label class="form-label">Search Users</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search by username or email..."
            class="form-input"
            @input="debouncedSearch"
          />
        </div>
        <div class="flex items-end">
          <button @click="loadUsers" :disabled="loading" class="btn btn-primary w-full">
            <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
            Search
          </button>
        </div>
      </div>
    </div>

    <!-- Users List -->
    <div class="bg-white shadow rounded-lg">
      <div class="card-header">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Users ({{ response?.total || 0 }})
        </h3>
      </div>
      <div class="card-body p-0">
        <!-- Loading state -->
        <div v-if="loading" class="text-center py-12">
          <div class="spinner w-8 h-8 mx-auto mb-4"></div>
          <p class="text-gray-500">Loading users...</p>
        </div>

        <!-- Empty state -->
        <div v-else-if="!loading && (!response?.users || response.users.length === 0)" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No users found</h3>
          <p class="mt-1 text-sm text-gray-500">
            Get started by creating a new user account.
          </p>
          <div class="mt-6" v-if="hasPermission('user:create')">
            <router-link to="/users/new" class="btn btn-primary">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              Add User
            </router-link>
          </div>
        </div>

        <!-- Users Table -->
        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead class="table-header">
              <tr>
                <th class="table-header-cell">User</th>
                <th class="table-header-cell">Email</th>
                <th class="table-header-cell">Roles</th>
                <th class="table-header-cell">Status</th>
                <th class="table-header-cell">Created</th>
                <th class="table-header-cell">Actions</th>
              </tr>
            </thead>
            <tbody class="table-body">
              <tr v-for="user in response?.users" :key="user.id">
                <td class="table-cell">
                  <div class="flex items-center">
                    <div class="h-10 w-10 flex-shrink-0">
                      <div class="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                        <span class="text-sm font-medium text-gray-700">
                          {{ user.username?.charAt(0).toUpperCase() }}
                        </span>
                      </div>
                    </div>
                    <div class="ml-4">
                      <div class="text-sm font-medium text-gray-900">{{ user.username }}</div>
                      <div class="text-sm text-gray-500">ID: {{ user.id.substring(0, 8) }}...</div>
                    </div>
                  </div>
                </td>
                <td class="table-cell">
                  <div class="text-sm text-gray-900">{{ user.email }}</div>
                </td>
                <td class="table-cell">
                  <div class="flex flex-wrap gap-1">
                    <span v-for="role in user.roles" :key="role" class="badge badge-info">
                      {{ formatRole(role) }}
                    </span>
                  </div>
                </td>
                <td class="table-cell">
                  <span :class="user.is_active ? 'badge-success' : 'badge-danger'">
                    {{ user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="table-cell">
                  {{ formatDate(user.created_at) }}
                </td>
                <td class="table-cell">
                  <div class="flex space-x-2">
                    <router-link
                      :to="`/users/${user.id}`"
                      class="text-blue-600 hover:text-blue-900"
                      title="View"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                      </svg>
                    </router-link>
                    <router-link
                      v-if="hasPermission('user:update')"
                      :to="`/users/${user.id}/edit`"
                      class="text-indigo-600 hover:text-indigo-900"
                      title="Edit"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                      </svg>
                    </router-link>
                    <button
                      v-if="hasPermission('user:delete') && user.id !== currentUser?.id"
                      @click="confirmDelete(user)"
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface User {
  id: string
  username: string
  email: string
  is_active: boolean
  roles: string[]
  created_at: string
  updated_at: string
}

interface UserListResponse {
  users: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}

const authStore = useAuthStore()

const loading = ref(false)
const response = ref<UserListResponse | null>(null)

const filters = reactive({
  search: '',
  sort: 'created_at',
  order: 'desc',
})

const pagination = reactive({
  page: 1,
  limit: 20,
})

const currentUser = computed(() => authStore.user)

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const formatRole = (role: string) => {
  return role
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

const debouncedSearch = debounce(() => {
  pagination.page = 1
  loadUsers()
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

const loadUsers = async () => {
  if (!hasPermission('user:read')) return

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

    response.value = await userAPI.list(params)
  } catch (error) {
    console.error('Failed to load users:', error)
    showErrorToast('Failed to load users')
  } finally {
    loading.value = false
  }
}

const goToPage = (page: number) => {
  pagination.page = page
  loadUsers()
}

const confirmDelete = async (user: User) => {
  if (confirm(`Are you sure you want to delete the user "${user.username}"? This action cannot be undone.`)) {
    try {
      await userAPI.delete(user.id)
      showSuccessToast('User deleted successfully')
      await loadUsers()
    } catch (error: any) {
      console.error('Failed to delete user:', error)
      const message = error.response?.data?.error || 'Failed to delete user'
      showErrorToast(message)
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>
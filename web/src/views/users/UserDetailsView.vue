<template>
  <div class="page-container page-content">
    <div class="page-header">
      <div class="mb-6">
        <router-link to="/users" class="text-blue-600 hover:text-blue-800 flex items-center">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
          </svg>
          Back to Users
        </router-link>
      </div>
      <div>
        <h1 class="page-title">User Details</h1>
        <p class="page-subtitle">View user information and permissions</p>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
      <p class="mt-4 text-gray-600">Loading user details...</p>
    </div>

    <div v-else-if="user" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- User Information -->
      <div class="lg:col-span-2">
        <div class="bg-white shadow rounded-lg p-6">
          <div class="flex justify-between items-start mb-6">
            <h2 class="text-xl font-semibold text-gray-900">User Information</h2>
            <router-link
              v-if="hasPermission('user:update')"
              :to="`/users/${user.id}/edit`"
              class="btn btn-outline btn-sm"
            >
              Edit User
            </router-link>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label class="form-label">Username</label>
              <div class="form-value">{{ user.username }}</div>
            </div>
            <div>
              <label class="form-label">Email</label>
              <div class="form-value">{{ user.email }}</div>
            </div>
            <div>
              <label class="form-label">Status</label>
              <span :class="user.is_active ? 'badge-success' : 'badge-danger'">
                {{ user.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <div>
              <label class="form-label">User ID</label>
              <div class="form-value font-mono text-sm">{{ user.id }}</div>
            </div>
          </div>
        </div>

        <!-- Roles -->
        <div class="bg-white shadow rounded-lg p-6 mt-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">Roles</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="role in user.roles"
              :key="role.id"
              class="badge badge-info"
            >
              {{ formatRole(role.name) }}
            </span>
            <span v-if="!user.roles || user.roles.length === 0" class="text-gray-500">
              No roles assigned
            </span>
          </div>
        </div>

        <!-- Permissions -->
        <div class="bg-white shadow rounded-lg p-6 mt-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">Permissions</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
            <span
              v-for="permission in user.permissions"
              :key="permission"
              class="text-sm bg-gray-100 px-2 py-1 rounded"
            >
              {{ permission }}
            </span>
            <span v-if="!user.permissions || user.permissions.length === 0" class="text-gray-500 col-span-full">
              No permissions assigned
            </span>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div>
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
          <div class="space-y-3">
            <router-link
              v-if="hasPermission('user:update')"
              :to="`/users/${user.id}/edit`"
              class="btn btn-primary btn-sm w-full"
            >
              Edit User
            </router-link>
            <button
              v-if="hasPermission('user:delete') && user.id !== currentUser?.id"
              @click="confirmDelete"
              class="btn btn-danger btn-sm w-full"
            >
              Delete User
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="error" class="text-center py-12">
      <svg class="w-16 h-16 mx-auto text-red-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Error Loading User</h3>
      <p class="text-gray-600 mb-4">{{ error }}</p>
      <router-link to="/users" class="btn btn-outline">
        Back to Users
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface User {
  id: string
  username: string
  email: string
  is_active: boolean
  roles: Array<{
    id: string
    name: string
    description: string
  }>
  permissions: string[]
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(true)
const user = ref<User | null>(null)
const error = ref('')

const currentUser = computed(() => authStore.user)

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatRole = (roleName: string) => {
  return roleName.charAt(0).toUpperCase() + roleName.slice(1)
}

const loadUser = async () => {
  const userId = route.params.id as string

  try {
    loading.value = true
    error.value = ''

    const response = await userAPI.get(userId)
    user.value = response.data
  } catch (err: any) {
    console.error('Failed to load user:', err)
    error.value = err.response?.data?.message || 'Failed to load user'
    showErrorToast(error.value)
  } finally {
    loading.value = false
  }
}

const confirmDelete = () => {
  if (!user.value) return

  if (confirm(`Are you sure you want to delete the user "${user.value.username}"? This action cannot be undone.`)) {
    deleteUser()
  }
}

const deleteUser = async () => {
  if (!user.value) return

  try {
    await userAPI.delete(user.value.id)
    showSuccessToast('User deleted successfully')
    router.push('/users')
  } catch (err: any) {
    console.error('Failed to delete user:', err)
    const errorMessage = err.response?.data?.message || 'Failed to delete user'
    showErrorToast(errorMessage)
  }
}

onMounted(async () => {
  if (!authStore.hasPermission('user:read')) {
    showErrorToast('You do not have permission to view user details')
    router.push('/dashboard')
    return
  }

  await loadUser()
})
</script>
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
        <h1 class="page-title">Edit User</h1>
        <p class="page-subtitle">Update user account information</p>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
      <p class="mt-4 text-gray-600">Loading user information...</p>
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

    <div v-else-if="user" class="max-w-2xl">
      <form @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6">
        <!-- Username (Read-only) -->
        <div class="mb-6">
          <label for="username" class="form-label">Username</label>
          <input
            id="username"
            :value="user.username"
            type="text"
            disabled
            class="form-input bg-gray-50"
            placeholder="Username (cannot be changed)"
          />
          <p class="form-help">Username cannot be changed after account creation</p>
        </div>

        <!-- Email -->
        <div class="mb-6">
          <label for="email" class="form-label">Email *</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            required
            :disabled="loading"
            class="form-input"
            placeholder="Enter email address"
          />
          <p v-if="errors.email" class="form-error">{{ errors.email }}</p>
        </div>

        <!-- Roles -->
        <div class="mb-6">
          <label class="form-label">Roles</label>
          <div class="space-y-2">
            <label v-for="role in availableRoles" :key="role.id" class="flex items-center">
              <input
                v-model="form.roles"
                :value="role.name"
                type="checkbox"
                :disabled="loading"
                class="form-checkbox"
              />
              <span class="ml-2">
                {{ formatRole(role.name) }}
                <span class="text-gray-500 text-sm">- {{ role.description }}</span>
              </span>
            </label>
          </div>
          <p v-if="errors.roles" class="form-error">{{ errors.roles }}</p>
          <p class="form-help">Select roles for the user. If no roles are selected, the user will have no permissions.</p>
        </div>

        <!-- Active Status -->
        <div class="mb-6">
          <label class="flex items-center">
            <input
              v-model="form.is_active"
              type="checkbox"
              :disabled="loading || user.id === currentUser?.id"
              class="form-checkbox"
            />
            <span class="ml-2">Active</span>
          </label>
          <p class="form-help">
            Inactive users cannot log in to the system
            <span v-if="user.id === currentUser?.id" class="text-orange-600">
              (You cannot deactivate your own account)
            </span>
          </p>
        </div>

        <!-- Form Actions -->
        <div class="flex justify-end space-x-4">
          <router-link
            :to="`/users/${user.id}`"
            class="btn btn-outline"
            :disabled="loading"
          >
            Cancel
          </router-link>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="loading || !isFormValid"
          >
            <span v-if="loading" class="animate-spin mr-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
              </svg>
            </span>
            {{ loading ? 'Updating...' : 'Update User' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface Role {
  id: string
  name: string
  description: string
}

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
const submitting = ref(false)
const user = ref<User | null>(null)
const error = ref('')
const availableRoles = ref<Role[]>([])

const form = reactive({
  email: '',
  roles: [] as string[],
  is_active: true,
})

const errors = reactive({
  email: '',
  roles: '',
})

const currentUser = computed(() => authStore.user)

const isFormValid = computed(() => {
  return (
    form.email.trim() !== '' &&
    !Object.values(errors).some(error => error !== '')
  )
})

const formatRole = (roleName: string) => {
  return roleName.charAt(0).toUpperCase() + roleName.slice(1)
}

const validateForm = () => {
  // Reset errors
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })

  // Validate email
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(form.email)) {
    errors.email = 'Please enter a valid email address'
  }

  return Object.values(errors).every(error => error === '')
}

const loadUser = async () => {
  const userId = route.params.id as string

  try {
    loading.value = true
    error.value = ''

    const response = await userAPI.get(userId)
    user.value = response.data

    // Populate form
    form.email = user.value.email
    form.roles = user.value.roles.map(role => role.name)
    form.is_active = user.value.is_active
  } catch (err: any) {
    console.error('Failed to load user:', err)
    error.value = err.response?.data?.message || 'Failed to load user'
    showErrorToast(error.value)
  } finally {
    loading.value = false
  }
}

const loadRoles = async () => {
  try {
    // For now, we'll use hardcoded roles since we don't have a roles API
    // In a real implementation, you'd fetch this from the API
    availableRoles.value = [
      { id: '1', name: 'admin', description: 'Full system access' },
      { id: '2', name: 'editor', description: 'Can edit and view content' },
      { id: '3', name: 'viewer', description: 'Read-only access' },
    ]
  } catch (error) {
    console.error('Failed to load roles:', error)
    showErrorToast('Failed to load available roles')
  }
}

const handleSubmit = async () => {
  if (!validateForm() || !user.value) {
    return
  }

  submitting.value = true

  try {
    const userData = {
      email: form.email.trim(),
      is_active: form.is_active,
      roles: form.roles.length > 0 ? form.roles : undefined,
    }

    await userAPI.update(user.value.id, userData)

    showSuccessToast('User updated successfully')
    router.push(`/users/${user.value.id}`)
  } catch (error: any) {
    console.error('Failed to update user:', error)

    let errorMessage = 'Failed to update user'
    if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    }

    showErrorToast(errorMessage)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (!authStore.hasPermission('user:update')) {
    showErrorToast('You do not have permission to edit users')
    router.push('/users')
    return
  }

  await Promise.all([loadUser(), loadRoles()])
})
</script>
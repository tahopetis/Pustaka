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
        <h1 class="page-title">Create User</h1>
        <p class="page-subtitle">Create a new user account</p>
      </div>
    </div>

    <div class="max-w-2xl">
      <form @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Username -->
          <div>
            <label for="username" class="form-label">Username *</label>
            <input
              id="username"
              v-model="form.username"
              type="text"
              required
              :disabled="loading"
              class="form-input"
              placeholder="Enter username"
            />
            <p v-if="errors.username" class="form-error">{{ errors.username }}</p>
          </div>

          <!-- Email -->
          <div>
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
        </div>

        <!-- Password -->
        <div class="mt-6">
          <label for="password" class="form-label">Password *</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            :disabled="loading"
            class="form-input"
            placeholder="Enter password"
          />
          <p v-if="errors.password" class="form-error">{{ errors.password }}</p>
          <p class="form-help">Password should be at least 8 characters long</p>
        </div>

        <!-- Confirm Password -->
        <div class="mt-6">
          <label for="confirmPassword" class="form-label">Confirm Password *</label>
          <input
            id="confirmPassword"
            v-model="form.confirmPassword"
            type="password"
            required
            :disabled="loading"
            class="form-input"
            placeholder="Confirm password"
          />
          <p v-if="errors.confirmPassword" class="form-error">{{ errors.confirmPassword }}</p>
        </div>

        <!-- Roles -->
        <div class="mt-6">
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
        <div class="mt-6">
          <label class="flex items-center">
            <input
              v-model="form.is_active"
              type="checkbox"
              :disabled="loading"
              class="form-checkbox"
            />
            <span class="ml-2">Active</span>
          </label>
          <p class="form-help">Inactive users cannot log in to the system</p>
        </div>

        <!-- Form Actions -->
        <div class="mt-8 flex justify-end space-x-4">
          <router-link
            to="/users"
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
            {{ loading ? 'Creating...' : 'Create User' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface Role {
  id: string
  name: string
  description: string
}

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const availableRoles = ref<Role[]>([])

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  roles: [] as string[],
  is_active: true,
})

const errors = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  roles: '',
})

const isFormValid = computed(() => {
  return (
    form.username.trim() !== '' &&
    form.email.trim() !== '' &&
    form.password.length >= 8 &&
    form.password === form.confirmPassword &&
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

  // Validate username
  if (form.username.trim().length < 3) {
    errors.username = 'Username must be at least 3 characters long'
  }

  // Validate email
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(form.email)) {
    errors.email = 'Please enter a valid email address'
  }

  // Validate password
  if (form.password.length < 8) {
    errors.password = 'Password must be at least 8 characters long'
  }

  // Validate confirm password
  if (form.password !== form.confirmPassword) {
    errors.confirmPassword = 'Passwords do not match'
  }

  return Object.values(errors).every(error => error === '')
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
  if (!validateForm()) {
    return
  }

  loading.value = true

  try {
    const userData = {
      username: form.username.trim(),
      email: form.email.trim(),
      password: form.password,
      roles: form.roles.length > 0 ? form.roles : undefined,
    }

    await userAPI.create(userData)

    showSuccessToast('User created successfully')
    router.push('/users')
  } catch (error: any) {
    console.error('Failed to create user:', error)

    let errorMessage = 'Failed to create user'
    if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    }

    showErrorToast(errorMessage)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!authStore.hasPermission('user:create')) {
    showErrorToast('You do not have permission to create users')
    router.push('/users')
    return
  }

  loadRoles()
})
</script>
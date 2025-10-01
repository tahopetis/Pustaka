<template>
  <div class="px-4 py-6 sm:px-0">
    <div class="max-w-3xl mx-auto">
      <!-- Page header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900">Profile</h1>
        <p class="mt-2 text-gray-600">Manage your account settings and preferences</p>
      </div>

      <div class="bg-white shadow rounded-lg">
        <div class="card-body">
          <div v-if="loading" class="text-center py-12">
            <div class="spinner w-8 h-8 mx-auto mb-4"></div>
            <p class="text-gray-500">Loading profile...</p>
          </div>

          <div v-else-if="user" class="space-y-6">
            <!-- User Information -->
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-4">User Information</h3>
              <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <div>
                  <dt class="text-sm font-medium text-gray-500">Username</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ user.username }}</dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">Email</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ user.email }}</dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">Status</dt>
                  <dd class="mt-1">
                    <span :class="user.is_active ? 'badge badge-success' : 'badge badge-danger'">
                      {{ user.is_active ? 'Active' : 'Inactive' }}
                    </span>
                  </dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">User ID</dt>
                  <dd class="mt-1 text-sm text-gray-900 font-mono text-xs">{{ user.id }}</dd>
                </div>
              </dl>
            </div>

            <!-- Roles -->
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-4">Roles</h3>
              <div v-if="user.roles.length === 0" class="text-gray-500">
                No roles assigned
              </div>
              <div v-else class="flex flex-wrap gap-2">
                <span v-for="role in user.roles" :key="role.id" class="badge badge-info">
                  {{ role.name }}
                </span>
              </div>
            </div>

            <!-- Permissions -->
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-4">Permissions</h3>
              <div v-if="user.permissions.length === 0" class="text-gray-500">
                No permissions assigned
              </div>
              <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
                <span v-for="permission in user.permissions" :key="permission" class="badge badge-success text-xs">
                  {{ permission }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { showErrorToast } from '@/utils/toast'

const authStore = useAuthStore()

const loading = ref(false)
const user = ref(authStore.user)

const loadProfile = async () => {
  loading.value = true
  try {
    user.value = await authStore.getUserProfile()
  } catch (error) {
    console.error('Failed to load profile:', error)
    showErrorToast('Failed to load profile')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!user.value) {
    loadProfile()
  }
})
</script>
<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <nav class="flex" aria-label="Breadcrumb">
          <ol class="flex items-center space-x-4">
            <li>
              <router-link to="/lifecycle-status" class="text-gray-500 hover:text-gray-700">
                Lifecycle Statuses
              </router-link>
            </li>
            <li>
              <div class="flex items-center">
                <svg class="flex-shrink-0 h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="ml-4 text-sm font-medium text-gray-900">Lifecycle Status Details</span>
              </div>
            </li>
          </ol>
        </nav>
        <div class="flex items-center mt-2">
          <div
            v-if="lifecycleStatus?.color"
            class="w-4 h-4 rounded-full mr-3"
            :style="{ backgroundColor: lifecycleStatus.color }"
          ></div>
          <h1 class="text-3xl font-bold text-gray-900">{{ lifecycleStatus?.display_name || lifecycleStatus?.name }}</h1>
        </div>
        <p class="mt-2 text-gray-600">{{ lifecycleStatus?.description || 'No description provided' }}</p>
      </div>
      <div class="flex space-x-3">
        <router-link
          v-if="hasPermission('lifecycle_status:update') && !lifecycleStatus?.is_system"
          :to="`/lifecycle-status/${$route.params.id}/edit`"
          class="btn btn-primary"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
          </svg>
          Edit
        </router-link>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading lifecycle status details...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="text-center py-12">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
        <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900">Error loading details</h3>
      <p class="mt-2 text-gray-500">{{ error }}</p>
      <div class="mt-6">
        <router-link to="/lifecycle-status" class="btn btn-outline">Back to Lifecycle Statuses</router-link>
      </div>
    </div>

    <!-- Details -->
    <div v-else-if="lifecycleStatus" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Details -->
      <div class="lg:col-span-2">
        <div class="bg-white shadow rounded-lg">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Status Details</h3>
          </div>
          <div class="card-body">
            <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
              <div>
                <dt class="text-sm font-medium text-gray-500">Name</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ lifecycleStatus.name }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Display Name</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ lifecycleStatus.display_name }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Sort Order</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ lifecycleStatus.sort_order }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Status</dt>
                <dd class="mt-1">
                  <span :class="lifecycleStatus.is_active ? 'badge badge-success' : 'badge badge-warning'">
                    {{ lifecycleStatus.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Type</dt>
                <dd class="mt-1">
                  <span v-if="lifecycleStatus.is_system" class="badge badge-gray">System</span>
                  <span v-else class="badge badge-blue">Custom</span>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Amortization Behavior</dt>
                <dd class="mt-1">
                  <span :class="getAmortizationBehaviorBadgeClass(lifecycleStatus.amortization_behavior)">
                    {{ formatAmortizationBehavior(lifecycleStatus.amortization_behavior) }}
                  </span>
                </dd>
              </div>
              <div v-if="lifecycleStatus.icon">
                <dt class="text-sm font-medium text-gray-500">Icon</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ lifecycleStatus.icon }}</dd>
              </div>
              <div v-if="lifecycleStatus.color">
                <dt class="text-sm font-medium text-gray-500">Color</dt>
                <dd class="mt-1 flex items-center">
                  <div
                    class="w-6 h-6 rounded border border-gray-300 mr-2"
                    :style="{ backgroundColor: lifecycleStatus.color }"
                  ></div>
                  <span class="text-sm text-gray-900">{{ lifecycleStatus.color }}</span>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Created</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ formatDate(lifecycleStatus.created_at) }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Last Updated</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ formatDate(lifecycleStatus.updated_at) }}</dd>
              </div>
            </dl>
            <div v-if="lifecycleStatus.description" class="mt-6">
              <dt class="text-sm font-medium text-gray-500">Description</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ lifecycleStatus.description }}</dd>
            </div>
          </div>
        </div>
      </div>

      <!-- Usage Statistics -->
      <div class="lg:col-span-1">
        <div v-if="usageStats" class="bg-white shadow rounded-lg">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Usage Statistics</h3>
          </div>
          <div class="card-body">
            <div class="space-y-4">
              <div class="flex items-center">
                <div class="p-2 bg-blue-100 rounded-lg">
                  <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
                  </svg>
                </div>
                <div class="ml-4">
                  <p class="text-sm font-medium text-gray-600">CIs with this status</p>
                  <p class="text-lg font-semibold text-gray-900">
                    {{ usageStats.status_usage?.find(u => u.lifecycle_status.id === lifecycleStatus.id)?.usage_count || 0 }}
                  </p>
                </div>
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
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLifecycleStatusStore } from '@/stores/lifecycleStatus'

const route = useRoute()
const authStore = useAuthStore()
const lifecycleStatusStore = useLifecycleStatusStore()

const loading = ref(false)
const error = ref('')
const lifecycleStatus = ref<any>(null)
const usageStats = ref<any>(null)

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const formatAmortizationBehavior = (behavior: string) => {
  const behaviors: Record<string, string> = {
    'pending': 'Pending',
    'active': 'Active',
    'terminal': 'Terminal'
  }
  return behaviors[behavior] || behavior
}

const getAmortizationBehaviorBadgeClass = (behavior: string) => {
  const classes: Record<string, string> = {
    'pending': 'badge badge-warning',
    'active': 'badge badge-success',
    'terminal': 'badge badge-gray'
  }
  return classes[behavior] || 'badge badge-gray'
}

const loadLifecycleStatus = async () => {
  try {
    loading.value = true
    error.value = ''
    lifecycleStatus.value = await lifecycleStatusStore.getLifecycleStatus(route.params.id as string)
  } catch (err: any) {
    console.error('Failed to load lifecycle status:', err)
    error.value = err.response?.data?.error || 'Failed to load lifecycle status'
  } finally {
    loading.value = false
  }
}

const loadUsageStats = async () => {
  try {
    usageStats.value = await lifecycleStatusStore.getLifecycleStatusUsage()
  } catch (err) {
    console.error('Failed to load usage stats:', err)
  }
}

onMounted(async () => {
  await Promise.all([
    loadLifecycleStatus(),
    loadUsageStats()
  ])
})
</script>
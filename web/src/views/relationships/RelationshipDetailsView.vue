<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <nav class="flex" aria-label="Breadcrumb">
          <ol class="flex items-center space-x-4">
            <li>
              <router-link to="/relationships" class="text-gray-500 hover:text-gray-700">
                Relationships
              </router-link>
            </li>
            <li>
              <div class="flex items-center">
                <svg class="flex-shrink-0 h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="ml-4 text-sm font-medium text-gray-900">Relationship Details</span>
              </div>
            </li>
          </ol>
        </nav>
        <h1 class="text-3xl font-bold text-gray-900 mt-2">Relationship Details</h1>
        <p class="mt-2 text-gray-600">View relationship information between configuration items</p>
      </div>
      <div class="flex space-x-3">
        <router-link
          v-if="hasPermission('relationship:update')"
          :to="`/relationships/${relationship?.id}/edit`"
          class="btn btn-primary"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
          </svg>
          Edit Relationship
        </router-link>
        <router-link to="/relationships" class="btn btn-outline">
          Back to Relationships
        </router-link>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading relationship details...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="text-center py-12">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
        <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Error Loading Relationship</h3>
      <p class="text-gray-500 mb-4">{{ error }}</p>
      <router-link to="/relationships" class="btn btn-primary">
        Back to Relationships
      </router-link>
    </div>

    <!-- Relationship Details -->
    <div v-else-if="relationship" class="max-w-4xl mx-auto">
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Relationship Information</h3>
        </div>
        <div class="card-body">
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- Source CI -->
            <div>
              <h4 class="text-sm font-medium text-gray-900 mb-3">Source Configuration Item</h4>
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center justify-between">
                  <div>
                    <router-link
                      :to="`/ci/${relationship.source_id}`"
                      class="text-lg font-medium text-blue-600 hover:text-blue-900"
                    >
                      {{ sourceCI?.name || 'Loading...' }}
                    </router-link>
                    <p class="text-sm text-gray-500 mt-1">{{ sourceCI?.ci_type || 'Unknown Type' }}</p>
                    <p class="text-xs text-gray-400 mt-1">ID: {{ relationship.source_id }}</p>
                  </div>
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                  </svg>
                </div>
              </div>
            </div>

            <!-- Target CI -->
            <div>
              <h4 class="text-sm font-medium text-gray-900 mb-3">Target Configuration Item</h4>
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center justify-between">
                  <div>
                    <router-link
                      :to="`/ci/${relationship.target_id}`"
                      class="text-lg font-medium text-blue-600 hover:text-blue-900"
                    >
                      {{ targetCI?.name || 'Loading...' }}
                    </router-link>
                    <p class="text-sm text-gray-500 mt-1">{{ targetCI?.ci_type || 'Unknown Type' }}</p>
                    <p class="text-xs text-gray-400 mt-1">ID: {{ relationship.target_id }}</p>
                  </div>
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Relationship Type -->
          <div class="mt-6">
            <h4 class="text-sm font-medium text-gray-900 mb-3">Relationship Type</h4>
            <div class="flex items-center">
              <span class="badge badge-info text-lg px-4 py-2">
                {{ formatRelationshipType(relationship.relationship_type) }}
              </span>
            </div>
          </div>

          <!-- Attributes -->
          <div v-if="relationship.attributes && Object.keys(relationship.attributes).length > 0" class="mt-6">
            <h4 class="text-sm font-medium text-gray-900 mb-3">Attributes</h4>
            <div class="bg-gray-50 rounded-lg p-4">
              <dl class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div v-for="(value, key) in relationship.attributes" :key="key" class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">{{ key }}</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ value }}</dd>
                </div>
              </dl>
            </div>
          </div>

          <!-- Timestamps -->
          <div class="mt-6 pt-6 border-t border-gray-200">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <h4 class="text-sm font-medium text-gray-900 mb-1">Created</h4>
                <p class="text-sm text-gray-500">{{ formatDate(relationship.created_at) }}</p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-900 mb-1">Last Updated</h4>
                <p class="text-sm text-gray-500">{{ formatDate(relationship.updated_at) }}</p>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="mt-6 pt-6 border-t border-gray-200">
            <div class="flex justify-between">
              <div class="flex space-x-3">
                <router-link
                  v-if="hasPermission('relationship:update')"
                  :to="`/relationships/${relationship.id}/edit`"
                  class="btn btn-primary"
                >
                  <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                  </svg>
                  Edit Relationship
                </router-link>
                <button
                  v-if="hasPermission('relationship:delete')"
                  @click="confirmDelete"
                  class="btn btn-danger"
                >
                  <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                  </svg>
                  Delete Relationship
                </button>
              </div>
              <router-link to="/relationships" class="btn btn-outline">
                Back to Relationships
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { relationshipAPI, ciAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface Relationship {
  id: string
  source_id: string
  target_id: string
  relationship_type: string
  attributes?: Record<string, any>
  created_at: string
  updated_at: string
}

interface CI {
  id: string
  name: string
  ci_type: string
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const error = ref('')
const relationship = ref<Relationship | null>(null)
const sourceCI = ref<CI | null>(null)
const targetCI = ref<CI | null>(null)

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatRelationshipType = (type: string) => {
  return type.split('_').map(word =>
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ')
}

const loadRelationshipDetails = async () => {
  const relationshipId = route.params.id as string
  if (!relationshipId) {
    error.value = 'Invalid relationship ID'
    return
  }

  loading.value = true
  error.value = ''

  try {
    // Load relationship details
    const relationshipResponse = await relationshipAPI.get(relationshipId)
    relationship.value = relationshipResponse.data

    // Load source and target CI details
    await Promise.all([
      loadCIDetails(relationship.value.source_id, 'source'),
      loadCIDetails(relationship.value.target_id, 'target')
    ])
  } catch (err: any) {
    console.error('Failed to load relationship details:', err)
    error.value = err.response?.data?.error || 'Failed to load relationship details'
  } finally {
    loading.value = false
  }
}

const loadCIDetails = async (ciId: string, type: 'source' | 'target') => {
  try {
    const ciResponse = await ciAPI.get(ciId)
    if (type === 'source') {
      sourceCI.value = ciResponse.data
    } else {
      targetCI.value = ciResponse.data
    }
  } catch (err) {
    console.error(`Failed to load ${type} CI details:`, err)
  }
}

const confirmDelete = async () => {
  if (!relationship.value) return

  if (confirm('Are you sure you want to delete this relationship? This action cannot be undone.')) {
    try {
      await relationshipAPI.delete(relationship.value.id)
      showSuccessToast('Relationship deleted successfully')
      router.push('/relationships')
    } catch (err: any) {
      console.error('Failed to delete relationship:', err)
      const message = err.response?.data?.error || 'Failed to delete relationship'
      showErrorToast(message)
    }
  }
}

onMounted(() => {
  loadRelationshipDetails()
})
</script>
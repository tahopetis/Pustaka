<template>
  <div class="px-4 py-6 sm:px-0">
    <div class="max-w-7xl mx-auto">
      <!-- Page header -->
      <div class="mb-8 flex justify-between items-center">
        <div>
          <nav class="flex" aria-label="Breadcrumb">
            <ol class="flex items-center space-x-4">
              <li>
                <router-link to="/ci" class="text-gray-500 hover:text-gray-700">
                  Configuration Items
                </router-link>
              </li>
              <li>
                <div class="flex items-center">
                  <svg class="flex-shrink-0 h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                    <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                  </svg>
                  <span class="ml-4 text-sm font-medium text-gray-900">{{ ci?.name }}</span>
                </div>
              </li>
            </ol>
          </nav>
          <h1 class="text-3xl font-bold text-gray-900 mt-2">{{ ci?.name }}</h1>
          <p class="mt-2 text-gray-600">{{ ci?.ci_type }} Configuration Item</p>
        </div>
        <div class="flex space-x-3">
          <router-link
            v-if="hasPermission('ci:update')"
            :to="`/ci/${$route.params.id}/edit`"
            class="btn btn-outline"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
            </svg>
            Edit
          </router-link>
          <button
            v-if="hasPermission('ci:delete')"
            @click="confirmDelete"
            class="btn btn-danger"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
            Delete
          </button>
        </div>
      </div>

      <div v-if="loading" class="text-center py-12">
        <div class="spinner w-8 h-8 mx-auto mb-4"></div>
        <p class="text-gray-500">Loading configuration item...</p>
      </div>

      <div v-else-if="ci" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Main Content -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Basic Information -->
          <div class="card">
            <div class="card-header">
              <h3 class="text-lg leading-6 font-medium text-gray-900">Basic Information</h3>
            </div>
            <div class="card-body">
              <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <div>
                  <dt class="text-sm font-medium text-gray-500">Name</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ ci.name }}</dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">Type</dt>
                  <dd class="mt-1">
                    <span class="badge badge-info">{{ ci.ci_type }}</span>
                  </dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">Created</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ formatDate(ci.created_at) }}</dd>
                </div>
                <div>
                  <dt class="text-sm font-medium text-gray-500">Last Updated</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ formatDate(ci.updated_at) }}</dd>
                </div>
              </dl>
            </div>
          </div>

          <!-- Attributes -->
          <div class="card">
            <div class="card-header">
              <h3 class="text-lg leading-6 font-medium text-gray-900">Attributes</h3>
            </div>
            <div class="card-body">
              <div v-if="!ciType || !ci" class="text-center py-8">
                <p class="text-gray-500">Loading attributes...</p>
              </div>
              <div v-else-if="Object.keys(ci.attributes).length === 0" class="text-center py-8">
                <p class="text-gray-500">No attributes defined</p>
              </div>
              <FlexibleAttributeDisplay
                v-else
                :ci="ci"
                :ci-type="ciType"
              />
            </div>
          </div>

          <!-- Relationships -->
          <div class="card">
            <div class="card-header">
              <h3 class="text-lg leading-6 font-medium text-gray-900">Relationships</h3>
            </div>
            <div class="card-body">
              <div v-if="loadingRelationships" class="text-center py-8">
                <div class="spinner w-6 h-6 mx-auto mb-2"></div>
                <p class="text-gray-500 text-sm">Loading relationships...</p>
              </div>
              <div v-else-if="relationships.length === 0" class="text-center py-8">
                <p class="text-gray-500">No relationships found</p>
                <router-link
                  v-if="hasPermission('relationship:create')"
                  to="/relationships"
                  class="text-blue-600 hover:text-blue-800 text-sm mt-2 inline-block"
                >
                  Create Relationship
                </router-link>
              </div>
              <div v-else class="space-y-4">
                <div v-for="rel in relationships" :key="rel.id" class="border border-gray-200 rounded-lg p-4">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-3">
                      <div class="flex-shrink-0">
                        <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                          <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
                          </svg>
                        </div>
                      </div>
                      <div>
                        <p class="text-sm font-medium text-gray-900">{{ rel.relationship_type }}</p>
                        <p class="text-xs text-gray-500">
                          {{ rel.source_ci?.name }} → {{ rel.target_ci?.name }}
                        </p>
                      </div>
                    </div>
                    <div class="flex space-x-2">
                      <router-link
                        v-if="hasPermission('relationship:update')"
                        :to="`/relationships/${rel.id}`"
                        class="text-indigo-600 hover:text-indigo-900"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                        </svg>
                      </router-link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Sidebar -->
        <div class="space-y-6">
          <!-- Tags -->
          <div class="card">
            <div class="card-header">
              <h3 class="text-lg leading-6 font-medium text-gray-900">Tags</h3>
            </div>
            <div class="card-body">
              <div v-if="ci.tags.length === 0" class="text-center py-4">
                <p class="text-gray-500 text-sm">No tags</p>
              </div>
              <div v-else class="flex flex-wrap gap-2">
                <span v-for="tag in ci.tags" :key="tag" class="badge badge-success">
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>

          <!-- Quick Actions -->
          <div class="card">
            <div class="card-header">
              <h3 class="text-lg leading-6 font-medium text-gray-900">Quick Actions</h3>
            </div>
            <div class="card-body space-y-3">
              <router-link
                v-if="hasPermission('relationship:create')"
                :to="`/relationships/new?source_id=${ci.id}`"
                class="block w-full text-center btn btn-outline"
              >
                Add Relationship
              </router-link>
              <router-link
                to="/graph"
                class="block w-full text-center btn btn-outline"
              >
                View in Graph
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
import { ciAPI, relationshipAPI, ciTypeAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'
import FlexibleAttributeDisplay from '@/components/ci/FlexibleAttributeDisplay.vue'
import type { CI, Relationship, CIType } from '@/types/ci'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const loadingRelationships = ref(false)
const ci = ref<CI | null>(null)
const ciType = ref<CIType | null>(null)
const relationships = ref<Relationship[]>([])

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const loadCI = async () => {
  if (!route.params.id) return

  loading.value = true
  try {
    const response = await ciAPI.get(route.params.id as string)
    ci.value = response.data
  } catch (error: any) {
    console.error('Failed to load CI:', error)
    const message = error.response?.data?.error || 'Failed to load configuration item'
    showErrorToast(message)
    router.push('/ci')
  } finally {
    loading.value = false
  }
}

const loadCIType = async () => {
  if (!ci.value) return

  try {
    const response = await ciTypeAPI.get(ci.value.ci_type)
    ciType.value = response.data
  } catch (error) {
    console.error('Failed to load CI type:', error)
  }
}

const loadRelationships = async () => {
  if (!ci.value) return

  loadingRelationships.value = true
  try {
    const response = await relationshipAPI.list({
      source_id: ci.value.id,
      limit: 50
    })
    relationships.value = response.data.relationships || []
  } catch (error) {
    console.error('Failed to load relationships:', error)
  } finally {
    loadingRelationships.value = false
  }
}

const confirmDelete = async () => {
  if (!ci.value) return

  if (confirm(`Are you sure you want to delete "${ci.value.name}"? This action cannot be undone.`)) {
    try {
      await ciAPI.delete(ci.value.id)
      showSuccessToast('Configuration item deleted successfully')
      router.push('/ci')
    } catch (error: any) {
      console.error('Failed to delete CI:', error)
      const message = error.response?.data?.error || 'Failed to delete configuration item'
      showErrorToast(message)
    }
  }
}

onMounted(async () => {
  await loadCI()
  await loadCIType()
  await loadRelationships()
})
</script>
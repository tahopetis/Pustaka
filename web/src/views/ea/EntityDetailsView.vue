<template>
  <div class="page-container page-content">
    <!-- Loading State -->
    <div v-if="eaStore.loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading entity details...</p>
    </div>

    <!-- Entity Details -->
    <div v-else-if="entity">
      <!-- Breadcrumbs and Header -->
      <div class="mb-6">
        <nav class="flex mb-4" aria-label="Breadcrumb">
          <ol class="flex items-center space-x-2">
            <li>
              <router-link to="/entities/business" class="text-gray-400 hover:text-gray-500">
                <svg class="flex-shrink-0 h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
                </svg>
              </router-link>
            </li>
            <li>
              <div class="flex items-center">
                <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <router-link :to="`/entities/${entity.domain}`" class="ml-2 text-sm font-medium text-gray-500 hover:text-gray-700">
                  {{ domainDisplay }}
                </router-link>
              </div>
            </li>
            <li>
              <div class="flex items-center">
                <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="ml-2 text-sm font-medium text-gray-900">{{ entity.name }}</span>
              </div>
            </li>
          </ol>
        </nav>

        <div class="flex items-center justify-between">
          <div>
            <h1 class="page-title">{{ entity.name }}</h1>
            <p class="page-subtitle">{{ entity.ci_type_display }}</p>
          </div>
          <div class="flex space-x-2">
            <router-link
              v-if="hasPermission('ea:update')"
              :to="`/entities/${entity.domain}/${entityId}/edit`"
              class="btn btn-outline"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
              </svg>
              Edit
            </router-link>
            <button
              v-if="hasPermission('ea:delete')"
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
      </div>

      <!-- Tabs -->
      <div class="border-b border-gray-200 mb-6">
        <nav class="-mb-px flex space-x-8">
          <button
            @click="activeTab = 'overview'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
            :class="activeTab === 'overview' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          >
            Overview
          </button>
          <button
            @click="activeTab = 'attributes'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
            :class="activeTab === 'attributes' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          >
            Attributes
          </button>
          <button
            @click="activeTab = 'relationships'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
            :class="activeTab === 'relationships' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          >
            Relationships
          </button>
          <button
            @click="activeTab = 'audit'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
            :class="activeTab === 'audit' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          >
            Audit History
          </button>
        </nav>
      </div>

      <!-- Overview Tab -->
      <div v-if="activeTab === 'overview'" class="space-y-6">
        <!-- Entity Info Card -->
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Entity Information</h3>
          <dl class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div>
              <dt class="text-sm font-medium text-gray-500">Name</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ entity.name }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">CI Type</dt>
              <dd class="mt-1">
                <span class="badge badge-info">{{ entity.ci_type_display }}</span>
              </dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Domain</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ entity.domain }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Lifecycle Status</dt>
              <dd class="mt-1">
                <span v-if="entity.lifecycle_status_display" class="badge">
                  {{ entity.lifecycle_status_display }}
                </span>
                <span v-else class="text-sm text-gray-400">No status</span>
              </dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Owner</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ entity.owner_name || 'Not assigned' }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Team</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ entity.team_name || 'Not assigned' }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Data Quality Score</dt>
              <dd class="mt-1">
                <span
                  class="font-medium"
                  :class="dataQualityColorClass"
                >
                  {{ entity.data_quality_score }}%
                </span>
              </dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Created</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatDate(entity.created_at) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">Last Updated</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatDate(entity.updated_at) }}</dd>
            </div>
          </dl>
        </div>

        <!-- Tags -->
        <div v-if="entity.tags && entity.tags.length > 0" class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Tags</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="tag in entity.tags"
              :key="tag"
              class="badge badge-success"
            >
              {{ tag }}
            </span>
          </div>
        </div>
      </div>

      <!-- Attributes Tab -->
      <div v-else-if="activeTab === 'attributes'" class="space-y-6">
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Attributes</h3>
          <FlexibleAttributeDisplay :attributes="entity.attributes" />
        </div>
      </div>

      <!-- Relationships Tab -->
      <div v-else-if="activeTab === 'relationships'" class="space-y-6">
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Relationships</h3>
          <p class="text-sm text-gray-500 mb-4">
            This feature is coming in a future phase. Stay tuned for relationship management capabilities.
          </p>
        </div>
      </div>

      <!-- Audit History Tab -->
      <div v-else-if="activeTab === 'audit'" class="space-y-6">
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Audit History</h3>
          <p class="text-sm text-gray-500 mb-4">
            Audit log viewing is coming in a future phase. Stay tuned for comprehensive audit trail capabilities.
          </p>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else class="text-center py-12">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">Entity not found</h3>
      <p class="mt-1 text-sm text-gray-500">
        The requested entity could not be found.
      </p>
      <div class="mt-6">
        <router-link
          :to="`/entities/${route.params.domain || 'business'}`"
          class="btn btn-primary"
        >
          Back to List
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useEaStore } from '@/stores/ea'
import FlexibleAttributeDisplay from '@/components/ci/FlexibleAttributeDisplay.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const eaStore = useEaStore()

const activeTab = ref('overview')
const entityId = computed(() => route.params.id as string)

const entity = computed(() => eaStore.currentEntity)

const domainDisplay = computed(() => {
  if (!entity.value) return 'Entity'
  const domainMap: Record<string, string> = {
    strategy: 'Strategy',
    business: 'Business',
    application: 'Application',
    data: 'Data',
    technology: 'Technology',
    infrastructure: 'Infrastructure',
    security: 'Security',
    governance: 'Governance'
  }
  return domainMap[entity.value.domain] || 'Entity'
})

const dataQualityColorClass = computed(() => {
  if (!entity.value) return ''
  const score = entity.value.data_quality_score
  return score >= 80 ? 'text-green-600' : score >= 60 ? 'text-yellow-600' : 'text-red-600'
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const confirmDelete = async () => {
  if (confirm(`Are you sure you want to delete "${entity.value?.name}"? This action cannot be undone.`)) {
    try {
      await eaStore.deleteEntity(entityId.value)
      router.push(`/entities/${entity.value?.domain || 'business'}`)
    } catch (error: any) {
      console.error('Failed to delete entity:', error)
      alert(error.response?.data?.error || 'Failed to delete entity')
    }
  }
}

onMounted(async () => {
  await eaStore.fetchEntity(entityId.value)
})
</script>

<template>
  <div class="page-container page-content">
    <div class="page-header flex justify-between items-center">
      <div>
        <nav class="flex" aria-label="Breadcrumb">
          <ol class="flex items-center space-x-4">
            <li>
              <router-link to="/relationship-types" class="text-gray-500 hover:text-gray-700">
                Relationship Types
              </router-link>
            </li>
            <li>
              <div class="flex items-center">
                <svg class="flex-shrink-0 h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="ml-4 text-sm font-medium text-gray-900">Relationship Type Details</span>
              </div>
            </li>
          </ol>
        </nav>
        <h1 class="text-3xl font-bold text-gray-900 mt-2">{{ relationshipType?.display_name || relationshipType?.name }}</h1>
        <p class="mt-2 text-gray-600">{{ relationshipType?.description || 'No description provided' }}</p>
      </div>
      <div class="flex space-x-3">
        <router-link
          v-if="hasPermission('relationship_type:update') && !relationshipType?.is_system"
          :to="`/relationship-types/${$route.params.id}/edit`"
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
      <p class="text-gray-500">Loading relationship type details...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="text-center py-12">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
        <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Error Loading Relationship Type</h3>
      <p class="text-gray-500 mb-4">{{ error }}</p>
      <router-link to="/relationship-types" class="btn btn-primary">
        Back to Relationship Types
      </router-link>
    </div>

    <!-- Relationship Type Details -->
    <div v-else-if="relationshipType" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Details -->
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
                <dd class="mt-1 text-sm text-gray-900">
                  <div class="flex items-center">
                    <code class="bg-gray-100 px-2 py-1 rounded text-xs">{{ relationshipType.name }}</code>
                    <span v-if="relationshipType.is_system" class="ml-2 badge badge-gray text-xs">System</span>
                  </div>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Status</dt>
                <dd class="mt-1">
                  <span :class="relationshipType.is_active ? 'badge badge-success' : 'badge badge-warning'">
                    {{ relationshipType.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Forward Label</dt>
                <dd class="mt-1">
                  <span
                    v-if="relationshipType.color"
                    class="badge"
                    :style="{ backgroundColor: relationshipType.color + '20', color: relationshipType.color, borderColor: relationshipType.color }"
                  >
                    {{ relationshipType.forward_label }}
                  </span>
                  <span v-else class="badge badge-info">
                    {{ relationshipType.forward_label }}
                  </span>
                </dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500">Reverse Label</dt>
                <dd class="mt-1">
                  <span
                    v-if="relationshipType.color"
                    class="badge"
                    :style="{ backgroundColor: relationshipType.color + '20', color: relationshipType.color, borderColor: relationshipType.color }"
                  >
                    {{ relationshipType.reverse_label }}
                  </span>
                  <span v-else class="badge badge-secondary">
                    {{ relationshipType.reverse_label }}
                  </span>
                </dd>
              </div>
              <div v-if="relationshipType.category">
                <dt class="text-sm font-medium text-gray-500">Category</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ relationshipType.category }}</dd>
              </div>
              <div v-if="relationshipType.bidirectional">
                <dt class="text-sm font-medium text-gray-500">Type</dt>
                <dd class="mt-1">
                  <span class="badge badge-gray">Bidirectional</span>
                </dd>
              </div>
              <div v-if="relationshipType.color">
                <dt class="text-sm font-medium text-gray-500">Color</dt>
                <dd class="mt-1 flex items-center">
                  <div
                    class="w-6 h-6 rounded border border-gray-300 mr-2"
                    :style="{ backgroundColor: relationshipType.color }"
                  ></div>
                  <code class="text-sm">{{ relationshipType.color }}</code>
                </dd>
              </div>
              <div v-if="relationshipType.icon">
                <dt class="text-sm font-medium text-gray-500">Icon</dt>
                <dd class="mt-1">
                  <code class="text-sm">{{ relationshipType.icon }}</code>
                </dd>
              </div>
              <div class="sm:col-span-2">
                <dt class="text-sm font-medium text-gray-500">Description</dt>
                <dd class="mt-1 text-sm text-gray-900">
                  {{ relationshipType.description || 'No description provided' }}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Relationship Constraints -->
        <div v-if="hasConstraints" class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Relationship Constraints</h3>
          </div>
          <div class="card-body">
            <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
              <div v-if="relationshipType.cardinality_source">
                <dt class="text-sm font-medium text-gray-500">Source Cardinality</dt>
                <dd class="mt-1 text-sm text-gray-900">
                  <span class="badge badge-outline">{{ getCardinalityLabel(relationshipType.cardinality_source) }}</span>
                </dd>
              </div>
              <div v-if="relationshipType.cardinality_target">
                <dt class="text-sm font-medium text-gray-500">Target Cardinality</dt>
                <dd class="mt-1 text-sm text-gray-900">
                  <span class="badge badge-outline">{{ getCardinalityLabel(relationshipType.cardinality_target) }}</span>
                </dd>
              </div>
              <div v-if="relationshipType.allowed_source_types && relationshipType.allowed_source_types.length > 0" class="sm:col-span-2">
                <dt class="text-sm font-medium text-gray-500">Allowed Source Types</dt>
                <dd class="mt-1">
                  <div class="flex flex-wrap gap-2">
                    <span v-for="type in relationshipType.allowed_source_types" :key="type" class="badge badge-outline text-xs">
                      {{ type }}
                    </span>
                  </div>
                </dd>
              </div>
              <div v-if="relationshipType.allowed_target_types && relationshipType.allowed_target_types.length > 0" class="sm:col-span-2">
                <dt class="text-sm font-medium text-gray-500">Allowed Target Types</dt>
                <dd class="mt-1">
                  <div class="flex flex-wrap gap-2">
                    <span v-for="type in relationshipType.allowed_target_types" :key="type" class="badge badge-outline text-xs">
                      {{ type }}
                    </span>
                  </div>
                </dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Usage Examples -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Usage Examples</h3>
          </div>
          <div class="card-body">
            <div class="space-y-4">
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center space-x-4">
                  <div class="flex-1 text-right">
                    <p class="text-sm font-medium text-gray-900">Server A</p>
                    <p class="text-xs text-gray-500">Source CI</p>
                  </div>
                  <div class="flex items-center space-x-2">
                    <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path>
                    </svg>
                    <span
                      v-if="relationshipType.color"
                      class="badge text-sm"
                      :style="{ backgroundColor: relationshipType.color + '20', color: relationshipType.color, borderColor: relationshipType.color }"
                    >
                      {{ relationshipType.forward_label }}
                    </span>
                    <span v-else class="badge badge-info text-sm">
                      {{ relationshipType.forward_label }}
                    </span>
                  </div>
                  <div class="flex-1">
                    <p class="text-sm font-medium text-gray-900">Database B</p>
                    <p class="text-xs text-gray-500">Target CI</p>
                  </div>
                </div>
                <p class="text-xs text-gray-600 mt-2 text-center">Server A <strong>{{ relationshipType.forward_label.toLowerCase() }}</strong> Database B</p>
              </div>

              <div v-if="!relationshipType.bidirectional" class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center space-x-4">
                  <div class="flex-1 text-right">
                    <p class="text-sm font-medium text-gray-900">Database B</p>
                    <p class="text-xs text-gray-500">Source CI</p>
                  </div>
                  <div class="flex items-center space-x-2">
                    <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path>
                    </svg>
                    <span
                      v-if="relationshipType.color"
                      class="badge text-sm"
                      :style="{ backgroundColor: relationshipType.color + '20', color: relationshipType.color, borderColor: relationshipType.color }"
                    >
                      {{ relationshipType.reverse_label }}
                    </span>
                    <span v-else class="badge badge-secondary text-sm">
                      {{ relationshipType.reverse_label }}
                    </span>
                  </div>
                  <div class="flex-1">
                    <p class="text-sm font-medium text-gray-900">Server A</p>
                    <p class="text-xs text-gray-500">Target CI</p>
                  </div>
                </div>
                <p class="text-xs text-gray-600 mt-2 text-center">Database B <strong>{{ relationshipType.reverse_label.toLowerCase() }}</strong> Server A</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Relationships -->
        <div class="card">
          <div class="card-header flex justify-between items-center">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Recent Relationships</h3>
            <router-link to="/relationships" class="text-sm text-blue-600 hover:text-blue-900">
              View All
            </router-link>
          </div>
          <div class="card-body">
            <div class="text-center py-8 text-gray-500">
              <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              <p class="mt-2 text-sm">Recent relationships will be shown here</p>
              <p class="text-xs text-gray-400">Integration with relationship API needed</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Stats Widget -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Usage Statistics</h3>
          </div>
          <div class="card-body">
            <div class="space-y-4">
              <div class="flex justify-between items-center">
                <span class="text-sm text-gray-500">Total Relationships</span>
                <span class="text-lg font-semibold text-gray-900">{{ stats?.total_relationships || 0 }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-sm text-gray-500">Usage Rank</span>
                <span class="text-lg font-semibold text-gray-900">#{{ getUsageRank() }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-sm text-gray-500">Created</span>
                <span class="text-sm text-gray-900">{{ formatDate(relationshipType.created_at) }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-sm text-gray-500">Last Updated</span>
                <span class="text-sm text-gray-900">{{ formatDate(relationshipType.updated_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Metadata -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Metadata</h3>
          </div>
          <div class="card-body">
            <dl class="space-y-4">
              <div>
                <dt class="text-sm font-medium text-gray-500">System Type</dt>
                <dd class="mt-1">
                  <span :class="relationshipType.is_system ? 'badge badge-gray' : 'badge badge-outline'">
                    {{ relationshipType.is_system ? 'Yes' : 'No' }}
                  </span>
                </dd>
              </div>
              <div v-if="relationshipType.created_by">
                <dt class="text-sm font-medium text-gray-500">Created By</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ relationshipType.created_by }}</dd>
              </div>
              <div v-if="relationshipType.updated_by">
                <dt class="text-sm font-medium text-gray-500">Updated By</dt>
                <dd class="mt-1 text-sm text-gray-900">{{ relationshipType.updated_by }}</dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Actions -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Actions</h3>
          </div>
          <div class="card-body space-y-3">
            <router-link
              v-if="hasPermission('relationship_type:update') && !relationshipType.is_system"
              :to="`/relationship-types/${relationshipType.id}/edit`"
              class="btn btn-outline w-full"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
              </svg>
              Edit Relationship Type
            </router-link>

            <router-link
              to="/relationships/new"
              class="btn btn-outline w-full"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              Create Relationship
            </router-link>

            <button
              v-if="hasPermission('relationship_type:delete') && !relationshipType.is_system"
              @click="confirmDelete"
              class="btn btn-outline w-full text-red-600 hover:text-red-700 hover:bg-red-50"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
              </svg>
              Delete Relationship Type
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useRelationshipTypeStore } from '@/stores/relationshipTypes'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const relationshipTypeStore = useRelationshipTypeStore()

const loading = ref(false)
const error = ref('')
const stats = ref<any>(null)

const relationshipType = computed(() => {
  const typeId = route.params.id as string
  return relationshipTypeStore.getRelationshipTypeById(typeId)
})

const hasConstraints = computed(() => {
  return relationshipType.value && (
    relationshipType.value.cardinality_source ||
    relationshipType.value.cardinality_target ||
    (relationshipType.value.allowed_source_types && relationshipType.value.allowed_source_types.length > 0) ||
    (relationshipType.value.allowed_target_types && relationshipType.value.allowed_target_types.length > 0)
  )
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const getCardinalityLabel = (cardinality: string) => {
  const labels: Record<string, string> = {
    '1': 'Exactly one',
    '0..1': 'Zero or one',
    '*': 'Zero or more',
    '1..*': 'One or more'
  }
  return labels[cardinality] || cardinality
}

const getUsageRank = () => {
  if (!stats.value?.most_used_types) return 0

  const index = stats.value.most_used_types.findIndex((item: any) =>
    item.type.id === relationshipType.value?.id
  )

  return index > -1 ? index + 1 : 0
}

const loadRelationshipType = async () => {
  const typeId = route.params.id as string
  if (!typeId) return

  loading.value = true
  try {
    // Check if already in store
    const existing = relationshipTypeStore.getRelationshipTypeById(typeId)
    if (existing) {
      return
    }

    // Load from API
    await relationshipTypeStore.loadRelationshipTypes()

    if (!relationshipTypeStore.getRelationshipTypeById(typeId)) {
      error.value = 'Relationship type not found'
    }
  } catch (err: any) {
    console.error('Failed to load relationship type:', err)
    error.value = err.response?.data?.error || 'Failed to load relationship type'
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    await relationshipTypeStore.loadStats()
    stats.value = relationshipTypeStore.stats
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
}

const confirmDelete = async () => {
  if (!relationshipType.value) return

  if (confirm(`Are you sure you want to delete the relationship type "${relationshipType.value.name}"? This action cannot be undone and may affect existing relationships.`)) {
    try {
      await relationshipTypeStore.deleteRelationshipType(relationshipType.value.id)
      showSuccessToast('Relationship type deleted successfully')
      router.push('/relationship-types')
    } catch (error: any) {
      console.error('Failed to delete relationship type:', error)
      const message = error.response?.data?.error || 'Failed to delete relationship type'
      showErrorToast(message)
    }
  }
}

onMounted(async () => {
  if (!hasPermission('relationship_type:read')) {
    showErrorToast('You do not have permission to view relationship types')
    router.push('/relationship-types')
    return
  }

  await loadRelationshipType()
  await loadStats()
})
</script>
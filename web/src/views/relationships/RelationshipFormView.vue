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
                <span class="ml-4 text-sm font-medium text-gray-900">{{ isEditing ? 'Edit Relationship' : 'Create Relationship' }}</span>
              </div>
            </li>
          </ol>
        </nav>
        <h1 class="text-3xl font-bold text-gray-900 mt-2">{{ isEditing ? 'Edit Relationship' : 'Create Relationship' }}</h1>
        <p class="mt-2 text-gray-600">{{ isEditing ? 'Update relationship attributes' : 'Create a new relationship between configuration items' }}</p>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading configuration items...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="text-center py-12">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
        <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Error Loading Data</h3>
      <p class="text-gray-500 mb-4">{{ error }}</p>
      <router-link to="/relationships" class="btn btn-primary">
        Back to Relationships
      </router-link>
    </div>

    <!-- Form -->
    <div v-else class="max-w-3xl mx-auto">
      <!-- Relationship Info Display (Edit Mode) -->
      <div v-if="isEditing && existingRelationship" class="card mb-6">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Relationship Information</h3>
        </div>
        <div class="card-body">
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- Source CI -->
            <div>
              <label class="form-label">Source Configuration Item</label>
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center justify-between">
                  <div>
                    <router-link
                      :to="`/ci/${existingRelationship.source_id}`"
                      class="text-lg font-medium text-blue-600 hover:text-blue-900"
                    >
                      {{ sourceCI?.name || 'Loading...' }}
                    </router-link>
                    <p class="text-sm text-gray-500 mt-1">{{ sourceCI?.ci_type || 'Unknown Type' }}</p>
                    <p class="text-xs text-gray-400 mt-1">ID: {{ existingRelationship.source_id }}</p>
                  </div>
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                  </svg>
                </div>
              </div>
            </div>

            <!-- Target CI -->
            <div>
              <label class="form-label">Target Configuration Item</label>
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="flex items-center justify-between">
                  <div>
                    <router-link
                      :to="`/ci/${existingRelationship.target_id}`"
                      class="text-lg font-medium text-blue-600 hover:text-blue-900"
                    >
                      {{ targetCI?.name || 'Loading...' }}
                    </router-link>
                    <p class="text-sm text-gray-500 mt-1">{{ targetCI?.ci_type || 'Unknown Type' }}</p>
                    <p class="text-xs text-gray-400 mt-1">ID: {{ existingRelationship.target_id }}</p>
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
            <label class="form-label">Relationship Type</label>
            <div class="flex items-center">
              <span
                v-if="relationshipTypeDetails?.color"
                class="badge text-lg px-4 py-2"
                :style="{ backgroundColor: relationshipTypeDetails.color + '20', color: relationshipTypeDetails.color, borderColor: relationshipTypeDetails.color }"
              >
                {{ relationshipTypeDetails.forward_label }}
              </span>
              <span v-else class="badge badge-info text-lg px-4 py-2">
                {{ formatRelationshipType(existingRelationship.relationship_type) }}
              </span>
            </div>
            <p v-if="relationshipTypeDetails?.description" class="text-sm text-gray-500 mt-2">
              {{ relationshipTypeDetails.description }}
            </p>
            <p class="text-sm text-gray-500 mt-2">Note: Relationship type and CIs cannot be modified after creation. Only attributes can be updated.</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ isEditing ? 'Edit Attributes' : 'Create Relationship' }}</h3>
        </div>
        <div class="card-body">
          <form @submit.prevent="handleSubmit" class="space-y-6">
            <!-- Create Mode Fields -->
            <div v-if="!isEditing">
              <!-- Bulk Mode Selection -->
              <div class="bg-blue-50 rounded-lg p-4 mb-6">
                <label class="form-label mb-2 block">Bulk Creation Mode</label>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <label class="flex items-center p-3 bg-white rounded-lg border cursor-pointer hover:bg-gray-50 transition-colors"
                         :class="bulkMode === 'none' ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200'">
                    <input
                      v-model="bulkMode"
                      type="radio"
                      value="none"
                      class="mr-3"
                    />
                    <div>
                      <div class="font-medium text-gray-900">Single Relationship</div>
                      <div class="text-sm text-gray-500">One source → One target</div>
                    </div>
                  </label>

                  <label class="flex items-center p-3 bg-white rounded-lg border cursor-pointer hover:bg-gray-50 transition-colors"
                         :class="bulkMode === 'targets' ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200'">
                    <input
                      v-model="bulkMode"
                      type="radio"
                      value="targets"
                      class="mr-3"
                    />
                    <div>
                      <div class="font-medium text-gray-900">Multiple Targets</div>
                      <div class="text-sm text-gray-500">One source → Multiple targets</div>
                    </div>
                  </label>

                  <label class="flex items-center p-3 bg-white rounded-lg border cursor-pointer hover:bg-gray-50 transition-colors"
                         :class="bulkMode === 'sources' ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200'">
                    <input
                      v-model="bulkMode"
                      type="radio"
                      value="sources"
                      class="mr-3"
                    />
                    <div>
                      <div class="font-medium text-gray-900">Multiple Sources</div>
                      <div class="text-sm text-gray-500">Multiple sources → One target</div>
                    </div>
                  </label>
                </div>

                <!-- Matrix Mode Option -->
                <div class="mt-3 pt-3 border-t border-blue-200">
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="matrixMode"
                      type="checkbox"
                      :disabled="bulkMode === 'none'"
                      class="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50 mr-2"
                    />
                    <span class="text-sm text-gray-700">Matrix mode: Create all combinations (multiple sources × multiple targets)</span>
                  </label>
                </div>
              </div>

              <!-- Source CI(s) -->
              <div>
                <label class="form-label">{{ bulkMode === 'sources' || matrixMode ? 'Source Configuration Items (Multiple)' : 'Source Configuration Item' }}</label>
                <SearchableCISelect
                  v-if="bulkMode === 'sources' || matrixMode"
                  v-model="form.source_ids"
                  :multiple="true"
                  placeholder="Search and select source CIs..."
                  :help-text="matrixMode ? 'Select multiple source CIs for matrix creation' : 'Select multiple configuration items as sources'"
                  :exclude-ids="getExcludeIdsForSources()"
                  :max-results="5"
                  @change="handleSourcesChange"
                />
                <SearchableCISelect
                  v-else
                  v-model="form.source_id"
                  placeholder="Search for source CI..."
                  help-text="The configuration item that is the source of this relationship"
                  :disabled="!!sourceId"
                  :exclude-ids="[]"
                  :max-results="5"
                  @change="handleSourceChange"
                />
              </div>

              <!-- Target CI(s) -->
              <div>
                <label class="form-label">{{ bulkMode === 'targets' || matrixMode ? 'Target Configuration Items (Multiple)' : 'Target Configuration Item' }}</label>
                <SearchableCISelect
                  v-if="bulkMode === 'targets' || matrixMode"
                  v-model="form.target_ids"
                  :multiple="true"
                  placeholder="Search and select target CIs..."
                  :help-text="matrixMode ? 'Select multiple target CIs for matrix creation' : 'Select multiple configuration items as targets'"
                  :exclude-ids="getExcludeIdsForTargets()"
                  :max-results="5"
                  @change="handleTargetsChange"
                />
                <SearchableCISelect
                  v-else
                  v-model="form.target_id"
                  placeholder="Search for target CI..."
                  help-text="The configuration item that is the target of this relationship"
                  :exclude-ids="form.source_id ? [form.source_id] : []"
                  :max-results="5"
                  @change="handleTargetChange"
                />
              </div>

              <!-- Relationship Type -->
              <RelationshipTypeSelect
                v-model="form.relationship_type"
                label="Relationship Type"
                placeholder="Select relationship type"
                help-text="The type of relationship between these configuration items"
                :required="true"
                @change="handleRelationshipTypeChange"
              />
            </div>

            <!-- Attributes (Always shown) -->
            <div>
              <label class="form-label">Attributes (Optional)</label>
              <div class="space-y-3">
                <div v-for="(attr, index) in attributesList" :key="index" class="flex space-x-2">
                  <input
                    v-model="attr.key"
                    type="text"
                    placeholder="Attribute name"
                    class="form-input flex-1"
                  />
                  <input
                    v-model="attr.value"
                    type="text"
                    placeholder="Attribute value"
                    class="form-input flex-1"
                  />
                  <button
                    type="button"
                    @click="removeAttribute(index)"
                    class="btn btn-outline"
                  >
                    Remove
                  </button>
                </div>
                <button
                  type="button"
                  @click="addAttribute"
                  class="btn btn-outline"
                >
                  Add Attribute
                </button>
              </div>
              <p class="form-help">Additional attributes that describe this relationship</p>
            </div>

            <!-- Form Actions -->
            <div class="flex flex-wrap gap-3 pt-4 border-t border-gray-200">
              <button
                v-if="!isEditing"
                type="button"
                :disabled="submitting"
                class="btn btn-outline"
                @click="handleSubmitAndAddMore"
              >
                <span v-if="submitting" class="spinner w-4 h-4 mr-2"></span>
                {{ submitting ? 'Creating...' : 'Create Relationship & Add More' }}
              </button>
              <button
                type="submit"
                :disabled="submitting"
                class="btn btn-primary"
              >
                <span v-if="submitting" class="spinner w-4 h-4 mr-2"></span>
                {{ submitButtonText }}
              </button>
              <router-link
                to="/relationships"
                class="btn btn-outline"
              >
                Cancel
              </router-link>
            </div>
          </form>
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
import { useNotificationStore } from '@/stores/notification'
import { ciAPI, relationshipAPI } from '@/services/api'
import RelationshipTypeSelect from '@/components/relationship/RelationshipTypeSelect.vue'
import SearchableCISelect from '@/components/ci/SearchableCISelect.vue'

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
const relationshipTypeStore = useRelationshipTypeStore()
const notificationStore = useNotificationStore()

const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const existingRelationship = ref<Relationship | null>(null)
const sourceCI = ref<CI | null>(null)
const targetCI = ref<CI | null>(null)
const bulkMode = ref<'none' | 'targets' | 'sources'>('none')
const matrixMode = ref(false)

// Check if we're editing an existing relationship
const isEditing = computed(() => !!route.params.id)

// Get source_id from query params if provided
const sourceId = computed(() => route.query.source_id as string || '')

const form = ref({
  source_id: '',
  source_ids: [] as string[],
  target_id: '',
  target_ids: [] as string[],
  relationship_type: ''
})

const attributesList = ref([{ key: '', value: '' }])

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

// Get relationship type details for display
const relationshipTypeDetails = computed(() => {
  if (!existingRelationship.value) return null
  return relationshipTypeStore.getRelationshipTypeByName(existingRelationship.value.relationship_type)
})

const submitButtonText = computed(() => {
  if (submitting.value) {
    return isEditing.value ? 'Updating...' : 'Creating...'
  }
  if (isEditing.value) {
    return 'Update Relationship'
  }

  const count = matrixMode.value
    ? form.value.source_ids.length * form.value.target_ids.length
    : bulkMode.value === 'targets'
      ? form.value.target_ids.length
      : bulkMode.value === 'sources'
        ? form.value.source_ids.length
        : 1

  if (count > 1) {
    return `Create ${count} Relationship${count > 1 ? 's' : ''}`
  }
  return 'Create Relationship'
})

const formatRelationshipType = (type: string) => {
  return type.split('_').map(word =>
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ')
}

const handleRelationshipTypeChange = (type: any) => {
  console.log('Relationship type changed:', type)
}

const handleSourceChange = (ci: CI | null) => {
  // Clear target if source changes to prevent self-referencing
  if (ci && form.value.target_id === ci.id) {
    form.value.target_id = ''
  }
}

const handleSourcesChange = (cis: CI[] | null) => {
  // Clear any targets that are in the source list
  if (cis) {
    const sourceIds = cis.map(c => c.id)
    if (bulkMode.value === 'sources') {
      if (form.value.target_id && sourceIds.includes(form.value.target_id)) {
        form.value.target_id = ''
      }
    } else if (matrixMode.value) {
      form.value.target_ids = form.value.target_ids.filter(id => !sourceIds.includes(id))
    }
  }
}

const handleTargetChange = (ci: CI | null) => {
  // Clear source if target changes to prevent self-referencing
  if (ci) {
    if (form.value.source_id === ci.id) {
      form.value.source_id = ''
    }
    if (form.value.source_ids.includes(ci.id)) {
      form.value.source_ids = form.value.source_ids.filter(id => id !== ci.id)
    }
  }
}

const handleTargetsChange = (cis: CI[] | null) => {
  // Clear any sources that are in the target list
  if (cis) {
    const targetIds = cis.map(c => c.id)
    if (bulkMode.value === 'targets') {
      if (form.value.source_id && targetIds.includes(form.value.source_id)) {
        form.value.source_id = ''
      }
    } else if (matrixMode.value) {
      form.value.source_ids = form.value.source_ids.filter(id => !targetIds.includes(id))
    }
  }
}

// Helper functions to get exclude IDs for the searchable selects
const getExcludeIdsForSources = () => {
  const excludes: string[] = []
  if (bulkMode.value === 'targets') {
    // Single source mode, exclude selected target
    if (form.value.target_id) {
      excludes.push(form.value.target_id)
    }
  } else if (matrixMode.value) {
    // Matrix mode, exclude all selected targets
    excludes.push(...form.value.target_ids)
  }
  return excludes
}

const getExcludeIdsForTargets = () => {
  const excludes: string[] = []
  if (bulkMode.value === 'sources') {
    // Single target mode, exclude selected sources
    if (form.value.source_id) {
      excludes.push(form.value.source_id)
    }
  } else if (matrixMode.value) {
    // Matrix mode, exclude all selected sources
    excludes.push(...form.value.source_ids)
  } else if (bulkMode.value === 'targets') {
    // Multiple targets mode, exclude selected source
    if (form.value.source_id) {
      excludes.push(form.value.source_id)
    }
  } else {
    // Single mode, exclude selected source
    if (form.value.source_id) {
      excludes.push(form.value.source_id)
    }
  }
  return excludes
}

const loadRelationship = async () => {
  const relationshipId = route.params.id as string
  if (!relationshipId) return

  loading.value = true
  try {
    const response = await relationshipAPI.get(relationshipId)
    existingRelationship.value = response.data

    // Populate form with existing data
    form.value = {
      source_id: response.data.source_id,
      source_ids: [],
      target_id: response.data.target_id,
      target_ids: [],
      relationship_type: response.data.relationship_type
    }

    // Populate attributes
    if (response.data.attributes && Object.keys(response.data.attributes).length > 0) {
      attributesList.value = Object.entries(response.data.attributes).map(([key, value]) => ({
        key,
        value: String(value)
      }))
    } else {
      attributesList.value = [{ key: '', value: '' }]
    }

    // Load CI details for display
    await Promise.all([
      loadCIDetails(response.data.source_id, 'source'),
      loadCIDetails(response.data.target_id, 'target')
    ])
  } catch (err: any) {
    console.error('Failed to load relationship:', err)
    error.value = err.response?.data?.error || 'Failed to load relationship'
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

const addAttribute = () => {
  attributesList.value.push({ key: '', value: '' })
}

const removeAttribute = (index: number) => {
  if (attributesList.value.length > 1) {
    attributesList.value.splice(index, 1)
  }
}

const handleSubmit = async (andAddMore = false) => {
  const permissionRequired = isEditing.value ? 'relationship:update' : 'relationship:create'
  if (!hasPermission(permissionRequired)) {
    notificationStore.showError(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationships`)
    return
  }

  // Validate form
  if (!isEditing.value) {
    if (!form.value.relationship_type) {
      notificationStore.showError('Please select a relationship type')
      return
    }

    // Validate based on mode
    if (matrixMode.value) {
      if (form.value.source_ids.length === 0) {
        notificationStore.showError('Please select at least one source CI')
        return
      }
      if (form.value.target_ids.length === 0) {
        notificationStore.showError('Please select at least one target CI')
        return
      }
    } else if (bulkMode.value === 'targets') {
      if (!form.value.source_id) {
        notificationStore.showError('Please select a source CI')
        return
      }
      if (form.value.target_ids.length === 0) {
        notificationStore.showError('Please select at least one target CI')
        return
      }
      if (form.value.target_ids.includes(form.value.source_id)) {
        notificationStore.showError('Source and target configuration items must be different')
        return
      }
    } else if (bulkMode.value === 'sources') {
      if (form.value.source_ids.length === 0) {
        notificationStore.showError('Please select at least one source CI')
        return
      }
      if (!form.value.target_id) {
        notificationStore.showError('Please select a target CI')
        return
      }
      if (form.value.source_ids.includes(form.value.target_id)) {
        notificationStore.showError('Source and target configuration items must be different')
        return
      }
    } else {
      // Single mode
      if (!form.value.source_id) {
        notificationStore.showError('Please select a source CI')
        return
      }
      if (!form.value.target_id) {
        notificationStore.showError('Please select a target CI')
        return
      }
      if (form.value.source_id === form.value.target_id) {
        notificationStore.showError('Source and target configuration items must be different')
        return
      }
    }
  }

  submitting.value = true
  try {
    // Build attributes object
    const attributes: Record<string, any> = {}
    attributesList.value.forEach(attr => {
      if (attr.key && attr.value) {
        attributes[attr.key] = attr.value
      }
    })

    if (isEditing.value) {
      // In edit mode, only update attributes
      await relationshipAPI.update(route.params.id as string, {
        attributes: Object.keys(attributes).length > 0 ? attributes : undefined
      })
      notificationStore.showSuccess('Relationship updated successfully')
    } else if (matrixMode.value) {
      // Matrix create mode: multiple sources × multiple targets
      const matrixData = {
        source_ids: form.value.source_ids,
        target_ids: form.value.target_ids,
        relationship_type: form.value.relationship_type,
        attributes: Object.keys(attributes).length > 0 ? attributes : undefined
      }

      const response = await relationshipAPI.createBulkMatrix(matrixData)
      notificationStore.showSuccess(
        `Successfully created ${response.data.total_created} relationship(s)`
      )
    } else if (bulkMode.value === 'targets') {
      // Bulk create mode: one source → multiple targets (use existing createBulk if available)
      for (const targetId of form.value.target_ids) {
        const relationshipData = {
          source_id: form.value.source_id,
          target_id: targetId,
          relationship_type: form.value.relationship_type,
          attributes: Object.keys(attributes).length > 0 ? attributes : undefined
        }
        await relationshipAPI.create(relationshipData)
      }
      notificationStore.showSuccess(
        `Successfully created ${form.value.target_ids.length} relationship(s)`
      )
    } else if (bulkMode.value === 'sources') {
      // Bulk sources mode: multiple sources → one target
      const bulkSourcesData = {
        source_ids: form.value.source_ids,
        target_id: form.value.target_id,
        relationship_type: form.value.relationship_type,
        attributes: Object.keys(attributes).length > 0 ? attributes : undefined
      }

      const response = await relationshipAPI.createBulkFromSources(bulkSourcesData)
      notificationStore.showSuccess(
        `Successfully created ${response.data.total_created} relationship(s)`
      )
    } else {
      // Single create mode
      const relationshipData = {
        source_id: form.value.source_id,
        target_id: form.value.target_id,
        relationship_type: form.value.relationship_type,
        attributes: Object.keys(attributes).length > 0 ? attributes : undefined
      }

      await relationshipAPI.create(relationshipData)
      notificationStore.showSuccess('Relationship created successfully')
    }

    if (andAddMore) {
      // Reset form for adding more relationships
      form.value.source_id = ''
      form.value.source_ids = []
      form.value.target_id = ''
      form.value.target_ids = []
      attributesList.value = [{ key: '', value: '' }]
    } else {
      router.push('/relationships')
    }
  } catch (err: any) {
    console.error(`Failed to ${isEditing.value ? 'update' : 'create'} relationship:`, err)
    const message = err.response?.data?.error || `Failed to ${isEditing.value ? 'update' : 'create'} relationship`
    notificationStore.showError(message)
  } finally {
    submitting.value = false
  }
}

const handleSubmitAndAddMore = async () => {
  await handleSubmit(true)
}

onMounted(async () => {
  // Load relationship types
  if (relationshipTypeStore.relationshipTypes.length === 0) {
    try {
      await relationshipTypeStore.loadRelationshipTypes()
    } catch (err) {
      console.error('Failed to load relationship types:', err)
    }
  }

  const permissionRequired = isEditing.value ? 'relationship:update' : 'relationship:create'
  if (!hasPermission(permissionRequired)) {
    notificationStore.showError(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationships`)
    router.push('/relationships')
    return
  }

  if (isEditing.value) {
    await loadRelationship()
  } else if (sourceId.value) {
    form.value.source_id = sourceId.value
  }
})
</script>

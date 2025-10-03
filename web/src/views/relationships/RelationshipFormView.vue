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
    <div v-else class="max-w-2xl mx-auto">
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
              <span class="badge badge-info text-lg px-4 py-2">
                {{ formatRelationshipType(existingRelationship.relationship_type) }}
              </span>
            </div>
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
              <!-- Source CI -->
              <div>
                <label class="form-label">Source Configuration Item</label>
                <select
                  v-model="form.source_id"
                  class="form-input"
                  required
                  :disabled="!!sourceId"
                >
                  <option value="">Select a source CI</option>
                  <option
                    v-for="ci in configurationItems"
                    :key="ci.id"
                    :value="ci.id"
                  >
                    {{ ci.name }} ({{ ci.ci_type }})
                  </option>
                </select>
                <p class="form-help">The configuration item that is the source of this relationship</p>
              </div>

              <!-- Target CI -->
              <div>
                <label class="form-label">Target Configuration Item</label>
                <select
                  v-model="form.target_id"
                  class="form-input"
                  required
                >
                  <option value="">Select a target CI</option>
                  <option
                    v-for="ci in configurationItems"
                    :key="ci.id"
                    :value="ci.id"
                    :disabled="ci.id === form.source_id"
                  >
                    {{ ci.name }} ({{ ci.ci_type }})
                  </option>
                </select>
                <p class="form-help">The configuration item that is the target of this relationship</p>
              </div>

              <!-- Relationship Type -->
              <div>
                <label class="form-label">Relationship Type</label>
                <select v-model="form.relationship_type" class="form-input" required>
                  <option value="">Select relationship type</option>
                  <option value="depends_on">Depends On</option>
                  <option value="connected_to">Connected To</option>
                  <option value="runs_on">Runs On</option>
                  <option value="contains">Contains</option>
                  <option value="managed_by">Managed By</option>
                  <option value="monitors">Monitors</option>
                  <option value="backed_up_by">Backed Up By</option>
                  <option value="secured_by">Secured By</option>
                </select>
                <p class="form-help">The type of relationship between these configuration items</p>
              </div>
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
            <div class="flex space-x-3 pt-4 border-t border-gray-200">
              <button
                type="submit"
                :disabled="submitting"
                class="btn btn-primary"
              >
                <span v-if="submitting" class="spinner w-4 h-4 mr-2"></span>
                {{ submitting ? (isEditing ? 'Updating...' : 'Creating...') : (isEditing ? 'Update Relationship' : 'Create Relationship') }}
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
import { useNotificationStore } from '@/stores/notification'
import { ciAPI, relationshipAPI } from '@/services/api'

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
const notificationStore = useNotificationStore()

const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const configurationItems = ref<CI[]>([])
const existingRelationship = ref<Relationship | null>(null)
const sourceCI = ref<CI | null>(null)
const targetCI = ref<CI | null>(null)

// Check if we're editing an existing relationship
const isEditing = computed(() => !!route.params.id)

// Get source_id from query params if provided
const sourceId = computed(() => route.query.source_id as string || '')

const form = ref({
  source_id: '',
  target_id: '',
  relationship_type: ''
})

const attributesList = ref([{ key: '', value: '' }])

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const formatRelationshipType = (type: string) => {
  return type.split('_').map(word =>
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ')
}

const loadConfigurationItems = async () => {
  loading.value = true
  try {
    const response = await ciAPI.list({ limit: 1000 })
    configurationItems.value = response.data.cis || []
  } catch (err) {
    console.error('Failed to load configuration items:', err)
    error.value = 'Failed to load configuration items'
  } finally {
    loading.value = false
  }
}

const loadRelationship = async () => {
  const relationshipId = route.params.id as string
  if (!relationshipId) return

  loading.value = true
  try {
    const response = await relationshipAPI.get(relationshipId)
    existingRelationship.value = response.data

    // Populate form with existing data (for create mode compatibility)
    form.value = {
      source_id: response.data.source_id,
      target_id: response.data.target_id,
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

const handleSubmit = async () => {
  const permissionRequired = isEditing.value ? 'relationship:update' : 'relationship:create'
  if (!hasPermission(permissionRequired)) {
    notificationStore.showError(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationships`)
    return
  }

  // Validate form
  if (!isEditing.value) {
    if (!form.value.source_id || !form.value.target_id || !form.value.relationship_type) {
      notificationStore.showError('Please fill in all required fields')
      return
    }

    if (form.value.source_id === form.value.target_id) {
      notificationStore.showError('Source and target configuration items must be different')
      return
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
    } else {
      // In create mode, create new relationship
      const relationshipData = {
        source_id: form.value.source_id,
        target_id: form.value.target_id,
        relationship_type: form.value.relationship_type,
        attributes: Object.keys(attributes).length > 0 ? attributes : undefined
      }

      await relationshipAPI.create(relationshipData)
      notificationStore.showSuccess('Relationship created successfully')
    }

    router.push('/relationships')
  } catch (err: any) {
    console.error(`Failed to ${isEditing.value ? 'update' : 'create'} relationship:`, err)
    const message = err.response?.data?.error || `Failed to ${isEditing.value ? 'update' : 'create'} relationship`
    notificationStore.showError(message)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  const permissionRequired = isEditing.value ? 'relationship:update' : 'relationship:create'
  if (!hasPermission(permissionRequired)) {
    notificationStore.showError(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationships`)
    router.push('/relationships')
    return
  }

  await loadConfigurationItems()

  if (isEditing.value) {
    await loadRelationship()
  } else if (sourceId.value) {
    form.value.source_id = sourceId.value
  }
})
</script>
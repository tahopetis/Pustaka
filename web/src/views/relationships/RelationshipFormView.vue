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
                <span class="ml-4 text-sm font-medium text-gray-900">Create Relationship</span>
              </div>
            </li>
          </ol>
        </nav>
        <h1 class="text-3xl font-bold text-gray-900 mt-2">Create Relationship</h1>
        <p class="mt-2 text-gray-600">Create a new relationship between configuration items</p>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading configuration items...</p>
    </div>

    <div v-else class="max-w-2xl mx-auto">
      <div class="card">
        <div class="card-body">
          <form @submit.prevent="handleSubmit" class="space-y-6">
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

            <!-- Attributes (Optional) -->
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
                {{ submitting ? 'Creating...' : 'Create Relationship' }}
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
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import { ciAPI, relationshipAPI } from '@/services/api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const loading = ref(false)
const submitting = ref(false)
const configurationItems = ref<any[]>([])

// Get source_id from query params if provided
const sourceId = computed(() => route.query.source_id as string || '')

const form = ref({
  source_id: sourceId.value,
  target_id: '',
  relationship_type: ''
})

const attributesList = ref([{ key: '', value: '' }])

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const loadConfigurationItems = async () => {
  loading.value = true
  try {
    const response = await ciAPI.list({ limit: 1000 })
    configurationItems.value = response.data.cis || []
  } catch (error) {
    console.error('Failed to load configuration items:', error)
    notificationStore.showError('Failed to load configuration items')
  } finally {
    loading.value = false
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
  if (!hasPermission('relationship:create')) {
    notificationStore.showError('You do not have permission to create relationships')
    return
  }

  // Validate form
  if (!form.value.source_id || !form.value.target_id || !form.value.relationship_type) {
    notificationStore.showError('Please fill in all required fields')
    return
  }

  if (form.value.source_id === form.value.target_id) {
    notificationStore.showError('Source and target configuration items must be different')
    return
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

    const relationshipData = {
      source_id: form.value.source_id,
      target_id: form.value.target_id,
      relationship_type: form.value.relationship_type,
      attributes: Object.keys(attributes).length > 0 ? attributes : undefined
    }

    await relationshipAPI.create(relationshipData)
    notificationStore.showSuccess('Relationship created successfully')
    router.push('/relationships')
  } catch (error: any) {
    console.error('Failed to create relationship:', error)
    const message = error.response?.data?.error || 'Failed to create relationship'
    notificationStore.showError(message)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (!hasPermission('relationship:create')) {
    notificationStore.showError('You do not have permission to create relationships')
    router.push('/relationships')
    return
  }

  loadConfigurationItems()
})
</script>
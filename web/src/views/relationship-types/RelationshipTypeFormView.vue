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
                <span class="ml-4 text-sm font-medium text-gray-900">{{ isEditing ? 'Edit Relationship Type' : 'Create Relationship Type' }}</span>
              </div>
            </li>
          </ol>
        </nav>
        <h1 class="text-3xl font-bold text-gray-900 mt-2">{{ isEditing ? 'Edit Relationship Type' : 'Create Relationship Type' }}</h1>
        <p class="mt-2 text-gray-600">{{ isEditing ? 'Update relationship type configuration' : 'Define a new type of relationship between configuration items' }}</p>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-12">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading relationship type...</p>
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

    <!-- Form -->
    <div v-else class="max-w-4xl mx-auto">
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">
            {{ isEditing ? 'Edit Relationship Type' : 'Create Relationship Type' }}
          </h3>
        </div>
        <div class="card-body">
          <form @submit.prevent="handleSubmit" class="space-y-8">
            <!-- Basic Information Section -->
            <div>
              <h4 class="text-md font-medium text-gray-900 mb-4">Basic Information</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Name -->
                <div>
                  <label class="form-label">Name *</label>
                  <input
                    v-model="form.name"
                    type="text"
                    class="form-input"
                    required
                    placeholder="e.g., scope, hosts, supports, managed_by"
                    :disabled="isEditing"
                  />
                  <p class="form-help">
                    Internal system name (underscore separated, lowercase). {{ isEditing ? 'Cannot be modified after creation.' : '' }}
                  </p>
                </div>

                <!-- Display Name -->
                <div>
                  <label class="form-label">Display Name</label>
                  <input
                    v-model="form.display_name"
                    type="text"
                    class="form-input"
                    placeholder="e.g., Scope, Hosts, Supports, Managed By"
                  />
                  <p class="form-help">
                    Human-readable name for display purposes (optional)
                  </p>
                </div>

                <!-- Forward Label -->
                <div>
                  <label class="form-label">Forward Label *</label>
                  <input
                    v-model="form.forward_label"
                    type="text"
                    class="form-input"
                    required
                    placeholder="e.g., scopes, hosts, supports, managed by"
                  />
                  <p class="form-help">
                    How the relationship is described from source to target
                  </p>
                </div>

                <!-- Reverse Label -->
                <div>
                  <label class="form-label">Reverse Label *</label>
                  <input
                    v-model="form.reverse_label"
                    type="text"
                    class="form-input"
                    required
                    placeholder="e.g., scoped by, hosted by, supported by, manages"
                  />
                  <p class="form-help">
                    How the relationship is described from target to source
                  </p>
                </div>

                <!-- Category -->
                <div>
                  <label class="form-label">Category</label>
                  <select v-model="form.category" class="form-input">
                    <option value="">Select a category...</option>
                    <option v-for="category in availableCategories" :key="category" :value="category">
                      {{ category }}
                    </option>
                  </select>
                  <p class="form-help">
                    Group related relationship types together
                  </p>
                </div>

                <!-- Bidirectional -->
                <div class="flex items-center space-x-4 pt-6">
                  <label class="flex items-center">
                    <input
                      v-model="form.bidirectional"
                      type="checkbox"
                      class="form-checkbox"
                    />
                    <span class="ml-2">Bidirectional</span>
                  </label>
                  <p class="text-sm text-gray-500">
                    Relationship works the same in both directions
                  </p>
                </div>
              </div>

              <!-- Description -->
              <div class="mt-6">
                <label class="form-label">Description</label>
                <textarea
                  v-model="form.description"
                  class="form-input"
                  rows="3"
                  placeholder="Describe when this relationship type should be used..."
                ></textarea>
                <p class="form-help">
                  Optional description to help users understand when to use this relationship type
                </p>
              </div>
            </div>

            <!-- Appearance Section -->
            <div class="border-t pt-8">
              <h4 class="text-md font-medium text-gray-900 mb-4">Appearance</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Color -->
                <div>
                  <label class="form-label">Color</label>
                  <div class="flex items-center space-x-3">
                    <input
                      v-model="form.color"
                      type="color"
                      class="h-10 w-20 border border-gray-300 rounded cursor-pointer"
                    />
                    <input
                      v-model="form.color"
                      type="text"
                      class="form-input flex-1"
                      placeholder="#3B82F6"
                      pattern="^#[0-9A-Fa-f]{6}$"
                    />
                  </div>
                  <p class="form-help">
                    Color used in graph visualization (hex color code)
                  </p>
                </div>

                <!-- Icon -->
                <div>
                  <label class="form-label">Icon</label>
                  <input
                    v-model="form.icon"
                    type="text"
                    class="form-input"
                    placeholder="e.g., arrow-right, link, database"
                  />
                  <p class="form-help">
                    Optional icon name for enhanced visualization
                  </p>
                </div>
              </div>
            </div>

            <!-- Constraints Section -->
            <div class="border-t pt-8">
              <h4 class="text-md font-medium text-gray-900 mb-4">Relationship Constraints</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Cardinality Source -->
                <div>
                  <label class="form-label">Source Cardinality</label>
                  <select v-model="form.cardinality_source" class="form-input">
                    <option value="">No restriction</option>
                    <option value="1">Exactly one</option>
                    <option value="0..1">Zero or one</option>
                    <option value="*">Zero or more</option>
                    <option value="1..*">One or more</option>
                  </select>
                  <p class="form-help">
                    How many relationships a source CI can have of this type
                  </p>
                </div>

                <!-- Cardinality Target -->
                <div>
                  <label class="form-label">Target Cardinality</label>
                  <select v-model="form.cardinality_target" class="form-input">
                    <option value="">No restriction</option>
                    <option value="1">Exactly one</option>
                    <option value="0..1">Zero or one</option>
                    <option value="*">Zero or more</option>
                    <option value="1..*">One or more</option>
                  </select>
                  <p class="form-help">
                    How many relationships a target CI can have of this type
                  </p>
                </div>
              </div>

              <!-- Allowed Source Types -->
              <div class="mt-6">
                <label class="form-label">Allowed Source Types</label>
                <div class="border rounded-lg p-4 bg-gray-50">
                  <p class="text-sm text-gray-600 mb-3">
                    Leave empty to allow all CI types as sources
                  </p>
                  <!-- This would need a CI type selector component -->
                  <div class="text-sm text-gray-500">
                    CI type selector component would go here
                  </div>
                </div>
              </div>

              <!-- Allowed Target Types -->
              <div class="mt-6">
                <label class="form-label">Allowed Target Types</label>
                <div class="border rounded-lg p-4 bg-gray-50">
                  <p class="text-sm text-gray-600 mb-3">
                    Leave empty to allow all CI types as targets
                  </p>
                  <!-- This would need a CI type selector component -->
                  <div class="text-sm text-gray-500">
                    CI type selector component would go here
                  </div>
                </div>
              </div>
            </div>

            <!-- Status (Edit mode only) -->
            <div v-if="isEditing" class="border-t pt-8">
              <h4 class="text-md font-medium text-gray-900 mb-4">Status</h4>
              <div class="flex items-center space-x-4">
                <label class="flex items-center">
                  <input
                    v-model="form.is_active"
                    type="radio"
                    :value="true"
                    class="form-radio"
                  />
                  <span class="ml-2">Active</span>
                </label>
                <label class="flex items-center">
                  <input
                    v-model="form.is_active"
                    type="radio"
                    :value="false"
                    class="form-radio"
                  />
                  <span class="ml-2">Inactive</span>
                </label>
              </div>
              <p class="form-help">
                Inactive types cannot be used for new relationships
              </p>
            </div>

            <!-- Preview -->
            <div class="border-t pt-8">
              <h4 class="text-md font-medium text-gray-900 mb-4">Preview</h4>
              <div class="bg-gray-50 rounded-lg p-4">
                <div class="space-y-3">
                  <div class="flex items-center space-x-3">
                    <span class="text-sm font-medium text-gray-700 w-24">Forward:</span>
                    <span
                      v-if="form.color"
                      class="badge"
                      :style="{ backgroundColor: form.color + '20', color: form.color, borderColor: form.color }"
                    >
                      {{ form.forward_label || 'Forward Label' }}
                    </span>
                    <span v-else class="badge badge-info">
                      {{ form.forward_label || 'Forward Label' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-3">
                    <span class="text-sm font-medium text-gray-700 w-24">Reverse:</span>
                    <span
                      v-if="form.color"
                      class="badge"
                      :style="{ backgroundColor: form.color + '20', color: form.color, borderColor: form.color }"
                    >
                      {{ form.reverse_label || 'Reverse Label' }}
                    </span>
                    <span v-else class="badge badge-secondary">
                      {{ form.reverse_label || 'Reverse Label' }}
                    </span>
                  </div>
                  <div v-if="form.bidirectional" class="flex items-center space-x-3">
                    <span class="text-sm font-medium text-gray-700 w-24">Type:</span>
                    <span class="badge badge-gray">Bidirectional</span>
                  </div>
                  <div v-if="form.category" class="flex items-center space-x-3">
                    <span class="text-sm font-medium text-gray-700 w-24">Category:</span>
                    <span class="text-sm text-gray-600">{{ form.category }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Form Actions -->
            <div class="flex space-x-3 pt-6 border-t border-gray-200">
              <button
                type="submit"
                :disabled="submitting"
                class="btn btn-primary"
              >
                <span v-if="submitting" class="spinner w-4 h-4 mr-2"></span>
                {{ submitting ? (isEditing ? 'Updating...' : 'Creating...') : (isEditing ? 'Update Relationship Type' : 'Create Relationship Type') }}
              </button>
              <router-link
                to="/relationship-types"
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
import { showSuccessToast, showErrorToast } from '@/utils/toast'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const relationshipTypeStore = useRelationshipTypeStore()

const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const availableCategories = ref<string[]>([])

// Check if we're editing an existing relationship type
const isEditing = computed(() => !!route.params.id)

const form = ref({
  name: '',
  display_name: '',
  forward_label: '',
  reverse_label: '',
  description: '',
  color: '#3B82F6',
  icon: '',
  category: '',
  bidirectional: false,
  cardinality_source: 'many',
  cardinality_target: 'many',
  allowed_source_types: [] as string[],
  allowed_target_types: [] as string[],
  is_active: true
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const loadCategories = async () => {
  try {
    await relationshipTypeStore.loadCategories()
    availableCategories.value = relationshipTypeStore.categories
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

const loadRelationshipType = async () => {
  const typeId = route.params.id as string
  if (!typeId) return

  loading.value = true
  try {
    const relationshipType = relationshipTypeStore.getRelationshipTypeById(typeId)

    if (relationshipType) {
      form.value = {
        name: relationshipType.name,
        display_name: relationshipType.display_name || '',
        forward_label: relationshipType.forward_label,
        reverse_label: relationshipType.reverse_label,
        description: relationshipType.description || '',
        color: relationshipType.color || '#3B82F6',
        icon: relationshipType.icon || '',
        category: relationshipType.category || '',
        bidirectional: relationshipType.bidirectional || false,
        cardinality_source: relationshipType.cardinality_source || '',
        cardinality_target: relationshipType.cardinality_target || '',
        allowed_source_types: relationshipType.allowed_source_types || [],
        allowed_target_types: relationshipType.allowed_target_types || [],
        is_active: relationshipType.is_active
      }
    } else {
      // Load from store if not in cache
      await relationshipTypeStore.loadRelationshipTypes()
      const loadedType = relationshipTypeStore.getRelationshipTypeById(typeId)
      if (loadedType) {
        form.value = {
          name: loadedType.name,
          display_name: loadedType.display_name || '',
          forward_label: loadedType.forward_label,
          reverse_label: loadedType.reverse_label,
          description: loadedType.description || '',
          color: loadedType.color || '#3B82F6',
          icon: loadedType.icon || '',
          category: loadedType.category || '',
          bidirectional: loadedType.bidirectional || false,
          cardinality_source: loadedType.cardinality_source || '',
          cardinality_target: loadedType.cardinality_target || '',
          allowed_source_types: loadedType.allowed_source_types || [],
          allowed_target_types: loadedType.allowed_target_types || [],
          is_active: loadedType.is_active
        }
      } else {
        error.value = 'Relationship type not found'
      }
    }
  } catch (err: any) {
    console.error('Failed to load relationship type:', err)
    error.value = err.response?.data?.error || 'Failed to load relationship type'
  } finally {
    loading.value = false
  }
}

const validateForm = () => {
  if (!form.value.name.trim()) {
    showErrorToast('Name is required')
    return false
  }

  if (!isEditing.value && !/^[a-z][a-z0-9_]*$/.test(form.value.name)) {
    showErrorToast('Name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores')
    return false
  }

  if (!form.value.forward_label.trim()) {
    showErrorToast('Forward label is required')
    return false
  }

  if (!form.value.reverse_label.trim()) {
    showErrorToast('Reverse label is required')
    return false
  }

  if (form.value.color && !/^#[0-9A-Fa-f]{6}$/.test(form.value.color)) {
    showErrorToast('Color must be a valid hex color code')
    return false
  }

  return true
}

const handleSubmit = async () => {
  const permissionRequired = isEditing.value ? 'relationship_type:update' : 'relationship_type:create'
  if (!hasPermission(permissionRequired)) {
    showErrorToast(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationship types`)
    return
  }

  if (!validateForm()) {
    return
  }

  submitting.value = true
  try {
    const submitData = {
      name: form.value.name.trim(),
      display_name: form.value.display_name.trim() || undefined,
      forward_label: form.value.forward_label.trim(),
      reverse_label: form.value.reverse_label.trim(),
      description: form.value.description.trim() || undefined,
      color: form.value.color || undefined,
      icon: form.value.icon.trim() || undefined,
      category: form.value.category || undefined,
      bidirectional: form.value.bidirectional,
      cardinality_source: form.value.cardinality_source || undefined,
      cardinality_target: form.value.cardinality_target || undefined,
      allowed_source_types: form.value.allowed_source_types.length > 0 ? form.value.allowed_source_types : undefined,
      allowed_target_types: form.value.allowed_target_types.length > 0 ? form.value.allowed_target_types : undefined,
      ...(isEditing.value && { is_active: form.value.is_active })
    }

    if (isEditing.value) {
      await relationshipTypeStore.updateRelationshipType(route.params.id as string, submitData)
      showSuccessToast('Relationship type updated successfully')
    } else {
      await relationshipTypeStore.createRelationshipType(submitData)
      showSuccessToast('Relationship type created successfully')
    }

    router.push('/relationship-types')
  } catch (err: any) {
    console.error(`Failed to ${isEditing.value ? 'update' : 'create'} relationship type:`, err)
    const message = err.response?.data?.error || `Failed to ${isEditing.value ? 'update' : 'create'} relationship type`
    showErrorToast(message)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadCategories()

  const permissionRequired = isEditing.value ? 'relationship_type:update' : 'relationship_type:create'
  if (!hasPermission(permissionRequired)) {
    showErrorToast(`You do not have permission to ${isEditing.value ? 'update' : 'create'} relationship types`)
    router.push('/relationship-types')
    return
  }

  if (isEditing.value) {
    await loadRelationshipType()
  }
})
</script>
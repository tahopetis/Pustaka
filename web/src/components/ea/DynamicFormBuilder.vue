<template>
  <div class="space-y-6">
    <!-- Validation Summary -->
    <ValidationSummary v-if="validationErrors.length > 0" :errors="validationErrors" />

    <!-- Entity Form -->
    <form @submit.prevent="handleSubmit">
      <!-- Basic Information Section -->
      <FormFieldGroup
        title="Basic Information"
        :collapsed="false"
        :persistent="true"
      >
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Entity Name -->
          <div :data-field="'name'" class="col-span-2">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Entity Name <span class="text-red-500">*</span>
            </label>
            <input
              v-model="formData.name"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              :class="{ 'border-red-500': fieldErrors.name }"
              placeholder="Enter entity name"
            />
            <p v-if="fieldErrors.name" class="text-sm text-red-600 mt-1">{{ fieldErrors.name }}</p>
          </div>

          <!-- CI Type (read-only in edit mode) -->
          <div :data-field="'ci_type'">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              CI Type <span class="text-red-500">*</span>
            </label>
            <input
              v-if="isEdit"
              :value="ciTypeDisplay"
              type="text"
              disabled
              class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500"
            />
            <div v-else-if="isLoadingCITypes" class="relative">
              <select
                disabled
                class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500 cursor-not-allowed"
              >
                <option>Loading CI Types...</option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                <svg class="animate-spin h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
            </div>
            <select
              v-else
              v-model="formData.ci_type"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Select CI Type</option>
              <option
                v-for="type in availableCITypes"
                :key="type.name"
                :value="type.name"
              >
                {{ type.name }}
              </option>
            </select>
            <p v-if="!isLoadingCITypes && availableCITypes.length === 0" class="text-sm text-red-600 mt-1">
              <span v-if="ciTypesError">{{ ciTypesError }}</span>
              <span v-else>No CI types available. Please check your connection or contact support.</span>
            </p>
          </div>

          <!-- Lifecycle Status -->
          <div :data-field="'lifecycle_status_id'">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Lifecycle Status
            </label>
            <select
              v-model="formData.lifecycle_status_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Select Status</option>
              <option
                v-for="status in lifecycleStatuses"
                :key="status.id"
                :value="status.id"
              >
                {{ status.display_name }}
              </option>
            </select>
          </div>

          <!-- Owner/Team (EA-specific) -->
          <div :data-field="'owner'">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Owner Team <span class="text-red-500">*</span>
            </label>
            <div v-if="isLoadingTeams" class="relative">
              <select
                disabled
                class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500 cursor-not-allowed"
              >
                <option>Loading Teams...</option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                <svg class="animate-spin h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
            </div>
            <select
              v-else
              v-model="formData.owner"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Select Team</option>
              <option
                v-for="team in eaTeams"
                :key="team.id"
                :value="team.name"
              >
                {{ team.name }}
              </option>
            </select>
            <p v-if="!isLoadingTeams && eaTeams.length === 0" class="text-sm text-red-600 mt-1">
              <span v-if="teamsError">{{ teamsError }}</span>
              <span v-else>No teams available. Please check your connection or contact support.</span>
            </p>
            <p v-else class="text-xs text-gray-500 mt-1">EA team responsible for this entity</p>
          </div>

          <!-- Tags -->
          <div class="col-span-2">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Tags
            </label>
            <tag-input
              v-model="formData.tags"
              placeholder="Add tags (press Enter to add)"
            />
          </div>
        </div>
      </FormFieldGroup>

      <!-- Dynamic Attributes Sections -->
      <FormFieldGroup
        v-for="group in fieldGroups"
        :key="group.name"
        :title="group.title"
        :collapsed="group.collapsed"
        @toggle="toggleGroupCollapse(group.name)"
      >
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="attr in group.attributes"
            :key="attr.name"
            :data-field="attr.name"
            :class="attr.type === 'object' || attr.type === 'array' ? 'col-span-2' : ''"
          >
            <DynamicAttributeField
              :attribute="attr"
              :model-value="formData.attributes[attr.name]"
              :required="attr.required"
              @update:model-value="updateAttribute(attr.name, $event)"
              @validation="onFieldValidation(attr.name, $event)"
            />
          </div>
        </div>
      </FormFieldGroup>

      <!-- Form Actions -->
      <div class="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          class="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          @click="handleCancel"
        >
          Cancel
        </button>
        <button
          v-if="!isEdit"
          type="button"
          class="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          @click="saveAsDraft"
        >
          Save as Draft
        </button>
        <button
          type="submit"
          :disabled="loading"
          class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="loading" class="flex items-center">
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Saving...
          </span>
          <span v-else>{{ isEdit ? 'Update Entity' : 'Create Entity' }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEaStore } from '@/stores/ea'
import { useEaTypesStore } from '@/stores/eaTypes'
import { lifecycleStatusAPI } from '@/services/api'
import FormFieldGroup from './FormFieldGroup.vue'
import ValidationSummary from './ValidationSummary.vue'
import DynamicAttributeField from '@/components/ci/DynamicAttributeField.vue'
import TagInput from '@/components/base/TagInput.vue'
import type { EACreateRequest, EAUpdateRequest, ValidationError, FieldGroup, AttributeSchema } from '@/types/ea'

interface Props {
  entityId?: string
  ciType?: string
  domain?: string
}

const props = defineProps<Props>()

const route = useRoute()
const router = useRouter()
const eaStore = useEaStore()
const eaTypesStore = useEaTypesStore()

// State
const formData = ref<{
  name: string
  ci_type: string
  lifecycle_status_id: string | null
  owner: string
  attributes: Record<string, any>
  tags: string[]
}>({
  name: '',
  ci_type: props.ciType || '',
  lifecycle_status_id: null,
  owner: '',
  attributes: {},
  tags: []
})

const loading = ref(false)
const validationErrors = ref<ValidationError[]>([])
const fieldValidation = ref<Record<string, boolean>>({})
const fieldErrors = ref<Record<string, string>>({})
const unsavedChanges = ref(false)
const lifecycleStatuses = ref<any[]>([])
const collapsedGroups = ref<Record<string, boolean>>({})

// Computed
const isEdit = computed(() => !!props.entityId)

const ciTypeDisplay = computed(() => {
  if (!formData.value.ci_type) return ''
  return formData.value.ci_type
})

const availableCITypes = computed(() => {
  if (props.domain) {
    return eaTypesStore.getCiTypesByDomain(props.domain)
  }
  return eaTypesStore.ciTypes
})

const isLoadingCITypes = computed(() => eaTypesStore.loading && eaTypesStore.ciTypes.length === 0)

const ciTypesError = computed(() => eaTypesStore.error)

const eaTeams = computed(() => eaTypesStore.teams || [])

const isLoadingTeams = computed(() => eaTypesStore.loading && eaTypesStore.teams.length === 0)

const teamsError = computed(() => eaTypesStore.error)

const ciTypeDefinition = computed(() => {
  if (!formData.value.ci_type) return null
  return eaTypesStore.getCiTypeByName(formData.value.ci_type)
})

const fieldGroups = computed((): FieldGroup[] => {
  if (!ciTypeDefinition.value) return []

  const allAttributes = [
    ...(ciTypeDefinition.value.required_attributes || []),
    ...(ciTypeDefinition.value.optional_attributes || [])
  ]

  // If fewer than 20 attributes, use a single group
  if (allAttributes.length < 20) {
    return [{
      name: 'all',
      title: 'Attributes',
      collapsed: false,
      attributes: allAttributes
    }]
  }

  // Otherwise, group by attribute category or create logical groups
  // For now, group into "Basic" and "Advanced" based on position
  const midPoint = Math.ceil(allAttributes.length / 2)
  return [
    {
      name: 'basic',
      title: 'Basic Attributes',
      collapsed: false,
      attributes: allAttributes.slice(0, midPoint)
    },
    {
      name: 'advanced',
      title: 'Advanced Attributes',
      collapsed: true,
      attributes: allAttributes.slice(midPoint)
    }
  ]
})

const hasChanges = computed(() => unsavedChanges.value)

// Methods
const updateAttribute = (name: string, value: any) => {
  formData.value.attributes[name] = value
  unsavedChanges.value = true
}

const onFieldValidation = (fieldName: string, isValid: boolean) => {
  fieldValidation.value[fieldName] = isValid
  updateValidationSummary()
}

const updateValidationSummary = () => {
  const errors: ValidationError[] = []

  // Check field validation
  Object.entries(fieldValidation.value).forEach(([field, isValid]) => {
    if (!isValid) {
      errors.push({
        field,
        message: `${field} has validation errors`,
        code: 'validation_error'
      })
    }
  })

  validationErrors.value = errors
}

const toggleGroupCollapse = (groupName: string) => {
  collapsedGroups.value[groupName] = !collapsedGroups.value[groupName]
}

const handleSubmit = async () => {
  loading.value = true
  validationErrors.value = []
  fieldErrors.value = {}

  try {
    const data = {
      name: formData.value.name,
      ci_type: formData.value.ci_type,
      owner: formData.value.owner,
      lifecycle_status_id: formData.value.lifecycle_status_id || undefined,
      attributes: formData.value.attributes,
      tags: formData.value.tags
    }

    if (isEdit.value) {
      const updateData: EAUpdateRequest = {
        name: data.name,
        lifecycle_status_id: data.lifecycle_status_id,
        attributes: data.attributes,
        tags: data.tags
      }
      await eaStore.updateEntity(props.entityId!, updateData)
    } else {
      await eaStore.createEntity(data as EACreateRequest)
    }

    unsavedChanges.value = false

    // Navigate back to list view
    const domain = props.domain || 'business'
    router.push(`/entities/${domain}`)
  } catch (error: any) {
    // Handle validation errors from backend
    if (error.response?.status === 422) {
      const errorDetails = error.response.data?.error?.details
      if (errorDetails) {
        Object.entries(errorDetails).forEach(([field, messages]) => {
          const message = Array.isArray(messages) ? messages.join('. ') : String(messages)
          validationErrors.value.push({
            field,
            message,
            code: 'validation_error'
          })
          fieldErrors.value[field] = message
        })
      }
    } else {
      console.error('Failed to save entity:', error)
    }
  } finally {
    loading.value = false
  }
}

const saveAsDraft = async () => {
  // Find the draft lifecycle status
  const draftStatus = lifecycleStatuses.value.find(s => s.name === 'draft')
  if (draftStatus) {
    formData.value.lifecycle_status_id = draftStatus.id
  }
  await handleSubmit()
}

const handleCancel = () => {
  if (hasChanges.value) {
    if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
      const domain = props.domain || 'business'
      router.push(`/entities/${domain}`)
    }
  } else {
    const domain = props.domain || 'business'
    router.push(`/entities/${domain}`)
  }
}

const loadLifecycleStatuses = async () => {
  try {
    const response = await lifecycleStatusAPI.getActive()
    lifecycleStatuses.value = response.data
  } catch (error) {
    console.error('Failed to load lifecycle statuses:', error)
  }
}

const loadEntity = async () => {
  if (props.entityId) {
    try {
      const entity = await eaStore.fetchEntity(props.entityId)
      if (entity) {
        formData.value = {
          name: entity.name,
          ci_type: entity.ci_type,
          lifecycle_status_id: entity.lifecycle_status_id,
          owner: entity.owner || '',
          attributes: { ...entity.attributes },
          tags: [...entity.tags]
        }
      }
    } catch (error) {
      console.error('Failed to load entity:', error)
    }
  }
}

// Lifecycle
onMounted(async () => {
  console.log('[DynamicFormBuilder] Component mounting, fetching data...')

  await loadLifecycleStatuses()

  // Load CI types if not already loaded
  if (eaTypesStore.ciTypes.length === 0) {
    console.log('[DynamicFormBuilder] Fetching CI types...')
    try {
      await eaTypesStore.fetchCiTypes()
      console.log('[DynamicFormBuilder] CI types loaded:', eaTypesStore.ciTypes.length)
    } catch (err) {
      console.error('[DynamicFormBuilder] Failed to load CI types:', err)
    }
  }

  // Load EA teams if not already loaded
  if (eaTypesStore.teams.length === 0) {
    console.log('[DynamicFormBuilder] Fetching EA teams...')
    try {
      await eaTypesStore.fetchTeams()
      console.log('[DynamicFormBuilder] EA teams loaded:', eaTypesStore.teams.length)
    } catch (err) {
      console.error('[DynamicFormBuilder] Failed to load EA teams:', err)
    }
  }

  // Load entity if editing
  if (props.entityId) {
    await loadEntity()
  }
})

// Watch for unsaved changes
watch(() => formData.value, () => {
  unsavedChanges.value = true
}, { deep: true })

// Warn before leaving with unsaved changes
watch(unsavedChanges, (hasChanges) => {
  if (hasChanges) {
    window.onbeforeunload = () => {
      return 'You have unsaved changes. Are you sure you want to leave?'
    }
  } else {
    window.onbeforeunload = null
  }
})
</script>

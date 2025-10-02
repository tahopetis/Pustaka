<template>
  <div class="page-container page-content" style="max-width: 1024px; margin: 0 auto;">
    <!-- Page header -->
    <div class="page-header">
      <h1 class="page-title">
        {{ isEdit ? 'Edit Configuration Item' : 'Create Configuration Item' }}
      </h1>
      <p class="page-subtitle">
        {{ isEdit ? 'Update the configuration item details' : 'Add a new configuration item to your CMDB' }}
      </p>
    </div>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        <!-- Basic Information -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Basic Information</h3>
          </div>
          <div class="card-body">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="form-label">Name *</label>
                <input
                  v-model="form.name"
                  type="text"
                  required
                  class="form-input"
                  placeholder="Enter CI name"
                  :disabled="loading"
                >
              </div>
              <div>
                <label class="form-label">CI Type *</label>
                <select
                  v-model="form.ci_type"
                  required
                  class="form-input"
                  :disabled="loading || isEdit"
                  @change="onCITypeChange"
                >
                  <option value="">Select CI Type</option>
                  <option v-for="type in ciTypes" :key="type.id" :value="type.name">
                    {{ type.name }} - {{ type.description }}
                  </option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- Attributes -->
        <div v-if="selectedCIType" class="bg-white shadow rounded-lg">
          <div class="px-4 py-5 border-b border-gray-200 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Attributes</h3>
            <p class="mt-1 text-sm text-gray-500">
              Configure the attributes for this {{ selectedCIType.name }}
            </p>
          </div>
          <div class="px-4 py-5 sm:p-6">
            <!-- Required Attributes Section -->
            <div v-if="selectedCIType.required_attributes.length > 0" class="mb-8">
              <h4 class="text-md font-medium text-gray-900 mb-4 flex items-center">
                <Icon name="required" class="w-4 h-4 text-red-500 mr-2" />
                Required Attributes
                <span class="ml-2 px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded">
                  {{ selectedCIType.required_attributes.length }}
                </span>
              </h4>
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <DynamicAttributeField
                  v-for="attr in selectedCIType.required_attributes"
                  :key="`req-${attr.name}`"
                  :attribute="attr"
                  :model-value="form.attributes[attr.name]"
                  @update:model-value="updateAttribute(attr.name, $event)"
                  @validation="onAttributeValidation(attr.name, $event)"
                  :required="true"
                  :disabled="loading"
                />
              </div>
            </div>

            <!-- Optional Attributes Section -->
            <div v-if="selectedCIType.optional_attributes.length > 0">
              <h4 class="text-md font-medium text-gray-900 mb-4 flex items-center">
                <Icon name="optional" class="w-4 h-4 text-blue-500 mr-2" />
                Optional Attributes
                <span class="ml-2 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
                  {{ selectedCIType.optional_attributes.length }}
                </span>
              </h4>
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <DynamicAttributeField
                  v-for="attr in selectedCIType.optional_attributes"
                  :key="`opt-${attr.name}`"
                  :attribute="attr"
                  :model-value="form.attributes[attr.name]"
                  @update:model-value="updateAttribute(attr.name, $event)"
                  @validation="onAttributeValidation(attr.name, $event)"
                  :required="false"
                  :disabled="loading"
                />
              </div>
            </div>

            <!-- Validation Summary -->
            <div v-if="validationErrors.length > 0" class="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
              <h4 class="text-sm font-medium text-red-800 mb-2">Please fix the following errors:</h4>
              <ul class="text-sm text-red-700 space-y-1">
                <li v-for="error in validationErrors" :key="error" class="flex items-center">
                  <Icon name="error" class="w-4 h-4 mr-2" />
                  {{ error }}
                </li>
              </ul>
            </div>

            <!-- Schema Summary -->
            <div v-if="selectedCIType" class="mt-6 bg-gray-50 border border-gray-200 rounded-lg p-4">
              <h4 class="text-sm font-medium text-gray-900 mb-2">Schema Summary</h4>
              <div class="text-sm text-gray-600">
                <p>{{ selectedCIType.name }} includes {{ selectedCIType.required_attributes.length }} required and {{ selectedCIType.optional_attributes.length }} optional attributes</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Tags -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Tags</h3>
            <p class="mt-1 text-sm text-gray-500">Add tags to help categorize and find this CI</p>
          </div>
          <div class="card-body">
            <div class="space-y-2">
              <div v-for="(tag, index) in form.tags" :key="index" class="flex items-center space-x-2">
                <input
                  v-model="form.tags[index]"
                  type="text"
                  class="form-input"
                  placeholder="Enter tag"
                  :disabled="loading"
                >
                <button
                  type="button"
                  @click="removeTag(index)"
                  class="text-red-600 hover:text-red-800"
                  :disabled="loading"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </button>
              </div>
              <button
                type="button"
                @click="addTag"
                class="text-blue-600 hover:text-blue-800 text-sm"
                :disabled="loading"
              >
                + Add Tag
              </button>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end space-x-3">
          <router-link
            to="/ci"
            class="btn btn-outline"
          >
            Cancel
          </router-link>
          <button
            type="submit"
            :disabled="loading || !isFormValid"
            class="btn btn-primary"
          >
            <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
            {{ isEdit ? 'Update' : 'Create' }} Configuration Item
          </button>
        </div>
      </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ciAPI, ciTypeAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'
import DynamicAttributeField from '@/components/ci/DynamicAttributeField.vue'
import Icon from '@/components/base/Icon.vue'
import type { CI, CIType, CreateCIData, UpdateCIData } from '@/types/ci'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const ciTypes = ref<CIType[]>([])
const existingCI = ref<CI | null>(null)
const attributeValidation = ref<Record<string, { isValid: boolean; error?: string }>>({})

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  name: '',
  ci_type: '',
  attributes: {} as Record<string, any>,
  tags: [] as string[],
})

const selectedCIType = computed(() => {
  return ciTypes.value.find(type => type.name === form.ci_type)
})

const validationErrors = computed(() => {
  const errors: string[] = []

  // Check required attributes validation
  if (selectedCIType.value) {
    selectedCIType.value.required_attributes.forEach(attr => {
      const validation = attributeValidation.value[attr.name]
      if (validation && !validation.isValid) {
        errors.push(validation.error || `${attr.name} is invalid`)
      }
    })
  }

  return errors
})

const isFormValid = computed(() => {
  if (!form.name.trim() || !form.ci_type || !selectedCIType.value) {
    return false
  }

  // Check all required attributes are valid
  if (selectedCIType.value) {
    for (const attr of selectedCIType.value.required_attributes) {
      const validation = attributeValidation.value[attr.name]
      if (validation && !validation.isValid) {
        return false
      }
    }
  }

  return true
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const onCITypeChange = () => {
  // Reset attributes and validation when CI type changes
  form.attributes = {}
  attributeValidation.value = {}

  // Set default values for required attributes if available
  if (selectedCIType.value) {
    selectedCIType.value.required_attributes.forEach((attr: any) => {
      if (attr.default !== undefined) {
        form.attributes[attr.name] = attr.default
      }
      // Initialize validation state
      attributeValidation.value[attr.name] = { isValid: !attr.default, error: undefined }
    })

    // Initialize validation for optional attributes
    selectedCIType.value.optional_attributes.forEach((attr: any) => {
      attributeValidation.value[attr.name] = { isValid: true, error: undefined }
    })
  }
}

const updateAttribute = (attributeName: string, value: any) => {
  form.attributes[attributeName] = value
}

const onAttributeValidation = (attributeName: string, validation: { isValid: boolean; error?: string } | boolean, error?: string) => {
  if (typeof validation === 'boolean') {
    attributeValidation.value[attributeName] = { isValid: validation, error }
  } else {
    attributeValidation.value[attributeName] = validation
  }
}

const addTag = () => {
  form.tags.push('')
}

const removeTag = (index: number) => {
  form.tags.splice(index, 1)
}

const loadCITypes = async () => {
  try {
    const response = await ciTypeAPI.list()
    ciTypes.value = response.data.ci_types || []
  } catch (error) {
    console.error('Failed to load CI types:', error)
    showErrorToast('Failed to load CI types')
  }
}

const loadCI = async () => {
  if (!route.params.id) return

  try {
    const response = await ciAPI.get(route.params.id as string)
    existingCI.value = response.data

    // Populate form with existing data
    form.name = existingCI.value.name
    form.ci_type = existingCI.value.ci_type
    form.attributes = { ...existingCI.value.attributes }
    form.tags = existingCI.value.tags ? [...existingCI.value.tags] : []

    // Wait for CI type to be loaded and set up validation
    await new Promise(resolve => setTimeout(resolve, 100))

    // Initialize validation for loaded CI
    if (selectedCIType.value) {
      // Initialize validation for required attributes
      selectedCIType.value.required_attributes.forEach((attr: any) => {
        const hasValue = form.attributes[attr.name] !== undefined && form.attributes[attr.name] !== null && form.attributes[attr.name] !== ''
        attributeValidation.value[attr.name] = {
          isValid: hasValue,
          error: hasValue ? undefined : `${attr.name} is required`
        }
      })

      // Initialize validation for optional attributes
      selectedCIType.value.optional_attributes.forEach((attr: any) => {
        attributeValidation.value[attr.name] = { isValid: true, error: undefined }
      })
    }
  } catch (error: any) {
    console.error('Failed to load CI:', error)
    const message = error.response?.data?.error || 'Failed to load configuration item'
    showErrorToast(message)
    router.push('/ci')
  }
}

const handleSubmit = async () => {
  if (!isFormValid.value) return

  loading.value = true

  try {
    const data = {
      name: form.name.trim(),
      ci_type: form.ci_type,
      attributes: form.attributes,
      tags: form.tags.filter(tag => tag.trim()),
    }

    if (isEdit.value) {
      await ciAPI.update(route.params.id as string, {
        attributes: data.attributes,
        tags: data.tags,
      } as UpdateCIData)
      showSuccessToast('Configuration item updated successfully')
    } else {
      await ciAPI.create(data as CreateCIData)
      showSuccessToast('Configuration item created successfully')
    }

    router.push('/ci')
  } catch (error: any) {
    console.error('Failed to save CI:', error)
    const message = error.response?.data?.error || 'Failed to save configuration item'
    showErrorToast(message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadCITypes()

  if (isEdit.value) {
    await loadCI()
  }
})
</script>
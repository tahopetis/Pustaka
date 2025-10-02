<template>
  <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
    <div class="relative top-20 mx-auto p-5 border w-11/12 md:w-4/5 lg:w-3/4 xl:w-2/3 shadow-lg rounded-lg bg-white">
      <!-- Header -->
      <div class="flex justify-between items-center pb-4 border-b">
        <h3 class="text-lg font-semibold text-gray-900">
          {{ isEdit ? 'Edit CI Type' : 'Create CI Type' }}
        </h3>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <Icon name="x" class="w-6 h-6" />
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="py-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Basic Information -->
          <div class="space-y-4">
            <h4 class="text-md font-medium text-gray-900">Basic Information</h4>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Name *
              </label>
              <BaseInput
                v-model="form.name"
                placeholder="e.g., Server, Application, Database"
                required
                :disabled="isEdit"
              />
              <p class="text-xs text-gray-500 mt-1">
                Unique identifier for this CI type
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <BaseTextarea
                v-model="form.description"
                placeholder="Describe what this CI type represents"
                rows="3"
              />
            </div>
          </div>

          <!-- Schema Preview -->
          <div class="space-y-4">
            <h4 class="text-md font-medium text-gray-900">Schema Preview</h4>
            <div class="bg-gray-50 border rounded-lg p-4 max-h-64 overflow-y-auto">
              <pre class="text-xs text-gray-700">{{ schemaPreview }}</pre>
            </div>
          </div>
        </div>

        <!-- Required Attributes -->
        <div class="mt-6 space-y-4">
          <div class="flex justify-between items-center">
            <h4 class="text-md font-medium text-gray-900">Required Attributes</h4>
            <BaseButton
              type="button"
              @click="addRequiredAttribute"
              variant="outline"
              size="sm"
            >
              <Icon name="plus" class="w-4 h-4 mr-1" />
              Add Required
            </BaseButton>
          </div>

          <div v-if="form.required_attributes.length === 0" class="text-center py-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-500">No required attributes defined</p>
          </div>

          <div v-else class="space-y-3">
            <AttributeEditor
              v-for="(attr, index) in form.required_attributes"
              :key="`req-${index}`"
              :attribute="attr"
              :index="index"
              :is-required="true"
              @update="updateRequiredAttribute"
              @remove="removeRequiredAttribute"
            />
          </div>
        </div>

        <!-- Optional Attributes -->
        <div class="mt-6 space-y-4">
          <div class="flex justify-between items-center">
            <h4 class="text-md font-medium text-gray-900">Optional Attributes</h4>
            <BaseButton
              type="button"
              @click="addOptionalAttribute"
              variant="outline"
              size="sm"
            >
              <Icon name="plus" class="w-4 h-4 mr-1" />
              Add Optional
            </BaseButton>
          </div>

          <div v-if="form.optional_attributes.length === 0" class="text-center py-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-500">No optional attributes defined</p>
          </div>

          <div v-else class="space-y-3">
            <AttributeEditor
              v-for="(attr, index) in form.optional_attributes"
              :key="`opt-${index}`"
              :attribute="attr"
              :index="index"
              :is-required="false"
              @update="updateOptionalAttribute"
              @remove="removeOptionalAttribute"
            />
          </div>
        </div>

        <!-- Form Actions -->
        <div class="mt-8 flex justify-end gap-3 pt-4 border-t">
          <BaseButton
            type="button"
            @click="$emit('close')"
            variant="outline"
          >
            Cancel
          </BaseButton>
          <BaseButton
            type="submit"
            :disabled="isSubmitting"
          >
            <BaseSpinner v-if="isSubmitting" class="w-4 h-4 mr-2" />
            {{ isEdit ? 'Update' : 'Create' }} CI Type
          </BaseButton>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useCITypesStore, type CIType } from '@/stores/ciTypes'
import { useNotificationStore } from '@/stores/notification'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import BaseSpinner from '@/components/base/BaseSpinner.vue'
import Icon from '@/components/base/Icon.vue'
import AttributeEditor from './AttributeEditor.vue'

interface AttributeDefinition {
  name: string
  type: string
  description: string
  validation?: {
    pattern?: string
    min_length?: number
    max_length?: number
    min?: number
    max?: number
    enum?: string[]
    format?: string
  }
}

interface CITypeForm {
  id?: string
  name: string
  description?: string
  required_attributes: AttributeDefinition[]
  optional_attributes: AttributeDefinition[]
}

const props = defineProps<{
  ciType?: CIType
  isEdit?: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const ciTypesStore = useCITypesStore()
const notificationStore = useNotificationStore()

const isSubmitting = ref(false)
const form = ref<CITypeForm>({
  name: '',
  description: '',
  required_attributes: [],
  optional_attributes: []
})

// Helper functions must be defined before they are used in watch
const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    required_attributes: [],
    optional_attributes: []
  }
}

// Initialize form with CI type data if editing - moved after resetForm definition
watch(() => props.ciType, (ciType) => {
  if (ciType) {
    form.value = {
      id: ciType.id,
      name: ciType.name,
      description: ciType.description || '',
      required_attributes: [...ciType.required_attributes],
      optional_attributes: [...ciType.optional_attributes]
    }
  } else {
    resetForm()
  }
}, { immediate: true })

const schemaPreview = computed(() => {
  return JSON.stringify({
    name: form.value.name,
    description: form.value.description,
    required_attributes: form.value.required_attributes,
    optional_attributes: form.value.optional_attributes
  }, null, 2)
})

const addRequiredAttribute = () => {
  form.value.required_attributes.push({
    name: '',
    type: 'string',
    description: '',
    validation: {}
  })
}

const addOptionalAttribute = () => {
  form.value.optional_attributes.push({
    name: '',
    type: 'string',
    description: '',
    validation: {}
  })
}

const updateRequiredAttribute = (index: number, attribute: AttributeDefinition) => {
  form.value.required_attributes[index] = { ...attribute }
}

const updateOptionalAttribute = (index: number, attribute: AttributeDefinition) => {
  form.value.optional_attributes[index] = { ...attribute }
}

const removeRequiredAttribute = (index: number) => {
  form.value.required_attributes.splice(index, 1)
}

const removeOptionalAttribute = (index: number) => {
  form.value.optional_attributes.splice(index, 1)
}

const validateForm = (): string[] => {
  const errors: string[] = []

  if (!form.value.name.trim()) {
    errors.push('Name is required')
  }

  // Check for duplicate attribute names
  const allNames = [
    ...form.value.required_attributes.map(attr => attr.name),
    ...form.value.optional_attributes.map(attr => attr.name)
  ]

  const duplicates = allNames.filter((name, index) => name && allNames.indexOf(name) !== index)
  if (duplicates.length > 0) {
    errors.push(`Duplicate attribute names: ${duplicates.join(', ')}`)
  }

  // Validate each attribute
  const validateAttributes = (attributes: AttributeDefinition[], type: string) => {
    attributes.forEach((attr, index) => {
      if (!attr.name.trim()) {
        errors.push(`${type} attribute ${index + 1}: Name is required`)
      }
      if (attr.name && !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(attr.name)) {
        errors.push(`${type} attribute "${attr.name}": Name must be a valid identifier`)
      }
    })
  }

  validateAttributes(form.value.required_attributes, 'Required')
  validateAttributes(form.value.optional_attributes, 'Optional')

  return errors
}

const cleanValidationData = (validation: any) => {
  if (!validation || typeof validation !== 'object') {
    return {}
  }

  const cleaned: any = {}
  Object.keys(validation).forEach(key => {
    const value = validation[key]
    if (value !== undefined && value !== null && value !== '') {
      cleaned[key] = value
    }
  })

  return Object.keys(cleaned).length > 0 ? cleaned : undefined
}

const cleanAttributeData = (attributes: AttributeDefinition[]): any[] => {
  return attributes.map(attr => ({
    name: attr.name.trim(),
    type: attr.type,
    description: attr.description.trim(),
    validation: cleanValidationData(attr.validation)
  })).filter(attr => attr.name) // Filter out empty attributes
}

const handleSubmit = async () => {
  try {
    // Validate form
    const errors = validateForm()
    if (errors.length > 0) {
      console.error('❌ Validation errors:', errors)
      notificationStore.showError(errors.join('\n'))
      return
    }

    console.log('✅ Form validation passed')
    isSubmitting.value = true

    // Clean and prepare the data
    const cleanedData = {
      description: form.value.description?.trim() || undefined,
      required_attributes: cleanAttributeData(form.value.required_attributes),
      optional_attributes: cleanAttributeData(form.value.optional_attributes)
    }

    console.log('📤 Submitting data:', {
      isEdit: props.isEdit,
      id: props.ciType?.id,
      data: cleanedData
    })

    if (props.isEdit && props.ciType?.id) {
      console.log('📝 Updating existing CI Type')
      await ciTypesStore.updateCIType(props.ciType.id, cleanedData)
      notificationStore.showSuccess('CI Type updated successfully')
    } else {
      console.log('➕ Creating new CI Type')
      // For create, include the name
      await ciTypesStore.createCIType({
        name: form.value.name.trim(),
        ...cleanedData
      })
      notificationStore.showSuccess('CI Type created successfully')
    }

    console.log('✅ CI Type saved successfully')

    // Emit saved event and close modal
    emit('saved')

  } catch (error: any) {
    console.error('❌ Error saving CI type:', error)

    // Enhanced error handling
    let errorMessage = 'Failed to save CI type. Please try again.'

    if (error.response) {
      const status = error.response.status
      const errorData = error.response.data

      console.error('Error response:', { status, data: errorData })

      if (status === 422) {
        const errorDetails = errorData?.error?.details
        if (errorDetails) {
          const errorMessages = Object.values(errorDetails).flat()
          errorMessage = `Validation errors: ${errorMessages.join('. ')}`
        } else {
          errorMessage = 'Validation failed. Please check your input.'
        }
      } else if (status === 409) {
        errorMessage = 'A CI type with this name already exists'
      } else if (status >= 500) {
        const serverMessage = errorData?.error?.message || 'Server error occurred'
        errorMessage = `${serverMessage}. Please try again later.`
      }
    } else if (error.message) {
      if (error.message.includes('already exists')) {
        errorMessage = 'A CI type with this name already exists'
      } else if (error.message.includes('network')) {
        errorMessage = 'Network error. Please check your connection.'
      } else {
        errorMessage = error.message
      }
    }

    notificationStore.showError(errorMessage)

  } finally {
    isSubmitting.value = false
  }
}
</script>
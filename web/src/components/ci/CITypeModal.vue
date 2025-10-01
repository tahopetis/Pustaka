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
import { useCITypesStore } from '@/stores/ciTypes'
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
  name: string
  description?: string
  required_attributes: AttributeDefinition[]
  optional_attributes: AttributeDefinition[]
}

const props = defineProps<{
  ciType?: CITypeForm
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

// Initialize form with CI type data if editing
watch(() => props.ciType, (ciType) => {
  if (ciType) {
    form.value = {
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

const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    required_attributes: [],
    optional_attributes: []
  }
}

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

const handleSubmit = async () => {
  const errors = validateForm()
  if (errors.length > 0) {
    notificationStore.showError(errors.join('\n'))
    return
  }

  try {
    isSubmitting.value = true

    if (props.isEdit && props.ciType) {
      await ciTypesStore.updateCIType(props.ciType.id!, form.value)
    } else {
      await ciTypesStore.createCIType(form.value)
    }

    emit('saved')
  } catch (error: any) {
    console.error('Error saving CI type:', error)
    if (error.message.includes('already exists')) {
      notificationStore.showError('A CI type with this name already exists')
    } else {
      notificationStore.showError('Failed to save CI type')
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>
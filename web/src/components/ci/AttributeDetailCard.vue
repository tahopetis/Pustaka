<template>
  <div class="bg-white border rounded-lg p-4">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <div class="flex items-center gap-2 mb-2">
          <h5 class="font-medium text-gray-900">{{ attribute.name }}</h5>
          <span
            :class="typeBadgeClass"
            class="px-2 py-1 text-xs font-medium rounded"
          >
            {{ attribute.type }}
          </span>
          <span
            v-if="isRequired"
            class="px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded"
          >
            Required
          </span>
        </div>

        <p v-if="attribute.description" class="text-sm text-gray-600 mb-3">
          {{ attribute.description }}
        </p>

        <!-- Validation Rules -->
        <div v-if="hasValidationRules" class="space-y-2">
          <h6 class="text-xs font-medium text-gray-700 uppercase tracking-wide">Validation Rules</h6>

          <div class="flex flex-wrap gap-2">
            <!-- String/Array/Object Length -->
            <span
              v-if="attribute.validation.min_length !== undefined"
              class="inline-flex items-center px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded"
            >
              Min Length: {{ attribute.validation.min_length }}
            </span>
            <span
              v-if="attribute.validation.max_length !== undefined"
              class="inline-flex items-center px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded"
            >
              Max Length: {{ attribute.validation.max_length }}
            </span>

            <!-- Number Range -->
            <span
              v-if="attribute.validation.min !== undefined"
              class="inline-flex items-center px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded"
            >
              Min: {{ attribute.validation.min }}
            </span>
            <span
              v-if="attribute.validation.max !== undefined"
              class="inline-flex items-center px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded"
            >
              Max: {{ attribute.validation.max }}
            </span>

            <!-- Pattern -->
            <span
              v-if="attribute.validation.pattern"
              class="inline-flex items-center px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded"
              title="Regular expression pattern"
            >
              <Icon name="hashtag" class="w-3 h-3 mr-1" />
              Pattern: {{ attribute.validation.pattern }}
            </span>

            <!-- Format -->
            <span
              v-if="attribute.validation.format"
              class="inline-flex items-center px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded"
            >
              <Icon name="check-circle" class="w-3 h-3 mr-1" />
              Format: {{ attribute.validation.format }}
            </span>

            <!-- Enum -->
            <span
              v-if="attribute.validation.enum && attribute.validation.enum.length > 0"
              class="inline-flex items-center px-2 py-1 text-xs bg-green-100 text-green-800 rounded"
              title="Allowed values"
            >
              <Icon name="list" class="w-3 h-3 mr-1" />
              Enum: {{ attribute.validation.enum.join(', ') }}
            </span>
          </div>
        </div>

        <!-- No Validation Rules -->
        <div v-else class="text-xs text-gray-500 italic mt-2">
          No validation rules defined
        </div>
      </div>

      <!-- Type Icon -->
      <div class="ml-4">
        <Icon :name="typeIcon" :class="typeIconClass" class="w-5 h-5" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/base/Icon.vue'

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

const props = defineProps<{
  attribute: AttributeDefinition
  isRequired: boolean
}>()

const typeBadgeClass = computed(() => {
  const typeClasses = {
    string: 'bg-blue-100 text-blue-800',
    integer: 'bg-green-100 text-green-800',
    boolean: 'bg-purple-100 text-purple-800',
    array: 'bg-orange-100 text-orange-800',
    object: 'bg-gray-100 text-gray-800'
  }
  return typeClasses[props.attribute.type as keyof typeof typeClasses] || 'bg-gray-100 text-gray-800'
})

const typeIcon = computed(() => {
  const typeIcons = {
    string: 'text',
    integer: 'hashtag',
    boolean: 'toggle',
    array: 'list',
    object: 'cube'
  }
  return typeIcons[props.attribute.type as keyof typeof typeIcons] || 'question-mark'
})

const typeIconClass = computed(() => {
  const typeIconClasses = {
    string: 'text-blue-600',
    integer: 'text-green-600',
    boolean: 'text-purple-600',
    array: 'text-orange-600',
    object: 'text-gray-600'
  }
  return typeIconClasses[props.attribute.type as keyof typeof typeIconClasses] || 'text-gray-600'
})

const hasValidationRules = computed(() => {
  if (!props.attribute.validation) return false

  const validation = props.attribute.validation
  return (
    validation.pattern !== undefined ||
    validation.min_length !== undefined ||
    validation.max_length !== undefined ||
    validation.min !== undefined ||
    validation.max !== undefined ||
    (validation.enum && validation.enum.length > 0) ||
    validation.format !== undefined
  )
})
</script>
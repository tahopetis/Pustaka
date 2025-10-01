<template>
  <div class="space-y-4">
    <!-- Required Attributes -->
    <div v-if="requiredAttributes.length > 0">
      <h4 class="text-md font-medium text-gray-900 mb-3 flex items-center">
        <Icon name="required" class="w-4 h-4 text-red-500 mr-2" />
        Required Attributes
        <span class="ml-2 px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded">
          {{ requiredAttributes.length }}
        </span>
      </h4>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="attr in requiredAttributes"
          :key="attr.name"
          class="bg-gray-50 border border-gray-200 rounded-lg p-4"
        >
          <div class="flex items-center justify-between mb-2">
            <h5 class="text-sm font-medium text-gray-900">{{ attr.name }}</h5>
            <span
              :class="getTypeBadgeClass(attr.type)"
              class="px-2 py-1 text-xs font-medium rounded"
            >
              {{ attr.type }}
            </span>
          </div>

          <div v-if="hasAttributeDescription(attr)" class="mb-3">
            <p class="text-xs text-gray-600">{{ attr.description }}</p>
          </div>

          <div class="space-y-1">
            <div class="text-sm text-gray-700">
              <strong>Value:</strong>
              <span class="ml-2 font-mono text-xs">{{ displayAttributeValue(attr, ci.attributes[attr.name]) }}</span>
            </div>

            <div v-if="showValidationRules(attr)" class="text-xs text-gray-500">
              <strong>Validation:</strong>
              <div class="ml-2 mt-1 space-y-1">
                <span
                  v-if="attr.validation?.min_length !== undefined || attr.validation?.max_length !== undefined"
                  class="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded"
                >
                  Length: {{ formatLengthValidation(attr.validation) }}
                </span>
                <span
                  v-else-if="attr.validation?.min !== undefined || attr.validation?.max !== undefined"
                  class="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded"
                >
                  Range: {{ formatRangeValidation(attr.validation) }}
                </span>
                <span
                  v-else-if="attr.validation?.pattern"
                  class="inline-block bg-purple-100 text-purple-700 px-2 py-1 rounded"
                  title="Pattern"
                >
                  <Icon name="hashtag" class="w-3 h-3 mr-1 inline" />
                  Pattern
                </span>
                <span
                  v-else-if="attr.validation?.format"
                  class="inline-block bg-blue-100 text-blue-700 px-2 py-1 rounded"
                  title="Format"
                >
                  <Icon name="check-circle" class="w-3 h-3 mr-1 inline" />
                  {{ attr.validation.format }}
                </span>
                <span
                  v-else-if="attr.validation?.enum && attr.validation.enum.length > 0"
                  class="inline-block bg-green-100 text-green-700 px-2 py-1 rounded"
                  title="Enum values"
                >
                  <Icon name="list" class="w-3 h-3 mr-1 inline" />
                  Enum: {{ attr.validation.enum.length }} values
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Optional Attributes -->
    <div v-if="optionalAttributes.length > 0">
      <h4 class="text-md font-medium text-gray-900 mb-3 flex items-center">
        <Icon name="optional" class="w-4 h-4 text-blue-500 mr-2" />
        Optional Attributes
        <span class="ml-2 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
          {{ optionalAttributes.filter(attr => ci.attributes[attr.name] !== undefined && ci.attributes[attr.name] !== null && ci.attributes[attr.name] !== '').length }}/{{ optionalAttributes.length }}
        </span>
      </h4>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="attr in optionalAttributes"
          :key="attr.name"
          class="bg-gray-50 border border-gray-200 rounded-lg p-4"
          :class="{ 'opacity-60': !hasAttributeValue(attr) }"
        >
          <div class="flex items-center justify-between mb-2">
            <h5 class="text-sm font-medium text-gray-900">
              {{ attr.name }}
              <span v-if="!hasAttributeValue(attr)" class="text-gray-400 text-xs ml-1">(not set)</span>
            </h5>
            <span
              :class="getTypeBadgeClass(attr.type)"
              class="px-2 py-1 text-xs font-medium rounded"
            >
              {{ attr.type }}
            </span>
          </div>

          <div v-if="hasAttributeDescription(attr)" class="mb-3">
            <p class="text-xs text-gray-600">{{ attr.description }}</p>
          </div>

          <div v-if="hasAttributeValue(attr)" class="space-y-1">
            <div class="text-sm text-gray-700">
              <strong>Value:</strong>
              <span class="ml-2 font-mono text-xs">{{ displayAttributeValue(attr, ci.attributes[attr.name]) }}</span>
            </div>

            <div v-if="showValidationRules(attr)" class="text-xs text-gray-500">
              <strong>Validation:</strong>
              <div class="ml-2 mt-1 space-y-1">
                <span
                  v-if="attr.validation?.min_length !== undefined || attr.validation?.max_length !== undefined"
                  class="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded"
                >
                  Length: {{ formatLengthValidation(attr.validation) }}
                </span>
                <span
                  v-else-if="attr.validation?.min !== undefined || attr.validation?.max !== undefined"
                  class="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded"
                >
                  Range: {{ formatRangeValidation(attr.validation) }}
                </span>
                <span
                  v-else-if="attr.validation?.pattern"
                  class="inline-block bg-purple-100 text-purple-700 px-2 py-1 rounded"
                  title="Pattern"
                >
                  <Icon name="hashtag" class="w-3 h-3 mr-1 inline" />
                  Pattern
                </span>
                <span
                  v-else-if="attr.validation?.format"
                  class="inline-block bg-blue-100 text-blue-700 px-2 py-1 rounded"
                  title="Format"
                >
                  <Icon name="check-circle" class="w-3 h-3 mr-1 inline" />
                  {{ attr.validation.format }}
                </span>
                <span
                  v-else-if="attr.validation?.enum && attr.validation.enum.length > 0"
                  class="inline-block bg-green-100 text-green-700 px-2 py-1 rounded"
                  title="Enum values"
                >
                  <Icon name="list" class="w-3 h-3 mr-1 inline" />
                  Enum: {{ attr.validation.enum.length }} values
                </span>
              </div>
            </div>
          </div>
          <div v-else class="text-xs text-gray-400 italic">
            No value set
          </div>
        </div>
      </div>
    </div>

    <!-- Raw JSON View -->
    <div class="mt-6">
      <div class="flex items-center justify-between mb-3">
        <h4 class="text-md font-medium text-gray-900 flex items-center">
          <Icon name="document-text" class="w-4 h-4 text-gray-600 mr-2" />
          Raw Attributes (JSON)
        </h4>
        <BaseButton
          @click="copyAttributesToClipboard"
          variant="outline"
          size="sm"
        >
          <Icon name="clipboard" class="w-4 h-4 mr-1" />
          Copy JSON
        </BaseButton>
      </div>
      <div class="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto">
        <pre class="text-xs font-mono whitespace-pre-wrap">{{ formattedAttributes }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/base/Icon.vue'
import { useNotificationStore } from '@/stores/notification'

interface AttributeDefinition {
  name: string
  type: string
  description?: string
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

interface CI {
  id: string
  name: string
  ci_type: string
  attributes: Record<string, any>
  tags: string[]
  created_at: string
  updated_at: string
}

interface Props {
  ci: CI
  ciType: {
    required_attributes: AttributeDefinition[]
    optional_attributes: AttributeDefinition[]
  }
}

const props = defineProps<Props>()

const notificationStore = useNotificationStore()

const requiredAttributes = computed(() => props.ciType.required_attributes)
const optionalAttributes = computed(() => props.ciType.optional_attributes)

const formattedAttributes = computed(() => {
  return JSON.stringify(props.ci.attributes, null, 2)
})

const hasAttributeValue = (attr: AttributeDefinition) => {
  const value = props.ci.attributes[attr.name]
  return value !== undefined && value !== null && value !== ''
}

const hasAttributeDescription = (attr: AttributeDefinition) => {
  return attr.description && attr.description.trim() !== ''
}

const showValidationRules = (attr: AttributeDefinition) => {
  if (!attr.validation) return false

  return (
    attr.validation.pattern !== undefined ||
    attr.validation.min_length !== undefined ||
    attr.validation.max_length !== undefined ||
    attr.validation.min !== undefined ||
    attr.validation.max !== undefined ||
    (attr.validation.enum && attr.validation.enum.length > 0) ||
    attr.validation.format !== undefined
  )
}

const displayAttributeValue = (attr: AttributeDefinition, value: any) => {
  if (value === undefined || value === null || value === '') {
    return '<empty>'
  }

  if (attr.type === 'boolean') {
    return value ? 'true' : 'false'
  }

  if (attr.type === 'array' || attr.type === 'object') {
    return JSON.stringify(value)
  }

  return String(value)
}

const getTypeBadgeClass = (type: string) => {
  const typeClasses = {
    string: 'bg-blue-100 text-blue-800',
    integer: 'bg-green-100 text-green-800',
    boolean: 'bg-purple-100 text-purple-800',
    array: 'bg-orange-100 text-orange-800',
    object: 'bg-gray-100 text-gray-800'
  }
  return typeClasses[type as keyof typeof typeClasses] || 'bg-gray-100 text-gray-800'
}

const formatLengthValidation = (validation: any) => {
  const parts = []
  if (validation.min_length !== undefined) parts.push(`min ${validation.min_length}`)
  if (validation.max_length !== undefined) parts.push(`max ${validation.max_length}`)
  return parts.join(' - ')
}

const formatRangeValidation = (validation: any) => {
  const parts = []
  if (validation.min !== undefined) parts.push(`min ${validation.min}`)
  if (validation.max !== undefined) parts.push(`max ${validation.max}`)
  return parts.join(' - ')
}

const copyAttributesToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(formattedAttributes.value)
    notificationStore.showSuccess('Attributes copied to clipboard')
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
    notificationStore.showError('Failed to copy attributes to clipboard')
  }
}
</script>
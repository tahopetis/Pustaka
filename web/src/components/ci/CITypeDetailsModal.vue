<template>
  <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
    <div class="relative top-20 mx-auto p-5 border w-11/12 md:w-4/5 lg:w-3/4 shadow-lg rounded-lg bg-white">
      <!-- Header -->
      <div class="flex justify-between items-center pb-4 border-b">
        <div>
          <h3 class="text-lg font-semibold text-gray-900">
            {{ ciType.name }}
          </h3>
          <p class="text-sm text-gray-500">
            CI Type Schema Details
          </p>
        </div>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <Icon name="x" class="w-6 h-6" />
        </button>
      </div>

      <!-- Content -->
      <div class="py-6 space-y-6">
        <!-- Basic Information -->
        <div>
          <h4 class="text-md font-medium text-gray-900 mb-3">Basic Information</h4>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">Name</label>
              <div class="mt-1 text-sm text-gray-900">{{ ciType.name }}</div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">ID</label>
              <div class="mt-1 text-sm text-gray-500 font-mono">{{ ciType.id }}</div>
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium text-gray-700">Description</label>
              <div class="mt-1 text-sm text-gray-900">
                {{ ciType.description || 'No description provided' }}
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">Created</label>
              <div class="mt-1 text-sm text-gray-900">
                {{ formatDate(ciType.created_at) }}
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">Last Updated</label>
              <div class="mt-1 text-sm text-gray-900">
                {{ formatDate(ciType.updated_at) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Required Attributes -->
        <div>
          <h4 class="text-md font-medium text-gray-900 mb-3">
            Required Attributes
            <span class="ml-2 px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded">
              {{ ciType.required_attributes.length }}
            </span>
          </h4>

          <div v-if="ciType.required_attributes.length === 0" class="text-center py-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-500">No required attributes defined</p>
          </div>

          <div v-else class="space-y-3">
            <AttributeDetailCard
              v-for="(attr, index) in ciType.required_attributes"
              :key="`req-${index}`"
              :attribute="attr"
              :is-required="true"
            />
          </div>
        </div>

        <!-- Optional Attributes -->
        <div>
          <h4 class="text-md font-medium text-gray-900 mb-3">
            Optional Attributes
            <span class="ml-2 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
              {{ ciType.optional_attributes.length }}
            </span>
          </h4>

          <div v-if="ciType.optional_attributes.length === 0" class="text-center py-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-500">No optional attributes defined</p>
          </div>

          <div v-else class="space-y-3">
            <AttributeDetailCard
              v-for="(attr, index) in ciType.optional_attributes"
              :key="`opt-${index}`"
              :attribute="attr"
              :is-required="false"
            />
          </div>
        </div>

        <!-- JSON Schema -->
        <div>
          <h4 class="text-md font-medium text-gray-900 mb-3">Complete Schema (JSON)</h4>
          <div class="bg-gray-50 border rounded-lg p-4">
            <pre class="text-xs text-gray-700 overflow-x-auto whitespace-pre-wrap">{{ jsonSchema }}</pre>
          </div>
          <div class="mt-2 flex justify-end">
            <BaseButton
              @click="copyToClipboard"
              variant="outline"
              size="sm"
            >
              <Icon name="clipboard" class="w-4 h-4 mr-1" />
              Copy Schema
            </BaseButton>
          </div>
        </div>

        <!-- Usage Statistics -->
        <div v-if="usageCount !== undefined">
          <h4 class="text-md font-medium text-gray-900 mb-3">Usage Statistics</h4>
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div class="flex items-center">
              <Icon name="cube" class="w-5 h-5 text-blue-600 mr-2" />
              <span class="text-sm text-blue-900">
                This CI type is used by <strong>{{ usageCount }}</strong> configuration item(s)
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end pt-4 border-t">
        <BaseButton
          @click="$emit('close')"
          variant="outline"
        >
          Close
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useNotificationStore } from '@/stores/notification'
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/base/Icon.vue'
import AttributeDetailCard from './AttributeDetailCard.vue'

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

interface CIType {
  id: string
  name: string
  description?: string
  required_attributes: AttributeDefinition[]
  optional_attributes: AttributeDefinition[]
  created_at: string
  updated_at: string
}

const props = defineProps<{
  ciType: CIType
  usageCount?: number
}>()

const emit = defineEmits<{
  close: []
}>()

const notificationStore = useNotificationStore()

const jsonSchema = computed(() => {
  return JSON.stringify({
    name: props.ciType.name,
    description: props.ciType.description,
    required_attributes: props.ciType.required_attributes,
    optional_attributes: props.ciType.optional_attributes,
    created_at: props.ciType.created_at,
    updated_at: props.ciType.updated_at
  }, null, 2)
})

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(jsonSchema.value)
    notificationStore.showSuccess('Schema copied to clipboard')
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
    notificationStore.showError('Failed to copy schema to clipboard')
  }
}
</script>
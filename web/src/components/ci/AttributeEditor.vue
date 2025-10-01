<template>
  <div class="bg-white border rounded-lg p-4 space-y-4">
    <div class="flex justify-between items-start">
      <div class="flex-1 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Attribute Name -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Name *
          </label>
          <BaseInput
            v-model="localAttribute.name"
            placeholder="e.g., hostname, ip_address"
            required
            @input="emitUpdate"
          />
        </div>

        <!-- Attribute Type -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Type *
          </label>
          <BaseSelect
            v-model="localAttribute.type"
            :options="attributeTypes"
            @change="onTypeChange"
          />
        </div>

        <!-- Description -->
        <div class="md:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <BaseInput
            v-model="localAttribute.description"
            placeholder="Describe this attribute"
            @input="emitUpdate"
          />
        </div>
      </div>

      <button
        type="button"
        @click="$emit('remove', index)"
        class="ml-4 text-red-600 hover:text-red-800 transition-colors"
        title="Remove attribute"
      >
        <Icon name="trash" class="w-5 h-5" />
      </button>
    </div>

    <!-- Validation Rules -->
    <div v-if="showValidation" class="border-t pt-4">
      <div class="flex justify-between items-center mb-3">
        <h5 class="text-sm font-medium text-gray-900">Validation Rules</h5>
        <BaseButton
          type="button"
          @click="toggleValidation"
          variant="outline"
          size="sm"
        >
          Hide Validation
        </BaseButton>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <!-- String/Array/Object Length Validation -->
        <div v-if="hasLengthValidation">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Min Length
          </label>
          <BaseInput
            v-model.number="localAttribute.validation.min_length"
            type="number"
            min="0"
            placeholder="0"
            @input="emitUpdate"
          />
        </div>

        <div v-if="hasLengthValidation">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Max Length
          </label>
          <BaseInput
            v-model.number="localAttribute.validation.max_length"
            type="number"
            min="0"
            placeholder="255"
            @input="emitUpdate"
          />
        </div>

        <!-- Number Range Validation -->
        <div v-if="localAttribute.type === 'integer'">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Min Value
          </label>
          <BaseInput
            v-model.number="localAttribute.validation.min"
            type="number"
            placeholder="0"
            @input="emitUpdate"
          />
        </div>

        <div v-if="localAttribute.type === 'integer'">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Max Value
          </label>
          <BaseInput
            v-model.number="localAttribute.validation.max"
            type="number"
            placeholder="100"
            @input="emitUpdate"
          />
        </div>

        <!-- Pattern Validation -->
        <div v-if="localAttribute.type === 'string'">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Pattern (Regex)
          </label>
          <BaseInput
            v-model="localAttribute.validation.pattern"
            placeholder="^[a-zA-Z0-9-]+$"
            @input="emitUpdate"
          />
        </div>

        <!-- Format Validation -->
        <div v-if="localAttribute.type === 'string'">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Format
          </label>
          <BaseSelect
            v-model="localAttribute.validation.format"
            :options="formatOptions"
            placeholder="No format"
            @change="emitUpdate"
          />
        </div>

        <!-- Enum Validation -->
        <div v-if="supportsEnum" class="md:col-span-2 lg:col-span-3">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Allowed Values (comma-separated)
          </label>
          <BaseInput
            v-model="enumInput"
            placeholder="value1, value2, value3"
            @input="updateEnum"
          />
          <p class="text-xs text-gray-500 mt-1">
            Enter comma-separated values for enumerated options
          </p>
        </div>
      </div>
    </div>

    <!-- Show Validation Button -->
    <div v-else class="text-center">
      <BaseButton
        type="button"
        @click="toggleValidation"
        variant="outline"
        size="sm"
      >
        <Icon name="shield-check" class="w-4 h-4 mr-1" />
        Add Validation Rules
      </BaseButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import BaseButton from '@/components/base/BaseButton.vue'
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
  index: number
  isRequired: boolean
}>()

const emit = defineEmits<{
  update: [index: number, attribute: AttributeDefinition]
  remove: [index: number]
}>()

const showValidation = ref(false)
const enumInput = ref('')

const localAttribute = ref<AttributeDefinition>({ ...props.attribute })

// Initialize validation object if it doesn't exist
if (!localAttribute.value.validation) {
  localAttribute.value.validation = {}
}

const attributeTypes = [
  { value: 'string', label: 'String' },
  { value: 'integer', label: 'Integer' },
  { value: 'boolean', label: 'Boolean' },
  { value: 'array', label: 'Array' },
  { value: 'object', label: 'Object' }
]

const formatOptions = [
  { value: '', label: 'No format' },
  { value: 'email', label: 'Email' },
  { value: 'url', label: 'URL' },
  { value: 'ipv4', label: 'IPv4 Address' },
  { value: 'date', label: 'Date (YYYY-MM-DD)' },
  { value: 'datetime', label: 'Datetime (ISO 8601)' }
]

const hasLengthValidation = computed(() => {
  return ['string', 'array', 'object'].includes(localAttribute.value.type)
})

const supportsEnum = computed(() => {
  return ['string', 'integer', 'array'].includes(localAttribute.value.type)
})

// Watch for changes from props only (not from local changes)
watch(() => props.attribute, (newAttribute) => {
  // Prevent infinite loops by checking if the change is coming from outside
  if (JSON.stringify(newAttribute) !== JSON.stringify(localAttribute.value)) {
    localAttribute.value = { ...newAttribute }
    if (!localAttribute.value.validation) {
      localAttribute.value.validation = {}
    }
    updateEnumInput()
  }
}, { deep: true })

// Only emit updates when user interacts with inputs, not on every internal change
const emitUpdate = () => {
  emit('update', props.index, { ...localAttribute.value })
}

const toggleValidation = () => {
  showValidation.value = !showValidation.value
  if (!showValidation.value) {
    // Clear validation when hiding
    localAttribute.value.validation = {}
  }
}

const onTypeChange = () => {
  // Clear validation rules that don't apply to the new type
  if (localAttribute.value.validation) {
    if (!hasLengthValidation.value) {
      delete localAttribute.value.validation.min_length
      delete localAttribute.value.validation.max_length
    }
    if (localAttribute.value.type !== 'integer') {
      delete localAttribute.value.validation.min
      delete localAttribute.value.validation.max
    }
    if (localAttribute.value.type !== 'string') {
      delete localAttribute.value.validation.pattern
      delete localAttribute.value.validation.format
    }
    if (!supportsEnum.value) {
      delete localAttribute.value.validation.enum
    }
  }
  emitUpdate()
}

const updateEnumInput = () => {
  if (localAttribute.value.validation?.enum) {
    enumInput.value = localAttribute.value.validation.enum.join(', ')
  } else {
    enumInput.value = ''
  }
}

const updateEnum = () => {
  const values = enumInput.value
    .split(',')
    .map(v => v.trim())
    .filter(v => v.length > 0)

  if (values.length > 0) {
    if (!localAttribute.value.validation) {
      localAttribute.value.validation = {}
    }
    localAttribute.value.validation.enum = values
  } else {
    if (localAttribute.value.validation) {
      delete localAttribute.value.validation.enum
    }
  }
  emitUpdate()
}

// Initialize enum input
updateEnumInput()
</script>
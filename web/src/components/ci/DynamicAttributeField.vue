<template>
  <div class="space-y-2">
    <label class="block text-sm font-medium text-gray-700">
      {{ attribute.name }}
      <span v-if="required" class="text-red-500 ml-1">*</span>
      <span class="text-gray-400 text-xs ml-1">({{ attribute.type }})</span>
    </label>

    <p v-if="attribute.description" class="text-xs text-gray-500 mb-1">
      {{ attribute.description }}
    </p>

    <!-- String Input -->
    <BaseInput
      v-if="attribute.type === 'string'"
      :model-value="modelValue"
      @update:model-value="updateValue"
      :placeholder="getPlaceholder()"
      :disabled="disabled"
      :required="required"
      :error="!!error"
    />

    <!-- Integer Input -->
    <BaseInput
      v-else-if="attribute.type === 'integer'"
      :model-value="modelValue"
      @update:model-value="updateValue"
      type="number"
      :placeholder="getPlaceholder()"
      :disabled="disabled"
      :required="required"
      :error="!!error"
      @blur="validateInteger"
    />

    <!-- Boolean Select -->
    <BaseSelect
      v-else-if="attribute.type === 'boolean'"
      :model-value="modelValue?.toString()"
      @update:model-value="updateBooleanValue"
      :disabled="disabled"
      :required="required"
      :error="!!error"
    >
      <option value="">Select {{ attribute.name }}</option>
      <option value="true">True</option>
      <option value="false">False</option>
    </BaseSelect>

    <!-- Enum Select -->
    <BaseSelect
      v-else-if="attribute.type === 'string' && hasEnum"
      :model-value="modelValue"
      @update:model-value="updateValue"
      :disabled="disabled"
      :required="required"
      :error="!!error"
    >
      <option value="">Select {{ attribute.name }}</option>
      <option v-for="option in attribute.validation?.enum" :key="option" :value="option">
        {{ option }}
      </option>
    </BaseSelect>

    <!-- Array Input (JSON format) -->
    <BaseTextarea
      v-else-if="attribute.type === 'array'"
      :model-value="modelValue"
      @update:model-value="updateArrayValue"
      :placeholder="getPlaceholder()"
      :rows="4"
      :disabled="disabled"
      :required="required"
      :error="!!error"
      @blur="validateArray"
    />

    <!-- Object Input (JSON format) -->
    <BaseTextarea
      v-else-if="attribute.type === 'object'"
      :model-value="modelValue"
      @update:model-value="updateObjectValue"
      :placeholder="getPlaceholder()"
      :rows="6"
      :disabled="disabled"
      :required="required"
      :error="!!error"
      @blur="validateObject"
    />

    <!-- Fallback for unknown types -->
    <BaseInput
      v-else
      :model-value="modelValue"
      @update:model-value="updateValue"
      :placeholder="getPlaceholder()"
      :disabled="disabled"
      :required="required"
      :error="!!error"
    />

    <!-- Validation Help Text -->
    <div v-if="hasValidationRules" class="text-xs text-gray-500">
      <span v-if="attribute.validation?.min_length !== undefined || attribute.validation?.max_length !== undefined">
        Length:
        <span v-if="attribute.validation?.min_length !== undefined">min {{ attribute.validation.min_length }}</span>
        <span v-if="attribute.validation?.min_length !== undefined && attribute.validation?.max_length !== undefined"> - </span>
        <span v-if="attribute.validation?.max_length !== undefined">max {{ attribute.validation.max_length }}</span>
      </span>
      <span v-else-if="attribute.validation?.min !== undefined || attribute.validation?.max !== undefined">
        Value:
        <span v-if="attribute.validation?.min !== undefined">min {{ attribute.validation.min }}</span>
        <span v-if="attribute.validation?.min !== undefined && attribute.validation?.max !== undefined"> - </span>
        <span v-if="attribute.validation?.max !== undefined">max {{ attribute.validation.max }}</span>
      </span>
      <span v-else-if="attribute.validation?.pattern" class="text-purple-600">
        Pattern: {{ attribute.validation.pattern }}
      </span>
      <span v-else-if="attribute.validation?.format" class="text-blue-600">
        Format: {{ attribute.validation.format }}
      </span>
      <span v-else-if="hasEnum" class="text-green-600">
        Values: {{ attribute.validation?.enum?.join(', ') }}
      </span>
    </div>

    <!-- Error Message -->
    <p v-if="error" class="text-xs text-red-600 mt-1">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'

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

interface Props {
  attribute: AttributeDefinition
  modelValue: any
  required?: boolean
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
  disabled: false
})

const emit = defineEmits<{
  'update:modelValue': [value: any]
  'validation': [isValid: boolean, error?: string]
}>()

const error = ref<string>('')

const hasEnum = computed(() => {
  return props.attribute.validation?.enum && props.attribute.validation.enum.length > 0
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

const getPlaceholder = () => {
  if (props.attribute.type === 'array') {
    return 'Enter JSON array, e.g., ["item1", "item2"]'
  }
  if (props.attribute.type === 'object') {
    return 'Enter JSON object, e.g., {"key": "value"}'
  }
  return `Enter ${props.attribute.name}`
}

const updateValue = (value: any) => {
  error.value = ''
  emit('update:modelValue', value)
  validateValue(value)
}

const updateBooleanValue = (value: string) => {
  const boolValue = value === 'true' ? true : value === 'false' ? false : null
  error.value = ''
  emit('update:modelValue', boolValue)
  emit('validation', boolValue !== null)
}

const updateArrayValue = (value: string) => {
  error.value = ''
  emit('update:modelValue', value)
  emit('validation', true) // Will validate on blur
}

const updateObjectValue = (value: string) => {
  error.value = ''
  emit('update:modelValue', value)
  emit('validation', true) // Will validate on blur
}

const validateInteger = () => {
  const value = props.modelValue
  if (value === null || value === undefined || value === '') {
    if (props.required) {
      error.value = `${props.attribute.name} is required`
      emit('validation', false, error.value)
    } else {
      emit('validation', true)
    }
    return
  }

  const numValue = Number(value)
  if (isNaN(numValue)) {
    error.value = `${props.attribute.name} must be a valid number`
    emit('validation', false, error.value)
    return
  }

  const validation = props.attribute.validation
  if (validation) {
    if (validation.min !== undefined && numValue < validation.min) {
      error.value = `${props.attribute.name} must be at least ${validation.min}`
      emit('validation', false, error.value)
      return
    }
    if (validation.max !== undefined && numValue > validation.max) {
      error.value = `${props.attribute.name} must be at most ${validation.max}`
      emit('validation', false, error.value)
      return
    }
  }

  emit('validation', true)
}

const validateArray = () => {
  const value = props.modelValue
  if (!value || value.trim() === '') {
    if (props.required) {
      error.value = `${props.attribute.name} is required`
      emit('validation', false, error.value)
    } else {
      emit('validation', true)
    }
    return
  }

  try {
    const parsed = JSON.parse(value)
    if (!Array.isArray(parsed)) {
      error.value = `${props.attribute.name} must be a valid JSON array`
      emit('validation', false, error.value)
      return
    }

    const validation = props.attribute.validation
    if (validation) {
      if (validation.min_length !== undefined && parsed.length < validation.min_length) {
        error.value = `${props.attribute.name} must have at least ${validation.min_length} items`
        emit('validation', false, error.value)
        return
      }
      if (validation.max_length !== undefined && parsed.length > validation.max_length) {
        error.value = `${props.attribute.name} must have at most ${validation.max_length} items`
        emit('validation', false, error.value)
        return
      }
    }

    // Update with parsed value
    emit('update:modelValue', parsed)
    emit('validation', true)
  } catch (e) {
    error.value = `${props.attribute.name} must be valid JSON array format`
    emit('validation', false, error.value)
  }
}

const validateObject = () => {
  const value = props.modelValue
  if (!value || value.trim() === '') {
    if (props.required) {
      error.value = `${props.attribute.name} is required`
      emit('validation', false, error.value)
    } else {
      emit('validation', true)
    }
    return
  }

  try {
    const parsed = JSON.parse(value)
    if (typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
      error.value = `${props.attribute.name} must be a valid JSON object`
      emit('validation', false, error.value)
      return
    }

    const validation = props.attribute.validation
    if (validation) {
      if (validation.min_length !== undefined && Object.keys(parsed).length < validation.min_length) {
        error.value = `${props.attribute.name} must have at least ${validation.min_length} properties`
        emit('validation', false, error.value)
        return
      }
      if (validation.max_length !== undefined && Object.keys(parsed).length > validation.max_length) {
        error.value = `${props.attribute.name} must have at most ${validation.max_length} properties`
        emit('validation', false, error.value)
        return
      }
    }

    // Update with parsed value
    emit('update:modelValue', parsed)
    emit('validation', true)
  } catch (e) {
    error.value = `${props.attribute.name} must be valid JSON object format`
    emit('validation', false, error.value)
  }
}

const validateValue = (value: any) => {
  if (props.attribute.type !== 'string') return

  if (!value || value.trim() === '') {
    if (props.required) {
      error.value = `${props.attribute.name} is required`
      emit('validation', false, error.value)
    } else {
      emit('validation', true)
    }
    return
  }

  const validation = props.attribute.validation
  if (!validation) {
    emit('validation', true)
    return
  }

  // Length validation
  if (validation.min_length !== undefined && value.length < validation.min_length) {
    error.value = `${props.attribute.name} must be at least ${validation.min_length} characters`
    emit('validation', false, error.value)
    return
  }
  if (validation.max_length !== undefined && value.length > validation.max_length) {
    error.value = `${props.attribute.name} must be at most ${validation.max_length} characters`
    emit('validation', false, error.value)
    return
  }

  // Pattern validation
  if (validation.pattern) {
    try {
      const regex = new RegExp(validation.pattern)
      if (!regex.test(value)) {
        error.value = `${props.attribute.name} format is invalid`
        emit('validation', false, error.value)
        return
      }
    } catch (e) {
      console.warn('Invalid regex pattern:', validation.pattern)
    }
  }

  // Format validation
  if (validation.format) {
    if (!validateFormat(value, validation.format)) {
      error.value = `${props.attribute.name} must be a valid ${validation.format}`
      emit('validation', false, error.value)
      return
    }
  }

  // Enum validation
  if (validation.enum && validation.enum.length > 0) {
    if (!validation.enum.includes(value)) {
      error.value = `${props.attribute.name} must be one of: ${validation.enum.join(', ')}`
      emit('validation', false, error.value)
      return
    }
  }

  emit('validation', true)
}

const validateFormat = (value: string, format: string): boolean => {
  switch (format) {
    case 'email':
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
    case 'url':
      try {
        new URL(value)
        return true
      } catch {
        return false
      }
    case 'ipv4':
      return /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(value)
    case 'date':
      return /^\d{4}-\d{2}-\d{2}$/.test(value)
    case 'datetime':
      return !isNaN(Date.parse(value))
    default:
      return true
  }
}

// Watch for changes in the attribute definition (when switching CI types)
watch(() => props.attribute, () => {
  error.value = ''
}, { deep: true })

// Initialize validation
watch(() => props.modelValue, (newValue) => {
  if (newValue !== undefined && newValue !== null && newValue !== '') {
    if (props.attribute.type === 'string') {
      validateValue(newValue)
    } else if (props.attribute.type === 'integer') {
      validateInteger()
    }
  }
}, { immediate: true })
</script>
<template>
  <select
    :id="id"
    :value="modelValue"
    :disabled="disabled"
    :required="required"
    :class="selectClasses"
    @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    @blur="$emit('blur')"
    @focus="$emit('focus')"
  >
    <option v-if="placeholder && !modelValue" value="">{{ placeholder }}</option>
    <option
      v-for="option in normalizedOptions"
      :key="option.value"
      :value="option.value"
      :disabled="option.disabled"
    >
      {{ option.label }}
    </option>
    <slot />
  </select>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Option {
  value: string | number
  label: string
  disabled?: boolean
}

interface Props {
  id?: string
  modelValue?: string | number
  options?: (Option | string | number)[]
  placeholder?: string
  disabled?: boolean
  required?: boolean
  error?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  required: false,
  error: false
})

defineEmits<{
  'update:modelValue': [value: string | number]
  blur: []
  focus: []
}>()

const normalizedOptions = computed(() => {
  if (!props.options) {
    return []
  }

  return props.options.map(option => {
    if (typeof option === 'object' && option !== null) {
      return option as Option
    }

    // Convert primitive values to option format
    return {
      value: option,
      label: String(option)
    }
  })
})

const selectClasses = computed(() => {
  const baseClasses = 'block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm'
  const errorClasses = props.error ? 'border-red-300 text-red-900 focus:ring-red-500 focus:border-red-500' : ''
  const disabledClasses = props.disabled ? 'bg-gray-50 text-gray-500' : ''

  return [baseClasses, errorClasses, disabledClasses].join(' ')
})
</script>
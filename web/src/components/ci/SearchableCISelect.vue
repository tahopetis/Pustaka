<template>
  <div class="searchable-ci-select">
    <select
      :id="id"
      v-model="internalValue"
      :multiple="multiple"
      :disabled="disabled"
      class="form-select"
      @change="handleChange"
    >
      <option v-if="!multiple && !internalValue" value="">{{ placeholder }}</option>
      <option v-for="ci in ciList" :key="ci.id" :value="ci.id" :disabled="excludeIds.includes(ci.id)">
        {{ ci.name }} ({{ ci.ci_type }})
      </option>
    </select>
    <p v-if="helpText" class="text-sm text-gray-500 mt-1">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

interface Props {
  modelValue: string | string[] | null
  multiple?: boolean
  placeholder?: string
  disabled?: boolean
  excludeIds?: string[]
  maxResults?: number
  helpText?: string
  id?: string
}

interface Props {
  modelValue: string | string[] | null
  multiple?: boolean
  placeholder?: string
  disabled?: boolean
  excludeIds?: string[]
  maxResults?: number
  helpText?: string
  id?: string
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  placeholder: 'Select...',
  disabled: false,
  excludeIds: () => [],
  maxResults: 10,
  helpText: '',
  id: undefined
})

const emit = defineEmits<{
  'update:modelValue': [value: string | string[] | null]
  'change': [value: string | string[] | null]
}>()

const internalValue = ref<string | string[] | null>(props.modelValue)
const ciList = ref<Array<{id: string, name: string, ci_type: string}>>([])

watch(() => props.modelValue, (newValue) => {
  internalValue.value = newValue
})

const handleChange = () => {
  emit('update:modelValue', internalValue.value)
  emit('change', internalValue.value)
}

onMounted(() => {
  // TODO: Load CI list from API
  // For now, just show a placeholder component
  console.warn('SearchableCISelect: API integration not implemented yet')
})
</script>

<style scoped>
.searchable-ci-select {
  width: 100%;
}

.form-select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
}

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-select:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}
</style>

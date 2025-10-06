<template>
  <div>
    <label class="form-label">{{ label }}</label>
    <select
      v-model="selectedValue"
      :required="required"
      :disabled="disabled"
      class="form-input"
      @change="handleChange"
    >
      <option value="">{{ placeholder }}</option>
      <option
        v-for="type in relationshipTypes"
        :key="type.id"
        :value="type.name"
        :disabled="!type.is_active"
      >
        {{ type.forward_label }}
        <template v-if="!type.is_active"> (Inactive)</template>
      </option>
    </select>
    <p v-if="helpText" class="form-help">{{ helpText }}</p>

    <!-- Relationship type preview -->
    <div v-if="selectedType && showPreview" class="mt-2 p-3 bg-gray-50 rounded-lg">
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600">Forward:</span>
        <span
          v-if="selectedType.color"
          class="badge text-sm"
          :style="{ backgroundColor: selectedType.color + '20', color: selectedType.color, borderColor: selectedType.color }"
        >
          {{ selectedType.forward_label }}
        </span>
        <span v-else class="badge badge-info text-sm">
          {{ selectedType.forward_label }}
        </span>
      </div>
      <div class="flex items-center justify-between text-sm mt-1">
        <span class="text-gray-600">Reverse:</span>
        <span
          v-if="selectedType.color"
          class="badge text-sm"
          :style="{ backgroundColor: selectedType.color + '20', color: selectedType.color, borderColor: selectedType.color }"
        >
          {{ selectedType.reverse_label }}
        </span>
        <span v-else class="badge badge-secondary text-sm">
          {{ selectedType.reverse_label }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRelationshipTypeStore } from '@/stores/relationshipTypes'

interface RelationshipTypeSelectProps {
  modelValue?: string
  label?: string
  placeholder?: string
  required?: boolean
  disabled?: boolean
  helpText?: string
  showPreview?: boolean
  filterActive?: boolean
}

const props = withDefaults(defineProps<RelationshipTypeSelectProps>(), {
  modelValue: '',
  label: 'Relationship Type',
  placeholder: 'Select relationship type...',
  required: false,
  disabled: false,
  helpText: '',
  showPreview: true,
  filterActive: true
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'change': [type: any]
}>()

const relationshipTypeStore = useRelationshipTypeStore()

const selectedValue = ref(props.modelValue)

// Get relationship types based on filter
const relationshipTypes = computed(() => {
  return props.filterActive
    ? relationshipTypeStore.activeRelationshipTypes
    : relationshipTypeStore.relationshipTypes
})

// Get selected relationship type for preview
const selectedType = computed(() => {
  return relationshipTypes.value.find(type => type.name === selectedValue.value)
})

const handleChange = () => {
  emit('update:modelValue', selectedValue.value)
  emit('change', selectedType.value)
}

// Watch for external changes
watch(() => props.modelValue, (newValue) => {
  if (newValue !== selectedValue.value) {
    selectedValue.value = newValue
  }
})

// Load relationship types on mount
onMounted(async () => {
  if (relationshipTypes.value.length === 0) {
    try {
      await relationshipTypeStore.loadRelationshipTypes()
    } catch (error) {
      console.error('Failed to load relationship types:', error)
    }
  }
})
</script>
<template>
  <div class="relative">
    <input
      :id="id"
      type="text"
      :value="displayValue"
      :placeholder="placeholder"
      :disabled="disabled"
      class="form-input"
      @input="onInput"
      @focus="handleFocus"
      @blur="handleBlur"
    />

    <!-- Autocomplete dropdown -->
    <div
      v-if="showDropdown && (searchResults.length > 0 || searching)"
      class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto"
    >
      <div v-if="searching" class="p-3 text-gray-500 text-sm">
        Searching...
      </div>
      <div
        v-for="result in searchResults"
        :key="result.name"
        class="px-4 py-2 hover:bg-gray-100 cursor-pointer"
        @mousedown.prevent="selectResult(result)"
      >
        <div class="font-medium">{{ result.name }}</div>
        <div v-if="result.description" class="text-sm text-gray-500">{{ result.description }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ciTypeAPI } from '@/services/api'

interface CITypeResult {
  name: string
  description?: string
}

interface Props {
  id?: string
  modelValue?: string
  domain?: 'asset' | 'ea' | 'all'
  eaDomain?: string
  disabled?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: 'Search CI Types...',
  disabled: false,
  domain: 'all'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'selected': [value: CITypeResult]
}>()

const searchQuery = ref('')
const searchResults = ref<CITypeResult[]>([])
const showDropdown = ref(false)
const searching = ref(false)
const searchTimeout = ref<number | null>(null)
const blurTimeout = ref<number | null>(null)

// Display value - show the modelValue if set, otherwise show search query
const displayValue = computed(() => {
  return props.modelValue || searchQuery.value
})

// Filter results based on domain
const filterByDomain = (results: CITypeResult[]): CITypeResult[] => {
  if (props.domain === 'all') {
    return results
  }

  if (props.domain === 'asset') {
    // Filter OUT types starting with 'EA.'
    return results.filter(r => !r.name.startsWith('EA.'))
  }

  if (props.domain === 'ea' && props.eaDomain) {
    // Only show types starting with 'EA.{eaDomain}' (capitalize first letter)
    const capitalizedDomain = props.eaDomain.charAt(0).toUpperCase() + props.eaDomain.slice(1).toLowerCase()
    const prefix = `EA.${capitalizedDomain}`
    return results.filter(r => r.name.startsWith(prefix))
  }

  // If domain is 'ea' but no eaDomain specified, show all EA.* types
  if (props.domain === 'ea') {
    return results.filter(r => r.name.startsWith('EA.'))
  }

  return results
}

// Fetch CI types from API with debounced search
const fetchCITypes = async () => {
  searching.value = true
  try {
    const response = await ciTypeAPI.list({
      search: searchQuery.value || undefined,
      limit: 5
    })

    let results: CITypeResult[] = []

    // Handle different response structures
    const data = response.data as any
    if (data.ci_types) {
      results = data.ci_types
    } else if (Array.isArray(data)) {
      results = data
    } else if (data.data && Array.isArray(data.data)) {
      results = data.data
    }

    // Apply domain filtering
    searchResults.value = filterByDomain(results)
  } catch (error) {
    console.error('Failed to fetch CI types:', error)
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

// Handle input changes with debouncing
const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  searchQuery.value = target.value

  // Clear existing timeout
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }

  // Show dropdown when typing
  showDropdown.value = true

  // Debounce the search
  searchTimeout.value = window.setTimeout(() => {
    fetchCITypes()
  }, 300)
}

// Handle focus - load initial suggestions
const handleFocus = () => {
  showDropdown.value = true

  // Load initial suggestions if search is empty
  if (!searchQuery.value && !props.modelValue) {
    fetchCITypes()
  }
}

// Handle blur - delay hiding to allow click events to register
const handleBlur = () => {
  blurTimeout.value = window.setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

// Select a result from the dropdown
const selectResult = (result: CITypeResult) => {
  searchQuery.value = result.name
  emit('update:modelValue', result.name)
  emit('selected', result)
  showDropdown.value = false
}

// Sync with external modelValue changes
watch(() => props.modelValue, (newValue) => {
  if (newValue && newValue !== searchQuery.value) {
    searchQuery.value = newValue
  }
}, { immediate: true })

// Cleanup timeouts on unmount
// Note: Vue 3's setup cleanup handles this automatically
</script>

<style scoped>
.form-input {
  @apply block w-full border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2;
}

.form-input:disabled {
  @apply bg-gray-50 text-gray-500 cursor-not-allowed;
}
</style>

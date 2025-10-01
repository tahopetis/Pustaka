<template>
  <div class="bg-white shadow rounded-lg">
    <div class="card-header">
      <div class="flex items-center justify-between">
        <h3 class="text-lg leading-6 font-medium text-gray-900">Advanced Search</h3>
        <button
          @click="toggleAdvanced"
          class="text-blue-600 hover:text-blue-800 text-sm"
        >
          {{ showAdvanced ? 'Hide Advanced' : 'Show Advanced' }}
          <svg
            class="w-4 h-4 inline ml-1 transition-transform"
            :class="{ 'rotate-180': showAdvanced }"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
          </svg>
        </button>
      </div>
    </div>
    <div class="card-body">
      <!-- Basic Search -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
        <div>
          <label class="form-label">Quick Search</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search by name, tags..."
            class="form-input"
            @input="onSearchChange"
          >
        </div>
        <div>
          <label class="form-label">CI Type</label>
          <select v-model="filters.ci_type" class="form-input" @change="onSearchChange">
            <option value="">All Types</option>
            <option v-for="type in ciTypes" :key="type.id" :value="type.name">
              {{ type.name }}
            </option>
          </select>
        </div>
        <div>
          <label class="form-label">Sort By</label>
          <select v-model="filters.sort" class="form-input" @change="onSearchChange">
            <option value="name">Name</option>
            <option value="type">Type</option>
            <option value="created_at">Created Date</option>
            <option value="updated_at">Updated Date</option>
          </select>
        </div>
        <div>
          <label class="form-label">Order</label>
          <select v-model="filters.order" class="form-input" @change="onSearchChange">
            <option value="asc">Ascending</option>
            <option value="desc">Descending</option>
          </select>
        </div>
      </div>

      <!-- Advanced Search -->
      <div v-if="showAdvanced" class="border-t pt-4">
        <h4 class="text-md font-medium text-gray-900 mb-4">Attribute Search</h4>

        <!-- CI Type Selection for Attribute Search -->
        <div class="mb-4">
          <label class="form-label">Search in Attributes for Type</label>
          <select
            v-model="selectedCIType"
            class="form-input"
            @change="onCITypeChange"
          >
            <option value="">Select CI type to search attributes</option>
            <option v-for="type in ciTypes" :key="type.id" :value="type.name">
              {{ type.name }} ({{ type.required_attributes.length + type.optional_attributes.length }} attributes)
            </option>
          </select>
        </div>

        <!-- Attribute Filters -->
        <div v-if="selectedCITypeObject" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div v-for="attr in availableAttributes" :key="attr.name" class="border border-gray-200 rounded-lg p-4">
              <div class="flex items-center justify-between mb-2">
                <label class="block text-sm font-medium text-gray-700">
                  {{ attr.name }}
                  <span class="text-gray-400 text-xs ml-1">({{ attr.type }})</span>
                </label>
                <button
                  @click="clearAttributeFilter(attr.name)"
                  class="text-gray-400 hover:text-gray-600"
                  title="Clear filter"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </button>
              </div>

              <!-- String search -->
              <input
                v-if="attr.type === 'string' && !hasEnum(attr)"
                v-model="attributeFilters[attr.name]"
                type="text"
                :placeholder="`Search ${attr.name}...`"
                class="form-input"
                @input="onAttributeSearchChange"
              >

              <!-- Enum select -->
              <select
                v-else-if="attr.type === 'string' && hasEnum(attr)"
                v-model="attributeFilters[attr.name]"
                class="form-input"
                @change="onAttributeSearchChange"
              >
                <option value="">Any value</option>
                <option v-for="option in attr.validation?.enum" :key="option" :value="option">
                  {{ option }}
                </option>
              </select>

              <!-- Integer range -->
              <div v-else-if="attr.type === 'integer'" class="flex space-x-2">
                <input
                  v-model="attributeFilters[attr.name + '_min']"
                  type="number"
                  :placeholder="`Min ${attr.name}`"
                  class="form-input flex-1"
                  @input="onAttributeSearchChange"
                >
                <input
                  v-model="attributeFilters[attr.name + '_max']"
                  type="number"
                  :placeholder="`Max ${attr.name}`"
                  class="form-input flex-1"
                  @input="onAttributeSearchChange"
                >
              </div>

              <!-- Boolean select -->
              <select
                v-else-if="attr.type === 'boolean'"
                v-model="attributeFilters[attr.name]"
                class="form-input"
                @change="onAttributeSearchChange"
              >
                <option value="">Any value</option>
                <option value="true">True</option>
                <option value="false">False</option>
              </select>

              <!-- Array/Object contains (JSON search) -->
              <input
                v-else-if="attr.type === 'array' || attr.type === 'object'"
                v-model="attributeFilters[attr.name]"
                type="text"
                :placeholder="`Search within ${attr.name}...`"
                class="form-input"
                @input="onAttributeSearchChange"
              >

              <!-- Validation hint -->
              <div v-if="attr.validation && attr.validation.description" class="text-xs text-gray-500 mt-1">
                {{ getValidationHint(attr) }}
              </div>
            </div>
          </div>

          <!-- Active Filters Summary -->
          <div v-if="hasActiveAttributeFilters" class="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
            <h5 class="text-sm font-medium text-blue-900 mb-2">Active Attribute Filters:</h5>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="(value, key) in activeAttributeFilters"
                :key="key"
                class="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded"
              >
                {{ formatFilterDisplay(key, value) }}
                <button
                  @click="clearAttributeFilter(key)"
                  class="ml-1 text-blue-600 hover:text-blue-800"
                >
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </button>
              </span>
            </div>
          </div>
        </div>

        <!-- Tags Search -->
        <div class="mt-4">
          <label class="form-label">Tags</label>
          <div class="flex space-x-2">
            <input
              v-model="tagInput"
              type="text"
              placeholder="Add tag filter..."
              class="form-input flex-1"
              @keyup.enter="addTagFilter"
            >
            <button
              @click="addTagFilter"
              class="btn btn-outline"
            >
              Add
            </button>
          </div>
          <div v-if="filters.tags.length > 0" class="flex flex-wrap gap-2 mt-2">
            <span
              v-for="tag in filters.tags"
              :key="tag"
              class="inline-flex items-center px-2 py-1 bg-green-100 text-green-800 text-xs rounded"
            >
              {{ tag }}
              <button
                @click="removeTagFilter(tag)"
                class="ml-1 text-green-600 hover:text-green-800"
              >
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
              </button>
            </span>
          </div>
        </div>
      </div>

      <!-- Search Actions -->
      <div class="flex justify-between items-center mt-4 pt-4 border-t">
        <div class="text-sm text-gray-500">
          <span v-if="isLoading">Searching...</span>
          <span v-else-if="totalResults !== undefined">
            Found {{ totalResults }} results
          </span>
        </div>
        <div class="space-x-2">
          <button
            @click="clearAllFilters"
            class="btn btn-outline"
          >
            Clear All
          </button>
          <button
            @click="applyFilters"
            class="btn btn-primary"
            :disabled="isLoading"
          >
            <span v-if="isLoading" class="spinner w-4 h-4 mr-2"></span>
            Search
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import type { CIType } from '@/types/ci'

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
  ciTypes: CIType[]
  isLoading?: boolean
  totalResults?: number
}

const props = withDefaults(defineProps<Props>(), {
  isLoading: false,
  totalResults: undefined
})

const emit = defineEmits<{
  'search': [filters: any]
  'clear': []
}>()

const showAdvanced = ref(false)
const selectedCIType = ref('')
const tagInput = ref('')

const filters = reactive({
  search: '',
  ci_type: '',
  sort: 'name',
  order: 'asc',
  tags: [] as string[],
  attributes: {} as Record<string, any>
})

const attributeFilters = reactive<Record<string, any>>({})

const selectedCITypeObject = computed(() => {
  return props.ciTypes.find(type => type.name === selectedCIType.value)
})

const availableAttributes = computed(() => {
  if (!selectedCITypeObject.value) return []
  return [
    ...selectedCITypeObject.value.required_attributes,
    ...selectedCITypeObject.value.optional_attributes
  ]
})

const activeAttributeFilters = computed(() => {
  const active: Record<string, any> = {}

  Object.keys(attributeFilters).forEach(key => {
    const value = attributeFilters[key]
    if (value !== undefined && value !== null && value !== '') {
      // Skip range filters that only have one value
      if (key.endsWith('_min') || key.endsWith('_max')) {
        const baseKey = key.replace(/_min|_max$/, '')
        if (!active[baseKey]) active[baseKey] = {}
        if (key.endsWith('_min')) active[baseKey].min = value
        if (key.endsWith('_max')) active[baseKey].max = value
      } else {
        active[key] = value
      }
    }
  })

  return active
})

const hasActiveAttributeFilters = computed(() => {
  return Object.keys(activeAttributeFilters.value).length > 0
})

const toggleAdvanced = () => {
  showAdvanced.value = !showAdvanced.value
}

const onSearchChange = () => {
  emit('search', { ...filters })
}

const onCITypeChange = () => {
  // Clear attribute filters when CI type changes
  Object.keys(attributeFilters).forEach(key => {
    delete attributeFilters[key]
  })
  onSearchChange()
}

const onAttributeSearchChange = () => {
  // Update filters object with attribute search values
  filters.attributes = { ...activeAttributeFilters.value }
  emit('search', { ...filters })
}

const hasEnum = (attr: AttributeDefinition) => {
  return attr.validation?.enum && attr.validation.enum.length > 0
}

const getValidationHint = (attr: AttributeDefinition) => {
  const validation = attr.validation
  if (!validation) return ''

  const hints = []
  if (validation.min_length !== undefined) hints.push(`min ${validation.min_length} chars`)
  if (validation.max_length !== undefined) hints.push(`max ${validation.max_length} chars`)
  if (validation.min !== undefined) hints.push(`min ${validation.min}`)
  if (validation.max !== undefined) hints.push(`max ${validation.max}`)
  if (validation.enum) hints.push(`${validation.enum.length} possible values`)
  if (validation.format) hints.push(`format: ${validation.format}`)
  if (validation.pattern) hints.push(`pattern required`)

  return hints.length > 0 ? `Constraints: ${hints.join(', ')}` : ''
}

const clearAttributeFilter = (attributeName: string) => {
  // Remove the attribute filter
  delete attributeFilters[attributeName]
  delete attributeFilters[attributeName + '_min']
  delete attributeFilters[attributeName + '_max']

  // Update filters and emit
  filters.attributes = { ...activeAttributeFilters.value }
  emit('search', { ...filters })
}

const formatFilterDisplay = (key: string, value: any) => {
  if (typeof value === 'object' && value.min !== undefined && value.max !== undefined) {
    return `${key}: ${value.min} - ${value.max}`
  } else if (typeof value === 'object' && value.min !== undefined) {
    return `${key}: ≥ ${value.min}`
  } else if (typeof value === 'object' && value.max !== undefined) {
    return `${key}: ≤ ${value.max}`
  } else {
    return `${key}: ${value}`
  }
}

const addTagFilter = () => {
  const tag = tagInput.value.trim()
  if (tag && !filters.tags.includes(tag)) {
    filters.tags.push(tag)
    tagInput.value = ''
    onSearchChange()
  }
}

const removeTagFilter = (tag: string) => {
  const index = filters.tags.indexOf(tag)
  if (index > -1) {
    filters.tags.splice(index, 1)
    onSearchChange()
  }
}

const clearAllFilters = () => {
  filters.search = ''
  filters.ci_type = ''
  filters.sort = 'name'
  filters.order = 'asc'
  filters.tags = []
  filters.attributes = {}

  Object.keys(attributeFilters).forEach(key => {
    delete attributeFilters[key]
  })

  selectedCIType.value = ''
  showAdvanced.value = false

  emit('clear')
}

const applyFilters = () => {
  emit('search', { ...filters })
}

// Watch for external filter changes
watch(() => props.ciTypes, () => {
  // Reset selected CI type if it no longer exists
  if (selectedCIType.value && !props.ciTypes.find(type => type.name === selectedCIType.value)) {
    selectedCIType.value = ''
  }
})
</script>
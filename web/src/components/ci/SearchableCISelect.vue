<template>
  <div class="searchable-ci-select">
    <!-- Filter box appears once the list is longer than `maxResults`, so short
         lists stay a plain dropdown and long ones stay browsable. -->
    <input
      v-if="showFilter"
      v-model="filter"
      type="search"
      class="form-input"
      :placeholder="`Filter ${ciList.length} configuration items...`"
      :disabled="disabled"
      :aria-label="placeholder"
    />

    <select
      :id="id"
      v-model="internalValue"
      :multiple="multiple"
      :disabled="disabled || loading"
      class="form-select"
      :size="multiple && visibleCIs.length > 8 ? 8 : undefined"
      @change="handleChange"
    >
      <option v-if="!multiple && !internalValue" value="">{{ placeholder }}</option>
      <optgroup v-for="group in groupedCIs" :key="group.ciType" :label="group.ciType">
        <option
          v-for="ci in group.items"
          :key="ci.id"
          :value="ci.id"
          :disabled="excludeIds.includes(ci.id)"
        >
          {{ ci.name }} ({{ ci.ci_type }})
        </option>
      </optgroup>
    </select>

    <p v-if="loading" class="text-sm text-gray-500 mt-1">Loading configuration items...</p>
    <p v-else-if="errorMessage" class="text-sm text-red-600 mt-1">{{ errorMessage }}</p>
    <p v-else-if="filter && visibleCIs.length === 0" class="text-sm text-gray-500 mt-1">
      No configuration items match "{{ filter }}".
    </p>
    <p v-if="helpText" class="text-sm text-gray-500 mt-1">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ciAPI } from '@/services/api'

interface Props {
  modelValue: string | string[] | null
  multiple?: boolean
  placeholder?: string
  disabled?: boolean
  excludeIds?: string[]
  /** Above this many configuration items a filter box is shown. */
  maxResults?: number
  helpText?: string
  id?: string
}

interface CIOption {
  id: string
  name: string
  ci_type: string
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
  // Mirrors the repo's other selects (e.g. RelationshipTypeSelect): `v-model`
  // carries the id(s), while `change` carries the selected CI object(s).
  'change': [value: CIOption | CIOption[] | null]
}>()

// The API caps `limit` at 100 per request; walk pages until everything is loaded.
const PAGE_SIZE = 100
const MAX_PAGES = 20

const internalValue = ref<string | string[] | null>(props.modelValue)
const ciList = ref<CIOption[]>([])
const loading = ref(false)
const errorMessage = ref('')
const filter = ref('')

watch(
  () => props.modelValue,
  (newValue) => {
    internalValue.value = newValue
  }
)

const showFilter = computed(() => ciList.value.length > props.maxResults)

const visibleCIs = computed(() => {
  const query = filter.value.trim().toLowerCase()
  if (!query) return ciList.value
  return ciList.value.filter(
    (ci) => ci.name.toLowerCase().includes(query) || ci.ci_type.toLowerCase().includes(query)
  )
})

// Group by CI type so long lists stay scannable in a native select.
const groupedCIs = computed(() => {
  const groups = new Map<string, CIOption[]>()
  for (const ci of visibleCIs.value) {
    const key = ci.ci_type || 'Uncategorized'
    const bucket = groups.get(key)
    if (bucket) bucket.push(ci)
    else groups.set(key, [ci])
  }
  return Array.from(groups.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([ciType, items]) => ({ ciType, items }))
})

const handleChange = () => {
  emit('update:modelValue', internalValue.value)

  if (props.multiple) {
    const ids = Array.isArray(internalValue.value) ? internalValue.value : []
    emit('change', ciList.value.filter((ci) => ids.includes(ci.id)))
  } else {
    const id = typeof internalValue.value === 'string' ? internalValue.value : ''
    emit('change', ciList.value.find((ci) => ci.id === id) || null)
  }
}

const loadCIs = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const collected: CIOption[] = []
    let page = 1
    let totalPages = 1

    do {
      const { data } = await ciAPI.list({ page, limit: PAGE_SIZE })
      const items: Array<Record<string, unknown>> = Array.isArray(data?.cis) ? data.cis : []
      for (const ci of items) {
        if (typeof ci.id !== 'string' || typeof ci.name !== 'string') continue
        collected.push({
          id: ci.id,
          name: ci.name,
          ci_type: typeof ci.ci_type === 'string' && ci.ci_type ? ci.ci_type : 'Unknown'
        })
      }
      totalPages = Number(data?.total_pages) || 1
      page += 1
    } while (page <= totalPages && page <= MAX_PAGES)

    ciList.value = collected
  } catch (error) {
    console.error('SearchableCISelect: failed to load configuration items', error)
    errorMessage.value = 'Failed to load configuration items. Please reload the page.'
  } finally {
    loading.value = false
  }
}

onMounted(loadCIs)

defineExpose({ refresh: loadCIs })
</script>

<style scoped>
.searchable-ci-select {
  width: 100%;
}

.form-select,
.form-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
}

.form-input {
  margin-bottom: 0.5rem;
}

.form-select:focus,
.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-select:disabled,
.form-input:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}
</style>

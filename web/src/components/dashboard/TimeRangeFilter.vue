<template>
  <div class="bg-white shadow rounded-lg p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Time Range Filter</h3>

    <!-- Quick Select Buttons -->
    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700 mb-2">Quick Select</label>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="preset in presets"
          :key="preset.value"
          @click="selectPreset(preset.value)"
          :class="[
            'px-4 py-2 text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500',
            selectedPreset === preset.value
              ? 'bg-blue-600 text-white hover:bg-blue-700'
              : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
          ]"
        >
          {{ preset.label }}
        </button>
      </div>
    </div>

    <!-- Custom Date Range -->
    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700 mb-2">Custom Date Range</label>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs text-gray-600 mb-1">Start Date</label>
          <input
            v-model="localStartDate"
            type="date"
            :max="maxStartDate"
            class="form-input w-full"
            @change="onCustomDateChange"
          />
        </div>
        <div>
          <label class="block text-xs text-gray-600 mb-1">End Date</label>
          <input
            v-model="localEndDate"
            type="date"
            :min="localStartDate || undefined"
            :max="today"
            class="form-input w-full"
            @change="onCustomDateChange"
          />
        </div>
      </div>
      <p v-if="validationError" class="mt-2 text-sm text-red-600">
        {{ validationError }}
      </p>
    </div>

    <!-- Action Buttons -->
    <div class="flex flex-wrap gap-3">
      <button
        @click="applyFilter"
        :disabled="!!validationError || !hasChanges"
        class="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-blue-300 disabled:cursor-not-allowed transition-colors"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        Apply Filter
      </button>

      <button
        @click="resetFilter"
        class="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md border border-gray-300 text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Reset
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// Types
export interface TimeRange {
  startDate: string | null  // ISO format YYYY-MM-DD
  endDate: string | null
  preset: '7' | '30' | '90' | 'custom'
}

type PresetValue = '7' | '30' | '90'

interface Preset {
  value: PresetValue
  label: string
  days: number
}

// Emits
const emit = defineEmits<{
  change: [timeRange: TimeRange]
}>()

// Preset configurations
const presets: Preset[] = [
  { value: '7', label: 'Last 7 Days', days: 7 },
  { value: '30', label: 'Last 30 Days', days: 30 },
  { value: '90', label: 'Last 90 Days', days: 90 }
]

// Reactive state
const selectedPreset = ref<PresetValue | 'custom'>('30')
const localStartDate = ref<string>('')
const localEndDate = ref<string>('')
const appliedTimeRange = ref<TimeRange | null>(null)

// Computed properties
const today = computed(() => {
  return new Date().toISOString().split('T')[0]
})

const maxStartDate = computed(() => {
  return localEndDate.value || today.value
})

const validationError = computed(() => {
  if (!localStartDate.value || !localEndDate.value) {
    return null
  }

  const start = new Date(localStartDate.value)
  const end = new Date(localEndDate.value)

  if (start > end) {
    return 'Start date must be before or equal to end date'
  }

  const now = new Date()
  now.setHours(0, 0, 0, 0)

  if (end > now) {
    return 'End date cannot be in the future'
  }

  return null
})

const hasChanges = computed(() => {
  if (!appliedTimeRange.value) {
    return true // Initial state, allow apply
  }

  return (
    localStartDate.value !== appliedTimeRange.value.startDate ||
    localEndDate.value !== appliedTimeRange.value.endDate ||
    selectedPreset.value !== appliedTimeRange.value.preset
  )
})

// Methods
const calculateDateRange = (days: number): { start: string; end: string } => {
  const end = new Date()
  const start = new Date()
  start.setDate(end.getDate() - days)

  return {
    start: start.toISOString().split('T')[0],
    end: end.toISOString().split('T')[0]
  }
}

const selectPreset = (preset: PresetValue) => {
  selectedPreset.value = preset
  const presetConfig = presets.find(p => p.value === preset)

  if (presetConfig) {
    const { start, end } = calculateDateRange(presetConfig.days)
    localStartDate.value = start
    localEndDate.value = end
  }
}

const onCustomDateChange = () => {
  // When user manually changes dates, switch to custom preset
  if (localStartDate.value && localEndDate.value) {
    selectedPreset.value = 'custom'
  }
}

const applyFilter = () => {
  if (validationError.value) {
    return
  }

  const timeRange: TimeRange = {
    startDate: localStartDate.value || null,
    endDate: localEndDate.value || null,
    preset: selectedPreset.value
  }

  appliedTimeRange.value = { ...timeRange }
  emit('change', timeRange)
}

const resetFilter = () => {
  // Reset to default 30 days preset
  selectedPreset.value = '30'
  const { start, end } = calculateDateRange(30)
  localStartDate.value = start
  localEndDate.value = end

  const timeRange: TimeRange = {
    startDate: start,
    endDate: end,
    preset: '30'
  }

  appliedTimeRange.value = { ...timeRange }
  emit('change', timeRange)
}

// Lifecycle
onMounted(() => {
  // Initialize with 30 days preset
  selectPreset('30')

  // Automatically apply the default filter
  const timeRange: TimeRange = {
    startDate: localStartDate.value,
    endDate: localEndDate.value,
    preset: '30'
  }

  appliedTimeRange.value = { ...timeRange }
  emit('change', timeRange)
})
</script>

<style scoped>
/* Additional custom styles if needed */
.form-input {
  @apply block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm;
}
</style>

<template>
  <div
    class="quality-metric-card bg-white rounded-lg shadow p-6 transition-all duration-200"
    :class="{
      'cursor-pointer hover:shadow-lg': clickable,
      'cursor-default': !clickable,
    }"
    @click="handleClick"
  >
    <div class="flex items-start">
      <!-- Icon -->
      <div class="flex-shrink-0 mr-4">
        <div
          class="w-12 h-12 rounded-lg flex items-center justify-center"
          :class="iconContainerClass"
        >
          <svg class="w-6 h-6" :class="iconClass" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              :d="iconPath"
            />
          </svg>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-gray-500 uppercase tracking-wide">
          {{ title }}
        </p>
        <div class="mt-2 flex items-baseline gap-2">
          <p class="text-3xl font-bold" :class="valueClass">
            {{ displayValue }}
          </p>
          <!-- Trend indicator (optional) -->
          <span
            v-if="trend && trend !== 'neutral'"
            class="text-sm font-medium flex items-center gap-1"
            :class="trendClass"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                v-if="trend === 'up'"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 10l7-7m0 0l7 7m-7-7v18"
              />
              <path
                v-else
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 14l-7 7m0 0l-7-7m7 7V3"
              />
            </svg>
            {{ trendText }}
          </span>
        </div>
        <!-- Optional description -->
        <p v-if="description" class="mt-1 text-sm text-gray-400">
          {{ description }}
        </p>
      </div>

      <!-- Click indicator -->
      <div v-if="clickable" class="flex-shrink-0 ml-2">
        <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 5l7 7-7 7"
          />
        </svg>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title: string
  value: string | number
  icon?: string
  clickable?: boolean
  trend?: 'up' | 'down' | 'neutral' | null
  trendValue?: number
  description?: string
}

const props = withDefaults(defineProps<Props>(), {
  icon: 'cube',
  clickable: false,
  trend: null,
})

const emit = defineEmits<{
  click: []
}>()

// ============================================================================
// Icon Paths
// ============================================================================

const iconPaths: Record<string, string> = {
  cube: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
  'check-circle': 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
  clock: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
  'exclamation-triangle': 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
  users: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
  chart: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z',
}

// ============================================================================
// Computed Properties
// ============================================================================

const iconPath = computed(() => {
  return iconPaths[props.icon] || iconPaths.cube
})

const displayValue = computed(() => {
  // Format number with percentage if title suggests it
  if (typeof props.value === 'number') {
    if (props.title.toLowerCase().includes('%') || props.title.toLowerCase().includes('completeness')) {
      return `${props.value.toFixed(1)}%`
    }
    return props.value.toLocaleString()
  }
  return props.value
})

const valueClass = computed(() => {
  // Determine color based on value and title
  const val = typeof props.value === 'number' ? props.value : 0

  if (props.title.toLowerCase().includes('completeness')) {
    if (val >= 80) return 'text-green-600'
    if (val >= 50) return 'text-yellow-600'
    return 'text-red-600'
  }

  if (props.title.toLowerCase().includes('stale')) {
    if (val === 0) return 'text-green-600'
    if (val < 10) return 'text-yellow-600'
    return 'text-red-600'
  }

  if (props.title.toLowerCase().includes('error')) {
    if (val === 0) return 'text-green-600'
    if (val < 20) return 'text-yellow-600'
    return 'text-red-600'
  }

  // Default color for other metrics
  return 'text-gray-900'
})

const iconContainerClass = computed(() => {
  const val = typeof props.value === 'number' ? props.value : 0

  if (props.title.toLowerCase().includes('completeness')) {
    if (val >= 80) return 'bg-green-100'
    if (val >= 50) return 'bg-yellow-100'
    return 'bg-red-100'
  }

  if (props.title.toLowerCase().includes('stale') || props.title.toLowerCase().includes('error')) {
    if (val === 0) return 'bg-green-100'
    if (val < (props.title.toLowerCase().includes('stale') ? 10 : 20)) return 'bg-yellow-100'
    return 'bg-red-100'
  }

  return 'bg-blue-100'
})

const iconClass = computed(() => {
  const val = typeof props.value === 'number' ? props.value : 0

  if (props.title.toLowerCase().includes('completeness')) {
    if (val >= 80) return 'text-green-600'
    if (val >= 50) return 'text-yellow-600'
    return 'text-red-600'
  }

  if (props.title.toLowerCase().includes('stale') || props.title.toLowerCase().includes('error')) {
    if (val === 0) return 'text-green-600'
    if (val < (props.title.toLowerCase().includes('stale') ? 10 : 20)) return 'text-yellow-600'
    return 'text-red-600'
  }

  return 'text-blue-600'
})

const trendClass = computed(() => {
  if (props.trend === 'up') {
    // Up could be good (green) or bad (red) depending on context
    // For stale/errors, up is bad. For completeness, up is good.
    if (props.title.toLowerCase().includes('stale') || props.title.toLowerCase().includes('error')) {
      return 'text-red-600'
    }
    return 'text-green-600'
  }

  if (props.trend === 'down') {
    // Down could be good (green) or bad (red) depending on context
    if (props.title.toLowerCase().includes('stale') || props.title.toLowerCase().includes('error')) {
      return 'text-green-600'
    }
    return 'text-red-600'
  }

  return 'text-gray-500'
})

const trendText = computed(() => {
  if (props.trendValue !== undefined) {
    return `${props.trendValue}%`
  }
  return ''
})

// ============================================================================
// Methods
// ============================================================================

function handleClick() {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<style scoped>
.quality-metric-card {
  position: relative;
}

.quality-metric-card:hover {
  transform: translateY(-2px);
}
</style>

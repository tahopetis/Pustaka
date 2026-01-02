<template>
  <div
    class="border rounded-lg p-4 transition-colors"
    :class="cardClasses"
  >
    <div class="flex items-start">
      <!-- Icon -->
      <div class="flex-shrink-0 mr-3">
        <svg
          class="w-6 h-6"
          :class="iconColor"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            :d="iconClass"
          />
        </svg>
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium" :class="textColor">{{ title }}</p>
        <div class="mt-2 flex items-baseline">
          <p class="text-2xl font-semibold" :class="countColor">
            {{ count }}
          </p>
          <p class="ml-2 text-sm" :class="percentageColor">
            ({{ percentage.toFixed(1) }}%)
          </p>
        </div>
        <div class="mt-2 w-full bg-gray-200 rounded-full h-1.5">
          <div
            class="h-1.5 rounded-full transition-all duration-500"
            :class="progressBarColor"
            :style="{ width: `${percentage}%` }"
          />
        </div>
      </div>

      <!-- Severity Badge -->
      <div class="flex-shrink-0 ml-2">
        <span
          class="inline-flex items-center px-2 py-1 rounded text-xs font-medium"
          :class="badgeClasses"
        >
          {{ severity }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title: string
  count: number
  percentage: number
  severity: 'low' | 'medium' | 'high' | 'critical'
  iconClass: string
}

const props = withDefaults(defineProps<Props>(), {
  count: 0,
  percentage: 0
})

const cardClasses = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'bg-red-50 border-red-200'
  if (severity === 'high') return 'bg-orange-50 border-orange-200'
  if (severity === 'medium') return 'bg-yellow-50 border-yellow-200'
  return 'bg-blue-50 border-blue-200'
})

const iconColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-600'
  if (severity === 'high') return 'text-orange-600'
  if (severity === 'medium') return 'text-yellow-600'
  return 'text-blue-600'
})

const textColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-900'
  if (severity === 'high') return 'text-orange-900'
  if (severity === 'medium') return 'text-yellow-900'
  return 'text-blue-900'
})

const countColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-700'
  if (severity === 'high') return 'text-orange-700'
  if (severity === 'medium') return 'text-yellow-700'
  return 'text-blue-700'
})

const percentageColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-600'
  if (severity === 'high') return 'text-orange-600'
  if (severity === 'medium') return 'text-yellow-600'
  return 'text-blue-600'
})

const progressBarColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'bg-red-600'
  if (severity === 'high') return 'bg-orange-600'
  if (severity === 'medium') return 'bg-yellow-600'
  return 'bg-blue-600'
})

const badgeClasses = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'bg-red-200 text-red-800'
  if (severity === 'high') return 'bg-orange-200 text-orange-800'
  if (severity === 'medium') return 'bg-yellow-200 text-yellow-800'
  return 'bg-blue-200 text-blue-800'
})
</script>

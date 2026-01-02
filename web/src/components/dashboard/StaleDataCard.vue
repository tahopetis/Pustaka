<template>
  <div
    class="border rounded-lg p-3 text-center transition-colors"
    :class="cardClasses"
  >
    <!-- Clock Icon -->
    <svg
      class="w-5 h-5 mx-auto mb-2"
      :class="iconColor"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>

    <!-- Title -->
    <p class="text-xs font-medium mb-1" :class="textColor">
      {{ title }}
    </p>

    <!-- Count -->
    <p class="text-xl font-semibold" :class="countColor">
      {{ count }}
    </p>

    <!-- Label -->
    <p class="text-xs mt-1" :class="labelColor">
      CI{{ count !== 1 ? 's' : '' }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title: string
  count: number
  severity: 'low' | 'medium' | 'high' | 'critical'
}

const props = withDefaults(defineProps<Props>(), {
  count: 0
})

const cardClasses = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'bg-red-50 border-red-200'
  if (severity === 'high') return 'bg-orange-50 border-orange-200'
  if (severity === 'medium') return 'bg-yellow-50 border-yellow-200'
  return 'bg-gray-50 border-gray-200'
})

const iconColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-600'
  if (severity === 'high') return 'text-orange-600'
  if (severity === 'medium') return 'text-yellow-600'
  return 'text-gray-600'
})

const textColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-900'
  if (severity === 'high') return 'text-orange-900'
  if (severity === 'medium') return 'text-yellow-900'
  return 'text-gray-900'
})

const countColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-700'
  if (severity === 'high') return 'text-orange-700'
  if (severity === 'medium') return 'text-yellow-700'
  return 'text-gray-700'
})

const labelColor = computed(() => {
  const severity = props.severity
  if (severity === 'critical') return 'text-red-600'
  if (severity === 'high') return 'text-orange-600'
  if (severity === 'medium') return 'text-yellow-600'
  return 'text-gray-600'
})
</script>

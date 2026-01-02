<template>
  <div class="space-y-1">
    <div class="flex items-center justify-between text-sm">
      <div class="flex items-center space-x-2">
        <svg class="w-4 h-4" :class="textColor" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="iconClass" />
        </svg>
        <span :class="textColor">{{ label }}</span>
      </div>
      <div class="flex items-center space-x-3">
        <span :class="countColor" class="font-semibold">{{ count }}</span>
        <span :class="percentageColor" class="text-xs">({{ percentageText }}%)</span>
      </div>
    </div>
    <div class="w-full bg-gray-200 rounded-full h-2">
      <div
        class="h-2 rounded-full transition-all duration-500"
        :class="[color, 'progress-bar']"
        :style="{ width: `${percentage}%` }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  label: string
  count: number
  total: number
  color: string
  iconClass: string
}

const props = withDefaults(defineProps<Props>(), {
  count: 0,
  total: 0
})

const percentage = computed(() => {
  if (props.total === 0) return 0
  return (props.count / props.total) * 100
})

const percentageText = computed(() => {
  return percentage.value.toFixed(1)
})

const textColor = computed(() => {
  if (percentage.value === 0) return 'text-gray-400'
  if (props.color.includes('green')) return 'text-green-700'
  if (props.color.includes('yellow')) return 'text-yellow-700'
  if (props.color.includes('orange')) return 'text-orange-700'
  if (props.color.includes('red')) return 'text-red-700'
  return 'text-gray-700'
})

const countColor = computed(() => {
  if (percentage.value === 0) return 'text-gray-400'
  if (props.color.includes('green')) return 'text-green-900'
  if (props.color.includes('yellow')) return 'text-yellow-900'
  if (props.color.includes('orange')) return 'text-orange-900'
  if (props.color.includes('red')) return 'text-red-900'
  return 'text-gray-900'
})

const percentageColor = computed(() => {
  if (percentage.value === 0) return 'text-gray-400'
  if (props.color.includes('green')) return 'text-green-600'
  if (props.color.includes('yellow')) return 'text-yellow-600'
  if (props.color.includes('orange')) return 'text-orange-600'
  if (props.color.includes('red')) return 'text-red-600'
  return 'text-gray-600'
})
</script>

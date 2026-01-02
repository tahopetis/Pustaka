<template>
  <DashboardWidget
    title="CMDB Health Score"
    :loading="loading"
    :error="error"
    @retry="fetchHealthScore"
  >
    <div v-if="healthScore" class="space-y-6">
      <!-- Main Score Display -->
      <div class="flex flex-col items-center justify-center">
        <!-- Circular Progress -->
        <div class="relative w-48 h-48">
          <svg class="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
            <!-- Background Circle -->
            <circle
              cx="50"
              cy="50"
              r="45"
              fill="none"
              stroke="#e5e7eb"
              stroke-width="8"
            />

            <!-- Progress Circle -->
            <circle
              cx="50"
              cy="50"
              r="45"
              fill="none"
              :stroke="scoreColor"
              stroke-width="8"
              stroke-linecap="round"
              :stroke-dasharray="circumference"
              :stroke-dashoffset="strokeDashoffset"
              class="transition-all duration-1000 ease-in-out"
            />
          </svg>

          <!-- Score Text in Center -->
          <div class="absolute inset-0 flex flex-col items-center justify-center">
            <div class="text-4xl font-bold" :class="textColor">
              {{ Math.round(healthScore.overall) }}
            </div>
            <div class="text-sm text-gray-500 mt-1">out of 100</div>
          </div>
        </div>

        <!-- Trend Indicator -->
        <div class="flex items-center mt-4 space-x-2">
          <component
            :is="trendIcon"
            class="w-5 h-5"
            :class="trendColor"
          />
          <span class="text-sm font-medium" :class="trendColor">
            {{ trendText }}
          </span>
        </div>

        <!-- Last Updated -->
        <div class="text-xs text-gray-400 mt-2">
          Updated {{ formattedTime }}
        </div>
      </div>

      <!-- Sub-scores -->
      <div class="grid grid-cols-3 gap-4">
        <!-- Completeness -->
        <div class="text-center">
          <div class="text-xs text-gray-500 mb-1">Completeness</div>
          <div class="text-lg font-semibold text-gray-900">
            {{ Math.round(healthScore.completeness) }}%
          </div>
          <div class="w-full bg-gray-200 rounded-full h-1.5 mt-1">
            <div
              class="bg-blue-500 h-1.5 rounded-full transition-all duration-500"
              :style="{ width: `${healthScore.completeness}%` }"
            />
          </div>
        </div>

        <!-- Correctness -->
        <div class="text-center">
          <div class="text-xs text-gray-500 mb-1">Correctness</div>
          <div class="text-lg font-semibold text-gray-900">
            {{ Math.round(healthScore.correctness) }}%
          </div>
          <div class="w-full bg-gray-200 rounded-full h-1.5 mt-1">
            <div
              class="bg-green-500 h-1.5 rounded-full transition-all duration-500"
              :style="{ width: `${healthScore.correctness}%` }"
            />
          </div>
        </div>

        <!-- Compliance -->
        <div class="text-center">
          <div class="text-xs text-gray-500 mb-1">Compliance</div>
          <div class="text-lg font-semibold text-gray-900">
            {{ Math.round(healthScore.compliance) }}%
          </div>
          <div class="w-full bg-gray-200 rounded-full h-1.5 mt-1">
            <div
              class="bg-purple-500 h-1.5 rounded-full transition-all duration-500"
              :style="{ width: `${healthScore.compliance}%` }"
            />
          </div>
        </div>
      </div>
    </div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { formatDistanceToNow } from 'date-fns'
import { api } from '@/services/api'
import DashboardWidget from './DashboardWidget.vue'

interface HealthScore {
  overall: number
  completeness: number
  correctness: number
  compliance: number
  trend: 'improving' | 'declining' | 'stable'
  calculated_at: string
}

const loading = ref(true)
const error = ref<string | null>(null)
const healthScore = ref<HealthScore | null>(null)

const circumference = 2 * Math.PI * 45 // radius is 45

const strokeDashoffset = computed(() => {
  if (!healthScore.value) return circumference
  const progress = healthScore.value.overall / 100
  return circumference * (1 - progress)
})

const scoreColor = computed(() => {
  if (!healthScore.value) return '#9CA3AF'
  const score = healthScore.value.overall
  if (score >= 80) return '#10B981' // green
  if (score >= 50) return '#F59E0B' // yellow
  return '#EF4444' // red
})

const textColor = computed(() => {
  if (!healthScore.value) return 'text-gray-400'
  const score = healthScore.value.overall
  if (score >= 80) return 'text-green-600'
  if (score >= 50) return 'text-yellow-600'
  return 'text-red-600'
})

const trendIcon = computed(() => {
  if (!healthScore.value) return 'div'
  const trend = healthScore.value.trend
  if (trend === 'improving') {
    return 'svg' // Will be rendered as arrow up
  } else if (trend === 'declining') {
    return 'svg' // Will be rendered as arrow down
  }
  return 'svg' // Will be rendered as minus
})

const trendColor = computed(() => {
  if (!healthScore.value) return 'text-gray-400'
  const trend = healthScore.value.trend
  if (trend === 'improving') return 'text-green-600'
  if (trend === 'declining') return 'text-red-600'
  return 'text-gray-400'
})

const trendText = computed(() => {
  if (!healthScore.value) return 'Unknown'
  const trend = healthScore.value.trend
  if (trend === 'improving') return 'Improving'
  if (trend === 'declining') return 'Declining'
  return 'Stable'
})

const formattedTime = computed(() => {
  if (!healthScore.value) return ''
  return formatDistanceToNow(new Date(healthScore.value.calculated_at), {
    addSuffix: true
  })
})

const fetchHealthScore = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await api.get<HealthScore>('/dashboard/health-score')
    healthScore.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchHealthScore()
})
</script>

<style scoped>
/* Add any component-specific styles here */
</style>

<template>
  <div class="space-y-8 p-6">
    <h1 class="text-3xl font-bold text-gray-900">DonutChart Component Demo</h1>

    <!-- Demo 1: Basic CI Type Distribution -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">1. CI Type Distribution</h2>
      <DonutChart
        :data="ciTypeData"
        title="Configuration Items by Type"
        :size="350"
      />
    </div>

    <!-- Demo 2: Relationship Type Distribution -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">2. Relationship Type Distribution</h2>
      <DonutChart
        :data="relationshipData"
        title="Relationships by Type"
        :size="350"
        :innerRadiusRatio="0.5"
      />
    </div>

    <!-- Demo 3: Auto-colored Data -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">3. Auto-Generated Colors</h2>
      <DonutChart
        :data="autoColoredData"
        title="Status Distribution (Auto Colors)"
        :size="300"
      />
    </div>

    <!-- Demo 4: Thin Ring -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">4. Thin Ring Style</h2>
      <DonutChart
        :data="statusData"
        title="System Health Status"
        :size="280"
        :innerRadiusRatio="0.75"
      />
    </div>

    <!-- Demo 5: Many Segments -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">5. Many Segments</h2>
      <DonutChart
        :data="detailedData"
        title="Detailed Infrastructure Breakdown"
        :size="450"
      />
    </div>

    <!-- Demo 6: Empty State -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">6. Empty State</h2>
      <DonutChart
        :data="emptyData"
        title="No Data Available"
        :size="300"
      />
    </div>

    <!-- Demo 7: Single Segment -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">7. Single Segment (100%)</h2>
      <DonutChart
        :data="singleSegmentData"
        title="Uniform System"
        :size="250"
      />
    </div>

    <!-- Demo 8: Integration with DashboardWidget -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">8. With DashboardWidget</h2>
      <DashboardWidget
        title="Protected with Widget Wrapper"
        :loading="isLoading"
        :error="error"
        @retry="handleRetry"
      >
        <DonutChart
          :data="ciTypeData"
          :size="320"
        />
      </DashboardWidget>

      <div class="mt-4 flex gap-2">
        <button
          @click="simulateLoading"
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Simulate Loading
        </button>
        <button
          @click="simulateError"
          class="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
        >
          Simulate Error
        </button>
        <button
          @click="clearState"
          class="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
        >
          Clear State
        </button>
      </div>
    </div>

    <!-- Interactive Demo -->
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">9. Interactive Demo</h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Chart -->
        <div>
          <DonutChart
            :data="interactiveData"
            title="Interactive Chart"
            :size="interactiveSize"
            :innerRadiusRatio="interactiveRatio"
          />
        </div>

        <!-- Controls -->
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Size: {{ interactiveSize }}px
            </label>
            <input
              v-model.number="interactiveSize"
              type="range"
              min="200"
              max="500"
              step="10"
              class="w-full"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Inner Radius Ratio: {{ interactiveRatio.toFixed(2) }}
            </label>
            <input
              v-model.number="interactiveRatio"
              type="range"
              min="0"
              max="0.9"
              step="0.05"
              class="w-full"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Add Random Segment
            </label>
            <button
              @click="addRandomSegment"
              class="px-4 py-2 bg-indigo-500 text-white rounded hover:bg-indigo-600"
            >
              Add Segment
            </button>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Remove Last Segment
            </label>
            <button
              @click="removeSegment"
              :disabled="interactiveData.length === 0"
              class="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Remove Segment
            </button>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Reset
            </label>
            <button
              @click="resetInteractive"
              class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
            >
              Reset to Default
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DonutChart from './DonutChart.vue'
import DashboardWidget from './DashboardWidget.vue'
import type { DonutChartData } from '@/types/dashboard'

// ============================================================================
// Demo 1: CI Type Distribution
// ============================================================================
const ciTypeData: DonutChartData[] = [
  { label: 'Servers', value: 45, color: '#3B82F6' },
  { label: 'Databases', value: 30, color: '#10B981' },
  { label: 'Applications', value: 25, color: '#F59E0B' },
  { label: 'Network Devices', value: 15, color: '#EF4444' },
  { label: 'Storage', value: 10, color: '#8B5CF6' },
]

// ============================================================================
// Demo 2: Relationship Type Distribution
// ============================================================================
const relationshipData: DonutChartData[] = [
  { label: 'Depends On', value: 120, color: '#3B82F6' },
  { label: 'Contains', value: 85, color: '#10B981' },
  { label: 'Connected To', value: 60, color: '#F59E0B' },
  { label: 'Hosts', value: 45, color: '#EC4899' },
]

// ============================================================================
// Demo 3: Auto-colored Data (no color specified)
// ============================================================================
const autoColoredData: DonutChartData[] = [
  { label: 'Active', value: 75 },
  { label: 'Pending', value: 15 },
  { label: 'Inactive', value: 8 },
  { label: 'Error', value: 2 },
]

// ============================================================================
// Demo 4: Status Data (Thin Ring)
// ============================================================================
const statusData: DonutChartData[] = [
  { label: 'Healthy', value: 85, color: '#10B981' },
  { label: 'Warning', value: 10, color: '#F59E0B' },
  { label: 'Critical', value: 5, color: '#EF4444' },
]

// ============================================================================
// Demo 5: Detailed Data (Many Segments)
// ============================================================================
const detailedData: DonutChartData[] = [
  { label: 'Web Servers', value: 25, color: '#3B82F6' },
  { label: 'App Servers', value: 20, color: '#10B981' },
  { label: 'Database Servers', value: 15, color: '#F59E0B' },
  { label: 'Cache Servers', value: 10, color: '#EF4444' },
  { label: 'Load Balancers', value: 8, color: '#8B5CF6' },
  { label: 'Message Queues', value: 7, color: '#EC4899' },
  { label: 'Storage Systems', value: 6, color: '#14B8A6' },
  { label: 'Monitoring Tools', value: 5, color: '#F97316' },
  { label: 'Security Tools', value: 4, color: '#6366F1' },
]

// ============================================================================
// Demo 6: Empty State
// ============================================================================
const emptyData: DonutChartData[] = []

// ============================================================================
// Demo 7: Single Segment
// ============================================================================
const singleSegmentData: DonutChartData[] = [
  { label: 'Linux Systems', value: 100, color: '#3B82F6' },
]

// ============================================================================
// Demo 8: DashboardWidget Integration
// ============================================================================
const isLoading = ref(false)
const error = ref<string | null>(null)

function simulateLoading() {
  isLoading.value = true
  error.value = null
  setTimeout(() => {
    isLoading.value = false
  }, 2000)
}

function simulateError() {
  isLoading.value = false
  error.value = 'Failed to load chart data. Please try again.'
}

function clearState() {
  isLoading.value = false
  error.value = null
}

function handleRetry() {
  simulateLoading()
}

// ============================================================================
// Demo 9: Interactive Demo
// ============================================================================
const interactiveSize = ref(300)
const interactiveRatio = ref(0.6)
const interactiveData = ref<DonutChartData[]>([
  { label: 'Category A', value: 40, color: '#3B82F6' },
  { label: 'Category B', value: 30, color: '#10B981' },
  { label: 'Category C', value: 30, color: '#F59E0B' },
])

const colorPalette = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444',
  '#8B5CF6', '#EC4899', '#14B8A6', '#F97316',
  '#6366F1', '#84CC16', '#06B6D4', '#F43F5E',
]

function addRandomSegment() {
  const index = interactiveData.value.length
  const value = Math.floor(Math.random() * 50) + 10
  const color = colorPalette[index % colorPalette.length]

  interactiveData.value.push({
    label: `Segment ${String.fromCharCode(65 + index)}`,
    value,
    color,
  })
}

function removeSegment() {
  if (interactiveData.value.length > 0) {
    interactiveData.value.pop()
  }
}

function resetInteractive() {
  interactiveSize.value = 300
  interactiveRatio.value = 0.6
  interactiveData.value = [
    { label: 'Category A', value: 40, color: '#3B82F6' },
    { label: 'Category B', value: 30, color: '#10B981' },
    { label: 'Category C', value: 30, color: '#F59E0B' },
  ]
}
</script>

<style scoped>
/* Demo-specific styles */
</style>

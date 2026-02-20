<template>
  <div class="quality-chart bg-white rounded-lg shadow p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-gray-900">
        {{ title }}
      </h3>
      <div v-if="total" class="text-sm text-gray-500">
        Total: {{ total }}
      </div>
    </div>

    <!-- Chart Container -->
    <div class="chart-container">
      <!-- Empty State -->
      <div
        v-if="!hasData"
        class="flex flex-col items-center justify-center h-64 text-gray-400"
      >
        <svg class="w-16 h-16 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
        <p class="text-sm">No data available</p>
      </div>

      <!-- Pie Chart with Legend -->
      <div v-else class="flex items-start gap-6">
        <!-- SVG Chart -->
        <div class="flex-shrink-0" :style="{ width: chartSize + 'px', height: chartSize + 'px' }">
          <svg
            :viewBox="`0 0 ${chartSize} ${chartSize}`"
            :width="chartSize"
            :height="chartSize"
            class="overflow-visible"
            role="img"
            :aria-label="ariaLabel"
          >
            <g :transform="`translate(${chartSize / 2}, ${chartSize / 2})`">
              <!-- Pie Segments -->
              <g class="segments">
                <path
                  v-for="(segment, index) in segments"
                  :key="index"
                  :d="segment.path"
                  :fill="segment.color"
                  :class="[
                    'segment',
                    'transition-all duration-200 cursor-pointer',
                    hoveredIndex === index ? 'opacity-100' : 'opacity-90',
                  ]"
                  :style="{
                    transform: hoveredIndex === index ? 'scale(1.05)' : 'scale(1)',
                    transformOrigin: 'center',
                  }"
                  @mouseenter="handleMouseEnter(index, $event)"
                  @mouseleave="handleMouseLeave"
                  @focus="handleMouseEnter(index, $event)"
                  @blur="handleMouseLeave"
                  tabindex="0"
                  :aria-label="`${segment.label}: ${segment.value} (${segment.percentage}%)`"
                />
              </g>

              <!-- Center Label (for donut style, optional) -->
              <g v-if="showDonut" class="center-label" pointer-events="none">
                <text
                  text-anchor="middle"
                  dominant-baseline="middle"
                  class="fill-gray-400 text-xs font-medium"
                  y="-8"
                >
                  Total
                </text>
                <text
                  text-anchor="middle"
                  dominant-baseline="middle"
                  class="fill-gray-900 text-2xl font-bold"
                  y="12"
                >
                  {{ total }}
                </text>
              </g>
            </g>
          </svg>

          <!-- Tooltip -->
          <div
            v-if="tooltip.visible"
            ref="tooltipRef"
            class="absolute z-10 px-3 py-2 text-sm bg-gray-900 text-white rounded-lg shadow-lg pointer-events-none"
            :style="{
              left: tooltip.x + 'px',
              top: tooltip.y + 'px',
              transform: 'translate(-50%, -100%) translateY(-8px)',
            }"
          >
            <div class="font-semibold">{{ tooltip.label }}</div>
            <div class="text-gray-300">
              {{ tooltip.value }} ({{ tooltip.percentage }}%)
            </div>
          </div>
        </div>

        <!-- Legend -->
        <div class="flex-1 min-w-0">
          <div class="space-y-2">
            <div
              v-for="(item, index) in legendItems"
              :key="index"
              class="flex items-center justify-between gap-2 text-sm cursor-pointer hover:bg-gray-50 rounded px-2 py-1.5 transition-colors"
              @mouseenter="handleMouseEnter(index, $event)"
              @mouseleave="handleMouseLeave"
              @click="handleMouseEnter(index, $event)"
              role="button"
              tabindex="0"
              :aria-label="`${item.label}: ${item.value} items, ${item.percentage}%`"
            >
              <div class="flex items-center gap-2 min-w-0 flex-1">
                <div
                  class="w-3 h-3 rounded-sm flex-shrink-0"
                  :style="{ backgroundColor: item.color }"
                />
                <span class="text-gray-700 truncate" :title="item.label">
                  {{ item.label }}
                </span>
              </div>
              <div class="flex items-center gap-2 flex-shrink-0">
                <span class="font-medium text-gray-900">{{ item.value }}</span>
                <span class="text-gray-500 text-xs min-w-[3rem] text-right">
                  {{ item.percentage }}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import * as d3 from 'd3'

interface ChartData {
  label: string
  value: number
  color?: string
}

interface Props {
  title: string
  data: Record<string, number>
  type?: 'pie' | 'donut'
  size?: number
}

const props = withDefaults(defineProps<Props>(), {
  type: 'pie',
  size: 300,
})

// ============================================================================
// Template Refs
// ============================================================================

const tooltipRef = ref<HTMLDivElement | null>(null)

// ============================================================================
// State
// ============================================================================

const hoveredIndex = ref<number | null>(null)
const tooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  label: '',
  value: 0,
  percentage: '',
})

// ============================================================================
// Computed Properties
// ============================================================================

const chartSize = computed(() => props.size)

const showDonut = computed(() => props.type === 'donut')

const total = computed(() => {
  return Object.values(props.data).reduce((sum, val) => sum + val, 0)
})

const hasData = computed(() => {
  return total.value > 0
})

const chartData = computed((): ChartData[] => {
  return Object.entries(props.data).map(([label, value]) => ({
    label,
    value,
  }))
})

const colorScale = computed(() => {
  return d3.scaleOrdinal<string>()
    .range([
      '#3B82F6', // blue-500
      '#10B981', // emerald-500
      '#F59E0B', // amber-500
      '#EF4444', // red-500
      '#8B5CF6', // violet-500
      '#EC4899', // pink-500
      '#14B8A6', // teal-500
      '#F97316', // orange-500
      '#6366F1', // indigo-500
      '#84CC16', // lime-500
      '#06B6D4', // cyan-500
      '#A855F7', // purple-500
    ])
})

const segments = computed(() => {
  if (!hasData.value) return []

  const radius = chartSize.value / 2 - 10
  const innerRadius = props.type === 'donut' ? radius * 0.6 : 0

  const pie = d3.pie<ChartData>()
    .value((d) => d.value)
    .sort(null)

  const arc = d3.arc<d3.PieArcDatum<ChartData>>()
    .innerRadius(innerRadius)
    .outerRadius(radius)

  const pieData = pie(chartData.value)

  return pieData.map((d, index) => ({
    path: arc(d) || '',
    color: d.data.color || colorScale.value(index.toString()),
    label: d.data.label,
    value: d.data.value,
    percentage: ((d.data.value / total.value) * 100).toFixed(1),
  }))
})

const legendItems = computed(() => {
  return segments.value.map((segment) => ({
    label: segment.label,
    value: segment.value,
    percentage: segment.percentage,
    color: segment.color,
  }))
})

const ariaLabel = computed(() => {
  return `${props.title} showing ${chartData.value.length} categories with total of ${total.value}`
})

// ============================================================================
// Methods
// ============================================================================

function handleMouseEnter(index: number, event: MouseEvent) {
  hoveredIndex.value = index
  showTooltip(index, event)
}

function handleMouseLeave() {
  hoveredIndex.value = null
  hideTooltip()
}

function showTooltip(index: number, event: MouseEvent) {
  const segment = segments.value[index]
  if (!segment) return

  const container = (event.currentTarget as HTMLElement).closest('.quality-chart')
  if (!container) return

  const rect = container.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  tooltip.value = {
    visible: true,
    x,
    y,
    label: segment.label,
    value: segment.value,
    percentage: segment.percentage,
  }
}

function hideTooltip() {
  tooltip.value.visible = false
}

// ============================================================================
// Lifecycle
// ============================================================================

watch(
  () => props.data,
  async () => {
    if (hasData.value) {
      await nextTick()
    }
  },
  { deep: true }
)

onMounted(async () => {
  await nextTick()
})
</script>

<style scoped>
.quality-chart {
  position: relative;
}

.chart-container {
  position: relative;
  min-width: 250px;
  min-height: 200px;
}

.segment {
  outline: none;
}

.segment:hover,
.segment:focus {
  filter: brightness(1.1);
}

.segment:focus-visible {
  stroke: #3B82F6;
  stroke-width: 2px;
}

.segment {
  transition: transform 200ms ease-in-out, opacity 200ms ease-in-out;
}

text {
  user-select: none;
}

[role="button"]:focus-visible {
  outline: 2px solid #3B82F6;
  outline-offset: 2px;
}
</style>

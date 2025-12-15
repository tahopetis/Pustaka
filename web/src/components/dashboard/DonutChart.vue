<template>
  <div class="donut-chart-container" :style="{ width: `${size}px`, height: `${size}px` }">
    <!-- Chart Title -->
    <h3 v-if="title" class="text-sm font-medium text-gray-700 mb-2 text-center">
      {{ title }}
    </h3>

    <div class="flex items-start gap-4">
      <!-- SVG Chart -->
      <div class="flex-shrink-0" :style="{ width: chartSize + 'px', height: chartSize + 'px' }">
        <svg
          ref="svgRef"
          :viewBox="`0 0 ${chartSize} ${chartSize}`"
          :width="chartSize"
          :height="chartSize"
          class="overflow-visible"
          role="img"
          :aria-label="ariaLabel"
        >
          <!-- Chart Group -->
          <g :transform="`translate(${chartSize / 2}, ${chartSize / 2})`">
            <!-- Donut Segments -->
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
                role="button"
                :aria-label="`${segment.label}: ${segment.value} (${segment.percentage}%)`"
              />
            </g>

            <!-- Center Label -->
            <g class="center-label" pointer-events="none">
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
                {{ totalValue }}
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
        <div class="space-y-1.5">
          <div
            v-for="(item, index) in legendItems"
            :key="index"
            class="flex items-center justify-between gap-2 text-sm cursor-pointer hover:bg-gray-50 rounded px-2 py-1 transition-colors"
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

    <!-- Empty State -->
    <div
      v-if="!hasData"
      class="flex flex-col items-center justify-center h-full text-gray-400"
    >
      <svg
        class="w-16 h-16 mb-2"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"
        />
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z"
        />
      </svg>
      <p class="text-sm">No data available</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import * as d3 from 'd3'
import type { DonutChartData } from '@/types/dashboard'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** Array of segments to display in the donut chart */
  data: DonutChartData[]
  /** Optional title for the chart */
  title?: string
  /** Size of the entire component in pixels */
  size?: number
  /** Ratio of inner radius to outer radius (0-1, where 0.6 creates donut hole) */
  innerRadiusRatio?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: undefined,
  size: 300,
  innerRadiusRatio: 0.6,
})

// ============================================================================
// Template Refs
// ============================================================================

const svgRef = ref<SVGSVGElement | null>(null)
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

/**
 * Chart size (reserving space for title and legend)
 */
const chartSize = computed(() => {
  return props.title ? props.size - 30 : props.size
})

/**
 * Check if there's any data to display
 */
const hasData = computed(() => {
  return props.data && props.data.length > 0 && totalValue.value > 0
})

/**
 * Total value across all segments
 */
const totalValue = computed(() => {
  return props.data.reduce((sum, item) => sum + item.value, 0)
})

/**
 * Color palette for segments (fallback if colors not provided)
 */
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
    ])
})

/**
 * Processed segment data with paths and percentages
 */
const segments = computed(() => {
  if (!hasData.value) return []

  const radius = chartSize.value / 2 - 10 // Margin for hover effect
  const innerRadius = radius * props.innerRadiusRatio

  // Create pie layout
  const pie = d3.pie<DonutChartData>()
    .value((d) => d.value)
    .sort(null) // Preserve data order

  // Create arc generator
  const arc = d3.arc<d3.PieArcDatum<DonutChartData>>()
    .innerRadius(innerRadius)
    .outerRadius(radius)

  // Generate segments
  const pieData = pie(props.data)

  return pieData.map((d, index) => ({
    path: arc(d) || '',
    color: d.data.color || colorScale.value(index.toString()),
    label: d.data.label,
    value: d.data.value,
    percentage: ((d.data.value / totalValue.value) * 100).toFixed(1),
  }))
})

/**
 * Legend items with formatted data
 */
const legendItems = computed(() => {
  return segments.value.map((segment) => ({
    label: segment.label,
    value: segment.value,
    percentage: segment.percentage,
    color: segment.color,
  }))
})

/**
 * Accessible label for the chart
 */
const ariaLabel = computed(() => {
  const title = props.title || 'Donut chart'
  const description = `showing ${props.data.length} categories with total of ${totalValue.value}`
  return `${title} ${description}`
})

// ============================================================================
// Methods
// ============================================================================

/**
 * Handle mouse enter on segment or legend item
 */
function handleMouseEnter(index: number, event: MouseEvent) {
  hoveredIndex.value = index
  showTooltip(index, event)
}

/**
 * Handle mouse leave from segment or legend item
 */
function handleMouseLeave() {
  hoveredIndex.value = null
  hideTooltip()
}

/**
 * Show tooltip at cursor position
 */
function showTooltip(index: number, event: MouseEvent) {
  const segment = segments.value[index]
  if (!segment) return

  // Get chart container position
  const container = (event.currentTarget as HTMLElement).closest('.donut-chart-container')
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

/**
 * Hide tooltip
 */
function hideTooltip() {
  tooltip.value.visible = false
}

// ============================================================================
// Lifecycle
// ============================================================================

/**
 * Re-render chart when data changes
 */
watch(
  () => props.data,
  async () => {
    if (hasData.value) {
      await nextTick()
      // Chart re-renders automatically via computed properties
    }
  },
  { deep: true }
)

/**
 * Initialize chart on mount
 */
onMounted(async () => {
  await nextTick()
  // Chart renders automatically via template
})
</script>

<style scoped>
.donut-chart-container {
  position: relative;
  min-width: 250px;
  min-height: 250px;
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

/* Smooth transitions for arc animations */
.segment {
  transition: transform 200ms ease-in-out, opacity 200ms ease-in-out;
}

/* Ensure SVG text is selectable for accessibility */
text {
  user-select: none;
}

/* Legend hover effect */
[role="button"]:focus-visible {
  outline: 2px solid #3B82F6;
  outline-offset: 2px;
}
</style>

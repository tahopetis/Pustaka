<template>
  <div class="trend-chart-container" ref="containerRef">
    <svg
      ref="svgRef"
      :viewBox="`0 0 ${width} ${height}`"
      class="w-full"
      role="img"
      :aria-label="ariaLabel"
    >
      <!-- Gradient definitions for fills -->
      <defs>
        <linearGradient
          v-for="(series, index) in data"
          :key="`gradient-${series.id}`"
          :id="`gradient-${series.id}`"
          x1="0%"
          y1="0%"
          x2="0%"
          y2="100%"
        >
          <stop offset="0%" :stop-color="series.color" :stop-opacity="0.3" />
          <stop offset="100%" :stop-color="series.color" :stop-opacity="0.05" />
        </linearGradient>
      </defs>

      <!-- Chart axes and lines will be rendered here by D3 -->
      <g :transform="`translate(${margin.left}, ${margin.top})`" ref="chartRef">
        <!-- X Axis -->
        <g
          :transform="`translate(0, ${innerHeight})`"
          ref="xAxisRef"
          class="x-axis"
        />

        <!-- Y Axis -->
        <g ref="yAxisRef" class="y-axis" />

        <!-- Grid lines (optional) -->
        <g ref="gridRef" class="grid" />

        <!-- Chart area (lines and areas) -->
        <g ref="linesRef" class="lines-group" />

        <!-- Interaction overlay -->
        <rect
          :width="innerWidth"
          :height="innerHeight"
          fill="transparent"
          class="interaction-overlay"
          @mousemove="handleMouseMove"
          @mouseleave="handleMouseLeave"
        />

        <!-- Hover line -->
        <line
          v-if="hoveredDate"
          :x1="hoverLineX"
          :y1="0"
          :x2="hoverLineX"
          :y2="innerHeight"
          stroke="#9CA3AF"
          stroke-width="1"
          stroke-dasharray="4,4"
          class="hover-line"
        />

        <!-- Hover circles -->
        <g v-if="hoveredDate" class="hover-circles">
          <circle
            v-for="point in hoveredPoints"
            :key="point.seriesId"
            :cx="hoverLineX"
            :cy="point.y"
            r="4"
            :fill="point.color"
            :stroke="point.color"
            stroke-width="2"
            class="hover-circle"
          />
        </g>
      </g>
    </svg>

    <!-- Legend -->
    <div v-if="showLegend && data.length > 1" class="legend mt-4 flex flex-wrap gap-4 justify-center">
      <div
        v-for="series in data"
        :key="`legend-${series.id}`"
        class="legend-item flex items-center space-x-2"
      >
        <div
          class="legend-color w-4 h-4 rounded"
          :style="{ backgroundColor: series.color }"
        />
        <span class="text-sm text-gray-700">{{ series.name }}</span>
      </div>
    </div>

    <!-- Tooltip -->
    <div
      v-if="hoveredDate && tooltipData"
      ref="tooltipRef"
      class="tooltip absolute bg-gray-900 text-white px-3 py-2 rounded shadow-lg text-sm pointer-events-none z-10"
      :style="{ left: `${tooltipX}px`, top: `${tooltipY}px` }"
      role="tooltip"
    >
      <div class="font-semibold mb-1">{{ formatTooltipDate(hoveredDate) }}</div>
      <div
        v-for="item in tooltipData"
        :key="item.seriesId"
        class="flex items-center justify-between space-x-3"
      >
        <div class="flex items-center space-x-2">
          <div
            class="w-2 h-2 rounded-full"
            :style="{ backgroundColor: item.color }"
          />
          <span>{{ item.name }}:</span>
        </div>
        <span class="font-medium">{{ item.value }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as d3 from 'd3'
import type { ChartSeries } from '@/types/dashboard'

// ============================================================================
// Props and Emits
// ============================================================================

interface Props {
  /** Chart data as array of series */
  data: ChartSeries[]
  /** Chart title for accessibility */
  title?: string
  /** Chart height in pixels */
  height?: number
  /** X-axis label */
  xAxisLabel?: string
  /** Y-axis label */
  yAxisLabel?: string
  /** Show legend */
  showLegend?: boolean
  /** Show grid lines */
  showGrid?: boolean
  /** Animation duration in milliseconds */
  animationDuration?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Trend Chart',
  height: 300,
  xAxisLabel: '',
  yAxisLabel: '',
  showLegend: true,
  showGrid: true,
  animationDuration: 750,
})

// ============================================================================
// Template Refs
// ============================================================================

const containerRef = ref<HTMLDivElement | null>(null)
const svgRef = ref<SVGSVGElement | null>(null)
const chartRef = ref<SVGGElement | null>(null)
const xAxisRef = ref<SVGGElement | null>(null)
const yAxisRef = ref<SVGGElement | null>(null)
const gridRef = ref<SVGGElement | null>(null)
const linesRef = ref<SVGGElement | null>(null)
const tooltipRef = ref<HTMLDivElement | null>(null)

// ============================================================================
// State
// ============================================================================

const width = ref(800)
const margin = { top: 20, right: 30, bottom: 40, left: 60 }

const hoveredDate = ref<Date | null>(null)
const hoverLineX = ref(0)
const tooltipX = ref(0)
const tooltipY = ref(0)

// ============================================================================
// Computed Properties
// ============================================================================

const innerWidth = computed(() => width.value - margin.left - margin.right)
const innerHeight = computed(() => props.height - margin.top - margin.bottom)

const ariaLabel = computed(() => {
  const seriesNames = props.data.map((s) => s.name).join(', ')
  return `${props.title}: Line chart showing trends for ${seriesNames}`
})

/**
 * Parse all data points and convert date strings to Date objects
 */
const parsedData = computed(() => {
  return props.data.map((series) => ({
    ...series,
    data: series.data.map((point) => ({
      ...point,
      date: new Date(point.x),
      value: point.y,
    })),
  }))
})

/**
 * Get all unique dates across all series
 */
const allDates = computed(() => {
  const dates = new Set<number>()
  parsedData.value.forEach((series) => {
    series.data.forEach((point) => {
      dates.add(point.date.getTime())
    })
  })
  return Array.from(dates)
    .map((time) => new Date(time))
    .sort((a, b) => a.getTime() - b.getTime())
})

/**
 * Get min and max values for Y-axis domain
 */
const yDomain = computed(() => {
  const allValues = parsedData.value.flatMap((series) =>
    series.data.map((point) => point.value)
  )
  const min = Math.min(...allValues, 0)
  const max = Math.max(...allValues)
  // Add 10% padding
  const padding = (max - min) * 0.1
  return [Math.max(0, min - padding), max + padding]
})

/**
 * Hovered data points for tooltip
 */
const hoveredPoints = computed(() => {
  if (!hoveredDate.value) return []

  const points: Array<{
    seriesId: string
    name: string
    color: string
    y: number
    value: number
  }> = []

  parsedData.value.forEach((series) => {
    const point = series.data.find(
      (p) => p.date.getTime() === hoveredDate.value!.getTime()
    )
    if (point && yScale.value) {
      points.push({
        seriesId: series.id,
        name: series.name,
        color: series.color,
        y: yScale.value(point.value),
        value: point.value,
      })
    }
  })

  return points
})

/**
 * Tooltip data for display
 */
const tooltipData = computed(() => {
  if (!hoveredDate.value) return null

  return parsedData.value
    .map((series) => {
      const point = series.data.find(
        (p) => p.date.getTime() === hoveredDate.value!.getTime()
      )
      if (point) {
        return {
          seriesId: series.id,
          name: series.name,
          color: series.color,
          value: point.value,
        }
      }
      return null
    })
    .filter(Boolean)
})

// ============================================================================
// D3 Scales
// ============================================================================

const xScale = ref<d3.ScaleTime<number, number> | null>(null)
const yScale = ref<d3.ScaleLinear<number, number> | null>(null)

/**
 * Initialize D3 scales
 */
function initScales() {
  if (allDates.value.length === 0) return

  xScale.value = d3
    .scaleTime()
    .domain(d3.extent(allDates.value) as [Date, Date])
    .range([0, innerWidth.value])

  yScale.value = d3
    .scaleLinear()
    .domain(yDomain.value)
    .range([innerHeight.value, 0])
    .nice()
}

// ============================================================================
// Rendering Functions
// ============================================================================

/**
 * Render X-axis
 */
function renderXAxis() {
  if (!xAxisRef.value || !xScale.value) return

  const xAxis = d3
    .axisBottom(xScale.value)
    .ticks(6)
    .tickFormat((d) => d3.timeFormat('%b %d')(d as Date))

  d3.select(xAxisRef.value)
    .transition()
    .duration(props.animationDuration)
    .call(xAxis as any)

  // Style axis
  d3.select(xAxisRef.value)
    .selectAll('text')
    .attr('fill', '#6B7280')
    .style('font-size', '12px')

  d3.select(xAxisRef.value).selectAll('line').attr('stroke', '#D1D5DB')

  d3.select(xAxisRef.value).select('.domain').attr('stroke', '#D1D5DB')
}

/**
 * Render Y-axis
 */
function renderYAxis() {
  if (!yAxisRef.value || !yScale.value) return

  const yAxis = d3.axisLeft(yScale.value).ticks(5)

  d3.select(yAxisRef.value)
    .transition()
    .duration(props.animationDuration)
    .call(yAxis as any)

  // Style axis
  d3.select(yAxisRef.value)
    .selectAll('text')
    .attr('fill', '#6B7280')
    .style('font-size', '12px')

  d3.select(yAxisRef.value).selectAll('line').attr('stroke', '#D1D5DB')

  d3.select(yAxisRef.value).select('.domain').attr('stroke', '#D1D5DB')
}

/**
 * Render grid lines
 */
function renderGrid() {
  if (!gridRef.value || !yScale.value || !props.showGrid) return

  const gridLines = d3
    .axisLeft(yScale.value)
    .ticks(5)
    .tickSize(-innerWidth.value)
    .tickFormat(() => '')

  d3.select(gridRef.value).call(gridLines as any)

  // Style grid
  d3.select(gridRef.value)
    .selectAll('line')
    .attr('stroke', '#F3F4F6')
    .attr('stroke-opacity', 0.7)

  d3.select(gridRef.value).select('.domain').remove()
}

/**
 * Render lines and area fills
 */
function renderLines() {
  if (!linesRef.value || !xScale.value || !yScale.value) return

  const linesGroup = d3.select(linesRef.value)

  // Line generator
  const lineGenerator = d3
    .line<any>()
    .x((d) => xScale.value!(d.date))
    .y((d) => yScale.value!(d.value))
    .curve(d3.curveMonotoneX) // Smooth curves

  // Area generator for gradient fill
  const areaGenerator = d3
    .area<any>()
    .x((d) => xScale.value!(d.date))
    .y0(innerHeight.value)
    .y1((d) => yScale.value!(d.value))
    .curve(d3.curveMonotoneX)

  // Bind data
  const seriesGroups = linesGroup
    .selectAll('.series-group')
    .data(parsedData.value, (d: any) => d.id)

  // Remove old series
  seriesGroups.exit().remove()

  // Add new series
  const seriesEnter = seriesGroups
    .enter()
    .append('g')
    .attr('class', 'series-group')

  // Merge enter + update
  const seriesMerge = seriesEnter.merge(seriesGroups as any)

  // Render area fills
  const areas = seriesMerge.selectAll('.area').data((d: any) => [d])

  areas.exit().remove()

  const areasEnter = areas
    .enter()
    .append('path')
    .attr('class', 'area')
    .attr('fill', (d: any) => `url(#gradient-${d.id})`)
    .attr('d', (d: any) => areaGenerator(d.data))
    .attr('opacity', 0)

  areasEnter
    .merge(areas as any)
    .transition()
    .duration(props.animationDuration)
    .attr('d', (d: any) => areaGenerator(d.data))
    .attr('opacity', 1)

  // Render lines
  const lines = seriesMerge.selectAll('.line').data((d: any) => [d])

  lines.exit().remove()

  const linesEnter = lines
    .enter()
    .append('path')
    .attr('class', 'line')
    .attr('fill', 'none')
    .attr('stroke', (d: any) => d.color)
    .attr('stroke-width', 2)
    .attr('d', (d: any) => lineGenerator(d.data))
    .attr('opacity', 0)

  linesEnter
    .merge(lines as any)
    .transition()
    .duration(props.animationDuration)
    .attr('stroke', (d: any) => d.color)
    .attr('d', (d: any) => lineGenerator(d.data))
    .attr('opacity', 1)
}

/**
 * Render the complete chart
 */
function renderChart() {
  initScales()
  renderXAxis()
  renderYAxis()
  renderGrid()
  renderLines()
}

// ============================================================================
// Interaction Handlers
// ============================================================================

/**
 * Handle mouse move for tooltip
 */
function handleMouseMove(event: MouseEvent) {
  if (!xScale.value || !containerRef.value) return

  const rect = (event.currentTarget as SVGRectElement).getBoundingClientRect()
  const x = event.clientX - rect.left

  // Find closest date
  const dateAtX = xScale.value.invert(x)
  const closestDate = allDates.value.reduce((prev, curr) => {
    return Math.abs(curr.getTime() - dateAtX.getTime()) <
      Math.abs(prev.getTime() - dateAtX.getTime())
      ? curr
      : prev
  })

  hoveredDate.value = closestDate
  hoverLineX.value = xScale.value(closestDate)

  // Position tooltip
  const containerRect = containerRef.value.getBoundingClientRect()
  tooltipX.value = event.clientX - containerRect.left + 10
  tooltipY.value = event.clientY - containerRect.top - 10
}

/**
 * Handle mouse leave
 */
function handleMouseLeave() {
  hoveredDate.value = null
}

/**
 * Format date for tooltip
 */
function formatTooltipDate(date: Date): string {
  return d3.timeFormat('%B %d, %Y')(date)
}

// ============================================================================
// Resize Handling
// ============================================================================

let resizeObserver: ResizeObserver | null = null

function handleResize() {
  if (!containerRef.value) return
  const containerWidth = containerRef.value.clientWidth
  width.value = Math.max(300, containerWidth)
  nextTick(() => {
    renderChart()
  })
}

// ============================================================================
// Lifecycle
// ============================================================================

onMounted(() => {
  handleResize()
  renderChart()

  // Setup resize observer
  if (containerRef.value) {
    resizeObserver = new ResizeObserver(handleResize)
    resizeObserver.observe(containerRef.value)
  }
})

onUnmounted(() => {
  if (resizeObserver && containerRef.value) {
    resizeObserver.unobserve(containerRef.value)
    resizeObserver.disconnect()
  }
})

// Watch for data changes
watch(
  () => props.data,
  () => {
    nextTick(() => {
      renderChart()
    })
  },
  { deep: true }
)

// Watch for height changes
watch(
  () => props.height,
  () => {
    nextTick(() => {
      renderChart()
    })
  }
)
</script>

<style scoped>
.trend-chart-container {
  position: relative;
  width: 100%;
}

/* Axis styling */
:deep(.x-axis text),
:deep(.y-axis text) {
  font-family: system-ui, -apple-system, sans-serif;
}

/* Grid lines */
:deep(.grid line) {
  shape-rendering: crispEdges;
}

/* Tooltip */
.tooltip {
  transform: translate(10px, -10px);
  white-space: nowrap;
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* Hover effects */
.hover-line {
  pointer-events: none;
}

.hover-circle {
  pointer-events: none;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

/* Legend */
.legend-item {
  cursor: default;
  user-select: none;
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .tooltip {
    animation: none;
  }
}
</style>

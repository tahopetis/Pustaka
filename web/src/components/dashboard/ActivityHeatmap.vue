<template>
  <div class="activity-heatmap">
    <h3 v-if="title" class="text-lg font-semibold mb-4 text-gray-800">
      {{ title }}
    </h3>

    <!-- Empty State -->
    <div
      v-if="!hasData"
      class="flex flex-col items-center justify-center py-12 text-gray-400"
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
          d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
        />
      </svg>
      <p class="text-sm">No activity data available</p>
    </div>

    <!-- Heatmap -->
    <div v-else class="heatmap-container">
      <!-- Month Labels -->
      <div class="flex gap-2 mb-2">
        <div class="w-12"></div>
        <div class="flex-1 flex justify-between text-xs text-gray-600">
          <span v-for="month in monthLabels" :key="month.label" :style="{ marginLeft: month.offset + '%' }">
            {{ month.label }}
          </span>
        </div>
      </div>

      <!-- Heatmap Grid -->
      <div class="flex gap-2">
        <!-- Day Labels -->
        <div class="flex flex-col justify-between text-xs text-gray-600 w-12">
          <span>Mon</span>
          <span>Wed</span>
          <span>Fri</span>
        </div>

        <!-- Calendar Grid -->
        <div class="flex-1 overflow-x-auto">
          <div class="inline-flex gap-1" :style="{ minWidth: gridWidth + 'px' }">
            <!-- Week Columns -->
            <div
              v-for="(week, weekIndex) in weeks"
              :key="weekIndex"
              class="flex flex-col gap-1"
            >
              <!-- Day Cells -->
              <div
                v-for="(day, dayIndex) in week"
                :key="dayIndex"
                class="cell transition-all duration-200 cursor-pointer"
                :class="getCellClass(day)"
                :style="getCellStyle(day)"
                :title="getTooltipText(day)"
                @mouseenter="handleCellHover(day, $event)"
                @mouseleave="handleCellLeave"
                @click="handleCellClick(day)"
                role="button"
                tabindex="0"
                :aria-label="getTooltipText(day)"
                @keydown.enter="handleCellClick(day)"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Legend -->
      <div class="flex items-center justify-end gap-2 mt-4 text-xs text-gray-600">
        <span>Less</span>
        <div class="flex gap-1">
          <div
            v-for="level in 5"
            :key="level"
            class="w-3 h-3 rounded-sm"
            :class="getLegendClass(level - 1)"
          ></div>
        </div>
        <span>More</span>
      </div>

      <!-- Tooltip -->
      <div
        v-if="hoveredDay"
        ref="tooltipRef"
        class="absolute z-10 px-3 py-2 text-sm bg-gray-900 text-white rounded-lg shadow-xl pointer-events-none"
        :style="{
          left: tooltipX + 'px',
          top: tooltipY + 'px',
          transform: 'translate(-50%, -120%)',
        }"
      >
        <div class="font-semibold">{{ formatDate(hoveredDay.date) }}</div>
        <div class="text-gray-300">
          {{ hoveredDay.count }} {{ hoveredDay.count === 1 ? 'event' : 'events' }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { HeatmapData } from '@/types/dashboard'

// ============================================================================
// Props and Emits
// ============================================================================

interface Props {
  /** Heatmap data with dates and counts */
  data: Record<string, number>
  /** Optional title for the heatmap */
  title?: string
  /** Number of weeks to display (default: 12 weeks = ~3 months) */
  weeks?: number
  /** Cell size in pixels */
  cellSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Activity Heatmap',
  weeks: 12,
  cellSize: 12,
})

interface Emits {
  /** Emitted when a day cell is clicked */
  (e: 'dayClick', date: string, count: number): void
}

const emit = defineEmits<Emits>()

// ============================================================================
// State
// ============================================================================

const hoveredDay = ref<{ date: Date; count: number } | null>(null)
const tooltipRef = ref<HTMLDivElement | null>(null)
const tooltipX = ref(0)
const tooltipY = ref(0)

// ============================================================================
// Computed Properties
// ============================================================================

/**
 * Check if there's any data
 */
const hasData = computed(() => {
  return Object.keys(props.data).length > 0
})

/**
 * Grid width based on number of weeks
 */
const gridWidth = computed(() => {
  return props.weeks * (props.cellSize + 4) // cell size + gap
})

/**
 * Generate weeks array for the heatmap
 * Each week contains 7 days (Sunday to Saturday)
 */
const weeks = computed(() => {
  const today = new Date()
  const endDate = new Date(today)
  endDate.setHours(23, 59, 59, 999)

  const startDate = new Date(endDate)
  startDate.setDate(startDate.getDate() - (props.weeks * 7) + 1)
  startDate.setHours(0, 0, 0, 0)

  // Adjust to start on Sunday
  const dayOfWeek = startDate.getDay()
  if (dayOfWeek !== 0) {
    startDate.setDate(startDate.getDate() - dayOfWeek)
  }

  const weeksArray: Array<Array<{ date: Date; count: number }>> = []
  let currentWeek: Array<{ date: Date; count: number }> = []

  const currentDate = new Date(startDate)
  while (currentDate <= endDate) {
    const dateString = formatDateKey(currentDate)
    const count = props.data[dateString] || 0

    currentWeek.push({
      date: new Date(currentDate),
      count,
    })

    if (currentWeek.length === 7) {
      weeksArray.push(currentWeek)
      currentWeek = []
    }

    currentDate.setDate(currentDate.getDate() + 1)
  }

  // Add remaining days if any
  if (currentWeek.length > 0) {
    while (currentWeek.length < 7) {
      currentWeek.push({
        date: new Date(currentDate),
        count: 0,
      })
      currentDate.setDate(currentDate.getDate() + 1)
    }
    weeksArray.push(currentWeek)
  }

  return weeksArray
})

/**
 * Month labels for the header
 */
const monthLabels = computed(() => {
  if (weeks.value.length === 0) return []

  const labels: Array<{ label: string; offset: number }> = []
  const months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
  ]

  let currentMonth = -1
  weeks.value.forEach((week, weekIndex) => {
    const firstDay = week[0]
    const month = firstDay.date.getMonth()

    if (month !== currentMonth) {
      currentMonth = month
      labels.push({
        label: months[month],
        offset: (weekIndex / weeks.value.length) * 100,
      })
    }
  })

  return labels
})

/**
 * Maximum activity count for color scaling
 */
const maxCount = computed(() => {
  const counts = Object.values(props.data)
  return counts.length > 0 ? Math.max(...counts) : 0
})

// ============================================================================
// Methods
// ============================================================================

/**
 * Format date as YYYY-MM-DD for data lookup
 */
function formatDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Format date for display in tooltip
 */
function formatDate(date: Date): string {
  const options: Intl.DateTimeFormatOptions = {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }
  return date.toLocaleDateString('en-US', options)
}

/**
 * Get activity level (0-4) for a given count
 */
function getActivityLevel(count: number): number {
  if (count === 0) return 0
  if (maxCount.value === 0) return 0

  const percentage = (count / maxCount.value) * 100
  if (percentage <= 20) return 1
  if (percentage <= 40) return 2
  if (percentage <= 60) return 3
  return 4
}

/**
 * Get CSS class for cell based on activity level
 */
function getCellClass(day: { date: Date; count: number }): string {
  const level = getActivityLevel(day.count)
  const isToday = formatDateKey(day.date) === formatDateKey(new Date())

  return `activity-level-${level} ${isToday ? 'ring-2 ring-blue-500 ring-offset-1' : ''}`
}

/**
 * Get inline styles for cell
 */
function getCellStyle(day: { date: Date; count: number }) {
  return {
    width: props.cellSize + 'px',
    height: props.cellSize + 'px',
  }
}

/**
 * Get CSS class for legend item
 */
function getLegendClass(level: number): string {
  return `activity-level-${level}`
}

/**
 * Get tooltip text for a day
 */
function getTooltipText(day: { date: Date; count: number }): string {
  return `${formatDate(day.date)}: ${day.count} ${day.count === 1 ? 'event' : 'events'}`
}

/**
 * Handle cell hover
 */
function handleCellHover(day: { date: Date; count: number }, event: MouseEvent) {
  hoveredDay.value = day

  const container = (event.currentTarget as HTMLElement).closest('.activity-heatmap')
  if (!container) return

  const rect = container.getBoundingClientRect()
  const targetRect = (event.currentTarget as HTMLElement).getBoundingClientRect()

  tooltipX.value = targetRect.left - rect.left + targetRect.width / 2
  tooltipY.value = targetRect.top - rect.top
}

/**
 * Handle cell leave
 */
function handleCellLeave() {
  hoveredDay.value = null
}

/**
 * Handle cell click
 */
function handleCellClick(day: { date: Date; count: number }) {
  emit('dayClick', formatDateKey(day.date), day.count)
}
</script>

<style scoped>
.activity-heatmap {
  position: relative;
  width: 100%;
}

.heatmap-container {
  position: relative;
}

/* Activity Level Colors (GitHub-style) */
.cell {
  border-radius: 2px;
  border: 1px solid rgba(0, 0, 0, 0.05);
}

.activity-level-0 {
  background-color: #ebedf0;
}

.activity-level-1 {
  background-color: #9be9a8;
}

.activity-level-2 {
  background-color: #40c463;
}

.activity-level-3 {
  background-color: #30a14e;
}

.activity-level-4 {
  background-color: #216e39;
}

/* Hover Effects */
.cell:hover {
  outline: 2px solid rgba(59, 130, 246, 0.5);
  outline-offset: -1px;
  transform: scale(1.1);
  z-index: 1;
}

.cell:focus-visible {
  outline: 2px solid #3B82F6;
  outline-offset: 1px;
  transform: scale(1.1);
  z-index: 1;
}

/* Scrollbar Styling */
.overflow-x-auto::-webkit-scrollbar {
  height: 6px;
}

.overflow-x-auto::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.overflow-x-auto::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.overflow-x-auto::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .cell {
    transition: none;
  }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .activity-level-0 {
    background-color: #161b22;
  }

  .activity-level-1 {
    background-color: #0e4429;
  }

  .activity-level-2 {
    background-color: #006d32;
  }

  .activity-level-3 {
    background-color: #26a641;
  }

  .activity-level-4 {
    background-color: #39d353;
  }
}
</style>

<template>
  <div class="network-analytics-card">
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
          d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
        />
      </svg>
      <p class="text-sm">No network data available</p>
    </div>

    <!-- Chart -->
    <div v-else class="space-y-3">
      <div
        v-for="(item, index) in sortedData"
        :key="item.id"
        class="relative cursor-pointer group transition-all"
        @mouseenter="handleMouseEnter(index, $event)"
        @mouseleave="handleMouseLeave"
        @click="handleItemClick(item)"
        role="button"
        tabindex="0"
        :aria-label="`${item.name}: ${item.connection_count} connections`"
        @keydown.enter="handleItemClick(item)"
        @keydown.space.prevent="handleItemClick(item)"
      >
        <!-- Bar Container -->
        <div class="flex items-center gap-3">
          <!-- Label -->
          <div class="flex-shrink-0 w-32 sm:w-40">
            <div class="text-sm font-medium text-gray-700 truncate" :title="item.name">
              {{ item.name }}
            </div>
            <div class="text-xs text-gray-500">
              {{ item.ci_type_name }}
            </div>
          </div>

          <!-- Bar -->
          <div class="flex-1 relative h-8 bg-gray-100 rounded-lg overflow-hidden">
            <!-- Background Bar -->
            <div
              class="absolute inset-0 transition-all duration-500 ease-out rounded-lg"
              :style="{
                width: `${getBarWidth(item.connection_count)}%`,
                background: getGradient(index),
                transform: hoveredIndex === index ? 'scaleY(1.1)' : 'scaleY(1)',
              }"
            ></div>

            <!-- Connection Count Label -->
            <div
              class="absolute inset-0 flex items-center justify-end pr-3 text-sm font-semibold transition-opacity"
              :class="
                getBarWidth(item.connection_count) > 20
                  ? 'text-white'
                  : 'text-gray-700'
              "
              :style="{
                opacity: hoveredIndex === index ? 1 : 0.9,
              }"
            >
              {{ item.connection_count }}
            </div>
          </div>

          <!-- Rank Badge -->
          <div
            class="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all"
            :class="getRankBadgeClass(index)"
            :style="{
              transform: hoveredIndex === index ? 'scale(1.1)' : 'scale(1)',
            }"
          >
            {{ index + 1 }}
          </div>
        </div>

        <!-- Hover Indicator -->
        <div
          v-if="hoveredIndex === index"
          class="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 rounded-r"
        ></div>
      </div>
    </div>

    <!-- Tooltip -->
    <div
      v-if="tooltip.visible"
      ref="tooltipRef"
      class="absolute z-20 px-3 py-2 text-sm bg-gray-900 text-white rounded-lg shadow-xl pointer-events-none"
      :style="{
        left: tooltip.x + 'px',
        top: tooltip.y + 'px',
        transform: 'translate(-50%, -120%)',
      }"
    >
      <div class="font-semibold mb-1">{{ tooltip.name }}</div>
      <div class="text-gray-300 text-xs space-y-0.5">
        <div>Type: {{ tooltip.type }}</div>
        <div>Connections: {{ tooltip.connections }}</div>
        <div v-if="tooltip.domain">Domain: {{ tooltip.domain }}</div>
      </div>
      <div class="text-xs text-gray-400 mt-1">Click to view details</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { MostConnectedCI } from '@/types/dashboard'

// ============================================================================
// Props and Emits
// ============================================================================

interface Props {
  /** Most connected CIs data */
  data: MostConnectedCI[]
  /** Optional title for the card */
  title?: string
  /** Maximum number of items to display */
  limit?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Most Connected CIs',
  limit: 10,
})

interface Emits {
  /** Emitted when a CI item is clicked */
  (e: 'itemClick', item: MostConnectedCI): void
}

const emit = defineEmits<Emits>()

// ============================================================================
// State
// ============================================================================

const hoveredIndex = ref<number | null>(null)
const tooltipRef = ref<HTMLDivElement | null>(null)
const tooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  name: '',
  type: '',
  connections: 0,
  domain: '',
})

// ============================================================================
// Computed Properties
// ============================================================================

/**
 * Check if there's any data
 */
const hasData = computed(() => {
  return props.data && props.data.length > 0
})

/**
 * Sorted and limited data
 */
const sortedData = computed(() => {
  return [...props.data]
    .sort((a, b) => b.connection_count - a.connection_count)
    .slice(0, props.limit)
})

/**
 * Maximum connection count for scaling bars
 */
const maxConnections = computed(() => {
  if (sortedData.value.length === 0) return 1
  return Math.max(...sortedData.value.map((item) => item.connection_count))
})

// ============================================================================
// Methods
// ============================================================================

/**
 * Calculate bar width as percentage
 */
function getBarWidth(count: number): number {
  if (maxConnections.value === 0) return 0
  return (count / maxConnections.value) * 100
}

/**
 * Get gradient color for bar based on rank
 */
function getGradient(index: number): string {
  const gradients = [
    'linear-gradient(90deg, #3B82F6, #2563EB)', // blue (1st)
    'linear-gradient(90deg, #10B981, #059669)', // green (2nd)
    'linear-gradient(90deg, #F59E0B, #D97706)', // amber (3rd)
    'linear-gradient(90deg, #8B5CF6, #7C3AED)', // violet
    'linear-gradient(90deg, #EC4899, #DB2777)', // pink
    'linear-gradient(90deg, #06B6D4, #0891B2)', // cyan
    'linear-gradient(90deg, #F97316, #EA580C)', // orange
    'linear-gradient(90deg, #6366F1, #4F46E5)', // indigo
    'linear-gradient(90deg, #14B8A6, #0D9488)', // teal
    'linear-gradient(90deg, #84CC16, #65A30D)', // lime
  ]
  return gradients[index % gradients.length]
}

/**
 * Get rank badge styling classes
 */
function getRankBadgeClass(index: number): string {
  if (index === 0) return 'bg-yellow-100 text-yellow-700 ring-2 ring-yellow-400'
  if (index === 1) return 'bg-gray-100 text-gray-700 ring-2 ring-gray-400'
  if (index === 2) return 'bg-orange-100 text-orange-700 ring-2 ring-orange-400'
  return 'bg-blue-50 text-blue-600'
}

/**
 * Handle mouse enter on item
 */
function handleMouseEnter(index: number, event: MouseEvent) {
  hoveredIndex.value = index
  showTooltip(index, event)
}

/**
 * Handle mouse leave
 */
function handleMouseLeave() {
  hoveredIndex.value = null
  hideTooltip()
}

/**
 * Show tooltip
 */
function showTooltip(index: number, event: MouseEvent) {
  const item = sortedData.value[index]
  if (!item) return

  const container = (event.currentTarget as HTMLElement).closest('.network-analytics-card')
  if (!container) return

  const rect = container.getBoundingClientRect()
  const targetRect = (event.currentTarget as HTMLElement).getBoundingClientRect()

  tooltip.value = {
    visible: true,
    x: targetRect.left - rect.left + targetRect.width / 2,
    y: targetRect.top - rect.top,
    name: item.name,
    type: item.ci_type_name,
    connections: item.connection_count,
    domain: item.domain_name || '',
  }
}

/**
 * Hide tooltip
 */
function hideTooltip() {
  tooltip.value.visible = false
}

/**
 * Handle item click
 */
function handleItemClick(item: MostConnectedCI) {
  emit('itemClick', item)
}
</script>

<style scoped>
.network-analytics-card {
  position: relative;
  width: 100%;
  min-height: 200px;
}

/* Hover Effects */
[role="button"]:hover {
  background-color: rgba(59, 130, 246, 0.02);
  border-radius: 0.5rem;
}

[role="button"]:focus-visible {
  outline: 2px solid #3B82F6;
  outline-offset: 2px;
  border-radius: 0.5rem;
}

/* Smooth Transitions */
.transition-all {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 200ms;
}

/* Animations */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

[role="button"] {
  animation: slideIn 0.3s ease-out forwards;
}

/* Stagger animation delay */
[role="button"]:nth-child(1) {
  animation-delay: 0ms;
}
[role="button"]:nth-child(2) {
  animation-delay: 50ms;
}
[role="button"]:nth-child(3) {
  animation-delay: 100ms;
}
[role="button"]:nth-child(4) {
  animation-delay: 150ms;
}
[role="button"]:nth-child(5) {
  animation-delay: 200ms;
}
[role="button"]:nth-child(6) {
  animation-delay: 250ms;
}
[role="button"]:nth-child(7) {
  animation-delay: 300ms;
}
[role="button"]:nth-child(8) {
  animation-delay: 350ms;
}
[role="button"]:nth-child(9) {
  animation-delay: 400ms;
}
[role="button"]:nth-child(10) {
  animation-delay: 450ms;
}
</style>

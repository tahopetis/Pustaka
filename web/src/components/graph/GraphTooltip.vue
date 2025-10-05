<template>
  <Teleport to="body">
    <div
      v-if="visible"
      ref="tooltipRef"
      :style="{
        position: 'fixed',
        left: x + 'px',
        top: y + 'px',
        zIndex: 9999,
        pointerEvents: 'auto'
      }"
      class="graph-tooltip"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
    >
      <div class="bg-white rounded-xl shadow-2xl border border-gray-200 p-0 min-w-80 max-w-96 overflow-hidden">
        <!-- Header -->
        <div class="bg-gradient-to-r from-blue-50 to-indigo-50 px-4 py-3 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div
                class="w-8 h-8 rounded-full border-2 border-white shadow-sm flex items-center justify-center"
                :style="{ backgroundColor: getNodeColor() }"
              >
                <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 text-sm">{{ node?.name }}</h3>
                <p class="text-xs text-gray-600 font-medium">{{ node?.type }}</p>
              </div>
            </div>
            <div class="flex items-center space-x-1">
              <!-- Copy ID button -->
              <button
                @click="copyToClipboard"
                class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition-colors"
                title="Copy ID"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                </svg>
              </button>
              <!-- View details button -->
              <button
                @click="viewDetails"
                class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition-colors"
                title="View Details"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Attributes Section -->
        <div class="p-4">
          <div v-if="hasAttributes" class="space-y-3">
            <h4 class="text-xs font-semibold text-gray-700 uppercase tracking-wider mb-3">Attributes</h4>

            <div class="space-y-2 max-h-64 overflow-y-auto">
              <div
                v-for="(value, key) in sortedAttributes"
                :key="key"
                class="group flex items-start justify-between py-2 px-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors"
              >
                <div class="flex-1 min-w-0">
                  <p class="text-xs font-medium text-gray-700 truncate">{{ formatAttributeName(key) }}</p>
                  <p class="text-xs text-gray-500 mt-0.5">{{ formatAttributeValue(value) }}</p>
                </div>
                <div class="flex items-center ml-2 space-x-1">
                  <!-- Attribute type icon -->
                  <div class="w-5 h-5 flex items-center justify-center">
                    <svg
                      v-if="typeof value === 'boolean'"
                      class="w-4 h-4"
                      :class="value ? 'text-green-500' : 'text-red-500'"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path v-if="value" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                      <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6"></path>
                    </svg>
                    <svg
                      v-else-if="typeof value === 'number'"
                      class="w-4 h-4 text-blue-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14"></path>
                    </svg>
                    <svg
                      v-else
                      class="w-4 h-4 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                    </svg>
                  </div>
                  <!-- Copy attribute button -->
                  <button
                    @click="copyAttributeToClipboard(key, value)"
                    class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-gray-600 hover:bg-white rounded transition-all"
                    title="Copy attribute"
                  >
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"></path>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- No attributes state -->
          <div v-else class="text-center py-6">
            <svg class="mx-auto h-8 w-8 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
            </svg>
            <p class="text-xs text-gray-500">No attributes defined</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-4 py-2 border-t border-gray-200">
          <div class="flex items-center justify-between">
            <p class="text-xs text-gray-500">ID: {{ node?.id }}</p>
            <p class="text-xs text-gray-400">{{ attributeCount }} attributes</p>
          </div>
        </div>
      </div>

      <!-- Tooltip arrow -->
      <div class="absolute -bottom-2 left-1/2 transform -translate-x-1/2">
        <div class="w-0 h-0 border-l-8 border-r-8 border-t-8 border-transparent border-t-white"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { showSuccessToast } from '@/utils/toast'

interface Props {
  node: any
  visible: boolean
  x: number
  y: number
}

const props = defineProps<Props>()
const emit = defineEmits<{
  hide: []
}>()

const tooltipRef = ref(null)
const hideTimeout = ref<NodeJS.Timeout | null>(null)

// Computed properties
const hasAttributes = computed(() => {
  return props.node?.attributes && Object.keys(props.node.attributes).length > 0
})

const sortedAttributes = computed(() => {
  if (!props.node?.attributes) return {}

  const entries = Object.entries(props.node.attributes)
  entries.sort(([keyA], [keyB]) => keyA.localeCompare(keyB))
  return Object.fromEntries(entries)
})

const attributeCount = computed(() => {
  return props.node?.attributes ? Object.keys(props.node.attributes).length : 0
})

// Methods
const getNodeColor = () => {
  const ciTypeColors = {
    'Server': '#3b82f6',
    'Application': '#10b981',
    'Database': '#f59e0b',
    'default': '#6b7280',
  }

  return ciTypeColors[props.node?.type as keyof typeof ciTypeColors] || ciTypeColors.default
}

const formatAttributeName = (key: string) => {
  return key
    .split(/[_\-\s]+/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

const formatAttributeValue = (value: any) => {
  if (value === null || value === undefined) {
    return '—'
  }

  if (typeof value === 'boolean') {
    return value ? 'True' : 'False'
  }

  if (typeof value === 'number') {
    return value.toLocaleString()
  }

  if (typeof value === 'string') {
    // Truncate long strings
    return value.length > 100 ? value.substring(0, 100) + '...' : value
  }

  if (Array.isArray(value)) {
    return `Array (${value.length} items)`
  }

  if (typeof value === 'object') {
    return `Object (${Object.keys(value).length} keys)`
  }

  return String(value)
}

const copyToClipboard = async () => {
  try {
    const text = `CI: ${props.node?.name}\nType: ${props.node?.type}\nID: ${props.node?.id}`
    await navigator.clipboard.writeText(text)
    showSuccessToast('CI information copied to clipboard')
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
  }
}

const copyAttributeToClipboard = async (key: string, value: any) => {
  try {
    const text = `${key}: ${formatAttributeValue(value)}`
    await navigator.clipboard.writeText(text)
    showSuccessToast('Attribute copied to clipboard')
  } catch (error) {
    console.error('Failed to copy attribute to clipboard:', error)
  }
}

const viewDetails = () => {
  if (props.node?.id) {
    window.location.href = `/ci/${props.node.id}`
  }
}

const onMouseEnter = () => {
  // Clear any pending hide timeout
  if (hideTimeout.value) {
    clearTimeout(hideTimeout.value)
    hideTimeout.value = null
  }
}

const onMouseLeave = () => {
  // Hide tooltip after a short delay to allow smooth transitions
  hideTimeout.value = setTimeout(() => {
    emit('hide')
  }, 100)
}

// Watch for position changes and adjust tooltip if it goes off-screen
watch([() => props.x, () => props.y], async () => {
  if (props.visible && tooltipRef.value) {
    await nextTick()
    adjustTooltipPosition()
  }
})

const adjustTooltipPosition = () => {
  if (!tooltipRef.value) return

  const tooltip = tooltipRef.value
  const rect = tooltip.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  let adjustedX = props.x
  let adjustedY = props.y

  // Adjust horizontal position if tooltip goes off screen
  if (rect.right > viewportWidth) {
    adjustedX = viewportWidth - rect.width - 20
  }
  if (rect.left < 20) {
    adjustedX = 20
  }

  // Adjust vertical position if tooltip goes off screen
  if (rect.bottom > viewportHeight) {
    adjustedY = viewportHeight - rect.height - 20
  }
  if (rect.top < 20) {
    adjustedY = 20
  }

  // Apply adjustments
  tooltip.style.left = adjustedX + 'px'
  tooltip.style.top = adjustedY + 'px'
}

// Clean up timeout on unmount
const cleanup = () => {
  if (hideTimeout.value) {
    clearTimeout(hideTimeout.value)
    hideTimeout.value = null
  }
}

// Expose cleanup method
defineExpose({
  cleanup
})
</script>

<style scoped>
.graph-tooltip {
  transition: opacity 0.2s ease-in-out;
}

/* Custom scrollbar for attributes */
.max-h-64::-webkit-scrollbar {
  width: 4px;
}

.max-h-64::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 2px;
}

.max-h-64::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}

.max-h-64::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* Smooth transitions for hover effects */
.group:hover .group-hover\:opacity-100 {
  opacity: 1;
  transition: opacity 0.2s ease-in-out;
}

/* Tooltip arrow shadow */
.tooltip-arrow {
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}
</style>
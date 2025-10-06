<template>
  <span
    :class="[
      'badge inline-flex items-center',
      size === 'small' ? 'text-xs px-2 py-1' : 'text-sm px-3 py-1'
    ]"
    :style="badgeStyle"
  >
    <svg
      v-if="relationshipType?.icon && showIcon"
      :class="[
        'w-3 h-3 mr-1',
        size === 'small' ? 'w-3 h-3' : 'w-4 h-4'
      ]"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path>
    </svg>
    {{ displayText }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRelationshipTypeStore } from '@/stores/relationshipTypes'

interface RelationshipType {
  id: string
  name: string
  forward_label: string
  reverse_label: string
  color?: string
  icon?: string
  is_active: boolean
  is_system: boolean
}

interface RelationshipTypeBadgeProps {
  type?: string | RelationshipType
  label?: string
  showReverse?: boolean
  size?: 'small' | 'medium'
  showIcon?: boolean
  customColor?: string
}

const props = withDefaults(defineProps<RelationshipTypeBadgeProps>(), {
  type: '',
  label: '',
  showReverse: false,
  size: 'medium',
  showIcon: true,
  customColor: ''
})

const relationshipTypeStore = useRelationshipTypeStore()

// Get relationship type details
const relationshipType = computed(() => {
  if (!props.type) return null

  // If it's already a RelationshipType object, use it directly
  if (typeof props.type === 'object') {
    return props.type
  }

  // Otherwise, look it up by name
  return relationshipTypeStore.getRelationshipTypeByName(props.type)
})

// Get display text
const displayText = computed(() => {
  if (props.label) return props.label

  if (relationshipType.value) {
    return props.showReverse ? relationshipType.value.reverse_label : relationshipType.value.forward_label
  }

  // Fallback: format the type name
  if (typeof props.type === 'string' && props.type) {
    return props.type.split('_').map(word =>
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ')
  }

  return 'Unknown Type'
})

// Get badge style
const badgeStyle = computed(() => {
  if (props.customColor) {
    return {
      backgroundColor: props.customColor + '20',
      color: props.customColor,
      borderColor: props.customColor
    }
  }

  if (relationshipType.value?.color) {
    const color = relationshipType.value.color
    return {
      backgroundColor: color + '20',
      color: color,
      borderColor: color
    }
  }

  // Dynamic color assignment based on type name hash for consistent colors
  if (typeof props.type === 'string' && props.type) {
    const color = generateColorFromName(props.type)
    return {
      backgroundColor: color + '20',
      color: color,
      borderColor: color
    }
  }

  // Default fallback color
  const defaultColor = '#6B7280' // gray
  return {
    backgroundColor: defaultColor + '20',
    color: defaultColor,
    borderColor: defaultColor
  }
})

// Generate consistent color based on type name hash
const generateColorFromName = (name: string) => {
  // Define a palette of visually distinct colors
  const colorPalette = [
    '#3B82F6', // blue
    '#10B981', // green
    '#F59E0B', // amber
    '#EF4444', // red
    '#8B5CF6', // purple
    '#06B6D4', // cyan
    '#84CC16', // lime
    '#F97316', // orange
    '#EC4899', // pink
    '#14B8A6', // teal
    '#6366F1', // indigo
    '#A855F7', // violet
    '#FACC15', // yellow
    '#22C55E', // emerald
    '#0EA5E9', // sky
    '#D946EF', // fuchsia
    '#F43F5E', // rose
    '#059669', // emerald-dark
    '#7C3AED', // violet-dark
    '#DC2626', // red-dark
  ]

  // Simple hash function to generate consistent index
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    const char = name.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32-bit integer
  }

  const index = Math.abs(hash) % colorPalette.length
  return colorPalette[index]
}
</script>
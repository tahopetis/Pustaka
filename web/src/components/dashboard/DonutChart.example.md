# DonutChart Component - Usage Examples

## Overview

The `DonutChart` component is a reusable D3.js-based donut chart (pie chart with inner radius) designed for displaying distribution data with interactive features, accessible design, and responsive layout.

## Features

- D3.js donut chart with customizable inner radius
- Color-coded segments with distinct colors
- Center label showing total count
- Interactive legend with percentages
- Hover effects with tooltips
- Keyboard accessible (WCAG 2.1 AA compliant)
- Responsive SVG with viewBox
- Empty state handling
- TypeScript support

## Basic Usage

```vue
<template>
  <DonutChart
    :data="ciTypeDistribution"
    title="CI Distribution by Type"
    :size="300"
  />
</template>

<script setup lang="ts">
import DonutChart from '@/components/dashboard/DonutChart.vue'
import type { DonutChartData } from '@/types/dashboard'

const ciTypeDistribution: DonutChartData[] = [
  { label: 'Servers', value: 45, color: '#3B82F6' },
  { label: 'Databases', value: 30, color: '#10B981' },
  { label: 'Applications', value: 25, color: '#F59E0B' },
]
</script>
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `data` | `DonutChartData[]` | **Required** | Array of segments to display |
| `title` | `string` | `undefined` | Optional chart title |
| `size` | `number` | `300` | Size of component in pixels |
| `innerRadiusRatio` | `number` | `0.6` | Ratio of inner to outer radius (0-1) |

## Data Structure

```typescript
interface DonutChartData {
  label: string      // Segment label
  value: number      // Numeric value
  color?: string     // Hex color (optional, auto-generated if not provided)
  percentage?: number // Calculated automatically
}
```

## Examples

### 1. CI Type Distribution

```vue
<template>
  <DonutChart
    :data="ciTypeUsage"
    title="Configuration Items by Type"
    :size="350"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import { useDashboardData } from '@/composables/useDashboardData'
import type { DonutChartData } from '@/types/dashboard'

const { ciTypeUsage } = useDashboardData()

// Color palette for CI types
const ciTypeColors: Record<string, string> = {
  'Server': '#3B82F6',      // blue
  'Database': '#10B981',    // emerald
  'Application': '#F59E0B', // amber
  'Network': '#EF4444',     // red
  'Storage': '#8B5CF6',     // violet
}

const ciTypeDistribution = computed<DonutChartData[]>(() => {
  return ciTypeUsage.value.map(item => ({
    label: item.type,
    value: item.count,
    color: ciTypeColors[item.type] || '#6B7280', // gray fallback
  }))
})
</script>
```

### 2. Relationship Type Distribution

```vue
<template>
  <DonutChart
    :data="relationshipDistribution"
    title="Relationships by Type"
    :size="300"
    :innerRadiusRatio="0.5"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import { useDashboardData } from '@/composables/useDashboardData'
import type { DonutChartData } from '@/types/dashboard'

const { relationshipTypeUsage } = useDashboardData()

const relationshipDistribution = computed<DonutChartData[]>(() => {
  return relationshipTypeUsage.value.map(item => ({
    label: item.relationship_type.display_name || item.relationship_type.name,
    value: item.usage_count,
    color: item.relationship_type.color || undefined, // Use type's color if available
  }))
})
</script>
```

### 3. Custom Colors (Auto-Generated)

```vue
<template>
  <DonutChart
    :data="dataWithoutColors"
    title="Auto-Colored Chart"
  />
</template>

<script setup lang="ts">
import DonutChart from '@/components/dashboard/DonutChart.vue'

// Colors will be automatically assigned from the default palette
const dataWithoutColors = [
  { label: 'Category A', value: 100 },
  { label: 'Category B', value: 75 },
  { label: 'Category C', value: 50 },
]
</script>
```

### 4. Small Donut (Thin Ring)

```vue
<template>
  <DonutChart
    :data="statusData"
    title="Status Distribution"
    :size="250"
    :innerRadiusRatio="0.75"
  />
</template>

<script setup lang="ts">
const statusData = [
  { label: 'Active', value: 85, color: '#10B981' },
  { label: 'Inactive', value: 10, color: '#EF4444' },
  { label: 'Pending', value: 5, color: '#F59E0B' },
]
</script>
```

### 5. Large Chart with Many Segments

```vue
<template>
  <DonutChart
    :data="detailedData"
    title="Detailed Breakdown"
    :size="450"
  />
</template>

<script setup lang="ts">
const detailedData = [
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
</script>
```

### 6. Integration with DashboardWidget

```vue
<template>
  <DashboardWidget
    title="CI Type Distribution"
    :loading="loading.ciTypes"
    :error="errors.ciTypes"
    @retry="retryDataSource('ciTypes')"
  >
    <DonutChart
      :data="ciTypeDistribution"
      :size="320"
    />
  </DashboardWidget>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import { useDashboardData } from '@/composables/useDashboardData'
import type { DonutChartData } from '@/types/dashboard'

const { ciTypeUsage, loading, errors, retryDataSource } = useDashboardData()

const ciTypeDistribution = computed<DonutChartData[]>(() => {
  return ciTypeUsage.value.map(item => ({
    label: item.type,
    value: item.count,
  }))
})
</script>
```

### 7. Dynamic Data Updates

```vue
<template>
  <div>
    <DonutChart
      :data="currentData"
      :title="currentTitle"
      :size="300"
    />

    <div class="mt-4 flex gap-2">
      <button @click="showCITypes" class="btn-primary">
        CI Types
      </button>
      <button @click="showRelationships" class="btn-primary">
        Relationships
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import { useDashboardData } from '@/composables/useDashboardData'
import type { DonutChartData } from '@/types/dashboard'

const { ciTypeUsage, relationshipTypeUsage } = useDashboardData()
const viewMode = ref<'ci' | 'relationship'>('ci')

const currentData = computed<DonutChartData[]>(() => {
  if (viewMode.value === 'ci') {
    return ciTypeUsage.value.map(item => ({
      label: item.type,
      value: item.count,
    }))
  } else {
    return relationshipTypeUsage.value.map(item => ({
      label: item.relationship_type.display_name || item.relationship_type.name,
      value: item.usage_count,
      color: item.relationship_type.color,
    }))
  }
})

const currentTitle = computed(() => {
  return viewMode.value === 'ci'
    ? 'CI Type Distribution'
    : 'Relationship Type Distribution'
})

function showCITypes() {
  viewMode.value = 'ci'
}

function showRelationships() {
  viewMode.value = 'relationship'
}
</script>
```

## Styling and Customization

### Custom Colors

The component uses a default color palette from D3.js, but you can provide custom colors:

```typescript
const customColorData: DonutChartData[] = [
  { label: 'Critical', value: 10, color: '#DC2626' },  // red-600
  { label: 'Warning', value: 25, color: '#F59E0B' },   // amber-500
  { label: 'Normal', value: 65, color: '#10B981' },    // emerald-500
]
```

### Inner Radius Customization

- `innerRadiusRatio: 0` - Full pie chart (no hole)
- `innerRadiusRatio: 0.5` - Medium donut
- `innerRadiusRatio: 0.6` - Default donut
- `innerRadiusRatio: 0.8` - Thin ring

## Accessibility Features

The component is fully accessible and follows WCAG 2.1 AA guidelines:

1. **ARIA Labels**: SVG has descriptive `aria-label` attribute
2. **Keyboard Navigation**: All segments and legend items are keyboard accessible
3. **Focus Indicators**: Clear focus styles for keyboard users
4. **Screen Reader Support**: Descriptive labels for all interactive elements
5. **Semantic HTML**: Proper role attributes (`role="img"`, `role="button"`)

### Keyboard Navigation

- `Tab`: Navigate between segments and legend items
- `Enter/Space`: Activate segment or legend item (shows tooltip)
- `Escape`: Close tooltip and remove focus

## Empty State

The component automatically displays an empty state when:
- `data` array is empty
- All values are zero

```vue
<DonutChart :data="[]" title="No Data" />
```

## Performance Considerations

1. **Reactive Updates**: Chart automatically re-renders when data changes
2. **Computed Properties**: Expensive calculations are memoized
3. **D3.js Optimization**: Uses D3's efficient arc generators
4. **Minimal Re-renders**: Only updates when necessary

## Best Practices

1. **Data Formatting**: Ensure data is sorted or ordered as needed before passing to component
2. **Color Consistency**: Use consistent colors across your dashboard
3. **Label Length**: Keep labels concise for better legend readability
4. **Value Ranges**: Works best with 3-10 segments; more may be hard to read
5. **Responsive Design**: Consider wrapping in a responsive container

## Common Use Cases

### Dashboard Statistics

```vue
<DonutChart
  :data="dashboardStats"
  title="System Overview"
  :size="300"
/>
```

### Resource Allocation

```vue
<DonutChart
  :data="resourceUsage"
  title="Resource Allocation"
  :size="350"
  :innerRadiusRatio="0.5"
/>
```

### Compliance Status

```vue
<DonutChart
  :data="complianceData"
  title="Compliance Status"
  :size="280"
/>
```

## Troubleshooting

### Chart Not Rendering

- Ensure `data` prop has at least one item with `value > 0`
- Check that D3.js is properly imported
- Verify TypeScript types match `DonutChartData` interface

### Colors Not Showing

- Provide valid hex color codes (e.g., `#3B82F6`)
- Fallback colors are automatically applied if not provided

### Legend Overflow

- Use shorter labels or increase component `size`
- Consider showing only top N items for large datasets

## Related Components

- `DashboardWidget.vue` - Wrapper with loading/error states
- `TrendChart.vue` - Time-series line chart
- `NetworkAnalyticsCard.vue` - Horizontal bar chart
- `ActivityHeatmap.vue` - Calendar-style heatmap

## API Reference

See `/web/src/types/dashboard.ts` for complete TypeScript definitions.

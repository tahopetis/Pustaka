# TrendChart Component - Usage Examples

## Basic Usage

```vue
<template>
  <TrendChart
    :data="chartData"
    title="Daily Activity Trend"
    :height="300"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import TrendChart from '@/components/dashboard/TrendChart.vue'
import type { ChartSeries } from '@/types/dashboard'
import { useDashboardData } from '@/composables/useDashboardData'

const { auditStats } = useDashboardData()

// Convert audit daily_activity to ChartSeries format
const chartData = computed<ChartSeries[]>(() => {
  if (!auditStats.value?.daily_activity) return []

  const dataPoints = Object.entries(auditStats.value.daily_activity)
    .map(([date, count]) => ({
      x: date,  // ISO date string (YYYY-MM-DD)
      y: count, // Event count
    }))
    .sort((a, b) => a.x.localeCompare(b.x))

  return [
    {
      id: 'activity',
      name: 'Daily Events',
      color: '#3B82F6', // Blue
      data: dataPoints,
    },
  ]
})
</script>
```

## Multi-Line Chart with Legend

```vue
<template>
  <TrendChart
    :data="multiSeriesData"
    title="CI and Relationship Growth"
    :height="400"
    :show-legend="true"
    :show-grid="true"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import TrendChart from '@/components/dashboard/TrendChart.vue'
import type { ChartSeries } from '@/types/dashboard'

// Example with multiple series
const multiSeriesData = computed<ChartSeries[]>(() => {
  return [
    {
      id: 'ci-growth',
      name: 'Configuration Items',
      color: '#10B981', // Green
      data: [
        { x: '2025-01-01', y: 10 },
        { x: '2025-01-02', y: 15 },
        { x: '2025-01-03', y: 22 },
        { x: '2025-01-04', y: 28 },
        { x: '2025-01-05', y: 35 },
      ],
    },
    {
      id: 'rel-growth',
      name: 'Relationships',
      color: '#8B5CF6', // Purple
      data: [
        { x: '2025-01-01', y: 5 },
        { x: '2025-01-02', y: 12 },
        { x: '2025-01-03', y: 18 },
        { x: '2025-01-04', y: 25 },
        { x: '2025-01-05', y: 30 },
      ],
    },
  ]
})
</script>
```

## Integration with DashboardWidget

```vue
<template>
  <DashboardWidget
    title="Activity Trend"
    :loading="loading.audit"
    :error="errors.audit"
    :empty="!hasData"
    empty-message="No activity data available"
    @retry="retryDataSource('audit')"
  >
    <TrendChart
      :data="activityChartData"
      title="Daily Activity"
      :height="300"
      :show-legend="false"
    />
  </DashboardWidget>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'
import TrendChart from '@/components/dashboard/TrendChart.vue'
import { useDashboardData } from '@/composables/useDashboardData'
import type { ChartSeries } from '@/types/dashboard'

const { auditStats, loading, errors, retryDataSource } = useDashboardData()

const hasData = computed(() => {
  return auditStats.value?.daily_activity && 
         Object.keys(auditStats.value.daily_activity).length > 0
})

const activityChartData = computed<ChartSeries[]>(() => {
  if (!auditStats.value?.daily_activity) return []

  const dataPoints = Object.entries(auditStats.value.daily_activity)
    .map(([date, count]) => ({
      x: date,
      y: count,
    }))
    .sort((a, b) => a.x.localeCompare(b.x))

  return [
    {
      id: 'daily-activity',
      name: 'Events',
      color: '#3B82F6',
      data: dataPoints,
    },
  ]
})
</script>
```

## Props Reference

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `data` | `ChartSeries[]` | **Required** | Array of data series to plot |
| `title` | `string` | `'Trend Chart'` | Chart title for accessibility |
| `height` | `number` | `300` | Chart height in pixels |
| `xAxisLabel` | `string` | `''` | X-axis label (optional) |
| `yAxisLabel` | `string` | `''` | Y-axis label (optional) |
| `showLegend` | `boolean` | `true` | Show legend for multi-series charts |
| `showGrid` | `boolean` | `true` | Show horizontal grid lines |
| `animationDuration` | `number` | `750` | Animation duration in milliseconds |

## Data Structure

```typescript
interface ChartSeries {
  /** Unique identifier for the series */
  id: string
  /** Display name shown in legend and tooltip */
  name: string
  /** Hex color code (e.g., '#3B82F6') */
  color: string
  /** Array of data points */
  data: ChartDataPoint[]
}

interface ChartDataPoint {
  /** X-axis value (date string in ISO format or number) */
  x: string | number
  /** Y-axis value (numeric) */
  y: number
  /** Optional label for custom tooltip text */
  label?: string
  /** Optional hex color override for this point */
  color?: string
  /** Optional metadata for custom interactions */
  metadata?: Record<string, any>
}
```

## Features

### Responsive Design
- Automatically adjusts to container width
- Uses SVG viewBox for perfect scaling
- ResizeObserver for dynamic resizing

### Interactive Tooltips
- Hover over the chart to see exact values
- Shows all series data at the hovered date
- Smart tooltip positioning

### Smooth Animations
- Transitions between data updates
- Respects prefers-reduced-motion for accessibility

### Accessibility
- Proper ARIA labels and roles
- Semantic SVG structure
- Keyboard-friendly (no interactive elements to tab through)

### Multi-Series Support
- Render multiple trend lines with different colors
- Automatic legend generation
- Independent data series

### Gradient Fills
- Beautiful gradient areas under each line
- Color-matched to series color
- Subtle visual enhancement

## Color Palette Suggestions

```typescript
// Tailwind CSS color equivalents
const colors = {
  blue: '#3B82F6',
  green: '#10B981',
  purple: '#8B5CF6',
  pink: '#EC4899',
  yellow: '#F59E0B',
  red: '#EF4444',
  indigo: '#6366F1',
  teal: '#14B8A6',
}
```

## Performance Considerations

- Uses D3.js enter/update/exit pattern for efficient re-rendering
- Debounced resize handling
- Memoized computed properties
- Virtual DOM optimizations with proper keys

## Browser Compatibility

Requires modern browser with:
- ES6+ JavaScript
- SVG support
- ResizeObserver API
- D3.js v7 compatibility

## Known Limitations

- Maximum recommended data points per series: ~365 (one year of daily data)
- For larger datasets, consider aggregation or sampling
- Tooltip positioning may need adjustment for small screens

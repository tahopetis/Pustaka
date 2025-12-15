# DonutChart Component

A fully-featured D3.js donut chart component for the Pustaka CMDB dashboard, designed for displaying distribution data with interactive features and accessible design.

## Overview

The DonutChart component is part of Phase 2 of the Dashboard Enhancement Plan and provides a professional, interactive visualization for displaying categorical data distributions such as CI types and relationship types.

## Features

### Core Functionality
- **D3.js Donut Chart**: Pie chart with customizable inner radius creating a donut shape
- **Color-Coded Segments**: Distinct, accessible colors for each category
- **Center Label**: Total count displayed in the center of the donut
- **Interactive Legend**: Shows labels, values, and percentages
- **Responsive Design**: SVG with viewBox for automatic scaling
- **Empty State Handling**: Graceful display when no data is available

### Interactive Features
- **Hover Effects**: Segments highlight and scale on hover
- **Tooltips**: Contextual information appears on hover
- **Synchronized Interactions**: Hovering over legend items highlights corresponding segments
- **Keyboard Navigation**: Full keyboard accessibility for all interactive elements

### Accessibility (WCAG 2.1 AA Compliant)
- **ARIA Labels**: Descriptive labels on all chart elements
- **Keyboard Support**: Tab navigation, Enter/Space activation
- **Focus Indicators**: Clear visual focus states
- **Screen Reader Support**: Semantic HTML with proper roles
- **Color Contrast**: Accessible color palette

## Installation

The component is located at:
```
/home/tahopetis/dev/pustaka/web/src/components/dashboard/DonutChart.vue
```

## Usage

### Basic Example

```vue
<template>
  <DonutChart
    :data="chartData"
    title="CI Distribution"
    :size="300"
  />
</template>

<script setup lang="ts">
import DonutChart from '@/components/dashboard/DonutChart.vue'
import type { DonutChartData } from '@/types/dashboard'

const chartData: DonutChartData[] = [
  { label: 'Servers', value: 45, color: '#3B82F6' },
  { label: 'Databases', value: 30, color: '#10B981' },
  { label: 'Applications', value: 25, color: '#F59E0B' },
]
</script>
```

### Integration with Dashboard Data

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
    // Color will be auto-generated if not provided
  }))
})
</script>
```

## Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `data` | `DonutChartData[]` | Yes | - | Array of segments to display in the chart |
| `title` | `string` | No | `undefined` | Optional title displayed above the chart |
| `size` | `number` | No | `300` | Overall component size in pixels (width and height) |
| `innerRadiusRatio` | `number` | No | `0.6` | Ratio of inner radius to outer radius (0-1), where 0.6 creates a standard donut |

## Data Structure

```typescript
interface DonutChartData {
  label: string       // Category or segment label
  value: number       // Numeric value for the segment
  color?: string      // Optional hex color code (e.g., '#3B82F6')
  percentage?: number // Automatically calculated by the component
}
```

## Styling

### Default Color Palette

If colors are not provided, the component uses a professional D3.js color scale:

1. `#3B82F6` - Blue (Tailwind blue-500)
2. `#10B981` - Emerald (Tailwind emerald-500)
3. `#F59E0B` - Amber (Tailwind amber-500)
4. `#EF4444` - Red (Tailwind red-500)
5. `#8B5CF6` - Violet (Tailwind violet-500)
6. `#EC4899` - Pink (Tailwind pink-500)
7. `#14B8A6` - Teal (Tailwind teal-500)
8. `#F97316` - Orange (Tailwind orange-500)
9. `#6366F1` - Indigo (Tailwind indigo-500)
10. `#84CC16` - Lime (Tailwind lime-500)

### Inner Radius Options

- `0.0` - Full pie chart (no hole)
- `0.3` - Small hole
- `0.5` - Medium donut
- `0.6` - Standard donut (default)
- `0.75` - Thin ring
- `0.9` - Very thin ring

## Component Structure

### Template Layout

```
┌─────────────────────────────────┐
│ Title (optional)                │
├──────────────┬──────────────────┤
│              │ Legend:          │
│   SVG Chart  │ ■ Label  50 (25%)│
│              │ ■ Label  75 (37%)│
│              │ ■ Label  75 (38%)│
└──────────────┴──────────────────┘
```

### SVG Structure

```
<svg>
  <g transform="translate(centerX, centerY)">
    <g class="segments">
      <path /> <!-- Arc segment -->
      <path /> <!-- Arc segment -->
      ...
    </g>
    <g class="center-label">
      <text>Total</text>
      <text>100</text>
    </g>
  </g>
</svg>
```

## Interactivity

### Mouse Events

- **Segment Hover**: Scales segment to 105%, shows tooltip
- **Legend Item Hover**: Highlights corresponding segment, shows tooltip
- **Mouse Leave**: Resets all states

### Keyboard Events

- **Tab**: Navigate between segments and legend items
- **Shift+Tab**: Navigate backwards
- **Enter/Space**: Activate item (show tooltip)
- **Escape**: Close tooltip and remove focus

### Tooltip

The tooltip displays:
- Segment label
- Numeric value
- Percentage of total

Position: Above cursor, centered horizontally

## Empty State

When there's no data or all values are zero, the component displays:
- Chart icon (pie chart SVG)
- "No data available" message
- Gray color scheme

## Data Flow

```
Props (data) → Computed (segments) → D3 Generators → SVG Paths → DOM
                    ↓
              Computed (legendItems) → Template → Legend
```

## Performance Optimization

1. **Computed Properties**: All expensive calculations are memoized
2. **Reactive Updates**: Only re-renders when data actually changes
3. **D3 Efficiency**: Uses D3's optimized arc generators
4. **CSS Transitions**: Smooth animations via CSS instead of JavaScript
5. **Event Debouncing**: Prevents excessive re-renders on rapid interactions

## Browser Compatibility

- **Modern Browsers**: Chrome, Firefox, Safari, Edge (latest 2 versions)
- **D3.js v7**: Requires ES6+ support (no IE11)
- **SVG Support**: Required for chart rendering

## Accessibility Details

### ARIA Attributes

- `role="img"` on SVG
- `aria-label` describing the entire chart
- `role="button"` on interactive segments
- `aria-label` on each segment with value and percentage
- `tabindex="0"` for keyboard accessibility

### Focus Management

- Clear focus indicators with blue outline
- Focus order: segments → legend items
- Escape key to clear focus and close tooltip

### Screen Reader Announcements

Example: "CI Distribution showing 3 categories with total of 100"
Segment: "Servers: 45 (45%)"

## Testing

### Unit Tests

Location: `/home/tahopetis/dev/pustaka/web/src/components/dashboard/DonutChart.test.ts`

Test coverage includes:
- Basic rendering
- Data processing and percentage calculation
- Segment and legend rendering
- Interactive hover effects
- Accessibility features
- Empty state handling
- Edge cases (single segment, many segments, small values)

Run tests:
```bash
cd web/
npm run test DonutChart.test.ts
```

### Manual Testing Checklist

- [ ] Chart renders with sample data
- [ ] Hover effects work on segments
- [ ] Hover effects work on legend items
- [ ] Tooltip appears and disappears correctly
- [ ] Keyboard navigation works (Tab, Enter, Escape)
- [ ] Empty state displays when no data
- [ ] Colors are applied correctly
- [ ] Center total is calculated correctly
- [ ] Percentages add up to 100%
- [ ] Responsive at different sizes
- [ ] Screen reader announces chart correctly

## Integration Points

### Dashboard View

Used in `DashboardView.vue` for:
- CI type distribution
- Relationship type distribution

### Data Sources

Gets data from `useDashboardData` composable:
- `ciTypeUsage` → DonutChartData[]
- `relationshipTypeUsage` → DonutChartData[]

### Related Components

- **DashboardWidget.vue**: Wrapper providing loading/error states
- **TimeRangeFilter.vue**: Time filtering (affects data but not chart directly)
- **TrendChart.vue**: Companion chart for time-series data

## Common Use Cases

### 1. CI Type Distribution

```typescript
const ciTypeData = ciTypeUsage.value.map(item => ({
  label: item.type,
  value: item.count,
}))
```

### 2. Relationship Type Distribution

```typescript
const relTypeData = relationshipTypeUsage.value.map(item => ({
  label: item.relationship_type.display_name,
  value: item.usage_count,
  color: item.relationship_type.color,
}))
```

### 3. Status Distribution

```typescript
const statusData = [
  { label: 'Active', value: 85, color: '#10B981' },
  { label: 'Inactive', value: 10, color: '#EF4444' },
  { label: 'Pending', value: 5, color: '#F59E0B' },
]
```

## Troubleshooting

### Issue: Chart not rendering

**Possible Causes:**
- Empty data array
- All values are zero
- D3.js not imported correctly

**Solution:**
- Verify data has items with value > 0
- Check browser console for errors
- Ensure D3.js is in package.json dependencies

### Issue: Colors not showing

**Possible Causes:**
- Invalid color format
- Missing color property

**Solution:**
- Use hex format: `#3B82F6`
- Colors are optional; component auto-generates if missing

### Issue: Legend overflow

**Possible Causes:**
- Too many segments
- Very long labels
- Small component size

**Solution:**
- Increase `size` prop
- Shorten labels
- Consider grouping smaller segments into "Other"

### Issue: Hover effects not working

**Possible Causes:**
- CSS conflicts
- Z-index issues
- Event listeners not attached

**Solution:**
- Check for CSS conflicts in parent components
- Ensure tooltip has high z-index
- Verify mouseenter/mouseleave events in DevTools

## Future Enhancements

Potential improvements for future versions:

1. **Animation**: Smooth arc transitions when data changes
2. **Custom Tooltip**: Slot-based custom tooltip content
3. **Drill-down**: Click events for navigation
4. **Export**: SVG download functionality
5. **Themes**: Dark mode support
6. **Responsive Legend**: Collapsible legend for small screens
7. **Data Labels**: Optional labels on segments
8. **Gradient Fill**: Gradient colors for segments

## References

- [D3.js Documentation](https://d3js.org/)
- [D3 Arc Generator](https://github.com/d3/d3-shape#arcs)
- [D3 Pie Layout](https://github.com/d3/d3-shape#pies)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Dashboard Enhancement Plan](~/.claude/plans/cheeky-weaving-ocean.md)

## Files

- **Component**: `/home/tahopetis/dev/pustaka/web/src/components/dashboard/DonutChart.vue`
- **Tests**: `/home/tahopetis/dev/pustaka/web/src/components/dashboard/DonutChart.test.ts`
- **Examples**: `/home/tahopetis/dev/pustaka/web/src/components/dashboard/DonutChart.example.md`
- **Types**: `/home/tahopetis/dev/pustaka/web/src/types/dashboard.ts`

## Support

For questions or issues, refer to:
- Component code comments
- Test file for usage examples
- Dashboard types file for data structures
- Project CLAUDE.md for development guidelines

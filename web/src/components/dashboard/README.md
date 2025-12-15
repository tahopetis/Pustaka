# Dashboard Components

This directory contains reusable components for the Pustaka CMDB dashboard, part of the dashboard enhancement plan documented in `~/.claude/plans/cheeky-weaving-ocean.md`.

## Components

### DashboardWidget.vue

A consistent wrapper component for all dashboard widgets providing:
- Loading states with skeleton animations
- Error states with retry functionality
- Empty states with customizable messages
- Header with title and optional actions
- Full accessibility support (ARIA attributes, keyboard navigation)

**Status**: ✅ Complete
**Documentation**: [DashboardWidget.example.md](./DashboardWidget.example.md)
**Tests**: [DashboardWidget.test.md](./DashboardWidget.test.md)

### TimeRangeFilter.vue

Time range selection component with:
- Quick preset buttons (7, 30, 90 days)
- Custom date range picker
- Date validation
- Apply/Reset functionality

**Status**: ✅ Complete

### DonutChart.vue

D3.js-based donut chart component for displaying distribution data:
- Interactive donut/pie chart with customizable inner radius
- Color-coded segments with distinct, accessible colors
- Center label showing total count
- Interactive legend with percentages
- Hover effects and tooltips
- Full keyboard accessibility (WCAG 2.1 AA compliant)
- Responsive SVG design
- Empty state handling

**Status**: ✅ Complete
**Documentation**: [DonutChart.README.md](./DonutChart.README.md), [DonutChart.example.md](./DonutChart.example.md)
**Tests**: [DonutChart.test.ts](./DonutChart.test.ts)
**Demo**: [DonutChartDemo.vue](./DonutChartDemo.vue)

## Implementation Plan Status

Based on `~/.claude/plans/cheeky-weaving-ocean.md`:

### Phase 1: Foundation Components ✅ Complete
- [x] TimeRangeFilter.vue
- [x] DashboardWidget.vue
- [x] useDashboardData.ts composable

### Phase 2: Chart Components (In Progress)
- [ ] TrendChart.vue
- [x] **DonutChart.vue** ✅ Complete
- [ ] NetworkAnalyticsCard.vue
- [ ] ActivityHeatmap.vue

### Phase 3: Dashboard Layout Integration (TODO)
- [ ] Enhanced DashboardView.vue

### Phase 4: Backend Enhancement (TODO)
- [ ] New analytics endpoints

## Usage Examples

### DashboardWidget

```vue
<template>
  <DashboardWidget
    title="Total CIs"
    :loading="loading"
    :error="error"
    :empty="isEmpty"
    @retry="fetchData"
  >
    <template #actions>
      <button @click="refresh">Refresh</button>
    </template>

    <div class="text-3xl font-bold">{{ data.total }}</div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'
// ... component logic
</script>
```

### DonutChart

```vue
<template>
  <DonutChart
    :data="ciTypeDistribution"
    title="CI Distribution by Type"
    :size="300"
    :innerRadiusRatio="0.6"
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

### Integration Example: DonutChart + DashboardWidget

```vue
<template>
  <DashboardWidget
    title="CI Type Distribution"
    :loading="loading.ciTypes"
    :error="errors.ciTypes"
    @retry="retryDataSource('ciTypes')"
  >
    <DonutChart
      :data="ciTypeData"
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

const ciTypeData = computed<DonutChartData[]>(() => {
  return ciTypeUsage.value.map(item => ({
    label: item.type,
    value: item.count,
  }))
})
</script>
```

## File Structure

```
dashboard/
├── README.md                          # This file
├── DashboardWidget.vue                # Reusable widget wrapper
├── DashboardWidget.example.md         # Widget usage examples
├── DashboardWidget.test.md            # Widget test cases
├── TimeRangeFilter.vue                # Time range selector
├── DonutChart.vue                     # D3.js donut chart component
├── DonutChart.README.md               # DonutChart comprehensive docs
├── DonutChart.example.md              # DonutChart usage examples
├── DonutChart.test.ts                 # DonutChart unit tests
├── DonutChartDemo.vue                 # DonutChart interactive demo
└── [future components...]
```

## Design System

All components follow the Pustaka design system:
- **Colors**: Tailwind CSS utility classes
- **Spacing**: Consistent px-4 py-5 (mobile) and sm:px-6 (desktop)
- **Shadows**: `shadow` class for cards
- **Borders**: `border-gray-200` for separators
- **Focus States**: Blue ring with `focus:ring-2 focus:ring-blue-500`
- **Hover States**: Subtle background changes
- **Transitions**: Smooth `transition-colors`

### DonutChart Color Palette

Default colors when not specified:
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

## Accessibility

All components meet WCAG 2.1 AA standards:
- Semantic HTML structure
- ARIA attributes for dynamic content
- Keyboard navigation support (Tab, Enter, Escape)
- Screen reader announcements
- Sufficient color contrast
- Focus indicators
- Interactive elements properly labeled

### DonutChart Accessibility Features
- SVG with `role="img"` and descriptive `aria-label`
- All segments keyboard accessible with `tabindex="0"`
- Segments have `role="button"` and descriptive `aria-label`
- Legend items fully keyboard navigable
- Focus visible styles with blue outline
- Tooltip appears on both mouse and keyboard interaction

## Testing

Components should be tested at multiple levels:
1. **Unit tests** (Vitest) - Component logic and rendering
2. **Integration tests** - Component interactions
3. **E2E tests** (Playwright) - User workflows
4. **Visual regression** - Screenshot comparisons
5. **Accessibility** - Automated ARIA/keyboard checks

### DonutChart Test Coverage

The DonutChart includes comprehensive test coverage:
- ✅ Basic rendering and SVG generation
- ✅ Data processing and percentage calculations
- ✅ Segment and legend rendering
- ✅ Interactive hover effects
- ✅ Tooltip display and positioning
- ✅ Accessibility features (ARIA, keyboard navigation)
- ✅ Empty state handling
- ✅ Color customization
- ✅ Edge cases (single segment, many segments, small values)

Run tests:
```bash
cd web/
npm run test DonutChart.test.ts
```

## Component API Reference

### DonutChart Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `data` | `DonutChartData[]` | Yes | - | Array of segments |
| `title` | `string` | No | `undefined` | Chart title |
| `size` | `number` | No | `300` | Component size (px) |
| `innerRadiusRatio` | `number` | No | `0.6` | Inner/outer radius ratio |

### DonutChartData Interface

```typescript
interface DonutChartData {
  label: string       // Segment label
  value: number       // Numeric value
  color?: string      // Optional hex color
  percentage?: number // Auto-calculated
}
```

## Data Sources

The dashboard components integrate with the following data sources:

### useDashboardData Composable

Provides reactive data from backend APIs:
- `stats` - Basic dashboard statistics
- `auditStats` - Audit log statistics with time filtering
- `ciTypeUsage` - CI type distribution (for DonutChart)
- `mostConnected` - Top connected CIs
- `relationshipTypeUsage` - Relationship type distribution (for DonutChart)
- `loading` - Loading states per data source
- `errors` - Error states per data source

### DonutChart Data Mapping

**CI Type Distribution:**
```typescript
const ciTypeData = ciTypeUsage.value.map(item => ({
  label: item.type,
  value: item.count,
}))
```

**Relationship Type Distribution:**
```typescript
const relationshipData = relationshipTypeUsage.value.map(item => ({
  label: item.relationship_type.display_name || item.relationship_type.name,
  value: item.usage_count,
  color: item.relationship_type.color,
}))
```

## Development Guidelines

### Adding New Components

1. Create component file: `ComponentName.vue`
2. Add TypeScript types to `/web/src/types/dashboard.ts`
3. Write unit tests: `ComponentName.test.ts`
4. Create usage examples: `ComponentName.example.md`
5. Update this README with component details
6. Add to plan status in `~/.claude/plans/cheeky-weaving-ocean.md`

### Component Checklist

- [ ] Vue 3 Composition API with `<script setup lang="ts">`
- [ ] TypeScript strict mode
- [ ] Tailwind CSS for styling
- [ ] Responsive design (mobile-first)
- [ ] WCAG 2.1 AA accessibility
- [ ] Keyboard navigation support
- [ ] Loading/error/empty states
- [ ] Unit test coverage >85%
- [ ] Documentation with examples
- [ ] ARIA labels and semantic HTML

## Performance Considerations

### DonutChart Performance

1. **Computed Properties**: Expensive D3 calculations are memoized
2. **Reactive Updates**: Only re-renders when data actually changes
3. **D3 Optimization**: Uses D3's efficient arc generators
4. **CSS Transitions**: Smooth animations without JavaScript
5. **Minimal DOM**: SVG paths generated efficiently

### General Guidelines

- Use `computed` for derived state
- Lazy load heavy components
- Debounce rapid data updates
- Minimize watchers
- Use `v-once` for static content

## Browser Compatibility

- **Modern Browsers**: Chrome, Firefox, Safari, Edge (latest 2 versions)
- **D3.js v7**: Requires ES6+ support (no IE11)
- **SVG Support**: Required for all chart components
- **CSS Grid**: Used for responsive layouts

## Next Steps

1. ✅ ~~Create DonutChart.vue component~~ Complete
2. Create TrendChart.vue for time-series data
3. Create NetworkAnalyticsCard.vue for bar charts
4. Create ActivityHeatmap.vue for calendar heatmap
5. Integrate all components into DashboardView.vue
6. Add backend analytics endpoints
7. Write E2E tests for dashboard workflows

## Related Documentation

- **Dashboard Plan**: `~/.claude/plans/cheeky-weaving-ocean.md`
- **Project Docs**: `/home/tahopetis/dev/pustaka/CLAUDE.md`
- **Dashboard Types**: `/home/tahopetis/dev/pustaka/web/src/types/dashboard.ts`
- **Dashboard Composable**: `/home/tahopetis/dev/pustaka/web/src/composables/useDashboardData.ts`
- **Dashboard View**: `/home/tahopetis/dev/pustaka/web/src/views/DashboardView.vue`

## Support

For questions or issues:
- Check component documentation files
- Review test files for usage examples
- Refer to TypeScript type definitions
- See project CLAUDE.md for development guidelines

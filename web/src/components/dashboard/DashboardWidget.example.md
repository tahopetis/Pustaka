# DashboardWidget Component

A reusable wrapper component for dashboard widgets with built-in loading, error, and empty states.

## Features

- Consistent card wrapper with white background and shadow
- Header with title and optional action slot
- Loading skeleton state with animations
- Error state with retry functionality
- Empty state with customizable message
- Fully accessible with ARIA attributes
- TypeScript support

## Props

| Prop | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `title` | `string` | - | Yes | Title displayed in widget header |
| `loading` | `boolean` | `false` | No | Shows loading skeleton when true |
| `error` | `string \| null` | `null` | No | Error message to display in error state |
| `empty` | `boolean` | `false` | No | Shows empty state when true |
| `emptyMessage` | `string` | `"No data available"` | No | Message displayed in empty state |

## Events

| Event | Payload | Description |
|-------|---------|-------------|
| `@retry` | - | Emitted when user clicks retry button in error state |

## Slots

| Slot | Description |
|------|-------------|
| `default` | Main content area of the widget |
| `actions` | Optional actions displayed in the header (e.g., refresh button) |

## Basic Usage

```vue
<template>
  <DashboardWidget
    title="Total CIs"
    :loading="loading"
    :error="error"
    :empty="isEmpty"
    empty-message="No configuration items found"
    @retry="fetchData"
  >
    <!-- Your content here -->
    <div class="text-3xl font-bold">{{ totalCIs }}</div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'

const loading = ref(false)
const error = ref<string | null>(null)
const isEmpty = ref(false)
const totalCIs = ref(0)

const fetchData = async () => {
  loading.value = true
  error.value = null

  try {
    // Fetch your data
    const response = await api.get('/dashboard/stats')
    totalCIs.value = response.data.total_cis
    isEmpty.value = totalCIs.value === 0
  } catch (err) {
    error.value = 'Failed to load dashboard data'
  } finally {
    loading.value = false
  }
}
</script>
```

## With Actions Slot

```vue
<template>
  <DashboardWidget
    title="Recent Activity"
    :loading="loading"
    @retry="refresh"
  >
    <template #actions>
      <button
        @click="refresh"
        class="inline-flex items-center px-3 py-1 text-sm text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Refresh
      </button>
    </template>

    <!-- Widget content -->
    <ul class="space-y-2">
      <li v-for="item in activities" :key="item.id">
        {{ item.description }}
      </li>
    </ul>
  </DashboardWidget>
</template>
```

## States Examples

### Loading State

```vue
<DashboardWidget
  title="CI Distribution"
  :loading="true"
>
  <!-- Content won't be shown while loading -->
  <div>This won't be visible</div>
</DashboardWidget>
```

The loading state shows an animated skeleton with placeholder bars.

### Error State

```vue
<DashboardWidget
  title="Network Analytics"
  :error="'Failed to connect to server'"
  @retry="handleRetry"
>
  <div>Chart content</div>
</DashboardWidget>
```

The error state displays:
- Red alert icon
- "Failed to load data" heading
- Custom error message
- Retry button that emits `@retry` event

### Empty State

```vue
<DashboardWidget
  title="Most Connected CIs"
  :empty="true"
  empty-message="No relationships configured yet"
>
  <div>This won't be shown when empty</div>
</DashboardWidget>
```

The empty state displays:
- Gray inbox icon
- Custom empty message

## Complete Example with All States

```vue
<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <DashboardWidget
      title="Total Configuration Items"
      :loading="stats.loading"
      :error="stats.error"
      :empty="stats.data.total === 0"
      empty-message="No CIs have been created yet"
      @retry="loadStats"
    >
      <template #actions>
        <button
          @click="loadStats"
          :disabled="stats.loading"
          class="text-sm text-blue-600 hover:text-blue-800"
        >
          Refresh
        </button>
      </template>

      <div class="space-y-4">
        <div class="text-4xl font-bold text-gray-900">
          {{ stats.data.total }}
        </div>
        <div class="flex items-center text-sm">
          <svg class="w-5 h-5 text-green-500 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd"/>
          </svg>
          <span class="text-green-600 font-medium">12%</span>
          <span class="text-gray-500 ml-1">from last month</span>
        </div>
      </div>
    </DashboardWidget>

    <DashboardWidget
      title="CI Types Distribution"
      :loading="distribution.loading"
      :error="distribution.error"
      @retry="loadDistribution"
    >
      <!-- Chart component goes here -->
      <DonutChart :data="distribution.data" />
    </DashboardWidget>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import { api } from '@/services/api'

const stats = reactive({
  loading: false,
  error: null as string | null,
  data: { total: 0 }
})

const distribution = reactive({
  loading: false,
  error: null as string | null,
  data: []
})

const loadStats = async () => {
  stats.loading = true
  stats.error = null

  try {
    const response = await api.get('/dashboard/stats')
    stats.data = response.data
  } catch (err) {
    stats.error = err instanceof Error ? err.message : 'Failed to load stats'
  } finally {
    stats.loading = false
  }
}

const loadDistribution = async () => {
  distribution.loading = true
  distribution.error = null

  try {
    const response = await api.get('/analytics/ci-types/usage')
    distribution.data = response.data
  } catch (err) {
    distribution.error = err instanceof Error ? err.message : 'Failed to load distribution'
  } finally {
    distribution.loading = false
  }
}
</script>
```

## Accessibility Features

The component includes several accessibility features:

- `role="region"` with `aria-label` for screen readers
- `aria-busy` attribute indicates loading state
- `role="alert"` and `aria-live="polite"` for error messages
- `role="status"` for empty state announcements
- `aria-hidden="true"` on decorative SVG icons
- Semantic HTML structure
- Keyboard-accessible retry button

## Styling

The component uses Tailwind CSS classes and follows the project's design system:

- White background (`bg-white`)
- Shadow for depth (`shadow`)
- Rounded corners (`rounded-lg`)
- Consistent padding (`px-4 py-5 sm:px-6` for header, `px-4 py-5 sm:p-6` for content)
- Border separator between header and content
- Focus ring for interactive elements

## Best Practices

1. **Always handle the retry event**: Provide a way to recover from errors
   ```vue
   @retry="fetchData"
   ```

2. **Use reactive error state**: Let the component handle error display
   ```ts
   const error = ref<string | null>(null)
   ```

3. **Combine states appropriately**: Loading takes precedence over error and empty
   - Loading state → shows skeleton
   - Error state (if not loading) → shows error UI
   - Empty state (if not loading and no error) → shows empty UI
   - Content (if not loading, no error, not empty) → shows actual content

4. **Provide meaningful messages**: Customize empty messages for context
   ```vue
   empty-message="No relationships have been created yet"
   ```

5. **Use actions slot for operations**: Add refresh, export, or filter actions
   ```vue
   <template #actions>
     <button @click="refresh">Refresh</button>
   </template>
   ```

## Integration with Dashboard Plan

This component is part of Phase 1 (Foundation Components) in the dashboard enhancement plan. It provides:

- Consistent widget wrapper for all dashboard cards
- Built-in state management UI
- Accessibility compliance
- Reusable across all dashboard visualizations

Use this component to wrap:
- Stat cards
- Trend charts (TrendChart.vue)
- Distribution charts (DonutChart.vue)
- Network analytics cards
- Activity heatmaps
- Any custom dashboard widgets

# DashboardWidget Component Test Cases

## Manual Testing Checklist

### Visual Tests

#### 1. Normal State
- [ ] White background is displayed
- [ ] Shadow/elevation is visible
- [ ] Header has proper padding and border
- [ ] Title is displayed correctly
- [ ] Content area has proper padding

#### 2. Loading State
- [ ] Skeleton animation is visible
- [ ] Multiple gray bars with different widths
- [ ] Pulsing animation is smooth
- [ ] Content is hidden during loading
- [ ] `aria-busy="true"` is set

#### 3. Error State
- [ ] Red error icon is displayed
- [ ] "Failed to load data" heading shows
- [ ] Error message is displayed
- [ ] Retry button is visible and clickable
- [ ] Blue hover state on retry button
- [ ] Content is hidden when error is shown

#### 4. Empty State
- [ ] Gray inbox icon is displayed
- [ ] Custom empty message shows
- [ ] Message is centered
- [ ] Content is hidden when empty

#### 5. With Actions Slot
- [ ] Actions appear in header
- [ ] Actions are aligned to the right
- [ ] Multiple actions space correctly
- [ ] Actions don't overlap with title

### Functional Tests

#### State Transitions
```
Initial → Loading
Loading → Success (with data)
Loading → Error
Error → Retry → Loading
Success → Empty (when data becomes empty)
```

#### Event Handling
- [ ] `@retry` event emits when retry button clicked
- [ ] Event handler receives correct parameters
- [ ] Multiple clicks don't cause issues

#### Props Validation
- [ ] `title` prop is required
- [ ] `loading` defaults to false
- [ ] `error` defaults to null
- [ ] `empty` defaults to false
- [ ] `emptyMessage` has default value

### Accessibility Tests

#### ARIA Attributes
- [ ] `role="region"` is present
- [ ] `aria-label` matches title
- [ ] `aria-busy` changes with loading state
- [ ] Error has `role="alert"` and `aria-live="polite"`
- [ ] Empty state has `role="status"` and `aria-live="polite"`
- [ ] SVG icons have `aria-hidden="true"`

#### Keyboard Navigation
- [ ] Retry button is focusable with Tab
- [ ] Enter/Space triggers retry
- [ ] Focus ring is visible
- [ ] Actions in slot are keyboard accessible

#### Screen Reader
- [ ] Title is announced when region enters view
- [ ] Loading state announces "busy"
- [ ] Error message is announced
- [ ] Empty message is announced
- [ ] Retry button has clear label

### Responsive Tests

#### Desktop (≥1024px)
- [ ] Full padding in header and content
- [ ] Actions display inline
- [ ] Error/empty states are well-sized

#### Tablet (768px - 1023px)
- [ ] Header padding adjusts correctly
- [ ] Content remains readable
- [ ] Buttons remain clickable

#### Mobile (<768px)
- [ ] Smaller padding used
- [ ] Title doesn't wrap awkwardly
- [ ] Actions stack if needed
- [ ] Error/empty icons scale appropriately

### Integration Tests

#### With Real Data
```vue
<DashboardWidget
  title="API Test"
  :loading="loading"
  :error="error"
  :empty="!data.length"
  @retry="fetchData"
>
  <div v-for="item in data" :key="item.id">
    {{ item.name }}
  </div>
</DashboardWidget>
```

Test:
- [ ] Loading shows on mount
- [ ] Content appears after successful fetch
- [ ] Error shows on fetch failure
- [ ] Retry refetches data
- [ ] Empty shows when data array is empty

#### Multiple Widgets
```vue
<div class="grid grid-cols-2 gap-6">
  <DashboardWidget title="Widget 1" />
  <DashboardWidget title="Widget 2" />
</div>
```

Test:
- [ ] Widgets have consistent styling
- [ ] Grid layout works correctly
- [ ] Gap spacing is uniform
- [ ] Each widget is independent

#### Nested with Chart Components
```vue
<DashboardWidget title="Chart" :loading="loading">
  <DonutChart :data="chartData" />
</DashboardWidget>
```

Test:
- [ ] Chart renders inside widget
- [ ] Chart has proper spacing
- [ ] Loading works with chart
- [ ] Chart respects widget dimensions

### Edge Cases

#### Long Titles
```vue
<DashboardWidget
  title="This is a Very Long Title That Should Handle Properly Without Breaking the Layout"
/>
```
- [ ] Title wraps gracefully
- [ ] Doesn't overflow container
- [ ] Actions still visible

#### Long Error Messages
```vue
<DashboardWidget
  error="A very long error message that explains in great detail what went wrong and provides context about the failure"
/>
```
- [ ] Error message wraps
- [ ] Doesn't break layout
- [ ] Retry button remains accessible

#### Rapid State Changes
```ts
// Quickly toggle states
loading.value = true
setTimeout(() => { error.value = "Error" }, 100)
setTimeout(() => { error.value = null; loading.value = false }, 200)
```
- [ ] No flickering
- [ ] States transition smoothly
- [ ] No console errors

#### No Data
```vue
<DashboardWidget title="Empty Widget" />
```
- [ ] Shows empty default slot area
- [ ] No errors in console
- [ ] Proper spacing maintained

### Performance Tests

#### Rendering
- [ ] Component mounts quickly (<100ms)
- [ ] Skeleton animation is smooth (60fps)
- [ ] No layout shift during state changes
- [ ] Transitions are smooth

#### Memory
- [ ] No memory leaks on mount/unmount
- [ ] Event listeners are cleaned up
- [ ] Multiple instances don't cause issues

### Browser Compatibility

Test in:
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)
- [ ] Chrome Mobile
- [ ] Safari Mobile

## Automated Test Suite (Vitest)

```typescript
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DashboardWidget from './DashboardWidget.vue'

describe('DashboardWidget', () => {
  it('renders title correctly', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test Widget' }
    })
    expect(wrapper.text()).toContain('Test Widget')
  })

  it('shows loading state when loading prop is true', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test', loading: true }
    })
    expect(wrapper.find('.animate-pulse').exists()).toBe(true)
  })

  it('shows error state when error prop is provided', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test', error: 'Test error' }
    })
    expect(wrapper.text()).toContain('Failed to load data')
    expect(wrapper.text()).toContain('Test error')
  })

  it('emits retry event when retry button clicked', async () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test', error: 'Test error' }
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('retry')).toBeTruthy()
  })

  it('shows empty state when empty prop is true', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test', empty: true, emptyMessage: 'Custom empty' }
    })
    expect(wrapper.text()).toContain('Custom empty')
  })

  it('renders default slot content', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test' },
      slots: { default: '<div>Test Content</div>' }
    })
    expect(wrapper.text()).toContain('Test Content')
  })

  it('renders actions slot in header', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test' },
      slots: { actions: '<button>Action</button>' }
    })
    expect(wrapper.find('button').text()).toBe('Action')
  })

  it('sets aria-busy when loading', () => {
    const wrapper = mount(DashboardWidget, {
      props: { title: 'Test', loading: true }
    })
    expect(wrapper.attributes('aria-busy')).toBe('true')
  })

  it('prioritizes loading over error and empty', () => {
    const wrapper = mount(DashboardWidget, {
      props: {
        title: 'Test',
        loading: true,
        error: 'Error',
        empty: true
      }
    })
    expect(wrapper.find('.animate-pulse').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Failed to load data')
  })

  it('prioritizes error over empty when not loading', () => {
    const wrapper = mount(DashboardWidget, {
      props: {
        title: 'Test',
        loading: false,
        error: 'Error',
        empty: true
      }
    })
    expect(wrapper.text()).toContain('Failed to load data')
    expect(wrapper.text()).not.toContain('No data available')
  })
})
```

## Visual Regression Tests (Playwright)

```typescript
import { test, expect } from '@playwright/test'

test.describe('DashboardWidget Component', () => {
  test('default state screenshot', async ({ page }) => {
    await page.goto('/component-library/dashboard-widget')
    await expect(page.locator('.dashboard-widget')).toHaveScreenshot('default.png')
  })

  test('loading state screenshot', async ({ page }) => {
    await page.goto('/component-library/dashboard-widget?state=loading')
    await expect(page.locator('.dashboard-widget')).toHaveScreenshot('loading.png')
  })

  test('error state screenshot', async ({ page }) => {
    await page.goto('/component-library/dashboard-widget?state=error')
    await expect(page.locator('.dashboard-widget')).toHaveScreenshot('error.png')
  })

  test('empty state screenshot', async ({ page }) => {
    await page.goto('/component-library/dashboard-widget?state=empty')
    await expect(page.locator('.dashboard-widget')).toHaveScreenshot('empty.png')
  })
})
```

## Coverage Goals

- [ ] Line coverage: >90%
- [ ] Branch coverage: >85%
- [ ] Function coverage: 100%
- [ ] Statement coverage: >90%

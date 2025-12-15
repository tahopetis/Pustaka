/**
 * DonutChart Component Tests
 *
 * Test coverage for the D3.js donut chart component including:
 * - Data rendering and segment generation
 * - Interactive hover effects
 * - Legend display and synchronization
 * - Accessibility features
 * - Empty state handling
 * - Color customization
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DonutChart from './DonutChart.vue'
import type { DonutChartData } from '@/types/dashboard'

describe('DonutChart', () => {
  // ============================================================================
  // Test Data
  // ============================================================================

  const mockData: DonutChartData[] = [
    { label: 'Servers', value: 45, color: '#3B82F6' },
    { label: 'Databases', value: 30, color: '#10B981' },
    { label: 'Applications', value: 25, color: '#F59E0B' },
  ]

  const mockDataWithoutColors: DonutChartData[] = [
    { label: 'Type A', value: 50 },
    { label: 'Type B', value: 30 },
    { label: 'Type C', value: 20 },
  ]

  // ============================================================================
  // Rendering Tests
  // ============================================================================

  describe('Basic Rendering', () => {
    it('renders the component', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      expect(wrapper.exists()).toBe(true)
    })

    it('renders SVG element', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const svg = wrapper.find('svg')
      expect(svg.exists()).toBe(true)
    })

    it('renders with title when provided', () => {
      const wrapper = mount(DonutChart, {
        props: {
          data: mockData,
          title: 'CI Distribution',
        },
      })

      expect(wrapper.text()).toContain('CI Distribution')
    })

    it('renders without title when not provided', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const title = wrapper.find('h3')
      expect(title.exists()).toBe(false)
    })

    it('applies custom size', () => {
      const customSize = 400
      const wrapper = mount(DonutChart, {
        props: {
          data: mockData,
          size: customSize,
        },
      })

      const container = wrapper.find('.donut-chart-container')
      expect(container.attributes('style')).toContain(`width: ${customSize}px`)
      expect(container.attributes('style')).toContain(`height: ${customSize}px`)
    })
  })

  // ============================================================================
  // Data Processing Tests
  // ============================================================================

  describe('Data Processing', () => {
    it('calculates total value correctly', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const expectedTotal = mockData.reduce((sum, item) => sum + item.value, 0)
      expect(wrapper.text()).toContain(expectedTotal.toString())
    })

    it('calculates percentages correctly', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const total = mockData.reduce((sum, item) => sum + item.value, 0)
      mockData.forEach((item) => {
        const percentage = ((item.value / total) * 100).toFixed(1)
        expect(wrapper.text()).toContain(`${percentage}%`)
      })
    })

    it('generates correct number of segments', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      expect(segments).toHaveLength(mockData.length)
    })

    it('uses fallback colors when colors not provided', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockDataWithoutColors },
      })

      const segments = wrapper.findAll('.segment')
      expect(segments.length).toBeGreaterThan(0)

      // Each segment should have a fill attribute
      segments.forEach((segment) => {
        expect(segment.attributes('fill')).toBeTruthy()
      })
    })

    it('uses custom colors when provided', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      segments.forEach((segment, index) => {
        expect(segment.attributes('fill')).toBe(mockData[index].color)
      })
    })
  })

  // ============================================================================
  // Legend Tests
  // ============================================================================

  describe('Legend', () => {
    it('renders legend items', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      mockData.forEach((item) => {
        expect(wrapper.text()).toContain(item.label)
        expect(wrapper.text()).toContain(item.value.toString())
      })
    })

    it('displays color indicators', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const colorIndicators = wrapper.findAll('.w-3.h-3.rounded-sm')
      expect(colorIndicators).toHaveLength(mockData.length)

      colorIndicators.forEach((indicator, index) => {
        const style = indicator.attributes('style')
        expect(style).toContain(`background-color: ${mockData[index].color}`)
      })
    })

    it('shows percentages in legend', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const total = mockData.reduce((sum, item) => sum + item.value, 0)
      mockData.forEach((item) => {
        const percentage = ((item.value / total) * 100).toFixed(1)
        expect(wrapper.text()).toContain(`${percentage}%`)
      })
    })
  })

  // ============================================================================
  // Interactive Tests
  // ============================================================================

  describe('Interactivity', () => {
    it('shows tooltip on segment hover', async () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segment = wrapper.find('.segment')
      await segment.trigger('mouseenter')

      // Tooltip should become visible
      const tooltip = wrapper.find('[class*="bg-gray-900"]')
      expect(tooltip.exists()).toBe(true)
    })

    it('hides tooltip on mouse leave', async () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segment = wrapper.find('.segment')
      await segment.trigger('mouseenter')
      await segment.trigger('mouseleave')

      // Check that tooltip visibility flag is false
      // Note: In actual DOM, v-if would remove the element
      const vm = wrapper.vm as any
      expect(vm.tooltip.visible).toBe(false)
    })

    it('highlights segment on legend item hover', async () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const legendItems = wrapper.findAll('[role="button"]')
      await legendItems[0].trigger('mouseenter')

      const vm = wrapper.vm as any
      expect(vm.hoveredIndex).toBe(0)
    })

    it('applies hover styles to segments', async () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      await segments[0].trigger('mouseenter')

      const vm = wrapper.vm as any
      expect(vm.hoveredIndex).toBe(0)
    })
  })

  // ============================================================================
  // Accessibility Tests
  // ============================================================================

  describe('Accessibility', () => {
    it('has proper ARIA label on SVG', () => {
      const wrapper = mount(DonutChart, {
        props: {
          data: mockData,
          title: 'Test Chart',
        },
      })

      const svg = wrapper.find('svg')
      const ariaLabel = svg.attributes('aria-label')
      expect(ariaLabel).toBeTruthy()
      expect(ariaLabel).toContain('Test Chart')
    })

    it('has role="img" on SVG', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const svg = wrapper.find('svg')
      expect(svg.attributes('role')).toBe('img')
    })

    it('segments are keyboard accessible', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      segments.forEach((segment) => {
        expect(segment.attributes('tabindex')).toBe('0')
        expect(segment.attributes('role')).toBe('button')
      })
    })

    it('segments have descriptive aria-labels', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      segments.forEach((segment, index) => {
        const ariaLabel = segment.attributes('aria-label')
        expect(ariaLabel).toContain(mockData[index].label)
        expect(ariaLabel).toContain(mockData[index].value.toString())
      })
    })

    it('legend items are keyboard accessible', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const legendItems = wrapper.findAll('[role="button"]')
      legendItems.forEach((item) => {
        expect(item.attributes('tabindex')).toBe('0')
        expect(item.attributes('aria-label')).toBeTruthy()
      })
    })

    it('legend items have descriptive aria-labels', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const legendItems = wrapper.findAll('[role="button"]')
      legendItems.forEach((item, index) => {
        const ariaLabel = item.attributes('aria-label')
        expect(ariaLabel).toContain(mockData[index].label)
        expect(ariaLabel).toContain(mockData[index].value.toString())
      })
    })
  })

  // ============================================================================
  // Empty State Tests
  // ============================================================================

  describe('Empty State', () => {
    it('shows empty state when no data', () => {
      const wrapper = mount(DonutChart, {
        props: { data: [] },
      })

      expect(wrapper.text()).toContain('No data available')
    })

    it('shows empty state when all values are zero', () => {
      const emptyData = [
        { label: 'A', value: 0, color: '#000' },
        { label: 'B', value: 0, color: '#000' },
      ]

      const wrapper = mount(DonutChart, {
        props: { data: emptyData },
      })

      expect(wrapper.text()).toContain('No data available')
    })

    it('does not show empty state when data is valid', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      expect(wrapper.text()).not.toContain('No data available')
    })
  })

  // ============================================================================
  // Props Tests
  // ============================================================================

  describe('Props', () => {
    it('uses default size when not specified', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const container = wrapper.find('.donut-chart-container')
      expect(container.attributes('style')).toContain('width: 300px')
      expect(container.attributes('style')).toContain('height: 300px')
    })

    it('applies custom inner radius ratio', async () => {
      const wrapper = mount(DonutChart, {
        props: {
          data: mockData,
          innerRadiusRatio: 0.5,
        },
      })

      // The component should render without errors
      expect(wrapper.exists()).toBe(true)
    })

    it('handles data updates reactively', async () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const newData = [
        { label: 'New Item', value: 100, color: '#000000' },
      ]

      await wrapper.setProps({ data: newData })

      expect(wrapper.text()).toContain('New Item')
      expect(wrapper.text()).toContain('100')
    })
  })

  // ============================================================================
  // SVG Generation Tests
  // ============================================================================

  describe('SVG Generation', () => {
    it('generates valid arc paths', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const segments = wrapper.findAll('.segment')
      segments.forEach((segment) => {
        const d = segment.attributes('d')
        expect(d).toBeTruthy()
        // Arc paths should contain M (moveto) and A (arc) commands
        expect(d).toMatch(/M/)
        expect(d).toMatch(/A/)
      })
    })

    it('renders center label with total value', () => {
      const wrapper = mount(DonutChart, {
        props: { data: mockData },
      })

      const total = mockData.reduce((sum, item) => sum + item.value, 0)
      const centerLabels = wrapper.findAll('.center-label text')

      expect(centerLabels.length).toBeGreaterThan(0)
      expect(wrapper.text()).toContain('Total')
      expect(wrapper.text()).toContain(total.toString())
    })

    it('applies correct viewBox', () => {
      const wrapper = mount(DonutChart, {
        props: {
          data: mockData,
          size: 400,
        },
      })

      const svg = wrapper.find('svg')
      const viewBox = svg.attributes('viewBox')
      expect(viewBox).toBeTruthy()
    })
  })

  // ============================================================================
  // Edge Cases
  // ============================================================================

  describe('Edge Cases', () => {
    it('handles single segment', () => {
      const singleData = [{ label: 'Only One', value: 100, color: '#3B82F6' }]

      const wrapper = mount(DonutChart, {
        props: { data: singleData },
      })

      const segments = wrapper.findAll('.segment')
      expect(segments).toHaveLength(1)
      expect(wrapper.text()).toContain('100%')
    })

    it('handles many segments', () => {
      const manyData = Array.from({ length: 20 }, (_, i) => ({
        label: `Item ${i + 1}`,
        value: Math.floor(Math.random() * 100) + 1,
        color: `#${Math.floor(Math.random() * 16777215).toString(16)}`,
      }))

      const wrapper = mount(DonutChart, {
        props: { data: manyData },
      })

      const segments = wrapper.findAll('.segment')
      expect(segments).toHaveLength(20)
    })

    it('handles very small values', () => {
      const smallData = [
        { label: 'A', value: 1, color: '#3B82F6' },
        { label: 'B', value: 999, color: '#10B981' },
      ]

      const wrapper = mount(DonutChart, {
        props: { data: smallData },
      })

      expect(wrapper.text()).toContain('1000')
      expect(wrapper.text()).toContain('0.1%')
    })

    it('handles equal values', () => {
      const equalData = [
        { label: 'A', value: 50, color: '#3B82F6' },
        { label: 'B', value: 50, color: '#10B981' },
      ]

      const wrapper = mount(DonutChart, {
        props: { data: equalData },
      })

      expect(wrapper.text()).toContain('50.0%')
    })
  })
})

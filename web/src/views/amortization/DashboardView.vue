<template>
  <div class="amortization-dashboard">
    <div class="page-header">
      <h1 class="page-title">Amortization Dashboard</h1>
      <p class="page-description">
        Overview of asset amortization and depreciation tracking
      </p>
    </div>

    <!-- Adjustment Dialog -->
    <AdjustmentDialog
      :is-open="showAdjustmentDialog"
      @close="showAdjustmentDialog = false"
      @created="handleAdjustmentCreated"
    />

    <!-- Restructure Dialog -->
    <RestructureDialog
      :show="showRestructureDialog"
      @close="showRestructureDialog = false"
      @success="handleRestructureSuccess"
    />

    <!-- Key Metrics -->
    <div class="metrics-grid">
      <!-- OCC Card -->
      <div class="metric-card metric-card-gray">
        <div class="metric-icon">
          <i class="fas fa-shopping-cart"></i>
        </div>
        <div class="metric-content">
          <div class="metric-label">Original Capitalized Cost</div>
          <div class="metric-value">{{ formatCurrency(metrics.total_original_cost || 0) }}</div>
          <div class="metric-sublabel">OCC</div>
        </div>
      </div>

      <!-- GVB Card -->
      <div class="metric-card metric-card-purple">
        <div class="metric-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="metric-content">
          <div class="metric-label">Gross Book Value</div>
          <div class="metric-value">{{ formatCurrency(metrics.total_gross_book_value || 0) }}</div>
          <div class="metric-sublabel">GVB</div>
        </div>
      </div>

      <!-- NBV Card -->
      <div class="metric-card metric-card-blue">
        <div class="metric-icon">
          <i class="fas fa-book"></i>
        </div>
        <div class="metric-content">
          <div class="metric-label">Net Book Value</div>
          <div class="metric-value">{{ formatCurrency(metrics.total_net_book_value || 0) }}</div>
          <div class="metric-sublabel">NBV</div>
        </div>
      </div>

      <!-- AD Card -->
      <div class="metric-card metric-card-orange">
        <div class="metric-icon">
          <i class="fas fa-chart-area"></i>
        </div>
        <div class="metric-content">
          <div class="metric-label">Accumulated Depreciation</div>
          <div class="metric-value">{{ formatCurrency(metrics.total_accumulated_depreciation || 0) }}</div>
          <div class="metric-sublabel">AD</div>
          <div v-if="metrics.total_gross_book_value > 0" class="metric-badge">
            {{ ((metrics.total_accumulated_depreciation || 0) / metrics.total_gross_book_value * 100).toFixed(1) }}%
          </div>
        </div>
      </div>

      <!-- SV Card -->
      <div class="metric-card metric-card-red">
        <div class="metric-icon">
          <i class="fas fa-anchor"></i>
        </div>
        <div class="metric-content">
          <div class="metric-label">Salvage Value</div>
          <div class="metric-value">{{ formatCurrency(metrics.total_salvage_value || 0) }}</div>
          <div class="metric-sublabel">SV</div>
        </div>
      </div>
    </div>

    <!-- Historical Book Value Chart -->
    <div class="chart-section">
      <h3>Historical Book Value Trend</h3>
      <div v-if="chartLoading" class="loading-placeholder">
        Loading chart data...
      </div>
      <div v-else-if="chartData.length > 0" class="chart-container">
        <svg ref="chartSvg" class="chart-svg"></svg>
      </div>
      <div v-else class="no-data">
        No historical data available yet. The chart will appear once amortization entries are recorded.
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions">
      <h2>Quick Actions</h2>
      <div class="actions-grid">
        <router-link to="/ci" class="action-card">
          <i class="fas fa-plus"></i>
          <h3>Manage CI Types</h3>
          <p>Configure amortization for CI types</p>
        </router-link>

        <router-link to="/ci" class="action-card">
          <i class="fas fa-dollar-sign"></i>
          <h3>Asset Financials</h3>
          <p>Set financial information for assets</p>
        </router-link>

        <router-link to="/amortization/ledger" class="action-card">
          <i class="fas fa-book"></i>
          <h3>View Ledger</h3>
          <p>Browse amortization ledger entries</p>
        </router-link>

        <router-link to="/amortization/reports" class="action-card">
          <i class="fas fa-chart-bar"></i>
          <h3>Reports</h3>
          <p>Generate amortization reports</p>
        </router-link>

        <button @click="showAdjustmentDialog = true" class="action-card action-button">
          <i class="fas fa-edit"></i>
          <h3>Create Adjustment</h3>
          <p>Record manual adjustment to asset book value</p>
        </button>

        <button @click="showRestructureDialog = true" class="action-card action-button">
          <i class="fas fa-clock"></i>
          <h3>Restructure Amortization</h3>
          <p>Change useful life with prospective recalculation</p>
        </button>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="recent-activity">
      <h2>Recent Amortization Activity</h2>
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading recent activity...
      </div>
      <div v-else-if="recentActivity.length === 0" class="no-data">
        No recent amortization activity found.
      </div>
      <div v-else class="activity-list">
        <div
          v-for="activity in recentActivity"
          :key="activity.id"
          class="activity-item"
        >
          <div class="activity-badge">
            <span :class="getBadgeClasses(activity.entry_type)">
              {{ getBadgeLabel(activity.entry_type) }}
            </span>
          </div>
          <div class="activity-content">
            <h4>{{ activity.description }}</h4>
            <p class="activity-details">
              {{ formatCurrency(activity.amount) }} - {{ formatDate(activity.entry_date) }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationLedgerEntry, AmortizationMetrics } from '@/types/amortization'
import AdjustmentDialog from '@/components/amortization/AdjustmentDialog.vue'
import RestructureDialog from '@/components/amortization/RestructureDialog.vue'

const amortizationStore = useAmortizationStore()

const loading = ref(true)
const showAdjustmentDialog = ref(false)
const showRestructureDialog = ref(false)
const chartLoading = ref(false)
const metrics = ref<AmortizationMetrics>({
  total_amortizable_assets: 0,
  total_book_value: 0,
  monthly_depreciation: 0,
  active_amortizations: 0,
  total_original_cost: 0,
  total_gross_book_value: 0,
  total_net_book_value: 0,
  total_accumulated_depreciation: 0,
  total_salvage_value: 0,
  total_monthly_depreciation: 0,
})
const recentActivity = ref<AmortizationLedgerEntry[]>([])
const chartData = ref<any[]>([])
const chartSvg = ref<SVGSVGElement | null>(null)

onMounted(async () => {
  await loadDashboardData()
  loading.value = false
  // Load chart data after metrics are loaded
  if (metrics.value.total_original_cost > 0) {
    await loadChartData()
  }
})

// Watch for metrics updates
watch(() => metrics.value.total_original_cost, async (newValue) => {
  if (newValue > 0) {
    await loadChartData()
  }
})

const loadChartData = async () => {
  chartLoading.value = true
  try {
    // Fetch historical ledger data aggregated by month (last 24 months)
    const endDate = new Date()
    const startDate = new Date()
    startDate.setMonth(startDate.getMonth() - 24)

    const response = await fetch('/api/v1/amortization/reports/depreciation-schedule?' +
      new URLSearchParams({
        date_from: startDate.toISOString().split('T')[0],
        date_to: endDate.toISOString().split('T')[0],
      }).toString(), {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      }
    })

    if (response.ok) {
      const data = await response.json()
      // Filter to only historical data
      chartData.value = data.monthly_data.filter((entry: any) => !entry.is_projected)

      // Wait for DOM update twice - first for chartData to update, second for SVG to render
      await nextTick()
      await nextTick()
      drawChart()
    } else {
      console.error('Failed to load chart data:', response.status, response.statusText)
    }
  } catch (error) {
    console.error('Failed to load chart data:', error)
  } finally {
    chartLoading.value = false
  }
}

const drawChart = () => {
  if (!chartSvg.value || !chartData.value?.length) {
    return
  }

  const svg = chartSvg.value
  const data = chartData.value
  const width = svg.clientWidth || 800
  const height = 400
  const padding = { top: 40, right: 40, bottom: 60, left: 80 }

  // Clear existing content
  while (svg.firstChild) {
    svg.removeChild(svg.firstChild)
  }

  // Set SVG dimensions
  svg.setAttribute('width', width.toString())
  svg.setAttribute('height', height.toString())
  svg.setAttribute('viewBox', `0 0 ${width} ${height}`)

  // Calculate scales
  const maxValue = Math.max(...data.map(d => d.opening_book_value), ...data.map(d => d.closing_book_value)) * 1.1
  const chartWidth = width - padding.left - padding.right
  const chartHeight = height - padding.top - padding.bottom

  // Create namespace
  const ns = 'http://www.w3.org/2000/svg'

  // Draw background
  const bg = document.createElementNS(ns, 'rect')
  bg.setAttribute('x', '0')
  bg.setAttribute('y', '0')
  bg.setAttribute('width', width.toString())
  bg.setAttribute('height', height.toString())
  bg.setAttribute('fill', '#f9fafb')
  svg.appendChild(bg)

  // Draw grid lines
  for (let i = 0; i <= 4; i++) {
    const y = padding.top + (i / 4) * chartHeight
    const gridLine = document.createElementNS(ns, 'line')
    gridLine.setAttribute('x1', padding.left.toString())
    gridLine.setAttribute('y1', y.toString())
    gridLine.setAttribute('x2', (width - padding.right).toString())
    gridLine.setAttribute('y2', y.toString())
    gridLine.setAttribute('stroke', '#e5e7eb')
    gridLine.setAttribute('stroke-width', '1')
    svg.appendChild(gridLine)
  }

  // Reference line values from metrics
  const totalOriginalCost = metrics.value.total_original_cost || 0
  const totalGrossBookValue = metrics.value.total_gross_book_value || 0
  const totalSalvageValue = metrics.value.total_salvage_value || 0

  // Draw OCC reference line (gray dashed)
  if (totalOriginalCost > 0 && totalOriginalCost < maxValue) {
    const occY = padding.top + chartHeight - (totalOriginalCost / maxValue) * chartHeight
    const occLine = document.createElementNS(ns, 'line')
    occLine.setAttribute('x1', padding.left.toString())
    occLine.setAttribute('y1', occY.toString())
    occLine.setAttribute('x2', (width - padding.right).toString())
    occLine.setAttribute('y2', occY.toString())
    occLine.setAttribute('stroke', '#6b7280')
    occLine.setAttribute('stroke-width', '2')
    occLine.setAttribute('stroke-dasharray', '12,4,4,4')
    svg.appendChild(occLine)

    const occLabel = document.createElementNS(ns, 'text')
    occLabel.setAttribute('x', (width - padding.right - 10).toString())
    occLabel.setAttribute('y', (occY - 8).toString())
    occLabel.setAttribute('text-anchor', 'end')
    occLabel.setAttribute('font-size', '10')
    occLabel.setAttribute('font-weight', '600')
    occLabel.setAttribute('fill', '#6b7280')
    occLabel.textContent = `OCC: ${formatCurrency(totalOriginalCost)}`
    svg.appendChild(occLabel)
  }

  // Draw GVB reference line (purple dashed)
  if (totalGrossBookValue > 0 && totalGrossBookValue < maxValue) {
    const gvbY = padding.top + chartHeight - (totalGrossBookValue / maxValue) * chartHeight
    const gvbLine = document.createElementNS(ns, 'line')
    gvbLine.setAttribute('x1', padding.left.toString())
    gvbLine.setAttribute('y1', gvbY.toString())
    gvbLine.setAttribute('x2', (width - padding.right).toString())
    gvbLine.setAttribute('y2', gvbY.toString())
    gvbLine.setAttribute('stroke', '#8b5cf6')
    gvbLine.setAttribute('stroke-width', '2')
    gvbLine.setAttribute('stroke-dasharray', '8,4')
    svg.appendChild(gvbLine)

    const gvbLabel = document.createElementNS(ns, 'text')
    gvbLabel.setAttribute('x', (width - padding.right - 10).toString())
    gvbLabel.setAttribute('y', (gvbY - 8).toString())
    gvbLabel.setAttribute('text-anchor', 'end')
    gvbLabel.setAttribute('font-size', '10')
    gvbLabel.setAttribute('font-weight', '600')
    gvbLabel.setAttribute('fill', '#8b5cf6')
    gvbLabel.textContent = `GVB: ${formatCurrency(totalGrossBookValue)}`
    svg.appendChild(gvbLabel)
  }

  // Draw salvage value reference line
  if (totalSalvageValue > 0 && totalSalvageValue < maxValue) {
    const svY = padding.top + chartHeight - (totalSalvageValue / maxValue) * chartHeight
    const svLine = document.createElementNS(ns, 'line')
    svLine.setAttribute('x1', padding.left.toString())
    svLine.setAttribute('y1', svY.toString())
    svLine.setAttribute('x2', (width - padding.right).toString())
    svLine.setAttribute('y2', svY.toString())
    svLine.setAttribute('stroke', '#ef4444')
    svLine.setAttribute('stroke-width', '2')
    svLine.setAttribute('stroke-dasharray', '8,4')
    svg.appendChild(svLine)

    const svLabel = document.createElementNS(ns, 'text')
    svLabel.setAttribute('x', (width - padding.right - 10).toString())
    svLabel.setAttribute('y', (svY - 8).toString())
    svLabel.setAttribute('text-anchor', 'end')
    svLabel.setAttribute('font-size', '10')
    svLabel.setAttribute('font-weight', '600')
    svLabel.setAttribute('fill', '#ef4444')
    svLabel.textContent = `SV: ${formatCurrency(totalSalvageValue)}`
    svg.appendChild(svLabel)
  }

  // Draw book value line
  const path = document.createElementNS(ns, 'path')
  let pathD = ''

  data.forEach((entry, index) => {
    // Handle single data point case
    const x = data.length === 1
      ? padding.left + chartWidth / 2  // Center position for single point
      : padding.left + (index / (data.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (entry.closing_book_value / maxValue) * chartHeight

    if (index === 0) {
      pathD += `M ${x} ${y}`
    } else {
      pathD += ` L ${x} ${y}`
    }

    // Draw depreciation bars
    const barHeight = (entry.depreciation_amount / maxValue) * chartHeight
    if (barHeight > 0) {
      const bar = document.createElementNS(ns, 'rect')
      bar.setAttribute('x', (x - 3).toString())
      bar.setAttribute('y', (padding.top + chartHeight - (entry.closing_book_value / maxValue) * chartHeight - barHeight).toString())
      bar.setAttribute('width', '6')
      bar.setAttribute('height', barHeight.toString())
      bar.setAttribute('fill', 'rgba(59, 130, 246, 0.3)')
      bar.setAttribute('rx', '2')
      svg.appendChild(bar)
    }
  })

  path.setAttribute('d', pathD)
  path.setAttribute('fill', 'none')
  path.setAttribute('stroke', '#3b82f6')
  path.setAttribute('stroke-width', '3')
  svg.appendChild(path)

  // Draw point markers for each data point with SVG tooltips
  data.forEach((entry, index) => {
    const x = data.length === 1
      ? padding.left + chartWidth / 2
      : padding.left + (index / (data.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (entry.closing_book_value / maxValue) * chartHeight

    // Create group for interactivity
    const group = document.createElementNS(ns, 'g')
    group.style.cursor = 'pointer'

    // Outer circle for hover effect
    const outerCircle = document.createElementNS(ns, 'circle')
    outerCircle.setAttribute('cx', x.toString())
    outerCircle.setAttribute('cy', y.toString())
    outerCircle.setAttribute('r', '8')
    outerCircle.setAttribute('fill', 'transparent')
    outerCircle.setAttribute('class', 'chart-point-hover')
    group.appendChild(outerCircle)

    // Inner circle
    const circle = document.createElementNS(ns, 'circle')
    circle.setAttribute('cx', x.toString())
    circle.setAttribute('cy', y.toString())
    circle.setAttribute('r', '5')
    circle.setAttribute('fill', '#3b82f6')
    circle.setAttribute('stroke', '#fff')
    circle.setAttribute('stroke-width', '2')
    group.appendChild(circle)

    // Tooltip group (hidden by default)
    const tooltipGroup = document.createElementNS(ns, 'g')
    tooltipGroup.setAttribute('class', 'tooltip-group')
    tooltipGroup.style.opacity = '0'
    tooltipGroup.style.transition = 'opacity 0.2s'

    // Tooltip background
    const tooltipBg = document.createElementNS(ns, 'rect')
    tooltipBg.setAttribute('x', (x + 10).toString())
    tooltipBg.setAttribute('y', (y - 60).toString())
    tooltipBg.setAttribute('width', '150')
    tooltipBg.setAttribute('height', '55')
    tooltipBg.setAttribute('rx', '4')
    tooltipBg.setAttribute('fill', '#1f2937')
    tooltipBg.setAttribute('opacity', '0.95')
    tooltipGroup.appendChild(tooltipBg)

    // Tooltip text - Date
    const date = new Date(entry.month)
    const tooltipTitle = document.createElementNS(ns, 'text')
    tooltipTitle.setAttribute('x', (x + 18).toString())
    tooltipTitle.setAttribute('y', (y - 43).toString())
    tooltipTitle.setAttribute('font-size', '11')
    tooltipTitle.setAttribute('font-weight', '600')
    tooltipTitle.setAttribute('fill', '#fff')
    tooltipTitle.textContent = date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
    tooltipGroup.appendChild(tooltipTitle)

    // Tooltip text - Book Value
    const tooltipValue = document.createElementNS(ns, 'text')
    tooltipValue.setAttribute('x', (x + 18).toString())
    tooltipValue.setAttribute('y', (y - 28).toString())
    tooltipValue.setAttribute('font-size', '10')
    tooltipValue.setAttribute('fill', '#e5e7eb')
    tooltipValue.textContent = `Book Value: ${formatCurrency(entry.closing_book_value)}`
    tooltipGroup.appendChild(tooltipValue)

    // Tooltip text - Depreciation
    const tooltipDepreciation = document.createElementNS(ns, 'text')
    tooltipDepreciation.setAttribute('x', (x + 18).toString())
    tooltipDepreciation.setAttribute('y', (y - 14).toString())
    tooltipDepreciation.setAttribute('font-size', '10')
    tooltipDepreciation.setAttribute('fill', '#93c5fd')
    tooltipDepreciation.textContent = `Depreciation: ${formatCurrency(entry.depreciation_amount)}`
    tooltipGroup.appendChild(tooltipDepreciation)

    // Show tooltip on hover
    group.addEventListener('mouseenter', () => {
      tooltipGroup.style.opacity = '1'
      outerCircle.setAttribute('fill', 'rgba(59, 130, 246, 0.1)')
    })

    group.addEventListener('mouseleave', () => {
      tooltipGroup.style.opacity = '0'
      outerCircle.setAttribute('fill', 'transparent')
    })

    group.appendChild(tooltipGroup)
    svg.appendChild(group)
  })

  // Draw axes
  const xAxis = document.createElementNS(ns, 'line')
  xAxis.setAttribute('x1', padding.left.toString())
  xAxis.setAttribute('y1', (height - padding.bottom).toString())
  xAxis.setAttribute('x2', (width - padding.right).toString())
  xAxis.setAttribute('y2', (height - padding.bottom).toString())
  xAxis.setAttribute('stroke', '#374151')
  xAxis.setAttribute('stroke-width', '2')
  svg.appendChild(xAxis)

  const yAxis = document.createElementNS(ns, 'line')
  yAxis.setAttribute('x1', padding.left.toString())
  yAxis.setAttribute('y1', padding.top.toString())
  yAxis.setAttribute('x2', padding.left.toString())
  yAxis.setAttribute('y2', (height - padding.bottom).toString())
  yAxis.setAttribute('stroke', '#374151')
  yAxis.setAttribute('stroke-width', '2')
  svg.appendChild(yAxis)

  // Y-axis labels
  for (let i = 0; i <= 4; i++) {
    const value = (maxValue * (4 - i)) / 4
    const y = padding.top + (i / 4) * chartHeight
    const text = document.createElementNS(ns, 'text')
    text.setAttribute('x', (padding.left - 10).toString())
    text.setAttribute('y', (y + 4).toString())
    text.setAttribute('text-anchor', 'end')
    text.setAttribute('font-size', '10')
    text.setAttribute('fill', '#6b7280')
    text.textContent = formatCurrency(value)
    svg.appendChild(text)
  }

  // X-axis labels (show 6 months)
  const labelInterval = Math.ceil(data.length / 6)
  data.forEach((entry, index) => {
    if (index % labelInterval === 0 || index === data.length - 1) {
      // Handle single data point case
      const x = data.length === 1
        ? padding.left + chartWidth / 2  // Center position for single point
        : padding.left + (index / (data.length - 1)) * chartWidth
      const monthLabel = document.createElementNS(ns, 'text')
      monthLabel.setAttribute('x', x.toString())
      monthLabel.setAttribute('y', (height - padding.bottom + 20).toString())
      monthLabel.setAttribute('text-anchor', 'middle')
      monthLabel.setAttribute('font-size', '10')
      monthLabel.setAttribute('fill', '#6b7280')
      const date = new Date(entry.month)
      monthLabel.textContent = date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' })
      svg.appendChild(monthLabel)
    }
  })

  // Add legend
  const legendY = height - 25
  let legendX = padding.left

  // OCC
  const legendOCCLine = document.createElementNS(ns, 'line')
  legendOCCLine.setAttribute('x1', legendX.toString())
  legendOCCLine.setAttribute('y1', legendY.toString())
  legendOCCLine.setAttribute('x2', (legendX + 30).toString())
  legendOCCLine.setAttribute('y2', legendY.toString())
  legendOCCLine.setAttribute('stroke', '#6b7280')
  legendOCCLine.setAttribute('stroke-width', '2')
  legendOCCLine.setAttribute('stroke-dasharray', '12,4,4,4')
  svg.appendChild(legendOCCLine)

  const legendTextOCC = document.createElementNS(ns, 'text')
  legendTextOCC.setAttribute('x', (legendX + 38).toString())
  legendTextOCC.setAttribute('y', (legendY + 4).toString())
  legendTextOCC.setAttribute('font-size', '11')
  legendTextOCC.setAttribute('fill', '#6b7280')
  legendTextOCC.textContent = 'OCC'
  svg.appendChild(legendTextOCC)

  legendX += 80

  // GVB
  const legendGVBLine = document.createElementNS(ns, 'line')
  legendGVBLine.setAttribute('x1', legendX.toString())
  legendGVBLine.setAttribute('y1', legendY.toString())
  legendGVBLine.setAttribute('x2', (legendX + 30).toString())
  legendGVBLine.setAttribute('y2', legendY.toString())
  legendGVBLine.setAttribute('stroke', '#8b5cf6')
  legendGVBLine.setAttribute('stroke-width', '2')
  legendGVBLine.setAttribute('stroke-dasharray', '8,4')
  svg.appendChild(legendGVBLine)

  const legendTextGVB = document.createElementNS(ns, 'text')
  legendTextGVB.setAttribute('x', (legendX + 38).toString())
  legendTextGVB.setAttribute('y', (legendY + 4).toString())
  legendTextGVB.setAttribute('font-size', '11')
  legendTextGVB.setAttribute('fill', '#6b7280')
  legendTextGVB.textContent = 'GVB'
  svg.appendChild(legendTextGVB)

  legendX += 80

  // Book Value
  const legendLine = document.createElementNS(ns, 'line')
  legendLine.setAttribute('x1', legendX.toString())
  legendLine.setAttribute('y1', legendY.toString())
  legendLine.setAttribute('x2', (legendX + 30).toString())
  legendLine.setAttribute('y2', legendY.toString())
  legendLine.setAttribute('stroke', '#3b82f6')
  legendLine.setAttribute('stroke-width', '3')
  svg.appendChild(legendLine)

  const legendText1 = document.createElementNS(ns, 'text')
  legendText1.setAttribute('x', (legendX + 38).toString())
  legendText1.setAttribute('y', (legendY + 4).toString())
  legendText1.setAttribute('font-size', '11')
  legendText1.setAttribute('fill', '#6b7280')
  legendText1.textContent = 'Book Value'
  svg.appendChild(legendText1)

  legendX += 120

  // Monthly Depreciation
  const legendBar = document.createElementNS(ns, 'rect')
  legendBar.setAttribute('x', legendX.toString())
  legendBar.setAttribute('y', (legendY - 6).toString())
  legendBar.setAttribute('width', '20')
  legendBar.setAttribute('height', '12')
  legendBar.setAttribute('fill', 'rgba(59, 130, 246, 0.3)')
  legendBar.setAttribute('rx', '2')
  svg.appendChild(legendBar)

  const legendText2 = document.createElementNS(ns, 'text')
  legendText2.setAttribute('x', (legendX + 28).toString())
  legendText2.setAttribute('y', (legendY + 4).toString())
  legendText2.setAttribute('font-size', '11')
  legendText2.setAttribute('fill', '#6b7280')
  legendText2.textContent = 'Monthly Depreciation'
  svg.appendChild(legendText2)
}

const loadDashboardData = async () => {
  try {
    // Load metrics
    const summaryResponse = await amortizationStore.loadMetrics()
    if (summaryResponse) {
      metrics.value = summaryResponse
    }

    // Load recent activity
    const ledgerResponse = await amortizationStore.loadLedgerEntries({
      page_size: 10,
      sort_by: 'entry_date',
      sort_order: 'desc'
    })
    if (ledgerResponse && ledgerResponse.entries) {
      recentActivity.value = ledgerResponse.entries || []
    }
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  }
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount)
}

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString()
}

const getBadgeLabel = (entryType: string): string => {
  const labels = {
    'depreciation': 'Monthly',
    'monthly_depreciation': 'Monthly',
    'catch_up_depreciation': 'Catch-up',
    'adjustment': 'Adjustment',
    'write_off': 'Write-off',
    'reversal': '↩️ Reversal',
    'restructuring': '📊 Restructuring'
  }
  return labels[entryType] || formatEntryType(entryType)
}

const getBadgeClasses = (entryType: string): string => {
  const classes = {
    'depreciation': 'badge badge-gray',
    'monthly_depreciation': 'badge badge-gray',
    'catch_up_depreciation': 'badge badge-blue',
    'adjustment': 'badge badge-purple',
    'write_off': 'badge badge-red',
    'reversal': 'badge badge-gray badge-faded',
    'restructuring': 'badge badge-gray badge-faded'
  }
  return classes[entryType] || 'badge badge-gray'
}

const formatEntryType = (entryType: string): string => {
  return entryType.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

const handleAdjustmentCreated = async () => {
  // Reload dashboard data to show the new adjustment
  await loadDashboardData()
}

const handleRestructureSuccess = async () => {
  // Reload dashboard data to show the restructuring
  await loadDashboardData()
}
</script>

<style scoped>
.amortization-dashboard {
  padding: 2rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.page-description {
  color: #6b7280;
  font-size: 1rem;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.metric-card {
  background: white;
  padding: 1.25rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 1rem;
  position: relative;
  overflow: hidden;
}

.metric-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
}

.metric-card-gray::before { background: #6b7280; }
.metric-card-purple::before { background: #8b5cf6; }
.metric-card-blue::before { background: #3b82f6; }
.metric-card-orange::before { background: #f97316; }
.metric-card-red::before { background: #ef4444; }

.metric-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.metric-card-gray .metric-icon { background: #f3f4f6; color: #6b7280; }
.metric-card-purple .metric-icon { background: #ede9fe; color: #8b5cf6; }
.metric-card-blue .metric-icon { background: #dbeafe; color: #3b82f6; }
.metric-card-orange .metric-icon { background: #ffedd5; color: #f97316; }
.metric-card-red .metric-icon { background: #fee2e2; color: #ef4444; }

.metric-content {
  flex: 1;
}

.metric-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.2;
}

.metric-sublabel {
  font-size: 0.7rem;
  font-weight: 600;
  color: #9ca3af;
  margin-top: 0.25rem;
}

.metric-badge {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  padding: 0.25rem 0.5rem;
  background: #ffedd5;
  color: #c2410c;
  font-size: 0.7rem;
  font-weight: 600;
  border-radius: 9999px;
}

.metric-content h3 {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0 0 0.5rem 0;
}

.quick-actions {
  margin-bottom: 2rem;
}

.quick-actions h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.action-card {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  text-decoration: none;
  color: inherit;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.5rem;
  transition: all 0.2s;
}

.action-card:hover {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.action-card i {
  font-size: 2rem;
  color: #4f46e5;
}

.action-card h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.action-card p {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
}

.action-button {
  cursor: pointer;
  border: none;
  background: white;
  width: 100%;
}

.recent-activity h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
}

.no-data {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 0.5rem;
}

.activity-list {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-badge {
  flex-shrink: 0;
}

.activity-content h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.25rem 0;
}

.activity-details {
  font-size: 0.75rem;
  color: #6b7280;
  margin: 0;
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 9999px;
  white-space: nowrap;
}

.badge-gray {
  background: #f3f4f6;
  color: #374151;
}

.badge-blue {
  background: #dbeafe;
  color: #1e40af;
}

.badge-purple {
  background: #f3e8ff;
  color: #7c3aed;
}

.badge-red {
  background: #fee2e2;
  color: #991b1b;
}

.badge-faded {
  opacity: 0.7;
}

.chart-section {
  margin-bottom: 2rem;
}

.chart-section h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
}

.chart-container {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.chart-svg {
  width: 100%;
  height: auto;
  display: block;
}

.loading-placeholder {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 0.5rem;
  font-size: 0.875rem;
}

@media (max-width: 768px) {
  .amortization-dashboard {
    padding: 1rem;
  }

  .metrics-grid,
  .actions-grid {
    grid-template-columns: 1fr;
  }

  .metric-card {
    flex-direction: column;
    text-align: center;
  }
}
</style>
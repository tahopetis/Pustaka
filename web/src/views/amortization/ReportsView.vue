<template>
  <div class="amortization-reports">
    <div class="page-header">
      <h1 class="page-title">Amortization Reports</h1>
      <p class="page-description">
        Financial reports and analysis for amortized assets
      </p>
    </div>

    <!-- Summary Cards -->
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-coins"></i>
        </div>
        <div class="metric-content">
          <h3>Total Amortizable Assets</h3>
          <div class="metric-value">{{ metrics.total_amortizable_assets || 0 }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="metric-content">
          <h3>Total Book Value</h3>
          <div class="metric-value">{{ formatCurrency(metrics.total_book_value || 0) }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-calculator"></i>
        </div>
        <div class="metric-content">
          <h3>Monthly Depreciation</h3>
          <div class="metric-value">{{ formatCurrency(metrics.monthly_depreciation || 0) }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-history"></i>
        </div>
        <div class="metric-content">
          <h3>Active Amortizations</h3>
          <div class="metric-value">{{ metrics.active_amortizations || 0 }}</div>
        </div>
      </div>
    </div>

    <!-- Report Sections -->
    <div class="reports-section">
      <div class="report-tabs">
        <button
          :class="['tab-btn', { active: activeTab === 'assets' }]"
          @click="activeTab = 'assets'"
        >
          Asset Summary
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'journal' }]"
          @click="activeTab = 'journal'"
        >
          Journal Report
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'depreciation' }]"
          @click="activeTab = 'depreciation'"
        >
          Depreciation Schedule
        </button>
      </div>

      <!-- Asset Summary Tab -->
      <div v-if="activeTab === 'assets'" class="tab-content">
        <div class="report-controls">
          <div class="control-group">
            <label>CI Type Filter</label>
            <select v-model="assetFilters.ci_type_id" @change="loadAssets" class="control-select">
              <option value="">All Types</option>
            </select>
          </div>
          <div class="control-group">
            <label>Status Filter</label>
            <select v-model="assetFilters.status" @change="loadAssets" class="control-select">
              <option value="">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="active">Active</option>
              <option value="terminal">Terminal</option>
            </select>
          </div>
          <button @click="exportAssetReport" class="btn btn-outline">Export CSV</button>
        </div>

        <div class="report-table">
          <table class="data-table">
            <thead>
              <tr>
                <th>Asset Name</th>
                <th>Type</th>
                <th>Purchase Cost</th>
                <th>Current Book Value</th>
                <th>Accumulated Depreciation</th>
                <th>Monthly Depreciation</th>
                <th>Remaining Months</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="asset in assets" :key="asset.ci_id">
                <td>
                  <router-link :to="`/ci/${asset.ci_id}`" class="asset-link">
                    {{ asset.ci_name }}
                  </router-link>
                </td>
                <td>{{ asset.ci_type_name }}</td>
                <td class="amount">{{ formatCurrency(asset.purchase_cost) }}</td>
                <td class="amount">{{ formatCurrency(asset.current_book_value) }}</td>
                <td class="amount">{{ formatCurrency(asset.accumulated_depreciation) }}</td>
                <td class="amount">{{ formatCurrency(asset.monthly_depreciation) }}</td>
                <td>{{ asset.remaining_months }}</td>
                <td>
                  <span :class="`status-badge ${asset.status}`">
                    {{ asset.status }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Journal Report Tab -->
      <div v-if="activeTab === 'journal'" class="tab-content">
        <div class="report-controls">
          <div class="control-group">
            <label>Search CI Name</label>
            <input
              v-model="journalFilters.ci_name_search"
              type="text"
              class="control-input"
              placeholder="Search by name..."
              @keyup.enter="loadJournalReport"
            />
          </div>
          <div class="control-group">
            <label>CI Type</label>
            <select v-model="journalFilters.ci_type_id" @change="loadJournalReport" class="control-select">
              <option value="">All Types</option>
              <option v-for="ciType in ciTypes" :key="ciType.id" :value="ciType.id">
                {{ ciType.name }}
              </option>
            </select>
          </div>
          <div class="control-group">
            <label>Date From</label>
            <input v-model="journalFilters.date_from" type="date" class="control-input" />
          </div>
          <div class="control-group">
            <label>Date To</label>
            <input v-model="journalFilters.date_to" type="date" class="control-input" />
          </div>
          <button @click="loadJournalReport" class="btn btn-primary">Update Report</button>
          <button @click="exportJournalReport" class="btn btn-outline">Export CSV</button>
        </div>

        <div class="report-table">
          <table class="data-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Entry Type</th>
                <th>Asset</th>
                <th>Description</th>
                <th>Amount</th>
                <th>Book Value Before</th>
                <th>Book Value After</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="entry in journalEntries" :key="entry.id">
                <td>{{ formatDate(entry.entry_date) }}</td>
                <td>
              <span :class="getBadgeClasses(entry.entry_type)">
                {{ getBadgeLabel(entry.entry_type) }}
              </span>
            </td>
                <td>{{ entry.ci_name || 'N/A' }}</td>
                <td>{{ entry.description || '-' }}</td>
                <td class="amount">{{ formatCurrency(entry.amount) }}</td>
                <td class="amount">{{ formatCurrency(entry.book_value_before) }}</td>
                <td class="amount">{{ formatCurrency(entry.book_value_after) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Depreciation Schedule Tab -->
      <div v-if="activeTab === 'depreciation'" class="tab-content">
        <div class="schedule-info">
          <h3>Depreciation Schedule Overview</h3>
          <p>This shows the projected depreciation for all active amortizable assets over the next 12 months.</p>
        </div>

        <div class="schedule-chart">
          <h4>Monthly Depreciation Forecast</h4>
          <div class="chart-placeholder">
            <i class="fas fa-chart-bar"></i>
            <p>Chart visualization would be implemented here</p>
            <small>Total projected monthly depreciation: {{ formatCurrency(metrics.monthly_depreciation || 0) }}</small>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AssetSummary, AmortizationLedgerEntry, AmortizationMetrics } from '@/types/amortization'

const amortizationStore = useAmortizationStore()

const activeTab = ref('assets')
const loading = ref(false)
const metrics = ref<AmortizationMetrics>({
  total_amortizable_assets: 0,
  total_book_value: 0,
  monthly_depreciation: 0,
  active_amortizations: 0,
})

const assets = ref<AssetSummary[]>([])
const journalEntries = ref<AmortizationLedgerEntry[]>([])

const assetFilters = ref({
  ci_type_id: '',
  status: '',
})

const journalFilters = ref({
  ci_name_search: '',
  ci_type_id: '',
  date_from: '',
  date_to: '',
})

const ciTypes = ref<{id: string, name: string}[]>([])

const loadCITypes = async () => {
  try {
    const response = await fetch('/api/v1/ci-types')
    const data = await response.json()
    ciTypes.value = data.data || []
  } catch (error) {
    console.error('Failed to load CI types:', error)
  }
}

onMounted(async () => {
  await loadCITypes()
  await loadReportData()
})

const loadReportData = async () => {
  loading.value = true
  try {
    // Load metrics
    const metricsResponse = await amortizationStore.loadMetrics()
    if (metricsResponse) {
      metrics.value = metricsResponse
    }

    // Load assets
    await loadAssets()

    // Load journal entries for current month
    const now = new Date()
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
    const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)

    journalFilters.value.date_from = firstDay.toISOString().split('T')[0]
    journalFilters.value.date_to = lastDay.toISOString().split('T')[0]

    await loadJournalReport()
  } catch (error) {
    console.error('Failed to load report data:', error)
  } finally {
    loading.value = false
  }
}

const loadAssets = async () => {
  try {
    const response = await amortizationStore.loadAssetSummaries(assetFilters.value)
    // Store returns { cis: AssetSummary[], pagination: {...} }
    if (response && response.cis) {
      assets.value = response.cis
    }
  } catch (error) {
    console.error('Failed to load assets:', error)
  }
}

const loadJournalReport = async () => {
  try {
    const filters = {
      date_from: journalFilters.value.date_from,
      date_to: journalFilters.value.date_to,
      ci_name_search: journalFilters.value.ci_name_search || undefined,
      ci_type_id: journalFilters.value.ci_type_id || undefined,
      sort_by: 'entry_date',
      sort_order: 'desc' as const,
      limit: 100
    }
    const response = await amortizationStore.loadLedgerEntries(filters)
    if (response && response.entries) {
      journalEntries.value = response.entries
    }
  } catch (error) {
    console.error('Failed to load journal report:', error)
  }
}

const exportAssetReport = () => {
  const headers = [
    'Asset Name',
    'Type',
    'Purchase Cost',
    'Current Book Value',
    'Accumulated Depreciation',
    'Monthly Depreciation',
    'Remaining Months',
    'Status',
  ]

  const csvContent = [
    headers.join(','),
    ...assets.value.map(asset => [
      asset.ci_name,
      asset.ci_type_name,
      asset.purchase_cost.toString(),
      asset.current_book_value.toString(),
      asset.accumulated_depreciation.toString(),
      asset.monthly_depreciation.toString(),
      asset.remaining_months.toString(),
      asset.status,
    ].map(field => `"${field}"`).join(','))
  ].join('\n')

  downloadCSV(csvContent, `asset-summary-${new Date().toISOString().split('T')[0]}.csv`)
}

const exportJournalReport = () => {
  const headers = [
    'Date',
    'Entry Type',
    'Asset',
    'Description',
    'Amount',
    'Book Value Before',
    'Book Value After',
  ]

  const csvContent = [
    headers.join(','),
    ...journalEntries.value.map(entry => [
      formatDate(entry.entry_date),
      formatEntryType(entry.entry_type),
      entry.ci_name || '',
      entry.description || '',
      entry.amount.toString(),
      entry.book_value_before.toString(),
      entry.book_value_after.toString(),
    ].map(field => `"${field}"`).join(','))
  ].join('\n')

  downloadCSV(csvContent, `journal-report-${new Date().toISOString().split('T')[0]}.csv`)
}

const downloadCSV = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/csv' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  window.URL.revokeObjectURL(url)
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
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
</script>

<style scoped>
.amortization-reports {
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
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.metric-icon {
  width: 3rem;
  height: 3rem;
  background: #f3f4f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4f46e5;
}

.metric-content h3 {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 0.25rem;
  font-weight: 500;
}

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
}

.reports-section {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.report-tabs {
  display: flex;
  border-bottom: 1px solid #e5e7eb;
}

.tab-btn {
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  color: #6b7280;
  font-weight: 500;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #374151;
}

.tab-btn.active {
  color: #4f46e5;
  border-bottom-color: #4f46e5;
}

.tab-content {
  padding: 1.5rem;
}

.report-controls {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  align-items: end;
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.control-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.control-select,
.control-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
}

.report-table {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 800px;
}

.data-table th {
  background: #f9fafb;
  padding: 0.75rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.data-table td {
  padding: 0.75rem;
  border-bottom: 1px solid #f3f4f6;
}

.data-table tr:hover {
  background: #f9fafb;
}

.asset-link {
  color: #4f46e5;
  text-decoration: none;
}

.asset-link:hover {
  text-decoration: underline;
}

.amount {
  font-family: 'Courier New', monospace;
  text-align: right;
}

.status-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-badge.pending {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.active {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge.terminal {
  background: #fee2e2;
  color: #991b1b;
}

.schedule-info {
  margin-bottom: 2rem;
}

.schedule-info h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.schedule-info p {
  color: #6b7280;
}

.schedule-chart {
  background: #f9fafb;
  padding: 2rem;
  border-radius: 0.5rem;
  text-align: center;
}

.chart-placeholder {
  color: #6b7280;
}

.chart-placeholder i {
  font-size: 3rem;
  margin-bottom: 1rem;
  display: block;
}

.btn {
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-primary:hover {
  background: #4338ca;
}

.btn-outline {
  background: white;
  color: #6b7280;
  border-color: #d1d5db;
}

.btn-outline:hover {
  background: #f9fafb;
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

@media (max-width: 768px) {
  .amortization-reports {
    padding: 1rem;
  }

  .metrics-grid {
    grid-template-columns: 1fr;
  }

  .metric-card {
    flex-direction: column;
    text-align: center;
  }

  .report-controls {
    flex-direction: column;
    align-items: stretch;
  }

  .report-tabs {
    overflow-x: auto;
  }

  .data-table {
    font-size: 0.75rem;
  }
}
</style>
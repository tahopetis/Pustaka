<template>
  <div class="amortization-reports">
    <div class="page-header">
      <h1 class="page-title">Amortization Reports</h1>
      <p class="page-description">
        Financial reports and analysis for amortized assets
      </p>
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
            <div class="input-with-autocomplete">
              <input
                v-model="journalFilters.ci_name_search"
                type="text"
                class="control-input"
                placeholder="Search by name..."
                @input="onCISearchInputChange"
                @focus="handleCISearchFocus"
                @blur="hideCIAutocomplete"
              />
              <!-- Autocomplete dropdown -->
              <div
                v-if="showCIAutocomplete && (ciSearchResults.length > 0 || isCISearching)"
                class="autocomplete-dropdown"
              >
                <div v-if="isCISearching" class="autocomplete-loading">
                  Searching...
                </div>
                <div
                  v-for="result in ciSearchResults"
                  :key="result.id"
                  class="autocomplete-item"
                  @mousedown="selectCIResult(result)"
                >
                  <div class="ci-name">{{ result.name }}</div>
                  <div class="ci-type">{{ result.ci_type }}</div>
                </div>
              </div>
            </div>
          </div>
          <div class="control-group">
            <label>CI Type</label>
            <div class="input-with-autocomplete">
              <input
                :value="selectedCIType?.name || ''"
                type="text"
                class="control-input"
                placeholder="All Types"
                @input="onCITypeSearchInputChange"
                @focus="handleCITypeSearchFocus"
                @blur="hideCITypeAutocomplete"
              />
              <!-- Clear button when type is selected -->
              <button
                v-if="selectedCIType"
                @click="clearCIType"
                class="clear-btn"
                type="button"
              >×</button>
              <!-- Autocomplete dropdown -->
              <div
                v-if="showCITypeAutocomplete && (ciTypeSearchResults.length > 0 || isCITypeSearching)"
                class="autocomplete-dropdown"
              >
                <div v-if="isCITypeSearching" class="autocomplete-loading">
                  Searching...
                </div>
                <div
                  v-for="result in ciTypeSearchResults"
                  :key="result.id"
                  class="autocomplete-item"
                  @mousedown="selectCITypeResult(result)"
                >
                  {{ result.name }}
                </div>
              </div>
            </div>
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
        <!-- Filters -->
        <div class="schedule-filters">
          <div class="filter-row">
            <div class="control-group">
              <label>Date Range</label>
              <div class="date-range-inputs">
                <input
                  v-model="scheduleFilters.date_from"
                  type="month"
                  class="control-input"
                  max="9999-12"
                />
                <span class="date-separator">to</span>
                <input
                  v-model="scheduleFilters.date_to"
                  type="month"
                  class="control-input"
                  max="9999-12"
                />
              </div>
              <div class="preset-buttons">
                <button @click.prevent="setDatePreset(6)" class="preset-btn">Last 6 + Next 6</button>
                <button @click.prevent="setDatePreset(12)" class="preset-btn">Last 12 + Next 12</button>
                <button @click.prevent="setDatePreset(24)" class="preset-btn">Last 24 + Next 24</button>
              </div>
              <div v-if="dateRangeError" class="filter-error">{{ dateRangeError }}</div>
            </div>
            <div class="control-group">
              <label>CI Type</label>
              <select v-model="scheduleFilters.ci_type_id" @change="loadScheduleData" class="control-select">
                <option value="">All Types</option>
                <option v-for="ciType in ciTypes" :key="ciType.id" :value="ciType.id">
                  {{ ciType.name }}
                </option>
              </select>
            </div>
            <div class="control-group">
              <label>CI Name</label>
              <input
                v-model="scheduleFilters.ci_name"
                type="text"
                class="control-input"
                placeholder="Filter by name..."
                @keyup.enter="loadScheduleData"
              />
            </div>
            <button @click="loadScheduleData" class="btn btn-primary">Update Schedule</button>
            <button @click="exportScheduleReport" class="btn btn-outline">Export CSV</button>
          </div>
        </div>

        <!-- Summary Metrics Cards -->
        <div class="metrics-grid">
          <!-- OCC Card -->
          <div class="metric-card metric-card-gray">
            <div class="metric-icon">
              <i class="fas fa-shopping-cart"></i>
            </div>
            <div class="metric-content">
              <div class="metric-label">Original Capitalized Cost</div>
              <div class="metric-value">{{ formatCurrency(scheduleData?.total_original_cost || 0) }}</div>
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
              <div class="metric-value">{{ formatCurrency(scheduleData?.total_gross_book_value || 0) }}</div>
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
              <div class="metric-value">{{ formatCurrency(scheduleData?.total_net_book_value || 0) }}</div>
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
              <div class="metric-value">{{ formatCurrency(scheduleData?.total_accumulated_depreciation || scheduleData?.summary.total_depreciation || 0) }}</div>
              <div class="metric-sublabel">AD</div>
              <div class="metric-badge">{{ (scheduleData?.summary.depreciation_percentage || 0).toFixed(1) }}%</div>
            </div>
          </div>

          <!-- SV Card -->
          <div class="metric-card metric-card-red">
            <div class="metric-icon">
              <i class="fas fa-anchor"></i>
            </div>
            <div class="metric-content">
              <div class="metric-label">Salvage Value</div>
              <div class="metric-value">{{ formatCurrency(scheduleData?.total_salvage_value || 0) }}</div>
              <div class="metric-sublabel">SV</div>
            </div>
          </div>
        </div>

        <!-- Period-Specific Metrics Cards -->
        <div class="metrics-section">
          <h3 class="metrics-section-title">Period Metrics ({{ scheduleData?.period_summary?.opening_date || '' }} to {{ scheduleData?.period_summary?.closing_date || '' }})</h3>
          <div class="metrics-grid">
            <!-- Opening NBV Card -->
            <div class="metric-card metric-card-teal">
              <div class="metric-icon">
                <i class="fas fa-play-circle"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Opening Book Value</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.opening_book_value || 0) }}</div>
                <div class="metric-sublabel">Start of Period</div>
              </div>
            </div>

            <!-- Closing NBV Card -->
            <div class="metric-card metric-card-indigo">
              <div class="metric-icon">
                <i class="fas fa-stop-circle"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Closing Book Value</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.closing_book_value || 0) }}</div>
                <div class="metric-sublabel">End of Period</div>
              </div>
            </div>

            <!-- Period Depreciation Card -->
            <div class="metric-card metric-card-yellow">
              <div class="metric-icon">
                <i class="fas fa-chart-line"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Period Depreciation</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.period_depreciation || 0) }}</div>
                <div class="metric-sublabel">Total for Period</div>
                <div class="metric-badge">{{ scheduleData?.period_summary?.months_count || 0 }} months</div>
              </div>
            </div>

            <!-- Period Write-offs Card -->
            <div class="metric-card metric-card-pink">
              <div class="metric-icon">
                <i class="fas fa-times-circle"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Period Write-offs</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.period_write_offs || 0) }}</div>
                <div class="metric-sublabel">Write-offs in Period</div>
              </div>
            </div>

            <!-- Period Adjustments Card -->
            <div class="metric-card metric-card-cyan">
              <div class="metric-icon">
                <i class="fas fa-edit"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Period Adjustments</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.period_adjustments || 0) }}</div>
                <div class="metric-sublabel">Net Adjustments</div>
              </div>
            </div>

            <!-- Average Monthly Expense Card -->
            <div class="metric-card metric-card-lime">
              <div class="metric-icon">
                <i class="fas fa-calculator"></i>
              </div>
              <div class="metric-content">
                <div class="metric-label">Avg Monthly Depreciation</div>
                <div class="metric-value">{{ formatCurrency(scheduleData?.period_summary?.average_monthly_expense || 0) }}</div>
                <div class="metric-sublabel">Per Month</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Chart Section -->
        <div class="schedule-chart-section">
          <h3>Monthly Depreciation Trend</h3>
          <div v-if="scheduleLoading" class="loading-placeholder">
            Loading chart data...
          </div>
          <div v-else-if="hasValidScheduleData" class="chart-container">
            <svg ref="chartSvg" class="depreciation-chart"></svg>
          </div>
          <div v-else class="no-data">
            No schedule data available. Try adjusting the filters or selecting a different CI type.
          </div>
        </div>

        <!-- Data Table -->
        <div v-if="hasValidScheduleData" class="schedule-table-section">
          <h3>Monthly Schedule Details</h3>
          <div v-if="scheduleLoading" class="loading-placeholder">
            Loading table data...
          </div>
          <div v-else class="report-table">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Month</th>
                  <th>Type</th>
                  <th>Opening Value</th>
                  <th>GVB</th>
                  <th>Depreciation</th>
                  <th>Write-offs</th>
                  <th>Adjustments</th>
                  <th>Accum. Deprec.</th>
                  <th>Closing Value</th>
                  <th>Active Assets</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="entry in scheduleData.monthly_data" :key="entry.month">
                  <td>{{ formatMonth(entry.month) }}</td>
                  <td>
                    <span :class="entry.is_projected ? 'badge badge-blue' : 'badge badge-gray'">
                      {{ entry.is_projected ? 'Projected' : 'Actual' }}
                    </span>
                  </td>
                  <td class="amount">{{ formatCurrency(entry.opening_book_value) }}</td>
                  <td class="amount">{{ formatCurrency(entry.gross_book_value) }}</td>
                  <td class="amount">{{ formatCurrency(entry.depreciation_amount) }}</td>
                  <td class="amount">{{ formatCurrency(entry.write_off_amount) }}</td>
                  <td class="amount">{{ formatCurrency(entry.adjustment_amount) }}</td>
                  <td class="amount">{{ formatCurrency(entry.accumulated_depreciation) }}</td>
                  <td class="amount">{{ formatCurrency(entry.closing_book_value) }}</td>
                  <td>{{ entry.active_assets_count }}</td>
                </tr>
                <!-- Summary Row -->
                <tr v-if="scheduleData.monthly_data.length > 0" class="summary-row">
                  <td colspan="2"><strong>Period Totals</strong></td>
                  <td class="amount">-</td>
                  <td class="amount">-</td>
                  <td class="amount">{{ formatCurrency(scheduleData.period_summary.period_depreciation) }}</td>
                  <td class="amount">{{ formatCurrency(scheduleData.period_summary.period_write_offs) }}</td>
                  <td class="amount">{{ formatCurrency(scheduleData.period_summary.period_adjustments) }}</td>
                  <td class="amount">-</td>
                  <td class="amount">-</td>
                  <td>-</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import { ciAPI, ciTypeAPI } from '@/services/api'
import type {
  AssetSummary,
  AmortizationLedgerEntry,
  AmortizationMetrics,
  DepreciationScheduleResponse,
  MonthlyScheduleEntry
} from '@/types/amortization'

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

// Schedule state
const scheduleData = ref<DepreciationScheduleResponse | null>(null)
const scheduleLoading = ref(false)
const dateRangeError = ref('')
const scheduleFilters = ref({
  date_from: '',
  date_to: '',
  ci_type_id: '',
  ci_name: '',
})
const chartSvg = ref<SVGSVGElement | null>(null)

// Computed property to check if schedule has valid data for display
const hasValidScheduleData = computed(() => {
  if (!scheduleData.value?.monthly_data?.length) return false
  return scheduleData.value.monthly_data.some(d =>
    (d.opening_book_value || 0) > 0 ||
    (d.closing_book_value || 0) > 0 ||
    (d.depreciation_amount || 0) > 0
  )
})

const ciTypes = ref<{id: string, name: string}[]>([])

// Date range helper functions
const setDatePreset = (months: number) => {
  const now = new Date()
  const startDate = new Date(now.getFullYear(), now.getMonth() - months, 1)
  const endDate = new Date(now.getFullYear(), now.getMonth() + months, 1)

  scheduleFilters.value.date_from = startDate.toISOString().slice(0, 7) // YYYY-MM
  scheduleFilters.value.date_to = endDate.toISOString().slice(0, 7)
  dateRangeError.value = ''
  loadScheduleData()
}

const validateDateRange = (): boolean => {
  if (!scheduleFilters.value.date_from || !scheduleFilters.value.date_to) {
    dateRangeError.value = 'Please select both start and end dates'
    return false
  }

  const startDate = new Date(scheduleFilters.value.date_from + '-01')
  const endDate = new Date(scheduleFilters.value.date_to + '-01')

  if (startDate >= endDate) {
    dateRangeError.value = 'Start date must be before end date'
    return false
  }

  // Calculate the difference in months
  const monthsDiff = (endDate.getFullYear() - startDate.getFullYear()) * 12 +
    (endDate.getMonth() - startDate.getMonth())

  if (monthsDiff > 120) { // 10 years = 120 months
    dateRangeError.value = 'Date range cannot exceed 10 years'
    return false
  }

  dateRangeError.value = ''
  return true
}

// CI Name Autocomplete state
const ciSearchResults = ref<any[]>([])
const showCIAutocomplete = ref(false)
const isCISearching = ref(false)
const ciSearchTimeout = ref<NodeJS.Timeout | null>(null)

// CI Type Autocomplete state
const ciTypeSearchResults = ref<any[]>([])
const showCITypeAutocomplete = ref(false)
const isCITypeSearching = ref(false)
const ciTypeSearchTimeout = ref<NodeJS.Timeout | null>(null)
const selectedCIType = ref<{id: string, name: string} | null>(null)

const loadCITypes = async () => {
  try {
    const response = await ciTypeAPI.list()
    ciTypes.value = response.data.ci_types || []
  } catch (error) {
    console.error('Failed to load CI types:', error)
  }
}

// CI Name Autocomplete Methods
const onCISearchInputChange = () => {
  if (ciSearchTimeout.value) {
    clearTimeout(ciSearchTimeout.value)
  }

  showCIAutocomplete.value = true

  // Debounce search
  ciSearchTimeout.value = setTimeout(() => {
    fetchCISearchSuggestions()
  }, 300)
}

const handleCISearchFocus = () => {
  showCIAutocomplete.value = true
  // Load default 5 CIs when search is empty
  if (!journalFilters.value.ci_name_search.trim()) {
    fetchCISearchSuggestions()
  }
}

const hideCIAutocomplete = () => {
  // Delay hiding to allow click events to register
  setTimeout(() => {
    showCIAutocomplete.value = false
  }, 200)
}

const fetchCISearchSuggestions = async () => {
  isCISearching.value = true
  try {
    const params: any = {
      limit: 5
    }

    // Add search term if provided
    if (journalFilters.value.ci_name_search.trim()) {
      params.search = journalFilters.value.ci_name_search
    }

    const response = await ciAPI.list(params)
    ciSearchResults.value = response.data.cis || []
  } catch (error) {
    console.error('Failed to fetch CI suggestions:', error)
    ciSearchResults.value = []
  } finally {
    isCISearching.value = false
  }
}

const selectCIResult = (result: any) => {
  journalFilters.value.ci_name_search = result.name
  showCIAutocomplete.value = false
  ciSearchResults.value = []
  loadJournalReport()
}

// CI Type Autocomplete Methods
const onCITypeSearchInputChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.value

  // Update the selectedCIType name for display
  if (selectedCIType.value) {
    selectedCIType.value.name = value
  } else {
    selectedCIType.value = { id: '', name: value }
  }

  if (ciTypeSearchTimeout.value) {
    clearTimeout(ciTypeSearchTimeout.value)
  }

  showCITypeAutocomplete.value = true

  // Debounce search
  ciTypeSearchTimeout.value = setTimeout(() => {
    fetchCITypeSearchSuggestions()
  }, 300)
}

const handleCITypeSearchFocus = () => {
  showCITypeAutocomplete.value = true
  // Load all CI types when search is empty
  if (!selectedCIType.value?.name) {
    fetchCITypeSearchSuggestions()
  }
}

const hideCITypeAutocomplete = () => {
  setTimeout(() => {
    showCITypeAutocomplete.value = false
  }, 200)
}

const fetchCITypeSearchSuggestions = async () => {
  isCITypeSearching.value = true
  try {
    const searchTerm = selectedCIType.value?.name || ''

    const response = await ciTypeAPI.list({ limit: 100 })
    let types = response.data.ci_types || []

    // Filter by search term if provided
    if (searchTerm.trim()) {
      types = types.filter(t =>
        t.name.toLowerCase().includes(searchTerm.toLowerCase())
      )
    }

    ciTypeSearchResults.value = types
  } catch (error) {
    console.error('Failed to fetch CI type suggestions:', error)
    ciTypeSearchResults.value = []
  } finally {
    isCITypeSearching.value = false
  }
}

const selectCITypeResult = (result: any) => {
  selectedCIType.value = result
  journalFilters.value.ci_type_id = result.id
  showCITypeAutocomplete.value = false
  ciTypeSearchResults.value = []
  loadJournalReport()
}

const clearCIType = () => {
  selectedCIType.value = null
  journalFilters.value.ci_type_id = ''
  loadJournalReport()
}

onMounted(async () => {
  await loadCITypes()

  // Initialize default date range (Last 12 + Next 12 months)
  setDatePreset(12)

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

// Schedule helper functions
const formatMonth = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' })
}

const getLatestAssetCount = (): number => {
  if (!scheduleData.value?.monthly_data.length) return 0
  return scheduleData.value.monthly_data[scheduleData.value.monthly_data.length - 1].active_assets_count
}

const loadScheduleData = async () => {
  // Validate date range
  if (!validateDateRange()) {
    return
  }

  scheduleLoading.value = true
  try {
    const params: any = {
      date_from: scheduleFilters.value.date_from + '-01',
      date_to: scheduleFilters.value.date_to + '-01',
    }

    if (scheduleFilters.value.ci_type_id) {
      params.ci_type_ids = scheduleFilters.value.ci_type_id
    }

    if (scheduleFilters.value.ci_name) {
      // For CI name search, we'd need to implement search by name
      // For now, we'll skip this filter
    }

    const response = await amortizationStore.loadDepreciationSchedule(params)
    scheduleData.value = response
    // Note: Chart will be drawn by the scheduleData watcher
  } catch (error: any) {
    console.error('Failed to load schedule data:', error)
    // Clear old data on error to avoid showing stale data
    scheduleData.value = null
  } finally {
    scheduleLoading.value = false
  }
}

const exportScheduleReport = () => {
  if (!scheduleData.value) return

  const headers = [
    'Month',
    'Type',
    'Opening Value',
    'GVB',
    'Depreciation',
    'Write-offs',
    'Adjustments',
    'Accumulated Depreciation',
    'Closing Value',
    'Active Assets'
  ]

  const csvContent = [
    headers.join(','),
    ...scheduleData.value.monthly_data.map(entry => [
      formatMonth(entry.month),
      entry.is_projected ? 'Projected' : 'Actual',
      entry.opening_book_value.toString(),
      entry.gross_book_value.toString(),
      entry.depreciation_amount.toString(),
      entry.write_off_amount.toString(),
      entry.adjustment_amount.toString(),
      entry.accumulated_depreciation.toString(),
      entry.closing_book_value.toString(),
      entry.active_assets_count.toString()
    ].map(field => `"${field}"`).join(','))
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `depreciation-schedule-${new Date().toISOString().split('T')[0]}.csv`
  link.click()
  window.URL.revokeObjectURL(url)
}

// Watch for tab changes to load schedule data when switching to depreciation tab
watch(activeTab, (newTab) => {
  if (newTab === 'depreciation' && !scheduleData.value) {
    loadScheduleData()
  }
})

// Watch chartSvg ref to draw chart when it becomes available
watch(chartSvg, () => {
  if (chartSvg.value && hasValidScheduleData.value) {
    // SVG ref is now available, draw the chart
    nextTick(() => {
      requestAnimationFrame(() => {
        drawSimpleChart()
      })
    })
  }
})

// Watch schedule data changes to redraw chart (for filter changes)
watch(scheduleData, async () => {
  if (scheduleData.value?.monthly_data?.length > 0 && chartSvg.value) {
    // Chart already exists, just redraw it
    requestAnimationFrame(() => {
      drawSimpleChart()
    })
  }
}, { deep: true })

const drawSimpleChart = () => {
  if (!chartSvg.value) {
    console.warn('Chart SVG ref is not available')
    return
  }

  if (!scheduleData.value?.monthly_data?.length) {
    console.warn('No monthly data available')
    return
  }

  const svg = chartSvg.value
  const data = scheduleData.value.monthly_data

  // Check if data has valid values (not all zeros)
  const hasValidData = data.some(d =>
    (d.opening_book_value || 0) > 0 ||
    (d.closing_book_value || 0) > 0 ||
    (d.depreciation_amount || 0) > 0
  )

  if (!hasValidData) {
    console.warn('No valid chart data available (all values are 0)')
    // Clear the chart and hide it
    while (svg.firstChild) {
      svg.removeChild(svg.firstChild)
    }
    return
  }

  // Get container dimensions for proper sizing
  const container = svg.parentElement
  const containerWidth = container ? container.clientWidth : 800
  const width = Math.max(containerWidth - 40, 800) // Account for padding
  const height = 400
  const padding = { top: 30, right: 40, bottom: 80, left: 80 }

  // Clear existing content
  while (svg.firstChild) {
    svg.removeChild(svg.firstChild)
  }

  // Set SVG dimensions explicitly
  svg.style.width = width + 'px'
  svg.style.height = height + 'px'
  svg.setAttribute('width', width.toString())
  svg.setAttribute('height', height.toString())
  svg.setAttribute('viewBox', `0 0 ${width} ${height}`)

  // Calculate scales with validation to prevent NaN
  const validValues = data.flatMap(d => [
    d.opening_book_value || 0,
    d.closing_book_value || 0
  ]).filter(v => isFinite(v) && v > 0)

  const maxValue = validValues.length > 0
    ? Math.max(...validValues) * 1.1
    : 100 // fallback default

  if (!isFinite(maxValue) || maxValue <= 0) {
    console.warn('Invalid max value for chart, skipping render')
    return
  }

  const validDepreciation = data
    .map(d => d.depreciation_amount || 0)
    .filter(v => isFinite(v) && v > 0)

  const maxDepreciation = validDepreciation.length > 0
    ? Math.max(...validDepreciation)
    : 1 // fallback default
  const chartWidth = width - padding.left - padding.right
  const chartHeight = height - padding.top - padding.bottom

  // Create namespace
  const ns = 'http://www.w3.org/2000/svg'

  // Find the transition point (last actual month)
  let lastActualIndex = -1
  data.forEach((entry, index) => {
    if (!entry.is_projected) lastActualIndex = index
  })

  // Draw background shading for historical vs forecast
  if (lastActualIndex >= 0 && lastActualIndex < data.length - 1) {
    const transitionX = padding.left + ((lastActualIndex + 0.5) / (data.length - 1)) * chartWidth

    // Historical background (blue tint)
    const historicalBg = document.createElementNS(ns, 'rect')
    historicalBg.setAttribute('x', padding.left.toString())
    historicalBg.setAttribute('y', padding.top.toString())
    historicalBg.setAttribute('width', (transitionX - padding.left).toString())
    historicalBg.setAttribute('height', chartHeight.toString())
    historicalBg.setAttribute('fill', 'rgba(59, 130, 246, 0.05)')
    svg.appendChild(historicalBg)

    // Forecast background (gray tint)
    const forecastBg = document.createElementNS(ns, 'rect')
    forecastBg.setAttribute('x', transitionX.toString())
    forecastBg.setAttribute('y', padding.top.toString())
    forecastBg.setAttribute('width', (width - padding.right - transitionX).toString())
    forecastBg.setAttribute('height', chartHeight.toString())
    forecastBg.setAttribute('fill', 'rgba(107, 114, 128, 0.05)')
    svg.appendChild(forecastBg)

    // Vertical division line
    const dividerLine = document.createElementNS(ns, 'line')
    dividerLine.setAttribute('x1', transitionX.toString())
    dividerLine.setAttribute('y1', padding.top.toString())
    dividerLine.setAttribute('x2', transitionX.toString())
    dividerLine.setAttribute('y2', (height - padding.bottom).toString())
    dividerLine.setAttribute('stroke', '#9ca3af')
    dividerLine.setAttribute('stroke-width', '2')
    dividerLine.setAttribute('stroke-dasharray', '5,5')
    svg.appendChild(dividerLine)

    // Add labels for historical/forecast
    const historicalLabel = document.createElementNS(ns, 'text')
    historicalLabel.setAttribute('x', (padding.left + 10).toString())
    historicalLabel.setAttribute('y', (padding.top - 10).toString())
    historicalLabel.setAttribute('font-size', '11')
    historicalLabel.setAttribute('font-weight', '600')
    historicalLabel.setAttribute('fill', '#3b82f6')
    historicalLabel.textContent = 'HISTORICAL'
    svg.appendChild(historicalLabel)

    const forecastLabel = document.createElementNS(ns, 'text')
    forecastLabel.setAttribute('x', (transitionX + 10).toString())
    forecastLabel.setAttribute('y', (padding.top - 10).toString())
    forecastLabel.setAttribute('font-size', '11')
    forecastLabel.setAttribute('font-weight', '600')
    forecastLabel.setAttribute('fill', '#6b7280')
    forecastLabel.textContent = 'FORECAST'
    svg.appendChild(forecastLabel)
  }

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

  // Calculate values for reference lines
  const totalSalvageValue = scheduleData.value?.total_salvage_value || 0
  const totalOriginalCost = scheduleData.value?.total_original_cost || 0  // OCC
  const totalGrossBookValue = scheduleData.value?.total_gross_book_value || 0  // GVB

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
    occLine.setAttribute('stroke-dasharray', '12,4,4,4')  // Long dash pattern
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

  // Draw salvage value reference line (if significant)
  if (totalSalvageValue > 0 && totalSalvageValue < maxValue) {
    const salvageY = padding.top + chartHeight - (totalSalvageValue / maxValue) * chartHeight

    const salvageLine = document.createElementNS(ns, 'line')
    salvageLine.setAttribute('x1', padding.left.toString())
    salvageLine.setAttribute('y1', salvageY.toString())
    salvageLine.setAttribute('x2', (width - padding.right).toString())
    salvageLine.setAttribute('y2', salvageY.toString())
    salvageLine.setAttribute('stroke', '#ef4444')
    salvageLine.setAttribute('stroke-width', '2')
    salvageLine.setAttribute('stroke-dasharray', '8,4')
    svg.appendChild(salvageLine)

    // Salvage value label
    const salvageLabel = document.createElementNS(ns, 'text')
    salvageLabel.setAttribute('x', (width - padding.right - 10).toString())
    salvageLabel.setAttribute('y', (salvageY - 8).toString())
    salvageLabel.setAttribute('text-anchor', 'end')
    salvageLabel.setAttribute('font-size', '10')
    salvageLabel.setAttribute('font-weight', '600')
    salvageLabel.setAttribute('fill', '#ef4444')
    salvageLabel.textContent = `Salvage: ${formatCurrency(totalSalvageValue)}`
    svg.appendChild(salvageLabel)
  }

  // Draw axes
  const xAxis = document.createElementNS(ns, 'line')
  xAxis.setAttribute('x1', padding.left.toString())
  xAxis.setAttribute('y1', (height - padding.bottom).toString())
  xAxis.setAttribute('x2', (width - padding.right).toString())
  xAxis.setAttribute('y2', (height - padding.bottom).toString())
  xAxis.setAttribute('stroke', '#9ca3af')
  xAxis.setAttribute('stroke-width', '2')
  svg.appendChild(xAxis)

  const yAxis = document.createElementNS(ns, 'line')
  yAxis.setAttribute('x1', padding.left.toString())
  yAxis.setAttribute('y1', padding.top.toString())
  yAxis.setAttribute('x2', padding.left.toString())
  yAxis.setAttribute('y2', (height - padding.bottom).toString())
  yAxis.setAttribute('stroke', '#9ca3af')
  yAxis.setAttribute('stroke-width', '2')
  svg.appendChild(yAxis)

  // Add Y-axis labels (currency)
  for (let i = 0; i <= 4; i++) {
    const value = maxValue * (1 - i / 4)
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

  // Draw monthly depreciation bars
  const barWidth = (chartWidth / data.length) * 0.6
  data.forEach((entry, index) => {
    const xCenter = padding.left + (index / (data.length - 1)) * chartWidth
    const barHeight = (entry.depreciation_amount / maxValue) * chartHeight * 0.3  // Scale to 30% of chart height
    const y = height - padding.bottom - barHeight

    if (barHeight > 0) {
      const bar = document.createElementNS(ns, 'rect')
      bar.setAttribute('x', (xCenter - barWidth / 2).toString())
      bar.setAttribute('y', y.toString())
      bar.setAttribute('width', barWidth.toString())
      bar.setAttribute('height', barHeight.toString())
      bar.setAttribute('fill', entry.is_projected ? 'rgba(16, 185, 129, 0.3)' : 'rgba(59, 130, 246, 0.3)')
      bar.setAttribute('rx', '2')
      svg.appendChild(bar)
    }
  })

  // Draw book value line
  let pathD = ''
  let pathDArea = ''

  data.forEach((entry, index) => {
    const x = padding.left + (index / (data.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (entry.closing_book_value / maxValue) * chartHeight

    if (index === 0) {
      pathD += `M ${x} ${y}`
      pathDArea += `M ${x} ${height - padding.bottom} L ${x} ${y}`
    } else {
      pathD += ` L ${x} ${y}`
      pathDArea += ` L ${x} ${y}`
    }
  })

  // Close the area for fill
  const lastIndex = data.length - 1
  const lastX = padding.left + chartWidth
  pathDArea += ` L ${lastX} ${height - padding.bottom} Z`

  // Draw area fill under the line
  const area = document.createElementNS(ns, 'path')
  area.setAttribute('d', pathDArea)
  area.setAttribute('fill', 'url(#gradient-fill)')
  area.setAttribute('opacity', '0.1')
  svg.appendChild(area)

  // Define gradient
  const defs = document.createElementNS(ns, 'defs')
  defs.innerHTML = `
    <linearGradient id="gradient-fill" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#3b82f6;stop-opacity:0.3" />
      <stop offset="100%" style="stop-color:#3b82f6;stop-opacity:0" />
    </linearGradient>
  `
  svg.appendChild(defs)

  // Draw the book value line
  const line = document.createElementNS(ns, 'path')
  line.setAttribute('d', pathD)
  line.setAttribute('fill', 'none')
  line.setAttribute('stroke', '#3b82f6')
  line.setAttribute('stroke-width', '3')
  line.setAttribute('stroke-linecap', 'round')
  line.setAttribute('stroke-linejoin', 'round')
  svg.appendChild(line)

  // Draw accumulated depreciation line
  let accumulatedPathD = ''
  data.forEach((entry, index) => {
    const x = padding.left + (index / (data.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (entry.accumulated_depreciation / maxValue) * chartHeight

    if (index === 0) {
      accumulatedPathD += `M ${x} ${y}`
    } else {
      accumulatedPathD += ` L ${x} ${y}`
    }
  })

  const accumulatedLine = document.createElementNS(ns, 'path')
  accumulatedLine.setAttribute('d', accumulatedPathD)
  accumulatedLine.setAttribute('fill', 'none')
  accumulatedLine.setAttribute('stroke', '#f97316') // Orange color
  accumulatedLine.setAttribute('stroke-width', '2')
  accumulatedLine.setAttribute('stroke-linecap', 'round')
  accumulatedLine.setAttribute('stroke-linejoin', 'round')
  accumulatedLine.setAttribute('stroke-dasharray', '4,2') // Dashed line to distinguish from book value
  svg.appendChild(accumulatedLine)

  // Draw data points with tooltips
  data.forEach((entry, index) => {
    const x = padding.left + (index / (data.length - 1)) * chartWidth
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
    circle.setAttribute('fill', entry.is_projected ? '#10b981' : '#3b82f6')
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
    tooltipBg.setAttribute('width', '140')
    tooltipBg.setAttribute('height', '55')
    tooltipBg.setAttribute('rx', '4')
    tooltipBg.setAttribute('fill', '#1f2937')
    tooltipBg.setAttribute('opacity', '0.95')
    tooltipGroup.appendChild(tooltipBg)

    // Tooltip text
    const tooltipTitle = document.createElementNS(ns, 'text')
    tooltipTitle.setAttribute('x', (x + 18).toString())
    tooltipTitle.setAttribute('y', (y - 43).toString())
    tooltipTitle.setAttribute('font-size', '11')
    tooltipTitle.setAttribute('font-weight', '600')
    tooltipTitle.setAttribute('fill', '#fff')
    tooltipTitle.textContent = formatMonth(entry.month)
    tooltipGroup.appendChild(tooltipTitle)

    const tooltipType = document.createElementNS(ns, 'text')
    tooltipType.setAttribute('x', (x + 18).toString())
    tooltipType.setAttribute('y', (y - 28).toString())
    tooltipType.setAttribute('font-size', '9')
    tooltipType.setAttribute('fill', entry.is_projected ? '#10b981' : '#93c5fd')
    tooltipType.textContent = entry.is_projected ? '● Projected' : '● Actual'
    tooltipGroup.appendChild(tooltipType)

    const tooltipValue = document.createElementNS(ns, 'text')
    tooltipValue.setAttribute('x', (x + 18).toString())
    tooltipValue.setAttribute('y', (y - 14).toString())
    tooltipValue.setAttribute('font-size', '10')
    tooltipValue.setAttribute('fill', '#e5e7eb')
    tooltipValue.textContent = formatCurrency(entry.closing_book_value)
    tooltipGroup.appendChild(tooltipValue)

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

  // Draw accumulated depreciation data points with tooltips
  data.forEach((entry, index) => {
    const x = padding.left + (index / (data.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (entry.accumulated_depreciation / maxValue) * chartHeight

    // Create group for interactivity
    const group = document.createElementNS(ns, 'g')
    group.style.cursor = 'pointer'

    // Outer circle for hover effect
    const outerCircle = document.createElementNS(ns, 'circle')
    outerCircle.setAttribute('cx', x.toString())
    outerCircle.setAttribute('cy', y.toString())
    outerCircle.setAttribute('r', '6')
    outerCircle.setAttribute('fill', 'transparent')
    outerCircle.setAttribute('class', 'chart-point-hover')
    group.appendChild(outerCircle)

    // Inner circle (smaller than book value points)
    const circle = document.createElementNS(ns, 'circle')
    circle.setAttribute('cx', x.toString())
    circle.setAttribute('cy', y.toString())
    circle.setAttribute('r', '4')
    circle.setAttribute('fill', entry.is_projected ? '#fb923c' : '#f97316')
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
    tooltipBg.setAttribute('width', '160')
    tooltipBg.setAttribute('height', '55')
    tooltipBg.setAttribute('rx', '4')
    tooltipBg.setAttribute('fill', '#1f2937')
    tooltipBg.setAttribute('opacity', '0.95')
    tooltipGroup.appendChild(tooltipBg)

    // Tooltip text
    const tooltipTitle = document.createElementNS(ns, 'text')
    tooltipTitle.setAttribute('x', (x + 18).toString())
    tooltipTitle.setAttribute('y', (y - 43).toString())
    tooltipTitle.setAttribute('font-size', '11')
    tooltipTitle.setAttribute('font-weight', '600')
    tooltipTitle.setAttribute('fill', '#fff')
    tooltipTitle.textContent = formatMonth(entry.month)
    tooltipGroup.appendChild(tooltipTitle)

    const tooltipType = document.createElementNS(ns, 'text')
    tooltipType.setAttribute('x', (x + 18).toString())
    tooltipType.setAttribute('y', (y - 28).toString())
    tooltipType.setAttribute('font-size', '9')
    tooltipType.setAttribute('fill', '#fb923c')
    tooltipType.textContent = '● Accumulated Deprec.'
    tooltipGroup.appendChild(tooltipType)

    const tooltipValue = document.createElementNS(ns, 'text')
    tooltipValue.setAttribute('x', (x + 18).toString())
    tooltipValue.setAttribute('y', (y - 14).toString())
    tooltipValue.setAttribute('font-size', '10')
    tooltipValue.setAttribute('fill', '#e5e7eb')
    tooltipValue.textContent = formatCurrency(entry.accumulated_depreciation)
    tooltipGroup.appendChild(tooltipValue)

    // Show tooltip on hover
    group.addEventListener('mouseenter', () => {
      tooltipGroup.style.opacity = '1'
      outerCircle.setAttribute('fill', 'rgba(249, 115, 22, 0.15)')
    })

    group.addEventListener('mouseleave', () => {
      tooltipGroup.style.opacity = '0'
      outerCircle.setAttribute('fill', 'transparent')
    })

    group.appendChild(tooltipGroup)
    svg.appendChild(group)
  })

  // Add X-axis labels (months)
  data.forEach((entry, index) => {
    if (index % Math.ceil(data.length / 8) === 0) {
      const x = padding.left + (index / (data.length - 1)) * chartWidth
      const text = document.createElementNS(ns, 'text')
      text.setAttribute('x', x.toString())
      text.setAttribute('y', (height - padding.bottom + 20).toString())
      text.setAttribute('text-anchor', 'middle')
      text.setAttribute('font-size', '10')
      text.setAttribute('fill', '#6b7280')
      text.textContent = formatMonth(entry.month)
      svg.appendChild(text)
    }
  })

  // Add legend
  const legendY = height - 25
  let legendX = padding.left

  // OCC (Gray dashed)
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

  // GVB (Purple dashed)
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

  // Book value legend
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

  // Monthly Depreciation legend
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

  legendX += 140

  // Salvage value legend (reuse totalSalvageValue from above)
  if (totalSalvageValue > 0 && totalSalvageValue < maxValue) {
    const legendSalvageLine = document.createElementNS(ns, 'line')
    legendSalvageLine.setAttribute('x1', legendX.toString())
    legendSalvageLine.setAttribute('y1', legendY.toString())
    legendSalvageLine.setAttribute('x2', (legendX + 30).toString())
    legendSalvageLine.setAttribute('y2', legendY.toString())
    legendSalvageLine.setAttribute('stroke', '#ef4444')
    legendSalvageLine.setAttribute('stroke-width', '2')
    legendSalvageLine.setAttribute('stroke-dasharray', '8,4')
    svg.appendChild(legendSalvageLine)

    const legendText3 = document.createElementNS(ns, 'text')
    legendText3.setAttribute('x', (legendX + 38).toString())
    legendText3.setAttribute('y', (legendY + 4).toString())
    legendText3.setAttribute('font-size', '11')
    legendText3.setAttribute('fill', '#6b7280')
    legendText3.textContent = 'Salvage Value'
    svg.appendChild(legendText3)

    legendX += 110
  }

  // Accumulated depreciation legend
  const legendAccumulatedLine = document.createElementNS(ns, 'line')
  legendAccumulatedLine.setAttribute('x1', legendX.toString())
  legendAccumulatedLine.setAttribute('y1', legendY.toString())
  legendAccumulatedLine.setAttribute('x2', (legendX + 30).toString())
  legendAccumulatedLine.setAttribute('y2', legendY.toString())
  legendAccumulatedLine.setAttribute('stroke', '#f97316')
  legendAccumulatedLine.setAttribute('stroke-width', '2')
  legendAccumulatedLine.setAttribute('stroke-dasharray', '4,2')
  svg.appendChild(legendAccumulatedLine)

  const legendText4 = document.createElementNS(ns, 'text')
  legendText4.setAttribute('x', (legendX + 38).toString())
  legendText4.setAttribute('y', (legendY + 4).toString())
  legendText4.setAttribute('font-size', '11')
  legendText4.setAttribute('fill', '#6b7280')
  legendText4.textContent = 'Accumulated Depreciation'
  svg.appendChild(legendText4)
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

/* New period-specific card colors */
.metric-card-teal::before { background: #14b8a6; }
.metric-card-indigo::before { background: #6366f1; }
.metric-card-yellow::before { background: #eab308; }
.metric-card-pink::before { background: #ec4899; }
.metric-card-cyan::before { background: #06b6d4; }
.metric-card-lime::before { background: #84cc16; }

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
.metric-card-teal .metric-icon { background: #ccfbf1; color: #14b8a6; }
.metric-card-indigo .metric-icon { background: #e0e7ff; color: #6366f1; }
.metric-card-yellow .metric-icon { background: #fef9c3; color: #ca8a04; }
.metric-card-pink .metric-icon { background: #fce7f3; color: #ec4899; }
.metric-card-cyan .metric-icon { background: #cffafe; color: #06b6d4; }
.metric-card-lime .metric-icon { background: #ecfccb; color: #84cc16; }

.metrics-section {
  margin-bottom: 2rem;
}

.metrics-section-title {
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 1rem;
}

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

.input-with-autocomplete {
  position: relative;
}

.input-with-autocomplete .control-input {
  width: 100%;
  padding-right: 32px;
}

.clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border: none;
  background: #d1d5db;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  color: #6b7280;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.clear-btn:hover {
  background: #9ca3af;
  color: #1f2937;
}

.autocomplete-dropdown {
  position: absolute;
  z-index: 100;
  width: 100%;
  left: 0;
  top: 100%;
  margin-top: 4px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-height: 240px;
  overflow: auto;
}

.autocomplete-loading {
  padding: 0.75rem;
  color: #6b7280;
  font-size: 0.875rem;
  text-align: center;
}

.autocomplete-item {
  padding: 0.75rem;
  cursor: pointer;
  border-bottom: 1px solid #f3f4f6;
}

.autocomplete-item:last-child {
  border-bottom: none;
}

.autocomplete-item:hover {
  background: #f9fafb;
}

.autocomplete-item .ci-name {
  font-weight: 500;
  color: #1f2937;
}

.autocomplete-item .ci-type {
  font-size: 0.875rem;
  color: #6b7280;
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

.summary-row {
  background: #f0fdf4 !important;
  font-weight: 600;
  border-top: 2px solid #16a34a;
}

.summary-row:hover {
  background: #f0fdf4 !important;
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

/* Schedule-specific styles */
.schedule-summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.summary-card {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.summary-card h4 {
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.summary-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
}

.schedule-filters {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 1.5rem;
}

.filter-row {
  display: flex;
  gap: 1rem;
  align-items: end;
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 150px;
}

.control-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.25rem;
}

.control-select,
.control-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
}

.date-range-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-range-inputs .control-input {
  flex: 1;
}

.date-separator {
  color: #6b7280;
  font-weight: 500;
}

.preset-buttons {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
}

.preset-btn {
  padding: 0.375rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background: white;
  font-size: 0.75rem;
  font-weight: 500;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
}

.preset-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.filter-error {
  color: #dc2626;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.schedule-chart-section {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 1.5rem;
}

.schedule-chart-section h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
}

.chart-container {
  width: 100%;
  overflow-x: auto;
  min-height: 400px;
}

.depreciation-chart {
  width: 100%;
  min-height: 400px;
  display: block;
}

.loading-placeholder {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.schedule-table-section {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.schedule-table-section h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
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
</style>
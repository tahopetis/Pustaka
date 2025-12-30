<template>
  <div class="amortization-timeline">
    <!-- Statistics Summary -->
    <div v-if="!loading && statistics" class="timeline-stats">
      <div class="stat-card">
        <div class="stat-icon stat-icon-blue">
          <i class="fas fa-calendar-check"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.total_entries }}</div>
          <div class="stat-label">Total Entries</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-green">
          <i class="fas fa-arrow-down"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatCurrency(statistics.total_depreciation) }}</div>
          <div class="stat-label">Total Depreciation</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-purple">
          <i class="fas fa-edit"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ statistics.adjustments_count }}</div>
          <div class="stat-label">Adjustments</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-orange">
          <i class="fas fa-book"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatCurrency(statistics.current_book_value) }}</div>
          <div class="stat-label">Current Book Value</div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-8">
      <div class="spinner w-8 h-8 mx-auto mb-4"></div>
      <p class="text-gray-500">Loading amortization journey...</p>
    </div>

    <!-- No Data State -->
    <div v-else-if="!entries || entries.length === 0" class="text-center py-8">
      <i class="fas fa-history text-4xl text-gray-300 mb-4"></i>
      <p class="text-gray-500">No amortization history available</p>
      <p class="text-sm text-gray-400 mt-2">Amortization entries will appear here once processing begins</p>
    </div>

    <!-- Timeline -->
    <div v-else class="timeline-container">
      <div class="timeline-header">
        <h3>Amortization Journey</h3>
        <button @click="exportTimeline" class="btn btn-outline btn-sm">
          <i class="fas fa-download mr-2"></i>
          Export CSV
        </button>
      </div>

      <div class="timeline">
        <div
          v-for="(entry, index) in entries"
          :key="entry.id"
          class="timeline-entry"
          :class="{ 'timeline-entry-first': index === 0 }"
        >
          <!-- Timeline dot -->
          <div class="timeline-dot">
            <i :class="getEntryIcon(entry.entry_type)"></i>
          </div>

          <!-- Timeline content -->
          <div class="timeline-content">
            <div class="entry-card">
              <!-- Entry header -->
              <div class="entry-header">
                <div class="entry-type">
                  <span :class="getBadgeClasses(entry.entry_type)">
                    {{ getBadgeLabel(entry.entry_type) }}
                  </span>
                </div>
                <div class="entry-date">
                  <i class="fas fa-clock mr-1"></i>
                  {{ formatDate(entry.entry_date) }}
                </div>
              </div>

              <!-- Entry details -->
              <div v-if="entry.description" class="entry-description">
                {{ entry.description }}
              </div>

              <!-- Financial impact -->
              <div class="entry-financials">
                <div class="financial-row">
                  <span class="financial-label">Amount:</span>
                  <span class="financial-value" :class="getAmountClass(entry.entry_type)">
                    {{ formatCurrency(entry.amount) }}
                  </span>
                </div>
                <div class="financial-row">
                  <span class="financial-label">Book Value Before:</span>
                  <span class="financial-value">{{ formatCurrency(entry.book_value_before) }}</span>
                </div>
                <div class="financial-row">
                  <span class="financial-label">Book Value After:</span>
                  <span class="financial-value financial-highlight">
                    {{ formatCurrency(entry.book_value_after) }}
                  </span>
                </div>
                <div class="financial-row">
                  <span class="financial-label">Accumulated Depreciation:</span>
                  <span class="financial-value">{{ formatCurrency(entry.accumulated_depreciation) }}</span>
                </div>
              </div>

              <!-- Metadata -->
              <div class="entry-meta">
                <span class="meta-item">
                  <i class="fas fa-user mr-1"></i>
                  {{ entry.created_by || 'System' }}
                </span>
                <span class="meta-item">
                  <i class="fas fa-clock mr-1"></i>
                  {{ formatDateTime(entry.created_at) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Timeline line (except for last entry) -->
          <div v-if="index < entries.length - 1" class="timeline-line"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationLedgerEntry } from '@/types/amortization'

interface Props {
  ciId: string
}

const props = defineProps<Props>()
const amortizationStore = useAmortizationStore()

const loading = ref(true)
const entries = ref<AmortizationLedgerEntry[]>([])

// Statistics computed from entries
const statistics = computed(() => {
  if (!entries.value || entries.value.length === 0) return null

  const totalDepreciation = entries.value.reduce((sum, entry) => {
    if (entry.entry_type === 'monthly_depreciation' ||
        entry.entry_type === 'catch_up_depreciation' ||
        entry.entry_type === 'depreciation') {
      return sum + Math.abs(entry.amount)
    }
    return sum
  }, 0)

  const adjustmentsCount = entries.value.filter(
    entry => entry.entry_type === 'adjustment'
  ).length

  const latestEntry = entries.value[0] // Entries are sorted by date desc
  const currentBookValue = latestEntry?.book_value_after || 0

  return {
    total_entries: entries.value.length,
    total_depreciation: totalDepreciation,
    adjustments_count: adjustmentsCount,
    current_book_value: currentBookValue
  }
})

const loadTimeline = async () => {
  loading.value = true
  try {
    const response = await amortizationStore.loadLedgerEntries({
      ci_id: props.ciId,
      sort_by: 'entry_date',
      sort_order: 'desc',
      limit: 1000 // Get all entries for this CI
    })
    if (response && response.entries) {
      entries.value = response.entries
    }
  } catch (error) {
    console.error('Failed to load amortization timeline:', error)
  } finally {
    loading.value = false
  }
}

const exportTimeline = () => {
  if (!entries.value || entries.value.length === 0) return

  const headers = [
    'Date',
    'Entry Type',
    'Description',
    'Amount',
    'Book Value Before',
    'Book Value After',
    'Accumulated Depreciation',
    'Created By',
    'Created At'
  ]

  const csvContent = [
    headers.join(','),
    ...entries.value.map(entry => [
      formatDate(entry.entry_date),
      formatEntryType(entry.entry_type),
      entry.description || '',
      entry.amount.toString(),
      entry.book_value_before.toString(),
      entry.book_value_after.toString(),
      entry.accumulated_depreciation.toString(),
      entry.created_by || 'System',
      formatDateTime(entry.created_at)
    ].map(field => `"${field}"`).join(','))
  ].join('\n')

  downloadCSV(csvContent, `amortization-journey-${props.ciId}-${new Date().toISOString().split('T')[0]}.csv`)
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
    currency: 'USD'
  }).format(amount)
}

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString()
}

const formatDateTime = (dateString: string): string => {
  return new Date(dateString).toLocaleString()
}

const formatEntryType = (entryType: string): string => {
  return entryType.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

const getBadgeLabel = (entryType: string): string => {
  const labels = {
    'depreciation': 'Monthly Depreciation',
    'monthly_depreciation': 'Monthly Depreciation',
    'catch_up_depreciation': 'Catch-Up Depreciation',
    'adjustment': 'Manual Adjustment',
    'write_off': 'Write-Off',
    'reversal': 'Reversal',
    'restructuring': 'Restructuring'
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
    'reversal': 'badge badge-yellow',
    'restructuring': 'badge badge-orange'
  }
  return classes[entryType] || 'badge badge-gray'
}

const getEntryIcon = (entryType: string): string => {
  const icons = {
    'depreciation': 'fas fa-calendar',
    'monthly_depreciation': 'fas fa-calendar',
    'catch_up_depreciation': 'fas fa-bolt',
    'adjustment': 'fas fa-edit',
    'write_off': 'fas fa-times-circle',
    'reversal': 'fas fa-undo',
    'restructuring': 'fas fa-exchange-alt'
  }
  return icons[entryType] || 'fas fa-circle'
}

const getAmountClass = (entryType: string): string => {
  if (entryType === 'adjustment') {
    return 'financial-value-neutral'
  }
  return 'financial-value-negative'
}

onMounted(() => {
  loadTimeline()
})

watch(() => props.ciId, () => {
  if (props.ciId) {
    loadTimeline()
  }
})
</script>

<style scoped>
.amortization-timeline {
  padding: 1.5rem;
}

/* Statistics */
.timeline-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: white;
  padding: 1rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.stat-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.stat-icon-blue {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon-green {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.stat-icon-purple {
  background: linear-gradient(135deg, #8e2de2 0%, #4a00e0 100%);
}

.stat-icon-orange {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1f2937;
}

.stat-label {
  font-size: 0.75rem;
  color: #6b7280;
  margin-top: 0.125rem;
}

/* Timeline Container */
.timeline-container {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.timeline-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.timeline {
  padding: 1.5rem;
}

/* Timeline Entry */
.timeline-entry {
  display: flex;
  position: relative;
}

.timeline-entry-first {
  margin-top: 0;
}

.timeline-dot {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 1rem;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
  z-index: 2;
}

.timeline-line {
  position: absolute;
  left: 1.5rem;
  top: 3rem;
  bottom: -1.5rem;
  width: 2px;
  background: linear-gradient(to bottom, #e5e7eb 0%, #f3f4f6 100%);
  transform: translateX(-50%);
}

.timeline-content {
  flex: 1;
  padding-left: 1.5rem;
  padding-bottom: 2rem;
}

.timeline-entry:last-child .timeline-content {
  padding-bottom: 0;
}

/* Entry Card */
.entry-card {
  background: #f9fafb;
  border-radius: 0.5rem;
  padding: 1rem;
  border: 1px solid #e5e7eb;
  transition: all 0.2s;
}

.entry-card:hover {
  background: #ffffff;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  transform: translateY(-2px);
}

.entry-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.entry-date {
  font-size: 0.875rem;
  color: #6b7280;
  display: flex;
  align-items: center;
}

.entry-description {
  font-size: 0.875rem;
  color: #374151;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
}

/* Financial Impact */
.entry-financials {
  background: white;
  border-radius: 0.375rem;
  padding: 0.75rem;
  margin-bottom: 0.75rem;
}

.financial-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.25rem 0;
  font-size: 0.875rem;
}

.financial-label {
  color: #6b7280;
  font-weight: 500;
}

.financial-value {
  color: #1f2937;
  font-family: 'Courier New', monospace;
  font-weight: 600;
}

.financial-highlight {
  color: #4f46e5;
  font-size: 1rem;
}

.financial-value-negative {
  color: #dc2626;
}

.financial-value-neutral {
  color: #f59e0b;
}

/* Metadata */
.entry-meta {
  display: flex;
  gap: 1rem;
  font-size: 0.75rem;
  color: #9ca3af;
}

.meta-item {
  display: flex;
  align-items: center;
}

/* Badges */
.badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
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

.badge-yellow {
  background: #fef3c7;
  color: #92400e;
}

.badge-orange {
  background: #fed7aa;
  color: #c2410c;
}

/* Buttons */
.btn {
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
}

.btn-outline {
  background: white;
  color: #6b7280;
  border-color: #d1d5db;
}

.btn-outline:hover {
  background: #f9fafb;
  color: #374151;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
}

/* Responsive */
@media (max-width: 768px) {
  .timeline-stats {
    grid-template-columns: 1fr;
  }

  .timeline-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .timeline-content {
    padding-left: 1rem;
  }

  .entry-financials {
    font-size: 0.75rem;
  }

  .entry-meta {
    flex-direction: column;
    gap: 0.25rem;
  }
}
</style>

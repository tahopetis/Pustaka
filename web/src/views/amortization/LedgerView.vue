<template>
  <div class="amortization-ledger">
    <div class="page-header">
      <h1 class="page-title">Amortization Ledger</h1>
      <p class="page-description">
        Complete audit trail of all amortization transactions
      </p>
    </div>

    <!-- Filters -->
    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label for="entryType">Entry Type</label>
          <select
            id="entryType"
            v-model="filters.entry_type"
            multiple
            class="filter-select"
          >
            <option value="monthly_depreciation">Monthly Depreciation</option>
            <option value="catch_up_depreciation">Catch-Up Depreciation</option>
            <option value="depreciation">Depreciation</option>
            <option value="adjustment">Adjustment</option>
            <option value="write_off">Write Off</option>
            <option value="reversal">Reversal</option>
          </select>
        </div>

        <div class="filter-group">
          <label for="dateFrom">Date From</label>
          <input
            id="dateFrom"
            v-model="filters.date_from"
            type="date"
            class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label for="dateTo">Date To</label>
          <input
            id="dateTo"
            v-model="filters.date_to"
            type="date"
            class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label for="sortBy">Sort By</label>
          <select id="sortBy" v-model="filters.sort_by" class="filter-select">
            <option value="entry_date">Entry Date</option>
            <option value="created_at">Created At</option>
            <option value="amount">Amount</option>
          </select>
        </div>

        <div class="filter-group">
          <label for="sortOrder">Sort Order</label>
          <select id="sortOrder" v-model="filters.sort_order" class="filter-select">
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </div>

        <div class="filter-actions">
          <button @click="applyFilters" class="btn btn-primary">
            Apply Filters
          </button>
          <button @click="clearFilters" class="btn btn-secondary">
            Clear
          </button>
          <button @click="exportLedger" class="btn btn-outline">
            Export
          </button>
        </div>
      </div>
    </div>

    <!-- Results Summary -->
    <div class="results-summary">
      <span v-if="pagination.total">
        Showing {{ ledgerEntries.length }} of {{ pagination.total }} entries
      </span>
      <span v-else>No entries found</span>
    </div>

    <!-- Ledger Table -->
    <div class="ledger-table">
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading ledger entries...
      </div>
      <div v-else-if="ledgerEntries.length === 0" class="no-data">
        No ledger entries found matching the current filters.
      </div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>Date</th>
            <th>CI Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Amount</th>
            <th>Book Value Before</th>
            <th>Book Value After</th>
            <th>Accumulated Depreciation</th>
            <th>Created By</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in ledgerEntries" :key="entry.id">
            <td>{{ formatDate(entry.entry_date) }}</td>
            <td>
              <router-link
                v-if="entry.ci_id"
                :to="`/ci/${entry.ci_id}`"
                class="ci-link"
              >
                {{ entry.ci_name || 'N/A' }}
              </router-link>
              <span v-else>N/A</span>
            </td>
            <td>
              <span :class="getBadgeClasses(entry.entry_type)">
                {{ getBadgeLabel(entry.entry_type) }}
              </span>
            </td>
            <td>{{ entry.description || '-' }}</td>
            <td class="amount">{{ formatCurrency(entry.amount) }}</td>
            <td class="amount">{{ formatCurrency(entry.book_value_before) }}</td>
            <td class="amount">{{ formatCurrency(entry.book_value_after) }}</td>
            <td class="amount">{{ formatCurrency(entry.accumulated_depreciation) }}</td>
            <td>{{ entry.created_by_name || entry.created_by || 'System' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="pagination.has_next || pagination.has_prev" class="pagination">
      <button
        :disabled="!pagination.has_prev"
        @click="previousPage"
        class="btn btn-outline"
      >
        Previous
      </button>
      <span class="page-info">
        Page {{ Math.floor(pagination.offset / pagination.limit) + 1 }}
      </span>
      <button
        :disabled="!pagination.has_next"
        @click="nextPage"
        class="btn btn-outline"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationLedgerEntry, AmortizationLedgerFilters } from '@/types/amortization'

const amortizationStore = useAmortizationStore()

const loading = ref(false)
const ledgerEntries = ref<AmortizationLedgerEntry[]>([])
const pagination = ref({
  total: 0,
  limit: 50,
  offset: 0,
  has_next: false,
  has_prev: false,
})

const filters = ref<AmortizationLedgerFilters>({
  entry_type: [],
  date_from: '',
  date_to: '',
  sort_by: 'entry_date',
  sort_order: 'desc',
  limit: 50,
  offset: 0,
})

onMounted(() => {
  loadLedgerEntries()
})

const loadLedgerEntries = async () => {
  loading.value = true
  try {
    const response = await amortizationStore.loadLedgerEntries(filters.value)
    ledgerEntries.value = amortizationStore.ledgerEntries
    pagination.value = {
      total: response.total_count,
      limit: response.page_size,
      offset: ((response.page - 1) * response.page_size),
      has_prev: response.page > 1,
      has_next: response.page < response.total_pages
    }
  } catch (error) {
    console.error('Failed to load ledger entries:', error)
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.value.offset = 0
  filters.value.offset = 0
  loadLedgerEntries()
}

const clearFilters = () => {
  filters.value = {
    entry_type: [],
    date_from: '',
    date_to: '',
    sort_by: 'entry_date',
    sort_order: 'desc',
    limit: 50,
    offset: 0,
  }
  pagination.value.offset = 0
  loadLedgerEntries()
}

const previousPage = () => {
  if (pagination.value.has_prev) {
    filters.value.offset = Math.max(0, filters.value.offset! - pagination.value.limit)
    pagination.value.offset = filters.value.offset
    loadLedgerEntries()
  }
}

const nextPage = () => {
  if (pagination.value.has_next) {
    filters.value.offset = filters.value.offset! + pagination.value.limit
    pagination.value.offset = filters.value.offset
    loadLedgerEntries()
  }
}

const exportLedger = () => {
  // Create CSV content
  const headers = [
    'Date',
    'CI Name',
    'Entry Type',
    'Description',
    'Amount',
    'Book Value Before',
    'Book Value After',
    'Accumulated Depreciation',
  ]

  const csvContent = [
    headers.join(','),
    ...ledgerEntries.value.map(entry => [
      formatDate(entry.entry_date),
      entry.ci_name || '',
      formatEntryType(entry.entry_type),
      entry.description || '',
      entry.amount.toString(),
      entry.book_value_before.toString(),
      entry.book_value_after.toString(),
      entry.accumulated_depreciation.toString(),
    ].map(field => `"${field}"`).join(','))
  ].join('\n')

  // Download CSV file
  const blob = new Blob([csvContent], { type: 'text/csv' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `amortization-ledger-${new Date().toISOString().split('T')[0]}.csv`
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
.amortization-ledger {
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

.filters-section {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 1.5rem;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  align-items: end;
}

.filter-group {
  display: flex;
  flex-direction: column;
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.25rem;
}

.filter-select,
.filter-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
}

.filter-actions {
  display: flex;
  gap: 0.5rem;
  align-items: end;
}

.results-summary {
  margin-bottom: 1rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.ledger-table {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.loading,
.no-data {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
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

.ci-link {
  color: #4f46e5;
  text-decoration: none;
}

.ci-link:hover {
  text-decoration: underline;
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

.amount {
  font-family: 'Courier New', monospace;
  text-align: right;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.page-info {
  color: #6b7280;
  font-size: 0.875rem;
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

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #4338ca;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
}

.btn-outline {
  background: white;
  color: #6b7280;
  border-color: #d1d5db;
}

.btn-outline:hover:not(:disabled) {
  background: #f9fafb;
}

@media (max-width: 768px) {
  .amortization-ledger {
    padding: 1rem;
  }

  .filters-grid {
    grid-template-columns: 1fr;
  }

  .data-table {
    font-size: 0.75rem;
  }

  .data-table th,
  .data-table td {
    padding: 0.5rem;
  }
}
</style>
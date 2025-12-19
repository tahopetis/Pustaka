<template>
  <div class="amortization-dashboard">
    <div class="page-header">
      <h1 class="page-title">Amortization Dashboard</h1>
      <p class="page-description">
        Overview of asset amortization and depreciation tracking
      </p>
    </div>

    <!-- Key Metrics -->
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-coins"></i>
        </div>
        <div class="metric-content">
          <h3>Total Amortizable Assets</h3>
          <div class="metric-value">{{ metrics.total_amortizable_assets }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="metric-content">
          <h3>Total Book Value</h3>
          <div class="metric-value">{{ formatCurrency(metrics.total_book_value) }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-calculator"></i>
        </div>
        <div class="metric-content">
          <h3>Monthly Depreciation</h3>
          <div class="metric-value">{{ formatCurrency(metrics.monthly_depreciation) }}</div>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon">
          <i class="fas fa-history"></i>
        </div>
        <div class="metric-content">
          <h3>Active Amortizations</h3>
          <div class="metric-value">{{ metrics.active_amortizations }}</div>
        </div>
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
          <div class="activity-icon">
            <i :class="getActivityIcon(activity.entry_type)"></i>
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
import { ref, onMounted } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationLedgerEntry, AmortizationMetrics } from '@/types/amortization'

const amortizationStore = useAmortizationStore()

const loading = ref(true)
const metrics = ref<AmortizationMetrics>({
  total_amortizable_assets: 0,
  total_book_value: 0,
  monthly_depreciation: 0,
  active_amortizations: 0,
})
const recentActivity = ref<AmortizationLedgerEntry[]>([])

onMounted(async () => {
  await loadDashboardData()
  loading.value = false
})

const loadDashboardData = async () => {
  try {
    // Load metrics
    const summaryResponse = await amortizationStore.loadMetrics()
    if (summaryResponse) {
      metrics.value = summaryResponse
    }

    // Load recent activity
    const ledgerResponse = await amortizationStore.loadLedgerEntries({
      limit: 10,
      sort_by: 'entry_date',
      sort_order: 'desc'
    })
    if (ledgerResponse && ledgerResponse.data) {
      recentActivity.value = ledgerResponse.data || []
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

const getActivityIcon = (entryType: string): string => {
  switch (entryType) {
    case 'depreciation':
      return 'fas fa-chart-line'
    case 'adjustment':
      return 'fas fa-edit'
    case 'write_off':
      return 'fas fa-times-circle'
    case 'reversal':
      return 'fas fa-undo'
    default:
      return 'fas fa-circle'
  }
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

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
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

.activity-icon {
  width: 2rem;
  height: 2rem;
  background: #f3f4f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4f46e5;
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
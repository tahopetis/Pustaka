<template>
  <div class="amortization-scheduler">
    <div class="page-header">
      <h1 class="page-title">Amortization Scheduler</h1>
      <p class="page-description">
        Manage automated amortization processing runs
      </p>
    </div>

    <!-- Scheduler Status -->
    <div class="status-card">
      <h2>Scheduler Status</h2>
      <div v-if="loadingStatus" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading scheduler status...
      </div>
      <div v-else class="status-content">
        <div class="status-grid">
          <div class="status-item">
            <label>Current Status</label>
            <div class="status-value">
              <span :class="`status-indicator ${schedulerStatus?.is_running ? 'running' : 'idle'}`">
                {{ schedulerStatus?.is_running ? 'Running' : 'Idle' }}
              </span>
            </div>
          </div>
          <div class="status-item">
            <label>Last Run</label>
            <div class="status-value">
              {{ schedulerStatus?.last_run_date ? formatDate(schedulerStatus.last_run_date) : 'Never' }}
            </div>
          </div>
          <div class="status-item">
            <label>Next Run</label>
            <div class="status-value">
              {{ schedulerStatus?.next_run_date ? formatDate(schedulerStatus.next_run_date) : 'Scheduled' }}
            </div>
          </div>
          <div class="status-item">
            <label>Run Frequency</label>
            <div class="status-value">
              {{ schedulerStatus?.run_frequency || 'Daily at 00:00 UTC' }}
            </div>
          </div>
        </div>

        <div class="scheduler-actions">
          <button
            @click="runScheduler"
            class="btn btn-primary"
            :disabled="schedulerStatus?.is_running || running"
          >
            <i v-if="running" class="fas fa-spinner fa-spin"></i>
            {{ running ? 'Running...' : 'Run Scheduler Now' }}
          </button>
          <button @click="loadSchedulerStatus" class="btn btn-secondary">
            Refresh Status
          </button>
        </div>
      </div>
    </div>

    <!-- Recent Runs -->
    <div class="runs-section">
      <h2>Recent Scheduler Runs</h2>
      <div v-if="loadingRuns" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading recent runs...
      </div>
      <div v-else-if="schedulerRuns.length === 0" class="no-data">
        No scheduler runs found.
      </div>
      <div v-else class="runs-table">
        <table class="data-table">
          <thead>
            <tr>
              <th>Run Date</th>
              <th>Status</th>
              <th>Started</th>
              <th>Completed</th>
              <th>Assets Processed</th>
              <th>Successful Entries</th>
              <th>Failed Entries</th>
              <th>Duration</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in schedulerRuns" :key="run.id">
              <td>{{ formatDate(run.run_date) }}</td>
              <td>
                <span :class="`status-badge ${run.status}`">
                  {{ run.status }}
                </span>
              </td>
              <td>{{ formatDateTime(run.started_at) }}</td>
              <td>{{ formatDateTime(run.completed_at) }}</td>
              <td>{{ run.total_assets_processed }}</td>
              <td>{{ run.successful_entries }}</td>
              <td>{{ run.failed_entries }}</td>
              <td>{{ calculateDuration(run.started_at, run.completed_at) }}</td>
              <td>
                <button
                  @click="viewRunDetails(run.id)"
                  class="btn btn-sm btn-outline"
                >
                  Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Run Details Modal -->
    <div v-if="selectedRun" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>Run Details</h3>
          <button @click="closeModal" class="btn-close">&times;</button>
        </div>
        <div class="modal-body">
          <div class="run-details">
            <div class="detail-group">
              <label>Run ID</label>
              <div>{{ selectedRun.id }}</div>
            </div>
            <div class="detail-group">
              <label>Status</label>
              <div>
                <span :class="`status-badge ${selectedRun.status}`">
                  {{ selectedRun.status }}
                </span>
              </div>
            </div>
            <div class="detail-group">
              <label>Run Date</label>
              <div>{{ formatDate(selectedRun.run_date) }}</div>
            </div>
            <div class="detail-group">
              <label>Started At</label>
              <div>{{ formatDateTime(selectedRun.started_at) }}</div>
            </div>
            <div class="detail-group">
              <label>Completed At</label>
              <div>{{ formatDateTime(selectedRun.completed_at) }}</div>
            </div>
            <div class="detail-group">
              <label>Duration</label>
              <div>{{ calculateDuration(selectedRun.started_at, selectedRun.completed_at) }}</div>
            </div>
            <div class="detail-group">
              <label>Total Assets Processed</label>
              <div>{{ selectedRun.total_assets_processed }}</div>
            </div>
            <div class="detail-group">
              <label>Successful Entries</label>
              <div>{{ selectedRun.successful_entries }}</div>
            </div>
            <div class="detail-group">
              <label>Failed Entries</label>
              <div>{{ selectedRun.failed_entries }}</div>
            </div>
            <div v-if="selectedRun.error_message" class="detail-group">
              <label>Error Message</label>
              <div class="error-message">{{ selectedRun.error_message }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationRun, SchedulerStatus } from '@/types/amortization'

const amortizationStore = useAmortizationStore()

const loadingStatus = ref(false)
const loadingRuns = ref(false)
const running = ref(false)
const schedulerStatus = ref<SchedulerStatus | null>(null)
const schedulerRuns = ref<AmortizationRun[]>([])
const selectedRun = ref<AmortizationRun | null>(null)

onMounted(() => {
  loadSchedulerStatus()
  loadSchedulerRuns()
})

const loadSchedulerStatus = async () => {
  loadingStatus.value = true
  try {
    const response = await amortizationStore.loadSchedulerStatus()
    if (response.data) {
      schedulerStatus.value = response.data
    }
  } catch (error) {
    console.error('Failed to load scheduler status:', error)
  } finally {
    loadingStatus.value = false
  }
}

const loadSchedulerRuns = async () => {
  loadingRuns.value = true
  try {
    const response = await amortizationStore.loadSchedulerRuns()
    if (response.data) {
      schedulerRuns.value = response.data.data
    }
  } catch (error) {
    console.error('Failed to load scheduler runs:', error)
  } finally {
    loadingRuns.value = false
  }
}

const runScheduler = async () => {
  running.value = true
  try {
    const response = await amortizationStore.runScheduler()
    if (response.data) {
      // Refresh status and runs
      await loadSchedulerStatus()
      await loadSchedulerRuns()
    }
  } catch (error) {
    console.error('Failed to run scheduler:', error)
  } finally {
    running.value = false
  }
}

const viewRunDetails = (runId: string) => {
  const run = schedulerRuns.value.find(r => r.id === runId)
  if (run) {
    selectedRun.value = run
  }
}

const closeModal = () => {
  selectedRun.value = null
}

const formatDate = (dateString?: string): string => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleDateString()
}

const formatDateTime = (dateString?: string): string => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleString()
}

const calculateDuration = (started?: string, completed?: string): string => {
  if (!started) return 'N/A'

  const startTime = new Date(started)
  const endTime = completed ? new Date(completed) : new Date()
  const duration = endTime.getTime() - startTime.getTime()

  if (duration < 1000) {
    return `${duration}ms`
  } else if (duration < 60000) {
    return `${Math.round(duration / 1000)}s`
  } else {
    return `${Math.round(duration / 60000)}m`
  }
}
</script>

<style scoped>
.amortization-scheduler {
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

.status-card {
  background: white;
  padding: 2rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.status-card h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #6b7280;
}

.status-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.status-item label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.25rem;
}

.status-value {
  font-size: 1rem;
  color: #1f2937;
}

.status-indicator {
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-indicator.running {
  background: #dbeafe;
  color: #1e40af;
}

.status-indicator.idle {
  background: #f3f4f6;
  color: #6b7280;
}

.scheduler-actions {
  display: flex;
  gap: 1rem;
}

.runs-section {
  background: white;
  padding: 2rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.runs-section h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.no-data {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 0.5rem;
}

.runs-table {
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

.status-badge.running {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge.completed {
  background: #d1fae5;
  color: #065f46;
}

.status-badge.failed {
  background: #fee2e2;
  color: #991b1b;
}

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
  gap: 0.5rem;
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

.btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-close:hover {
  color: #1f2937;
}

.modal-body {
  padding: 1.5rem;
}

.run-details {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
}

.detail-group > div {
  font-size: 0.875rem;
  color: #1f2937;
}

.error-message {
  color: #991b1b;
  background: #fef2f2;
  padding: 0.75rem;
  border-radius: 0.375rem;
  border: 1px solid #fecaca;
}

@media (max-width: 768px) {
  .amortization-scheduler {
    padding: 1rem;
  }

  .status-grid {
    grid-template-columns: 1fr;
  }

  .scheduler-actions {
    flex-direction: column;
  }

  .data-table {
    font-size: 0.75rem;
    min-width: 600px;
  }
}
</style>
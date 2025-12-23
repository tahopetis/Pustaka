<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="modal-overlay" @click.self="close">
        <div class="modal-container" @click.stop>
          <div class="modal-header">
            <h2 class="modal-title">Create Amortization Adjustment</h2>
            <button @click="close" class="modal-close">
              <i class="fas fa-times"></i>
            </button>
          </div>

          <div class="modal-body">
            <form @submit.prevent="handleSubmit" class="adjustment-form">
              <!-- CI Selection -->
              <div class="form-group">
                <label for="ciSelect" class="form-label">
                  Configuration Item <span class="required">*</span>
                </label>
                <select
                  id="ciSelect"
                  v-model="form.ci_id"
                  class="form-select"
                  :class="{ 'is-invalid': errors.ci_id }"
                  required
                  @change="handleCIChange"
                >
                  <option value="">Select an asset...</option>
                  <option
                    v-for="asset in amortizableAssets"
                    :key="asset.id || asset.ci_id"
                    :value="asset.id || asset.ci_id"
                  >
                    {{ asset.name || asset.ci_name }} ({{ asset.ci_type || asset.ci_type_name || 'N/A' }})
                    - Current Book Value: {{ formatCurrency(asset.current_book_value) }}
                  </option>
                </select>
                <div v-if="errors.ci_id" class="invalid-feedback">
                  {{ errors.ci_id }}
                </div>
              </div>

              <!-- CI Details (shown when CI is selected) -->
              <div v-if="selectedCI" class="ci-details">
                <div class="detail-row">
                  <span class="detail-label">Purchase Cost:</span>
                  <span class="detail-value">{{ formatCurrency(selectedCI.purchase_cost) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Current Book Value:</span>
                  <span class="detail-value">{{ formatCurrency(selectedCI.current_book_value) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Accumulated Depreciation:</span>
                  <span class="detail-value">{{ formatCurrency(selectedCI.accumulated_depreciation) }}</span>
                </div>
              </div>

              <!-- Adjustment Amount -->
              <div class="form-group">
                <label for="amount" class="form-label">
                  Adjustment Amount <span class="required">*</span>
                </label>
                <div class="input-group">
                  <span class="input-prefix">$</span>
                  <input
                    id="amount"
                    v-model.number="form.amount"
                    type="number"
                    step="0.01"
                    class="form-input"
                    :class="{ 'is-invalid': errors.amount }"
                    placeholder="0.00"
                    required
                  />
                </div>
                <div v-if="errors.amount" class="invalid-feedback">
                  {{ errors.amount }}
                </div>
                <small class="form-hint">
                  Enter positive value to increase book value, negative to decrease.
                  New book value will be:
                  <strong>{{ formatCurrency(newBookValue) }}</strong>
                </small>
              </div>

              <!-- Effective Date -->
              <div class="form-group">
                <label for="effectiveDate" class="form-label">Effective Date</label>
                <input
                  id="effectiveDate"
                  v-model="form.effective_date"
                  type="date"
                  class="form-input"
                />
                <small class="form-hint">Leave empty for today's date</small>
              </div>

              <!-- Description -->
              <div class="form-group">
                <label for="description" class="form-label">
                  Description <span class="required">*</span>
                </label>
                <textarea
                  id="description"
                  v-model="form.description"
                  class="form-textarea"
                  :class="{ 'is-invalid': errors.description }"
                  rows="3"
                  placeholder="Explain the reason for this adjustment..."
                  required
                ></textarea>
                <div v-if="errors.description" class="invalid-feedback">
                  {{ errors.description }}
                </div>
              </div>

              <!-- Form Actions -->
              <div class="form-actions">
                <button type="button" @click="close" class="btn btn-secondary">
                  Cancel
                </button>
                <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
                  <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
                  <span v-if="isSubmitting">Creating...</span>
                  <span v-else>Create Adjustment</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AssetSummary } from '@/types/amortization'

interface Props {
  isOpen: boolean
  ciId?: string
}

interface Emits {
  (e: 'close'): void
  (e: 'created'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const amortizationStore = useAmortizationStore()

// Form state
const form = ref({
  ci_id: props.ciId || '',
  amount: 0,
  effective_date: '',
  description: ''
})

const errors = ref<Record<string, string>>({})
const isSubmitting = ref(false)
const amortizableAssets = ref<AssetSummary[]>([])

// Computed
const selectedCI = computed(() => {
  if (!form.value.ci_id) return null
  return amortizableAssets.value.find(a => (a.id || a.ci_id) === form.value.ci_id)
})

const newBookValue = computed(() => {
  if (!selectedCI.value) return 0
  return selectedCI.value.current_book_value + (form.value.amount || 0)
})

// Load amortizable assets
const loadAssets = async () => {
  try {
    const response = await amortizationStore.loadAssetSummaries({
      has_financial_data: true,
      limit: 1000
    })
    if (response?.cis) {
      amortizableAssets.value = response.cis.filter(
        asset => asset.current_book_value !== undefined && asset.purchase_cost > 0
      )
    }
  } catch (error) {
    console.error('Failed to load amortizable assets:', error)
  }
}

// Watch for modal open
watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    loadAssets()
    if (props.ciId) {
      form.value.ci_id = props.ciId
    }
  }
})

// Methods
const close = () => {
  emit('close')
  // Reset form after transition
  setTimeout(() => {
    form.value = {
      ci_id: props.ciId || '',
      amount: 0,
      effective_date: '',
      description: ''
    }
    errors.value = {}
  }, 300)
}

const handleCIChange = () => {
  // Clear amount when CI changes
  form.value.amount = 0
  errors.value = {}
}

const validateForm = (): boolean => {
  errors.value = {}

  if (!form.value.ci_id) {
    errors.value.ci_id = 'Please select a configuration item'
  }

  if (form.value.amount === 0 || !form.value.amount) {
    errors.value.amount = 'Please enter an adjustment amount'
  }

  if (!form.value.description?.trim()) {
    errors.value.description = 'Please provide a description'
  }

  // Validate new book value won't be negative
  if (selectedCI.value && newBookValue.value < 0) {
    errors.value.amount = `Adjustment would result in negative book value (${formatCurrency(newBookValue.value)})`
  }

  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }

  isSubmitting.value = true

  try {
    await amortizationStore.createAdjustment(form.value.ci_id, {
      ci_id: form.value.ci_id,
      entry_type: 'adjustment',
      amount: form.value.amount,
      description: form.value.description,
      entry_date: form.value.effective_date || new Date().toISOString().split('T')[0]
    })

    emit('created')
    close()
  } catch (error: any) {
    console.error('Failed to create adjustment:', error)
    if (error.response?.data?.message) {
      errors.value.amount = error.response.data.message
    } else {
      errors.value.amount = 'Failed to create adjustment. Please try again.'
    }
  } finally {
    isSubmitting.value = false
  }
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount || 0)
}
</script>

<style scoped>
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
  padding: 1rem;
}

.modal-container {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.375rem;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
}

.adjustment-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
}

.form-select,
.form-input,
.form-textarea {
  padding: 0.625rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-select:focus,
.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.form-select.is-invalid,
.form-input.is-invalid,
.form-textarea.is-invalid {
  border-color: #ef4444;
}

.invalid-feedback {
  font-size: 0.75rem;
  color: #ef4444;
}

.form-hint {
  font-size: 0.75rem;
  color: #6b7280;
}

.input-group {
  display: flex;
  align-items: stretch;
}

.input-prefix {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  background: #f9fafb;
  border: 1px solid #d1d5db;
  border-right: none;
  border-radius: 0.375rem 0 0 0.375rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.input-group .form-input {
  border-radius: 0 0.375rem 0.375rem 0;
  flex: 1;
}

.ci-details {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.detail-label {
  color: #6b7280;
}

.detail-value {
  font-weight: 500;
  color: #1f2937;
  font-family: 'Courier New', monospace;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding-top: 0.5rem;
  border-top: 1px solid #e5e7eb;
  margin-top: 0.5rem;
}

.btn {
  padding: 0.625rem 1rem;
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

.btn-secondary {
  background: white;
  color: #374151;
  border-color: #d1d5db;
}

.btn-secondary:hover:not(:disabled) {
  background: #f9fafb;
  border-color: #9ca3af;
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #4338ca;
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95) translateY(-10px);
}

.modal-enter-to .modal-container,
.modal-leave-from .modal-container {
  transform: scale(1) translateY(0);
}

@media (max-width: 640px) {
  .modal-container {
    max-height: 95vh;
  }

  .modal-header,
  .modal-body {
    padding: 1rem;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions .btn {
    width: 100%;
  }
}
</style>

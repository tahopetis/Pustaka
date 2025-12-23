<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import { amortizationApi } from '@/services/amortizationApi'
import type { AssetSummary, RestructuringCalculation } from '@/types/amortization'

interface Props {
  show: boolean
  ciId?: string
}

interface Emits {
  (e: 'close'): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const amortizationStore = useAmortizationStore()

const selectedCIId = ref<string>('')
const newUsefulLifeMonths = ref<number>(60)
const reason = ref<string>('')
const loading = ref(false)
const previewLoading = ref(false)
const error = ref<string | null>(null)
const preview = ref<RestructuringCalculation | null>(null)
const effectiveDate = ref<string>(new Date().toISOString().split('T')[0])
const amortizableAssets = ref<AssetSummary[]>([])

// Computed
const selectedCI = computed(() => {
  if (!selectedCIId.value) return null
  return amortizableAssets.value.find(a => (a.id || a.ci_id) === selectedCIId.value)
})

const isValid = computed(() => {
  return selectedCIId.value && newUsefulLifeMonths.value > 0 && reason.value.trim().length > 0
})

// Load amortizable assets when dialog opens
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
  } catch (err) {
    console.error('Failed to load amortizable assets:', err)
  }
}

// Reset form when dialog opens/closes
watch(() => props.show, (isOpen) => {
  if (isOpen) {
    loadAssets()
    selectedCIId.value = props.ciId || ''
    reason.value = ''
    preview.value = null
    error.value = null
    effectiveDate.value = new Date().toISOString().split('T')[0]
    if (props.ciId) {
      // Load preview for pre-selected CI
      loadPreview()
    }
  }
})

// When CI changes, reset and load new preview
watch(selectedCIId, () => {
  preview.value = null
  if (selectedCI.value) {
    newUsefulLifeMonths.value = Math.max(1, selectedCI.value.useful_life_months + 12)
    loadPreview()
  }
})

const monthlyChangeFormatted = computed(() => {
  if (!preview.value) return ''
  const change = preview.value.monthly_depreciation_change
  const sign = change >= 0 ? '+' : ''
  return `${sign}$${change.toFixed(2)}`
})

const monthlyChangeClass = computed(() => {
  if (!preview.value) return ''
  return preview.value.monthly_depreciation_change <= 0 ? 'text-green-600' : 'text-red-600'
})

const percentChangeFormatted = computed(() => {
  if (!preview.value) return ''
  const sign = preview.value.percent_change >= 0 ? '+' : ''
  return `${sign}${preview.value.percent_change.toFixed(1)}%`
})

const percentChangeClass = computed(() => {
  if (!preview.value) return ''
  return preview.value.percent_change <= 0 ? 'text-green-600' : 'text-red-600'
})

const close = () => {
  emit('close')
}

const loadPreview = async () => {
  if (!selectedCI.value || newUsefulLifeMonths.value <= 0) return

  previewLoading.value = true
  error.value = null
  preview.value = null

  try {
    const ciId = selectedCIId.value || selectedCI.value.id || selectedCI.value.ci_id
    const result = await amortizationApi.previewRestructuring(
      ciId,
      newUsefulLifeMonths.value
    )

    if (!result.is_valid) {
      error.value = result.validation_message || 'Invalid restructuring request'
      return
    }

    preview.value = result
  } catch (err: any) {
    error.value = err.response?.data?.error?.message || 'Failed to preview restructuring'
  } finally {
    previewLoading.value = false
  }
}

const executeRestructuring = async () => {
  if (!selectedCI.value || !isValid.value || !preview.value?.is_valid) return

  loading.value = true
  error.value = null

  try {
    const ciId = selectedCIId.value || selectedCI.value.id || selectedCI.value.ci_id
    const effectiveDateTime = effectiveDate.value ? new Date(effectiveDate.value) : undefined

    await amortizationApi.executeRestructuring(
      ciId,
      newUsefulLifeMonths.value,
      reason.value,
      effectiveDateTime
    )

    emit('success')
    close()
  } catch (err: any) {
    error.value = err.response?.data?.error?.message || 'Failed to execute restructuring'
  } finally {
    loading.value = false
  }
}

// Debounced preview load when useful life changes
let previewTimeout: ReturnType<typeof setTimeout> | null = null
watch(newUsefulLifeMonths, () => {
  if (previewTimeout) clearTimeout(previewTimeout)
  previewTimeout = setTimeout(() => {
    loadPreview()
  }, 500)
})

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount || 0)
}
</script>

<template>
  <Transition name="modal">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" @click.self="close">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <!-- Header -->
        <div class="flex items-center justify-between p-6 border-b">
          <div>
            <h2 class="text-xl font-semibold text-gray-900">Restructure Amortization</h2>
            <p class="text-sm text-gray-500 mt-1">Change useful life with prospective recalculation</p>
          </div>
          <button @click="close" class="text-gray-400 hover:text-gray-600">
            <i class="fas fa-times text-xl"></i>
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-6">
          <!-- Error Alert -->
          <div v-if="error" class="p-4 bg-red-50 border border-red-200 rounded-md">
            <div class="flex">
              <i class="fas fa-exclamation-circle text-red-400 mt-0.5"></i>
              <div class="ml-3">
                <p class="text-sm text-red-800">{{ error }}</p>
              </div>
            </div>
          </div>

          <!-- CI Selection -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Configuration Item <span class="text-red-500">*</span>
            </label>
            <select
              v-model="selectedCIId"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              :disabled="!!props.ciId"
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
          </div>

          <!-- Current Info -->
          <div v-if="selectedCI" class="p-4 bg-gray-50 rounded-md">
            <h3 class="text-sm font-medium text-gray-700 mb-3">Current Configuration</h3>
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500">Current Useful Life:</span>
                <span class="ml-2 font-medium">{{ selectedCI.useful_life_months }} months</span>
              </div>
              <div>
                <span class="text-gray-500">Monthly Depreciation:</span>
                <span class="ml-2 font-medium">{{ formatCurrency(selectedCI.monthly_depreciation) }}</span>
              </div>
              <div>
                <span class="text-gray-500">Current Book Value:</span>
                <span class="ml-2 font-medium">{{ formatCurrency(selectedCI.current_book_value) }}</span>
              </div>
              <div>
                <span class="text-gray-500">Accumulated Depreciation:</span>
                <span class="ml-2 font-medium">{{ formatCurrency(selectedCI.accumulated_depreciation) }}</span>
              </div>
            </div>
          </div>

          <!-- Form -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">
                New Useful Life (months)
              </label>
              <input
                v-model.number="newUsefulLifeMonths"
                type="number"
                min="1"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Enter new useful life in months"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">
                Effective Date
              </label>
              <input
                v-model="effectiveDate"
                type="date"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">
                Reason <span class="text-red-500">*</span>
              </label>
              <textarea
                v-model="reason"
                rows="3"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Explain why the useful life is being changed..."
              ></textarea>
            </div>
          </div>

          <!-- Preview -->
          <div v-if="previewLoading" class="flex items-center justify-center py-8">
            <i class="fas fa-spinner fa-spin text-blue-500 text-2xl"></i>
            <span class="ml-3 text-gray-600">Calculating impact...</span>
          </div>

          <div v-else-if="preview && preview.is_valid" class="p-4 bg-blue-50 border border-blue-200 rounded-md">
            <h3 class="text-sm font-medium text-gray-700 mb-3">
              <i class="fas fa-chart-line mr-2"></i>
              Restructuring Preview
            </h3>

            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500">Old Monthly Depreciation:</span>
                <span class="ml-2 font-medium">{{ formatCurrency(preview.current_monthly_depreciation) }}</span>
              </div>
              <div>
                <span class="text-gray-500">New Monthly Depreciation:</span>
                <span class="ml-2 font-medium" :class="monthlyChangeClass">
                  {{ formatCurrency(preview.new_monthly_depreciation) }}
                </span>
              </div>
              <div>
                <span class="text-gray-500">Monthly Change:</span>
                <span class="ml-2 font-medium" :class="monthlyChangeClass">
                  {{ monthlyChangeFormatted }}
                </span>
              </div>
              <div>
                <span class="text-gray-500">Percent Change:</span>
                <span class="ml-2 font-medium" :class="percentChangeClass">
                  {{ percentChangeFormatted }}
                </span>
              </div>
              <div>
                <span class="text-gray-500">Remaining Life (Old):</span>
                <span class="ml-2 font-medium">{{ preview.remaining_months_old }} months</span>
              </div>
              <div>
                <span class="text-gray-500">Remaining Life (New):</span>
                <span class="ml-2 font-medium text-blue-600">{{ preview.remaining_months_new }} months</span>
              </div>
              <div v-if="preview.new_end_date" class="col-span-2">
                <span class="text-gray-500">New End Date:</span>
                <span class="ml-2 font-medium">{{ new Date(preview.new_end_date).toLocaleDateString() }}</span>
              </div>
            </div>

            <div class="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p class="text-xs text-yellow-800">
                <i class="fas fa-info-circle mr-1"></i>
                This uses <strong>prospective recalculation</strong>: future depreciation will be calculated based on the current book value spread over the remaining new useful life. Historical depreciation is preserved.
              </p>
            </div>
          </div>

          <!-- Validation Message -->
          <div v-else-if="preview && !preview.is_valid" class="p-4 bg-red-50 border border-red-200 rounded-md">
            <p class="text-sm text-red-800">
              <i class="fas fa-exclamation-triangle mr-2"></i>
              {{ preview.validation_message }}
            </p>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end gap-3 p-6 border-t bg-gray-50">
          <button
            @click="close"
            class="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            :disabled="loading"
          >
            Cancel
          </button>
          <button
            @click="executeRestructuring"
            :disabled="loading || !isValid || !preview?.is_valid"
            class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
            <i v-else class="fas fa-check mr-2"></i>
            {{ loading ? 'Processing...' : 'Apply Restructure' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .bg-white,
.modal-leave-active .bg-white {
  transition: transform 0.3s ease;
}

.modal-enter-from .bg-white,
.modal-leave-to .bg-white {
  transform: scale(0.95);
}
</style>

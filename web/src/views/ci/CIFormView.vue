<template>
  <div class="page-container page-content" style="max-width: 1024px; margin: 0 auto;">
    <!-- Page header -->
    <div class="page-header">
      <h1 class="page-title">
        {{
          isEdit
            ? 'Edit Configuration Item'
            : context === 'ea' && domain
              ? `Create ${domain.charAt(0).toUpperCase() + domain.slice(1)} Entity`
              : context === 'asset'
                ? 'Create Asset Management CI'
                : 'Create Configuration Item'
        }}
      </h1>
      <p class="page-subtitle">
        {{ isEdit ? 'Update the configuration item details' : 'Add a new configuration item to your CMDB' }}
      </p>
    </div>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        <!-- Basic Information -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Basic Information</h3>
          </div>
          <div class="card-body">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="form-label">Name *</label>
                <input
                  v-model="form.name"
                  type="text"
                  required
                  class="form-input"
                  placeholder="Enter CI name"
                  :disabled="loading"
                >
              </div>
              <div>
                <label class="form-label">CI Type *</label>
                <CITypeAutocomplete
                  v-model="form.ci_type"
                  :domain="context === 'ea' ? 'ea' : context === 'asset' ? 'asset' : 'all'"
                  :ea-domain="domain"
                  placeholder="Search CI types..."
                  :disabled="loading || isEdit"
                  @update:model-value="onCITypeChange"
                />
              </div>
              <div>
                <label class="form-label">Lifecycle Status</label>
                <select
                  v-model="form.lifecycle_status_id"
                  class="form-input"
                  :disabled="loading"
                >
                  <option value="">Select Lifecycle Status</option>
                  <option v-for="status in activeLifecycleStatuses" :key="status.id" :value="status.id">
                    {{ status.display_name }}
                  </option>
                </select>
                <p class="text-xs text-gray-500 mt-1">
                  Current lifecycle status of this CI (optional)
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Attributes -->
        <div v-if="selectedCIType" class="bg-white shadow rounded-lg">
          <div class="px-4 py-5 border-b border-gray-200 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Attributes</h3>
            <p class="mt-1 text-sm text-gray-500">
              Configure the attributes for this {{ selectedCIType.name }}
            </p>
          </div>
          <div class="px-4 py-5 sm:p-6">
            <!-- Required Attributes Section -->
            <div v-if="selectedCIType.required_attributes.length > 0" class="mb-8">
              <h4 class="text-md font-medium text-gray-900 mb-4 flex items-center">
                <Icon name="required" class="w-4 h-4 text-red-500 mr-2" />
                Required Attributes
                <span class="ml-2 px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded">
                  {{ selectedCIType.required_attributes.length }}
                </span>
              </h4>
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <DynamicAttributeField
                  v-for="attr in selectedCIType.required_attributes"
                  :key="`req-${attr.name}`"
                  :attribute="attr"
                  :model-value="form.attributes[attr.name]"
                  @update:model-value="updateAttribute(attr.name, $event)"
                  @validation="onAttributeValidation(attr.name, $event)"
                  :required="true"
                  :disabled="loading"
                />
              </div>
            </div>

            <!-- Optional Attributes Section -->
            <div v-if="selectedCIType.optional_attributes.length > 0">
              <h4 class="text-md font-medium text-gray-900 mb-4 flex items-center">
                <Icon name="optional" class="w-4 h-4 text-blue-500 mr-2" />
                Optional Attributes
                <span class="ml-2 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
                  {{ selectedCIType.optional_attributes.length }}
                </span>
              </h4>
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <DynamicAttributeField
                  v-for="attr in selectedCIType.optional_attributes"
                  :key="`opt-${attr.name}`"
                  :attribute="attr"
                  :model-value="form.attributes[attr.name]"
                  @update:model-value="updateAttribute(attr.name, $event)"
                  @validation="onAttributeValidation(attr.name, $event)"
                  :required="false"
                  :disabled="loading"
                />
              </div>
            </div>

            <!-- Validation Summary -->
            <div v-if="validationErrors.length > 0" class="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
              <h4 class="text-sm font-medium text-red-800 mb-2">Please fix the following errors:</h4>
              <ul class="text-sm text-red-700 space-y-1">
                <li v-for="error in validationErrors" :key="error" class="flex items-center">
                  <Icon name="error" class="w-4 h-4 mr-2" />
                  {{ error }}
                </li>
              </ul>
            </div>

            <!-- Schema Summary -->
            <div v-if="selectedCIType" class="mt-6 bg-gray-50 border border-gray-200 rounded-lg p-4">
              <h4 class="text-sm font-medium text-gray-900 mb-2">Schema Summary</h4>
              <div class="text-sm text-gray-600">
                <p>{{ selectedCIType.name }} includes {{ selectedCIType.required_attributes.length }} required and {{ selectedCIType.optional_attributes.length }} optional attributes</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Tags -->
        <div class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Tags</h3>
            <p class="mt-1 text-sm text-gray-500">Add tags to help categorize and find this CI</p>
          </div>
          <div class="card-body">
            <div class="space-y-2">
              <div v-for="(tag, index) in form.tags" :key="index" class="flex items-center space-x-2">
                <input
                  v-model="form.tags[index]"
                  type="text"
                  class="form-input"
                  placeholder="Enter tag"
                  :disabled="loading"
                >
                <button
                  type="button"
                  @click="removeTag(index)"
                  class="text-red-600 hover:text-red-800"
                  :disabled="loading"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </button>
              </div>
              <button
                type="button"
                @click="addTag"
                class="text-blue-600 hover:text-blue-800 text-sm"
                :disabled="loading"
              >
                + Add Tag
              </button>
            </div>
          </div>
        </div>

        <!-- Financials - Only for amortizable CI types -->
        <div v-if="selectedCIType && selectedCIType.is_amortizable" class="card">
          <div class="card-header">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Financial Information</h3>
            <p class="mt-1 text-sm text-gray-500">
              Configure amortization and depreciation settings for this {{ selectedCIType.name }}
              <span v-if="hasFinancialData" class="ml-2 text-amber-600 font-medium">
                (Read-only - use /amortization page for adjustments)
              </span>
            </p>
          </div>
          <div class="card-body">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="form-label">Purchase Cost</label>
                <div class="relative">
                  <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500 font-medium">$</span>
                  <input
                    v-model="rawCurrencyInputs.purchase_cost"
                    @input="handleCurrencyInput('purchase_cost', $event)"
                    @blur="handleCurrencyBlur('purchase_cost')"
                    type="text"
                    :class="['form-input pl-8', hasFinancialData ? 'bg-gray-50' : '']"
                    placeholder="1,000.00"
                    :disabled="loading || hasFinancialData"
                    :title="hasFinancialData ? 'Financial data already exists. Use /amortization page for adjustments.' : ''"
                  >
                </div>
                 <p class="text-xs text-gray-500 mt-1">
                   Initial purchase cost of the asset
                   <span v-if="hasFinancialData" class="text-amber-600">
                     - Cannot edit directly (amortization in progress)
                   </span>
                 </p>
              </div>
              <div>
                <label class="form-label">Salvage Value</label>
                <div class="relative">
                  <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500 font-medium">$</span>
                   <input
                    v-model="rawCurrencyInputs.salvage_value"
                    @input="handleCurrencyInput('salvage_value', $event)"
                    @blur="handleCurrencyBlur('salvage_value')"
                    type="text"
                    :class="['form-input pl-8', hasFinancialData ? 'bg-gray-50' : '']"
                    placeholder="100.00"
                    :disabled="loading || hasFinancialData"
                    :title="hasFinancialData ? 'Financial data already exists. Use /amortization page for adjustments.' : ''"
                  >
                </div>
                 <p class="text-xs text-gray-500 mt-1">
                   Expected value at end of useful life
                   <span v-if="hasFinancialData" class="text-amber-600">
                     - Cannot edit directly (amortization in progress)
                   </span>
                 </p>
              </div>
              <div>
                <label class="form-label">Amortization Start Date</label>
                 <input
                  v-model="form.financials.amort_start_date"
                  type="date"
                  :class="['form-input', hasFinancialData ? 'bg-gray-50' : '']"
                  :disabled="loading || hasFinancialData"
                  :title="hasFinancialData ? 'Financial data already exists. Use /amortization page for adjustments.' : ''"
                >
                 <p class="text-xs text-gray-500 mt-1">
                   Date when depreciation calculation begins
                   <span v-if="hasFinancialData" class="text-amber-600">
                     - Cannot edit directly (amortization in progress)
                   </span>
                 </p>
              </div>
              <div>
                <label class="form-label">Useful Life (months)</label>
                 <input
                  v-model.number="form.financials.useful_life_months"
                  type="number"
                  min="1"
                  max="600"
                  :class="['form-input', hasFinancialData ? 'bg-gray-50' : '']"
                  placeholder="60"
                  :disabled="loading || hasFinancialData"
                  :title="hasFinancialData ? 'Financial data already exists. Use /amortization page for adjustments.' : ''"
                >
                 <p class="text-xs text-gray-500 mt-1">
                   Expected useful life in months (1-600)
                   <span v-if="hasFinancialData" class="text-amber-600">
                     - Cannot edit directly (amortization in progress)
                   </span>
                 </p>
              </div>
            </div>

            <!-- Calculated Values Display -->
            <div v-if="form.financials.purchase_cost && form.financials.useful_life_months" class="mt-6 p-4 bg-gray-50 rounded-lg">
              <h4 class="text-sm font-medium text-gray-900 mb-3">Calculated Values</h4>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div>
                  <span class="text-gray-500">Monthly Depreciation:</span>
                  <div class="font-medium text-gray-900">{{ formatCurrency(calculateMonthlyDepreciation()) }}</div>
                </div>
                <div>
                  <span class="text-gray-500">Current Book Value:</span>
                  <div class="font-medium text-gray-900">{{ formatCurrency(form.financials.current_book_value || calculateCurrentBookValue()) }}</div>
                </div>
                <div>
                  <span class="text-gray-500">Remaining Months:</span>
                  <div class="font-medium text-gray-900">{{ calculateRemainingMonths() }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end space-x-3">
          <router-link
            to="/ci"
            class="btn btn-outline"
          >
            Cancel
          </router-link>
          <button
            type="submit"
            :disabled="loading || !isFormValid"
            class="btn btn-primary"
          >
            <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
            {{ isEdit ? 'Update' : 'Create' }} Configuration Item
          </button>
        </div>
      </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLifecycleStatusStore } from '@/stores/lifecycleStatus'
import { useAmortizationStore } from '@/stores/amortization'
import { ciAPI, ciTypeAPI } from '@/services/api'
import { showSuccessToast, showErrorToast } from '@/utils/toast'
import DynamicAttributeField from '@/components/ci/DynamicAttributeField.vue'
import Icon from '@/components/base/Icon.vue'
import CITypeAutocomplete from '@/components/base/CITypeAutocomplete.vue'
import type { CI, CIType, CreateCIData, UpdateCIData } from '@/types/ci'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const lifecycleStatusStore = useLifecycleStatusStore()
const amortizationStore = useAmortizationStore()

const context = computed(() => route.query.context as string | undefined)
const domain = computed(() => route.query.domain as string | undefined)

const loading = ref(false)
const ciTypes = ref<CIType[]>([])
const existingCI = ref<CI | null>(null)
const attributeValidation = ref<Record<string, { isValid: boolean; error?: string }>>({})
const existingFinancialData = ref<any>(null)

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  name: '',
  ci_type: '',
  attributes: {} as Record<string, any>,
  tags: [] as string[],
  lifecycle_status_id: '',
  financials: {
    purchase_cost: null as number | null,
    salvage_value: null as number | null,
    amort_start_date: '' as string,
    useful_life_months: null as number | null,
    current_book_value: null as number | null,
  },
})

const activeLifecycleStatuses = computed(() => lifecycleStatusStore.activeLifecycleStatuses)

const selectedCIType = computed(() => {
  return ciTypes.value.find(type => type.name === form.ci_type)
})

const validationErrors = computed(() => {
  const errors: string[] = []

  // Check required attributes validation
  if (selectedCIType.value) {
    selectedCIType.value.required_attributes.forEach(attr => {
      const validation = attributeValidation.value[attr.name]
      if (validation && !validation.isValid) {
        errors.push(validation.error || `${attr.name} is invalid`)
      }
    })
  }

  // Check financial fields validation for amortizable CI types
  if (selectedCIType.value && selectedCIType.value.is_amortizable) {
    const hasAnyFinancialData = (
      form.financials.purchase_cost !== null && form.financials.purchase_cost > 0 ||
      form.financials.salvage_value !== null && form.financials.salvage_value > 0 ||
      form.financials.amort_start_date !== null && form.financials.amort_start_date !== '' ||
      form.financials.useful_life_months !== null && form.financials.useful_life_months > 0
    )

    if (hasAnyFinancialData) {
      // If any financial field is filled, all required fields must be filled
      if (!form.financials.purchase_cost || form.financials.purchase_cost <= 0) {
        errors.push('Purchase Cost is required when filling financial information')
      }
      if (!form.financials.amort_start_date || form.financials.amort_start_date === '') {
        errors.push('Amortization Start Date is required when filling financial information')
      }
      if (!form.financials.useful_life_months || form.financials.useful_life_months <= 0) {
        errors.push('Useful Life (months) is required when filling financial information')
      }
    }
  }

  return errors
})

const isFormValid = computed(() => {
  if (!form.name.trim() || !form.ci_type || !selectedCIType.value) {
    return false
  }

  // Check all required attributes are valid
  if (selectedCIType.value) {
    for (const attr of selectedCIType.value.required_attributes) {
      const validation = attributeValidation.value[attr.name]
      if (validation && !validation.isValid) {
        return false
      }
    }
  }

  // Check financial fields validation for amortizable CI types
  if (selectedCIType.value && selectedCIType.value.is_amortizable) {
    const hasAnyFinancialData = (
      form.financials.purchase_cost !== null && form.financials.purchase_cost > 0 ||
      form.financials.salvage_value !== null && form.financials.salvage_value > 0 ||
      form.financials.amort_start_date !== null && form.financials.amort_start_date !== '' ||
      form.financials.useful_life_months !== null && form.financials.useful_life_months > 0
    )

    if (hasAnyFinancialData) {
      // If any financial field is filled, all required fields must be filled
      if (!form.financials.purchase_cost || form.financials.purchase_cost <= 0) {
        return false
      }
      if (!form.financials.amort_start_date || form.financials.amort_start_date === '') {
        return false
      }
      if (!form.financials.useful_life_months || form.financials.useful_life_months <= 0) {
        return false
      }
    }
  }

  return true
})

const hasFinancialData = computed(() => {
  // Only show as read-only if editing AND existing financial data exists
  // Check that at least one field has a meaningful value (not null, not 0/empty)
  return (
    isEdit.value &&
    existingFinancialData.value && (
      (existingFinancialData.value.purchase_cost != null && existingFinancialData.value.purchase_cost > 0) ||
      (existingFinancialData.value.salvage_value != null && existingFinancialData.value.salvage_value > 0) ||
      (existingFinancialData.value.amort_start_date != null && existingFinancialData.value.amort_start_date !== '') ||
      (existingFinancialData.value.useful_life_months != null && existingFinancialData.value.useful_life_months > 0)
    )
  )
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const onCITypeChange = async () => {
  // Reset attributes and validation when CI type changes
  form.attributes = {}
  attributeValidation.value = {}

  // If selectedCIType is not in ciTypes, we need to fetch it
  if (!selectedCIType.value && form.ci_type) {
    try {
      // Fetch the CI type details from API
      const response = await ciTypeAPI.list({ search: form.ci_type, limit: 100 })
      const matchedType = response.data.ci_types.find((t: CIType) => t.name === form.ci_type)
      if (matchedType) {
        // Add to ciTypes if not already present
        if (!ciTypes.value.find(t => t.name === matchedType.name)) {
          ciTypes.value.push(matchedType)
        }
      }
    } catch (error) {
      console.error('Failed to load CI type details:', error)
    }
  }

  // Set default values for required attributes if available
  if (selectedCIType.value) {
    selectedCIType.value.required_attributes.forEach((attr: any) => {
      if (attr.default !== undefined) {
        form.attributes[attr.name] = attr.default
      }
      // Initialize validation state
      attributeValidation.value[attr.name] = { isValid: !attr.default, error: undefined }
    })

    // Initialize validation for optional attributes
    selectedCIType.value.optional_attributes.forEach((attr: any) => {
      attributeValidation.value[attr.name] = { isValid: true, error: undefined }
    })
  }
}

const updateAttribute = (attributeName: string, value: any) => {
  form.attributes[attributeName] = value
}

const onAttributeValidation = (attributeName: string, validation: { isValid: boolean; error?: string } | boolean, error?: string) => {
  if (typeof validation === 'boolean') {
    attributeValidation.value[attributeName] = { isValid: validation, error }
  } else {
    attributeValidation.value[attributeName] = validation
  }
}

const addTag = () => {
  form.tags.push('')
}

const removeTag = (index: number) => {
  form.tags.splice(index, 1)
}

// Financial calculation functions
const formatCurrency = (amount: number | null): string => {
  if (amount === null || amount === undefined) return '$0.00'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

// Store raw input values for smooth editing experience
const rawCurrencyInputs = ref({
  purchase_cost: '',
  salvage_value: ''
})

// Format currency for display in input fields (without currency symbol)
const formatCurrencyDisplay = (amount: number | null): string => {
  if (amount === null || amount === undefined) return ''
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
    useGrouping: true,
  }).format(amount)
}

// Handle currency input - allow raw typing, only parse on blur
const handleCurrencyInput = (field: 'purchase_cost' | 'salvage_value', event: Event) => {
  const target = event.target as HTMLInputElement
  const inputValue = target.value

  // Update raw input for display
  rawCurrencyInputs.value[field] = inputValue
}

// Format currency when leaving the field
const handleCurrencyBlur = (field: 'purchase_cost' | 'salvage_value') => {
  const rawValue = rawCurrencyInputs.value[field]

  // Parse the raw input to numeric value
  const numericValue = rawValue.replace(/[^0-9.]/g, '')
  const parsedValue = numericValue ? parseFloat(numericValue) : null

  // Update the form model
  form.financials[field] = parsedValue

  // Update the display with formatted value
  if (parsedValue !== null && parsedValue !== undefined) {
    rawCurrencyInputs.value[field] = formatCurrencyDisplay(parsedValue)
  } else {
    rawCurrencyInputs.value[field] = ''
  }
}


const calculateMonthlyDepreciation = (): number => {
  if (!form.financials.purchase_cost || !form.financials.useful_life_months) return 0
  const depreciableAmount = form.financials.purchase_cost - (form.financials.salvage_value || 0)
  return depreciableAmount / form.financials.useful_life_months
}

const calculateCurrentBookValue = (): number => {
  if (!form.financials.purchase_cost || !form.financials.amort_start_date) return 0

  const monthlyDepreciation = calculateMonthlyDepreciation()
  const startDate = new Date(form.financials.amort_start_date)
  const currentDate = new Date()

  const monthsPassed = Math.max(0,
    Math.floor((currentDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24 * 30))
  )

  const totalDepreciation = Math.min(
    monthsPassed * monthlyDepreciation,
    form.financials.purchase_cost - (form.financials.salvage_value || 0)
  )

  return form.financials.purchase_cost - totalDepreciation
}

const calculateRemainingMonths = (): number => {
  if (!form.financials.useful_life_months || !form.financials.amort_start_date) return 0

  const startDate = new Date(form.financials.amort_start_date)
  const currentDate = new Date()
  const monthsPassed = Math.floor((currentDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24 * 30))

  return Math.max(0, form.financials.useful_life_months - monthsPassed)
}

const loadCITypes = async () => {
  try {
    const response = await ciTypeAPI.list()
    let allTypes = response.data.ci_types || []

    // Filter based on context
    if (context.value === 'asset') {
      // Show only non-EA types
      allTypes = allTypes.filter((type: CIType) => !type.name.startsWith('EA.'))
    } else if (context.value === 'ea' && domain.value) {
      // Show only EA types for specific domain
      const domainPrefix = `EA.${domain.value.charAt(0).toUpperCase() + domain.value.slice(1)}`
      allTypes = allTypes.filter((type: CIType) => type.name.startsWith(domainPrefix))
    }

    ciTypes.value = allTypes
  } catch (error) {
    console.error('Failed to load CI types:', error)
    showErrorToast('Failed to load CI types')
  }
}

const loadCI = async () => {
  if (!route.params.id) return

  try {
    const response = await ciAPI.get(route.params.id as string)
    existingCI.value = response.data

    // Populate form with existing data
    form.name = existingCI.value.name
    form.ci_type = existingCI.value.ci_type
    form.attributes = { ...existingCI.value.attributes }
    form.tags = existingCI.value.tags ? [...existingCI.value.tags] : []
    form.lifecycle_status_id = existingCI.value.lifecycle_status_id || ''

    // Wait for CI type to be loaded and set up validation
    await new Promise(resolve => setTimeout(resolve, 100))

    // Load financial data if this is an amortizable CI type (after CI type is resolved)
    if (selectedCIType.value && selectedCIType.value.is_amortizable) {
      try {
        const financialResponse = await amortizationStore.loadAssetFinancials(existingCI.value.id)
        if (financialResponse) {
          // Store existing financial data for read-only check
          existingFinancialData.value = { ...financialResponse }

          form.financials.purchase_cost = financialResponse.purchase_cost || null
          form.financials.salvage_value = financialResponse.salvage_value || null
          // Format date for HTML date input (yyyy-MM-dd)
          form.financials.amort_start_date = financialResponse.amort_start_date
            ? new Date(financialResponse.amort_start_date).toISOString().split('T')[0]
            : ''
          form.financials.useful_life_months = financialResponse.useful_life_months || null
          form.financials.current_book_value = financialResponse.current_book_value || null

          // Initialize raw currency inputs with formatted values
          if (financialResponse.purchase_cost) {
            rawCurrencyInputs.value.purchase_cost = formatCurrencyDisplay(financialResponse.purchase_cost)
          }
          if (financialResponse.salvage_value) {
            rawCurrencyInputs.value.salvage_value = formatCurrencyDisplay(financialResponse.salvage_value)
          }
        } else {
          // No existing financial data found
          existingFinancialData.value = null
        }
      } catch (error) {
        console.error('Failed to load financial data:', error)
        existingFinancialData.value = null
        // Don't fail the entire form load if financial data fails
      }
    }

    // Initialize validation for loaded CI
    if (selectedCIType.value) {
      // Initialize validation for required attributes
      selectedCIType.value.required_attributes.forEach((attr: any) => {
        const hasValue = form.attributes[attr.name] !== undefined && form.attributes[attr.name] !== null && form.attributes[attr.name] !== ''
        attributeValidation.value[attr.name] = {
          isValid: hasValue,
          error: hasValue ? undefined : `${attr.name} is required`
        }
      })

      // Initialize validation for optional attributes
      selectedCIType.value.optional_attributes.forEach((attr: any) => {
        attributeValidation.value[attr.name] = { isValid: true, error: undefined }
      })
    }
  } catch (error: any) {
    console.error('Failed to load CI:', error)
    const message = error.response?.data?.error || 'Failed to load configuration item'
    showErrorToast(message)
    router.push('/ci')
  }
}

const handleSubmit = async () => {
  if (!isFormValid.value) return

  loading.value = true

  try {
    const data = {
      name: form.name.trim(),
      ci_type: form.ci_type,
      attributes: form.attributes,
      tags: form.tags.filter(tag => tag.trim()),
      lifecycle_status_id: form.lifecycle_status_id || undefined,
    }

    if (isEdit.value) {
      await ciAPI.update(route.params.id as string, {
        attributes: data.attributes,
        tags: data.tags,
        lifecycle_status_id: data.lifecycle_status_id,
      } as UpdateCIData)

      // Update financial data if CI type is amortizable and financial data is provided
      if (selectedCIType.value && selectedCIType.value.is_amortizable) {
        const financialData = {
          purchase_cost: form.financials.purchase_cost,
          salvage_value: form.financials.salvage_value,
          amort_start_date: form.financials.amort_start_date || undefined,
          useful_life_months: form.financials.useful_life_months,
        }

        // Only send financial data that has values
        const hasFinancialData = Object.values(financialData).some(val => val !== null && val !== '' && val !== undefined)

        if (hasFinancialData) {
          try {
            const existingFinancials = await amortizationStore.loadAssetFinancials(route.params.id as string)
            if (existingFinancials) {
              await amortizationStore.updateAssetFinancials(route.params.id as string, financialData)
            } else {
              await amortizationStore.createAssetFinancials(route.params.id as string, financialData)
            }
          } catch (error) {
            console.error('Failed to save financial data:', error)
            showErrorToast('Configuration item updated but financial data failed to save')
          }
        }
      }

      showSuccessToast('Configuration item updated successfully')
    } else {
      const createResponse = await ciAPI.create(data as CreateCIData)

      // Save financial data if CI type is amortizable and financial data is provided
      if (selectedCIType.value && selectedCIType.value.is_amortizable) {
        const financialData = {
          purchase_cost: form.financials.purchase_cost,
          salvage_value: form.financials.salvage_value,
          amort_start_date: form.financials.amort_start_date || undefined,
          useful_life_months: form.financials.useful_life_months,
        }

        // Only send financial data that has values
        const hasFinancialData = Object.values(financialData).some(val => val !== null && val !== '' && val !== undefined)

        if (hasFinancialData) {
          try {
            await amortizationStore.createAssetFinancials(createResponse.data.id, financialData)
          } catch (error) {
            console.error('Failed to save financial data:', error)
            showErrorToast('Configuration item created but financial data failed to save')
          }
        }
      }

      showSuccessToast('Configuration item created successfully')
    }

    router.push('/ci')
  } catch (error: any) {
    console.error('Failed to save CI:', error)
    const message = error.response?.data?.error || 'Failed to save configuration item'
    showErrorToast(message)
  } finally {
    loading.value = false
  }
}

// Watch for changes in CI type to handle financial fields visibility
watch(() => selectedCIType.value, (newType) => {
  if (newType && newType.is_amortizable && isEdit.value && existingCI.value) {
    // Load financial data when switching to an amortizable CI type
    amortizationStore.loadAssetFinancials(existingCI.value.id)
      .then(response => {
        if (response) {
          // Store existing financial data for read-only check
          existingFinancialData.value = { ...response }

          form.financials.purchase_cost = response.purchase_cost || null
          form.financials.salvage_value = response.salvage_value || null
          // Format date for HTML date input (yyyy-MM-dd)
          form.financials.amort_start_date = response.amort_start_date
            ? new Date(response.amort_start_date).toISOString().split('T')[0]
            : ''
          form.financials.useful_life_months = response.useful_life_months || null
          form.financials.current_book_value = response.current_book_value || null

          // Initialize raw currency inputs with formatted values
          if (response.purchase_cost) {
            rawCurrencyInputs.value.purchase_cost = formatCurrencyDisplay(response.purchase_cost)
          }
          if (response.salvage_value) {
            rawCurrencyInputs.value.salvage_value = formatCurrencyDisplay(response.salvage_value)
          }
        } else {
          existingFinancialData.value = null
        }
      })
      .catch(error => {
        console.error('Failed to load financial data:', error)
        existingFinancialData.value = null
      })
  } else {
    // Clear financial data and raw inputs when switching to non-amortizable CI type
    form.financials = {
      purchase_cost: null,
      salvage_value: null,
      amort_start_date: '',
      useful_life_months: null,
      current_book_value: null
    }
    rawCurrencyInputs.value = {
      purchase_cost: '',
      salvage_value: ''
    }
    // Also clear existing financial data
    existingFinancialData.value = null
  }
}, { immediate: true })

onMounted(async () => {
  // Initialize raw inputs as empty for create mode
  rawCurrencyInputs.value = {
    purchase_cost: '',
    salvage_value: ''
  }

  await Promise.all([
    loadCITypes(),
    lifecycleStatusStore.getActiveLifecycleStatuses()
  ])

  if (isEdit.value) {
    await loadCI()
  }
})
</script>
<template>
  <div class="amortization-settings">
    <div class="page-header">
      <h1 class="page-title">Amortization Settings</h1>
      <p class="page-description">
        Configure amortization methods and default values
      </p>
    </div>

    <div class="settings-container">
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading settings...
      </div>

      <form v-else @submit.prevent="saveSettings" class="settings-form">


        <div class="form-section">
          <h2>Financial Settings</h2>
          <div class="form-group">
            <label for="currency">Currency</label>
            <select id="currency" v-model="settings.currency" class="form-select">
              <option value="USD">USD - US Dollar</option>
              <option value="EUR">EUR - Euro</option>
              <option value="GBP">GBP - British Pound</option>
            </select>
          </div>

          <div class="form-group">
            <label for="usefulLife">Default Useful Life (months)</label>
            <input
              id="usefulLife"
              v-model.number="settings.default_useful_life_months"
              type="number"
              min="1"
              max="600"
              class="form-input"
            />
            <small class="form-help">
              Default useful life for new amortizable assets (1-600 months)
            </small>
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" class="btn btn-primary" :disabled="saving">
            <i v-if="saving" class="fas fa-spinner fa-spin"></i>
            {{ saving ? 'Saving...' : 'Save Settings' }}
          </button>
          <button type="button" @click="resetSettings" class="btn btn-secondary">
            Reset to Defaults
          </button>
        </div>

        <div v-if="successMessage" class="alert alert-success">
          {{ successMessage }}
        </div>
        <div v-if="errorMessage" class="alert alert-error">
          {{ errorMessage }}
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAmortizationStore } from '@/stores/amortization'
import type { AmortizationSettings } from '@/types/amortization'

const amortizationStore = useAmortizationStore()

const loading = ref(false)
const saving = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

const settings = ref<AmortizationSettings>({
  id: 'global',
  currency: 'USD',
  default_useful_life_months: 36,
  created_at: '',
  updated_at: '',
})

onMounted(async () => {
  await loadSettings()
})

const loadSettings = async () => {
  loading.value = true
  try {
    const response = await amortizationStore.loadSettings()
    if (response) {
      settings.value = response
    }
  } catch (error) {
    console.error('Failed to load settings:', error)
    errorMessage.value = 'Failed to load amortization settings'
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  successMessage.value = ''
  errorMessage.value = ''

  try {
    const response = await amortizationStore.updateSettings(settings.value)
    if (response.data) {
      settings.value = response.data
      successMessage.value = 'Amortization settings saved successfully!'
      setTimeout(() => {
        successMessage.value = ''
      }, 3000)
    }
  } catch (error: any) {
    console.error('Failed to save settings:', error)
    errorMessage.value = error.response?.data?.message || 'Failed to save settings'
  } finally {
    saving.value = false
  }
}

const resetSettings = () => {
  settings.value = {
    id: 'global',
    currency: 'USD',
    default_useful_life_months: 36,
    created_at: '',
    updated_at: '',
  }
  errorMessage.value = ''
  successMessage.value = ''
}
</script>

<style scoped>
.amortization-settings {
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

.settings-container {
  max-width: 600px;
  margin: 0 auto;
}

.settings-form {
  background: white;
  padding: 2rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.form-section {
  margin-bottom: 2rem;
}

.form-section h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-select,
.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.form-select:focus,
.form-input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.form-help {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: #6b7280;
  line-height: 1.4;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 0.75rem 1.5rem;
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

.alert {
  padding: 1rem;
  border-radius: 0.375rem;
  margin-top: 1rem;
}

.alert-success {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.alert-error {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

@media (max-width: 768px) {
  .amortization-settings {
    padding: 1rem;
  }

  .settings-form {
    padding: 1.5rem;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import importApi, { ImportError, ImportResult } from '@/services/importApi'

export interface ImportRow {
  row_number?: number
  Name: string
  CI_Type: string
  Domain: string
  Lifecycle_Status: string
  Owner: string
  Team: string
  Attributes: string
  Tags: string
}

export const useEaImportStore = defineStore('eaImport', () => {
  // State
  const currentStep = ref(1) // 1=upload, 2=preview, 3=validate, 4=import
  const file = ref<File | null>(null)
  const ciType = ref<string>('')
  const parsedData = ref<ImportRow[]>([])
  const validationResult = ref<ImportResult | null>(null)
  const importResult = ref<ImportResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const hasErrors = computed(() => {
    return validationResult.value && validationResult.value.error_count > 0
  })

  const isValid = computed(() => {
    return validationResult.value && validationResult.value.error_count === 0
  })

  const errorCount = computed(() => {
    return validationResult.value?.error_count || 0
  })

  const successCount = computed(() => {
    return validationResult.value?.success_count || 0
  })

  const totalRows = computed(() => {
    return parsedData.value.length
  })

  // Actions
  function setStep(step: number) {
    currentStep.value = step
  }

  function setFile(newFile: File | null) {
    file.value = newFile
  }

  function setCiType(type: string) {
    ciType.value = type
  }

  function setParsedData(data: ImportRow[]) {
    parsedData.value = data
  }

  async function downloadTemplate() {
    try {
      loading.value = true
      error.value = null

      const blob = await importApi.generateTemplate(ciType.value)
      const filename = `ea-${ciType.value.replace(/\./g, '-')}-template.csv`
      importApi.downloadBlob(blob, filename)
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to download template'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function validateImport() {
    if (!file.value) {
      error.value = 'No file selected'
      return
    }

    try {
      loading.value = true
      error.value = null

      const result = await importApi.validateImport(file.value, ciType.value)
      validationResult.value = result
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Validation failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function executeImport() {
    if (!file.value) {
      error.value = 'No file selected'
      return
    }

    try {
      loading.value = true
      error.value = null

      const result = await importApi.executeImport(file.value, ciType.value)
      importResult.value = result
      return result
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Import failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function downloadErrorCSV() {
    if (!validationResult.value?.errors) {
      error.value = 'No errors to download'
      return
    }

    try {
      loading.value = true
      error.value = null

      const blob = await importApi.downloadErrorCSV(validationResult.value.errors)
      importApi.downloadBlob(blob, 'import-errors.csv')
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to download error CSV'
      throw err
    } finally {
      loading.value = false
    }
  }

  function reset() {
    currentStep.value = 1
    file.value = null
    ciType.value = ''
    parsedData.value = []
    validationResult.value = null
    importResult.value = null
    loading.value = false
    error.value = null
  }

  return {
    // State
    currentStep,
    file,
    ciType,
    parsedData,
    validationResult,
    importResult,
    loading,
    error,

    // Computed
    hasErrors,
    isValid,
    errorCount,
    successCount,
    totalRows,

    // Actions
    setStep,
    setFile,
    setCiType,
    setParsedData,
    downloadTemplate,
    validateImport,
    executeImport,
    downloadErrorCSV,
    reset,
  }
})

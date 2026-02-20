<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="mb-8">
      <nav class="flex" aria-label="Breadcrumb">
        <ol class="flex items-center space-x-4">
          <li>
            <router-link to="/entities/business" class="text-gray-400 hover:text-gray-500">
              <svg class="flex-shrink-0 h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
              </svg>
            </router-link>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
              <span class="ml-4 text-sm font-medium text-gray-500">EA Entities</span>
            </div>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
              <span class="ml-4 text-sm font-medium text-gray-900">Import</span>
            </div>
          </li>
        </ol>
      </nav>

      <div class="mt-4">
        <h1 class="text-3xl font-bold text-gray-900">Import EA Entities</h1>
        <p class="mt-2 text-sm text-gray-600">
          Import EA entities from a CSV file. Follow the wizard to validate and import your data.
        </p>
      </div>
    </div>

    <!-- Progress Stepper -->
    <div class="mb-8">
      <div class="relative">
        <div class="absolute inset-0 flex items-center" aria-hidden="true">
          <div class="w-full border-t border-gray-300"></div>
        </div>
        <div class="relative flex justify-between">
          <button
            v-for="step in steps"
            :key="step.number"
            @click="canNavigateToStep(step.number) && importStore.setStep(step.number)"
            :disabled="!canNavigateToStep(step.number)"
            :class="[
              currentStep >= step.number ? 'text-blue-600' : 'text-gray-500',
              !canNavigateToStep(step.number) && 'cursor-not-allowed opacity-50',
              'group flex items-center'
            ]"
          >
            <span class="flex items-center">
              <span
                :class="[
                  currentStep >= step.number ? 'bg-blue-600 border-blue-600' : 'bg-white border-gray-300',
                  'h-10 w-10 flex items-center justify-center border-2 rounded-full text-sm font-medium transition-colors'
                ]"
              >
                <span v-if="currentStep > step.number">
                  <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                </span>
                <span v-else>{{ step.number }}</span>
              </span>
              <span class="ml-3 text-sm font-medium hidden sm:inline">{{ step.name }}</span>
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="importStore.error" class="mb-6 rounded-md bg-red-50 p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error</h3>
          <div class="mt-2 text-sm text-red-700">
            {{ importStore.error }}
          </div>
        </div>
      </div>
    </div>

    <!-- Step Content -->
    <div class="bg-white shadow rounded-lg">
      <!-- Step 1: Upload -->
      <div v-show="currentStep === 1" class="p-6">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">Upload CSV File</h2>

        <div class="space-y-6">
          <!-- CI Type Selection -->
          <div>
            <label for="ciType" class="block text-sm font-medium text-gray-700">
              CI Type <span class="text-red-500">*</span>
            </label>
            <select
              id="ciType"
              v-model="importStore.ciType"
              class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md border"
              required
            >
              <option value="">Select a CI type...</option>
              <option value="EA.Business-BusinessCapability">EA.Business-BusinessCapability</option>
              <option value="EA.Business-Process">EA.Business-Process</option>
              <option value="EA.Application-BusinessApp">EA.Application-BusinessApp</option>
              <option value="EA.Application-Component">EA.Application-Component</option>
              <option value="EA.Data-DataObject">EA.Data-DataObject</option>
              <option value="EA.Technology-Component">EA.Technology-Component</option>
              <option value="EA.Infrastructure-Node">EA.Infrastructure-Node</option>
            </select>
          </div>

          <!-- File Upload -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              CSV File <span class="text-red-500">*</span>
            </label>

            <!-- Drag & Drop Zone -->
            <div
              @drop="handleDrop"
              @dragover.prevent
              @dragenter.prevent
              :class="[
                dragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300',
                'border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer hover:border-blue-400'
              ]"
              @click="$refs.fileInput.click()"
            >
              <input
                ref="fileInput"
                type="file"
                accept=".csv"
                class="hidden"
                @change="handleFileSelect"
              />

              <svg v-if="!importStore.file" class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <div v-else class="mx-auto h-12 w-12 text-green-500">
                <svg class="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>

              <p v-if="!importStore.file" class="mt-4 text-sm text-gray-600">
                Drag and drop your CSV file here, or click to browse
              </p>
              <p v-else class="mt-4 text-sm font-medium text-gray-900">
                {{ importStore.file.name }}
              </p>
              <p class="mt-1 text-xs text-gray-500">
                Maximum file size: 32MB
              </p>
            </div>
          </div>

          <!-- Download Template Button -->
          <div class="flex items-center justify-between">
            <button
              @click="handleDownloadTemplate"
              :disabled="!importStore.ciType || importStore.loading"
              class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download Template
            </button>

            <button
              @click="handleNext"
              :disabled="!importStore.file || !importStore.ciType || importStore.loading"
              class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next: Preview
              <svg class="ml-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Step 2: Preview -->
      <div v-show="currentStep === 2" class="p-6">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">Preview Data</h2>

        <ImportPreview
          v-if="importStore.parsedData.length > 0"
          :data="importStore.parsedData"
          :columns="previewColumns"
        />

        <div class="mt-6 flex justify-between">
          <button
            @click="importStore.setStep(1)"
            class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <svg class="mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
            </svg>
            Back
          </button>

          <button
            @click="handleValidate"
            :disabled="importStore.loading"
            class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
          >
            <span v-if="importStore.loading" class="flex items-center">
              <svg class="animate-spin -ml-1 mr-2 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Validating...
            </span>
            <span v-else class="flex items-center">
              Validate
              <svg class="ml-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </span>
          </button>
        </div>
      </div>

      <!-- Step 3: Validate -->
      <div v-show="currentStep === 3" class="p-6">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">Validation Results</h2>

        <ImportValidationErrors
          v-if="importStore.validationResult"
          :errors="importStore.validationResult.errors || []"
          :error-count="importStore.errorCount"
          :has-data="!!importStore.validationResult"
          @download-errors="handleDownloadErrors"
          @retry="importStore.setStep(1)"
        />

        <div class="mt-6 flex justify-between">
          <button
            @click="importStore.setStep(2)"
            class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <svg class="mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
            </svg>
            Back
          </button>

          <button
            v-if="importStore.isValid"
            @click="handleImport"
            :disabled="importStore.loading"
            class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 disabled:opacity-50"
          >
            <span v-if="importStore.loading" class="flex items-center">
              <svg class="animate-spin -ml-1 mr-2 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Importing...
            </span>
            <span v-else class="flex items-center">
              Import {{ importStore.successCount }} Entities
              <svg class="ml-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
              </svg>
            </span>
          </button>
        </div>
      </div>

      <!-- Step 4: Import Complete -->
      <div v-show="currentStep === 4" class="p-6">
        <div class="text-center">
          <div v-if="importStore.importResult" class="mb-6">
            <!-- Success Icon -->
            <div class="mx-auto flex items-center justify-center h-20 w-20 rounded-full bg-green-100 mb-4">
              <svg class="h-12 w-12 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>

            <h2 class="text-2xl font-bold text-gray-900 mb-2">Import Complete!</h2>
            <p class="text-gray-600 mb-6">
              Successfully imported <span class="font-semibold text-green-600">{{ importStore.importResult.success_count }}</span> entities.
              <span v-if="importStore.importResult.error_count > 0">
                {{ importStore.importResult.error_count }} rows failed.
              </span>
            </p>
          </div>

          <!-- Action Buttons -->
          <div class="flex justify-center gap-4">
            <router-link
              to="/entities/business"
              class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <svg class="mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
              View Imported Entities
            </router-link>

            <button
              @click="handleReset"
              class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
            >
              <svg class="mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              Import More
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Papa from 'papaparse'
import { useEaImportStore } from '@/stores/eaImport'
import ImportPreview from '@/components/ea/ImportPreview.vue'
import ImportValidationErrors from '@/components/ea/ImportValidationErrors.vue'
import type { ImportRow } from '@/stores/eaImport'

const route = useRoute()
const router = useRouter()
const importStore = useEaImportStore()

const dragActive = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const steps = [
  { number: 1, name: 'Upload' },
  { number: 2, name: 'Preview' },
  { number: 3, name: 'Validate' },
  { number: 4, name: 'Import' },
]

const currentStep = computed(() => importStore.currentStep)

const previewColumns = computed(() => {
  if (importStore.parsedData.length === 0) return []
  return Object.keys(importStore.parsedData[0]).filter(key => key !== 'Attributes')
})

onMounted(() => {
  // Set CI type from query params if provided
  const ciType = route.query.ci_type as string
  if (ciType) {
    importStore.setCiType(ciType)
  }
})

function canNavigateToStep(step: number): boolean {
  // Can only go to step 2 if file is selected
  if (step === 2) return !!importStore.file
  // Can only go to step 3 if validation is done
  if (step === 3) return !!importStore.validationResult
  // Can only go to step 4 if import is done
  if (step === 4) return !!importStore.importResult
  return true
}

function handleDragOver(event: DragEvent) {
  event.preventDefault()
  dragActive.value = true
}

function handleDragLeave() {
  dragActive.value = false
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  dragActive.value = false

  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (file.type === 'text/csv' || file.name.endsWith('.csv')) {
      importStore.setFile(file)
      parseFileForPreview(file)
    }
  }
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    const file = files[0]
    importStore.setFile(file)
    parseFileForPreview(file)
  }
}

function parseFileForPreview(file: File) {
  Papa.parse(file, {
    header: true,
    skipEmptyLines: true,
    complete: (results) => {
      importStore.setParsedData(results.data as ImportRow[])
    },
    error: (error) => {
      console.error('Error parsing CSV:', error)
    },
  })
}

async function handleDownloadTemplate() {
  try {
    await importStore.downloadTemplate()
  } catch (error) {
    console.error('Failed to download template:', error)
  }
}

async function handleNext() {
  if (importStore.file && importStore.parsedData.length === 0) {
    parseFileForPreview(importStore.file)
  }
  importStore.setStep(2)
}

async function handleValidate() {
  try {
    await importStore.validateImport()
    importStore.setStep(3)
  } catch (error) {
    console.error('Validation failed:', error)
  }
}

async function handleDownloadErrors() {
  try {
    await importStore.downloadErrorCSV()
  } catch (error) {
    console.error('Failed to download errors:', error)
  }
}

async function handleImport() {
  try {
    await importStore.executeImport()
    importStore.setStep(4)
  } catch (error) {
    console.error('Import failed:', error)
  }
}

function handleReset() {
  importStore.reset()
  router.push('/entities/import')
}
</script>

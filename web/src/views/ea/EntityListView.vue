<template>
  <div class="page-container page-content">
    <!-- Page header -->
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">{{ domainDisplay }} Entities</h1>
        <p class="page-subtitle">Manage {{ domainDisplay.toLowerCase() }} entities and their attributes</p>
      </div>
      <div class="flex space-x-2">
        <button
          v-if="hasPermission('ea:delete')"
          @click="exportToCSV"
          class="btn btn-outline"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path>
          </svg>
          Export CSV
        </button>
        <router-link
          v-if="hasPermission('ea:create')"
          :to="`/entities/${currentDomain}/create`"
          class="btn btn-primary"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
          Create {{ domainDisplay }} Entity
        </router-link>
      </div>
    </div>

    <div class="flex gap-6">
      <!-- Domain Sidebar -->
      <div class="w-1/6">
        <div class="bg-white shadow rounded-lg p-4">
          <h3 class="text-sm font-medium text-gray-900 mb-3">Domains</h3>
          <nav class="space-y-1">
            <router-link
              v-for="domain in domains"
              :key="domain.id"
              :to="`/entities/${domain.id}`"
              class="flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors duration-150"
              :class="currentDomain === domain.id ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-50'"
              @click="switchDomain(domain.id)"
            >
              <component :is="domain.icon" class="w-4 h-4 mr-2" />
              {{ domain.name }}
            </router-link>
          </nav>
        </div>
      </div>

      <!-- Main Content -->
      <div class="flex-1">
        <!-- Search and Filters -->
        <div class="bg-white shadow rounded-lg p-4 mb-4">
          <div class="flex flex-wrap gap-4">
            <!-- Global Search -->
            <div class="flex-1 min-w-[200px]">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search entities by name, description..."
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                @input="onSearchChange"
              />
            </div>

            <!-- CI Type Filter -->
            <div class="w-48">
              <select
                v-model="filter.ci_type"
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                @change="loadEntities"
              >
                <option value="">All CI Types</option>
                <option
                  v-for="type in availableCITypes"
                  :key="type.name"
                  :value="type.name"
                >
                  {{ type.name }}
                </option>
              </select>
            </div>

            <!-- Lifecycle Status Filter -->
            <div class="w-48">
              <select
                v-model="filter.lifecycle_status_id"
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                @change="loadEntities"
              >
                <option value="">All Statuses</option>
                <option
                  v-for="status in lifecycleStatuses"
                  :key="status.id"
                  :value="status.id"
                >
                  {{ status.display_name }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- ag-Grid Entity Table -->
        <div class="bg-white shadow rounded-lg">
          <div class="ag-theme-alpine" style="height: 600px;">
            <ag-grid-vue
              :column-defs="columnDefs"
              :row-data="rowData"
              :default-col-def="defaultColDef"
              :pagination="true"
              :pagination-page-size="paginationPageSize"
              :pagination-page-size-selector="paginationPageSizeSelector"
              :row-selection="rowSelection"
              :suppress-row-click-selection="true"
              :row-height="48"
              @grid-ready="onGridReady"
              @selection-changed="onSelectionChanged"
            />
          </div>

          <!-- Bulk Actions Bar -->
          <div
            v-if="selectedRows.length > 0"
            class="px-4 py-3 bg-blue-50 border-t border-blue-200 flex items-center justify-between"
          >
            <span class="text-sm font-medium text-blue-900">
              {{ selectedRows.length }} item{{ selectedRows.length > 1 ? 's' : '' }} selected
            </span>
            <div class="flex space-x-2">
              <button
                v-if="hasPermission('ea:update')"
                @click="bulkChangeStatus"
                class="btn btn-sm btn-outline"
              >
                Change Status
              </button>
              <button
                v-if="hasPermission('ea:delete')"
                @click="bulkDelete"
                class="btn btn-sm btn-outline btn-danger"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AgGridVue } from 'ag-grid-vue3'
import 'ag-grid-community/styles/ag-grid.css'
import 'ag-grid-community/styles/ag-theme-alpine.css'
import { useAuthStore } from '@/stores/auth'
import { useEaStore } from '@/stores/ea'
import { useEaTypesStore } from '@/stores/eaTypes'
import { lifecycleStatusAPI } from '@/services/api'
import type { EAEntity } from '@/types/ea'
import type { GridApi, ColumnApi } from 'ag-grid-community'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const eaStore = useEaStore()
const eaTypesStore = useEaTypesStore()

// Domain definitions
const domains = ref([
  { id: 'strategy', name: 'Strategy', icon: 'LightBulbIcon' },
  { id: 'business', name: 'Business', icon: 'BriefcaseIcon' },
  { id: 'application', name: 'Application', icon: 'ChipIcon' },
  { id: 'data', name: 'Data', icon: 'DatabaseIcon' },
  { id: 'technology', name: 'Technology', icon: 'ServerIcon' },
  { id: 'infrastructure', name: 'Infrastructure', icon: 'CloudIcon' },
  { id: 'security', name: 'Security', icon: 'ShieldIcon' },
  { id: 'governance', name: 'Governance', icon: 'ScaleIcon' }
])

// State
const currentDomain = ref((route.params.domain as string) || 'business')
const searchQuery = ref('')
const gridApi = ref<GridApi | null>(null)
const columnApi = ref<ColumnApi | null>(null)
const selectedRows = ref<EAEntity[]>([])
const lifecycleStatuses = ref<any[]>([])

const filter = ref({
  ci_type: '',
  lifecycle_status_id: ''
})

const paginationPageSize = ref(25)
const paginationPageSizeSelector = ref([25, 50, 100])

// ag-Grid configuration
const columnDefs = ref([
  {
    headerName: 'Name',
    field: 'name',
    filter: 'agTextColumnFilter',
    sortable: true,
    resizable: true,
    flex: 2,
    cellRenderer: (params: any) => {
      return `<a href="#/entities/${currentDomain.value}/${params.data.id}" class="text-blue-600 hover:text-blue-900 font-medium">${params.value}</a>`
    }
  },
  {
    headerName: 'CI Type',
    field: 'ci_type_display',
    filter: 'agTextColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1
  },
  {
    headerName: 'Lifecycle Status',
    field: 'lifecycle_status_display',
    filter: 'agTextColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1
  },
  {
    headerName: 'Owner',
    field: 'owner_name',
    filter: 'agTextColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1
  },
  {
    headerName: 'Team',
    field: 'team_name',
    filter: 'agTextColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1
  },
  {
    headerName: 'Data Quality',
    field: 'data_quality_score',
    filter: 'agNumberColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1,
    cellRenderer: (params: any) => {
      const score = params.value
      const colorClass = score >= 80 ? 'text-green-600' : score >= 60 ? 'text-yellow-600' : 'text-red-600'
      return `<span class="${colorClass} font-medium">${score}%</span>`
    }
  },
  {
    headerName: 'Last Updated',
    field: 'updated_at',
    filter: 'agDateColumnFilter',
    sortable: true,
    resizable: true,
    flex: 1,
    valueFormatter: (params: any) => {
      return new Date(params.value).toLocaleDateString()
    }
  },
  {
    headerName: 'Actions',
    sortable: false,
    filter: false,
    resizable: false,
    flex: 1,
    cellRenderer: (params: any) => {
      const entityId = params.data.id
      return `
        <div class="flex space-x-2">
          <a href="#/entities/${currentDomain.value}/${entityId}" class="text-blue-600 hover:text-blue-900" title="View">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
            </svg>
          </a>
          ${hasPermission('ea:update') ? `
            <a href="#/entities/${currentDomain.value}/${entityId}/edit" class="text-indigo-600 hover:text-indigo-900" title="Edit">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
              </svg>
            </a>
          ` : ''}
        </div>
      `
    }
  }
])

const defaultColDef = {
  sortable: true,
  filter: true,
  resizable: true,
  flex: 1
}

const rowSelection = ref('multiple')

const rowData = ref<EAEntity[]>([])

// Computed
const domainDisplay = computed(() => {
  const domain = domains.value.find(d => d.id === currentDomain.value)
  return domain?.name || 'Entities'
})

const availableCITypes = computed(() => {
  return eaTypesStore.getCiTypesByDomain(currentDomain.value)
})

// Methods
const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const onGridReady = (params: any) => {
  gridApi.value = params.api
  columnApi.value = params.columnApi
  loadEntities()
}

const onSelectionChanged = () => {
  if (gridApi.value) {
    const selectedNodes = gridApi.value.getSelectedNodes()
    selectedRows.value = selectedNodes.map(node => node.data)
  }
}

const switchDomain = (domainId: string) => {
  currentDomain.value = domainId
  router.push(`/entities/${domainId}`)
}

const loadEntities = async () => {
  try {
    await eaStore.fetchEntities({
      domain: currentDomain.value,
      ci_type: filter.value.ci_type || undefined,
      lifecycle_status_id: filter.value.lifecycle_status_id || undefined,
      search: searchQuery.value || undefined,
      page: 1,
      page_size: paginationPageSize.value
    })

    rowData.value = eaStore.entities
  } catch (error) {
    console.error('Failed to load entities:', error)
  }
}

const onSearchChange = () => {
  loadEntities()
}

const exportToCSV = () => {
  if (gridApi.value) {
    gridApi.value.exportDataAsCsv({
      fileName: `${currentDomain.value}-entities-${new Date().toISOString().split('T')[0]}.csv`
    })
  }
}

const bulkChangeStatus = async () => {
  // Implement bulk status change
  console.log('Bulk change status for:', selectedRows.value)
}

const bulkDelete = async () => {
  if (confirm(`Are you sure you want to delete ${selectedRows.value.length} entities?`)) {
    try {
      for (const entity of selectedRows.value) {
        await eaStore.deleteEntity(entity.id)
      }
      await loadEntities()
      selectedRows.value = []
    } catch (error) {
      console.error('Failed to delete entities:', error)
    }
  }
}

const loadLifecycleStatuses = async () => {
  try {
    const response = await lifecycleStatusAPI.getActive()
    lifecycleStatuses.value = response.data
  } catch (error) {
    console.error('Failed to load lifecycle statuses:', error)
  }
}

// Lifecycle
onMounted(async () => {
  await loadLifecycleStatuses()

  // Load CI types if not already loaded
  if (eaTypesStore.ciTypes.length === 0) {
    await eaTypesStore.fetchCiTypes()
  }
})

// Watch for route changes
watch(() => route.params.domain, (newDomain) => {
  if (newDomain) {
    currentDomain.value = newDomain as string
    filter.value.ci_type = ''
    filter.value.lifecycle_status_id = ''
    loadEntities()
  }
}, { immediate: true })
</script>

<style scoped>
.ag-theme-alpine {
  font-size: 14px;
}
</style>

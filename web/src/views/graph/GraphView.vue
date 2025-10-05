<template>
  <div class="page-container page-content">
    <!-- Page header -->
    <div class="page-header">
      <h1 class="page-title">Graph Visualization</h1>
      <p class="page-subtitle">Explore relationships between your configuration items</p>
    </div>

      <!-- Controls -->
      <div class="bg-white shadow rounded-lg mb-6">
        <div class="card-body">
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label class="form-label">CI Types</label>
              <select v-model="filters.ci_types" multiple class="form-input" @change="loadGraphData">
                <option v-for="type in ciTypes" :key="type.id" :value="type.name">
                  {{ type.name }}
                </option>
              </select>
              <p class="text-xs text-gray-500 mt-1">Hold Ctrl/Cmd to select multiple</p>
            </div>
            <div>
              <label class="form-label">Search</label>
              <div class="relative">
                <input
                  v-model="filters.search"
                  type="text"
                  placeholder="Search CI names..."
                  class="form-input pr-10"
                  @input="onSearchInputChange"
                  @focus="handleSearchFocus"
                  @blur="hideAutocomplete"
                >
                <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                  </svg>
                </div>
                <!-- Autocomplete dropdown -->
                <div
                  v-if="showAutocomplete && (searchResults.length > 0 || isSearching)"
                  class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto"
                >
                  <div v-if="isSearching" class="p-3 text-gray-500 text-sm">
                    Searching...
                  </div>
                  <div
                    v-for="result in searchResults"
                    :key="result.id"
                    class="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center justify-between"
                    @mousedown="selectSearchResult(result)"
                  >
                    <div class="flex items-center flex-1">
                      <div v-if="result.isShowAll" class="mr-3">
                        <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
                        </svg>
                      </div>
                      <div v-else-if="result.isCIOption" class="mr-3">
                        <div class="w-4 h-4 rounded-full border border-gray-300" :style="{ backgroundColor: getCITypeColor(result.ci_type) }"></div>
                      </div>
                      <div v-else class="mr-3">
                        <div class="w-4 h-4 rounded-full border border-gray-300" :style="{ backgroundColor: getCITypeColor(result.ci_type) }"></div>
                      </div>
                      <div>
                        <div class="font-medium">{{ result.name }}</div>
                        <div class="text-sm text-gray-500">{{ result.ci_type }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div>
              <label class="form-label">Max Nodes</label>
              <select v-model="filters.limit" class="form-input" @change="loadGraphData">
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="200">200</option>
                <option value="500">500</option>
              </select>
            </div>
            <div class="flex items-end">
              <button @click="loadGraphData" :disabled="loading" class="btn btn-primary w-full">
                <span v-if="loading" class="spinner w-4 h-4 mr-2"></span>
                Refresh Graph
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Graph Container -->
      <div class="bg-white shadow rounded-lg">
        <div class="card-header flex justify-between items-center">
          <h3 class="text-lg leading-6 font-medium text-gray-900">
            Configuration Item Graph
          </h3>
          <div class="flex space-x-2">
            <button @click="centerGraph" class="btn btn-outline text-sm">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
              </svg>
              Center
            </button>
            <button @click="fitGraph" class="btn btn-outline text-sm">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4"></path>
              </svg>
              Fit
            </button>
            <button @click="clearGraph" class="btn btn-outline text-sm">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
              </svg>
              Clear
            </button>
          </div>
        </div>
        <div class="relative">
          <!-- Loading State -->
          <div v-if="loading" class="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-10">
            <div class="text-center">
              <div class="spinner w-8 h-8 mx-auto mb-4"></div>
              <p class="text-gray-500">Loading graph data...</p>
            </div>
          </div>

          <!-- Empty State - No results for selected CI -->
          <div v-if="!loading && filters.search && (!graphData || graphData.nodes.length === 0)" class="text-center py-16">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No results found for "{{ filters.search }}"</h3>
            <p class="mt-1 text-sm text-gray-500">
              Try searching for a different configuration item or adjust your filters.
            </p>
          </div>

          <!-- Empty State - Initial state -->
        <div v-else-if="!loading && !filters.search && (!graphData || graphData.nodes.length === 0)" class="text-center py-16">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">Search for Configuration Items</h3>
            <p class="mt-1 text-sm text-gray-500">
              Start typing in the search box above to explore configuration items and their relationships.
            </p>
            <div class="mt-6">
              <div class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white">
                <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                </svg>
                Search and select a CI to begin
              </div>
            </div>
          </div>

          <!-- Graph Canvas -->
          <div
            v-show="!loading && graphData && graphData.nodes.length > 0"
            ref="graphContainer"
            class="w-full"
            style="height: 600px;"
          ></div>

          <!-- Context Menu -->
          <div
            v-if="contextMenu.visible"
            :style="{
              left: contextMenu.x + 'px',
              top: contextMenu.y + 'px',
              position: 'fixed',
              zIndex: 9999
            }"
            class="bg-white border border-gray-200 rounded-lg shadow-lg py-1 min-w-48"
            @click="hideContextMenu"
          >
            <button
              v-if="contextMenu.node"
              @click="showRelationshipOptions(contextMenu.node)"
              class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 flex items-center"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
              </svg>
              Expand Relationships
            </button>
            <button
              v-if="contextMenu.node"
              @click="showCollapseOptions(contextMenu.node)"
              class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 flex items-center"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
              </svg>
              Collapse Relationships
            </button>
            <button
              v-if="contextMenu.node"
              @click="viewNodeDetails(contextMenu.node)"
              class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 flex items-center"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              View Details
            </button>
          </div>

          <!-- Custom Tooltip Component -->
          <GraphTooltip
            v-if="tooltipState.visible"
            :node="tooltipState.node"
            :visible="tooltipState.visible"
            :x="tooltipState.x"
            :y="tooltipState.y"
            @hide="hideTooltip"
            ref="tooltipRef"
          />
        </div>
      </div>

      <!-- Graph Statistics -->
      <div v-if="graphData && graphData.nodes.length > 0" class="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-white shadow rounded-lg p-6">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
                </svg>
              </div>
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">Total Nodes</dt>
                <dd class="text-lg font-medium text-gray-900">{{ graphData.nodes.length }}</dd>
              </dl>
            </div>
          </div>
        </div>

        <div class="bg-white shadow rounded-lg p-6">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <div class="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
                </svg>
              </div>
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">Total Edges</dt>
                <dd class="text-lg font-medium text-gray-900">{{ graphData.edges.length }}</dd>
              </dl>
            </div>
          </div>
        </div>

        <div class="bg-white shadow rounded-lg p-6">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <div class="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                </svg>
              </div>
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">CI Types</dt>
                <dd class="text-lg font-medium text-gray-900">{{ uniqueCITypes.length }}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>

      <!-- Legend -->
      <div v-if="graphData && graphData.nodes.length > 0" class="mt-6 bg-white shadow rounded-lg">
        <div class="card-header">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Legend</h3>
        </div>
        <div class="card-body">
          <div class="flex flex-wrap gap-4">
            <div v-for="type in uniqueCITypes" :key="type" class="flex items-center space-x-2">
              <div
                class="w-4 h-4 rounded-full border-2 border-gray-300"
                :style="{ backgroundColor: getCITypeColor(type) }"
              ></div>
              <span class="text-sm text-gray-700">{{ type }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Relationship Options Modal -->
      <div
        v-if="relationshipModal.visible"
        class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
        @click="hideRelationshipModal"
      >
        <div
          class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4"
          @click.stop
        >
          <div class="p-6">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-medium text-gray-900">
                {{ relationshipModal.mode === 'expand' ? 'Expand' : 'Collapse' }} Relationships for {{ relationshipModal.node?.name }}
              </h3>
              <button
                @click="hideRelationshipModal"
                class="text-gray-400 hover:text-gray-500"
              >
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
              </button>
            </div>

            <div v-if="relationshipModal.loading" class="text-center py-8">
              <div class="spinner w-8 h-8 mx-auto mb-4"></div>
              <p class="text-gray-500">Loading relationship types...</p>
            </div>

            <div v-else-if="hasGroupedRelationships">
              <p class="text-sm text-gray-600 mb-4">
                Select relationship types to {{ relationshipModal.mode === 'expand' ? 'expand' : 'collapse' }}:
              </p>
              <div class="space-y-4 max-h-60 overflow-y-auto">
                <!-- Using section (outgoing relationships) -->
                <div v-if="relationshipModal.relationshipTypes.outgoing.length > 0">
                  <div class="flex items-center mb-2">
                    <svg class="w-4 h-4 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"></path>
                    </svg>
                    <h4 class="text-sm font-semibold text-gray-700">Using (outgoing)</h4>
                    <span class="ml-2 text-xs text-gray-500">{{ getTotalCount('outgoing') }} relationships</span>
                  </div>
                  <div class="space-y-1 ml-6">
                    <label
                      v-for="type in relationshipModal.relationshipTypes.outgoing"
                      :key="`outgoing-${type.type}`"
                      class="flex items-center p-2 border rounded hover:bg-blue-50 cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        :value="`outgoing-${type.type}`"
                        v-model="relationshipModal.selectedType"
                        class="mr-3 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      >
                      <div class="flex-1">
                        <div class="font-medium text-gray-900">{{ type.type }}</div>
                        <div class="text-xs text-gray-500">{{ type.count }} relationships</div>
                      </div>
                    </label>
                  </div>
                </div>

                <!-- Used by section (incoming relationships) -->
                <div v-if="relationshipModal.relationshipTypes.incoming.length > 0">
                  <div class="flex items-center mb-2">
                    <svg class="w-4 h-4 mr-2 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 17l-5-5m0 0l5-5m-5 5h12"></path>
                    </svg>
                    <h4 class="text-sm font-semibold text-gray-700">Used by (incoming)</h4>
                    <span class="ml-2 text-xs text-gray-500">{{ getTotalCount('incoming') }} relationships</span>
                  </div>
                  <div class="space-y-1 ml-6">
                    <label
                      v-for="type in relationshipModal.relationshipTypes.incoming"
                      :key="`incoming-${type.type}`"
                      class="flex items-center p-2 border rounded hover:bg-green-50 cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        :value="`incoming-${type.type}`"
                        v-model="relationshipModal.selectedType"
                        class="mr-3 h-4 w-4 text-green-600 focus:ring-green-500 border-gray-300 rounded"
                      >
                      <div class="flex-1">
                        <div class="font-medium text-gray-900">{{ type.type }}</div>
                        <div class="text-xs text-gray-500">{{ type.count }} relationships</div>
                      </div>
                    </label>
                  </div>
                </div>
              </div>

              <div class="mt-6 flex justify-end space-x-3">
                <button
                  @click="hideRelationshipModal"
                  class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
                >
                  Cancel
                </button>
                <button
                  @click="relationshipModal.mode === 'expand' ? expandSelectedRelationships() : collapseSelectedRelationships()"
                  :disabled="!relationshipModal.selectedType || relationshipModal.selectedType.length === 0"
                  class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 rounded-md"
                >
                  {{ relationshipModal.mode === 'expand' ? 'Expand' : 'Collapse' }} Selected ({{ relationshipModal.selectedType?.length || 0 }})
                </button>
              </div>
            </div>

            <div v-else class="text-center py-8">
              <svg class="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              <h3 class="text-sm font-medium text-gray-900 mb-2">No relationships found</h3>
              <p class="text-sm text-gray-500">
                This CI doesn't have any relationships to expand.
              </p>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { graphAPI, ciAPI, ciTypeAPI } from '@/services/api'
import { showErrorToast, showSuccessToast } from '@/utils/toast'
import type { GraphData, CIType, CI } from '@/types/ci'
import GraphTooltip from '@/components/graph/GraphTooltip.vue'

// Import vis-network
import { Network, DataSet } from 'vis-network/standalone/umd/vis-network.min'

const authStore = useAuthStore()

const loading = ref(false)
const graphContainer = ref(null)
const network = ref(null)
const graphData = ref(null)
const ciTypes = ref([])
const searchResults = ref([])
const showAutocomplete = ref(false)
const isSearching = ref(false)
const searchTimeout = ref<NodeJS.Timeout | null>(null)
const tooltipRef = ref(null)

// Custom tooltip state
const tooltipState = reactive({
  visible: false,
  node: null,
  x: 0,
  y: 0
})

// Context menu state
const contextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  node: null
})

// Relationship options modal state
const relationshipModal = reactive({
  visible: false,
  node: null,
  loading: false,
  relationshipTypes: [],
  selectedType: [],
  mode: 'expand' // 'expand' or 'collapse'
})

const filters = reactive({
  ci_types: [],
  search: '',
  limit: 100,
})

const uniqueCITypes = computed(() => {
  if (!graphData.value) return []
  return [...new Set(graphData.value.nodes.map(node => node.type))]
})

const hasGroupedRelationships = computed(() => {
  if (!relationshipModal.relationshipTypes || typeof relationshipModal.relationshipTypes !== 'object') {
    return false
  }
  return relationshipModal.relationshipTypes.outgoing.length > 0 ||
         relationshipModal.relationshipTypes.incoming.length > 0
})

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const ciTypeColors = {
  'Server': '#3b82f6',
  'Application': '#10b981',
  'Database': '#f59e0b',
  'default': '#6b7280',
}

const getCITypeColor = (type: string) => {
  return ciTypeColors[type as keyof typeof ciTypeColors] || ciTypeColors.default
}

const handleSearchFocus = () => {
  showAutocomplete.value = true
  // Load initial suggestions if search is empty
  if (!filters.search.trim()) {
    fetchSearchSuggestions()
  }
}

const onSearchInputChange = () => {
  // Clear existing timeout
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }

  // Always show autocomplete when typing
  showAutocomplete.value = true

  // Debounce search suggestions
  searchTimeout.value = setTimeout(() => {
    fetchSearchSuggestions()
  }, 300)
}

const fetchSearchSuggestions = async () => {
  if (!filters.search.trim()) {
    // When search is empty, show top 5 CIs + "Show All" option
    isSearching.value = true
    try {
      const response = await ciAPI.list({ limit: 5 })
      let ciResults = []
      if (response.data && response.data.cis) {
        ciResults = response.data.cis.map(ci => ({
          id: ci.id,
          name: ci.name,
          ci_type: ci.ci_type,
          isCIOption: true,
          ci: ci
        }))
      }

      // Add "Show All" option at the top
      ciResults.unshift({
        id: 'show-all',
        name: 'Show All CIs',
        ci_type: 'All',
        isShowAll: true
      })

      searchResults.value = ciResults
    } catch (error) {
      console.error('Failed to fetch CI suggestions:', error)
      searchResults.value = []
    } finally {
      isSearching.value = false
    }
    return
  }

  isSearching.value = true
  try {
    // Search for specific CIs when user types using the list endpoint with search parameter
    const response = await ciAPI.list({
      search: filters.search,
      limit: 5
    })

    // Handle different response structures
    let ciResults = []
    if (response.data && response.data.cis) {
      ciResults = response.data.cis.map(ci => ({
        id: ci.id,
        name: ci.name,
        ci_type: ci.ci_type,
        isCIOption: true,
        ci: ci
      }))
    } else if (response.data) {
      // Handle case where response is directly the data
      ciResults = Array.isArray(response.data) ? response.data.map(ci => ({
        id: ci.id,
        name: ci.name,
        ci_type: ci.ci_type,
        isCIOption: true,
        ci: ci
      })) : []
    }

    searchResults.value = ciResults
    console.log('Search results:', searchResults.value) // Debug log
  } catch (error) {
    console.error('Failed to fetch search suggestions:', error)
    searchResults.value = []
  } finally {
    isSearching.value = false
  }
}

const hideAutocomplete = () => {
  // Delay hiding to allow click events to register
  setTimeout(() => {
    showAutocomplete.value = false
  }, 200)
}

const loadGraphData = async () => {
  if (!hasPermission('ci:read')) return

  loading.value = true
  try {
    const params = {
      ci_types: filters.ci_types.length > 0 ? filters.ci_types : undefined,
      search: filters.search || undefined,
      limit: filters.limit,
    }

    const response = await graphAPI.explore(params)
    graphData.value = response.data

    if (graphData.value) {
      await nextTick()
      renderGraph()
    }
  } catch (error) {
    console.error('Failed to load graph data:', error)
    showErrorToast('Failed to load graph data')
  } finally {
    loading.value = false
  }
}

const selectSearchResult = async (result: any) => {
  if (!hasPermission('ci:read')) return

  let searchQuery = ''
  let searchParams = {}

  if (result.isShowAll) {
    // Show all CIs - clear filters
    filters.search = ''
    filters.ci_types = []
    searchQuery = 'All CIs'
    searchParams = {
      limit: 100
    }
  } else if (result.isCIOption) {
    // Specific CI selected - show only the single CI node initially
    searchQuery = result.name
    filters.search = result.name

    // Create a single node graph for the selected CI
    graphData.value = {
      nodes: [{
        id: result.id,
        name: result.name,
        type: result.ci_type,
        attributes: result.ci.attributes || {},
        x: 0,
        y: 0
      }],
      edges: []
    }

    await nextTick()
    renderGraph()
    showSuccessToast(`Loaded ${searchQuery}`)

    // Don't set loading to false here since we're not making an API call
    loading.value = false
    return
  } else {
    // Fallback for any other type
    searchQuery = result.name
    filters.search = result.name
    searchParams = {
      search: result.name,
      limit: 50
    }
  }

  showAutocomplete.value = false
  searchResults.value = []

  loading.value = true
  try {
    // Use the explore API with the appropriate search parameters
    const response = await graphAPI.explore(searchParams)

    if (response.data) {
      graphData.value = response.data
      await nextTick()
      renderGraph()
      showSuccessToast(`Loaded ${searchQuery}`)
    }
  } catch (error) {
    console.error('Failed to load graph data:', error)
    showErrorToast('Failed to load graph data')
  } finally {
    loading.value = false
  }
}

const loadCITypes = async () => {
  try {
    const response = await ciTypeAPI.list()
    ciTypes.value = response.data.ci_types || []
  } catch (error) {
    console.error('Failed to load CI types:', error)
  }
}

const renderGraph = () => {
  if (!graphContainer.value || !graphData.value) return

  const nodes = new DataSet(
    graphData.value.nodes.map(node => ({
      id: node.id,
      label: node.name,
      color: {
        background: getCITypeColor(node.type),
        border: '#1f2937',
      },
      font: {
        color: '#1f2937',
        size: 14,
      },
      shape: 'dot',
      size: 20,
      // Remove the default tooltip by setting title to undefined
      title: undefined,
    }))
  )

  const edges = new DataSet(
    graphData.value.edges.map(edge => ({
      from: edge.source,
      to: edge.target,
      label: edge.relationship_type, // Show relationship type on arrow
      color: {
        color: '#6b7280',
        highlight: '#3b82f6',
      },
      font: {
        size: 14,
        color: '#1f2937',
        strokeWidth: 3,
        strokeColor: '#ffffff',
        align: 'middle',
        background: '#ffffff',
      },
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 0.8,
          type: 'arrow',
        },
        middle: {
          enabled: false,
        },
        from: {
          enabled: false,
        },
      },
      smooth: {
        type: 'curvedCW',
        roundness: 0.2,
      },
      title: `${edge.relationship_type}: ${graphData.value.nodes.find(n => n.id === edge.source)?.name} → ${graphData.value.nodes.find(n => n.id === edge.target)?.name}`,
    }))
  )

  const data = { nodes, edges }

  const options = {
    layout: {
      improvedLayout: true,
      hierarchical: {
        enabled: false,
      },
    },
    physics: {
      stabilization: {
        iterations: 200,
      },
      barnesHut: {
        gravitationalConstant: -8000,
        centralGravity: 0.3,
        springLength: 95,
        springConstant: 0.04,
        damping: 0.09,
        avoidOverlap: 0.1,
      },
    },
    interaction: {
      hover: true,
      tooltipDelay: 200,
      zoomView: true,
      dragView: true,
      // Disable default tooltips
      hideEdgesOnDrag: true,
    },
    nodes: {
      borderWidth: 2,
      shadow: true,
    },
    edges: {
      width: 2,
      shadow: true,
      smooth: {
        enabled: true,
        type: 'curvedCW',
        roundness: 0.1,
      },
      font: {
        size: 14,
        color: '#1f2937',
        strokeWidth: 3,
        strokeColor: '#ffffff',
        align: 'middle',
        background: '#ffffff',
      },
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 0.8,
        },
      },
    },
  }

  // Destroy existing network if it exists
  if (network.value) {
    network.value.destroy()
  }

  // Create new network
  network.value = new Network(graphContainer.value, data, options)

  // Add event listeners
  network.value.on('click', (params) => {
    if (params.nodes.length > 0) {
      const nodeId = params.nodes[0]
      const node = graphData.value?.nodes.find(n => n.id === nodeId)
      if (node) {
        // Just select the node, don't navigate automatically
        // Users can right-click for details or use other interactions
      }
    }
  })

  // Add hover event listener for custom tooltips
  network.value.on('hoverNode', (params) => {
    const nodeId = params.node
    const node = graphData.value?.nodes.find(n => n.id === nodeId)
    if (node) {
      showTooltip(params.event, node)
    }
  })

  // Add blur event listener to hide tooltip
  network.value.on('blurNode', () => {
    hideTooltip()
  })

  // Hide tooltip when dragging starts
  network.value.on('dragStart', () => {
    hideTooltip()
  })

  // Hide tooltip when zooming
  network.value.on('zoom', () => {
    hideTooltip()
  })

  // Add right-click event listener with better event handling
  network.value.on('oncontext', (params) => {
    if (params.nodes.length > 0) {
      const nodeId = params.nodes[0]
      const node = graphData.value?.nodes.find(n => n.id === nodeId)
      if (node) {
        // Prevent default context menu
        const event = params.event
        if (event) {
          event.preventDefault()
          event.stopPropagation()
          showContextMenu(event, node)
        }
      }
    }
  })
}

// Custom tooltip methods
const showTooltip = (event, node) => {
  hideTooltip() // Clear any existing tooltip

  if (!network.value) {
    return
  }

  // Use the mouse event position directly and offset it slightly above the cursor
  const x = event.clientX
  const y = event.clientY - 10 // Position slightly above cursor

  tooltipState.node = node
  tooltipState.x = x
  tooltipState.y = y
  tooltipState.visible = true
}

const hideTooltip = () => {
  tooltipState.visible = false
  tooltipState.node = null

  // Clean up tooltip component if needed
  if (tooltipRef.value?.cleanup) {
    tooltipRef.value.cleanup()
  }
}

const showRelationshipOptions = async (node: any) => {
  if (!hasPermission('ci:read')) return

  relationshipModal.node = node
  relationshipModal.visible = true
  relationshipModal.loading = true
  relationshipModal.selectedType = []
  relationshipModal.mode = 'expand'
  hideContextMenu()

  try {
    // Get relationships for this node to determine available relationship types
    const response = await graphAPI.explore({
      center_id: node.id,
      limit: 200,
      depth: 1
    })

    if (response.data && response.data.edges) {
      // Filter edges to only include those directly connected to the selected node
      const connectedEdges = response.data.edges.filter(edge =>
        edge.source === node.id || edge.target === node.id
      )

      // Group relationships by direction and type
      const groupedRelationships = {
        outgoing: {}, // Node is source: "Using"
        incoming: {}  // Node is target: "Used by"
      }

      connectedEdges.forEach(edge => {
        const type = edge.relationship_type
        const direction = edge.source === node.id ? 'outgoing' : 'incoming'

        if (!groupedRelationships[direction][type]) {
          groupedRelationships[direction][type] = {
            type,
            count: 0,
            direction,
            edges: []
          }
        }
        groupedRelationships[direction][type].count++
        groupedRelationships[direction][type].edges.push(edge)
      })

      // Convert to arrays and sort
      const outgoing = Object.values(groupedRelationships.outgoing)
      const incoming = Object.values(groupedRelationships.incoming)

      relationshipModal.relationshipTypes = {
        outgoing,
        incoming
      }
      console.log('Grouped relationships for node:', node.name, relationshipModal.relationshipTypes) // Debug log
    } else {
      relationshipModal.relationshipTypes = { outgoing: [], incoming: [] }
    }
  } catch (error) {
    console.error('Failed to load relationship types:', error)
    relationshipModal.relationshipTypes = { outgoing: [], incoming: [] }
  } finally {
    relationshipModal.loading = false
  }
}

const showCollapseOptions = async (node: any) => {
  if (!hasPermission('ci:read')) return

  relationshipModal.node = node
  relationshipModal.visible = true
  relationshipModal.loading = true
  relationshipModal.selectedType = []
  relationshipModal.mode = 'collapse'
  hideContextMenu()

  try {
    // Get current graph data to determine what relationship types are currently visible
    if (graphData.value && graphData.value.edges) {
      // Group relationships by type and count them
      const relationshipTypes = graphData.value.edges.reduce((acc, edge) => {
        const type = edge.relationship_type
        if (!acc[type]) {
          acc[type] = { type, count: 0 }
        }
        acc[type].count++
        return acc
      }, {})

      relationshipModal.relationshipTypes = Object.values(relationshipTypes)
      console.log('Available relationship types for collapse:', relationshipModal.relationshipTypes) // Debug log
    } else {
      relationshipModal.relationshipTypes = []
    }
  } catch (error) {
    console.error('Failed to load relationship types for collapse:', error)
    relationshipModal.relationshipTypes = []
  } finally {
    relationshipModal.loading = false
  }
}

const collapseSelectedRelationships = async () => {
  if (!relationshipModal.selectedType || relationshipModal.selectedType.length === 0) return

  try {
    const node = relationshipModal.node

    if (graphData.value && graphData.value.edges) {
      const allNodes = graphData.value.nodes
      const allEdges = graphData.value.edges

      // Filter out edges that match the selected relationship types
      const remainingEdges = allEdges.filter(edge =>
        !relationshipModal.selectedType.includes(edge.relationship_type)
      )

      // Get nodes that should remain (either original node or connected to remaining edges)
      const nodeIdsToKeep = new Set([node.id])
      remainingEdges.forEach(edge => {
        nodeIdsToKeep.add(edge.source)
        nodeIdsToKeep.add(edge.target)
      })

      const remainingNodes = allNodes.filter(n => nodeIdsToKeep.has(n.id))

      // Update graph data with filtered results
      graphData.value.nodes = remainingNodes
      graphData.value.edges = remainingEdges

      await nextTick()
      renderGraph()
      showSuccessToast(`Collapsed ${relationshipModal.selectedType.join(', ')} relationships`)
    }
  } catch (error) {
    console.error('Failed to collapse relationships:', error)
    showErrorToast('Failed to collapse relationships')
  } finally {
    hideRelationshipModal()
  }
}

const getTotalCount = (direction: 'outgoing' | 'incoming') => {
  if (!relationshipModal.relationshipTypes || !relationshipModal.relationshipTypes[direction]) {
    return 0
  }
  return relationshipModal.relationshipTypes[direction].reduce((total, type) => total + type.count, 0)
}

const parseSelectedTypes = () => {
  const result = {
    types: [],
    directions: {}
  }

  if (!relationshipModal.selectedType) return result

  relationshipModal.selectedType.forEach(selection => {
    const [direction, type] = selection.split('-')
    if (!result.types.includes(type)) {
      result.types.push(type)
    }
    result.directions[type] = direction
  })

  return result
}

const hideRelationshipModal = () => {
  relationshipModal.visible = false
  relationshipModal.node = null
  relationshipModal.selectedType = []
  relationshipModal.relationshipTypes = []
}

const expandSelectedRelationships = async () => {
  if (!relationshipModal.selectedType || relationshipModal.selectedType.length === 0) return

  try {
    relationshipModal.loading = true
    const node = relationshipModal.node
    const parsedSelections = parseSelectedTypes()

    // Get relationships for the node
    const response = await graphAPI.explore({
      center_id: node.id,
      limit: 200,
      depth: 1
    })

    if (response.data) {
      const allNodes = response.data.nodes || []
      const allEdges = response.data.edges || []

      // Filter edges based on selected types and directions
      const filteredEdges = allEdges.filter(edge => {
        // Only include edges connected to the selected node
        if (edge.source !== node.id && edge.target !== node.id) {
          return false
        }

        // Check if this edge type is selected and matches the direction
        const edgeDirection = edge.source === node.id ? 'outgoing' : 'incoming'
        const selectionKey = `${edgeDirection}-${edge.relationship_type}`

        return relationshipModal.selectedType.includes(selectionKey)
      })

      // Get only nodes that are connected by the filtered edges
      const connectedNodeIds = new Set([node.id])
      filteredEdges.forEach(edge => {
        connectedNodeIds.add(edge.source)
        connectedNodeIds.add(edge.target)
      })

      const filteredNodes = allNodes.filter(n => connectedNodeIds.has(n.id))

      if (graphData.value) {
        // Create sets to avoid duplicates
        const existingNodeIds = new Set(graphData.value.nodes.map(n => n.id))
        const existingEdgeIds = new Set(graphData.value.edges.map(e => `${e.source}-${e.target}`))

        // Add only new nodes and filtered edges
        const uniqueNewNodes = filteredNodes.filter(n => !existingNodeIds.has(n.id))
        const uniqueNewEdges = filteredEdges.filter(e => !existingEdgeIds.has(`${e.source}-${e.target}`))

        graphData.value.nodes.push(...uniqueNewNodes)
        graphData.value.edges.push(...uniqueNewEdges)
      } else {
        graphData.value = {
          nodes: filteredNodes,
          edges: filteredEdges
        }
      }

      await nextTick()
      renderGraph()

      const relationshipCount = relationshipModal.selectedType.length
      showSuccessToast(`Expanded ${relationshipCount} relationship group(s) for ${node.name}`)
    }
  } catch (error) {
    console.error('Failed to expand relationships:', error)
    showErrorToast('Failed to expand relationships')
  } finally {
    relationshipModal.loading = false
    hideRelationshipModal()
  }
}

const viewNodeDetails = (node: any) => {
  window.location.href = `/ci/${node.id}`
  hideContextMenu()
}

const showContextMenu = (event, node) => {
  // Prevent default browser context menu
  if (event.preventDefault) {
    event.preventDefault()
  }
  if (event.stopPropagation) {
    event.stopPropagation()
  }

  contextMenu.visible = true
  contextMenu.x = event.clientX
  contextMenu.y = event.clientY
  contextMenu.node = node
}

const hideContextMenu = () => {
  contextMenu.visible = false
  contextMenu.node = null
}

const centerGraph = () => {
  if (network.value) {
    network.value.fit()
  }
}

const fitGraph = () => {
  if (network.value) {
    network.value.fit({
      animation: {
        duration: 1000,
        easingFunction: 'easeInOutQuad',
      },
    })
  }
}

const clearGraph = () => {
  graphData.value = { nodes: [], edges: [] }
  if (network.value) {
    network.value.destroy()
    network.value = null
  }
  filters.search = ''
  filters.ci_types = []
  searchResults.value = []
  showAutocomplete.value = false
  hideTooltip()
}

// Close context menu when clicking outside
document.addEventListener('click', (e) => {
  if (!e.target || !(e.target as Element).closest('[style*="z-index: 9999"]')) {
    hideContextMenu()
  }
})

onMounted(async () => {
  await loadCITypes()
  // Don't load graph data initially - wait for user to search
  // This ensures the graph starts empty
})

onUnmounted(() => {
  if (network.value) {
    network.value.destroy()
  }
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }
  hideTooltip()
})
</script>

<style scoped>
/* Add vis-network styles */
/* @import url('vis-network/styles/dist/vis-network.min.css'); */
</style>
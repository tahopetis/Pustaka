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
                  @input="debouncedSearch"
                  @focus="showAutocomplete = true"
                  @blur="hideAutocomplete"
                >
                <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                  </svg>
                </div>
                <!-- Autocomplete dropdown -->
                <div v-if="showAutocomplete && searchResults.length > 0" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
                  <div
                    v-for="result in searchResults"
                    :key="result.id"
                    class="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center justify-between"
                    @mousedown="selectSearchResult(result)"
                  >
                    <div>
                      <div class="font-medium">{{ result.name }}</div>
                      <div class="text-sm text-gray-500">{{ result.type }}</div>
                    </div>
                    <div class="w-2 h-2 rounded-full" :style="{ backgroundColor: getCITypeColor(result.type) }"></div>
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

          <!-- Empty State -->
          <div v-if="!loading && (!graphData || graphData.nodes.length === 0)" class="text-center py-16">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No graph data available</h3>
            <p class="mt-1 text-sm text-gray-500">
              Try adjusting your filters or create some configuration items with relationships.
            </p>
          </div>

          <!-- Graph Canvas -->
          <div
            v-show="!loading && graphData && graphData.nodes.length > 0"
            ref="graphContainer"
            class="w-full"
            style="height: 600px;"
            @contextmenu.prevent="handleRightClick"
          ></div>

          <!-- Context Menu -->
          <div
            v-if="contextMenu.visible"
            :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
            class="absolute z-20 bg-white border border-gray-200 rounded-lg shadow-lg py-1"
            @click="hideContextMenu"
          >
            <button
              v-if="contextMenu.node"
              @click="expandNode(contextMenu.node)"
              class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 flex items-center"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"></path>
              </svg>
              Expand Node
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { graphAPI, ciAPI, ciTypeAPI } from '@/services/api'
import { showErrorToast, showSuccessToast } from '@/utils/toast'
import type { GraphData, CIType, CI } from '@/types/ci'

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

// Context menu state
const contextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  node: null
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

const debouncedSearch = debounce(() => {
  loadGraphData()
  if (filters.search.length >= 2) {
    searchCIs()
  } else {
    searchResults.value = []
  }
}, 500)

function debounce(func, wait) {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
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

const searchCIs = async () => {
  if (!hasPermission('ci:read') || filters.search.length < 2) return

  try {
    const response = await ciAPI.search({
      search: filters.search,
      limit: 10
    })
    searchResults.value = response.data.cis || []
  } catch (error) {
    console.error('Failed to search CIs:', error)
  }
}

const selectSearchResult = (ci) => {
  filters.search = ci.name
  showAutocomplete.value = false
  loadGraphData()
}

const hideAutocomplete = () => {
  setTimeout(() => {
    showAutocomplete.value = false
  }, 200)
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
      title: `${node.name} (${node.type})\n${JSON.stringify(node.attributes, null, 2)}`,
    }))
  )

  const edges = new DataSet(
    graphData.value.edges.map(edge => ({
      from: edge.source,
      to: edge.target,
      label: edge.type,
      color: {
        color: '#6b7280',
        highlight: '#3b82f6',
      },
      font: {
        size: 12,
        align: 'middle',
        background: '#ffffff',
      },
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 0.5,
        },
      },
      title: `${edge.type}\n${JSON.stringify(edge.attributes || {}, null, 2)}`,
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
    },
    nodes: {
      borderWidth: 2,
      shadow: true,
    },
    edges: {
      width: 2,
      shadow: true,
      smooth: {
        type: 'continuous',
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
        // Navigate to CI details
        window.location.href = `/ci/${nodeId}`
      }
    }
  })

  // Add right-click event listener
  network.value.on('oncontext', (params) => {
    if (params.nodes.length > 0) {
      const nodeId = params.nodes[0]
      const node = graphData.value?.nodes.find(n => n.id === nodeId)
      if (node) {
        showContextMenu(params.event, node)
      }
    }
  })
}

const expandNode = async (node: any) => {
  if (!hasPermission('ci:read')) return

  try {
    showSuccessToast(`Expanding connections for ${node.name}...`)
    // TODO: Implement node expansion logic
    // This could call a new API endpoint to get connected nodes
    console.log('Expand node:', node)
  } catch (error) {
    console.error('Failed to expand node:', error)
    showErrorToast('Failed to expand node')
  }
  hideContextMenu()
}

const viewNodeDetails = (node: any) => {
  window.location.href = `/ci/${node.id}`
  hideContextMenu()
}

const handleRightClick = (event) => {
  event.preventDefault()
}

const showContextMenu = (event, node) => {
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
}

// Close context menu when clicking outside
document.addEventListener('click', (e) => {
  if (!e.target || !(e.target as Element).closest('.absolute.z-20')) {
    hideContextMenu()
  }
})

onMounted(async () => {
  await loadCITypes()
  await loadGraphData()
})

onUnmounted(() => {
  if (network.value) {
    network.value.destroy()
  }
})
</script>

<style scoped>
/* Add vis-network styles */
/* @import url('vis-network/styles/dist/vis-network.min.css'); */
</style>
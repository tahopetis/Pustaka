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
              <input
                v-model="filters.search"
                type="text"
                placeholder="Search CI names..."
                class="form-input"
                @input="debouncedSearch"
              >
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
          ></div>
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
import { graphAPI, ciTypeAPI } from '@/services/api'
import { showErrorToast } from '@/utils/toast'
import type { GraphData, CIType } from '@/types/ci'

// Import vis-network
import { Network, DataSet } from 'vis-network/standalone/umd/vis-network.min'

const authStore = useAuthStore()

const loading = ref(false)
const graphContainer = ref<HTMLElement | null>(null)
const network = ref<Network | null>(null)
const graphData = ref<GraphData | null>(null)
const ciTypes = ref<CIType[]>([])

const filters = reactive({
  ci_types: [] as string[],
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
}, 500)

function debounce(func: Function, wait: number) {
  let timeout: NodeJS.Timeout
  return function executedFunction(...args: any[]) {
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
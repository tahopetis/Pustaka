<template>
  <div class="relative">
    <!-- Search Button -->
    <button
      @click="openSearch"
      class="flex items-center px-4 py-2 bg-white border border-gray-300 rounded-lg shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
      :class="{ 'w-full': fullWidth }"
    >
      <svg class="w-5 h-5 text-gray-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <span class="text-gray-500">Search...</span>
      <kbd class="ml-auto hidden sm:inline-flex items-center px-2 py-1 text-xs font-semibold text-gray-400 bg-gray-100 border border-gray-300 rounded">
        ⌘K
      </kbd>
    </button>

    <!-- Search Modal -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-50 overflow-y-auto"
        @click.self="closeSearch"
      >
        <div class="flex items-start justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
          <!-- Background overlay -->
          <Transition
            enter-active-class="transition ease-out duration-200"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition ease-in duration-150"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <div
              v-if="isOpen"
              class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              @click="closeSearch"
            />
          </Transition>

          <!-- Modal panel -->
          <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full">
            <!-- Search Input -->
            <div class="p-4 border-b border-gray-200">
              <div class="relative">
                <input
                  ref="searchInput"
                  v-model="searchQuery"
                  @input="handleSearch"
                  @keydown.escape="closeSearch"
                  @keydown.down="moveSelection('down')"
                  @keydown.up="moveSelection('up')"
                  @keydown.enter="selectResult"
                  type="text"
                  class="w-full pl-10 pr-4 py-3 text-lg border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Search CIs by name, ID, or tag..."
                />
                <svg class="absolute left-3 top-3.5 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <div v-if="searchQuery" class="absolute right-3 top-3 flex items-center space-x-2">
                  <span class="text-xs text-gray-400">{{ results?.length ?? 0 }} results</span>
                  <button
                    @click="clearSearch"
                    class="text-gray-400 hover:text-gray-600"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <!-- Search Results -->
            <div class="max-h-96 overflow-y-auto">
              <!-- Loading State -->
              <div v-if="loading" class="p-8 text-center">
                <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                <p class="mt-2 text-sm text-gray-500">Searching...</p>
              </div>

              <!-- Error State -->
              <div v-else-if="error" class="p-8 text-center">
                <svg class="w-12 h-12 mx-auto text-red-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="text-sm text-gray-900 font-medium">Search failed</p>
                <p class="text-xs text-gray-500 mt-1">{{ error }}</p>
              </div>

              <!-- No Results -->
              <div v-else-if="searchQuery && results.length === 0" class="p-8 text-center">
                <svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <p class="text-sm text-gray-900 font-medium">No results found</p>
                <p class="text-xs text-gray-500 mt-1">Try different keywords</p>
              </div>

              <!-- Empty State -->
              <div v-else-if="!searchQuery" class="p-8 text-center">
                <svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <p class="text-sm text-gray-900 font-medium">Start typing to search</p>
                <p class="text-xs text-gray-500 mt-1">Search by CI name, ID, or tag</p>
                <div class="mt-4 flex items-center justify-center space-x-4 text-xs text-gray-500">
                  <span class="flex items-center">
                    <kbd class="px-2 py-1 bg-gray-100 border border-gray-300 rounded mr-1">↑↓</kbd>
                    Navigate
                  </span>
                  <span class="flex items-center">
                    <kbd class="px-2 py-1 bg-gray-100 border border-gray-300 rounded mr-1">↵</kbd>
                    Select
                  </span>
                  <span class="flex items-center">
                    <kbd class="px-2 py-1 bg-gray-100 border border-gray-300 rounded mr-1">esc</kbd>
                    Close
                  </span>
                </div>
              </div>

              <!-- Results List -->
              <ul v-else class="divide-y divide-gray-200">
                <li
                  v-for="(result, index) in results"
                  :key="result.id"
                  @click="navigateToCI(result.id)"
                  class="cursor-pointer hover:bg-blue-50"
                  :class="{ 'bg-blue-50': selectedIndex === index }"
                >
                  <div class="px-4 py-3 sm:px-6">
                    <div class="flex items-center justify-between">
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-blue-600 truncate">
                          {{ result.name }}
                        </p>
                        <p class="text-xs text-gray-500 truncate">
                          {{ result.ci_type }} • ID: {{ result.id.slice(0, 8) }}
                        </p>
                        <div class="mt-1 flex flex-wrap gap-1">
                          <span
                            v-for="tag in result.tags?.slice(0, 3)"
                            :key="tag"
                            class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800"
                          >
                            {{ tag }}
                          </span>
                        </div>
                      </div>
                      <svg class="w-5 h-5 text-gray-400 ml-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>
                </li>
              </ul>
            </div>

            <!-- Recent Searches -->
            <div v-if="!searchQuery && recentSearches.length > 0" class="p-4 border-t border-gray-200 bg-gray-50">
              <p class="text-xs font-medium text-gray-500 mb-2">Recent Searches</p>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="(search, index) in recentSearches.slice(0, 5)"
                  :key="index"
                  @click="searchQuery = search; handleSearch()"
                  class="inline-flex items-center px-3 py-1 text-sm bg-white border border-gray-300 rounded-full hover:bg-gray-50"
                >
                  <svg class="w-3 h-3 text-gray-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ search }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/services/api'

interface SearchResult {
  id: string
  name: string
  ci_type: string
  tags: string[]
}

interface Props {
  fullWidth?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullWidth: false
})

const router = useRouter()
const isOpen = ref(false)
const searchQuery = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const selectedIndex = ref(0)
const searchInput = ref<HTMLInputElement | null>(null)

// Load recent searches from localStorage
const recentSearches = ref<string[]>([])
try {
  const stored = localStorage.getItem('recentSearches')
  if (stored) {
    recentSearches.value = JSON.parse(stored)
  }
} catch {
  recentSearches.value = []
}

// Debounced search
let searchTimeout: NodeJS.Timeout | null = null
const handleSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  searchTimeout = setTimeout(async () => {
    if (!searchQuery.value.trim()) {
      results.value = []
      return
    }

    loading.value = true
    error.value = null

    try {
      const response = await api.get<{ cis: SearchResult[] }>('/ci', {
        params: {
          search: searchQuery.value,
          limit: 10
        }
      })
      results.value = response.data.cis || []
      selectedIndex.value = 0

      // Save to recent searches
      if (!recentSearches.value.includes(searchQuery.value)) {
        recentSearches.value = [searchQuery.value, ...recentSearches.value].slice(0, 10)
        localStorage.setItem('recentSearches', JSON.stringify(recentSearches.value))
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Search failed'
      results.value = []
    } finally {
      loading.value = false
    }
  }, 300)
}

const openSearch = () => {
  isOpen.value = true
  nextTick(() => {
    searchInput.value?.focus()
  })
}

const closeSearch = () => {
  isOpen.value = false
  searchQuery.value = ''
  results.value = []
  selectedIndex.value = 0
}

const clearSearch = () => {
  searchQuery.value = ''
  results.value = []
  searchInput.value?.focus()
}

const moveSelection = (direction: 'up' | 'down') => {
  if (results.value.length === 0) return

  if (direction === 'down') {
    selectedIndex.value = (selectedIndex.value + 1) % results.value.length
  } else {
    selectedIndex.value = selectedIndex.value === 0 ? results.value.length - 1 : selectedIndex.value - 1
  }
}

const selectResult = () => {
  if (results.value[selectedIndex.value]) {
    navigateToCI(results.value[selectedIndex.value].id)
  }
}

const navigateToCI = (ciId: string) => {
  closeSearch()
  router.push({ name: 'CIDetails', params: { id: ciId } })
}

// Keyboard shortcut handler
const handleKeydown = (e: KeyboardEvent) => {
  // Cmd/Ctrl + K to open search
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    if (isOpen.value) {
      closeSearch()
    } else {
      openSearch()
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

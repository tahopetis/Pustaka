<template>
  <div class="page-container page-content">
    <!-- Header -->
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">CI Type Management</h1>
        <p class="page-subtitle">
          Define and manage configuration item type schemas
        </p>
      </div>
      <BaseButton @click="showCreateModal = true">
        <Icon name="plus" class="w-4 h-4 mr-2" />
        New CI Type
      </BaseButton>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white shadow rounded-lg p-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <BaseInput
            v-model="searchQuery"
            placeholder="Search CI types..."
            type="search"
            clearable
          >
            <template #prefix>
              <Icon name="search" class="w-5 h-5 text-gray-400" />
            </template>
          </BaseInput>
        </div>
      </div>
    </div>

    <!-- CI Types List -->
    <div class="bg-white shadow rounded-lg">
      <div v-if="loading" class="flex justify-center items-center py-12">
        <BaseSpinner class="w-8 h-8" />
      </div>

      <div v-else-if="ciTypes.length === 0" class="text-center py-12">
        <Icon name="document-text" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">No CI types found</h3>
        <p class="text-gray-500 mb-4">Get started by creating your first CI type schema</p>
        <BaseButton @click="showCreateModal = true" variant="outline">
          <Icon name="plus" class="w-4 h-4 mr-2" />
          Create CI Type
        </BaseButton>
      </div>

      <div v-else class="overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Description
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Required Attributes
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Optional Attributes
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Usage Count
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="ciType in filteredCITypes" :key="ciType.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div>
                    <div class="text-sm font-medium text-gray-900">
                      {{ ciType.name }}
                    </div>
                    <div class="text-sm text-gray-500">
                      ID: {{ ciType.id.substring(0, 8) }}...
                    </div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="text-sm text-gray-900">
                  {{ ciType.description || 'No description' }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="attr in ciType.required_attributes.slice(0, 3)"
                    :key="attr.name"
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-800"
                  >
                    {{ attr.name }}
                  </span>
                  <span
                    v-if="ciType.required_attributes.length > 3"
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800"
                  >
                    +{{ ciType.required_attributes.length - 3 }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="attr in ciType.optional_attributes.slice(0, 3)"
                    :key="attr.name"
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800"
                  >
                    {{ attr.name }}
                  </span>
                  <span
                    v-if="ciType.optional_attributes.length > 3"
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800"
                  >
                    +{{ ciType.optional_attributes.length - 3 }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center text-sm text-gray-900">
                  <Icon name="cube" class="w-4 h-4 mr-1 text-gray-400" />
                  {{ getUsageCount(ciType.name) }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <div class="flex justify-end gap-2">
                  <button
                    @click="viewCIType(ciType)"
                    class="text-indigo-600 hover:text-indigo-900"
                    title="View details"
                  >
                    <Icon name="eye" class="w-4 h-4" />
                  </button>
                  <button
                    @click="editCIType(ciType)"
                    class="text-indigo-600 hover:text-indigo-900"
                    title="Edit"
                  >
                    <Icon name="pencil" class="w-4 h-4" />
                  </button>
                  <button
                    @click="deleteCIType(ciType)"
                    class="text-red-600 hover:text-red-900"
                    title="Delete"
                  >
                    <Icon name="trash" class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-between">
      <div class="text-sm text-gray-700">
        Showing {{ (pagination.page - 1) * pagination.limit + 1 }} to
        {{ Math.min(pagination.page * pagination.limit, pagination.total) }} of
        {{ pagination.total }} results
      </div>
      <div class="flex gap-2">
        <BaseButton
          @click="previousPage"
          :disabled="pagination.page <= 1"
          variant="outline"
          size="sm"
        >
          Previous
        </BaseButton>
        <BaseButton
          @click="nextPage"
          :disabled="pagination.page >= pagination.total_pages"
          variant="outline"
          size="sm"
        >
          Next
        </BaseButton>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <CITypeModal
      v-if="showCreateModal || showEditModal"
      :ci-type="editingCIType"
      :is-edit="showEditModal"
      @close="closeModal"
      @saved="onCISaved"
    />

    <!-- View Details Modal -->
    <CITypeDetailsModal
      v-if="showViewModal"
      :ci-type="viewingCIType"
      @close="showViewModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useCITypesStore } from '@/stores/ciTypes'
import { useNotificationStore } from '@/stores/notification'
import CITypeModal from '@/components/ci/CITypeModal.vue'
import CITypeDetailsModal from '@/components/ci/CITypeDetailsModal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSpinner from '@/components/base/BaseSpinner.vue'
import Icon from '@/components/base/Icon.vue'

interface CIType {
  id: string
  name: string
  description?: string
  required_attributes: Array<{
    name: string
    type: string
    description: string
    validation?: any
  }>
  optional_attributes: Array<{
    name: string
    type: string
    description: string
    validation?: any
  }>
  created_at: string
  updated_at: string
}

interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

const ciTypesStore = useCITypesStore()
const notificationStore = useNotificationStore()

const ciTypes = ref<CIType[]>([])
const loading = ref(false)
const searchQuery = ref('')
const pagination = ref<Pagination>({
  page: 1,
  limit: 20,
  total: 0,
  total_pages: 0
})

// Modal states
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showViewModal = ref(false)
const editingCIType = ref<CIType | null>(null)
const viewingCIType = ref<CIType | null>(null)

const usageStats = ref<Record<string, number>>({})

const filteredCITypes = computed(() => {
  if (!searchQuery.value) return ciTypes.value

  const query = searchQuery.value.toLowerCase()
  return ciTypes.value.filter(ciType =>
    ciType.name.toLowerCase().includes(query) ||
    (ciType.description && ciType.description.toLowerCase().includes(query))
  )
})

const loadCITypes = async () => {
  try {
    loading.value = true
    const response = await ciTypesStore.listCITypes(
      pagination.value.page,
      pagination.value.limit,
      searchQuery.value
    )
    ciTypes.value = response.ci_types
    pagination.value = {
      page: response.page,
      limit: response.limit,
      total: response.total,
      total_pages: response.total_pages
    }
  } catch (error) {
    notificationStore.showError('Failed to load CI types')
    console.error('Error loading CI types:', error)
  } finally {
    loading.value = false
  }
}

const loadUsageStats = async () => {
  try {
    const stats = await ciTypesStore.getCITypesByUsage()
    if (stats && Array.isArray(stats)) {
      usageStats.value = stats.reduce((acc, stat) => {
        acc[stat.type] = stat.count
        return acc
      }, {} as Record<string, number>)
    } else {
      usageStats.value = {}
    }
  } catch (error) {
    console.error('Error loading usage stats:', error)
    usageStats.value = {}
  }
}

const getUsageCount = (ciTypeName: string) => {
  return usageStats.value[ciTypeName] || 0
}

const viewCIType = (ciType: CIType) => {
  viewingCIType.value = ciType
  showViewModal.value = true
}

const editCIType = (ciType: CIType) => {
  editingCIType.value = { ...ciType }
  showEditModal.value = true
}

const deleteCIType = async (ciType: CIType) => {
  if (!confirm(`Are you sure you want to delete the CI type "${ciType.name}"? This action cannot be undone.`)) {
    return
  }

  try {
    await ciTypesStore.deleteCIType(ciType.id)
    notificationStore.showSuccess('CI type deleted successfully')
    await loadCITypes()
    await loadUsageStats()
  } catch (error: any) {
    if (error.message.includes('cannot delete CI type with existing CIs')) {
      notificationStore.showError('Cannot delete CI type with existing configuration items')
    } else {
      notificationStore.showError('Failed to delete CI type')
    }
    console.error('Error deleting CI type:', error)
  }
}

const closeModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingCIType.value = null
}

const onCISaved = async () => {
  closeModal()
  await loadCITypes()
  await loadUsageStats()
  notificationStore.showSuccess(
    showEditModal.value ? 'CI type updated successfully' : 'CI type created successfully'
  )
}

const previousPage = () => {
  if (pagination.value.page > 1) {
    pagination.value.page--
    loadCITypes()
  }
}

const nextPage = () => {
  if (pagination.value.page < pagination.value.total_pages) {
    pagination.value.page++
    loadCITypes()
  }
}

onMounted(() => {
  loadCITypes()
  loadUsageStats()
})
</script>
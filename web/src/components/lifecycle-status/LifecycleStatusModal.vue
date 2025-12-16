<template>
  <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
    <div class="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-lg bg-white">
      <!-- Header -->
      <div class="flex justify-between items-center pb-4 border-b">
        <h3 class="text-lg font-semibold text-gray-900">
          {{ isEdit ? 'Edit Lifecycle Status' : 'Create Lifecycle Status' }}
        </h3>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="py-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Left Column - Basic Information -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Name *
              </label>
              <input
                v-model="form.name"
                type="text"
                placeholder="e.g., planned, operational, retired"
                class="form-input"
                required
                :disabled="isEdit"
                pattern="[a-z0-9_]+"
                title="Must contain only lowercase letters, numbers, and underscores"
              />
              <p class="text-xs text-gray-500 mt-1">
                Unique identifier (lowercase, numbers, underscores only)
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Display Name *
              </label>
              <input
                v-model="form.display_name"
                type="text"
                placeholder="e.g., Planned, Operational, Retired"
                class="form-input"
                required
              />
              <p class="text-xs text-gray-500 mt-1">
                Human-readable name for this status
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                v-model="form.description"
                placeholder="Describe when this status applies to a CI"
                rows="3"
                class="form-input"
              ></textarea>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Sort Order
              </label>
              <input
                v-model.number="form.sort_order"
                type="number"
                min="0"
                placeholder="10"
                class="form-input"
              />
              <p class="text-xs text-gray-500 mt-1">
                Order in which statuses appear (lower numbers appear first)
              </p>
            </div>
          </div>

          <!-- Right Column - Visual Settings -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Color
              </label>
              <div class="flex items-center space-x-2">
                <input
                  v-model="form.color"
                  type="color"
                  class="h-10 w-20"
                />
                <input
                  v-model="form.color"
                  type="text"
                  placeholder="#22c55e"
                  pattern="^#[0-9A-Fa-f]{6}$"
                  class="form-input flex-1"
                />
              </div>
              <p class="text-xs text-gray-500 mt-1">
                Color for visual representation in UI
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Icon
              </label>
              <select v-model="form.icon" class="form-input">
                <option value="">Select an icon</option>
                <option value="calendar">Calendar (📅)</option>
                <option value="package">Package (📦)</option>
                <option value="archive">Archive (📋)</option>
                <option value="clock">Clock (🕐)</option>
                <option value="check-circle">Check Circle (✅)</option>
                <option value="wrench">Wrench (🔧)</option>
                <option value="alert-triangle">Alert Triangle (⚠️)</option>
                <option value="power-off">Power Off (🔌)</option>
                <option value="trash-2">Trash (🗑️)</option>
                <option value="x-circle">X Circle (❌)</option>
              </select>
              <p class="text-xs text-gray-500 mt-1">
                Icon to represent this status
              </p>
            </div>

            <!-- Preview -->
            <div v-if="form.display_name || form.name" class="border rounded-lg p-4 bg-gray-50">
              <h4 class="text-sm font-medium text-gray-700 mb-2">Preview</h4>
              <div class="flex items-center">
                <div
                  v-if="form.color"
                  class="w-3 h-3 rounded-full mr-2"
                  :style="{ backgroundColor: form.color }"
                ></div>
                <span class="font-medium">{{ form.display_name || form.name }}</span>
                <span v-if="form.display_name && form.name" class="text-xs text-gray-500 ml-1">({{ form.name }})</span>
                <div v-if="form.icon" class="ml-2 flex items-center text-xs text-gray-500">
                  <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="getIconPath(form.icon)"></path>
                  </svg>
                  {{ form.icon }}
                </div>
              </div>
              <p v-if="form.description" class="text-sm text-gray-600 mt-1">
                {{ form.description }}
              </p>
            </div>
          </div>
        </div>

        <!-- Status Settings -->
        <div v-if="!isEdit" class="mt-6 p-4 bg-blue-50 rounded-lg">
          <h4 class="text-sm font-medium text-blue-900 mb-2">New Status Information</h4>
          <p class="text-sm text-blue-700">
            This will be created as a custom status. System statuses cannot be modified after creation.
          </p>
        </div>

        <div v-else class="mt-6 space-y-4">
          <div class="flex items-center">
            <input
              v-model="form.is_active"
              type="checkbox"
              id="is_active"
              class="form-checkbox"
            />
            <label for="is_active" class="ml-2 text-sm text-gray-700">
              Active Status
            </label>
          </div>
          <p class="text-xs text-gray-500">
            Inactive statuses cannot be assigned to new CIs but remain for historical data.
          </p>
        </div>

        <!-- Form Actions -->
        <div class="mt-8 flex justify-end gap-3 pt-4 border-t">
          <button
            type="button"
            @click="$emit('close')"
            class="btn btn-outline"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="btn btn-primary"
          >
            <span v-if="isSubmitting" class="spinner w-4 h-4 mr-2"></span>
            {{ isEdit ? 'Update' : 'Create' }} Lifecycle Status
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useLifecycleStatusStore, type LifecycleStatus } from '@/stores/lifecycleStatus'
import { showSuccessToast, showErrorToast } from '@/utils/toast'

interface Props {
  lifecycleStatus?: LifecycleStatus | null
  isOpen: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'saved', lifecycleStatus: LifecycleStatus): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const lifecycleStatusStore = useLifecycleStatusStore()

const isSubmitting = ref(false)
const isEdit = computed(() => !!props.lifecycleStatus)

const form = ref({
  name: '',
  display_name: '',
  description: '',
  color: '#22c55e',
  icon: '',
  sort_order: 10,
  is_active: true
})

// Icon path mapping
const getIconPath = (iconName: string) => {
  const iconPaths: Record<string, string> = {
    'calendar': 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
    'package': 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
    'archive': 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4',
    'clock': 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    'check-circle': 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    'wrench': 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z',
    'alert-triangle': 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.314 16.5c-.77.833.192 2.5 1.732 2.5z',
    'power-off': 'M18.364 5.636l-1.414 1.414M9.172 9.172L5.636 5.636M12 2v4m0 4v4m0 8a7 7 0 110-14 7 7 0 010 14z',
    'trash-2': 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
    'x-circle': 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
  }
  return iconPaths[iconName] || 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1'
}

// Watch for changes in props
watch(() => props.lifecycleStatus, (newStatus) => {
  if (newStatus) {
    form.value = {
      name: newStatus.name,
      display_name: newStatus.display_name,
      description: newStatus.description || '',
      color: newStatus.color || '#22c55e',
      icon: newStatus.icon || '',
      sort_order: newStatus.sort_order,
      is_active: newStatus.is_active
    }
  } else {
    // Reset form for new status
    form.value = {
      name: '',
      display_name: '',
      description: '',
      color: '#22c55e',
      icon: '',
      sort_order: 10,
      is_active: true
    }
  }
}, { immediate: true })

const validateForm = () => {
  if (!form.value.name.trim()) {
    showErrorToast('Name is required')
    return false
  }

  if (!form.value.display_name.trim()) {
    showErrorToast('Display name is required')
    return false
  }

  // Validate name format
  if (!/^[a-z0-9_]+$/.test(form.value.name)) {
    showErrorToast('Name must contain only lowercase letters, numbers, and underscores')
    return false
  }

  return true
}

const handleSubmit = async () => {
  if (!validateForm()) return

  isSubmitting.value = true

  try {
    let savedLifecycleStatus: LifecycleStatus

    if (isEdit.value && props.lifecycleStatus) {
      // Update existing status
      const updateData = {
        display_name: form.value.display_name,
        description: form.value.description,
        color: form.value.color,
        icon: form.value.icon,
        sort_order: form.value.sort_order,
        is_active: form.value.is_active
      }
      savedLifecycleStatus = await lifecycleStatusStore.updateLifecycleStatus(
        props.lifecycleStatus.id,
        updateData
      )
      showSuccessToast('Lifecycle status updated successfully')
    } else {
      // Create new status
      const createData = {
        name: form.value.name,
        display_name: form.value.display_name,
        description: form.value.description,
        color: form.value.color,
        icon: form.value.icon,
        sort_order: form.value.sort_order
      }
      savedLifecycleStatus = await lifecycleStatusStore.createLifecycleStatus(createData)
      showSuccessToast('Lifecycle status created successfully')
    }

    emit('saved', savedLifecycleStatus)
    emit('close')
  } catch (error: any) {
    console.error('Failed to save lifecycle status:', error)
    // Error handling is already done in the store
  } finally {
    isSubmitting.value = false
  }
}
</script>
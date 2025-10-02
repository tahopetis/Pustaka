<template>
  <Transition
    enter-active-class="transform ease-out duration-300 transition"
    enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
    enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
    leave-active-class="transition ease-in duration-100"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="notification"
      :class="[
        'max-w-sm w-full bg-white shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden',
        positionClasses
      ]"
    >
      <div class="p-4">
        <div class="flex items-start">
          <div class="flex-shrink-0">
            <Icon
              :name="iconName"
              :class="iconClasses"
              class="h-6 w-6"
            />
          </div>
          <div class="ml-3 w-0 flex-1 pt-0.5">
            <p class="text-sm font-medium text-gray-900">
              {{ notification.title }}
            </p>
            <p v-if="notification.message" class="mt-1 text-sm text-gray-500">
              {{ notification.message }}
            </p>
          </div>
          <div class="ml-4 flex-shrink-0 flex">
            <button
              @click="close"
              class="bg-white rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <span class="sr-only">Close</span>
              <Icon name="x" class="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import type { Notification } from '@/stores/notification'
import Icon from '@/components/base/Icon.vue'

interface Props {
  notification: Notification
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'
}

const props = withDefaults(defineProps<Props>(), {
  position: 'top-right'
})

const emit = defineEmits<{
  close: []
}>()

const iconName = computed(() => {
  switch (props.notification.type) {
    case 'success':
      return 'check-circle'
    case 'error':
      return 'x-circle'
    case 'warning':
      return 'exclamation-triangle'
    case 'info':
      return 'information-circle'
    default:
      return 'information-circle'
  }
})

const iconClasses = computed(() => {
  switch (props.notification.type) {
    case 'success':
      return 'text-green-400'
    case 'error':
      return 'text-red-400'
    case 'warning':
      return 'text-yellow-400'
    case 'info':
      return 'text-blue-400'
    default:
      return 'text-blue-400'
  }
})

const positionClasses = computed(() => {
  switch (props.position) {
    case 'top-right':
      return 'fixed top-4 right-4'
    case 'top-left':
      return 'fixed top-4 left-4'
    case 'bottom-right':
      return 'fixed bottom-4 right-4'
    case 'bottom-left':
      return 'fixed bottom-4 left-4'
    default:
      return 'fixed top-4 right-4'
  }
})

let closeTimer: number | null = null

const close = () => {
  emit('close')
}

const startTimer = () => {
  if (props.notification.duration && props.notification.duration > 0) {
    closeTimer = window.setTimeout(() => {
      close()
    }, props.notification.duration)
  }
}

const clearTimer = () => {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

onMounted(() => {
  startTimer()
})

onUnmounted(() => {
  clearTimer()
})
</script>
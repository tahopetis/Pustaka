<template>
  <Teleport to="body">
    <div>
      <!-- Top Right Notifications -->
      <div class="fixed top-4 right-4 z-50 space-y-2">
        <NotificationToast
          v-for="notification in topRightNotifications"
          :key="notification.id"
          :notification="notification"
          position="top-right"
          @close="removeNotification(notification.id)"
        />
      </div>

      <!-- Top Left Notifications -->
      <div class="fixed top-4 left-4 z-50 space-y-2">
        <NotificationToast
          v-for="notification in topLeftNotifications"
          :key="notification.id"
          :notification="notification"
          position="top-left"
          @close="removeNotification(notification.id)"
        />
      </div>

      <!-- Bottom Right Notifications -->
      <div class="fixed bottom-4 right-4 z-50 space-y-2">
        <NotificationToast
          v-for="notification in bottomRightNotifications"
          :key="notification.id"
          :notification="notification"
          position="bottom-right"
          @close="removeNotification(notification.id)"
        />
      </div>

      <!-- Bottom Left Notifications -->
      <div class="fixed bottom-4 left-4 z-50 space-y-2">
        <NotificationToast
          v-for="notification in bottomLeftNotifications"
          :key="notification.id"
          :notification="notification"
          position="bottom-left"
          @close="removeNotification(notification.id)"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useNotificationStore } from '@/stores/notification'
import NotificationToast from './NotificationToast.vue'

const notificationStore = useNotificationStore()

// Group notifications by position (for now, all go to top-right)
const topRightNotifications = computed(() => notificationStore.notifications.slice(0, 3))
const topLeftNotifications = computed(() => [])
const bottomRightNotifications = computed(() => [])
const bottomLeftNotifications = computed(() => [])

const removeNotification = (id: string) => {
  notificationStore.removeNotification(id)
}
</script>
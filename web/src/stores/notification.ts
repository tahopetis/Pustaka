import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<Notification[]>([])

  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const id = crypto.randomUUID()
    const newNotification = { ...notification, id }

    notifications.value.push(newNotification)

    // Auto remove after duration (default 5 seconds)
    const duration = notification.duration ?? 5000
    if (duration > 0) {
      setTimeout(() => {
        removeNotification(id)
      }, duration)
    }

    return id
  }

  const removeNotification = (id: string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }

  const clearAll = () => {
    notifications.value = []
  }

  const showSuccess = (message: string, title?: string) => {
    return addNotification({
      type: 'success',
      title: title ?? 'Success',
      message,
      duration: 5000
    })
  }

  const showError = (message: string, title?: string) => {
    return addNotification({
      type: 'error',
      title: title ?? 'Error',
      message,
      duration: 8000 // Longer duration for errors
    })
  }

  const showWarning = (message: string, title?: string) => {
    return addNotification({
      type: 'warning',
      title: title ?? 'Warning',
      message,
      duration: 6000
    })
  }

  const showInfo = (message: string, title?: string) => {
    return addNotification({
      type: 'info',
      title: title ?? 'Info',
      message,
      duration: 5000
    })
  }

  return {
    notifications,
    addNotification,
    removeNotification,
    clearAll,
    showSuccess,
    showError,
    showWarning,
    showInfo
  }
})
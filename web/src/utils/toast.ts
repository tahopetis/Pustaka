// Simple toast utility using browser console for now
// Replace with actual toast library when needed

export const showToast = (message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info') => {
  const logMethod = type === 'error' ? 'error' : type === 'warning' ? 'warn' : 'log'
  console[logMethod](`[${type.toUpperCase()}] ${message}`)
}

// Specific toast helpers
export const showSuccessToast = (message: string) => showToast(message, 'success')
export const showErrorToast = (message: string) => showToast(message, 'error')
export const showInfoToast = (message: string) => showToast(message, 'info')
export const showWarningToast = (message: string) => showToast(message, 'warning')
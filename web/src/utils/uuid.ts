// UUID utility with fallback for older browsers

export function generateUUID(): string {
  // Check if crypto.randomUUID() is available
  if (typeof crypto !== 'undefined' && crypto.randomUUID && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  // Fallback implementation for older browsers
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

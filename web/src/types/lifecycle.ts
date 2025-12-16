// Lifecycle Status Types for Pustaka CMDB

export interface LifecycleStatus {
  id: string
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  sort_order: number
  is_active: boolean
  is_system: boolean
  created_at: string
  updated_at?: string
  created_by: string
  updated_by?: string
}

export interface CreateLifecycleStatusRequest {
  name: string
  display_name: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
}

export interface UpdateLifecycleStatusRequest {
  display_name?: string
  description?: string
  color?: string
  icon?: string
  sort_order?: number
  is_active?: boolean
}

export interface LifecycleStatusListResponse {
  lifecycle_statuses: LifecycleStatus[]
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface ListLifecycleStatusFilters {
  search?: string
  is_active?: boolean
  is_system?: boolean
  sort?: string
  order?: string
}

export interface LifecycleStatusUsage {
  lifecycle_status: LifecycleStatus
  usage_count: number
}

export interface LifecycleStatusUsageResponse {
  total_cis: number
  cis_with_status: number
  cis_without_status: number
  status_usage: LifecycleStatusUsage[]
  status_distribution: Record<string, number>
}

export interface CIStatusDistribution {
  status_name: string
  display_name: string
  color: string
  icon: string
  count: number
  percentage: number
}

export interface CIStatusDistributionByType {
  ci_type_name: string
  statuses: CIStatusDistribution[]
  total_cis: number
}

// Validation helpers
export interface LifecycleStatusValidationError {
  field: string
  message: string
}

export interface LifecycleStatusValidationResult {
  is_valid: boolean
  errors: LifecycleStatusValidationError[]
}

// Form validation rules
export const LifecycleStatusValidationRules = {
  name: {
    required: true,
    minLength: 2,
    maxLength: 100,
    pattern: /^[a-z0-9_]+$/,
    message: 'Name must contain only lowercase letters, numbers, and underscores'
  },
  display_name: {
    required: true,
    minLength: 2,
    maxLength: 100,
    message: 'Display name is required and must be between 2 and 100 characters'
  },
  description: {
    maxLength: 500,
    message: 'Description must be less than 500 characters'
  },
  color: {
    pattern: /^#[0-9A-Fa-f]{6}$/,
    message: 'Color must be a valid hex color code (e.g., #FF5733)'
  },
  icon: {
    maxLength: 50,
    message: 'Icon must be less than 50 characters'
  },
  sort_order: {
    min: 0,
    message: 'Sort order must be 0 or greater'
  }
}

// Default values
export const DefaultLifecycleStatusColors = [
  '#94a3b8', // slate-400 (Planned)
  '#3b82f6', // blue-500 (On Order)
  '#10b981', // emerald-500 (In Stock)
  '#f59e0b', // amber-500 (Pending Install)
  '#22c55e', // green-500 (Operational)
  '#f97316', // orange-500 (In Maintenance)
  '#ef4444', // red-500 (Defective/Repair)
  '#6b7280', // gray-500 (Retired)
  '#4b5563', // gray-600 (Disposed)
  '#991b1b'  // red-800 (Missing/Stolen)
]

export const DefaultLifecycleStatusIcons = [
  'calendar',    // Planned
  'package',     // On Order
  'archive',     // In Stock
  'clock',       // Pending Install
  'check-circle', // Operational
  'wrench',      // In Maintenance
  'alert-triangle', // Defective/Repair
  'power-off',   // Retired
  'trash-2',     // Disposed
  'x-circle'     // Missing/Stolen
]

// Helper functions
export function validateLifecycleStatusName(name: string): string | null {
  if (!name) {
    return 'Name is required'
  }
  if (name.length < 2 || name.length > 100) {
    return 'Name must be between 2 and 100 characters'
  }
  if (!/^[a-z0-9_]+$/.test(name)) {
    return 'Name must contain only lowercase letters, numbers, and underscores'
  }
  return null
}

export function validateLifecycleStatusDisplayName(displayName: string): string | null {
  if (!displayName) {
    return 'Display name is required'
  }
  if (displayName.length < 2 || displayName.length > 100) {
    return 'Display name must be between 2 and 100 characters'
  }
  return null
}

export function validateLifecycleStatusColor(color: string): string | null {
  if (color && !/^#[0-9A-Fa-f]{6}$/.test(color)) {
    return 'Color must be a valid hex color code (e.g., #FF5733)'
  }
  return null
}

export function validateLifecycleStatusSortOrder(sortOrder: number): string | null {
  if (sortOrder < 0) {
    return 'Sort order must be 0 or greater'
  }
  return null
}

export function validateCreateLifecycleStatusRequest(data: Partial<CreateLifecycleStatusRequest>): LifecycleStatusValidationResult {
  const errors: LifecycleStatusValidationError[] = []

  const nameError = validateLifecycleStatusName(data.name || '')
  if (nameError) {
    errors.push({ field: 'name', message: nameError })
  }

  const displayNameError = validateLifecycleStatusDisplayName(data.display_name || '')
  if (displayNameError) {
    errors.push({ field: 'display_name', message: displayNameError })
  }

  const colorError = validateLifecycleStatusColor(data.color || '')
  if (colorError) {
    errors.push({ field: 'color', message: colorError })
  }

  if (data.sort_order !== undefined) {
    const sortOrderError = validateLifecycleStatusSortOrder(data.sort_order)
    if (sortOrderError) {
      errors.push({ field: 'sort_order', message: sortOrderError })
    }
  }

  if (data.description && data.description.length > 500) {
    errors.push({ field: 'description', message: 'Description must be less than 500 characters' })
  }

  if (data.icon && data.icon.length > 50) {
    errors.push({ field: 'icon', message: 'Icon must be less than 50 characters' })
  }

  return {
    is_valid: errors.length === 0,
    errors
  }
}

export function validateUpdateLifecycleStatusRequest(data: Partial<UpdateLifecycleStatusRequest>): LifecycleStatusValidationResult {
  const errors: LifecycleStatusValidationError[] = []

  if (data.display_name !== undefined) {
    const displayNameError = validateLifecycleStatusDisplayName(data.display_name || '')
    if (displayNameError) {
      errors.push({ field: 'display_name', message: displayNameError })
    }
  }

  if (data.color !== undefined) {
    const colorError = validateLifecycleStatusColor(data.color || '')
    if (colorError) {
      errors.push({ field: 'color', message: colorError })
    }
  }

  if (data.sort_order !== undefined) {
    const sortOrderError = validateLifecycleStatusSortOrder(data.sort_order)
    if (sortOrderError) {
      errors.push({ field: 'sort_order', message: sortOrderError })
    }
  }

  if (data.description !== undefined && data.description.length > 500) {
    errors.push({ field: 'description', message: 'Description must be less than 500 characters' })
  }

  if (data.icon !== undefined && data.icon.length > 50) {
    errors.push({ field: 'icon', message: 'Icon must be less than 50 characters' })
  }

  return {
    is_valid: errors.length === 0,
    errors
  }
}
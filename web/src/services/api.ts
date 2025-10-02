import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { showToast } from '@/utils/toast'

// Create axios instance
export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()

    // Add authorization header if token exists
    if (authStore.accessToken) {
      config.headers['Authorization'] = `Bearer ${authStore.accessToken}`
    }

    // Add request ID for tracking
    config.headers['X-Request-ID'] = crypto.randomUUID()

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const authStore = useAuthStore()

    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      // Don't retry on auth endpoints to prevent infinite loops
      if (error.config?.url?.includes('/auth/login') ||
          error.config?.url?.includes('/auth/refresh')) {
        return Promise.reject(error)
      }

      // Try to refresh the token
      const refreshed = await authStore.refreshAccessToken()

      if (refreshed && error.config) {
        // Retry the original request with new token
        error.config.headers['Authorization'] = `Bearer ${authStore.accessToken}`
        return api.request(error.config)
      } else {
        // Refresh failed, logout user but don't force redirect
        await authStore.logout()
      }
    }

    // Handle 403 Forbidden
    if (error.response?.status === 403) {
      showToast('You do not have permission to perform this action', 'error')
    }

    // Handle 422 Validation errors
    if (error.response?.status === 422) {
      const errors = error.response.data?.error?.details || {}
      const errorMessage = Object.values(errors).flat().join('. ') || 'Validation failed'
      showToast(errorMessage, 'error')
    }

    // Handle 500 Server errors
    if (error.response?.status >= 500) {
      const errorMessage = error.response.data?.error?.message || 'Server error occurred. Please try again later.'
      showToast(errorMessage, 'error')
      console.error('Server Error Details:', {
        status: error.response.status,
        statusText: error.response.statusText,
        data: error.response.data,
        url: error.config?.url,
        method: error.config?.method,
        requestData: error.config?.data
      })
    }

    // Handle network errors
    if (!error.response) {
      showToast('Network error. Please check your connection.', 'error')
    }

    return Promise.reject(error)
  }
)

// API service methods
export const authAPI = {
  login: (credentials: { username: string; password: string }) =>
    api.post('/auth/login', credentials),

  refreshToken: (refreshToken: string) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),

  getCurrentUser: () => api.get('/me'),
}

export const ciAPI = {
  list: (params?: {
    page?: number
    limit?: number
    ci_type?: string
    search?: string
    tags?: string[]
    sort?: string
    order?: string
    attributes?: Record<string, any>
  }) => api.get('/ci', { params }),

  get: (id: string) => api.get(`/ci/${id}`),

  create: (data: {
    name: string
    ci_type: string
    attributes: Record<string, any>
    tags?: string[]
  }) => api.post('/ci', data),

  update: (id: string, data: {
    attributes?: Record<string, any>
    tags?: string[]
  }) => api.put(`/ci/${id}`, data),

  delete: (id: string) => api.delete(`/ci/${id}`),
}

export const ciTypeAPI = {
  list: (params?: { page?: number; limit?: number; search?: string }) =>
    api.get('/ci-types', { params }),

  get: (id: string) => api.get(`/ci-types/${id}`),

  create: (data: {
    name: string
    description?: string
    required_attributes: any[]
    optional_attributes?: any[]
  }) => api.post('/ci-types', data),

  update: (id: string, data: {
    description?: string
    required_attributes?: any[]
    optional_attributes?: any[]
  }) => api.put(`/ci-types/${id}`, data),

  delete: (id: string) => api.delete(`/ci-types/${id}`),

  getUsage: () => api.get('/analytics/ci-types/usage'),
}

export const relationshipAPI = {
  list: (params?: {
    page?: number
    limit?: number
    source_id?: string
    target_id?: string
    relationship_type?: string
    search?: string
    sort?: string
    order?: string
  }) => api.get('/relationships', { params }),

  get: (id: string) => api.get(`/relationships/${id}`),

  create: (data: {
    source_id: string
    target_id: string
    relationship_type: string
    attributes?: Record<string, any>
  }) => api.post('/relationships', data),

  update: (id: string, data: {
    attributes?: Record<string, any>
  }) => api.put(`/relationships/${id}`, data),

  delete: (id: string) => api.delete(`/relationships/${id}`),
}

export const graphAPI = {
  getNodes: (id: string) => api.get(`/graph/nodes/${id}`),
  explore: (params?: {
    center_id?: string
    depth?: number
    node_types?: string[]
    relationship_types?: string[]
  }) => api.get('/graph/explore', { params }),
}

export const auditAPI = {
  list: (params?: {
    page?: number
    limit?: number
    entity?: string
    entity_id?: string
    action?: string
    performed_by?: string
    from_date?: string
    to_date?: string
  }) => api.get('/audit', { params }),

  get: (id: string) => api.get(`/audit/${id}`),
}

export const userAPI = {
  list: (params?: { page?: number; limit?: number; search?: string }) =>
    api.get('/users', { params }),

  get: (id: string) => api.get(`/users/${id}`),

  create: (data: {
    username: string
    email: string
    password: string
    roles?: string[]
  }) => api.post('/users', data),

  update: (id: string, data: {
    email?: string
    is_active?: boolean
    roles?: string[]
  }) => api.put(`/users/${id}`, data),

  delete: (id: string) => api.delete(`/users/${id}`),
}
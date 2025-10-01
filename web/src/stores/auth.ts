import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/services/api'

export interface User {
  id: string
  username: string
  email: string
  is_active: boolean
  roles: Array<{
    id: string
    name: string
    description: string
  }>
  permissions: string[]
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)

  const isAuthenticated = computed(() => {
    // Check if we have both token and user data, or just token (for initial refresh state)
    if (accessToken.value && user.value && user.value.id) {
      return true
    }
    // For refresh scenarios, if we have a token in localStorage, consider as potentially authenticated
    // until checkAuth() confirms the user data
    if (accessToken.value && !user.value && !isInitialized.value) {
      return false // Don't consider authenticated until user is loaded
    }
    return false
  })

  // Add states to track if auth has been initialized and is loading
  const isInitialized = ref(false)
  const isLoading = ref(true)

  const hasPermission = (permission: string) => {
    return user.value?.permissions.includes(permission) || false
  }

  const hasRole = (role: string) => {
    return user.value?.roles.some(r => r.name === role) || false
  }

  const hasAnyPermission = (permissions: string[]) => {
    return permissions.some(permission => hasPermission(permission))
  }

  const hasAllPermissions = (permissions: string[]) => {
    return permissions.every(permission => hasPermission(permission))
  }

  const login = async (credentials: LoginCredentials): Promise<void> => {
    try {
      const response = await api.post<AuthResponse>('/auth/login', credentials)

      const { access_token, refresh_token, user: userData } = response.data

      accessToken.value = access_token
      refreshToken.value = refresh_token || null
      user.value = userData

      // Store tokens in localStorage
      localStorage.setItem('access_token', access_token)
      if (refresh_token) {
        localStorage.setItem('refresh_token', refresh_token)
      }

      // Set default Authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`

    } catch (error) {
      throw error
    }
  }

  const logout = async () => {
    try {
      // Clear tokens
      accessToken.value = null
      refreshToken.value = null
      user.value = null

      // Remove from localStorage
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')

      // Clear Authorization header
      delete api.defaults.headers.common['Authorization']

    } catch (error) {
      console.error('Logout error:', error)
    }
  }

  const refreshAccessToken = async (): Promise<boolean> => {
    try {
      const storedRefreshToken = localStorage.getItem('refresh_token')
      if (!storedRefreshToken) {
        return false
      }

      const response = await api.post<{ access_token: string; refresh_token: string }>('/auth/refresh', {
        refresh_token: storedRefreshToken
      })

      const { access_token, refresh_token } = response.data

      accessToken.value = access_token
      refreshToken.value = refresh_token

      // Update localStorage
      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)

      // Update Authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`

      return true
    } catch (error) {
      // Refresh failed - clear tokens without calling logout to prevent infinite loops
      accessToken.value = null
      refreshToken.value = null
      user.value = null

      // Remove from localStorage
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')

      // Clear Authorization header
      delete api.defaults.headers.common['Authorization']

      return false
    }
  }

  const checkAuth = async (): Promise<boolean> => {
    console.log('checkAuth: Starting authentication check...')
    isLoading.value = true
    const storedToken = localStorage.getItem('access_token')

    if (!storedToken) {
      console.log('checkAuth: No stored token found')
      isInitialized.value = true
      isLoading.value = false
      return false
    }

    console.log('checkAuth: Found stored token, validating with API...')
    try {
      // Set token for API requests
      api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`
      accessToken.value = storedToken

      // Get current user info
      console.log('checkAuth: Calling /me endpoint...')
      const response = await api.get<{ user: User }>('/me')
      console.log('checkAuth: Successfully got user data:', response.data.user)
      user.value = response.data.user
      isInitialized.value = true
      isLoading.value = false

      return true
    } catch (error) {
      console.log('checkAuth: API call failed, error:', error)
      // Token is invalid, try to refresh
      console.log('checkAuth: Attempting to refresh token...')
      const refreshed = await refreshAccessToken()
      if (refreshed) {
        console.log('checkAuth: Token refresh successful, getting user info...')
        // After successful refresh, try to get user info again
        try {
          const response = await api.get<{ user: User }>('/me')
          console.log('checkAuth: Successfully got user data after refresh:', response.data.user)
          user.value = response.data.user
          isInitialized.value = true
          isLoading.value = false
          return true
        } catch (userError) {
          console.log('checkAuth: Still failed to get user info after refresh:', userError)
          // If we still can't get user info, clear auth without calling logout
          accessToken.value = null
          refreshToken.value = null
          user.value = null
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          delete api.defaults.headers.common['Authorization']
          isInitialized.value = true
          isLoading.value = false
          return false
        }
      }
      console.log('checkAuth: Token refresh failed, clearing auth state')
      isInitialized.value = true
      isLoading.value = false
      return false
    }
  }

  const getUserProfile = async (): Promise<User> => {
    try {
      const response = await api.get<{ user: User }>('/me')
      user.value = response.data.user
      return user.value
    } catch (error) {
      throw error
    }
  }

  // Load session from localStorage and determine initial auth state
  const loadSessionFromLocalStorage = (): boolean => {
    const storedToken = localStorage.getItem('access_token')
    const storedRefreshToken = localStorage.getItem('refresh_token')

    if (storedToken) {
      accessToken.value = storedToken
      refreshToken.value = storedRefreshToken
      api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`
      return true
    }

    // No token found, mark as initialized and not loading
    isInitialized.value = true
    isLoading.value = false
    return false
  }

  // Initialize auth state
  const init = () => {
    const hasToken = loadSessionFromLocalStorage()
    if (hasToken) {
      // We have a token, but need to validate it and get user data
      // Keep loading state true until checkAuth() completes
      console.log('Found stored token, validating...')
    } else {
      console.log('No stored token found')
    }
  }

  // Initialize on store creation
  init()

  return {
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    isInitialized,
    isLoading,
    login,
    logout,
    refreshAccessToken,
    checkAuth,
    getUserProfile,
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAllPermissions
  }
})
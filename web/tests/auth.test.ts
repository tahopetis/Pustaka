import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import axios from 'axios'

// Mock axios
vi.mock('axios')

describe('Auth Store', () => {
  let authStore: any

  beforeEach(() => {
    // Create a fresh pinia instance for each test
    setActivePinia(createPinia())
    authStore = useAuthStore()

    // Clear localStorage
    localStorage.clear()

    // Reset all mocks
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial state', () => {
      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
      expect(authStore.refreshToken).toBeNull()
      expect(authStore.isAuthenticated).toBe(false)
    })
  })

  describe('Login', () => {
    it('should login successfully and set tokens', async () => {
      const mockUser = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        roles: [],
        permissions: []
      }

      const mockResponse = {
        data: {
          access_token: 'mock-access-token',
          token_type: 'Bearer',
          expires_in: 3600,
          user: mockUser
        }
      }

      vi.mocked(axios.post).mockResolvedValueOnce(mockResponse)

      await authStore.login({
        username: 'testuser',
        password: 'password123'
      })

      expect(authStore.user).toEqual(mockUser)
      expect(authStore.accessToken).toBe('mock-access-token')
      expect(authStore.isAuthenticated).toBe(true)
      expect(localStorage.getItem('access_token')).toBe('mock-access-token')
      expect(axios.defaults.headers.common['Authorization']).toBe('Bearer mock-access-token')
    })

    it('should handle login failure', async () => {
      const mockError = new Error('Invalid credentials')
      mockError.response = { status: 401, data: { error: 'Invalid credentials' } }

      vi.mocked(axios.post).mockRejectedValueOnce(mockError)

      await expect(authStore.login({
        username: 'testuser',
        password: 'wrong-password'
      })).rejects.toThrow('Invalid credentials')

      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
      expect(authStore.isAuthenticated).toBe(false)
    })
  })

  describe('Logout', () => {
    it('should clear all auth data', async () => {
      // Set up initial authenticated state
      authStore.user = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        roles: [],
        permissions: []
      }
      authStore.accessToken = 'mock-token'
      authStore.refreshToken = 'mock-refresh-token'
      localStorage.setItem('access_token', 'mock-token')
      localStorage.setItem('refresh_token', 'mock-refresh-token')

      await authStore.logout()

      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
      expect(authStore.refreshToken).toBeNull()
      expect(authStore.isAuthenticated).toBe(false)
      expect(localStorage.getItem('access_token')).toBeNull()
      expect(localStorage.getItem('refresh_token')).toBeNull()
      expect(axios.defaults.headers.common['Authorization']).toBeUndefined()
    })
  })

  describe('Permission Checks', () => {
    beforeEach(() => {
      authStore.user = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        roles: [],
        permissions: ['ci:read', 'ci:create', 'user:read']
      }
    })

    it('should check individual permissions correctly', () => {
      expect(authStore.hasPermission('ci:read')).toBe(true)
      expect(authStore.hasPermission('ci:create')).toBe(true)
      expect(authStore.hasPermission('user:read')).toBe(true)
      expect(authStore.hasPermission('ci:delete')).toBe(false)
    })

    it('should check role permissions correctly', () => {
      authStore.user.roles = [
        { id: '1', name: 'admin', description: 'Administrator' },
        { id: '2', name: 'editor', description: 'Editor' }
      ]

      expect(authStore.hasRole('admin')).toBe(true)
      expect(authStore.hasRole('editor')).toBe(true)
      expect(authStore.hasRole('viewer')).toBe(false)
    })

    it('should check any permission correctly', () => {
      expect(authStore.hasAnyPermission('ci:read', 'ci:delete')).toBe(true)
      expect(authStore.hasAnyPermission('ci:delete', 'user:delete')).toBe(false)
    })

    it('should check all permissions correctly', () => {
      expect(authStore.hasAllPermissions('ci:read', 'ci:create')).toBe(true)
      expect(authStore.hasAllPermissions('ci:read', 'ci:delete')).toBe(false)
    })
  })

  describe('Token Refresh', () => {
    it('should refresh access token successfully', async () => {
      localStorage.setItem('refresh_token', 'mock-refresh-token')

      const mockResponse = {
        data: {
          access_token: 'new-access-token',
          refresh_token: 'new-refresh-token'
        }
      }

      vi.mocked(axios.post).mockResolvedValueOnce(mockResponse)

      const result = await authStore.refreshAccessToken()

      expect(result).toBe(true)
      expect(authStore.accessToken).toBe('new-access-token')
      expect(authStore.refreshToken).toBe('new-refresh-token')
      expect(localStorage.getItem('access_token')).toBe('new-access-token')
      expect(localStorage.getItem('refresh_token')).toBe('new-refresh-token')
    })

    it('should handle token refresh failure', async () => {
      localStorage.setItem('refresh_token', 'expired-refresh-token')

      const mockError = new Error('Token expired')
      mockError.response = { status: 401 }

      vi.mocked(axios.post).mockRejectedValueOnce(mockError)

      const result = await authStore.refreshAccessToken()

      expect(result).toBe(false)
      expect(authStore.accessToken).toBeNull()
      expect(authStore.refreshToken).toBeNull()
      expect(localStorage.getItem('access_token')).toBeNull()
      expect(localStorage.getItem('refresh_token')).toBeNull()
    })
  })

  describe('Check Auth', () => {
    it('should verify existing valid token', async () => {
      localStorage.setItem('access_token', 'valid-token')

      const mockUser = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        roles: [],
        permissions: []
      }

      vi.mocked(axios.get).mockResolvedValueOnce({ data: { user: mockUser } })

      const result = await authStore.checkAuth()

      expect(result).toBe(true)
      expect(authStore.user).toEqual(mockUser)
      expect(authStore.accessToken).toBe('valid-token')
    })

    it('should handle invalid token by trying refresh', async () => {
      localStorage.setItem('access_token', 'invalid-token')
      localStorage.setItem('refresh_token', 'valid-refresh-token')

      const mockUser = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        roles: [],
        permissions: []
      }

      // First call fails (invalid token)
      vi.mocked(axios.get).mockRejectedValueOnce(new Error('Invalid token'))

      // Refresh succeeds
      vi.mocked(axios.post).mockResolvedValueOnce({
        data: {
          access_token: 'new-token',
          refresh_token: 'new-refresh-token'
        }
      })

      // Get user with new token succeeds
      vi.mocked(axios.get).mockResolvedValueOnce({ data: { user: mockUser } })

      const result = await authStore.checkAuth()

      expect(result).toBe(true)
      expect(authStore.accessToken).toBe('new-token')
    })
  })
})
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/ci',
      name: 'ConfigurationItems',
      component: () => import('@/views/ci/CIListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/ci/new',
      name: 'CreateCI',
      component: () => import('@/views/ci/CIFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'ci:create' }
    },
    {
      path: '/ci/:id',
      name: 'CIDetails',
      component: () => import('@/views/ci/CIDetailsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/ci/:id/edit',
      name: 'EditCI',
      component: () => import('@/views/ci/CIFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'ci:update' }
    },
    {
      path: '/ci-types',
      name: 'CITypes',
      component: () => import('@/views/ci/CITypeManagementView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/relationships',
      name: 'Relationships',
      component: () => import('@/views/relationships/RelationshipListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/graph',
      name: 'Graph',
      component: () => import('@/views/graph/GraphView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/audit',
      name: 'Audit',
      component: () => import('@/views/audit/AuditView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'audit:read' }
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('@/views/users/UserListView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'user:read' }
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/profile/ProfileView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/NotFoundView.vue')
    }
  ]
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  console.log('Router guard:', {
    to: to.path,
    isLoading: authStore.isLoading,
    isInitialized: authStore.isInitialized,
    isAuthenticated: authStore.isAuthenticated,
    hasAccessToken: !!authStore.accessToken,
    hasUser: !!authStore.user
  })

  // If authentication is not initialized, ensure checkAuth is called
  if (!authStore.isInitialized) {
    console.log('Auth not initialized, calling checkAuth...')
    await authStore.checkAuth()
    console.log('checkAuth completed, auth state updated')
  }

  // If authentication is still loading after checkAuth, wait for it
  if (authStore.isLoading) {
    console.log('Auth is still loading, waiting for completion...')

    // Create a promise that resolves when auth is initialized
    await new Promise<void>((resolve) => {
      const checkAuth = () => {
        if (!authStore.isLoading) {
          console.log('Auth loading completed')
          resolve()
        } else {
          setTimeout(checkAuth, 50) // Check every 50ms
        }
      }
      checkAuth()
    })
  }

  console.log('Final auth state:', {
    isLoading: authStore.isLoading,
    isInitialized: authStore.isInitialized,
    isAuthenticated: authStore.isAuthenticated,
    hasUser: !!authStore.user
  })

  // Check if route requires authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    console.log('Route requires auth but user not authenticated, redirecting to login')
    next('/login')
    return
  }

  // Check if route requires specific permission
  if (to.meta.requiresPermission && !authStore.hasPermission(to.meta.requiresPermission)) {
    console.log('Route requires specific permission but user lacks it, redirecting to dashboard')
    next('/dashboard')
    return
  }

  // Redirect authenticated users away from login page
  if (to.name === 'Login' && authStore.isAuthenticated) {
    console.log('Authenticated user accessing login, redirecting to dashboard')
    next('/dashboard')
    return
  }

  console.log('Router guard allowing navigation to:', to.path)
  next()
})

export default router
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
      path: '/lifecycle-status',
      name: 'LifecycleStatus',
      component: () => import('@/views/lifecycle-statuses/LifecycleStatusListView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'lifecycle_status:read' }
    },
    {
      path: '/lifecycle-status/new',
      name: 'CreateLifecycleStatus',
      component: () => import('@/views/lifecycle-statuses/LifecycleStatusFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'lifecycle_status:create' }
    },
    {
      path: '/lifecycle-status/:id',
      name: 'LifecycleStatusDetails',
      component: () => import('@/views/lifecycle-statuses/LifecycleStatusDetailsView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'lifecycle_status:read' }
    },
    {
      path: '/lifecycle-status/:id/edit',
      name: 'EditLifecycleStatus',
      component: () => import('@/views/lifecycle-statuses/LifecycleStatusFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'lifecycle_status:update' }
    },
    {
      path: '/relationships',
      name: 'Relationships',
      component: () => import('@/views/relationships/RelationshipListView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/relationships/new',
      name: 'CreateRelationship',
      component: () => import('@/views/relationships/RelationshipFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship:create' }
    },
    {
      path: '/relationships/:id',
      name: 'RelationshipDetails',
      component: () => import('@/views/relationships/RelationshipDetailsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/relationships/:id/edit',
      name: 'EditRelationship',
      component: () => import('@/views/relationships/RelationshipFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship:update' }
    },
    {
      path: '/relationship-types',
      name: 'RelationshipTypes',
      component: () => import('@/views/relationship-types/RelationshipTypeListView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship_type:read' }
    },
    {
      path: '/relationship-types/new',
      name: 'CreateRelationshipType',
      component: () => import('@/views/relationship-types/RelationshipTypeFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship_type:create' }
    },
    {
      path: '/relationship-types/:id',
      name: 'RelationshipTypeDetails',
      component: () => import('@/views/relationship-types/RelationshipTypeDetailsView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship_type:read' }
    },
    {
      path: '/relationship-types/:id/edit',
      name: 'EditRelationshipType',
      component: () => import('@/views/relationship-types/RelationshipTypeFormView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'relationship_type:update' }
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
      path: '/users/new',
      name: 'CreateUser',
      component: () => import('@/views/users/UserCreateView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'user:create' }
    },
    {
      path: '/users/:id',
      name: 'UserDetails',
      component: () => import('@/views/users/UserDetailsView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'user:read' }
    },
    {
      path: '/users/:id/edit',
      name: 'EditUser',
      component: () => import('@/views/users/UserEditView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'user:update' }
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/profile/ProfileView.vue'),
      meta: { requiresAuth: true }
    },
    // Amortization Module Routes
    {
      path: '/amortization',
      name: 'AmortizationDashboard',
      component: () => import('@/views/amortization/DashboardView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'amortization:read' }
    },
    {
      path: '/amortization/ledger',
      name: 'AmortizationLedger',
      component: () => import('@/views/amortization/LedgerView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'amortization:read' }
    },
    {
      path: '/amortization/reports',
      name: 'AmortizationReports',
      component: () => import('@/views/amortization/ReportsView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'amortization:read' }
    },
    {
      path: '/amortization/settings',
      name: 'AmortizationSettings',
      component: () => import('@/views/amortization/SettingsView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'amortization:admin' }
    },
    {
      path: '/amortization/scheduler',
      name: 'AmortizationScheduler',
      component: () => import('@/views/amortization/SchedulerView.vue'),
      meta: { requiresAuth: true, requiresPermission: 'amortization:admin' }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/NotFoundView.vue'),
      meta: { requiresAuth: false }
    }
  ]
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Initialize auth store if not already done
  if (!authStore.isInitialized) {
    try {
      await authStore.checkAuth()
    } catch (error) {
      console.error('Auth initialization failed:', error)
    }
  }

  // Wait a bit for user data to be loaded if we have a token but no user yet
  if (authStore.accessToken && !authStore.user && authStore.isLoading) {
    // Wait for auth check to complete
    let attempts = 0
    while (authStore.isLoading && attempts < 50) { // Max 5 seconds
      await new Promise(resolve => setTimeout(resolve, 100))
      attempts++
    }
  }

  // Check if route requires authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  // Check if route requires specific permission
  if (to.meta.requiresPermission && typeof to.meta.requiresPermission === 'string') {
    if (!authStore.hasPermission(to.meta.requiresPermission)) {
      console.log(`Router guard: Missing permission ${to.meta.requiresPermission}, redirecting to dashboard`)
      console.log('User permissions:', authStore.user?.permissions || [])
      // For unauthorized access, you could redirect to a dedicated "unauthorized" page
      // For now, we'll redirect to dashboard
      next('/dashboard')
      return
    }
  }

  // If user is authenticated and tries to access login, redirect to dashboard
  if (to.path === '/login' && authStore.isAuthenticated) {
    next('/dashboard')
    return
  }

  next()
})

export default router
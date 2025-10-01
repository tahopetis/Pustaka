<template>
  <div id="app" class="min-h-screen bg-gray-50">
    <!-- Loading spinner while authenticating -->
    <div v-if="authStore.isLoading" class="fixed inset-0 bg-gray-50 flex items-center justify-center z-50">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p class="text-gray-600">Loading...</p>
      </div>
    </div>

    <!-- Navigation -->
    <AppNavigation v-if="isAuthenticated && !authStore.isLoading" />

    <!-- Main content -->
    <main class="flex" :class="isAuthenticated ? '' : 'h-screen'">
      <!-- Sidebar -->
      <AppSidebar v-if="isAuthenticated && !authStore.isLoading" />

      <!-- Page content -->
      <div :class="isAuthenticated ? 'flex-1' : 'w-full h-full flex items-center justify-center'">
        <router-view />
      </div>
    </main>

    <!-- Toast notifications -->
    <ToastContainer />
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import AppNavigation from '@/components/layout/AppNavigation.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import ToastContainer from '@/components/common/ToastContainer.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)

// Initialize authentication immediately and wait for completion
const initAuth = async () => {
  try {
    console.log('App.vue: Starting auth initialization...')
    await authStore.checkAuth()
    console.log('App.vue: Auth initialization completed')
  } catch (error) {
    console.error('Auth initialization error:', error)
  }
}

onMounted(async () => {
  // Double-check auth state if still loading
  if (authStore.isLoading) {
    const isAuthenticated = await authStore.checkAuth()
    if (!isAuthenticated) {
      console.log('No valid authentication found, user needs to login')
    }
  }
})
</script>
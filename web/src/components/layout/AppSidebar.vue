<template>
  <!-- Sidebar for desktop -->
  <div class="hidden md:flex md:flex-shrink-0">
    <div class="flex flex-col w-64">
      <!-- Sidebar component, swap this element with another sidebar if you like -->
      <div class="flex-1 flex flex-col min-h-0 border-r border-gray-200 bg-gray-50">
        <div class="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
          <nav class="mt-5 flex-1 px-2 bg-white space-y-1">
            <!-- Dashboard -->
            <router-link
              to="/dashboard"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/dashboard')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/dashboard') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
              </svg>
              Dashboard
            </router-link>

            <!-- Configuration Items -->
            <router-link
              v-if="hasAnyPermission(['ci:read'])"
              to="/ci"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/ci')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/ci') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              Configuration Items
            </router-link>

            <!-- CI Types -->
            <router-link
              v-if="hasAnyPermission(['ci_type:read'])"
              to="/ci-types"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/ci-types')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/ci-types') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
              CI Types
            </router-link>

            <!-- Relationships -->
            <router-link
              v-if="hasAnyPermission(['relationship:read'])"
              to="/relationships"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/relationships')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/relationships') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              Relationships
            </router-link>

            <!-- Graph Visualization -->
            <router-link
              v-if="hasAnyPermission(['ci:read'])"
              to="/graph"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/graph')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/graph') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              Graph
            </router-link>

            <!-- Users -->
            <router-link
              v-if="hasAnyPermission(['user:read'])"
              to="/users"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/users')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/users') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              Users
            </router-link>

            <!-- Audit Logs -->
            <router-link
              v-if="hasAnyPermission(['audit:read'])"
              to="/audit"
              class="group flex items-center px-2 py-2 text-sm font-medium rounded-md"
              :class="isCurrentRoute('/audit')
                ? 'bg-blue-100 text-blue-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'"
            >
              <svg
                class="mr-3 h-5 w-5"
                :class="isCurrentRoute('/audit') ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Audit Logs
            </router-link>
          </nav>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const isCurrentRoute = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}

const hasAnyPermission = (permissions: string[]) => {
  return authStore.hasAnyPermission(permissions)
}
</script>
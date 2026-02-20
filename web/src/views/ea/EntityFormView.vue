<template>
  <div class="page-container page-content">
    <!-- Breadcrumbs and Header -->
    <div class="mb-6">
      <nav class="flex mb-4" aria-label="Breadcrumb">
        <ol class="flex items-center space-x-2">
          <li>
            <router-link to="/entities/business" class="text-gray-400 hover:text-gray-500">
              <svg class="flex-shrink-0 h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
              </svg>
            </router-link>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
              <router-link :to="`/entities/${domain}`" class="ml-2 text-sm font-medium text-gray-500 hover:text-gray-700">
                {{ domainDisplay }}
              </router-link>
            </div>
          </li>
          <li v-if="ciType">
            <div class="flex items-center">
              <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
              <span class="ml-2 text-sm font-medium text-gray-500">{{ ciType }}</span>
            </div>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="flex-shrink-0 h-5 w-5 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
              <span class="ml-2 text-sm font-medium text-gray-900">{{ isEdit ? 'Edit' : 'Create' }}</span>
            </div>
          </li>
        </ol>
      </nav>

      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">
            {{ isEdit ? 'Edit Entity' : 'Create Entity' }}
          </h1>
          <p class="page-subtitle">
            {{ isEdit ? 'Update entity attributes and relationships' : `Create a new ${domainDisplay.toLowerCase()} entity` }}
          </p>
        </div>
        <button
          class="btn btn-outline"
          @click="goBack"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
          </svg>
          Back to List
        </button>
      </div>
    </div>

    <!-- Form -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Form -->
      <div class="lg:col-span-2">
        <div class="bg-white shadow rounded-lg p-6">
          <DynamicFormBuilder
            :entity-id="entityId"
            :ci-type="ciType"
            :domain="domain"
          />
        </div>
      </div>

      <!-- Help Sidebar -->
      <div class="lg:col-span-1">
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Help & Information</h3>

          <div v-if="ciTypeDefinition" class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700">CI Type</h4>
              <p class="text-sm text-gray-600 mt-1">{{ ciTypeDefinition.name }}</p>
              <p v-if="ciTypeDefinition.description" class="text-xs text-gray-500 mt-1">
                {{ ciTypeDefinition.description }}
              </p>
            </div>

            <div class="border-t border-gray-200 pt-4">
              <h4 class="text-sm font-medium text-gray-700">Required Fields</h4>
              <ul class="mt-2 space-y-1">
                <li
                  v-for="attr in ciTypeDefinition.required_attributes"
                  :key="attr.name"
                  class="text-xs text-gray-600 flex items-center"
                >
                  <svg class="w-3 h-3 mr-1 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
                  </svg>
                  {{ attr.name }}
                </li>
              </ul>
            </div>

            <div class="border-t border-gray-200 pt-4">
              <h4 class="text-sm font-medium text-gray-700">Tips</h4>
              <ul class="mt-2 space-y-2 text-xs text-gray-600">
                <li class="flex items-start">
                  <svg class="w-4 h-4 mr-1 text-blue-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                  <span>Fill in all required fields marked with <span class="text-red-500">*</span></span>
                </li>
                <li class="flex items-start">
                  <svg class="w-4 h-4 mr-1 text-blue-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                  <span>Use "Save as Draft" to create entities without completing all fields</span>
                </li>
                <li class="flex items-start">
                  <svg class="w-4 h-4 mr-1 text-blue-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                  <span>Data Quality Score will be calculated automatically based on filled fields</span>
                </li>
              </ul>
            </div>
          </div>

          <div v-else class="text-sm text-gray-500">
            Select a CI Type to see help information
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEaTypesStore } from '@/stores/eaTypes'
import DynamicFormBuilder from '@/components/ea/DynamicFormBuilder.vue'

const route = useRoute()
const router = useRouter()
const eaTypesStore = useEaTypesStore()

// Extract route params
const domain = computed(() => route.params.domain as string)
const ciType = computed(() => route.params.ciType as string || '')
const entityId = computed(() => route.params.id as string || '')

const isEdit = computed(() => !!entityId.value)

const domainDisplay = computed(() => {
  const domainMap: Record<string, string> = {
    strategy: 'Strategy',
    business: 'Business',
    application: 'Application',
    data: 'Data',
    technology: 'Technology',
    infrastructure: 'Infrastructure',
    security: 'Security',
    governance: 'Governance'
  }
  return domainMap[domain.value] || 'Entity'
})

const ciTypeDefinition = computed(() => {
  if (!ciType.value) return null
  return eaTypesStore.getCiTypeByName(ciType.value)
})

const goBack = () => {
  router.push(`/entities/${domain.value}`)
}
</script>

<template>
  <DashboardWidget
    title="Risk & Compliance"
    :loading="loading"
    :error="error"
    @retry="fetchRiskMetrics"
  >
    <div v-if="metrics" class="space-y-6">
      <!-- Risk Score Gauge -->
      <div class="flex items-center justify-center">
        <div class="relative">
          <div class="w-40 h-40">
            <svg viewBox="0 0 100 100" class="transform -rotate-90">
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#e5e7eb"
                stroke-width="10"
              />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                :stroke="riskScoreColor"
                stroke-width="10"
                stroke-dasharray="283"
                :stroke-dashoffset="283 - (283 * metrics.risk_score) / 100"
                class="transition-all duration-1000 ease-out"
              />
            </svg>
          </div>
          <div class="absolute inset-0 flex flex-col items-center justify-center">
            <div class="text-4xl font-bold" :class="riskScoreColor">
              {{ metrics.risk_score }}
            </div>
            <div class="text-xs text-gray-500 mt-1">Risk Score</div>
          </div>
        </div>
      </div>

      <!-- Risk Stats Grid -->
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-red-50 border border-red-200 rounded-lg p-4">
          <div class="flex items-center mb-2">
            <svg class="w-5 h-5 text-red-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div class="text-sm font-medium text-red-700">SPOFs</div>
          </div>
          <div class="text-2xl font-bold text-red-900">
            {{ metrics.spof_count }}
          </div>
          <div class="text-xs text-red-600 mt-1">
            Single points of failure
          </div>
        </div>

        <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <div class="flex items-center mb-2">
            <svg class="w-5 h-5 text-yellow-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <div class="text-sm font-medium text-yellow-700">Critical</div>
          </div>
          <div class="text-2xl font-bold text-yellow-900">
            {{ metrics.critical_assets_count }}
          </div>
          <div class="text-xs text-yellow-600 mt-1">
            High-impact assets
          </div>
        </div>

        <div class="bg-orange-50 border border-orange-200 rounded-lg p-4">
          <div class="flex items-center mb-2">
            <svg class="w-5 h-5 text-orange-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div class="text-sm font-medium text-orange-700">No Redundancy</div>
          </div>
          <div class="text-2xl font-bold text-orange-900">
            {{ metrics.no_redundancy_count }}
          </div>
          <div class="text-xs text-orange-600 mt-1">
            Assets without backup
          </div>
        </div>

        <div class="bg-gray-50 border border-gray-200 rounded-lg p-4">
          <div class="flex items-center mb-2">
            <svg class="w-5 h-5 text-gray-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div class="text-sm font-medium text-gray-700">Compliance</div>
          </div>
          <div class="text-2xl font-bold text-gray-900">
            {{ metrics.compliance_violations }}
          </div>
          <div class="text-xs text-gray-600 mt-1">
            Violations found
          </div>
        </div>
      </div>

      <!-- High Risk CIs Table -->
      <div v-if="metrics.high_risk_cis && metrics.high_risk_cis.length > 0">
        <h4 class="text-sm font-medium text-gray-700 mb-3">High-Risk Assets</h4>
        <div class="overflow-hidden border border-gray-200 rounded-lg">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Asset
                </th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Risk Score
                </th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Age
                </th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Tags
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr
                v-for="ci in metrics.high_risk_cis"
                :key="ci.id"
                class="hover:bg-gray-50 cursor-pointer"
                @click="navigateToCI(ci.id)"
              >
                <td class="px-4 py-2 whitespace-nowrap">
                  <div class="flex items-center">
                    <div class="text-sm font-medium text-blue-600">
                      {{ ci.name }}
                    </div>
                    <div v-if="ci.is_critical" class="ml-2">
                      <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                        Critical
                      </span>
                    </div>
                  </div>
                  <div class="text-xs text-gray-500">{{ ci.ci_type }}</div>
                </td>
                <td class="px-4 py-2 whitespace-nowrap">
                  <span
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium"
                    :class="getRiskScoreClass(ci.risk_score)"
                  >
                    {{ ci.risk_score }}/100
                  </span>
                </td>
                <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-900">
                  {{ ci.age_months }} months
                </td>
                <td class="px-4 py-2 whitespace-nowrap">
                  <div v-if="ci.has_redundancy">
                    <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                      Has Backup
                    </span>
                  </div>
                  <div v-else class="text-xs text-gray-400">
                    No redundancy
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- No High Risk Assets Message -->
      <div
        v-else
        class="text-center py-4 bg-green-50 border border-green-200 rounded-lg"
      >
        <svg
          class="w-8 h-8 mx-auto text-green-600 mb-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <p class="text-sm font-medium text-green-900">
          No high-risk assets detected
        </p>
        <p class="text-xs text-green-700 mt-1">
          All assets have acceptable risk levels
        </p>
      </div>
    </div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/services/api'
import DashboardWidget from './DashboardWidget.vue'

interface HighRiskCI {
  id: string
  name: string
  ci_type: string
  risk_score: number
  is_amortizable: boolean
  has_redundancy: boolean
  is_critical: boolean
  age_months: number
  has_tags: boolean
}

interface RiskMetrics {
  risk_score: number
  spof_count: number
  critical_assets_count: number
  no_redundancy_count: number
  compliance_violations: number
  high_risk_cis: HighRiskCI[]
  last_updated: string
}

const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const metrics = ref<RiskMetrics | null>(null)

const riskScoreColor = computed(() => {
  if (!metrics.value) return 'text-gray-500'
  const score = metrics.value.risk_score
  if (score >= 80) return 'text-red-600'
  if (score >= 60) return 'text-orange-600'
  if (score >= 40) return 'text-yellow-600'
  return 'text-green-600'
})

const getRiskScoreClass = (score: number) => {
  if (score >= 80) return 'bg-red-100 text-red-800'
  if (score >= 60) return 'bg-orange-100 text-orange-800'
  if (score >= 40) return 'bg-yellow-100 text-yellow-800'
  return 'bg-green-100 text-green-800'
}

const navigateToCI = (ciId: string) => {
  router.push({ name: 'CIDetails', params: { id: ciId } })
}

const fetchRiskMetrics = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await api.get<RiskMetrics>('/dashboard/risk-metrics', {
      params: {
        limit: 10
      }
    })
    metrics.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchRiskMetrics()
})
</script>

<template>
  <DashboardWidget
    title="Asset Aging"
    :loading="loading"
    :error="error"
    @retry="fetchAssetAging"
  >
    <div v-if="metrics" class="space-y-6">
      <!-- Summary Stats -->
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div class="text-sm text-blue-700 font-medium mb-1">Average Age</div>
          <div class="text-2xl font-bold text-blue-900">
            {{ metrics.average_age_months.toFixed(1) }} months
          </div>
        </div>

        <div
          v-if="metrics.oldest_asset"
          class="bg-purple-50 border border-purple-200 rounded-lg p-4"
        >
          <div class="text-sm text-purple-700 font-medium mb-1">Oldest Asset</div>
          <div class="text-lg font-bold text-purple-900 truncate">
            {{ metrics.oldest_asset.name }}
          </div>
          <div class="text-xs text-purple-600 mt-1">
            {{ metrics.oldest_asset.age_months }} months old
          </div>
        </div>
      </div>

      <!-- Age Distribution Chart -->
      <div v-if="metrics.distribution">
        <h4 class="text-sm font-medium text-gray-700 mb-3">Age Distribution</h4>
        <div class="space-y-2">
          <DistributionBar
            label="Less than 1 year"
            :count="metrics.distribution.less_than_1_year"
            :total="totalAssets"
            color="bg-green-500"
            :icon-class="ageIcons.lessThan1Year"
          />
          <DistributionBar
            label="1-3 years"
            :count="metrics.distribution.one_to_3_years"
            :total="totalAssets"
            color="bg-yellow-500"
            :icon-class="ageIcons.oneTo3Years"
          />
          <DistributionBar
            label="3-5 years"
            :count="metrics.distribution.three_to_5_years"
            :total="totalAssets"
            color="bg-orange-500"
            :icon-class="ageIcons.threeTo5Years"
          />
          <DistributionBar
            label="More than 5 years"
            :count="metrics.distribution.more_than_5_years"
            :total="totalAssets"
            color="bg-red-500"
            :icon-class="ageIcons.moreThan5Years"
          />
        </div>
      </div>

      <!-- Approaching EOL -->
      <div v-if="metrics.approaching_eol && metrics.approaching_eol.length > 0">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-sm font-medium text-gray-700">Approaching End-of-Life</h4>
          <span class="text-xs text-gray-500">
            Next {{ eolThresholdMonths }} months
          </span>
        </div>

        <div class="overflow-hidden border border-gray-200 rounded-lg">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Asset
                </th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  Type
                </th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                  EOL Date
                </th>
                <th class="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase">
                  Days Left
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr
                v-for="asset in metrics.approaching_eol"
                :key="asset.ci.id"
                class="hover:bg-gray-50 cursor-pointer"
                @click="navigateToCI(asset.ci.id)"
              >
                <td class="px-4 py-2 whitespace-nowrap">
                  <div class="text-sm font-medium text-blue-600">
                    {{ asset.ci.name }}
                  </div>
                </td>
                <td class="px-4 py-2 whitespace-nowrap">
                  <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800">
                    {{ asset.type }}
                  </span>
                </td>
                <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-900">
                  {{ formatDate(asset.eol_date) }}
                </td>
                <td class="px-4 py-2 whitespace-nowrap text-right">
                  <span
                    class="inline-flex items-center px-2 py-1 rounded text-xs font-medium"
                    :class="getEOLSeverityClass(asset.days_until_eol)"
                  >
                    {{ asset.days_until_eol }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- No EOL Assets Message -->
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
          No assets approaching end-of-life
        </p>
        <p class="text-xs text-green-700 mt-1">
          All assets are within their useful life period
        </p>
      </div>
    </div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { format } from 'date-fns'
import { api } from '@/services/api'
import DashboardWidget from './DashboardWidget.vue'
import DistributionBar from './DistributionBar.vue'

interface AgeDistribution {
  less_than_1_year: number
  one_to_3_years: number
  three_to_5_years: number
  more_than_5_years: number
}

interface CIReference {
  id: string
  name: string
}

interface ApproachingEOLAsset {
  ci: CIReference
  eol_date: string
  days_until_eol: number
  type: string
}

interface OldestAsset {
  id: string
  name: string
  age_months: number
  created_at: string
}

interface AssetAgingMetrics {
  distribution: AgeDistribution
  approaching_eol: ApproachingEOLAsset[]
  average_age_months: number
  oldest_asset?: OldestAsset
}

const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const metrics = ref<AssetAgingMetrics | null>(null)
const eolThresholdMonths = ref(6)

const totalAssets = computed(() => {
  if (!metrics.value || !metrics.value.distribution) return 0
  const d = metrics.value.distribution
  return d.less_than_1_year + d.one_to_3_years + d.three_to_5_years + d.more_than_5_years
})

const ageIcons = {
  lessThan1Year: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
  oneTo3Years: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
  threeTo5Years: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z',
  moreThan5Years: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z'
}

const formatDate = (dateStr: string) => {
  return format(new Date(dateStr), 'MMM d, yyyy')
}

const getEOLSeverityClass = (days: number) => {
  if (days <= 30) return 'bg-red-100 text-red-800'
  if (days <= 90) return 'bg-orange-100 text-orange-800'
  return 'bg-yellow-100 text-yellow-800'
}

const navigateToCI = (ciId: string) => {
  router.push({ name: 'ci-details', params: { id: ciId } })
}

const fetchAssetAging = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await api.get<AssetAgingMetrics>('/dashboard/asset-aging', {
      params: {
        eol_threshold_months: eolThresholdMonths.value,
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
  fetchAssetAging()
})
</script>

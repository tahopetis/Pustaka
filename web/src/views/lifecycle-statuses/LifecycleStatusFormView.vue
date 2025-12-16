<template>
  <div class="page-container page-content">
    <LifecycleStatusModal
      :lifecycle-status="existingLifecycleStatus"
      :is-open="true"
      @close="handleClose"
      @saved="handleSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLifecycleStatusStore, type LifecycleStatus } from '@/stores/lifecycleStatus'
import LifecycleStatusModal from '@/components/lifecycle-status/LifecycleStatusModal.vue'

const route = useRoute()
const router = useRouter()
const lifecycleStatusStore = useLifecycleStatusStore()

const existingLifecycleStatus = ref<LifecycleStatus | null>(null)
const loading = ref(false)

const isEdit = computed(() => !!route.params.id && route.name === 'EditLifecycleStatus')

const loadLifecycleStatus = async () => {
  if (!route.params.id) return

  loading.value = true
  try {
    const status = await lifecycleStatusStore.getLifecycleStatus(route.params.id as string)
    existingLifecycleStatus.value = status
  } catch (error) {
    console.error('Failed to load lifecycle status:', error)
    router.push('/lifecycle-statuses')
  } finally {
    loading.value = false
  }
}

const handleClose = () => {
  router.push('/lifecycle-status')
}

const handleSaved = (savedLifecycleStatus: LifecycleStatus) => {
  // Show success toast is handled in the modal
  router.push('/lifecycle-status')
}

onMounted(async () => {
  if (isEdit.value) {
    await loadLifecycleStatus()
  }
})
</script>
<template>
  <div class="border border-gray-200 rounded-lg mb-4 overflow-hidden">
    <button
      v-if="!persistent"
      type="button"
      class="w-full px-4 py-3 bg-gray-50 hover:bg-gray-100 flex items-center justify-between transition-colors duration-150"
      @click="toggleCollapse"
    >
      <h3 class="text-lg font-medium text-gray-900">{{ title }}</h3>
      <svg
        class="w-5 h-5 text-gray-500 transform transition-transform duration-200"
        :class="{ 'rotate-180': !collapsed }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
      </svg>
    </button>
    <div v-else class="px-4 py-3 bg-gray-50 border-b border-gray-200">
      <h3 class="text-lg font-medium text-gray-900">{{ title }}</h3>
    </div>

    <div
      v-show="!collapsed || persistent"
      class="p-4 transition-all duration-200"
    >
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  collapsed?: boolean
  persistent?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
  persistent: false
})

const emit = defineEmits<{
  'toggle': [collapsed: boolean]
}>()

const toggleCollapse = () => {
  emit('toggle', !props.collapsed)
}
</script>

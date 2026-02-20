<template>
  <div class="flex flex-wrap items-center gap-2 p-2 border border-gray-300 rounded-md shadow-sm bg-white">
    <span
      v-for="(tag, index) in tags"
      :key="index"
      class="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 text-sm rounded"
    >
      {{ tag }}
      <button
        type="button"
        @click="removeTag(index)"
        class="hover:text-blue-600 focus:outline-none"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
        </svg>
      </button>
    </span>
    <input
      v-model="inputValue"
      type="text"
      class="flex-1 min-w-[120px] outline-none text-sm"
      :placeholder="placeholder"
      @keydown.enter.prevent="addTag"
      @keydown.backspace="handleBackspace"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  modelValue: string[]
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: 'Add tag...'
})

const emit = defineEmits<{
  'update:modelValue': [tags: string[]]
}>()

const tags = ref<string[]>([...props.modelValue])
const inputValue = ref('')

const addTag = () => {
  const tag = inputValue.value.trim()
  if (tag && !tags.value.includes(tag)) {
    tags.value.push(tag)
    emit('update:modelValue', tags.value)
    inputValue.value = ''
  }
}

const removeTag = (index: number) => {
  tags.value.splice(index, 1)
  emit('update:modelValue', tags.value)
}

const handleBackspace = () => {
  if (inputValue.value === '' && tags.value.length > 0) {
    tags.value.pop()
    emit('update:modelValue', tags.value)
  }
}
</script>

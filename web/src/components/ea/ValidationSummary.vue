<template>
  <div v-if="errors.length > 0" class="mb-4 bg-red-50 border-l-4 border-red-500 p-4 rounded">
    <div class="flex">
      <div class="flex-shrink-0">
        <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="ml-3 flex-1">
        <h3 class="text-sm font-medium text-red-800">
          {{ errorCount }} error{{ errorCount > 1 ? 's' : '' }} need fixing
        </h3>
        <div class="mt-2 text-sm text-red-700">
          <ul class="list-disc list-inside space-y-1">
            <li
              v-for="(error, index) in errors"
              :key="index"
              class="cursor-pointer hover:underline"
              @click="scrollToField(error.field)"
            >
              <span class="font-medium">{{ error.field }}:</span> {{ error.message }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ValidationError } from '@/types/ea'

interface Props {
  errors: ValidationError[]
}

const props = defineProps<Props>()

const errorCount = computed(() => props.errors.length)

const scrollToField = (fieldName: string) => {
  const element = document.querySelector(`[data-field="${fieldName}"]`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'center' })

    // Highlight the field briefly
    const htmlElement = element as HTMLElement
    htmlElement.classList.add('ring-2', 'ring-red-500')
    setTimeout(() => {
      htmlElement.classList.remove('ring-2', 'ring-red-500')
    }, 2000)
  }
}
</script>

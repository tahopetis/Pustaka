<template>
  <div class="qr-code-display">
    <!-- QR Code Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal"></div>

        <!-- Modal panel -->
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <!-- Modal header -->
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4 border-b border-gray-200">
            <div class="flex items-center justify-between">
              <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                QR Code - {{ ciName }}
              </h3>
              <button @click="closeModal" class="text-gray-400 hover:text-gray-500">
                <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Modal body -->
          <div class="px-4 pt-5 pb-4 sm:p-6">
            <!-- Loading state -->
            <div v-if="loading" class="text-center py-8">
              <div class="spinner w-12 h-12 mx-auto mb-4"></div>
              <p class="text-gray-500">Loading QR code...</p>
            </div>

            <!-- Error state -->
            <div v-else-if="error" class="text-center py-8">
              <svg class="mx-auto h-12 w-12 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="mt-2 text-red-600">{{ error }}</p>
            </div>

            <!-- QR Code display -->
            <div v-else class="text-center">
              <img
                v-if="qrCodeData"
                :src="qrCodeData"
                :alt="'QR Code for ' + ciName"
                class="mx-auto border-2 border-gray-200 rounded-lg p-2 bg-white"
                style="width: 256px; height: 256px;"
              />

              <!-- Public URL -->
              <div v-if="publicUrl" class="mt-4 p-3 bg-gray-50 rounded-lg">
                <div class="text-sm text-gray-500 mb-1">Public URL:</div>
                <a :href="publicUrl" target="_blank" class="text-sm text-blue-600 hover:text-blue-800 break-all">
                  {{ publicUrl }}
                </a>
              </div>
            </div>
          </div>

          <!-- Modal footer -->
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              @click="downloadPNG"
              :disabled="loading || !qrCodeData"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Download PNG
            </button>
            <button
              @click="copyToClipboard"
              :disabled="loading || !qrCodeData"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Copy to Clipboard
            </button>
            <button
              @click="closeModal"
              type="button"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

const props = defineProps<{
  ciId: string;
  ciName: string;
  modelValue: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const showModal = ref(props.modelValue);
const loading = ref(false);
const qrCodeData = ref<string>('');
const publicUrl = ref<string>('');
const error = ref<string>('');

// Watch for modal open/close
watch(() => props.modelValue, (newValue) => {
  showModal.value = newValue;
  if (newValue) {
    loadQRCode();
  }
});

watch(showModal, (newValue) => {
  emit('update:modelValue', newValue);
});

// Load QR code from backend
const loadQRCode = async () => {
  loading.value = true;
  error.value = '';

  try {
    const token = localStorage.getItem('access_token');
    const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    const size = 256;

    const response = await fetch(
      `${baseURL}/api/v1/qr/ci/${props.ciId}?size=${size}`,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error('Failed to load QR code');
    }

    const data = await response.json();
    qrCodeData.value = data.qr_code;
    publicUrl.value = data.public_url;
  } catch (err) {
    console.error('Error loading QR code:', err);
    error.value = 'Failed to load QR code. Please try again.';
  } finally {
    loading.value = false;
  }
};

// Copy QR code to clipboard
const copyToClipboard = async () => {
  try {
    // Convert base64 to blob
    const base64Data = qrCodeData.value.split(',')[1];
    const byteCharacters = atob(base64Data);
    const byteNumbers = new Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i);
    }
    const byteArray = new Uint8Array(byteNumbers);
    const blob = new Blob([byteArray], { type: 'image/png' });

    // Copy to clipboard
    const item = new ClipboardItem({ 'image/png': blob });
    await navigator.clipboard.write([item]);

    alert('QR Code copied to clipboard!');
  } catch (err) {
    console.error('Error copying to clipboard:', err);
    alert('Failed to copy QR code to clipboard');
  }
};

// Download QR code as PNG
const downloadPNG = () => {
  try {
    const link = document.createElement('a');
    link.href = qrCodeData.value;
    link.download = `ci-${props.ciId}-qr.png`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  } catch (err) {
    console.error('Error downloading QR code:', err);
    alert('Failed to download QR code');
  }
};

const closeModal = () => {
  showModal.value = false;
};
</script>

<style scoped>
.spinner {
  border: 3px solid #f3f3f3;
  border-top: 3px solid #3498db;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>

<template>
  <div class="status-container">
    <h3>System Status</h3>
    <button @click="fetchStatus" :disabled="loading">
      {{ loading ? 'Loading...' : 'Refresh Status' }}
    </button>
    <div v-if="error" class="error">{{ error }}</div>
    <pre v-if="status" class="status-display">{{ status }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const status = ref<string | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

const fetchStatus = async () => {
  loading.value = true;
  error.value = null;
  try {
    const response = await fetch('http://localhost:8080/api/v1/status');
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    status.value = JSON.stringify(data, null, 2);
  } catch (e: any) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchStatus();
});
</script>

<style scoped>
.status-container {
  border: 1px solid #ccc;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 8px;
}
.status-display {
  background: #f4f4f4;
  padding: 0.5rem;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 300px;
}
.error {
  color: red;
}
</style>

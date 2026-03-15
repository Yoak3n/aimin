<template>
  <div class="status-container">
    <div class="header">
      <h3>System Status</h3>
      <div v-if="realtimeState" class="state-badge">{{ realtimeState }}</div>
    </div>
    <button @click="fetchStatus" :disabled="loading">
      {{ loading ? 'Loading...' : 'Refresh Status' }}
    </button>
    <div v-if="error" class="error">{{ error }}</div>
    <pre v-if="status" class="status-display">{{ status }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const props = defineProps<{
  realtimeState?: string | null
}>();

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
  border: 1px solid #4d4d4f;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 8px;
  background-color: #343541;
  color: white;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}
.state-badge {
  background-color: #10a37f;
  color: white;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: bold;
}
.status-display {
  background: #202123;
  padding: 0.5rem;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 300px;
  color: #ececf1;
}
.error {
  color: red;
}
</style>

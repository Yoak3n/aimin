<template>
  <div class="logs-container">
    <h3>Live Logs</h3>
    <div class="logs-window" ref="logsWindow">
      <div v-for="(log, index) in logs" :key="index" class="log-item">
        <span class="log-time">[{{ log.time }}]</span>
        <span class="log-content">{{ log.content }}</span>
      </div>
      <div v-if="logs.length === 0" class="no-logs">Waiting for logs...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';

interface LogMessage {
  time: string;
  content: string;
}

const props = defineProps<{
  logs: LogMessage[]
}>();

const logsWindow = ref<HTMLElement | null>(null);

watch(() => props.logs.length, () => {
  nextTick(() => {
    if (logsWindow.value) {
      logsWindow.value.scrollTop = logsWindow.value.scrollHeight;
    }
  });
});
</script>

<style scoped>
.logs-container {
  border: 1px solid #ccc;
  padding: 1rem;
  border-radius: 8px;
  height: 400px;
  display: flex;
  flex-direction: column;
}
.logs-window {
  flex: 1;
  overflow-y: auto;
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 0.5rem;
  border-radius: 4px;
  font-family: monospace;
}
.log-item {
  margin-bottom: 0.25rem;
  border-bottom: 1px solid #333;
  padding-bottom: 0.25rem;
}
.log-time {
  color: #569cd6;
  margin-right: 0.5rem;
}
.no-logs {
  color: #888;
  font-style: italic;
  text-align: center;
  margin-top: 1rem;
}
</style>

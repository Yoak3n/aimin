<template>
  <div class="thought-section">
    <div class="thought-section__title">Thought</div>
    <div class="thought-section__body">{{ reasoning }}</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  reasoning: string;
}>();
</script>

<style scoped>
.thought-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px dashed #d6d6d6;
  background: #fbfbfb;
  color: #444;
}

.thought-section__title {
  font-size: 11px;
  opacity: 0.75;
  letter-spacing: 0.2px;
  text-transform: uppercase;
}

.thought-section__body {
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.5;
}
</style>

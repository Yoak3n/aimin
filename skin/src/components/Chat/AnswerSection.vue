<template>
  <div class="answer-section">
    <div class="answer-section__title">Answer</div>
    <div class="answer-section__body">
      {{ content }}<span v-if="streaming" class="answer-section__cursor"></span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  content: string;
  streaming?: boolean;
}>();
</script>

<style scoped>
.answer-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #d7e6ff;
  background: #f4f9ff;
  color: #111;
}

.answer-section__title {
  font-size: 11px;
  opacity: 0.75;
  letter-spacing: 0.2px;
  text-transform: uppercase;
}

.answer-section__body {
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.55;
  font-weight: 600;
}

.answer-section__cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  background: #4f8df5;
  margin-left: 2px;
  vertical-align: text-bottom;
  animation: cursor-blink 0.8s step-end infinite;
}

@keyframes cursor-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>

<template>
  <div class="toolcall-section">
    <div class="toolcall-section__title">Tool Calls</div>
    <div class="toolcall-section__body">
      <div v-for="(tc, idx) in toolCalls" :key="idx" class="toolcall-item">
        <div class="toolcall-item__name">{{ tc.name }}</div>
        <div class="toolcall-item__args">{{ tc.arguments }}</div>
        <details v-if="tc.id && getToolResult(tc.id)" class="toolcall-item__result">
          <summary class="toolcall-item__summary">
            {{ getToolResult(tc.id)?.hasError ? "Result (error)" : "Result" }}
          </summary>
          <div class="toolcall-item__result-body">
            <div v-if="getToolResult(tc.id)?.action" class="toolcall-item__result-meta">
              {{ getToolResult(tc.id)?.action }}
            </div>
            <div v-if="getToolResult(tc.id)?.hasError" class="toolcall-item__result-error">
              {{ getToolResult(tc.id)?.error }}
            </div>
            <div class="toolcall-item__result-text">{{ getToolResult(tc.id)?.result }}</div>
          </div>
        </details>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ToolCallItem, ToolResultItem } from "./types";

const props = defineProps<{
  toolCalls: ToolCallItem[];
  toolResults?: ToolResultItem[];
}>();

function getToolResult(toolCallId: string) {
  if (!props.toolResults?.length) return null;
  return props.toolResults.find((x) => x.toolCallId === toolCallId) ?? null;
}
</script>

<style scoped>
.toolcall-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #e8e0f6;
  background: #f9f7fd;
  color: #444;
}

.toolcall-section__title {
  font-size: 11px;
  opacity: 0.75;
  letter-spacing: 0.2px;
  text-transform: uppercase;
}

.toolcall-section__body {
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.5;
}

.toolcall-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid #eee;
  background: #fff;
}

.toolcall-item + .toolcall-item {
  margin-top: 10px;
}

.toolcall-item__name {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
  opacity: 0.9;
}

.toolcall-item__args {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
  opacity: 0.85;
  white-space: pre-wrap;
  word-break: break-word;
}

.toolcall-item__result {
  margin-top: 6px;
}

.toolcall-item__summary {
  cursor: pointer;
  font-size: 12px;
  opacity: 0.9;
}

.toolcall-item__result-body {
  margin-top: 6px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.toolcall-item__result-meta {
  font-size: 12px;
  opacity: 0.9;
}

.toolcall-item__result-error {
  font-size: 12px;
  color: #c41c1c;
  white-space: pre-wrap;
  word-break: break-word;
}

.toolcall-item__result-text {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
  opacity: 0.9;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>

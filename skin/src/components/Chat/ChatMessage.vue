<template>
  <div class="chat-msg" :data-role="message.role">
    <div class="chat-msg__meta">
      <span class="chat-msg__role">{{ message.role }}</span>
      <span class="chat-msg__time">{{ formatTime(message.time) }}</span>
    </div>
    <div class="chat-msg__content">
      <template v-if="isAgentStructured">
        <ThoughtSection v-if="message.reasoning" :reasoning="message.reasoning" />
        <ToolCallSection
          v-if="message.toolCalls?.length"
          :tool-calls="message.toolCalls"
          :tool-results="message.toolResults"
        />
        <AnswerSection v-if="answerContent" :content="answerContent" :streaming="message.streaming" />
        <button
          v-if="message.audio"
          class="chat-msg__audio-btn"
          :class="{
            'chat-msg__audio-btn--playing': playingId === message.id && pausedId !== message.id,
            'chat-msg__audio-btn--paused': pausedId === message.id,
          }"
          type="button"
          @click.stop="$emit('toggleAudio', message)"
        >
          {{ playingId === message.id ? (pausedId === message.id ? "继续播放" : "暂停播放") : "播放语音" }}
        </button>
      </template>
      <template v-else>
        <span>{{ message.content }}</span>
        <button
          v-if="message.audio"
          class="chat-msg__audio-btn"
          :class="{
            'chat-msg__audio-btn--playing': playingId === message.id && pausedId !== message.id,
            'chat-msg__audio-btn--paused': pausedId === message.id,
          }"
          type="button"
          @click.stop="$emit('toggleAudio', message)"
        >
          {{ playingId === message.id ? (pausedId === message.id ? "继续播放" : "暂停播放") : "播放语音" }}
        </button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ChatMessage } from "./types";
import ThoughtSection from "./ThoughtSection.vue";
import ToolCallSection from "./ToolCallSection.vue";
import AnswerSection from "./AnswerSection.vue";

const props = defineProps<{
  message: ChatMessage;
  playingId?: string | null;
  pausedId?: string | null;
}>();

defineEmits<{
  toggleAudio: [message: ChatMessage];
}>();

const isAgentStructured = computed(() => {
  const m = props.message;
  return m.role === "agent" && !!(m.reasoning || m.toolCalls?.length || m.content || m.finalAnswer);
});

const answerContent = computed(() => {
  const m = props.message;
  return m.finalAnswer || m.content || "";
});

function formatTime(ts: number) {
  return new Date(ts).toLocaleTimeString();
}
</script>

<style scoped>
.chat-msg {
  max-width: 85%;
  padding: 10px 12px;
  border-radius: 12px;
  background: #f6f6f6;
  color: #111;
  align-self: flex-start;
  word-break: break-word;
  white-space: pre-wrap;
}

.chat-msg[data-role="user"] {
  background: #e7f1ff;
  align-self: flex-end;
}

.chat-msg[data-role="system"] {
  background: #faf1e5;
  align-self: center;
  max-width: 95%;
}

.chat-msg__meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  opacity: 0.8;
  margin-bottom: 4px;
}

.chat-msg__content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.chat-msg__audio-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
  padding: 6px 14px;
  border: none;
  border-radius: 999px;
  background: #4f8df5;
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.15s;
}

.chat-msg__audio-btn:hover:not(:disabled) {
  background: #3b72d9;
}

.chat-msg__audio-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.chat-msg__audio-btn--playing {
  background: #999;
}

.chat-msg__audio-btn--paused {
  background: #e8a735;
}

.chat-msg__audio-btn--paused:hover:not(:disabled) {
  background: #d0932a;
}
</style>

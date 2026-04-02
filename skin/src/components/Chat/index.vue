<template>
  <div class="chat">
    <div class="chat__header">
      <div class="chat__title">Conversation Agent</div>
      <div class="chat__status" :data-connected="appStore.isConnected">
        {{ appStore.isConnected ? "Connected" : "Disconnected" }}
      </div>
      <div class="chat__client">
        <span class="chat__client-label">Client</span>
        <span class="chat__client-id">{{ appStore.clientId || "-" }}</span>
      </div>
      <button class="chat__btn" type="button" @click="reconnect" :disabled="appStore.isConnected">
        Reconnect
      </button>
    </div>

    <div ref="listEl" class="chat__list">
      <div v-for="m in messages" :key="m.id" class="chat__msg" :data-role="m.role">
        <div class="chat__meta">
          <span class="chat__role">{{ m.role }}</span>
          <span class="chat__time">{{ formatTime(m.time) }}</span>
        </div>
        <div class="chat__content">
          <template v-if="m.role === 'agent' && (m.thought || m.finalAnswer)">
            <div v-if="m.thought" class="chat__section chat__section--thought">
              <div class="chat__section-title">Thought</div>
              <div class="chat__section-body">{{ m.thought }}</div>
            </div>
            <div v-if="m.finalAnswer" class="chat__section chat__section--answer">
              <div class="chat__section-title">Answer</div>
              <div class="chat__section-body">{{ m.finalAnswer }}</div>
            </div>
          </template>
          <template v-else>
            {{ m.content }}
          </template>
        </div>
      </div>
    </div>

    <form class="chat__composer" @submit.prevent="send">
      <input
        v-model="draft"
        class="chat__input"
        type="text"
        placeholder="输入消息，回车发送"
        :disabled="!appStore.isConnected || isSending"
      />
      <button class="chat__btn" type="submit" :disabled="!appStore.isConnected || isSending || !draft.trim()">
        Send
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useAppStore } from "@/store/module/app";
import type { WsIncomingMessage, WsReplyMessage, WsReplyMessageData } from "@/types/ws";

const props = defineProps<{
  managed?: boolean;
}>();

type ChatRole = "user" | "agent" | "system";

interface ChatMessage {
  id: string;
  role: ChatRole;
  content: string;
  time: number;
  taskId?: string;
  streaming?: boolean;
  thought?: string;
  finalAnswer?: string;
}

const appStore = useAppStore();

const listEl = ref<HTMLElement | null>(null);
const draft = ref("");
const isSending = ref(false);
const messages = ref<ChatMessage[]>([]);

const taskToMessageId = new Map<string, string>();
const taskToRaw = new Map<string, string>();

function now() {
  return Date.now();
}

function makeId(prefix: string) {
  return `${prefix}_${Math.random().toString(36).slice(2)}_${now()}`;
}

function pushMessage(partial: Omit<ChatMessage, "id" | "time"> & Partial<Pick<ChatMessage, "id" | "time">>) {
  const message: ChatMessage = {
    id: partial.id ?? makeId(partial.role),
    time: partial.time ?? now(),
    role: partial.role,
    content: partial.content,
    taskId: partial.taskId,
    streaming: partial.streaming,
  };
  messages.value.push(message);
  return message;
}

function findMessageById(id: string) {
  return messages.value.find((m) => m.id === id) ?? null;
}

function upsertAgentStructuredReply(
  taskId: string,
  parsed: { thought?: string; finalAnswer?: string; content: string },
  streaming: boolean
) {
  const existingId = taskToMessageId.get(taskId);
  if (!existingId) {
    const created = pushMessage({
      role: "agent",
      content: parsed.content,
      thought: parsed.thought,
      finalAnswer: parsed.finalAnswer,
      taskId,
      streaming,
    });
    taskToMessageId.set(taskId, created.id);
    return;
  }

  const target = findMessageById(existingId);
  if (!target) {
    taskToMessageId.delete(taskId);
    const created = pushMessage({
      role: "agent",
      content: parsed.content,
      thought: parsed.thought,
      finalAnswer: parsed.finalAnswer,
      taskId,
      streaming,
    });
    taskToMessageId.set(taskId, created.id);
    return;
  }

  target.content = parsed.content;
  if (typeof parsed.thought === "string") target.thought = parsed.thought;
  if (typeof parsed.finalAnswer === "string") target.finalAnswer = parsed.finalAnswer;
  target.streaming = streaming;
}

function extractXmlTagContent(raw: string, tag: string) {
  const open = `<${tag}>`;
  const close = `</${tag}>`;
  const start = raw.indexOf(open);
  if (start === -1) return null;
  const contentStart = start + open.length;
  const end = raw.indexOf(close, contentStart);
  if (end === -1) return raw.slice(contentStart);
  return raw.slice(contentStart, end);
}

function parseAgentXmlOrPlain(raw: string) {
  const thought = extractXmlTagContent(raw, "thought");
  const finalAnswer = extractXmlTagContent(raw, "final_answer");
  if (thought !== null || finalAnswer !== null) {
    const content = (finalAnswer ?? thought ?? "").trim();
    return {
      thought: thought?.trim() || undefined,
      finalAnswer: finalAnswer?.trim() || undefined,
      content,
    };
  }
  return { content: raw };
}

function handleReply(data: WsReplyMessageData) {
  if (data.status === 0 && data.chunk) {
    const prev = taskToRaw.get(data.task_id) ?? "";
    const nextRaw = prev + data.chunk.content;
    taskToRaw.set(data.task_id, nextRaw);
    const parsed = parseAgentXmlOrPlain(nextRaw);
    upsertAgentStructuredReply(data.task_id, parsed, true);
    return;
  }

  if (data.status === 1 && data.result) {
    const prevRaw = taskToRaw.get(data.task_id) ?? "";
    const prevParsed = parseAgentXmlOrPlain(prevRaw);
    const nextParsed = parseAgentXmlOrPlain(data.result.content);
    const combined = {
      content: (nextParsed.finalAnswer ?? nextParsed.thought ?? nextParsed.content ?? "").trim(),
      thought: nextParsed.thought ?? prevParsed.thought,
      finalAnswer: nextParsed.finalAnswer ?? prevParsed.finalAnswer,
    };
    taskToRaw.set(data.task_id, prevRaw ? `${prevRaw}${data.result.content}` : data.result.content);
    upsertAgentStructuredReply(data.task_id, combined, false);
  }
}

function handleIncoming(message: WsIncomingMessage) {
  if (message.action === "Connected") {
    return;
  }

  if (message.action === "Close") {
    return;
  }

  if (message.action === "State") {
    return;
  }

  if (message.action === "Log") {
    return;
  }

  if (message.action === "Ask") {
    return;
  }

  if (message.action === "Reply") {
    handleReply((message as WsReplyMessage).data);
  }
}

function receiveWsMessage(message: WsIncomingMessage) {
  handleIncoming(message);
}

defineExpose({
  receiveWsMessage,
});

function reconnect() {
  appStore.initWebSocket(appStore.clientId || undefined);
}

async function send() {
  const text = draft.value.trim();
  if (!text) return;
  if (!appStore.isConnected) return;

  isSending.value = true;
  try {
    pushMessage({ role: "user", content: text });
    draft.value = "";

    appStore.sendTask(text, 0);
  } finally {
    isSending.value = false;
  }
}

function formatTime(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

let unsubscribe: null | (() => void) = null;

onMounted(() => {
  if (props.managed) return;
  if (!appStore.socket) {
    appStore.initWebSocket();
  }

  unsubscribe = appStore.onIncomingMessage((message) => {
    handleIncoming(message);
  });
});

onBeforeUnmount(() => {
  unsubscribe?.();
  unsubscribe = null;
});

watch(
  () => messages.value.length,
  async () => {
    await nextTick();
    const el = listEl.value;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }
);

</script>


<style scoped>
.chat {
  display: flex;
  flex-direction: column;
  height: 100vh;
  max-width: 900px;
  margin: 0 auto;
  padding: 16px;
  box-sizing: border-box;
  gap: 12px;
}

.chat__header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.chat__title {
  font-size: 18px;
  font-weight: 600;
}

.chat__status {
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 999px;
  background: #f2f2f2;
  color: #333;
}

.chat__status[data-connected="true"] {
  background: #e8f7ee;
  color: #1c7a3d;
}

.chat__client {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 999px;
  background: #f2f2f2;
  color: #333;
  max-width: 320px;
}

.chat__client-label {
  opacity: 0.75;
}

.chat__client-id {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat__list {
  flex: 1;
  overflow: auto;
  padding: 12px;
  border: 1px solid #e5e5e5;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: #fff;
}

.chat__msg {
  max-width: 85%;
  padding: 10px 12px;
  border-radius: 12px;
  background: #f6f6f6;
  color: #111;
  align-self: flex-start;
  word-break: break-word;
  white-space: pre-wrap;
}

.chat__msg[data-role="user"] {
  background: #e7f1ff;
  align-self: flex-end;
}

.chat__msg[data-role="system"] {
  background: #faf1e5;
  align-self: center;
  max-width: 95%;
}

.chat__meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  opacity: 0.8;
  margin-bottom: 4px;
}

.chat__composer {
  display: flex;
  gap: 10px;
}

.chat__input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #e5e5e5;
  border-radius: 10px;
  outline: none;
}

.chat__input:disabled {
  background: #f3f3f3;
}

.chat__btn {
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid #e5e5e5;
  background: #fff;
  cursor: pointer;
}

.chat__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.chat__section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.chat__section + .chat__section {
  margin-top: 10px;
}

.chat__section-title {
  font-size: 11px;
  opacity: 0.75;
  letter-spacing: 0.2px;
  text-transform: uppercase;
}

.chat__section-body {
  white-space: pre-wrap;
}

.chat__section--thought {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px dashed #d6d6d6;
  background: #fbfbfb;
  color: #444;
}

.chat__section--thought .chat__section-body {
  font-size: 13px;
  line-height: 1.5;
}

.chat__section--answer {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #d7e6ff;
  background: #f4f9ff;
  color: #111;
}

.chat__section--answer .chat__section-body {
  font-size: 14px;
  line-height: 1.55;
  font-weight: 600;
}
</style>

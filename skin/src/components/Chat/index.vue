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
      <ChatMessage
        v-for="m in messages"
        :key="m.id"
        :message="m"
        :playing-id="playingId"
        :paused-id="pausedId"
        @toggle-audio="onToggleAudio"
      />
    </div>

    <form class="chat__composer" @submit.prevent="send">
      <input v-model="draft" class="chat__input" type="text" placeholder="输入消息，回车发送"
        :disabled="!appStore.isConnected || isSending" />
      <button class="chat__btn" type="button" @click="interruptCurrentRound" :disabled="!appStore.isConnected">
        Interrupt
      </button>
      <button class="chat__btn" type="submit" :disabled="!appStore.isConnected || isSending || !draft.trim()">
        Send
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useAppStore } from "@/store/module/app";
import type { WsIncomingMessage, WsReplyMessage, WsReplyMessageData, WsToolResultMessage, WsToolResultMessageData, WsAudioMessage, WsAudioMessageData } from "@/types/ws";
import type { ChatMessage as ChatMessageType, ToolCallItem, ToolResultItem } from "./types";
import ChatMessage from "./ChatMessage.vue";
import { scanJsonObject } from "./utils";

const props = defineProps<{
  managed?: boolean;
}>();

const appStore = useAppStore();

const listEl = ref<HTMLElement | null>(null);
const draft = ref("");
const isSending = ref(false);
const messages = ref<ChatMessageType[]>([]);
const playingId = ref<string | null>(null);
const playingAudio = ref<HTMLAudioElement | null>(null);
const playingSource = ref<AudioBufferSourceNode | null>(null);
const playingCtx = ref<AudioContext | null>(null);

const taskToMessageId = new Map<string, string>();
const taskToRaw = new Map<string, string>();
const taskToReasoningRaw = new Map<string, string>();
const taskToToolResults = new Map<string, ToolResultItem[]>();

function now() {
  return Date.now();
}

function makeId(prefix: string) {
  return `${prefix}_${Math.random().toString(36).slice(2)}_${now()}`;
}

function pushMessage(partial: Omit<ChatMessageType, "id" | "time"> & Partial<Pick<ChatMessageType, "id" | "time">>) {
  const message: ChatMessageType = {
    id: partial.id ?? makeId(partial.role),
    time: partial.time ?? now(),
    role: partial.role,
    content: partial.content,
    taskId: partial.taskId,
    streaming: partial.streaming,
    reasoning: partial.reasoning,
    finalAnswer: partial.finalAnswer,
    toolCalls: partial.toolCalls,
    toolResults: partial.toolResults,
    audio: partial.audio,
  };
  messages.value.push(message);
  return message;
}

function findMessageById(id: string) {
  return messages.value.find((m) => m.id === id) ?? null;
}

function upsertAgentStructuredReply(
  taskId: string,
  parsed: { toolCalls: ToolCallItem[]; reasoning?: string; finalAnswer?: string; content: string },
  streaming: boolean
) {
  const existingId = taskToMessageId.get(taskId);
  if (!existingId) {
    const created = pushMessage({
      role: "agent",
      content: parsed.content,
      reasoning: parsed.reasoning,
      finalAnswer: parsed.finalAnswer,
      toolCalls: parsed.toolCalls,
      toolResults: taskToToolResults.get(taskId) ?? [],
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
      reasoning: parsed.reasoning,
      finalAnswer: parsed.finalAnswer,
      toolCalls: parsed.toolCalls,
      toolResults: taskToToolResults.get(taskId) ?? [],
      taskId,
      streaming,
    });
    taskToMessageId.set(taskId, created.id);
    return;
  }

  target.content = parsed.content;
  if (typeof parsed.reasoning === "string") target.reasoning = parsed.reasoning;
  if (typeof parsed.finalAnswer === "string") target.finalAnswer = parsed.finalAnswer;
  if (Array.isArray(parsed.toolCalls)) target.toolCalls = parsed.toolCalls;
  target.toolResults = taskToToolResults.get(taskId) ?? target.toolResults;
  target.streaming = streaming;
}

function parseAgentTextWithToolCalls(raw: string) {
  const marker = "[tool_call]";
  const toolCalls: ToolCallItem[] = [];
  let out = "";
  let i = 0;

  while (i < raw.length) {
    const idx = raw.indexOf(marker, i);
    if (idx === -1) {
      out += raw.slice(i);
      break;
    }
    out += raw.slice(i, idx);
    let p = idx + marker.length;
    while (p < raw.length && /\s/.test(raw[p])) p++;
    if (p >= raw.length) {
      out += raw.slice(idx);
      break;
    }
    let nameEnd = p;
    while (nameEnd < raw.length && !/\s/.test(raw[nameEnd])) nameEnd++;
    const firstToken = raw.slice(p, nameEnd).trim();
    if (!firstToken) {
      out += raw.slice(idx);
      break;
    }
    p = nameEnd;
    while (p < raw.length && /\s/.test(raw[p])) p++;

    let toolCallId: string | undefined;
    let name = "";
    if (p < raw.length && raw[p] === "{") {
      name = firstToken;
    } else {
      let secondEnd = p;
      while (secondEnd < raw.length && !/\s/.test(raw[secondEnd])) secondEnd++;
      const secondToken = raw.slice(p, secondEnd).trim();
      if (!secondToken) {
        out += raw.slice(idx);
        break;
      }
      toolCallId = firstToken;
      name = secondToken;
      p = secondEnd;
      while (p < raw.length && /\s/.test(raw[p])) p++;
    }

    if (p >= raw.length || raw[p] !== "{") {
      out += raw.slice(idx);
      break;
    }
    const parsedJson = scanJsonObject(raw, p);
    if (!parsedJson) {
      out += raw.slice(idx);
      break;
    }
    if (name !== "final_answer") {
      toolCalls.push({ id: toolCallId, name, arguments: parsedJson.json });
    }
    i = parsedJson.end;
  }

  return {
    content: out.trim(),
    toolCalls,
  };
}

function parseAgentRaw(raw: string) {
  const parsed = parseAgentTextWithToolCalls(raw);
  return {
    content: parsed.content,
    toolCalls: parsed.toolCalls,
  };
}

function handleReply(data: WsReplyMessageData) {
  if (data.status === 0 && data.chunk) {
    const prev = taskToRaw.get(data.task_id) ?? "";
    const cot = prev + data.chunk.content;
    taskToRaw.set(data.task_id, cot);
    const prevReasoning = taskToReasoningRaw.get(data.task_id) ?? "";
    const nextReasoning = prevReasoning + String(data.chunk.reasoning_content ?? "");
    taskToReasoningRaw.set(data.task_id, nextReasoning);
    const parsed = parseAgentRaw(cot);
    upsertAgentStructuredReply(
      data.task_id,
      {
        content: parsed.content,
        toolCalls: parsed.toolCalls,
        reasoning: nextReasoning.trim() || undefined,
      },
      true
    );
    return;
  }

  if (data.status === 1 && data.result) {
    const prevRaw = taskToRaw.get(data.task_id) ?? "";
    const parsedPrev = parseAgentRaw(prevRaw);
    taskToRaw.set(data.task_id, prevRaw ? `${prevRaw}${data.result.content}` : data.result.content);
    const reasoning = (taskToReasoningRaw.get(data.task_id) ?? "").trim() || undefined;
    const finalText = String(data.result.content ?? "").trim();
    const prevText = String(parsedPrev.content ?? "").trim();

    upsertAgentStructuredReply(
      data.task_id,
      {
        content: prevText || finalText,
        toolCalls: parsedPrev.toolCalls,
        reasoning,
        finalAnswer: finalText && finalText !== prevText ? finalText : undefined,
      },
      false
    );
  }
}

function normalizeToolResultData(data: unknown): WsToolResultMessageData | null {
  if (!data || typeof data !== "object") return null;
  const maybe = data as Record<string, unknown>;
  const taskId = typeof maybe.task_id === "string" ? maybe.task_id.trim() : "";
  const toolCallId = typeof maybe.tool_call_id === "string" ? maybe.tool_call_id.trim() : "";
  const action = typeof maybe.action === "string" ? maybe.action : "";
  const result = typeof maybe.result === "string" ? maybe.result : "";
  const err = typeof maybe.error === "string" ? maybe.error : undefined;
  const hasError = typeof maybe.has_error === "boolean" ? maybe.has_error : Boolean(err);
  if (!taskId || !toolCallId) return null;
  return { task_id: taskId, tool_call_id: toolCallId, action, result, error: err, has_error: hasError };
}

function handleToolResultMessage(message: WsIncomingMessage) {
  if (message.action !== "ToolResult") return false;
  const data = normalizeToolResultData((message as WsToolResultMessage).data);
  if (!data) return true;

  const nextItem: ToolResultItem = {
    toolCallId: data.tool_call_id,
    action: data.action,
    result: data.result,
    error: data.error,
    hasError: data.has_error,
  };
  const prev = taskToToolResults.get(data.task_id) ?? [];
  const idx = prev.findIndex((x) => x.toolCallId === nextItem.toolCallId);
  const next = idx >= 0 ? prev.map((x, i) => (i === idx ? nextItem : x)) : [...prev, nextItem];
  taskToToolResults.set(data.task_id, next);

  const msgId = taskToMessageId.get(data.task_id);
  if (msgId) {
    const target = findMessageById(msgId);
    if (target) {
      target.toolResults = next;
    }
  }
  return true;
}

function base64ToArrayBuffer(base64: string): ArrayBuffer {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

function stopCurrentAudio() {
  if (playingAudio.value) {
    playingAudio.value.pause();
    playingAudio.value = null;
  }
  if (playingSource.value) {
    try { playingSource.value.stop(); } catch {}
    playingSource.value = null;
  }
  if (playingCtx.value) {
    playingCtx.value.close();
    playingCtx.value = null;
  }
  playingId.value = null;
}

function playAudioBase64(audioBase64: string, format: string, msgId?: string) {
  try {
    stopCurrentAudio();

    if (format === "pcm16") {
      const pcmBuffer = base64ToArrayBuffer(audioBase64);
      const pcmData = new Int16Array(pcmBuffer);
      const sampleRate = 24000;
      const float32 = new Float32Array(pcmData.length);
      for (let i = 0; i < pcmData.length; i++) {
        float32[i] = pcmData[i] / 32768.0;
      }
      const ctx = new AudioContext({ sampleRate });
      const audioBuffer = ctx.createBuffer(1, float32.length, sampleRate);
      audioBuffer.getChannelData(0).set(float32);
      const source = ctx.createBufferSource();
      source.buffer = audioBuffer;
      source.connect(ctx.destination);
      source.start();

      playingCtx.value = ctx;
      playingSource.value = source;
      if (msgId) playingId.value = msgId;

      source.onended = () => {
        ctx.close();
        if (playingSource.value === source) {
          playingSource.value = null;
          playingCtx.value = null;
        }
        if (playingId.value === msgId) playingId.value = null;
      };
    } else {
      const audioBuffer = base64ToArrayBuffer(audioBase64);
      const blob = new Blob([audioBuffer], { type: `audio/${format}` });
      const url = URL.createObjectURL(blob);
      const audio = new Audio(url);

      playingAudio.value = audio;
      if (msgId) playingId.value = msgId;

      audio.onended = () => {
        URL.revokeObjectURL(url);
        if (playingAudio.value === audio) playingAudio.value = null;
        if (playingId.value === msgId) playingId.value = null;
      };
      audio.onerror = () => {
        URL.revokeObjectURL(url);
        if (playingAudio.value === audio) playingAudio.value = null;
        if (playingId.value === msgId) playingId.value = null;
      };
      audio.play();
    }
  } catch (e) {
    console.error("Audio playback error:", e);
    stopCurrentAudio();
  }
}

function pauseCurrentAudio() {
  if (playingAudio.value) {
    playingAudio.value.pause();
  } else if (playingCtx.value && playingCtx.value.state === "running") {
    playingCtx.value.suspend();
  }
}

function resumeCurrentAudio() {
  if (playingAudio.value) {
    playingAudio.value.play();
  } else if (playingCtx.value && playingCtx.value.state === "suspended") {
    playingCtx.value.resume();
  }
}

const pausedId = ref<string | null>(null);

function onToggleAudio(m: ChatMessageType) {
  if (!m.audio) return;

  if (playingId.value === m.id) {
    if (pausedId.value === m.id) {
      resumeCurrentAudio();
      pausedId.value = null;
    } else {
      pauseCurrentAudio();
      pausedId.value = m.id;
    }
    return;
  }

  pausedId.value = null;
  playAudioBase64(m.audio.base64, m.audio.format, m.id);
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

  if (message.action === "ToolResult") {
    handleToolResultMessage(message);
    return;
  }

  if (message.action === "Audio") {
    const data = (message as WsAudioMessage).data as WsAudioMessageData | undefined;
    if (data?.audio_base64) {
      pushMessage({
        role: "agent",
        content: `语音回复 · ${data.voice ?? ""} · ${(data.bytes / 1024).toFixed(1)} KB`,
        taskId: data.task_id,
        audio: {
          base64: data.audio_base64,
          format: data.format || "wav",
          voice: data.voice ?? "",
          bytes: data.bytes ?? 0,
        },
      });
    }
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

function interruptCurrentRound() {
  if (!appStore.isConnected) return;
  appStore.sendInterrupt("Interrupt");
  pushMessage({ role: "system", content: "已请求打断当前轮次" });
}

async function send() {
  const text = draft.value.trim();
  if (!text) return;
  if (!appStore.isConnected) return;

  isSending.value = true;
  try {
    if (messages.value.some((m) => m.role === "agent" && m.streaming)) {
      appStore.sendInterrupt("Interrupt");
    }
    pushMessage({ role: "user", content: text });
    draft.value = "";

    appStore.sendTask(text, 0);
  } finally {
    isSending.value = false;
  }
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
  stopCurrentAudio();
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
</style>

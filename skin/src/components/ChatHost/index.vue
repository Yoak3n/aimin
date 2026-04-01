<template>
  <div class="chat-host">
    <div class="chat-host__main">
      <Chat ref="chatRef" :managed="true" />
    </div>

    <aside v-if="hasSystemPanel" class="chat-host__system">
      <div class="chat-host__system-header">
        <div class="chat-host__system-title">System</div>
        <button class="chat-host__system-btn" type="button" @click="clearSystemNotices">Clear</button>
      </div>

      <div v-if="latestState" class="chat-host__section chat-host__state">
        <div class="chat-host__section-header">
          <div class="chat-host__section-title">State</div>
          <div class="chat-host__section-meta">{{ formatTime(latestState.time) }}</div>
        </div>
        <div class="chat-host__state-body" :title="latestState.content">{{ latestState.content }}</div>
      </div>

      <div v-if="logCount" class="chat-host__section chat-host__logs">
        <div class="chat-host__section-header">
          <div class="chat-host__section-title">Log</div>
          <div class="chat-host__section-meta">{{ logCount }}</div>
          <button class="chat-host__system-btn" type="button" @click="toggleLogs">
            {{ isLogExpanded ? "Collapse" : "Expand" }}
          </button>
        </div>
        <div class="chat-host__logs-body">
          <template v-if="isLogExpanded">
            <div class="chat-host__system-list">
              <div
                v-for="l in visibleLogs"
                :key="l.id"
                class="chat-host__system-item"
                data-action="Log"
                :title="l.content"
              >
                <span class="chat-host__system-time">{{ formatTime(l.time) }}</span>
                <span class="chat-host__system-action">Log</span>
                <span class="chat-host__system-content">{{ l.content }}</span>
              </div>
            </div>
          </template>
          <template v-else>
            <div v-if="latestLog" class="chat-host__logs-collapsed" :title="latestLog.content">
              <span class="chat-host__system-time">{{ formatTime(latestLog.time) }}</span>
              <span class="chat-host__system-content">{{ latestLog.content }}</span>
            </div>
          </template>
        </div>
      </div>

      <div v-if="visibleSystemNotices.length" class="chat-host__system-list">
        <div
          v-for="n in visibleSystemNotices"
          :key="n.id"
          class="chat-host__system-item"
          :data-action="n.action"
          :title="n.content"
        >
          <span class="chat-host__system-time">{{ formatTime(n.time) }}</span>
          <span class="chat-host__system-action">{{ n.action }}</span>
          <span class="chat-host__system-content">{{ n.content }}</span>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import Chat from "@/components/Chat/index.vue";
import { useAppStore } from "@/store/module/app";
import type { WsIncomingMessage, WsLogMessageData } from "@/types/ws";

type ChatExposed = {
  receiveWsMessage: (message: WsIncomingMessage) => void;
};

type SystemNotice = {
  id: string;
  time: number;
  action: string;
  content: string;
};

type StateSnapshot = {
  time: number;
  content: string;
};

const appStore = useAppStore();
const chatRef = ref<ChatExposed | null>(null);

const systemNotices = ref<SystemNotice[]>([]);
const latestState = ref<StateSnapshot | null>(null);

const logItems = ref<SystemNotice[]>([]);
const isLogExpanded = ref(false);

function now() {
  return Date.now();
}

function makeId(prefix: string) {
  return `${prefix}_${Math.random().toString(36).slice(2)}_${now()}`;
}

function formatTime(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

function pushSystemNotice(action: string, content: string, time: number = now()) {
  systemNotices.value.push({
    id: makeId("sys"),
    time,
    action,
    content,
  });
  if (systemNotices.value.length > 200) {
    systemNotices.value.splice(0, systemNotices.value.length - 200);
  }
}

function clearSystemNotices() {
  systemNotices.value = [];
  logItems.value = [];
  latestState.value = null;
}

function normalizeLogData(data: WsLogMessageData | unknown) {
  if (!data || typeof data !== "object") return null;
  const maybe = data as { time?: unknown; content?: unknown };
  if (typeof maybe.content !== "string") return null;
  const time = typeof maybe.time === "string" ? Date.parse(maybe.time) : NaN;
  return {
    time: Number.isFinite(time) ? time : now(),
    content: maybe.content,
  };
}

function handleSystemMessage(message: WsIncomingMessage) {
  if (message.action === "Connected") {
    pushSystemNotice("Connected", `clientId=${appStore.clientId || "-"}`);
    return true;
  }

  if (message.action === "Close") {
    pushSystemNotice("Close", String(message.data ?? "Close"));
    return true;
  }

  if (message.action === "State") {
    latestState.value = { time: now(), content: String(message.data ?? "") };
    return true;
  }

  if (message.action === "Log") {
    const normalized = normalizeLogData(message.data);
    if (normalized) {
      logItems.value.push({
        id: makeId("log"),
        time: normalized.time,
        action: "Log",
        content: normalized.content,
      });
    } else {
      logItems.value.push({
        id: makeId("log"),
        time: now(),
        action: "Log",
        content: String(message.data ?? ""),
      });
    }
    if (logItems.value.length > 500) {
      logItems.value.splice(0, logItems.value.length - 500);
    }
    return true;
  }

  return false;
}

function dispatchToChat(message: WsIncomingMessage) {
  chatRef.value?.receiveWsMessage(message);
}

let unsubscribe: null | (() => void) = null;

onMounted(() => {
  if (!appStore.socket) {
    appStore.initWebSocket();
  }

  unsubscribe = appStore.onIncomingMessage((message) => {
    if (handleSystemMessage(message)) return;
    dispatchToChat(message);
  });
});

onBeforeUnmount(() => {
  unsubscribe?.();
  unsubscribe = null;
});

const visibleSystemNotices = computed(() => {
  const list = systemNotices.value;
  if (list.length <= 20) return list;
  return list.slice(list.length - 20);
});

const logCount = computed(() => logItems.value.length);

const latestLog = computed(() => {
  const list = logItems.value;
  return list.length ? list[list.length - 1] : null;
});

const visibleLogs = computed(() => {
  const list = logItems.value;
  if (list.length <= 50) return list;
  return list.slice(list.length - 50);
});

const hasSystemPanel = computed(() => {
  return Boolean(latestState.value) || logCount.value > 0 || visibleSystemNotices.value.length > 0;
});

function toggleLogs() {
  isLogExpanded.value = !isLogExpanded.value;
}
</script>

<style scoped>
.chat-host {
  display: flex;
  align-items: stretch;
  gap: 16px;
  min-height: 100vh;
  padding: 0 16px;
  box-sizing: border-box;
}

.chat-host__main {
  flex: 1 1 auto;
  min-width: 0;
}

.chat-host__system {
  flex: 0 0 420px;
  width: 420px;
  max-width: 420px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  border-radius: 10px;
  border: 1px solid #e5e5e5;
  background: #fff;
  overflow: auto;
  box-sizing: border-box;
}

.chat-host__system-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.chat-host__system-title {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.2px;
  text-transform: uppercase;
  opacity: 0.75;
}

.chat-host__system-btn {
  padding: 6px 10px;
  border-radius: 10px;
  border: 1px solid #e5e5e5;
  background: #fff;
  cursor: pointer;
  font-size: 12px;
}

.chat-host__system-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.chat-host__section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 10px;
  border-radius: 10px;
  border: 1px solid #e5e5e5;
  background: #fafafa;
}

.chat-host__section-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.chat-host__section-title {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.2px;
  text-transform: uppercase;
  opacity: 0.75;
}

.chat-host__section-meta {
  font-size: 11px;
  opacity: 0.6;
}

.chat-host__state-body {
  font-size: 12px;
  opacity: 0.9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-host__logs-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chat-host__logs-collapsed {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px;
  align-items: baseline;
  padding: 6px 8px;
  border-radius: 10px;
  background: #fafafa;
  border: 1px solid #ededed;
}

.chat-host__system-item {
  display: grid;
  grid-template-columns: auto auto 1fr;
  gap: 8px;
  align-items: baseline;
  padding: 6px 8px;
  border-radius: 10px;
  background: #fafafa;
  border: 1px solid #ededed;
}

.chat-host__system-time {
  font-size: 11px;
  opacity: 0.7;
}

.chat-host__system-action {
  font-size: 11px;
  font-weight: 700;
  opacity: 0.85;
}

.chat-host__system-content {
  font-size: 12px;
  opacity: 0.9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1100px) {
  .chat-host {
    flex-direction: column;
    padding: 0;
    gap: 0;
  }

  .chat-host__system {
    width: auto;
    max-width: none;
    height: auto;
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
}
</style>

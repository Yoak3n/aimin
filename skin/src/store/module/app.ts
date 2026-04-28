import { defineStore } from "pinia";
import type { WsAnswerMessageData, WsIncomingMessage, WsOutgoingMessage, WsTaskType } from "@/types/ws";

export const useAppStore = defineStore("app", {
  state: () => ({
    isMobile: false,
    socket: null as WebSocket | null,
    isConnected: false,
    clientId: "",
    lastIncomingMessage: null as WsIncomingMessage | null,
    lastIncomingRaw: "",
  }),
  actions: {
    initWebSocket(id?: string) {
      if (this.socket) {
        this.socket.close();
      }

      const clientId = id || `client_${Math.random().toString(36).substring(2, 9)}`;
      this.clientId = clientId;
      const baseUrl = import.meta.env.VITE_WS_BASE_URL;
      const wsUrl = `${baseUrl}/ws/${clientId}`;

      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        this.isConnected = true;
      };

      this.socket.onmessage = (event) => {
        const raw = typeof event.data === "string" ? event.data : String(event.data ?? "");
        const message = parseIncomingMessage(raw);
        if (!message) return;

        this.lastIncomingMessage = message;
        this.lastIncomingRaw = raw;

        if (message.action === "Ping") {
          this.sendMessage({ action: "Pong", data: "Pong" });
        }

        notifyIncoming(message, raw);
      };

      this.socket.onclose = () => {
        this.isConnected = false;
        this.socket = null;
      };

      this.socket.onerror = (error) => {
        console.error("WebSocket error:", error);
        this.isConnected = false;
      };
    },
    onIncomingMessage(listener: (message: WsIncomingMessage, raw: string) => void) {
      incomingListeners.add(listener);
      return () => {
        incomingListeners.delete(listener);
      };
    },
    sendMessage(message: WsOutgoingMessage) {
      if (this.socket && this.isConnected) {
        this.socket.send(JSON.stringify(message));
      } else {
        console.error("WebSocket is not connected");
      }
    },
    sendPing(payload: string = "Ping") {
      this.sendMessage({ action: "Ping", data: payload });
    },
    sendInterrupt(payload: string = "Interrupt") {
      this.sendMessage({ action: "Interrupt", data: payload });
    },
    sendAnswer(payload: WsAnswerMessageData) {
      this.sendMessage({ action: "Answer", data: payload });
    },
    skipAnswer(id: string) {
      this.sendAnswer({ id, skip: true });
    },
    sendTask(payload: string, taskType: WsTaskType = 0) {
      this.sendMessage({ action: "Task", data: { type: taskType, payload, from: this.clientId } });
    },
  },
});

const incomingListeners = new Set<(message: WsIncomingMessage, raw: string) => void>();

function notifyIncoming(message: WsIncomingMessage, raw: string) {
  incomingListeners.forEach((listener) => {
    listener(message, raw);
  });
}

function parseIncomingMessage(raw: string): WsIncomingMessage | null {
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== "object") return null;
    const maybeMessage = parsed as { action?: unknown; data?: unknown };
    if (typeof maybeMessage.action !== "string") return null;
    return maybeMessage as WsIncomingMessage;
  } catch {
    return null;
  }
}

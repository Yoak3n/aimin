import { defineStore } from "pinia";

export const useAppStore = defineStore("app", {
  state: () => ({
    isMobile: false,
    socket: null as WebSocket | null,
    isConnected: false,
    clientId: "",
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
        console.log("WebSocket connected");
        this.isConnected = true;
      };

      this.socket.onmessage = (event) => {
        console.log("WebSocket message received:", event.data);
        // 处理接收到的消息
      };

      this.socket.onclose = () => {
        console.log("WebSocket disconnected");
        this.isConnected = false;
        this.socket = null;
      };

      this.socket.onerror = (error) => {
        console.error("WebSocket error:", error);
        this.isConnected = false;
      };
    },
    sendMessage(message: any) {
      if (this.socket && this.isConnected) {
        this.socket.send(JSON.stringify(message));
      } else {
        console.error("WebSocket is not connected");
      }
    },
  },
});

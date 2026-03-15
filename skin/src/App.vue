<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import Chat from './components/Chat.vue';
import Status from './components/Status.vue';
import Logs from './components/Logs.vue';
import Sidebar from './components/Sidebar.vue';

interface LogMessage {
  time: string;
  content: string;
}

const logs = ref<LogMessage[]>([]);
const isConnected = ref(false);
const currentQuestion = ref<string | null>(null);
const answerInput = ref('');
const lastReply = ref<any>(null);
const sidebarRef = ref<any>(null);
const selectedConversationId = ref<string | null>(null);
const currentFsmState = ref<string | null>(null);

let socket: WebSocket | null = null;
const clientId = `web-client-${Math.floor(Math.random() * 1000)}`;

const connectWebSocket = () => {
  socket = new WebSocket(`ws://localhost:8080/ws/${clientId}`);

  socket.onopen = () => {
    console.log('WebSocket connected');
    isConnected.value = true;
    logs.value.push({
      time: new Date().toLocaleTimeString(),
      content: 'Connected to WebSocket server'
    });
  };

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      console.log('Received:', msg);

      if (msg.action === 'Log') {
        logs.value.push(msg.data);
      } else if (msg.action === 'Ping') {
        if (msg.action === 'Ping') {
           socket?.send(JSON.stringify({ action: 'Pong', data: 'Pong' }));
        }
      } else if (msg.action === 'Ask') {
        currentQuestion.value = msg.data;
        logs.value.push({
            time: new Date().toLocaleTimeString(),
            content: `Received Question: ${msg.data}`
        });
      } else if (msg.action === 'Reply') {
        lastReply.value = msg.data;
        logs.value.push({
            time: new Date().toLocaleTimeString(),
            content: `Received Reply: ${msg.data.content}`
        });
      } else if (msg.action === 'State') {
        currentFsmState.value = msg.data;
      }
    } catch (e) {
      console.error('Error parsing message:', e);
    }
  };

  socket.onclose = () => {
    console.log('WebSocket disconnected');
    isConnected.value = false;
    logs.value.push({
      time: new Date().toLocaleTimeString(),
      content: 'Disconnected from WebSocket server'
    });
    // Optional: reconnect logic
  };

  socket.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
};

const sendAnswer = () => {
  if (socket && isConnected.value && currentQuestion.value) {
    const message = {
      action: 'Answer',
      data: answerInput.value
    };
    socket.send(JSON.stringify(message));
    logs.value.push({
        time: new Date().toLocaleTimeString(),
        content: `Sent Answer: ${answerInput.value}`
    });
    currentQuestion.value = null;
    answerInput.value = '';
  }
};

const sendTask = (taskData: any) => {
  if (socket && isConnected.value) {
    const message = {
      action: 'Task',
      data: taskData
    };
    socket.send(JSON.stringify(message));
    logs.value.push({
        time: new Date().toLocaleTimeString(),
        content: `Sent task: ${taskData.Id}`
    });
  } else {
    alert('WebSocket is not connected');
  }
};

const handleSelectConversation = (id: string) => {
  selectedConversationId.value = id;
};

const handleNewChat = () => {
  selectedConversationId.value = null;
};

const handleConversationCreated = (id: string) => {
  selectedConversationId.value = id;
  // Refresh sidebar list
  if (sidebarRef.value) {
    sidebarRef.value.fetchConversations();
  }
};

onMounted(() => {
  connectWebSocket();
});

onUnmounted(() => {
  if (socket) {
    socket.close();
  }
});
</script>

<template>
  <div class="app-container">
    <Sidebar 
      ref="sidebarRef"
      @select-conversation="handleSelectConversation"
      @new-chat="handleNewChat"
    />
    
    <div class="main-content">
      <div class="chat-wrapper">
         <Chat 
           :sendTask="sendTask" 
           :lastReply="lastReply" 
           :conversationId="selectedConversationId"
           @conversation-created="handleConversationCreated"
         />
      </div>
    </div>
    
    <div class="right-panel">
      <div class="status-bar">
        <h3>Status</h3>
        <span :class="['status-indicator', isConnected ? 'connected' : 'disconnected']">
          {{ isConnected ? 'Connected' : 'Disconnected' }}
        </span>
      </div>
      <Status :realtimeState="currentFsmState" />
      <div class="logs-container">
        <h3>Logs</h3>
        <Logs :logs="logs" />
      </div>
    </div>

    <div v-if="currentQuestion" class="modal-overlay">
      <div class="modal-content">
        <h3>New Question!</h3>
        <p>{{ currentQuestion }}</p>
        <div class="input-group">
          <input v-model="answerInput" placeholder="Type your answer..." @keyup.enter="sendAnswer" />
          <button @click="sendAnswer">Submit Answer</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: #343541;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
}

.chat-wrapper {
  flex: 1;
  height: 100%;
  overflow: hidden;
}

.right-panel {
  width: 300px;
  background-color: #202123;
  border-left: 1px solid #4d4d4f;
  display: flex;
  flex-direction: column;
  padding: 1rem;
  overflow-y: auto;
  color: white;
  flex-shrink: 0;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #4d4d4f;
}

.status-indicator {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: bold;
}

.status-indicator.connected {
  background-color: #4caf50;
  color: white;
}

.status-indicator.disconnected {
  background-color: #f44336;
  color: white;
}

.logs-container {
  margin-top: 1rem;
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: #343541;
  color: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  width: 90%;
  max-width: 500px;
  text-align: center;
  border: 1px solid #565869;
}

.modal-content h3 {
  color: #ff9800;
  margin-top: 0;
}

.modal-content p {
  font-size: 1.1rem;
  margin: 1.5rem 0;
}

.modal-content .input-group {
  display: flex;
  gap: 0.5rem;
}

.modal-content input {
  flex: 1;
  padding: 0.8rem;
  border: 1px solid #565869;
  border-radius: 4px;
  font-size: 1rem;
  background-color: #40414f;
  color: white;
}

.modal-content button {
  padding: 0.8rem 1.2rem;
  background-color: #ff9800;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

.modal-content button:hover {
  background-color: #e68900;
}
</style>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import Chat from './components/Chat.vue';
import Status from './components/Status.vue';
import Logs from './components/Logs.vue';

interface LogMessage {
  time: string;
  content: string;
}

const logs = ref<LogMessage[]>([]);
const isConnected = ref(false);
const currentQuestion = ref<string | null>(null);
const answerInput = ref('');
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
    <header>
      <h1>Aimin Dashboard</h1>
      <span :class="['status-indicator', isConnected ? 'connected' : 'disconnected']">
        {{ isConnected ? 'Connected' : 'Disconnected' }}
      </span>
    </header>

    <main>
      <div class="left-panel">
        <Status />
        <Chat :sendTask="sendTask" />
      </div>
      <div class="right-panel">
        <Logs :logs="logs" />
      </div>
    </main>

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
  background-color: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 500px;
  text-align: center;
}

.modal-content h3 {
  color: #e65100;
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
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
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
  background-color: #f57c00;
}

.app-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  font-family: Arial, sans-serif;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  border-bottom: 2px solid #eee;
  padding-bottom: 1rem;
}

.status-indicator {
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-weight: bold;
  font-size: 0.9rem;
}

.connected {
  background-color: #e6fffa;
  color: #047857;
  border: 1px solid #047857;
}

.disconnected {
  background-color: #fff5f5;
  color: #c53030;
  border: 1px solid #c53030;
}

main {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
}

@media (max-width: 768px) {
  main {
    grid-template-columns: 1fr;
  }
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.right-panel {
  display: flex;
  flex-direction: column;
}
</style>

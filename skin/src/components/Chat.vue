<template>
  <div class="chat-container">
    <h3>Chat</h3>
    <div class="input-group">
      <input 
        v-model="message" 
        placeholder="Type your message..." 
        @keyup.enter="send"
      />
      <button @click="send" :disabled="!message.trim()">Send</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  sendTask: (data: any) => void
}>();

const message = ref('');

const send = () => {
  if (!message.value.trim()) return;
  
  const taskId = `task-${Date.now()}`;
  props.sendTask({
    Id: taskId,
    Type: 'conversation',
    Payload: message.value
  });
  
  message.value = '';
};
</script>

<style scoped>
.chat-container {
  border: 1px solid #ccc;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 8px;
}
.input-group {
  display: flex;
  gap: 0.5rem;
}
.input-group input {
  flex: 1;
  padding: 0.5rem;
}
</style>

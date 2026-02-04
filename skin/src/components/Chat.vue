<template>
  <div class="chat-container">
    <div v-if="!currentConversationId && messages.length === 0" class="welcome-screen">
      <h1>How can I help you today?</h1>
    </div>

    <div class="messages" ref="messagesContainer">
      <div v-for="(msg, index) in messages" :key="index" :class="['message', msg.role]">
        <div class="message-inner">
          <div class="avatar">
            <div v-if="msg.role === 'user'" class="user-avatar">U</div>
            <div v-else class="ai-avatar">AI</div>
          </div>
          <div class="content-wrapper">
             <div class="role-name">{{ msg.role === 'user' ? 'User' : 'AI' }}</div>
             <div class="message-content">{{ msg.content }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="input-area">
      <div class="input-group">
        <textarea 
          v-model="message" 
          placeholder="Send a message..." 
          @keydown.enter.prevent="handleEnter"
          rows="1"
          ref="textarea"
        ></textarea>
        <button @click="send" :disabled="!canSend" class="send-btn">
          <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';

const props = defineProps<{
  sendTask: (data: any) => void;
  lastReply: any;
  conversationId: string | null;
}>();

const emit = defineEmits(['conversation-created']);

const message = ref('');
const messages = ref<Array<{role: string, content: string}>>([]);
const messagesContainer = ref<HTMLElement | null>(null);
const currentConversationId = ref<string | null>(props.conversationId);

const canSend = computed(() => {
  return message.value.trim().length > 0;
});

const fetchHistory = async (id: string) => {
  try {
    const response = await fetch(`http://localhost:8080/api/v1/conversations/${id}/messages`);
    if (response.ok) {
      const res = await response.json();
      if (res.code === 0) {
        messages.value = (res.data || []).map((m: any) => ({
          role: m.role,
          content: m.content
        }));
        scrollToBottom();
      }
    }
  } catch (error) {
    console.error('Failed to fetch history:', error);
  }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
};

const handleEnter = (e: KeyboardEvent) => {
  if (!e.shiftKey) {
    send();
  }
};

watch(() => props.conversationId, (newId) => {
  // If the new ID matches what we just set locally via lastReply (new conversation flow),
  // and we already have messages (user input + AI reply), prevent reloading to avoid race conditions
  if (newId && newId === currentConversationId.value && messages.value.length > 0) {
      return;
  }

  currentConversationId.value = newId;
  if (newId) {
    fetchHistory(newId);
  } else {
    messages.value = [];
  }
});

watch(() => props.lastReply, (newReply) => {
  if (newReply) {
    // Check if this reply belongs to the current conversation
    // Or if it's a new conversation start
    if (!currentConversationId.value) {
        // This was a new conversation
        currentConversationId.value = newReply.conversation_id;
        emit('conversation-created', newReply.conversation_id);
    }
    
    if (currentConversationId.value === newReply.conversation_id) {
       messages.value.push({
         role: 'assistant',
         content: newReply.content
       });
       scrollToBottom();
    }
  }
});

const send = () => {
  if (!canSend.value) return;
  
  const taskId = `task-${Date.now()}`;
  let type = '';
  let payload: any = null;

  if (!currentConversationId.value) {
    type = 'conversation_create';
    payload = message.value;
  } else {
    type = 'conversation_continue';
    payload = {
      conversation_id: currentConversationId.value,
      question: message.value
    };
  }

  props.sendTask({ Type: type, Payload: payload, ID: taskId });
  messages.value.push({ role: 'user', content: message.value });
  message.value = '';
  scrollToBottom();
};
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
  background-color: #343541;
  color: #ececf1;
}

.welcome-screen {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ececf1;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding-bottom: 120px; /* Space for input */
}

.message {
  padding: 1.5rem;
  display: flex;
  justify-content: center;
  border-bottom: 1px solid #2d2d3e;
}

.message-inner {
  display: flex;
  gap: 1.5rem;
  max-width: 800px;
  width: 100%;
}

.avatar {
  flex-shrink: 0;
  width: 30px;
  height: 30px;
}

.user-avatar, .ai-avatar {
  width: 100%;
  height: 100%;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.8rem;
}

.user-avatar {
  background-color: #5436DA;
  color: white;
}

.ai-avatar {
  background-color: #10a37f;
  color: white;
}

.content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.role-name {
  font-weight: bold;
  font-size: 0.9rem;
  opacity: 0.9;
}

.message.user .message-inner {
  flex-direction: row-reverse;
}

.message.user .content-wrapper {
  align-items: flex-end;
}

.message.user .role-name {
  text-align: right;
}

.message.user {
  background-color: #343541;
}

.message.assistant {
  background-color: #444654;
}

.message-content {
  width: 100%;
  line-height: 1.6;
}

.input-area {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  background-image: linear-gradient(180deg, rgba(53,55,64,0), #353740 58.85%);
  padding: 2rem 0;
  display: flex;
  justify-content: center;
}

.input-group {
  position: relative;
  width: 100%;
  max-width: 768px;
  background-color: #40414f;
  border-radius: 0.75rem;
  box-shadow: 0 0 15px rgba(0,0,0,0.1);
  display: flex;
  align-items: flex-end;
  padding: 0.75rem;
  border: 1px solid rgba(32,33,35,0.5);
}

textarea {
  width: 100%;
  background-color: transparent;
  border: none;
  color: white;
  font-family: inherit;
  font-size: 1rem;
  resize: none;
  max-height: 200px;
  outline: none;
  padding-right: 2rem;
}

.send-btn {
  position: absolute;
  right: 0.75rem;
  bottom: 0.75rem;
  background-color: transparent;
  border: none;
  color: #ececf1;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: background-color 0.2s;
}

.send-btn:hover:not(:disabled) {
  background-color: #202123;
}

.send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>

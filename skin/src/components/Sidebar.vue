<script setup lang="ts">
import { ref, onMounted } from 'vue';

const conversations = ref<any[]>([]);
const selectedId = ref<string | null>(null);
const emit = defineEmits(['select-conversation', 'new-chat']);

const fetchConversations = async () => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/conversations');
    const res = await response.json();
    if (res.code === 0) {
      conversations.value = res.data;
    }
  } catch (error) {
    console.error('Failed to fetch conversations:', error);
  }
};

const selectConversation = (id: string) => {
  selectedId.value = id;
  emit('select-conversation', id);
};

const newChat = () => {
  selectedId.value = null;
  emit('new-chat');
};

onMounted(() => {
  fetchConversations();
});

defineExpose({
  fetchConversations
});
</script>

<template>
  <div class="sidebar">
    <button class="new-chat-btn" @click="newChat">+ New Chat</button>
    <div class="conversation-list">
      <div 
        v-for="conv in conversations" 
        :key="conv.id" 
        :class="['conversation-item', { active: selectedId === conv.id }]"
        @click="selectConversation(conv.id)"
      >
        <div class="conv-title">{{ conv.topic || conv.id }}</div>
        <div class="conv-date">{{ new Date(conv.updated_at).toLocaleString() }}</div>
      </div>
    </div>
    <button class="refresh-btn" @click="fetchConversations">Refresh List</button>
  </div>
</template>

<style scoped>
.sidebar {
  width: 260px;
  background-color: #202123;
  color: white;
  display: flex;
  flex-direction: column;
  padding: 0.5rem;
  height: 100%;
  border-right: 1px solid #4d4d4f;
  flex-shrink: 0;
}

.new-chat-btn {
  background-color: transparent;
  border: 1px solid #565869;
  color: white;
  padding: 0.75rem;
  border-radius: 0.375rem;
  text-align: left;
  margin-bottom: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
  width: 100%;
}

.new-chat-btn:hover {
  background-color: #2a2b32;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 0.5rem;
}

.conversation-item {
  padding: 0.75rem;
  border-radius: 0.375rem;
  cursor: pointer;
  margin-bottom: 0.25rem;
  transition: background-color 0.2s;
}

.conversation-item:hover {
  background-color: #2a2b32;
}

.conversation-item.active {
  background-color: #343541;
}

.conv-title {
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conv-date {
  font-size: 0.75rem;
  color: #8e8ea0;
  margin-top: 0.25rem;
}

.refresh-btn {
    background-color: #343541;
    color: white;
    border: 1px solid #565869;
    padding: 0.5rem;
    cursor: pointer;
    border-radius: 0.375rem;
    width: 100%;
}

.refresh-btn:hover {
    background-color: #40414f;
}

/* Custom scrollbar */
.conversation-list::-webkit-scrollbar {
  width: 6px;
}

.conversation-list::-webkit-scrollbar-track {
  background: transparent;
}

.conversation-list::-webkit-scrollbar-thumb {
  background-color: #565869;
  border-radius: 3px;
}
</style>

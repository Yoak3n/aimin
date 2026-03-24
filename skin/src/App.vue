<script setup lang="ts">
import { useAppStore } from "@/store/module/app";

const appStore = useAppStore();
appStore.initWebSocket();
</script>

<template>
  <router-view />
</template>

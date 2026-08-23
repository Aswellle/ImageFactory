<script setup lang="ts">
import { RouterView } from 'vue-router'
import PageLoadingBar from './components/PageLoadingBar.vue'
</script>

<template>
  <PageLoadingBar />
  <RouterView />
</template>

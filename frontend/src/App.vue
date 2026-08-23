<script setup lang="ts">
import { RouterView } from 'vue-router'
import PageLoadingBar from './components/PageLoadingBar.vue'
import ToastContainer from './components/ToastContainer.vue'
</script>

<template>
  <PageLoadingBar />
  <RouterView />
  <ToastContainer />
</template>

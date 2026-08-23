<script setup lang="ts">
defineProps<{
  type?: 'card' | 'list' | 'text'
}>()
</script>

<template>
  <!-- Card skeleton: image + text lines -->
  <div v-if="type === 'card'" class="card overflow-hidden animate-fade-in">
    <div class="skeleton aspect-[4/3]" />
    <div class="p-3 space-y-2">
      <div class="skeleton h-3 w-3/4" />
      <div class="skeleton h-3 w-1/2" />
    </div>
  </div>

  <!-- List skeleton: avatar + text lines -->
  <div v-else-if="type === 'list'" class="flex items-center gap-4 px-4 py-3 animate-fade-in">
    <div class="skeleton w-16 h-16 rounded-md shrink-0" />
    <div class="flex-1 space-y-2">
      <div class="skeleton h-3 w-3/4" />
      <div class="skeleton h-3 w-1/2" />
    </div>
  </div>

  <!-- Text skeleton: simple lines -->
  <div v-else class="space-y-2 animate-fade-in">
    <div class="skeleton h-3 w-full" />
    <div class="skeleton h-3 w-5/6" />
    <div class="skeleton h-3 w-4/6" />
  </div>
</template>

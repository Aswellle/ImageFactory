<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
  hint?: string
  tone?: 'default' | 'success' | 'danger' | 'accent'
}>()

const ring: Record<string, string> = {
  default: 'text-text-secondary',
  success: 'text-green-500',
  danger: 'text-red-500',
  accent: 'text-accent',
}
</script>

<template>
  <div class="if-card p-5">
    <div class="text-xs text-text-secondary uppercase tracking-wide">{{ label }}</div>
    <div
      class="text-2xl font-semibold mt-1.5 tabular-nums"
      :class="tone ? ring[tone] : 'text-text'"
    >
      {{ value }}
    </div>
    <div v-if="hint" class="text-xs text-text-secondary mt-1">{{ hint }}</div>
  </div>
</template>

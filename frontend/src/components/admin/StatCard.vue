<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
  hint?: string
  tone?: 'default' | 'success' | 'danger' | 'accent'
}>()

const ring: Record<string, string> = {
  default: 'text-[var(--text-secondary)]',
  success: 'text-[var(--success)]',
  danger: 'text-[var(--danger)]',
  accent: 'text-[var(--accent)]',
}
</script>

<template>
  <div class="bento-tile">
    <div class="text-caption uppercase tracking-wide">{{ label }}</div>
    <div
      class="metric-value mt-2"
      :class="tone ? ring[tone] : 'text-[var(--text)]'"
    >
      {{ value }}
    </div>
    <div v-if="hint" class="text-caption mt-1.5">{{ hint }}</div>
  </div>
</template>

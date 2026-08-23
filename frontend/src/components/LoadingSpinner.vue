<script setup lang="ts">
defineProps<{
  size?: 'sm' | 'md' | 'lg'
  label?: string
}>()

const sizeClass = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-10 w-10',
}
</script>

<template>
  <svg
    :class="[sizeClass[size ?? 'md'], 'animate-spin text-[var(--accent)]']"
    viewBox="0 0 24 24"
    fill="none"
    role="status"
    :aria-label="label ?? 'Loading'"
    aria-live="polite"
  >
    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
  </svg>
</template>

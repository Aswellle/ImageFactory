<script setup lang="ts">
import { computed } from 'vue'

export type FieldHintType = 'error' | 'success' | 'warning' | 'info'

interface Props {
  type: FieldHintType
  message: string
}

const props = defineProps<Props>()

const icon = computed((): string => {
  const map: Record<FieldHintType, string> = {
    error: 'exclamationCircle',
    success: 'checkCircle',
    warning: 'exclamationTriangle',
    info: 'infoCircle'
  }
  return map[props.type]
})

const color = computed((): string => {
  const map: Record<FieldHintType, string> = {
    error: 'var(--danger)',
    success: 'var(--success)',
    warning: 'var(--warning)',
    info: 'var(--accent)'
  }
  return map[props.type]
})
</script>

<template>
  <Transition name="field-hint">
    <p v-if="message" class="field-hint" role="alert">
      <svg
        v-if="icon === 'exclamationCircle'"
        class="field-hint-icon"
        :style="{ color }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <svg
        v-else-if="icon === 'checkCircle'"
        class="field-hint-icon"
        :style="{ color }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
        <polyline points="22 4 12 14.01 9 11.01" />
      </svg>
      <svg
        v-else-if="icon === 'exclamationTriangle'"
        class="field-hint-icon"
        :style="{ color }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      <svg
        v-else-if="icon === 'infoCircle'"
        class="field-hint-icon"
        :style="{ color }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
      <span>{{ message }}</span>
    </p>
  </Transition>
</template>

<style scoped>
.field-hint-enter-active,
.field-hint-leave-active {
  transition: all 0.2s ease;
}

.field-hint-enter-from,
.field-hint-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.field-hint {
  display: flex;
  align-items: flex-start;
  gap: 0.375rem;
  margin-top: 0.375rem;
  font-size: 0.75rem;
  line-height: 1.4;
  color: var(--text-secondary);
}

.field-hint-icon {
  flex-shrink: 0;
  width: 0.875rem;
  height: 0.875rem;
  margin-top: 0.125rem;
}
</style>

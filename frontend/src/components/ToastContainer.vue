<script setup lang="ts">
import { computed } from 'vue'
import { useToastStore, type ToastType } from '@/stores/toast'

const store = useToastStore()

const iconMap: Record<ToastType, string> = {
  success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
  error: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
  warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z',
  info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
}

const colorMap: Record<ToastType, { bg: string; border: string; icon: string; accent: string }> = {
  success: {
    bg: 'bg-[color-mix(in_srgb,var(--success)_8%,var(--surface))]',
    border: 'border-[color-mix(in_srgb,var(--success)_20%,transparent)]',
    icon: 'text-[var(--success)]',
    accent: 'var(--success)',
  },
  error: {
    bg: 'bg-[color-mix(in_srgb,var(--danger)_8%,var(--surface))]',
    border: 'border-[color-mix(in_srgb,var(--danger)_20%,transparent)]',
    icon: 'text-[var(--danger)]',
    accent: 'var(--danger)',
  },
  warning: {
    bg: 'bg-[color-mix(in_srgb,var(--warning)_8%,var(--surface))]',
    border: 'border-[color-mix(in_srgb,var(--warning)_20%,transparent)]',
    icon: 'text-[var(--warning)]',
    accent: 'var(--warning)',
  },
  info: {
    bg: 'bg-[color-mix(in_srgb,var(--accent)_8%,var(--surface))]',
    border: 'border-[color-mix(in_srgb,var(--accent)_20%,transparent)]',
    icon: 'text-[var(--accent)]',
    accent: 'var(--accent)',
  },
}

function getColors(type: ToastType) {
  return colorMap[type]
}

function handleMouseEnter(id: number) {
  store.pause(id)
}

function handleMouseLeave(id: number) {
  store.resume(id)
}

function handleAction(id: number, onClick: () => void) {
  onClick()
  store.dismiss(id)
}

const orderedToasts = computed(() => store.toasts.slice(0, 5))
</script>

<template>
  <div
    aria-live="polite"
    aria-atomic="true"
    class="fixed bottom-6 right-6 z-[var(--z-toast)] flex flex-col gap-3 max-w-sm w-full pointer-events-none"
  >
    <TransitionGroup
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 translate-x-8 scale-95"
      enter-to-class="opacity-100 translate-x-0 scale-100"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 translate-x-0 scale-100"
      leave-to-class="opacity-0 translate-x-8 scale-95"
      move-class="transition-all duration-300 ease-out"
    >
      <div
        v-for="toast in orderedToasts"
        :key="toast.id"
        class="pointer-events-auto rounded-2xl border backdrop-blur-xl shadow-lg overflow-hidden"
        :class="[
          getColors(toast.type).bg,
          getColors(toast.type).border,
          toast.duration === 0 ? 'pr-2' : 'pr-3',
        ]"
        @mouseenter="handleMouseEnter(toast.id)"
        @mouseleave="handleMouseLeave(toast.id)"
      >
        <div class="flex items-start gap-3 p-4">
          <!-- Icon -->
          <div class="shrink-0 mt-0.5">
            <svg
              class="h-5 w-5"
              :class="getColors(toast.type).icon"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path :d="iconMap[toast.type]" />
            </svg>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-[var(--text)] leading-snug">
              {{ toast.title }}
            </p>
            <p
              v-if="toast.message"
              class="mt-0.5 text-sm text-[var(--text-secondary)] leading-relaxed"
            >
              {{ toast.message }}
            </p>
            <!-- Action button -->
            <button
              v-if="toast.action"
              class="mt-2 text-sm font-medium underline-offset-2 hover:underline focus:outline-none focus-underline"
              :style="{ color: getColors(toast.type).accent }"
              @click="handleAction(toast.id, toast.action.onClick)"
            >
              {{ toast.action.label }}
            </button>
          </div>

          <!-- Close button -->
          <button
            class="shrink-0 p-1 rounded-lg text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--surface-2)] transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
            :aria-label="`Dismiss notification: ${toast.title}`"
            @click="store.dismiss(toast.id)"
          >
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>

          <!-- Progress bar -->
          <div
            v-if="toast.duration > 0"
            class="absolute bottom-0 left-0 right-0 h-0.5 bg-[var(--surface-2)]"
          >
            <div
              class="h-full origin-left"
              :style="{
                backgroundColor: getColors(toast.type).accent,
                animation: `shrink ${toast.duration}ms linear forwards`,
              }"
            />
          </div>

          <!-- Spinner for persistent toasts -->
          <div
            v-if="toast.duration === 0"
            class="absolute bottom-0 left-0 right-0 h-0.5"
          >
            <div
              class="h-full w-1/3 animate-pulse rounded-full"
              :style="{ backgroundColor: getColors(toast.type).accent }"
            />
          </div>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style>
@keyframes shrink {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}
</style>

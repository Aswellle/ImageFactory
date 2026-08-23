<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useFavoriteStore } from '@/stores/favorite'

const props = defineProps<{
  assetId: number
  size?: 'sm' | 'md' | 'lg'
  /** Initial favorited state from parent (avoids per-button API call). */
  initialFavorited?: boolean
}>()

const emit = defineEmits<{
  toggled: [isFavorited: boolean]
}>()

const store = useFavoriteStore()
const isFav = ref(props.initialFavorited ?? false)
const popping = ref(false)
let popTimer: ReturnType<typeof setTimeout> | null = null
const loading = ref(false)

// Only check individually if parent didn't provide initial state.
onMounted(() => {
  if (props.initialFavorited !== undefined) return
  isFav.value = store.isFavorited(props.assetId)
})

onUnmounted(() => {
  if (popTimer) clearTimeout(popTimer)
})

async function toggle() {
  if (loading.value) return
  loading.value = true
  try {
    if (isFav.value) {
      await store.removeFavorite(props.assetId)
      isFav.value = false
      emit('toggled', false)
    } else {
      await store.addFavorite(props.assetId)
      isFav.value = true
      popping.value = true
      popTimer = setTimeout(() => { popping.value = false }, 400)
      emit('toggled', true)
    }
  } finally {
    loading.value = false
  }
}

const sizeClass = {
  sm: 'w-4 h-4',
  md: 'w-5 h-5',
  lg: 'w-6 h-6',
}[props.size ?? 'md']
</script>

<template>
  <button
    type="button"
    :disabled="loading"
    class="inline-flex items-center justify-center rounded-md transition-all duration-150 hover:bg-surface-hover active:scale-90 disabled:opacity-50"
    :aria-label="isFav ? 'Remove from favorites' : 'Add to favorites'"
    @click.prevent="toggle"
  >
    <svg
      :class="[
        sizeClass,
        isFav ? 'text-[var(--danger)]' : 'text-text-muted',
        popping ? 'animate-heart-pop' : ''
      ]"
      :fill="isFav ? 'currentColor' : 'none'"
      viewBox="0 0 24 24"
      stroke="currentColor"
      stroke-width="2"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
      />
    </svg>
  </button>
</template>

<style scoped>
@keyframes heartPop {
  0% { transform: scale(1); }
  25% { transform: scale(1.3); }
  50% { transform: scale(0.95); }
  100% { transform: scale(1); }
}
.animate-heart-pop {
  animation: heartPop 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
@media (prefers-reduced-motion: reduce) {
  .animate-heart-pop { animation: none; }
}
</style>

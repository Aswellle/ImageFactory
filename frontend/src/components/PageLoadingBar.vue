<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, type RouteLocationNormalized } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const isLoading = ref(false)

// Determine if a navigation will be an auth redirect (no loading bar needed).
function willRedirect(to: RouteLocationNormalized): boolean {
  const auth = useAuthStore()
  if (to.meta.public) return false
  if (!auth.isAuthenticated) return true
  if (to.meta.requiresAdmin && !auth.isAdmin) return true
  if (to.name === 'landing' && auth.isAuthenticated) return true
  return false
}

router.beforeEach((to) => {
  if (willRedirect(to)) return
  isLoading.value = true
})

router.afterEach(() => {
  isLoading.value = false
})

router.onError(() => {
  isLoading.value = false
})
</script>

<template>
  <Transition
    enter-active-class="transition-opacity duration-150"
    leave-active-class="transition-opacity duration-150"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
  >
    <div
      v-if="isLoading"
      class="fixed top-0 left-0 right-0 z-[var(--z-toast)] flex justify-center"
      role="status"
      aria-live="polite"
      aria-label="Loading page"
    >
      <div class="h-1 w-32 bg-[var(--surface-2)] rounded-full overflow-hidden mt-0">
        <div class="h-full w-1/2 bg-[var(--accent)] animate-loading-bar rounded-full" />
      </div>
    </div>
  </Transition>
</template>

<style scoped>
@keyframes loading-bar {
  0% {
    transform: translateX(-100%);
  }
  50% {
    transform: translateX(100%);
  }
  100% {
    transform: translateX(300%);
  }
}
.animate-loading-bar {
  animation: loading-bar 1s ease-in-out infinite;
}
</style>

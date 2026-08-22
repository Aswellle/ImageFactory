<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'
import { computed } from 'vue'

interface NavItem {
  to: string
  label: string
  // Simple text-icon abbreviations keep the sidebar dependency-free (no icon lib).
  icon: string
}

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()

const nav: NavItem[] = [
  { to: '/admin', label: 'Dashboard', icon: '◧' },
  { to: '/admin/users', label: 'Users', icon: '⬢' },
  { to: '/admin/jobs', label: 'Jobs', icon: '⟳' },
  { to: '/admin/api-keys', label: 'API Keys', icon: '🗝' },
]

function isActive(to: string): boolean {
  if (to === '/admin') return router.currentRoute.value.path === '/admin'
  return router.currentRoute.value.path.startsWith(to)
}

const currentTitle = computed(() => {
  const path = router.currentRoute.value.path
  if (path === '/admin') return 'Dashboard'
  if (path.startsWith('/admin/users')) return 'Users'
  if (path.startsWith('/admin/jobs')) return 'Jobs'
  if (path.startsWith('/admin/api-keys')) return 'API Keys'
  return 'Admin'
})

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen flex bg-bg">
    <!-- Sidebar -->
    <aside
      class="hidden md:flex flex-col w-56 border-r border-border bg-surface px-4 py-5 shrink-0"
    >
      <div class="flex items-center gap-2 px-2 mb-8">
        <div class="w-7 h-7 rounded-md bg-accent"></div>
        <span class="text-base font-semibold tracking-tight">ImageForge</span>
        <span class="ml-auto text-[10px] px-1.5 py-0.5 rounded bg-accent/10 text-accent font-medium uppercase">Admin</span>
      </div>

      <nav class="flex-1 space-y-1">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
          :class="{ '!bg-accent/10 !text-accent': isActive(item.to) }"
        >
          <span class="w-4 text-center text-xs opacity-70">{{ item.icon }}</span>
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="border-t border-border pt-3 mt-4 space-y-1">
        <RouterLink
          to="/"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
        >
          <span class="w-4 text-center text-xs opacity-70">←</span>
          Back to app
        </RouterLink>
        <button
          class="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          @click="app.setTheme(app.theme === 'dark' ? 'light' : 'dark')"
        >
          <span class="w-4 text-center text-xs opacity-70">{{ app.theme === 'dark' ? '☀' : '☾' }}</span>
          {{ app.theme === 'dark' ? 'Light' : 'Dark' }} mode
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 border-b border-border bg-surface px-6 flex items-center justify-between shrink-0">
        <div class="text-base font-semibold tracking-tight">{{ currentTitle }}</div>
        <div class="flex items-center gap-3">
          <span class="text-sm text-text-secondary hidden sm:inline">{{ auth.user?.email }}</span>
          <button class="if-btn-ghost text-sm" @click="logout">Sign out</button>
        </div>
      </header>

      <main class="flex-1 overflow-auto">
        <div class="max-w-6xl mx-auto px-6 py-8">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

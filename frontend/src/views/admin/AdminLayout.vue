<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'
import { computed } from 'vue'

interface NavItem {
  to: string
  label: string
}

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()

const nav: NavItem[] = [
  { to: '/admin', label: 'Dashboard' },
  { to: '/admin/users', label: 'Users' },
  { to: '/admin/jobs', label: 'Jobs' },
  { to: '/admin/api-keys', label: 'API Keys' },
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
  <div class="min-h-[100dvh] flex bg-[var(--bg)]">
    <!-- Sidebar -->
    <aside class="hidden md:flex flex-col w-56 border-r border-[var(--border)] bg-[var(--surface)] px-4 py-5 shrink-0">
      <div class="flex items-center gap-2.5 px-2 mb-8">
        <div class="w-7 h-7 rounded-lg bg-[var(--accent)]"></div>
        <span class="text-sm font-semibold tracking-tight text-[var(--text)]">ImageForge</span>
        <span class="ml-auto text-[10px] px-1.5 py-0.5 rounded-md bg-[var(--accent-soft)] text-[var(--accent)] font-medium uppercase">Admin</span>
      </div>

      <nav class="flex-1 space-y-1">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-colors"
          :class="{ 'bg-[var(--accent-soft)] text-[var(--accent)]': isActive(item.to) }"
        >
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="border-t border-[var(--border)] pt-3 mt-4 space-y-1">
        <RouterLink
          to="/app"
          class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
          Back to app
        </RouterLink>
        <button
          class="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-colors"
          @click="app.setTheme(app.theme === 'dark' ? 'light' : 'dark')"
        >
          <svg v-if="app.theme === 'dark'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5" /><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" /></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" /></svg>
          {{ app.theme === 'dark' ? 'Light' : 'Dark' }} mode
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 border-b border-[var(--border)] bg-[var(--surface)] px-6 flex items-center justify-between shrink-0">
        <div class="text-sm font-semibold tracking-tight text-[var(--text)]">{{ currentTitle }}</div>
        <div class="flex items-center gap-4">
          <span class="text-caption hidden sm:inline">{{ auth.user?.email }}</span>
          <button class="btn btn-ghost text-xs" @click="logout">Sign out</button>
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

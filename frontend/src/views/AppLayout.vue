<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'
import { useKeyboard } from '@/composables/useKeyboard'

const auth = useAuthStore()
const app = useAppStore()
const router = useRouter()
useKeyboard(router)

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
      </div>

      <nav class="flex-1 space-y-1">
        <RouterLink
          to="/"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
        >
          Workspace
        </RouterLink>
        <RouterLink
          to="/create"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
        >
          Create
        </RouterLink>
        <RouterLink
          to="/assets"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
        >
          Gallery
        </RouterLink>
        <RouterLink
          to="/projects"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
        >
          Projects
        </RouterLink>
        <RouterLink
          to="/api-keys"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          active-class="!bg-surface-2 !text-text"
        >
          API
        </RouterLink>
      </nav>

      <div class="border-t border-border pt-3 mt-4 space-y-1">
        <button
          class="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-text-secondary hover:bg-surface-2 hover:text-text transition-colors"
          @click="app.setTheme(app.theme === 'dark' ? 'light' : 'dark')"
        >
          {{ app.theme === 'dark' ? 'Light' : 'Dark' }} mode
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 border-b border-border bg-surface px-6 flex items-center justify-between shrink-0">
        <div class="md:hidden text-base font-semibold tracking-tight">ImageForge</div>
        <div class="flex items-center gap-3 ml-auto">
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

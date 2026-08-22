<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useKeyboard } from '@/composables/useKeyboard'

const { t } = useI18n()
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
  <div class="min-h-screen flex bg-[var(--bg)]">
    <!-- Sidebar -->
    <aside
      class="hidden md:flex flex-col w-56 border-r border-[var(--border)] bg-[var(--surface)] px-4 py-5 shrink-0"
    >
      <div class="flex items-center gap-2 px-2 mb-8">
        <div class="w-7 h-7 rounded-md bg-[var(--accent)]"></div>
        <span class="text-base font-semibold tracking-tight">{{ t('common.appName') }}</span>
      </div>

      <nav class="flex-1 space-y-1">
        <RouterLink
          to="/"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.workspace') }}
        </RouterLink>
        <RouterLink
          to="/create"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.create') }}
        </RouterLink>
        <RouterLink
          to="/assets"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.gallery') }}
        </RouterLink>
        <RouterLink
          to="/projects"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.projects') }}
        </RouterLink>
        <RouterLink
          to="/templates"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.templates') }}
        </RouterLink>
        <RouterLink
          to="/api-keys"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          active-class="!bg-[var(--surface-2)] !text-[var(--text)]"
        >
          {{ t('nav.api') }}
        </RouterLink>
      </nav>

      <div class="border-t border-[var(--border)] pt-3 mt-4 space-y-1">
        <!-- Language switcher -->
        <div class="flex items-center gap-1 px-2 py-1.5">
          <button
            v-for="loc in app.availableLocales"
            :key="loc.code"
            type="button"
            class="flex-1 text-center text-xs px-2 py-1 rounded transition-colors"
            :class="app.locale === loc.code ? 'bg-[var(--surface-2)] text-[var(--text)] font-medium' : 'text-[var(--text-secondary)] hover:text-[var(--text)]'"
            @click="app.setLocale(loc.code)"
          >
            {{ loc.flag }} {{ loc.code.toUpperCase() }}
          </button>
        </div>
        <button
          class="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-[var(--text-secondary)] hover:bg-[var(--surface-2)] hover:text-[var(--text)] transition-all duration-150"
          @click="app.setTheme(app.theme === 'dark' ? 'light' : 'dark')"
        >
          {{ app.theme === 'dark' ? t('common.lightMode') : t('common.darkMode') }}
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 border-b border-[var(--border)] bg-[var(--surface)] px-6 flex items-center justify-between shrink-0">
        <div class="md:hidden text-base font-semibold tracking-tight">{{ t('common.appName') }}</div>
        <div class="flex items-center gap-3 ml-auto">
          <!-- Mobile language switcher -->
          <div class="md:hidden flex items-center gap-1 mr-2">
            <button
              v-for="loc in app.availableLocales"
              :key="loc.code"
              type="button"
              class="text-xs px-1.5 py-0.5 rounded transition-colors"
              :class="app.locale === loc.code ? 'bg-[var(--surface-2)] text-[var(--text)] font-medium' : 'text-[var(--text-secondary)]'"
              @click="app.setLocale(loc.code)"
            >
              {{ loc.code.toUpperCase() }}
            </button>
          </div>
          <span class="text-sm text-[var(--text-secondary)] hidden sm:inline">{{ auth.user?.email }}</span>
          <button class="if-btn-ghost text-sm" @click="logout">{{ t('common.signOut') }}</button>
        </div>
      </header>

      <main class="flex-1 overflow-auto">
        <div class="max-w-6xl mx-auto px-6 py-8">
          <RouterView v-slot="{ Component }">
            <Transition name="route-fade" mode="out-in">
              <component :is="Component" />
            </Transition>
          </RouterView>
        </div>
      </main>
    </div>
  </div>
</template>

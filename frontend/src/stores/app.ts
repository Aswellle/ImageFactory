import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getLocale, setLocale, availableLocales } from '@/i18n'

type LocaleCode = 'en' | 'zh'

// App store holds lightweight UI state (sidebar, theme). It does NOT hold
// secrets; per the security rules API keys/tokens never live in component
// state beyond the auth store.
export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const theme = ref<'light' | 'dark'>(
    (localStorage.getItem('imageforge_theme') as 'light' | 'dark') || 'light',
  )

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setTheme(next: 'light' | 'dark') {
    theme.value = next
    localStorage.setItem('imageforge_theme', next)
    document.documentElement.classList.toggle('dark', next === 'dark')
  }

  const locale = computed<LocaleCode>(() => getLocale())

  function initTheme() {
    const saved = (localStorage.getItem('imageforge_theme') as 'light' | 'dark') || null
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    setTheme(saved ?? (prefersDark ? 'dark' : 'light'))
  }

  return { sidebarCollapsed, theme, toggleSidebar, setTheme, initTheme, locale, setLocale, availableLocales }
})

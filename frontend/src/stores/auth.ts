import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { getToken, setToken, parseApiError } from '@/api/client'
import type { AuthResponse, User } from '@/types'
import { useToastStore } from '@/stores/toast'
import { useI18n } from 'vue-i18n'

// Auth store owns the current user + JWT. Token persists in localStorage so a
// refresh keeps the session. The plaintext token never leaves the client; it is
// attached to requests by the axios interceptor.
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function persist(res: AuthResponse) {
    token.value = res.token
    user.value = res.user
    setToken(res.token)
  }

  async function register(email: string, password: string, name: string) {
    loading.value = true
    error.value = null
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      persist(await authApi.register(email, password, name))
      toast.success(t('auth.welcomeTitle', { name: user.value?.name || email }))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('auth.registerFailed'), error.value)
      throw e
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      persist(await authApi.login(email, password))
      toast.success(t('auth.welcomeBack', { name: user.value?.name || email }))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('auth.loginFailed'), error.value)
      throw e
    } finally {
      loading.value = false
    }
  }

  function logout() {
    token.value = null
    user.value = null
    setToken(null)
  }

  return { token, user, loading, error, isAuthenticated, isAdmin, register, login, logout }
})

import axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import type { ApiError } from '@/types'

// API_BASE is proxied to the backend in vite.config.ts during dev.
const BASE_URL = import.meta.env.IF_API_BASE || '/v1'

const TOKEN_KEY = 'imageforge_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null): void {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

function authHeader(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const token = getToken()
  if (token) config.headers.set('Authorization', `Bearer ${token}`)
  config.headers.set('X-Request-ID', crypto.randomUUID())
  return config
}

// api is the shared axios instance.
export const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 30_000,
  withCredentials: false,
})

api.interceptors.request.use(authHeader)

// Response interceptor: surface auth failures and network errors as toasts.
// This runs outside of any component, so we lazy-import the toast store to
// avoid a Pinia circular-dependency at module-load time.
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (!axios.isAxiosError(error)) return Promise.reject(error)

    const status = error.response?.status
    const data = error.response?.data as ApiError | undefined

// Lazy import avoids Pinia initialization order issues.
void Promise.all([import('@/stores/toast'), import('@/i18n/index')]).then(([{ useToastStore }, i18nMod]) => {
  const toast = useToastStore()
  const t = i18nMod?.i18n?.global?.t ?? ((k: string) => k)

  if (status === 401) {
    setToken(null)
    toast.warning(t('toast.sessionExpired') as string, undefined, 5000)
  } else if (status === 403) {
    toast.error((data?.error?.message || t('toast.somethingWrong')) as string)
  } else if (!error.response) {
    toast.error(t('toast.networkError') as string, undefined, 6000)
  }
})

    return Promise.reject(error)
  },
)

// parseApiError extracts a stable ImageForge error from any upstream failure.
export function parseApiError(err: unknown): { code: string; message: string; requestID?: string } {
  const axiosErr = err as AxiosError | undefined
  const data = axiosErr?.response?.data as ApiError | undefined
  if (data?.error) {
    return { code: data.error.code, message: data.error.message, requestID: data.error.request_id }
  }
  return { code: 'NETWORK_ERROR', message: axiosErr?.message || (err instanceof Error ? err.message : 'Network error') }
}

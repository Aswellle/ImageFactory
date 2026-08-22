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

// parseApiError extracts a stable ImageForge error from any upstream failure.
export function parseApiError(err: unknown): { code: string; message: string; requestID?: string } {
  const axiosErr = err as AxiosError | undefined
  const data = axiosErr?.response?.data as ApiError | undefined
  if (data?.error) {
    return { code: data.error.code, message: data.error.message, requestID: data.error.request_id }
  }
  return { code: 'NETWORK_ERROR', message: axiosErr?.message || (err instanceof Error ? err.message : 'Network error') }
}

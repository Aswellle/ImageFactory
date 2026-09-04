import { api } from './client'
import type { AuthResponse } from '@/types'

export const authApi = {
  register(email: string, password: string, name: string): Promise<AuthResponse> {
    return api.post('/auth/register', { email, password, name }).then((r) => r.data.data)
  },
  login(email: string, password: string): Promise<AuthResponse> {
    return api.post('/auth/login', { email, password }).then((r) => r.data.data)
  },
  sendResetCode(email: string): Promise<void> {
    return api.post('/auth/send-reset-code', { email }).then((r) => r.data.data)
  },
  resetPassword(code: string, newPassword: string): Promise<void> {
    return api.post('/auth/reset-password', { code, newPassword }).then((r) => r.data.data)
  },
}

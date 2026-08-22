import { api } from './client'
import type { AuthResponse } from '@/types'

export const authApi = {
  register(email: string, password: string, name: string): Promise<AuthResponse> {
    return api.post('/auth/register', { email, password, name }).then((r) => r.data.data)
  },
  login(email: string, password: string): Promise<AuthResponse> {
    return api.post('/auth/login', { email, password }).then((r) => r.data.data)
  },
}

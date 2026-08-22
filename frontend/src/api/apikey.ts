import { api } from './client'

export interface APIKey {
  id: number
  name: string
  prefix: string
  status: string
  last_used?: string
  expires_at?: string
  created_at: string
}

export interface CreateAPIKeyResponse {
  id: number
  name: string
  prefix: string
  key: string // ONLY returned once at creation
  status: string
  created_at: string
}

export interface UsageStats {
  total_requests: number
  total_images: number
  total_tokens: number
  period_start: string
  period_end: string
}

export const apikeyApi = {
  create(name: string): Promise<CreateAPIKeyResponse> {
    return api.post('/api-keys', { name }).then((r) => r.data.data)
  },
  list(): Promise<APIKey[]> {
    return api.get('/api-keys').then((r) => r.data.data)
  },
  revoke(id: number): Promise<void> {
    return api.delete(`/api-keys/${id}`).then(() => undefined)
  },
  usage(): Promise<UsageStats> {
    return api.get('/usage').then((r) => r.data.data)
  },
}

import { api } from './client'
import type { Paginated } from '@/types'

// Admin API client. All calls require an admin-authenticated session; the backend
// enforces this with the adminAuth middleware (JWT admin role OR x-api-key
// header matching cfg.Auth.AdminPanelKey). Paths live under /v1/admin.
//
// Contract note: mutating actions (user status, job retry, key revoke) use POST
// and return 204 NoContent; lists and the dashboard use the { data } /
// { data, pagination } envelope.

export interface AdminUser {
  id: number
  email: string
  name: string
  role: 'user' | 'admin'
  status: 'active' | 'suspended' | 'deleted'
  last_login_at?: string
  created_at: string
}

export interface AdminJob {
  id: string
  user_id: number
  username?: string
  type: 'generation' | 'edit'
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
  provider?: string
  model?: string
  prompt?: string
  image_count?: number
  retry_count?: number
  error_code?: string
  error_message?: string
  created_at: string
  completed_at?: string
}

export interface AdminApiKey {
  id: number
  user_id: number
  username?: string
  key_prefix: string
  name: string
  status: string
  created_at: string
  last_used_at?: string
}

export interface DashboardStats {
  total_users: number
  active_users: number
  total_jobs: number
  completed_jobs: number
  failed_jobs: number
  total_assets: number
  total_api_keys: number
  active_api_keys: number
  jobs_last_24h: number
  users_last_24h: number
  assets_generated_24h: number
}

export interface ActivityEvent {
  type: string
  job_id?: string
  user_id?: number
  status?: string
  prompt?: string
  model?: string
  provider?: string
  error?: string
  created_at: string
}

export interface DashboardHealth {
  database: 'ok' | 'down'
  timestamp: string
  checks: { database: string }
}

export interface ListParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
  role?: string
}

export interface UserListParams extends ListParams {
  role?: string
}

export interface JobListParams extends ListParams {
  type?: string
}

export const adminApi = {
  // --- Dashboard (split into stats / activity / health) ---
  dashboardStats(): Promise<DashboardStats> {
    return api.get('/admin/dashboard/stats').then((r) => r.data.data)
  },
  dashboardActivity(limit = 20): Promise<ActivityEvent[]> {
    return api.get('/admin/dashboard/activity', { params: { limit } }).then((r) => r.data.data)
  },
  dashboardHealth(): Promise<DashboardHealth> {
    return api.get('/admin/dashboard/health').then((r) => r.data.data)
  },

  // --- Users ---
  listUsers(params?: UserListParams): Promise<Paginated<AdminUser>> {
    return api.get('/admin/users', { params }).then((r) => r.data)
  },
  setUserStatus(id: number, status: string): Promise<void> {
    return api.post(`/admin/users/${id}/status`, { status }).then(() => undefined)
  },

  // --- Jobs ---
  listJobs(params?: JobListParams): Promise<Paginated<AdminJob>> {
    return api.get('/admin/jobs', { params }).then((r) => r.data)
  },
  retryJob(id: string): Promise<void> {
    return api.post(`/admin/jobs/${id}/retry`).then(() => undefined)
  },

  // --- API Keys ---
  listApiKeys(params?: ListParams): Promise<Paginated<AdminApiKey>> {
    return api.get('/admin/api-keys', { params }).then((r) => r.data)
  },
  revokeApiKey(id: number): Promise<void> {
    return api.post(`/admin/api-keys/${id}/revoke`).then(() => undefined)
  },
}

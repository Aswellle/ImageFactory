import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminApi } from '@/api/admin'
import { parseApiError } from '@/api/client'
import type {
  ActivityEvent,
  AdminApiKey,
  AdminJob,
  AdminUser,
  DashboardHealth,
  DashboardStats,
} from '@/api/admin'

// Admin store owns the admin panel's server-backed state: dashboard stats +
// activity feed plus the user/job/api-key list pages. Each view resets the
// relevant slice on mount so stale data never leaks between views.
export const useAdminStore = defineStore('admin', () => {
  // --- Dashboard ---
  const stats = ref<DashboardStats | null>(null)
  const activity = ref<ActivityEvent[]>([])
  const health = ref<DashboardHealth | null>(null)
  const dashboardLoading = ref(false)

  // --- Users ---
  const users = ref<AdminUser[]>([])
  const usersTotal = ref(0)
  const usersPage = ref(1)
  const usersPageSize = ref(20)
  const usersLoading = ref(false)

  // --- Jobs ---
  const jobs = ref<AdminJob[]>([])
  const jobsTotal = ref(0)
  const jobsPage = ref(1)
  const jobsPageSize = ref(20)
  const jobsLoading = ref(false)

  // --- API Keys ---
  const apiKeys = ref<AdminApiKey[]>([])
  const apiKeysTotal = ref(0)
  const apiKeysPage = ref(1)
  const apiKeysPageSize = ref(20)
  const apiKeysLoading = ref(false)

  const error = ref<string | null>(null)

  function fail(e: unknown): string {
    const msg = parseApiError(e).message
    error.value = msg
    return msg
  }

  async function loadDashboard() {
    dashboardLoading.value = true
    error.value = null
    try {
      const [s, a, h] = await Promise.all([
        adminApi.dashboardStats(),
        adminApi.dashboardActivity(),
        adminApi.dashboardHealth(),
      ])
      stats.value = s
      activity.value = a
      health.value = h
    } catch (e) {
      fail(e)
    } finally {
      dashboardLoading.value = false
    }
  }

  async function loadUsers(params?: {
    page?: number
    page_size?: number
    search?: string
    role?: string
    status?: string
  }) {
    usersLoading.value = true
    error.value = null
    try {
      if (params?.page) usersPage.value = params.page
      if (params?.page_size) usersPageSize.value = params.page_size
      const res = await adminApi.listUsers({
        page: usersPage.value,
        page_size: usersPageSize.value,
        search: params?.search,
        role: params?.role,
        status: params?.status,
      })
      users.value = res.data
      usersTotal.value = res.pagination.total
    } catch (e) {
      fail(e)
    } finally {
      usersLoading.value = false
    }
  }

  async function setUserStatus(id: number, status: 'active' | 'suspended') {
    try {
      await adminApi.setUserStatus(id, status)
      const idx = users.value.findIndex((u) => u.id === id)
      if (idx >= 0) users.value[idx] = { ...users.value[idx], status }
    } catch (e) {
      throw new Error(fail(e))
    }
  }

  async function loadJobs(params?: {
    page?: number
    page_size?: number
    status?: string
    type?: string
    search?: string
  }) {
    jobsLoading.value = true
    error.value = null
    try {
      if (params?.page) jobsPage.value = params.page
      if (params?.page_size) jobsPageSize.value = params.page_size
      const res = await adminApi.listJobs({
        page: jobsPage.value,
        page_size: jobsPageSize.value,
        status: params?.status,
        type: params?.type,
        search: params?.search,
      })
      jobs.value = res.data
      jobsTotal.value = res.pagination.total
    } catch (e) {
      fail(e)
    } finally {
      jobsLoading.value = false
    }
  }

  async function retryJob(id: string) {
    try {
      await adminApi.retryJob(id)
    } catch (e) {
      throw new Error(fail(e))
    }
  }

  async function loadApiKeys(params?: {
    page?: number
    page_size?: number
    status?: string
    search?: string
  }) {
    apiKeysLoading.value = true
    error.value = null
    try {
      if (params?.page) apiKeysPage.value = params.page
      if (params?.page_size) apiKeysPageSize.value = params.page_size
      const res = await adminApi.listApiKeys({
        page: apiKeysPage.value,
        page_size: apiKeysPageSize.value,
        status: params?.status,
        search: params?.search,
      })
      apiKeys.value = res.data
      apiKeysTotal.value = res.pagination.total
    } catch (e) {
      fail(e)
    } finally {
      apiKeysLoading.value = false
    }
  }

  async function revokeApiKey(id: number) {
    try {
      await adminApi.revokeApiKey(id)
      apiKeys.value = apiKeys.value.filter((k) => k.id !== id)
      apiKeysTotal.value = Math.max(0, apiKeysTotal.value - 1)
    } catch (e) {
      throw new Error(fail(e))
    }
  }

  function reset() {
    error.value = null
  }

  return {
    // dashboard
    stats,
    activity,
    health,
    dashboardLoading,
    loadDashboard,
    // users
    users,
    usersTotal,
    usersPage,
    usersPageSize,
    usersLoading,
    loadUsers,
    setUserStatus,
    // jobs
    jobs,
    jobsTotal,
    jobsPage,
    jobsPageSize,
    jobsLoading,
    loadJobs,
    retryJob,
    // api keys
    apiKeys,
    apiKeysTotal,
    apiKeysPage,
    apiKeysPageSize,
    apiKeysLoading,
    loadApiKeys,
    revokeApiKey,
    // shared
    error,
    reset,
  }
})

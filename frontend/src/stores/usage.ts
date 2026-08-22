import { defineStore } from 'pinia'
import { ref } from 'vue'
import { usageApi } from '@/api/usage'
import { parseApiError } from '@/api/client'
import type { UsageStat, UsageSummary } from '@/types'

// useUsageStore loads and caches the authenticated user's usage dashboard
// data. The summary covers the trailing 30-day period; history is a per-day
// breakdown used by the chart.
export const useUsageStore = defineStore('usage', () => {
  const summary = ref<UsageSummary | null>(null)
  const history = ref<UsageStat[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchSummary() {
    loading.value = true
    error.value = null
    try {
      summary.value = await usageApi.get()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function fetchHistory(days = 30) {
    try {
      history.value = await usageApi.history(days)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  // Load everything the dashboard needs.
  async function init(days = 30) {
    await Promise.all([fetchSummary(), fetchHistory(days)])
  }

  function reset() {
    summary.value = null
    history.value = []
    error.value = null
  }

  return { summary, history, loading, error, fetchSummary, fetchHistory, init, reset }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { generationApi, type GenerateRequest } from '@/api/generation'
import { parseApiError } from '@/api/client'
import type { GenerationJob } from '@/types'

// Generation store manages the user's image-generation jobs: submitting new
// requests, polling status, and surfacing results. The actual work happens
// server-side (ImageForge → Sub2API); this store only tracks job state.
export const useGenerationStore = defineStore('generation', () => {
  const jobs = ref<GenerationJob[]>([])
  const activeJobId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeJob = computed(() =>
    jobs.value.find((j) => j.job_id === activeJobId.value) ?? null,
  )

  const isProcessing = computed(() =>
    activeJob.value?.status === 'processing' || activeJob.value?.status === 'pending',
  )

  async function generate(req: GenerateRequest) {
    loading.value = true
    error.value = null
    try {
      const res = await generationApi.create(req)
      activeJobId.value = res.job_id
      await fetchActive(res.job_id)
      await refreshList()
      return res.job_id
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchActive(jobId: string) {
    try {
      const job = await generationApi.get(jobId)
      upsert(job)
      activeJobId.value = job.job_id
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function refreshList() {
    try {
      const list = await generationApi.list()
      jobs.value = list
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  // Poll the active job until it reaches a terminal state. Returns when the
  // job completes, fails, or the max attempts are exhausted.
  // Check cancelled.value to stop polling early (e.g., when user navigates away).
  const cancelled = ref(false)

  async function pollUntilDone(jobId: string, maxAttempts = 60) {
    cancelled.value = false
    try {
      for (let i = 0; i < maxAttempts; i++) {
        if (cancelled.value) return null
        const job = await generationApi.get(jobId)
        upsert(job)
        if (job.status === 'completed' || job.status === 'failed') {
          return job
        }
        await new Promise((r) => setTimeout(r, 3000))
      }
      return generationApi.get(jobId).then(upsert)
    } finally {
      cancelled.value = false
    }
  }

  function cancelPolling() {
    cancelled.value = true
  }

  function upsert(job: GenerationJob) {
    const idx = jobs.value.findIndex((j) => j.job_id === job.job_id)
    if (idx >= 0) jobs.value[idx] = job
    else jobs.value.unshift(job)
  }

  function reset() {
    error.value = null
  }

  return {
    jobs,
    activeJobId,
    activeJob,
    loading,
    error,
    isProcessing,
    generate,
    fetchActive,
    refreshList,
    pollUntilDone,
    cancelPolling,
    reset,
  }
})

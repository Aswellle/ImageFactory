import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectApi, type Project } from '@/api/project'
import { parseApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useI18n } from 'vue-i18n'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchProjects() {
    loading.value = true
    error.value = null
    try {
      projects.value = await projectApi.list()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function createProject(name: string, description?: string) {
    const toast = useToastStore()
    const { t } = useI18n()
    loading.value = true
    error.value = null
    try {
      const p = await projectApi.create({ name, description })
      projects.value.unshift(p)
      toast.success(t('toast.projectCreated'))
      return p
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.projectCreateFailed'), error.value)
      throw e
    } finally {
      loading.value = false
    }
  }

  return { projects, loading, error, fetchProjects, createProject }
})

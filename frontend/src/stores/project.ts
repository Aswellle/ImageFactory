import { defineStore } from 'pinia'
import { ref } from 'vue'
import { projectApi, type Project } from '@/api/project'
import { parseApiError } from '@/api/client'

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
    loading.value = true
    error.value = null
    try {
      const p = await projectApi.create({ name, description })
      projects.value.unshift(p)
      return p
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    } finally {
      loading.value = false
    }
  }

  return { projects, loading, error, fetchProjects, createProject }
})

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { promptTemplateApi, type BuiltInTemplate } from '@/api/promptTemplate'
import type { PromptTemplate } from '@/types'
import { parseApiError } from '@/api/client'

export const usePromptTemplateStore = defineStore('promptTemplate', () => {
  const templates = ref<PromptTemplate[]>([])
  const builtIn = ref<BuiltInTemplate[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Templates grouped by their category for display.
  const byCategory = computed(() => {
    const groups: Record<string, PromptTemplate[]> = {}
    for (const t of templates.value) {
      const key = t.category || 'custom'
      ;(groups[key] ??= []).push(t)
    }
    return groups
  })

  const categories = computed(() => Object.keys(byCategory.value))

  async function fetchTemplates() {
    loading.value = true
    error.value = null
    try {
      templates.value = await promptTemplateApi.list()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function fetchBuiltIn() {
    try {
      builtIn.value = await promptTemplateApi.listBuiltIn()
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function createTemplate(input: {
    name: string
    description?: string
    content: string
    variables?: string[]
    category?: string
  }) {
    try {
      const t = await promptTemplateApi.create(input)
      templates.value.unshift(t)
      return t
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  async function updateTemplate(
    id: number,
    input: {
      name?: string
      description?: string
      content?: string
      variables?: string[]
      category?: string
      clear_variables?: boolean
    },
  ) {
    try {
      const t = await promptTemplateApi.update(id, input)
      const idx = templates.value.findIndex((x) => x.id === id)
      if (idx >= 0) templates.value[idx] = t
      return t
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  async function deleteTemplate(id: number) {
    try {
      await promptTemplateApi.delete(id)
      templates.value = templates.value.filter((t) => t.id !== id)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }


  // Apply variables client-side (used for built-in templates not stored in the DB).
  function applyVariables(content: string, variables: Record<string,string>): string {
    let out = content
    for (const [k, v] of Object.entries(variables)) {
      out = out.split(`{{${k}}}`).join(v)
    }
    return out
  }


  // Apply a template's variables and return the rendered prompt.
  async function applyTemplate(id: number, variables: Record<string, string>) {
    try {
      const res = await promptTemplateApi.apply(id, variables)
      return res.prompt
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  // Parse a comma-separated variables string into a clean list.
  function parseVariables(variables?: string): string[] {
    if (!variables) return []
    return variables
      .split(',')
      .map((v) => v.trim())
      .filter(Boolean)
  }

  return {
    templates,
    builtIn,
    loading,
    error,
    byCategory,
    categories,
    fetchTemplates,
    fetchBuiltIn,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    applyTemplate,
    applyVariables,
    parseVariables,
  }
})

import { api } from './client'
import type { PromptTemplate } from '@/types'

// BuiltInTemplate mirrors the backend's system-provided templates.
export interface BuiltInTemplate {
  id: number
  name: string
  description: string
  content: string
  variables: string[]
  category: string
  examples?: Record<string, Record<string, string>>
}

// ApplyResult is the response from filling a template's variables.
export interface ApplyResult {
  prompt: string
}

export const promptTemplateApi = {
  list(): Promise<PromptTemplate[]> {
    return api.get('/prompt-templates').then((r) => r.data.data)
  },
  get(id: number): Promise<PromptTemplate> {
    return api.get(`/prompt-templates/${id}`).then((r) => r.data.data)
  },
  create(input: {
    name: string
    description?: string
    content: string
    variables?: string[]
    category?: string
  }): Promise<PromptTemplate> {
    return api.post('/prompt-templates', input).then((r) => r.data.data)
  },
  update(
    id: number,
    input: {
      name?: string
      description?: string
      content?: string
      variables?: string[]
      category?: string
      clear_variables?: boolean
    },
  ): Promise<PromptTemplate> {
    return api.put(`/prompt-templates/${id}`, input).then((r) => r.data.data)
  },
  delete(id: number): Promise<void> {
    return api.delete(`/prompt-templates/${id}`).then(() => undefined)
  },
  apply(id: number, variables: Record<string, string>): Promise<ApplyResult> {
    return api.post(`/prompt-templates/${id}/apply`, { variables }).then((r) => r.data.data)
  },
  listBuiltIn(): Promise<BuiltInTemplate[]> {
    return api.get('/prompt-templates?built_in=true').then((r) => r.data.data)
  },
}

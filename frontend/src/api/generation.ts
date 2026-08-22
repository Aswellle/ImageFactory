import { api } from './client'
import type { GenerationJob } from '@/types'

export interface GenerateRequest {
  prompt: string
  negative_prompt?: string
  model?: string
  size?: string
  quality?: string
  style?: string
  n?: number
  response_format?: string
  project_id?: number
}

export interface GenerateResponse {
  job_id: string
  status: string
  sub2api_task_id?: string
  created_at: string
}

export const generationApi = {
  create(req: GenerateRequest): Promise<GenerateResponse> {
    return api.post('/images/generations', req).then((r) => r.data.data)
  },
  get(jobId: string): Promise<GenerationJob> {
    return api.get(`/images/jobs/${jobId}`).then((r) => r.data.data)
  },
  list(): Promise<GenerationJob[]> {
    return api.get('/images/jobs').then((r) => r.data.data)
  },
}

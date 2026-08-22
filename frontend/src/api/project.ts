import { api } from './client'

export interface Project {
  id: number
  name: string
  description: string
  status: string
  created_at: string
}

export interface CreateProjectRequest {
  name: string
  description?: string
}

export const projectApi = {
  create(req: CreateProjectRequest): Promise<Project> {
    return api.post('/projects', req).then((r) => r.data.data)
  },
  list(): Promise<Project[]> {
    return api.get('/projects').then((r) => r.data.data)
  },
  get(id: string): Promise<Project> {
    return api.get(`/projects/${id}`).then((r) => r.data.data)
  },
}

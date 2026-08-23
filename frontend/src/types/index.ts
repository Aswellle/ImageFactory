export interface User {
  id: number
  email: string
  name: string
  role: 'user' | 'admin'
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: {
    code: string
    message: string
    request_id?: string
  }
}

export interface PageParams {
  page?: number
  pageSize?: number
}

export interface Paginated<T> {
  data: T[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

// ---- Generation ----
export type JobType = 'generation' | 'edit'
export type JobStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
export type AssetSource = 'generated' | 'uploaded' | 'edited'

export interface Asset {
  id: number
  title?: string
  description?: string
  source: AssetSource
  prompt?: string
  negative_prompt?: string
  model?: string
  model_provider?: string
  width?: number
  height?: number
  aspect_ratio?: string
  mime_type?: string
  file_size?: number
  storage_key: string
  thumbnail_key?: string
  medium_key?: string
  current_version: number
  created_at: string
  updated_at: string
}

export interface Project {
  id: number
  name: string
  description?: string
  status: 'active' | 'archived'
  created_at: string
}

export interface PromptTemplate {
  id: number
  name: string
  description?: string
  content: string
  variables?: string
  category: string
}

export interface ModelConfig {
  id: string
  display_name: string
  provider: string
  capabilities: string[]
  supported_sizes: string[]
  enabled: boolean
}


export interface GenerationJob {
  job_id: string
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
  prompt?: string
  model?: string
  image_count?: number
  sub2api_task_id?: string
  error_code?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  created_at: string
}

// ---- Usage ----

export interface UsageStat {
  date: string
  requests: number
  images: number
  tokens: number
  cost: number
}

export interface UsageSummary {
  total_requests: number
  total_images: number
  total_tokens: number
  total_cost: number
  period_start: string
  period_end: string
  daily: UsageStat[]
}

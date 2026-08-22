import { api } from './client'

export interface Asset {
  id: number
  title?: string
  description?: string
  source: string
  status: string
  prompt?: string
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

export interface AssetListResponse {
  data: Asset[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface AssetListParams {
  project_id?: string
  tag?: string
  page?: number
  page_size?: number
}

export const assetApi = {
  list(params?: AssetListParams): Promise<AssetListResponse> {
    return api.get('/assets', { params }).then((r) => r.data)
  },
  get(id: string): Promise<Asset> {
    return api.get(`/assets/${id}`).then((r) => r.data.data)
  },
  delete(id: string): Promise<void> {
    return api.delete(`/assets/${id}`).then(() => undefined)
  },
}

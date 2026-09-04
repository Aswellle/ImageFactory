import { api } from './client'
import type { Asset } from '@/types'

export type { Asset }

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

import { api } from './client'

export interface Collection {
  id: number
  user_id: number
  name: string
  description?: string
  created_at: string
}

export interface CollectionListResponse {
  data: Collection[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export const collectionApi = {
  list(): Promise<Collection[]> {
    return api.get('/collections').then((r) => r.data.data)
  },
  get(id: number): Promise<Collection> {
    return api.get(`/collections/${id}`).then((r) => r.data.data)
  },
  create(name: string, description?: string): Promise<Collection> {
    return api.post('/collections', { name, description }).then((r) => r.data.data)
  },
  update(id: number, name: string, description?: string): Promise<Collection> {
    return api.put(`/collections/${id}`, { name, description }).then((r) => r.data.data)
  },
  delete(id: number): Promise<void> {
    return api.delete(`/collections/${id}`).then(() => undefined)
  },
  addAsset(collectionId: number, assetId: number): Promise<void> {
    return api.post(`/collections/${collectionId}/assets`, { asset_id: assetId }).then(() => undefined)
  },
  removeAsset(collectionId: number, assetId: number): Promise<void> {
    return api.delete(`/collections/${collectionId}/assets/${assetId}`).then(() => undefined)
  },
  listAssets(collectionId: number): Promise<unknown[]> {
    return api.get(`/collections/${collectionId}/assets`).then((r) => r.data.data)
  },
}

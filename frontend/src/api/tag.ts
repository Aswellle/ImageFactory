import { api } from './client'

export interface Tag {
  id: number
  user_id: number
  name: string
  color?: string
  created_at: string
}

export const tagApi = {
  list(): Promise<Tag[]> {
    return api.get('/tags').then((r) => r.data.data)
  },
  get(id: number): Promise<Tag> {
    return api.get(`/tags/${id}`).then((r) => r.data.data)
  },
  create(name: string, color?: string): Promise<Tag> {
    return api.post('/tags', { name, color }).then((r) => r.data.data)
  },
  update(id: number, name: string, color?: string): Promise<Tag> {
    return api.put(`/tags/${id}`, { name, color }).then((r) => r.data.data)
  },
  delete(id: number): Promise<void> {
    return api.delete(`/tags/${id}`).then(() => undefined)
  },
  tagAsset(tagId: number, assetId: number): Promise<void> {
    return api.post(`/tags/${tagId}/assets`, { asset_id: assetId }).then(() => undefined)
  },
  untagAsset(tagId: number, assetId: number): Promise<void> {
    return api.delete(`/tags/${tagId}/assets/${assetId}`).then(() => undefined)
  },
  listAssets(tagId: number): Promise<unknown[]> {
    return api.get(`/tags/${tagId}/assets`).then((r) => r.data.data)
  },
  listAssetTags(assetId: number): Promise<Tag[]> {
    return api.get(`/assets/${assetId}/tags`).then((r) => r.data.data)
  },
}

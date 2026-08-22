import { api } from './client'

export interface Favorite {
  id: number
  user_id: number
  asset_id: number
  created_at: string
}

export const favoriteApi = {
  add(assetId: number): Promise<Favorite> {
    return api.get('/favorites', { params: { asset_id: assetId } }).then((r) => r.data)
  },
  list(): Promise<Favorite[]> {
    return api.get('/favorites').then((r) => r.data.data)
  },
  remove(assetId: number): Promise<void> {
    return api.delete(`/favorites/${assetId}`).then(() => undefined)
  },
  check(assetId: number): Promise<boolean> {
    return api.get('/favorites/check', { params: { asset_id: assetId } }).then((r) => r.data.data)
  },
}

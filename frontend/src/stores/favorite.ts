import { defineStore } from 'pinia'
import { ref } from 'vue'
import { favoriteApi, type Favorite } from '@/api/favorite'
import { parseApiError } from '@/api/client'

export const useFavoriteStore = defineStore('favorite', () => {
  const favorites = ref<Favorite[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchFavorites() {
    loading.value = true
    error.value = null
    try {
      favorites.value = await favoriteApi.list()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function addFavorite(assetId: number) {
    try {
      const fav = await favoriteApi.add(assetId)
      favorites.value.unshift(fav)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function removeFavorite(assetId: number) {
    try {
      await favoriteApi.remove(assetId)
      favorites.value = favorites.value.filter((f) => f.asset_id !== assetId)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function checkFavorite(assetId: number): Promise<boolean> {
    try {
      return await favoriteApi.check(assetId)
    } catch {
      return false
    }
  }

  function isFavorited(assetId: number): boolean {
    return favorites.value.some((f) => f.asset_id === assetId)
  }

  return { favorites, loading, error, fetchFavorites, addFavorite, removeFavorite, checkFavorite, isFavorited }
})

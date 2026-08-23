import { defineStore } from 'pinia'
import { ref } from 'vue'
import { favoriteApi, type Favorite } from '@/api/favorite'
import { parseApiError } from '@/api/client'

export const useFavoriteStore = defineStore('favorite', () => {
  const favorites = ref<Favorite[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // O(1) lookup set for favorited asset IDs.
  const favoritedIds = ref<Set<number>>(new Set())

  // Track whether we've attempted to sync, to avoid redundant fetches
  // while still allowing retries after errors.
  const synced = ref(false)

  async function fetchFavorites() {
    loading.value = true
    error.value = null
    try {
      favorites.value = await favoriteApi.list()
      favoritedIds.value = new Set(favorites.value.map((f) => f.asset_id))
      synced.value = true
    } catch (e) {
      error.value = parseApiError(e).message
      // Reset synced so retry can happen
      synced.value = false
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addFavorite(assetId: number) {
    try {
      const fav = await favoriteApi.add(assetId)
      favorites.value.unshift(fav)
      favoritedIds.value.add(assetId)
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  async function removeFavorite(assetId: number) {
    try {
      await favoriteApi.remove(assetId)
      favorites.value = favorites.value.filter((f) => f.asset_id !== assetId)
      favoritedIds.value.delete(assetId)
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  // Batch sync: ensure favorites are loaded so isFavorited() works for lists.
  // Uses synced flag to avoid redundant fetches; resets on error for retry.
  async function syncFavorites() {
    if (synced.value && favorites.value.length > 0) return
    await fetchFavorites()
  }

  function isFavorited(assetId: number): boolean {
    return favoritedIds.value.has(assetId)
  }

  return { favorites, loading, error, favoritedIds, fetchFavorites, addFavorite, removeFavorite, syncFavorites, isFavorited }
})

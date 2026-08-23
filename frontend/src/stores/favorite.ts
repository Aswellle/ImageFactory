import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { favoriteApi, type Favorite } from '@/api/favorite'
import { parseApiError } from '@/api/client'

export const useFavoriteStore = defineStore('favorite', () => {
  const favorites = ref<Favorite[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // O(1) lookup set for favorited asset IDs.
  const favoritedIds = ref<Set<number>>(new Set())

  const favoritedIdsSet = computed(() => favoritedIds.value)

  async function fetchFavorites() {
    loading.value = true
    error.value = null
    try {
      favorites.value = await favoriteApi.list()
      favoritedIds.value = new Set(favorites.value.map((f) => f.asset_id))
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

  // Batch check: sync favoritedIds for the given asset IDs.
  // Batch sync: ensure favorites are loaded so isFavorited() works for lists.
  async function syncFavorites() {
    if (favorites.value.length > 0) return
    await fetchFavorites()
  }

  function isFavorited(assetId: number): boolean {
    return favoritedIds.value.has(assetId)
  }

  return { favorites, loading, error, favoritedIdsSet, fetchFavorites, addFavorite, removeFavorite, syncFavorites, isFavorited }
})

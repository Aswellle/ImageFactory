import { defineStore } from 'pinia'
import { ref } from 'vue'
import { collectionApi, type Collection } from '@/api/collection'
import { parseApiError } from '@/api/client'

export const useCollectionStore = defineStore('collection', () => {
  const collections = ref<Collection[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchCollections() {
    loading.value = true
    error.value = null
    try {
      collections.value = await collectionApi.list()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function getCollection(id: number) {
    try {
      return await collectionApi.get(id)
    } catch (e) {
      error.value = parseApiError(e).message
      return null
    }
  }

  async function createCollection(name: string, description?: string) {
    try {
      const col = await collectionApi.create(name, description)
      collections.value.unshift(col)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function updateCollection(id: number, name: string, description?: string) {
    try {
      const col = await collectionApi.update(id, name, description)
      const idx = collections.value.findIndex((c) => c.id === id)
      if (idx >= 0) collections.value[idx] = col
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function deleteCollection(id: number) {
    try {
      await collectionApi.delete(id)
      collections.value = collections.value.filter((c) => c.id !== id)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function addAssetToCollection(collectionId: number, assetId: number) {
    try {
      await collectionApi.addAsset(collectionId, assetId)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function removeAssetFromCollection(collectionId: number, assetId: number) {
    try {
      await collectionApi.removeAsset(collectionId, assetId)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function listAssets(collectionId: number) {
    try {
      return await collectionApi.listAssets(collectionId)
    } catch (e) {
      error.value = parseApiError(e).message
      return []
    }
  }

  return {
    collections, loading, error,
    fetchCollections, getCollection, createCollection, updateCollection, deleteCollection,
    addAssetToCollection, removeAssetFromCollection, listAssets,
  }
})

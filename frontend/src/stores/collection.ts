import { defineStore } from 'pinia'
import { ref } from 'vue'
import { collectionApi, type Collection } from '@/api/collection'
import { parseApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useI18n } from 'vue-i18n'

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
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      const col = await collectionApi.create(name, description)
      collections.value.unshift(col)
      toast.success(t('toast.collectionCreated'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.collectionCreateFailed'), error.value)
    }
  }

  async function updateCollection(id: number, name: string, description?: string) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      const col = await collectionApi.update(id, name, description)
      const idx = collections.value.findIndex((c) => c.id === id)
      if (idx >= 0) collections.value[idx] = col
      toast.success(t('toast.collectionUpdated'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.collectionUpdated'), error.value)
    }
  }

  async function deleteCollection(id: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await collectionApi.delete(id)
      collections.value = collections.value.filter((c) => c.id !== id)
      toast.success(t('toast.collectionDeleted'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.collectionDeleted'), error.value)
    }
  }

  async function addAssetToCollection(collectionId: number, assetId: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await collectionApi.addAsset(collectionId, assetId)
      toast.success(t('toast.assetAddedToCollection'), undefined, 2500)
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.assetAddedToCollection'), error.value)
    }
  }

  async function removeAssetFromCollection(collectionId: number, assetId: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await collectionApi.removeAsset(collectionId, assetId)
      toast.success(t('toast.assetRemovedFromCollection'), undefined, 2500)
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.assetRemovedFromCollection'), error.value)
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

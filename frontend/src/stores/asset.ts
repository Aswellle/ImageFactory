import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { assetApi, type Asset, type AssetListParams } from '@/api/asset'
import { parseApiError } from '@/api/client'

export const useAssetStore = defineStore('asset', () => {
  const assets = ref<Asset[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeAssetId = ref<number | null>(null)

  const activeAsset = computed(() =>
    assets.value.find((a) => a.id === activeAssetId.value) ?? null,
  )

  async function fetchAssets(params?: AssetListParams) {
    loading.value = true
    error.value = null
    try {
      const res = await assetApi.list(params)
      assets.value = res.data
      total.value = res.pagination.total
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function fetchAsset(id: number) {
    try {
      const a = await assetApi.get(String(id))
      const idx = assets.value.findIndex((x) => x.id === id)
      if (idx >= 0) assets.value[idx] = a
      else assets.value.unshift(a)
      activeAssetId.value = id
      return a
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  async function deleteAsset(id: number) {
    try {
      await assetApi.delete(String(id))
      assets.value = assets.value.filter((a) => a.id !== id)
    } catch (e) {
      error.value = parseApiError(e).message
      throw e
    }
  }

  return { assets, total, loading, error, activeAssetId, activeAsset, fetchAssets, fetchAsset, deleteAsset }
})

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { tagApi, type Tag } from '@/api/tag'
import { parseApiError } from '@/api/client'

export const useTagStore = defineStore('tag', () => {
  const tags = ref<Tag[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchTags() {
    loading.value = true
    error.value = null
    try {
      tags.value = await tagApi.list()
    } catch (e) {
      error.value = parseApiError(e).message
    } finally {
      loading.value = false
    }
  }

  async function createTag(name: string, color?: string) {
    try {
      const t = await tagApi.create(name, color)
      tags.value.push(t)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function updateTag(id: number, name: string, color?: string) {
    try {
      const t = await tagApi.update(id, name, color)
      const idx = tags.value.findIndex((x) => x.id === id)
      if (idx >= 0) tags.value[idx] = t
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function deleteTag(id: number) {
    try {
      await tagApi.delete(id)
      tags.value = tags.value.filter((t) => t.id !== id)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function tagAsset(tagId: number, assetId: number) {
    try {
      await tagApi.tagAsset(tagId, assetId)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function untagAsset(tagId: number, assetId: number) {
    try {
      await tagApi.untagAsset(tagId, assetId)
    } catch (e) {
      error.value = parseApiError(e).message
    }
  }

  async function fetchAssetTags(assetId: number) {
    try {
      return await tagApi.listAssetTags(assetId)
    } catch (e) {
      error.value = parseApiError(e).message
      return []
    }
  }

  return {
    tags, loading, error,
    fetchTags, createTag, updateTag, deleteTag,
    tagAsset, untagAsset, fetchAssetTags,
  }
})

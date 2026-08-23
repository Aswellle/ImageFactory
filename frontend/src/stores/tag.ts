import { defineStore } from 'pinia'
import { ref } from 'vue'
import { tagApi, type Tag } from '@/api/tag'
import { parseApiError } from '@/api/client'
import { useToastStore } from '@/stores/toast'
import { useI18n } from 'vue-i18n'

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
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      const tag = await tagApi.create(name, color)
      tags.value.push(tag)
      toast.success(t('toast.tagCreated'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.tagCreateFailed'), error.value)
    }
  }

  async function updateTag(id: number, name: string, color?: string) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      const tag = await tagApi.update(id, name, color)
      const idx = tags.value.findIndex((x) => x.id === id)
      if (idx >= 0) tags.value[idx] = tag
      toast.success(t('toast.tagUpdated'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.tagUpdated'), error.value)
    }
  }

  async function deleteTag(id: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await tagApi.delete(id)
      tags.value = tags.value.filter((x) => x.id !== id)
      toast.success(t('toast.tagDeleted'))
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.tagDeleted'), error.value)
    }
  }

  async function tagAsset(tagId: number, assetId: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await tagApi.tagAsset(tagId, assetId)
      toast.success(t('toast.tagApplied'), undefined, 2000)
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.tagApplied'), error.value)
    }
  }

  async function untagAsset(tagId: number, assetId: number) {
    const toast = useToastStore()
    const { t } = useI18n()
    try {
      await tagApi.untagAsset(tagId, assetId)
      toast.success(t('toast.tagRemoved'), undefined, 2000)
    } catch (e) {
      error.value = parseApiError(e).message
      toast.error(t('toast.tagRemoved'), error.value)
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

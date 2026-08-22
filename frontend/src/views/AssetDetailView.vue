<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { useTagStore } from '@/stores/tag'
import { useI18n } from 'vue-i18n'
import type { Asset } from '@/api/asset'
import type { Tag } from '@/api/tag'
import FavoriteButton from '@/components/FavoriteButton.vue'
import TagBadge from '@/components/TagBadge.vue'
import CollectionPicker from '@/components/CollectionPicker.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()
const tagStore = useTagStore()

const asset = ref<Asset | null>(null)
const loading = ref(true)
const assetTags = ref<Tag[]>([])
const tagInput = ref('')
const tagColor = ref('#6366f1')
const showTagForm = ref(false)

const assetId = computed(() => asset.value ? Number(asset.value.id) : Number(route.params.id))

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    asset.value = await assetStore.fetchAsset(id)
    assetTags.value = await tagStore.fetchAssetTags(id)
  } catch {
    router.push('/assets')
  } finally {
    loading.value = false
  }
})

watch(() => route.params.id, async (newId) => {
  if (!newId) return
  loading.value = true
  try {
    const id = Number(newId)
    asset.value = await assetStore.fetchAsset(id)
    assetTags.value = await tagStore.fetchAssetTags(id)
  } catch {
    router.push('/assets')
  } finally {
    loading.value = false
  }
})

const versions = computed(() => {
  return asset.value ? [asset.value] : []
})

function formatSize(bytes?: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function addTag() {
  if (!tagInput.value.trim()) return
  try {
    await tagStore.createTag(tagInput.value.trim(), tagColor.value)
    assetTags.value = await tagStore.fetchAssetTags(assetId.value)
    tagInput.value = ''
    showTagForm.value = false
  } catch {
    // error handled in store
  }
}

async function removeTag(tag: Tag) {
  try {
    await tagStore.untagAsset(tag.id, assetId.value)
    assetTags.value = assetTags.value.filter((t) => t.id !== tag.id)
  } catch {
    // error handled in store
  }
}

async function applyExistingTag(tag: Tag) {
  try {
    await tagStore.tagAsset(tag.id, assetId.value)
    if (!assetTags.value.some((t) => t.id === tag.id)) {
      assetTags.value = await tagStore.fetchAssetTags(assetId.value)
    }
  } catch {
    // error handled in store
  }
}
</script>

<template>
  <!-- Loading state -->
  <div v-if="loading" class="space-y-6 animate-fade-in">
    <div class="skeleton h-6 w-32"></div>
    <div class="skeleton aspect-[16/10] rounded-2xl"></div>
    <div class="grid grid-cols-2 gap-6">
      <div class="skeleton h-40 rounded-2xl"></div>
      <div class="skeleton h-40 rounded-2xl"></div>
    </div>
  </div>

  <!-- Content -->
  <div v-else-if="asset" class="space-y-6 animate-fade-in">
    <!-- Header -->
    <header class="flex items-start justify-between">
      <div>
        <button class="mb-3 text-sm text-text-secondary transition-colors hover:text-text" @click="$router.back()">{{ t('common.back') }}</button>
        <h1 class="text-heading text-[#1d1d1f] dark:text-white">{{ asset.title || t('common.untitled') }}</h1>
      </div>
      <FavoriteButton :asset-id="asset.id" size="lg" />
    </header>

    <!-- Image preview -->
    <section class="bento-tile overflow-hidden">
      <div class="flex min-h-[300px] items-center justify-center bg-surface-2">
        <img
          v-if="asset.storage_key"
          :src="`/v1/assets/${asset.id}/content`"
          class="max-h-[70vh] max-w-full object-contain img-loading"
          @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
        />
        <div v-else class="p-8 text-center text-sm text-text-secondary">
          {{ asset.prompt?.slice(0, 200) || t('asset.noPreviewAvailable') }}
        </div>
      </div>
    </section>

    <!-- Tags & Collections -->
    <section class="bento-tile space-y-4">
      <h3 class="text-sm font-medium">{{ t('asset.tags') }}</h3>
      <div class="flex flex-wrap gap-2">
        <TagBadge
          v-for="tag in assetTags"
          :key="tag.id"
          :tag="tag"
          @click="removeTag(tag)"
        />
        <button
          type="button"
          class="badge border-dashed text-text-secondary transition-colors hover:text-text hover:bg-surface-2"
          @click="showTagForm = !showTagForm"
        >
          {{ t('asset.addTag') }}
        </button>
      </div>

      <!-- Tag form -->
      <div v-if="showTagForm" class="space-y-3 rounded-xl border border-border p-4 animate-scale-in">
        <div class="flex gap-2">
          <input
            v-model="tagInput"
            class="input flex-1"
            :placeholder="t('asset.tagNamePlaceholder')"
            @keyup.enter="addTag"
          />
          <input
            v-model="tagColor"
            type="color"
            class="h-9 w-9 cursor-pointer rounded-lg border-0 bg-transparent"
          />
          <button
            type="button"
            class="btn-primary"
            @click="addTag"
          >
            {{ t('common.add') }}
          </button>
        </div>

        <!-- Quick-pick existing tags -->
        <div v-if="tagStore.tags.length > 0" class="space-y-2">
          <p class="text-caption">{{ t('asset.quickAddExisting') }}</p>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="tag in tagStore.tags.filter((t) => !assetTags.some((at: Tag) => at.id === t.id))"
              :key="tag.id"
              type="button"
              @click="applyExistingTag(tag)"
            >
              <TagBadge :tag="tag" clickable />
            </button>
          </div>
        </div>
      </div>

      <!-- Collection picker -->
      <div class="pt-2">
        <CollectionPicker :asset-id="asset.id" />
      </div>
    </section>

    <!-- Details -->
    <section class="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div class="bento-tile space-y-4">
        <h3 class="text-sm font-medium">{{ t('asset.details') }}</h3>
        <dl class="space-y-2 text-sm">
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.model') }}</dt>
            <dd>{{ asset.model || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.provider') }}</dt>
            <dd>{{ asset.model_provider || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.size') }}</dt>
            <dd>{{ asset.width && asset.height ? `${asset.width}×${asset.height}` : '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.aspectRatio') }}</dt>
            <dd>{{ asset.aspect_ratio || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.format') }}</dt>
            <dd>{{ asset.mime_type || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.fileSize') }}</dt>
            <dd>{{ formatSize(asset.file_size) }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">{{ t('asset.created') }}</dt>
            <dd>{{ asset.created_at }}</dd>
          </div>
        </dl>
      </div>

      <div class="bento-tile space-y-4">
        <h3 class="text-sm font-medium">{{ t('asset.prompt') }}</h3>
        <p class="whitespace-pre-wrap text-sm text-text-secondary">{{ asset.prompt || '-' }}</p>
      </div>
    </section>

    <!-- Versions -->
    <section v-if="versions.length > 0" class="bento-tile space-y-4">
      <h3 class="text-sm font-medium">{{ t('asset.versionHistory') }}</h3>
      <div class="flex gap-3">
        <div v-for="v in versions" :key="v.current_version" class="group w-24">
          <div class="aspect-square overflow-hidden rounded-lg bg-surface-2 transition-transform duration-150 group-hover:scale-[1.03]">
            <img
              v-if="v.thumbnail_key"
              :src="`/v1/assets/${v.id}/content`"
              class="h-full w-full object-cover img-loading"
              @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
            />
          </div>
          <div class="mt-1 text-center text-xs text-text-muted">v{{ v.current_version }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

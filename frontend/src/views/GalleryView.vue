<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { useFavoriteStore } from '@/stores/favorite'
import { useTagStore } from '@/stores/tag'
import { useI18n } from 'vue-i18n'
import type { Asset } from '@/api/asset'
import FavoriteButton from '@/components/FavoriteButton.vue'
import TagBadge from '@/components/TagBadge.vue'

const { t } = useI18n()
const router = useRouter()
const assetStore = useAssetStore()
const favoriteStore = useFavoriteStore()
const tagStore = useTagStore()

const viewMode = ref<'grid' | 'list'>('grid')
const searchQuery = ref('')
const onlyFavorites = ref(false)
const selectedTagId = ref<number | null>(null)

onMounted(async () => {
  assetStore.fetchAssets()
  await favoriteStore.syncFavorites()
  if (tagStore.tags.length === 0) tagStore.fetchTags()
})

const filteredAssets = computed(() => {
  let list = assetStore.assets

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(
      (a) => a.prompt?.toLowerCase().includes(q) || a.title?.toLowerCase().includes(q),
    )
  }

  if (onlyFavorites.value) {
    list = list.filter((a) => favoriteStore.isFavorited(a.id))
  }

  if (selectedTagId.value !== null) {
    // Client-side filtering for tag is not possible without tag data on asset.
    // This placeholder is for future backend support.
  }

  return list
})

function openAsset(asset: Asset) {
  router.push(`/assets/${asset.id}`)
}

function formatSize(bytes?: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<template>
  <div class="space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-heading">{{ t('gallery.title') }}</h1>
        <p class="text-caption mt-1">{{ t('gallery.imagesCount', { count: assetStore.total }) }}</p>
      </div>
      <div class="flex items-center gap-3">
        <input
          v-model="searchQuery"
          class="input w-56"
          :placeholder="t('gallery.searchPlaceholder')"
          data-search-input
        />
        <button
          type="button"
          class="btn btn-ghost"
          :class="onlyFavorites ? 'badge-accent !rounded-md' : ''"
          @click="onlyFavorites = !onlyFavorites"
        >
          <svg class="h-4 w-4" :fill="onlyFavorites ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
        </button>
        <div class="flex border border-[var(--border)] rounded-lg overflow-hidden">
          <button
            class="px-2.5 py-1.5 text-sm transition-colors"
            :class="viewMode === 'grid' ? 'bg-[var(--surface-2)] text-[var(--text)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-2)]'"
            @click="viewMode = 'grid'"
          >
            {{ t('gallery.grid') }}
          </button>
          <button
            class="px-2.5 py-1.5 text-sm border-l border-[var(--border)] transition-colors"
            :class="viewMode === 'list' ? 'bg-[var(--surface-2)] text-[var(--text)]' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-2)]'"
            @click="viewMode = 'list'"
          >
            {{ t('gallery.list') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Tag filter bar -->
    <div v-if="tagStore.tags.length > 0" class="flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="badge"
        :class="selectedTagId === null ? 'badge-accent' : ''"
        @click="selectedTagId = null"
      >
        {{ t('gallery.all') }}
      </button>
      <button
        v-for="tag in tagStore.tags"
        :key="tag.id"
        type="button"
        @click="selectedTagId = selectedTagId === tag.id ? null : tag.id"
      >
        <TagBadge :tag="tag" :clickable="true" />
      </button>
    </div>

    <!-- Loading -->
    <div v-if="assetStore.loading" class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4 stagger">
      <div v-for="n in 8" :key="n" class="break-inside-avoid mb-4">
        <div class="card overflow-hidden">
          <div class="skeleton aspect-[4/3]"></div>
          <div class="p-3 space-y-2">
            <div class="skeleton h-3 w-3/4"></div>
            <div class="skeleton h-3 w-1/2"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredAssets.length === 0">
      <div class="card animate-fade-in">
        <div class="py-20 text-center">
          <div class="text-body mb-1">
            {{ searchQuery || onlyFavorites ? t('gallery.noMatchingImages') : t('gallery.noImagesYet') }}
          </div>
          <div class="text-caption">
            {{ searchQuery || onlyFavorites ? t('gallery.tryDifferentFilter') : t('gallery.generateFirstImage') }}
          </div>
        </div>
      </div>
    </div>
    <!-- Masonry Grid -->
    <div v-else-if="viewMode === 'grid'" class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4">
      <div
        v-for="asset in filteredAssets"
        :key="asset.id"
        class="break-inside-avoid mb-4 group"
        @click="openAsset(asset)"
      >
        <div class="card card-interactive overflow-hidden break-inside-avoid">
          <div class="relative overflow-hidden">
            <div class="aspect-[4/3] bg-[var(--surface-2)] flex items-center justify-center overflow-hidden">
              <img
                v-if="asset.thumbnail_key || asset.storage_key"
                :src="`/v1/assets/${asset.id}/content`"
                class="w-full h-auto object-cover img-loading transition-transform duration-300 group-hover:scale-[1.02]"
                loading="lazy"
                @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
              />
              <div v-else class="text-caption p-4 text-center">
                {{ asset.prompt?.slice(0, 60) || t('common.noPreview') }}
              </div>
            </div>
            <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200" @click.stop>
              <FavoriteButton :asset-id="asset.id" size="sm" :initial-favorited="favoriteStore.isFavorited(asset.id)" />
            </div>
          </div>
          <div class="p-3 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
            <p class="text-xs text-[var(--text-secondary)] line-clamp-2">{{ asset.prompt || t('common.untitled') }}</p>
          </div>
        </div>
      </div>
    </div>


    <!-- List View -->
    <div v-else class="card divide-y divide-[var(--border-subtle)] overflow-hidden">
      <div
        v-for="asset in filteredAssets"
        :key="asset.id"
        class="flex items-center gap-4 px-4 py-3 hover:bg-[var(--surface-2)] cursor-pointer transition-colors duration-150"
        @click="openAsset(asset)"
      >
        <div class="w-16 h-16 bg-[var(--surface-2)] rounded-md shrink-0 overflow-hidden">
          <img
            v-if="asset.thumbnail_key || asset.storage_key"
            :src="`/v1/assets/${asset.id}/content`"
            class="w-full h-full object-cover img-loading"
            @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
          />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm truncate">{{ asset.prompt || t('common.untitled') }}</p>
          <p class="text-caption mt-0.5">{{ asset.model }} · {{ formatSize(asset.file_size) }}</p>
        </div>
        <FavoriteButton :asset-id="asset.id" size="sm" :initial-favorited="favoriteStore.isFavorited(asset.id)" @click.stop />
        <span class="text-caption">{{ asset.created_at }}</span>
      </div>
    </div>

  </div>
</template>

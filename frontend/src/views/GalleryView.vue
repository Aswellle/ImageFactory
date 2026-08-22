<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import type { Asset } from '@/api/asset'

const router = useRouter()
const store = useAssetStore()

const viewMode = ref<'grid' | 'list'>('grid')
const searchQuery = ref('')

onMounted(() => {
  store.fetchAssets()
})

const filteredAssets = computed(() => {
  if (!searchQuery.value) return store.assets
  const q = searchQuery.value.toLowerCase()
  return store.assets.filter(
    (a) => a.prompt?.toLowerCase().includes(q) || a.title?.toLowerCase().includes(q),
  )
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
        <h1 class="text-xl font-semibold tracking-tight">Gallery</h1>
        <p class="text-sm text-text-secondary mt-1">{{ store.total }} images</p>
      </div>
      <div class="flex items-center gap-3">
        <input
          v-model="searchQuery"
          class="if-input w-56"
          placeholder="Search by prompt…"
        />
        <div class="flex border border-border rounded-md overflow-hidden">
          <button
            class="px-2.5 py-1.5 text-sm transition-colors"
            :class="viewMode === 'grid' ? 'bg-surface-2 text-text' : 'text-text-secondary hover:text-text'"
            @click="viewMode = 'grid'"
          >
            Grid
          </button>
          <button
            class="px-2.5 py-1.5 text-sm border-l border-border transition-colors"
            :class="viewMode === 'list' ? 'bg-surface-2 text-text' : 'text-text-secondary hover:text-text'"
            @click="viewMode = 'list'"
          >
            List
          </button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-20 text-text-secondary text-sm">
      Loading…
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredAssets.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">
          {{ searchQuery ? 'No matching images' : 'No images yet' }}
        </div>
        <div class="text-xs text-text-secondary">
          {{ searchQuery ? 'Try a different search term' : 'Generate your first image to get started' }}
        </div>
      </div>
    </div>

    <!-- Masonry Grid -->
    <div v-else-if="viewMode === 'grid'" class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4">
      <div
        v-for="asset in filteredAssets"
        :key="asset.id"
        class="break-inside-avoid mb-4 cursor-pointer group"
        @click="openAsset(asset)"
      >
        <div class="if-card overflow-hidden">
          <div class="aspect-auto min-h-[120px] bg-surface-2 flex items-center justify-center">
            <img
              v-if="asset.thumbnail_key || asset.storage_key"
              :src="`/v1/assets/${asset.id}/content`"
              class="w-full h-auto object-cover"
              loading="lazy"
            />
            <div v-else class="text-text-secondary text-xs p-4 text-center">
              {{ asset.prompt?.slice(0, 60) || 'No preview' }}
            </div>
          </div>
          <div class="p-3 opacity-0 group-hover:opacity-100 transition-opacity">
            <p class="text-xs text-text-secondary line-clamp-2">{{ asset.prompt || 'Untitled' }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="if-card divide-y divide-border">
      <div
        v-for="asset in filteredAssets"
        :key="asset.id"
        class="flex items-center gap-4 px-4 py-3 hover:bg-surface-2 cursor-pointer transition-colors"
        @click="openAsset(asset)"
      >
        <div class="w-16 h-16 bg-surface-2 rounded shrink-0 overflow-hidden">
          <img
            v-if="asset.thumbnail_key || asset.storage_key"
            :src="`/v1/assets/${asset.id}/content`"
            class="w-full h-full object-cover"
          />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm truncate">{{ asset.prompt || 'Untitled' }}</p>
          <p class="text-xs text-text-secondary mt-0.5">{{ asset.model }} · {{ formatSize(asset.file_size) }}</p>
        </div>
        <span class="text-xs text-text-secondary">{{ asset.created_at }}</span>
      </div>
    </div>
  </div>
</template>

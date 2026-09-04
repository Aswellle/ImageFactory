<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { useFavoriteStore } from '@/stores/favorite'
import { useTagStore } from '@/stores/tag'
import type { Asset } from '@/api/asset'

const router = useRouter()
const assetStore = useAssetStore()
const favoriteStore = useFavoriteStore()
const tagStore = useTagStore()

const viewMode = ref<'grid' | 'list'>('grid')
const searchQuery = ref('')
const onlyFavorites = ref(false)
const lightboxAsset = ref<Asset | null>(null)

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

  return list
})

onMounted(async () => {
  assetStore.fetchAssets()
  await favoriteStore.syncFavorites()
  if (tagStore.tags.length === 0) tagStore.fetchTags()
})

function openAsset(asset: Asset) {
  lightboxAsset.value = asset
}

function closeLightbox() {
  lightboxAsset.value = null
}

function formatSize(bytes?: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function onImageError(event: Event) {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
}

// Vary aspect ratios for editorial feel
function getAspectRatio(index: number): string {
  const ratios = ['aspect-[4/3]', 'aspect-[3/4]', 'aspect-square', 'aspect-[4/5]', 'aspect-[16/10]']
  return ratios[index % ratios.length]
}
</script>

<template>
  <div class="studio-theme">
    <div class="gallery-studio">
      <!-- Sidebar (shared with Dashboard) -->
      <aside class="studio-sidebar">
        <div class="studio-sidebar-logo">
          <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
            <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="#0C0A09" />
          </svg>
        </div>
        <nav class="studio-sidebar-nav">
          <RouterLink to="/app" class="studio-nav-item" title="Generate">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/gallery" class="studio-nav-item active" title="Gallery">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M3 15l4-4a2 2 0 012.8 0L14 15" />
              <path d="M14 13l1-1a2 2 0 012.8 0L21 15" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/projects" class="studio-nav-item" title="Projects">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/api" class="studio-nav-item" title="API">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 18l6-6-6-6M8 6l-6 6 6 6" />
            </svg>
          </RouterLink>
        </nav>
      </aside>

      <!-- Main -->
      <div class="gallery-main">
        <!-- Top bar -->
        <header class="gallery-topbar">
          <div class="gallery-topbar-left">
            <span class="gallery-topbar-title">Gallery</span>
            <span class="gallery-topbar-count">{{ filteredAssets.length }} items</span>
          </div>
          <div class="gallery-topbar-right">
            <div class="gallery-search">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
              <input v-model="searchQuery" type="text" placeholder="Search images..." />
            </div>
            <button class="gallery-filter-chip" :class="{ active: onlyFavorites }" @click="onlyFavorites = !onlyFavorites" style="cursor:pointer">
              <svg width="10" height="10" viewBox="0 0 24 24" :fill="onlyFavorites ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2"><path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
              Favorites
            </button>
            <div class="gallery-view-toggle">
              <button class="gallery-view-btn" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
              </button>
              <button class="gallery-view-btn" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
              </button>
            </div>
          </div>
        </header>

        <!-- Gallery content -->
        <div class="gallery-content">
          <!-- Loading -->
          <div v-if="assetStore.loading" class="gallery-skeleton-grid">
            <div v-for="n in 8" :key="n" class="gallery-skeleton-item">
              <div class="gallery-skeleton-mat">
                <div class="gallery-skeleton-thumb"></div>
                <div class="gallery-skeleton-label"></div>
              </div>
            </div>
          </div>

          <!-- Empty -->
          <div v-else-if="filteredAssets.length === 0" class="gallery-empty">
            <div class="gallery-empty-icon">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <p class="gallery-empty-title">{{ searchQuery || onlyFavorites ? 'No matching images' : 'Your gallery is empty' }}</p>
            <p class="gallery-empty-desc">{{ searchQuery || onlyFavorites ? 'Try adjusting your search or filters' : 'Start creating to see your images here' }}</p>
            <button v-if="!searchQuery && !onlyFavorites" class="gallery-empty-btn" @click="router.push('/app')">
              Create your first image
            </button>
            <button v-else class="gallery-empty-btn" @click="searchQuery = ''; onlyFavorites = false">
              Clear filters
            </button>
          </div>

          <!-- Grid with Print Frames -->
          <div v-else class="gallery-grid">
            <div
              v-for="(asset, idx) in filteredAssets"
              :key="asset.id"
              class="print-frame"
              @click="openAsset(asset)"
            >
              <div class="print-mat">
                <div class="print-image-container" :class="getAspectRatio(idx)">
                  <img
                    v-if="asset.thumbnail_key || asset.storage_key"
                    :src="`/v1/assets/${asset.id}/content`"
                    :alt="asset.prompt"
                    class="img-loading"
                    loading="lazy"
                    @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
                    @error="onImageError"
                  />
                  <div class="print-overlay">
                    <button class="print-overlay-btn" @click.stop="router.push(`/assets/${asset.id}`)">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M7 17L17 7M17 7H8M17 7v9"/></svg>
                    </button>
                  </div>
                </div>
              </div>
              <div class="print-label">
                <p class="print-label-text">{{ asset.prompt || 'Untitled' }}</p>
                <div class="print-label-meta">
                  <span>{{ asset.model || 'unknown' }}</span>
                  <span class="print-label-dot"></span>
                  <span>{{ formatSize(asset.file_size) }}</span>
                  <span class="print-label-dot"></span>
                  <span>{{ formatDate(asset.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Lightbox -->
      <div v-if="lightboxAsset" class="gallery-lightbox" @click.self="closeLightbox">
        <button class="gallery-lightbox-close" @click="closeLightbox">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        </button>
        <div class="gallery-lightbox-image">
          <img
            :src="`/v1/assets/${lightboxAsset.id}/content`"
            :alt="lightboxAsset.prompt"
          />
        </div>
        <div class="gallery-lightbox-info">
          <p class="gallery-lightbox-prompt">{{ lightboxAsset.prompt || 'Untitled' }}</p>
          <p class="gallery-lightbox-meta">{{ lightboxAsset.model }} · {{ lightboxAsset.width }}x{{ lightboxAsset.height }} · {{ formatSize(lightboxAsset.file_size) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { useTagStore } from '@/stores/tag'
import type { Tag } from '@/api/tag'

const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()
const tagStore = useTagStore()

interface Asset {
  id: number
  title?: string
  description?: string
  prompt?: string
  negative_prompt?: string
  model?: string
  model_provider?: string
  width?: number
  height?: number
  aspect_ratio?: string
  mime_type?: string
  file_size?: number
  storage_key: string
  thumbnail_key?: string
  current_version: number
  created_at: string
  updated_at: string
}

const asset = ref<Asset | null>(null)
const loading = ref(true)
const assetTags = ref<Tag[]>([])
const tagInput = ref('')
const tagColor = ref('#D4953A')
const showTagForm = ref(false)
const isFavorited = ref(false)

const assetId = computed(() => asset.value ? Number(asset.value.id) : Number(route.params.id))

onMounted(async () => {
  await loadAsset()
})

watch(() => route.params.id, async (newId) => {
  if (newId) await loadAsset()
})

async function loadAsset() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    asset.value = await assetStore.fetchAsset(id) as unknown as Asset
    if (tagStore.tags.length === 0) await tagStore.fetchTags()
    assetTags.value = await tagStore.fetchAssetTags(id)
  } catch {
    router.push('/app/gallery')
  } finally {
    loading.value = false
  }
}

const versions = computed(() => {
  return asset.value ? [asset.value] : []
})

function formatSize(bytes?: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
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

function downloadImage() {
  if (!asset.value) return
  const link = document.createElement('a')
  link.href = `/v1/assets/${asset.value.id}/content`
  link.download = `${asset.value.title || 'image'}.png`
  link.click()
}

function toggleFavorite() {
  isFavorited.value = !isFavorited.value
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="studio-theme">
    <div class="asset-detail-studio">
      <!-- Sidebar -->
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

      <!-- Main viewport -->
      <div class="asset-detail-viewport">
        <!-- Top bar -->
        <div class="asset-detail-topbar">
          <button class="asset-detail-back" @click="goBack">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
            Back
          </button>
          <div class="asset-detail-actions">
            <button class="asset-detail-action-btn" :class="{ active: isFavorited }" @click="toggleFavorite" title="Favorite">
              <svg width="16" height="16" viewBox="0 0 24 24" :fill="isFavorited ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2"><path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
            </button>
            <button class="asset-detail-action-btn" @click="downloadImage" title="Download">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
            </button>
          </div>
        </div>

        <!-- Stage (image viewer) -->
        <div class="asset-detail-stage">
          <!-- Loading -->
          <div v-if="loading" class="asset-detail-loading">
            <div class="generation-developing-dots">
              <span class="generation-developing-dot"></span>
              <span class="generation-developing-dot"></span>
              <span class="generation-developing-dot"></span>
            </div>
          </div>

          <!-- Image -->
          <div v-else-if="asset?.storage_key" class="asset-detail-image-wrap">
            <img
              :src="`/v1/assets/${asset.id}/content`"
              :alt="asset.prompt || 'Image'"
              class="asset-detail-image"
            />
          </div>

          <!-- No image -->
          <div v-else class="generation-empty" style="color: var(--studio-text-muted);">
            <p style="font-size: 13px;">No preview available</p>
          </div>
        </div>

        <!-- Bottom film strip (versions) -->
        <div v-if="versions.length > 0" class="asset-detail-filmstrip">
          <div
            v-for="v in versions"
            :key="v.current_version"
            class="asset-detail-filmstrip-item active"
          >
            <img
              v-if="v.thumbnail_key || v.storage_key"
              :src="`/v1/assets/${v.id}/content`"
              :alt="`Version ${v.current_version}`"
            />
          </div>
        </div>
      </div>

      <!-- Right info panel -->
      <aside class="asset-detail-panel">
        <!-- Title -->
        <div class="asset-detail-panel-section">
          <h1 class="asset-detail-title">{{ asset?.title || 'Untitled' }}</h1>
          <p class="asset-detail-subtitle">{{ asset?.model || 'Unknown model' }} · v{{ asset?.current_version || 1 }}</p>
        </div>

        <!-- Prompt -->
        <div class="asset-detail-panel-section">
          <div class="asset-detail-panel-label">Prompt</div>
          <p class="asset-detail-prompt">{{ asset?.prompt || '—' }}</p>
        </div>

        <!-- Tags -->
        <div class="asset-detail-panel-section">
          <div class="asset-detail-panel-label">Tags</div>
          <div class="asset-detail-tags">
            <span
              v-for="tag in assetTags"
              :key="tag.id"
              class="asset-detail-tag"
              :style="{ background: (tag.color || '#6366f1') + '20', color: tag.color || '#6366f1' }"
              @click="removeTag(tag)"
            >
              {{ tag.name }}
              <span class="asset-detail-tag-remove">×</span>
            </span>
            <span
              class="asset-detail-tag"
              style="background: var(--studio-surface-2); color: var(--studio-text-muted); border: 1px dashed var(--studio-border);"
              @click="showTagForm = !showTagForm"
            >
              + Add
            </span>
          </div>

          <!-- Tag form -->
          <div v-if="showTagForm" class="asset-detail-tag-input">
            <input
              v-model="tagInput"
              placeholder="Tag name..."
              @keydown.enter="addTag"
            />
            <input
              v-model="tagColor"
              type="color"
              style="width: 32px; height: 32px; padding: 0; border: none; border-radius: 4px; cursor: pointer;"
            />
            <button class="generation-submit-btn" style="padding: 6px 14px;" @click="addTag">Add</button>
          </div>

          <!-- Quick-pick existing tags -->
          <div v-if="showTagForm && tagStore.tags.length > 0" style="margin-top: 8px;">
            <p style="font-size: 10px; color: var(--studio-text-muted); margin-bottom: 6px;">Existing tags:</p>
            <div class="asset-detail-tags">
              <span
                v-for="tag in tagStore.tags.filter((t) => !assetTags.some((at) => at.id === t.id)).slice(0, 8)"
                :key="tag.id"
                class="asset-detail-tag"
                :style="{ background: (tag.color || '#6366f1') + '20', color: tag.color || '#6366f1', opacity: 0.8 }"
                @click="applyExistingTag(tag)"
              >
                {{ tag.name }}
              </span>
            </div>
          </div>
        </div>

        <!-- Details -->
        <div class="asset-detail-panel-section">
          <div class="asset-detail-panel-label">Details</div>
          <div class="asset-detail-meta-grid">
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">Model</div>
              <div class="asset-detail-meta-value">{{ asset?.model || '-' }}</div>
            </div>
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">Provider</div>
              <div class="asset-detail-meta-value">{{ asset?.model_provider || '-' }}</div>
            </div>
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">Size</div>
              <div class="asset-detail-meta-value">{{ asset?.width && asset?.height ? `${asset.width}×${asset.height}` : '-' }}</div>
            </div>
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">Format</div>
              <div class="asset-detail-meta-value">{{ asset?.mime_type?.split('/')[1] || '-' }}</div>
            </div>
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">File Size</div>
              <div class="asset-detail-meta-value">{{ formatSize(asset?.file_size) }}</div>
            </div>
            <div class="asset-detail-meta-item">
              <div class="asset-detail-meta-key">Created</div>
              <div class="asset-detail-meta-value">{{ formatDate(asset?.created_at || '') }}</div>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="asset-detail-panel-section">
          <div class="asset-detail-panel-label">Actions</div>
          <div class="asset-detail-btn-row">
            <button class="asset-detail-btn primary" @click="downloadImage">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
              Download
            </button>
            <button class="asset-detail-btn" @click="router.push('/app/gallery')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/></svg>
              View in Gallery
            </button>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

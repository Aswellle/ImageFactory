<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { projectApi } from '@/api/project'

const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()

interface Project {
  id: number
  name: string
  description?: string
  status: string
  created_at: string
}

const project = ref<Project | null>(null)
const loading = ref(true)

const assets = computed(() => {
  return assetStore.assets
})

onMounted(async () => {
  try {
    project.value = await projectApi.get(route.params.id as string)
    await assetStore.fetchAssets({ project_id: route.params.id as string })
  } catch {
    router.push('/app/projects')
  } finally {
    loading.value = false
  }
})

function openAsset(id: number) {
  router.push(`/assets/${id}`)
}

function deleteAsset(id: number) {
  if (confirm('Delete this asset?')) {
    assetStore.deleteAsset(id)
  }
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>

<template>
  <div class="studio-theme">
    <div class="project-detail-studio">
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
          <RouterLink to="/app/gallery" class="studio-nav-item" title="Gallery">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M3 15l4-4a2 2 0 012.8 0L14 15" />
              <path d="M14 13l1-1a2 2 0 012.8 0L21 15" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/projects" class="studio-nav-item active" title="Projects">
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
      <div class="project-detail-main">
        <!-- Top bar -->
        <header class="project-detail-topbar">
          <button class="project-detail-back" @click="router.push('/app/projects')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
            Back to Projects
          </button>
          <span class="project-detail-title">{{ project?.name || 'Loading...' }}</span>
          <div class="project-detail-actions">
            <button class="project-detail-action-btn" @click="router.push('/app/gallery')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/></svg>
              View All
            </button>
          </div>
        </header>

        <!-- Loading -->
        <div v-if="loading" class="project-detail-content">
          <div class="projects-skeleton-grid">
            <div v-for="n in 6" :key="n" class="projects-skeleton-card">
              <div class="projects-skeleton-previews">
                <div v-for="i in 9" :key="i" class="projects-skeleton-preview"></div>
              </div>
              <div class="projects-skeleton-info">
                <div class="projects-skeleton-line"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Content -->
        <div v-else-if="project" class="project-detail-content">
          <!-- Project info -->
          <div class="project-info-header">
            <h1 class="project-info-name">{{ project.name }}</h1>
            <p v-if="project.description" class="project-info-desc">{{ project.description }}</p>
            <div class="project-info-meta">
              <span>{{ assets.length }} images</span>
              <span class="contact-sheet-meta-dot"></span>
              <span>Created {{ formatDate(project.created_at) }}</span>
            </div>
          </div>

          <!-- Empty state -->
          <div v-if="assets.length === 0" class="project-detail-empty">
            <div class="project-detail-empty-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <p class="project-detail-empty-text">No images in this project yet</p>
            <button class="project-detail-empty-btn" @click="router.push('/app')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14M5 12h14"/></svg>
              Generate images
            </button>
          </div>

          <!-- Light Table Grid -->
          <div v-else class="light-table-grid">
            <div
              v-for="asset in assets"
              :key="asset.id"
              class="light-table-item"
              @click="openAsset(asset.id)"
            >
              <div class="light-table-frame">
                <div class="light-table-img-wrap">
                  <img
                    v-if="asset.thumbnail_key || asset.storage_key"
                    :src="`/v1/assets/${asset.id}/content`"
                    :alt="asset.prompt"
                    loading="lazy"
                  />
                </div>
              </div>
              <div class="light-table-overlay">
                <button class="light-table-overlay-btn" @click.stop="openAsset(asset.id)" title="View">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                </button>
                <button class="light-table-overlay-btn" @click.stop="deleteAsset(asset.id)" title="Delete" style="background: #ef4444; color: white;">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

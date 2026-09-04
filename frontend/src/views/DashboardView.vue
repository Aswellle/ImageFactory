<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAssetStore } from '@/stores/asset'
import { useGenerationStore } from '@/stores/generation'

const assetStore = useAssetStore()
const generationStore = useGenerationStore()

const prompt = ref('')
const generating = ref(false)
const activeIndex = ref(-1)

const recentAssets = computed(() => assetStore.assets.slice(0, 12))
const selectedAsset = computed(() => activeIndex.value >= 0 ? recentAssets.value[activeIndex.value] : null)

onMounted(async () => {
  await assetStore.fetchAssets({ page_size: 12 })
})

async function handleGenerate() {
  if (!prompt.value.trim() || generating.value) return
  generating.value = true
  try {
    await generationStore.generate({ prompt: prompt.value.trim() })
    prompt.value = ''
    await assetStore.fetchAssets({ page_size: 12 })
  } finally {
    generating.value = false
  }
}

function onImageError(event: Event) {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
}
</script>

<template>
  <div class="studio-theme">
    <div class="dashboard-studio">
      <!-- Sidebar -->
      <aside class="studio-sidebar">
        <div class="studio-sidebar-logo">
          <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
            <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="#0C0A09" />
          </svg>
        </div>
        <nav class="studio-sidebar-nav">
          <button class="studio-nav-item active" title="Generate">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
          </button>
          <button class="studio-nav-item" title="Gallery">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M3 15l4-4a2 2 0 012.8 0L14 15" />
              <path d="M14 13l1-1a2 2 0 012.8 0L21 15" />
            </svg>
          </button>
          <button class="studio-nav-item" title="Projects">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
            </svg>
          </button>
          <button class="studio-nav-item" title="API">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 18l6-6-6-6M8 6l-6 6 6 6" />
            </svg>
          </button>
        </nav>
      </aside>

      <!-- Main -->
      <div class="studio-main">
        <!-- Top bar -->
        <header class="studio-topbar">
          <span class="studio-topbar-title">Generate</span>
          <div class="studio-topbar-actions">
            <button class="studio-param-chip" style="cursor:default">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
              flux-1-pro
            </button>
            <button class="studio-param-chip" style="cursor:default">
              1024 x 1024
            </button>
          </div>
        </header>

        <!-- Viewport -->
        <div class="studio-viewport">
          <div class="studio-viewport-canvas">
            <!-- Empty state -->
            <div v-if="!selectedAsset && !generating && recentAssets.length === 0" class="studio-empty">
              <div class="studio-empty-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                  <rect x="3" y="3" width="18" height="18" rx="2" />
                  <circle cx="8.5" cy="8.5" r="1.5" />
                  <path d="M21 15l-5-5L5 21" />
                </svg>
              </div>
              <p class="studio-empty-title">开始创作</p>
              <p class="studio-empty-desc">输入描述，让 AI 为你生成独一无二的视觉作品</p>
            </div>

            <!-- Generating state -->
            <div v-else-if="generating" class="studio-generating">
              <div class="studio-spinner"></div>
              <p class="studio-generating-text">正在生成中...</p>
            </div>

            <!-- Generated result -->
            <div v-else-if="selectedAsset" class="studio-result">
              <img :src="`/v1/assets/${selectedAsset.id}/content`"] :alt="selectedAsset.prompt" @error="onImageError" />
            </div>
          </div>

          <!-- Floating prompt bar -->
          <div class="studio-prompt-bar">
            <div class="studio-prompt-inner">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
              <input
                v-model="prompt"
                type="text"
                class="studio-prompt-input"
                placeholder="描述你想要生成的图片..."
                @keydown.enter="handleGenerate"
              />
              <button
                class="studio-prompt-btn"
                :disabled="!prompt.trim() || generating"
                @click="handleGenerate"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M7 17L17 7M17 7H8M17 7v9" />
                </svg>
              </button>
            </div>
            <div class="studio-params">
              <span class="studio-param-chip">flux-1-pro</span>
              <span class="studio-param-chip">1024 x 1024</span>
              <span class="studio-param-chip">1 image</span>
            </div>
          </div>
        </div>

        <!-- Filmstrip -->
        <div v-if="recentAssets.length > 0" class="studio-filmstrip">
          <span class="studio-filmstrip-label">Recent</span>
          <div
            v-for="(asset, idx) in recentAssets"
            :key="asset.id"
            class="studio-filmstrip-item"
            :class="{ active: activeIndex === idx }"
            @click="activeIndex = idx"
          >
            <img :src="`/v1/assets/${asset.id}/content`"] :alt="asset.prompt" @error="onImageError" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

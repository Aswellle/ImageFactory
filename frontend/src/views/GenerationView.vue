<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useGenerationStore } from '@/stores/generation'
import { useRouter } from 'vue-router'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const router = useRouter()
const store = useGenerationStore()

const prompt = ref('')
const negativePrompt = ref('')
const model = ref('gpt-image-1')
const size = ref('1024x1024')
const imageCount = ref(1)
const showAdvanced = ref(false)
const showPanel = ref(true)

const sizes = [
  { label: 'Square', value: '1024x1024', ratio: '1:1' },
  { label: 'Wide', value: '1792x1024', ratio: '16:9' },
  { label: 'Tall', value: '1024x1792', ratio: '9:16' },
]

const models = [
  { label: 'GPT Image', value: 'gpt-image-1', desc: 'Fast · High quality' },
  { label: 'DALL·E 3', value: 'dall-e-3', desc: 'Creative · Detailed' },
]

const charCount = computed(() => prompt.value.length)
const canSubmit = computed(() => prompt.value.trim().length > 0 && !store.loading)

async function submit() {
  if (!canSubmit.value) return
  try {
    const jobId = await store.generate({
      prompt: prompt.value.trim(),
      negative_prompt: negativePrompt.value.trim() || undefined,
      model: model.value,
      size: size.value,
      n: imageCount.value,
    })
    await store.pollUntilDone(jobId)
  } catch {
    // error surfaced via store.error
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.metaKey && !e.ctrlKey) {
    e.preventDefault()
    submit()
  }
}

function goToGallery() {
  router.push('/app/gallery')
}

function regenerate() {
  submit()
}

onUnmounted(() => {
  store.cancelPolling()
})

function formatSizeLabel(sizeValue: string): string {
  const found = sizes.find(s => s.value === sizeValue)
  return found ? `${found.label} (${found.ratio})` : sizeValue
}
</script>

<template>
  <div class="studio-theme">
    <div class="generation-studio">
      <!-- Sidebar -->
      <aside class="studio-sidebar">
        <div class="studio-sidebar-logo">
          <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
            <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="#0C0A09" />
          </svg>
        </div>
        <nav class="studio-sidebar-nav">
          <RouterLink to="/app" class="studio-nav-item active" title="Generate">
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
      <div class="generation-viewport">
        <!-- Stage (developing area) -->
        <div class="generation-stage">
          <!-- Panel toggle button -->
          <button class="generation-panel-toggle" @click="showPanel = !showPanel" :title="showPanel ? 'Hide panel' : 'Show panel'">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path :d="showPanel ? 'M15 18l-6-6 6-6' : 'M9 18l6-6-6-6'" />
            </svg>
          </button>

          <!-- Empty state -->
          <div v-if="!store.loading && !store.activeJob && !store.error" class="generation-empty">
            <div class="generation-empty-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <p class="generation-empty-title">Start creating</p>
            <p class="generation-empty-desc">Type a prompt below to generate your first image. The canvas will come alive as your image develops.</p>
          </div>

          <!-- Developing state -->
          <div v-else-if="store.loading && store.activeJob" class="generation-developing">
            <div class="generation-developing-frame">
              <div class="generation-developing-overlay">
                <div class="generation-developing-dots">
                  <span class="generation-developing-dot"></span>
                  <span class="generation-developing-dot"></span>
                  <span class="generation-developing-dot"></span>
                </div>
                <span class="generation-developing-label">{{ store.activeJob.status === 'pending' ? 'In queue' : 'Developing' }}</span>
              </div>
            </div>
          </div>

          <!-- Result state (single) -->
          <div v-else-if="store.activeJob?.status === 'completed' && imageCount === 1" class="generation-result">
            <div class="generation-result-frame">
              <div class="generation-result-placeholder">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color: var(--studio-text-muted);">
                  <rect x="3" y="3" width="18" height="18" rx="2" />
                  <circle cx="8.5" cy="8.5" r="1.5" />
                  <path d="M21 15l-5-5L5 21" />
                </svg>
              </div>
              <div class="generation-result-info">
                <span class="generation-result-prompt">{{ store.activeJob.prompt }}</span>
                <span class="generation-result-meta">{{ store.activeJob.model }} · {{ size }}</span>
              </div>
            </div>
            <div class="generation-actions">
              <button class="generation-action-btn primary" @click="regenerate">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/></svg>
                Regenerate
              </button>
              <button class="generation-action-btn" @click="goToGallery">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
                View in Gallery
              </button>
            </div>
          </div>

          <!-- Result state (multiple) -->
          <div v-else-if="store.activeJob?.status === 'completed' && imageCount > 1" class="generation-results-grid">
            <div v-for="n in imageCount" :key="n" class="generation-result-frame">
              <div class="generation-result-placeholder" style="height: 200px; display: flex; align-items: center; justify-content: center;">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color: var(--studio-text-muted);">
                  <rect x="3" y="3" width="18" height="18" rx="2" />
                  <circle cx="8.5" cy="8.5" r="1.5" />
                  <path d="M21 15l-5-5L5 21" />
                </svg>
              </div>
            </div>
          </div>

          <!-- Error state -->
          <div v-else-if="store.error" class="generation-error">
            <div class="generation-error-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/></svg>
            </div>
            <p class="generation-error-text">{{ store.error }}</p>
            <button class="generation-action-btn" @click="store.error = null">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/></svg>
              Try again
            </button>
          </div>
        </div>

        <!-- Bottom prompt bar -->
        <div class="generation-prompt-bar">
          <input
            v-model="prompt"
            type="text"
            class="generation-prompt-input"
            placeholder="Describe the image you want to create..."
            :disabled="store.loading"
            @keydown="handleKeydown"
          />
          <span class="generation-prompt-counter">{{ charCount }}</span>
          <button class="generation-submit-btn" :disabled="!canSubmit" @click="submit">
            <svg v-if="!store.loading" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
            <LoadingSpinner v-else size="sm" />
            <span>{{ store.loading ? 'Creating...' : 'Generate' }}</span>
          </button>
        </div>
      </div>

      <!-- Right parameter panel -->
      <aside class="generation-panel" :class="{ collapsed: !showPanel, open: showPanel }">
        <div class="generation-panel-header">
          <span class="generation-panel-title">Parameters</span>
        </div>
        <div class="generation-panel-body">
          <!-- Model -->
          <div class="generation-field">
            <label class="generation-label">Model</label>
            <div class="generation-model-row">
              <button
                v-for="m in models"
                :key="m.value"
                class="generation-model-btn"
                :class="{ active: model === m.value }"
                @click="model = m.value"
              >
                <div class="generation-model-name">{{ m.label }}</div>
                <div class="generation-model-desc">{{ m.desc }}</div>
              </button>
            </div>
          </div>

          <!-- Size -->
          <div class="generation-field">
            <label class="generation-label">Size</label>
            <div class="generation-size-grid">
              <button
                v-for="s in sizes"
                :key="s.value"
                class="generation-size-btn"
                :class="{ active: size === s.value }"
                @click="size = s.value"
              >
                {{ s.label }}
                <br />
                <span style="font-size: 9px; opacity: 0.7;">{{ s.ratio }}</span>
              </button>
            </div>
          </div>

          <!-- Count -->
          <div class="generation-field">
            <label class="generation-label">Count</label>
            <div class="generation-count-row">
              <button
                v-for="n in [1, 2, 4]"
                :key="n"
                class="generation-count-btn"
                :class="{ active: imageCount === n }"
                @click="imageCount = n"
              >
                {{ n }}
              </button>
            </div>
          </div>

          <!-- Advanced toggle -->
          <button class="generation-advanced-toggle" :class="{ open: showAdvanced }" @click="showAdvanced = !showAdvanced">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 9l6 6 6-6"/></svg>
            Advanced
          </button>

          <!-- Advanced fields -->
          <div v-if="showAdvanced" style="margin-top: 12px;">
            <div class="generation-field">
              <label class="generation-label">Negative Prompt</label>
              <input
                v-model="negativePrompt"
                class="generation-input"
                placeholder="What to avoid..."
              />
            </div>
          </div>

          <!-- Active job info -->
          <div v-if="store.activeJob" style="margin-top: 20px; padding-top: 16px; border-top: 1px solid var(--studio-border);">
            <div class="generation-label" style="margin-bottom: 8px;">Status</div>
            <div style="display: flex; align-items: center; gap: 8px;">
              <span
                style="width: 8px; height: 8px; border-radius: 50%;"
                :style="{
                  background: store.activeJob.status === 'completed' ? '#22c55e' : store.activeJob.status === 'failed' ? '#ef4444' : 'var(--studio-accent)',
                  animation: store.activeJob.status === 'processing' || store.activeJob.status === 'pending' ? 'dotPulse 1.5s ease-in-out infinite' : 'none'
                }"
              ></span>
              <span style="font-size: 12px; color: var(--studio-text-secondary); font-family: 'JetBrains Mono', monospace;">
                {{ store.activeJob.status }}
              </span>
            </div>
            <div style="margin-top: 8px; font-size: 11px; color: var(--studio-text-muted); font-family: 'JetBrains Mono', monospace;">
              {{ formatSizeLabel(size) }} · {{ model }} · ×{{ imageCount }}
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

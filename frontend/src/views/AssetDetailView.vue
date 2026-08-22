<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import type { Asset } from '@/api/asset'

const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()

const asset = ref<Asset | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    asset.value = await assetStore.fetchAsset(id)
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
</script>

<template>
  <div v-if="loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>
  <div v-else-if="asset" class="space-y-6">
    <header>
      <button class="text-sm text-text-secondary hover:text-text mb-3" @click="$router.back()">← Back</button>
      <h1 class="text-xl font-semibold tracking-tight">{{ asset.title || 'Untitled' }}</h1>
    </header>

    <!-- Image preview -->
    <section class="if-card overflow-hidden">
      <div class="bg-surface-2 flex items-center justify-center min-h-[300px]">
        <img v-if="asset.storage_key" :src="`/v1/assets/${asset.id}/content`" class="max-w-full max-h-[70vh] object-contain" />
        <div v-else class="text-text-secondary text-sm p-8 text-center">
          {{ asset.prompt?.slice(0, 200) || 'No preview available' }}
        </div>
      </div>
    </section>

    <!-- Details -->
    <section class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="if-card p-5 space-y-4">
        <h3 class="text-sm font-medium">Details</h3>
        <dl class="space-y-2 text-sm">
          <div class="flex justify-between">
            <dt class="text-text-secondary">Model</dt>
            <dd>{{ asset.model || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">Provider</dt>
            <dd>{{ asset.model_provider || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">Size</dt>
            <dd>{{ asset.width && asset.height ? `${asset.width}×${asset.height}` : '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">Aspect Ratio</dt>
            <dd>{{ asset.aspect_ratio || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">Format</dt>
            <dd>{{ asset.mime_type || '-' }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">File Size</dt>
            <dd>{{ formatSize(asset.file_size) }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-text-secondary">Created</dt>
            <dd>{{ asset.created_at }}</dd>
          </div>
        </dl>
      </div>

      <div class="if-card p-5 space-y-4">
        <h3 class="text-sm font-medium">Prompt</h3>
        <p class="text-sm text-text-secondary whitespace-pre-wrap">{{ asset.prompt || '-' }}</p>
        <!-- TODO: Add negative_prompt field to asset schema -->
      </div>
    </section>

    <!-- Versions -->
    <section v-if="versions.length > 0" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Version History</h3>
      <div class="flex gap-3">
        <div v-for="v in versions" :key="v.current_version" class="w-24">
          <div class="aspect-square bg-surface-2 rounded overflow-hidden">
            <img v-if="v.thumbnail_key" :src="`/v1/assets/${v.id}/content`" class="w-full h-full object-cover" />
          </div>
          <div class="text-xs text-text-secondary mt-1 text-center">v{{ v.current_version }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

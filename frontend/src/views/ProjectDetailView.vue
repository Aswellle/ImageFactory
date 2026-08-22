<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { projectApi } from '@/api/project'
import type { Project } from '@/api/project'

const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()

const project = ref<Project | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    project.value = await projectApi.get(route.params.id as string)
    await assetStore.fetchAssets({ project_id: route.params.id as string })
  } catch {
    router.push('/projects')
  } finally {
    loading.value = false
  }
})

function deleteAsset(id: number) {
  if (confirm('Delete this asset?')) {
    assetStore.deleteAsset(id)
  }
}
</script>

<template>
  <div v-if="loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>
  <div v-else-if="project" class="space-y-6">
    <header>
      <button class="text-sm text-text-secondary hover:text-text mb-3" @click="$router.push('/projects')">← Back</button>
      <h1 class="text-xl font-semibold tracking-tight">{{ project.name }}</h1>
      <p v-if="project.description" class="text-sm text-text-secondary mt-1">{{ project.description }}</p>
    </header>

    <div v-if="assetStore.assets.length === 0" class="if-card">
      <div class="py-16 text-center text-text-secondary text-sm">No images in this project.</div>
    </div>

    <div v-else class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4">
      <div v-for="asset in assetStore.assets" :key="asset.id" class="break-inside-avoid mb-4">
        <div class="if-card overflow-hidden group relative cursor-pointer" @click="$router.push(`/assets/${asset.id}`)">
          <div class="aspect-auto min-h-[100px] bg-surface-2">
            <img v-if="asset.thumbnail_key || asset.storage_key" :src="`/v1/assets/${asset.id}/content`" class="w-full h-auto object-cover" loading="lazy" />
          </div>
          <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-end opacity-0 group-hover:opacity-100">
            <div class="p-2 w-full flex justify-end gap-1">
              <button class="bg-white text-xs px-2 py-1 rounded" @click.stop="deleteAsset(asset.id)">Delete</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAssetStore } from '@/stores/asset'
import { useProjectStore } from '@/stores/project'

const assetStore = useAssetStore()
const projectStore = useProjectStore()

onMounted(() => {
  assetStore.fetchAssets({ page_size: 8 })
  projectStore.fetchProjects()
})

</script>

<template>
  <div class="space-y-8">
    <!-- Hero -->
    <section class="text-center max-w-2xl mx-auto pt-4">
      <h1 class="text-2xl font-semibold tracking-tight">Create your next image</h1>
      <p class="text-text-secondary mt-2 text-sm">
        Describe the image you want, or edit an existing one.
      </p>
      <div class="flex justify-center gap-3 mt-5">
        <RouterLink to="/create" class="if-btn-primary">Start generating →</RouterLink>
        <RouterLink to="/assets" class="if-btn-ghost">View Gallery</RouterLink>
      </div>
    </section>

    <!-- Stats -->
    <section class="grid grid-cols-3 gap-4">
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Total images</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">{{ assetStore.total || 0 }}</div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Projects</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">{{ projectStore.projects.length || 0 }}</div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Storage</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">0 B</div>
      </div>
    </section>

    <!-- Recent images -->
    <section v-if="assetStore.assets.length > 0">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base font-medium">Recent</h2>
        <RouterLink to="/assets" class="text-sm text-accent hover:underline">View all →</RouterLink>
      </div>
      <div class="columns-2 md:columns-4 gap-4 space-y-4">
        <div v-for="asset in assetStore.assets.slice(0, 8)" :key="asset.id" class="break-inside-avoid mb-4 cursor-pointer" @click="$router.push(`/assets/${asset.id}`)">
          <div class="if-card overflow-hidden">
            <div class="aspect-auto min-h-[80px] bg-surface-2">
              <img v-if="asset.thumbnail_key || asset.storage_key" :src="`/v1/assets/${asset.id}/content`" class="w-full h-auto object-cover" loading="lazy" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Empty state -->
    <section v-else class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">Your images will appear here</div>
        <div class="text-xs text-text-secondary">Generate your first image to get started.</div>
      </div>
    </section>
  </div>
</template>

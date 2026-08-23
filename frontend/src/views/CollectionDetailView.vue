<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCollectionStore } from '@/stores/collection'
import { useFavoriteStore } from '@/stores/favorite'
import type { Asset } from '@/api/asset'
import FavoriteButton from '@/components/FavoriteButton.vue'

const route = useRoute()
const router = useRouter()
const store = useCollectionStore()
const assets = ref<Asset[]>([])
const loading = ref(true)
const collection = ref<{ id: number; name: string; description?: string } | null>(null)

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    collection.value = await store.getCollection(id)
    assets.value = (await store.listAssets(id)) as Asset[]
    await useFavoriteStore().syncFavorites()
  } catch {
    router.push('/collections')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="text-center py-20 text-body">Loading…</div>
  <div v-else-if="collection" class="space-y-6">
    <header>
      <button class="text-sm text-text-secondary hover:text-text mb-3 transition-colors" @click="router.push('/collections')">← Collections</button>
      <h1 class="text-heading">{{ collection.name }}</h1>
      <p v-if="collection.description" class="text-body mt-1">{{ collection.description }}</p>
    </header>

    <!-- Empty state -->
    <div v-if="assets.length === 0" class="card animate-fade-in">
      <div class="py-20 text-center">
        <div class="text-body mb-1">No assets in this collection</div>
        <div class="text-caption">Add assets from the asset detail page</div>
      </div>
    </div>

    <!-- Asset grid -->
    <div v-else class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4">
      <div
        v-for="asset in assets"
        :key="asset.id"
        class="break-inside-avoid mb-4 cursor-pointer group relative"
        @click="router.push(`/assets/${asset.id}`)"
      >
        <div class="card overflow-hidden">
          <div class="relative">
            <div class="aspect-auto min-h-[120px] bg-surface-2 flex items-center justify-center">
              <img
                v-if="asset.thumbnail_key || asset.storage_key"
                :src="`/v1/assets/${asset.id}/content`"
                class="w-full h-auto object-cover"
                loading="lazy"
              />
            </div>
            <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
              <FavoriteButton :asset-id="asset.id" size="sm" :initial-favorited="useFavoriteStore().isFavorited(asset.id)" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

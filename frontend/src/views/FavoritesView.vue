<script setup lang="ts">
import { onMounted } from 'vue'
import { useFavoriteStore } from '@/stores/favorite'
import FavoriteButton from '@/components/FavoriteButton.vue'

const store = useFavoriteStore()

onMounted(() => {
  store.fetchFavorites()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-heading">Favorites</h1>
      <p class="text-body mt-1">{{ store.favorites.length }} favorite images</p>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-20 text-body">Loading…</div>

    <!-- Empty state -->
    <div v-else-if="store.favorites.length === 0" class="card animate-fade-in">
      <div class="py-20 text-center">
        <div class="text-body mb-1">No favorites yet</div>
        <div class="text-caption">Click the heart icon on any image to save it here</div>
      </div>
    </div>

    <!-- Favorites grid -->
    <div v-else class="columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4">
      <div
        v-for="fav in store.favorites"
        :key="fav.id"
        class="break-inside-avoid mb-4 cursor-pointer group relative"
        @click="$router.push(`/assets/${fav.asset_id}`)"
      >
        <div class="card overflow-hidden">
          <div class="relative">
            <div class="aspect-auto min-h-[120px] bg-surface-2 flex items-center justify-center">
              <img
                :src="`/v1/assets/${fav.asset_id}/content`"
                class="w-full h-auto object-cover"
                loading="lazy"
              />
            </div>
            <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
              <FavoriteButton :asset-id="fav.asset_id" size="sm" />
            </div>
          </div>
          <div class="p-3">
            <p class="text-caption">Asset #{{ fav.asset_id }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

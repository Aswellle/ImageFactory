<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAssetStore } from '@/stores/asset'
import { useProjectStore } from '@/stores/project'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const assetStore = useAssetStore()
const projectStore = useProjectStore()

const ready = ref(false)

onMounted(async () => {
  await Promise.all([
    assetStore.fetchAssets({ page_size: 8 }),
    projectStore.fetchProjects(),
  ])
  ready.value = true
})
</script>

<template>
  <div class="space-y-8">
    <!-- Hero -->
    <section class="text-center max-w-2xl mx-auto pt-4 animate-fade-in">
      <h1 class="text-display">
        {{ t('dashboard.heroTitle') }}
      </h1>
      <p class="text-body mt-3">
        {{ t('dashboard.heroSubtitle') }}
      </p>
      <div class="flex justify-center gap-3 mt-7">
        <RouterLink to="/create" class="btn btn-primary">
          {{ t('dashboard.startGenerating') }}
          <span class="ml-0.5">→</span>
        </RouterLink>
        <RouterLink to="/assets" class="btn btn-ghost">
          {{ t('dashboard.viewGallery') }}
        </RouterLink>
      </div>
    </section>

    <!-- Stats -->
    <section class="grid grid-cols-1 sm:grid-cols-3 gap-4 stagger">
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">{{ t('dashboard.totalImages') }}</div>
        <div class="metric-value mt-2">{{ assetStore.total || 0 }}</div>
      </div>
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">{{ t('dashboard.projects') }}</div>
        <div class="metric-value mt-2">{{ projectStore.projects.length || 0 }}</div>
      </div>
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">{{ t('dashboard.storage') }}</div>
        <div class="metric-value mt-2">0 B</div>
      </div>
    </section>

    <!-- Recent images -->
    <section v-if="assetStore.assets.length > 0">
      <div class="flex items-center justify-between mb-5">
        <h2 class="text-subheading">{{ t('dashboard.recent') }}</h2>
        <RouterLink to="/assets" class="text-sm text-accent hover:underline-offset-4 hover:underline transition-all">{{ t('dashboard.viewAll') }} →</RouterLink>
      </div>
      <div class="columns-2 md:columns-4 gap-4 space-y-4">
        <div
          v-for="asset in assetStore.assets.slice(0, 8)"
          :key="asset.id"
          class="break-inside-avoid mb-4 group"
          @click="$router.push(`/assets/${asset.id}`)"
        >
          <div class="card card-interactive overflow-hidden">
            <div class="relative aspect-auto min-h-[80px] bg-surface-2 overflow-hidden">
              <img
                v-if="asset.thumbnail_key || asset.storage_key"
                :src="`/v1/assets/${asset.id}/content`"
                class="w-full h-auto object-cover img-loading"
                loading="lazy"
                @load="($event.target as HTMLImageElement)?.classList.add('img-loaded')"
              />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Empty state -->
    <section v-else-if="ready">
      <div class="card animate-fade-in">
        <div class="py-20 text-center">
          <div class="text-body mb-1">{{ t('dashboard.emptyPrimary') }}</div>
          <div class="text-caption">{{ t('dashboard.emptySecondary') }}</div>
        </div>
      </div>
    </section>

    <!-- Loading skeleton -->
    <section v-else class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="n in 8" :key="n" class="skeleton aspect-[4/3]"></div>
    </section>
  </div>
</template>

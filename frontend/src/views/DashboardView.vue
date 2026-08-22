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
    <section class="text-center max-w-2xl mx-auto pt-6 animate-fade-in">
      <h1 class="text-2xl font-semibold tracking-tight text-text">
        {{ t('dashboard.heroTitle') }}
      </h1>
      <p class="text-text-secondary mt-2 text-sm leading-relaxed">
        {{ t('dashboard.heroSubtitle') }}
      </p>
      <div class="flex justify-center gap-3 mt-6">
        <RouterLink to="/create" class="if-btn-primary !px-5 !py-2.5">
          {{ t('dashboard.startGenerating') }}
          <span class="ml-0.5">→</span>
        </RouterLink>
        <RouterLink to="/assets" class="if-btn-ghost !px-5 !py-2.5">
          {{ t('dashboard.viewGallery') }}
        </RouterLink>
      </div>
    </section>

    <!-- Stats -->
    <section class="grid grid-cols-3 gap-4 animate-stagger">
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide font-medium">{{ t('dashboard.totalImages') }}</div>
        <div class="text-2xl font-semibold mt-2 tabular-nums">{{ assetStore.total || 0 }}</div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide font-medium">{{ t('dashboard.projects') }}</div>
        <div class="text-2xl font-semibold mt-2 tabular-nums">{{ projectStore.projects.length || 0 }}</div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide font-medium">{{ t('dashboard.storage') }}</div>
        <div class="text-2xl font-semibold mt-2 tabular-nums">0 B</div>
      </div>
    </section>

    <!-- Recent images -->
    <section v-if="assetStore.assets.length > 0">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base font-medium">{{ t('dashboard.recent') }}</h2>
        <RouterLink to="/assets" class="text-sm text-accent hover:underline-offset-4 hover:underline transition-all">{{ t('dashboard.viewAll') }} →</RouterLink>
      </div>
      <div class="columns-2 md:columns-4 gap-4 space-y-4">
        <div
          v-for="asset in assetStore.assets.slice(0, 8)"
          :key="asset.id"
          class="break-inside-avoid mb-4 group"
          @click="$router.push(`/assets/${asset.id}`)"
        >
          <div class="if-card if-card-interactive overflow-hidden">
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
    <section v-else-if="ready" class="if-card animate-fade-in">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">{{ t('dashboard.emptyPrimary') }}</div>
        <div class="text-xs text-text-muted">{{ t('dashboard.emptySecondary') }}</div>
      </div>
    </section>

    <!-- Loading skeleton -->
    <section v-else class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="n in 8" :key="n" class="skeleton aspect-[4/3]"></div>
    </section>
  </div>
</template>

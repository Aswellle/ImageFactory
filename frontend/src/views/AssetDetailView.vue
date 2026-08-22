<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAssetStore } from '@/stores/asset'
import { useTagStore } from '@/stores/tag'
import type { Asset } from '@/api/asset'
import type { Tag } from '@/api/tag'
import FavoriteButton from '@/components/FavoriteButton.vue'
import TagBadge from '@/components/TagBadge.vue'
import CollectionPicker from '@/components/CollectionPicker.vue'

const route = useRoute()
const router = useRouter()
const assetStore = useAssetStore()
const tagStore = useTagStore()

const asset = ref<Asset | null>(null)
const loading = ref(true)
const assetTags = ref<Tag[]>([])
const tagInput = ref('')
const tagColor = ref('#6366f1')
const showTagForm = ref(false)

const assetId = computed(() => asset.value ? Number(asset.value.id) : Number(route.params.id))

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    asset.value = await assetStore.fetchAsset(id)
    assetTags.value = await tagStore.fetchAssetTags(id)
  } catch {
    router.push('/assets')
  } finally {
    loading.value = false
  }
})

watch(() => route.params.id, async (newId) => {
  if (!newId) return
  loading.value = true
  try {
    const id = Number(newId)
    asset.value = await assetStore.fetchAsset(id)
    assetTags.value = await tagStore.fetchAssetTags(id)
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

async function addTag() {
  if (!tagInput.value.trim()) return
  try {
    await tagStore.createTag(tagInput.value.trim(), tagColor.value)
    assetTags.value = await tagStore.fetchAssetTags(assetId.value)
    tagInput.value = ''
    showTagForm.value = false
  } catch {
    // error handled in store
  }
}

async function removeTag(tag: Tag) {
  try {
    await tagStore.untagAsset(tag.id, assetId.value)
    assetTags.value = assetTags.value.filter((t) => t.id !== tag.id)
  } catch {
    // error handled in store
  }
}

async function applyExistingTag(tag: Tag) {
  try {
    await tagStore.tagAsset(tag.id, assetId.value)
    if (!assetTags.value.some((t) => t.id === tag.id)) {
      assetTags.value = await tagStore.fetchAssetTags(assetId.value)
    }
  } catch {
    // error handled in store
  }
}
</script>

<template>
  <div v-if="loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>
  <div v-else-if="asset" class="space-y-6">
    <header class="flex items-start justify-between">
      <div>
        <button class="text-sm text-text-secondary hover:text-text mb-3" @click="$router.back()">← Back</button>
        <h1 class="text-xl font-semibold tracking-tight">{{ asset.title || 'Untitled' }}</h1>
      </div>
      <FavoriteButton :asset-id="asset.id" size="lg" />
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

    <!-- Tags & Collections -->
    <section class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Tags</h3>
      <div class="flex flex-wrap gap-2">
        <TagBadge
          v-for="tag in assetTags"
          :key="tag.id"
          :tag="tag"
          @click="removeTag(tag)"
        />
        <button
          type="button"
          class="inline-flex items-center gap-1 rounded-full border border-dashed border-border px-2.5 py-0.5 text-xs text-text-secondary transition-colors hover:border-text-muted hover:text-text"
          @click="showTagForm = !showTagForm"
        >
          + Add Tag
        </button>
      </div>

      <!-- Tag form -->
      <div v-if="showTagForm" class="space-y-3 rounded-lg border border-border p-3">
        <div class="flex gap-2">
          <input
            v-model="tagInput"
            class="if-input flex-1"
            placeholder="Tag name…"
            @keyup.enter="addTag"
          />
          <input
            v-model="tagColor"
            type="color"
            class="h-9 w-9 cursor-pointer rounded border-0 bg-transparent"
          />
          <button
            type="button"
            class="rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-primary/90"
            @click="addTag"
          >
            Add
          </button>
        </div>

        <!-- Quick-pick existing tags -->
        <div v-if="tagStore.tags.length > 0" class="space-y-2">
          <p class="text-xs text-text-secondary">Quick add existing tag:</p>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="tag in tagStore.tags.filter((t) => !assetTags.some((at) => at.id === t.id))"
              :key="tag.id"
              type="button"
              @click="applyExistingTag(tag)"
            >
              <TagBadge :tag="tag" clickable />
            </button>
          </div>
        </div>
      </div>

      <!-- Collection picker -->
      <div class="pt-2">
        <CollectionPicker :asset-id="asset.id" />
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

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useCollectionStore } from '@/stores/collection'
import type { Collection } from '@/api/collection'

const props = defineProps<{
  assetId: number
}>()

const emit = defineEmits<{
  added: [collectionId: number]
}>()

const store = useCollectionStore()
const showDropdown = ref(false)
const loading = ref(false)

onMounted(() => {
  if (store.collections.length === 0) {
    store.fetchCollections()
  }
})

watch(() => store.collections, () => {})

async function addToCollection(collection: Collection) {
  loading.value = true
  try {
    await store.addAssetToCollection(collection.id, props.assetId)
    showDropdown.value = false
    emit('added', collection.id)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative">
    <button
      type="button"
      class="inline-flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-medium text-text-secondary transition-colors hover:bg-surface-hover hover:text-text-primary"
      @click="showDropdown = !showDropdown"
    >
      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
      </svg>
      Add to Collection
    </button>

    <div
      v-if="showDropdown"
      class="absolute right-0 z-50 mt-1 max-h-60 w-56 overflow-auto rounded-lg border border-border bg-surface shadow-lg"
    >
      <div v-if="store.collections.length === 0" class="px-3 py-2 text-sm text-text-muted">
        No collections yet
      </div>
      <button
        v-for="col in store.collections"
        :key="col.id"
        type="button"
        class="flex w-full items-center px-3 py-2 text-sm text-text-secondary transition-colors hover:bg-surface-hover hover:text-text-primary"
        :disabled="loading"
        @click="addToCollection(col)"
      >
        {{ col.name }}
      </button>
    </div>

    <!-- Backdrop to close dropdown -->
    <div
      v-if="showDropdown"
      class="fixed inset-0 z-40"
      @click="showDropdown = false"
    />
  </div>
</template>

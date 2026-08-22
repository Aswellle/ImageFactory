<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useCollectionStore } from '@/stores/collection'
import { useRouter } from 'vue-router'

const router = useRouter()
const store = useCollectionStore()
const showForm = ref(false)
const newName = ref('')
const newDescription = ref('')

onMounted(() => {
  store.fetchCollections()
})

async function createCollection() {
  if (!newName.value.trim()) return
  await store.createCollection(newName.value.trim(), newDescription.value.trim() || undefined)
  newName.value = ''
  newDescription.value = ''
  showForm.value = false
}

async function deleteCollection(id: number) {
  if (confirm('Delete this collection?')) {
    await store.deleteCollection(id)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Collections</h1>
        <p class="text-sm text-text-secondary mt-1">{{ store.collections.length }} collections</p>
      </div>
      <button
        type="button"
        class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90"
        @click="showForm = !showForm"
      >
        + New Collection
      </button>
    </div>

    <!-- Create form -->
    <div v-if="showForm" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Create Collection</h3>
      <div class="space-y-3">
        <input v-model="newName" class="if-input w-full" placeholder="Collection name…" />
        <input v-model="newDescription" class="if-input w-full" placeholder="Description (optional)…" />
        <div class="flex gap-2">
          <button
            type="button"
            class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90"
            @click="createCollection"
          >
            Create
          </button>
          <button
            type="button"
            class="rounded-lg border border-border px-4 py-2 text-sm font-medium text-text-secondary transition-colors hover:bg-surface-hover"
            @click="showForm = false"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>

    <!-- Empty state -->
    <div v-else-if="store.collections.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">No collections yet</div>
        <div class="text-xs text-text-secondary">Create your first collection to organize your assets</div>
      </div>
    </div>

    <!-- Collection list -->
    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="col in store.collections"
        :key="col.id"
        class="if-card p-5 space-y-3 transition-shadow hover:shadow-md"
      >
        <div class="flex items-start justify-between">
          <div class="cursor-pointer flex-1 min-w-0" @click="router.push(`/collections/${col.id}`)">
            <h3 class="font-medium truncate">{{ col.name }}</h3>
            <p v-if="col.description" class="text-sm text-text-secondary mt-1 line-clamp-2">{{ col.description }}</p>
          </div>
          <button
            type="button"
            class="ml-2 shrink-0 text-text-secondary hover:text-red-500 transition-colors"
            title="Delete collection"
            @click="deleteCollection(col.id)"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
        <div class="text-xs text-text-secondary">{{ col.created_at }}</div>
      </div>
    </div>
  </div>
</template>

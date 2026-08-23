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
    <!-- Header -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-heading">Collections</h1>
        <p class="text-body mt-1">{{ store.collections.length }} collections</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">+ New Collection</button>
    </div>

    <!-- Create form -->
    <div v-if="showForm" class="card p-5 space-y-4 animate-scale-in">
      <h3 class="text-subheading">Create Collection</h3>
      <div class="space-y-3">
        <input v-model="newName" class="input" placeholder="Collection name…" />
        <input v-model="newDescription" class="input" placeholder="Description (optional)…" />
        <div class="flex gap-2">
          <button class="btn btn-primary" @click="createCollection">Create</button>
          <button class="btn btn-ghost" @click="showForm = false">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="text-center py-20 text-body">Loading…</div>

    <!-- Empty state -->
    <div v-else-if="store.collections.length === 0" class="card animate-fade-in">
      <div class="py-20 text-center">
        <div class="text-body mb-1">No collections yet</div>
        <div class="text-caption">Create your first collection to organize your assets</div>
      </div>
    </div>

    <!-- Collection list -->
    <div v-else class="card divide-y divide-border overflow-hidden stagger">
      <div
        v-for="col in store.collections"
        :key="col.id"
        class="flex items-center justify-between gap-4 px-5 py-4"
      >
        <div class="min-w-0 cursor-pointer flex-1" @click="router.push(`/collections/${col.id}`)">
          <h3 class="font-medium truncate">{{ col.name }}</h3>
          <p v-if="col.description" class="text-caption mt-0.5 line-clamp-2">{{ col.description }}</p>
          <span class="text-caption mt-1 block">{{ col.created_at }}</span>
        </div>
        <button
          type="button"
          class="btn btn-ghost !p-2 text-text-muted hover:text-danger transition-colors"
          title="Delete collection"
          @click="deleteCollection(col.id)"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useTagStore } from '@/stores/tag'

const store = useTagStore()
const showForm = ref(false)
const newName = ref('')
const newColor = ref('#6366f1')

const presetColors = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4', '#3b82f6', '#6366f1', '#a855f7', '#ec4899', '#6b7280']

onMounted(() => {
  store.fetchTags()
})

async function createTag() {
  if (!newName.value.trim()) return
  await store.createTag(newName.value.trim(), newColor.value)
  newName.value = ''
  newColor.value = '#6366f1'
  showForm.value = false
}

async function deleteTag(id: number) {
  if (confirm('Delete this tag? It will be removed from all assets.')) {
    await store.deleteTag(id)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Tags</h1>
        <p class="text-sm text-text-secondary mt-1">{{ store.tags.length }} tags</p>
      </div>
      <button
        type="button"
        class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90"
        @click="showForm = !showForm"
      >
        + New Tag
      </button>
    </div>

    <!-- Create form -->
    <div v-if="showForm" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Create Tag</h3>
      <div class="space-y-3">
        <input v-model="newName" class="if-input w-full" placeholder="Tag name…" />
        <div class="flex items-center gap-2">
          <span class="text-sm text-text-secondary">Color:</span>
          <div class="flex gap-1.5">
            <button
              v-for="color in presetColors"
              :key="color"
              type="button"
              class="h-6 w-6 rounded-full border-2 transition-transform hover:scale-110"
              :class="newColor === color ? 'border-text scale-110' : 'border-transparent'"
              :style="{ backgroundColor: color }"
              @click="newColor = color"
            />
          </div>
          <input
            v-model="newColor"
            type="color"
            class="ml-2 h-8 w-8 cursor-pointer rounded border-0 bg-transparent"
          />
        </div>
        <div class="flex gap-2">
          <button
            type="button"
            class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90"
            @click="createTag"
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
    <div v-else-if="store.tags.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">No tags yet</div>
        <div class="text-xs text-text-secondary">Create tags to organize your assets</div>
      </div>
    </div>

    <!-- Tag list -->
    <div v-else class="space-y-2">
      <div
        v-for="tag in store.tags"
        :key="tag.id"
        class="if-card flex items-center justify-between p-4"
      >
        <div class="flex items-center gap-3">
          <span
            v-if="tag.color"
            class="inline-block h-3 w-3 rounded-full"
            :style="{ backgroundColor: tag.color }"
          />
          <span class="font-medium">{{ tag.name }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs text-text-secondary">{{ tag.created_at }}</span>
          <button
            type="button"
            class="text-text-secondary hover:text-red-500 transition-colors"
            title="Delete tag"
            @click="deleteTag(tag.id)"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

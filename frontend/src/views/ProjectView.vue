<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useProjectStore } from '@/stores/project'

const projectStore = useProjectStore()

const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')

onMounted(() => {
  projectStore.fetchProjects()
})

async function createProject() {
  if (!newName.value.trim()) return
  try {
    await projectStore.createProject(newName.value.trim(), newDesc.value.trim() || undefined)
    newName.value = ''
    newDesc.value = ''
    showCreate.value = false
  } catch {
    // error in store
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Projects</h1>
        <p class="text-sm text-text-secondary mt-1">{{ projectStore.projects.length }} projects</p>
      </div>
      <button class="if-btn-primary" @click="showCreate = true">+ New Project</button>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Create project</h3>
      <input v-model="newName" class="if-input" placeholder="Project name" />
      <textarea v-model="newDesc" class="if-input resize-none" rows="2" placeholder="Description (optional)" />
      <div class="flex justify-end gap-2">
        <button class="if-btn-ghost" @click="showCreate = false">Cancel</button>
        <button class="if-btn-primary" :disabled="!newName.trim()" @click="createProject">Create</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="projectStore.loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>

    <!-- Empty -->
    <div v-else-if="projectStore.projects.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">No projects yet</div>
        <div class="text-xs text-text-secondary">Create a project to organize your images.</div>
      </div>
    </div>

    <!-- Project grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="p in projectStore.projects"
        :key="p.id"
        class="if-card p-5 hover:shadow-md transition-shadow cursor-pointer"
        @click="$router.push(`/projects/${p.id}`)"
      >
        <h3 class="text-sm font-medium truncate">{{ p.name }}</h3>
        <p v-if="p.description" class="text-xs text-text-secondary mt-1 line-clamp-2">{{ p.description }}</p>
        <div class="text-xs text-text-secondary mt-3">{{ p.created_at }}</div>
      </div>
    </div>
  </div>
</template>

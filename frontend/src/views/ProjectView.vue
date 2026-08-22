<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
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
    <!-- Header -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-heading">{{ t('projects.title') }}</h1>
        <p class="text-body mt-1">{{ t('projects.projectsCount', { count: projectStore.projects.length }) }}</p>
      </div>
      <button class="btn btn-primary" @click="showCreate = true">{{ t('projects.newProject') }}</button>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="card p-5 space-y-4 animate-scale-in">
      <h3 class="text-subheading">{{ t('projects.createProject') }}</h3>
      <input v-model="newName" class="input" :placeholder="t('projects.projectNamePlaceholder')" />
      <textarea v-model="newDesc" class="input resize-none" rows="2" :placeholder="t('projects.projectDescPlaceholder')" />
      <div class="flex justify-end gap-2">
        <button class="btn btn-ghost" @click="showCreate = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!newName.trim()" @click="createProject">{{ t('common.create') }}</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="projectStore.loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 stagger">
      <div v-for="n in 6" :key="n" class="card p-5 space-y-3">
        <div class="skeleton h-4 w-2/3"></div>
        <div class="skeleton h-3 w-full"></div>
        <div class="skeleton h-3 w-1/2"></div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else-if="projectStore.projects.length === 0" class="card animate-fade-in">
      <div class="py-20 text-center">
        <div class="text-body mb-1">{{ t('projects.noProjectsYet') }}</div>
        <div class="text-caption">{{ t('projects.createProjectToOrganize') }}</div>
      </div>
    </div>

    <!-- Project grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 stagger">
      <div
        v-for="p in projectStore.projects"
        :key="p.id"
        class="card card-interactive p-5"
        @click="$router.push(`/projects/${p.id}`)"
      >
        <h3 class="text-sm font-medium truncate">{{ p.name }}</h3>
        <p v-if="p.description" class="text-caption mt-1 line-clamp-2">{{ p.description }}</p>
        <div class="text-caption mt-3">{{ p.created_at }}</div>
      </div>
    </div>
  </div>
</template>

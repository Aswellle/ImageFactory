<script setup lang="ts">
import type { PromptTemplate } from '@/types'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const props = defineProps<{
  template: PromptTemplate
}>()

const emit = defineEmits<{
  edit: [template: PromptTemplate]
  delete: [id: number]
  apply: [template: PromptTemplate]
}>()

// Derive a stable accent color from the category.
const categoryColor = computed(() => {
  switch (props.template.category) {
    case 'product':
      return 'bg-blue-500/15 text-blue-400'
    case 'scene':
      return 'bg-emerald-500/15 text-emerald-400'
    case 'style':
      return 'bg-purple-500/15 text-purple-400'
    default:
      return 'bg-surface-2 text-text-secondary'
  }
})

const variableList = computed(() => {
  if (!props.template.variables) return []
  return props.template.variables
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
})
</script>

<template>
  <div class="if-card p-5 flex flex-col gap-3 hover:shadow-md transition-shadow animate-fade-in">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <h3 class="text-sm font-medium truncate">{{ template.name }}</h3>
        <p v-if="template.description" class="text-xs text-text-secondary mt-1 line-clamp-2">
          {{ template.description }}
        </p>
      </div>
      <span
        class="shrink-0 text-[11px] font-medium px-2 py-0.5 rounded-full"
        :class="categoryColor"
      >
        {{ template.category }}
      </span>
    </div>

    <p class="text-xs text-text-secondary font-mono leading-relaxed line-clamp-3 bg-surface rounded-md px-3 py-2">
      {{ template.content }}
    </p>

    <div v-if="variableList.length" class="flex flex-wrap gap-1.5">
      <span
        v-for="v in variableList"
        :key="v"
        class="text-[11px] px-2 py-0.5 rounded-md bg-accent/10 text-accent font-mono"
        v-text="'{{' + v + '}}'"
      ></span>
    </div>

    <div class="flex items-center gap-2 pt-1">
      <button class="if-btn-primary text-xs px-3 py-1.5" @click="emit('apply', template)">
        {{ t('promptTemplates.usePrompt') }}
      </button>
      <button class="if-btn-ghost text-xs px-3 py-1.5" @click="emit('edit', template)">
        {{ t('promptTemplates.editTemplate') }}
      </button>
      <button
        class="if-btn-ghost text-xs px-3 py-1.5 text-red-400 hover:text-red-300 hover:bg-red-500/10 ml-auto"
        @click="emit('delete', template.id)"
      >
        {{ t('promptTemplates.deleteAction') }}
      </button>
    </div>
  </div>
</template>

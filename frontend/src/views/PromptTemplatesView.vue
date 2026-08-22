<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { usePromptTemplateStore } from '@/stores/promptTemplate'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { PromptTemplate } from '@/types'
import type { BuiltInTemplate } from '@/api/promptTemplate'
import PromptTemplateCard from '@/components/PromptTemplateCard.vue'
import PromptTemplateModal from '@/components/PromptTemplateModal.vue'

const { t } = useI18n()
const store = usePromptTemplateStore()
const router = useRouter()


const showModal = ref(false)
const editing = ref<PromptTemplate | null>(null)
const activeFilter = ref<string>('all')

// Apply flow state. Built-in templates carry variables as an array and use a
// negative placeholder ID; user templates use a CSV string and a real DB id.
type ApplyTarget = PromptTemplate | (BuiltInTemplate & { id: number; variables: string[] })
const applyTarget = ref<ApplyTarget | null>(null)
const applyInputs = ref<Record<string, string>>({})
const applyResult = ref<string | null>(null)
const applyLoading = ref(false)


onMounted(() => {
  store.fetchTemplates()
  store.fetchBuiltIn()
})

const variableList = computed(() => {
  const raw = applyTarget.value?.variables
  if (!raw) return []
  // Built-in templates carry variables as an array; user templates as CSV.
  if (Array.isArray(raw)) return raw.filter(Boolean)
  return raw.split(',').map((v) => v.trim()).filter(Boolean)
})

const filteredTemplates = computed(() => {
  if (activeFilter.value === 'all') return store.templates
  return store.templates.filter((t) => t.category === activeFilter.value)
})

const filterOptions = computed(() => {
  const cats = store.categories
  return [{ label: 'All', value: 'all' }, ...cats.map((c) => ({ label: c, value: c }))]
})

function openCreate() {
  editing.value = null
  showModal.value = true
}

function openEdit(t: PromptTemplate) {
  editing.value = t
  showModal.value = true
}

async function handleSubmit(input: {
  name: string
  description: string
  content: string
  variables: string[]
  category: string
  clear_variables?: boolean
}) {
  try {
    if (editing.value) {
      await store.updateTemplate(editing.value.id, input)
    } else {
      await store.createTemplate(input)
    }
    showModal.value = false
    editing.value = null
  } catch {
    // error surfaced via store.error
  }
}

async function handleDelete(id: number) {
  if (confirm('Delete this template? This cannot be undone.')) {
    await store.deleteTemplate(id)
  }
}

function openApply(t: ApplyTarget) {
  applyTarget.value = t
  applyInputs.value = {}
  applyResult.value = null
  for (const v of variableList.value) {
    applyInputs.value[v] = ''
  }
}

async function submitApply() {
  if (!applyTarget.value) return
  applyLoading.value = true
  applyResult.value = null
  try {
    const filled: Record<string, string> = {}
    for (const [k, v] of Object.entries(applyInputs.value)) {
      if (v.trim()) filled[k] = v.trim()
    }
    // Built-in templates (negative IDs) are not stored in the DB, so render client-side.
    if (applyTarget.value.id < 0) {
      applyResult.value = store.applyVariables(applyTarget.value.content, filled)
    } else {
      applyResult.value = await store.applyTemplate(applyTarget.value.id, filled)
    }
  } catch {
    // error surfaced via store.error
  } finally {
    applyLoading.value = false
  }
}

function useResult() {
  if (applyResult.value) {
    // Navigate to Create with the rendered prompt. GenerationView reads `prompt`.
    router.push({ name: 'create', query: { prompt: applyResult.value } })
  }
}
</script>

<template>
  <div class="space-y-8">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">{{ t('promptTemplates.title') }}</h1>
        <p class="text-sm text-text-secondary mt-1">
          {{ t('promptTemplates.subtitle') }}
        </p>
      </div>
      <button class="if-btn-primary" @click="openCreate">{{ t('promptTemplates.newTemplate') }}</button>
    </div>


    <!-- Built-in gallery -->
    <section v-if="store.builtIn.length">
      <h2 class="text-sm font-medium text-text-secondary mb-3 uppercase tracking-wide">
        {{ t('promptTemplates.builtInTemplates') }}
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="b in store.builtIn"
          :key="b.id"
          class="if-card p-5 flex flex-col gap-3 cursor-pointer hover:shadow-md transition-shadow"
          @click="openApply(b)"
        >
          <div class="flex items-start justify-between gap-2">
            <h3 class="text-sm font-medium">{{ b.name }}</h3>
            <span class="text-[11px] font-medium px-2 py-0.5 rounded-full bg-accent/10 text-accent">
              {{ b.category }}
            </span>
          </div>
          <p class="text-xs text-text-secondary">{{ b.description }}</p>
          <div v-if="b.variables.length" class="flex flex-wrap gap-1.5">
            <span
              v-for="v in b.variables"
              :key="v"
              class="text-[11px] px-2 py-0.5 rounded-md bg-surface-2 text-text-secondary font-mono"
            >
              {{ `{{${v}}}` }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- Filter + user templates -->
    <section>
      <div class="flex items-center gap-2 mb-3">
        <h2 class="text-sm font-medium text-text-secondary uppercase tracking-wide">
          {{ t('promptTemplates.yourTemplates') }}
        </h2>
        <div class="flex gap-1 ml-auto">
          <button
            v-for="opt in filterOptions"
            :key="opt.value"
            class="text-xs px-2.5 py-1 rounded-md transition-colors"
            :class="activeFilter === opt.value
              ? 'bg-accent text-white'
              : 'bg-surface-2 text-text-secondary hover:bg-surface-2/80'"
            @click="activeFilter = opt.value"
          >
            {{ opt.value === 'all' ? t('promptTemplates.filterAll') : opt.label }}
          </button>
        </div>
      </div>

      <div v-if="store.loading" class="text-center py-20 text-text-secondary text-sm">{{ t('common.loading') }}</div>

      <div v-else-if="filteredTemplates.length === 0" class="if-card">
        <div class="py-20 text-center">
          <div class="text-text-secondary text-sm mb-1">{{ t('promptTemplates.noTemplatesYet') }}</div>
          <div class="text-xs text-text-secondary">
            {{ t('promptTemplates.createOrTryBuiltIn') }}
          </div>
        </div>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <PromptTemplateCard
          v-for="t in filteredTemplates"
          :key="t.id"
          :template="t"
          @edit="openEdit"
          @delete="handleDelete"
          @apply="openApply"
        />
      </div>
    </section>

    <!-- Create / Edit modal -->
    <PromptTemplateModal
      :open="showModal"
      :template="editing"
      @close="showModal = false; editing = null"
      @submit="handleSubmit"
    />

    <!-- Apply drawer -->
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="applyTarget"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
        @click.self="applyTarget = null"
      >
        <div class="if-card w-full max-w-xl max-h-[90vh] overflow-y-auto p-6 space-y-5 animate-scale-in">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold tracking-tight">{{ applyTarget.name }}</h2>
              <p v-if="applyTarget.description" class="text-sm text-text-secondary mt-1">
                {{ applyTarget.description }}
              </p>
            </div>
            <button class="if-btn-ghost text-xs px-2 py-1" aria-label="Close" @click="applyTarget = null">
              ✕
            </button>
          </div>

          <div class="bg-surface rounded-md px-3 py-2.5">
            <p class="text-xs font-mono leading-relaxed text-text-secondary whitespace-pre-wrap">
              {{ applyTarget.content }}
            </p>
          </div>

          <div v-if="variableList.length" class="space-y-3">
            <div v-for="v in variableList" :key="v">
              <label class="if-label">{{ `{{${v}}}` }}</label>
              <input
                v-model="applyInputs[v]"
                class="if-input"
                :placeholder="t('promptTemplates.enterValue', { name: v })"
              />
            </div>
          </div>

          <div v-if="applyResult" class="space-y-2">
            <label class="if-label">{{ t('promptTemplates.result') }}</label>
            <p class="text-sm font-mono leading-relaxed bg-surface rounded-md px-3 py-2.5 whitespace-pre-wrap">
              {{ applyResult }}
            </p>
          </div>

          <div class="flex justify-end gap-2 pt-1">
            <button class="if-btn-ghost" @click="applyTarget = null">{{ t('promptTemplates.cancel') }}</button>
            <button
              v-if="applyResult"
              class="if-btn-primary"
              @click="useResult"
            >
              {{ t('promptTemplates.usePrompt') }}
            </button>
            <button
              class="if-btn-primary"
              :disabled="applyLoading"
              @click="submitApply"
            >
              {{ applyLoading ? t('promptTemplates.rendering') : t('promptTemplates.render') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

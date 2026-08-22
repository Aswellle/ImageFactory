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
    router.push({ name: 'create', query: { prompt: applyResult.value } })
  }
}
</script>

<template>
  <div class="page-container space-y-10">
    <!-- Header -->
    <div class="flex items-end justify-between">
      <div>
        <h1 class="text-heading text-[#1d1d1f] dark:text-white">{{ t('promptTemplates.title') }}</h1>
        <p class="text-body mt-1">{{ t('promptTemplates.subtitle') }}</p>
      </div>
      <button class="btn btn-primary" @click="openCreate">{{ t('promptTemplates.newTemplate') }}</button>
    </div>

    <!-- Built-in gallery -->
    <section v-if="store.builtIn.length">
      <h2 class="text-caption uppercase tracking-wide mb-4">
        {{ t('promptTemplates.builtInTemplates') }}
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="b in store.builtIn"
          :key="b.id"
          class="bento-tile flex flex-col gap-3 cursor-pointer card-interactive"
          @click="openApply(b)"
        >
          <div class="flex items-start justify-between gap-2">
            <h3 class="text-subheading text-[#1d1d1f] dark:text-white">{{ b.name }}</h3>
            <span class="badge badge-accent">{{ b.category }}</span>
          </div>
          <p class="text-body">{{ b.description }}</p>
          <div v-if="b.variables.length" class="flex flex-wrap gap-1.5">
            <span
              v-for="v in b.variables"
              :key="v"
              class="text-[11px] px-2 py-0.5 rounded-md bg-[var(--surface-2)] text-[var(--text-secondary)] text-mono"
            >
              {{ `{{${v}}}` }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- Filter + user templates -->
    <section>
      <div class="flex items-center gap-3 mb-4">
        <h2 class="text-caption uppercase tracking-wide">
          {{ t('promptTemplates.yourTemplates') }}
        </h2>
        <div class="flex gap-1 ml-auto">
          <button
            v-for="opt in filterOptions"
            :key="opt.value"
            class="text-xs px-3 py-1.5 rounded-lg transition-colors"
            :class="activeFilter === opt.value
              ? 'bg-[var(--accent)] text-white'
              : 'bg-[var(--surface-2)] text-[var(--text-secondary)] hover:opacity-80'"
            @click="activeFilter = opt.value"
          >
            {{ opt.value === 'all' ? t('promptTemplates.filterAll') : opt.label }}
          </button>
        </div>
      </div>

      <div v-if="store.loading" class="space-y-4">
        <div v-for="i in 3" :key="i" class="skeleton h-40 rounded-2xl" />
      </div>

      <div v-else-if="filteredTemplates.length === 0" class="bento-tile text-center py-20">
        <div class="text-body mb-1">{{ t('promptTemplates.noTemplatesYet') }}</div>
        <div class="text-caption">{{ t('promptTemplates.createOrTryBuiltIn') }}</div>
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
        <div class="surface w-full max-w-xl max-h-[90vh] overflow-y-auto rounded-2xl p-6 space-y-5 animate-scale-in">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-lg font-semibold tracking-tight text-[#1d1d1f] dark:text-white">{{ applyTarget.name }}</h2>
              <p v-if="applyTarget.description" class="text-body mt-1">
                {{ applyTarget.description }}
              </p>
            </div>
            <button class="btn btn-ghost !px-2 !py-1" aria-label="Close" @click="applyTarget = null">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12" /></svg>
            </button>
          </div>

          <div class="bg-[var(--surface-2)] rounded-lg px-4 py-3">
            <p class="text-xs text-mono leading-relaxed text-[var(--text-secondary)] whitespace-pre-wrap">
              {{ applyTarget.content }}
            </p>
          </div>

          <div v-if="variableList.length" class="space-y-4">
            <div v-for="v in variableList" :key="v">
              <label class="text-xs font-medium text-[var(--text-secondary)] mb-1.5 block">{{ `{{${v}}}` }}</label>
              <input
                v-model="applyInputs[v]"
                class="input"
                :placeholder="t('promptTemplates.enterValue', { name: v })"
              />
            </div>
          </div>

          <div v-if="applyResult" class="space-y-2">
            <label class="text-xs font-medium text-[var(--text-secondary)] mb-1.5 block">{{ t('promptTemplates.result') }}</label>
            <p class="text-sm text-mono leading-relaxed bg-[var(--surface-2)] rounded-lg px-4 py-3 whitespace-pre-wrap">
              {{ applyResult }}
            </p>
          </div>

          <div class="flex justify-end gap-3 pt-2">
            <button class="btn btn-ghost" @click="applyTarget = null">{{ t('promptTemplates.cancel') }}</button>
            <button
              v-if="applyResult"
              class="btn btn-primary"
              @click="useResult"
            >
              {{ t('promptTemplates.usePrompt') }}
            </button>
            <button
              class="btn btn-primary"
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

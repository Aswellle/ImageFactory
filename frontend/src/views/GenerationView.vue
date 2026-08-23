<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGenerationStore } from '@/stores/generation'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const store = useGenerationStore()

const prompt = ref('')
const negativePrompt = ref('')
const model = ref('gpt-image-1')
const size = ref('1024x1024')
const imageCount = ref(1)
const showAdvanced = ref(false)

const sizes = [
  { label: 'Square', value: '1024x1024' },
  { label: 'Landscape', value: '1792x1024' },
  { label: 'Portrait', value: '1024x1792' },
]

const models = [
  { label: 'GPT Image 1', value: 'gpt-image-1' },
  { label: 'DALL·E 3', value: 'dall-e-3' },
]

const charCount = computed(() => prompt.value.length)
const canSubmit = computed(() => prompt.value.trim().length > 0 && !store.loading)

async function submit() {
  if (!canSubmit.value) return
  try {
    const jobId = await store.generate({
      prompt: prompt.value.trim(),
      negative_prompt: negativePrompt.value.trim() || undefined,
      model: model.value,
      size: size.value,
      n: imageCount.value,
    })
    // Poll until done.
    await store.pollUntilDone(jobId)
  } catch {
    // error surfaced via store.error
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'pending': return t('generation.statusQueued')
    case 'processing': return t('generation.statusGenerating')
    case 'completed': return t('generation.statusCompleted')
    case 'failed': return t('generation.statusFailed')
    default: return status
  }
}
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-8">
    <!-- Header -->
    <header class="animate-fade-in">
    <h1 class="text-heading text-[var(--text)]">{{ t('generation.title') }}</h1>
      <p class="mt-2 text-body max-w-lg">{{ t('generation.subtitle') }}</p>
    </header>

    <!-- Form -->
    <form class="space-y-5" @submit.prevent="submit">
      <!-- Prompt -->
      <div>
        <div class="mb-2 flex items-center justify-between">
        <label class="text-sm font-medium text-[var(--text)]" for="prompt">{{ t('generation.promptLabel') }}</label>
          <span class="text-caption tabular-nums">{{ charCount }}</span>
        </div>
        <textarea
          id="prompt"
          v-model="prompt"
          rows="4"
          class="input resize-none"
          :placeholder="t('generation.promptPlaceholder')"
        />
      </div>

      <!-- Model & Size -->
      <div class="grid grid-cols-2 gap-4">
        <div>
        <label class="mb-2 block text-sm font-medium text-[var(--text)]" for="model">{{ t('generation.modelLabel') }}</label>
          <select id="model" v-model="model" class="input">
            <option v-for="m in models" :key="m.value" :value="m.value">{{ m.label }}</option>
          </select>
        </div>
        <div>
        <label class="mb-2 block text-sm font-medium text-[var(--text)]" for="size">{{ t('generation.sizeLabel') }}</label>
          <select id="size" v-model="size" class="input">
            <option v-for="s in sizes" :key="s.value" :value="s.value">{{ s.label }}</option>
          </select>
        </div>
      </div>

      <!-- Image count -->
      <div>
      <label class="mb-2 block text-sm font-medium text-[var(--text)]">{{ t('generation.numberOfImages') }}</label>
        <div class="flex gap-2">
          <button
            v-for="n in [1, 2, 4]"
            :key="n"
            type="button"
            class="btn flex-1"
            :class="imageCount === n ? 'btn-primary' : 'btn-ghost'"
            @click="imageCount = n"
          >
            {{ n }}
          </button>
        </div>
      </div>

      <!-- Advanced toggle -->
      <button type="button" class="text-sm text-[var(--text-secondary)] transition-colors hover:text-[var(--text)]" @click="showAdvanced = !showAdvanced">
        {{ showAdvanced ? t('generation.hideAdvanced') : t('generation.showAdvanced') }} {{ t('generation.advanced') }}
      </button>
      <div v-if="showAdvanced" class="bento-tile space-y-4 animate-scale-in">
        <div>
        <label class="mb-2 block text-sm font-medium text-[var(--text)]" for="negative">{{ t('generation.negativePromptLabel') }}</label>
          <input id="negative" v-model="negativePrompt" class="input" :placeholder="t('generation.negativePromptPlaceholder')" />
        </div>
      </div>

      <!-- Error -->
      <p v-if="store.error" class="text-sm text-danger animate-fade-in">{{ store.error }}</p>

      <!-- Submit -->
      <button type="submit" class="btn btn-primary w-full" :disabled="!canSubmit">
        <span v-if="store.loading" class="flex items-center justify-center gap-2">
          <svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {{ t('generation.generating') }}
        </span>
        <span v-else>{{ t('generation.generate') }}</span>
      </button>
    </form>

    <!-- Active job status -->
    <div v-if="store.activeJob" class="bento-tile space-y-3 animate-fade-in">
      <div class="flex items-center justify-between">
        <span class="text-sm font-medium">{{ t('generation.jobStatus', { id: store.activeJob.job_id.slice(0, 12) }) }}</span>
        <span
          class="badge"
          :class="{
            'badge-accent': store.activeJob.status === 'processing' || store.activeJob.status === 'pending',
            'badge-success': store.activeJob.status === 'completed',
            'badge-danger': store.activeJob.status === 'failed',
          }"
        >
          {{ statusLabel(store.activeJob.status) }}
        </span>
      </div>
      <p v-if="store.activeJob.status === 'failed'" class="text-sm text-danger">
        {{ store.activeJob.error_message || t('generation.generationFailed') }}
      </p>
      <p v-else-if="store.activeJob.status === 'completed'" class="text-sm text-[var(--text-secondary)]">
        {{ t('generation.imagesSaved') }}
      </p>
      <p v-else class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
        <span class="relative flex h-2 w-2">
          <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent opacity-60"></span>
          <span class="relative inline-flex h-2 w-2 rounded-full bg-accent"></span>
        </span>
        {{ t('generation.mayTakeAMinute') }}
      </p>
    </div>
  </div>
</template>

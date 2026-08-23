<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PromptTemplate } from '@/types'

const { t } = useI18n()

const props = defineProps<{
  open: boolean
  // When editing, the template being edited; when creating, null.
  template: PromptTemplate | null
}>()


const emit = defineEmits<{
  close: []
  submit: [input: {
    name: string
    description: string
    content: string
    variables: string[]
    category: string
    clear_variables?: boolean
  }]
}>()

const variableOptions = ['product', 'scene', 'style', 'lighting', 'requirements']

const categories = [
  { label: 'Product', value: 'product' },
  { label: 'Scene', value: 'scene' },
  { label: 'Style', value: 'style' },
  { label: 'Custom', value: 'custom' },
]

const name = ref('')
const description = ref('')
const content = ref('')
const category = ref('custom')
const selectedVariables = ref<string[]>([])

const isEditing = computed(() => !!props.template)
const title = computed(() => (isEditing.value ? t('promptTemplates.editTemplate') : t('promptTemplates.createTemplate')))
const canSubmit = computed(() => name.value.trim().length > 0 && content.value.trim().length > 0)

// Reset the form whenever the modal opens or the edited template changes.
watch(
  () => [props.open, props.template] as const,
  () => {
    if (!props.open) return
    name.value = props.template?.name ?? ''
    description.value = props.template?.description ?? ''
    content.value = props.template?.content ?? ''
    category.value = props.template?.category ?? 'custom'
    selectedVariables.value = props.template?.variables
      ? props.template.variables.split(',').map((v) => v.trim()).filter(Boolean)
      : []
  },
  { immediate: true },
)

function toggleVariable(v: string) {
  const idx = selectedVariables.value.indexOf(v)
  if (idx >= 0) selectedVariables.value.splice(idx, 1)
  else selectedVariables.value.push(v)
}


// Live preview: fill {{variables}} with sample placeholders.
const preview = computed(() => {
  let out = content.value
  for (const v of selectedVariables.value) {
    out = out.split(`{{${v}}}`).join(`[${v}]`)
  }
  return out
})

function submit() {
  if (!canSubmit.value) return
  emit('submit', {
    name: name.value.trim(),
    description: description.value.trim(),
    content: content.value,
    variables: selectedVariables.value,
    category: category.value,
    clear_variables: isEditing.value && selectedVariables.value.length === 0,
  })
}
</script>

<template>
  <Transition
    enter-active-class="transition duration-150 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-100 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
      @click.self="emit('close')"
    >
      <div class="if-card w-full max-w-2xl max-h-[90vh] overflow-y-auto p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold tracking-tight">{{ title }}</h2>
          <button
            class="if-btn-ghost text-xs px-2 py-1"
            aria-label="Close"
            @click="emit('close')"
          >
            ✕
          </button>
        </div>

        <div class="space-y-4">
          <!-- Name + category row -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="sm:col-span-2">
              <label class="if-label">{{ t('promptTemplates.name') }}</label>
              <input
                v-model="name"
                class="if-input"
                :placeholder="t('promptTemplates.namePlaceholder')"
              />
            </div>
            <div>
              <label class="if-label">{{ t('promptTemplates.category') }}</label>
              <select v-model="category" class="if-input">
                <option v-for="c in categories" :key="c.value" :value="c.value">
                  {{ c.label }}
                </option>
              </select>
            </div>
          </div>

          <div>
            <label class="if-label">{{ t('promptTemplates.description') }}</label>
            <input
              v-model="description"
              class="if-input"
              :placeholder="t('promptTemplates.descriptionPlaceholder')"
            />
          </div>

          <div>
            <label class="if-label">{{ t('promptTemplates.templateContent') }}</label>
            <textarea
              v-model="content"
              class="if-input resize-none font-mono text-xs leading-relaxed"
              rows="5"
              :placeholder="t('promptTemplates.templateContentPlaceholder')"
            />
            <p class="text-xs text-text-secondary mt-1.5">
              {{ t('promptTemplates.variableHint') }}
            </p>
          </div>

          <div>
            <label class="if-label">{{ t('promptTemplates.variables') }}</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="v in variableOptions"
                :key="v"
                type="button"
                class="text-xs px-2.5 py-1 rounded-md font-mono transition-colors"
                :class="selectedVariables.includes(v)
                  ? 'bg-accent text-white'
                  : 'bg-surface-2 text-text-secondary hover:bg-surface-2/80'"
                @click="toggleVariable(v)"
              >
                v-text="'{{' + v + '}}'"
              </button>
            </div>
          </div>


          <!-- Live preview -->
          <div v-if="preview">
            <label class="if-label">{{ t('promptTemplates.preview') }}</label>
            <p class="text-xs font-mono leading-relaxed bg-surface rounded-md px-3 py-2.5 text-text-secondary whitespace-pre-wrap">
              {{ preview }}
            </p>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-1">
          <button class="if-btn-ghost" @click="emit('close')">{{ t('promptTemplates.cancel') }}</button>
          <button class="if-btn-primary" :disabled="!canSubmit" @click="submit">
            {{ isEditing ? t('promptTemplates.saveChanges') : t('promptTemplates.createAction') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

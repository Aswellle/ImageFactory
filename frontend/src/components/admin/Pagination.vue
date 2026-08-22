<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'change', page: number, pageSize: number): void
}>()

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
)

const from = computed(() =>
  props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1,
)
const to = computed(() => Math.min(props.page * props.pageSize, props.total))

// Show a window around the current page.
const pages = computed<number[]>(() => {
  const max = totalPages.value
  const cur = props.page
  const window = 2
  const start = Math.max(1, cur - window)
  const end = Math.min(max, cur + window)
  const out: number[] = []
  for (let p = start; p <= end; p++) out.push(p)
  return out
})

const pageSizeOptions = [10, 20, 50, 100]

function goto(page: number) {
  if (page < 1 || page > totalPages.value || page === props.page) return
  emit('change', page, props.pageSize)
}

function setPageSize(size: number) {
  if (size === props.pageSize) return
  emit('change', 1, size)
}

function onChangeSize(event: Event) {
  setPageSize(Number((event.target as HTMLSelectElement).value))
}
</script>

<template>
  <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
    <div class="flex items-center gap-3 text-sm text-text-secondary">
      <span>{{ from }}–{{ to }} of {{ total }}</span>
      <select
        :value="pageSize"
        class="if-input !w-auto !py-1 !px-2 text-xs"
        @change="onChangeSize"
      >
        <option v-for="opt in pageSizeOptions" :key="opt" :value="opt">
          {{ opt }} / page
        </option>
      </select>
    </div>

    <div class="flex items-center gap-1">
      <button
        class="if-btn-ghost !px-2 !py-1 text-xs"
        :disabled="page <= 1"
        @click="goto(1)"
      >
        «
      </button>
      <button
        class="if-btn-ghost !px-2 !py-1 text-xs"
        :disabled="page <= 1"
        @click="goto(page - 1)"
      >
        ‹
      </button>
      <button
        v-for="p in pages"
        :key="p"
        class="if-btn-ghost !px-2.5 !py-1 text-xs min-w-[2rem]"
        :class="{ '!bg-accent !text-white hover:!bg-accent-hover': p === page }"
        @click="goto(p)"
      >
        {{ p }}
      </button>
      <button
        class="if-btn-ghost !px-2 !py-1 text-xs"
        :disabled="page >= totalPages"
        @click="goto(page + 1)"
      >
        ›
      </button>
      <button
        class="if-btn-ghost !px-2 !py-1 text-xs"
        :disabled="page >= totalPages"
        @click="goto(totalPages)"
      >
        »
      </button>
    </div>
  </div>
</template>

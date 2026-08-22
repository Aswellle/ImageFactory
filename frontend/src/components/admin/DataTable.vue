<script setup lang="ts" generic="T extends Record<string, any>">
import { computed } from 'vue'

export interface Column {
  key: string
  label: string
  sortable?: boolean
  width?: string
  align?: 'left' | 'center' | 'right'
}

const props = defineProps<{
  columns: Column[]
  rows: T[]
  rowKey: keyof T
  sortKey?: string
  sortAsc?: boolean
  selectable?: boolean
  selected?: Array<T[keyof T]>
  loading?: boolean
  emptyText?: string
}>()

const emit = defineEmits<{
  (e: 'sort', key: string): void
  (e: 'toggle-select', row: T): void
  (e: 'toggle-select-all'): void
  (e: 'row-click', row: T): void
}>()

const allSelected = computed(() => {
  if (!props.selectable || props.rows.length === 0) return false
  const sel = new Set(props.selected ?? [])
  return props.rows.every((r) => sel.has(r[props.rowKey]))
})

const colAlign = (c: Column) =>
  c.align === 'center' ? 'text-center' : c.align === 'right' ? 'text-right' : 'text-left'

function onSort(c: Column) {
  if (c.sortable) emit('sort', c.key)
}
</script>

<template>
  <div class="if-card overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-surface-2/50 text-text-secondary text-xs uppercase tracking-wide">
          <tr>
            <th v-if="selectable" class="w-10 px-3 py-3">
              <input
                type="checkbox"
                :checked="allSelected"
                class="accent-[var(--accent)]"
                @change="emit('toggle-select-all')"
              />
            </th>
            <th
              v-for="col in columns"
              :key="col.key"
              class="px-3 py-3 font-medium"
              :class="[colAlign(col), col.sortable ? 'cursor-pointer hover:text-text select-none' : '']"
              :style="col.width ? { width: col.width } : undefined"
              @click="onSort(col)"
            >
              <span class="inline-flex items-center gap-1">
                {{ col.label }}
                <template v-if="col.sortable">
                  <span class="text-[10px] opacity-60">
                    {{ sortKey === col.key ? (sortAsc ? '▲' : '▼') : '⇅' }}
                  </span>
                </template>
              </span>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          <tr v-if="loading">
            <td :colspan="columns.length + (selectable ? 1 : 0)" class="px-3 py-12 text-center text-text-secondary">
              <span class="inline-block animate-pulse">Loading…</span>
            </td>
          </tr>
          <tr v-else-if="rows.length === 0">
            <td :colspan="columns.length + (selectable ? 1 : 0)" class="px-3 py-12 text-center text-text-secondary">
              {{ emptyText ?? 'No data' }}
            </td>
          </tr>
          <tr
            v-for="row in rows"
            v-else
            :key="String(row[rowKey])"
            class="hover:bg-surface-2/30 transition-colors"
            :class="{ 'bg-accent/5': selectable && selected?.includes(row[rowKey]) }"
            @click="emit('row-click', row)"
          >
            <td v-if="selectable" class="px-3 py-3" @click.stop>
              <input
                type="checkbox"
                :checked="selected?.includes(row[rowKey]) ?? false"
                class="accent-[var(--accent)]"
                @change="emit('toggle-select', row)"
              />
            </td>
            <td
              v-for="col in columns"
              :key="col.key"
              class="px-3 py-3"
              :class="colAlign(col)"
            >
              <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">
                {{ row[col.key] ?? '—' }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

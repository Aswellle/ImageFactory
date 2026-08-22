<script setup lang="ts">
import type { Tag } from '@/api/tag'

const props = defineProps<{
  tag: Tag
  clickable?: boolean
}>()

defineEmits<{
  click: [tag: Tag]
}>()

const colorMap: Record<string, string> = {
  red: 'bg-red-100 text-red-800 border-red-200',
  blue: 'bg-blue-100 text-blue-800 border-blue-200',
  green: 'bg-green-100 text-green-800 border-green-200',
  yellow: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  purple: 'bg-purple-100 text-purple-800 border-purple-200',
  pink: 'bg-pink-100 text-pink-800 border-pink-200',
  indigo: 'bg-indigo-100 text-indigo-800 border-indigo-200',
  gray: 'bg-gray-100 text-gray-800 border-gray-200',
}

const defaultClass = 'bg-gray-100 text-gray-800 border-gray-200'
const colorClass = colorMap[props.tag.color?.toLowerCase() ?? ''] ?? defaultClass
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors',
      colorClass,
      clickable ? 'cursor-pointer hover:opacity-80' : '',
    ]"
    @click="clickable ? $emit('click', tag) : undefined"
  >
    <template v-if="tag.color">
      <span
        class="inline-block h-2 w-2 rounded-full"
        :style="{ backgroundColor: tag.color }"
      />
    </template>
    {{ tag.name }}
  </span>
</template>

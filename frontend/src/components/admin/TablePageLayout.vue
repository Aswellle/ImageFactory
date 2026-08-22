<script setup lang="ts">
// TablePageLayout is a thin wrapper that gives every admin list page the same
// vertical rhythm: a header row, an optional filter bar, the table, and a
// footer slot (usually Pagination). Content slots let each page fill in the
// specifics without reinventing the structure.
</script>

<template>
  <div class="space-y-5">
    <!-- Header: title + page-level actions -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <slot name="title" />
      </div>
      <div class="flex items-center gap-2">
        <slot name="actions" />
      </div>
    </div>

    <!-- Optional toolbar / filters -->
    <div v-if="$slots.filters" class="if-card p-4">
      <slot name="filters" />
    </div>

    <!-- Table region -->
    <slot name="table" />

    <!-- Footer: usually pagination -->
    <div v-if="$slots.footer" class="if-card p-3">
      <slot name="footer" />
    </div>
  </div>
</template>

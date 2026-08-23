<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/stores/admin'
import TablePageLayout from '@/components/admin/TablePageLayout.vue'
import DataTable from '@/components/admin/DataTable.vue'
import type { Column } from '@/components/admin/DataTable.vue'
import Pagination from '@/components/admin/Pagination.vue'
import type { AdminApiKey } from '@/api/admin'

const store = useAdminStore()
const { apiKeys, apiKeysTotal, apiKeysPage, apiKeysPageSize, apiKeysLoading } = storeToRefs(store)

const statusFilter = ref('')
const search = ref('')
const showRevoke = ref(false)
const target = ref<AdminApiKey | null>(null)

const columns: Column[] = [
  { key: 'id', label: 'ID', width: '60px' },
  { key: 'username', label: 'User' },
  { key: 'name', label: 'Name' },
  { key: 'key_prefix', label: 'Prefix', width: '140px' },
  { key: 'status', label: 'Status', width: '100px' },
  { key: 'last_used_at', label: 'Last used', width: '160px' },
  { key: 'created_at', label: 'Created', width: '160px' },
  { key: 'actions', label: 'Actions', width: '100px' },
]

async function load() {
  await store.loadApiKeys({
    page: apiKeysPage.value,
    page_size: apiKeysPageSize.value,
    status: statusFilter.value || undefined,
    search: search.value || undefined,
  })
}

function onPage(page: number, size: number) {
  apiKeysPage.value = page
  apiKeysPageSize.value = size
  load()
}

function applyFilters() {
  apiKeysPage.value = 1
  load()
}

function askRevoke(k: AdminApiKey) {
  target.value = k
  showRevoke.value = true
}

async function confirmRevoke() {
  if (!target.value) return
  await store.revokeApiKey(target.value.id)
  showRevoke.value = false
  target.value = null
}

function fmt(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleString()
}

watch([statusFilter, search], applyFilters)
onMounted(load)
</script>

<template>
  <TablePageLayout>
    <template #title>
      <h1 class="text-xl font-semibold tracking-tight">API Keys</h1>
      <p class="text-sm text-[var(--text-secondary)] mt-1">Manage programmatic access keys across all users.</p>
    </template>

    <template #filters>
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[200px] max-w-md">
          <input
            v-model="search"
            type="text"
            placeholder="Search by name, prefix, or user…"
            class="input pl-9"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] text-xs">⌕</span>
        </div>
        <select v-model="statusFilter" class="input !w-auto">
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="revoked">Revoked</option>
          <option value="expired">Expired</option>
        </select>
      </div>
    </template>

    <template #table>
<DataTable
  :columns="columns"
  :rows="apiKeys"
  row-key="id"
  :loading="apiKeysLoading"
  empty-text="No API keys found"
>
        <template #cell-username="{ value }">
          <span class="text-xs font-mono text-[var(--text-secondary)]">{{ value || '—' }}</span>
        </template>
        <template #cell-key_prefix="{ value }">
          <span class="font-mono text-xs">{{ value }}…</span>
        </template>
        <template #cell-status="{ value }">
          <span
            class="text-xs px-2 py-0.5 rounded-full"
            :class="{
              'bg-green-100 text-green-800': value === 'active',
              'bg-red-100 text-red-800': value === 'revoked',
              'bg-gray-100 text-gray-700': value === 'expired',
            }"
          >
            {{ value }}
          </span>
        </template>
        <template #cell-last_used_at="{ value }">
          <span class="text-xs text-[var(--text-secondary)]">{{ fmt(value) }}</span>
        </template>
        <template #cell-created_at="{ value }">
          <span class="text-xs text-[var(--text-secondary)]">{{ fmt(value) }}</span>
        </template>
<template #cell-actions="{ row: r }">
  <button
    class="btn btn-ghost !px-2 !py-1 text-xs text-danger hover:text-danger"
    :disabled="(r as AdminApiKey).status === 'revoked'"
    @click.stop="askRevoke(r as AdminApiKey)"
  >
    {{ (r as AdminApiKey).status === 'revoked' ? 'Revoked' : 'Revoke' }}
  </button>
</template>
      </DataTable>
    </template>

    <template #footer>
      <Pagination
        :page="apiKeysPage"
        :page-size="apiKeysPageSize"
        :total="apiKeysTotal"
        @change="onPage"
      />
    </template>
  </TablePageLayout>

  <!-- Revoke confirmation -->
  <div v-if="showRevoke" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 animate-fade-in">
    <div class="surface rounded-2xl p-6 w-full max-w-sm animate-scale-in">
      <h3 class="text-base font-semibold">Revoke API key?</h3>
      <p class="text-sm text-[var(--text-secondary)] mt-2">
        This will immediately revoke the key <span class="font-mono">{{ target?.key_prefix }}…</span>
        belonging to {{ target?.username ?? 'unknown' }}. This cannot be undone.
      </p>
      <div class="flex justify-end gap-2 mt-5">
        <button class="if-btn-ghost" @click="showRevoke = false">Cancel</button>
        <button class="btn btn-primary !bg-red-500 hover:!bg-red-600" @click="confirmRevoke">Revoke</button>
      </div>
    </div>
  </div>
</template>

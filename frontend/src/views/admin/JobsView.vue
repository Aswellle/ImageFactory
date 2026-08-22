<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/stores/admin'
import TablePageLayout from '@/components/admin/TablePageLayout.vue'
import DataTable from '@/components/admin/DataTable.vue'
import type { Column } from '@/components/admin/DataTable.vue'
import Pagination from '@/components/admin/Pagination.vue'
import type { AdminJob } from '@/api/admin'

const store = useAdminStore()
const { jobs, jobsTotal, jobsPage, jobsPageSize, jobsLoading } = storeToRefs(store)

const statusFilter = ref('')
const typeFilter = ref('')
const search = ref('')

const sortKey = ref<string>('created_at')
const sortAsc = ref(false)

const columns: Column[] = [
  { key: 'id', label: 'Job ID', sortable: false, width: '180px' },
  { key: 'username', label: 'User', sortable: true },
  { key: 'type', label: 'Type', sortable: true, width: '90px' },
  { key: 'status', label: 'Status', sortable: true, width: '110px' },
  { key: 'provider', label: 'Provider', sortable: true, width: '110px' },
  { key: 'model', label: 'Model', sortable: true, width: '140px' },
  { key: 'image_count', label: 'Images', sortable: true, width: '80px' },
  { key: 'retry_count', label: 'Retries', sortable: true, width: '80px' },
  { key: 'created_at', label: 'Created', sortable: true, width: '160px' },
  { key: 'actions', label: 'Actions', sortable: false, width: '100px' },
]

function clientSort(rows: AdminJob[]): AdminJob[] {
  const key = sortKey.value
  const dir = sortAsc.value ? 1 : -1
  return [...rows].sort((a, b) => {
    const av = (a as any)[key] ?? ''
    const bv = (b as any)[key] ?? ''
    return String(av).localeCompare(String(bv)) * dir
  })
}

async function load() {
  await store.loadJobs({
    page: jobsPage.value,
    page_size: jobsPageSize.value,
    status: statusFilter.value || undefined,
    type: typeFilter.value || undefined,
    search: search.value || undefined,
  })
}

function onSort(key: string) {
  if (sortKey.value === key) sortAsc.value = !sortAsc.value
  else {
    sortKey.value = key
    sortAsc.value = true
  }
}

function onPage(page: number, size: number) {
  jobsPage.value = page
  jobsPageSize.value = size
  load()
}

function applyFilters() {
  jobsPage.value = 1
  load()
}

async function onRetry(job: AdminJob) {
  await store.retryJob(job.id)
  await load()
}

function fmt(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleString()
}

watch([statusFilter, typeFilter, search], applyFilters)
onMounted(load)
</script>

<template>
  <TablePageLayout>
    <template #title>
      <h1 class="text-xl font-semibold tracking-tight">Jobs</h1>
      <p class="text-sm text-text-secondary mt-1">Monitor image generation and edit jobs across all users.</p>
    </template>

    <template #filters>
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[200px] max-w-md">
          <input
            v-model="search"
            type="text"
            placeholder="Search by prompt, model, or ID…"
            class="if-input pl-9"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary text-xs">⌕</span>
        </div>
        <select v-model="statusFilter" class="if-input !w-auto">
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="processing">Processing</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select v-model="typeFilter" class="if-input !w-auto">
          <option value="">All types</option>
          <option value="generation">Generation</option>
          <option value="edit">Edit</option>
        </select>
      </div>
    </template>

    <template #table>
      <DataTable
        :columns="columns"
        :rows="clientSort(jobs)"
        row-key="id"
        :sort-key="sortKey"
        :sort-asc="sortAsc"
        :loading="jobsLoading"
        empty-text="No jobs found"
        @sort="onSort"
      >
        <template #cell-id="{ value }">
          <span class="font-mono text-xs">{{ value }}</span>
        </template>
        <template #cell-username="{ value }">
          <span class="text-xs text-text-secondary">{{ value || '—' }}</span>
        </template>
        <template #cell-type="{ value }">
          <span class="text-xs px-2 py-0.5 rounded-full bg-surface-2">{{ value }}</span>
        </template>
        <template #cell-status="{ value }">
          <span
            class="text-xs px-2 py-0.5 rounded-full"
            :class="{
              'bg-green-100 text-green-800': value === 'completed',
              'bg-red-100 text-red-800': value === 'failed',
              'bg-blue-100 text-blue-800': value === 'processing',
              'bg-yellow-100 text-yellow-800': value === 'pending',
              'bg-gray-100 text-gray-700': value === 'cancelled',
            }"
          >
            {{ value }}
          </span>
        </template>
        <template #cell-provider="{ value }">
          <span class="text-xs">{{ value || '—' }}</span>
        </template>
        <template #cell-model="{ value }">
          <span class="text-xs font-mono">{{ value || '—' }}</span>
        </template>
        <template #cell-image_count="{ value }">
          <span class="text-xs tabular-nums">{{ value ?? 0 }}</span>
        </template>
        <template #cell-retry_count="{ value }">
          <span class="text-xs tabular-nums">{{ value ?? 0 }}</span>
        </template>
        <template #cell-created_at="{ value }">
          <span class="text-xs text-text-secondary">{{ fmt(value) }}</span>
        </template>
        <template #cell-actions="{ row }">
          <button
            class="if-btn-ghost !px-2 !py-1 text-xs"
            :disabled="row.status !== 'failed'"
            title="Retry failed job"
            @click.stop="onRetry(row)"
          >
            Retry
          </button>
        </template>
      </DataTable>
    </template>

    <template #footer>
      <Pagination
        :page="jobsPage"
        :page-size="jobsPageSize"
        :total="jobsTotal"
        @change="onPage"
      />
    </template>
  </TablePageLayout>
</template>

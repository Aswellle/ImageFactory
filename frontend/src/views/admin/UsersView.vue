<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/stores/admin'
import TablePageLayout from '@/components/admin/TablePageLayout.vue'
import DataTable from '@/components/admin/DataTable.vue'
import type { Column } from '@/components/admin/DataTable.vue'
import Pagination from '@/components/admin/Pagination.vue'
import type { AdminUser } from '@/api/admin'

const store = useAdminStore()
const { users, usersTotal, usersPage, usersPageSize, usersLoading } = storeToRefs(store)

const search = ref('')
const roleFilter = ref('')
const statusFilter = ref('')

const sortKey = ref<string>('created_at')
const sortAsc = ref(false)

const selected = ref<number[]>([])
const showSuspend = ref(false)
const target = ref<AdminUser | null>(null)
const pendingStatus = ref<'active' | 'suspended'>('active')

const columns: Column[] = [
  { key: 'id', label: 'ID', sortable: true, width: '60px' },
  { key: 'email', label: 'Email', sortable: true },
  { key: 'name', label: 'Name', sortable: true },
  { key: 'role', label: 'Role', sortable: true, width: '90px' },
  { key: 'status', label: 'Status', sortable: true, width: '110px' },
  { key: 'last_login_at', label: 'Last login', sortable: true, width: '160px' },
  { key: 'created_at', label: 'Created', sortable: true, width: '160px' },
  { key: 'actions', label: 'Actions', sortable: false, width: '120px' },
]

function clientSort(rows: AdminUser[]): AdminUser[] {
  const key = sortKey.value
  const dir = sortAsc.value ? 1 : -1
  return [...rows].sort((a, b) => {
    const av = (a as any)[key] ?? ''
    const bv = (b as any)[key] ?? ''
    return av.localeCompare(bv) * dir
  })
}

async function load() {
  await store.loadUsers({
    page: usersPage.value,
    page_size: usersPageSize.value,
    search: search.value || undefined,
    role: roleFilter.value || undefined,
    status: statusFilter.value || undefined,
  })
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
function onSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    usersPage.value = 1
    load()
  }, 300)
}

function applyFilters() {
  usersPage.value = 1
  load()
}

function onSort(key: string) {
  if (sortKey.value === key) sortAsc.value = !sortAsc.value
  else {
    sortKey.value = key
    sortAsc.value = true
  }
}

function onPage(page: number, size: number) {
  usersPage.value = page
  usersPageSize.value = size
  load()
}

function toggleSelect(row: AdminUser) {
  const idx = selected.value.indexOf(row.id)
  if (idx >= 0) selected.value.splice(idx, 1)
  else selected.value.push(row.id)
}

function toggleSelectAll() {
  if (selected.value.length === users.value.length) selected.value = []
  else selected.value = users.value.map((u) => u.id)
}

function askToggleSuspend(u: AdminUser) {
  target.value = u
  pendingStatus.value = u.status === 'suspended' ? 'active' : 'suspended'
  showSuspend.value = true
}

async function confirmSuspend() {
  if (!target.value) return
  await store.setUserStatus(target.value.id, pendingStatus.value)
  showSuspend.value = false
  target.value = null
}

function fmt(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleDateString()
}

watch([roleFilter, statusFilter], applyFilters)
onMounted(load)
</script>

<template>
  <TablePageLayout>
    <template #title>
      <h1 class="text-xl font-semibold tracking-tight">Users</h1>
      <p class="text-sm text-[var(--text-secondary)] mt-1">Manage user accounts and access.</p>
    </template>

    <template #filters>
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[200px] max-w-md">
          <input
            v-model="search"
            type="text"
            placeholder="Search by email or name…"
            class="input pl-9"
            @input="onSearch"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] text-xs">⌕</span>
        </div>
        <select v-model="roleFilter" class="input !w-auto">
          <option value="">All roles</option>
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
        <select v-model="statusFilter" class="input !w-auto">
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="suspended">Suspended</option>
          <option value="deleted">Deleted</option>
        </select>
      </div>
    </template>

    <template #table>
      <DataTable
        :columns="columns"
        :rows="clientSort(users)"
        row-key="id"
        :sort-key="sortKey"
        :sort-asc="sortAsc"
        :selectable="true"
        :selected="selected"
        :loading="usersLoading"
        empty-text="No users found"
        @sort="onSort"
        @toggle-select="toggleSelect"
        @toggle-select-all="toggleSelectAll"
      >
        <template #cell-email="{ row }">
          <span class="font-mono text-xs">{{ row.email }}</span>
        </template>
        <template #cell-name="{ value }">
          <span class="text-[var(--text-secondary)]">{{ value || '—' }}</span>
        </template>
        <template #cell-role="{ value }">
          <span
            class="text-xs px-2 py-0.5 rounded-full font-medium"
            :class="value === 'admin' ? 'bg-[var(--accent-soft)] text-[var(--accent)]' : 'bg-[var(--surface-2)] text-[var(--text-secondary)]'"
          >
            {{ value }}
          </span>
        </template>
        <template #cell-status="{ value }">
          <span
            class="text-xs px-2 py-0.5 rounded-full"
            :class="{
              'bg-green-100 text-green-800': value === 'active',
              'bg-yellow-100 text-yellow-800': value === 'suspended',
              'bg-red-100 text-red-800': value === 'deleted',
            }"
          >
            {{ value }}
          </span>
        </template>
        <template #cell-last_login_at="{ value }">
          <span class="text-xs text-[var(--text-secondary)]">{{ fmt(value) }}</span>
        </template>
        <template #cell-created_at="{ value }">
          <span class="text-xs text-[var(--text-secondary)]">{{ fmt(value) }}</span>
        </template>
        <template #cell-actions="{ row }">
          <button
            class="btn btn-ghost !px-2 !py-1 text-xs"
            :disabled="row.status === 'deleted'"
            @click.stop="askToggleSuspend(row)"
          >
            {{ row.status === 'suspended' ? 'Restore' : 'Suspend' }}
          </button>
        </template>
      </DataTable>
    </template>

    <template #footer>
      <Pagination
        :page="usersPage"
        :page-size="usersPageSize"
        :total="usersTotal"
        @change="onPage"
      />
    </template>
  </TablePageLayout>

  <!-- Status change confirmation -->
  <div v-if="showSuspend" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 animate-fade-in">
    <div class="surface rounded-2xl p-6 w-full max-w-sm animate-scale-in">
      <h3 class="text-base font-semibold">
        {{ pendingStatus === 'suspended' ? 'Suspend' : 'Restore' }} user?
      </h3>
      <p class="text-sm text-[var(--text-secondary)] mt-2">
        This will set <span class="font-mono">{{ target?.email }}</span> to
        <span class="font-medium">{{ pendingStatus }}</span>.
      </p>
      <div class="flex justify-end gap-2 mt-5">
        <button class="if-btn-ghost" @click="showSuspend = false">Cancel</button>
        <button
          class="if-btn-primary"
          :class="pendingStatus === 'suspended' ? '!bg-yellow-500 hover:!bg-yellow-600' : ''"
          @click="confirmSuspend"
        >
          Confirm
        </button>
      </div>
    </div>
  </div>
</template>

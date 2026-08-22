<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/stores/admin'
import { useI18n } from 'vue-i18n'
import StatCard from '@/components/admin/StatCard.vue'

const store = useAdminStore()
const { t } = useI18n()
const { stats, activity, health, dashboardLoading } = storeToRefs(store)
onMounted(() => {
  store.loadDashboard()
})

function fmt(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleTimeString()
}
</script>

<template>
  <div class="space-y-8">
    <!-- Health -->
    <div
      v-if="health"
      class="flex items-center gap-2.5 text-sm px-4 py-2.5 rounded-xl"
      :class="health.database === 'ok' ? 'text-[var(--success)] bg-[color-mix(in_srgb,var(--success)_8%,transparent)]' : 'text-[var(--danger)] bg-[color-mix(in_srgb,var(--danger)_8%,transparent)]'"
    >
      <span class="w-2 h-2 rounded-full" :class="health.database === 'ok' ? 'bg-[var(--success)]' : 'bg-[var(--danger)]'"></span>
      Database: {{ health.database }}
    </div>

    <!-- Stats grid -->
    <section class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard label="Total users" :value="stats?.total_users ?? 0" tone="accent" />
      <StatCard label="Active users" :value="stats?.active_users ?? 0" tone="success" />
      <StatCard label="Total jobs" :value="stats?.total_jobs ?? 0" />
      <StatCard label="API keys" :value="stats?.total_api_keys ?? 0" />
    </section>

    <!-- Outcome breakdown -->
    <section v-if="stats" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard label="Completed" :value="stats.completed_jobs" tone="success" />
      <StatCard label="Failed" :value="stats.failed_jobs" :tone="stats.failed_jobs > 0 ? 'danger' : 'default'" />
      <StatCard label="Jobs (24h)" :value="stats.jobs_last_24h" />
      <StatCard label="New users (24h)" :value="stats.users_last_24h" tone="accent" />
    </section>

    <!-- Recent activity -->
    <section class="surface rounded-2xl overflow-hidden">
      <div class="px-6 py-4 border-b border-[var(--border)] flex items-center justify-between">
        <h2 class="text-sm font-medium text-[var(--text)]">Recent activity</h2>
        <span class="text-caption">{{ activity.length }} events</span>
      </div>
      <div v-if="dashboardLoading" class="p-6 text-center text-[var(--text-secondary)] text-sm">{{ t('common.loading') }}</div>
      <div v-else-if="activity.length === 0" class="p-6 text-center text-[var(--text-secondary)] text-sm">
        No recent activity
      </div>
      <div v-else class="divide-y divide-[var(--border-subtle)] max-h-80 overflow-y-auto">
        <div
          v-for="(ev, i) in activity"
          :key="i"
          class="flex items-center justify-between px-6 py-3.5 text-sm"
        >
          <div class="min-w-0">
            <div class="truncate text-mono text-xs text-[var(--text-muted)]">{{ ev.job_id ?? ev.type }}</div>
            <div class="truncate text-[var(--text)]">{{ ev.prompt || ev.provider || ev.model || '—' }}</div>
          </div>
          <div class="flex items-center gap-4 shrink-0 ml-4">
            <span
              v-if="ev.status"
              class="badge"
              :class="{
                'badge-accent': ev.status === 'completed',
                '': ev.status === 'failed',
              }"
              :style="ev.status === 'failed' ? 'color: var(--danger)' : ev.status === 'processing' || ev.status === 'pending' ? 'color: var(--warning)' : ''"
            >
              {{ ev.status }}
            </span>
            <span class="text-caption text-mono">{{ fmt(ev.created_at) }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

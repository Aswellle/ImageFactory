<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAdminStore } from '@/stores/admin'
import StatCard from '@/components/admin/StatCard.vue'

const store = useAdminStore()
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
  <div class="space-y-6">
    <!-- Health -->
    <section
      v-if="health"
      class="if-card px-4 py-2.5 flex items-center gap-2 text-sm"
      :class="health.database === 'ok' ? 'text-success' : 'text-danger'"
    >
      <span class="w-2 h-2 rounded-full" :class="health.database === 'ok' ? 'bg-success' : 'bg-danger'"></span>
      Database: {{ health.database }}
    </section>

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
    <section class="if-card overflow-hidden">
      <div class="px-5 py-4 border-b border-border flex items-center justify-between">
        <h2 class="text-sm font-medium">Recent activity</h2>
        <span class="text-xs text-text-secondary">{{ activity.length }} events</span>
      </div>
      <div v-if="dashboardLoading" class="p-5 text-center text-text-secondary text-sm animate-pulse">Loading…</div>
      <div v-else-if="activity.length === 0" class="p-5 text-center text-text-secondary text-sm">
        No recent activity
      </div>
      <div v-else class="divide-y divide-border max-h-80 overflow-y-auto">
        <div
          v-for="(ev, i) in activity"
          :key="i"
          class="flex items-center justify-between px-5 py-3 text-sm"
        >
          <div class="min-w-0">
            <div class="truncate font-mono text-xs text-text-secondary">{{ ev.job_id ?? ev.type }}</div>
            <div class="truncate">{{ ev.prompt || ev.provider || ev.model || '—' }}</div>
          </div>
          <div class="flex items-center gap-3 shrink-0 ml-3">
            <span
              v-if="ev.status"
              class="text-xs px-2 py-0.5 rounded-full"
              :class="{
                'bg-green-100 text-green-800': ev.status === 'completed',
                'bg-red-100 text-red-800': ev.status === 'failed',
                'bg-yellow-100 text-yellow-800': ev.status === 'processing' || ev.status === 'pending',
                'bg-gray-100 text-gray-700': ev.status === 'cancelled',
              }"
            >
              {{ ev.status }}
            </span>
            <span class="text-xs text-text-secondary">{{ fmt(ev.created_at) }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

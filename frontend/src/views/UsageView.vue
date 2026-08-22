<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useUsageStore } from '@/stores/usage'

// UsageView renders the authenticated user's API usage dashboard: headline
// stat cards for the trailing 30-day period, a per-day activity chart, and a
// period selector that refetches history.
const store = useUsageStore()
const { summary, history, loading, error } = storeToRefs(store)

const period = ref(30)

onMounted(() => {
  store.init(period.value)
})

async function changePeriod(days: number) {
  period.value = days
  await store.fetchHistory(days)
}

// Chart scaling: bar height is relative to the busiest day so an empty period
// still renders a usable axis.
const maxRequests = computed(() =>
  history.value.reduce((max, d) => Math.max(max, d.requests), 0),
)

function requestsHeight(d: { requests: number }) {
  const peak = maxRequests.value || 1
  return Math.round((d.requests / peak) * 100)
}

function formatCost(cost: number) {
  return cost.toFixed(4)
}

function formatNumber(n: number) {
  return n.toLocaleString()
}
</script>

<template>
  <div class="space-y-8">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Usage</h1>
        <span
          v-if="summary"
          class="text-xs text-text-secondary"
        >
          {{ summary.period_start }} — {{ summary.period_end }}
        </span>
        <p v-else class="text-sm text-text-secondary">Your API usage over time.</p>
      </div>

      <div class="flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          v-for="d in [7, 30, 90]"
          :key="d"
          class="px-3 py-1 text-xs font-medium rounded transition-colors"
          :class="period === d ? 'bg-accent text-white' : 'text-text-secondary hover:bg-surface-2'"
          @click="changePeriod(d)"
        >
          {{ d }}d
        </button>
      </div>
    </header>

    <!-- Stat cards -->
    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Requests</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">
          {{ summary ? formatNumber(summary.total_requests) : '—' }}
        </div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Images</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">
          {{ summary ? formatNumber(summary.total_images) : '—' }}
        </div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Tokens</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">
          {{ summary ? formatNumber(summary.total_tokens) : '—' }}
        </div>
      </div>
      <div class="if-card p-5">
        <div class="text-xs text-text-secondary uppercase tracking-wide">Cost</div>
        <div class="text-2xl font-semibold mt-1.5 tabular-nums">
          {{ summary ? formatCost(summary.total_cost) : '—' }}
        </div>
      </div>
    </section>

    <!-- Daily activity chart -->
    <section class="if-card p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base font-medium">Daily requests</h2>
        <span v-if="loading" class="text-xs text-text-secondary">Loading…</span>
      </div>

      <div v-if="error" class="text-sm text-red-500">{{ error }}</div>

      <div
        v-else-if="history.length"
        class="flex items-end gap-1 h-40"
        role="img"
        :aria-label="`Daily request usage for the last ${period} days`"
      >
        <div
          v-for="d in history"
          :key="d.date"
          class="flex-1 flex flex-col items-end justify-end group"
        >
          <div class="w-full flex justify-center">
            <div
              class="w-full max-w-[18px] rounded-t bg-accent/80 transition-all group-hover:bg-accent"
              :style="{ height: requestsHeight(d) + '%' }"
              :title="`${d.date}: ${d.requests} requests`"
            ></div>
          </div>
        </div>
      </div>

      <div v-else-if="!loading" class="text-sm text-text-secondary py-10 text-center">
        No usage recorded in this period yet.
      </div>

      <!-- X-axis labels: show first, middle, last to avoid clutter -->
      <div v-if="history.length" class="flex justify-between mt-3 text-[10px] text-text-secondary">
        <span>{{ history[0]?.date }}</span>
        <span>{{ history[Math.floor(history.length / 2)]?.date }}</span>
        <span>{{ history[history.length - 1]?.date }}</span>
      </div>
    </section>

    <!-- Daily breakdown table -->
    <section v-if="history.length" class="if-card overflow-hidden">
      <table class="w-full text-sm">
        <thead class="text-left text-xs uppercase tracking-wide text-text-secondary border-b border-border">
          <tr>
            <th class="px-5 py-3 font-medium">Date</th>
            <th class="px-5 py-3 font-medium text-right">Requests</th>
            <th class="px-5 py-3 font-medium text-right">Images</th>
            <th class="px-5 py-3 font-medium text-right">Tokens</th>
            <th class="px-5 py-3 font-medium text-right">Cost</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="d in [...history].reverse()"
            :key="d.date"
            class="border-b border-border last:border-0 hover:bg-surface-2/50"
          >
            <td class="px-5 py-2.5 tabular-nums">{{ d.date }}</td>
            <td class="px-5 py-2.5 text-right tabular-nums">{{ formatNumber(d.requests) }}</td>
            <td class="px-5 py-2.5 text-right tabular-nums">{{ formatNumber(d.images) }}</td>
            <td class="px-5 py-2.5 text-right tabular-nums">{{ formatNumber(d.tokens) }}</td>
            <td class="px-5 py-2.5 text-right tabular-nums">{{ formatCost(d.cost) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useUsageStore } from '@/stores/usage'

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
  <div class="page-container space-y-8">
    <!-- Header -->
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-heading text-[#1d1d1f] dark:text-white">Usage</h1>
        <span v-if="summary" class="text-caption">
          {{ summary.period_start }} — {{ summary.period_end }}
        </span>
        <p v-else class="text-body">Your API usage over time.</p>
      </div>

      <!-- Period selector -->
      <div class="flex items-center gap-1 rounded-lg border border-[var(--border)] p-1">
        <button
          v-for="d in [7, 30, 90]"
          :key="d"
          class="px-3 py-1.5 text-xs font-medium rounded-md transition-colors"
          :class="period === d ? 'bg-[var(--accent)] text-white' : 'text-[var(--text-secondary)] hover:bg-[var(--surface-2)]'"
          @click="changePeriod(d)"
        >
          {{ d }}d
        </button>
      </div>
    </header>

    <!-- Stat cards -->
    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">Requests</div>
        <div class="metric-value mt-2 text-[#1d1d1f] dark:text-white">
          {{ summary ? formatNumber(summary.total_requests) : '—' }}
        </div>
      </div>
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">Images</div>
        <div class="metric-value mt-2 text-[#1d1d1f] dark:text-white">
          {{ summary ? formatNumber(summary.total_images) : '—' }}
        </div>
      </div>
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">Tokens</div>
        <div class="metric-value mt-2 text-[#1d1d1f] dark:text-white">
          {{ summary ? formatNumber(summary.total_tokens) : '—' }}
        </div>
      </div>
      <div class="bento-tile">
        <div class="text-caption uppercase tracking-wide">Cost</div>
        <div class="metric-value mt-2 text-[#1d1d1f] dark:text-white">
          {{ summary ? formatCost(summary.total_cost) : '—' }}
        </div>
      </div>
    </section>

    <!-- Daily activity chart -->
    <section class="bento-tile">
      <div class="flex items-center justify-between mb-5">
        <h2 class="text-subheading text-[#1d1d1f] dark:text-white">Daily requests</h2>
        <span v-if="loading" class="text-caption">Loading…</span>
      </div>

      <div v-if="error" class="text-sm text-[var(--danger)]">{{ error }}</div>

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
              class="w-full max-w-[18px] rounded-t bg-[var(--accent)]/80 transition-all group-hover:bg-[var(--accent)]"
              :style="{ height: requestsHeight(d) + '%' }"
              :title="`${d.date}: ${d.requests} requests`"
            ></div>
          </div>
        </div>
      </div>

      <div v-else-if="!loading" class="text-body py-10 text-center">
        No usage recorded in this period yet.
      </div>

      <!-- X-axis labels -->
      <div v-if="history.length" class="flex justify-between mt-4 text-[10px] text-[var(--text-muted)]">
        <span>{{ history[0]?.date }}</span>
        <span>{{ history[Math.floor(history.length / 2)]?.date }}</span>
        <span>{{ history[history.length - 1]?.date }}</span>
      </div>
    </section>

    <!-- Daily breakdown table -->
    <section v-if="history.length" class="surface rounded-2xl overflow-hidden">
      <table class="w-full text-sm">
        <thead class="text-left text-xs uppercase tracking-wide text-[var(--text-secondary)] border-b border-[var(--border)]">
          <tr>
            <th class="px-6 py-3.5 font-medium">Date</th>
            <th class="px-6 py-3.5 font-medium text-right">Requests</th>
            <th class="px-6 py-3.5 font-medium text-right">Images</th>
            <th class="px-6 py-3.5 font-medium text-right">Tokens</th>
            <th class="px-6 py-3.5 font-medium text-right">Cost</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(d, idx) in [...history].reverse()"
            :key="d.date"
            :class="idx > 0 ? 'border-t border-[var(--border-subtle)]' : ''"
            class="hover:bg-[var(--surface-2)]/50 transition-colors"
          >
            <td class="px-6 py-3 text-mono">{{ d.date }}</td>
            <td class="px-6 py-3 text-right text-mono">{{ formatNumber(d.requests) }}</td>
            <td class="px-6 py-3 text-right text-mono">{{ formatNumber(d.images) }}</td>
            <td class="px-6 py-3 text-right text-mono">{{ formatNumber(d.tokens) }}</td>
            <td class="px-6 py-3 text-right text-mono">{{ formatCost(d.cost) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

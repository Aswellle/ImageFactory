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

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="studio-theme">
    <div class="usage-page">
      <!-- Header -->
      <div class="usage-header">
        <div>
          <h1 class="usage-title">Usage</h1>
          <p style="font-size: 13px; color: var(--studio-text-secondary); margin-top: 4px;">Track your API usage and costs</p>
        </div>
        <div class="usage-period">
          <button
            v-for="d in [7, 30, 90]"
            :key="d"
            class="usage-period-btn"
            :class="{ active: period === d }"
            @click="changePeriod(d)"
          >
            {{ d }}d
          </button>
        </div>
      </div>

      <!-- Error -->
      <p v-if="error" class="auth-error-text" style="margin-bottom: 16px;">{{ error }}</p>

      <!-- Stat cards -->
      <div class="usage-stats">
        <div class="usage-stat">
          <div class="usage-stat-label">Total Requests</div>
          <div class="usage-stat-value">{{ formatNumber(summary?.total_requests || 0) }}</div>
          <div class="usage-stat-sub">Last {{ period }} days</div>
        </div>
        <div class="usage-stat">
          <div class="usage-stat-label">Images Generated</div>
          <div class="usage-stat-value">{{ formatNumber(summary?.total_images || 0) }}</div>
          <div class="usage-stat-sub">Last {{ period }} days</div>
        </div>
        <div class="usage-stat">
          <div class="usage-stat-label">Tokens Used</div>
          <div class="usage-stat-value">{{ formatNumber(summary?.total_tokens || 0) }}</div>
          <div class="usage-stat-sub">Last {{ period }} days</div>
        </div>
        <div class="usage-stat">
          <div class="usage-stat-label">Total Cost</div>
          <div class="usage-stat-value">${{ formatCost(summary?.total_cost || 0) }}</div>
          <div class="usage-stat-sub">Last {{ period }} days</div>
        </div>
      </div>

      <!-- Chart -->
      <div class="usage-chart">
        <div class="usage-chart-title">Daily Activity</div>
        <div class="usage-chart-bars">
          <div
            v-for="d in history"
            :key="d.date"
            class="usage-chart-bar"
            :style="{ height: requestsHeight(d) + '%' }"
            :title="`${d.date}: ${d.requests} requests`"
          ></div>
        </div>
        <div class="usage-chart-labels">
          <span v-if="history.length > 0">{{ formatDate(history[0]?.date || '') }}</span>
          <span v-if="history.length > 0">{{ formatDate(history[history.length - 1]?.date || '') }}</span>
        </div>
      </div>

      <!-- Daily breakdown table -->
      <div v-if="history.length" style="background: var(--studio-surface); border: 1px solid var(--studio-border); border-radius: var(--studio-radius-sm); overflow: hidden;">
        <table class="usage-table">
          <thead>
            <tr>
              <th>Date</th>
              <th>Requests</th>
              <th>Images</th>
              <th>Tokens</th>
              <th>Cost</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in history" :key="d.date">
              <td style="font-family: var(--studio-font-mono);">{{ formatDate(d.date) }}</td>
              <td>{{ formatNumber(d.requests) }}</td>
              <td>{{ formatNumber(d.images) }}</td>
              <td>{{ formatNumber(d.tokens) }}</td>
              <td>${{ formatCost(d.cost) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty state -->
      <div v-else-if="!loading" class="studio-empty">
        <div class="studio-empty-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
            <path d="M12 20V10M18 20V4M6 20v-4"/>
          </svg>
        </div>
        <p class="studio-empty-title">No usage data</p>
        <p class="studio-empty-desc">Your API usage statistics will appear here once you start making requests</p>
      </div>
    </div>
  </div>
</template>

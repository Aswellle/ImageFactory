<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apikeyApi, type APIKey, type CreateAPIKeyResponse } from '@/api/apikey'

const keys = ref<APIKey[]>([])
const loading = ref(false)
const showCreate = ref(false)
const newName = ref('')
const createdKey = ref<CreateAPIKeyResponse | null>(null)
const copied = ref(false)
const error = ref<string | null>(null)

onMounted(fetchKeys)

async function fetchKeys() {
  loading.value = true
  try {
    keys.value = await apikeyApi.list()
  } catch {
    error.value = 'Failed to load API keys'
  } finally {
    loading.value = false
  }
}

async function createKey() {
  if (!newName.value.trim()) return
  try {
    createdKey.value = await apikeyApi.create(newName.value.trim())
    newName.value = ''
    showCreate.value = false
    await fetchKeys()
  } catch {
    error.value = 'Failed to create key'
  }
}

async function revokeKey(id: number) {
  if (!confirm('Revoke this API key? This cannot be undone.')) return
  try {
    await apikeyApi.revoke(id)
    await fetchKeys()
  } catch {
    error.value = 'Failed to revoke key'
  }
}

async function copyKey() {
  if (createdKey.value) {
    await navigator.clipboard.writeText(createdKey.value.key)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">API Keys</h1>
        <p class="text-sm text-text-secondary mt-1">Manage programmatic access to your account.</p>
      </div>
      <button class="if-btn-primary" @click="showCreate = true">+ Create Key</button>
    </div>

    <!-- Error -->
    <p v-if="error" class="text-sm text-danger">{{ error }}</p>

    <!-- Created key modal -->
    <div v-if="createdKey" class="if-card p-5 space-y-4 border border-warning/30 bg-warning/5">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-warning">Key Created — Copy Now</h3>
        <button class="text-sm text-text-secondary" @click="createdKey = null">✕</button>
      </div>
      <p class="text-xs text-text-secondary">This key will only be shown once. Store it securely.</p>
      <div class="flex items-center gap-2">
        <code class="flex-1 text-sm bg-surface-2 px-3 py-2 rounded font-mono break-all select-all">{{ createdKey.key }}</code>
        <button class="if-btn-ghost text-sm shrink-0" @click="copyKey">{{ copied ? 'Copied!' : 'Copy' }}</button>
      </div>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">Create API Key</h3>
      <input v-model="newName" class="if-input" placeholder="Key name (e.g., Production)" />
      <div class="flex justify-end gap-2">
        <button class="if-btn-ghost" @click="showCreate = false">Cancel</button>
        <button class="if-btn-primary" :disabled="!newName.trim()" @click="createKey">Create</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-20 text-text-secondary text-sm">Loading…</div>

    <!-- Empty -->
    <div v-else-if="keys.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">No API keys yet</div>
        <div class="text-xs text-text-secondary">Create a key to access the API programmatically.</div>
      </div>
    </div>

    <!-- Key list -->
    <div v-else class="if-card divide-y divide-border">
      <div v-for="k in keys" :key="k.id" class="flex items-center justify-between px-5 py-4">
        <div>
          <div class="text-sm font-medium">{{ k.name || 'Untitled' }}</div>
          <div class="text-xs text-text-secondary mt-0.5 font-mono">{{ k.prefix }}…</div>
        </div>
        <div class="flex items-center gap-4">
          <span
            class="text-xs px-2 py-0.5 rounded-full"
            :class="k.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
          >
            {{ k.status }}
          </span>
          <span class="text-xs text-text-secondary">{{ k.created_at }}</span>
          <button class="if-btn-ghost text-xs" @click="revokeKey(k.id)">Revoke</button>
        </div>
      </div>
    </div>

    <!-- API usage section -->
    <section class="if-card p-5 space-y-4">
      <h3 className="text-sm font-medium">Usage (Last 30 days)</h3>
      <div class="grid grid-cols-3 gap-4">
        <div><div class="text-xs text-text-secondary">Requests</div><div class="text-lg font-semibold">0</div></div>
        <div><div class="text-xs text-text-secondary">Images</div><div class="text-lg font-semibold">0</div></div>
        <div><div class="text-xs text-text-secondary">Tokens</div><div class="text-lg font-semibold">0</div></div>
      </div>
    </section>
  </div>
</template>

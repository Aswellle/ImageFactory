<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apikeyApi, type APIKey, type CreateAPIKeyResponse } from '@/api/apikey'

const { t } = useI18n()
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
    error.value = t('apiKeys.failedToLoad')
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
    error.value = t('apiKeys.failedToCreate')
  }
}

async function revokeKey(id: number) {
  if (!confirm(t('apiKeys.confirmRevoke'))) return
  try {
    await apikeyApi.revoke(id)
    await fetchKeys()
  } catch {
    error.value = t('apiKeys.failedToRevoke')
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
        <h1 class="text-xl font-semibold tracking-tight">{{ t('apiKeys.title') }}</h1>
        <p class="text-sm text-text-secondary mt-1">{{ t('apiKeys.subtitle') }}</p>
      </div>
      <button class="if-btn-primary" @click="showCreate = true">{{ t('apiKeys.createKey') }}</button>
    </div>

    <!-- Error -->
    <p v-if="error" class="text-sm text-danger">{{ error }}</p>

    <!-- Created key modal -->
    <div v-if="createdKey" class="if-card p-5 space-y-4 border border-warning/30 bg-warning/5">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-warning">{{ t('apiKeys.keyCreatedTitle') }}</h3>
        <button class="text-sm text-text-secondary" @click="createdKey = null">✕</button>
      </div>
      <p class="text-xs text-text-secondary">{{ t('apiKeys.keyCreatedNotice') }}</p>
      <div class="flex items-center gap-2">
        <code class="flex-1 text-sm bg-surface-2 px-3 py-2 rounded font-mono break-all select-all">{{ createdKey.key }}</code>
        <button class="if-btn-ghost text-sm shrink-0" @click="copyKey">{{ copied ? t('common.copied') : t('common.copy') }}</button>
      </div>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="if-card p-5 space-y-4">
      <h3 class="text-sm font-medium">{{ t('apiKeys.createApiKey') }}</h3>
      <input v-model="newName" class="if-input" :placeholder="t('apiKeys.keyNamePlaceholder')" />
      <div class="flex justify-end gap-2">
        <button class="if-btn-ghost" @click="showCreate = false">{{ t('common.cancel') }}</button>
        <button class="if-btn-primary" :disabled="!newName.trim()" @click="createKey">{{ t('common.create') }}</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-20 text-text-secondary text-sm">{{ t('common.loading') }}</div>

    <!-- Empty -->
    <div v-else-if="keys.length === 0" class="if-card">
      <div class="py-20 text-center">
        <div class="text-text-secondary text-sm mb-1">{{ t('apiKeys.noApiKeysYet') }}</div>
        <div class="text-xs text-text-secondary">{{ t('apiKeys.createKeyToAccess') }}</div>
      </div>
    </div>

    <!-- Key list -->
    <div v-else class="if-card divide-y divide-border">
      <div v-for="k in keys" :key="k.id" class="flex items-center justify-between px-5 py-4">
        <div>
          <div class="text-sm font-medium">{{ k.name || t('common.untitled') }}</div>
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
          <button class="if-btn-ghost text-xs" @click="revokeKey(k.id)">{{ t('common.revoke') }}</button>
        </div>
      </div>
    </div>

    <!-- API usage section -->
    <section class="if-card p-5 space-y-4">
      <h3 className="text-sm font-medium">{{ t('apiKeys.usageLast30Days') }}</h3>
      <div class="grid grid-cols-3 gap-4">
        <div><div class="text-xs text-text-secondary">{{ t('apiKeys.requests') }}</div><div class="text-lg font-semibold">0</div></div>
        <div><div class="text-xs text-text-secondary">{{ t('apiKeys.images') }}</div><div class="text-lg font-semibold">0</div></div>
        <div><div class="text-xs text-text-secondary">{{ t('apiKeys.tokens') }}</div><div class="text-lg font-semibold">0</div></div>
      </div>
    </section>
  </div>
</template>

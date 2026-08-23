<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apikeyApi, type APIKey, type CreateAPIKeyResponse } from '@/api/apikey'
import SkeletonCard from '@/components/SkeletonCard.vue'

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
  <div class="space-y-8">
    <!-- Header -->
    <div class="flex items-end justify-between">
      <div>
        <h1 class="text-heading text-[var(--text)]">{{ t('apiKeys.title') }}</h1>
        <p class="text-body mt-1">{{ t('apiKeys.subtitle') }}</p>
      </div>
      <button class="btn btn-primary" @click="showCreate = true">{{ t('apiKeys.createKey') }}</button>
    </div>

    <!-- Error -->
    <p v-if="error" class="text-sm text-[var(--danger)]">{{ error }}</p>

    <!-- Created key notice -->
    <div v-if="createdKey" class="bento-tile border-[color-mix(in_srgb,var(--warning)_30%,transparent)] bg-[color-mix(in_srgb,var(--warning)_5%,transparent)]">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-medium" style="color: var(--warning)">{{ t('apiKeys.keyCreatedTitle') }}</h3>
        <button class="text-[var(--text-muted)] hover:text-[var(--text)] transition" @click="createdKey = null">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12" /></svg>
        </button>
      </div>
      <p class="text-caption mb-3">{{ t('apiKeys.keyCreatedNotice') }}</p>
      <div class="flex items-center gap-3">
        <code class="flex-1 text-sm bg-[var(--surface-2)] px-4 py-2.5 rounded-lg text-mono break-all select-all">{{ createdKey.key }}</code>
        <button class="btn btn-ghost shrink-0" @click="copyKey">{{ copied ? t('common.copied') : t('common.copy') }}</button>
      </div>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="bento-tile space-y-4">
      <h3 class="text-subheading text-[var(--text)]">{{ t('apiKeys.createApiKey') }}</h3>
      <div>
        <input v-model="newName" class="input" :placeholder="t('apiKeys.keyNamePlaceholder')" @keyup.enter="createKey" />
      </div>
      <div class="flex justify-end gap-3 pt-2">
        <button class="btn btn-ghost" @click="showCreate = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!newName.trim()" @click="createKey">{{ t('common.create') }}</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <SkeletonCard v-for="i in 3" :key="i" type="list" />
    </div>

    <!-- Empty -->
    <div v-else-if="keys.length === 0" class="bento-tile text-center py-16">
      <div class="text-body mb-1">{{ t('apiKeys.noApiKeysYet') }}</div>
      <div class="text-caption">{{ t('apiKeys.createKeyToAccess') }}</div>
    </div>

    <!-- Key list -->
    <div v-else class="surface rounded-2xl overflow-hidden">
      <div v-for="(k, idx) in keys" :key="k.id" class="flex items-center justify-between px-6 py-4" :class="idx > 0 ? 'border-t border-[var(--border-subtle)]' : ''">
        <div>
          <div class="text-sm font-medium text-[var(--text)]">{{ k.name || t('common.untitled') }}</div>
          <div class="text-caption mt-0.5 text-mono">{{ k.prefix }}…</div>
        </div>
        <div class="flex items-center gap-5">
          <span class="badge" :class="k.status === 'active' ? 'badge-accent' : ''">{{ k.status }}</span>
          <span class="text-caption">{{ k.created_at }}</span>
          <button class="btn btn-ghost text-xs" @click="revokeKey(k.id)">{{ t('common.revoke') }}</button>
        </div>
      </div>
    </div>

    <!-- Usage stats -->
    <div class="bento-tile space-y-5">
      <h3 class="text-subheading text-[var(--text)]">{{ t('apiKeys.usageLast30Days') }}</h3>
      <div class="grid grid-cols-3 gap-6">
        <div>
          <div class="text-caption mb-1">{{ t('apiKeys.requests') }}</div>
          <div class="metric-value text-[var(--text)]">0</div>
        </div>
        <div>
          <div class="text-caption mb-1">{{ t('apiKeys.images') }}</div>
          <div class="metric-value text-[var(--text)]">0</div>
        </div>
        <div>
          <div class="text-caption mb-1">{{ t('apiKeys.tokens') }}</div>
          <div class="metric-value text-[var(--text)]">0</div>
        </div>
      </div>
    </div>
  </div>
</template>

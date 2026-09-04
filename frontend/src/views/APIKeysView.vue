<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apikeyApi, type APIKey, type CreateAPIKeyResponse } from '@/api/apikey'
import SkeletonCard from '@/components/SkeletonCard.vue'
import { useToastStore } from '@/stores/toast'

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
    const toast = useToastStore()
    const { t } = useI18n()
    toast.success(t('toast.apiKeyCopied'), undefined, 2000)
  }
}
</script>

<template>
  <div class="studio-theme">
    <div class="api-keys-page">
      <!-- Header -->
      <div class="api-keys-header">
        <div>
          <h1 class="api-keys-title">API Keys</h1>
          <p class="api-keys-desc">Manage your API keys for programmatic access</p>
        </div>
        <button class="studio-btn studio-btn-primary" @click="showCreate = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14M5 12h14"/></svg>
          Create Key
        </button>
      </div>

      <!-- Error -->
      <p v-if="error" class="auth-error-text" style="margin-bottom: 16px;">{{ error }}</p>

      <!-- Created key notice -->
      <div v-if="createdKey" class="api-key-created">
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
          <span class="api-key-created-label">Key Created — Copy it now!</span>
          <button style="background: transparent; border: none; color: var(--studio-text-muted); cursor: pointer;" @click="createdKey = null">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="api-key-created-value">
          <code>{{ createdKey.key }}</code>
          <button class="api-key-copy-btn" @click="copyKey">{{ copied ? 'Copied!' : 'Copy' }}</button>
        </div>
      </div>

      <!-- Create form -->
      <div v-if="showCreate" style="margin-bottom: 20px; padding: 20px; background: var(--studio-surface); border: 1px solid var(--studio-border); border-radius: var(--studio-radius-sm);">
        <h3 style="font-size: 14px; font-weight: 500; color: var(--studio-text); margin-bottom: 12px;">Create API Key</h3>
        <div style="display: flex; gap: 8px;">
          <input
            v-model="newName"
            class="studio-input"
            style="flex: 1;"
            placeholder="Key name (e.g. Production)"
            @keydown.enter="createKey"
          />
          <button class="studio-btn" @click="showCreate = false">Cancel</button>
          <button class="studio-btn studio-btn-primary" :disabled="!newName.trim()" @click="createKey">Create</button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" style="display: flex; flex-direction: column; gap: 8px;">
        <SkeletonCard v-for="i in 3" :key="i" type="list" />
      </div>

      <!-- Empty -->
      <div v-else-if="keys.length === 0" class="studio-empty" style="padding: 60px;">
        <div class="studio-empty-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
            <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>
          </svg>
        </div>
        <p class="studio-empty-title">No API keys yet</p>
        <p class="studio-empty-desc">Create a key to access the ImageForge API programmatically</p>
      </div>

      <!-- Key list -->
      <div v-else class="api-keys-list">
        <div v-for="k in keys" :key="k.id" class="api-key-item">
          <div class="api-key-info">
            <div class="api-key-name">{{ k.name || 'Untitled' }}</div>
            <div class="api-key-meta">{{ k.prefix }}… · {{ k.status }} · {{ k.created_at }}</div>
          </div>
          <div class="api-key-actions">
            <button class="api-key-revoke" @click="revokeKey(k.id)">Revoke</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

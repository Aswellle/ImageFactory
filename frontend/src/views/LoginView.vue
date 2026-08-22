<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const mode = ref<'login' | 'register'>('login')
const email = ref('')
const password = ref('')
const name = ref('')

async function submit() {
  try {
    if (mode.value === 'register') {
      await auth.register(email.value, password.value, name.value)
    } else {
      await auth.login(email.value, password.value)
    }
    router.push('/')
  } catch {
    // error surfaced via auth.error
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg px-4">
    <div class="w-full max-w-sm">
      <!-- Logo / wordmark -->
      <div class="text-center mb-8">
        <h1 class="text-2xl font-semibold tracking-tight">{{ t('common.appName') }}</h1>
        <p class="text-sm text-text-secondary mt-1.5">
          {{ mode === 'login' ? t('auth.signInTitle') : t('auth.createAccountTitle') }}
        </p>
      </div>

      <!-- Card -->
      <form class="if-card p-6 space-y-4" @submit.prevent="submit">
        <div v-if="mode === 'register'">
          <label class="if-label" for="name">{{ t('auth.nameLabel') }}</label>
          <input id="name" v-model="name" type="text" class="if-input" autocomplete="name" :placeholder="t('auth.namePlaceholder')"
            :aria-invalid="auth.error ? 'true' : undefined" />
        </div>

        <div>
          <label class="if-label" for="email">{{ t('auth.emailLabel') }}</label>
          <input id="email" v-model="email" type="email" class="if-input" autocomplete="email" required :placeholder="t('auth.emailPlaceholder')"
            :aria-invalid="auth.error ? 'true' : undefined"
            :aria-describedby="auth.error ? 'auth-error' : undefined" />
        </div>

        <div>
          <label class="if-label" for="password">{{ t('auth.passwordLabel') }}</label>
          <input id="password" v-model="password" type="password" class="if-input" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" required :placeholder="t('auth.passwordPlaceholder')"
            :aria-invalid="auth.error ? 'true' : undefined"
            :aria-describedby="auth.error ? 'auth-error' : undefined" />
        </div>

        <!-- Error state -->
        <p v-if="auth.error" id="auth-error" role="alert" class="text-sm text-danger">{{ auth.error }}</p>

        <button type="submit" class="if-btn-primary w-full" :disabled="auth.loading">
          {{ auth.loading ? t('auth.pleaseWait') : mode === 'login' ? t('auth.signIn') : t('auth.createAccount') }}
        </button>
      </form>

      <!-- Mode toggle -->
      <p class="text-center text-sm text-text-secondary mt-5">
        <template v-if="mode === 'login'">
          {{ t('auth.dontHaveAccount') }}
          <button class="text-accent font-medium hover:underline" @click="mode = 'register'">{{ t('auth.signUp') }}</button>
        </template>
        <template v-else>
          {{ t('auth.alreadyHaveAccount') }}
          <button class="text-accent font-medium hover:underline" @click="mode = 'login'">{{ t('auth.signIn') }}</button>
        </template>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { authApi } from '@/api/auth'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import FieldHint from '@/components/FieldHint.vue'
import { passwordScore, EMAIL_PLACEHOLDER } from '@/utils/password'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const email = ref('')
const code = ref('')
const newPassword = ref('')
const showPassword = ref(false)

const step = ref<'email' | 'reset'>('email')
const errors = reactive({
  email: '',
  code: '',
  password: ''
})

async function sendCode() {
  errors.email = ''

  if (!email.value.trim()) {
    errors.email = t('auth.emailRequired')
    return
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value)) {
    errors.email = t('auth.invalidEmail')
    return
  }

  try {
    await authApi.sendResetCode(email.value)
    step.value = 'reset'
  } catch {
    errors.email = t('auth.sendCodeFailed')
  }
}

async function resetPassword() {
  errors.code = ''
  errors.password = ''

  let isValid = true

  if (!code.value.trim()) {
    errors.code = t('auth.codeRequired')
    isValid = false
  }

  if (!newPassword.value) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (passwordScore(newPassword.value) < 3) {
    errors.password = t('auth.passwordTooWeak')
    isValid = false
  }

  if (!isValid) return

  try {
    await authApi.resetPassword(code.value, newPassword.value)
    router.push('/login')
  } catch {
    errors.code = t('auth.resetFailed')
  }
}
</script>

<template>
  <div class="min-h-[100dvh] flex items-center justify-center bg-[var(--bg)] px-4 py-8">
    <div class="w-full max-w-sm">
      <!-- Logo / wordmark -->
      <div class="text-center mb-6">
        <h1 class="text-2xl font-semibold tracking-tight text-[var(--text)]">{{ t('common.appName') }}</h1>
        <p class="text-sm text-[var(--text-secondary)] mt-1.5">
          {{ step === 'email' ? t('auth.resetPasswordTitle') : t('auth.resetPasswordNew') }}
        </p>
      </div>

      <!-- Card -->
      <form class="surface rounded-2xl p-6 space-y-4" @submit.prevent="step === 'email' ? sendCode() : resetPassword()">
        <!-- Step 1: Enter email -->
        <template v-if="step === 'email'">
          <p class="text-sm text-[var(--text-secondary)]">
            {{ t('auth.resetPasswordDesc') }}
          </p>
          <div>
            <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="reset-email">{{ t('auth.emailLabel') }}</label>
            <input
              id="reset-email"
              v-model="email"
              type="email"
              class="input"
              :class="{ 'input-error': errors.email }"
              autocomplete="email"
              required
              :placeholder="EMAIL_PLACEHOLDER"
              :aria-invalid="!!errors.email"
            />
            <FieldHint v-if="errors.email" type="error" :message="errors.email" />
          </div>
        </template>

        <!-- Step 2: Enter code and new password -->
        <template v-else>
          <p class="text-sm text-[var(--text-secondary)]">
            {{ t('auth.resetPasswordSent', { email }) }}
          </p>
          <div>
            <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="code">{{ t('auth.verificationCodeLabel') }}</label>
            <input
              id="code"
              v-model="code"
              type="text"
              class="input"
              :class="{ 'input-error': errors.code }"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              required
              :placeholder="t('auth.verificationCodePlaceholder')"
              :aria-invalid="!!errors.code"
              @input="code = code.replace(/\D/g, '')"
            />
            <FieldHint v-if="errors.code" type="error" :message="errors.code" />
          </div>
          <div>
            <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="new-password">{{ t('auth.newPasswordLabel') }}</label>
            <div class="relative">
              <input
                id="new-password"
                v-model="newPassword"
                :type="showPassword ? 'text' : 'password'"
                class="input pr-10"
                :class="{ 'input-error': errors.password }"
                autocomplete="new-password"
                required
                :placeholder="t('auth.passwordPlaceholder')"
                :aria-invalid="!!errors.password"
              />
              <button
                type="button"
                class="absolute inset-y-0 right-0 flex items-center justify-center w-10 text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
                :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
                :aria-pressed="showPassword"
                @click="showPassword = !showPassword"
              >
                <svg v-if="!showPassword" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
                <svg v-else class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
              </button>
            </div>
            <FieldHint v-if="errors.password" type="error" :message="errors.password" />
          </div>
        </template>

        <!-- Error state -->
        <p v-if="auth.error" id="auth-error" role="alert" class="text-sm text-[var(--danger)]">{{ auth.error }}</p>

        <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading">
          <span v-if="auth.loading" class="flex items-center justify-center gap-2">
            <LoadingSpinner size="sm" />
            {{ t('auth.pleaseWait') }}
          </span>
          <span v-else>{{ step === 'email' ? t('auth.sendCode') : t('auth.resetPasswordSubmit') }}</span>
        </button>

        <!-- Back to login -->
        <p class="text-center">
          <router-link to="/login" class="text-sm text-[var(--accent)] hover:underline">
            {{ t('auth.backToLogin') }}
          </router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<style scoped>
.input-error {
  border-color: var(--danger);
  box-shadow: 0 0 0 1px var(--danger);
}
</style>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import FieldHint from '@/components/FieldHint.vue'
import PasswordStrengthMeter from '@/components/PasswordStrengthMeter.vue'
import { passwordScore, EMAIL_PLACEHOLDER } from '@/utils/password'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const mode = ref<'login' | 'register'>('login')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const name = ref('')

const showPassword = ref(false)
const showConfirmPassword = ref(false)

const errors = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// Clear validation errors and password visibility when toggling modes
watch(mode, () => {
  errors.name = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
  showPassword.value = false
  showConfirmPassword.value = false
})

function validate(): boolean {
  errors.name = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''

  let isValid = true

  if (mode.value === 'register') {
    if (!name.value.trim()) {
      errors.name = t('auth.nameRequired')
      isValid = false
    }
  }

  if (!email.value.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  if (!password.value) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (mode.value === 'register' && passwordScore(password.value) < 3) {
    errors.password = t('auth.passwordTooWeak')
    isValid = false
  }

  if (mode.value === 'register') {
    if (!confirmPassword.value) {
      errors.confirmPassword = t('auth.confirmPasswordRequired')
      isValid = false
    } else if (confirmPassword.value !== password.value) {
      errors.confirmPassword = t('auth.passwordMismatch')
      isValid = false
    }
  }

  return isValid
}

async function submit() {
  if (!validate()) return

  try {
    if (mode.value === 'register') {
      await auth.register(email.value, password.value, name.value)
    } else {
      await auth.login(email.value, password.value)
    }
    router.push('/app')
  } catch {
    // error surfaced via auth.error
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
          {{ mode === 'login' ? t('auth.signInTitle') : t('auth.createAccountTitle') }}
        </p>
      </div>

      <!-- Card -->
      <form class="surface rounded-2xl p-6 space-y-4" @submit.prevent="submit">
        <!-- Name (register only) -->
        <div v-if="mode === 'register'">
          <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="name">{{ t('auth.nameLabel') }}</label>
          <input
            id="name"
            v-model="name"
            type="text"
            class="input"
            :class="{ 'input-error': errors.name }"
            autocomplete="name"
            :placeholder="t('auth.namePlaceholder')"
            :aria-invalid="!!errors.name"
          />
          <FieldHint v-if="errors.name" type="error" :message="errors.name" />
        </div>

        <!-- Email -->
        <div>
          <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="email">{{ t('auth.emailLabel') }}</label>
          <input
            id="email"
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

        <!-- Password -->
        <div>
          <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="password">{{ t('auth.passwordLabel') }}</label>
          <div class="relative">
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              class="input pr-10"
              :class="{ 'input-error': errors.password }"
              :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
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
          <PasswordStrengthMeter v-if="mode === 'register'" :model-value="password" />
        </div>

        <!-- Confirm Password (register only) -->
        <div v-if="mode === 'register'">
          <label class="text-sm font-medium text-[var(--text)] mb-1.5 block" for="confirmPassword">{{ t('auth.confirmPasswordLabel') }}</label>
          <div class="relative">
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              class="input pr-10"
              :class="{ 'input-error': errors.confirmPassword }"
              autocomplete="new-password"
              required
              :placeholder="t('auth.confirmPasswordPlaceholder')"
              :aria-invalid="!!errors.confirmPassword"
            />
            <button
              type="button"
              class="absolute inset-y-0 right-0 flex items-center justify-center w-10 text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
              :aria-label="showConfirmPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              :aria-pressed="showConfirmPassword"
              @click="showConfirmPassword = !showConfirmPassword"
            >
              <svg v-if="!showConfirmPassword" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
              <svg v-else class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
            </button>
          </div>
          <FieldHint v-if="errors.confirmPassword" type="error" :message="errors.confirmPassword" />
        </div>

        <!-- Forgot password (login only) -->
        <div v-if="mode === 'login'" class="flex justify-end">
          <router-link
            to="/forgot-password"
            class="text-sm text-[var(--accent)] hover:underline"
          >
            {{ t('auth.forgotPassword') }}
          </router-link>
        </div>

        <!-- Error state -->
        <p v-if="auth.error" id="auth-error" role="alert" class="text-sm text-[var(--danger)]">{{ auth.error }}</p>

        <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading">
          <span v-if="auth.loading" class="flex items-center justify-center gap-2">
            <LoadingSpinner size="sm" />
            {{ t('auth.pleaseWait') }}
          </span>
          <span v-else>{{ mode === 'login' ? t('auth.signIn') : t('auth.createAccount') }}</span>
        </button>
      </form>

      <!-- Mode toggle -->
      <p class="text-center text-sm text-[var(--text-secondary)] mt-5">
        <template v-if="mode === 'login'">
          {{ t('auth.dontHaveAccount') }}
          <button class="text-[var(--accent)] font-medium hover:underline" @click="mode = 'register'">{{ t('auth.signUp') }}</button>
        </template>
        <template v-else>
          {{ t('auth.alreadyHaveAccount') }}
          <button class="text-[var(--accent)] font-medium hover:underline" @click="mode = 'login'">{{ t('auth.signIn') }}</button>
        </template>
      </p>
    </div>
  </div>
</template>

<style scoped>
.input-error {
  border-color: var(--danger);
  box-shadow: 0 0 0 1px var(--danger);
}
</style>

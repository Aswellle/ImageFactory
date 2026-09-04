<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
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

watch(mode, () => {
  errors.name = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
  showPassword.value = false
  showConfirmPassword.value = false
})

const strengthLevel = computed(() => passwordScore(password.value))

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

function strengthLabel(score: number): string {
  if (score <= 1) return 'Weak'
  if (score <= 2) return 'Fair'
  if (score <= 3) return 'Good'
  return 'Strong'
}

import { computed } from 'vue'
</script>

<template>
  <div class="studio-theme">
    <div class="auth-screen">
      <div class="auth-container">
        <!-- Header -->
        <div class="auth-header">
          <div class="auth-logo">
            <svg width="24" height="24" viewBox="0 0 32 32" fill="none">
              <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="#0C0A09" />
            </svg>
          </div>
          <h1 class="auth-title">ImageForge</h1>
          <p class="auth-subtitle">
            {{ mode === 'login' ? 'Sign in to your account' : 'Create your account' }}
          </p>
        </div>

        <!-- Card -->
        <form class="auth-card" @submit.prevent="submit">
          <!-- Name (register only) -->
          <div v-if="mode === 'register'" class="auth-field">
            <label class="auth-label" for="name">Name</label>
            <div class="auth-input-wrap">
              <input
                id="name"
                v-model="name"
                class="auth-input"
                :class="{ 'auth-input-error': errors.name }"
                type="text"
                placeholder="Your name"
                autocomplete="name"
              />
            </div>
            <p v-if="errors.name" class="auth-error-text">{{ errors.name }}</p>
          </div>

          <!-- Email -->
          <div class="auth-field">
            <label class="auth-label" for="email">Email</label>
            <div class="auth-input-wrap">
              <input
                id="email"
                v-model="email"
                class="auth-input"
                :class="{ 'auth-input-error': errors.email }"
                type="email"
                :placeholder="EMAIL_PLACEHOLDER"
                autocomplete="email"
              />
            </div>
            <p v-if="errors.email" class="auth-error-text">{{ errors.email }}</p>
          </div>

          <!-- Password -->
          <div class="auth-field">
            <label class="auth-label" for="password">Password</label>
            <div class="auth-input-wrap">
              <input
                id="password"
                v-model="password"
                class="auth-input"
                :class="{ 'auth-input-error': errors.password }"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Enter password"
                :autocomplete="mode === 'register' ? 'new-password' : 'current-password'"
              />
              <button
                type="button"
                class="auth-input-icon"
                @click="showPassword = !showPassword"
              >
                <svg v-if="!showPassword" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
              </button>
            </div>
            <p v-if="errors.password" class="auth-error-text">{{ errors.password }}</p>

            <!-- Password strength (register only) -->
            <div v-if="mode === 'register' && password.length > 0" class="auth-strength">
              <div
                v-for="i in 4"
                :key="i"
                class="auth-strength-bar"
                :class="{
                  active: i <= strengthLevel,
                  weak: strengthLevel <= 1 && i <= strengthLevel,
                  medium: strengthLevel === 2 && i <= strengthLevel,
                  strong: strengthLevel >= 3 && i <= strengthLevel
                }"
              ></div>
            </div>
            <p v-if="mode === 'register' && password.length > 0" style="font-size: 11px; color: var(--studio-text-muted); margin-top: 4px;">
              {{ strengthLabel(strengthLevel) }}
            </p>
          </div>

          <!-- Confirm Password (register only) -->
          <div v-if="mode === 'register'" class="auth-field">
            <label class="auth-label" for="confirm">Confirm Password</label>
            <div class="auth-input-wrap">
              <input
                id="confirm"
                v-model="confirmPassword"
                class="auth-input"
                :class="{ 'auth-input-error': errors.confirmPassword }"
                :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="Confirm password"
                autocomplete="new-password"
              />
              <button
                type="button"
                class="auth-input-icon"
                @click="showConfirmPassword = !showConfirmPassword"
              >
                <svg v-if="!showConfirmPassword" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
              </button>
            </div>
            <p v-if="errors.confirmPassword" class="auth-error-text">{{ errors.confirmPassword }}</p>
          </div>

          <!-- Forgot password (login only) -->
          <div v-if="mode === 'login'" class="auth-forgot-link">
            <button type="button" @click="router.push('/forgot-password')">
              Forgot password?
            </button>
          </div>

          <!-- Error -->
          <p v-if="auth.error" id="auth-error" role="alert" class="auth-error-text" style="margin-bottom: 12px;">
            {{ auth.error }}
          </p>

          <!-- Submit -->
          <button type="submit" class="auth-submit-btn" :disabled="auth.loading">
            <LoadingSpinner v-if="auth.loading" size="sm" />
            <span v-else>{{ mode === 'login' ? 'Sign in' : 'Create account' }}</span>
          </button>
        </form>

        <!-- Mode toggle -->
        <p class="auth-footer">
          {{ mode === 'login' ? "Don't have an account?" : 'Already have an account?' }}
          <button type="button" @click="mode = mode === 'login' ? 'register' : 'login'">
            {{ mode === 'login' ? 'Sign up' : 'Sign in' }}
          </button>
        </p>
      </div>
    </div>
  </div>
</template>

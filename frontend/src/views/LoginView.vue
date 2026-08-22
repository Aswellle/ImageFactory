<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

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
        <h1 class="text-2xl font-semibold tracking-tight">ImageForge</h1>
        <p class="text-sm text-text-secondary mt-1.5">
          {{ mode === 'login' ? 'Sign in to your workspace' : 'Create your workspace' }}
        </p>
      </div>

      <!-- Card -->
      <form class="if-card p-6 space-y-4" @submit.prevent="submit">
        <div v-if="mode === 'register'">
          <label class="if-label" for="name">Name</label>
          <input id="name" v-model="name" type="text" class="if-input" autocomplete="name" placeholder="Your name"
            :aria-invalid="auth.error ? 'true' : undefined" />
        </div>

        <div>
          <label class="if-label" for="email">Email</label>
          <input id="email" v-model="email" type="email" class="if-input" autocomplete="email" required placeholder="you@company.com"
            :aria-invalid="auth.error ? 'true' : undefined"
            :aria-describedby="auth.error ? 'auth-error' : undefined" />
        </div>

        <div>
          <label class="if-label" for="password">Password</label>
          <input id="password" v-model="password" type="password" class="if-input" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" required placeholder="••••••••"
            :aria-invalid="auth.error ? 'true' : undefined"
            :aria-describedby="auth.error ? 'auth-error' : undefined" />
        </div>

        <!-- Error state -->
        <p v-if="auth.error" id="auth-error" role="alert" class="text-sm text-danger">{{ auth.error }}</p>

        <button type="submit" class="if-btn-primary w-full" :disabled="auth.loading">
          {{ auth.loading ? 'Please wait…' : mode === 'login' ? 'Sign In' : 'Create Account' }}
        </button>
      </form>


      <!-- Mode toggle -->
      <p class="text-center text-sm text-text-secondary mt-5">
        <template v-if="mode === 'login'">
          Don't have an account?
          <button class="text-accent font-medium hover:underline" @click="mode = 'register'">Sign up</button>
        </template>
        <template v-else>
          Already have an account?
          <button class="text-accent font-medium hover:underline" @click="mode = 'login'">Sign in</button>
        </template>
      </p>
    </div>
  </div>
</template>

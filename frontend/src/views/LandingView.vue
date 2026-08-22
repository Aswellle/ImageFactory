<template>
  <div class="min-h-[100dvh] bg-[var(--bg)] text-[var(--text)]">
    <!-- Navigation -->
    <header class="fixed inset-x-0 top-0 z-[var(--z-overlay)]">
      <nav class="mx-auto flex h-12 max-w-[1024px] items-center justify-between px-6">
        <a href="/" class="flex items-center gap-2 text-sm font-medium text-[var(--text)]">
          <svg width="20" height="20" viewBox="0 0 32 32" fill="none" class="opacity-90">
            <rect width="32" height="32" rx="7" fill="currentColor" opacity="0.08" />
            <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="currentColor" opacity="0.9" />
          </svg>
          ImageForge
        </a>
        <div class="flex items-center gap-6">
          <a href="#features" class="text-xs font-medium text-[var(--text-secondary)] transition hover:text-[var(--text)]">{{ t('nav.features') }}</a>
          <a href="#workflow" class="text-xs font-medium text-[var(--text-secondary)] transition hover:text-[var(--text)]">{{ t('nav.workflow') }}</a>
          <button @click="openDesktop" class="inline-flex h-8 items-center gap-1.5 rounded-full bg-[var(--text)] px-4 text-xs font-medium text-[var(--bg)] transition hover:opacity-80">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="20" height="14" rx="2" />
              <path d="M8 21h8M12 17v4" />
            </svg>
            {{ t('nav.openApp') }}
          </button>
        </div>
      </nav>
    </header>

    <!-- Hero -->
    <section class="flex min-h-[100dvh] items-center justify-center px-6 pt-12">
      <div class="mx-auto max-w-[880px] text-center">
        <p class="mb-4 text-sm font-medium tracking-wide text-[var(--text-secondary)]">{{ t('hero.tagline') }}</p>
        <h1 class="text-display mb-6 text-[var(--text)]">
          {{ t('hero.title.line1') }}<br />
          <span class="text-[var(--text-secondary)]">{{ t('hero.title.line2') }}</span>
        </h1>
        <p class="mx-auto mb-10 max-w-[560px] text-lg leading-relaxed text-[var(--text-secondary)]">
          {{ t('hero.description') }}
        </p>
        <div class="flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4">
          <button @click="openDesktop" class="inline-flex h-12 items-center gap-2 rounded-full bg-[var(--accent)] px-8 text-sm font-medium text-white transition hover:opacity-90 active:scale-[0.98]">
            {{ t('hero.cta.primary') }}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M7 17L17 7M17 7H8M17 7v9" />
            </svg>
          </button>
          <a href="#features" class="inline-flex h-12 items-center gap-2 rounded-full border border-[var(--border)] px-8 text-sm font-medium text-[var(--text)] transition hover:bg-[var(--surface-2)]">
            {{ t('hero.cta.secondary') }}
          </a>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section id="features" class="px-6 py-24">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('features.tagline') }}</p>
        <h2 class="text-heading mb-16 max-w-[480px] text-[var(--text)]">{{ t('features.title') }}</h2>

        <!-- Feature grid: asymmetric -->
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="bento-tile md:col-span-2">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('features.gen.title') }}</h3>
            <p class="text-body max-w-[420px]">{{ t('features.gen.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('features.versions.title') }}</h3>
            <p class="text-body">{{ t('features.versions.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('features.organize.title') }}</h3>
            <p class="text-body">{{ t('features.organize.desc') }}</p>
          </div>

          <div class="bento-tile md:col-span-2">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('features.api.title') }}</h3>
            <p class="text-body max-w-[420px]">{{ t('features.api.desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Workflow -->
    <section id="workflow" class="px-6 py-24">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('workflow.tagline') }}</p>
        <h2 class="text-heading mb-16 text-[var(--text)]">{{ t('workflow.title') }}</h2>

        <div class="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">01</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('workflow.step1.title') }}</h3>
            <p class="text-body">{{ t('workflow.step1.desc') }}</p>
          </div>
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">02</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('workflow.step2.title') }}</h3>
            <p class="text-body">{{ t('workflow.step2.desc') }}</p>
          </div>
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">03</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('workflow.step3.title') }}</h3>
            <p class="text-body">{{ t('workflow.step3.desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="px-6 py-24">
      <div class="mx-auto max-w-[640px] text-center">
        <h2 class="text-heading mb-4 text-[var(--text)]">{{ t('cta.title') }}</h2>
        <p class="mb-8 text-body">{{ t('cta.desc') }}</p>
        <button @click="openDesktop" class="inline-flex h-12 items-center gap-2 rounded-full bg-[var(--text)] px-8 text-sm font-medium text-[var(--bg)] transition hover:opacity-80 active:scale-[0.98]">
          {{ t('cta.button') }}
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M7 17L17 7M17 7H8M17 7v9" />
          </svg>
        </button>
      </div>
    </section>

    <!-- Footer -->
    <footer class="border-t border-[var(--border)] px-6 py-8">
      <div class="mx-auto flex max-w-[1024px] flex-col items-center justify-between gap-4 text-xs text-[var(--text-muted)] sm:flex-row">
        <span>© 2026 ImageForge. All rights reserved.</span>
        <div class="flex items-center gap-6">
          <a href="#" class="transition hover:text-[var(--text)]">Privacy</a>
          <a href="#" class="transition hover:text-[var(--text)]">Terms</a>
          <a href="#" class="transition hover:text-[var(--text)]">Contact</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useDesktopLauncher } from '@/composables/useDesktopLauncher'

const { t } = useI18n()
const { openDesktop } = useDesktopLauncher()
</script>

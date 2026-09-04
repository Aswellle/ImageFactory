<template>
  <div class="min-h-[100dvh] bg-[var(--bg)] text-[var(--text)]">
    <!-- Navigation - Apple-style Liquid Glass -->
    <header 
      class="fixed inset-x-0 top-0 z-[var(--z-overlay)] flex justify-center px-4 pt-3"
      @mousemove="handleNavMouseMove"
      @mouseleave="handleNavMouseLeave"
    >
      <nav 
        ref="navGlass"
        class="nav-glass-container"
      >
        <!-- Glass highlight layers -->
        <span class="glass-highlight" ref="highlight"></span>
        <span class="glass-border"></span>
        
        <!-- Navigation content -->
        <div class="nav-content">
          <a href="/" class="flex items-center gap-2 text-sm font-medium text-[var(--text)]">
            <svg width="20" height="20" viewBox="0 0 32 32" fill="none" class="opacity-90">
              <rect width="32" height="32" rx="7" fill="currentColor" opacity="0.08" />
              <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="currentColor" opacity="0.9" />
            </svg>
            ImageForge
          </a>
          <div class="flex items-center gap-6">
            <a href="#features" class="nav-link text-xs font-medium text-[var(--text-secondary)] transition-colors hover:text-[var(--text)]">{{ t('landing.nav.features') }}</a>
            <a href="#workflow" class="nav-link text-xs font-medium text-[var(--text-secondary)] transition-colors hover:text-[var(--text)]">{{ t('landing.nav.workflow') }}</a>
            <a href="#pricing" class="nav-link text-xs font-medium text-[var(--text-secondary)] transition-colors hover:text-[var(--text)]">{{ t('landing.pricing.tagline') }}</a>
            <button @click="openDesktop" class="inline-flex h-8 items-center gap-1.5 rounded-full bg-[var(--text)] px-4 text-xs font-medium text-[var(--bg)] transition hover:opacity-80">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2" />
                <path d="M8 21h8M12 17v4" />
              </svg>
              {{ t('landing.nav.openApp') }}
            </button>
          </div>
        </div>
      </nav>
    </header>

    <!-- Hero -->
    <section class="flex min-h-[100dvh] items-center justify-center px-6 pt-12">
      <div class="mx-auto max-w-[880px] text-center">
        <p class="mb-4 text-sm font-medium tracking-wide text-[var(--text-secondary)]">{{ t('landing.hero.tagline') }}</p>
        <h1 class="text-display mb-6 text-[var(--text)]">
          {{ t('landing.hero.title.line1') }}<br />
          <span class="text-[var(--text-secondary)]">{{ t('landing.hero.title.line2') }}</span>
        </h1>
        <p class="mx-auto mb-10 max-w-[560px] text-lg leading-relaxed text-[var(--text-secondary)]">
          {{ t('landing.hero.description') }}
        </p>
        <div class="flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4">
          <button @click="openDesktop" class="inline-flex h-12 items-center gap-2 rounded-full bg-[var(--accent)] px-8 text-sm font-medium text-white transition hover:opacity-90 active:scale-[0.98]">
            {{ t('landing.hero.cta.primary') }}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M7 17L17 7M17 7H8M17 7v9" />
            </svg>
          </button>
          <a href="#features" class="inline-flex h-12 items-center gap-2 rounded-full border border-[var(--border)] px-8 text-sm font-medium text-[var(--text)] transition hover:bg-[var(--surface-2)]">
            {{ t('landing.hero.cta.secondary') }}
          </a>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section id="features" class="px-6 py-24">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.features.tagline') }}</p>
        <h2 class="text-heading mb-16 max-w-[480px] text-[var(--text)]">{{ t('landing.features.title') }}</h2>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="bento-tile md:col-span-2">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.features.gen.title') }}</h3>
            <p class="text-body max-w-[420px]">{{ t('landing.features.gen.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.features.versions.title') }}</h3>
            <p class="text-body">{{ t('landing.features.versions.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.features.organize.title') }}</h3>
            <p class="text-body">{{ t('landing.features.organize.desc') }}</p>
          </div>

          <div class="bento-tile md:col-span-2">
            <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--surface-2)]">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--accent)]">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.features.api.title') }}</h3>
            <p class="text-body max-w-[420px]">{{ t('landing.features.api.desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Use Cases -->
    <section id="useCases" class="px-6 py-24 bg-[var(--surface-2)]">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.useCases.tagline') }}</p>
        <h2 class="text-heading mb-16 text-[var(--text)]">{{ t('landing.useCases.title') }}</h2>

        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div class="bento-tile">
            <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--accent-soft)]">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--accent)]">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.useCases.marketing.title') }}</h3>
            <p class="text-body">{{ t('landing.useCases.marketing.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--accent-soft)]">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--accent)]">
                <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z" />
                <line x1="3" y1="6" x2="21" y2="6" />
                <path d="M16 10a4 4 0 01-8 0" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.useCases.ecommerce.title') }}</h3>
            <p class="text-body">{{ t('landing.useCases.ecommerce.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--accent-soft)]">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--accent)]">
                <rect x="5" y="2" width="14" height="20" rx="2" />
                <line x1="12" y1="18" x2="12" y2="18" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.useCases.social.title') }}</h3>
            <p class="text-body">{{ t('landing.useCases.social.desc') }}</p>
          </div>

          <div class="bento-tile">
            <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--accent-soft)]">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--accent)]">
                <path d="M12 19l7-7 3 3-7 7-3-3z" />
                <path d="M18 13l-1.5-7.5L2 2l3.5 14.5L13 18l5-5z" />
                <path d="M2 2l7.586 7.586" />
                <circle cx="11" cy="11" r="2" />
              </svg>
            </div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.useCases.design.title') }}</h3>
            <p class="text-body">{{ t('landing.useCases.design.desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Workflow -->
    <section id="workflow" class="px-6 py-24">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.workflow.tagline') }}</p>
        <h2 class="text-heading mb-16 text-[var(--text)]">{{ t('landing.workflow.title') }}</h2>

        <div class="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">01</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.workflow.step1.title') }}</h3>
            <p class="text-body">{{ t('landing.workflow.step1.desc') }}</p>
          </div>
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">02</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.workflow.step2.title') }}</h3>
            <p class="text-body">{{ t('landing.workflow.step2.desc') }}</p>
          </div>
          <div>
            <div class="metric-value mb-3 text-[var(--text)]">03</div>
            <h3 class="mb-2 text-subheading text-[var(--text)]">{{ t('landing.workflow.step3.title') }}</h3>
            <p class="text-body">{{ t('landing.workflow.step3.desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- API Showcase -->
    <section id="api" class="px-6 py-24 bg-[var(--surface-2)]">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.api.tagline') }}</p>
        <h2 class="text-heading mb-4 text-[var(--text)]">{{ t('landing.api.title') }}</h2>
        <p class="text-body mb-8">{{ t('landing.api.desc') }}</p>

        <div class="bento-tile bg-[var(--text)] text-[var(--bg)] overflow-hidden">
          <div class="flex items-center gap-2 mb-4">
            <div class="w-3 h-3 rounded-full bg-[#ff5f56]"></div>
            <div class="w-3 h-3 rounded-full bg-[#ffbd2e]"></div>
            <div class="w-3 h-3 rounded-full bg-[#27c93f]"></div>
            <span class="ml-2 text-xs opacity-60">Python</span>
          </div>
          <pre class="text-sm font-mono overflow-x-auto"><code>from imageforge import ImageForge

client = ImageForge(api_key="your_api_key")

# Generate an image
result = client.images.generate(
    prompt="A serene mountain landscape at sunset",
    model="flux-1-pro",
    width=1024,
    height=1024
)

result.save("landscape.png")
print(f"Image URL: {result.url}")</code></pre>
        </div>
      </div>
    </section>

    <!-- Pricing -->
    <section id="pricing" class="px-6 py-24 bg-[var(--surface-2)]">
      <div class="mx-auto max-w-[1024px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.pricing.tagline') }}</p>
        <h2 class="text-heading mb-16 text-[var(--text)]">{{ t('landing.pricing.title') }}</h2>

        <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
          <!-- Free Tier -->
          <div class="bento-tile flex flex-col">
            <h3 class="text-subheading text-[var(--text)]">{{ t('landing.pricing.free.name') }}</h3>
            <div class="my-4">
              <span class="text-display text-[var(--text)]">{{ t('landing.pricing.free.price') }}</span>
              <span class="text-sm text-[var(--text-secondary)]">{{ t('landing.pricing.free.period') }}</span>
            </div>
            <ul class="mb-6 flex-1 space-y-3">
              <li v-for="(feature, idx) in freeFeatures" :key="idx" class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-[var(--success)]">
                  <path d="M20 6L9 17l-5-5" />
                </svg>
                {{ feature }}
              </li>
            </ul>
            <button @click="openDesktop" class="w-full inline-flex h-10 items-center justify-center rounded-full border border-[var(--border)] text-sm font-medium text-[var(--text)] transition hover:bg-[var(--surface)]">
              {{ t('landing.pricing.free.cta') }}
            </button>
          </div>

          <!-- Pro Tier -->
          <div class="bento-tile flex flex-col border-[var(--accent)] relative">
            <div class="absolute -top-3 left-1/2 -translate-x-1/2">
              <span class="inline-flex h-6 items-center rounded-full bg-[var(--accent)] px-3 text-xs font-medium text-white">{{ t('landing.pricing.pro.popular') }}</span>
            </div>
            <h3 class="text-subheading text-[var(--text)]">{{ t('landing.pricing.pro.name') }}</h3>
            <div class="my-4">
              <span class="text-display text-[var(--text)]">{{ t('landing.pricing.pro.price') }}</span>
              <span class="text-sm text-[var(--text-secondary)]">{{ t('landing.pricing.pro.period') }}</span>
            </div>
            <ul class="mb-6 flex-1 space-y-3">
              <li v-for="(feature, idx) in proFeatures" :key="idx" class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-[var(--success)]">
                  <path d="M20 6L9 17l-5-5" />
                </svg>
                {{ feature }}
              </li>
            </ul>
            <button @click="openDesktop" class="w-full inline-flex h-10 items-center justify-center rounded-full bg-[var(--accent)] text-sm font-medium text-white transition hover:opacity-90">
              {{ t('landing.pricing.pro.cta') }}
            </button>
          </div>

          <!-- Enterprise Tier -->
          <div class="bento-tile flex flex-col">
            <h3 class="text-subheading text-[var(--text)]">{{ t('landing.pricing.enterprise.name') }}</h3>
            <div class="my-4">
              <span class="text-display text-[var(--text)]">{{ t('landing.pricing.enterprise.price') }}</span>
              <span class="text-sm text-[var(--text-secondary)]">{{ t('landing.pricing.enterprise.period') }}</span>
            </div>
            <ul class="mb-6 flex-1 space-y-3">
              <li v-for="(feature, idx) in enterpriseFeatures" :key="idx" class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-[var(--success)]">
                  <path d="M20 6L9 17l-5-5" />
                </svg>
                {{ feature }}
              </li>
            </ul>
            <button @click="openDesktop" class="w-full inline-flex h-10 items-center justify-center rounded-full border border-[var(--border)] text-sm font-medium text-[var(--text)] transition hover:bg-[var(--surface)]">
              {{ t('landing.pricing.enterprise.cta') }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section id="faq" class="px-6 py-24">
      <div class="mx-auto max-w-[720px]">
        <p class="mb-2 text-sm font-medium text-[var(--text-secondary)]">{{ t('landing.faq.tagline') }}</p>
        <h2 class="text-heading mb-16 text-[var(--text)]">{{ t('landing.faq.title') }}</h2>

        <div class="space-y-4">
          <div class="bento-tile">
            <h3 class="text-subheading text-[var(--text)] mb-2">{{ t('landing.faq.q1.q') }}</h3>
            <p class="text-body">{{ t('landing.faq.q1.a') }}</p>
          </div>
          <div class="bento-tile">
            <h3 class="text-subheading text-[var(--text)] mb-2">{{ t('landing.faq.q2.q') }}</h3>
            <p class="text-body">{{ t('landing.faq.q2.a') }}</p>
          </div>
          <div class="bento-tile">
            <h3 class="text-subheading text-[var(--text)] mb-2">{{ t('landing.faq.q3.q') }}</h3>
            <p class="text-body">{{ t('landing.faq.q3.a') }}</p>
          </div>
          <div class="bento-tile">
            <h3 class="text-subheading text-[var(--text)] mb-2">{{ t('landing.faq.q4.q') }}</h3>
            <p class="text-body">{{ t('landing.faq.q4.a') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="px-6 py-24">
      <div class="mx-auto max-w-[640px] text-center">
        <h2 class="text-heading mb-4 text-[var(--text)]">{{ t('landing.cta.title') }}</h2>
        <p class="mb-8 text-body">{{ t('landing.cta.desc') }}</p>
        <button @click="openDesktop" class="inline-flex h-12 items-center gap-2 rounded-full bg-[var(--text)] px-8 text-sm font-medium text-[var(--bg)] transition hover:opacity-80 active:scale-[0.98]">
          {{ t('landing.cta.button') }}
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M7 17L17 7M17 7H8M17 7v9" />
          </svg>
        </button>
      </div>
    </section>

    <!-- Footer -->
    <footer class="border-t border-[var(--border)] px-6 py-12">
      <div class="mx-auto max-w-[1024px]">
        <div class="flex flex-col items-center justify-between gap-4 sm:flex-row">
          <div class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
            <svg width="16" height="16" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="7" fill="currentColor" opacity="0.08" />
              <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="currentColor" opacity="0.9" />
            </svg>
            © 2026 ImageForge. All rights reserved.
          </div>
          <div class="flex items-center gap-4">
            <a href="#" class="text-[var(--text-secondary)] hover:text-[var(--text)] transition-colors">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
            </a>
            <a href="#" class="text-[var(--text-secondary)] hover:text-[var(--text)] transition-colors">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
              </svg>
            </a>
            <a href="#" class="text-[var(--text-secondary)] hover:text-[var(--text)] transition-colors">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
              </svg>
            </a>
          </div>
        </div>
      </div>
    </footer>

  </div>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { i18n } from '@/i18n'
import { useDesktopLauncher } from '@/composables/useDesktopLauncher'

const { t, tm } = useI18n()
const { openDesktop } = useDesktopLauncher()

// Pricing features (using ref + watch for language switching)
const freeFeatures = ref<string[]>([])
const proFeatures = ref<string[]>([])
const enterpriseFeatures = ref<string[]>([])

function updatePricingFeatures() {
  freeFeatures.value = JSON.parse(JSON.stringify(tm('landing.pricing.free.features'))) || []
  proFeatures.value = JSON.parse(JSON.stringify(tm('landing.pricing.pro.features'))) || []
  enterpriseFeatures.value = JSON.parse(JSON.stringify(tm('landing.pricing.enterprise.features'))) || []
}

// Update on language change
watch(() => i18n.global.locale.value, updatePricingFeatures, { immediate: true })


// Liquid glass mouse tracking
const navGlass = ref<HTMLElement | null>(null)
const highlight = ref<HTMLElement | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)
const targetX = ref(0)
const targetY = ref(0)
let animationFrame: number | null = null

function handleNavMouseMove(e: MouseEvent) {
  if (!navGlass.value) return
  const rect = navGlass.value.getBoundingClientRect()
  
  // Calculate mouse position relative to nav center (-10 to 10)
  const centerX = rect.left + rect.width / 2
  const centerY = rect.top + rect.height / 2
  targetX.value = ((e.clientX - centerX) / (rect.width / 2)) * 10
  targetY.value = ((e.clientY - centerY) / (rect.height / 2)) * 10
  
  // Clamp values
  targetX.value = Math.max(-10, Math.min(10, targetX.value))
  targetY.value = Math.max(-10, Math.min(10, targetY.value))
  
  startAnimation()
}

function handleNavMouseLeave() {
  targetX.value = 0
  targetY.value = 0
  startAnimation()
}

function animate() {
  // Smooth interpolation (easing)
  const ease = 0.12
  mouseX.value += (targetX.value - mouseX.value) * ease
  mouseY.value += (targetY.value - mouseY.value) * ease
  
  // Update CSS custom properties
  if (highlight.value) {
    highlight.value.style.setProperty('--mx', `${mouseX.value.toFixed(2)}px`)
    highlight.value.style.setProperty('--my', `${mouseY.value.toFixed(2)}px`)
  }
  
  // Continue animation if there's still movement
  if (Math.abs(mouseX.value - targetX.value) > 0.01 || Math.abs(mouseY.value - targetY.value) > 0.01) {
    animationFrame = requestAnimationFrame(animate)
  } else {
    animationFrame = null
  }
}

function startAnimation() {
  if (!animationFrame) {
    animationFrame = requestAnimationFrame(animate)
  }
}

onUnmounted(() => {
  if (animationFrame) {
    cancelAnimationFrame(animationFrame)
  }
})
</script>

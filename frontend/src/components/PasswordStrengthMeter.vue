<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

interface Props {
  modelValue: string
}

const props = defineProps<Props>()
const { t } = useI18n()

interface Rule {
  key: string
  labelKey: string
  passed: boolean
}

const rules = computed<Rule[]>(() => {
  const pw = props.modelValue || ''
  return [
    { key: 'length', labelKey: 'auth.passwordRuleLength', passed: pw.length >= 8 },
    { key: 'lowercase', labelKey: 'auth.passwordRuleLowercase', passed: /[a-z]/.test(pw) },
    { key: 'uppercase', labelKey: 'auth.passwordRuleUppercase', passed: /[A-Z]/.test(pw) },
    { key: 'number', labelKey: 'auth.passwordRuleNumber', passed: /\d/.test(pw) },
    { key: 'symbol', labelKey: 'auth.passwordRuleSymbol', passed: /[^a-zA-Z0-9]/.test(pw) }
  ]
})

const score = computed((): number => rules.value.filter(r => r.passed).length)

const strengthLevel = computed((): number => {
  const s = score.value
  if (s <= 1) return 0
  if (s === 2) return 1
  if (s === 3) return 2
  if (s === 4) return 3
  return 4
})

const strengthLabelKey = computed((): string => {
  const keys = [
    'auth.passwordStrengthVeryWeak',
    'auth.passwordStrengthWeak',
    'auth.passwordStrengthMedium',
    'auth.passwordStrengthStrong',
    'auth.passwordStrengthVeryStrong'
  ]
  return keys[strengthLevel.value]
})

const barColor = computed((): string => {
  const colors = ['var(--danger)', 'var(--warning)', 'var(--warning)', 'var(--success)', 'var(--success)']
  return colors[strengthLevel.value]
})

function segmentColor(index: number): string {
  const level = strengthLevel.value
  if (level < index) return 'var(--border)'
  return barColor.value
}

const successMessage = computed((): string | null => {
  const s = score.value
  if (s === 3) return t('auth.passwordSuccessMedium')
  if (s === 4) return t('auth.passwordSuccessStrong')
  if (s === 5) return t('auth.passwordSuccessVeryStrong')
  return null
})
</script>

<template>
  <div v-if="modelValue" class="password-strength">
    <div class="strength-bars" aria-hidden="true">
      <div
        v-for="i in 4"
        :key="i"
        class="strength-segment"
        :style="{ backgroundColor: segmentColor(i) }"
      />
    </div>
    <p class="strength-label" :style="{ color: barColor }">
      {{ t('auth.passwordStrengthLabel', { level: t(strengthLabelKey) }) }}
    </p>
    <p v-if="successMessage" class="strength-success">
      {{ successMessage }}
    </p>
    <ul class="strength-rules">
      <li
        v-for="rule in rules"
        :key="rule.key"
        class="strength-rule"
        :class="{ 'rule-passed': rule.passed }"
      >
        <svg
          class="rule-icon"
          :style="{ color: rule.passed ? 'var(--success)' : 'var(--text-muted)' }"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          aria-hidden="true"
        >
          <polyline v-if="rule.passed" points="20 6 9 17 4 12" />
          <g v-else>
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </g>
        </svg>
        <span>{{ t(rule.labelKey) }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.password-strength {
  margin-top: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.strength-bars {
  display: flex;
  gap: 0.25rem;
}

.strength-segment {
  flex: 1;
  height: 0.375rem;
  border-radius: 9999px;
  transition: background-color 0.3s ease;
}

.strength-label {
  font-size: 0.75rem;
  font-weight: 500;
  transition: color 0.3s ease;
}

.strength-success {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--success);
  margin: 0;
}

.strength-rules {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.strength-rule {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.6875rem;
  line-height: 1.4;
  color: var(--text-muted);
  transition: color 0.2s ease;
}

.strength-rule.rule-passed {
  color: var(--success);
}

.rule-icon {
  flex-shrink: 0;
  width: 0.75rem;
  height: 0.75rem;
}
</style>

import { onMounted, onUnmounted, ref } from 'vue'
import type { Router } from 'vue-router'

// useKeyboard provides global keyboard shortcut handling.
// Shortcuts:
//   /       -> focus search (when not in input)
//   g n     -> go to generation
//   g g     -> go to gallery
//   g p     -> go to projects
//   g a     -> go to api keys
const SEQUENCE_TIMEOUT_MS = 600

export function useKeyboard(router: Router) {
  const sequence = ref<string | null>(null)
  let seqTimer: ReturnType<typeof setTimeout> | null = null

  function handleKeydown(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null
    const isInput =
      target &&
      (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)

    if (e.key === 'Escape') return

    // '/' focuses search when not in an input.
    if (e.key === '/' && !isInput) {
      e.preventDefault()
      const searchInput = document.querySelector<HTMLInputElement>('input[data-search-input]')
      searchInput?.focus()
      return
    }

    // 'g' prefix sequences (vim-style navigation).
    if (e.key === 'g' && !isInput) {
      if (seqTimer) clearTimeout(seqTimer)
      sequence.value = 'g'
      seqTimer = setTimeout(() => {
        sequence.value = null
        seqTimer = null
      }, SEQUENCE_TIMEOUT_MS)
      return
    }

    if (sequence.value === 'g' && !isInput) {
      if (seqTimer) clearTimeout(seqTimer)
      seqTimer = null
      sequence.value = null
      switch (e.key) {
        case 'n': router.push('/app/create'); break
        case 'g': router.push('/app/assets'); break
        case 'p': router.push('/app/projects'); break
        case 'a': router.push('/app/api-keys'); break
      }
      e.preventDefault()
    }
  }

  onMounted(() => window.addEventListener('keydown', handleKeydown))
  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
    if (seqTimer) clearTimeout(seqTimer)
  })
}

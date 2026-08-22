import { onMounted, onUnmounted, ref } from 'vue'
import type { Router } from 'vue-router'

// useKeyboard provides global keyboard shortcut handling.
// Shortcuts:
//   /       -> focus search (when not in input)
//   g n     -> go to generation
//   g g     -> go to gallery
//   g p     -> go to projects
//   g a     -> go to api keys
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
      const searchInput = document.querySelector<HTMLInputElement>('input[placeholder*="Search"]')
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
      }, 600)
      return
    }

    if (sequence.value === 'g' && !isInput) {
      if (seqTimer) clearTimeout(seqTimer)
      seqTimer = null
      sequence.value = null
      switch (e.key) {
        case 'n': router.push('/create'); break
        case 'g': router.push('/assets'); break
        case 'p': router.push('/projects'); break
        case 'a': router.push('/api-keys'); break
      }
      e.preventDefault()
    }
  }

  onMounted(() => window.addEventListener('keydown', handleKeydown))
  onUnmounted(() => window.removeEventListener('keydown', handleKeydown))
}

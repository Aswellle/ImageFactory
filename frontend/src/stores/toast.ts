import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  type: ToastType
  title: string
  message?: string
  duration: number
  action?: { label: string; onClick: () => void }
  createdAt: number
  paused: boolean
  pausedAt?: number
  remaining: number
}

const DEFAULT_DURATION = 4000
const timers = new Map<number, ReturnType<typeof setTimeout>>()

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  let nextId = 1

  function clearTimer(id: number) {
    const t = timers.get(id)
    if (t !== undefined) {
      clearTimeout(t)
      timers.delete(id)
    }
  }

  function scheduleDismiss(toast: Toast) {
    clearTimer(toast.id)
    if (toast.duration <= 0) return
    if (toast.paused) return
    const t = setTimeout(() => dismiss(toast.id), toast.remaining)
    timers.set(toast.id, t)
  }
function show(options: {
  type: ToastType
  title: string
  message?: string
  duration?: number
  action?: { label: string; onClick: () => void }
}): number {
  const id = nextId++
  const duration = options.duration ?? DEFAULT_DURATION

  // Suppress duplicate toasts with the same title within 1 second
  const now = Date.now()
  const recentDuplicate = toasts.value.find(
    (t) => t.title === options.title && now - t.createdAt < 1000,
  )
  if (recentDuplicate) {
    return recentDuplicate.id
  }

  const toast: Toast = {
    id,
    type: options.type,
    title: options.title,
    message: options.message,
    duration,
    action: options.action,
    createdAt: now,
    paused: false,
    remaining: duration,
  }
  toasts.value.push(toast)
  scheduleDismiss(toast)
  return id
}

  function dismiss(id: number) {
    clearTimer(id)
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function pause(id: number) {
    const toast = toasts.value.find((t) => t.id === id)
    if (!toast || toast.paused || toast.duration <= 0) return
    toast.paused = true
    toast.pausedAt = Date.now()
    clearTimer(id)
  }

  function resume(id: number) {
    const toast = toasts.value.find((t) => t.id === id)
    if (!toast || !toast.paused || toast.duration <= 0) return
    if (toast.pausedAt) {
      toast.remaining = Math.max(0, toast.remaining - (Date.now() - toast.pausedAt))
    }
    toast.paused = false
    toast.pausedAt = undefined
    toast.createdAt = Date.now()
    scheduleDismiss(toast)
  }

  function success(title: string, message?: string, duration?: number): number {
    return show({ type: 'success', title, message, duration })
  }

  function error(title: string, message?: string, duration?: number): number {
    return show({ type: 'error', title, message, duration: duration ?? 6000 })
  }

  function warning(title: string, message?: string, duration?: number): number {
    return show({ type: 'warning', title, message, duration })
  }

  function info(title: string, message?: string, duration?: number): number {
    return show({ type: 'info', title, message, duration })
  }

  function clearAll() {
    for (const id of timers.keys()) clearTimer(id)
    toasts.value = []
  }

  return {
    toasts,
    show,
    dismiss,
    pause,
    resume,
    success,
    error,
    warning,
    info,
    clearAll,
  }
})

export function useToast() {
  const store = useToastStore()
  return {
    toasts: store.toasts,
    show: store.show,
    dismiss: store.dismiss,
    pause: store.pause,
    resume: store.resume,
    success: store.success,
    error: store.error,
    warning: store.warning,
    info: store.info,
    clearAll: store.clearAll,
  }
}

/**
 * Creates a debounced version of a function that delays invoking `fn` until
 * after `wait` milliseconds have elapsed since the last time the debounced
 * function was invoked.
 */
export function debounce<T extends (...args: unknown[]) => void>(fn: T, wait: number): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null

  return function (this: unknown, ...args: Parameters<T>) {
    if (timeout !== null) {
      clearTimeout(timeout)
    }
    timeout = setTimeout(() => {
      fn.apply(this, args)
      timeout = null
    }, wait)
  }
}

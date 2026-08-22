import { onUnmounted } from 'vue'

/**
 * Composable to launch the desktop app via custom protocol.
 * Falls back to the web app if the protocol handler is not registered.
 *
 * Usage:
 *   const { openDesktop } = useDesktopLauncher()
 *   <button @click="openDesktop">Open App</button>
 */
export function useDesktopLauncher() {
  let hiddenFrame: HTMLIFrameElement | null = null
  let cleanup: (() => void) | null = null

  function openDesktop() {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
    const webUrl = window.location.origin + '/app'

    if (isLocal) {
      window.location.href = webUrl
      return
    }

    const desktopUrl = 'imageforge://open'
    let didHide = false

    const onVisibility = () => {
      if (document.visibilityState === 'hidden') {
        didHide = true
        performCleanup()
      }
    }

    cleanup = () => {
      document.removeEventListener('visibilitychange', onVisibility)
      if (hiddenFrame) {
        hiddenFrame.remove()
        hiddenFrame = null
      }
    }

    function performCleanup() {
      if (cleanup) {
        cleanup()
        cleanup = null
      }
    }

    document.addEventListener('visibilitychange', onVisibility)

    // Attempt protocol launch via hidden iframe.
    hiddenFrame = document.createElement('iframe')
    hiddenFrame.style.display = 'none'
    document.body.appendChild(hiddenFrame)
    try {
      hiddenFrame.contentWindow?.location.assign(desktopUrl)
    } catch {
      // Protocol not handled — fall through to web.
    }

    // Fallback: if page is still visible after 2s, redirect to web app.
    setTimeout(() => {
      if (!didHide) {
        performCleanup()
        window.location.href = webUrl
      }
    }, 2000)
  }

  onUnmounted(() => {
    if (cleanup) {
      cleanup()
    }
  })

  return { openDesktop }
}

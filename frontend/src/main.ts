import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { useAppStore } from './stores/app'
import { i18n } from './i18n'
import './styles/main.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)

// Global error handler: log errors to console in development, avoid crashing the app.
app.config.errorHandler = (err, _instance, info) => {
  if (import.meta.env.DEV) {
    console.error(`[Vue Error] ${info}:`, err)
  }
  // In production, you could send this to an error tracking service.
}

// Initialize theme before mount to avoid a flash of the wrong theme.
useAppStore().initTheme()

app.mount('#app')

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

// Initialize theme before mount to avoid a flash of the wrong theme.
useAppStore().initTheme()

app.mount('#app')

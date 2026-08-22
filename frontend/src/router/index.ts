import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Lazy-loaded views keep the initial bundle small.
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/views/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
      { path: 'create', name: 'create', component: () => import('@/views/GenerationView.vue') },
      { path: 'assets', name: 'assets', component: () => import('@/views/GalleryView.vue') },
      { path: 'assets/:id', name: 'asset-detail', component: () => import('@/views/AssetDetailView.vue') },
      { path: 'projects', name: 'projects', component: () => import('@/views/ProjectView.vue') },
      { path: 'projects/:id', name: 'project-detail', component: () => import('@/views/ProjectDetailView.vue') },
    ],
  },
  // Fallback: send unknown routes home (or to login).
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth guard: enforce the auth/non-auth split. The auth store hydrates its
// token from localStorage on init, so a refresh keeps the session.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isAuthenticated) return { name: 'login', query: { redirect: to.fullPath } }
  return true
})

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
      { path: 'collections', name: 'collections', component: () => import('@/views/CollectionsView.vue') },
      { path: 'collections/:id', name: 'collection-detail', component: () => import('@/views/CollectionDetailView.vue') },
      { path: 'tags', name: 'tags', component: () => import('@/views/TagsView.vue') },
      { path: 'templates', name: 'templates', component: () => import('@/views/PromptTemplatesView.vue') },
      { path: 'favorites', name: 'favorites', component: () => import('@/views/FavoritesView.vue') },
      { path: 'usage', name: 'usage', component: () => import('@/views/UsageView.vue') },
    ],
  },
  // Admin panel. Lazy-loaded; access enforced by requiresAdmin meta + the
  // auth guard below (which also checks auth.isAdmin client-side).
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      { path: '', redirect: { name: 'admin-dashboard' } },
      { path: 'dashboard', name: 'admin-dashboard', component: () => import('@/views/admin/DashboardView.vue') },
      { path: 'users', name: 'admin-users', component: () => import('@/views/admin/UsersView.vue') },
      { path: 'jobs', name: 'admin-jobs', component: () => import('@/views/admin/JobsView.vue') },
      { path: 'api-keys', name: 'admin-api-keys', component: () => import('@/views/admin/ApiKeysView.vue') },
    ],
  },
  // Fallback: send unknown routes home (or to login).
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth guard: enforce the auth/non-auth split and the admin-only gate. The auth
// store hydrates its token from localStorage on init, so a refresh keeps the
// session. Admin routes additionally require the admin role; non-admins are
// bounced back to the dashboard.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isAuthenticated) return { name: 'login', query: { redirect: to.fullPath } }
  if (to.meta.requiresAdmin && !auth.isAdmin) return { name: 'dashboard' }
  return true
})

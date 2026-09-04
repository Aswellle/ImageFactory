import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Lazy-loaded views keep the initial bundle small.
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'landing',
    component: () => import('@/views/LandingView.vue'),
    meta: { public: true, layout: 'landing' },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('@/views/ForgotPasswordView.vue'),
    meta: { public: true },
  },
  {
    path: '/app',
    component: () => import('@/views/AppLayout.vue'),
    meta: { requiresAuth: true, layout: 'app' },
    children: [
      { path: '', name: 'dashboard', component: () => import(/* webpackPrefetch: true */ '@/views/DashboardView.vue') },
      { path: 'create', name: 'create', component: () => import(/* webpackPrefetch: true */ '@/views/GenerationView.vue') },
      { path: 'assets', name: 'assets', component: () => import(/* webpackPrefetch: true */ '@/views/GalleryView.vue') },
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
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'app' },
    children: [
      { path: '', redirect: { name: 'admin-dashboard' } },
      { path: 'dashboard', name: 'admin-dashboard', component: () => import('@/views/admin/DashboardView.vue') },
      { path: 'users', name: 'admin-users', component: () => import('@/views/admin/UsersView.vue') },
      { path: 'jobs', name: 'admin-jobs', component: () => import('@/views/admin/JobsView.vue') },
      { path: 'api-keys', name: 'admin-api-keys', component: () => import('@/views/admin/ApiKeysView.vue') },
    ],
  },
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
  // Redirect authenticated users from landing to app dashboard.
  if (to.name === 'landing' && auth.isAuthenticated) return { name: 'dashboard' }
  return true
})

import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guest?: boolean
  }
}

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/sign-in',
      name: 'sign-in',
      component: () => import('@/views/SignInView.vue'),
      meta: { guest: true },
    },
    {
      path: '/sign-up',
      name: 'sign-up',
      component: () => import('@/views/SignUpView.vue'),
      meta: { guest: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { guest: true },
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { guest: true },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/cron',
      name: 'cron-monitors',
      component: () => import('@/views/CronMonitorListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/cron/new',
      name: 'cron-monitor-create',
      component: () => import('@/views/CronMonitorCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/cron/:id',
      name: 'cron-monitor-detail',
      component: () => import('@/views/CronMonitorDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/cron/:id/edit',
      name: 'cron-monitor-edit',
      component: () => import('@/views/CronMonitorEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/uptime',
      name: 'uptime-monitors',
      component: () => import('@/views/UptimeMonitorListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/uptime/new',
      name: 'uptime-monitor-create',
      component: () => import('@/views/UptimeMonitorCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/uptime/:id',
      name: 'uptime-monitor-detail',
      component: () => import('@/views/UptimeMonitorDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/uptime/:id/edit',
      name: 'uptime-monitor-edit',
      component: () => import('@/views/UptimeMonitorEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

let authInitialized = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!authInitialized) {
    await auth.init()
    authInitialized = true
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'sign-in' }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
})

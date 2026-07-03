import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useConsent } from '@/lib/consent'
import { trackPageview } from '@/lib/analytics'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guest?: boolean
  }
}

export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      return { el: to.hash, behavior: 'smooth' }
    }
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('@/views/PricingView.vue'),
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('@/views/DocsView.vue'),
    },
    {
      path: '/faq',
      name: 'faq',
      component: () => import('@/views/FaqView.vue'),
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('@/views/AboutView.vue'),
    },
    {
      path: '/terms',
      name: 'terms',
      component: () => import('@/views/TermsView.vue'),
    },
    {
      path: '/privacy',
      name: 'privacy',
      component: () => import('@/views/PrivacyView.vue'),
    },
    {
      path: '/refund',
      name: 'refund-policy',
      component: () => import('@/views/RefundPolicyView.vue'),
    },
    {
      path: '/blog',
      name: 'blog',
      component: () => import('@/views/BlogView.vue'),
    },
    {
      path: '/blog/:slug',
      name: 'blog-post',
      component: () => import('@/views/BlogPostView.vue'),
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
      path: '/monitors/ssl',
      name: 'ssl-monitors',
      component: () => import('@/views/SSLMonitorListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/ssl/new',
      name: 'ssl-monitor-create',
      component: () => import('@/views/SSLMonitorCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/ssl/:id',
      name: 'ssl-monitor-detail',
      component: () => import('@/views/SSLMonitorDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/ssl/:id/edit',
      name: 'ssl-monitor-edit',
      component: () => import('@/views/SSLMonitorEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/domain',
      name: 'domain-monitors',
      component: () => import('@/views/DomainMonitorListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/domain/new',
      name: 'domain-monitor-create',
      component: () => import('@/views/DomainMonitorCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/domain/:id',
      name: 'domain-monitor-detail',
      component: () => import('@/views/DomainMonitorDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/domain/:id/edit',
      name: 'domain-monitor-edit',
      component: () => import('@/views/DomainMonitorEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/port',
      name: 'port-monitors',
      component: () => import('@/views/PortMonitorListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/port/new',
      name: 'port-monitor-create',
      component: () => import('@/views/PortMonitorCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/port/:id',
      name: 'port-monitor-detail',
      component: () => import('@/views/PortMonitorDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/monitors/port/:id/edit',
      name: 'port-monitor-edit',
      component: () => import('@/views/PortMonitorEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/status-pages',
      name: 'status-pages',
      component: () => import('@/views/StatusPageListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/status-pages/new',
      name: 'status-page-create',
      component: () => import('@/views/StatusPageCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/status-pages/:id',
      name: 'status-page-detail',
      component: () => import('@/views/StatusPageDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/status-pages/:id/edit',
      name: 'status-page-edit',
      component: () => import('@/views/StatusPageEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/maintenance',
      name: 'maintenance',
      component: () => import('@/views/MaintenanceWindowListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/maintenance/new',
      name: 'maintenance-create',
      component: () => import('@/views/MaintenanceWindowCreateView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/maintenance/:id/edit',
      name: 'maintenance-edit',
      component: () => import('@/views/MaintenanceWindowEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/billing',
      name: 'billing',
      component: () => import('@/views/BillingView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/accept-terms',
      name: 'accept-terms',
      component: () => import('@/views/AcceptTermsView.vue'),
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

  if (
    to.meta.requiresAuth &&
    auth.isAuthenticated &&
    auth.user?.needsTermsAcceptance &&
    to.name !== 'accept-terms'
  ) {
    return { name: 'accept-terms' }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }

  if (to.name === 'home' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
})

router.afterEach((to) => {
  const { status } = useConsent()
  if (status.value === 'granted') trackPageview(to.fullPath)
})

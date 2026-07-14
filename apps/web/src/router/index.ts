import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useConsent } from '@/lib/consent'
import { trackPageview } from '@/lib/analytics'
import { routes } from './routes'

export { routes } from './routes'

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
  routes,
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

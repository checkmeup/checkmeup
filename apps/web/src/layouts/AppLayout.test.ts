import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive } from 'vue'
import AppLayout from './AppLayout.vue'
import { appVersion } from '@/lib/version'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => reactive({ path: '/dashboard' }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { email: 'andrew@example.com' }, clear: vi.fn() }),
}))

function mountLayout() {
  return mount(AppLayout, {
    slots: { default: '<p>page</p>' },
    global: {
      stubs: { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } },
    },
  })
}

describe('AppLayout', () => {
  it('shows the running version in the footer, linking to the changelog', () => {
    const link = mountLayout().find('footer a')
    expect(link.text()).toBe(appVersion)
    expect(link.attributes('href')).toBe('/changelog')
  })

  it('keeps the version out of the account menu', async () => {
    const wrapper = mountLayout()
    await wrapper.find('button[aria-haspopup]').trigger('click')
    const menu = wrapper.find('[role="menu"]')
    expect(menu.text()).toContain('andrew@example.com')
    expect(menu.text()).not.toContain(appVersion)
  })
})

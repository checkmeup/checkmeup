import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import NotificationChannelPicker from './NotificationChannelPicker.vue'

const channels = ref<unknown[] | undefined>(undefined)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channels, isPending: ref(false) }),
}))

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

const list = [
  { id: 'tg', type: 'telegram', name: 'Team chat', config: {}, enabled: true, createdAt: '' },
  { id: 'old', type: 'email', name: 'Old inbox', config: {}, enabled: false, createdAt: '' },
  { id: 'ops', type: 'email', name: 'Ops', config: {}, enabled: true, createdAt: '' },
]

// Mounted under a parent holding the v-model, as the monitor forms do.
function mountPicker(initial: string[], newMonitor = false) {
  const model = ref<string[]>(initial)
  const wrapper = mount({
    components: { NotificationChannelPicker },
    setup: () => ({ model, newMonitor }),
    template: '<NotificationChannelPicker v-model="model" :new-monitor="newMonitor" />',
  })
  return { wrapper, model }
}

describe('NotificationChannelPicker', () => {
  it("starts a new monitor with every enabled channel selected once they've loaded", async () => {
    channels.value = undefined
    const { wrapper, model } = mountPicker([], true)
    channels.value = list
    await nextTick()

    expect(model.value).toEqual(['tg', 'ops'])
    expect(wrapper.text()).toContain('2 selected')
    expect(wrapper.text()).not.toContain('fall back to your account email')
  })

  it("doesn't select again what the user cleared when the list reloads", async () => {
    channels.value = list
    const { wrapper, model } = mountPicker([], true)
    await nextTick()
    model.value = []
    channels.value = [...list]
    await nextTick()

    expect(model.value).toEqual([])
    expect(wrapper.text()).toContain('fall back to your account email')
  })

  it("leaves an existing monitor's channels as they are", async () => {
    channels.value = list
    const { model } = mountPicker(['ops'])
    await nextTick()

    expect(model.value).toEqual(['ops'])
  })
})

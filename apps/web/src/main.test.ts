import { beforeEach, describe, expect, it, vi } from 'vitest'

const {
  mountMock,
  useMock,
  createAppMock,
  piniaInstance,
  createPiniaMock,
  routerInstance,
  queryClientCtor,
} = vi.hoisted(() => {
  const piniaInstance = { __brand: 'pinia' }
  const routerInstance = { __brand: 'router' }
  const mountMock = vi.fn()
  const useMock = vi.fn()
  return {
    mountMock,
    useMock,
    createAppMock: vi.fn(),
    piniaInstance,
    createPiniaMock: vi.fn(() => piniaInstance),
    routerInstance,
    queryClientCtor: vi.fn(),
  }
})

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue')>()
  return { ...actual, createApp: createAppMock }
})

vi.mock('pinia', () => ({ createPinia: createPiniaMock }))

vi.mock('@tanstack/vue-query', () => ({
  VueQueryPlugin: { __brand: 'vue-query-plugin' },
  QueryClient: vi.fn().mockImplementation(function (this: object, options: unknown) {
    queryClientCtor(options)
    Object.assign(this, { __brand: 'query-client' })
  }),
}))

vi.mock('./App.vue', () => ({ default: { __brand: 'app-component' } }))
vi.mock('./router', () => ({ router: routerInstance }))
vi.mock('./style.css', () => ({}))

describe('main.ts', () => {
  beforeEach(() => {
    vi.resetModules()
    mountMock.mockClear()
    useMock.mockClear().mockReturnValue({ use: useMock, mount: mountMock })
    createAppMock.mockClear().mockReturnValue({ use: useMock, mount: mountMock })
    createPiniaMock.mockClear()
    queryClientCtor.mockClear()
  })

  it('wires pinia, router, and vue-query onto the app, in order, then mounts to #app', async () => {
    await import('./main')

    expect(createAppMock).toHaveBeenCalledTimes(1)
    expect(createAppMock.mock.calls[0][0]).toEqual({ __brand: 'app-component' })

    expect(useMock).toHaveBeenCalledTimes(3)
    expect(useMock.mock.calls[0][0]).toBe(piniaInstance)
    expect(useMock.mock.calls[1][0]).toBe(routerInstance)
    expect(useMock.mock.calls[2][0]).toEqual({ __brand: 'vue-query-plugin' })
    expect(useMock.mock.calls[2][1]).toMatchObject({ queryClient: { __brand: 'query-client' } })

    expect(mountMock).toHaveBeenCalledExactlyOnceWith('#app')
  })

  it('configures the query client to retry once and skip refetch-on-focus', async () => {
    await import('./main')

    expect(queryClientCtor).toHaveBeenCalledExactlyOnceWith({
      defaultOptions: {
        queries: { retry: false, refetchOnWindowFocus: false },
      },
    })
  })
})

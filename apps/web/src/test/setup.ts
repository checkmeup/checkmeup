import { config } from '@vue/test-utils'
import { createHead } from '@unhead/vue/client'

// useSeo() (src/composables/useSeo.ts) calls useHead()/useSeoMeta(), which
// throw outside an app that has the @unhead/vue plugin installed — every
// @vue/test-utils mount() needs it, so install it globally here instead of
// per test file.
config.global.plugins.push(createHead())

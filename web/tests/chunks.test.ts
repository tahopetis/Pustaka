import { describe, expect, it } from 'vitest'
import { chunkNameFor } from '../config/chunks'

describe('chunkNameFor', () => {
  it.each([
    // Vue runtime packages must land in the vendor chunk, including scoped ones
    // (the previous regex required a doubled slash and silently missed these).
    ['/app/node_modules/vue/dist/vue.runtime.esm-bundler.js', 'vendor'],
    ['/app/node_modules/@vue/runtime-dom/dist/runtime-dom.esm-bundler.js', 'vendor'],
    ['/app/node_modules/@vue/shared/dist/shared.esm-bundler.js', 'vendor'],
    ['/app/node_modules/@vue/devtools-api/dist/index.js', 'vendor'],
    ['/app/node_modules/vue-router/dist/vue-router.js', 'vendor'],
    ['/app/node_modules/pinia/dist/pinia.mjs', 'vendor'],
    // UI libraries.
    ['/app/node_modules/@headlessui/vue/dist/headlessui.esm.js', 'ui'],
    ['/app/node_modules/@heroicons/vue/24/solid/esm/HomeIcon.js', 'ui'],
    // Chart libraries and their production dependencies.
    ['/app/node_modules/d3/dist/d3.js', 'charts'],
    ['/app/node_modules/d3-array/src/index.js', 'charts'],
    ['/app/node_modules/d3-scale/src/index.js', 'charts'],
    ['/app/node_modules/vis-network/standalone/esm/vis-network.js', 'charts'],
    ['/app/node_modules/internmap/src/index.js', 'charts'],
    ['/app/node_modules/@egjs/hammerjs/dist/hammer.esm.js', 'charts'],
    // Everything else is left to Vite's chunking.
    ['/app/node_modules/axios/index.js', undefined],
    ['/app/src/main.ts', undefined],
  ] as const)('%s -> %s', (id, expected) => {
    expect(chunkNameFor(id)).toBe(expected)
  })
})

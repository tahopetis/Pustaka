// Chunk groups for the Vite/Rolldown production build.
//
// Vite 8 bundles with Rolldown: `build.rolldownOptions.output.codeSplitting`
// replaces the deprecated `build.rollupOptions.output.manualChunks`. The groups
// are kept as data so the mapping is unit-testable (see tests/chunks.test.ts).

export interface ChunkGroup {
  name: string
  test: RegExp
}

const VENDOR_PACKAGES = ['vue', 'vue-router', 'pinia'].join('|')

// d3-* siblings and the graph libraries' direct runtime dependencies.
const CHART_PACKAGES = [
  'd3',
  'd3-[^/]+',
  'vis-network',
  'vis-data',
  'vis-util',
  'internmap',
  'delaunator',
  'robust-predicates',
  'rw',
  'iconv-lite',
  'safer-buffer',
  '@egjs/hammerjs',
  'keycharm',
].join('|')

/** Groups are evaluated in order; the first match wins. */
export const chunkGroups: readonly ChunkGroup[] = [
  {
    name: 'vendor',
    test: new RegExp(`node_modules/(?:${VENDOR_PACKAGES}|@vue/[^/]+)/`),
  },
  {
    name: 'ui',
    test: /node_modules\/(?:@headlessui\/vue|@heroicons\/vue)\//,
  },
  {
    name: 'charts',
    test: new RegExp(`node_modules/(?:${CHART_PACKAGES})/`),
  },
]

/** Returns the chunk name for a module id, or undefined to let Vite decide. */
export function chunkNameFor(id: string): string | undefined {
  for (const group of chunkGroups) {
    if (group.test.test(id)) return group.name
  }
  return undefined
}

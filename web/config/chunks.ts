// Chunk classifier for the Vite/Rolldown production build.
//
// Kept as a pure function so the package grouping is unit-testable without
// running a build (see tests/chunks.test.ts). Vite 8/Rolldown requires the
// function form of `manualChunks`.

type PackageMatcher = (pkgName: string) => boolean

const CHART_PACKAGES = new Set([
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
])

const CHUNK_GROUPS: ReadonlyArray<readonly [string, PackageMatcher]> = [
  [
    'vendor',
    pkg =>
      pkg === 'vue' ||
      pkg === 'vue-router' ||
      pkg === 'pinia' ||
      pkg.startsWith('@vue/'),
  ],
  ['ui', pkg => pkg === '@headlessui/vue' || pkg === '@heroicons/vue'],
  ['charts', pkg => pkg === 'd3' || pkg.startsWith('d3-') || CHART_PACKAGES.has(pkg)],
]

/** Returns the chunk name for a module id, or undefined to let Vite decide. */
export function chunkNameFor(id: string): string | undefined {
  const match = id.match(/node_modules\/(.+)$/)
  if (!match) return undefined

  const rest = match[1]
  const pkgName = rest.startsWith('@')
    ? rest.split('/').slice(0, 2).join('/')
    : rest.split('/')[0]

  for (const [chunkName, matches] of CHUNK_GROUPS) {
    if (matches(pkgName)) return chunkName
  }
  return undefined
}

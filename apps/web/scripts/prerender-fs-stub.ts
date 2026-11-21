// node:fs for running scripts/prerender.mts in a test: it reads the source
// index.html in place of the built dist/index.html (only the hashed asset
// tags differ) and its writes land in `written` instead of dist/, so each
// page is checked as a crawler that doesn't run JavaScript fetches it,
// without a full `vite build`.
import type * as NodeFs from 'node:fs'

// Kept on globalThis so a second run after vi.resetModules() fills the same map.
const store = globalThis as typeof globalThis & { prerenderWritten?: Map<string, string | Buffer> }
export const written = (store.prerenderWritten ??= new Map<string, string | Buffer>())

export function stubNodeFs(fs: typeof NodeFs) {
  const readFileSync = ((path: string, ...rest: unknown[]) =>
    String(path).endsWith('dist/index.html')
      ? fs.readFileSync(new URL('../index.html', import.meta.url), 'utf-8')
      : (fs.readFileSync as (...args: unknown[]) => unknown)(
          path,
          ...rest,
        )) as typeof fs.readFileSync
  const writeFileSync = ((path: string, data: string | Uint8Array) => {
    written.set(String(path), typeof data === 'string' ? data : Buffer.from(data))
  }) as typeof fs.writeFileSync
  const mkdirSync = (() => undefined) as typeof fs.mkdirSync
  const stubs = { readFileSync, writeFileSync, mkdirSync }
  return { ...fs, ...stubs, default: { ...fs, ...stubs } }
}

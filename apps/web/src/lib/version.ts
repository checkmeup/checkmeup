// Build-time version, baked in via VITE_APP_VERSION (see Dockerfile/deploy.yml).
// `git describe --tags --always --dirty` produces e.g. "v1.22.2-11-g6feb70e"
// once commits land after a tag — shortVersion strips that down to just the
// tag ("v1.22.2") for display, since the commit count/hash isn't meaningful
// to a user and only the exact-tag case ("v1.22.2") has no suffix to strip.
export function shortVersion(raw: string): string {
  return raw.match(/^v\d+\.\d+\.\d+/)?.[0] ?? raw
}

export const appVersion = shortVersion(import.meta.env.VITE_APP_VERSION ?? 'dev')

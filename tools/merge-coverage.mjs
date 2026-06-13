// tools/merge-coverage.mjs
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'fs'
import { glob } from 'glob'
import { dirname, resolve, relative } from 'path'

const ROOT = process.cwd()

const files = await glob('**/coverage/lcov.info', {
  ignore: ['**/node_modules/**', '**/.turbo/**', 'coverage/lcov.info'],
})

const merged = files
  .map((lcovPath) => {
    const pkgDir = resolve(ROOT, dirname(dirname(lcovPath))) // apps/ingest
    // eslint-disable-next-line security/detect-non-literal-fs-filename
    const content = readFileSync(lcovPath, 'utf8')

    return content
      .split('\n')
      .map((line) => {
        if (!line.startsWith('SF:')) return line
        const filePath = line.slice(3)
        // gcov2lcov (Go) already emits repo-root-relative paths; JS tools emit package-relative paths
        // eslint-disable-next-line security/detect-non-literal-fs-filename
        const absPath = existsSync(resolve(ROOT, filePath))
          ? resolve(ROOT, filePath)
          : resolve(pkgDir, filePath)
        const relPath = relative(ROOT, absPath)
        return `SF:${relPath}`
      })
      .join('\n')
  })
  .join('\n')

mkdirSync('coverage', { recursive: true })
writeFileSync('coverage/lcov.info', merged)
console.log(`Merged ${files.length} files → coverage/lcov.info`)

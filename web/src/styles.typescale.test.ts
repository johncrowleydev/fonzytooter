import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

/**
 * Guards the 12px floor documented in styles.css.
 *
 * This is worth a test because the failure is silent: `--text-2xs` no longer exists, so a
 * reintroduced `text-2xs` generates no utility at all. The class would look deliberate in the JSX
 * and simply not apply a font size.
 */
const sourceRoot = dirname(fileURLToPath(import.meta.url))

function collectSourceFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)

    if (statSync(path).isDirectory()) {
      // Generated API clients are not hand-styled.
      return entry === 'generated' ? [] : collectSourceFiles(path)
    }

    return /\.tsx?$/.test(path) && !/\.test\.tsx?$/.test(path) ? [path] : []
  })
}

const sourceFiles = collectSourceFiles(sourceRoot).map((path) => ({
  path: path.slice(sourceRoot.length + 1).replaceAll('\\', '/'),
  contents: readFileSync(path, 'utf8'),
}))

const stylesheet = readFileSync(join(sourceRoot, 'styles.css'), 'utf8')

describe('type scale floor', () => {
  it('found source files to check', () => {
    expect(sourceFiles.length).toBeGreaterThan(20)
  })

  it('does not define a size below the 12px floor', () => {
    expect(stylesheet).not.toMatch(/--text-2xs\s*:/)
  })

  it('never uses text-2xs, which would silently apply no font size', () => {
    const offenders = sourceFiles
      .filter(({ contents }) => /\btext-2xs\b/.test(contents))
      .map(({ path }) => path)

    expect(offenders).toEqual([])
  })

  it('does not reintroduce a sub-12px arbitrary font size', () => {
    const offenders: string[] = []

    for (const { path, contents } of sourceFiles) {
      for (const [utility, size, unit] of contents.matchAll(
        /\btext-\[(\d+(?:\.\d+)?)(px|rem)\]/g,
      ) as Iterable<[string, string, string]>) {
        const pixels = unit === 'rem' ? Number(size) * 16 : Number(size)

        if (pixels < 12) {
          offenders.push(`${path}: ${utility}`)
        }
      }
    }

    expect(offenders).toEqual([])
  })
})

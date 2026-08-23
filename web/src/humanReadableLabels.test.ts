import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

/**
 * Curriculum identifiers -- `python.execution-model`, `01-scientific-python` -- are internal
 * vocabulary. They were reaching the UI in six places, so this guards the rule going forward:
 * a learner should only read authored prose, never a key.
 *
 * The check is deliberately narrow. It looks at expressions in JSX *children* position, where the
 * value becomes visible text, and ignores attributes: `key={module.id}` and
 * `to={lessonPath(course.id, ...)}` are correct uses of the same field.
 */
const sourceRoot = dirname(fileURLToPath(import.meta.url))

function collectComponentFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)

    if (statSync(path).isDirectory()) {
      return entry === 'generated' ? [] : collectComponentFiles(path)
    }

    return path.endsWith('.tsx') && !path.endsWith('.test.tsx') ? [path] : []
  })
}

const componentFiles = collectComponentFiles(sourceRoot).map((path) => ({
  path: path.slice(sourceRoot.length + 1).replaceAll('\\', '/'),
  contents: readFileSync(path, 'utf8'),
}))

/** `>{module.id}` or `>{moduleId}` -- an identifier rendered as visible text. */
const RENDERED_IDENTIFIER = />\s*\{\s*[A-Za-z_$][\w$.?]*(?:\.id|Id)\s*\}/g

/** `moduleId.replaceAll('-', ' ')` -- dressing an identifier up as prose. */
const DE_SLUGIFIED_IDENTIFIER = /[\w$.?]*(?:\.id|Id)\s*\.\s*replace(?:All)?\s*\(/g

function findMatches(pattern: RegExp) {
  return componentFiles.flatMap(({ path, contents }) =>
    Array.from(contents.matchAll(pattern), (match) => `${path}: ${match[0].trim()}`),
  )
}

describe('human-readable labels', () => {
  it('found component files to check', () => {
    expect(componentFiles.length).toBeGreaterThan(20)
  })

  it('never renders a raw identifier as visible text', () => {
    expect(findMatches(RENDERED_IDENTIFIER)).toEqual([])
  })

  it('never reformats an identifier to look like prose', () => {
    expect(findMatches(DE_SLUGIFIED_IDENTIFIER)).toEqual([])
  })
})

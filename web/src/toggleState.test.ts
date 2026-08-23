import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

/**
 * Guards that a button whose appearance is chosen from state also reports that state.
 *
 * Seventeen toggle groups across the lesson interactives signalled their selection purely through
 * `variant={x === y ? … : …}`. Nothing was broken visually, which is why it went unnoticed through
 * four pull requests -- but the selected item was announced identically to the unselected ones, and
 * three older components had already established `aria-pressed` as the convention.
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

/** A `variant` chosen by a conditional is a toggle, whatever the two branches happen to be. */
const CONDITIONAL_VARIANT = /variant=\{[^}]*\?/

const componentFiles = collectComponentFiles(sourceRoot).map((path) => ({
  path: path.slice(sourceRoot.length + 1).replaceAll('\\', '/'),
  lines: readFileSync(path, 'utf8').split('\n'),
}))

describe('toggle buttons report their state', () => {
  it('found component files to check', () => {
    expect(componentFiles.length).toBeGreaterThan(20)
  })

  it('pairs every conditional variant with a pressed prop', () => {
    const offenders: string[] = []

    for (const { path, lines } of componentFiles) {
      lines.forEach((line, index) => {
        if (!CONDITIONAL_VARIANT.test(line)) return

        // The prop is written on the line immediately after, which is how Prettier lays it out.
        const following = lines.slice(index + 1, index + 3).join(' ')
        if (!/\bpressed=/.test(following)) {
          offenders.push(`${path}:${index + 1}: ${line.trim()}`)
        }
      })
    }

    expect(offenders).toEqual([])
  })

  it('actually found the toggles it is guarding', () => {
    // Without this, the check above would pass trivially if the pattern ever stopped matching.
    const toggles = componentFiles.flatMap(({ lines }) =>
      lines.filter((line) => CONDITIONAL_VARIANT.test(line)),
    )

    expect(toggles.length).toBeGreaterThanOrEqual(17)
  })
})

import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

/**
 * The palette in styles.css is derived to hit WCAG AA rather than chosen by eye, so the ratios are
 * a contract. This reads the stylesheet and re-measures it: a hand-edited token that looks fine on
 * one background but fails on another gets caught here instead of in the UI.
 *
 * Read from disk rather than imported: `styles.css?raw` resolves to an empty string because the
 * Tailwind Vite plugin claims the stylesheet before the raw loader sees it. This is the only place
 * that needs Node APIs, which is why `node` was added to tsconfig `types`.
 */
const stylesheetPath = join(dirname(fileURLToPath(import.meta.url)), 'styles.css')
const stylesheet = readFileSync(stylesheetPath, 'utf8')

const AA_NORMAL_TEXT = 4.5

function channelToLinear(value: number) {
  const srgb = value / 255
  return srgb <= 0.03928 ? srgb / 12.92 : Math.pow((srgb + 0.055) / 1.055, 2.4)
}

function relativeLuminance(hex: string) {
  const normalized = hex.replace('#', '')
  const [r, g, b] = [0, 2, 4]
    .map((index) => parseInt(normalized.slice(index, index + 2), 16))
    .map(channelToLinear)

  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

function contrastRatio(foreground: string, background: string) {
  const a = relativeLuminance(foreground)
  const b = relativeLuminance(background)
  const [lighter, darker] = a > b ? [a, b] : [b, a]

  return (lighter + 0.05) / (darker + 0.05)
}

/** Pulls the `--ft-*` declarations out of one theme's rule block. */
function readThemeTokens(selector: string) {
  const start = stylesheet.indexOf(selector)
  expect(start, `${selector} block is missing from styles.css`).toBeGreaterThan(-1)

  const block = stylesheet.slice(start, stylesheet.indexOf('}', start))
  const tokens: Record<string, string> = {}

  for (const [, name, value] of block.matchAll(/--ft-([\w-]+):\s*(#[0-9a-fA-F]{6})\s*;/g)) {
    tokens[name] = value
  }

  return tokens
}

function readRootToken(name: string) {
  const match = stylesheet.match(new RegExp(`--color-${name}:\\s*(#[0-9a-fA-F]{6})\\s*;`))
  expect(match, `--color-${name} is missing from styles.css`).toBeTruthy()

  return match![1]
}

const themes = {
  dark: readThemeTokens(":root[data-theme='dark']"),
  light: readThemeTokens(":root[data-theme='light']"),
}

/** Opaque surfaces that content actually sits on. Translucent tokens cannot be measured directly. */
const surfaceTokens = ['canvas', 'panel', 'panel-soft'] as const
const textTokens = ['ink', 'body', 'muted', 'faint'] as const
const accentTokens = [
  'accent-teal',
  'accent-teal-light',
  'accent-gold',
  'accent-coral',
  'accent-violet',
  'accent-blue',
  'accent-slate',
] as const

describe.each(Object.entries(themes))('%s theme', (themeName, tokens) => {
  it('defines every token the contract covers', () => {
    for (const name of [...surfaceTokens, ...textTokens, ...accentTokens]) {
      expect(tokens[name], `--ft-${name} missing in ${themeName}`).toMatch(/^#[0-9a-fA-F]{6}$/)
    }
  })

  describe.each(surfaceTokens)('on %s', (surface) => {
    it.each([...textTokens, ...accentTokens])('%s clears AA', (token) => {
      const ratio = contrastRatio(tokens[token], tokens[surface])

      expect(
        ratio,
        `--ft-${token} on --ft-${surface} in ${themeName} is ${ratio.toFixed(2)}:1`,
      ).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
    })
  })
})

describe('role-specific pairings', () => {
  it('keeps brand-ink readable on every vivid fill', () => {
    // Solid fills are theme-invariant precisely because this pairing holds in both themes.
    const ink = readRootToken('brand-ink')

    for (const hue of ['teal', 'teal-light', 'gold', 'coral', 'violet', 'blue']) {
      const ratio = contrastRatio(ink, readRootToken(`brand-${hue}`))

      expect(ratio, `brand-ink on brand-${hue} is ${ratio.toFixed(2)}:1`).toBeGreaterThanOrEqual(
        AA_NORMAL_TEXT,
      )
    }
  })

  it('keeps code text readable on the code surface', () => {
    const ratio = contrastRatio(readRootToken('code-ink'), readRootToken('code-surface'))

    expect(ratio).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
  })

  it('deepens the teal hover partner in light mode and lightens it in dark mode', () => {
    // hover:text-accent-teal-light must move away from the surface, not toward it.
    const lightBase = contrastRatio(themes.light['accent-teal'], themes.light.panel)
    const lightHover = contrastRatio(themes.light['accent-teal-light'], themes.light.panel)
    expect(lightHover).toBeGreaterThan(lightBase)

    const darkBase = contrastRatio(themes.dark['accent-teal'], themes.dark.panel)
    const darkHover = contrastRatio(themes.dark['accent-teal-light'], themes.dark.panel)
    expect(darkHover).toBeGreaterThan(darkBase)
  })
})

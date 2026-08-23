import { act, cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ThemeProvider, useTheme } from './ThemeContext'
import { THEME_STORAGE_KEY } from './theme'

/**
 * jsdom has no `matchMedia`. This stub reports a fixed system preference and records listeners so
 * a live OS light/dark switch can be simulated.
 */
function stubMatchMedia(systemPrefersLight: boolean) {
  const listeners = new Set<(event: MediaQueryListEvent) => void>()

  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: systemPrefersLight,
      media: query,
      addEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) => {
        listeners.add(listener)
      },
      removeEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) => {
        listeners.delete(listener)
      },
    })),
  )

  return {
    emitSystemChange(prefersLight: boolean) {
      act(() => {
        for (const listener of listeners) {
          listener({ matches: prefersLight } as MediaQueryListEvent)
        }
      })
    },
  }
}

function ThemeProbe() {
  const { theme, followsSystem, toggleTheme } = useTheme()

  return (
    <button type="button" onClick={toggleTheme}>
      {theme}
      {followsSystem ? ' (system)' : ' (chosen)'}
    </button>
  )
}

function renderProbe() {
  render(
    <ThemeProvider>
      <ThemeProbe />
    </ThemeProvider>,
  )
}

function probeText() {
  return screen.getByRole('button').textContent
}

beforeEach(() => {
  localStorage.clear()
  document.documentElement.removeAttribute('data-theme')
  document.documentElement.style.colorScheme = ''
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('ThemeProvider', () => {
  it('follows the system preference when nothing has been chosen', () => {
    stubMatchMedia(true)
    renderProbe()

    expect(probeText()).toBe('light (system)')
    expect(document.documentElement.dataset.theme).toBe('light')
  })

  it('prefers a stored choice over the system preference', () => {
    localStorage.setItem(THEME_STORAGE_KEY, 'dark')
    stubMatchMedia(true)
    renderProbe()

    expect(probeText()).toBe('dark (chosen)')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('persists a toggle so the theme survives a reload', async () => {
    stubMatchMedia(false)
    renderProbe()

    await userEvent.click(screen.getByRole('button'))

    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe('light')
    expect(document.documentElement.dataset.theme).toBe('light')
    // color-scheme drives native scrollbars and form controls.
    expect(document.documentElement.style.colorScheme).toBe('light')

    // Re-mounting stands in for a page reload.
    cleanup()
    renderProbe()

    expect(probeText()).toBe('light (chosen)')
  })

  it('ignores an unrecognised stored value instead of theming on it', () => {
    localStorage.setItem(THEME_STORAGE_KEY, 'solarized')
    stubMatchMedia(false)
    renderProbe()

    expect(probeText()).toBe('dark (system)')
  })

  it('tracks a live system change only while no choice has been made', async () => {
    const media = stubMatchMedia(false)
    renderProbe()

    media.emitSystemChange(true)
    expect(probeText()).toBe('light (system)')

    await userEvent.click(screen.getByRole('button'))
    expect(probeText()).toBe('dark (chosen)')

    // An explicit choice must win over later system changes.
    media.emitSystemChange(true)
    expect(probeText()).toBe('dark (chosen)')
  })
})

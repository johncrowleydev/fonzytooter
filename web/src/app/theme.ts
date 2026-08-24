export type Theme = 'dark' | 'light'

/**
 * Read by the inline bootstrap script in `index.html` before first paint, so this key is part of
 * the app's external contract. Changing it silently resets everyone's saved theme.
 */
export const THEME_STORAGE_KEY = 'fonzytooter.theme'

/** Kept in step with `--ft-canvas` so mobile browser chrome matches the page background. */
const themeColors: Record<Theme, string> = {
  dark: '#08111e',
  light: '#f3f6f9',
}

export function isTheme(value: unknown): value is Theme {
  return value === 'dark' || value === 'light'
}

export function readStoredTheme(): Theme | undefined {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return isTheme(stored) ? stored : undefined
  } catch {
    // Storage can throw when cookies/site data are blocked. Fall back to the system preference.
    return undefined
  }
}

export function storeTheme(theme: Theme) {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, theme)
  } catch {
    // A non-persisted theme is still usable for this session.
  }
}

export function readSystemTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

/**
 * Applies the theme to the document element rather than an app-level `div` so `color-scheme` in
 * `styles.css` can reach native UI: scrollbars, form controls, and overscroll background.
 */
export function applyTheme(theme: Theme) {
  const root = document.documentElement

  root.dataset.theme = theme
  // The bootstrap script sets this inline, which outranks the stylesheet, so it must be updated
  // here too or a toggle would leave native UI on the previous theme.
  root.style.colorScheme = theme
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', themeColors[theme])
}

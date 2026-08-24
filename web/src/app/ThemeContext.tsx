import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react'
import { applyTheme, readStoredTheme, readSystemTheme, storeTheme, type Theme } from './theme'

type ThemeContextValue = {
  theme: Theme
  /** True while the theme still follows the operating system because nothing was chosen yet. */
  followsSystem: boolean
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)

export function ThemeProvider({ children }: PropsWithChildren) {
  const [chosenTheme, setChosenTheme] = useState<Theme | undefined>(readStoredTheme)
  const [systemTheme, setSystemTheme] = useState<Theme>(readSystemTheme)
  const theme = chosenTheme ?? systemTheme

  // The bootstrap script owns the first paint; this keeps the document in step with later changes.
  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  // Track the OS preference so an unchosen theme follows a system light/dark switch live.
  useEffect(() => {
    const query = window.matchMedia('(prefers-color-scheme: light)')
    const handleChange = (event: MediaQueryListEvent) =>
      setSystemTheme(event.matches ? 'light' : 'dark')

    query.addEventListener('change', handleChange)
    return () => query.removeEventListener('change', handleChange)
  }, [])

  const toggleTheme = useCallback(() => {
    const next: Theme = theme === 'dark' ? 'light' : 'dark'
    setChosenTheme(next)
    storeTheme(next)
  }, [theme])

  const value = useMemo(
    () => ({ theme, followsSystem: chosenTheme === undefined, toggleTheme }),
    [theme, chosenTheme, toggleTheme],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const value = useContext(ThemeContext)

  if (!value) {
    throw new Error('useTheme must be used inside a ThemeProvider')
  }

  return value
}

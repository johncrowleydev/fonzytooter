export type NavItem = {
  label: string
  path: string
  icon: string
}

export const navItems: NavItem[] = [
  { label: 'Home', path: '/', icon: '⌂' },
  { label: 'Curriculum', path: '/curriculum', icon: '◫' },
  { label: 'Review', path: '/review', icon: '↺' },
  { label: 'Progress', path: '/progress', icon: '◒' },
  { label: 'Projects', path: '/projects', icon: '⌁' },
]

export function navigateTo(path: string) {
  window.history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function useCurrentPath() {
  const [path, setPath] = React.useState(window.location.pathname)

  React.useEffect(() => {
    const update = () => setPath(window.location.pathname)
    window.addEventListener('popstate', update)
    return () => window.removeEventListener('popstate', update)
  }, [])

  return path
}

import * as React from 'react'


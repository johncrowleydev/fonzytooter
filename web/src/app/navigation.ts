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

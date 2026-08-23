import type { PropsWithChildren } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import { navItems, type NavItem } from './navigation'
import { useTheme } from './ThemeContext'
import type { Theme } from './theme'
import { TutorButton } from '../features/tutor/TutorButton'
import { TutorOverlay } from '../features/tutor/TutorOverlay'

const navLinkStyles = {
  compact: {
    active: 'text-accent-teal',
    inactive: 'text-muted',
  },
  desktop: {
    active: 'border-l-2 border-accent-teal bg-accent-slate/10 text-ink',
    inactive: 'text-muted hover:bg-accent-slate/10 hover:text-ink',
  },
} as const

const navIconStyles = {
  compact: 'w-5 text-center text-lg leading-none',
  desktop: 'w-5 text-center text-lg leading-none text-faint',
} as const

export function AppShell({ children }: PropsWithChildren) {
  const { theme, toggleTheme } = useTheme()

  return (
    <div className="flex min-h-screen bg-canvas font-sans text-ink [background-image:radial-gradient(circle_at_80%_-10%,rgba(53,83,107,0.13),transparent_32rem)]">
      <aside className="sticky top-0 hidden h-screen w-60 shrink-0 flex-col overflow-hidden border-r border-line bg-shell px-4 py-6 lg:flex">
        <Link className="mb-14 flex items-center gap-3 px-2 text-left text-ink no-underline" to="/">
          <BrandMark />
          <strong className="text-base tracking-tight">Fonzytooter</strong>
        </Link>

        <div className="mb-2 px-3 text-xs font-bold uppercase tracking-widest text-faint">
          Workspace
        </div>
        <nav className="grid gap-1" aria-label="Primary navigation">
          {navItems.map((item) => (
            <ShellNavLink key={item.path} item={item} />
          ))}
        </nav>

        <div className="min-h-24 flex-1" />
        <ThemeToggle theme={theme} onClick={toggleTheme} />
        <div className="mt-4 flex items-center gap-2 border-t border-line px-2 pt-3">
          {/* A single initial inside a 28px circle, so it stays on the label tier. */}
          <span className="grid size-7 place-items-center rounded-full bg-avatar text-xs font-bold">
            F
          </span>
          <span>
            <strong className="block text-sm">Fonzy</strong>
            <small className="mt-0.5 block text-sm text-faint">Learning mode</small>
          </span>
          <span className="ml-auto text-faint">···</span>
        </div>
      </aside>

      <div className="min-w-0 flex-1">
        <header className="flex items-center justify-between border-b border-line px-4 py-4 lg:hidden">
          <Link className="flex items-center gap-2.5 text-ink no-underline" to="/">
            <BrandMark />
            <strong className="text-base tracking-tight">Fonzytooter</strong>
          </Link>
          <div className="flex items-center gap-2">
            <button
              className="grid size-8 place-items-center rounded-full border border-line-strong bg-accent-gold/10 text-accent-gold"
              onClick={toggleTheme}
              type="button"
              aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
            >
              {theme === 'dark' ? '☼' : '☾'}
            </button>
          </div>
        </header>

        <main className="mx-auto max-w-7xl px-4 pb-24 pt-8 sm:px-8 lg:px-14 lg:pt-12">
          {children}
        </main>

        <nav
          className="fixed inset-x-0 bottom-0 z-20 grid grid-cols-5 border-t border-line bg-shell-nav px-2 py-2 backdrop-blur lg:hidden"
          aria-label="Mobile navigation"
        >
          {navItems.map((item) => (
            <ShellNavLink key={item.path} item={item} compact />
          ))}
        </nav>
      </div>

      <TutorButton />
      <TutorOverlay />
    </div>
  )
}

function BrandMark() {
  return (
    <span className="grid size-7 place-items-center rounded-[9px_9px_9px_3px] bg-brand-teal font-serif text-xl font-bold text-brand-ink">
      ƒ
    </span>
  )
}

function ThemeToggle({ theme, onClick }: { theme: Theme; onClick: () => void }) {
  return (
    <button
      className="mt-2 flex w-full items-center gap-2 rounded-lg border border-line px-3 py-2.5 text-left text-sm text-muted transition hover:bg-accent-teal/10 hover:text-ink"
      type="button"
      onClick={onClick}
      aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
    >
      <span className="text-base leading-none text-accent-gold" aria-hidden="true">
        {theme === 'dark' ? '☼' : '☾'}
      </span>
      <span>{theme === 'dark' ? 'Light theme' : 'Dark theme'}</span>
    </button>
  )
}

function ShellNavLink({ item, compact = false }: { item: NavItem; compact?: boolean }) {
  const { pathname } = useLocation()

  return (
    <NavLink
      end={item.path === '/'}
      to={item.path}
      className={({ isActive }) => {
        const isSectionActive =
          isActive ||
          (item.activePathPrefix !== undefined &&
            (pathname === item.activePathPrefix ||
              pathname.startsWith(`${item.activePathPrefix}/`)))

        // The compact bar is a tab-bar label tier, not body copy: five columns on a narrow phone
        // cannot fit "Curriculum" at 14px.
        return `${compact ? 'flex min-w-0 flex-col items-center gap-1 rounded-lg px-1 py-1.5 text-xs' : 'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm'} no-underline transition ${navLinkStyles[compact ? 'compact' : 'desktop'][isSectionActive ? 'active' : 'inactive']}`
      }}
    >
      <span className={navIconStyles[compact ? 'compact' : 'desktop']}>{item.icon}</span>
      <span>{item.label}</span>
    </NavLink>
  )
}

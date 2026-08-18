import { useState, type PropsWithChildren } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { navItems, type NavItem } from './navigation'
import { TutorButton } from '../features/tutor/TutorButton'
import { TutorOverlay } from '../features/tutor/TutorOverlay'
import { useTutor } from '../features/tutor/TutorContext'

export function AppShell({ children }: PropsWithChildren) {
  const { openTutor } = useTutor()
  const [theme, setTheme] = useState<'dark' | 'light'>(() =>
    window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark',
  )

  const toggleTheme = () => setTheme((current) => (current === 'dark' ? 'light' : 'dark'))

  return (
    <div
      className={`theme-${theme} flex min-h-screen bg-[var(--canvas)] text-[var(--ink)] [background-image:radial-gradient(circle_at_80%_-10%,rgba(53,83,107,0.13),transparent_32rem)]`}
    >
      <aside
        className={`sticky top-0 hidden h-screen w-[242px] shrink-0 flex-col overflow-hidden border-r border-[var(--line)] px-4 py-6 lg:flex ${theme === 'light' ? 'bg-white/[0.86]' : 'bg-[rgba(9,18,31,0.78)]'}`}
      >
        <Link
          className="mb-14 flex items-center gap-3 px-2 text-left text-[var(--ink)] no-underline"
          to="/"
        >
          <BrandMark />
          <strong className="text-[15px] tracking-[-0.025em]">Fonzytooter</strong>
        </Link>

        <div className="mb-2 px-3 text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
          Workspace
        </div>
        <nav className="grid gap-1" aria-label="Primary navigation">
          {navItems.map((item) => (
            <ShellNavLink key={item.path} item={item} />
          ))}
        </nav>

        <div className="min-h-[100px] flex-1" />
        <button
          className="flex w-full items-center gap-2 rounded-lg border border-[var(--line)] bg-white/[0.025] px-3 py-2.5 text-left text-xs text-[var(--muted)] transition hover:text-[var(--ink)]"
          type="button"
          onClick={openTutor}
        >
          <span className="text-[var(--gold)]">✦</span>
          Open tutor
          <kbd className="ml-auto rounded border border-[var(--line)] px-1.5 py-0.5 text-[9px] text-[var(--faint)]">
            ⌘K
          </kbd>
        </button>
        <ThemeToggle theme={theme} onClick={toggleTheme} />
        <div className="mt-4 flex items-center gap-2 border-t border-[var(--line)] px-2 pt-3.5">
          <span className="grid size-7 place-items-center rounded-full bg-[#2f4c64] text-[11px] font-bold">
            F
          </span>
          <span>
            <strong className="block text-[11px]">Fonzy</strong>
            <small className="mt-0.5 block text-[10px] text-[var(--faint)]">Learning mode</small>
          </span>
          <span className="ml-auto text-[var(--faint)]">···</span>
        </div>
      </aside>

      <div className="min-w-0 flex-1">
        <header className="flex items-center justify-between border-b border-[var(--line)] px-4 py-4 lg:hidden">
          <Link className="flex items-center gap-2.5 text-[var(--ink)] no-underline" to="/">
            <BrandMark />
            <strong className="text-[15px] tracking-[-0.025em]">Fonzytooter</strong>
          </Link>
          <div className="flex items-center gap-2">
            <button
              className="grid size-8 place-items-center rounded-full border border-[var(--line-strong)] text-[var(--gold)]"
              onClick={toggleTheme}
              type="button"
              aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
            >
              {theme === 'dark' ? '☼' : '☾'}
            </button>
            <button
              className="grid size-8 place-items-center rounded-full border border-[var(--line-strong)] text-[var(--gold)]"
              onClick={openTutor}
              type="button"
              aria-label="Open tutor"
            >
              ✦
            </button>
          </div>
        </header>

        <main className="mx-auto max-w-[1380px] px-4 pb-24 pt-8 sm:px-8 lg:px-14 lg:pt-13">
          {children}
        </main>

        <nav
          className="fixed inset-x-0 bottom-0 z-20 grid grid-cols-5 border-t border-[var(--line)] bg-[rgba(9,18,31,0.96)] px-2 py-2 backdrop-blur lg:hidden"
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
    <span className="grid size-[29px] place-items-center rounded-[9px_9px_9px_3px] bg-[var(--teal)] font-serif text-[19px] font-bold text-[#0b151f]">
      ƒ
    </span>
  )
}

function ThemeToggle({ theme, onClick }: { theme: 'dark' | 'light'; onClick: () => void }) {
  return (
    <button
      className="mt-2 flex w-full items-center gap-2 rounded-lg border border-[var(--line)] px-3 py-2.5 text-left text-[11px] text-[var(--muted)] transition hover:bg-[rgba(118,208,192,0.08)] hover:text-[var(--ink)]"
      type="button"
      onClick={onClick}
    >
      <span className="text-[15px] leading-none text-[var(--gold)]">
        {theme === 'dark' ? '☼' : '☾'}
      </span>
      <span>{theme === 'dark' ? 'Light theme' : 'Dark theme'}</span>
      <kbd className="ml-auto rounded border border-[var(--line)] px-1.5 py-0.5 text-[9px] text-[var(--faint)]">
        {theme === 'dark' ? '☼' : '☾'}
      </kbd>
    </button>
  )
}

function ShellNavLink({ item, compact = false }: { item: NavItem; compact?: boolean }) {
  return (
    <NavLink
      end={item.path === '/'}
      to={item.path}
      className={({ isActive }) =>
        compact
          ? `flex min-w-0 flex-col items-center gap-1 rounded-lg px-1 py-1.5 text-[9px] no-underline transition ${isActive ? 'text-[var(--teal)]' : 'text-[var(--muted)]'}`
          : `flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-[13px] no-underline transition ${isActive ? 'bg-[rgba(142,183,195,0.1)] text-[var(--ink)] shadow-[inset_2px_0_0_var(--teal)]' : 'text-[var(--muted)] hover:bg-[rgba(142,183,195,0.08)] hover:text-[var(--ink)]'}`
      }
    >
      <span
        className={`w-[18px] text-center text-[18px] leading-none ${compact ? '' : 'text-[var(--faint)]'}`}
      >
        {item.icon}
      </span>
      <span>{item.label}</span>
    </NavLink>
  )
}

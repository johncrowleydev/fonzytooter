import { useState, type PropsWithChildren } from 'react'
import { navItems, navigateTo } from './navigation'
import { TutorButton } from '../features/tutor/TutorButton'
import { TutorOverlay } from '../features/tutor/TutorOverlay'
import { useTutor } from '../features/tutor/TutorContext'

export function AppShell({ children, currentPath }: PropsWithChildren<{ currentPath: string }>) {
  const { openTutor } = useTutor()
  const [theme, setTheme] = useState<'dark' | 'light'>(() => window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
  const activePath = currentPath === '/' ? '/' : `/${currentPath.split('/')[1]}`

  return (
    <div className={`app-shell theme-${theme}`}>
      <aside className="sidebar">
        <div className="brand-block">
          <button className="brand" onClick={() => navigateTo('/')} type="button">
            <span className="brand-mark">ƒ</span>
            <span><strong>Fonzytooter</strong></span>
          </button>
        </div>
        <div className="nav-label">Workspace</div>
        <nav className="side-nav" aria-label="Primary navigation">
          {navItems.map((item) => <NavLink key={item.path} item={item} active={activePath === item.path} />)}
        </nav>
        <div className="sidebar-spacer" />
        <button className="sidebar-tutor" type="button" onClick={openTutor}><span>✦</span> Open tutor <kbd>⌘K</kbd></button>
        <button className="theme-toggle" type="button" onClick={() => setTheme((current) => current === 'dark' ? 'light' : 'dark')}><span>{theme === 'dark' ? '☼' : '☾'}</span><span>{theme === 'dark' ? 'Light theme' : 'Dark theme'}</span><kbd>{theme === 'dark' ? '☼' : '☾'}</kbd></button>
        <div className="profile-chip"><span className="avatar">F</span><span><strong>Fonzy</strong><small>Learning mode</small></span><span className="profile-more">···</span></div>
      </aside>

      <div className="main-column">
        <header className="mobile-header">
          <button className="brand compact" onClick={() => navigateTo('/')} type="button"><span className="brand-mark">ƒ</span><strong>Fonzytooter</strong></button>
          <div className="mobile-actions"><button className="mobile-theme" onClick={() => setTheme((current) => current === 'dark' ? 'light' : 'dark')} type="button" aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}>{theme === 'dark' ? '☼' : '☾'}</button><button className="mobile-tutor" onClick={openTutor} type="button" aria-label="Open tutor">✦</button></div>
        </header>
        <main className="main-content">{children}</main>
        <nav className="mobile-nav" aria-label="Mobile navigation">
          {navItems.map((item) => <NavLink key={item.path} item={item} active={activePath === item.path} compact />)}
        </nav>
      </div>

      <TutorButton />
      <TutorOverlay />
    </div>
  )
}

function NavLink({ item, active, compact = false }: { item: (typeof navItems)[number]; active: boolean; compact?: boolean }) {
  return <button type="button" onClick={() => navigateTo(item.path)} className={`${compact ? 'mobile-nav-item' : 'side-nav-item'} ${active ? 'active' : ''}`}><span className="nav-icon">{item.icon}</span><span>{item.label}</span></button>
}

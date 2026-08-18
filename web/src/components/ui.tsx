import type { PropsWithChildren, ReactNode } from 'react'

export function Button({
  children,
  onClick,
  variant = 'primary',
  className = '',
  type = 'button',
  disabled = false,
}: PropsWithChildren<{
  onClick?: () => void
  variant?: 'primary' | 'secondary' | 'quiet' | 'outline'
  className?: string
  type?: 'button' | 'submit'
  disabled?: boolean
}>) {
  return (
    <button type={type} onClick={onClick} disabled={disabled} className={`button button-${variant} ${className}`}>
      {children}
    </button>
  )
}

export function Card({ children, className = '', muted = false }: PropsWithChildren<{ className?: string; muted?: boolean }>) {
  return <section className={`surface ${muted ? 'surface-muted' : ''} ${className}`}>{children}</section>
}

export function Badge({ children, tone = 'neutral' }: PropsWithChildren<{ tone?: 'neutral' | 'teal' | 'gold' | 'coral' | 'violet' | 'blue' }>) {
  return <span className={`badge badge-${tone}`}>{children}</span>
}

export function SectionHeading({ eyebrow, title, detail, action }: { eyebrow?: string; title: string; detail?: string; action?: ReactNode }) {
  return (
    <div className="section-heading">
      <div>
        {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
        <h2>{title}</h2>
        {detail ? <p className="section-detail">{detail}</p> : null}
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  )
}

export function ProgressBar({ value, tone = 'teal' }: { value: number; tone?: 'teal' | 'gold' | 'coral' | 'violet' | 'blue' }) {
  return (
    <div className="progress-track" aria-label={`${value}% complete`}>
      <span className={`progress-fill progress-${tone}`} style={{ width: `${Math.max(0, Math.min(value, 100))}%` }} />
    </div>
  )
}

export function StatusDot({ state, size = 'normal' }: { state: 'locked' | 'available' | 'in-progress' | 'completed' | 'done' | 'working' | 'todo'; size?: 'small' | 'normal' }) {
  return <span className={`status-dot status-${state} status-${size}`} aria-label={state} />
}

export function EmptyState({ title, children }: PropsWithChildren<{ title: string }>) {
  return <div className="empty-state"><span className="empty-mark">·</span><h3>{title}</h3><p>{children}</p></div>
}

export function PageIntro({ eyebrow, title, detail, compact = false, children }: PropsWithChildren<{ eyebrow?: string; title: string; detail?: string; compact?: boolean }>) {
  return (
    <header className={`page-intro ${compact ? 'page-intro-compact' : ''}`}>
      {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
      <h1>{title}</h1>
      {detail ? <p>{detail}</p> : null}
      {children}
    </header>
  )
}

export function IconLabel({ icon, children }: PropsWithChildren<{ icon: string }>) {
  return <span className="icon-label"><span aria-hidden="true">{icon}</span>{children}</span>
}

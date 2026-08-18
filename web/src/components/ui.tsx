import type { PropsWithChildren, ReactNode } from 'react'

const badgeTones = {
  neutral: 'border-line bg-white/5 text-muted',
  teal: 'border-brand-teal/30 bg-brand-teal/10 text-brand-teal',
  gold: 'border-brand-gold/30 bg-brand-gold/10 text-brand-gold',
  coral: 'border-brand-coral/30 bg-brand-coral/10 text-brand-coral',
  violet: 'border-brand-violet/30 bg-brand-violet/10 text-brand-violet',
  blue: 'border-brand-blue/30 bg-brand-blue/10 text-brand-blue',
} as const

const progressTones = {
  teal: 'bg-brand-teal',
  gold: 'bg-brand-gold',
  coral: 'bg-brand-coral',
  violet: 'bg-brand-violet',
  blue: 'bg-brand-blue',
} as const

const statusStyles = {
  locked: 'border-slate-600',
  todo: 'border-slate-600',
  available: 'border-brand-teal',
  'in-progress': 'border-brand-gold bg-brand-gold ring-2 ring-inset ring-panel',
  working: 'border-brand-gold bg-brand-gold ring-2 ring-inset ring-panel',
  completed: 'border-brand-teal bg-brand-teal',
  done: 'border-brand-teal bg-brand-teal',
} as const

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
  const variants = {
    primary:
      'border-0 bg-brand-teal text-brand-ink hover:-translate-y-px hover:bg-brand-teal-light',
    secondary: 'border border-line-strong bg-brand-slate/10 text-ink hover:bg-brand-slate/20',
    quiet: 'border-0 bg-transparent text-muted',
    outline: 'border border-line bg-transparent text-ink hover:bg-brand-slate/20',
  }

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`inline-flex items-center justify-center gap-2.5 rounded-lg px-4 py-2.5 text-xs font-bold transition ${variants[variant]} ${className}`}
    >
      {children}
    </button>
  )
}

export function Card({
  children,
  className = '',
  muted = false,
}: PropsWithChildren<{ className?: string; muted?: boolean }>) {
  return (
    <section
      className={`rounded-xl border border-line px-6 py-5 shadow-lg ${muted ? 'bg-panel-muted' : 'bg-panel'} ${className}`}
    >
      {children}
    </section>
  )
}

export function Badge({
  children,
  tone = 'neutral',
}: PropsWithChildren<{ tone?: keyof typeof badgeTones }>) {
  return (
    <span
      className={`inline-flex w-max items-center rounded-full border px-2 py-1 text-2xs font-bold uppercase leading-none tracking-wide ${badgeTones[tone]}`}
    >
      {children}
    </span>
  )
}

export function SectionHeading({
  eyebrow,
  title,
  detail,
  action,
}: {
  eyebrow?: string
  title: string
  detail?: string
  action?: ReactNode
}) {
  return (
    <div className="mb-4 flex items-start justify-between gap-4">
      <div>
        {eyebrow ? (
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
        ) : null}
        <h2 className="mt-1.5 text-lg font-semibold leading-tight tracking-tight">{title}</h2>
        {detail ? <p className="mt-1.5 text-xs leading-normal text-faint">{detail}</p> : null}
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  )
}

export function ProgressBar({
  value,
  tone = 'teal',
}: {
  value: number
  tone?: keyof typeof progressTones
}) {
  return (
    <div
      className="h-1.5 w-full overflow-hidden rounded-full bg-brand-slate/15"
      aria-label={`${value}% complete`}
    >
      <span
        className={`block h-full rounded-full ${progressTones[tone]}`}
        style={{ width: `${Math.max(0, Math.min(value, 100))}%` }}
      />
    </div>
  )
}

export function StatusDot({
  state,
  size = 'normal',
}: {
  state: keyof typeof statusStyles
  size?: 'small' | 'normal'
}) {
  const sizeClass = size === 'small' ? 'size-2' : 'size-3'
  return (
    <span
      className={`inline-block shrink-0 rounded-full border ${sizeClass} ${statusStyles[state]}`}
      aria-label={state}
    />
  )
}

export function EmptyState({ title, children }: PropsWithChildren<{ title: string }>) {
  return (
    <div className="text-center">
      <span className="block text-xl text-faint">·</span>
      <h3 className="text-lg font-semibold">{title}</h3>
      <p className="text-muted">{children}</p>
    </div>
  )
}

export function PageIntro({
  eyebrow,
  title,
  detail,
  compact = false,
  children,
}: PropsWithChildren<{ eyebrow?: string; title: string; detail?: string; compact?: boolean }>) {
  return (
    <header className={`max-w-3xl ${compact ? 'max-w-none' : ''}`}>
      {eyebrow ? (
        <p className="mb-3 text-2xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
      ) : null}
      <h1
        className={`${compact ? 'text-3xl sm:text-4xl' : 'text-4xl sm:text-5xl'} m-0 font-semibold leading-none tracking-tight`}
      >
        {title}
      </h1>
      {detail ? (
        <p className="mt-4 max-w-2xl text-sm leading-relaxed text-muted">{detail}</p>
      ) : null}
      {children}
    </header>
  )
}

export function IconLabel({ icon, children }: PropsWithChildren<{ icon: string }>) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden="true">{icon}</span>
      {children}
    </span>
  )
}

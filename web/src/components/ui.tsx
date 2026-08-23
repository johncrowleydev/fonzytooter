import type { PropsWithChildren, ReactNode } from 'react'

const badgeTones = {
  neutral: 'border-line bg-raised text-muted',
  teal: 'border-accent-teal/30 bg-accent-teal/10 text-accent-teal',
  gold: 'border-accent-gold/30 bg-accent-gold/10 text-accent-gold',
  coral: 'border-accent-coral/30 bg-accent-coral/10 text-accent-coral',
  violet: 'border-accent-violet/30 bg-accent-violet/10 text-accent-violet',
  blue: 'border-accent-blue/30 bg-accent-blue/10 text-accent-blue',
} as const

/*
 * Accent rather than brand: a progress fill carries no text, so it has to contrast with its track
 * instead. The vivid fill against the light-mode track measures about 1.4:1, which reads as an
 * empty bar.
 */
const progressTones = {
  teal: 'bg-accent-teal',
  gold: 'bg-accent-gold',
  coral: 'bg-accent-coral',
  violet: 'bg-accent-violet',
  blue: 'bg-accent-blue',
} as const

/*
 * A status dot carries meaning, so its rim needs to stay perceivable. The empty states use
 * `border-faint` rather than a line token: `line-strong` is ~1.7:1 on a white panel, which reads
 * as no dot at all. The filled states pair an accent rim with a vivid fill for the same reason —
 * the fill alone measures under 2:1 in light mode.
 */
const statusStyles = {
  locked: 'border-faint',
  todo: 'border-faint',
  available: 'border-accent-teal',
  'in-progress': 'border-accent-gold bg-brand-gold ring-2 ring-inset ring-panel',
  working: 'border-accent-gold bg-brand-gold ring-2 ring-inset ring-panel',
  completed: 'border-accent-teal bg-brand-teal',
  done: 'border-accent-teal bg-brand-teal',
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
      'border-0 bg-brand-teal text-brand-ink hover:-translate-y-px hover:bg-brand-teal-light active:translate-y-0 active:bg-brand-teal-light',
    secondary:
      'border border-line-strong bg-accent-slate/10 text-ink hover:bg-accent-slate/20 active:bg-accent-slate/30',
    quiet: 'border-0 bg-transparent text-muted active:bg-accent-slate/20',
    outline:
      'border border-line bg-transparent text-ink hover:bg-accent-slate/20 active:bg-accent-slate/30',
  }

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      /*
       * `pointer-coarse:min-h-11` reaches the 44px touch target without loosening the layout for a
       * mouse, where 40px is comfortable. `active:` matters more on touch than hover does: a
       * finger never hovers, so the press state is the only feedback a tap gets.
       */
      className={`inline-flex items-center justify-center gap-2.5 rounded-lg px-4 py-2.5 text-sm font-bold transition pointer-coarse:min-h-11 ${variants[variant]} ${className}`}
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
      className={`inline-flex w-max items-center rounded-full border px-2 py-1 text-xs font-bold uppercase leading-none tracking-wide ${badgeTones[tone]}`}
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
          <p className="text-xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
        ) : null}
        <h2 className="mt-1.5 text-lg font-semibold leading-tight tracking-tight">{title}</h2>
        {detail ? <p className="mt-1.5 text-sm leading-normal text-faint">{detail}</p> : null}
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
      className="h-1.5 w-full overflow-hidden rounded-full bg-accent-slate/15"
      aria-label={`${value}% complete`}
    >
      <span
        className={`block h-full rounded-full ${progressTones[tone]}`}
        style={{ width: `${Math.max(0, Math.min(value, 100))}%` }}
      />
    </div>
  )
}

/*
 * A screen reader reads this label out, so it cannot be the state key: "in-progress" and
 * "not-assessed" are internal vocabulary that happens to look like words.
 */
const statusLabels: Record<keyof typeof statusStyles, string> = {
  locked: 'Locked',
  todo: 'Not started',
  available: 'Available',
  'in-progress': 'In progress',
  working: 'In progress',
  completed: 'Completed',
  done: 'Completed',
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
      role="img"
      aria-label={statusLabels[state]}
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
        <p className="mb-3 text-xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
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

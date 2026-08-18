import type { PropsWithChildren, ReactNode } from 'react'

const badgeTones = {
  neutral: 'border-[var(--line)] bg-white/[0.03] text-[var(--muted)]',
  teal: 'border-[rgba(118,208,192,0.27)] bg-[rgba(118,208,192,0.09)] text-[var(--teal)]',
  gold: 'border-[rgba(225,184,106,0.27)] bg-[rgba(225,184,106,0.09)] text-[var(--gold)]',
  coral: 'border-[rgba(239,145,110,0.28)] bg-[rgba(239,145,110,0.09)] text-[var(--coral)]',
  violet: 'border-[rgba(169,155,231,0.27)] bg-[rgba(169,155,231,0.09)] text-[var(--violet)]',
  blue: 'border-[rgba(128,175,228,0.27)] bg-[rgba(128,175,228,0.09)] text-[var(--blue)]',
} as const

const progressTones = {
  teal: 'bg-[var(--teal)]',
  gold: 'bg-[var(--gold)]',
  coral: 'bg-[var(--coral)]',
  violet: 'bg-[var(--violet)]',
  blue: 'bg-[var(--blue)]',
} as const

const statusStyles = {
  locked: 'border-[#48576a]',
  todo: 'border-[#48576a]',
  available: 'border-[var(--teal)]',
  'in-progress': 'border-[var(--gold)] bg-[var(--gold)] shadow-[inset_0_0_0_2px_var(--panel)]',
  working: 'border-[var(--gold)] bg-[var(--gold)] shadow-[inset_0_0_0_2px_var(--panel)]',
  completed: 'border-[var(--teal)] bg-[var(--teal)]',
  done: 'border-[var(--teal)] bg-[var(--teal)]',
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
    primary: 'border-0 bg-[var(--teal)] text-[#0b171e] hover:-translate-y-px hover:bg-[#9ce1d5]',
    secondary:
      'border border-[var(--line-strong)] bg-[rgba(157,185,194,0.1)] text-[var(--ink)] hover:bg-[rgba(157,185,194,0.17)]',
    quiet: 'border-0 bg-transparent text-[var(--muted)]',
    outline:
      'border border-[var(--line)] bg-transparent text-[var(--ink)] hover:bg-[rgba(157,185,194,0.17)]',
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
      className={`rounded-[14px] border border-[var(--line)] px-6 py-5 shadow-[0_16px_40px_rgba(0,0,0,0.06)] ${muted ? 'bg-[rgba(18,31,47,0.52)]' : 'bg-[var(--panel)]'} ${className}`}
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
      className={`inline-flex w-max items-center rounded-full border px-2 py-1 text-[9px] font-bold uppercase leading-none tracking-[0.07em] ${badgeTones[tone]}`}
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
    <div className="mb-[17px] flex items-start justify-between gap-[18px]">
      <div>
        {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
        <h2 className="mt-1.5 text-[18px] font-semibold leading-tight tracking-[-0.025em]">
          {title}
        </h2>
        {detail ? (
          <p className="mt-1.5 text-[11px] leading-normal text-[var(--faint)]">{detail}</p>
        ) : null}
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
      className="h-[5px] w-full overflow-hidden rounded-full bg-[rgba(157,185,194,0.12)]"
      aria-label={`${value}% complete`}
    >
      <span
        className={`block h-full rounded-[inherit] ${progressTones[tone]}`}
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
  const sizeClass = size === 'small' ? 'size-[7px]' : 'size-[11px]'
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
      <span className="block text-xl text-[var(--faint)]">·</span>
      <h3>{title}</h3>
      <p className="text-[var(--muted)]">{children}</p>
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
    <header className={`max-w-[720px] ${compact ? 'max-w-none' : ''}`}>
      {eyebrow ? <p className="eyebrow mb-3">{eyebrow}</p> : null}
      <h1
        className={`${compact ? 'text-[clamp(25px,3vw,34px)]' : 'text-[clamp(32px,4vw,50px)]'} m-0 font-semibold leading-none tracking-[-0.055em]`}
      >
        {title}
      </h1>
      {detail ? (
        <p className="mt-4 max-w-[620px] text-sm leading-relaxed text-[var(--muted)]">{detail}</p>
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

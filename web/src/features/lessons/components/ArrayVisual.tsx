import type { ReactNode } from 'react'

export type Shape = readonly number[]

export function formatShape(shape: Shape) {
  if (shape.length === 0) return 'scalar'
  return `(${shape.join(', ')}${shape.length === 1 ? ',' : ''})`
}

export function ShapeChips({ shape, activeAxis }: { shape: Shape; activeAxis?: number }) {
  return (
    <div className="flex flex-wrap gap-2" aria-label={`shape ${formatShape(shape)}`}>
      {shape.map((length, axis) => (
        <span
          key={axis}
          className={
            axis === activeAxis
              ? 'rounded-md border border-brand-teal bg-brand-teal/10 px-3 py-2 font-mono text-xs text-ink'
              : 'rounded-md border border-line bg-panel-soft px-3 py-2 font-mono text-xs text-muted'
          }
        >
          axis {axis}: {length}
        </span>
      ))}
    </div>
  )
}

export function ValueGrid({
  values,
  columns,
  selected,
  label,
}: {
  values: readonly number[]
  columns: number
  selected?: readonly number[]
  label: string
}) {
  const selectedSet = new Set(selected)

  return (
    <div className="overflow-x-auto rounded-lg border border-line bg-panel-soft p-3">
      <div
        className="grid min-w-max gap-1.5"
        style={{ gridTemplateColumns: `repeat(${columns}, minmax(2.5rem, 1fr))` }}
        role="img"
        aria-label={label}
      >
        {values.map((value, index) => (
          <span
            key={index}
            className={
              selectedSet.has(index)
                ? 'grid min-h-10 place-items-center rounded-md border border-brand-teal bg-brand-teal/15 px-2 font-mono text-xs font-semibold text-ink'
                : 'grid min-h-10 place-items-center rounded-md border border-line bg-panel px-2 font-mono text-xs text-muted'
            }
            aria-label={`${value}${selectedSet.has(index) ? ', selected' : ''}`}
          >
            {value}
          </span>
        ))}
      </div>
    </div>
  )
}

export function InteractiveHeader({
  title,
  description,
  badge,
}: {
  title: string
  description: string
  badge?: ReactNode
}) {
  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p className="text-2xs font-bold uppercase tracking-widest text-faint">Interactive model</p>
        <h2 className="mt-1.5 text-lg font-semibold tracking-tight">{title}</h2>
        <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">{description}</p>
      </div>
      {badge}
    </div>
  )
}

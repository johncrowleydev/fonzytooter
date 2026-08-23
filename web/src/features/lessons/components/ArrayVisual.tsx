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
              ? 'rounded-md border border-accent-teal bg-accent-teal/10 px-3 py-2 font-mono text-sm text-ink'
              : 'rounded-md border border-line bg-panel-soft px-3 py-2 font-mono text-sm text-muted'
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
  const selectedValues = values.filter((_, index) => selectedSet.has(index))

  /*
   * `role="img"` makes this element a leaf in the accessibility tree, so its children are pruned.
   * Each cell used to carry its own aria-label announcing ", selected", none of which was
   * reachable -- and which cells are selected is the entire point of the indexing lesson.
   *
   * Rather than exposing twelve cells one at a time, the selection is summarised into the image's
   * own label, and the cells are explicitly hidden so the pruning is intentional rather than
   * incidental.
   */
  const accessibleLabel =
    selectedValues.length > 0 ? `${label}. Selected values: ${selectedValues.join(', ')}.` : label

  return (
    <div className="overflow-x-auto overscroll-x-contain rounded-lg border border-line bg-panel-soft p-3">
      <div
        className="grid min-w-max gap-1.5"
        style={{ gridTemplateColumns: `repeat(${columns}, minmax(2.5rem, 1fr))` }}
        role="img"
        aria-label={accessibleLabel}
      >
        {values.map((value, index) => (
          <span
            key={index}
            className={
              selectedSet.has(index)
                ? 'grid min-h-10 place-items-center rounded-md border border-accent-teal bg-accent-teal/15 px-2 font-mono text-sm font-semibold text-ink'
                : 'grid min-h-10 place-items-center rounded-md border border-line bg-panel px-2 font-mono text-sm text-muted'
            }
            aria-hidden="true"
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
        <p className="text-xs font-bold uppercase tracking-widest text-faint">Interactive model</p>
        <h2 className="mt-1.5 text-lg font-semibold tracking-tight">{title}</h2>
        <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted">{description}</p>
      </div>
      {badge}
    </div>
  )
}

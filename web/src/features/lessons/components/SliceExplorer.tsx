import { useState } from 'react'
import { Badge, Card } from '../../../components/ui'

const values = [2, 4, 6, 8, 10]
const sliceIndexOptions = [-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5]
const includedCellClass =
  'rounded-lg border border-accent-teal/50 bg-accent-teal/10 px-2 py-3 text-ink'
const excludedCellClass = 'rounded-lg border border-line bg-raised px-2 py-3 text-muted'

function normalizeSliceBound(bound: number, length: number) {
  return bound < 0 ? Math.max(bound + length, 0) : Math.min(bound, length)
}

function getIncludedIndices(start: number, stop: number, length: number) {
  const normalizedStart = normalizeSliceBound(start, length)
  const normalizedStop = normalizeSliceBound(stop, length)

  if (normalizedStart >= normalizedStop) return []

  return Array.from(
    { length: normalizedStop - normalizedStart },
    (_, offset) => normalizedStart + offset,
  )
}

export function SliceExplorer() {
  const [start, setStart] = useState(1)
  const [stop, setStop] = useState(-1)
  const includedIndices = getIncludedIndices(start, stop, values.length)
  const includedValues = includedIndices.map((index) => values[index])

  return (
    <Card className="grid gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">Read a Python slice</h2>
          <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">
            The start index is included. The stop index is excluded. Negative indices count from the
            end.
          </p>
        </div>
        <Badge tone="violet">No step</Badge>
      </div>

      <fieldset className="grid gap-4 rounded-lg border border-line bg-raised p-4 sm:grid-cols-2">
        <legend className="px-1 text-2xs font-bold uppercase tracking-widest text-faint">
          Slice controls
        </legend>
        <label className="grid gap-1.5 text-xs text-muted" htmlFor="slice-start">
          <span className="font-semibold text-ink">start</span>
          <select
            className="rounded-md border border-line-strong bg-panel px-3 py-2 text-sm text-ink outline-0 focus-visible:border-accent-teal focus-visible:ring-2 focus-visible:ring-accent-teal/30"
            id="slice-start"
            value={start}
            onChange={(event) => setStart(Number(event.target.value))}
          >
            {sliceIndexOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="grid gap-1.5 text-xs text-muted" htmlFor="slice-stop">
          <span className="font-semibold text-ink">stop</span>
          <select
            className="rounded-md border border-line-strong bg-panel px-3 py-2 text-sm text-ink outline-0 focus-visible:border-accent-teal focus-visible:ring-2 focus-visible:ring-accent-teal/30"
            id="slice-stop"
            value={stop}
            onChange={(event) => setStop(Number(event.target.value))}
          >
            {sliceIndexOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
      </fieldset>

      <div className="overflow-x-auto rounded-lg border border-line bg-panel-soft p-4">
        <div className="grid min-w-96 grid-cols-5 gap-2 text-center">
          {values.map((value, index) => {
            const included = includedIndices.includes(index)

            return (
              <div
                key={value}
                className={included ? includedCellClass : excludedCellClass}
                aria-label={`value ${value}, positive index ${index}, negative index ${index - values.length}, ${included ? 'included' : 'not included'}`}
              >
                <span className="block text-2xs uppercase tracking-wide text-faint">
                  index {index}
                </span>
                <strong className="mt-2 block text-lg">{value}</strong>
                <span className="mt-2 block text-2xs uppercase tracking-wide text-faint">
                  {included ? 'included' : 'outside slice'}
                </span>
                <span className="mt-1 block text-2xs text-muted">
                  negative {index - values.length}
                </span>
              </div>
            )
          })}
        </div>
        <div className="mt-3 flex items-center justify-between gap-3 text-2xs text-muted">
          <span>positive indices: 0 → 4</span>
          <span>negative indices: -5 → -1</span>
        </div>
      </div>

      <div
        className="rounded-lg border border-accent-violet/30 bg-accent-violet/10 p-4"
        role="status"
        aria-live="polite"
      >
        <p className="text-2xs font-bold uppercase tracking-widest text-faint">Result</p>
        <pre className="mt-3 overflow-x-auto font-mono text-sm leading-relaxed text-ink">
          <code>{`values[${start}:${stop}]\n→ [${includedValues.join(', ')}]`}</code>
        </pre>
        <p className="mt-3 text-xs leading-relaxed text-muted">
          {includedValues.length > 0
            ? `Indices ${includedIndices.join(', ')} are included; the stop bound is not.`
            : 'This pair of bounds contains no values.'}
        </p>
      </div>
    </Card>
  )
}

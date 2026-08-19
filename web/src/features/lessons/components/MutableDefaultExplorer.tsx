import { useState } from 'react'
import { Badge, Card } from '../../../components/ui'

const callValues = [1, 2, 3] as const
const snapshots = [[], [1], [1, 2], [1, 2, 3]] as const
const callStatusClasses = {
  completed: 'border-brand-teal/40 bg-brand-teal/10 text-ink',
  next: 'border-brand-gold/50 bg-brand-gold/10 text-ink',
  waiting: 'border-line bg-white/5 text-muted',
} as const
const buttonClass =
  'inline-flex items-center justify-center gap-2 rounded-lg border px-3 py-2 text-xs font-semibold transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-teal disabled:cursor-not-allowed disabled:opacity-50'

function formatValues(values: readonly number[]) {
  return `[${values.join(', ')}]`
}

export function MutableDefaultExplorer() {
  const [callCount, setCallCount] = useState(0)

  function runCall(callValue: number) {
    if (callCount === callValue - 1) setCallCount(callValue)
  }

  return (
    <Card className="grid gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">A default list is reused</h2>
          <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">
            Python evaluates a default argument when the function is defined, so this list is not
            recreated for every call.
          </p>
        </div>
        <Badge tone="coral">Simulated calls</Badge>
      </div>

      <div className="rounded-lg border border-brand-coral/30 bg-brand-coral/10 p-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-2xs font-bold uppercase tracking-widest text-brand-coral">
            Function-definition time
          </p>
          <span className="text-2xs font-semibold text-ink">created once</span>
        </div>
        <pre className="mt-3 overflow-x-auto font-mono text-xs leading-relaxed text-ink">
          <code>{`def record(value, values=[]):\n    values.append(value)\n    return values`}</code>
        </pre>
        <div className="mt-4 flex flex-wrap items-center gap-3 rounded-md border border-brand-coral/30 bg-panel px-3 py-2">
          <span className="text-xs text-muted">default list object</span>
          <code className="text-sm text-ink">{formatValues(snapshots[callCount])}</code>
          <span className="text-2xs text-muted">same object across omitted-argument calls</span>
        </div>
      </div>

      <div className="grid gap-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">
              Individual calls
            </p>
            <p className="mt-1 text-xs text-muted">
              Run them in order to reveal each returned list.
            </p>
          </div>
          <span className="text-2xs text-muted" role="status" aria-live="polite">
            {callCount} / 3 calls shown
          </span>
        </div>

        <div className="grid gap-3">
          {callValues.map((callValue) => {
            const completed = callCount >= callValue
            const next = callCount === callValue - 1
            const status = completed ? 'completed' : next ? 'next' : 'waiting'

            return (
              <div key={callValue} className={`rounded-lg border p-3 ${callStatusClasses[status]}`}>
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <code className="text-sm">record({callValue})</code>
                    <p className="mt-1 text-2xs text-muted">
                      {completed
                        ? 'call returned the shared default list'
                        : next
                          ? 'next call in the sequence'
                          : 'waiting for the previous call'}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <code className="text-sm">
                      {completed ? formatValues(snapshots[callValue]) : 'not shown yet'}
                    </code>
                    <button
                      className={`${buttonClass} border-brand-gold/40 bg-brand-gold/10 text-brand-gold hover:bg-brand-gold/20`}
                      type="button"
                      onClick={() => runCall(callValue)}
                      disabled={!next}
                      aria-label={`Run record(${callValue})`}
                    >
                      {completed ? 'Shown' : `Run record(${callValue})`}
                    </button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      <div className="rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-4">
        <p className="text-2xs font-bold uppercase tracking-widest text-faint">Safe contrast</p>
        <pre className="mt-3 overflow-x-auto font-mono text-xs leading-relaxed text-ink">
          <code>{`def record(value, values=None):\n    if values is None:\n        values = []\n    values.append(value)\n    return values`}</code>
        </pre>
        <p className="mt-3 text-xs leading-relaxed text-muted">
          The <code className="text-ink">None</code> pattern creates a fresh list inside each call
          that omits <code className="text-ink">values</code>.
        </p>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
        <p className="text-xs text-muted">
          This visual simulates the sequence; it does not execute Python.
        </p>
        <button
          className={`${buttonClass} border-line-strong bg-brand-slate/10 text-ink hover:bg-brand-slate/20`}
          type="button"
          onClick={() => setCallCount(0)}
        >
          Reset
        </button>
      </div>
    </Card>
  )
}

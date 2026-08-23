import { useState } from 'react'
import { Badge, Card } from '../../../components/ui'

type BindingName = 'samples' | 'backup'

type BindingState = {
  samples: number[]
  backup: number[]
  isCopied: boolean
}

const controlClass =
  'inline-flex items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-50'

function createInitialState(): BindingState {
  const samples = [10, 20, 30]

  return {
    samples,
    backup: samples,
    isCopied: false,
  }
}

function ListObjectCard({
  label,
  names,
  values,
  shared,
}: {
  label: string
  names: string
  values: number[]
  shared: boolean
}) {
  return (
    <div className="rounded-lg border border-accent-blue/30 bg-accent-blue/10 p-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-bold uppercase tracking-widest text-accent-blue">
          {label}
        </span>
        <span className="text-sm text-muted">
          {shared ? 'shared storage' : 'independent storage'}
        </span>
      </div>
      <code className="mt-3 block text-sm text-ink">[{values.join(', ')}]</code>
      <p className="mt-2 text-sm leading-normal text-muted">
        {shared
          ? `${names} both refer to this list object.`
          : `${names} refers to this list object.`}
      </p>
    </div>
  )
}

export function ReferenceBindingExplorer() {
  const [bindingState, setBindingState] = useState<BindingState>(createInitialState)

  function appendThrough(name: BindingName) {
    setBindingState((current) => {
      const updatedValues = [...current[name], 40]

      if (!current.isCopied) {
        return {
          samples: updatedValues,
          backup: updatedValues,
          isCopied: false,
        }
      }

      if (name === 'samples') {
        return { ...current, samples: updatedValues }
      }

      return { ...current, backup: updatedValues }
    })
  }

  function copyBackup() {
    setBindingState((current) => ({
      ...current,
      backup: [...current.backup],
      isCopied: true,
    }))
  }

  return (
    <Card className="grid gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">Reference or copy?</h2>
          <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted">
            Assignment binds another name to an object. It does not automatically create new list
            storage.
          </p>
        </div>
        <Badge tone={bindingState.isCopied ? 'gold' : 'teal'}>
          {bindingState.isCopied ? 'Two list objects' : 'One list object'}
        </Badge>
      </div>

      <div
        className={`rounded-lg border p-3 text-sm leading-relaxed ${bindingState.isCopied ? 'border-accent-gold/30 bg-accent-gold/10 text-ink' : 'border-accent-teal/30 bg-accent-teal/10 text-ink'}`}
        role="status"
        aria-live="polite"
      >
        <strong>
          {bindingState.isCopied
            ? 'backup was copied, so the names now reach independent storage.'
            : 'samples and backup reach the same mutable list.'}
        </strong>
        <p className="mt-1 text-muted">
          {bindingState.isCopied
            ? 'Appending through one name changes only that name’s list.'
            : 'Appending through either name changes what both names show.'}
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-3 md:items-stretch">
        <div className="rounded-lg border border-line bg-raised p-4">
          <p className="text-xs font-bold uppercase tracking-widest text-faint">Names</p>
          <div className="mt-3 grid gap-3">
            <div className="flex items-center justify-between gap-3 rounded-md border border-line px-3 py-2">
              <code className="text-sm text-ink">samples</code>
              <span className="text-sm text-accent-teal" aria-hidden="true">
                → object A
              </span>
            </div>
            <div className="flex items-center justify-between gap-3 rounded-md border border-line px-3 py-2">
              <code className="text-sm text-ink">backup</code>
              <span className="text-sm text-accent-teal" aria-hidden="true">
                → {bindingState.isCopied ? 'object B' : 'object A'}
              </span>
            </div>
          </div>
          <p className="mt-3 text-sm leading-normal text-muted">
            The arrows show where each name looks for its list object.
          </p>
        </div>

        <div className="grid place-items-center text-2xl text-accent-teal" aria-hidden="true">
          →
        </div>

        <div className="rounded-lg border border-line bg-raised p-4">
          <p className="text-xs font-bold uppercase tracking-widest text-faint">List storage</p>
          <div className="mt-3 grid gap-3 sm:grid-cols-2 md:grid-cols-1">
            <ListObjectCard
              label="object A"
              names={bindingState.isCopied ? 'samples' : 'samples and backup'}
              values={bindingState.samples}
              shared={!bindingState.isCopied}
            />
            {bindingState.isCopied ? (
              <ListObjectCard
                label="object B"
                names="backup"
                values={bindingState.backup}
                shared={false}
              />
            ) : null}
          </div>
        </div>
      </div>

      <pre className="overflow-x-auto overscroll-x-contain rounded-lg border border-line bg-code-surface p-4 font-mono text-sm leading-relaxed text-code-ink">
        <code>
          {bindingState.isCopied
            ? `samples = [10, 20, 30]\nbackup = samples\nbackup = samples.copy()`
            : `samples = [10, 20, 30]\nbackup = samples`}
        </code>
      </pre>

      <div className="flex flex-wrap gap-2 border-t border-line pt-4">
        <button
          className={`${controlClass} border-accent-teal/40 bg-accent-teal/10 text-accent-teal hover:bg-accent-teal/20`}
          type="button"
          onClick={() => appendThrough('samples')}
        >
          Append 40 through <code>samples</code>
        </button>
        <button
          className={`${controlClass} border-accent-blue/40 bg-accent-blue/10 text-accent-blue hover:bg-accent-blue/20`}
          type="button"
          onClick={() => appendThrough('backup')}
        >
          Append 40 through <code>backup</code>
        </button>
        <button
          className={`${controlClass} border-accent-gold/40 bg-accent-gold/10 text-accent-gold hover:bg-accent-gold/20`}
          type="button"
          onClick={copyBackup}
          disabled={bindingState.isCopied}
        >
          Copy <code>backup</code>
        </button>
        <button
          className={`${controlClass} border-line-strong bg-accent-slate/10 text-ink hover:bg-accent-slate/20`}
          type="button"
          onClick={() => setBindingState(createInitialState())}
        >
          Reset
        </button>
      </div>
    </Card>
  )
}

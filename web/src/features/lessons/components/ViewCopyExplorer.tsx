import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader, ValueGrid } from './ArrayVisual'

const initial = [10, 20, 30, 40]

export function ViewCopyExplorer() {
  const [source, setSource] = useState(initial)
  const [view, setView] = useState(initial.slice(1, 3))
  const [copy, setCopy] = useState(initial.slice(1, 3))

  function mutateView() {
    setView([999, view[1]])
    setSource([source[0], 999, source[2], source[3]])
  }

  function mutateCopy() {
    setCopy([555, copy[1]])
  }

  function reset() {
    setSource(initial)
    setView(initial.slice(1, 3))
    setCopy(initial.slice(1, 3))
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="View, copy, and underlying storage"
        description="Different array objects can share numerical storage. An explicit copy owns independent storage."
        badge={<Badge tone="gold">basic slice</Badge>}
      />
      <div className="grid gap-4 lg:grid-cols-3">
        <StoragePanel title="original" storage="storage A" values={source} />
        <StoragePanel title="middle = original[1:3]" storage="storage A (shared)" values={view} />
        <StoragePanel
          title="independent = original[1:3].copy()"
          storage="storage B"
          values={copy}
        />
      </div>
      <div className="flex flex-wrap gap-2">
        <Button onClick={mutateView}>Set middle[0] = 999</Button>
        <Button variant="secondary" onClick={mutateCopy}>
          Set independent[0] = 555
        </Button>
        <Button variant="outline" onClick={reset}>
          Reset
        </Button>
      </div>
      <div
        className="rounded-lg border border-line bg-raised p-4 text-sm leading-relaxed text-muted"
        role="status"
        aria-live="polite"
      >
        <strong className="text-ink">Storage result:</strong> the basic slice mutation appears in{' '}
        <code>original</code>; the <code>.copy()</code> mutation does not. This demonstrates a
        common basic-slice rule, not a claim that every derived array is a view.
      </div>
    </Card>
  )
}

function StoragePanel({
  title,
  storage,
  values,
}: {
  title: string
  storage: string
  values: readonly number[]
}) {
  return (
    <section className="grid gap-3 rounded-lg border border-line bg-raised p-4">
      <h3 className="break-words font-mono text-sm text-ink">{title}</h3>
      <ValueGrid
        values={values}
        columns={values.length}
        label={`${title} contains ${values.join(', ')}`}
      />
      <p className="text-xs font-bold uppercase tracking-widest text-faint">
        array object → {storage}
      </p>
    </section>
  )
}

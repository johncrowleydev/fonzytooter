import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader } from './ArrayVisual'

export function seededSequence(seed: number, count: number, skip = 0) {
  let state = seed >>> 0
  const values: number[] = []
  for (let index = 0; index < count + skip; index += 1) {
    state = (Math.imul(1664525, state) + 1013904223) >>> 0
    if (index >= skip) values.push(Number((state / 2 ** 32).toFixed(4)))
  }
  return values
}

export function SeedReproducibilityExplorer() {
  const [mode, setMode] = useState<'same' | 'different'>('same')
  const [advanceA, setAdvanceA] = useState(0)
  const seedA = 7
  const seedB = mode === 'same' ? 7 : 99
  const sequenceA = seededSequence(seedA, 5, advanceA)
  const sequenceB = seededSequence(seedB, 5)
  const match = sequenceA.every((value, index) => value === sequenceB[index])

  function setComparison(nextMode: 'same' | 'different') {
    setMode(nextMode)
    setAdvanceA(0)
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Seeds are reproducible starting states"
        description="Compare two deterministic pseudorandom generators, then advance one generator and observe how its state changes subsequent draws."
        badge={<Badge tone="blue">same algorithm</Badge>}
      />
      <div className="flex flex-wrap gap-2">
        <Button
          variant={mode === 'same' ? 'secondary' : 'outline'}
          pressed={mode === 'same'}
          onClick={() => setComparison('same')}
        >
          Use the same seed
        </Button>
        <Button
          variant={mode === 'different' ? 'secondary' : 'outline'}
          pressed={mode === 'different'}
          onClick={() => setComparison('different')}
        >
          Use different seeds
        </Button>
        <Button variant="outline" onClick={() => setAdvanceA((current) => current + 1)}>
          Advance generator A once
        </Button>
        <Button
          variant="quiet"
          onClick={() => {
            setMode('same')
            setAdvanceA(0)
          }}
        >
          Reset
        </Button>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <SequencePanel name="Generator A" seed={seedA} advance={advanceA} values={sequenceA} />
        <SequencePanel name="Generator B" seed={seedB} advance={0} values={sequenceB} />
      </div>
      <p
        className={
          match
            ? 'rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-4 text-sm leading-relaxed text-ink'
            : 'rounded-lg border border-accent-gold/30 bg-accent-gold/10 p-4 text-sm leading-relaxed text-ink'
        }
        role="status"
        aria-live="polite"
      >
        <strong>{match ? 'Sequences match.' : 'Sequences differ.'}</strong>{' '}
        {mode === 'same' && advanceA === 0
          ? 'The same algorithm and seed begin from the same state.'
          : mode === 'same'
            ? 'The seed is still the same, but generator A has advanced to a later state.'
            : 'Different seeds select different starting states.'}
      </p>
      <p className="text-sm leading-relaxed text-muted">
        A seed is not statistical magic. Reproducibility means documenting and controlling intended
        variability; experiments may deliberately compare many seeds rather than forcing every
        outcome to be identical.
      </p>
    </Card>
  )
}

function SequencePanel({
  name,
  seed,
  advance,
  values,
}: {
  name: string
  seed: number
  advance: number
  values: readonly number[]
}) {
  return (
    <section className="grid gap-3 rounded-lg border border-line bg-raised p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-sm font-semibold text-ink">{name}</h3>
        <span className="font-mono text-sm text-faint">
          seed={seed}; skipped={advance}
        </span>
      </div>
      <ol className="grid grid-cols-5 gap-2" aria-label={`${name} draws`}>
        {values.map((value, index) => (
          <li
            key={index}
            className="rounded-md border border-line bg-panel px-2 py-2 text-center font-mono text-sm text-muted"
          >
            {value}
          </li>
        ))}
      </ol>
    </section>
  )
}

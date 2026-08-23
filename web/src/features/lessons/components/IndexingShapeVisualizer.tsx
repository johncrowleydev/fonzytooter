import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { formatShape, InteractiveHeader, ValueGrid, type Shape } from './ArrayVisual'

type Selection = {
  expression: string
  indices: readonly number[]
  values: string
  shape: Shape
  reason: string
}

const selections: readonly Selection[] = [
  {
    expression: 'x[1]',
    indices: [4, 5, 6, 7],
    values: '[4, 5, 6, 7]',
    shape: [4],
    reason: 'Integer 1 fixes and removes axis 0.',
  },
  {
    expression: 'x[1, 2]',
    indices: [6],
    values: '6',
    shape: [],
    reason: 'Both axes are fixed, leaving a scalar.',
  },
  {
    expression: 'x[:, 1]',
    indices: [1, 5, 9],
    values: '[1, 5, 9]',
    shape: [3],
    reason: 'The slice preserves axis 0; integer 1 removes axis 1.',
  },
  {
    expression: 'x[1, :]',
    indices: [4, 5, 6, 7],
    values: '[4, 5, 6, 7]',
    shape: [4],
    reason: 'Integer 1 removes axis 0; the slice preserves axis 1.',
  },
  {
    expression: 'x[:, 1:3]',
    indices: [1, 2, 5, 6, 9, 10],
    values: '[[1, 2], [5, 6], [9, 10]]',
    shape: [3, 2],
    reason: 'Both slices preserve axes; the second slice has length 2.',
  },
  {
    expression: 'x[0]',
    indices: [0, 1, 2, 3],
    values: '[0, 1, 2, 3]',
    shape: [4],
    reason: 'Integer indexing fixes and removes axis 0.',
  },
  {
    expression: 'x[0:1]',
    indices: [0, 1, 2, 3],
    values: '[[0, 1, 2, 3]]',
    shape: [1, 4],
    reason: 'A one-position slice preserves axis 0 with length 1.',
  },
]

export function IndexingShapeVisualizer() {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const selection = selections[selectedIndex]
  const values = Array.from({ length: 12 }, (_, index) => index)

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Indexing and slicing change shape"
        description="Use the highlighted values and result shape together. Integer indices remove axes; slices preserve them."
        badge={<Badge tone="violet">x.shape = (3, 4)</Badge>}
      />
      <div className="flex flex-wrap gap-2" role="group" aria-label="Choose an indexing expression">
        {selections.map((candidate, index) => (
          <Button
            key={candidate.expression}
            variant={index === selectedIndex ? 'primary' : 'outline'}
            onClick={() => setSelectedIndex(index)}
          >
            <code>{candidate.expression}</code>
          </Button>
        ))}
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <ValueGrid
          values={values}
          columns={4}
          selected={selection.indices}
          label={`Source x; values selected by ${selection.expression} are marked`}
        />
        <div
          className="grid content-start gap-3 rounded-lg border border-brand-violet/30 bg-brand-violet/10 p-4"
          role="status"
          aria-live="polite"
        >
          <p className="font-mono text-sm text-ink">
            {selection.expression} → {selection.values}
          </p>
          <p className="font-mono text-sm text-brand-violet">
            result shape: {formatShape(selection.shape)}
          </p>
          <p className="text-xs leading-relaxed text-muted">{selection.reason}</p>
        </div>
      </div>
      <p className="rounded-lg border border-brand-gold/30 bg-brand-gold/10 p-4 text-xs leading-relaxed text-ink">
        Compare <code>x[0]</code> with <code>x[0:1]</code>: they select the same values, but only
        the slice keeps axis 0.
      </p>
    </Card>
  )
}

import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { formatShape, InteractiveHeader, ShapeChips, ValueGrid, type Shape } from './ArrayVisual'

const examples: readonly { name: string; shape: Shape; dtype: string }[] = [
  { name: 'signal', shape: [4], dtype: 'float64' },
  { name: 'measurements', shape: [3, 4], dtype: 'int64' },
  { name: 'batch', shape: [2, 3, 4], dtype: 'float32' },
]

export function ArrayStructureExplorer() {
  const [exampleIndex, setExampleIndex] = useState(0)
  const [axis, setAxis] = useState(0)
  const [prediction, setPrediction] = useState<number | null>(null)
  const example = examples[exampleIndex]
  const size = example.shape.reduce((total, length) => total * length, 1)
  const values = Array.from({ length: size }, (_, index) => index)
  const predictionCorrect = prediction === example.shape[axis]

  function chooseExample(index: number) {
    setExampleIndex(index)
    setAxis(0)
    setPrediction(null)
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Array structure and axes"
        description="Inspect 1D, 2D, and 3D arrays. Shape records one length per axis; dtype is a separate property."
        badge={<Badge tone="blue">predict first</Badge>}
      />

      <div className="flex flex-wrap gap-2" role="group" aria-label="Choose an array">
        {examples.map((candidate, index) => (
          <Button
            key={candidate.name}
            variant={index === exampleIndex ? 'primary' : 'outline'}
            pressed={index === exampleIndex}
            onClick={() => chooseExample(index)}
          >
            {candidate.name} {formatShape(candidate.shape)}
          </Button>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(15rem,0.7fr)]">
        <ValueGrid
          values={values}
          columns={example.shape.at(-1) ?? 1}
          label={`${example.name} array with shape ${formatShape(example.shape)}`}
        />
        <div className="grid content-start gap-3 rounded-lg border border-line bg-raised p-4">
          <dl className="grid grid-cols-2 gap-3 text-sm">
            <dt className="text-muted">shape</dt>
            <dd className="font-mono text-ink">{formatShape(example.shape)}</dd>
            <dt className="text-muted">ndim</dt>
            <dd className="font-mono text-ink">{example.shape.length}</dd>
            <dt className="text-muted">size</dt>
            <dd className="font-mono text-ink">{size}</dd>
            <dt className="text-muted">dtype</dt>
            <dd className="font-mono text-ink">{example.dtype}</dd>
          </dl>
          <ShapeChips shape={example.shape} activeAxis={axis} />
        </div>
      </div>

      <fieldset className="grid gap-3 rounded-lg border border-line p-4">
        <legend className="px-1 text-sm font-semibold text-ink">
          Select an axis, then predict its length
        </legend>
        <div className="flex flex-wrap gap-2">
          {example.shape.map((_, candidateAxis) => (
            <Button
              key={candidateAxis}
              variant={candidateAxis === axis ? 'secondary' : 'outline'}
              pressed={candidateAxis === axis}
              onClick={() => {
                setAxis(candidateAxis)
                setPrediction(null)
              }}
            >
              axis {candidateAxis}
            </Button>
          ))}
        </div>
        <div className="flex flex-wrap gap-2">
          {[1, 2, 3, 4].map((length) => (
            <Button key={length} variant="outline" onClick={() => setPrediction(length)}>
              {length}
            </Button>
          ))}
        </div>
        {prediction !== null ? (
          <p
            className={
              predictionCorrect
                ? 'rounded-md border border-accent-teal/30 bg-accent-teal/10 p-3 text-sm text-ink'
                : 'rounded-md border border-accent-coral/30 bg-accent-coral/10 p-3 text-sm text-ink'
            }
            role={predictionCorrect ? 'status' : 'alert'}
          >
            {predictionCorrect
              ? `Correct: axis ${axis} has ${prediction} positions.`
              : `Not yet. Read position ${axis} in the shape tuple ${formatShape(example.shape)}.`}
          </p>
        ) : null}
      </fieldset>
    </Card>
  )
}

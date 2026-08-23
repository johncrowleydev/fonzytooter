import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { formatShape, InteractiveHeader, ShapeChips, type Shape } from './ArrayVisual'

const examples: readonly { label: string; shape: Shape; values: readonly number[] }[] = [
  { label: 'scores', shape: [2, 3], values: [80, 90, 100, 70, 85, 95] },
  { label: 'batch', shape: [2, 3, 4], values: Array.from({ length: 24 }, (_, index) => index) },
]

export function reducedShape(shape: Shape, axis: number): Shape {
  return shape.filter((_, candidateAxis) => candidateAxis !== axis)
}

function unravelIndex(index: number, shape: Shape) {
  const coordinates = Array<number>(shape.length)
  let remainder = index
  for (let dimension = shape.length - 1; dimension >= 0; dimension -= 1) {
    coordinates[dimension] = remainder % shape[dimension]
    remainder = Math.floor(remainder / shape[dimension])
  }
  return coordinates
}

function ravelIndex(coordinates: readonly number[], shape: Shape) {
  return coordinates.reduce(
    (flatIndex, coordinate, dimension) => flatIndex * shape[dimension] + coordinate,
    0,
  )
}

export function reductionGroups(shape: Shape, axis: number, values: readonly number[]) {
  const outputShape = reducedShape(shape, axis)
  const outputSize = outputShape.reduce((total, length) => total * length, 1)
  const axisLength = shape[axis]
  return Array.from({ length: outputSize }, (_, groupIndex) => {
    const outputCoordinates = unravelIndex(groupIndex, outputShape)
    return Array.from({ length: axisLength }, (_, axisPosition) => {
      const sourceCoordinates = [...outputCoordinates]
      sourceCoordinates.splice(axis, 0, axisPosition)
      return values[ravelIndex(sourceCoordinates, shape)]
    })
  })
}

export function ReductionAxisExplorer() {
  const [exampleIndex, setExampleIndex] = useState(0)
  const [axis, setAxis] = useState(0)
  const [prediction, setPrediction] = useState<string | null>(null)
  const example = examples[exampleIndex]
  const resultShape = reducedShape(example.shape, axis)
  const groups = reductionGroups(example.shape, axis, example.values)
  const correct = prediction === formatShape(resultShape)

  function chooseExample(index: number) {
    setExampleIndex(index)
    setAxis(0)
    setPrediction(null)
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Collapse the named reduction axis"
        description="Choose an axis, predict the result shape, then inspect which values each output summarizes."
        badge={<Badge tone="gold">axis collapses</Badge>}
      />
      <div className="flex flex-wrap gap-2">
        {examples.map((candidate, index) => (
          <Button
            key={candidate.label}
            variant={exampleIndex === index ? 'primary' : 'outline'}
            onClick={() => chooseExample(index)}
          >
            {candidate.label} {formatShape(candidate.shape)}
          </Button>
        ))}
      </div>
      <ShapeChips shape={example.shape} activeAxis={axis} />
      <fieldset className="grid gap-3 rounded-lg border border-line p-4">
        <legend className="px-1 text-xs font-semibold text-ink">
          Which axis should mean() collapse?
        </legend>
        <div className="flex flex-wrap gap-2">
          {example.shape.map((_, candidateAxis) => (
            <Button
              key={candidateAxis}
              variant={axis === candidateAxis ? 'secondary' : 'outline'}
              onClick={() => {
                setAxis(candidateAxis)
                setPrediction(null)
              }}
            >
              axis {candidateAxis}
            </Button>
          ))}
        </div>
        <p className="text-xs text-muted">Predict the result shape after axis {axis} is removed:</p>
        <div className="flex flex-wrap gap-2">
          {Array.from(
            new Set(
              example.shape.map((_, candidateAxis) =>
                formatShape(reducedShape(example.shape, candidateAxis)),
              ),
            ),
          ).map((shape) => (
            <Button key={shape} variant="outline" onClick={() => setPrediction(shape)}>
              {shape}
            </Button>
          ))}
        </div>
      </fieldset>
      {prediction !== null ? (
        <div
          className={
            correct
              ? 'grid gap-3 rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-4'
              : 'rounded-lg border border-brand-coral/30 bg-brand-coral/10 p-4'
          }
          role={correct ? 'status' : 'alert'}
        >
          <p className="text-xs leading-relaxed text-ink">
            <strong>{correct ? 'Correct.' : 'Not yet.'}</strong> axis {axis} is the dimension being
            collapsed; the other axes keep their order.
          </p>
          {correct ? (
            <>
              <p className="font-mono text-sm text-ink">
                {formatShape(example.shape)} → {formatShape(resultShape)}
              </p>
              <div className="grid gap-2 sm:grid-cols-2">
                {groups.slice(0, 6).map((group, index) => (
                  <div
                    key={index}
                    className="rounded-md border border-line bg-panel px-3 py-2 text-xs text-muted"
                  >
                    output {index}: mean({group.join(', ')})
                  </div>
                ))}
              </div>
              {groups.length > 6 ? (
                <p className="text-2xs text-muted">
                  Showing the first 6 of {groups.length} output groups.
                </p>
              ) : null}
            </>
          ) : null}
        </div>
      ) : null}
    </Card>
  )
}

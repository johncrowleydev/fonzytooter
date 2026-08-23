import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { formatShape, InteractiveHeader, type Shape } from './ArrayVisual'

export type BroadcastComparison = {
  left: number
  right: number
  compatible: boolean
  result: number | null
}
export type BroadcastResult = {
  compatible: boolean
  shape: Shape | null
  comparisons: readonly BroadcastComparison[]
}

export function broadcastShapes(left: Shape, right: Shape): BroadcastResult {
  const width = Math.max(left.length, right.length)
  const paddedLeft = [...Array(width - left.length).fill(1), ...left]
  const paddedRight = [...Array(width - right.length).fill(1), ...right]
  const comparisons = paddedLeft.map((leftLength, index) => {
    const rightLength = paddedRight[index]
    const compatible = leftLength === rightLength || leftLength === 1 || rightLength === 1
    return {
      left: leftLength,
      right: rightLength,
      compatible,
      result: compatible ? Math.max(leftLength, rightLength) : null,
    }
  })
  const compatible = comparisons.every((comparison) => comparison.compatible)
  return {
    compatible,
    shape: compatible ? comparisons.map((comparison) => comparison.result!) : null,
    comparisons,
  }
}

const cases: readonly { left: Shape; right: Shape; shapeOptions: readonly Shape[] }[] = [
  { left: [5, 4], right: [4], shapeOptions: [[4], [1, 4], [5, 4]] },
  {
    left: [2, 5, 4],
    right: [5, 4],
    shapeOptions: [
      [5, 4],
      [2, 4],
      [2, 5, 4],
    ],
  },
  {
    left: [2, 5, 4],
    right: [1, 4],
    shapeOptions: [
      [2, 1, 4],
      [5, 4],
      [2, 5, 4],
    ],
  },
  {
    left: [3, 1],
    right: [1, 7],
    shapeOptions: [
      [3, 1],
      [1, 7],
      [3, 7],
    ],
  },
  { left: [8, 3], right: [8], shapeOptions: [] },
  {
    left: [4, 1, 6],
    right: [3, 6],
    shapeOptions: [
      [4, 1, 6],
      [3, 6],
      [4, 3, 6],
    ],
  },
]

export function BroadcastShapeLab() {
  const [caseIndex, setCaseIndex] = useState(0)
  const [compatibilityPrediction, setCompatibilityPrediction] = useState<'works' | 'fails' | null>(
    null,
  )
  const [shapePrediction, setShapePrediction] = useState<string | null>(null)
  const current = cases[caseIndex]
  const result = broadcastShapes(current.left, current.right)
  const compatibilityCorrect = compatibilityPrediction === (result.compatible ? 'works' : 'fails')
  const resultShapeLabel = result.shape ? formatShape(result.shape) : null
  const shapeCorrect = shapePrediction !== null && shapePrediction === resultShapeLabel
  const terminal = compatibilityCorrect && (!result.compatible || shapeCorrect)
  const width = Math.max(current.left.length, current.right.length)
  const paddedLeft = [...Array(width - current.left.length).fill(1), ...current.left]
  const paddedRight = [...Array(width - current.right.length).fill(1), ...current.right]

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Broadcasting shape lab"
        description="Right-align first, compare every dimension, then decide whether the operation works and predict its result shape."
        badge={<Badge tone="blue">equal or 1</Badge>}
      />
      <div className="flex flex-wrap gap-2" aria-label="Broadcasting cases">
        {cases.map((candidate, index) => (
          <Button
            key={`${formatShape(candidate.left)}-${formatShape(candidate.right)}`}
            variant={index === caseIndex ? 'secondary' : 'outline'}
            onClick={() => {
              setCaseIndex(index)
              setCompatibilityPrediction(null)
              setShapePrediction(null)
            }}
          >
            Case {index + 1}
          </Button>
        ))}
      </div>
      <div
        className="overflow-x-auto rounded-lg border border-line bg-brand-ink p-4 font-mono text-sm text-slate-100"
        aria-label="Right-aligned shapes"
      >
        <div
          className="grid min-w-max gap-2"
          style={{ gridTemplateColumns: `auto repeat(${width}, 3rem)` }}
        >
          <span>left</span>
          {paddedLeft.map((dimension, index) => (
            <span key={index} className="rounded border border-slate-600 px-2 py-1 text-center">
              {dimension}
            </span>
          ))}
          <span>right</span>
          {paddedRight.map((dimension, index) => (
            <span key={index} className="rounded border border-slate-600 px-2 py-1 text-center">
              {dimension}
            </span>
          ))}
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() => {
            setCompatibilityPrediction('works')
            setShapePrediction(null)
          }}
          disabled={compatibilityCorrect}
        >
          Compatible
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            setCompatibilityPrediction('fails')
            setShapePrediction(null)
          }}
          disabled={compatibilityCorrect}
        >
          Incompatible
        </Button>
      </div>
      {compatibilityPrediction !== null && !compatibilityCorrect ? (
        <div className="rounded-lg border border-brand-coral/30 bg-brand-coral/10 p-4" role="alert">
          <p className="text-xs text-ink">
            <strong>Not yet.</strong> Compare from the right; missing leading dimensions are shown
            as 1.
          </p>
        </div>
      ) : null}
      {compatibilityCorrect && result.compatible ? (
        <fieldset className="grid gap-3 rounded-lg border border-brand-blue/30 bg-brand-blue/10 p-4">
          <legend className="px-1 text-xs font-semibold text-ink">Predict the result shape</legend>
          <p className="text-xs leading-relaxed text-muted">
            Compatible. For each aligned position, take the larger compatible dimension.
          </p>
          <div className="flex flex-wrap gap-2">
            {current.shapeOptions.map((shape) => {
              const label = formatShape(shape)
              return (
                <Button
                  key={label}
                  variant={shapePrediction === label ? 'secondary' : 'outline'}
                  onClick={() => setShapePrediction(label)}
                  disabled={shapeCorrect}
                >
                  {label}
                </Button>
              )
            })}
          </div>
          {shapePrediction !== null && !shapeCorrect ? (
            <p
              className="rounded-md border border-brand-coral/30 bg-brand-coral/10 p-3 text-xs text-ink"
              role="alert"
            >
              <strong>Not yet.</strong> Build the result from the larger compatible dimension at
              every aligned position.
            </p>
          ) : null}
        </fieldset>
      ) : null}
      {terminal ? (
        <div
          className="grid gap-3 rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-4"
          role="status"
        >
          <p className="text-xs text-ink">
            <strong>Correct.</strong>{' '}
            {result.compatible
              ? 'The compatibility decision and result-shape prediction both follow the right-aligned rule.'
              : 'The first unequal pair without a 1 makes failure the terminal result.'}
          </p>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {result.comparisons.map((comparison, index) => (
              <p
                key={index}
                className="rounded-md border border-line bg-panel p-3 text-xs text-muted"
              >
                aligned axis {index}: {comparison.left} vs {comparison.right} —{' '}
                {comparison.compatible ? 'compatible' : 'incompatible'}
                {comparison.result === null ? '' : ` → ${comparison.result}`}
              </p>
            ))}
          </div>
          <p className="font-mono text-sm text-ink">
            result: {resultShapeLabel ?? 'broadcasting fails'}
          </p>
        </div>
      ) : null}
    </Card>
  )
}

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

const cases: readonly { left: Shape; right: Shape }[] = [
  { left: [5, 4], right: [4] },
  { left: [2, 5, 4], right: [5, 4] },
  { left: [2, 5, 4], right: [1, 4] },
  { left: [3, 1], right: [1, 7] },
  { left: [8, 3], right: [8] },
  { left: [4, 1, 6], right: [3, 6] },
]

export function BroadcastShapeLab() {
  const [caseIndex, setCaseIndex] = useState(0)
  const [prediction, setPrediction] = useState<'works' | 'fails' | null>(null)
  const current = cases[caseIndex]
  const result = broadcastShapes(current.left, current.right)
  const correct = prediction === (result.compatible ? 'works' : 'fails')
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
              setPrediction(null)
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
        <Button onClick={() => setPrediction('works')}>Compatible</Button>
        <Button variant="outline" onClick={() => setPrediction('fails')}>
          Incompatible
        </Button>
      </div>
      {prediction !== null ? (
        <div
          className={
            correct
              ? 'grid gap-3 rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-4'
              : 'rounded-lg border border-brand-coral/30 bg-brand-coral/10 p-4'
          }
          role={correct ? 'status' : 'alert'}
        >
          <p className="text-xs text-ink">
            <strong>{correct ? 'Correct.' : 'Not yet.'}</strong> Compare from the right; missing
            leading dimensions are shown as 1.
          </p>
          {correct ? (
            <>
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
                result: {result.shape ? formatShape(result.shape) : 'broadcasting fails'}
              </p>
            </>
          ) : null}
        </div>
      ) : null}
    </Card>
  )
}

import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader } from './ArrayVisual'
import { HighlightedCode } from '../../../components/HighlightedCode'

type Kind = 'element-wise transformation' | 'reduction' | 'vectorized comparison' | 'composition'

const questions: readonly { expression: string; kind: Kind; shape: string; explanation: string }[] =
  [
    {
      expression: 'x * 2',
      kind: 'element-wise transformation',
      shape: '(2, 3)',
      explanation: 'Every input position produces one transformed output position.',
    },
    {
      expression: 'np.sqrt(x)',
      kind: 'element-wise transformation',
      shape: '(2, 3)',
      explanation: 'The ufunc applies independently and preserves shape.',
    },
    {
      expression: 'x.mean(axis=1)',
      kind: 'reduction',
      shape: '(2,)',
      explanation: 'Axis 1 is collapsed from x.shape == (2, 3).',
    },
    {
      expression: 'x > 0',
      kind: 'vectorized comparison',
      shape: '(2, 3)',
      explanation: 'Each position produces one boolean result.',
    },
    {
      expression: '(x > 0).sum()',
      kind: 'composition',
      shape: 'scalar',
      explanation: 'A comparison creates a boolean array, then sum reduces all positions.',
    },
  ]

const kinds: readonly Kind[] = [
  'element-wise transformation',
  'reduction',
  'vectorized comparison',
  'composition',
]

export function OperationKindCheck() {
  const [index, setIndex] = useState(0)
  const [answer, setAnswer] = useState<Kind | null>(null)
  const [complete, setComplete] = useState<Set<number>>(() => new Set())
  const question = questions[index]
  const correct = answer === question.kind

  function choose(kind: Kind) {
    setAnswer(kind)
    if (kind === question.kind) setComplete((current) => new Set(current).add(index))
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Classify the array expression"
        description="Assume x.shape == (2, 3). Classify each expression and connect its operation kind to its result shape."
        badge={
          <Badge tone={complete.size === questions.length ? 'teal' : 'neutral'}>
            {complete.size} / {questions.length}
          </Badge>
        }
      />
      <div className="flex flex-wrap gap-2">
        {questions.map((candidate, candidateIndex) => (
          <Button
            key={candidate.expression}
            variant={candidateIndex === index ? 'secondary' : 'outline'}
            pressed={candidateIndex === index}
            onClick={() => {
              setIndex(candidateIndex)
              setAnswer(null)
            }}
          >
            {candidateIndex + 1}
            {complete.has(candidateIndex) ? ' ✓' : ''}
          </Button>
        ))}
      </div>
      <code className="rounded-lg border border-line bg-code-surface p-4 text-sm text-code-ink">
        <HighlightedCode code={question.expression} language="python" />
      </code>
      <div className="grid gap-2 sm:grid-cols-2" role="group" aria-label="Operation kinds">
        {kinds.map((kind) => (
          <Button
            key={kind}
            variant={answer === kind ? 'secondary' : 'outline'}
            pressed={answer === kind}
            onClick={() => choose(kind)}
            disabled={correct}
          >
            {kind}
          </Button>
        ))}
      </div>
      {answer !== null ? (
        <p
          className={
            correct
              ? 'rounded-md border border-accent-teal/30 bg-accent-teal/10 p-3 text-sm text-ink'
              : 'rounded-md border border-accent-coral/30 bg-accent-coral/10 p-3 text-sm text-ink'
          }
          role={correct ? 'status' : 'alert'}
        >
          <strong>{correct ? 'Correct.' : 'Try again.'}</strong> {question.explanation} Result
          shape: <code>{question.shape}</code>.
        </p>
      ) : null}
    </Card>
  )
}

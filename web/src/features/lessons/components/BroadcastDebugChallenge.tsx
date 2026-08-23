import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader } from './ArrayVisual'
import { HighlightedCode } from '../../../components/HighlightedCode'

type Challenge = 'row-means' | 'silent-grid'

const challengeContent = {
  'row-means': {
    title: 'Center each row of x.shape == (4, 3)',
    expression: 'x - x.mean(axis=1)',
    shapes: '(4, 3) with (4,) → compares 3 with 4 → fails',
    options: ['Use mean(axis=0)', 'Use mean(axis=1, keepdims=True)', 'Tile x four times'],
    correct: 1,
    feedback:
      'keepdims=True preserves the collapsed axis as length 1: (4, 1). That shape broadcasts across the three values in each row.',
  },
  'silent-grid': {
    title: 'Produce three pairwise sums, not a grid',
    expression: 'left.shape == (3, 1); right.shape == (3,); left + right',
    shapes: '(3, 1) with (1, 3) → legal result (3, 3)',
    options: [
      'Flatten left to (3,) before adding',
      'Add keepdims=True to right',
      'Trust the legal broadcast',
    ],
    correct: 0,
    feedback:
      'Matching both operands as (3,) expresses pairwise intent and produces (3,). Legal broadcasting is not proof that the operation is semantically right.',
  },
} as const

export function BroadcastDebugChallenge() {
  const [challenge, setChallenge] = useState<Challenge>('row-means')
  const [answer, setAnswer] = useState<number | null>(null)
  const content = challengeContent[challenge]
  const correct = answer === content.correct

  function chooseChallenge(next: Challenge) {
    setChallenge(next)
    setAnswer(null)
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Debug intent with shapes"
        description="One operation fails despite clear human intent; another succeeds while doing the wrong thing. Diagnose both from shape."
        badge={<Badge tone="coral">shape debugger</Badge>}
      />
      <div className="flex flex-wrap gap-2">
        <Button
          variant={challenge === 'row-means' ? 'secondary' : 'outline'}
          pressed={challenge === 'row-means'}
          onClick={() => chooseChallenge('row-means')}
        >
          Row means failure
        </Button>
        <Button
          variant={challenge === 'silent-grid' ? 'secondary' : 'outline'}
          pressed={challenge === 'silent-grid'}
          onClick={() => chooseChallenge('silent-grid')}
        >
          Legal but wrong
        </Button>
      </div>
      <section className="grid gap-3 rounded-lg border border-line bg-raised p-4">
        <h3 className="text-sm font-semibold text-ink">{content.title}</h3>
        <code className="overflow-x-auto overscroll-x-contain rounded-md bg-code-surface p-3 text-sm text-code-ink">
          <HighlightedCode code={content.expression} language="python" />
        </code>
        <p className="font-mono text-sm text-muted">{content.shapes}</p>
      </section>
      <div className="grid gap-2" role="group" aria-label="Debugging fixes">
        {content.options.map((option, index) => (
          <Button
            key={option}
            variant={answer === index ? 'secondary' : 'outline'}
            pressed={answer === index}
            onClick={() => setAnswer(index)}
            disabled={correct}
          >
            {option}
          </Button>
        ))}
      </div>
      {answer !== null ? (
        <p
          className={
            correct
              ? 'rounded-md border border-accent-teal/30 bg-accent-teal/10 p-3 text-sm leading-relaxed text-ink'
              : 'rounded-md border border-accent-coral/30 bg-accent-coral/10 p-3 text-sm leading-relaxed text-ink'
          }
          role={correct ? 'status' : 'alert'}
        >
          <strong>{correct ? 'Correct.' : 'Try again.'}</strong> {content.feedback}
        </p>
      ) : null}
    </Card>
  )
}

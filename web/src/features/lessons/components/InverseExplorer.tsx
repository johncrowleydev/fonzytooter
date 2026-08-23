import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'

type SquareDomain = 'all' | 'nonnegative' | 'nonpositive'
type Feedback = { tone: 'success' | 'error'; message: string } | null

const domainLabels: Record<SquareDomain, string> = {
  all: 'all real numbers',
  nonnegative: 'x ≥ 0',
  nonpositive: 'x ≤ 0',
}
const feedbackClasses = {
  success: 'border-accent-teal/30 bg-accent-teal/10',
  error: 'border-accent-coral/30 bg-accent-coral/10',
} as const

export function InverseExplorer() {
  const [stage, setStage] = useState(0)
  const [domain, setDomain] = useState<SquareDomain>('all')
  const [feedback, setFeedback] = useState<Feedback>(null)
  const [reverseStep, setReverseStep] = useState(0)
  const stageCount = 4
  const complete = stage === stageCount

  function answerInformationLoss(answer: string) {
    setFeedback(
      answer === 'either'
        ? {
            tone: 'success',
            message:
              'Squaring erased the sign. The output 9 does not contain enough information to determine whether 3 or -3 produced it.',
          }
        : {
            tone: 'error',
            message:
              'Both 3 and -3 square to 9. Over all real numbers, the output alone cannot distinguish them.',
          },
    )
  }

  function chooseDomain(nextDomain: SquareDomain) {
    setDomain(nextDomain)
    if (nextDomain === 'all') {
      setFeedback({
        tone: 'error',
        message:
          'Both branches remain, so 9 still has two possible predecessors. The square function is not injective on all real numbers.',
      })
      return
    }

    setFeedback({
      tone: 'success',
      message:
        nextDomain === 'nonnegative'
          ? 'The negative branch is excluded. With codomain [0, ∞), the inverse is f⁻¹(x) = √x.'
          : 'The positive branch is excluded. With codomain [0, ∞), the inverse is f⁻¹(x) = -√x.',
    })
  }

  function answerCodomain(canHandleNegative: boolean) {
    setFeedback(
      canHandleNegative
        ? {
            tone: 'error',
            message:
              'No real input in [0, ∞) squares to -4, so an inverse has no original input to return.',
          }
        : {
            tone: 'success',
            message:
              'Correct. The function is not surjective onto ℝ because negative values are gaps. Declaring codomain [0, ∞) removes those gaps.',
          },
    )
  }

  function advance() {
    setStage((current) => current + 1)
    setFeedback(null)
  }

  function resetStage() {
    if (stage <= 1) setDomain('all')
    if (stage === 3) setReverseStep(0)
    setFeedback(null)
  }

  function applyReverseOperation() {
    setReverseStep((current) => Math.min(current + 1, 2))
  }

  return (
    <Card className="my-8 grid min-w-0 gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">Inverse explorer</h2>
          <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted">
            Reversibility depends on whether the output preserves enough information and covers the
            declared codomain.
          </p>
        </div>
        <Badge tone={complete ? 'teal' : 'blue'}>
          {complete ? 'Complete' : `${stage + 1} / ${stageCount}`}
        </Badge>
      </div>

      {complete ? (
        <div className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-5">
          <p className="text-xs font-bold uppercase tracking-widest text-accent-teal">
            Reversibility
          </p>
          <p className="mt-2 text-sm leading-relaxed text-ink">
            Injectivity removes ambiguity going backward; surjectivity ensures every destination has
            a predecessor. Together, bijectivity makes the declared mapping fully reversible.
          </p>
        </div>
      ) : (
        <>
          {stage <= 1 ? <SquareGraph domain={domain} /> : null}
          {stage === 0 ? <InformationLossStage onAnswer={answerInformationLoss} /> : null}
          {stage === 1 ? <DomainStage domain={domain} onChoose={chooseDomain} /> : null}
          {stage === 2 ? <CodomainStage onAnswer={answerCodomain} /> : null}
          {stage === 3 ? (
            <ReverseOperationsStage step={reverseStep} onStep={applyReverseOperation} />
          ) : null}

          {feedback ? (
            <div
              className={`rounded-lg border p-3 text-sm leading-relaxed text-ink ${feedbackClasses[feedback.tone]}`}
              role={feedback.tone === 'error' ? 'alert' : 'status'}
              aria-live="polite"
            >
              <strong>{feedback.tone === 'success' ? '✓ Correct.' : '✗ Look again.'}</strong>{' '}
              {feedback.message}
            </div>
          ) : null}

          {stage === 3 && reverseStep === 2 ? (
            <div
              className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-3 text-sm leading-relaxed text-ink"
              role="status"
              aria-live="polite"
            >
              <strong>✓ Reversed.</strong> f⁻¹(y) = (y - 3) / 2, and f⁻¹(11) = 4.
            </div>
          ) : null}

          <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
            <Button variant="secondary" onClick={resetStage}>
              Reset stage
            </Button>
            {(feedback?.tone === 'success' || (stage === 3 && reverseStep === 2)) && (
              <Button onClick={advance}>Continue</Button>
            )}
          </div>
        </>
      )}
    </Card>
  )
}

function SquareGraph({ domain }: { domain: SquareDomain }) {
  const values = Array.from({ length: 81 }, (_, index) => -4 + index / 10).filter((x) => {
    if (domain === 'nonnegative') return x >= 0
    if (domain === 'nonpositive') return x <= 0
    return true
  })
  const path = values
    .map((x, index) => `${index === 0 ? 'M' : 'L'} ${300 + x * 60} ${260 - x * x * 13}`)
    .join(' ')
  const showPositive = domain !== 'nonpositive'
  const showNegative = domain !== 'nonnegative'

  return (
    <div className="grid gap-3 rounded-lg border border-line bg-panel-soft p-3 sm:grid-cols-2 sm:items-center">
      {/*
        Text inside this viewBox is in user units, scaled to the rendered width, so it is not on
        the document type scale. Label coordinates are hand-tuned around their widths; growing the
        size would push "(-3, 9)" over its own point.
      */}
      <svg
        className="block w-full min-w-0 max-w-full"
        viewBox="0 0 600 300"
        role="img"
        aria-label={`Graph of y equals x squared with domain ${domainLabels[domain]}. ${showPositive ? 'The point 3, 9 is shown.' : ''} ${showNegative ? 'The point negative 3, 9 is shown.' : ''}`}
      >
        <line className="stroke-line-strong" x1="40" y1="260" x2="560" y2="260" strokeWidth="2" />
        <line className="stroke-line-strong" x1="300" y1="20" x2="300" y2="280" strokeWidth="2" />
        <path className="fill-none stroke-accent-blue" d={path} strokeWidth="5" />
        {showPositive ? (
          <g>
            <circle className="fill-accent-gold" cx="480" cy="143" r="8" />
            <text className="fill-ink text-xs" x="490" y="133">
              (3, 9)
            </text>
          </g>
        ) : null}
        {showNegative ? (
          <g>
            <circle className="fill-accent-gold" cx="120" cy="143" r="8" />
            <text className="fill-ink text-xs" x="52" y="133">
              (-3, 9)
            </text>
          </g>
        ) : null}
        <text className="fill-muted text-xs" x="548" y="280">
          x
        </text>
        <text className="fill-muted text-xs" x="315" y="28">
          y
        </text>
      </svg>
      <div className="rounded-lg border border-line bg-panel p-4">
        <p className="text-xs font-bold uppercase tracking-widest text-faint">Declared mapping</p>
        <code className="mt-3 block text-sm text-ink">f(x) = x²</code>
        <p className="mt-3 text-sm leading-relaxed text-muted">Domain: {domainLabels[domain]}</p>
        <p className="mt-2 text-sm leading-relaxed text-muted">Codomain: [0, ∞)</p>
        <div className="mt-3 grid gap-1 font-mono text-sm text-ink">
          {showPositive ? <span>3 → 9</span> : null}
          {showNegative ? <span>-3 → 9</span> : null}
        </div>
      </div>
    </div>
  )
}

function InformationLossStage({ onAnswer }: { onAnswer: (answer: string) => void }) {
  return (
    <div className="grid gap-3">
      <div>
        <h3 className="text-base font-semibold text-ink">Information loss</h3>
        <p className="mt-2 text-sm leading-relaxed text-muted">
          If all you know is that the output was 9, what was the input?
        </p>
      </div>
      <div className="grid gap-2 sm:grid-cols-3" role="group" aria-label="Possible input">
        <Button variant="outline" onClick={() => onAnswer('positive')}>
          3
        </Button>
        <Button variant="outline" onClick={() => onAnswer('negative')}>
          -3
        </Button>
        <Button onClick={() => onAnswer('either')}>Either 3 or -3</Button>
      </div>
    </div>
  )
}

function DomainStage({
  domain,
  onChoose,
}: {
  domain: SquareDomain
  onChoose: (domain: SquareDomain) => void
}) {
  return (
    <div className="grid gap-3">
      <div>
        <h3 className="text-base font-semibold text-ink">Change the domain</h3>
        <p className="mt-2 text-sm leading-relaxed text-muted">
          Choose a domain. Which restrictions keep only one branch and make squaring injective?
        </p>
      </div>
      <div className="grid gap-2 sm:grid-cols-3" role="group" aria-label="Domain restriction">
        {(Object.keys(domainLabels) as SquareDomain[]).map((option) => (
          <Button
            key={option}
            variant={domain === option ? 'primary' : 'outline'}
            onClick={() => onChoose(option)}
          >
            {domainLabels[option]}
          </Button>
        ))}
      </div>
    </div>
  )
}

function CodomainStage({ onAnswer }: { onAnswer: (answer: boolean) => void }) {
  return (
    <div className="grid gap-4">
      <div>
        <h3 className="text-base font-semibold text-ink">Codomain matters too</h3>
        <p className="mt-2 text-sm leading-relaxed text-muted">
          Compare f : [0, ∞) → ℝ with f : [0, ∞) → [0, ∞), using f(x) = x².
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="rounded-lg border border-accent-coral/30 bg-accent-coral/10 p-4">
          <code className="text-sm text-ink">f : [0, ∞) → ℝ</code>
          <p className="mt-2 text-sm text-muted">Negative codomain values are gaps.</p>
        </div>
        <div className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-4">
          <code className="text-sm text-ink">f : [0, ∞) → [0, ∞)</code>
          <p className="mt-2 text-sm text-muted">Every declared destination is reached.</p>
        </div>
      </div>
      <p className="text-sm leading-relaxed text-muted">
        Can a full inverse f⁻¹ : ℝ → [0, ∞) handle the input -4?
      </p>
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => onAnswer(true)}>
          Yes
        </Button>
        <Button onClick={() => onAnswer(false)}>No — it has no predecessor</Button>
      </div>
    </div>
  )
}

function ReverseOperationsStage({ step, onStep }: { step: number; onStep: () => void }) {
  return (
    <div className="grid gap-4">
      <div>
        <h3 className="text-base font-semibold text-ink">Reverse the operations</h3>
        <p className="mt-2 text-sm leading-relaxed text-muted">
          f(x) = 2x + 3 sends 4 → 11. Walk backward from 11.
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2 rounded-lg border border-line bg-panel-soft p-4 font-mono text-sm text-ink">
        <span>11</span>
        {step >= 1 ? (
          <>
            <span aria-hidden="true">→</span>
            <span>subtract 3 → 8</span>
          </>
        ) : null}
        {step >= 2 ? (
          <>
            <span aria-hidden="true">→</span>
            <span>divide by 2 → 4</span>
          </>
        ) : null}
      </div>
      {step < 2 ? (
        <Button className="w-max" onClick={onStep}>
          {step === 0 ? 'Subtract 3' : 'Divide by 2'}
        </Button>
      ) : null}
    </div>
  )
}

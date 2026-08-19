import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'

type Feedback = { tone: 'success' | 'error'; message: string } | null
const feedbackClasses = {
  success: 'border-brand-teal/30 bg-brand-teal/10',
  error: 'border-brand-coral/30 bg-brand-coral/10',
} as const

export function CompositionPipeline() {
  const [stage, setStage] = useState(0)
  const [input, setInput] = useState(4)
  const [feedback, setFeedback] = useState<Feedback>(null)
  const [swappedRevealed, setSwappedRevealed] = useState(false)
  const stageCount = 4
  const complete = stage === stageCount

  function answerOrder(willMatch: boolean) {
    setSwappedRevealed(true)
    setFeedback(
      willMatch
        ? {
            tone: 'error',
            message:
              'Swapping changes which operation sees the original input. Compare the two pipeline results.',
          }
        : {
            tone: 'success',
            message:
              'Correct. g ∘ f applies doubling before adding 3; f ∘ g adds 3 before doubling. The order generally changes the result.',
          },
    )
  }

  function answerCompatibility(choice: 'uppercase' | 'formatAge') {
    setFeedback(
      choice === 'formatAge'
        ? {
            tone: 'success',
            message:
              'formatAge(parseAge("42")) composes: parseAge produces a number, and formatAge accepts a number. This is f : A → B followed by g : B → C.',
          }
        : {
            tone: 'error',
            message:
              'parseAge outputs a number, but uppercase expects a string. uppercase(parseAge("42")) is structurally incompatible.',
          },
    )
  }

  function advance() {
    setStage((current) => current + 1)
    setFeedback(null)
  }

  function resetStage() {
    if (stage <= 1) setInput(4)
    if (stage === 1) setSwappedRevealed(false)
    setFeedback(null)
  }

  return (
    <Card className="my-8 grid min-w-0 gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">Composition pipeline</h2>
          <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">
            Follow one structure through mathematical notation, nested calls, typed stages, and an
            ML model.
          </p>
        </div>
        <Badge tone={complete ? 'teal' : 'gold'}>
          {complete ? 'Complete' : `${stage + 1} / ${stageCount}`}
        </Badge>
      </div>

      {complete ? (
        <div className="rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-5">
          <p className="text-2xs font-bold uppercase tracking-widest text-brand-teal">
            Scaled-up composition
          </p>
          <p className="mt-2 text-sm leading-relaxed text-ink">
            A neural network is a composition of parameterized functions. Its notation is a larger
            version of the same pipeline structure.
          </p>
        </div>
      ) : (
        <>
          {stage === 0 ? <ValuePipeline input={input} onInput={setInput} /> : null}
          {stage === 1 ? (
            <OrderStage input={input} revealed={swappedRevealed} onAnswer={answerOrder} />
          ) : null}
          {stage === 2 ? <CompatibilityStage onAnswer={answerCompatibility} /> : null}
          {stage === 3 ? <MlPipeline /> : null}

          {feedback ? (
            <div
              className={`rounded-lg border p-3 text-xs leading-relaxed text-ink ${feedbackClasses[feedback.tone]}`}
              role={feedback.tone === 'error' ? 'alert' : 'status'}
              aria-live="polite"
            >
              <strong>{feedback.tone === 'success' ? '✓ Correct.' : '✗ Not yet.'}</strong>{' '}
              {feedback.message}
            </div>
          ) : null}

          <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
            <Button variant="secondary" onClick={resetStage}>
              Reset stage
            </Button>
            {stage === 0 ? <Button onClick={advance}>Compare the other order</Button> : null}
            {stage === 3 ? <Button onClick={advance}>Complete pipeline</Button> : null}
            {(stage === 1 || stage === 2) && feedback?.tone === 'success' ? (
              <Button onClick={advance}>Continue</Button>
            ) : null}
          </div>
        </>
      )}
    </Card>
  )
}

function ValuePipeline({ input, onInput }: { input: number; onInput: (value: number) => void }) {
  const afterF = input * 2
  const result = afterF + 3

  return (
    <div className="grid gap-5">
      <div>
        <h3 className="text-base font-semibold text-ink">Follow a value</h3>
        <p className="mt-2 text-xs leading-relaxed text-muted">
          f(x) = 2x and g(x) = x + 3. Change x and watch every representation stay synchronized.
        </p>
      </div>
      <label className="grid gap-2 rounded-lg border border-line bg-white/5 p-4 text-xs text-muted">
        <span className="flex items-center justify-between gap-3">
          <strong className="text-ink">Input x</strong>
          <output className="font-mono text-sm text-brand-gold">{input}</output>
        </span>
        <input
          className="w-full accent-brand-teal"
          type="range"
          min="-5"
          max="10"
          value={input}
          onChange={(event) => onInput(Number(event.target.value))}
        />
      </label>
      <PipelineNodes
        nodes={[String(input), 'f(x) = 2x', String(afterF), 'g(x) = x + 3', String(result)]}
      />
      <div
        className="grid gap-2 rounded-lg border border-brand-blue/30 bg-brand-blue/10 p-4 font-mono text-sm text-ink sm:grid-cols-2"
        role="status"
        aria-live="polite"
      >
        <span>
          (g ∘ f)({input}) = {result}
        </span>
        <span>
          g(f({input})) = {result}
        </span>
      </div>
    </div>
  )
}

function OrderStage({
  input,
  revealed,
  onAnswer,
}: {
  input: number
  revealed: boolean
  onAnswer: (willMatch: boolean) => void
}) {
  const original = input * 2 + 3
  const afterG = input + 3
  const swapped = afterG * 2

  return (
    <div className="grid gap-4">
      <div>
        <h3 className="text-base font-semibold text-ink">Order matters</h3>
        <p className="mt-2 text-xs leading-relaxed text-muted">
          (g ∘ f)({input}) produced {original}. Will swapping the order still produce {original}?
        </p>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => onAnswer(true)}>
          Yes
        </Button>
        <Button onClick={() => onAnswer(false)}>No</Button>
      </div>
      {revealed ? (
        <div className="grid gap-3" role="status" aria-live="polite">
          <PipelineNodes
            nodes={[String(input), 'g(x) = x + 3', String(afterG), 'f(x) = 2x', String(swapped)]}
          />
          <div className="rounded-lg border border-brand-violet/30 bg-brand-violet/10 p-4 font-mono text-sm text-ink">
            (f ∘ g)({input}) = f(g({input})) = {swapped}
          </div>
        </div>
      ) : null}
    </div>
  )
}

function CompatibilityStage({
  onAnswer,
}: {
  onAnswer: (choice: 'uppercase' | 'formatAge') => void
}) {
  return (
    <div className="grid gap-5">
      <div>
        <h3 className="text-base font-semibold text-ink">Pipeline compatibility</h3>
        <p className="mt-2 text-xs leading-relaxed text-muted">
          parseAge("42") produces a number. Which stage can consume that intermediate value?
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-3">
        <FunctionSignature name="parseAge" signature="string → number" />
        <FunctionSignature name="uppercase" signature="string → string" />
        <FunctionSignature name="formatAge" signature="number → string" />
      </div>
      <div className="grid gap-2 sm:grid-cols-2">
        <Button variant="outline" onClick={() => onAnswer('uppercase')}>
          uppercase(parseAge("42"))
        </Button>
        <Button onClick={() => onAnswer('formatAge')}>formatAge(parseAge("42"))</Button>
      </div>
      <div className="rounded-lg border border-line bg-panel-soft p-4 font-mono text-sm text-ink">
        f : A → B&nbsp;&nbsp;&nbsp; g : B → C&nbsp;&nbsp;&nbsp; therefore g ∘ f : A → C
      </div>
    </div>
  )
}

function FunctionSignature({ name, signature }: { name: string; signature: string }) {
  return (
    <div className="rounded-lg border border-line bg-white/5 p-4 text-center">
      <code className="text-sm text-ink">{name}</code>
      <p className="mt-2 text-xs text-muted">{signature}</p>
    </div>
  )
}

function MlPipeline() {
  return (
    <div className="grid gap-5">
      <div>
        <h3 className="text-base font-semibold text-ink">Scale the pipeline up to ML</h3>
        <p className="mt-2 text-xs leading-relaxed text-muted">
          Keep the composition structure; replace the small arithmetic stages with model layers.
        </p>
      </div>
      <PipelineNodes
        nodes={['input', 'linear', 'activation', 'linear', 'activation', 'prediction']}
      />
      <div className="grid gap-3 rounded-lg border border-brand-blue/30 bg-brand-blue/10 p-4">
        <code className="text-sm text-ink">x → f₁ → f₂ → f₃ → f₄ → ŷ</code>
        <code className="text-sm text-ink">ŷ = f₄(f₃(f₂(f₁(x))))</code>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="rounded-lg border border-line bg-panel-soft p-4">
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Model composition
          </p>
          <p className="mt-3 text-xs leading-relaxed text-muted">
            linear → activation → linear → activation → prediction ŷ
          </p>
        </div>
        <div className="rounded-lg border border-brand-gold/30 bg-brand-gold/10 p-4">
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Prediction plus target
          </p>
          <p className="mt-3 text-xs leading-relaxed text-ink">ŷ + target y → loss</p>
          <code className="mt-2 block text-sm text-ink">L(fθ(x), y)</code>
        </div>
      </div>
      <p className="text-xs leading-relaxed text-muted">
        No matrices or gradients are needed yet to see the structure: a neural network is a
        composition of parameterized functions.
      </p>
    </div>
  )
}

function PipelineNodes({ nodes }: { nodes: string[] }) {
  return (
    <div className="flex flex-wrap items-center gap-2" aria-label={nodes.join(' then ')}>
      {nodes.map((node, index) => (
        <div key={`${node}-${index}`} className="contents">
          {index > 0 ? (
            <span className="text-brand-teal" aria-hidden="true">
              →
            </span>
          ) : null}
          <span className="rounded-lg border border-line bg-panel px-3 py-2 font-mono text-xs text-ink">
            {node}
          </span>
        </div>
      ))}
    </div>
  )
}

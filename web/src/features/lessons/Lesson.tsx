import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

export function Lesson() {
  const { lessonId = 'backpropagation' } = useParams()
  const { setPageContext, openTutorWithContext } = useTutor()
  const [selectedText, setSelectedText] = useState('')
  const [isCheckComplete, setIsCheckComplete] = useState(false)
  const isBackprop = lessonId === 'backpropagation'
  const title = getLessonTitle(lessonId)
  const objectiveIds = useMemo(
    () => (isBackprop ? ['nn.chain-rule', 'nn.backpropagation'] : ['nn.neuron']),
    [isBackprop],
  )

  useEffect(() => {
    setPageContext({
      type: 'lesson',
      title,
      lessonId,
      lessonTitle: title,
      moduleId: 'neural-networks',
      moduleTitle: 'Neural Networks From Scratch',
      objectiveIds,
    })
  }, [lessonId, objectiveIds, setPageContext, title])

  function handleSelection() {
    const selection = window.getSelection()?.toString().trim() ?? ''
    if (selection.length > 8) setSelectedText(selection)
  }

  const openSelectionTutor = () =>
    openTutorWithContext({
      type: 'lesson',
      title,
      lessonId,
      lessonTitle: title,
      moduleId: 'neural-networks',
      moduleTitle: 'Neural Networks From Scratch',
      selectedText,
      objectiveIds,
    })

  return (
    <div className="grid max-w-none gap-7 max-sm:gap-5">
      <div className="flex items-center gap-2.5 text-2xs text-faint max-sm:gap-2 max-sm:text-2xs">
        <Link
          className="text-muted no-underline hover:text-brand-teal"
          to="/curriculum/neural-networks"
        >
          Neural Networks From Scratch
        </Link>
        <span>/</span>
        <span>Lesson 04</span>
        <span className="ml-auto text-muted">4 / 7</span>
      </div>

      <div className="grid grid-cols-[minmax(0,760px)_220px] justify-center gap-16 max-xl:grid-cols-[minmax(0,680px)] max-sm:block">
        <article className="relative" onMouseUp={handleSelection}>
          <PageIntro compact title={title} detail="Gradients through a computational graph." />

          {selectedText ? (
            <SelectionPopover
              onAskTutor={openSelectionTutor}
              onDismiss={() => setSelectedText('')}
            />
          ) : null}

          <div className="mx-auto mt-12 max-w-2xl text-base leading-loose text-body max-sm:mt-9 max-sm:text-sm max-sm:leading-8">
            <p className="m-0 mb-7 text-xl leading-relaxed tracking-tight text-ink max-sm:text-lg">
              Backpropagation is the bookkeeping that lets a network turn an error at its output
              into useful updates for the parameters that produced it.
            </p>
            <p className="m-0 mb-7">
              Rather than treating a network as a black box, we can read it as a sequence of small
              functions. Each operation knows how its output changes when its inputs change. The
              chain rule lets us compose those local answers.
            </p>

            <Card className="my-8 grid grid-cols-[35px_1fr] gap-3 border-brand-teal/30 bg-brand-teal/10 p-5">
              <div className="grid size-8 place-items-center rounded-full bg-brand-teal font-serif text-lg font-semibold text-brand-ink">
                ∂
              </div>
              <div>
                <p className="text-2xs font-bold uppercase tracking-widest text-faint">
                  The useful question
                </p>
                <h3 className="my-2 text-base tracking-tight text-ink">
                  What did each operation contribute to the final error?
                </h3>
                <p className="m-0 text-xs leading-relaxed text-body">
                  Keep the forward pass and backward pass conceptually separate: first compute the
                  values, then carry local sensitivities back through the same graph.
                </p>
              </div>
            </Card>

            <div className="my-9 border-y border-line px-2.5 pb-5 pt-6 text-center">
              <span className="text-2xs uppercase tracking-widest text-brand-coral">
                chain rule
              </span>
              <div className="my-3 flex items-center justify-center gap-2 font-serif text-2xl text-ink max-sm:gap-1.5 max-sm:text-lg">
                <span>∂L</span>
                <i className="text-base font-normal text-faint">/</i>
                <span>∂w</span>
                <b className="px-1 font-normal text-brand-gold">=</b>
                <span>∂L</span>
                <i className="text-base font-normal text-faint">/</i>
                <span>∂y</span>
                <span className="text-brand-teal">·</span>
                <span>∂y</span>
                <i className="text-base font-normal text-faint">/</i>
                <span>∂w</span>
              </div>
              <p className="m-0 text-2xs text-faint">
                Each factor answers a local question. Multiplication composes them along the path.
              </p>
            </div>

            <p className="m-0 mb-7">
              For a single weight, this is enough to build an update. For a whole network, the same
              pattern repeats over a graph, reusing intermediate values rather than recalculating
              every possibility from scratch.
            </p>

            <GradientSketch />
            <KnowledgeCheck
              isComplete={isCheckComplete}
              onComplete={() => setIsCheckComplete(true)}
            />
            <LessonSources />
          </div>

          <div className="mx-auto mt-9 flex max-w-2xl items-center justify-between gap-3 border-t border-line pt-5 max-sm:gap-2">
            <Link
              className="inline-flex items-center justify-center gap-2.5 rounded-lg border border-line-strong bg-brand-slate/10 px-4 py-2.5 text-xs font-bold text-ink no-underline transition hover:bg-brand-slate/20"
              to="/lesson/computational-graphs"
            >
              ← Previous
            </Link>
            <span className="text-2xs text-faint max-sm:hidden">Lesson 4 of 7</span>
            <Link
              className="inline-flex items-center justify-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-xs font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
              to="/exercise/gradient-descent-exercise"
            >
              Next: Exercise →
            </Link>
          </div>
        </article>

        <aside className="max-xl:hidden">
          <Card className="sticky top-6 mt-44 p-4">
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">
              In this lesson
            </p>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-line py-2.5 text-brand-teal">
              <span className="text-2xs">01</span>
              <strong className="text-2xs font-medium leading-normal">
                Forward &amp; backward passes
              </strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-line py-2.5 text-ink">
              <span className="text-2xs text-brand-coral">02</span>
              <strong className="text-2xs font-medium leading-normal">Chain rule intuition</strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-line py-2.5 text-faint">
              <span className="text-2xs">03</span>
              <strong className="text-2xs font-medium leading-normal">One update by hand</strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-line py-2.5 text-faint">
              <span className="text-2xs">04</span>
              <strong className="text-2xs font-medium leading-normal">Knowledge check</strong>
            </div>
          </Card>
        </aside>
      </div>
    </div>
  )
}

function getLessonTitle(lessonId: string) {
  if (lessonId === 'backpropagation') return 'Backpropagation'
  if (lessonId === 'computational-graphs') return 'Computational graphs'
  if (lessonId === 'activation-functions') return 'Activation functions'
  return 'From linear models to neurons'
}

function SelectionPopover({
  onAskTutor,
  onDismiss,
}: {
  onAskTutor: () => void
  onDismiss: () => void
}) {
  return (
    <div className="absolute top-40 right-5 z-4 flex items-center gap-2 rounded-lg border border-line-strong bg-slate-800 px-2 py-2 text-2xs shadow-2xl max-sm:top-36 max-sm:left-2.5 max-sm:right-2.5">
      <span className="text-faint">Text selected</span>
      <button
        className="border-0 bg-transparent p-0 text-2xs font-bold text-brand-teal"
        type="button"
        onClick={onAskTutor}
      >
        Ask tutor about this ↗
      </button>
      <button
        className="border-0 bg-transparent p-0 text-2xs text-faint"
        type="button"
        onClick={onDismiss}
        aria-label="Dismiss selection"
      >
        ×
      </button>
    </div>
  )
}

function GradientSketch() {
  return (
    <div className="my-9 overflow-hidden rounded-xl border border-line bg-panel">
      <div className="flex items-start justify-between gap-3 px-5 pb-1 pt-5">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive sketch
          </p>
          <h3 className="mt-1.5 text-sm">Follow one gradient backward</h3>
        </div>
        <Badge tone="teal">Conceptual</Badge>
      </div>
      <div className="relative mx-2 my-1 h-56 bg-[linear-gradient(rgba(125,156,174,0.06)_1px,transparent_1px),linear-gradient(90deg,rgba(125,156,174,0.06)_1px,transparent_1px)] bg-[length:28px_28px] max-sm:mx-[-3px] max-sm:h-48 max-sm:scale-95">
        <div className="absolute top-[22px] left-[8%] text-2xs uppercase tracking-wide text-faint">
          forward pass
        </div>
        <div className="absolute bottom-[17px] left-[49%] text-2xs uppercase tracking-wide text-brand-gold">
          gradient signal
        </div>
        <div className="absolute top-[45%] left-[10%] z-2 grid size-14 place-items-center rounded-full bg-brand-teal font-serif text-lg font-bold text-brand-ink max-sm:left-[5%]">
          <span>x</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-2xs font-normal text-muted">
            input
          </small>
        </div>
        <div className="absolute top-[17%] left-[31%] z-2 grid size-14 place-items-center rounded-full bg-brand-gold font-serif text-lg font-bold text-brand-ink">
          <span>w</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-2xs font-normal text-muted">
            parameter
          </small>
        </div>
        <div className="absolute top-[45%] left-[48%] z-2 grid size-14 place-items-center rounded-full bg-brand-coral font-serif text-lg font-bold text-brand-ink">
          <span>Σ</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-2xs font-normal text-muted">
            weighted sum
          </small>
        </div>
        <div className="absolute top-[45%] right-[10%] z-2 grid size-14 place-items-center rounded-full bg-brand-violet font-serif text-lg font-bold text-brand-ink max-sm:right-[5%]">
          <span>L</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-2xs font-normal text-muted">
            loss
          </small>
        </div>
        <div className="absolute top-[49%] left-[24%] text-xl text-muted">→</div>
        <div className="absolute top-[49%] left-[42%] text-xl text-muted">→</div>
        <div className="absolute top-[49%] left-[68%] text-xl text-muted">→</div>
        <div className="absolute top-[27%] left-[25%] rotate-[-38deg] text-xl text-brand-gold">
          ←
        </div>
        <div className="absolute top-[27%] left-[45%] rotate-[38deg] text-xl text-brand-gold">
          ←
        </div>
        <div className="absolute top-[28%] left-[67%] rotate-[58deg] text-xl text-brand-gold">
          ←
        </div>
      </div>
      <div className="flex items-center gap-2 border-t border-line px-3 py-2.5 max-sm:flex-wrap">
        <button
          className="rounded border-0 bg-brand-slate/15 px-2 py-1.5 text-2xs text-ink"
          type="button"
        >
          Show values
        </button>
        <button
          className="rounded border-0 bg-transparent px-2 py-1.5 text-2xs text-faint"
          type="button"
        >
          Show derivatives
        </button>
        <span className="ml-auto text-2xs text-faint max-sm:basis-full max-sm:ml-0 max-sm:mt-0.5">
          drag the graph to explore <span className="ml-1 text-brand-teal">↗</span>
        </span>
      </div>
    </div>
  )
}

function KnowledgeCheck({
  isComplete,
  onComplete,
}: {
  isComplete: boolean
  onComplete: () => void
}) {
  const [answer, setAnswer] = useState<string | null>(null)
  const correct = answer?.startsWith('It sets') ?? false
  const answers = [
    'It sets the size of the parameter update.',
    'It makes the gradient point toward the loss.',
    'It removes the need for a loss function.',
  ]

  return (
    <div className="my-10 rounded-xl border border-brand-gold/30 bg-brand-gold/10 p-5">
      <div className="flex justify-between gap-4">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">Knowledge check</p>
          <h3 className="mt-2 text-base">Why is the learning rate multiplied by the gradient?</h3>
        </div>
        <span className="text-2xs text-faint">1 of 1</span>
      </div>
      <div className="my-5 grid gap-2">
        {answers.map((option) => (
          <button
            key={option}
            onClick={() => setAnswer(option)}
            type="button"
            className={`rounded-lg border px-2.5 py-2.5 text-left text-xs ${answer === option ? 'border-brand-gold/50 bg-brand-gold/10 text-ink' : 'border-line bg-white/5 text-muted hover:border-brand-gold/50 hover:bg-brand-gold/10 hover:text-ink'}`}
          >
            {answer === option ? '●' : '○'} {option}
          </button>
        ))}
      </div>
      {answer ? (
        <div
          className={`mb-4 rounded-md p-2.5 text-xs leading-normal ${correct ? 'bg-brand-teal/10 text-brand-teal' : 'bg-brand-coral/10 text-brand-coral'}`}
        >
          {correct
            ? 'Correct. The gradient gives a direction; the learning rate chooses the step size.'
            : 'Not quite. Try connecting the learning rate to how far the optimizer moves.'}
        </div>
      ) : null}
      <Button variant="secondary" disabled={!answer} onClick={onComplete}>
        {isComplete ? 'Check complete ✓' : 'Mark check complete'}
      </Button>
    </div>
  )
}

function LessonSources() {
  return (
    <div className="mt-11 border-t border-line pt-7">
      <SectionHeading title="Sources" />
      <div className="grid grid-cols-[27px_1fr_20px] gap-2.5 border-t border-line py-3">
        <span className="text-2xs text-faint">[01]</span>
        <div>
          <strong className="block text-xs">Deep Learning</strong>
          <p className="mt-1 text-2xs text-faint">
            Ian Goodfellow, Yoshua Bengio, Aaron Courville · MIT Press
          </p>
        </div>
        <span className="text-right text-sm text-brand-teal">↗</span>
      </div>
      <div className="grid grid-cols-[27px_1fr_20px] gap-2.5 border-t border-line py-3">
        <span className="text-2xs text-faint">[02]</span>
        <div>
          <strong className="block text-xs">Automatic differentiation in machine learning</strong>
          <p className="mt-1 text-2xs text-faint">
            Baydin et al. · Journal of Machine Learning Research
          </p>
        </div>
        <span className="text-right text-sm text-brand-teal">↗</span>
      </div>
    </div>
  )
}

import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

export function Lesson() {
  const navigate = useNavigate()
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
    <div className="grid max-w-none gap-[29px] max-[640px]:gap-[21px]">
      <div className="flex items-center gap-2.5 text-[10px] text-[var(--faint)] max-[640px]:gap-[7px] max-[640px]:text-[9px]">
        <button
          className="border-0 bg-transparent p-0 text-[var(--muted)] hover:text-[var(--teal)]"
          onClick={() => navigate('/curriculum/neural-networks')}
          type="button"
        >
          Neural Networks From Scratch
        </button>
        <span>/</span>
        <span>Lesson 04</span>
        <span className="ml-auto text-[var(--muted)]">4 / 7</span>
      </div>

      <div className="grid grid-cols-[minmax(0,760px)_220px] justify-center gap-16 max-[1120px]:grid-cols-[minmax(0,680px)] max-[640px]:block">
        <article className="relative" onMouseUp={handleSelection}>
          <PageIntro compact title={title} detail="Gradients through a computational graph." />

          {selectedText ? (
            <SelectionPopover
              onAskTutor={openSelectionTutor}
              onDismiss={() => setSelectedText('')}
            />
          ) : null}

          <div className="mx-auto mt-[50px] max-w-[630px] text-[15px] leading-[1.9] text-[var(--body)] max-[640px]:mt-[34px] max-[640px]:text-sm max-[640px]:leading-[1.8]">
            <p className="m-0 mb-7 text-[19px] leading-[1.6] tracking-[-0.015em] text-[var(--ink)] max-[640px]:text-[17px]">
              Backpropagation is the bookkeeping that lets a network turn an error at its output
              into useful updates for the parameters that produced it.
            </p>
            <p className="m-0 mb-7">
              Rather than treating a network as a black box, we can read it as a sequence of small
              functions. Each operation knows how its output changes when its inputs change. The
              chain rule lets us compose those local answers.
            </p>

            <Card className="my-8 grid grid-cols-[35px_1fr] gap-[13px] border-[rgba(118,208,192,0.24)] bg-[rgba(118,208,192,0.065)] p-[19px_20px]">
              <div className="grid size-8 place-items-center rounded-full bg-[var(--teal)] font-serif text-lg font-semibold text-[#0b171e]">
                ∂
              </div>
              <div>
                <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
                  The useful question
                </p>
                <h3 className="my-[5px_8px] text-[15px] tracking-[-0.02em] text-[var(--ink)]">
                  What did each operation contribute to the final error?
                </h3>
                <p className="m-0 text-xs leading-[1.6] text-[var(--body)]">
                  Keep the forward pass and backward pass conceptually separate: first compute the
                  values, then carry local sensitivities back through the same graph.
                </p>
              </div>
            </Card>

            <div className="my-[35px] border-y border-[var(--line)] px-2.5 pb-[19px] pt-[25px] text-center">
              <span className="text-[9px] uppercase tracking-[0.16em] text-[var(--coral)]">
                chain rule
              </span>
              <div className="my-[13px_9px] flex items-center justify-center gap-[9px] font-serif text-[22px] text-[var(--ink)] max-[640px]:gap-1.5 max-[640px]:text-lg">
                <span>∂L</span>
                <i className="text-base font-normal text-[var(--faint)]">/</i>
                <span>∂w</span>
                <b className="px-[5px] font-normal text-[var(--gold)]">=</b>
                <span>∂L</span>
                <i className="text-base font-normal text-[var(--faint)]">/</i>
                <span>∂y</span>
                <span className="text-[var(--teal)]">·</span>
                <span>∂y</span>
                <i className="text-base font-normal text-[var(--faint)]">/</i>
                <span>∂w</span>
              </div>
              <p className="m-0 text-[10px] text-[var(--faint)]">
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

          <div className="mx-auto mt-[37px] flex max-w-[630px] items-center justify-between gap-3 border-t border-[var(--line)] pt-5 max-[640px]:gap-[9px]">
            <Button variant="secondary" onClick={() => navigate('/lesson/computational-graphs')}>
              ← Previous
            </Button>
            <span className="text-[10px] text-[var(--faint)] max-[640px]:hidden">
              Lesson 4 of 7
            </span>
            <Button onClick={() => navigate('/exercise/gradient-descent-exercise')}>
              Next: Exercise →
            </Button>
          </div>
        </article>

        <aside className="pt-[175px] max-[1120px]:hidden">
          <Card className="p-[17px]">
            <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
              In this lesson
            </p>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-[var(--line)] py-2.5 text-[var(--teal)]">
              <span className="text-[9px]">01</span>
              <strong className="text-[10px] font-medium leading-[1.3]">
                Forward &amp; backward passes
              </strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-[var(--line)] py-2.5 text-[var(--ink)]">
              <span className="text-[9px] text-[var(--coral)]">02</span>
              <strong className="text-[10px] font-medium leading-[1.3]">
                Chain rule intuition
              </strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-[var(--line)] py-2.5 text-[var(--faint)]">
              <span className="text-[9px]">03</span>
              <strong className="text-[10px] font-medium leading-[1.3]">One update by hand</strong>
            </div>
            <div className="grid grid-cols-[23px_1fr] gap-1.5 border-t border-[var(--line)] py-2.5 text-[var(--faint)]">
              <span className="text-[9px]">04</span>
              <strong className="text-[10px] font-medium leading-[1.3]">Knowledge check</strong>
            </div>
          </Card>
          <Card className="mt-4 p-[18px]">
            <span className="text-[var(--gold)]">✦</span>
            <h3 className="my-[13px_7px] text-[13px]">Tutor</h3>
            <button
              className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]"
              onClick={() =>
                openTutorWithContext({
                  type: 'lesson',
                  title,
                  lessonId,
                  lessonTitle: title,
                  moduleId: 'neural-networks',
                  moduleTitle: 'Neural Networks From Scratch',
                  objectiveIds,
                })
              }
            >
              Open tutor <span>→</span>
            </button>
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
    <div className="absolute top-[155px] right-5 z-[4] flex items-center gap-2 rounded-lg border border-[var(--line-strong)] bg-[#142334] px-[9px] py-[7px] text-[10px] shadow-[0_12px_26px_rgba(0,0,0,0.25)] max-[640px]:top-[143px] max-[640px]:left-2.5 max-[640px]:right-2.5">
      <span className="text-[var(--faint)]">Text selected</span>
      <button
        className="border-0 bg-transparent p-0 text-[10px] font-bold text-[var(--teal)]"
        type="button"
        onClick={onAskTutor}
      >
        Ask tutor about this ↗
      </button>
      <button
        className="border-0 bg-transparent p-0 text-[10px] text-[var(--faint)]"
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
    <div className="my-[34px] overflow-hidden rounded-xl border border-[var(--line)] bg-[var(--panel)]">
      <div className="flex items-start justify-between gap-3 px-[19px] pb-1 pt-[18px]">
        <div>
          <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
            Interactive sketch
          </p>
          <h3 className="mt-1.5 text-sm">Follow one gradient backward</h3>
        </div>
        <Badge tone="teal">Conceptual</Badge>
      </div>
      <div className="relative mx-[10px] my-[3px] h-[225px] bg-[linear-gradient(rgba(125,156,174,0.06)_1px,transparent_1px),linear-gradient(90deg,rgba(125,156,174,0.06)_1px,transparent_1px)] bg-[length:28px_28px] max-[640px]:mx-[-3px] max-[640px]:h-[195px] max-[640px]:scale-[0.95]">
        <div className="absolute top-[22px] left-[8%] text-[9px] uppercase tracking-[0.1em] text-[var(--faint)]">
          forward pass
        </div>
        <div className="absolute bottom-[17px] left-[49%] text-[9px] uppercase tracking-[0.1em] text-[var(--gold)]">
          gradient signal
        </div>
        <div className="absolute top-[45%] left-[10%] z-[2] grid size-[55px] place-items-center rounded-full bg-[var(--teal)] font-serif text-lg font-bold text-[#0c1721] max-[640px]:left-[5%]">
          <span>x</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-[10px] font-normal text-[var(--muted)]">
            input
          </small>
        </div>
        <div className="absolute top-[17%] left-[31%] z-[2] grid size-[55px] place-items-center rounded-full bg-[var(--gold)] font-serif text-lg font-bold text-[#0c1721]">
          <span>w</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-[10px] font-normal text-[var(--muted)]">
            parameter
          </small>
        </div>
        <div className="absolute top-[45%] left-[48%] z-[2] grid size-[55px] place-items-center rounded-full bg-[var(--coral)] font-serif text-lg font-bold text-[#0c1721]">
          <span>Σ</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-[10px] font-normal text-[var(--muted)]">
            weighted sum
          </small>
        </div>
        <div className="absolute top-[45%] right-[10%] z-[2] grid size-[55px] place-items-center rounded-full bg-[var(--violet)] font-serif text-lg font-bold text-[#0c1721] max-[640px]:right-[5%]">
          <span>L</span>
          <small className="absolute top-[61px] whitespace-nowrap font-sans text-[10px] font-normal text-[var(--muted)]">
            loss
          </small>
        </div>
        <div className="absolute top-[49%] left-[24%] text-xl text-[var(--muted)]">→</div>
        <div className="absolute top-[49%] left-[42%] text-xl text-[var(--muted)]">→</div>
        <div className="absolute top-[49%] left-[68%] text-xl text-[var(--muted)]">→</div>
        <div className="absolute top-[27%] left-[25%] rotate-[-38deg] text-xl text-[var(--gold)]">
          ←
        </div>
        <div className="absolute top-[27%] left-[45%] rotate-[38deg] text-xl text-[var(--gold)]">
          ←
        </div>
        <div className="absolute top-[28%] left-[67%] rotate-[58deg] text-xl text-[var(--gold)]">
          ←
        </div>
      </div>
      <div className="flex items-center gap-[9px] border-t border-[var(--line)] px-[13px] py-2.5 max-[640px]:flex-wrap">
        <button
          className="rounded-[5px] border-0 bg-[rgba(157,185,194,0.13)] px-2 py-1.5 text-[9px] text-[var(--ink)]"
          type="button"
        >
          Show values
        </button>
        <button
          className="rounded-[5px] border-0 bg-transparent px-2 py-1.5 text-[9px] text-[var(--faint)]"
          type="button"
        >
          Show derivatives
        </button>
        <span className="ml-auto text-[9px] text-[var(--faint)] max-[640px]:basis-full max-[640px]:ml-0 max-[640px]:mt-0.5">
          drag the graph to explore <span className="ml-[3px] text-[var(--teal)]">↗</span>
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
    <div className="my-[38px] rounded-xl border border-[rgba(225,184,106,0.3)] bg-[rgba(225,184,106,0.055)] p-5">
      <div className="flex justify-between gap-[15px]">
        <div>
          <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
            Knowledge check
          </p>
          <h3 className="mt-[7px] text-[15px]">
            Why is the learning rate multiplied by the gradient?
          </h3>
        </div>
        <span className="text-[10px] text-[var(--faint)]">1 of 1</span>
      </div>
      <div className="my-5 grid gap-[7px]">
        {answers.map((option) => (
          <button
            key={option}
            onClick={() => setAnswer(option)}
            type="button"
            className={`rounded-[7px] border px-2.5 py-2.5 text-left text-[11px] ${answer === option ? 'border-[rgba(225,184,106,0.45)] bg-[rgba(225,184,106,0.09)] text-[var(--ink)]' : 'border-[var(--line)] bg-white/[0.025] text-[var(--muted)] hover:border-[rgba(225,184,106,0.45)] hover:bg-[rgba(225,184,106,0.09)] hover:text-[var(--ink)]'}`}
          >
            {answer === option ? '●' : '○'} {option}
          </button>
        ))}
      </div>
      {answer ? (
        <div
          className={`mb-[15px] rounded-md p-2.5 text-[11px] leading-[1.5] ${correct ? 'bg-[rgba(118,208,192,0.1)] text-[var(--teal)]' : 'bg-[rgba(239,145,110,0.1)] text-[var(--coral)]'}`}
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
    <div className="mt-[42px] border-t border-[var(--line)] pt-[26px]">
      <SectionHeading title="Sources" />
      <div className="grid grid-cols-[27px_1fr_20px] gap-2.5 border-t border-[var(--line)] py-3">
        <span className="text-[10px] text-[var(--faint)]">[01]</span>
        <div>
          <strong className="block text-[11px]">Deep Learning</strong>
          <p className="mt-1 text-[10px] text-[var(--faint)]">
            Ian Goodfellow, Yoshua Bengio, Aaron Courville · MIT Press
          </p>
        </div>
        <span className="text-right text-[13px] text-[var(--teal)]">↗</span>
      </div>
      <div className="grid grid-cols-[27px_1fr_20px] gap-2.5 border-t border-[var(--line)] py-3">
        <span className="text-[10px] text-[var(--faint)]">[02]</span>
        <div>
          <strong className="block text-[11px]">
            Automatic differentiation in machine learning
          </strong>
          <p className="mt-1 text-[10px] text-[var(--faint)]">
            Baydin et al. · Journal of Machine Learning Research
          </p>
        </div>
        <span className="text-right text-[13px] text-[var(--teal)]">↗</span>
      </div>
    </div>
  )
}

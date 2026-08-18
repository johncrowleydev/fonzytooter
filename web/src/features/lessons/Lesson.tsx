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
    <div className="lesson-page page-stack">
      <div className="lesson-breadcrumb">
        <button onClick={() => navigate('/curriculum/neural-networks')} type="button">
          Neural Networks From Scratch
        </button>
        <span>/</span>
        <span>Lesson 04</span>
        <span className="lesson-progress-label">4 / 7</span>
      </div>

      <div className="lesson-layout">
        <article className="lesson-reader" onMouseUp={handleSelection}>
          <PageIntro compact title={title} detail="Gradients through a computational graph." />

          {selectedText ? (
            <SelectionPopover
              onAskTutor={openSelectionTutor}
              onDismiss={() => setSelectedText('')}
            />
          ) : null}

          <div className="lesson-prose">
            <p className="lead-prose">
              Backpropagation is the bookkeeping that lets a network turn an error at its output
              into useful updates for the parameters that produced it.
            </p>
            <p>
              Rather than treating a network as a black box, we can read it as a sequence of small
              functions. Each operation knows how its output changes when its inputs change. The
              chain rule lets us compose those local answers.
            </p>

            <Card className="concept-callout">
              <div className="callout-icon">∂</div>
              <div>
                <p className="eyebrow">The useful question</p>
                <h3>What did each operation contribute to the final error?</h3>
                <p>
                  Keep the forward pass and backward pass conceptually separate: first compute the
                  values, then carry local sensitivities back through the same graph.
                </p>
              </div>
            </Card>

            <div className="equation-block">
              <span className="equation-label">chain rule</span>
              <div className="equation">
                <span>∂L</span>
                <i>/</i>
                <span>∂w</span>
                <b>=</b>
                <span>∂L</span>
                <i>/</i>
                <span>∂y</span>
                <span className="dot">·</span>
                <span>∂y</span>
                <i>/</i>
                <span>∂w</span>
              </div>
              <p>
                Each factor answers a local question. Multiplication composes them along the path.
              </p>
            </div>

            <p>
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

          <div className="lesson-nav">
            <Button variant="secondary" onClick={() => navigate('/lesson/computational-graphs')}>
              ← Previous
            </Button>
            <span>Lesson 4 of 7</span>
            <Button onClick={() => navigate('/exercise/gradient-descent-exercise')}>
              Next: Exercise →
            </Button>
          </div>
        </article>

        <aside className="lesson-aside">
          <Card className="lesson-outline">
            <p className="eyebrow">In this lesson</p>
            <div className="outline-item done">
              <span>01</span>
              <strong>Forward &amp; backward passes</strong>
            </div>
            <div className="outline-item active">
              <span>02</span>
              <strong>Chain rule intuition</strong>
            </div>
            <div className="outline-item">
              <span>03</span>
              <strong>One update by hand</strong>
            </div>
            <div className="outline-item">
              <span>04</span>
              <strong>Knowledge check</strong>
            </div>
          </Card>
          <Card className="lesson-tutor-card">
            <span className="tutor-spark">✦</span>
            <h3>Tutor</h3>
            <button
              className="text-link"
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
    <div className="selection-popover">
      <span>Text selected</span>
      <button type="button" onClick={onAskTutor}>
        Ask tutor about this ↗
      </button>
      <button type="button" onClick={onDismiss} aria-label="Dismiss selection">
        ×
      </button>
    </div>
  )
}

function GradientSketch() {
  return (
    <div className="visualization-card">
      <div className="visualization-header">
        <div>
          <p className="eyebrow">Interactive sketch</p>
          <h3>Follow one gradient backward</h3>
        </div>
        <Badge tone="teal">Conceptual</Badge>
      </div>
      <div className="graph-visual">
        <div className="graph-label label-forward">forward pass</div>
        <div className="graph-label label-backward">gradient signal</div>
        <div className="graph-node g-input">
          <span>x</span>
          <small>input</small>
        </div>
        <div className="graph-node g-weight">
          <span>w</span>
          <small>parameter</small>
        </div>
        <div className="graph-node g-sum">
          <span>Σ</span>
          <small>weighted sum</small>
        </div>
        <div className="graph-node g-loss">
          <span>L</span>
          <small>loss</small>
        </div>
        <div className="graph-arrow a1">→</div>
        <div className="graph-arrow a2">→</div>
        <div className="graph-arrow a3">→</div>
        <div className="graph-arrow back1">←</div>
        <div className="graph-arrow back2">←</div>
        <div className="graph-arrow back3">←</div>
      </div>
      <div className="visualization-controls">
        <button type="button" className="active">
          Show values
        </button>
        <button type="button">Show derivatives</button>
        <span>
          drag the graph to explore <span>↗</span>
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
    <div className="knowledge-check">
      <div className="check-header">
        <div>
          <p className="eyebrow">Knowledge check</p>
          <h3>Why is the learning rate multiplied by the gradient?</h3>
        </div>
        <span className="check-count">1 of 1</span>
      </div>
      <div className="check-options">
        {answers.map((option) => (
          <button
            key={option}
            onClick={() => setAnswer(option)}
            type="button"
            className={answer === option ? 'check-option selected' : 'check-option'}
          >
            {answer === option ? '●' : '○'} {option}
          </button>
        ))}
      </div>
      {answer ? (
        <div className={`check-feedback ${correct ? 'correct' : 'incorrect'}`}>
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
    <div className="lesson-sources">
      <SectionHeading title="Sources" />
      <div className="source-item">
        <span>[01]</span>
        <div>
          <strong>Deep Learning</strong>
          <p>Ian Goodfellow, Yoshua Bengio, Aaron Courville · MIT Press</p>
        </div>
        <span className="source-arrow">↗</span>
      </div>
      <div className="source-item">
        <span>[02]</span>
        <div>
          <strong>Automatic differentiation in machine learning</strong>
          <p>Baydin et al. · Journal of Machine Learning Research</p>
        </div>
        <span className="source-arrow">↗</span>
      </div>
    </div>
  )
}

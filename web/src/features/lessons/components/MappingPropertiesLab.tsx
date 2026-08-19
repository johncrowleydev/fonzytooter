import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import {
  analyzeMapping,
  MappingDiagram,
  setFunctionOutput,
  type MappingEdge,
} from './MappingDiagram'

const fourDomain = ['1', '2', '3', '4'] as const
const fourCodomain = ['A', 'B', 'C', 'D'] as const
const threeCodomain = ['A', 'B', 'C'] as const
const collisionMapping: MappingEdge[] = [
  { from: '1', to: 'A' },
  { from: '2', to: 'B' },
  { from: '3', to: 'B' },
  { from: '4', to: 'D' },
]

type MappingCategory = 'injective-only' | 'surjective-only' | 'bijective' | 'neither'

const classificationMappings: {
  id: string
  label: string
  domain: readonly string[]
  codomain: readonly string[]
  edges: MappingEdge[]
  category: MappingCategory
}[] = [
  {
    id: 'alpha',
    label: 'Mapping α',
    domain: ['1', '2', '3'],
    codomain: ['A', 'B', 'C', 'D'],
    edges: [
      { from: '1', to: 'A' },
      { from: '2', to: 'B' },
      { from: '3', to: 'C' },
    ],
    category: 'injective-only',
  },
  {
    id: 'beta',
    label: 'Mapping β',
    domain: fourDomain,
    codomain: threeCodomain,
    edges: [
      { from: '1', to: 'A' },
      { from: '2', to: 'B' },
      { from: '3', to: 'B' },
      { from: '4', to: 'C' },
    ],
    category: 'surjective-only',
  },
  {
    id: 'gamma',
    label: 'Mapping γ',
    domain: ['1', '2', '3'],
    codomain: ['A', 'B', 'C'],
    edges: [
      { from: '1', to: 'B' },
      { from: '2', to: 'C' },
      { from: '3', to: 'A' },
    ],
    category: 'bijective',
  },
  {
    id: 'delta',
    label: 'Mapping δ',
    domain: fourDomain,
    codomain: fourCodomain,
    edges: [
      { from: '1', to: 'A' },
      { from: '2', to: 'B' },
      { from: '3', to: 'B' },
      { from: '4', to: 'A' },
    ],
    category: 'neither',
  },
]

const categoryLabels: Record<MappingCategory, string> = {
  'injective-only': 'Injective only',
  'surjective-only': 'Surjective only',
  bijective: 'Bijective',
  neither: 'Neither',
}

type Feedback = { tone: 'success' | 'error'; message: string } | null
const feedbackClasses = {
  success: 'border-brand-teal/30 bg-brand-teal/10',
  error: 'border-brand-coral/30 bg-brand-coral/10',
} as const
const evidenceButtonClass =
  'rounded-lg border px-3 py-2 text-left text-xs font-semibold transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-teal disabled:cursor-not-allowed disabled:opacity-50'

export function MappingPropertiesLab() {
  const [stage, setStage] = useState(0)
  const [edges, setEdges] = useState<MappingEdge[]>(collisionMapping)
  const [selectedInput, setSelectedInput] = useState('1')
  const [feedback, setFeedback] = useState<Feedback>(null)
  const [classifications, setClassifications] = useState<Record<string, string>>({})
  const [diagnosticAnswer, setDiagnosticAnswer] = useState<'yes' | 'no' | null>(null)
  const [collisionPair, setCollisionPair] = useState<string | null>(null)
  const [diagnosticTarget, setDiagnosticTarget] = useState<string | null>(null)
  const stageCount = 6
  const complete = stage === stageCount
  const codomain = stage === 3 ? threeCodomain : fourCodomain
  const analysis = analyzeMapping(fourDomain, codomain, edges)

  function checkDiagnostic() {
    if (stage === 0) {
      if (diagnosticAnswer === null) {
        setFeedback({
          tone: 'error',
          message: 'First decide whether the function is injective.',
        })
      } else if (diagnosticAnswer === 'yes') {
        setFeedback({
          tone: 'error',
          message:
            'The mapping is a valid function, but it contains a collision. Choose “No,” then locate the inputs and output that form it.',
        })
      } else if (!collisionPair || !diagnosticTarget) {
        setFeedback({
          tone: 'error',
          message: 'Now identify both the pair of inputs that collide and the output they reach.',
        })
      } else if (collisionPair !== '2-and-3' || diagnosticTarget !== 'B') {
        setFeedback({
          tone: 'error',
          message:
            'That evidence does not match the arrows. Trace each selected input to its destination and look for one output reached twice.',
        })
      } else {
        setFeedback({
          tone: 'success',
          message:
            'This is a valid function, but it is not injective: distinct inputs 2 and 3 collide at B. “Not injective” does not mean “not a function.”',
        })
      }
      return
    }

    if (diagnosticAnswer === null) {
      setFeedback({
        tone: 'error',
        message: 'First decide whether the function is surjective.',
      })
    } else if (diagnosticAnswer === 'yes') {
      setFeedback({
        tone: 'error',
        message:
          'The image does not fill the codomain. Choose “No,” then locate the codomain value with no incoming arrow.',
      })
    } else if (!diagnosticTarget) {
      setFeedback({
        tone: 'error',
        message: 'Now identify the codomain value that the mapping misses.',
      })
    } else if (diagnosticTarget !== 'C') {
      setFeedback({
        tone: 'error',
        message: `${diagnosticTarget} has at least one incoming arrow, so it belongs to the image. Look for the codomain value with none.`,
      })
    } else {
      setFeedback({
        tone: 'success',
        message:
          'C belongs to the codomain but not the image. Because image ≠ codomain, the function is not surjective.',
      })
    }
  }

  function selectDiagnosticAnswer(answer: 'yes' | 'no') {
    setDiagnosticAnswer(answer)
    if (answer === 'yes') {
      setCollisionPair(null)
      setDiagnosticTarget(null)
    }
    setFeedback(null)
  }

  function chooseOutput(input: string, output: string) {
    setEdges((current) => setFunctionOutput(current, input, output))
    setFeedback(null)
  }

  function checkEditableMapping() {
    if (stage === 2) {
      if (!analysis.isFunction) {
        setFeedback({
          tone: 'error',
          message: 'Keep exactly one output for every input before checking injectivity.',
        })
      } else if (!analysis.isInjective) {
        const collision = codomain.find((output) => analysis.incomingCounts[output] > 1)
        const inputs = edges.filter((edge) => edge.to === collision).map((edge) => edge.from)
        setFeedback({
          tone: 'error',
          message: `Inputs ${inputs.join(' and ')} still collide at ${collision}. An injective function has no collisions.`,
        })
      } else {
        setFeedback({
          tone: 'success',
          message:
            'Every input has one output and every output receives at most one input. This function is injective.',
        })
      }
      return
    }

    if (!analysis.isSurjective) {
      const gaps = codomain.filter((output) => analysis.incomingCounts[output] === 0)
      setFeedback({
        tone: 'error',
        message: `${gaps.join(' and ')} ${gaps.length === 1 ? 'is a gap' : 'are gaps'}. A surjection must reach every codomain value.`,
      })
      return
    }

    const shared = codomain.find((output) => analysis.incomingCounts[output] > 1)
    setFeedback({
      tone: 'success',
      message: `Every codomain value is reached, so this function is surjective. ${shared} receives multiple inputs, so it is not injective. The two properties are independent.`,
    })
  }

  function checkClassifications() {
    const incorrect = classificationMappings.filter(
      (mapping) => classifications[mapping.id] !== mapping.category,
    )
    setFeedback(
      incorrect.length === 0
        ? {
            tone: 'success',
            message:
              'All four categories are correct. Injectivity tracks collisions; surjectivity tracks gaps.',
          }
        : {
            tone: 'error',
            message: `${incorrect.length} classification${incorrect.length === 1 ? '' : 's'} still need attention. Check each diagram for both collisions and gaps.`,
          },
    )
  }

  function answerReversibility(answer: MappingCategory) {
    setFeedback(
      answer === 'bijective'
        ? {
            tone: 'success',
            message:
              'A bijection can be perfectly reversed. Collisions would create ambiguity backward, and gaps would leave codomain values with no predecessor.',
          }
        : {
            tone: 'error',
            message:
              'Only the bijective mapping has neither collisions nor gaps, so only it reverses across the full declared sets.',
          },
    )
  }

  function advance() {
    const next = stage + 1
    if (next === 2) setEdges(collisionMapping)
    if (next === 3) {
      setEdges([
        { from: '1', to: 'A' },
        { from: '2', to: 'A' },
        { from: '3', to: 'A' },
        { from: '4', to: 'A' },
      ])
    }
    setStage(next)
    setSelectedInput('1')
    setDiagnosticAnswer(null)
    setCollisionPair(null)
    setDiagnosticTarget(null)
    setFeedback(null)
  }

  function resetCurrentStage() {
    if (stage <= 2) setEdges(collisionMapping)
    if (stage === 3) {
      setEdges([
        { from: '1', to: 'A' },
        { from: '2', to: 'A' },
        { from: '3', to: 'A' },
        { from: '4', to: 'A' },
      ])
    }
    if (stage === 4) setClassifications({})
    if (stage <= 1) {
      setDiagnosticAnswer(null)
      setCollisionPair(null)
      setDiagnosticTarget(null)
    }
    setFeedback(null)
  }

  return (
    <Card className="my-8 grid min-w-0 gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">
            Collisions, gaps, and reversibility
          </h2>
          <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">
            Read injective and surjective as independent structural properties of a function.
          </p>
        </div>
        <Badge tone={complete ? 'teal' : 'violet'}>
          {complete ? 'Complete' : `${stage + 1} / ${stageCount}`}
        </Badge>
      </div>

      {complete ? (
        <div className="rounded-lg border border-brand-teal/30 bg-brand-teal/10 p-5">
          <p className="text-2xs font-bold uppercase tracking-widest text-brand-teal">
            Transition to inverses
          </p>
          <p className="mt-2 text-sm leading-relaxed text-ink">
            A bijection is reversible because it has neither backward ambiguity nor missing
            predecessors.
          </p>
        </div>
      ) : (
        <>
          <StageContent
            stage={stage}
            edges={edges}
            selectedInput={selectedInput}
            classifications={classifications}
            onSelectInput={setSelectedInput}
            onChooseOutput={chooseOutput}
            onClassify={(id, category) => {
              setClassifications((current) => ({ ...current, [id]: category }))
              setFeedback(null)
            }}
          />

          {stage <= 1 ? (
            <DiagnosticControls
              stage={stage}
              answer={diagnosticAnswer}
              collisionPair={collisionPair}
              target={diagnosticTarget}
              onAnswer={selectDiagnosticAnswer}
              onCollisionPair={(pair) => {
                setCollisionPair(pair)
                setFeedback(null)
              }}
              onTarget={(target) => {
                setDiagnosticTarget(target)
                setFeedback(null)
              }}
            />
          ) : null}
          {stage <= 1 ? (
            <Button className="w-max" onClick={checkDiagnostic}>
              Check evidence
            </Button>
          ) : null}
          {stage === 2 || stage === 3 ? (
            <Button className="w-max" onClick={checkEditableMapping}>
              Check mapping
            </Button>
          ) : null}
          {stage === 4 ? (
            <Button className="w-max" onClick={checkClassifications}>
              Check classifications
            </Button>
          ) : null}
          {stage === 5 ? (
            <div className="grid gap-2 sm:grid-cols-2" role="group" aria-label="Reversible mapping">
              {(Object.keys(categoryLabels) as MappingCategory[]).map((category) => (
                <Button
                  key={category}
                  variant="outline"
                  onClick={() => answerReversibility(category)}
                >
                  {categoryLabels[category]}
                </Button>
              ))}
            </div>
          ) : null}

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
            <Button variant="secondary" onClick={resetCurrentStage}>
              Reset stage
            </Button>
            {feedback?.tone === 'success' ? <Button onClick={advance}>Continue</Button> : null}
          </div>
        </>
      )}
    </Card>
  )
}

function DiagnosticControls({
  stage,
  answer,
  collisionPair,
  target,
  onAnswer,
  onCollisionPair,
  onTarget,
}: {
  stage: number
  answer: 'yes' | 'no' | null
  collisionPair: string | null
  target: string | null
  onAnswer: (answer: 'yes' | 'no') => void
  onCollisionPair: (pair: string) => void
  onTarget: (target: string) => void
}) {
  const pairOptions = [
    { value: '1-and-4', label: 'Inputs 1 and 4' },
    { value: '2-and-3', label: 'Inputs 2 and 3' },
    { value: '3-and-4', label: 'Inputs 3 and 4' },
  ]

  return (
    <div className="grid gap-4 rounded-lg border border-line bg-white/5 p-4">
      <fieldset>
        <legend className="text-xs font-semibold text-ink">
          {stage === 0 ? 'Is this function injective?' : 'Is this function surjective?'}
        </legend>
        <div className="mt-3 flex flex-wrap gap-2">
          {(['yes', 'no'] as const).map((option) => (
            <button
              key={option}
              className={`${evidenceButtonClass} ${evidenceSelectionClass(answer === option)}`}
              type="button"
              onClick={() => onAnswer(option)}
              aria-pressed={answer === option}
            >
              {answer === option ? '✓ Selected: ' : ''}
              {option === 'yes' ? 'Yes' : 'No'}
            </button>
          ))}
        </div>
      </fieldset>

      {stage === 0 ? (
        <fieldset disabled={answer !== 'no'}>
          <legend className="text-xs font-semibold text-ink">Which inputs collide?</legend>
          <div className="mt-3 flex flex-wrap gap-2">
            {pairOptions.map((option) => (
              <button
                key={option.value}
                className={`${evidenceButtonClass} ${evidenceSelectionClass(
                  collisionPair === option.value,
                )}`}
                type="button"
                onClick={() => onCollisionPair(option.value)}
                aria-pressed={collisionPair === option.value}
              >
                {collisionPair === option.value ? '✓ Selected: ' : ''}
                {option.label}
              </button>
            ))}
          </div>
        </fieldset>
      ) : null}

      <fieldset disabled={answer !== 'no'}>
        <legend className="text-xs font-semibold text-ink">
          {stage === 0 ? 'Which output receives the collision?' : 'Which codomain value is missed?'}
        </legend>
        <div className="mt-3 flex flex-wrap gap-2">
          {fourCodomain.map((output) => (
            <button
              key={output}
              className={`${evidenceButtonClass} ${evidenceSelectionClass(target === output)}`}
              type="button"
              onClick={() => onTarget(output)}
              aria-pressed={target === output}
            >
              {target === output ? '✓ Selected: ' : ''}
              Output {output}
            </button>
          ))}
        </div>
      </fieldset>
    </div>
  )
}

function evidenceSelectionClass(selected: boolean) {
  return selected
    ? 'border-brand-violet/60 bg-brand-violet/10 text-ink'
    : 'border-line bg-panel text-muted hover:border-brand-violet/50 hover:text-ink'
}

function StageContent({
  stage,
  edges,
  selectedInput,
  classifications,
  onSelectInput,
  onChooseOutput,
  onClassify,
}: {
  stage: number
  edges: MappingEdge[]
  selectedInput: string
  classifications: Record<string, string>
  onSelectInput: (input: string) => void
  onChooseOutput: (input: string, output: string) => void
  onClassify: (id: string, category: string) => void
}) {
  if (stage <= 1) {
    return (
      <div className="grid gap-4">
        <div>
          <h3 className="text-base font-semibold text-ink">
            {stage === 0 ? 'Find the collision' : 'Find the gap'}
          </h3>
          <p className="mt-2 text-xs leading-relaxed text-muted">
            {stage === 0
              ? 'Is this function injective? Choose the answer supported by the diagram.'
              : 'Is this function surjective? Choose the answer supported by the diagram.'}
          </p>
        </div>
        <MappingDiagram domain={fourDomain} codomain={fourCodomain} edges={collisionMapping} />
      </div>
    )
  }

  if (stage === 2 || stage === 3) {
    const currentCodomain = stage === 3 ? threeCodomain : fourCodomain
    return (
      <div className="grid gap-4">
        <div>
          <h3 className="text-base font-semibold text-ink">
            {stage === 2 ? 'Make it injective' : 'Build a surjection with unequal set sizes'}
          </h3>
          <p className="mt-2 text-xs leading-relaxed text-muted">
            {stage === 2
              ? 'Rearrange the valid mapping so that no two inputs collide.'
              : 'Build a surjective function from {1, 2, 3, 4} to {A, B, C}.'}
          </p>
        </div>
        <MappingDiagram
          domain={fourDomain}
          codomain={currentCodomain}
          edges={edges}
          selectedInput={selectedInput}
          onSelectInput={onSelectInput}
          onToggleEdge={onChooseOutput}
          singleOutput
        />
      </div>
    )
  }

  if (stage === 4) {
    return (
      <div className="grid gap-5">
        <div>
          <h3 className="text-base font-semibold text-ink">Classify four mappings</h3>
          <p className="mt-2 text-xs leading-relaxed text-muted">
            For each diagram, check separately for collisions and gaps.
          </p>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          {classificationMappings.map((mapping) => (
            <section key={mapping.id} className="grid gap-3 rounded-lg border border-line p-3">
              <h4 className="text-sm font-semibold text-ink">{mapping.label}</h4>
              <MappingDiagram
                domain={mapping.domain}
                codomain={mapping.codomain}
                edges={mapping.edges}
                compact
              />
              <label className="grid gap-1.5 text-xs text-muted">
                Classification
                <select
                  className="rounded-md border border-line-strong bg-panel px-3 py-2 text-xs text-ink outline-0 focus-visible:border-brand-teal focus-visible:ring-2 focus-visible:ring-brand-teal/30"
                  value={classifications[mapping.id] ?? ''}
                  onChange={(event) => onClassify(mapping.id, event.target.value)}
                >
                  <option value="">Choose…</option>
                  {(Object.keys(categoryLabels) as MappingCategory[]).map((category) => (
                    <option key={category} value={category}>
                      {categoryLabels[category]}
                    </option>
                  ))}
                </select>
              </label>
            </section>
          ))}
        </div>
        <PropertyMatrix />
      </div>
    )
  }

  return (
    <div>
      <h3 className="text-base font-semibold text-ink">Which mapping is perfectly reversible?</h3>
      <p className="mt-2 text-xs leading-relaxed text-muted">
        Choose the category whose outputs can each map back to exactly one original input.
      </p>
    </div>
  )
}

function PropertyMatrix() {
  return (
    <div className="rounded-lg border border-line bg-panel-soft p-3">
      <p className="text-2xs font-bold uppercase tracking-widest text-faint">
        Two independent questions
      </p>
      <table className="mt-3 w-full table-fixed text-center text-xs">
        <thead className="text-muted">
          <tr>
            <th className="p-2 text-left">Injective?</th>
            <th className="p-2">Not surjective</th>
            <th className="p-2">Surjective</th>
          </tr>
        </thead>
        <tbody className="text-ink">
          <tr className="border-t border-line">
            <th className="p-2 text-left">Yes</th>
            <td className="p-2">Injective only</td>
            <td className="p-2">Bijective</td>
          </tr>
          <tr className="border-t border-line">
            <th className="p-2 text-left">No</th>
            <td className="p-2">Neither</td>
            <td className="p-2">Surjective only</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

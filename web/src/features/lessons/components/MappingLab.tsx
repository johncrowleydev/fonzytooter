import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { analyzeMapping, MappingDiagram, type MappingEdge } from './MappingDiagram'

const domain = ['1', '2', '3', '4'] as const
const codomain = ['A', 'B', 'C', 'D', 'E'] as const
const exampleFunction: MappingEdge[] = [
  { from: '1', to: 'A' },
  { from: '2', to: 'C' },
  { from: '3', to: 'C' },
  { from: '4', to: 'E' },
]
const constantFunction: MappingEdge[] = domain.map((from) => ({ from, to: 'C' }))

const challenges = [
  'Build a function',
  'Share an output',
  'Break the function precisely',
  'Constant-function diagnostic',
] as const

type Feedback = {
  tone: 'success' | 'error'
  message: string
} | null

const feedbackClasses = {
  success: 'border-accent-teal/30 bg-accent-teal/10',
  error: 'border-accent-coral/30 bg-accent-coral/10',
} as const

export function MappingLab() {
  const [stage, setStage] = useState(0)
  const [edges, setEdges] = useState<MappingEdge[]>([])
  const [selectedInput, setSelectedInput] = useState<string>('1')
  const [feedback, setFeedback] = useState<Feedback>(null)
  const analysis = analyzeMapping(domain, codomain, edges)
  const complete = stage === challenges.length

  function toggleEdge(input: string, output: string) {
    setEdges((current) => {
      const exists = current.some((edge) => edge.from === input && edge.to === output)
      return exists
        ? current.filter((edge) => edge.from !== input || edge.to !== output)
        : [...current, { from: input, to: output }]
    })
    setFeedback(null)
  }

  function checkMapping() {
    if (stage === 0) {
      const unassigned = domain.filter((input) => analysis.outgoingCounts[input] === 0)
      const ambiguous = domain.filter((input) => analysis.outgoingCounts[input] > 1)

      if (unassigned.length > 0) {
        setFeedback({
          tone: 'error',
          message: `Input${unassigned.length > 1 ? 's' : ''} ${unassigned.join(', ')} ${unassigned.length > 1 ? 'have' : 'has'} no assigned output. Every domain input must have one.`,
        })
      } else if (ambiguous.length > 0) {
        setFeedback({
          tone: 'error',
          message: `Input ${ambiguous[0]} has ${analysis.outgoingCounts[ambiguous[0]]} outputs. One input cannot have two different outputs in a function.`,
        })
      } else {
        setFeedback({
          tone: 'success',
          message:
            'Every input has exactly one outgoing assignment, so this relation is a function.',
        })
      }
      return
    }

    if (stage === 1) {
      const sharedOutput = codomain.find((output) => analysis.incomingCounts[output] >= 2)
      if (!analysis.isFunction) {
        setFeedback({
          tone: 'error',
          message: describeFunctionProblem(analysis.outgoingCounts),
        })
      } else if (!sharedOutput) {
        setFeedback({
          tone: 'error',
          message:
            'This is a function, but no output is shared yet. Send two different inputs to the same output.',
        })
      } else {
        const inputs = edges
          .filter((edge) => edge.to === sharedOutput)
          .map((edge) => edge.from)
          .slice(0, 2)
        setFeedback({
          tone: 'success',
          message: `${inputs[0]} and ${inputs[1]} both map to ${sharedOutput}. This is still a function: uniqueness applies to the output assigned to each input, not to how many inputs may share an output.`,
        })
      }
      return
    }

    if (stage === 2) {
      const doubleInputs = domain.filter((input) => analysis.outgoingCounts[input] === 2)
      const allOthersSingle = domain.every(
        (input) => analysis.outgoingCounts[input] === (doubleInputs.includes(input) ? 2 : 1),
      )

      if (doubleInputs.length === 1 && allOthersSingle) {
        setFeedback({
          tone: 'success',
          message: `Input ${doubleInputs[0]} has two outputs, so f(${doubleInputs[0]}) is not uniquely determined. You created the requested precise failure.`,
        })
      } else {
        setFeedback({
          tone: 'error',
          message:
            'Give exactly one input exactly two outputs, while every other input keeps exactly one. Missing outputs or several ambiguous inputs do not match this challenge.',
        })
      }
    }
  }

  function answerConstant(isFunction: boolean) {
    setFeedback(
      isFunction
        ? {
            tone: 'success',
            message:
              'Yes. Every input has exactly one output. Sharing C does not violate functionhood.',
          }
        : {
            tone: 'error',
            message:
              'This is a function. Every input has exactly one output; several inputs are allowed to share C.',
          },
    )
  }

  function advance() {
    if (stage === 2) setEdges(constantFunction)
    setStage((current) => current + 1)
    setFeedback(null)
  }

  function resetCurrentStage() {
    setEdges(stage === 0 ? [] : stage === 3 ? constantFunction : exampleFunction)
    setSelectedInput('1')
    setFeedback(null)
  }

  return (
    <Card className="my-8 grid min-w-0 gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">
            Interactive model
          </p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">Mapping lab</h2>
          <p className="mt-2 max-w-2xl text-xs leading-relaxed text-muted">
            Edit the arrows to test the rule from domain to codomain. Invalid relations are allowed
            on purpose.
          </p>
        </div>
        <Badge tone={complete ? 'teal' : 'gold'}>
          {complete ? 'Complete' : `${stage + 1} / ${challenges.length}`}
        </Badge>
      </div>

      {complete ? (
        <div className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-5">
          <p className="text-2xs font-bold uppercase tracking-widest text-accent-teal">
            Final takeaway
          </p>
          <p className="mt-2 text-sm leading-relaxed text-ink">
            A relation A → B is a function exactly when every element of A has exactly one outgoing
            assignment.
          </p>
        </div>
      ) : (
        <>
          <div>
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">
              Challenge {stage + 1}
            </p>
            <h3 className="mt-1.5 text-base font-semibold text-ink">{challenges[stage]}</h3>
            <p className="mt-2 text-xs leading-relaxed text-muted">{instructionFor(stage)}</p>
          </div>

          {stage === 3 ? (
            <MappingDiagram domain={domain} codomain={codomain} edges={constantFunction} />
          ) : (
            <MappingDiagram
              domain={domain}
              codomain={codomain}
              edges={edges}
              selectedInput={selectedInput}
              onSelectInput={setSelectedInput}
              onToggleEdge={toggleEdge}
            />
          )}

          {stage === 3 ? (
            <div className="flex flex-wrap gap-2" role="group" aria-label="Is this a function?">
              <Button onClick={() => answerConstant(true)}>Yes, it is</Button>
              <Button variant="outline" onClick={() => answerConstant(false)}>
                No, it is not
              </Button>
            </div>
          ) : (
            <Button className="w-max" onClick={checkMapping}>
              Check mapping
            </Button>
          )}

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
              Reset challenge
            </Button>
            {feedback?.tone === 'success' ? (
              <Button onClick={advance}>Next challenge</Button>
            ) : null}
          </div>
        </>
      )}
    </Card>
  )
}

function instructionFor(stage: number) {
  if (stage === 0) return 'Connect the sets so that the relation defines a function from A to B.'
  if (stage === 1) {
    return 'Modify your function so that at least two different inputs map to the same output. It must still be a function.'
  }
  if (stage === 2) {
    return 'Now make the relation stop being a function by giving exactly one input two different outputs.'
  }
  return 'Every input currently maps to C. Is this a function?'
}

function describeFunctionProblem(outgoingCounts: Record<string, number>) {
  const missing = domain.find((input) => outgoingCounts[input] === 0)
  if (missing) return `Input ${missing} has no output. Every domain input must have one.`

  const ambiguous = domain.find((input) => outgoingCounts[input] > 1)
  return `Input ${ambiguous} has more than one output, so its result is not uniquely determined.`
}

import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { DEFAULT_COURSE_ID, modulePath } from '../../app/routes'
import { Badge, Button, Card, PageIntro, SectionHeading } from '../../components/ui'
import type { MockCheckResult } from '../../prototype/types'
import { useTutor } from '../tutor/TutorContext'

const starterCode = `def gradient_descent(gradient, start, learning_rate, steps):
    """Return the final value after a sequence of updates."""
    value = start
    for _ in range(steps):
        value = value - learning_rate * gradient(value)
    return value
`

const exerciseTabStyles = {
  active: 'border-brand-coral text-ink',
  inactive: 'border-transparent text-faint',
} as const

export function Exercise() {
  const { exerciseId = 'gradient-descent-exercise' } = useParams()
  const { setPageContext, openTutorWithContext } = useTutor()
  const [code, setCode] = useState(starterCode)
  const [output, setOutput] = useState<string[]>([])
  const [check, setCheck] = useState<MockCheckResult | null>(null)
  const [activeTab, setActiveTab] = useState<'prompt' | 'tests'>('prompt')
  useEffect(() => {
    setPageContext({
      type: 'exercise',
      title: 'Implement gradient descent',
      exerciseId,
      exerciseTitle: 'Implement gradient descent',
      objectiveIds: ['nn.backpropagation'],
      code,
      lastExecution: check ?? undefined,
    })
  }, [check, code, exerciseId, setPageContext])

  const run = () =>
    setOutput([
      '> running example on f(x) = x²',
      'step 00  x = 4.0000  loss = 16.0000',
      'step 08  x = 0.1678  loss = 0.0282',
      'done in 18ms',
    ])
  const runCheck = () => {
    setCheck({ passed: 2, failed: 1, summary: 'One convergence property still needs attention.' })
    setOutput([
      '> checking gradient_descent',
      '✓ function returns a numeric value',
      '✓ does not mutate input',
      '✗ expected final loss < 0.01, got 0.037',
      '3 tests completed in 22ms',
    ])
  }

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to={modulePath(DEFAULT_COURSE_ID, 'neural-networks')}
      >
        ← Neural Networks From Scratch
      </Link>
      <div className="flex items-start justify-between gap-5 max-sm:block">
        <PageIntro compact eyebrow="Exercise" title="Implement gradient descent" />
      </div>
      <div className="grid grid-cols-[minmax(0,1fr)_245px] gap-5 max-xl:grid-cols-1">
        <main className="grid gap-3.5">
          <Card className="overflow-hidden p-0">
            <div className="flex border-b border-line px-5">
              <button
                className={`mr-4 border-0 border-b-2 bg-transparent px-2 pb-3 pt-3.5 text-xs ${activeTab === 'prompt' ? exerciseTabStyles.active : exerciseTabStyles.inactive}`}
                onClick={() => setActiveTab('prompt')}
                type="button"
              >
                Prompt
              </button>
              <button
                className={`mr-4 border-0 border-b-2 bg-transparent px-2 pb-3 pt-3.5 text-xs ${activeTab === 'tests' ? exerciseTabStyles.active : exerciseTabStyles.inactive}`}
                onClick={() => setActiveTab('tests')}
                type="button"
              >
                Tests
              </button>
            </div>
            {activeTab === 'prompt' ? (
              <div className="px-6 pb-5 pt-4">
                <p className="m-0 text-xs leading-relaxed text-muted">
                  Write <code>gradient_descent()</code> so that it repeatedly updates a value using
                  the gradient and a learning rate.
                </p>
                <div className="mt-4 grid grid-cols-[24px_1fr] gap-x-2.5 gap-y-2">
                  <span className="font-mono text-2xs text-brand-coral">01</span>
                  <p className="m-0 text-xs text-muted">Start from the supplied value.</p>
                  <span className="font-mono text-2xs text-brand-coral">02</span>
                  <p className="m-0 text-xs text-muted">
                    Take exactly <code>steps</code> updates.
                  </p>
                  <span className="font-mono text-2xs text-brand-coral">03</span>
                  <p className="m-0 text-xs text-muted">
                    Return a value close to the minimum of a simple quadratic.
                  </p>
                </div>
              </div>
            ) : (
              <div className="px-6 pb-5 pt-4">
                <p className="m-0 text-xs leading-relaxed text-muted">
                  The checker looks for behavior rather than a magic string.
                </p>
                <div className="mt-4 grid grid-cols-[24px_1fr] gap-x-2.5 gap-y-2">
                  <span className="font-mono text-2xs text-brand-coral">✓</span>
                  <p className="m-0 text-xs text-muted">Converges on a simple quadratic.</p>
                  <span className="font-mono text-2xs text-brand-coral">✓</span>
                  <p className="m-0 text-xs text-muted">Preserves the input value.</p>
                  <span className="font-mono text-2xs text-brand-coral">✓</span>
                  <p className="m-0 text-xs text-muted">
                    Reaches a loss below the target tolerance.
                  </p>
                </div>
              </div>
            )}
          </Card>
          <Card className="overflow-hidden p-0">
            <div className="flex items-center justify-between gap-3 border-b border-line px-5 py-3.5">
              <div>
                <p className="mb-1 font-mono text-xs text-ink">workspace.py</p>
                <span className="flex items-center gap-1.5 text-2xs text-faint">
                  <span className="size-2 rounded-full border border-brand-gold bg-brand-gold ring-2 ring-inset ring-panel" />
                  Unsaved changes
                </span>
              </div>
              <span className="font-mono text-2xs text-faint">Python · 8 lines</span>
            </div>
            <textarea
              className="block min-h-72 w-full resize-y border-0 bg-slate-950 px-6 py-5 font-mono text-xs leading-relaxed text-slate-200 outline-0 focus:ring-1 focus:ring-brand-teal/30 max-sm:min-h-64 max-sm:p-4 max-sm:text-xs"
              spellCheck={false}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              aria-label="Python exercise editor"
            />
            <div className="flex items-center justify-between gap-4 border-t border-line px-4 py-3 max-sm:items-start max-sm:flex-col">
              <div className="flex gap-2">
                <Button onClick={run}>
                  Run <span>▶</span>
                </Button>
                <Button variant="secondary" onClick={runCheck}>
                  Check <span>✓</span>
                </Button>
              </div>
              <button
                className="border-0 bg-transparent p-0 text-xs text-brand-gold"
                onClick={() =>
                  openTutorWithContext({
                    type: 'exercise',
                    title: 'Implement gradient descent',
                    exerciseId,
                    exerciseTitle: 'Implement gradient descent',
                    objectiveIds: ['nn.backpropagation'],
                    code,
                    lastExecution: check ?? undefined,
                  })
                }
                type="button"
              >
                <span>✦</span> Ask tutor
              </button>
            </div>
          </Card>
          <Card className="overflow-hidden p-0">
            <div className="flex items-center justify-between gap-3 px-5 pt-4">
              <SectionHeading
                eyebrow="Feedback"
                title="Output"
                action={
                  check ? (
                    <Badge tone="gold">
                      {check.passed} passed · {check.failed} failed
                    </Badge>
                  ) : (
                    <span className="text-2xs text-faint">Run or check to see output</span>
                  )
                }
              />
            </div>
            <div className="min-h-44 px-5 pb-5 pt-3 font-mono text-xs leading-loose">
              {output.length ? (
                output.map((line, index) => (
                  <div
                    key={`${line}-${index}`}
                    className={`flex gap-2 ${line.startsWith('✓') ? 'text-brand-teal' : line.startsWith('✗') ? 'text-brand-coral' : 'text-muted'}`}
                  >
                    <span className="w-3 text-faint">
                      {line.startsWith('>')
                        ? ''
                        : line.startsWith('✓')
                          ? '✓'
                          : line.startsWith('✗')
                            ? '×'
                            : '·'}
                    </span>
                    {line}
                  </div>
                ))
              ) : (
                <div className="grid min-h-32 place-items-center content-center text-center text-faint">
                  <span className="text-2xl">⌁</span>
                  <p className="font-sans text-xs leading-normal">
                    Nothing run yet.
                    <br />
                    Output appears here.
                  </p>
                </div>
              )}
            </div>
          </Card>
        </main>
        <aside className="grid content-start gap-3.5 max-xl:grid-cols-2 max-sm:grid-cols-1">
          <Card>
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">Objective</p>
            <h3 className="my-2 text-base">Gradient descent</h3>
            <div className="mt-5 flex gap-2 border-t border-line pt-4 text-xs text-brand-gold">
              <span>◐</span>
              <strong className="font-medium leading-relaxed text-muted">
                Implement an update from a gradient
              </strong>
            </div>
          </Card>
        </aside>
      </div>
    </div>
  )
}

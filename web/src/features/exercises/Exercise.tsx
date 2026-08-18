import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
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

export function Exercise() {
  const navigate = useNavigate()
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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <button
        className="justify-self-start border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--muted)] hover:text-[var(--ink)]"
        onClick={() => navigate('/curriculum/neural-networks')}
        type="button"
      >
        ← Neural Networks From Scratch
      </button>
      <div className="flex items-start justify-between gap-5 max-[640px]:block">
        <PageIntro compact eyebrow="Exercise" title="Implement gradient descent" />
      </div>
      <div className="grid grid-cols-[minmax(0,1fr)_245px] gap-[18px] max-[1120px]:grid-cols-1">
        <main className="grid gap-3.5">
          <Card className="overflow-hidden p-0">
            <div className="flex border-b border-[var(--line)] px-[18px]">
              <button
                className={`mr-[15px] border-0 border-b-2 bg-transparent px-[9px] pb-3 pt-3.5 text-[11px] ${activeTab === 'prompt' ? 'border-[var(--coral)] text-[var(--ink)]' : 'border-transparent text-[var(--faint)]'}`}
                onClick={() => setActiveTab('prompt')}
                type="button"
              >
                Prompt
              </button>
              <button
                className={`mr-[15px] border-0 border-b-2 bg-transparent px-[9px] pb-3 pt-3.5 text-[11px] ${activeTab === 'tests' ? 'border-[var(--coral)] text-[var(--ink)]' : 'border-transparent text-[var(--faint)]'}`}
                onClick={() => setActiveTab('tests')}
                type="button"
              >
                Tests
              </button>
            </div>
            {activeTab === 'prompt' ? (
              <div className="px-[25px] pb-[21px] pt-[17px]">
                <p className="m-0 text-xs leading-[1.6] text-[var(--muted)]">
                  Write <code>gradient_descent()</code> so that it repeatedly updates a value using
                  the gradient and a learning rate.
                </p>
                <div className="mt-[15px] grid grid-cols-[24px_1fr] gap-x-2.5 gap-y-[7px]">
                  <span className="font-mono text-[10px] text-[var(--coral)]">01</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">
                    Start from the supplied value.
                  </p>
                  <span className="font-mono text-[10px] text-[var(--coral)]">02</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">
                    Take exactly <code>steps</code> updates.
                  </p>
                  <span className="font-mono text-[10px] text-[var(--coral)]">03</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">
                    Return a value close to the minimum of a simple quadratic.
                  </p>
                </div>
              </div>
            ) : (
              <div className="px-[25px] pb-[21px] pt-[17px]">
                <p className="m-0 text-xs leading-[1.6] text-[var(--muted)]">
                  The checker looks for behavior rather than a magic string.
                </p>
                <div className="mt-[15px] grid grid-cols-[24px_1fr] gap-x-2.5 gap-y-[7px]">
                  <span className="font-mono text-[10px] text-[var(--coral)]">✓</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">
                    Converges on a simple quadratic.
                  </p>
                  <span className="font-mono text-[10px] text-[var(--coral)]">✓</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">Preserves the input value.</p>
                  <span className="font-mono text-[10px] text-[var(--coral)]">✓</span>
                  <p className="m-0 text-[11px] text-[var(--muted)]">
                    Reaches a loss below the target tolerance.
                  </p>
                </div>
              </div>
            )}
          </Card>
          <Card className="overflow-hidden p-0">
            <div className="flex items-center justify-between gap-3 border-b border-[var(--line)] px-[18px] py-3.5">
              <div>
                <p className="mb-[5px] font-mono text-xs text-[var(--ink)]">workspace.py</p>
                <span className="flex items-center gap-1.5 text-[9px] text-[var(--faint)]">
                  <span className="size-[7px] rounded-full border border-[var(--gold)] bg-[var(--gold)] shadow-[inset_0_0_0_2px_var(--panel)]" />
                  Unsaved changes
                </span>
              </div>
              <span className="font-mono text-[9px] text-[var(--faint)]">Python · 8 lines</span>
            </div>
            <textarea
              className="block min-h-[290px] w-full resize-y border-0 bg-[#09131f] px-[22px] py-5 font-mono text-xs leading-[1.75] text-[#c9d5df] outline-0 focus:shadow-[inset_0_0_0_1px_rgba(118,208,192,0.3)] max-[640px]:min-h-[250px] max-[640px]:p-[15px] max-[640px]:text-[11px]"
              spellCheck={false}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              aria-label="Python exercise editor"
            />
            <div className="flex items-center justify-between gap-[15px] border-t border-[var(--line)] px-[17px] py-3 max-[640px]:items-start max-[640px]:flex-col">
              <div className="flex gap-[7px]">
                <Button onClick={run}>
                  Run <span>▶</span>
                </Button>
                <Button variant="secondary" onClick={runCheck}>
                  Check <span>✓</span>
                </Button>
              </div>
              <button
                className="border-0 bg-transparent p-0 text-[11px] text-[var(--gold)]"
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
            <div className="flex items-center justify-between gap-3 px-[18px] pt-[15px]">
              <SectionHeading
                eyebrow="Feedback"
                title="Output"
                action={
                  check ? (
                    <Badge tone="gold">
                      {check.passed} passed · {check.failed} failed
                    </Badge>
                  ) : (
                    <span className="text-[10px] text-[var(--faint)]">
                      Run or check to see output
                    </span>
                  )
                }
              />
            </div>
            <div className="min-h-[170px] px-[19px] pb-[19px] pt-3 font-mono text-[11px] leading-[1.8]">
              {output.length ? (
                output.map((line, index) => (
                  <div
                    key={`${line}-${index}`}
                    className={`flex gap-2 ${line.startsWith('✓') ? 'text-[var(--teal)]' : line.startsWith('✗') ? 'text-[var(--coral)]' : 'text-[var(--muted)]'}`}
                  >
                    <span className="w-[11px] text-[var(--faint)]">
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
                <div className="grid min-h-[130px] place-items-center content-center text-center text-[var(--faint)]">
                  <span className="text-[22px]">⌁</span>
                  <p className="font-sans text-[11px] leading-[1.5]">
                    Nothing run yet.
                    <br />
                    Output appears here.
                  </p>
                </div>
              )}
            </div>
          </Card>
        </main>
        <aside className="grid content-start gap-3.5 max-[1120px]:grid-cols-2 max-[640px]:grid-cols-1">
          <Card>
            <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
              Objective
            </p>
            <h3 className="my-[9px] text-base">Gradient descent</h3>
            <div className="mt-5 flex gap-2 border-t border-[var(--line)] pt-[15px] text-[11px] text-[var(--gold)]">
              <span>◐</span>
              <strong className="font-medium leading-[1.45] text-[var(--muted)]">
                Implement an update from a gradient
              </strong>
            </div>
          </Card>
        </aside>
      </div>
    </div>
  )
}

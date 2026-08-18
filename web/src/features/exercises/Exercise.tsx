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
    <div className="page-stack exercise-page">
      <button
        className="back-link"
        onClick={() => navigate('/curriculum/neural-networks')}
        type="button"
      >
        ← Neural Networks From Scratch
      </button>
      <div className="exercise-header">
        <PageIntro compact eyebrow="Exercise" title="Implement gradient descent" />
      </div>
      <div className="exercise-layout">
        <main>
          <Card className="prompt-card">
            <div className="exercise-tabs">
              <button
                className={activeTab === 'prompt' ? 'active' : ''}
                onClick={() => setActiveTab('prompt')}
                type="button"
              >
                Prompt
              </button>
              <button
                className={activeTab === 'tests' ? 'active' : ''}
                onClick={() => setActiveTab('tests')}
                type="button"
              >
                Tests
              </button>
            </div>
            {activeTab === 'prompt' ? (
              <div className="exercise-prompt">
                <p>
                  Write <code>gradient_descent()</code> so that it repeatedly updates a value using
                  the gradient and a learning rate.
                </p>
                <div className="prompt-list">
                  <span>01</span>
                  <p>Start from the supplied value.</p>
                  <span>02</span>
                  <p>
                    Take exactly <code>steps</code> updates.
                  </p>
                  <span>03</span>
                  <p>Return a value close to the minimum of a simple quadratic.</p>
                </div>
              </div>
            ) : (
              <div className="exercise-prompt">
                <p>The checker looks for behavior rather than a magic string.</p>
                <div className="prompt-list">
                  <span>✓</span>
                  <p>Converges on a simple quadratic.</p>
                  <span>✓</span>
                  <p>Preserves the input value.</p>
                  <span>✓</span>
                  <p>Reaches a loss below the target tolerance.</p>
                </div>
              </div>
            )}
          </Card>
          <Card className="editor-card">
            <div className="editor-header">
              <div>
                <p className="eyebrow">workspace.py</p>
                <span className="editor-status">
                  <span className="status-dot status-working status-small" />
                  Unsaved changes
                </span>
              </div>
              <span className="line-count">Python · 8 lines</span>
            </div>
            <textarea
              className="code-editor"
              spellCheck={false}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              aria-label="Python exercise editor"
            />
            <div className="editor-footer">
              <div className="editor-actions">
                <Button onClick={run}>
                  Run <span>▶</span>
                </Button>
                <Button variant="secondary" onClick={runCheck}>
                  Check <span>✓</span>
                </Button>
              </div>
              <button
                className="ask-tutor-inline"
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
          <Card className="output-card">
            <div className="output-header">
              <SectionHeading
                eyebrow="Feedback"
                title="Output"
                action={
                  check ? (
                    <Badge tone="gold">
                      {check.passed} passed · {check.failed} failed
                    </Badge>
                  ) : (
                    <span className="small-muted">Run or check to see output</span>
                  )
                }
              />
            </div>
            <div className="output-body">
              {output.length ? (
                output.map((line, index) => (
                  <div
                    key={`${line}-${index}`}
                    className={`output-line ${line.startsWith('✓') ? 'pass' : line.startsWith('✗') ? 'fail' : ''}`}
                  >
                    <span className="output-prompt">
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
                <div className="empty-output">
                  <span>⌁</span>
                  <p>
                    Nothing run yet.
                    <br />
                    Output appears here.
                  </p>
                </div>
              )}
            </div>
          </Card>
        </main>
        <aside className="exercise-aside">
          <Card>
            <p className="eyebrow">Objective</p>
            <h3>Gradient descent</h3>
            <div className="aside-objective">
              <span>◐</span>
              <strong>Implement an update from a gradient</strong>
            </div>
          </Card>
        </aside>
      </div>
    </div>
  )
}

import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, ProgressBar, SectionHeading } from '../../components/ui'
import { recentActivity } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function Dashboard() {
  const navigate = useNavigate()
  const { setPageContext } = useTutor()
  useEffect(() => {
    setPageContext({ type: 'dashboard', title: 'Home' })
  }, [setPageContext])

  return (
    <div className="page-stack dashboard-page">
      <PageIntro
        eyebrow="Monday, August 17"
        title="Good evening, Fonzy."
        detail="A focused place to pick up where you left off."
      />

      <section className="hero-grid">
        <Card className="continue-card">
          <div className="card-topline">
            <Badge tone="coral">Continue learning</Badge>
          </div>
          <div className="continue-layout">
            <div>
              <h2>Backpropagation</h2>
              <p className="card-kicker">
                Neural Networks From Scratch <span>·</span> Lesson 4
              </p>
              <p className="body-muted">
                Trace how a small local derivative becomes a useful update across a computational
                graph.
              </p>
              <Button onClick={() => navigate('/lesson/backpropagation')}>
                Continue lesson <span>→</span>
              </Button>
            </div>
            <div className="orbit-visual" aria-hidden="true">
              <span className="orbit-ring ring-one" />
              <span className="orbit-ring ring-two" />
              <span className="orbit-node node-a">∂</span>
              <span className="orbit-node node-b">w</span>
              <span className="orbit-node node-c">L</span>
              <span className="orbit-line line-a" />
              <span className="orbit-line line-b" />
            </div>
          </div>
        </Card>
        <div className="quick-stack">
          <Card className="review-ready-card">
            <div className="quick-icon teal">↺</div>
            <div>
              <p className="eyebrow">Ready when you are</p>
              <h3>14 reviews ready</h3>
              <button className="text-link" onClick={() => navigate('/review')}>
                Start review <span>→</span>
              </button>
            </div>
          </Card>
          <Card className="mastery-ready-card">
            <div className="quick-icon gold">◎</div>
            <div>
              <p className="eyebrow">Ready to test</p>
              <h3>2 concepts ready for a check</h3>
              <p className="small-muted">Matrix multiplication · Partial derivatives</p>
            </div>
          </Card>
        </div>
      </section>

      <section className="dashboard-columns">
        <Card className="project-card">
          <SectionHeading
            eyebrow="Current project"
            title="Neural Network From Scratch"
            action={
              <button className="text-link" onClick={() => navigate('/projects/nn-scratch')}>
                Open project →
              </button>
            }
          />
          <p className="body-muted">
            A repository-based lab for turning the pieces of this module into a small, inspectable
            implementation.
          </p>
          <div className="project-progress-line">
            <strong>3 of 7</strong>
            <span>objectives demonstrated</span>
            <span className="progress-percent">43%</span>
          </div>
          <ProgressBar value={43} tone="coral" />
          <div className="mini-objectives">
            <span>✓ Forward propagation</span>
            <span>◐ Gradient computation</span>
            <span>○ Training loop</span>
          </div>
        </Card>
        <Card className="activity-card">
          <SectionHeading
            eyebrow="Recent activity"
            title="A quiet trail of progress"
            action={<button className="text-link">View all</button>}
          />
          <div className="activity-list">
            {recentActivity.map((item) => (
              <div className="activity-item" key={item.id}>
                <span className={`activity-icon activity-${item.kind}`}>
                  {item.kind === 'review'
                    ? '↺'
                    : item.kind === 'lesson'
                      ? '✓'
                      : item.kind === 'exercise'
                        ? '⌘'
                        : '✦'}
                </span>
                <div>
                  <strong>{item.label}</strong>
                  <span>{item.detail}</span>
                </div>
                <time>{item.time}</time>
              </div>
            ))}
          </div>
        </Card>
      </section>

      <section className="dashboard-footer-row">
        <div className="focus-note">
          <span className="focus-mark">✦</span>
          <div>
            <strong>One useful next step</strong>
            <p>
              Finish the backpropagation lesson, then try the gradient descent exercise while the
              idea is fresh.
            </p>
          </div>
        </div>
        <div className="dashboard-stats">
          <div>
            <strong>42</strong>
            <span>introduced</span>
          </div>
          <div>
            <strong>31</strong>
            <span>recall strong</span>
          </div>
          <div>
            <strong>18</strong>
            <span>applied</span>
          </div>
        </div>
      </section>
    </div>
  )
}

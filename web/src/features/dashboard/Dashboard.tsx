import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, ProgressBar, SectionHeading } from '../../components/ui'
import { recentActivity } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

const activityKindStyles = {
  review: 'bg-brand-teal/10 text-brand-teal',
  lesson: 'bg-brand-gold/10 text-brand-gold',
  exercise: 'bg-brand-coral/10 text-brand-coral',
  tutor: 'bg-brand-violet/10 text-brand-violet',
  project: 'bg-brand-blue/10 text-brand-blue',
} as const

const activityKindIcons = {
  review: '↺',
  lesson: '✓',
  exercise: '⌘',
  tutor: '✦',
  project: '✦',
} as const

export function Dashboard() {
  const navigate = useNavigate()
  const { setPageContext } = useTutor()
  useEffect(() => {
    setPageContext({ type: 'dashboard', title: 'Home' })
  }, [setPageContext])

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro
        eyebrow="Monday, August 17"
        title="Good evening, Fonzy."
        detail="A focused place to pick up where you left off."
      />

      <section className="grid grid-cols-[minmax(0,1.58fr)_minmax(260px,0.84fr)] gap-3.5 max-lg:grid-cols-1">
        <Card className="min-h-72 border-brand-coral/30 bg-panel p-6 max-sm:min-h-0 max-sm:p-5">
          <div className="flex items-center justify-between gap-3">
            <Badge tone="coral">Continue learning</Badge>
          </div>
          <div className="relative flex min-h-52 items-center justify-between max-sm:min-h-56">
            <div>
              <h2 className="my-5 text-4xl font-semibold leading-none tracking-tight">
                Backpropagation
              </h2>
              <p className="m-0 text-xs font-semibold text-brand-coral">
                Neural Networks From Scratch <span className="px-1 text-faint">·</span> Lesson 4
              </p>
              <p className="my-4 max-w-sm text-xs leading-relaxed text-muted max-sm:max-w-56">
                Trace how a small local derivative becomes a useful update across a computational
                graph.
              </p>
              <Button onClick={() => navigate('/lesson/backpropagation')}>
                Continue lesson <span>→</span>
              </Button>
            </div>
            <div
              className="relative mr-3 ml-5 h-[170px] w-48 opacity-90 max-sm:absolute max-sm:right-[-9px] max-sm:bottom-0 max-sm:m-0 max-sm:origin-bottom-right max-sm:scale-75"
              aria-hidden="true"
            >
              <span className="absolute inset-[27px_11px_26px_12px] rotate-[-24deg] rounded-full border border-brand-coral/30" />
              <span className="absolute inset-[12px_42px_10px_43px] rotate-[62deg] rounded-full border border-brand-teal/30" />
              <span className="absolute top-[17px] right-8 grid size-8 place-items-center rounded-full bg-brand-coral font-serif text-base font-semibold text-brand-ink">
                ∂
              </span>
              <span className="absolute bottom-[19px] left-[29px] grid size-8 place-items-center rounded-full bg-brand-teal font-serif text-base font-semibold text-brand-ink">
                w
              </span>
              <span className="absolute top-[70px] left-20 grid size-8 place-items-center rounded-full bg-brand-gold font-serif text-base font-semibold text-brand-ink">
                L
              </span>
              <span className="absolute top-[85px] left-[45px] h-px w-[73px] origin-left rotate-[-32deg] bg-white/35" />
              <span className="absolute top-[89px] left-[94px] h-px w-[62px] origin-left rotate-[52deg] bg-white/35" />
            </div>
          </div>
        </Card>
        <div className="grid gap-3.5 max-sm:gap-2">
          <Card className="flex min-h-32 items-start gap-4 p-5 max-sm:min-h-28">
            <div className="grid size-9 place-items-center rounded-lg bg-brand-teal/10 text-xl text-brand-teal">
              ↺
            </div>
            <div>
              <p className="mt-px text-2xs font-bold uppercase tracking-widest text-faint">
                Ready when you are
              </p>
              <h3 className="my-2 text-lg tracking-tight">14 reviews ready</h3>
              <button
                className="border-0 bg-transparent p-0 text-xs font-bold text-brand-teal hover:text-ink"
                onClick={() => navigate('/review')}
              >
                Start review <span>→</span>
              </button>
            </div>
          </Card>
          <Card className="flex min-h-32 items-start gap-4 p-5 max-sm:min-h-28">
            <div className="grid size-9 place-items-center rounded-lg bg-brand-gold/10 text-xl text-brand-gold">
              ◎
            </div>
            <div>
              <p className="mt-px text-2xs font-bold uppercase tracking-widest text-faint">
                Ready to test
              </p>
              <h3 className="my-2 text-lg tracking-tight">2 concepts ready for a check</h3>
              <p className="text-2xs text-faint">Matrix multiplication · Partial derivatives</p>
            </div>
          </Card>
        </div>
      </section>

      <section className="grid grid-cols-[1.05fr_0.95fr] gap-3.5 max-lg:grid-cols-1">
        <Card className="min-h-72">
          <SectionHeading
            eyebrow="Current project"
            title="Neural Network From Scratch"
            action={
              <button
                className="border-0 bg-transparent p-0 text-xs font-bold text-brand-teal hover:text-ink"
                onClick={() => navigate('/projects/nn-scratch')}
              >
                Open project →
              </button>
            }
          />
          <p className="text-xs leading-relaxed text-muted">
            A repository-based lab for turning the pieces of this module into a small, inspectable
            implementation.
          </p>
          <div className="mt-6 flex items-baseline gap-2">
            <strong className="text-3xl tracking-tight">3 of 7</strong>
            <span className="text-xs text-muted">objectives demonstrated</span>
            <span className="ml-auto text-xs font-bold text-brand-coral">43%</span>
          </div>
          <ProgressBar value={43} tone="coral" />
          <div className="flex flex-wrap gap-x-4 gap-y-2 text-2xs text-muted">
            <span className="text-brand-teal">✓ Forward propagation</span>
            <span>◐ Gradient computation</span>
            <span>○ Training loop</span>
          </div>
        </Card>
        <Card className="min-h-72">
          <SectionHeading
            eyebrow="Recent activity"
            title="A quiet trail of progress"
            action={
              <button className="border-0 bg-transparent p-0 text-xs font-bold text-brand-teal hover:text-ink">
                View all
              </button>
            }
          />
          <div className="grid">
            {recentActivity.map((item) => (
              <div
                className="grid grid-cols-[24px_1fr_auto] items-center gap-2.5 border-t border-line py-2.5"
                key={item.id}
              >
                <span
                  className={`grid size-6 place-items-center rounded-lg text-xs ${activityKindStyles[item.kind]}`}
                >
                  {activityKindIcons[item.kind]}
                </span>
                <div>
                  <strong className="block text-xs font-semibold">{item.label}</strong>
                  <span className="mt-1 block text-2xs text-faint">{item.detail}</span>
                </div>
                <time className="text-2xs text-faint">{item.time}</time>
              </div>
            ))}
          </div>
        </Card>
      </section>

      <section className="flex items-center justify-between gap-5 border-t border-line pt-6 max-sm:block">
        <div className="flex max-w-xl items-start gap-3">
          <span className="text-base text-brand-gold">✦</span>
          <div>
            <strong className="text-xs">One useful next step</strong>
            <p className="mt-1.5 text-xs leading-normal text-muted">
              Finish the backpropagation lesson, then try the gradient descent exercise while the
              idea is fresh.
            </p>
          </div>
        </div>
        <div className="flex gap-7 max-sm:mt-5 max-sm:justify-between">
          <div className="grid gap-1">
            <strong className="text-2xl tracking-tight">42</strong>
            <span className="text-2xs uppercase tracking-wide text-faint">introduced</span>
          </div>
          <div className="grid gap-1">
            <strong className="text-2xl tracking-tight">31</strong>
            <span className="text-2xs uppercase tracking-wide text-faint">recall strong</span>
          </div>
          <div className="grid gap-1">
            <strong className="text-2xl tracking-tight">18</strong>
            <span className="text-2xs uppercase tracking-wide text-faint">applied</span>
          </div>
        </div>
      </section>
    </div>
  )
}

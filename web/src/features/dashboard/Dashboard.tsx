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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <PageIntro
        eyebrow="Monday, August 17"
        title="Good evening, Fonzy."
        detail="A focused place to pick up where you left off."
      />

      <section className="grid grid-cols-[minmax(0,1.58fr)_minmax(260px,0.84fr)] gap-3.5 max-[860px]:grid-cols-1">
        <Card className="min-h-[282px] border-[rgba(239,145,110,0.29)] bg-[var(--panel)] p-[25px_28px] max-[640px]:min-h-0 max-[640px]:p-5">
          <div className="flex items-center justify-between gap-3">
            <Badge tone="coral">Continue learning</Badge>
          </div>
          <div className="relative flex min-h-[205px] items-center justify-between max-[640px]:min-h-[225px]">
            <div>
              <h2 className="my-[22px_4px] text-[clamp(26px,3.3vw,39px)] font-semibold leading-none tracking-[-0.055em]">
                Backpropagation
              </h2>
              <p className="m-0 text-[11px] font-semibold text-[var(--coral)]">
                Neural Networks From Scratch <span className="px-1 text-[var(--faint)]">·</span>{' '}
                Lesson 4
              </p>
              <p className="my-[17px_21px] max-w-[370px] text-xs leading-[1.65] text-[var(--muted)] max-[640px]:max-w-[230px]">
                Trace how a small local derivative becomes a useful update across a computational
                graph.
              </p>
              <Button onClick={() => navigate('/lesson/backpropagation')}>
                Continue lesson <span>→</span>
              </Button>
            </div>
            <div
              className="relative mr-3 ml-[22px] h-[170px] w-[190px] opacity-[0.92] max-[640px]:absolute max-[640px]:right-[-9px] max-[640px]:bottom-0 max-[640px]:m-0 max-[640px]:origin-bottom-right max-[640px]:scale-[0.72]"
              aria-hidden="true"
            >
              <span className="absolute inset-[27px_11px_26px_12px] rotate-[-24deg] rounded-[50%] border border-[rgba(239,145,110,0.28)]" />
              <span className="absolute inset-[12px_42px_10px_43px] rotate-[62deg] rounded-[50%] border border-[rgba(118,208,192,0.28)]" />
              <span className="absolute top-[17px] right-8 grid size-[31px] place-items-center rounded-full bg-[var(--coral)] font-serif text-base font-semibold text-[#0c1721]">
                ∂
              </span>
              <span className="absolute bottom-[19px] left-[29px] grid size-[31px] place-items-center rounded-full bg-[var(--teal)] font-serif text-base font-semibold text-[#0c1721]">
                w
              </span>
              <span className="absolute top-[70px] left-20 grid size-[31px] place-items-center rounded-full bg-[var(--gold)] font-serif text-base font-semibold text-[#0c1721]">
                L
              </span>
              <span className="absolute top-[85px] left-[45px] h-px w-[73px] origin-left rotate-[-32deg] bg-[rgba(231,237,243,0.34)]" />
              <span className="absolute top-[89px] left-[94px] h-px w-[62px] origin-left rotate-[52deg] bg-[rgba(231,237,243,0.34)]" />
            </div>
          </div>
        </Card>
        <div className="grid gap-3.5 max-[640px]:gap-[9px]">
          <Card className="flex min-h-[134px] items-start gap-[15px] p-[21px] max-[640px]:min-h-[112px]">
            <div className="grid size-[34px] place-items-center rounded-[9px] bg-[rgba(118,208,192,0.11)] text-[19px] text-[var(--teal)]">
              ↺
            </div>
            <div>
              <p className="mt-px text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
                Ready when you are
              </p>
              <h3 className="my-[6px_10px] text-[17px] tracking-[-0.03em]">14 reviews ready</h3>
              <button
                className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]"
                onClick={() => navigate('/review')}
              >
                Start review <span>→</span>
              </button>
            </div>
          </Card>
          <Card className="flex min-h-[134px] items-start gap-[15px] p-[21px] max-[640px]:min-h-[112px]">
            <div className="grid size-[34px] place-items-center rounded-[9px] bg-[rgba(225,184,106,0.11)] text-[19px] text-[var(--gold)]">
              ◎
            </div>
            <div>
              <p className="mt-px text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
                Ready to test
              </p>
              <h3 className="my-[6px_10px] text-[17px] tracking-[-0.03em]">
                2 concepts ready for a check
              </h3>
              <p className="text-[10px] text-[var(--faint)]">
                Matrix multiplication · Partial derivatives
              </p>
            </div>
          </Card>
        </div>
      </section>

      <section className="grid grid-cols-[1.05fr_0.95fr] gap-3.5 max-[860px]:grid-cols-1">
        <Card className="min-h-[286px]">
          <SectionHeading
            eyebrow="Current project"
            title="Neural Network From Scratch"
            action={
              <button
                className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]"
                onClick={() => navigate('/projects/nn-scratch')}
              >
                Open project →
              </button>
            }
          />
          <p className="text-xs leading-[1.65] text-[var(--muted)]">
            A repository-based lab for turning the pieces of this module into a small, inspectable
            implementation.
          </p>
          <div className="mt-[26px] flex items-baseline gap-[7px]">
            <strong className="text-[26px] tracking-[-0.04em]">3 of 7</strong>
            <span className="text-[11px] text-[var(--muted)]">objectives demonstrated</span>
            <span className="ml-auto text-[11px] font-bold text-[var(--coral)]">43%</span>
          </div>
          <ProgressBar value={43} tone="coral" />
          <div className="flex flex-wrap gap-x-4 gap-y-[7px] text-[10px] text-[var(--muted)]">
            <span className="text-[var(--teal)]">✓ Forward propagation</span>
            <span>◐ Gradient computation</span>
            <span>○ Training loop</span>
          </div>
        </Card>
        <Card className="min-h-[286px]">
          <SectionHeading
            eyebrow="Recent activity"
            title="A quiet trail of progress"
            action={
              <button className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]">
                View all
              </button>
            }
          />
          <div className="grid">
            {recentActivity.map((item) => (
              <div
                className="grid grid-cols-[24px_1fr_auto] items-center gap-2.5 border-t border-[var(--line)] py-2.5"
                key={item.id}
              >
                <span
                  className={`grid size-[23px] place-items-center rounded-[7px] text-xs ${item.kind === 'review' ? 'bg-[rgba(118,208,192,0.1)] text-[var(--teal)]' : item.kind === 'lesson' ? 'bg-[rgba(225,184,106,0.1)] text-[var(--gold)]' : item.kind === 'exercise' ? 'bg-[rgba(239,145,110,0.1)] text-[var(--coral)]' : 'bg-[rgba(169,155,231,0.1)] text-[var(--violet)]'}`}
                >
                  {item.kind === 'review'
                    ? '↺'
                    : item.kind === 'lesson'
                      ? '✓'
                      : item.kind === 'exercise'
                        ? '⌘'
                        : '✦'}
                </span>
                <div>
                  <strong className="block text-[11px] font-semibold">{item.label}</strong>
                  <span className="mt-[3px] block text-[10px] text-[var(--faint)]">
                    {item.detail}
                  </span>
                </div>
                <time className="text-[9px] text-[var(--faint)]">{item.time}</time>
              </div>
            ))}
          </div>
        </Card>
      </section>

      <section className="flex items-center justify-between gap-5 border-t border-[var(--line)] pt-[23px] max-[640px]:block">
        <div className="flex max-w-[590px] items-start gap-[13px]">
          <span className="text-[15px] text-[var(--gold)]">✦</span>
          <div>
            <strong className="text-[11px]">One useful next step</strong>
            <p className="mt-1.5 text-[11px] leading-[1.5] text-[var(--muted)]">
              Finish the backpropagation lesson, then try the gradient descent exercise while the
              idea is fresh.
            </p>
          </div>
        </div>
        <div className="flex gap-[27px] max-[640px]:mt-5 max-[640px]:justify-between">
          <div className="grid gap-[3px]">
            <strong className="text-[21px] tracking-[-0.04em]">42</strong>
            <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
              introduced
            </span>
          </div>
          <div className="grid gap-[3px]">
            <strong className="text-[21px] tracking-[-0.04em]">31</strong>
            <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
              recall strong
            </span>
          </div>
          <div className="grid gap-[3px]">
            <strong className="text-[21px] tracking-[-0.04em]">18</strong>
            <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
              applied
            </span>
          </div>
        </div>
      </section>
    </div>
  )
}

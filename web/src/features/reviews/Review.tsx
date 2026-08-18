import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, ProgressBar } from '../../components/ui'
import { reviewCards } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function Review() {
  const navigate = useNavigate()
  const { setPageContext } = useTutor()
  const [cardIndex, setCardIndex] = useState(0)
  const [revealed, setRevealed] = useState(false)
  const [rated, setRated] = useState<string[]>([])
  const current = reviewCards[cardIndex]
  const complete = cardIndex >= reviewCards.length
  useEffect(() => {
    setPageContext({ type: 'review', title: 'Review', objectiveId: current?.objectiveId })
  }, [current?.objectiveId, setPageContext])
  const progress = Math.round((Math.min(cardIndex, reviewCards.length) / reviewCards.length) * 100)
  const remaining = useMemo(() => Math.max(reviewCards.length - rated.length, 0), [rated.length])

  if (complete)
    return (
      <ReviewComplete
        reviewed={rated.length}
        revisit={rated.filter((item) => item === 'again' || item === 'hard').length}
      />
    )

  function rate(rating: string) {
    setRated((items) => [...items, rating])
    setRevealed(false)
    setCardIndex((index) => index + 1)
  }

  return (
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <PageIntro compact title="Review">
        <div className="mt-[18px] flex gap-2.5 text-[10px] text-[var(--faint)]">
          <span>
            <strong className="text-[var(--teal)]">{remaining}</strong> remaining
          </span>
          <span>·</span>
          <span>Approx. 8 min</span>
        </div>
      </PageIntro>
      <div className="grid grid-cols-[100px_1fr_45px] items-center gap-[15px] text-[10px] text-[var(--faint)] max-[640px]:grid-cols-[83px_1fr_32px] max-[640px]:gap-[9px]">
        <span>Session progress</span>
        <ProgressBar value={progress} tone="teal" />
        <span className="text-right text-[var(--muted)]">
          {Math.min(cardIndex + 1, reviewCards.length)} / {reviewCards.length}
        </span>
      </div>
      <div className="grid grid-cols-[minmax(0,680px)_230px] justify-center gap-[42px] max-[860px]:grid-cols-1">
        <div>
          <Card className="flex min-h-[390px] flex-col p-[25px] max-[640px]:min-h-[420px] max-[640px]:p-[18px]">
            <div className="flex items-center justify-between gap-3">
              <Badge tone="teal">{current.objectiveLabel}</Badge>
              <span className="text-[9px] text-[var(--faint)]">
                Last reviewed {current.lastReviewed}
              </span>
            </div>
            <div className="my-auto mx-2.5 grid justify-items-center gap-3.5 text-center">
              <span className="grid size-[42px] place-items-center rounded-full border border-[rgba(118,208,192,0.37)] font-serif text-xl text-[var(--teal)]">
                ?
              </span>
              <h2 className="max-w-[490px] text-[clamp(22px,3vw,31px)] font-semibold leading-[1.22] tracking-[-0.04em]">
                {current.prompt}
              </h2>
            </div>
            {revealed ? (
              <div className="border-t border-[var(--line)] bg-[rgba(118,208,192,0.04)] px-[15px] py-[18px]">
                <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
                  Answer
                </p>
                <p className="my-2 text-[13px] leading-[1.7] text-[var(--ink)]">{current.answer}</p>
              </div>
            ) : (
              <div className="border-t border-[var(--line)] pt-[15px] text-[11px] text-[var(--faint)]">
                <span className="mr-[9px] text-[9px] font-bold uppercase tracking-[0.1em] text-[var(--gold)]">
                  Hint
                </span>
                {current.hint}
              </div>
            )}
            <div className="mt-[18px] flex justify-center">
              {!revealed ? (
                <Button onClick={() => setRevealed(true)}>
                  Reveal answer <span>↓</span>
                </Button>
              ) : (
                <div className="flex w-full items-center gap-[7px] max-[640px]:gap-1">
                  <span className="mr-auto text-[9px] text-[var(--faint)] max-[640px]:hidden">
                    How did that feel?
                  </span>
                  <button
                    onClick={() => rate('again')}
                    className="min-w-[61px] rounded-[7px] border border-[var(--line)] bg-white/[0.025] px-[7px] py-2 text-[var(--muted)] hover:border-[var(--coral)] hover:text-[var(--ink)] max-[640px]:min-w-0 max-[640px]:flex-1"
                    type="button"
                  >
                    <strong className="block text-[10px]">Again</strong>
                    <small className="mt-[3px] block text-[8px] text-[var(--faint)]">now</small>
                  </button>
                  <button
                    onClick={() => rate('hard')}
                    className="min-w-[61px] rounded-[7px] border border-[var(--line)] bg-white/[0.025] px-[7px] py-2 text-[var(--muted)] hover:border-[var(--gold)] hover:text-[var(--ink)] max-[640px]:min-w-0 max-[640px]:flex-1"
                    type="button"
                  >
                    <strong className="block text-[10px]">Hard</strong>
                    <small className="mt-[3px] block text-[8px] text-[var(--faint)]">soon</small>
                  </button>
                  <button
                    onClick={() => rate('good')}
                    className="min-w-[61px] rounded-[7px] border border-[var(--line)] bg-white/[0.025] px-[7px] py-2 text-[var(--muted)] hover:border-[var(--teal)] hover:text-[var(--ink)] max-[640px]:min-w-0 max-[640px]:flex-1"
                    type="button"
                  >
                    <strong className="block text-[10px]">Good</strong>
                    <small className="mt-[3px] block text-[8px] text-[var(--faint)]">later</small>
                  </button>
                  <button
                    onClick={() => rate('easy')}
                    className="min-w-[61px] rounded-[7px] border border-[var(--line)] bg-white/[0.025] px-[7px] py-2 text-[var(--muted)] hover:border-[var(--teal)] hover:text-[var(--ink)] max-[640px]:min-w-0 max-[640px]:flex-1"
                    type="button"
                  >
                    <strong className="block text-[10px]">Easy</strong>
                    <small className="mt-[3px] block text-[8px] text-[var(--faint)]">further</small>
                  </button>
                </div>
              )}
            </div>
          </Card>
          <div className="mt-3 flex items-center gap-[9px] text-[10px] text-[var(--faint)]">
            <span className="mt-1 size-[7px] shrink-0 rounded-full bg-[var(--teal)] shadow-[0_0_0_4px_rgba(118,208,192,0.1)]" />
            <strong className="font-medium text-[var(--muted)]">{current.objectiveLabel}</strong>
            <button
              className="ml-auto border-0 bg-transparent p-0 text-[10px] text-[var(--teal)]"
              type="button"
              onClick={() => navigate('/progress')}
            >
              View objective →
            </button>
            <button
              className="border-0 bg-transparent p-0 text-[10px] text-[var(--teal)]"
              type="button"
              onClick={() => navigate('/')}
            >
              Exit
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function ReviewComplete({ reviewed, revisit }: { reviewed: number; revisit: number }) {
  const navigate = useNavigate()
  return (
    <div className="grid min-h-[70vh] max-w-[1140px] content-center justify-items-center gap-[29px] text-center max-[640px]:min-h-[72vh] max-[640px]:gap-[21px]">
      <div className="grid size-16 place-items-center rounded-full border border-[rgba(118,208,192,0.35)] bg-[rgba(118,208,192,0.1)] text-[28px] text-[var(--teal)]">
        ✓
      </div>
      <p className="mt-7 text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
        Session complete
      </p>
      <h1 className="my-2.5 text-[45px] tracking-[-0.06em] max-[640px]:text-[38px]">
        Review complete.
      </h1>
      <p className="text-sm text-[var(--muted)]">
        You gave {reviewed} ideas your attention.{' '}
        {revisit ? `${revisit} are worth revisiting soon.` : 'Nice clean pass.'}
      </p>
      <div className="my-[27px] flex gap-px">
        <div className="grid min-w-[125px] gap-[5px] bg-[rgba(157,185,194,0.06)] p-4 max-[640px]:min-w-[95px] max-[640px]:px-2 max-[640px]:py-[13px]">
          <strong className="text-[23px] max-[640px]:text-xl">{reviewed}</strong>
          <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
            reviewed
          </span>
        </div>
        <div className="grid min-w-[125px] gap-[5px] bg-[rgba(157,185,194,0.06)] p-4 max-[640px]:min-w-[95px] max-[640px]:px-2 max-[640px]:py-[13px]">
          <strong className="text-[23px] max-[640px]:text-xl">{revisit}</strong>
          <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
            worth revisiting
          </span>
        </div>
        <div className="grid min-w-[125px] gap-[5px] bg-[rgba(157,185,194,0.06)] p-4 max-[640px]:min-w-[95px] max-[640px]:px-2 max-[640px]:py-[13px]">
          <strong className="text-[23px] max-[640px]:text-xl">3m</strong>
          <span className="text-[9px] uppercase tracking-[0.08em] text-[var(--faint)]">
            focused time
          </span>
        </div>
      </div>
      <div className="flex gap-[9px]">
        <Button onClick={() => navigate('/')}>Back home</Button>
        <Button variant="secondary" onClick={() => navigate('/progress')}>
          View progress
        </Button>
      </div>
    </div>
  )
}

import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, Button, Card, PageIntro, ProgressBar } from '../../components/ui'
import { reviewCards } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function Review() {
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
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Review">
        <div className="mt-4 text-2xs text-faint">
          <strong className="text-brand-teal">{remaining}</strong> remaining
        </div>
      </PageIntro>
      <div className="grid grid-cols-[100px_1fr_45px] items-center gap-4 text-2xs text-faint max-sm:grid-cols-[83px_1fr_32px] max-sm:gap-2">
        <span>Session progress</span>
        <ProgressBar value={progress} tone="teal" />
        <span className="text-right text-muted">
          {Math.min(cardIndex + 1, reviewCards.length)} / {reviewCards.length}
        </span>
      </div>
      <div className="grid grid-cols-[minmax(0,680px)_230px] justify-center gap-10 max-lg:grid-cols-1">
        <div>
          <Card className="flex min-h-96 flex-col p-6 max-sm:min-h-104 max-sm:p-5">
            <div className="flex items-center justify-between gap-3">
              <Badge tone="teal">{current.objectiveLabel}</Badge>
              <span className="text-2xs text-faint">Last reviewed {current.lastReviewed}</span>
            </div>
            <div className="mx-2.5 my-auto grid justify-items-center gap-3.5 text-center">
              <span className="grid size-11 place-items-center rounded-full border border-brand-teal/40 font-serif text-xl text-brand-teal">
                ?
              </span>
              <h2 className="max-w-lg text-2xl font-semibold leading-tight tracking-tight sm:text-3xl">
                {current.prompt}
              </h2>
            </div>
            {revealed ? (
              <div className="border-t border-line bg-brand-teal/5 px-4 py-5">
                <p className="text-2xs font-bold uppercase tracking-widest text-faint">Answer</p>
                <p className="my-2 text-sm leading-relaxed text-ink">{current.answer}</p>
              </div>
            ) : (
              <div className="border-t border-line pt-4 text-xs text-faint">
                <span className="mr-2 text-2xs font-bold uppercase tracking-wide text-brand-gold">
                  Hint
                </span>
                {current.hint}
              </div>
            )}
            <div className="mt-4 flex justify-center">
              {!revealed ? (
                <Button onClick={() => setRevealed(true)}>
                  Reveal answer <span>↓</span>
                </Button>
              ) : (
                <div className="flex w-full items-center gap-2 max-sm:gap-1">
                  <span className="mr-auto text-2xs text-faint max-sm:hidden">
                    How did that feel?
                  </span>
                  <button
                    onClick={() => rate('again')}
                    className="min-w-16 rounded-lg border border-line bg-white/5 px-2 py-2 text-muted hover:border-brand-coral hover:text-ink max-sm:min-w-0 max-sm:flex-1"
                    type="button"
                  >
                    <strong className="block text-2xs">Again</strong>
                    <small className="mt-1 block text-2xs text-faint">now</small>
                  </button>
                  <button
                    onClick={() => rate('hard')}
                    className="min-w-16 rounded-lg border border-line bg-white/5 px-2 py-2 text-muted hover:border-brand-gold hover:text-ink max-sm:min-w-0 max-sm:flex-1"
                    type="button"
                  >
                    <strong className="block text-2xs">Hard</strong>
                    <small className="mt-1 block text-2xs text-faint">soon</small>
                  </button>
                  <button
                    onClick={() => rate('good')}
                    className="min-w-16 rounded-lg border border-line bg-white/5 px-2 py-2 text-muted hover:border-brand-teal hover:text-ink max-sm:min-w-0 max-sm:flex-1"
                    type="button"
                  >
                    <strong className="block text-2xs">Good</strong>
                    <small className="mt-1 block text-2xs text-faint">later</small>
                  </button>
                  <button
                    onClick={() => rate('easy')}
                    className="min-w-16 rounded-lg border border-line bg-white/5 px-2 py-2 text-muted hover:border-brand-teal hover:text-ink max-sm:min-w-0 max-sm:flex-1"
                    type="button"
                  >
                    <strong className="block text-2xs">Easy</strong>
                    <small className="mt-1 block text-2xs text-faint">further</small>
                  </button>
                </div>
              )}
            </div>
          </Card>
          <div className="mt-3 flex items-center gap-2 text-2xs text-faint">
            <span className="mt-1 size-2 shrink-0 rounded-full bg-brand-teal ring-4 ring-brand-teal/10" />
            <strong className="font-medium text-muted">{current.objectiveLabel}</strong>
            <Link className="ml-auto text-2xs text-brand-teal no-underline" to="/progress">
              View objective →
            </Link>
            <Link className="text-2xs text-brand-teal no-underline" to="/">
              Exit
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

function ReviewComplete({ reviewed, revisit }: { reviewed: number; revisit: number }) {
  return (
    <div className="grid min-h-[70vh] max-w-6xl content-center justify-items-center gap-7 text-center max-sm:min-h-[72vh] max-sm:gap-5">
      <div className="grid size-16 place-items-center rounded-full border border-brand-teal/40 bg-brand-teal/10 text-3xl text-brand-teal">
        ✓
      </div>
      <p className="mt-7 text-2xs font-bold uppercase tracking-widest text-faint">
        Session complete
      </p>
      <h1 className="my-2.5 text-5xl tracking-tight max-sm:text-4xl">Review complete.</h1>
      <p className="text-sm text-muted">
        You gave {reviewed} ideas your attention.{' '}
        {revisit ? `${revisit} are worth revisiting soon.` : 'Nice clean pass.'}
      </p>
      <div className="my-7 flex gap-px">
        <div className="grid min-w-32 gap-1 bg-brand-slate/10 p-4 max-sm:min-w-24 max-sm:px-2 max-sm:py-3">
          <strong className="text-2xl max-sm:text-xl">{reviewed}</strong>
          <span className="text-2xs uppercase tracking-wide text-faint">reviewed</span>
        </div>
        <div className="grid min-w-32 gap-1 bg-brand-slate/10 p-4 max-sm:min-w-24 max-sm:px-2 max-sm:py-3">
          <strong className="text-2xl max-sm:text-xl">{revisit}</strong>
          <span className="text-2xs uppercase tracking-wide text-faint">worth revisiting</span>
        </div>
      </div>
      <div className="flex gap-2">
        <Link
          className="inline-flex items-center justify-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-xs font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
          to="/"
        >
          Back home
        </Link>
        <Link
          className="inline-flex items-center justify-center gap-2.5 rounded-lg border border-line-strong bg-brand-slate/10 px-4 py-2.5 text-xs font-bold text-ink no-underline transition hover:bg-brand-slate/20"
          to="/progress"
        >
          View progress
        </Link>
      </div>
    </div>
  )
}

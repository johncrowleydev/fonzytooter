import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  getListReviewCardsQueryKey,
  useCreateReviewCardReview,
  useListReviewCards,
} from '../../api/generated/endpoints'
import type { RatingPreviewResource } from '../../api/generated/schemas/ratingPreviewResource.zod'
import type { ReviewCardResource } from '../../api/generated/schemas/reviewCardResource.zod'
import type { ReviewSubmission } from '../../api/generated/schemas/reviewSubmission.zod'
import { DEFAULT_COURSE_ID } from '../../app/routes'
import { Badge, Button, Card, PageIntro, ProgressBar } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

const ratingStyles: Record<ReviewSubmission['rating'], string> = {
  again:
    'min-w-20 rounded-lg border border-line bg-white/5 px-3 py-2 text-muted transition hover:border-brand-coral hover:text-ink disabled:cursor-wait disabled:opacity-60 max-sm:min-w-0 max-sm:flex-1 max-sm:px-2',
  hard: 'min-w-20 rounded-lg border border-line bg-white/5 px-3 py-2 text-muted transition hover:border-brand-gold hover:text-ink disabled:cursor-wait disabled:opacity-60 max-sm:min-w-0 max-sm:flex-1 max-sm:px-2',
  good: 'min-w-20 rounded-lg border border-line bg-white/5 px-3 py-2 text-muted transition hover:border-brand-teal hover:text-ink disabled:cursor-wait disabled:opacity-60 max-sm:min-w-0 max-sm:flex-1 max-sm:px-2',
  easy: 'min-w-20 rounded-lg border border-line bg-white/5 px-3 py-2 text-muted transition hover:border-brand-teal hover:text-ink disabled:cursor-wait disabled:opacity-60 max-sm:min-w-0 max-sm:flex-1 max-sm:px-2',
}

const ratingLabels: Record<ReviewSubmission['rating'], string> = {
  again: 'Again',
  hard: 'Hard',
  good: 'Good',
  easy: 'Easy',
}

export function Review() {
  const queryClient = useQueryClient()
  const { setPageContext } = useTutor()
  const dueCards = useListReviewCards(DEFAULT_COURSE_ID, { due: true })
  const createReview = useCreateReviewCardReview()
  const [sessionCards, setSessionCards] = useState<ReviewCardResource[] | null>(null)

  useEffect(() => {
    setPageContext({ type: 'review', title: 'Review', courseId: DEFAULT_COURSE_ID })
  }, [setPageContext])

  useEffect(() => {
    if (sessionCards === null && dueCards.data?.data) {
      setSessionCards(dueCards.data.data)
    }
  }, [dueCards.data?.data, sessionCards])

  if (dueCards.isError) {
    return (
      <ReviewStatus
        title="Reviews are unavailable"
        detail="The review queue could not be loaded. Your existing scheduling state was not changed."
        action={<Button onClick={() => void dueCards.refetch()}>Try again</Button>}
      />
    )
  }
  if (dueCards.isPending || sessionCards === null) {
    return <ReviewStatus title="Loading reviews…" detail="Checking the current review queue." />
  }
  if (sessionCards.length === 0) {
    return <NoDueReviews />
  }

  return (
    <ReviewSessionView
      cards={sessionCards}
      onRate={async (card, rating) => {
        await createReview.mutateAsync({
          courseId: card.courseId,
          reviewItemId: card.id,
          data: { rating },
        })
        await queryClient.invalidateQueries({
          queryKey: getListReviewCardsQueryKey(DEFAULT_COURSE_ID, { due: true }),
        })
      }}
    />
  )
}

export function ReviewSessionView({
  cards,
  onRate,
}: {
  cards: ReviewCardResource[]
  onRate: (card: ReviewCardResource, rating: ReviewSubmission['rating']) => Promise<void>
}) {
  const { setPageContext } = useTutor()
  const [cardIndex, setCardIndex] = useState(0)
  const [revealed, setRevealed] = useState(false)
  const [ratings, setRatings] = useState<ReviewSubmission['rating'][]>([])
  const [pendingRating, setPendingRating] = useState<ReviewSubmission['rating'] | null>(null)
  const [submitError, setSubmitError] = useState(false)
  const current = cards[cardIndex]
  const complete = cardIndex >= cards.length

  useEffect(() => {
    setPageContext({
      type: 'review',
      title: 'Review',
      courseId: current?.courseId,
      objectiveId: current?.objectiveIds[0],
    })
  }, [current?.courseId, current?.objectiveIds, setPageContext])

  const remaining = useMemo(() => Math.max(cards.length - cardIndex, 0), [cardIndex, cards.length])
  const progress = Math.round((Math.min(cardIndex, cards.length) / cards.length) * 100)

  if (complete) {
    return (
      <ReviewComplete
        reviewed={ratings.length}
        revisit={ratings.filter((rating) => rating === 'again' || rating === 'hard').length}
      />
    )
  }

  async function rate(rating: ReviewSubmission['rating']) {
    setPendingRating(rating)
    setSubmitError(false)
    try {
      await onRate(current, rating)
      setRatings((items) => [...items, rating])
      setRevealed(false)
      setCardIndex((index) => index + 1)
    } catch {
      setSubmitError(true)
    } finally {
      setPendingRating(null)
    }
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
          {cardIndex + 1} / {cards.length}
        </span>
      </div>
      <div className="grid justify-center">
        <div className="w-full max-w-3xl">
          <Card className="flex min-h-96 flex-col p-6 max-sm:min-h-104 max-sm:p-5">
            <div className="flex items-center justify-between gap-3">
              <Badge tone="teal">{current.objectiveIds[0]}</Badge>
              <span className="text-2xs text-faint">
                {current.lastReviewedAt
                  ? `Last reviewed ${formatLocalDate(current.lastReviewedAt)}`
                  : 'New review item'}
              </span>
            </div>
            <div className="mx-2.5 my-auto grid justify-items-center gap-3.5 py-8 text-center">
              <span className="grid size-11 place-items-center rounded-full border border-brand-teal/40 font-serif text-xl text-brand-teal">
                ?
              </span>
              <h2 className="max-w-xl text-2xl font-semibold leading-tight tracking-tight sm:text-3xl">
                {current.prompt}
              </h2>
            </div>
            {revealed ? (
              <div className="border-t border-line bg-brand-teal/5 px-4 py-5">
                <p className="text-2xs font-bold uppercase tracking-widest text-faint">Answer</p>
                <p className="my-2 text-sm leading-relaxed text-ink">{current.answer}</p>
              </div>
            ) : current.hint ? (
              <div className="border-t border-line pt-4 text-xs text-faint">
                <span className="mr-2 text-2xs font-bold uppercase tracking-wide text-brand-gold">
                  Hint
                </span>
                {current.hint}
              </div>
            ) : null}
            <div className="mt-4 flex justify-center">
              {!revealed ? (
                <Button onClick={() => setRevealed(true)}>
                  Reveal answer <span>↓</span>
                </Button>
              ) : (
                <div className="grid w-full gap-3">
                  <p className="text-center text-2xs text-faint">How well did you recall it?</p>
                  <div className="flex w-full items-center justify-center gap-2 max-sm:gap-1">
                    {current.previews.map((preview) => (
                      <RatingButton
                        key={preview.rating}
                        preview={preview}
                        pending={pendingRating === preview.rating}
                        disabled={pendingRating !== null}
                        onClick={() => void rate(preview.rating)}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
            {submitError ? (
              <p role="alert" className="mt-4 text-center text-xs text-brand-coral">
                This rating could not be saved. The current card is still here; try again.
              </p>
            ) : null}
          </Card>
          <div className="mt-3 flex items-center gap-2 text-2xs text-faint">
            <span className="mt-1 size-2 shrink-0 rounded-full bg-brand-teal ring-4 ring-brand-teal/10" />
            <strong className="font-medium text-muted">FSRS scheduling</strong>
            <Link className="ml-auto text-brand-teal no-underline" to="/progress">
              View progress →
            </Link>
            <Link className="text-brand-teal no-underline" to="/">
              Exit
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

function RatingButton({
  preview,
  pending,
  disabled,
  onClick,
}: {
  preview: RatingPreviewResource
  pending: boolean
  disabled: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={ratingStyles[preview.rating]}
      type="button"
      disabled={disabled}
      title={`Next due ${formatLocalDate(preview.dueAt)}`}
    >
      <strong className="block text-2xs">
        {pending ? 'Saving…' : ratingLabels[preview.rating]}
      </strong>
      <small className="mt-1 block text-2xs text-faint">
        {formatInterval(preview.intervalSeconds)}
      </small>
    </button>
  )
}

export function NoDueReviews() {
  return (
    <ReviewStatus
      title="No reviews due"
      detail="There is nothing in the current FSRS queue. Continue learning and return when a card is due."
      action={
        <Link
          className="inline-flex items-center justify-center rounded-lg bg-brand-teal px-4 py-2.5 text-xs font-bold text-brand-ink no-underline"
          to="/curriculum"
        >
          Continue learning
        </Link>
      }
    />
  )
}

function ReviewStatus({
  title,
  detail,
  action,
}: {
  title: string
  detail: string
  action?: ReactNode
}) {
  return (
    <div className="grid min-h-[65vh] content-center justify-items-center gap-4 text-center">
      <span className="grid size-14 place-items-center rounded-full border border-brand-teal/40 bg-brand-teal/10 text-2xl text-brand-teal">
        ↺
      </span>
      <h1 className="text-4xl tracking-tight max-sm:text-3xl">{title}</h1>
      <p className="max-w-lg text-sm leading-relaxed text-muted">{detail}</p>
      {action ? <div className="mt-3">{action}</div> : null}
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
        You reviewed {reviewed} {reviewed === 1 ? 'idea' : 'ideas'}.{' '}
        {revisit ? `${revisit} will return sooner.` : 'The scheduler recorded a clean pass.'}
      </p>
      <div className="my-7 flex gap-px">
        <div className="grid min-w-32 gap-1 bg-brand-slate/10 p-4 max-sm:min-w-24 max-sm:px-2 max-sm:py-3">
          <strong className="text-2xl max-sm:text-xl">{reviewed}</strong>
          <span className="text-2xs uppercase tracking-wide text-faint">reviewed</span>
        </div>
        <div className="grid min-w-32 gap-1 bg-brand-slate/10 p-4 max-sm:min-w-24 max-sm:px-2 max-sm:py-3">
          <strong className="text-2xl max-sm:text-xl">{revisit}</strong>
          <span className="text-2xs uppercase tracking-wide text-faint">returning sooner</span>
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

function formatInterval(seconds: number) {
  if (seconds < 90) return '1 min'
  if (seconds < 60 * 60) return `${Math.round(seconds / 60)} min`
  if (seconds < 36 * 60 * 60) return `${Math.round(seconds / (60 * 60))} hr`
  const days = Math.round(seconds / (24 * 60 * 60))
  return `${days} ${days === 1 ? 'day' : 'days'}`
}

function formatLocalDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

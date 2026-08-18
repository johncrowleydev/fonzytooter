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
    <div className="page-stack review-page">
      <PageIntro compact title="Review">
        <div className="review-session-meta">
          <span>
            <strong>{remaining}</strong> remaining
          </span>
          <span>·</span>
          <span>Approx. 8 min</span>
        </div>
      </PageIntro>
      <div className="review-progress">
        <span>Session progress</span>
        <ProgressBar value={progress} tone="teal" />
        <span>
          {Math.min(cardIndex + 1, reviewCards.length)} / {reviewCards.length}
        </span>
      </div>
      <div className="review-layout">
        <div>
          <Card className={`review-card ${revealed ? 'revealed' : ''}`}>
            <div className="review-card-top">
              <Badge tone="teal">{current.objectiveLabel}</Badge>
              <span>Last reviewed {current.lastReviewed}</span>
            </div>
            <div className="review-prompt">
              <span className="prompt-mark">?</span>
              <h2>{current.prompt}</h2>
            </div>
            {revealed ? (
              <div className="review-answer">
                <p className="eyebrow">Answer</p>
                <p>{current.answer}</p>
              </div>
            ) : (
              <div className="review-hint">
                <span>Hint</span>
                {current.hint}
              </div>
            )}
            <div className="review-card-action">
              {!revealed ? (
                <Button onClick={() => setRevealed(true)}>
                  Reveal answer <span>↓</span>
                </Button>
              ) : (
                <div className="rating-row">
                  <span className="rating-label">How did that feel?</span>
                  <button onClick={() => rate('again')} className="rating again" type="button">
                    <strong>Again</strong>
                    <small>now</small>
                  </button>
                  <button onClick={() => rate('hard')} className="rating hard" type="button">
                    <strong>Hard</strong>
                    <small>soon</small>
                  </button>
                  <button onClick={() => rate('good')} className="rating good" type="button">
                    <strong>Good</strong>
                    <small>later</small>
                  </button>
                  <button onClick={() => rate('easy')} className="rating easy" type="button">
                    <strong>Easy</strong>
                    <small>further</small>
                  </button>
                </div>
              )}
            </div>
          </Card>
          <div className="review-context">
            <span className="context-pulse" />
            <strong>{current.objectiveLabel}</strong>
            <button type="button" onClick={() => navigate('/progress')}>
              View objective →
            </button>
            <button className="text-link" type="button" onClick={() => navigate('/')}>
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
    <div className="page-stack completion-page">
      <div className="completion-mark">✓</div>
      <p className="eyebrow">Session complete</p>
      <h1>Review complete.</h1>
      <p className="completion-lead">
        You gave {reviewed} ideas your attention.{' '}
        {revisit ? `${revisit} are worth revisiting soon.` : 'Nice clean pass.'}
      </p>
      <div className="completion-stats">
        <div>
          <strong>{reviewed}</strong>
          <span>reviewed</span>
        </div>
        <div>
          <strong>{revisit}</strong>
          <span>worth revisiting</span>
        </div>
        <div>
          <strong>3m</strong>
          <span>focused time</span>
        </div>
      </div>
      <div className="completion-actions">
        <Button onClick={() => navigate('/')}>Back home</Button>
        <Button variant="secondary" onClick={() => navigate('/progress')}>
          View progress
        </Button>
      </div>
    </div>
  )
}

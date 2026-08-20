import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReviewCardResource } from '../../api/generated/schemas/reviewCardResource.zod'
import { TutorProvider } from '../tutor/TutorContext'
import { NoDueReviews, ReviewSessionView } from './Review'

afterEach(cleanup)

const card: ReviewCardResource = {
  courseId: 'ai-ml',
  moduleId: 'foundations',
  id: 'neutral-card',
  order: 0,
  objectiveIds: ['objective.neutral'],
  sourceLessonId: 'neutral-lesson',
  prompt: 'What is the neutral prompt?',
  answer: 'The neutral answer.',
  hint: 'Use the neutral hint.',
  state: 'new',
  dueAt: '2026-08-20T14:30:00Z',
  virtual: true,
  due: true,
  previews: [
    {
      rating: 'again',
      dueAt: '2026-08-20T14:31:00Z',
      intervalSeconds: 60,
      intervalDays: 1 / 1440,
    },
    {
      rating: 'hard',
      dueAt: '2026-08-20T14:35:30Z',
      intervalSeconds: 330,
      intervalDays: 330 / 86400,
    },
    {
      rating: 'good',
      dueAt: '2026-08-20T14:40:00Z',
      intervalSeconds: 600,
      intervalDays: 1 / 144,
    },
    {
      rating: 'easy',
      dueAt: '2026-08-23T14:30:00Z',
      intervalSeconds: 259200,
      intervalDays: 3,
    },
  ],
}

describe('review session', () => {
  it('keeps the answer and real rating previews hidden until reveal', async () => {
    renderReview(vi.fn().mockResolvedValue(undefined))

    expect(screen.queryByText(card.answer)).toBeNull()
    expect(screen.queryByRole('button', { name: /Again/ })).toBeNull()
    await userEvent.click(screen.getByRole('button', { name: /Reveal answer/ }))

    expect(screen.getByText(card.answer)).toBeDefined()
    expect(screen.getByRole('button', { name: /Again1 min/ })).toBeDefined()
    expect(screen.getByRole('button', { name: /Hard6 min/ })).toBeDefined()
    expect(screen.getByRole('button', { name: /Good10 min/ })).toBeDefined()
    expect(screen.getByRole('button', { name: /Easy3 days/ })).toBeDefined()
  })

  it('shows pending submission and advances only after success', async () => {
    let resolveSubmission: (() => void) | undefined
    const onRate = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveSubmission = resolve
        }),
    )
    renderReview(onRate)
    await userEvent.click(screen.getByRole('button', { name: /Reveal answer/ }))
    await userEvent.click(screen.getByRole('button', { name: /Good10 min/ }))

    expect(
      (screen.getByRole('button', { name: /Saving…10 min/ }) as HTMLButtonElement).disabled,
    ).toBe(true)
    expect(screen.getByText(card.prompt)).toBeDefined()
    resolveSubmission?.()
    await waitFor(() => expect(screen.getByText('Review complete.')).toBeDefined())
    expect(onRate).toHaveBeenCalledWith(card, 'good')
  })

  it('retains the current revealed card when submission fails', async () => {
    renderReview(vi.fn().mockRejectedValue(new Error('offline')))
    await userEvent.click(screen.getByRole('button', { name: /Reveal answer/ }))
    await userEvent.click(screen.getByRole('button', { name: /Hard6 min/ }))

    expect((await screen.findByRole('alert')).textContent).toContain('current card is still here')
    expect(screen.getByText(card.prompt)).toBeDefined()
    expect(screen.getByText(card.answer)).toBeDefined()
    expect((screen.getByRole('button', { name: /Hard6 min/ }) as HTMLButtonElement).disabled).toBe(
      false,
    )
  })

  it('renders the no-due queue state truthfully', () => {
    render(
      <MemoryRouter>
        <NoDueReviews />
      </MemoryRouter>,
    )
    expect(screen.getByText('No reviews due')).toBeDefined()
    expect(screen.getByRole('link', { name: 'Continue learning' })).toBeDefined()
  })
})

function renderReview(
  onRate: (
    current: ReviewCardResource,
    rating: 'again' | 'hard' | 'good' | 'easy',
  ) => Promise<void>,
) {
  return render(
    <MemoryRouter>
      <TutorProvider>
        <ReviewSessionView cards={[card]} onRate={onRate} />
      </TutorProvider>
    </MemoryRouter>,
  )
}

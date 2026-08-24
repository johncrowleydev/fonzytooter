import { useState } from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { CourseProgressResource } from '../api/generated/schemas/courseProgressResource.zod'
import { ActivityList, DashboardView } from './dashboard/Dashboard'
import { LessonCompletionControl } from './lessons/LessonCompletionControl'
import { ProgressView } from './progress/Progress'

afterEach(cleanup)

const progress: CourseProgressResource = {
  courseId: 'ai-ml',
  completedLessonCount: 1,
  totalLessonCount: 2,
  dueReviewCount: 1,
  practiceExercise: {
    courseId: 'ai-ml',
    moduleId: 'foundations',
    moduleTitle: 'Foundations',
    exerciseId: 'exercise.one',
    exerciseTitle: 'First exercise',
  },
  objectives: [
    {
      courseId: 'ai-ml',
      moduleId: 'foundations',
      id: 'objective.one',
      title: 'First objective',
      description: 'The first objective.',
      introduced: true,
      linkedLessonCount: 1,
      completedLessonCount: 1,
      recall: { reviewItemCount: 1, reviewsCompleted: 2, dueReviewCount: 1 },
      application: { exerciseCount: 1, attempts: 1, fullyPassedAttempts: 0 },
      transferAssessed: false,
    },
    {
      courseId: 'ai-ml',
      moduleId: 'foundations',
      id: 'objective.two',
      title: 'Second objective',
      description: 'The second objective.',
      introduced: false,
      linkedLessonCount: 1,
      completedLessonCount: 0,
      recall: { reviewItemCount: 0, reviewsCompleted: 0, dueReviewCount: 0 },
      application: { exerciseCount: 0, attempts: 0, fullyPassedAttempts: 0 },
      transferAssessed: false,
    },
  ],
  nextLesson: {
    courseId: 'ai-ml',
    moduleId: 'foundations',
    moduleTitle: 'Foundations',
    lessonId: 'lesson-two',
    lessonTitle: 'Second lesson',
  },
}

describe('lesson completion UI', () => {
  it('offers sign-in instead of a progress mutation while anonymous', () => {
    const onChange = vi.fn()
    render(
      <MemoryRouter initialEntries={['/courses/ai-ml/modules/python/lessons/intro']}>
        <LessonCompletionControl
          authenticated={false}
          completed={false}
          pending={false}
          onChange={onChange}
        />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'Sign in to track progress' })).toBeDefined()
    expect(onChange).not.toHaveBeenCalled()
  })

  it('requires an explicit click to mark a lesson complete', async () => {
    const onChange = vi.fn()
    render(<LessonCompletionControl completed={false} pending={false} onChange={onChange} />)

    expect(onChange).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Mark complete' }))
    expect(onChange).toHaveBeenCalledWith(true)
  })

  it('offers an explicit undo for completed lessons', async () => {
    const onChange = vi.fn()
    render(<LessonCompletionControl completed pending={false} onChange={onChange} />)

    await userEvent.click(screen.getByRole('button', { name: 'Mark incomplete' }))
    expect(onChange).toHaveBeenCalledWith(false)
  })
})

describe('truthful progress UI', () => {
  it('shows factual recall and application evidence without mastery claims', async () => {
    function Harness() {
      const [selectedId, setSelectedId] = useState('objective.one')
      return <ProgressView progress={progress} selectedId={selectedId} onSelect={setSelectedId} />
    }

    render(<Harness />)
    expect(screen.getByText('2 reviews · 1 due')).toBeDefined()
    expect(screen.getByText('1 attempts · 0 passed')).toBeDefined()
    expect(screen.getAllByText('Not assessed').length).toBeGreaterThanOrEqual(1)
    expect(screen.queryByText('Mastered')).toBeNull()
    expect(screen.queryByText('Strong')).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: /Second objective/ }))
    expect(screen.getAllByText('Not introduced').length).toBeGreaterThan(0)
  })
})

describe('dashboard learner state', () => {
  it('links to the derived next incomplete lesson', () => {
    render(
      <MemoryRouter>
        <DashboardView progress={progress} activities={[]} />
      </MemoryRouter>,
    )

    const link = screen.getByRole('link', { name: /Continue lesson/ })
    expect(link.getAttribute('href')).toBe('/courses/ai-ml/modules/foundations/lessons/lesson-two')
    expect(screen.getAllByRole('link', { name: /Open/ })[0].getAttribute('href')).toBe('/review')
    expect(screen.getByText('No activity yet')).toBeDefined()
  })

  it('shows a restrained completed state when no next lesson exists', () => {
    render(
      <MemoryRouter>
        <DashboardView
          progress={{
            ...progress,
            completedLessonCount: 2,
            nextLesson: undefined,
          }}
          activities={[]}
        />
      </MemoryRouter>,
    )

    expect(screen.getByText('All current lessons complete')).toBeDefined()
    expect(screen.getByText('1 due now')).toBeDefined()
  })

  it('renders an explicit empty activity state', () => {
    render(<ActivityList activities={[]} />)
    expect(screen.getByText('No activity yet')).toBeDefined()
  })

  it('renders exercise checks truthfully with a course-aware link', () => {
    render(
      <MemoryRouter>
        <ActivityList
          activities={[
            {
              id: 9,
              kind: 'exercise_checked',
              courseId: 'ai-ml',
              courseTitle: 'AI & Machine Learning',
              moduleId: 'scientific-python',
              moduleTitle: 'Scientific Python',
              exerciseId: 'python.double',
              occurredAt: '2026-08-20T12:00:00Z',
            },
          ]}
        />
      </MemoryRouter>,
    )

    const link = screen.getByRole('link', { name: /Checked python.double/ })
    expect(link.getAttribute('href')).toBe(
      '/courses/ai-ml/modules/scientific-python/exercises/python.double',
    )
  })
})

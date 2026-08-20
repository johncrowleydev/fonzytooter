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
  objectives: [
    {
      courseId: 'ai-ml',
      moduleId: 'foundations',
      id: 'objective.one',
      title: 'First objective',
      description: 'The first objective.',
      introduced: true,
      recall: 'not_assessed',
      application: 'not_assessed',
      transfer: 'not_assessed',
    },
    {
      courseId: 'ai-ml',
      moduleId: 'foundations',
      id: 'objective.two',
      title: 'Second objective',
      description: 'The second objective.',
      introduced: false,
      recall: 'not_assessed',
      application: 'not_assessed',
      transfer: 'not_assessed',
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
  it('shows only introduction and not-assessed evidence states', async () => {
    function Harness() {
      const [selectedId, setSelectedId] = useState('objective.one')
      return <ProgressView progress={progress} selectedId={selectedId} onSelect={setSelectedId} />
    }

    render(<Harness />)
    expect(screen.getAllByText('Not assessed').length).toBeGreaterThanOrEqual(3)
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
    expect(screen.queryByText(/reviews ready/i)).toBeNull()
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
